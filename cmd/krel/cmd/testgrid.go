/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/gcp/gcs"
	"k8s.io/release/pkg/http"
	"k8s.io/release/pkg/util"
	"k8s.io/utils/pointer"
)

type TestGridOptions struct {
	branch        string
	boards        []string
	states        []string
	bucket        string
	testgridURL   string
	renderTronURL string
}

var testGridOpts = &TestGridOptions{}

type TestgridJobInfo struct {
	OverallStatus string `mapstructure:"overall_status"`
	DashboardName string `mapstructure:"dashboard_name"`
}

type TestGridJob struct {
	DashboardName string
	JobName       string
	Status        string
	GCSLocation   string
}

const (
	statePassing = "PASSING"
	stateFlaky   = "FLAKY"
	stateFailing = "FAILING"

	boardInforming = "informing"
	boardBlocking  = "blocking"

	layoutISO = "2006-01-02"
)

// ffCmd represents the base command when called without any subcommands
var testGridCmd = &cobra.Command{
	Use:           "testgridshot --branch <release-branch>",
	Short:         "Take a screenshot of the testgrid dashboards",
	Example:       "krel testgridshot --branch 1.17",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTestGridShot(testGridOpts)
	},
}

func init() {
	testGridCmd.PersistentFlags().StringVar(&testGridOpts.branch, "branch",
		"master", "From which release branch will get the testgrid dashboard")

	testGridCmd.PersistentFlags().StringSliceVar(&testGridOpts.boards, "boards", []string{boardBlocking, boardInforming},
		"Which Boards to retrieve the dashboards, defaults to blocking and informing")

	testGridCmd.PersistentFlags().StringSliceVar(&testGridOpts.states, "states", []string{stateFailing},
		"Which States to watch for each dashboard, default to failing")

	testGridCmd.PersistentFlags().StringVar(&testGridOpts.testgridURL,
		"testgrid-url", "https://testgrid.k8s.io", "The TestGrid URL")

	testGridCmd.PersistentFlags().StringVar(&testGridOpts.renderTronURL,
		"rendertron-url", "https://render-tron.appspot.com/screenshot", "The RenderTron URL service")

	testGridCmd.PersistentFlags().StringVar(&testGridOpts.bucket, "bucket", "k8s-staging-releng",
		"The name of the bucket to upload the images to. The files will be put into '/testgridshot/<release>/<datetime>_<rand>/...'. Defaults to k8s-staging-releng")

	rootCmd.AddCommand(testGridCmd)
}

func runTestGridShot(opts *TestGridOptions) error {
	logrus.Info("Starting krel testgrishot...")

	if err := opts.Validate(); err != nil {
		return errors.Wrap(err, "validating testgridshot options")
	}

	testgridJobs := []TestGridJob{}
	for _, board := range opts.boards {
		testGridDashboard := fmt.Sprintf("%s/sig-release-%s-%s/summary", opts.testgridURL, opts.branch, board)
		content, err := http.GetURLResponse(testGridDashboard, false)
		if err != nil {
			return errors.Wrapf(err,
				"unable to retrieve release announcement form url: %s", testGridDashboard,
			)
		}

		var result map[string]interface{}
		err = json.Unmarshal([]byte(content), &result)
		if err != nil {
			return errors.Wrap(err, "unable unmarshal the testgrid response")
		}

		testgridJobsTemp := []TestGridJob{}
		for jobName, jobData := range result {
			result := TestgridJobInfo{}
			err = mapstructure.Decode(jobData, &result)
			if err != nil {
				return errors.Wrap(err, "failed to decode testgrid data")
			}

			for _, state := range opts.states {
				if state == result.OverallStatus {
					testgridJobsTemp = append(testgridJobsTemp, TestGridJob{
						JobName:       jobName,
						DashboardName: result.DashboardName,
						Status:        result.OverallStatus,
					})
				}
			}
		}
		testgridJobs = append(testgridJobs, testgridJobsTemp...)

		dateNow := fmt.Sprintf("%s_%s", time.Now().UTC().Format(layoutISO), util.RandomString(6))
		testgridJobs, err = processDashboards(testgridJobs, dateNow, opts)
		if err != nil {
			return errors.Wrap(err, "failed to process the dashboards")
		}
	}

	generateIssueComment(testgridJobs, opts)
	return nil
}

func processDashboards(testgridJobs []TestGridJob, date string, opts *TestGridOptions) ([]TestGridJob, error) {
	for i, job := range testgridJobs {
		testGridJobURL := fmt.Sprintf("%s/%s#%s&width=30", opts.testgridURL, job.DashboardName, job.JobName)
		rendertronURL := fmt.Sprintf("%s/%s?width=3000&height=2500", opts.renderTronURL, url.PathEscape(testGridJobURL))
		logrus.Infof("rendertronURL for %s: %s", testGridJobURL, rendertronURL)

		content, err := http.GetURLResponseWithTimeOut(rendertronURL, 300*time.Second)
		if err != nil {
			return testgridJobs, errors.Wrapf(err, "failed to get the testgrid screenshot")
		}

		jobFile := fmt.Sprintf("/tmp/%s-%s-%s.jpg", job.DashboardName, strings.ReplaceAll(job.JobName, " ", "_"), job.Status)
		f, err := os.Create(jobFile)
		if err != nil {
			return testgridJobs, errors.Wrapf(err, "failed to create the file %s", jobFile)
		}
		defer f.Close()
		defer os.Remove(jobFile)

		_, err = f.Write(content)
		if err != nil {
			return testgridJobs, errors.Wrapf(err, "failed to write the content to the file %s", jobFile)
		}
		logrus.Infof("Screenshot saved for %s: %s", job.JobName, jobFile)

		gcsPath := filepath.Join(opts.bucket, "testgridshot", opts.branch, date, filepath.Base(jobFile))

		GCSCopyOptions := &gcs.Options{
			Concurrent:   pointer.BoolPtr(true),
			Recursive:    pointer.BoolPtr(true),
			NoClobber:    pointer.BoolPtr(true),
			AllowMissing: pointer.BoolPtr(true),
		}
		if err := gcs.CopyToGCS(jobFile, gcsPath, GCSCopyOptions); err != nil {
			return testgridJobs, errors.Wrapf(err, "failed to upload the file %s to GCS bucket %s", jobFile, gcsPath)
		}
		testgridJobs[i].GCSLocation = fmt.Sprintf("https://storage.googleapis.com/%s", gcsPath)
		logrus.Infof("Screenshot will be available for job %s at %s", job.JobName, testgridJobs[i].GCSLocation)
	}

	return testgridJobs, nil
}

func generateIssueComment(testgridJobs []TestGridJob, opts *TestGridOptions) {
	// Generate comment to GH
	output := []string{}
	output = append(output, fmt.Sprintf("<!-- ----[ issue comment ]---- -->\n### Testgrid dashboards for %s\n", opts.branch))

	timeNow := time.Now().UTC()
	for _, state := range opts.states {
		output = append(output, fmt.Sprintf("Boards checked for %s:", state))
		for _, board := range opts.boards {
			output = append(output, fmt.Sprintf("- [sig-release-%[1]s-%[2]s](%[3]s/sig-release-%[1]s-%[2]s)", opts.branch, board, opts.testgridURL))
		}
		output = append(output, "\n")

		haveState := false
		for _, job := range testgridJobs {
			if state == job.Status {
				output = append(output,
					fmt.Sprintf(
						"<details><summary><tt>%[1]s</tt> %[2]s <a href=\"%[5]s/%[3]s#%[4]s&width=30\">%[3]s#%[4]s - TestGrid</a></summary><p>\n"+
							"![%[3]s#%[4]s](%[6]s)\n"+
							"</p></details>",
						timeNow, state, job.DashboardName, job.JobName, opts.testgridURL, job.GCSLocation),
				)

				haveState = true
			}
		}

		if !haveState {
			output = append(output, fmt.Sprintf("**No %s Jobs**", state))
		}

		output = append(output, "\n\n")
	}

	output = append(output, "<!-- ----[ issue comment ]---- -->")

	logrus.Info("Please copy the lines below and paste in the Github Issue for the Release cut. Thanks for using krel!")
	fmt.Println(strings.Join(output, "\n"))
}

func (o *TestGridOptions) Validate() error {
	for i, state := range o.states {
		o.states[i] = strings.ToUpper(state)
		if o.states[i] != stateFailing &&
			o.states[i] != statePassing &&
			o.states[i] != stateFlaky {
			return errors.New(
				fmt.Sprintf("invalid state %s option. Valid options are: %s, %s, %s",
					o.states[i], stateFailing, stateFlaky, statePassing),
			)
		}
	}

	for _, board := range o.boards {
		if board != boardBlocking &&
			board != boardInforming {
			return errors.New(
				fmt.Sprintf("invalid board %s option. Valid options are: %s, %s",
					board, boardBlocking, boardInforming),
			)
		}
	}

	return nil
}

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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-sdk/github"
	"sigs.k8s.io/release-sdk/object"
	"sigs.k8s.io/release-utils/http"
)

const (
	k8sSigReleaseRepo = "sig-release"
)

type TestGridOptions struct {
	branch      string
	boards      []string
	states      []string
	bucket      string
	testgridURL string
	gitHubIssue int
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

// testGridCmd represents the base command when called without any subcommands
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
		git.DefaultBranch, "From which release branch will get the testgrid dashboard")

	testGridCmd.PersistentFlags().StringSliceVar(&testGridOpts.boards, "boards", []string{boardBlocking, boardInforming},
		"Which Boards to retrieve the dashboards, defaults to blocking and informing")

	testGridCmd.PersistentFlags().StringSliceVar(&testGridOpts.states, "states", []string{stateFailing},
		"Which States to watch for each dashboard, default to failing")

	testGridCmd.PersistentFlags().StringVar(&testGridOpts.testgridURL,
		"testgrid-url", "https://testgrid.k8s.io", "The TestGrid URL")

	testGridCmd.PersistentFlags().IntVar(&testGridOpts.gitHubIssue,
		"github-issue", -1, "The GitHub Issue for the release cut")

	testGridCmd.PersistentFlags().StringVar(&testGridOpts.bucket, "bucket", "k8s-staging-releng",
		"The name of the bucket to upload the images to. The files will be put into '/testgridshot/<release>/<datetime>_<rand>/...'. Defaults to k8s-staging-releng")

	rootCmd.AddCommand(testGridCmd)
}

func runTestGridShot(opts *TestGridOptions) error {
	logrus.Info("Starting krel testgrishot...")

	if err := opts.Validate(); err != nil {
		return fmt.Errorf("validating testgridshot options: %w", err)
	}

	testgridJobs := []TestGridJob{}
	for _, board := range opts.boards {
		testGridDashboard := fmt.Sprintf("%s/sig-release-%s-%s/summary", opts.testgridURL, opts.branch, board)
		content, err := http.NewAgent().WithTimeout(30 * time.Second).Get(testGridDashboard)
		if err != nil {
			return fmt.Errorf("unable to retrieve release announcement form url: %s: %w", testGridDashboard, err)
		}

		var result map[string]interface{}
		err = json.Unmarshal(content, &result)
		if err != nil {
			return fmt.Errorf("unable unmarshal the testgrid response: %w", err)
		}

		testgridJobsTemp := []TestGridJob{}
		for jobName, jobData := range result {
			result := TestgridJobInfo{}
			err = mapstructure.Decode(jobData, &result)
			if err != nil {
				return fmt.Errorf("decode testgrid data: %w", err)
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

		dateNow := fmt.Sprintf("%s-%s", time.Now().UTC().Format(layoutISO), uuid.NewString())
		testgridJobs, err = processDashboards(testgridJobs, dateNow, opts)
		if err != nil {
			return fmt.Errorf("processing the dashboards: %w", err)
		}
	}

	err := generateIssueComment(testgridJobs, opts)
	if err != nil {
		return fmt.Errorf("generating the GitHub issue comment: %w", err)
	}

	return nil
}

func processDashboards(testgridJobs []TestGridJob, date string, opts *TestGridOptions) ([]TestGridJob, error) {
	for i, job := range testgridJobs {
		testGridJobURL := fmt.Sprintf("%s/%s#%s&width=30", opts.testgridURL, job.DashboardName, job.JobName)
		screenshot, err := createScreenshot(testGridJobURL)
		if err != nil {
			return testgridJobs, fmt.Errorf("creating screenshot of %q: %w", testGridJobURL, err)
		}

		jobFile := fmt.Sprintf("/tmp/%s-%s-%s.jpg", job.DashboardName, strings.ReplaceAll(job.JobName, " ", "_"), job.Status)
		f, err := os.Create(jobFile)
		if err != nil {
			return testgridJobs, fmt.Errorf("failed to create the file %s: %w", jobFile, err)
		}

		_, err = f.Write(screenshot)
		f.Close()
		if err != nil {
			return testgridJobs, fmt.Errorf("failed to write the content to the file %s: %w", jobFile, err)
		}
		logrus.Infof("Screenshot saved for %s: %s", job.JobName, jobFile)

		gcsPath := filepath.Join(opts.bucket, "testgridshot", opts.branch, date, filepath.Base(jobFile))

		gcs := object.NewGCS()
		if err := gcs.CopyToRemote(jobFile, gcsPath); err != nil {
			return testgridJobs, fmt.Errorf("failed to upload the file %s to GCS bucket %s: %w", jobFile, gcsPath, err)
		}
		if err := os.Remove(jobFile); err != nil {
			return testgridJobs, fmt.Errorf("remove jobfile: %w", err)
		}
		testgridJobs[i].GCSLocation = fmt.Sprintf("https://storage.googleapis.com/%s", gcsPath)
		logrus.Infof("Screenshot will be available for job %s at %s", job.JobName, testgridJobs[i].GCSLocation)
	}

	return testgridJobs, nil
}

func createScreenshot(testGridJobURL string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var screenshot []byte

	screenshotTasks := chromedp.Tasks{
		chromedp.EmulateViewport(3000, 2500),
		chromedp.Navigate(testGridJobURL),
		chromedp.Sleep(5 * time.Second), // Give time to testgrid to load completely
		chromedp.CaptureScreenshot(&screenshot),
	}

	if err := chromedp.Run(ctx, screenshotTasks); err != nil {
		return nil, err
	}

	logrus.Infof("Successfully screenshotted testgrid dashboard %q", testGridJobURL)

	return screenshot, nil
}

func generateIssueComment(testgridJobs []TestGridJob, opts *TestGridOptions) error {
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
							"\n![%[3]s#%[4]s](%[6]s)\n\n"+
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

	output = append(output, "\n**comment generated by [krel](https://github.com/kubernetes/release/tree/master/docs/krel)**\n\n<!-- ----[ issue comment ]---- -->")

	if opts.gitHubIssue != -1 {
		gh := github.New()

		_, _, err := gh.Client().CreateComment(context.Background(), git.DefaultGithubOrg, k8sSigReleaseRepo, opts.gitHubIssue, strings.Join(output, "\n"))
		if err != nil {
			return fmt.Errorf("creating the GitHub comment: %w", err)
		}
		logrus.Infof("Comment created in the GitHub Issue https://github.com/%s/%s/issues/%d. Thanks for using krel!", git.DefaultGithubOrg, k8sSigReleaseRepo, opts.gitHubIssue)
	} else {
		logrus.Info("Please copy the lines below and paste in the Github Issue for the Release cut. Thanks for using krel!")
		fmt.Println(strings.Join(output, "\n"))
	}

	return nil
}

func (o *TestGridOptions) Validate() error {
	for i, state := range o.states {
		o.states[i] = strings.ToUpper(state)
		if o.states[i] != stateFailing &&
			o.states[i] != statePassing &&
			o.states[i] != stateFlaky {
			return fmt.Errorf(
				"invalid state %s option. Valid options are: %s, %s, %s",
				o.states[i],
				stateFailing,
				stateFlaky,
				statePassing,
			)
		}
	}

	for _, board := range o.boards {
		if board != boardBlocking &&
			board != boardInforming {
			return fmt.Errorf(
				"invalid board %s option. Valid options are: %s, %s",
				board,
				boardBlocking,
				boardInforming,
			)
		}
	}

	if o.gitHubIssue != -1 {
		token, isSet := os.LookupEnv(github.TokenEnvKey)
		if !isSet || token == "" {
			return fmt.Errorf("cannot send the screenshots if %s environment variable is not set", github.TokenEnvKey)
		}

		gh := github.New()

		issue, _, err := gh.Client().GetIssue(context.Background(), git.DefaultGithubOrg, k8sSigReleaseRepo, o.gitHubIssue)
		if err != nil || issue == nil {
			return fmt.Errorf("getting the GitHub Issue %d: %w", o.gitHubIssue, err)
		}

		// The issue needs to be in open state
		if issue.GetState() != "open" {
			return fmt.Errorf("GitHub Issue %d is %s needs to be a open issue", o.gitHubIssue, issue.GetState())
		}

		// Should be a Issue and not a Pull Request
		if issue.PullRequestLinks != nil {
			return errors.New("this is a Pull Request and not a GitHub Issue")
		}
	}

	return nil
}

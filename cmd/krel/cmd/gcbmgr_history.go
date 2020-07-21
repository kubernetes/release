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
	"fmt"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/gcp/build"
)

type HistoryOptions struct {
	DateFrom       string
	DateTo         string
	DateFromParsed string
	DateToParsed   string
}

var (
	historyOpts = &HistoryOptions{}

	status = map[string]string{
		"SUCCESS":   "Yes",
		"FAILURE":   "No",
		"CANCELLED": "No, Canceled",
		"TIMEOUT":   "No, Timeout",
	}
)

// historyCmd is a krel gcbmgr subcommand which generates information about the command that the operator ran for a specific release cut
var historyCmd = &cobra.Command{
	Use:           "history --branch release-1.18 --date-from 2020-06-18 [--date-to 2020-06-19]",
	Short:         "Run history to build a list of commands that ran when cutting a specific Kubernetes release",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunHistory(historyOpts)
	},
}

func init() {
	historyCmd.PersistentFlags().StringVar(
		&historyOpts.DateFrom,
		"date-from",
		"",
		"Get the jobs starting from a specific date. Format to use yyyy-mm-dd",
	)

	historyCmd.PersistentFlags().StringVar(
		&historyOpts.DateTo,
		"date-to",
		"",
		"Get the jobs ending from a specific date. Format to use yyyy-mm-dd",
	)

	gcbmgrCmd.AddCommand(historyCmd)
}

// RunHistory is the function invoked by 'krel gcbmgr history', responsible for
// getting the jobs and builind the list of commands to be added in the GitHub issue
func RunHistory(opts *HistoryOptions) error {
	if err := opts.Validate(); err != nil {
		return errors.Wrap(err, "validating history options")
	}

	logrus.Infof("Running history with the following options: %+v", opts)
	logrus.Infof("Build options: %v", *buildOpts)

	tagFilter := fmt.Sprintf("tags=%q create_time>%q create_time<%q", gcbmgrOpts.Branch, historyOpts.DateFromParsed, historyOpts.DateToParsed)
	jobs, err := build.GetJobsByTag(buildOpts.Project, tagFilter)
	if err != nil {
		return errors.Wrap(err, "getting the GCP build jobs")
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)

	table.SetHeader([]string{"Step", "Command", "Link", "Start", "Duration", "Succeeded?"})
	for i := len(jobs) - 1; i >= 0; i-- {
		job := jobs[i]
		stage := ""
		for _, stageTag := range job.Tags {
			if stageTag == "RELEASE" {
				stage = stageTag
				break
			}
			if stageTag == "STAGE" {
				stage = stageTag
				break
			}
		}

		// Build the command that was executed
		command := fmt.Sprintf("`krel gcbmgr --%s --type %s --branch %s %s",
			strings.ToLower(stage),
			job.Substitutions["_TYPE"],
			job.Substitutions["_RELEASE_BRANCH"],
			job.Substitutions["_BUILDVERSION"],
		)

		var mock string
		if job.Substitutions["_NOMOCK"] != "" {
			command = fmt.Sprintf("%s %s`", command, job.Substitutions["_NOMOCK"])
			mock = ""
		} else {
			command = fmt.Sprintf("%s`", command)
			mock = "MOCK "
		}

		start := job.Timing["BUILD"].StartTime
		end := job.Timing["BUILD"].EndTime
		logs := job.LogUrl

		// Calculate the duration of the job
		layout := "2006-01-02T15:04:05.000000000Z"
		tStart, err := time.Parse(layout, start)
		if err != nil {
			return errors.Wrap(err, "parsing the start job time")
		}
		tEnd, err := time.Parse(layout, end)
		if err != nil {
			return errors.Wrap(err, "parsing the end job time")
		}
		diff := tEnd.Sub(tStart)
		out := time.Time{}.Add(diff)

		step := fmt.Sprintf("`%s%s`", mock, stage)
		table.Append([]string{step, command, logs, start, out.Format("15:04:05"), status[job.Status]})
	}

	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.Render()

	fmt.Print(tableString.String())

	return nil
}

func (o *HistoryOptions) Validate() error {
	if o.DateFrom == "" {
		return errors.New("need to specify a starting date")
	}

	layOut := "2006-01-02"
	timeStampFrom, err := time.Parse(layOut, o.DateFrom)
	if err != nil {
		return errors.Wrapf(err, "failed to convert the date from flag")
	}
	o.DateFromParsed = timeStampFrom.Format("2006-01-02T15:04:05.000Z")

	if o.DateTo == "" {
		// Set the ending date to midnight of the next day
		dateTo := time.Date(timeStampFrom.Year(), timeStampFrom.Month(), timeStampFrom.Day(), 24, 0, 0, 0, timeStampFrom.Location())
		o.DateToParsed = dateTo.Format("2006-01-02T15:04:05.000Z")
	} else {
		timeStampTo, err := time.Parse(layOut, o.DateFrom)
		if err != nil {
			return errors.Wrapf(err, "failed to convert the date from flag")
		}
		// Set the ending date to midnight of the next day
		dateTo := time.Date(timeStampTo.Year(), timeStampTo.Month(), timeStampTo.Day(), 24, 0, 0, 0, timeStampTo.Location())
		o.DateToParsed = dateTo.Format("2006-01-02T15:04:05.000Z")
	}

	return nil
}

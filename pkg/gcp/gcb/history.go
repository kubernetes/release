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

package gcb

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/cloudbuild/v1"

	"sigs.k8s.io/release-sdk/git"

	"k8s.io/release/pkg/gcp/build"
	"k8s.io/release/pkg/release"
)

// History is the main structure for retrieving the GCB history output.
type History struct {
	opts *HistoryOptions
	impl historyImpl
}

// NewHistory creates a new `History` instance.
func NewHistory(opts *HistoryOptions) *History {
	return &History{
		opts: opts,
		impl: &defaultHistoryImpl{},
	}
}

// SetImpl can be used to set the internal History implementation.
func (h *History) SetImpl(impl historyImpl) {
	h.impl = impl
}

// HistoryOptions are the main settings for the `History`.
type HistoryOptions struct {
	// Branch is the release branch for filtering the jobs.
	Branch string

	// Project is the GCB project to be used.
	Project string

	// DateFrom is the string date for selecting the start of the range.
	DateFrom string

	// DateTo is the string date for selecting the end of the range.
	DateTo string
}

//counterfeiter:generate . historyImpl
type historyImpl interface {
	ParseTime(layout, value string) (time.Time, error)
	GetJobsByTag(project, tagsFilter string) ([]*cloudbuild.Build, error)
}

type defaultHistoryImpl struct{}

func (*defaultHistoryImpl) ParseTime(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

func (*defaultHistoryImpl) GetJobsByTag(
	project, tagsFilter string,
) ([]*cloudbuild.Build, error) {
	return build.GetJobsByTag(project, tagsFilter)
}

// NewHistoryOptions creates a new default HistoryOptions instance.
func NewHistoryOptions() *HistoryOptions {
	return &HistoryOptions{
		Branch:   git.DefaultBranch,
		Project:  release.DefaultKubernetesStagingProject,
		DateFrom: time.Now().Format("2006-01-02"),
		DateTo:   time.Now().Format("2006-01-02"),
	}
}

var status = map[string]string{
	"SUCCESS":   "Yes",
	"FAILURE":   "No",
	"CANCELLED": "No, Canceled",
	"TIMEOUT":   "No, Timeout",
}

// RunHistory is the function invoked by 'krel history', responsible for
// getting the jobs and builind the list of commands to be added in the GitHub issue.
func (h *History) Run() error {
	from, to, err := h.parseDateRange()
	if err != nil {
		return fmt.Errorf("parse from and to dates: %w", err)
	}

	logrus.Infof("Running history with the following options: %+v", h.opts)

	tagFilter := fmt.Sprintf(
		"tags=%q create_time>%q create_time<%q", h.opts.Branch, from, to,
	)

	jobs, err := h.impl.GetJobsByTag(h.opts.Project, tagFilter)
	if err != nil {
		return fmt.Errorf("get GCP build jobs by tag: %w", err)
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewTable(tableString,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			},
		}),
		tablewriter.WithHeader([]string{"Step", "Command", "Link", "Start", "Duration", "Succeeded?"}),
		tablewriter.WithRenderer(renderer.NewMarkdown()),
		tablewriter.WithRendition(tw.Rendition{
			Symbols: tw.NewSymbols(tw.StyleMarkdown),
			Borders: tw.Border{
				Left:   tw.On,
				Top:    tw.Off,
				Right:  tw.On,
				Bottom: tw.Off,
			},
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenRows: tw.On,
				},
			},
		}),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
	)

	for i := len(jobs) - 1; i >= 0; i-- {
		job := jobs[i]
		subcommand := ""

		for _, tag := range job.Tags {
			if tag == "RELEASE" || tag == "STAGE" {
				subcommand = strings.ToLower(tag)

				break
			}
		}

		// Build the command that was executed
		command := fmt.Sprintf("`krel %s --type %s --branch %s --build-version %s",
			subcommand,
			job.Substitutions["_TYPE"],
			job.Substitutions["_RELEASE_BRANCH"],
			job.Substitutions["_BUILDVERSION"],
		)

		var mock string

		if job.Substitutions["_NOMOCK"] != "" {
			command = fmt.Sprintf("%s %s`", command, job.Substitutions["_NOMOCK"])
			mock = ""
		} else {
			command += "`"
			mock = "mock "
		}

		start := job.Timing["BUILD"].StartTime
		end := job.Timing["BUILD"].EndTime
		logs := job.LogUrl

		if start == "" || end == "" {
			logrus.Infof("Skipping unfinished job from %s with ID: %s", job.CreateTime, job.Id)

			continue
		}

		// Calculate the duration of the job
		const layout = "2006-01-02T15:04:05.99Z"

		tStart, err := h.impl.ParseTime(layout, start)
		if err != nil {
			return fmt.Errorf("parsing the start job time: %w", err)
		}

		tEnd, err := h.impl.ParseTime(layout, end)
		if err != nil {
			return fmt.Errorf("parsing the end job time: %w", err)
		}

		diff := tEnd.Sub(tStart)
		out := time.Time{}.Add(diff)

		step := fmt.Sprintf("`%s%s`", mock, subcommand)
		if err := table.Append([]string{
			step, command, logs, start,
			out.Format("15:04:05"), status[job.Status],
		}); err != nil {
			return fmt.Errorf("append row to table: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return err
	}

	fmt.Print(tableString.String())

	return nil
}

func (h *History) parseDateRange() (from, to string, err error) {
	if h.opts.DateFrom == "" {
		return "", "", errors.New("need to specify a start date")
	}

	const (
		parseLayout  = "2006-01-02"
		resultLayout = "2006-01-02T15:04:05.000Z"
	)

	timeStampFrom, err := h.impl.ParseTime(parseLayout, h.opts.DateFrom)
	if err != nil {
		return "", "", fmt.Errorf("convert date from: %w", err)
	}

	from = timeStampFrom.Format(resultLayout)

	if h.opts.DateTo == "" {
		// Set the ending date to midnight of the next day
		dateTo := time.Date(
			timeStampFrom.Year(), timeStampFrom.Month(), timeStampFrom.Day(),
			24, 0, 0, 0, timeStampFrom.Location(),
		)
		to = dateTo.Format(resultLayout)
	} else {
		timeStampTo, err := h.impl.ParseTime(parseLayout, h.opts.DateTo)
		if err != nil {
			return "", "", fmt.Errorf("convert date to: %w", err)
		}
		// Set the ending date to midnight of the next day
		dateTo := time.Date(
			timeStampTo.Year(), timeStampTo.Month(), timeStampTo.Day(),
			24, 0, 0, 0, timeStampTo.Location(),
		)
		to = dateTo.Format(resultLayout)
	}

	return from, to, nil
}

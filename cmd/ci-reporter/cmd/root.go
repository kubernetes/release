/*
Copyright 2021 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/tj/go-spin"
	"golang.org/x/net/context"

	"sigs.k8s.io/release-utils/util"
)

var rootCmd = &cobra.Command{
	Use:   "reporter",
	Short: "Github and Testgrid report generator",
	Long:  "CI-Signal reporter that generates github and testgrid reports.",
	RunE: func(cmd *cobra.Command, args []string) error {
		setGithubConfig(cmd, args)
		// all available reporters are used by default that are used to generate the report
		// CLI sub commands can be used to specify a specific reporter
		selectedReporters := AllImplementedReporters

		return RunReport(cmd.Context(), cfg, &selectedReporters)
	},
}

// Execute executes the ci-reporter root command.
func Execute() error {
	return rootCmd.Execute()
}

// Config configuration that is getting injected into ci-signal report functions.
type Config struct {
	GithubClient   *githubv4.Client
	GithubToken    string
	ReleaseVersion string
	ShortReport    bool
	JSONOutput     bool
	Filepath       string
}

var cfg = &Config{}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfg.ReleaseVersion, "release-version", "v", "", "Specify a Kubernetes release versions like '1.22' which will populate the report additionally")
	rootCmd.PersistentFlags().BoolVarP(&cfg.ShortReport, "short", "s", false, "A short report for mails and slack")
	rootCmd.PersistentFlags().BoolVar(&cfg.JSONOutput, "json", false, "Report output in json format")
	rootCmd.PersistentFlags().StringVarP(&cfg.Filepath, "file", "f", "", "Specify a filepath to write the report to a file")
}

// RunReport used to execute.
func RunReport(ctx context.Context, cfg *Config, reporters *CIReporters) error {
	go func() {
		s := spin.New()

		for {
			fmt.Printf("\rloading data %s ", s.Next())
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// collect data from filtered reporters
	reports, err := reporters.CollectReportDataFromReporters(ctx, cfg)
	if err != nil {
		return err
	}

	// visualize data
	if err := PrintReporterData(cfg, reports); err != nil {
		return fmt.Errorf("printing report data: %w", err)
	}

	return nil
}

//
// Generic reporter types
//

// CIReportDataFields used so specify multiple reports.
type CIReportDataFields []CIReportData

// Marshal used to marshal CIReports into bytes.
func (d *CIReportDataFields) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

// CIReportData format of the ci report data that is being generated.
type CIReportData struct {
	Info    CIReporterInfo    `json:"info"`
	Records []*CIReportRecord `json:"records"`
}

// CIReporterInfo meta information about a reporter implementation.
type CIReporterInfo struct {
	Name CIReporterName `json:"name"`
}

// CIReporterName identifying name of a reporter.
type CIReporterName string

// CIReportRecord generic report data format.
type CIReportRecord struct {
	Title            string `json:"title"`
	TestgridBoard    string `json:"testgrid_board"`
	URL              string `json:"url"`
	Status           string `json:"status"`
	StatusDetails    string `json:"status_details"`
	CreatedTimestamp string `json:"created_timestamp"`
	UpdatedTimestamp string `json:"updated_timestamp"`
}

//
// Generic CIReporter interface and related functions
//

// CIReporter interface that is used to implement a new reporter.
type CIReporter interface {
	// GetCIReporterHead sets meta information which is used to differentiate reporters
	GetCIReporterHead() CIReporterInfo
	// CollectReportData is used to request / collect all report data
	CollectReportData(context.Context, *Config) ([]*CIReportRecord, error)
}

// CIReporters used to specify multiple CIReports, type gets extended by helper functions to collect and visualize report data.
type CIReporters []CIReporter

// AllImplementedReporters list of implemented reports that are used to generate ci-reports.
var AllImplementedReporters = CIReporters{GithubReporter{}, TestgridReporter{}}

// SearchReporter used to filter a implemented reporter by name.
func SearchReporter(ctx context.Context, reporterName string) (CIReporter, error) {
	var reporter CIReporter

	reporterFound := false

	for _, r := range AllImplementedReporters {
		if strings.EqualFold(string(r.GetCIReporterHead().Name), reporterName) {
			reporter = r
			reporterFound = true

			break
		}
	}

	if !reporterFound {
		return nil, errors.New("could not find a implemented reporter")
	}

	return reporter, nil
}

// CollectReportDataFromReporters used to collect data for multiple reporters.
func (r *CIReporters) CollectReportDataFromReporters(ctx context.Context, cfg *Config) (*CIReportDataFields, error) {
	collectedReports := CIReportDataFields{}

	for i := range *r {
		reporters := *r
		reporter := reporters[i]
		reporterHead := reporter.GetCIReporterHead()

		reportData, err := reporter.CollectReportData(ctx, cfg)
		if err != nil {
			return nil, err
		}

		collectedReports = append(collectedReports, CIReportData{
			Info:    reporterHead,
			Records: reportData,
		})
	}

	return &collectedReports, nil
}

// PrintReporterData used to print report data
//  1. Get a output stream to write the data to
//  2. Write data to stream
//     2.1. Write data in JSON format if set so
//     2.2. Write data in table format
func PrintReporterData(cfg *Config, reports *CIReportDataFields) error {
	// Get a stream to write the data to (file stream / standard out stream)
	var out *os.File

	if cfg.Filepath != "" {
		// open output file
		fileOut, err := os.OpenFile(cfg.Filepath, os.O_WRONLY|os.O_CREATE, 0o666)
		if err != nil {
			return fmt.Errorf("could not open or create a file at %s to write the ci signal report to: %w", cfg.Filepath, err)
		}

		out = fileOut
	} else {
		out = os.Stdout
	}

	defer func() {
		if err := out.Close(); err != nil {
			panic(err)
		}
	}()

	// Write data to stream
	if cfg.JSONOutput {
		// print report in json format
		d, err := reports.Marshal()
		if err != nil {
			return fmt.Errorf("could not marshal report data: %w", err)
		}

		_, err = out.Write(d)
		if err != nil {
			return fmt.Errorf("could not write to output stream: %w", err)
		}

		return nil
	}

	// print report in table format, (short table differs)
	for _, r := range *reports {
		// write header
		_, err := fmt.Fprintf(out, "\n%s REPORT\n\n", strings.ToUpper(string(r.Info.Name)))
		if err != nil {
			return fmt.Errorf("could not write to output stream: %w", err)
		}

		data := [][]string{}
		table := util.NewTableWriter(out, tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			},
		}),
			tablewriter.WithHeader([]string{"TESTGRID BOARD", "TITLE", "STATUS", "STATUS DETAILS"}),
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

		// table in short version differs from regular table
		if cfg.ShortReport {
			for _, record := range r.Records {
				data = append(data, []string{record.TestgridBoard, record.Title, record.Status, record.StatusDetails})
			}
		} else {
			table.Options(tablewriter.WithHeader([]string{"TESTGRID BOARD", "TITLE", "STATUS", "STATUS DETAILS", "URL", "UPDATED AT"}))

			for _, record := range r.Records {
				data = append(data, []string{
					record.TestgridBoard,
					record.Title, record.Status,
					record.StatusDetails,
					record.URL,
					strings.ReplaceAll(record.UpdatedTimestamp, "T00:00:00+00:00", ""),
				})
			}
		}

		if err := table.Bulk(data); err != nil {
			return err
		}

		if err := table.Render(); err != nil {
			return err
		}

		// write a summary
		countCategories := map[string]int{}
		categoryIndex := 2

		for i := range data {
			countCategories[data[i][categoryIndex]]++
		}

		categoryCounts := ""
		for category, categoryCount := range countCategories {
			categoryCounts += fmt.Sprintf("%s:%d ", category, categoryCount)
		}

		if _, err := fmt.Fprintf(out, "\nSUMMARY - Total:%d %s\n", len(data), categoryCounts); err != nil {
			return fmt.Errorf("could not write to output stream: %w", err)
		}
	}

	return nil
}

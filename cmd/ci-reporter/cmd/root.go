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

	"github.com/google/go-github/v34/github"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
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
		return RunReport(cfg, &selectedReporters)
	},
}

// Execute executes the ci-reporter root command.
func Execute() error {
	return rootCmd.Execute()
}

// Config configuration that is getting injected into ci-signal report functions
type Config struct {
	GithubClient   *github.Client
	GithubToken    string
	ReleaseVersion string
	ShortReport    bool
	JSONOutput     bool
}

var cfg = &Config{
	GithubClient:   &github.Client{},
	GithubToken:    "",
	ReleaseVersion: "",
	ShortReport:    false,
	JSONOutput:     false,
}

func init() {
	rootCmd.Flags().StringVarP(&cfg.ReleaseVersion, "release-version", "v", "", "Specify a Kubernetes release versions like '1.22' which will populate the report additionally")
	rootCmd.PersistentFlags().BoolVarP(&cfg.ShortReport, "short", "s", false, "A short report for mails and slack")
	rootCmd.PersistentFlags().BoolVar(&cfg.JSONOutput, "json", false, "Report output in json format")
}

// RunReport used to execute
func RunReport(cfg *Config, reporters *CIReporters) error {
	// collect data from filtered reporters
	reports, err := reporters.CollectReportDataFromReporters(cfg)
	if err != nil {
		return err
	}

	// visualize data
	err = PrintReporterData(cfg, reports)
	if err != nil {
		return err
	}

	return nil
}

//
// Generic reporter types
//

// CIReportDataFields used so specify multiple reports
type CIReportDataFields []CIReportData

// Marshal used to marshal CIReports into bytes
func (d *CIReportDataFields) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

// CIReportData format of the ci report data that is being generated
type CIReportData struct {
	Info    CIReporterInfo    `json:"info"`
	Records []*CIReportRecord `json:"records"`
}

// CIReporterInfo meta information about a reporter implementation
type CIReporterInfo struct {
	Name CIReporterName `json:"name"`
}

// CIReporterName identifying name of a reporter
type CIReporterName string

// CIReportRecord generic report data format
type CIReportRecord struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	URL              string   `json:"url"`
	Category         string   `json:"category"`
	Sigs             []string `json:"sigs"`
	Status           string   `json:"status"`
	CreatedTimestamp string   `json:"created_timestamp"`
}

//
// Generic CIReporter interface and related functions
//

// CIReporter interface that is used to implement a new reporter
type CIReporter interface {
	// GetCIReporterHead sets meta information which is used to differentiate reporters
	GetCIReporterHead() CIReporterInfo
	// CollectReportData is used to request / collect all report data
	CollectReportData(*Config) ([]*CIReportRecord, error)
}

// CIReporters used to specify multiple CIReports, type gets extended by helper functions to collect and visualize report data
type CIReporters []CIReporter

// AllImplementedReporters list of implemented reports that are used to generate ci-reports
var AllImplementedReporters = CIReporters{GithubReporter{}, TestgridReporter{}}

// SearchReporter used to filter a implemented reporter by name
func SearchReporter(reporterName string) (CIReporter, error) {
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

// CollectReportDataFromReporters used to collect data for multiple reporters
func (r *CIReporters) CollectReportDataFromReporters(cfg *Config) (*CIReportDataFields, error) {
	collectedReports := CIReportDataFields{}
	for i := range *r {
		reporters := *r
		reporter := reporters[i]
		reporterHead := reporter.GetCIReporterHead()
		reportData, err := reporter.CollectReportData(cfg)
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
func PrintReporterData(cfg *Config, reports *CIReportDataFields) error {
	if cfg.JSONOutput {
		// print report in json format
		d, err := reports.Marshal()
		if err != nil {
			return nil
		}
		fmt.Print(string(d))
	} else {
		// print report in table format, (short table differs)
		for _, r := range *reports {
			fmt.Printf("\n%s REPORT\n", strings.ToUpper(string(r.Info.Name)))
			table := tablewriter.NewWriter(os.Stdout)
			data := [][]string{}

			// table in short version differs from regular table
			if cfg.ShortReport {
				table.SetHeader([]string{"ID", "TITLE", "CATEGORY", "STATUS"})
				for _, record := range r.Records {
					data = append(data, []string{record.ID, record.Title, record.Category, record.Status})
				}
			} else {
				table.SetHeader([]string{"ID", "TITLE", "CATEGORY", "STATUS", "SIGS", "URL", "TS"})
				for _, record := range r.Records {
					data = append(data, []string{record.ID, record.Title, record.Category, record.Status, fmt.Sprintf("%v", record.Sigs), record.URL, record.CreatedTimestamp})
				}
			}

			countCategories := map[string]int{}
			categoryIndex := 2
			for i := range data {
				countCategories[data[i][categoryIndex]]++
			}
			categoryCounts := ""
			for category, categoryCount := range countCategories {
				categoryCounts += fmt.Sprintf("%s:%d\n", category, categoryCount)
			}
			if cfg.ShortReport {
				table.SetFooter([]string{fmt.Sprintf("Total: %d", len(data)), "", categoryCounts, ""})
			} else {
				table.SetFooter([]string{fmt.Sprintf("Total: %d", len(data)), "", categoryCounts, "", "", "", ""})
			}
			table.SetBorder(false)
			table.AppendBulk(data)
			table.SetAutoMergeCells(true)
			table.Render()
		}
	}
	return nil
}

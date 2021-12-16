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
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/release/pkg/testgrid"
)

var testgridCmd = &cobra.Command{
	Use:    "testgrid",
	Short:  "Testgrid report generator",
	Long:   "CI-Signal reporter that generates only a testgrid report.",
	PreRun: setGithubConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunReport(cfg, &CIReporters{TestgridReporter{}})
	},
}

// TestgridReporterName used to identify github reporter
var TestgridReporterName CIReporterName = "testgrid"

func init() {
	testgridCmd.Flags().StringVarP(&cfg.ReleaseVersion, "release-version", "v", "", "Specify a Kubernetes release versions like '1.22' which will populate the report additionally")
	rootCmd.AddCommand(testgridCmd)
}

// TestgridReporter github CIReporter implementation
type TestgridReporter struct{}

// GetCIReporterHead implementation from CIReporter
func (r TestgridReporter) GetCIReporterHead() CIReporterInfo {
	return CIReporterInfo{Name: TestgridReporterName}
}

// CollectReportData implementation from CIReporter
func (r TestgridReporter) CollectReportData(cfg *Config) ([]*CIReportRecord, error) {
	testgridReportData, err := GetTestgridReportData(*cfg)
	if err != nil {
		return nil, err
	}
	records := []*CIReportRecord{}
	for dashboardName, jobData := range testgridReportData {
		for jobName := range jobData {
			jobSummary := jobData[jobName]
			if !cfg.ShortReport || jobSummary.OverallStatus != testgrid.Passing {
				records = append(records, &CIReportRecord{
					ID:               string(jobName),
					Title:            string(dashboardName),
					URL:              jobSummary.GetJobURL(jobName),
					Category:         string(jobSummary.OverallStatus),
					Sigs:             jobSummary.FilterSigs(),
					Status:           jobSummary.FilterSuccessRateForLastRuns(),
					CreatedTimestamp: time.Unix(jobSummary.LastRunTimestamp, 0).Format("2006-01-02 15:04:05 CET"),
				})
			}
		}
	}
	return records, nil
}

// GetTestgridReportData used to request the raw report data from testgrid
func GetTestgridReportData(cfg Config) (testgrid.DashboardData, error) {
	testgridURLs := []testgrid.DashboardName{"sig-release-master-blocking", "sig-release-master-informing"}
	if cfg.ReleaseVersion != "" {
		testgridURLs = append(testgridURLs, []testgrid.DashboardName{
			testgrid.DashboardName(fmt.Sprintf("sig-release-%s-blocking", cfg.ReleaseVersion)),
			testgrid.DashboardName(fmt.Sprintf("sig-release-%s-informing", cfg.ReleaseVersion)),
		}...)
	}
	return testgrid.ReqTestgridDashboardSummaries(testgridURLs)
}

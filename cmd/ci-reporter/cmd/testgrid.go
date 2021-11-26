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
	"reflect"
	"regexp"

	"github.com/pkg/errors"
	"k8s.io/release/pkg/testgrid"
)

// GetTestgridReportData used to request the raw report data from testgrid
func GetTestgridReportData(cfg ReporterConfig) (testgrid.DashboardData, error) {
	testgridURLs := []testgrid.DashboardName{"sig-release-master-blocking", "sig-release-master-informing"}

	if cfg.ReleaseVersion != "" {
		testgridURLs = append(testgridURLs, []testgrid.DashboardName{
			testgrid.DashboardName(fmt.Sprintf("sig-release-%s-blocking", cfg.ReleaseVersion)),
			testgrid.DashboardName(fmt.Sprintf("sig-release-%s-informing", cfg.ReleaseVersion)),
		}...)
	}

	return testgrid.ReqTestgridDashboardSummaries(testgridURLs)
}

// PrintTestgridReportData used to print testgrid report data out to the console
func PrintTestgridReportData(cfg ReporterConfig, stats *testgrid.DashboardData) error {
	printShortReport := func(name testgrid.DashboardName, jobData *testgrid.JobData) error {
		overview, err := jobData.Overview()
		if err != nil {
			return errors.Wrap(err, "could not get testgrid data overview")
		}
		fmt.Printf("\nOverview for %s\n", name)
		fmt.Printf("\t%d jobs total\n", len(overview.FailingJobs)+len(overview.FlakyJobs)+len(overview.PassingJobs)+len(overview.StaleJobs))
		fmt.Printf("\t%d are passing\n", len(overview.PassingJobs))
		fmt.Printf("\t%d are flaking\n", len(overview.FlakyJobs))
		fmt.Printf("\t%d are failing\n", len(overview.FailingJobs))
		if len(overview.StaleJobs) > 0 {
			fmt.Printf("\t%d are stale\n", len(overview.StaleJobs))
		}
		return nil
	}

	printLongReport := func(name testgrid.DashboardName, jobData *testgrid.JobData) {
		fmt.Printf("\nDetails for %s\n", name)
		for jobName := range *jobData {
			j := *jobData
			jobData := j[jobName]
			if jobData.OverallStatus != testgrid.Passing {
				fmt.Printf("%s: %s\n", jobData.OverallStatus, jobName)
				fmt.Printf("- %s\n", jobData.GetJobURL(jobName))
				fmt.Printf("- %s\n", jobData.Status)
			} else if jobData.OverallStatus == testgrid.Failing {
				// Filter sigs
				sigRegex := regexp.MustCompile(`sig-[a-zA-Z]+`)
				sigsInvolved := map[string]int{}
				for i := range jobData.Tests {
					sigs := sigRegex.FindAllString(jobData.Tests[i].TestName, -1)
					for _, sig := range sigs {
						sigsInvolved[sig]++
					}
				}
				sigs := reflect.ValueOf(sigsInvolved).MapKeys()
				fmt.Printf("- Currently %d test are failing\n", len(jobData.Tests))
				fmt.Printf("- Sig's involved %v\n", sigs)
			}
		}
	}

	// if the short flag ist set, only print the short report otherwise print both the short & long report
	for name := range *stats {
		s := *stats
		data := s[name]
		err := printShortReport(name, &data)
		if err != nil {
			return err
		}
		if !cfg.ShortReport {
			printLongReport(name, &data)
		}
	}
	return nil
}

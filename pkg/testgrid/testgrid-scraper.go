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

package testgrid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// SummaryLookup this type is used if multiple testgrid summaries are getting requested concurrently
type SummaryLookup struct {
	Dashboard DashboardName
	Error     error
	Summary   JobData
}

// ReqTestgridDashboardSummaries this function requests multiple testgrid summaries concurrently
// This function implements a concurrency pattern to send http requests concurrently
func ReqTestgridDashboardSummaries(dashboardNames []DashboardName) (DashboardData, error) {
	// Worker
	requestData := func(done <-chan interface{}, dashboardNames ...DashboardName) <-chan SummaryLookup {
		summaryLookups := make(chan SummaryLookup)
		go func() {
			defer close(summaryLookups)
			for _, dashboardName := range dashboardNames {
				summary, err := ReqTestgridDashboardSummary(dashboardName)
				select {
				case <-done:
					return
				case summaryLookups <- SummaryLookup{
					Dashboard: dashboardName,
					Error:     err,
					Summary:   summary,
				}:
				}
			}
		}()
		return summaryLookups
	}

	done := make(chan interface{})
	defer close(done)
	dashboardData := DashboardData{}
	var err error

	// Collect data from buffered channel
	for lookups := range requestData(done, dashboardNames...) {
		if lookups.Error != nil {
			err = multierror.Append(err, errors.Wrapf(lookups.Error, "error requesting summary for dashboard %s", lookups.Dashboard))
		} else {
			dashboardData[lookups.Dashboard] = lookups.Summary
		}
	}
	return dashboardData, err
}

// ReqTestgridDashboardSummary used to retrieve summary information about a testgrid dashboard
func ReqTestgridDashboardSummary(dashboardName DashboardName) (JobData, error) {
	resp, err := http.Get(fmt.Sprintf("https://testgrid.k8s.io/%s/summary", dashboardName))
	if err != nil {
		return nil, errors.Wrap(err, "request remote content")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}
	summary, err := UnmarshalTestgridSummary(body)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal response body")
	}
	return summary, nil
}

// Overview used to get an overview about a testgrid board without additional information
type Overview struct {
	PassingJobs []JobName
	FlakyJobs   []JobName
	FailingJobs []JobName
	StaleJobs   []JobName
}

// Overview used to get an overview about a testgrid dashboard
func (d *JobData) Overview() (Overview, error) {
	overview := Overview{}
	for job := range *d {
		data := *d
		switch data[job].OverallStatus {
		case Passing:
			overview.PassingJobs = append(overview.PassingJobs, job)
		case Flaky:
			overview.FlakyJobs = append(overview.FlakyJobs, job)
		case Failing:
			overview.FailingJobs = append(overview.FailingJobs, job)
		case Stale:
			overview.StaleJobs = append(overview.StaleJobs, job)
		default:
			return Overview{}, fmt.Errorf("unrecognized job status: %s with summary info %v", data[job].OverallStatus, data[job])
		}
	}
	return overview, nil
}

//
// Types that reflect from testgrid summary
//

// JobName type for testgrid jobs
type JobName string

// DashboardData used to store testgrid dashboards
type DashboardData map[DashboardName]JobData

// JobData used to store multiple TestgridJobs TestgridJobSummary
type JobData map[JobName]JobSummary

// UnmarshalTestgridSummary used to unmarshal bytes into TestgridSummary
func UnmarshalTestgridSummary(data []byte) (JobData, error) {
	var r JobData
	err := json.Unmarshal(data, &r)
	return r, err
}

// Marshal used to marshal TestgridSummary into bytes
func (d *JobData) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

// JobSummary contains information about a testgrid job
type JobSummary struct {
	Alert               string        `json:"alert"`
	LastRunTimestamp    int64         `json:"last_run_timestamp"`
	LastUpdateTimestamp int64         `json:"last_update_timestamp"`
	LatestGreen         string        `json:"latest_green"`
	OverallStatus       OverallStatus `json:"overall_status"`
	OverallStatusIcon   string        `json:"overall_status_icon"`
	Status              string        `json:"status"`
	Tests               []Test        `json:"tests"`
	DashboardName       DashboardName `json:"dashboard_name"`
	BugURL              string        `json:"bug_url"`
}

// GetJobURL used to get the testgrid job url
func (j *JobSummary) GetJobURL(jobName JobName) string {
	return fmt.Sprintf("https://testgrid.k8s.io/%s#%s", j.DashboardName, strings.ReplaceAll(string(jobName), " ", "%20"))
}

// Test contains information about tests if the status if the Job is failing
type Test struct {
	DisplayName    string        `json:"display_name"`
	TestName       string        `json:"test_name"`
	FailCount      int64         `json:"fail_count"`
	FailTimestamp  int64         `json:"fail_timestamp"`
	PassTimestamp  int64         `json:"pass_timestamp"`
	BuildLink      string        `json:"build_link"`
	BuildURLText   string        `json:"build_url_text"`
	BuildLinkText  string        `json:"build_link_text"`
	FailureMessage string        `json:"failure_message"`
	LinkedBugs     []interface{} `json:"linked_bugs"`
	FailTestLink   string        `json:"fail_test_link"`
}

// DashboardName type for the testgrid dashboard (like sig-release-master-blocking)
type DashboardName string

const (
	// SigReleaseMasterInforming one of the dashboard names that can be used to scrapo a testgrid summary
	SigReleaseMasterInforming DashboardName = "sig-release-master-informing"
	// SigReleaseMasterBlocking one of the dashboard names that can be used to scrapo a testgrid summary
	SigReleaseMasterBlocking DashboardName = "sig-release-master-blocking"
)

// FilterSigs used to filter sigs from failing tests
func (j *JobSummary) FilterSigs() []string {
	// Filter sigs
	sigRegex := regexp.MustCompile(`sig-[a-zA-Z]+`)
	sigsInvolved := map[string]int{}
	for i := range j.Tests {
		sigs := sigRegex.FindAllString(j.Tests[i].TestName, -1)
		for _, sig := range sigs {
			sigsInvolved[sig]++
		}
	}
	sigs := []string{}
	for k := range sigsInvolved {
		sigs = append(sigs, k)
	}
	return sigs
}

// FilterSuccessRateForLastRuns used to parse last runs from test
// example: 8 of 9 (88.9%) recent columns passed (296791 of 296793 or 100.0% cells) -> 8 of 9 (88.9%)
func (j *JobSummary) FilterSuccessRateForLastRuns() string {
	successRateForLastRunsRegex := regexp.MustCompile(`\d+\sof\s\d+\s\(\d+\.\d+%\)`)
	return successRateForLastRunsRegex.FindString(j.Status)
}

// OverallStatus of a job in testgrid
type OverallStatus string

const (
	// Failing job status
	Failing OverallStatus = "FAILING"
	// Flaky job status
	Flaky OverallStatus = "FLAKY"
	// Passing job status
	Passing OverallStatus = "PASSING"
	// Stale job status
	Stale OverallStatus = "STALE"
)

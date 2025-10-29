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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// SummaryLookup this type is used if multiple testgrid summaries are getting requested concurrently.
type SummaryLookup struct {
	Dashboard DashboardName
	Error     error
	Summary   JobData
}

// ReqTestgridDashboardSummaries this function requests multiple testgrid summaries concurrently
// This function implements a concurrency pattern to send http requests concurrently.
func ReqTestgridDashboardSummaries(ctx context.Context, dashboardNames []DashboardName) (DashboardData, error) {
	// Worker
	requestData := func(done <-chan any, dashboardNames ...DashboardName) <-chan SummaryLookup {
		summaryLookups := make(chan SummaryLookup)

		go func() {
			defer close(summaryLookups)

			for _, dashboardName := range dashboardNames {
				summary, err := ReqTestgridDashboardSummary(ctx, dashboardName)
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

	done := make(chan any)
	defer close(done)

	dashboardData := DashboardData{}

	var err error

	// Collect data from buffered channel
	for lookups := range requestData(done, dashboardNames...) {
		if lookups.Error != nil {
			err = errors.Join(err, fmt.Errorf("error requesting summary for dashboard %s: %w", lookups.Dashboard, lookups.Error))
		} else {
			dashboardData[lookups.Dashboard] = lookups.Summary
		}
	}

	return dashboardData, err
}

type NotFound error

var ErrDashboardNotFound NotFound = errors.New("testgrid dashboard not found")

// ReqTestgridDashboardSummary used to retrieve summary information about a testgrid dashboard.
func ReqTestgridDashboardSummary(ctx context.Context, dashboardName DashboardName) (JobData, error) {
	url := fmt.Sprintf("https://testgrid.k8s.io/%s/summary", dashboardName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request remote content: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if strings.Contains(string(body), fmt.Sprintf("Dashboard %s not found", dashboardName)) {
		return nil, ErrDashboardNotFound
	}

	summary, err := UnmarshalTestgridSummary(body)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body: %w", err)
	}

	return summary, nil
}

// Overview used to get an overview about a testgrid board without additional information.
type Overview struct {
	PassingJobs []JobName
	FlakyJobs   []JobName
	FailingJobs []JobName
	StaleJobs   []JobName
}

//
// Types that reflect from testgrid summary
//

// JobName type for testgrid jobs.
type JobName string

// DashboardData used to store testgrid dashboards.
type DashboardData map[DashboardName]JobData

// JobData used to store multiple TestgridJobs TestgridJobSummary.
type JobData map[JobName]JobSummary

// Overview used to get an overview about a testgrid dashboard.
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

// UnmarshalTestgridSummary used to unmarshal bytes into TestgridSummary.
func UnmarshalTestgridSummary(data []byte) (JobData, error) {
	var r JobData

	err := json.Unmarshal(data, &r)

	return r, err
}

// Marshal used to marshal TestgridSummary into bytes.
func (d *JobData) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

// JobSummary contains information about a testgrid job.
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

// GetJobURL used to get the testgrid job url.
func (j *JobSummary) GetJobURL(jobName JobName) string {
	return fmt.Sprintf("https://testgrid.k8s.io/%s#%s", j.DashboardName, strings.ReplaceAll(string(jobName), " ", "%20"))
}

// Test contains information about tests if the status if the Job is failing.
type Test struct {
	DisplayName    string `json:"display_name"`
	TestName       string `json:"test_name"`
	FailCount      int64  `json:"fail_count"`
	FailTimestamp  int64  `json:"fail_timestamp"`
	PassTimestamp  int64  `json:"pass_timestamp"`
	BuildLink      string `json:"build_link"`
	BuildURLText   string `json:"build_url_text"`
	BuildLinkText  string `json:"build_link_text"`
	FailureMessage string `json:"failure_message"`
	LinkedBugs     []any  `json:"linked_bugs"`
	FailTestLink   string `json:"fail_test_link"`
}

// DashboardName type for the testgrid dashboard (like sig-release-master-blocking).
type DashboardName string

const (
	// SigReleaseMasterInforming one of the dashboard names that can be used to scrapo a testgrid summary.
	SigReleaseMasterInforming DashboardName = "sig-release-master-informing"
	// SigReleaseMasterBlocking one of the dashboard names that can be used to scrapo a testgrid summary.
	SigReleaseMasterBlocking DashboardName = "sig-release-master-blocking"
	sigRegexStr              string        = "sig-[a-zA-Z]+"
)

var sigRegex = regexp.MustCompile(sigRegexStr)

// FilterSigs used to filter sigs from failing tests.
func (j *JobSummary) FilterSigs() []string {
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

// OverallStatus of a job in testgrid.
type OverallStatus string

const (
	// Failing job status.
	Failing OverallStatus = "FAILING"
	// Flaky job status.
	Flaky OverallStatus = "FLAKY"
	// Passing job status.
	Passing OverallStatus = "PASSING"
	// Stale job status.
	Stale OverallStatus = "STALE"
)

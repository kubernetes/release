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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	negDashboardNames = []DashboardName{DashboardName(""), DashboardName("sig-release-master"), DashboardName("no-dashboard-here")}
	posDashboardNames = []DashboardName{SigReleaseMasterBlocking, SigReleaseMasterInforming}
)

func TestRequestTestgridSummaryPos(t *testing.T) {
	// Given
	// positive dashboard names

	for _, dashboardName := range posDashboardNames {
		// When
		summary, err := ReqTestgridDashboardSummary(dashboardName)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		for _, jobs := range summary {
			assert.Equal(t, dashboardName, jobs.DashboardName)
		}
	}
}

func TestRequestTestgridSummaryNeg(t *testing.T) {
	// Given
	// negative dashboard names

	for _, dashboardName := range negDashboardNames {
		// When
		summary, err := ReqTestgridDashboardSummary(dashboardName)

		// Then
		assert.Error(t, err)
		assert.Nil(t, summary)
	}
}

func TestRequestTestgridSummariesPos(t *testing.T) {
	// Given
	// positive dashboard names

	// When
	data, err := ReqTestgridDashboardSummaries(posDashboardNames)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, len(posDashboardNames), len(data))
}

func TestRequestTestgridSummariesNeg(t *testing.T) {
	// Given
	// negative dashboard names

	// When
	data, err := ReqTestgridDashboardSummaries(negDashboardNames)

	// Then
	assert.Error(t, err)
	assert.Empty(t, data)
}

func TestRequestTestgridSummariesPosNeg(t *testing.T) {
	// Given
	// Request positive and negative dashboard names, expect to get an error and receive positive dashboard name summaries

	// When
	data, err := ReqTestgridDashboardSummaries(append(negDashboardNames, posDashboardNames...))

	// Then
	assert.Error(t, err, "an error should be returned as not all dashboard name references are correct")
	assert.NotEmpty(t, data, "response shouldn't be empty because valid data das been added alongside faulty data - the correct data should be getting processed nonetheless")
	assert.Equal(t, len(posDashboardNames), len(data))
}

type jobGeneratorDef struct {
	amountOfJobs  int
	overallStatus OverallStatus
}

func TestOverviewPos(t *testing.T) {
	// Given
	passingJobs := 4
	flakyJobs := 3
	failingJobs := 2
	staleJobs := 1
	jobGeneratorDef := []jobGeneratorDef{
		{amountOfJobs: passingJobs, overallStatus: Passing},
		{amountOfJobs: flakyJobs, overallStatus: Flaky},
		{amountOfJobs: failingJobs, overallStatus: Failing},
		{amountOfJobs: staleJobs, overallStatus: Stale},
	}
	data := JobData{}
	for _, jobDef := range jobGeneratorDef {
		for i := 0; i < jobDef.amountOfJobs; i++ {
			data[JobName(fmt.Sprintf("%s-%d", jobDef.overallStatus, i))] = JobSummary{
				OverallStatus: jobDef.overallStatus,
				DashboardName: "sample-dashboard",
			}
		}
	}

	// When
	o, err := data.Overview()

	// Then
	assert.NoError(t, err)
	assert.Equal(t, failingJobs, len(o.FailingJobs))
	assert.Equal(t, flakyJobs, len(o.FlakyJobs))
	assert.Equal(t, passingJobs, len(o.PassingJobs))
	assert.Equal(t, staleJobs, len(o.StaleJobs))
}

func TestOverviewNeg(t *testing.T) {
	// Given
	data := []JobData{{"sampleTest": {OverallStatus: OverallStatus("")}}}

	for _, s := range data {
		// When
		o, err := s.Overview()

		// Then
		assert.Error(t, err)
		assert.Empty(t, o)
	}
}

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
	"fmt"
	"io/ioutil"
	"net/http"
)

// RequiredTestgridJob specifies a testgrid job like 'Master-Blocking'
type RequiredTestgridJob struct {
	OutputName string
	URLName    string
}

// GetTestgridReportData used to request the raw report data from testgrid
func GetTestgridReportData(cfg ReporterConfig) ([]TestgridStatistics, error) {
	requiredJobs := []RequiredTestgridJob{
		{OutputName: "Master-Blocking", URLName: "sig-release-master-blocking"},
		{OutputName: "Master-Informing", URLName: "sig-release-master-informing"},
	}

	if cfg.ReleaseVersion != "" {
		requiredJobs = append(requiredJobs, []RequiredTestgridJob{
			{OutputName: cfg.ReleaseVersion + "-blocking", URLName: "sig-release-" + cfg.ReleaseVersion + "-blocking"},
			{OutputName: cfg.ReleaseVersion + "-informing", URLName: "sig-release-" + cfg.ReleaseVersion + "-informing"},
		}...)
	}
	result := make([]TestgridStatistics, 0)
	for _, job := range requiredJobs {
		resp, err := http.Get(fmt.Sprintf("https://testgrid.k8s.io/%s/summary", job.URLName))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		jobs := make(map[string]overview)
		err = json.Unmarshal(body, &jobs)
		if err != nil {
			return nil, err
		}

		statistics := getStatistics(jobs)
		statistics.Name = job.OutputName
		result = append(result, statistics)
	}
	return result, nil
}

// PrintTestgridReportData used to print testgrid report data out to the console
func PrintTestgridReportData(stats []TestgridStatistics) {
	for _, stat := range stats {
		fmt.Printf("Failures in %s\n", stat.Name)
		fmt.Printf("\t%d jobs total\n", stat.Total)
		fmt.Printf("\t%d are passing\n", stat.Passing)
		fmt.Printf("\t%d are flaking\n", stat.Flaking)
		fmt.Printf("\t%d are failing\n", stat.Failing)
		fmt.Printf("\t%d are stale\n", stat.Stale)
		fmt.Print("\n\n")
	}
}

func getStatistics(jobs map[string]overview) TestgridStatistics {
	result := TestgridStatistics{}
	for _, v := range jobs {
		if v.OverallStatus == "PASSING" {
			result.Passing++
		} else if v.OverallStatus == "FAILING" {
			result.Failing++
		} else if v.OverallStatus == "FLAKY" {
			result.Flaking++
		} else {
			result.Stale++
		}
		result.Total++
	}
	return result
}

// TestgridStatistics specifies a testgrid job overview
type TestgridStatistics struct {
	Name    string
	Total   int
	Passing int
	Flaking int
	Failing int
	Stale   int
}

type overview struct {
	OverallStatus string `json:"overall_status"`
}

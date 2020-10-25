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

package main

import (
	"flag"
	"fmt"

	"k8s.io/klog"
	adapter "k8s.io/release/pkg/vulndash/adapter"
)

// nolint[gocyclo]
func main() {
	// klog uses the "v" flag in order to set the verbosity level
	klog.InitFlags(nil)
	dashboardPath := flag.String(
		"dashboard-file-path",
		"",
		"The path to the local dashboard files")
	vulnProject := flag.String(
		"vuln-target-project",
		"",
		"The project which the vulnerability dashboard will display information for")
	dashboardBucket := flag.String(
		"dashboard-bucket",
		"",
		"GCS bucket where dashboard files are stored")
	flag.Parse()

	if *dashboardPath == "" || *vulnProject == "" || *dashboardBucket == "" {
		klog.Fatal(fmt.Errorf("all of -dashboard-file-path, -vuln-target-project, -dashboard-bucket are required flags"))
	}

	klog.Info("********** STARTING (UPDATING VULN DASHBOARD) **********")
	adapter.UpdateVulnerabilityDashboard(*dashboardPath, *vulnProject, *dashboardBucket)
	klog.Info("********** FINISHED (UPDATING VULN DASHBOARD) **********")
}

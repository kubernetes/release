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

package adapter_test

import (
	"fmt"
	"reflect"
	"testing"

	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"
	adapter "k8s.io/release/pkg/vulndash/adapter"
)

func checkEqual(got, expected interface{}) error {
	if !reflect.DeepEqual(got, expected) {
		return fmt.Errorf(
			`<<<<<<< got (type %T)
%v
=======
%v
>>>>>>> expected (type %T)`,
			got,
			got,
			expected,
			expected)
	}
	return nil
}

func checkError(t *testing.T, err error, msg string) {
	if err != nil {
		fmt.Printf("\n%v", msg)
		fmt.Println(err)
		fmt.Println()
		t.Fail()
	}
}

func TestGenerateVulnerabilityBreakdown(t *testing.T) {
	tests := []struct {
		name            string
		vulnerabilities []*grafeaspb.Occurrence
		expected        map[string]adapter.ImageVulnBreakdown
	}{
		{
			"No critical and no fixable vulnerabilities",
			[]*grafeaspb.Occurrence{
				{
					ResourceUri: "foo/bar1@sha256:111",
					NoteName:    "vuln/CVE-2000",
					Details: &grafeaspb.Occurrence_Vulnerability{
						Vulnerability: &grafeaspb.VulnerabilityOccurrence{
							Severity:     grafeaspb.Severity_HIGH,
							FixAvailable: false,
						},
					},
				},
				{
					ResourceUri: "foo/bar2@sha256:000",
					NoteName:    "vuln/CVE-2000",
					Details: &grafeaspb.Occurrence_Vulnerability{
						Vulnerability: &grafeaspb.VulnerabilityOccurrence{
							Severity:     grafeaspb.Severity_HIGH,
							FixAvailable: false,
						},
					},
				},
			},
			map[string]adapter.ImageVulnBreakdown{
				"foo/bar1@sha256:111": {
					"foo/bar1@sha256:111",
					"bar1",
					"sha256:111",
					1,
					[]string{},
					[]string{},
				},
				"foo/bar2@sha256:000": {
					"foo/bar2@sha256:000",
					"bar2",
					"sha256:000",
					1,
					[]string{},
					[]string{},
				},
			},
		},
		{
			"No critical and Multiple fixable vulnerabilities",
			[]*grafeaspb.Occurrence{
				{
					ResourceUri: "foo/bar1@sha256:111",
					NoteName:    "vuln/CVE-2000",
					Details: &grafeaspb.Occurrence_Vulnerability{
						Vulnerability: &grafeaspb.VulnerabilityOccurrence{
							Severity:     grafeaspb.Severity_HIGH,
							FixAvailable: true,
						},
					},
				},
				{
					ResourceUri: "foo/bar2@sha256:000",
					NoteName:    "vuln/CVE-2000",
					Details: &grafeaspb.Occurrence_Vulnerability{
						Vulnerability: &grafeaspb.VulnerabilityOccurrence{
							Severity:     grafeaspb.Severity_MEDIUM,
							FixAvailable: true,
						},
					},
				},
			},
			map[string]adapter.ImageVulnBreakdown{
				"foo/bar1@sha256:111": {
					"foo/bar1@sha256:111",
					"bar1",
					"sha256:111",
					1,
					[]string{},
					[]string{
						"CVE-2000",
					},
				},
				"foo/bar2@sha256:000": {
					"foo/bar2@sha256:000",
					"bar2",
					"sha256:000",
					1,
					[]string{},
					[]string{
						"CVE-2000",
					},
				},
			},
		},
		{
			"Multiple critical and no fixable vulnerabilities",
			[]*grafeaspb.Occurrence{
				{
					ResourceUri: "foo/bar1@sha256:111",
					NoteName:    "vuln/CVE-2000",
					Details: &grafeaspb.Occurrence_Vulnerability{
						Vulnerability: &grafeaspb.VulnerabilityOccurrence{
							Severity:     grafeaspb.Severity_CRITICAL,
							FixAvailable: false,
						},
					},
				},
				{
					ResourceUri: "foo/bar2@sha256:000",
					NoteName:    "vuln/CVE-2000",
					Details: &grafeaspb.Occurrence_Vulnerability{
						Vulnerability: &grafeaspb.VulnerabilityOccurrence{
							Severity:     grafeaspb.Severity_CRITICAL,
							FixAvailable: false,
						},
					},
				},
			},
			map[string]adapter.ImageVulnBreakdown{
				"foo/bar1@sha256:111": {
					"foo/bar1@sha256:111",
					"bar1",
					"sha256:111",
					1,
					[]string{
						"CVE-2000",
					},
					[]string{},
				},
				"foo/bar2@sha256:000": {
					"foo/bar2@sha256:000",
					"bar2",
					"sha256:000",
					1,
					[]string{
						"CVE-2000",
					},
					[]string{},
				},
			},
		},
		{
			"Multiple critical and fixable vulnerabilities",
			[]*grafeaspb.Occurrence{
				{
					ResourceUri: "foo/bar1@sha256:111",
					NoteName:    "vuln/CVE-2000",
					Details: &grafeaspb.Occurrence_Vulnerability{
						Vulnerability: &grafeaspb.VulnerabilityOccurrence{
							Severity:     grafeaspb.Severity_CRITICAL,
							FixAvailable: true,
						},
					},
				},
				{
					ResourceUri: "foo/bar2@sha256:000",
					NoteName:    "vuln/CVE-2001",
					Details: &grafeaspb.Occurrence_Vulnerability{
						Vulnerability: &grafeaspb.VulnerabilityOccurrence{
							Severity:     grafeaspb.Severity_CRITICAL,
							FixAvailable: true,
						},
					},
				},
			},
			map[string]adapter.ImageVulnBreakdown{
				"foo/bar1@sha256:111": {
					"foo/bar1@sha256:111",
					"bar1",
					"sha256:111",
					1,
					[]string{
						"CVE-2000",
					},
					[]string{
						"CVE-2000",
					},
				},
				"foo/bar2@sha256:000": {
					"foo/bar2@sha256:000",
					"bar2",
					"sha256:000",
					1,
					[]string{
						"CVE-2001",
					},
					[]string{
						"CVE-2001",
					},
				},
			},
		},
	}

	for _, test := range tests {
		testBreakdown := adapter.GenerateVulnerabilityBreakdown(test.vulnerabilities)
		err := checkEqual(testBreakdown, test.expected)
		checkError(t, err, fmt.Sprintf("checkError: test: %v (GenerateDashboardJSON)\n", test.name))
	}
}

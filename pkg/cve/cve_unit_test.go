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

package cve

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCVEValidation(t *testing.T) {
	var sut CVE
	cve := CVE{
		ID:            "CVE-2020-8559",
		Title:         "Privilege escalation from compromised node to cluster",
		Description:   "If an attacker is able to intercept certain requests to the Kubelet, they",
		TrackingIssue: "https://github.com/kubernetes/kubernetes/issues/92914",
		CVSSVector:    "CVSS:3.1/AV:N/AC:H/PR:H/UI:R/S:U/C:H/I:H/A:H",
		CVSSScore:     6.4,
		CVSSRating:    "Medium",
		LinkedPRs:     []int{92941, 92969, 92970, 92971},
	}

	// As is, the CVE should validate
	require.Nil(t, cve.Validate())

	// Check each value
	sut = cve
	sut.ID = "CVE-123"
	require.NotNil(t, sut.Validate(), "checking cve id")

	sut = cve
	sut.Title = ""
	require.NotNil(t, sut.Validate(), "checking title")

	sut = cve
	sut.Description = ""
	require.NotNil(t, sut.Validate(), "checking description")

	sut = cve
	for _, testVector := range []string{
		"CVSS:3.1/AV:N/AC:H/P", //  too short
		"",                     // empty
		"CVSS:3.1/AV:N/AC:Á/PR:H/UI:R/S:U/C:H/I:H/A:H", //  invalid value
	} {
		sut.CVSSVector = testVector // too short
		require.NotNil(t, sut.Validate(), "checking vector string")
	}

	sut = cve
	for _, testScore := range []float32{
		0,    // cannot be 0 (nil value)
		-1,   // under
		10.1, // over
	} {
		sut.CVSSScore = testScore
		require.NotNil(t, sut.Validate(), "checking vector string")
	}

	sut = cve
	for _, tc := range []struct {
		Valid bool
		Value string
	}{
		{true, "None"},
		{true, "Low"},
		{true, "Medium"},
		{true, "High"},
		{true, "Critical"},
		{false, ""},
		{false, "Superbad"},
	} {
		sut.CVSSRating = tc.Value // too short
		if tc.Valid {
			require.Nil(t, sut.Validate(), "checking valid rating string")
		} else {
			require.NotNil(t, sut.Validate(), "checking invalid rating string")
		}
	}
}

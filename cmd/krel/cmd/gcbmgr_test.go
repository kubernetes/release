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

package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/release/pkg/gcp/build"
)

func TestRunGcbmgrSuccess(t *testing.T) {
	testcases := []struct {
		name       string
		gcbmgrOpts gcbmgrOptions
		buildOpts  build.Options
		expected   map[string]string
	}{
		{
			name: "list only",
			gcbmgrOpts: gcbmgrOptions{
				branch:  "master",
				gcpUser: "test-user",
			},
		},
		{
			name: "stream the job",
			gcbmgrOpts: gcbmgrOptions{
				branch:  "master",
				stream:  true,
				gcpUser: "test-user",
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		gcbmgrOpts = &tc.gcbmgrOpts

		err := runGcbmgr()
		assert.Nil(t, err)
	}
}

func TestRunGcbmgrFailure(t *testing.T) {
	testcases := []struct {
		name       string
		gcbmgrOpts gcbmgrOptions
		expected   map[string]string
	}{
		{
			name: "no release branch",
			gcbmgrOpts: gcbmgrOptions{
				branch:  "",
				gcpUser: "test-user",
			},
		},
		{
			name: "specify stage and release",
			gcbmgrOpts: gcbmgrOptions{
				stage:   true,
				release: true,
				gcpUser: "test-user",
			},
		},
	}

	for _, tc := range testcases {
		fmt.Printf("Test case: %s", tc.name)

		gcbmgrOpts = &tc.gcbmgrOpts

		err := runGcbmgr()
		assert.Error(t, err)
	}
}

func TestSetGCBSubstitutionsSuccess(t *testing.T) {
	testcases := []struct {
		name       string
		gcbmgrOpts gcbmgrOptions
		toolOrg    string
		toolRepo   string
		toolBranch string
		expected   map[string]string
	}{
		{
			name: "master prerelease",
			gcbmgrOpts: gcbmgrOptions{
				branch:      "master",
				releaseType: "prerelease",
				gcpUser:     "test-user",
			},
			expected: map[string]string{
				"BUILDVERSION":   "--buildversion=",
				"BUILD_AT_HEAD":  "",
				"BUILD_POINT":    "",
				"OFFICIAL":       "",
				"OFFICIAL_TAG":   "",
				"RC":             "",
				"RC_TAG":         "",
				"RELEASE_BRANCH": "master",
				"TOOL_ORG":       "kubernetes",
				"TOOL_REPO":      "release",
				"TOOL_BRANCH":    "master",
			},
		},
		{
			name: "release-1.14 RC",
			gcbmgrOpts: gcbmgrOptions{
				branch:      "release-1.14",
				releaseType: "rc",
				gcpUser:     "test-user",
			},
			expected: map[string]string{
				"BUILDVERSION":   "--buildversion=",
				"BUILD_AT_HEAD":  "",
				"BUILD_POINT":    "",
				"OFFICIAL":       "",
				"OFFICIAL_TAG":   "",
				"RC":             "--rc",
				"RC_TAG":         "rc",
				"RELEASE_BRANCH": "release-1.14",
				"TOOL_ORG":       "kubernetes",
				"TOOL_REPO":      "release",
				"TOOL_BRANCH":    "master",
			},
		},
		{
			name: "release-1.15 official",
			gcbmgrOpts: gcbmgrOptions{
				branch:      "release-1.15",
				releaseType: "official",
				gcpUser:     "test-user",
			},
			expected: map[string]string{
				"BUILDVERSION":   "--buildversion=",
				"BUILD_AT_HEAD":  "",
				"BUILD_POINT":    "",
				"OFFICIAL":       "--official",
				"OFFICIAL_TAG":   "official",
				"RC":             "",
				"RC_TAG":         "",
				"RELEASE_BRANCH": "release-1.15",
				"TOOL_ORG":       "kubernetes",
				"TOOL_REPO":      "release",
				"TOOL_BRANCH":    "master",
			},
		},
		{
			name: "release-1.16 official with custom tool org, repo, and branch",
			gcbmgrOpts: gcbmgrOptions{
				branch:      "release-1.16",
				releaseType: "official",
				gcpUser:     "test-user",
			},
			toolOrg:    "honk",
			toolRepo:   "best-tools",
			toolBranch: "tool-branch",
			expected: map[string]string{
				"BUILDVERSION":   "--buildversion=",
				"BUILD_AT_HEAD":  "",
				"BUILD_POINT":    "",
				"OFFICIAL":       "--official",
				"OFFICIAL_TAG":   "official",
				"RC":             "",
				"RC_TAG":         "",
				"RELEASE_BRANCH": "release-1.16",
				"TOOL_ORG":       "honk",
				"TOOL_REPO":      "best-tools",
				"TOOL_BRANCH":    "tool-branch",
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		opts := tc.gcbmgrOpts
		os.Setenv("TOOL_ORG", tc.toolOrg)
		os.Setenv("TOOL_REPO", tc.toolRepo)
		os.Setenv("TOOL_BRANCH", tc.toolBranch)

		subs, err := setGCBSubstitutions(&opts)
		actual := dropDynamicSubstitutions(subs)

		if err != nil {
			t.Fatalf("did not expect an error: %v", err)
		}

		assert.Equal(t, tc.expected, actual)
	}
}

func TestSetGCBSubstitutionsFailure(t *testing.T) {
	testcases := []struct {
		name       string
		gcbmgrOpts gcbmgrOptions
		expected   map[string]string
	}{
		{
			name: "no release branch",
			gcbmgrOpts: gcbmgrOptions{
				branch:  "",
				gcpUser: "test-user",
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		opts := tc.gcbmgrOpts

		_, err := setGCBSubstitutions(&opts)
		assert.Error(t, err)
	}
}

func dropDynamicSubstitutions(orig map[string]string) (result map[string]string) {
	result = orig

	for k := range result {
		if k == "GCP_USER_TAG" || k == "KUBE_CROSS_VERSION" {
			delete(result, k)
		}
	}

	return result
}

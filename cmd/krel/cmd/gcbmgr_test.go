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
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/gcp/build"
)

type fakeListJobs struct {
	expectedProject  string
	expectedLastJobs int64
	t                *testing.T
	err              error
}

func (f *fakeListJobs) ListJobs(project string, lastJobs int64) error {
	require.Equal(f.t, project, f.expectedProject)
	require.Equal(f.t, lastJobs, f.expectedLastJobs)
	return f.err
}

func TestRunGcbmgrList(t *testing.T) {
	testcases := []struct {
		name        string
		gcbmgrOpts  gcbmgrOptions
		buildOpts   build.Options
		listJobOpts fakeListJobs
		expectedErr bool
	}{
		{
			name: "list only",
			gcbmgrOpts: gcbmgrOptions{
				branch:   "master",
				gcpUser:  "test-user",
				lastJobs: 5,
			},
			listJobOpts: fakeListJobs{
				expectedProject:  "kubernetes-release-test",
				expectedLastJobs: int64(5),
				err:              nil,
			},
			expectedErr: false,
		},
		{
			name: "error on list jobs",
			gcbmgrOpts: gcbmgrOptions{
				branch:   "master",
				gcpUser:  "test-user",
				lastJobs: 10,
			},
			listJobOpts: fakeListJobs{
				expectedProject:  "kubernetes-release-test",
				expectedLastJobs: int64(10),
				err:              errors.New("Generic Error"),
			},
			expectedErr: true,
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		gcbmgrOpts = &tc.gcbmgrOpts

		f := &tc.listJobOpts
		f.t = t
		buildListJobs = f.ListJobs

		err := runGcbmgr()
		if tc.expectedErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
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
		require.Error(t, err)
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
			name: "master prerelease - stage",
			gcbmgrOpts: gcbmgrOptions{
				stage:       true,
				branch:      "master",
				releaseType: "prerelease",
				gcpUser:     "test-user",
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
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
			name: "master prerelease - release",
			gcbmgrOpts: gcbmgrOptions{
				release:      true,
				branch:       "master",
				releaseType:  "prerelease",
				buildVersion: "v1.33.7",
				gcpUser:      "test-user",
			},
			expected: map[string]string{
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
				stage:       true,
				branch:      "release-1.14",
				releaseType: "rc",
				gcpUser:     "test-user",
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
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
				stage:       true,
				branch:      "release-1.15",
				releaseType: "official",
				gcpUser:     "test-user",
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
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
				stage:       true,
				branch:      "release-1.16",
				releaseType: "official",
				gcpUser:     "test-user",
			},
			toolOrg:    "honk",
			toolRepo:   "best-tools",
			toolBranch: "tool-branch",
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
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

		require.Equal(t, tc.expected, actual)
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
		{
			name: "no build version",
			gcbmgrOpts: gcbmgrOptions{
				release: true,
				branch:  "",
				gcpUser: "test-user",
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		opts := tc.gcbmgrOpts

		_, err := setGCBSubstitutions(&opts)
		require.Error(t, err)
	}
}

func dropDynamicSubstitutions(orig map[string]string) (result map[string]string) {
	result = orig

	for k := range result {
		if k == "BUILDVERSION" || k == "BUILD_POINT" || k == "GCP_USER_TAG" || k == "KUBE_CROSS_VERSION" {
			delete(result, k)
		}
	}

	return result
}

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

package cmd_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/cmd/krel/cmd"
	"k8s.io/release/cmd/krel/cmd/cmdfakes"
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
	mock := &cmdfakes.FakeRepository{}
	mock.OpenReturns(nil)
	mock.CheckStateReturns(nil)
	mock.GetTagReturns("v1.0.0-20201010", nil)

	testcases := []struct {
		name        string
		gcbmgrOpts  cmd.GcbmgrOptions
		buildOpts   build.Options
		listJobOpts fakeListJobs
		expectedErr bool
	}{
		{
			name: "list only",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Branch:   "master",
				GcpUser:  "test-user",
				LastJobs: 5,
				Repo:     mock,
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
			gcbmgrOpts: cmd.GcbmgrOptions{
				Branch:   "master",
				GcpUser:  "test-user",
				LastJobs: 10,
				Repo:     mock,
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

		cmd.GcbmgrOpts = &tc.gcbmgrOpts

		f := &tc.listJobOpts
		f.t = t
		cmd.BuildListJobs = f.ListJobs

		err := cmd.RunGcbmgr()
		if tc.expectedErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestRunGcbmgrFailure(t *testing.T) {
	mock := &cmdfakes.FakeRepository{}
	mock.OpenReturns(nil)
	mock.CheckStateReturns(nil)
	mock.GetTagReturns("v1.0.0-20201010", nil)

	testcases := []struct {
		name       string
		gcbmgrOpts cmd.GcbmgrOptions
		expected   map[string]string
	}{
		{
			name: "no release branch",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Branch:  "",
				GcpUser: "test-user",
				Repo:    mock,
			},
		},
		{
			name: "specify stage and release",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Stage:   true,
				Release: true,
				GcpUser: "test-user",
				Repo:    mock,
			},
		},
	}

	for _, tc := range testcases {
		fmt.Printf("Test case: %s", tc.name)

		cmd.GcbmgrOpts = &tc.gcbmgrOpts

		err := cmd.RunGcbmgr()
		require.Error(t, err)
	}
}

func TestSetGCBSubstitutionsSuccess(t *testing.T) {
	mock := &cmdfakes.FakeRepository{}
	mock.OpenReturns(nil)
	mock.CheckStateReturns(nil)
	mock.GetTagReturns("v1.0.0-20201010", nil)

	testcases := []struct {
		name       string
		gcbmgrOpts cmd.GcbmgrOptions
		toolOrg    string
		toolRepo   string
		toolBranch string
		expected   map[string]string
	}{
		{
			name: "master prerelease - stage",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "master",
				ReleaseType: "prerelease",
				GcpUser:     "test-user",
				Repo:        mock,
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
				"OFFICIAL":       "",
				"OFFICIAL_TAG":   "",
				"RC":             "",
				"RC_TAG":         "",
				"RELEASE_BRANCH": "master",
				"TOOL_ORG":       "",
				"TOOL_REPO":      "",
				"TOOL_BRANCH":    "",
			},
		},
		{
			name: "master prerelease - release",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Release:      true,
				Branch:       "master",
				ReleaseType:  "prerelease",
				BuildVersion: "v1.33.7",
				GcpUser:      "test-user",
				Repo:         mock,
			},
			expected: map[string]string{
				"OFFICIAL":       "",
				"OFFICIAL_TAG":   "",
				"RC":             "",
				"RC_TAG":         "",
				"RELEASE_BRANCH": "master",
				"TOOL_ORG":       "",
				"TOOL_REPO":      "",
				"TOOL_BRANCH":    "",
			},
		},
		{
			name: "release-1.14 RC",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.14",
				ReleaseType: "rc",
				GcpUser:     "test-user",
				Repo:        mock,
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
				"OFFICIAL":       "",
				"OFFICIAL_TAG":   "",
				"RC":             "--rc",
				"RC_TAG":         "rc",
				"RELEASE_BRANCH": "release-1.14",
				"TOOL_ORG":       "",
				"TOOL_REPO":      "",
				"TOOL_BRANCH":    "",
			},
		},
		{
			name: "release-1.15 official",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: "official",
				GcpUser:     "test-user",
				Repo:        mock,
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
				"OFFICIAL":       "--official",
				"OFFICIAL_TAG":   "official",
				"RC":             "",
				"RC_TAG":         "",
				"RELEASE_BRANCH": "release-1.15",
				"TOOL_ORG":       "",
				"TOOL_REPO":      "",
				"TOOL_BRANCH":    "",
			},
		},
		{
			name: "release-1.16 official with custom tool org, repo, and branch",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.16",
				ReleaseType: "official",
				GcpUser:     "test-user",
				Repo:        mock,
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
		subs, err := cmd.SetGCBSubstitutions(
			&opts, tc.toolOrg, tc.toolRepo, tc.toolBranch,
		)
		actual := dropDynamicSubstitutions(subs)

		if err != nil {
			t.Fatalf("did not expect an error: %v", err)
		}

		require.Equal(t, tc.expected, actual)
	}
}

func TestSetGCBSubstitutionsFailure(t *testing.T) {
	mock := &cmdfakes.FakeRepository{}
	mock.OpenReturns(nil)
	mock.CheckStateReturns(nil)
	mock.GetTagReturns("v1.0.0-20201010", nil)

	testcases := []struct {
		name       string
		gcbmgrOpts cmd.GcbmgrOptions
		expected   map[string]string
	}{
		{
			name: "no release branch",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Branch:  "",
				GcpUser: "test-user",
				Repo:    mock,
			},
		},
		{
			name: "no build version",
			gcbmgrOpts: cmd.GcbmgrOptions{
				Release: true,
				Branch:  "",
				GcpUser: "test-user",
				Repo:    mock,
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		opts := tc.gcbmgrOpts

		_, err := cmd.SetGCBSubstitutions(&opts, "", "", "")
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

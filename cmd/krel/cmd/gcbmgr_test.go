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
	"k8s.io/release/pkg/git"
)

func mockRepo() cmd.Repository {
	mock := &cmdfakes.FakeRepository{}
	mock.OpenReturns(nil)
	mock.CheckStateReturns(nil)
	mock.GetTagReturns("v1.0.0-20201010", nil)
	return mock
}

func mockVersion() cmd.Version {
	mock := &cmdfakes.FakeVersion{}
	mock.GetKubeVersionForBranchReturns("v1.17.0", nil)
	return mock
}

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
		gcbmgrOpts  *cmd.GcbmgrOptions
		buildOpts   build.Options
		listJobOpts fakeListJobs
		expectedErr bool
	}{
		{
			name: "list only",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:   "master",
				GcpUser:  "test-user",
				LastJobs: 5,
				Repo:     mockRepo(),
				Version:  mockVersion(),
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
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:   "master",
				GcpUser:  "test-user",
				LastJobs: 10,
				Repo:     mockRepo(),
				Version:  mockVersion(),
			},
			listJobOpts: fakeListJobs{
				expectedProject:  "kubernetes-release-test",
				expectedLastJobs: int64(10),
				err:              errors.New("Generic Error"),
			},
			expectedErr: true,
		},
	}

	// Restore the previous state
	defer func() { cmd.BuildListJobs = build.ListJobs }()
	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)
		f := &tc.listJobOpts
		f.t = t
		cmd.BuildListJobs = f.ListJobs

		err := cmd.RunGcbmgr(tc.gcbmgrOpts)
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
		gcbmgrOpts *cmd.GcbmgrOptions
		expected   map[string]string
	}{
		{
			name: "no release branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:  "",
				GcpUser: "test-user",
				Repo:    mockRepo(),
				Version: mockVersion(),
			},
		},
		{
			name: "specify stage and release",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:   true,
				Release: true,
				GcpUser: "test-user",
				Repo:    mockRepo(),
				Version: mockVersion(),
			},
		},
	}

	for _, tc := range testcases {
		fmt.Printf("Test case: %s\n", tc.name)
		err := cmd.RunGcbmgr(tc.gcbmgrOpts)
		require.Error(t, err)
	}
}

func TestSetGCBSubstitutionsSuccess(t *testing.T) {
	testcases := []struct {
		name       string
		gcbmgrOpts *cmd.GcbmgrOptions
		toolOrg    string
		toolRepo   string
		toolBranch string
		expected   map[string]string
	}{
		{
			name: "master alpha - stage",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "master",
				ReleaseType: cmd.ReleaseTypeAlpha,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion(),
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
				"RELEASE_BRANCH": "master",
				"TOOL_ORG":       "",
				"TOOL_REPO":      "",
				"TOOL_BRANCH":    "",
				"TYPE":           cmd.ReleaseTypeAlpha,
				"TYPE_TAG":       cmd.ReleaseTypeAlpha,
			},
		},
		{
			name: "master beta - release",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Release:      true,
				Branch:       "master",
				ReleaseType:  cmd.ReleaseTypeBeta,
				BuildVersion: "v1.33.7",
				GcpUser:      "test-user",
				Repo:         mockRepo(),
				Version:      mockVersion(),
			},
			expected: map[string]string{
				"RELEASE_BRANCH": "master",
				"TOOL_ORG":       "",
				"TOOL_REPO":      "",
				"TOOL_BRANCH":    "",
				"TYPE":           cmd.ReleaseTypeBeta,
				"TYPE_TAG":       cmd.ReleaseTypeBeta,
			},
		},
		{
			name: "release-1.15 RC",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: cmd.ReleaseTypeRC,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion(),
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
				"RELEASE_BRANCH": "release-1.15",
				"TOOL_ORG":       "",
				"TOOL_REPO":      "",
				"TOOL_BRANCH":    "",
				"TYPE":           cmd.ReleaseTypeRC,
				"TYPE_TAG":       cmd.ReleaseTypeRC,
			},
		},
		{
			name: "release-1.15 official",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: cmd.ReleaseTypeOfficial,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion(),
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
				"RELEASE_BRANCH": "release-1.15",
				"TOOL_ORG":       "",
				"TOOL_REPO":      "",
				"TOOL_BRANCH":    "",
				"TYPE":           cmd.ReleaseTypeOfficial,
				"TYPE_TAG":       cmd.ReleaseTypeOfficial,
			},
		},
		{
			name: "release-1.16 official with custom tool org, repo, and branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.16",
				ReleaseType: cmd.ReleaseTypeOfficial,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion(),
			},
			toolOrg:    "honk",
			toolRepo:   "best-tools",
			toolBranch: "tool-branch",
			expected: map[string]string{
				"BUILD_AT_HEAD":  "",
				"RELEASE_BRANCH": "release-1.16",
				"TOOL_ORG":       "honk",
				"TOOL_REPO":      "best-tools",
				"TOOL_BRANCH":    "tool-branch",
				"TYPE":           cmd.ReleaseTypeOfficial,
				"TYPE_TAG":       cmd.ReleaseTypeOfficial,
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		subs, err := cmd.SetGCBSubstitutions(
			tc.gcbmgrOpts, tc.toolOrg, tc.toolRepo, tc.toolBranch,
		)
		require.Nil(t, err)

		actual := dropDynamicSubstitutions(subs)
		require.Equal(t, tc.expected, actual)
	}
}

func TestSetGCBSubstitutionsFailure(t *testing.T) {
	testcases := []struct {
		name       string
		gcbmgrOpts *cmd.GcbmgrOptions
		expected   map[string]string
	}{
		{
			name: "no release branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:  "",
				GcpUser: "test-user",
				Repo:    mockRepo(),
				Version: func() cmd.Version {
					m := &cmdfakes.FakeVersion{}
					m.GetKubeVersionForBranchReturns("", errors.New(""))
					return m
				}(),
			},
		},
		{
			name: "no build version",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Release: true,
				Branch:  "",
				GcpUser: "test-user",
				Repo:    mockRepo(),
				Version: mockVersion(),
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)
		_, err := cmd.SetGCBSubstitutions(tc.gcbmgrOpts, "", "", "")
		require.Error(t, err)
	}
}

func TestValidateSuccess(t *testing.T) {
	testcases := []struct {
		name       string
		gcbmgrOpts *cmd.GcbmgrOptions
	}{
		{
			name: "master alpha - stage",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "master",
				ReleaseType: cmd.ReleaseTypeAlpha,
			},
		},
		{
			name: "master beta - release",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Release:      true,
				Branch:       "master",
				ReleaseType:  cmd.ReleaseTypeBeta,
				BuildVersion: "v1.33.7",
			},
		},
		{
			name: "release-1.15 RC",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: cmd.ReleaseTypeRC,
			},
		},
		{
			name: "release-1.15 official",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: cmd.ReleaseTypeOfficial,
			},
		},
		{
			name: "release-1.16 official with custom tool org, repo, and branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.16",
				ReleaseType: cmd.ReleaseTypeOfficial,
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		err := tc.gcbmgrOpts.Validate()
		require.Nil(t, err)
	}
}

func TestValidateFailure(t *testing.T) {
	testcases := []struct {
		name       string
		gcbmgrOpts *cmd.GcbmgrOptions
	}{
		{
			name: "RC on master",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:      git.Master,
				ReleaseType: cmd.ReleaseTypeRC,
			},
		},
		{
			name: "official release on master",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:      git.Master,
				ReleaseType: cmd.ReleaseTypeOfficial,
			},
		},
		{
			name: "alpha on release branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:      "release-1.16",
				ReleaseType: cmd.ReleaseTypeAlpha,
			},
		},
		{
			name: "beta on release branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:      "release-1.19",
				ReleaseType: cmd.ReleaseTypeBeta,
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		err := tc.gcbmgrOpts.Validate()
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

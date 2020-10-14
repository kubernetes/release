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
	"k8s.io/release/pkg/release"
)

func mockRepo() cmd.Repository {
	mock := &cmdfakes.FakeRepository{}
	mock.OpenReturns(nil)
	mock.CheckStateReturns(nil)
	mock.GetTagReturns("v1.0.0-20201010", nil)
	return mock
}

func mockVersion(version string) cmd.Version {
	mock := &cmdfakes.FakeVersion{}
	mock.GetKubeVersionForBranchReturns(version, nil)
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
				Branch:   git.DefaultBranch,
				GcpUser:  "test-user",
				LastJobs: 5,
				Repo:     mockRepo(),
				Version:  mockVersion("v1.17.0"),
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
				Branch:   git.DefaultBranch,
				GcpUser:  "test-user",
				LastJobs: 10,
				Repo:     mockRepo(),
				Version:  mockVersion("v1.17.0"),
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
				Version: mockVersion("v1.17.0"),
			},
		},
		{
			name: "specify stage and release",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:   true,
				Release: true,
				GcpUser: "test-user",
				Repo:    mockRepo(),
				Version: mockVersion("v1.17.0"),
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
			name: "main branch alpha - stage",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      git.DefaultBranch,
				ReleaseType: release.ReleaseTypeAlpha,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion("v1.17.0"),
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":          "",
				"RELEASE_BRANCH":         git.DefaultBranch,
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_BRANCH":            "",
				"TYPE":                   release.ReleaseTypeAlpha,
				"TYPE_TAG":               release.ReleaseTypeAlpha,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "17",
				"PATCH_VERSION_TAG":      "0",
				"KUBERNETES_VERSION_TAG": "1.17.0",
			},
		},
		{
			name: "main branch beta - release",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Release:      true,
				Branch:       git.DefaultBranch,
				ReleaseType:  release.ReleaseTypeBeta,
				BuildVersion: "v1.33.7",
				GcpUser:      "test-user",
				Repo:         mockRepo(),
				Version:      mockVersion("v1.33.7"),
			},
			expected: map[string]string{
				"RELEASE_BRANCH":         git.DefaultBranch,
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_BRANCH":            "",
				"TYPE":                   release.ReleaseTypeBeta,
				"TYPE_TAG":               release.ReleaseTypeBeta,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "33",
				"PATCH_VERSION_TAG":      "7",
				"KUBERNETES_VERSION_TAG": "1.33.7",
			},
		},
		{
			name: "release-1.15 RC",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: release.ReleaseTypeRC,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion("v1.15.0-rc.1"),
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":          "",
				"RELEASE_BRANCH":         "release-1.15",
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_BRANCH":            "",
				"TYPE":                   release.ReleaseTypeRC,
				"TYPE_TAG":               release.ReleaseTypeRC,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "15",
				"KUBERNETES_VERSION_TAG": "1.15.0-rc.2",
				"PATCH_VERSION_TAG":      "0-rc.2",
			},
		},
		{
			name: "release-1.15 official",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: release.ReleaseTypeOfficial,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion("v1.15.1"),
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":          "",
				"RELEASE_BRANCH":         "release-1.15",
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_BRANCH":            "",
				"TYPE":                   release.ReleaseTypeOfficial,
				"TYPE_TAG":               release.ReleaseTypeOfficial,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "15",
				"PATCH_VERSION_TAG":      "1",
				"KUBERNETES_VERSION_TAG": "1.15.1",
			},
		},
		{
			name: "release-1.16 official with custom tool org, repo, and branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.16",
				ReleaseType: release.ReleaseTypeOfficial,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion("v1.16.0"),
			},
			toolOrg:    "honk",
			toolRepo:   "best-tools",
			toolBranch: "tool-branch",
			expected: map[string]string{
				"BUILD_AT_HEAD":          "",
				"RELEASE_BRANCH":         "release-1.16",
				"TOOL_ORG":               "honk",
				"TOOL_REPO":              "best-tools",
				"TOOL_BRANCH":            "tool-branch",
				"TYPE":                   release.ReleaseTypeOfficial,
				"TYPE_TAG":               release.ReleaseTypeOfficial,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "16",
				"PATCH_VERSION_TAG":      "0",
				"KUBERNETES_VERSION_TAG": "1.16.0",
			},
		},
		{
			name: "release-1.19 beta 1 with custom tool org, repo, branch and full build point",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.19",
				ReleaseType: release.ReleaseTypeBeta,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion("v1.19.0-alpha.2.763+2da917d3701904"),
			},
			toolOrg:    "honk",
			toolRepo:   "best-tools",
			toolBranch: "tool-branch",
			expected: map[string]string{
				"BUILD_AT_HEAD":          "",
				"RELEASE_BRANCH":         "release-1.19",
				"TOOL_ORG":               "honk",
				"TOOL_REPO":              "best-tools",
				"TOOL_BRANCH":            "tool-branch",
				"TYPE":                   release.ReleaseTypeBeta,
				"TYPE_TAG":               release.ReleaseTypeBeta,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "19",
				"PATCH_VERSION_TAG":      "0-beta.0",
				"KUBERNETES_VERSION_TAG": "1.19.0-beta.0",
			},
		},
		{
			name: "release-1.18 RC 1",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.18",
				ReleaseType: release.ReleaseTypeRC,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion("v1.18.6-rc.0.15+e38139724f8f00"),
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":          "",
				"RELEASE_BRANCH":         "release-1.18",
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_BRANCH":            "",
				"TYPE":                   release.ReleaseTypeRC,
				"TYPE_TAG":               release.ReleaseTypeRC,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "18",
				"KUBERNETES_VERSION_TAG": "1.18.6-rc.1",
				"PATCH_VERSION_TAG":      "6-rc.1",
			},
		},
		{
			name: "release-1.18 RC 1 from Beta",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.18",
				ReleaseType: release.ReleaseTypeRC,
				GcpUser:     "test-user",
				Repo:        mockRepo(),
				Version:     mockVersion("v1.18.0-beta.4.15+e38139724f8f00"),
			},
			expected: map[string]string{
				"BUILD_AT_HEAD":          "",
				"RELEASE_BRANCH":         "release-1.18",
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_BRANCH":            "",
				"TYPE":                   release.ReleaseTypeRC,
				"TYPE_TAG":               release.ReleaseTypeRC,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "18",
				"KUBERNETES_VERSION_TAG": "1.18.0-rc.1",
				"PATCH_VERSION_TAG":      "0-rc.1",
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
				Version: mockVersion("v1.17.0"),
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
			name: "main branch alpha - stage",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      git.DefaultBranch,
				ReleaseType: release.ReleaseTypeAlpha,
			},
		},
		{
			name: "main branch beta - release",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Release:      true,
				Branch:       git.DefaultBranch,
				ReleaseType:  release.ReleaseTypeBeta,
				BuildVersion: "v1.33.7",
			},
		},
		{
			name: "release-1.15 RC",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: release.ReleaseTypeRC,
			},
		},
		{
			name: "release-1.15 official",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: release.ReleaseTypeOfficial,
			},
		},
		{
			name: "release-1.16 official with custom tool org, repo, and branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Stage:       true,
				Branch:      "release-1.16",
				ReleaseType: release.ReleaseTypeOfficial,
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
			name: "RC on main branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:      git.DefaultBranch,
				ReleaseType: release.ReleaseTypeRC,
			},
		},
		{
			name: "official release on main branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:      git.DefaultBranch,
				ReleaseType: release.ReleaseTypeOfficial,
			},
		},
		{
			name: "alpha on release branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:      "release-1.16",
				ReleaseType: release.ReleaseTypeAlpha,
			},
		},
		{
			name: "beta on release branch",
			gcbmgrOpts: &cmd.GcbmgrOptions{
				Branch:      "release-1.19",
				ReleaseType: release.ReleaseTypeBeta,
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

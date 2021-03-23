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

package gcb_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/gcp/gcb/gcbfakes"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

func mockRepo() gcb.Repository {
	mock := &gcbfakes.FakeRepository{}
	mock.OpenReturns(nil)
	mock.CheckStateReturns(nil)
	mock.GetTagReturns("v1.0.0-20201010", nil)
	return mock
}

func mockVersion(version string) gcb.Version {
	mock := &gcbfakes.FakeVersion{}
	mock.GetKubeVersionForBranchReturns(version, nil)
	return mock
}

func mockRelease(version string) gcb.Release {
	mock := &gcbfakes.FakeRelease{}
	mock.GenerateReleaseVersionReturns(
		release.NewReleaseVersions(version, "", "", "", ""), nil,
	)
	return mock
}

func TestSubmitList(t *testing.T) {
	testcases := []struct {
		name         string
		gcbOpts      *gcb.Options
		expectedErr  bool
		repoMock     gcb.Repository
		versionMock  gcb.Version
		listJobsMock gcb.ListJobs
		releaseMock  gcb.Release
	}{
		{
			name: "list only",
			gcbOpts: &gcb.Options{
				Branch:   git.DefaultBranch,
				GcpUser:  "test-user",
				LastJobs: 5,
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.17.0"),
			listJobsMock: func() gcb.ListJobs {
				return &gcbfakes.FakeListJobs{}
			}(),
			releaseMock: mockRelease("v1.17.0"),
			expectedErr: false,
		},
		{
			name: "error on list jobs",
			gcbOpts: &gcb.Options{
				Branch:   git.DefaultBranch,
				GcpUser:  "test-user",
				LastJobs: 10,
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.17.0"),
			listJobsMock: func() gcb.ListJobs {
				m := &gcbfakes.FakeListJobs{}
				m.ListJobsReturns(errors.New(""))
				return m
			}(),
			releaseMock: mockRelease("v1.17.0"),
			expectedErr: true,
		},
	}

	// Restore the previous state
	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		sut := gcb.New(tc.gcbOpts)
		sut.SetRepoClient(tc.repoMock)
		sut.SetVersionClient(tc.versionMock)
		sut.SetListJobsClient(tc.listJobsMock)
		sut.SetReleaseClient(tc.releaseMock)
		err := sut.Submit()

		if tc.expectedErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestSubmitGcbFailure(t *testing.T) {
	testcases := []struct {
		name        string
		gcbOpts     *gcb.Options
		expected    map[string]string
		repoMock    gcb.Repository
		versionMock gcb.Version
	}{
		{
			name: "no release branch",
			gcbOpts: &gcb.Options{
				Branch:  "",
				GcpUser: "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.17.0"),
		},
		{
			name: "specify stage and release",
			gcbOpts: &gcb.Options{
				Stage:   true,
				Release: true,
				GcpUser: "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.17.0"),
		},
	}

	for _, tc := range testcases {
		fmt.Printf("Test case: %s\n", tc.name)
		sut := gcb.New(tc.gcbOpts)
		sut.SetRepoClient(tc.repoMock)
		sut.SetVersionClient(tc.versionMock)
		err := sut.Submit()
		require.Error(t, err)
	}
}

func TestSetGCBSubstitutionsSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		gcbOpts     *gcb.Options
		toolOrg     string
		toolRepo    string
		toolRef     string
		expected    map[string]string
		repoMock    gcb.Repository
		versionMock gcb.Version
		releaseMock gcb.Release
	}{
		{
			name: "main branch alpha - stage",
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      git.DefaultBranch,
				ReleaseType: release.ReleaseTypeAlpha,
				GcpUser:     "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.17.0"),
			releaseMock: mockRelease("v1.17.0"),
			expected: map[string]string{
				"RELEASE_BRANCH":         git.DefaultBranch,
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_REF":               "",
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
			gcbOpts: &gcb.Options{
				Release:      true,
				Branch:       git.DefaultBranch,
				ReleaseType:  release.ReleaseTypeBeta,
				BuildVersion: "v1.33.7",
				GcpUser:      "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.33.7"),
			releaseMock: mockRelease("v1.33.7"),
			expected: map[string]string{
				"RELEASE_BRANCH":         git.DefaultBranch,
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_REF":               "",
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
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: release.ReleaseTypeRC,
				GcpUser:     "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.15.0-rc.1"),
			releaseMock: mockRelease("v1.15.0-rc.2"),
			expected: map[string]string{
				"RELEASE_BRANCH":         "release-1.15",
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_REF":               "",
				"TYPE":                   release.ReleaseTypeRC,
				"TYPE_TAG":               release.ReleaseTypeRC,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "15",
				"KUBERNETES_VERSION_TAG": "1.15.0-rc.2",
				"PATCH_VERSION_TAG":      "0",
			},
		},
		{
			name: "release-1.15 official",
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: release.ReleaseTypeOfficial,
				GcpUser:     "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.15.1"),
			releaseMock: mockRelease("v1.15.1"),
			expected: map[string]string{
				"RELEASE_BRANCH":         "release-1.15",
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_REF":               "",
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
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      "release-1.16",
				ReleaseType: release.ReleaseTypeOfficial,
				GcpUser:     "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.16.0"),
			releaseMock: mockRelease("v1.16.0"),
			toolOrg:     "honk",
			toolRepo:    "best-tools",
			toolRef:     "tool-branch",
			expected: map[string]string{
				"RELEASE_BRANCH":         "release-1.16",
				"TOOL_ORG":               "honk",
				"TOOL_REPO":              "best-tools",
				"TOOL_REF":               "tool-branch",
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
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      "release-1.19",
				ReleaseType: release.ReleaseTypeBeta,
				GcpUser:     "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.19.0-alpha.2.763+2da917d3701904"),
			releaseMock: mockRelease("1.19.0-beta.0"),
			toolOrg:     "honk",
			toolRepo:    "best-tools",
			toolRef:     "tool-branch",
			expected: map[string]string{
				"RELEASE_BRANCH":         "release-1.19",
				"TOOL_ORG":               "honk",
				"TOOL_REPO":              "best-tools",
				"TOOL_REF":               "tool-branch",
				"TYPE":                   release.ReleaseTypeBeta,
				"TYPE_TAG":               release.ReleaseTypeBeta,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "19",
				"PATCH_VERSION_TAG":      "0",
				"KUBERNETES_VERSION_TAG": "1.19.0-beta.0",
			},
		},
		{
			name: "release-1.18 RC 1",
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      "release-1.18",
				ReleaseType: release.ReleaseTypeRC,
				GcpUser:     "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.18.6-rc.0.15+e38139724f8f00"),
			releaseMock: mockRelease("1.18.6-rc.1"),
			expected: map[string]string{
				"RELEASE_BRANCH":         "release-1.18",
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_REF":               "",
				"TYPE":                   release.ReleaseTypeRC,
				"TYPE_TAG":               release.ReleaseTypeRC,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "18",
				"KUBERNETES_VERSION_TAG": "1.18.6-rc.1",
				"PATCH_VERSION_TAG":      "6",
			},
		},
		{
			name: "release-1.18 RC 1 from Beta",
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      "release-1.18",
				ReleaseType: release.ReleaseTypeRC,
				GcpUser:     "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.18.0-beta.4.15+e38139724f8f00"),
			releaseMock: mockRelease("1.18.0-rc.1"),
			expected: map[string]string{
				"RELEASE_BRANCH":         "release-1.18",
				"TOOL_ORG":               "",
				"TOOL_REPO":              "",
				"TOOL_REF":               "",
				"TYPE":                   release.ReleaseTypeRC,
				"TYPE_TAG":               release.ReleaseTypeRC,
				"MAJOR_VERSION_TAG":      "1",
				"MINOR_VERSION_TAG":      "18",
				"KUBERNETES_VERSION_TAG": "1.18.0-rc.1",
				"PATCH_VERSION_TAG":      "0",
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		sut := gcb.New(tc.gcbOpts)
		sut.SetRepoClient(tc.repoMock)
		sut.SetVersionClient(tc.versionMock)
		sut.SetReleaseClient(tc.releaseMock)

		subs, err := sut.SetGCBSubstitutions(
			tc.toolOrg, tc.toolRepo, tc.toolRef,
		)
		require.Nil(t, err)

		actual := dropDynamicSubstitutions(subs)
		require.Equal(t, tc.expected, actual)
	}
}

func TestSetGCBSubstitutionsFailure(t *testing.T) {
	testcases := []struct {
		name        string
		gcbOpts     *gcb.Options
		expected    map[string]string
		repoMock    gcb.Repository
		versionMock gcb.Version
	}{
		{
			name: "no release branch",
			gcbOpts: &gcb.Options{
				Branch:  "",
				GcpUser: "test-user",
			},
			repoMock: mockRepo(),
			versionMock: func() gcb.Version {
				m := &gcbfakes.FakeVersion{}
				m.GetKubeVersionForBranchReturns("", errors.New(""))
				return m
			}(),
		},
		{
			name: "no build version",
			gcbOpts: &gcb.Options{
				Release: true,
				Branch:  "",
				GcpUser: "test-user",
			},
			repoMock:    mockRepo(),
			versionMock: mockVersion("v1.17.0"),
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)
		sut := gcb.New(tc.gcbOpts)
		sut.SetRepoClient(tc.repoMock)
		sut.SetVersionClient(tc.versionMock)
		_, err := sut.SetGCBSubstitutions("", "", "")
		require.Error(t, err)
	}
}

func TestValidateSuccess(t *testing.T) {
	testcases := []struct {
		name    string
		gcbOpts *gcb.Options
	}{
		{
			name: "main branch alpha - stage",
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      git.DefaultBranch,
				ReleaseType: release.ReleaseTypeAlpha,
			},
		},
		{
			name: "main branch beta - release",
			gcbOpts: &gcb.Options{
				Release:      true,
				Branch:       git.DefaultBranch,
				ReleaseType:  release.ReleaseTypeBeta,
				BuildVersion: "v1.33.7",
			},
		},
		{
			name: "release-1.15 RC",
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: release.ReleaseTypeRC,
			},
		},
		{
			name: "release-1.15 official",
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      "release-1.15",
				ReleaseType: release.ReleaseTypeOfficial,
			},
		},
		{
			name: "release-1.16 official with custom tool org, repo, and branch",
			gcbOpts: &gcb.Options{
				Stage:       true,
				Branch:      "release-1.16",
				ReleaseType: release.ReleaseTypeOfficial,
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		err := tc.gcbOpts.Validate()
		require.Nil(t, err)
	}
}

func TestValidateFailure(t *testing.T) {
	testcases := []struct {
		name    string
		gcbOpts *gcb.Options
	}{
		{
			name: "RC on main branch",
			gcbOpts: &gcb.Options{
				Branch:      git.DefaultBranch,
				ReleaseType: release.ReleaseTypeRC,
			},
		},
		{
			name: "official release on main branch",
			gcbOpts: &gcb.Options{
				Branch:      git.DefaultBranch,
				ReleaseType: release.ReleaseTypeOfficial,
			},
		},
		{
			name: "alpha on release branch",
			gcbOpts: &gcb.Options{
				Branch:      "release-1.16",
				ReleaseType: release.ReleaseTypeAlpha,
			},
		},
		{
			name: "beta on release branch",
			gcbOpts: &gcb.Options{
				Branch:      "release-1.19",
				ReleaseType: release.ReleaseTypeBeta,
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		err := tc.gcbOpts.Validate()
		require.Error(t, err)
	}
}

func dropDynamicSubstitutions(orig map[string]string) (result map[string]string) {
	result = orig

	for k := range result {
		if k == "BUILDVERSION" || k == "GCP_USER_TAG" || strings.HasPrefix(k, "KUBE_CROSS_VERSION") {
			delete(result, k)
		}
	}

	return result
}

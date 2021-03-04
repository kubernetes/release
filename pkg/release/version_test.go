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

package release_test

import (
	"errors"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
	"sigs.k8s.io/release-utils/util"
)

func newVersionSUT() (*release.Version, *releasefakes.FakeVersionClient) {
	client := &releasefakes.FakeVersionClient{}
	sut := release.NewVersion()
	sut.SetClient(client)

	return sut, client
}

func TestGetKubeVersionSuccess(t *testing.T) {
	testcases := []struct {
		behavior    func(*releasefakes.FakeVersionClient)
		versionType release.VersionType
		assertion   func(semver.Version)
	}{
		{
			behavior: func(mock *releasefakes.FakeVersionClient) {
				mock.GetURLResponseReturns("v1.17.3", nil)
			},
			versionType: release.VersionTypeStable,
			assertion:   func(s semver.Version) { require.Empty(t, s.Pre) },
		},
		{
			behavior: func(mock *releasefakes.FakeVersionClient) {
				mock.GetURLResponseReturns("v1.19.0-alpha.0.721+f8ff8f44206ff4", nil)
			},
			versionType: release.VersionTypeCILatest,
			assertion:   func(s semver.Version) { require.Len(t, s.Pre, 3) },
		},
		{
			behavior: func(mock *releasefakes.FakeVersionClient) {
				mock.GetURLResponseReturns("v1.19.0-alpha.0", nil)
			},
			versionType: release.VersionTypeStablePreRelease,
			assertion:   func(s semver.Version) { require.Len(t, s.Pre, 2) },
		},
	}

	for _, tc := range testcases {
		sut, client := newVersionSUT()
		tc.behavior(client)

		tag, err := sut.GetKubeVersion(tc.versionType)
		require.Nil(t, err)

		s, err := util.TagStringToSemver(tag)
		require.Nil(t, err)

		require.EqualValues(t, s.Major, 1)
		tc.assertion(s)
	}
}

func TestGetKubeVersionForBranchSuccess(t *testing.T) {
	testcases := []struct {
		versionType release.VersionType
		branch      string
		expected    string
	}{
		{
			versionType: release.VersionTypeStable,
			branch:      "release-1.13",
			expected:    "v1.13.12",
		},
		{
			versionType: release.VersionTypeCILatest,
			branch:      "release-1.15",
			expected:    "v1.15.12-beta.0.33+5f400ccfa32aff",
		},
		{
			versionType: release.VersionTypeStablePreRelease,
			branch:      "release-1.15",
			expected:    "v1.15.12-beta.0",
		},
	}

	for _, tc := range testcases {
		sut, client := newVersionSUT()
		client.GetURLResponseReturns(tc.expected, nil)

		actual, err := sut.GetKubeVersionForBranch(tc.versionType, tc.branch)

		require.Nil(t, err, string(tc.versionType))
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetKubeVersionForBranchFailure(t *testing.T) {
	testcases := []struct {
		behavior    func(*releasefakes.FakeVersionClient)
		versionType release.VersionType
		branch      string
	}{
		{
			behavior: func(mock *releasefakes.FakeVersionClient) {
				mock.GetURLResponseReturns("", errors.New(""))
			},
			versionType: release.VersionTypeStable,
			branch:      "wrong-branch",
		},
	}

	for _, tc := range testcases {
		sut, client := newVersionSUT()
		tc.behavior(client)

		_, err := sut.GetKubeVersionForBranch(tc.versionType, tc.branch)
		require.NotNil(t, err, string(tc.versionType))
	}
}

func TestURL(t *testing.T) {
	testcases := []struct {
		versionType release.VersionType
		version     string
		expected    string
	}{
		{
			versionType: release.VersionTypeStable,
			version:     "1.13",
			expected:    "https://dl.k8s.io/release/stable-1.13.txt",
		},
		{
			versionType: release.VersionTypeStable,
			expected:    "https://dl.k8s.io/release/stable.txt",
		},
		{
			versionType: release.VersionTypeCILatest,
			version:     "1.15",
			expected:    "https://dl.k8s.io/ci/latest-1.15.txt",
		},
		{
			versionType: release.VersionTypeCILatest,
			expected:    "https://dl.k8s.io/ci/latest.txt",
		},
		{
			versionType: release.VersionTypeStablePreRelease,
			version:     "1.15",
			expected:    "https://dl.k8s.io/release/latest-1.15.txt",
		},
		{
			versionType: release.VersionTypeStablePreRelease,
			expected:    "https://dl.k8s.io/release/latest.txt",
		},
	}

	for _, tc := range testcases {
		url := tc.versionType.URL(tc.version)
		require.Equal(t, tc.expected, url)
	}
}

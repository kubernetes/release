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

package release

import (
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/http"
	"k8s.io/release/pkg/util"
)

func TestGetKubeVersionSuccess(t *testing.T) {
	testcases := []struct {
		versionType VersionType
		assertion   func(semver.Version)
	}{
		{
			// for example: v1.17.3
			versionType: VersionTypeStable,
			assertion:   func(s semver.Version) { require.Empty(t, s.Pre) },
		},
		{
			// for example: v1.19.0-alpha.0.721+f8ff8f44206ff4
			versionType: VersionTypeCILatest,
			assertion:   func(s semver.Version) { require.Len(t, s.Pre, 3) },
		},
		{
			// for example: v1.19.0-alpha.0
			versionType: VersionTypeStablePreRelease,
			assertion:   func(s semver.Version) { require.Len(t, s.Pre, 2) },
		},
	}

	for _, tc := range testcases {
		tag, err := GetKubeVersion(tc.versionType)
		require.Nil(t, err)

		s, err := util.TagStringToSemver(tag)
		require.Nil(t, err)

		require.EqualValues(t, s.Major, 1)
		tc.assertion(s)
	}
}

func TestGetKubeVersionForBranchSuccess(t *testing.T) {
	testcases := []struct {
		versionType VersionType
		branch      string
		expected    string
	}{
		{
			versionType: VersionTypeStable,
			branch:      "release-1.13",
			expected:    "v1.13.12",
		},
		{
			versionType: VersionTypeCILatest,
			branch:      "release-1.14",
			expected:    "v1.14.11-beta.1.2+c8b135d0b49c44",
		},
		{
			versionType: VersionTypeStablePreRelease,
			branch:      "release-1.14",
			expected:    "v1.14.11-beta.0",
		},
	}

	for _, tc := range testcases {
		actual, err := GetKubeVersionForBranch(tc.versionType, tc.branch)
		require.Nil(t, err, string(tc.versionType))
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetKubeVersionForBranchFailure(t *testing.T) {
	testcases := []struct {
		versionType VersionType
		branch      string
	}{
		{
			versionType: VersionTypeStable,
			branch:      "wrong-branch",
		},
	}

	for _, tc := range testcases {
		_, err := GetKubeVersionForBranch(tc.versionType, tc.branch)
		require.NotNil(t, err, string(tc.versionType))
	}
}

func TestURL(t *testing.T) {
	testcases := []struct {
		versionType VersionType
		version     string
		expected    string
	}{
		{
			versionType: VersionTypeStable,
			version:     "1.13",
			expected:    "https://dl.k8s.io/release/stable-1.13.txt",
		},
		{
			versionType: VersionTypeStable,
			expected:    "https://dl.k8s.io/release/stable.txt",
		},
		{
			versionType: VersionTypeCILatest,
			version:     "1.14",
			expected:    "https://dl.k8s.io/ci/latest-1.14.txt",
		},
		{
			versionType: VersionTypeCILatest,
			expected:    "https://dl.k8s.io/ci/latest.txt",
		},
		{
			versionType: VersionTypeStablePreRelease,
			version:     "1.15",
			expected:    "https://dl.k8s.io/release/latest-1.15.txt",
		},
		{
			versionType: VersionTypeStablePreRelease,
			expected:    "https://dl.k8s.io/release/latest.txt",
		},
	}

	for _, tc := range testcases {
		url := tc.versionType.url(tc.version)
		require.Equal(t, tc.expected, url)

		response, err := http.GetURLResponse(url, true)
		require.Nil(t, err)
		require.NotEmpty(t, response)
	}
}

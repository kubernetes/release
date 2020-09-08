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
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

func TestSetReleaseVersion(t *testing.T) {
	for _, tc := range []struct {
		releaseType  string
		version      string
		branch       string
		parentBranch string
		expect       func(*release.Versions, error)
	}{
		{
			// new final patch release
			releaseType:  release.ReleaseTypeOfficial,
			version:      "v1.18.4-rc.0.3+3ff09514d162b0",
			branch:       "release-1.18",
			parentBranch: "release-1.18",
			expect: func(res *release.Versions, err error) {
				require.Nil(t, err)
				require.Equal(t, "v1.18.4", res.Prime())
				require.Equal(t, "v1.18.4", res.Official())
				require.Equal(t, "v1.18.5-rc.0", res.RC())
				require.Empty(t, res.Beta())
				require.Empty(t, res.Alpha())
				require.Len(t, res.Slice(), 2)
			},
		},
		{
			// new release candidate from release branch
			releaseType:  release.ReleaseTypeRC,
			version:      "v1.18.4-beta.0.2+3ff09514d162b0",
			branch:       "release-1.18",
			parentBranch: "release-1.18",
			expect: func(res *release.Versions, err error) {
				require.Nil(t, err)
				require.Equal(t, "v1.18.4-rc.1", res.Prime())
				require.Empty(t, res.Official())
				require.Equal(t, "v1.18.4-rc.1", res.RC())
				require.Empty(t, res.Beta())
				require.Empty(t, res.Alpha())
				require.Len(t, res.Slice(), 1)
			},
		},
		{
			// new release candidate from master
			releaseType:  release.ReleaseTypeRC,
			version:      "v1.18.0-beta.3.2+3ff09514d162b0",
			branch:       "release-1.18",
			parentBranch: git.Master,
			expect: func(res *release.Versions, err error) {
				require.Nil(t, err)
				require.Equal(t, "v1.18.0-rc.0", res.Prime())
				require.Empty(t, res.Official())
				require.Equal(t, "v1.18.0-rc.0", res.RC())
				require.Empty(t, res.Beta())
				require.Equal(t, "v1.19.0-alpha.0", res.Alpha())
				require.Len(t, res.Slice(), 2)
			},
		},
		{
			// new beta from beta
			releaseType:  release.ReleaseTypeBeta,
			version:      "v1.18.4-beta.1.2+3ff09514d162b0",
			branch:       git.Master,
			parentBranch: "",
			expect: func(res *release.Versions, err error) {
				require.Nil(t, err)
				require.Equal(t, "v1.18.4-beta.1", res.Prime())
				require.Empty(t, res.Official())
				require.Empty(t, res.RC())
				require.Equal(t, "v1.18.4-beta.1", res.Beta())
				require.Empty(t, res.Alpha())
				require.Len(t, res.Slice(), 1)
			},
		},
		{
			// new beta from alpha
			releaseType:  release.ReleaseTypeBeta,
			version:      "v1.18.0-alpha.1.2+3ff09514d162b0",
			branch:       git.Master,
			parentBranch: "",
			expect: func(res *release.Versions, err error) {
				require.Nil(t, err)
				require.Equal(t, "v1.18.0-beta.0", res.Prime())
				require.Empty(t, res.Official())
				require.Empty(t, res.RC())
				require.Equal(t, "v1.18.0-beta.0", res.Beta())
				require.Empty(t, res.Alpha())
				require.Len(t, res.Slice(), 1)
			},
		},
		{
			// new alpha
			releaseType:  release.ReleaseTypeAlpha,
			version:      "v1.18.4-alpha.1.2+3ff09514d162b0",
			branch:       git.Master,
			parentBranch: "",
			expect: func(res *release.Versions, err error) {
				require.Nil(t, err)
				require.Equal(t, "v1.18.4-alpha.1", res.Prime())
				require.Empty(t, res.Official())
				require.Empty(t, res.RC())
				require.Empty(t, res.Beta())
				require.Equal(t, "v1.18.4-alpha.1", res.Alpha())
				require.Len(t, res.Slice(), 1)
			},
		},
		{
			// new alpha after beta
			releaseType:  release.ReleaseTypeAlpha,
			version:      "v1.18.4-beta.1.2+3ff09514d162b0",
			branch:       git.Master,
			parentBranch: "",
			expect: func(res *release.Versions, err error) {
				require.NotNil(t, err)
				require.Nil(t, res)
			},
		},
		{
			// invalid branch
			releaseType:  release.ReleaseTypeOfficial,
			version:      "",
			branch:       "wrong",
			parentBranch: git.Master,
			expect: func(res *release.Versions, err error) {
				require.NotNil(t, err)
				require.Nil(t, res)
			},
		},
	} {
		tc.expect(release.SetReleaseVersion(
			tc.releaseType, tc.version, tc.branch, tc.parentBranch,
		))
	}
}

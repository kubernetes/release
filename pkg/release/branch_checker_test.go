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

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
)

func TestNeedsCreation(t *testing.T) {
	for _, tc := range []struct {
		prepare      func(*releasefakes.FakeBranchCheckerImpl)
		branch       string
		releaseType  string
		buildVersion semver.Version
		shouldErr    bool
		shouldReturn bool
	}{
		{ // success
			branch:       "release-1.19",
			releaseType:  release.ReleaseTypeAlpha,
			buildVersion: semver.MustParse("1.19.0-alpha.2"),
			prepare:      func(*releasefakes.FakeBranchCheckerImpl) {},
			shouldErr:    false,
			shouldReturn: true,
		},
		{ // success branch exists
			branch:       "release-1.19",
			releaseType:  release.ReleaseTypeAlpha,
			buildVersion: semver.MustParse("1.19.0-alpha.2"),
			prepare: func(mock *releasefakes.FakeBranchCheckerImpl) {
				mock.LSRemoteExecReturns("commit", nil)
			},
			shouldErr:    false,
			shouldReturn: false,
		},
		{ // success branch exists default branch
			branch:       git.DefaultBranch,
			releaseType:  release.ReleaseTypeAlpha,
			buildVersion: semver.MustParse("1.19.0-alpha.2"),
			prepare: func(mock *releasefakes.FakeBranchCheckerImpl) {
				mock.LSRemoteExecReturns("commit", nil)
			},
			shouldErr:    false,
			shouldReturn: false,
		},
		{ // failure wrong build version
			branch:       "release-1.20",
			releaseType:  release.ReleaseTypeAlpha,
			buildVersion: semver.MustParse("1.19.0-alpha.2"),
			prepare:      func(*releasefakes.FakeBranchCheckerImpl) {},
			shouldErr:    true,
			shouldReturn: false,
		},
		{ // failure official
			branch:       "release-1.19",
			releaseType:  release.ReleaseTypeOfficial,
			buildVersion: semver.MustParse("1.19.0"),
			prepare:      func(*releasefakes.FakeBranchCheckerImpl) {},
			shouldErr:    true,
			shouldReturn: false,
		},
		{ // failure LSRemoteExec errors
			branch:       "release-1.19",
			releaseType:  release.ReleaseTypeAlpha,
			buildVersion: semver.MustParse("1.19.0-alpha.2"),
			prepare: func(mock *releasefakes.FakeBranchCheckerImpl) {
				mock.LSRemoteExecReturns("", errors.New(""))
			},
			shouldErr:    true,
			shouldReturn: false,
		},
	} {
		mock := &releasefakes.FakeBranchCheckerImpl{}
		sut := release.NewBranchChecker()
		tc.prepare(mock)
		sut.SetImpl(mock)

		res, err := sut.NeedsCreation(tc.branch, tc.releaseType, tc.buildVersion)
		if tc.shouldErr {
			require.NotNil(t, err)
			require.False(t, res)
		} else {
			require.Nil(t, err)
			require.Equal(t, tc.shouldReturn, res)
		}
	}
}

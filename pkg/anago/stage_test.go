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

// nolint: dupl
package anago_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/anago"
	"k8s.io/release/pkg/anago/anagofakes"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

func generateTestingStageState(params *testStateParameters) *anago.StageState {
	state := anago.DefaultStageState()
	if params.versionsTag != nil {
		state.SetVersions(release.NewReleaseVersions("", *params.versionsTag, "", "", ""))
	}

	if params.createReleaseBranch != nil {
		state.SetCreateReleaseBranch(*params.createReleaseBranch)
	}
	return state
}

func TestInitLogFileStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // ToFile fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.ToFileReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)

		sut.SetState(
			generateTestingStageState(&testStateParameters{versionsTag: &testVersionTag}),
		)

		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.InitLogFile()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestCheckPrerequisitesStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // CheckPrerequisites fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.CheckPrerequisitesReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)

		sut.SetState(
			generateTestingStageState(&testStateParameters{versionsTag: &testVersionTag}),
		)

		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.CheckPrerequisites()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestCheckReleaseBranchStateStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // BranchNeedsCreation fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.BranchNeedsCreationReturns(false, err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)

		sut.SetState(
			generateTestingStageState(&testStateParameters{versionsTag: &testVersionTag}),
		)

		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.CheckReleaseBranchState()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestGenerateReleaseVersionStage(t *testing.T) {
	for _, tc := range []struct {
		prepare             func(*anagofakes.FakeStageImpl)
		createReleaseBranch bool
		shouldError         bool
	}{
		{ // success
			createReleaseBranch: true,
			prepare:             func(*anagofakes.FakeStageImpl) {},
			shouldError:         false,
		},
		{ // PrepareWorkspaceStage fails
			createReleaseBranch: true,
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.GenerateReleaseVersionReturns(nil, err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)

		sut.SetState(
			generateTestingStageState(&testStateParameters{
				createReleaseBranch: &tc.createReleaseBranch,
			}),
		)

		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)
		err := sut.GenerateReleaseVersion()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestPrepareWorkspaceStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // PrepareWorkspaceStage fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.PrepareWorkspaceStageReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)

		sut.SetState(
			generateTestingStageState(&testStateParameters{}),
		)

		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)
		err := sut.PrepareWorkspace()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestTagRepository(t *testing.T) {
	newRCVersions := release.NewReleaseVersions(
		"v1.20.0-rc.0", "", "v1.20.0-rc.0", "", "v1.21.0-alpha.0",
	)
	newBetaVersions := release.NewReleaseVersions(
		"v1.20.0-beta.1", "", "", "v1.20.0-beta.1", "",
	)
	for _, tc := range []struct {
		prepare             func(*anagofakes.FakeStageImpl)
		versions            *release.Versions
		releaseBranch       string
		createReleaseBranch bool
		shouldError         bool
	}{
		{ // success new rc creating release branch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
				mock.CurrentBranchReturnsOnCall(0, "release-1.20", nil)
			},
			versions:            newRCVersions,
			releaseBranch:       "release-1.20",
			createReleaseBranch: true,
			shouldError:         false,
		},
		{ // failure on CommitEmpty new rc creating release branch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
				mock.CurrentBranchReturnsOnCall(0, "release-1.20", nil)
				mock.CommitEmptyReturns(err)
			},
			versions:            newRCVersions,
			releaseBranch:       "release-1.20",
			createReleaseBranch: true,
			shouldError:         true,
		},
		{ // failure on CurrentBranch new rc creating release branch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
				mock.CurrentBranchReturnsOnCall(0, "release-1.20", nil)
				mock.CurrentBranchReturns("", err)
			},
			versions:            newRCVersions,
			releaseBranch:       "release-1.20",
			createReleaseBranch: true,
			shouldError:         true,
		},
		{ // failure on Checkout new rc creating release branch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
				mock.CurrentBranchReturnsOnCall(0, "release-1.20", nil)
				mock.CheckoutReturns(err)
			},
			versions:            newRCVersions,
			releaseBranch:       "release-1.20",
			createReleaseBranch: true,
			shouldError:         true,
		},
		{ // failure on RevParse new rc creating release branch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", nil)
			},
			versions:            newRCVersions,
			releaseBranch:       "release-1.20",
			createReleaseBranch: true,
			shouldError:         true,
		},
		{ // failure on OpenRepo new rc creating release branch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.OpenRepoReturns(nil, err)
			},
			versions:            newRCVersions,
			releaseBranch:       "release-1.20",
			createReleaseBranch: true,
			shouldError:         true,
		},
		{ // success new rc checking out release branch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
			},
			versions:            newRCVersions,
			releaseBranch:       "release-1.20",
			createReleaseBranch: true,
			shouldError:         false,
		},
		{ // failure on Checkout new rc checking out release branch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
				mock.CheckoutReturns(err)
			},
			versions:            newRCVersions,
			releaseBranch:       "release-1.20",
			createReleaseBranch: true,
			shouldError:         true,
		},
		{ // failure on RevParse new rc checking out release branch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", nil)
			},
			versions:            newRCVersions,
			releaseBranch:       "release-1.20",
			createReleaseBranch: true,
			shouldError:         true,
		},
		{ // success new beta
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
			},
			versions:      newBetaVersions,
			releaseBranch: git.DefaultBranch,
			shouldError:   false,
		},
		{ // new beta failure on Tag
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
				mock.TagReturns(err)
			},
			versions:      newBetaVersions,
			releaseBranch: git.DefaultBranch,
			shouldError:   true,
		},
		{ // new beta failure on CurrentBranch
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
				mock.CurrentBranchReturns("", err)
			},
			versions:      newBetaVersions,
			releaseBranch: git.DefaultBranch,
			shouldError:   true,
		},
		{ // new beta failure on Checkout
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.RevParseTagReturns("", err)
				mock.CheckoutReturns(err)
			},
			versions:      newBetaVersions,
			releaseBranch: git.DefaultBranch,
			shouldError:   true,
		},
	} {
		opts := anago.DefaultStageOptions()
		opts.BuildVersion = "v1.20.0-beta.1.358+4628c605aadb9b"
		opts.ReleaseBranch = tc.releaseBranch
		state := anago.DefaultState()
		err := opts.Validate(state)
		require.Nil(t, err)

		sut := anago.NewDefaultStage(opts)

		state.SetVersions(tc.versions)
		state.SetCreateReleaseBranch(tc.createReleaseBranch)
		sut.SetState(&anago.StageState{state})

		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err = sut.TagRepository()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestBuild(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // MakeCross fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.MakeCrossReturns(err)
			},
			shouldError: true,
		},
		{ // DockerHubLogin fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.DockerHubLoginReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)

		sut.SetState(
			generateTestingStageState(&testStateParameters{versionsTag: &testVersionTag}),
		)

		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.Build()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestGenerateChangelog(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // GenerateChangelog fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.GenerateChangelogReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)

		etag := ""
		sut.SetState(generateTestingStageState(&testStateParameters{
			versionsTag: &etag,
		}))

		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.GenerateChangelog()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestStageArtifacts(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // CheckReleaseBucket fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.CheckReleaseBucketReturns(err)
			},
			shouldError: true,
		},
		{ // StageLocalSourceTree fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.StageLocalSourceTreeReturns(err)
			},
			shouldError: true,
		},
		{ // StageLocalArtifacts fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.StageLocalArtifactsReturns(err)
			},
			shouldError: true,
		},
		{ // PushReleaseArtifacts fails on first
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.PushReleaseArtifactsReturns(err)
			},
			shouldError: true,
		},
		{ // PushReleaseArtifacts fails on second
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.PushReleaseArtifactsReturnsOnCall(1, err)
			},
			shouldError: true,
		},
		{ // PushContainerImages fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.PushContainerImagesReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)
		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		sut.SetState(
			generateTestingStageState(
				&testStateParameters{versionsTag: &testVersionTag},
			),
		)

		err := sut.StageArtifacts()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestSubmitStageImpl(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // Submit fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.SubmitReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)
		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)
		err := sut.Submit(false)
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

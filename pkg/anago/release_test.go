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
	"k8s.io/release/pkg/release"
)

func generateTestingReleaseState(params *testStateParameters) *anago.ReleaseState {
	state := anago.DefaultReleaseState()
	if params.versionsTag != nil {
		state.SetVersions(release.NewReleaseVersions(*params.versionsTag, *params.versionsTag, "", "", ""))
	}

	if params.createReleaseBranch != nil {
		state.SetCreateReleaseBranch(*params.createReleaseBranch)
	}
	return state
}

func TestInitLogFileRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // ToFile fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.ToFileReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()
		sut := anago.NewDefaultRelease(opts)

		sut.SetState(
			generateTestingReleaseState(&testStateParameters{versionsTag: &testVersionTag}),
		)

		mock := &anagofakes.FakeReleaseImpl{}
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

func TestCheckPrerequisitesRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // CheckPrerequisites fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.CheckPrerequisitesReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()
		sut := anago.NewDefaultRelease(opts)

		sut.SetState(
			generateTestingReleaseState(&testStateParameters{versionsTag: &testVersionTag}),
		)

		mock := &anagofakes.FakeReleaseImpl{}
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

func TestCheckReleaseBranchStateRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // BranchNeedsCreation fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.BranchNeedsCreationReturns(false, err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()
		sut := anago.NewDefaultRelease(opts)

		sut.SetState(
			generateTestingReleaseState(&testStateParameters{versionsTag: &testVersionTag}),
		)

		mock := &anagofakes.FakeReleaseImpl{}
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

func TestGenerateReleaseVersionRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare             func(*anagofakes.FakeReleaseImpl)
		createReleaseBranch bool
		shouldError         bool
	}{
		{ // success
			createReleaseBranch: true,
			prepare:             func(*anagofakes.FakeReleaseImpl) {},
			shouldError:         false,
		},
		{ // PrepareWorkspaceRelease fails
			createReleaseBranch: true,
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.GenerateReleaseVersionReturns(nil, err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()

		sut := anago.NewDefaultRelease(opts)
		sut.SetState(
			generateTestingReleaseState(&testStateParameters{
				createReleaseBranch: &tc.createReleaseBranch,
			}),
		)

		mock := &anagofakes.FakeReleaseImpl{}
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

func TestPushArtifacts(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // CheckReleaseBucket fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.CheckReleaseBucketReturns(err)
			},
			shouldError: true,
		},
		{ // CopyStagedFromGCSReturns fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.CopyStagedFromGCSReturns(err)
			},
			shouldError: true,
		},
		{ // ValidateImages fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.ValidateImagesReturns(err)
			},
			shouldError: true,
		},
		{ // PusblishVersion fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.PublishVersionReturns(err)
			},
			shouldError: true,
		},
		{ // NormalizePath fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.NormalizePathReturns("", err)
			},
			shouldError: true,
		},
		{ // CopyToRemote fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.CopyToRemoteReturns(err)
			},
			shouldError: true,
		},
		{ // PublishReleaseNotesIndex fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.PublishReleaseNotesIndexReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()

		sut := anago.NewDefaultRelease(opts)

		sut.SetState(
			generateTestingReleaseState(&testStateParameters{versionsTag: &testVersionTag}),
		)

		mock := &anagofakes.FakeReleaseImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.PushArtifacts()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestPrepareWorkspaceRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // PrepareWorkspaceRelease fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.PrepareWorkspaceReleaseReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()
		sut := anago.NewDefaultRelease(opts)
		mock := &anagofakes.FakeReleaseImpl{}
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

func TestSubmitReleaseImpl(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // Submit fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.SubmitReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()
		sut := anago.NewDefaultRelease(opts)
		mock := &anagofakes.FakeReleaseImpl{}
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

func TestCreateAnnouncement(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // Create announcement fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.CreateAnnouncementReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()
		sut := anago.NewDefaultRelease(opts)
		sut.SetState(
			generateTestingReleaseState(&testStateParameters{versionsTag: &testVersionTag}),
		)
		mock := &anagofakes.FakeReleaseImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)
		err := sut.CreateAnnouncement()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestPushGitObjects(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // Pushing list of branches fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.PushBranchesReturns(err)
			},
			shouldError: true,
		},
		{ // Pushing list of tags fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.PushTagsReturns(err)
			},
			shouldError: true,
		},
		{ // Pushing the main branch fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.PushMainBranchReturns(err)
			},
			shouldError: true,
		},
		{ // Creating the git pusher object fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.NewGitPusherReturns(nil, err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()
		sut := anago.NewDefaultRelease(opts)
		sut.SetState(
			generateTestingReleaseState(&testStateParameters{versionsTag: &testVersionTag}),
		)
		mock := &anagofakes.FakeReleaseImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)
		err := sut.PushGitObjects()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestUpdateGitHubPage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // Pushing list of branches fails
			prepare: func(mock *anagofakes.FakeReleaseImpl) {
				mock.UpdateGitHubPageReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()
		sut := anago.NewDefaultRelease(opts)
		sut.SetState(
			generateTestingReleaseState(&testStateParameters{versionsTag: &testVersionTag}),
		)
		mock := &anagofakes.FakeReleaseImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)
		err := sut.UpdateGitHubPage()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

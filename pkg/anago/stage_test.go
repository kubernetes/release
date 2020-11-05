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
)

func TestCheckPrerequisitesStage(t *testing.T) {
	opts := anago.DefaultStageOptions()
	sut := anago.NewDefaultStage(opts)
	mock := &anagofakes.FakeStageImpl{}
	sut.SetClient(mock)
	require.Nil(t, sut.CheckPrerequisites())
}

func TestGenerateReleaseVersionStage(t *testing.T) {
	for _, tc := range []struct {
		parentBranch string
		prepare      func(*anagofakes.FakeStageImpl)
		shouldError  bool
	}{
		{ // success
			parentBranch: git.DefaultBranch,
			prepare:      func(*anagofakes.FakeStageImpl) {},
			shouldError:  false,
		},
		{ // PrepareWorkspaceStage fails
			parentBranch: git.DefaultBranch,
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.GenerateReleaseVersionReturns(nil, err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)
		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetClient(mock)
		_, err := sut.GenerateReleaseVersion(tc.parentBranch)
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
		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetClient(mock)
		err := sut.PrepareWorkspace()
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
		sut.SetClient(mock)
		err := sut.StageArtifacts([]string{"v1.20.0"})
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

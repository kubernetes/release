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

package anago_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/anago"
	"k8s.io/release/pkg/anago/anagofakes"
)

var err = errors.New("error")

func TestStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageClient)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageClient) {},
			shouldError: false,
		},
		{ // CheckPrerequisites fails
			prepare: func(mock *anagofakes.FakeStageClient) {
				mock.CheckPrerequisitesReturns(err)
			},
			shouldError: true,
		},
		{ // SetBuildCandidate fails
			prepare: func(mock *anagofakes.FakeStageClient) {
				mock.SetBuildCandidateReturns(err)
			},
			shouldError: true,
		},
		{ // PrepareWorkspace fails
			prepare: func(mock *anagofakes.FakeStageClient) {
				mock.PrepareWorkspaceReturns(err)
			},
			shouldError: true,
		},
		{ // Build fails
			prepare: func(mock *anagofakes.FakeStageClient) {
				mock.BuildReturns(err)
			},
			shouldError: true,
		},
		{ // GenerateReleaseNotes fails
			prepare: func(mock *anagofakes.FakeStageClient) {
				mock.GenerateReleaseNotesReturns(err)
			},
			shouldError: true,
		},
		{ // StageArtifacts fails
			prepare: func(mock *anagofakes.FakeStageClient) {
				mock.StageArtifactsReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewStage(opts)
		mock := &anagofakes.FakeStageClient{}
		tc.prepare(mock)
		sut.SetClient(mock)

		if tc.shouldError {
			require.NotNil(t, sut.Run())
		} else {
			require.Nil(t, sut.Run())
		}
	}
}

func TestRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeReleaseClient)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeReleaseClient) {},
			shouldError: false,
		},
		{ // CheckPrerequisites fails
			prepare: func(mock *anagofakes.FakeReleaseClient) {
				mock.CheckPrerequisitesReturns(err)
			},
			shouldError: true,
		},
		{ // SetBuildCandidate fails
			prepare: func(mock *anagofakes.FakeReleaseClient) {
				mock.SetBuildCandidateReturns(err)
			},
			shouldError: true,
		},
		{ // PrepareWorkspace fails
			prepare: func(mock *anagofakes.FakeReleaseClient) {
				mock.PrepareWorkspaceReturns(err)
			},
			shouldError: true,
		},
		{ // PushArtifacts fails
			prepare: func(mock *anagofakes.FakeReleaseClient) {
				mock.PushArtifactsReturns(err)
			},
			shouldError: true,
		},
		{ // PushGitObjects fails
			prepare: func(mock *anagofakes.FakeReleaseClient) {
				mock.PushGitObjectsReturns(err)
			},
			shouldError: true,
		},
		{ // CreateAnnouncement fails
			prepare: func(mock *anagofakes.FakeReleaseClient) {
				mock.CreateAnnouncementReturns(err)
			},
			shouldError: true,
		},
		{ // Archive fails
			prepare: func(mock *anagofakes.FakeReleaseClient) {
				mock.ArchiveReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultReleaseOptions()
		sut := anago.NewRelease(opts)
		mock := &anagofakes.FakeReleaseClient{}
		tc.prepare(mock)
		sut.SetClient(mock)

		if tc.shouldError {
			require.NotNil(t, sut.Run())
		} else {
			require.Nil(t, sut.Run())
		}
	}
}

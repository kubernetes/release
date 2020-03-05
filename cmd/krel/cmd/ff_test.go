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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/git"
)

func (s *sut) getFfOptions() *ffOptions {
	return &ffOptions{
		masterRef:      git.Remotify(git.Master),
		nonInteractive: true,
	}
}

func TestFfFailedWithoutReleaseBranch(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	ff := &ffOptions{}

	// When
	err := runFf(ff, s.getRootOptions())

	// Then
	require.NotNil(t, err)
}

func TestFfFailedNoReleaseBranch(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	ffo := s.getFfOptions()
	ffo.branch = "not-a-release-branch"

	// When
	err := runFf(ffo, s.getRootOptions())

	// Then
	require.NotNil(t, err)
}

func TestFfFailedReleaseBranchDoesNotExist(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	ffo := s.getFfOptions()
	ffo.branch = "release-1.999"

	// When
	err := runFf(ffo, s.getRootOptions())

	// Then
	require.NotNil(t, err)
}

func TestFfFailedOldReleaseBranch(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	ffo := s.getFfOptions()
	ffo.branch = "release-1.17"

	// When
	err := runFf(ffo, s.getRootOptions())

	// Then
	require.NotNil(t, err)
}

func TestFfSuccessDryRun(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	ffo := s.getFfOptions()
	ffo.branch = pseudoReleaseBranch

	// When
	err := runFf(ffo, s.getRootOptions())

	// Then
	require.Nil(t, err)

	// Local should contain the commit
	lastLocalCommit := s.lastCommit(t, pseudoReleaseBranch)
	require.Contains(t, lastLocalCommit, testCommitMessage)

	// Remote should not be modified
	lastRemoteCommit := s.lastCommit(t, git.Remotify(pseudoReleaseBranch))
	require.NotContains(t, lastRemoteCommit, testCommitMessage)
}

func TestFfSuccess(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	ffo := s.getFfOptions()
	ffo.branch = pseudoReleaseBranch

	ro := s.getRootOptions()
	ro.nomock = true

	// When
	err := runFf(ffo, ro)

	// Then
	require.Nil(t, err)

	// Local should contain the commit
	lastLocalCommit := s.lastCommit(t, pseudoReleaseBranch)
	require.Contains(t, lastLocalCommit, testCommitMessage)

	// Remote should be modified
	lastRemoteCommit := s.lastCommit(t, git.Remotify(pseudoReleaseBranch))
	require.Contains(t, lastRemoteCommit, testCommitMessage)
}

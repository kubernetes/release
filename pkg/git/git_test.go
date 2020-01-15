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

package git_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/git/gitfakes"
)

func newSUT() (*git.Repo, *gitfakes.FakeRepository, *gitfakes.FakeWorktree) {
	repoMock := &gitfakes.FakeRepository{}
	worktreeMock := &gitfakes.FakeWorktree{}

	repo := &git.Repo{}
	repo.SetWorktree(worktreeMock)
	repo.SetInnerRepo(repoMock)

	return repo, repoMock, worktreeMock
}

func TestCommit(t *testing.T) {
	repo, _, worktreeMock := newSUT()
	require.Nil(t, repo.Commit("msg"))
	require.Equal(t, worktreeMock.CommitCallCount(), 1)
}

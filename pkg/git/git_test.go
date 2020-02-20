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

	"github.com/stretchr/testify/assert"
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

func TestGetDefaultKubernetesRepoURLSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		org      string
		useSSH   bool
		expected string
	}{
		{
			name:     "default HTTPS",
			expected: "https://github.com/kubernetes/kubernetes",
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		actual, err := git.GetDefaultKubernetesRepoURL()
		assert.Equal(t, tc.expected, actual)
		assert.Nil(t, err)
	}
}

func TestGetKubernetesRepoURLSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		org      string
		useSSH   bool
		expected string
	}{
		{
			name:     "default HTTPS",
			expected: "https://github.com/kubernetes/kubernetes",
		},
		{
			name:     "ssh with custom org",
			org:      "fake-org",
			useSSH:   true,
			expected: "git@github.com:fake-org/kubernetes",
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		actual, err := git.GetKubernetesRepoURL(tc.org, tc.useSSH)
		assert.Equal(t, tc.expected, actual)
		assert.Nil(t, err)
	}
}

func TestGetRepoURLSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		org      string
		repo     string
		useSSH   bool
		expected string
	}{
		{
			name:     "default Kubernetes HTTPS",
			org:      "kubernetes",
			repo:     "kubernetes",
			expected: "https://github.com/kubernetes/kubernetes",
		},
		{
			name:     "ssh with custom org",
			org:      "fake-org",
			repo:     "repofoo",
			useSSH:   true,
			expected: "git@github.com:fake-org/repofoo",
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		actual, err := git.GetRepoURL(tc.org, tc.repo, tc.useSSH)
		assert.Equal(t, tc.expected, actual)
		assert.Nil(t, err)
	}
}

func TestRemotify(t *testing.T) {
	testcases := []struct{ provided, expected string }{
		{provided: git.Master, expected: git.DefaultRemote + "/" + git.Master},
		{provided: "origin/ref", expected: "origin/ref"},
		{provided: "base/another_ref", expected: "base/another_ref"},
	}

	for _, tc := range testcases {
		require.Equal(t, git.Remotify(tc.provided), tc.expected)
	}
}

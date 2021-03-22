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
	"os"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
)

type sut struct {
	repo *release.Repo
	mock *releasefakes.FakeRepository
	dir  string
	t    *testing.T
}

func newSUT(t *testing.T) *sut {
	dir, err := os.MkdirTemp("", "k8s-test-")
	require.Nil(t, err)

	_, err = gogit.PlainInit(dir, false)
	require.Nil(t, err)
	require.Nil(t, os.Chdir(dir))

	repo := release.NewRepo()
	err = repo.Open()
	require.Nil(t, err)
	require.NotNil(t, repo)

	mock := &releasefakes.FakeRepository{}
	repo.SetRepo(mock)
	return &sut{repo, mock, dir, t}
}

func (s *sut) cleanup() {
	require.Nil(s.t, os.RemoveAll(s.dir))
}

func TestGetTagSuccess(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.DescribeReturns("v1.0.0", nil)

	// When
	tag, err := sut.repo.GetTag()

	// Then
	require.Nil(t, err)
	require.Contains(t, tag, "v1.0.0")
}

func TestGetTagFailure(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.DescribeReturns("", errors.New(""))

	// When
	tag, err := sut.repo.GetTag()

	// Then
	require.NotNil(t, err)
	require.Empty(t, tag)
}

func TestCheckStateSuccess(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.CurrentBranchReturns("branch", nil)
	sut.mock.RemotesReturns([]*git.Remote{
		git.NewRemote("origin", []string{"github.com:org/repo"}),
	}, nil)
	sut.mock.HeadReturns("dbade8e", nil)
	sut.mock.RevParseReturns("dbade8e", nil)
	sut.mock.LsRemoteReturns("dbade8e refs/heads/master", nil)

	// When
	err := sut.repo.CheckState("org", "repo", "branch", false)

	// Then
	require.Nil(t, err)
}

func TestCheckStateFailedNoRemoteFound(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.CurrentBranchReturns("branch", nil)
	sut.mock.RemotesReturns([]*git.Remote{
		git.NewRemote("origin", []string{"some-other-url"}),
	}, nil)

	// When
	err := sut.repo.CheckState("org", "repo", "branch", false)

	// Then
	require.NotNil(t, err)
}

func TestCheckStateFailedRemoteFailed(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.CurrentBranchReturns("branch", nil)
	sut.mock.RemotesReturns(nil, errors.New(""))

	// When
	err := sut.repo.CheckState("org", "repo", "branch", false)

	// Then
	require.NotNil(t, err)
}

func TestCheckStateFailedWrongBranch(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.CurrentBranchReturns("wrong", nil)

	// When
	err := sut.repo.CheckState("org", "repo", "branch", false)

	// Then
	require.NotNil(t, err)
}

func TestCheckStateFailedBranchFailed(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.CurrentBranchReturns("", errors.New(""))

	// When
	err := sut.repo.CheckState("org", "repo", "branch", false)

	// Then
	require.NotNil(t, err)
}

func TestCheckStateFailedLsRemote(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.CurrentBranchReturns("branch", nil)
	sut.mock.RemotesReturns([]*git.Remote{
		git.NewRemote("origin", []string{"github.com:org/repo"}),
	}, nil)
	sut.mock.LsRemoteReturns("", errors.New(""))

	// When
	err := sut.repo.CheckState("org", "repo", "branch", false)

	// Then
	require.NotNil(t, err)
}

func TestCheckStateFailedBranchHeadRetrievalFails(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.CurrentBranchReturns("branch", nil)
	sut.mock.RemotesReturns([]*git.Remote{
		git.NewRemote("origin", []string{"github.com:org/repo"}),
	}, nil)
	sut.mock.LsRemoteReturns("dbade8e refs/heads/master", nil)
	sut.mock.HeadReturns("", errors.New("no such commit"))

	// When
	err := sut.repo.CheckState("org", "repo", "branch", false)

	// Then
	require.NotNil(t, err)
}

func TestCheckStateFailedBranchHeadRetrievalNotEqual(t *testing.T) {
	// Given
	sut := newSUT(t)
	defer sut.cleanup()
	sut.mock.CurrentBranchReturns("branch", nil)
	sut.mock.RemotesReturns([]*git.Remote{
		git.NewRemote("origin", []string{"github.com:org/repo"}),
	}, nil)
	sut.mock.LsRemoteReturns("321 refs/heads/master", nil)
	sut.mock.HeadReturns("123", nil)

	// When
	err := sut.repo.CheckState("org", "repo", "branch", false)

	// Then
	require.NotNil(t, err)
}

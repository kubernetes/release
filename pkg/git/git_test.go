/*
Copyright 2019 The Kubernetes Authors.

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

package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/release/pkg/command"

	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type testRepo struct {
	sut                *Repo
	dir                string
	firstCommit        string
	firstBranchCommit  string
	secondBranchCommit string
	branchName         string
	firstTagCommit     string
	firstTagName       string
	secondTagCommit    string
	secondTagName      string
	thirdTagCommit     string
	thirdTagName       string
}

// newTestRepo creates a test repo with the following structure:
//
// * commit `secondBranchCommit` (tag: `thirdTagName`, HEAD -> `branchName`, origin/`branchName`)
// | Author: John Doe <john@doe.org>
// |
// |     Third commit
// |
// * commit `firstBranchCommit` (tag: `secondTagName`, HEAD -> `branchName`, origin/`branchName`)
// | Author: John Doe <john@doe.org>
// |
// |     Second commit
// |
// * commit `firstCommit` (tag: `firstTagName`, origin/master, origin/HEAD, master)
//   Author: John Doe <john@doe.org>
//
//       First commit
//
func newTestRepo(t *testing.T) *testRepo {
	// Setup the bare repo as base
	bareTempDir, err := ioutil.TempDir("", "k8s-test-bare-")
	require.Nil(t, err)

	bareRepo, err := git.PlainInit(bareTempDir, true)
	require.Nil(t, err)
	require.NotNil(t, bareRepo)

	// Clone from the bare to be able to add our test data
	cloneTempDir, err := ioutil.TempDir("", "k8s-test-clone-")
	require.Nil(t, err)
	cloneRepo, err := git.PlainInit(cloneTempDir, false)
	require.Nil(t, err)

	// Add the test data set
	const testFileName = "test-file"
	require.Nil(t, ioutil.WriteFile(
		filepath.Join(cloneTempDir, testFileName),
		[]byte("test-content"),
		0644,
	))

	worktree, err := cloneRepo.Worktree()
	require.Nil(t, err)
	_, err = worktree.Add(testFileName)
	require.Nil(t, err)

	author := &object.Signature{
		Name:  "John Doe",
		Email: "john@doe.org",
		When:  time.Now(),
	}
	firstCommit, err := worktree.Commit("First commit", &git.CommitOptions{
		Author: author,
	})
	require.Nil(t, err)

	firstTagName := "v0.1.0"
	firstTagRef, err := cloneRepo.CreateTag(firstTagName, firstCommit,
		&git.CreateTagOptions{
			Tagger:  author,
			Message: firstTagName,
		},
	)
	require.Nil(t, err)

	// Create a test branch and a test commit on top
	branchName := "first-branch"
	require.Nil(t, command.NewWithWorkDir(
		cloneTempDir, "git", "checkout", "-b", branchName,
	).RunSuccess())

	const branchTestFileName = "branch-test-file"
	require.Nil(t, ioutil.WriteFile(
		filepath.Join(cloneTempDir, branchTestFileName),
		[]byte("test-content"),
		0644,
	))
	_, err = worktree.Add(branchTestFileName)
	require.Nil(t, err)

	firstBranchCommit, err := worktree.Commit("Second commit", &git.CommitOptions{
		Author: author,
		All:    true,
	})
	require.Nil(t, err)

	secondTagName := "v0.1.1"
	secondTagRef, err := cloneRepo.CreateTag(secondTagName, firstBranchCommit,
		&git.CreateTagOptions{
			Tagger:  author,
			Message: firstTagName,
		},
	)
	require.Nil(t, err)

	const secondBranchTestFileName = "branch-test-file-2"
	require.Nil(t, ioutil.WriteFile(
		filepath.Join(cloneTempDir, secondBranchTestFileName),
		[]byte("test-content"),
		0644,
	))
	_, err = worktree.Add(secondBranchTestFileName)
	require.Nil(t, err)

	secondBranchCommit, err := worktree.Commit("Third commit", &git.CommitOptions{
		Author: author,
		All:    true,
	})
	require.Nil(t, err)

	thirdTagName := "v0.1.2"
	thirdTagRef, err := cloneRepo.CreateTag(thirdTagName, secondBranchCommit,
		&git.CreateTagOptions{
			Tagger:  author,
			Message: firstTagName,
		},
	)
	require.Nil(t, err)

	// Push the test content into the bare repo
	_, err = cloneRepo.CreateRemote(&config.RemoteConfig{
		Name: DefaultRemote,
		URLs: []string{bareTempDir},
	})
	require.Nil(t, err)
	require.Nil(t, cloneRepo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/*:refs/*"},
	}))

	require.Nil(t, os.RemoveAll(cloneTempDir))

	// Provide a system under test inside the test repo
	sut, err := CloneOrOpenRepo("", bareTempDir, false)
	require.Nil(t, err)
	require.Nil(t, command.NewWithWorkDir(
		sut.Dir(), "git", "checkout", branchName,
	).RunSuccess())

	return &testRepo{
		sut:                sut,
		dir:                bareTempDir,
		firstCommit:        firstCommit.String(),
		firstBranchCommit:  firstBranchCommit.String(),
		secondBranchCommit: secondBranchCommit.String(),
		branchName:         branchName,
		firstTagName:       firstTagName,
		firstTagCommit:     firstTagRef.Hash().String(),
		secondTagName:      secondTagName,
		secondTagCommit:    secondTagRef.Hash().String(),
		thirdTagName:       thirdTagName,
		thirdTagCommit:     thirdTagRef.Hash().String(),
	}
}

func (r *testRepo) cleanup(t *testing.T) {
	require.Nil(t, os.RemoveAll(r.dir))
	require.Nil(t, os.RemoveAll(r.sut.Dir()))
}

func TestSuccessCloneOrOpen(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	secondRepo, err := CloneOrOpenRepo(testRepo.sut.Dir(), testRepo.dir, false)
	require.Nil(t, err)

	require.Equal(t, testRepo.sut.Dir(), secondRepo.Dir())
	require.Nil(t, secondRepo.Cleanup())
}

func TestSuccessDescribeTag(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	tag, err := testRepo.sut.DescribeTag(testRepo.firstTagCommit)
	require.Nil(t, err)
	require.Equal(t, tag, testRepo.firstTagName)
}

func TestFailureDescribeTag(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	_, err := testRepo.sut.DescribeTag("wrong")
	require.NotNil(t, err)
}

func TestSuccessHasRemoteBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.Nil(t, testRepo.sut.HasRemoteBranch(testRepo.branchName))
	require.Nil(t, testRepo.sut.HasRemoteBranch(Master))
}

func TestFailureHasRemoteBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.HasRemoteBranch("wrong")
	require.NotNil(t, err)
}

func TestSuccessHead(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	head, err := testRepo.sut.Head()
	require.Nil(t, err)
	require.Equal(t, head, testRepo.secondBranchCommit)
}

func TestSuccessMerge(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Merge(Master)
	require.Nil(t, err)
}

func TestFailureMerge(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Merge("wrong")
	require.NotNil(t, err)
}

func TestSuccessMergeBase(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	mergeBase, err := testRepo.sut.MergeBase(Master, testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, mergeBase, testRepo.firstCommit)
}

func TestSuccessRevParse(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	masterRev, err := testRepo.sut.RevParse(Master)
	require.Nil(t, err)
	require.Equal(t, masterRev, testRepo.firstCommit)

	branchRev, err := testRepo.sut.RevParse(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, branchRev, testRepo.secondBranchCommit)

	tagRev, err := testRepo.sut.RevParse(testRepo.firstTagName)
	require.Nil(t, err)
	require.Equal(t, tagRev, testRepo.firstCommit)
}

func TestFailureRevParse(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	_, err := testRepo.sut.RevParse("wrong")
	require.NotNil(t, err)
}

func TestSuccessRevParseShort(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	masterRev, err := testRepo.sut.RevParseShort(Master)
	require.Nil(t, err)
	require.Equal(t, masterRev, testRepo.firstCommit[:10])

	branchRev, err := testRepo.sut.RevParseShort(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, branchRev, testRepo.secondBranchCommit[:10])

	tagRev, err := testRepo.sut.RevParseShort(testRepo.firstTagName)
	require.Nil(t, err)
	require.Equal(t, tagRev, testRepo.firstCommit[:10])
}

func TestFailureRevParseShort(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	_, err := testRepo.sut.RevParseShort("wrong")
	require.NotNil(t, err)
}

func TestSuccessPush(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Push(Master)
	require.Nil(t, err)
}

func TestFailurePush(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Push("wrong")
	require.NotNil(t, err)
}

func TestSuccessRemotify(t *testing.T) {
	newRemote := Remotify(Master)
	require.Equal(t, newRemote, DefaultRemote+"/"+Master)
}

func TestSuccessIsReleaseBranch(t *testing.T) {
	require.True(t, IsReleaseBranch("release-1.17"))
}

func TestFailureIsReleaseBranch(t *testing.T) {
	require.False(t, IsReleaseBranch("wrong-branch"))
}

func TestSuccessLatestTagForBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	version, err := testRepo.sut.latestTagForBranch(Master)
	require.Nil(t, err)
	require.Equal(t, addTagPrefix(version.String()), testRepo.firstTagName)
}

func TestFailureLatestTagForBranchInvalidBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	version, err := testRepo.sut.latestTagForBranch("wrong-branch")
	require.NotNil(t, err)
	require.Nil(t, version)
}

func TestSuccessLatestPatchToPatch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.LatestPatchToPatch(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, result.StartSHA(), testRepo.firstBranchCommit)
	require.Equal(t, result.StartRev(), testRepo.secondTagName)
	require.Equal(t, result.EndSHA(), testRepo.secondBranchCommit)
	require.Equal(t, result.EndRev(), testRepo.thirdTagName)
}

func TestFailureLatestPatchToPatchWrongBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.LatestPatchToPatch("wrong-branch")
	require.NotNil(t, err)
	require.Equal(t, DiscoverResult{}, result)
}

func TestSuccessDry(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	testRepo.sut.SetDry()

	err := testRepo.sut.Push(Master)
	require.Nil(t, err)
}

func TestSuccessLatestNonPatchFinalToLatest(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.LatestNonPatchFinalToLatest()
	require.Nil(t, err)
	require.Equal(t, result.StartSHA(), testRepo.firstCommit)
	require.Equal(t, result.StartRev(), testRepo.firstTagName)
	require.Equal(t, result.EndSHA(), testRepo.firstCommit)
	require.Equal(t, result.EndRev(), Master)
}

func TestFailureLatestNonPatchFinalToLatestNoLatestTag(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.Nil(t, command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "tag", "-d", testRepo.firstTagName,
	).RunSuccess())

	result, err := testRepo.sut.LatestNonPatchFinalToLatest()
	require.NotNil(t, err)
	require.Equal(t, DiscoverResult{}, result)
}

func TestSuccessLatestNonPatchFinalToMinor(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	nextMinorTag := "v0.2.0"
	require.Nil(t, command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "tag", nextMinorTag,
	).RunSuccess())

	result, err := testRepo.sut.LatestNonPatchFinalToMinor()
	require.Nil(t, err)
	require.Equal(t, result.StartSHA(), testRepo.firstCommit)
	require.Equal(t, result.StartRev(), testRepo.firstTagName)
	require.Equal(t, result.EndRev(), nextMinorTag)
}

func TestFailureLatestNonPatchFinalToMinor(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.LatestNonPatchFinalToMinor()
	require.NotNil(t, err)
	require.Equal(t, DiscoverResult{}, result)
}

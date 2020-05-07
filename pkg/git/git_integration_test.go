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

package git_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/stretchr/testify/require"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/util"
)

type testRepo struct {
	sut                *git.Repo
	dir                string
	firstCommit        string
	firstBranchCommit  string
	secondBranchCommit string
	thirdBranchCommit  string
	branchName         string
	firstTagCommit     string
	firstTagName       string
	secondTagCommit    string
	secondTagName      string
	thirdTagCommit     string
	thirdTagName       string
	testFileName       string
}

// newTestRepo creates a test repo with the following structure:
//
// * commit `thirdBranchCommit` (HEAD -> `branchName`, origin/`branchName`)
// | Author: John Doe <john@doe.org>
// |
// |     Fourth commit
// |
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

	bareRepo, err := gogit.PlainInit(bareTempDir, true)
	require.Nil(t, err)
	require.NotNil(t, bareRepo)

	// Clone from the bare to be able to add our test data
	cloneTempDir, err := ioutil.TempDir("", "k8s-test-clone-")
	require.Nil(t, err)
	cloneRepo, err := gogit.PlainInit(cloneTempDir, false)
	require.Nil(t, err)

	// Add the test data set
	const testFileName = "test-file"
	require.Nil(t, ioutil.WriteFile(
		filepath.Join(cloneTempDir, testFileName),
		[]byte("test-content"),
		os.FileMode(0644),
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
	firstCommit, err := worktree.Commit("First commit", &gogit.CommitOptions{
		Author: author,
	})
	require.Nil(t, err)

	firstTagName := "v1.17.0"
	firstTagRef, err := cloneRepo.CreateTag(firstTagName, firstCommit,
		&gogit.CreateTagOptions{
			Tagger:  author,
			Message: firstTagName,
		},
	)
	require.Nil(t, err)

	// Create a test branch and a test commit on top
	branchName := "release-1.17"
	require.Nil(t, command.NewWithWorkDir(
		cloneTempDir, "git", "checkout", "-b", branchName,
	).RunSuccess())

	const branchTestFileName = "branch-test-file"
	require.Nil(t, ioutil.WriteFile(
		filepath.Join(cloneTempDir, branchTestFileName),
		[]byte("test-content"),
		os.FileMode(0644),
	))
	_, err = worktree.Add(branchTestFileName)
	require.Nil(t, err)

	firstBranchCommit, err := worktree.Commit("Second commit", &gogit.CommitOptions{
		Author: author,
		All:    true,
	})
	require.Nil(t, err)

	secondTagName := "v0.1.1"
	secondTagRef, err := cloneRepo.CreateTag(secondTagName, firstBranchCommit,
		&gogit.CreateTagOptions{
			Tagger:  author,
			Message: secondTagName,
		},
	)
	require.Nil(t, err)

	const secondBranchTestFileName = "branch-test-file-2"
	require.Nil(t, ioutil.WriteFile(
		filepath.Join(cloneTempDir, secondBranchTestFileName),
		[]byte("test-content"),
		os.FileMode(0644),
	))
	_, err = worktree.Add(secondBranchTestFileName)
	require.Nil(t, err)

	secondBranchCommit, err := worktree.Commit("Third commit", &gogit.CommitOptions{
		Author: author,
		All:    true,
	})
	require.Nil(t, err)

	thirdTagName := "v1.17.1"
	thirdTagRef, err := cloneRepo.CreateTag(thirdTagName, secondBranchCommit,
		&gogit.CreateTagOptions{
			Tagger:  author,
			Message: thirdTagName,
		},
	)
	require.Nil(t, err)

	const thirdBranchTestFileName = "branch-test-file-3"
	require.Nil(t, ioutil.WriteFile(
		filepath.Join(cloneTempDir, thirdBranchTestFileName),
		[]byte("test-content"),
		os.FileMode(0644),
	))
	_, err = worktree.Add(thirdBranchTestFileName)
	require.Nil(t, err)

	thirdBranchCommit, err := worktree.Commit("Fourth commit", &gogit.CommitOptions{
		Author: author,
		All:    true,
	})
	require.Nil(t, err)

	// Push the test content into the bare repo
	_, err = cloneRepo.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemote,
		URLs: []string{bareTempDir},
	})
	require.Nil(t, err)
	require.Nil(t, cloneRepo.Push(&gogit.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/*:refs/*"},
	}))

	require.Nil(t, os.RemoveAll(cloneTempDir))

	// Provide a system under test inside the test repo
	sut, err := git.CloneOrOpenRepo("", bareTempDir, false)
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
		thirdBranchCommit:  thirdBranchCommit.String(),
		branchName:         branchName,
		firstTagName:       firstTagName,
		firstTagCommit:     firstTagRef.Hash().String(),
		secondTagName:      secondTagName,
		secondTagCommit:    secondTagRef.Hash().String(),
		thirdTagName:       thirdTagName,
		thirdTagCommit:     thirdTagRef.Hash().String(),
		testFileName:       filepath.Join(sut.Dir(), testFileName),
	}
}

func (r *testRepo) cleanup(t *testing.T) {
	require.Nil(t, os.RemoveAll(r.dir))
	require.Nil(t, os.RemoveAll(r.sut.Dir()))
}

func TestSuccessCloneOrOpen(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	secondRepo, err := git.CloneOrOpenRepo(testRepo.sut.Dir(), testRepo.dir, false)
	require.Nil(t, err)

	require.Equal(t, testRepo.sut.Dir(), secondRepo.Dir())
	require.Nil(t, secondRepo.Cleanup())
}

func TestSuccessDescribeTags(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	tag, err := testRepo.sut.Describe(
		git.NewDescribeOptions().
			WithRevision(testRepo.firstTagCommit).
			WithAbbrev(0).
			WithTags(),
	)
	require.Nil(t, err)
	require.Equal(t, tag, testRepo.firstTagName)
}

func TestFailureDescribeTags(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	_, err := testRepo.sut.Describe(
		git.NewDescribeOptions().
			WithRevision("wrong").
			WithAbbrev(0).
			WithTags(),
	)
	require.NotNil(t, err)
}

func TestSuccessHasRemoteBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.Nil(t, testRepo.sut.HasRemoteBranch(testRepo.branchName))
	require.Nil(t, testRepo.sut.HasRemoteBranch(git.Master))
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
	require.Equal(t, head, testRepo.thirdBranchCommit)
}

func TestSuccessMerge(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Merge(git.Master)
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

	mergeBase, err := testRepo.sut.MergeBase(git.Master, testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, mergeBase, testRepo.firstCommit)
}

func TestSuccessRevParse(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	masterRev, err := testRepo.sut.RevParse(git.Master)
	require.Nil(t, err)
	require.Equal(t, masterRev, testRepo.firstCommit)

	branchRev, err := testRepo.sut.RevParse(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, branchRev, testRepo.thirdBranchCommit)

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

	masterRev, err := testRepo.sut.RevParseShort(git.Master)
	require.Nil(t, err)
	require.Equal(t, masterRev, testRepo.firstCommit[:10])

	branchRev, err := testRepo.sut.RevParseShort(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, branchRev, testRepo.thirdBranchCommit[:10])

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

	err := testRepo.sut.Push(git.Master)
	require.Nil(t, err)
}

func TestFailurePush(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Push("wrong")
	require.NotNil(t, err)
}

func TestSuccessRemotify(t *testing.T) {
	newRemote := git.Remotify(git.Master)
	require.Equal(t, newRemote, git.DefaultRemote+"/"+git.Master)
}

func TestSuccessIsReleaseBranch(t *testing.T) {
	require.True(t, git.IsReleaseBranch("release-1.17"))
}

func TestFailureIsReleaseBranch(t *testing.T) {
	require.False(t, git.IsReleaseBranch("wrong-branch"))
}

func TestSuccessLatestTagForBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	version, err := testRepo.sut.LatestTagForBranch(git.Master)
	require.Nil(t, err)
	require.Equal(t, util.SemverToTagString(version), testRepo.firstTagName)
}

func TestSuccessLatestTagForBranchRelease(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	version, err := testRepo.sut.LatestTagForBranch("release-1.17")
	require.Nil(t, err)
	require.Equal(t, util.SemverToTagString(version), testRepo.thirdTagName)
}

func TestFailureLatestTagForBranchInvalidBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	version, err := testRepo.sut.LatestTagForBranch("wrong-branch")
	require.NotNil(t, err)
	require.Equal(t, version, semver.Version{})
}

func TestSuccessLatestPatchToPatch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	// This test case gets commits from v1.17.0 to v1.17.1
	result, err := testRepo.sut.LatestPatchToPatch(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, result.StartSHA(), testRepo.firstCommit)
	require.Equal(t, result.StartRev(), testRepo.firstTagName)
	require.Equal(t, result.EndRev(), testRepo.thirdTagName)
}

func TestSuccessLatestPatchToPatchNewTag(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	// This test case gets commits from v1.17.1 to a new v1.17.2
	nextMinorTag := "v1.17.2"
	require.Nil(t, command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "tag", nextMinorTag,
	).RunSuccess())

	result, err := testRepo.sut.LatestPatchToPatch(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, result.StartSHA(), testRepo.secondBranchCommit)
	require.Equal(t, result.StartRev(), testRepo.thirdTagName)
	require.Equal(t, result.EndRev(), nextMinorTag)
}

func TestFailureLatestPatchToPatchWrongBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.LatestPatchToPatch("wrong-branch")
	require.NotNil(t, err)
	require.Equal(t, git.DiscoverResult{}, result)
}

func TestSuccessLatestPatchToLatest(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	// This test case gets commits from v1.17.1 to head of release-1.17
	result, err := testRepo.sut.LatestPatchToLatest(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, result.StartSHA(), testRepo.secondBranchCommit)
	require.Equal(t, result.StartRev(), testRepo.thirdTagName)
	require.Equal(t, result.EndSHA(), testRepo.thirdBranchCommit)
}

func TestSuccessDry(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	testRepo.sut.SetDry()

	err := testRepo.sut.Push(git.Master)
	require.Nil(t, err)
}

func TestSuccessLatestReleaseBranchMergeBaseToLatest(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.LatestReleaseBranchMergeBaseToLatest()
	require.Nil(t, err)
	require.Equal(t, result.StartSHA(), testRepo.firstCommit)
	require.Equal(t, result.StartRev(), testRepo.firstTagName)
	require.Equal(t, result.EndSHA(), testRepo.firstCommit)
	require.Equal(t, result.EndRev(), git.Master)
}

func TestFailureLatestReleaseBranchMergeBaseToLatestNoLatestTag(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.Nil(t, command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "tag", "-d", testRepo.firstTagName,
	).RunSuccess())

	result, err := testRepo.sut.LatestReleaseBranchMergeBaseToLatest()
	require.NotNil(t, err)
	require.Equal(t, git.DiscoverResult{}, result)
}

func TestSuccessLatestNonPatchFinalToMinor(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	nextMinorTag := "v1.18.0"
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
	require.Equal(t, git.DiscoverResult{}, result)
}

func TestTagsForBranchMaster(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.TagsForBranch(git.Master)
	require.Nil(t, err)
	require.Equal(t, result, []string{testRepo.firstTagName})
}

func TestTagsForBranchOnBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.TagsForBranch(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, result, []string{
		testRepo.thirdTagName,
		testRepo.firstTagName,
		testRepo.secondTagName,
	})
}

func TestTagsForBranchFailureWrongBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.TagsForBranch("wrong-branch")
	require.NotNil(t, err)
	require.Nil(t, result)
}

func TestCheckoutSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.Nil(t, ioutil.WriteFile(
		testRepo.testFileName,
		[]byte("hello world"),
		os.FileMode(0644),
	))
	res, err := command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "diff", "--name-only").Run()
	require.Nil(t, err)
	require.True(t, res.Success())
	require.Contains(t, res.Output(), filepath.Base(testRepo.testFileName))

	err = testRepo.sut.Checkout(git.Master, testRepo.testFileName)
	require.Nil(t, err)

	res, err = command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "diff", "--name-only").Run()
	require.Nil(t, err)
	require.True(t, res.Success())
	require.Empty(t, res.Output())
}

func TestCheckoutFailureWrongRevision(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Checkout("wrong")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "checkout wrong did not succeed")
}

func TestAddSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	f, err := ioutil.TempFile(testRepo.sut.Dir(), "test")
	require.Nil(t, err)
	require.Nil(t, f.Close())

	filename := filepath.Base(f.Name())
	err = testRepo.sut.Add(filename)
	require.Nil(t, err)

	res, err := command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "diff", "--cached", "--name-only").Run()
	require.Nil(t, err)
	require.True(t, res.Success())
	require.Contains(t, res.Output(), filename)
}

func TestAddFailureWrongPath(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Add("wrong")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "adding file wrong to repository")
}

func TestCommitSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	commitMessage := "My commit message for this test"
	err := testRepo.sut.Commit(commitMessage)
	require.Nil(t, err)

	res, err := command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "log", "-1",
	).Run()
	require.Nil(t, err)
	require.True(t, res.Success())
	require.Contains(t, res.Output(), "Author: Anago GCB <nobody@k8s.io>")
	require.Contains(t, res.Output(), commitMessage)
}

func TestCurrentBranchDefault(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	branch, err := testRepo.sut.CurrentBranch()
	require.Nil(t, err)
	require.Equal(t, testRepo.branchName, branch)
}

func TestCurrentBranchMaster(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)
	require.Nil(t, testRepo.sut.Checkout(git.Master))

	branch, err := testRepo.sut.CurrentBranch()
	require.Nil(t, err)
	require.Equal(t, git.Master, branch)
}

func TestRmSuccessForce(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)
	require.Nil(t, ioutil.WriteFile(testRepo.testFileName,
		[]byte("test"), os.FileMode(0755)),
	)

	require.Nil(t, testRepo.sut.Rm(true, testRepo.testFileName))

	_, err := os.Stat(testRepo.testFileName)
	require.True(t, os.IsNotExist(err))
}

func TestHasRemoteSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.AddRemote("test", "owner", "repo")
	require.Nil(t, err)

	remotes, err := testRepo.sut.Remotes()
	require.Nil(t, err)

	require.Len(t, remotes, 2)

	// The origin remote
	require.Equal(t, git.DefaultRemote, remotes[0].Name())
	require.Len(t, remotes[0].URLs(), 1)
	require.Equal(t, testRepo.dir, remotes[0].URLs()[0])

	// Or via the API
	require.True(t, testRepo.sut.HasRemote("origin", testRepo.dir))

	// The added test remote
	require.Equal(t, "test", remotes[1].Name())
	require.Len(t, remotes[1].URLs(), 1)

	url := git.GetRepoURL("owner", "repo", true)
	require.Equal(t, url, remotes[1].URLs()[0])

	// Or via the API
	require.True(t, testRepo.sut.HasRemote("test", url))
}

func TestHasRemoteFailure(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.False(t, testRepo.sut.HasRemote("name", "some-url.com"))
}

func TestRmFailureForce(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)
	require.NotNil(t, testRepo.sut.Rm(true, "invalid"))
}

func TestRmSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.Nil(t, testRepo.sut.Rm(true, testRepo.testFileName))

	_, err := os.Stat(testRepo.testFileName)
	require.True(t, os.IsNotExist(err))
}

func TestRmFailureModified(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)
	require.Nil(t, ioutil.WriteFile(testRepo.testFileName,
		[]byte("test"), os.FileMode(0755)),
	)
	require.NotNil(t, testRepo.sut.Rm(false, testRepo.testFileName))
}

func TestOpenRepoSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	repo, err := git.OpenRepo(testRepo.sut.Dir())
	require.Nil(t, err)
	require.Equal(t, testRepo.sut.Dir(), repo.Dir())
}

func TestOpenRepoSuccessSearchGitDot(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	repo, err := git.OpenRepo(filepath.Join(testRepo.sut.Dir(), "not-existing"))
	require.Nil(t, err)
	require.Equal(t, testRepo.sut.Dir(), repo.Dir())
}

func TestOpenRepoFailure(t *testing.T) {
	repo, err := git.OpenRepo("/invalid")
	require.NotNil(t, err)
	require.Nil(t, repo)
}

func TestAddRemoteSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.AddRemote("remote", "owner", "repo")
	require.Nil(t, err)
}

func TestAddRemoteFailureAlreadyExisting(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.AddRemote(git.DefaultRemote, "owner", "repo")
	require.NotNil(t, err)
}

func TestPushToRemoteSuccessRemoteMaster(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.PushToRemote(git.DefaultRemote, git.Remotify(git.Master))
	require.Nil(t, err)
}

func TestPushToRemoteSuccessBranchTracked(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.PushToRemote(git.DefaultRemote, testRepo.branchName)
	require.Nil(t, err)
}

func TestPushToRemoteFailureBranchNotExisting(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.PushToRemote(git.DefaultRemote, "some-branch")
	require.NotNil(t, err)
}

func TestLSRemoteSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	res, err := testRepo.sut.LsRemote()
	require.Nil(t, err)
	require.Contains(t, res, testRepo.firstCommit)
	require.Contains(t, res, testRepo.secondBranchCommit)
	require.Contains(t, res, testRepo.thirdBranchCommit)
}

func TestLSRemoteFailure(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	res, err := testRepo.sut.LsRemote("invalid")
	require.NotNil(t, err)
	require.Empty(t, res)
}

func TestBranchSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	res, err := testRepo.sut.Branch()
	require.Nil(t, err)
	require.Contains(t, res, testRepo.branchName)
}

func TestBranchFailure(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	res, err := testRepo.sut.Branch("--invalid")
	require.NotNil(t, err)
	require.Empty(t, res)
}

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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/blang/semver"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/git"
	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/util"
)

var testAuthor = &object.Signature{
	Name:  "John Doe",
	Email: "john@doe.org",
	When:  time.Now(),
}

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
	bareTempDir, err := os.MkdirTemp("", "k8s-test-bare-")
	require.Nil(t, err)

	bareRepo, err := gogit.PlainInit(bareTempDir, true)
	require.Nil(t, err)
	require.NotNil(t, bareRepo)

	// Clone from the bare to be able to add our test data
	cloneTempDir, err := os.MkdirTemp("", "k8s-test-clone-")
	require.Nil(t, err)
	cloneRepo, err := gogit.PlainInit(cloneTempDir, false)
	require.Nil(t, err)

	// Add the test data set
	const testFileName = "test-file"
	require.Nil(t, os.WriteFile(
		filepath.Join(cloneTempDir, testFileName),
		[]byte("test-content"),
		os.FileMode(0o644),
	))

	worktree, err := cloneRepo.Worktree()
	require.Nil(t, err)
	_, err = worktree.Add(testFileName)
	require.Nil(t, err)

	firstCommit, err := worktree.Commit("First commit", &gogit.CommitOptions{
		Author: testAuthor,
	})
	require.Nil(t, err)

	firstTagName := "v1.17.0"
	firstTagRef, err := cloneRepo.CreateTag(firstTagName, firstCommit,
		&gogit.CreateTagOptions{
			Tagger:  testAuthor,
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
	require.Nil(t, os.WriteFile(
		filepath.Join(cloneTempDir, branchTestFileName),
		[]byte("test-content"),
		os.FileMode(0o644),
	))
	_, err = worktree.Add(branchTestFileName)
	require.Nil(t, err)

	firstBranchCommit, err := worktree.Commit("Second commit", &gogit.CommitOptions{
		Author: testAuthor,
		All:    true,
	})
	require.Nil(t, err)

	secondTagName := "v0.1.1"
	secondTagRef, err := cloneRepo.CreateTag(secondTagName, firstBranchCommit,
		&gogit.CreateTagOptions{
			Tagger:  testAuthor,
			Message: secondTagName,
		},
	)
	require.Nil(t, err)

	const secondBranchTestFileName = "branch-test-file-2"
	require.Nil(t, os.WriteFile(
		filepath.Join(cloneTempDir, secondBranchTestFileName),
		[]byte("test-content"),
		os.FileMode(0o644),
	))
	_, err = worktree.Add(secondBranchTestFileName)
	require.Nil(t, err)

	secondBranchCommit, err := worktree.Commit("Third commit", &gogit.CommitOptions{
		Author: testAuthor,
		All:    true,
	})
	require.Nil(t, err)

	thirdTagName := "v1.17.1"
	thirdTagRef, err := cloneRepo.CreateTag(thirdTagName, secondBranchCommit,
		&gogit.CreateTagOptions{
			Tagger:  testAuthor,
			Message: thirdTagName,
		},
	)
	require.Nil(t, err)

	const thirdBranchTestFileName = "branch-test-file-3"
	require.Nil(t, os.WriteFile(
		filepath.Join(cloneTempDir, thirdBranchTestFileName),
		[]byte("test-content"),
		os.FileMode(0o644),
	))
	_, err = worktree.Add(thirdBranchTestFileName)
	require.Nil(t, err)

	thirdBranchCommit, err := worktree.Commit("Fourth commit", &gogit.CommitOptions{
		Author: testAuthor,
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

	require.Equal(t, secondRepo.Dir(), testRepo.sut.Dir())
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
	require.Equal(t, testRepo.firstTagName, tag)
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

	for _, repo := range []string{testRepo.branchName, git.DefaultBranch} {
		branchExists, err := testRepo.sut.HasRemoteBranch(repo)
		require.Nil(t, err)
		require.Equal(t, branchExists, true)
	}
}

func TestFailureHasRemoteBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	// TODO: Let's simulate an actual git/network failure

	branchExists, err := testRepo.sut.HasRemoteBranch("wrong")
	require.Equal(t, branchExists, false)
	require.Nil(t, err)
}

func TestSuccessHead(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	head, err := testRepo.sut.Head()
	require.Nil(t, err)
	require.Equal(t, testRepo.thirdBranchCommit, head)
}

func TestSuccessMerge(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Merge(git.DefaultBranch)
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

	mergeBase, err := testRepo.sut.MergeBase(git.DefaultBranch, testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit, mergeBase)
}

func TestSuccessRevParse(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	mainRev, err := testRepo.sut.RevParse(git.DefaultBranch)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit, mainRev)

	branchRev, err := testRepo.sut.RevParse(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, testRepo.thirdBranchCommit, branchRev)

	tagRev, err := testRepo.sut.RevParse(testRepo.firstTagName)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit, tagRev)

	tagRev, err = testRepo.sut.RevParse(testRepo.firstCommit)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit, tagRev)
}

func TestSuccessRevTagParse(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	mainRev, err := testRepo.sut.RevParseTag(git.DefaultBranch)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit, mainRev)

	branchRev, err := testRepo.sut.RevParseTag(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, testRepo.thirdBranchCommit, branchRev)

	tagRev, err := testRepo.sut.RevParseTag(testRepo.firstTagName)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit, tagRev)
}

func TestFailureRevParse(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	_, err := testRepo.sut.RevParse("wrong")
	require.NotNil(t, err)
}

func TestFailureRevParseTag(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	_, err := testRepo.sut.RevParseTag("wrong")
	require.NotNil(t, err)

	_, err = testRepo.sut.RevParseTag(testRepo.firstCommit)
	require.NotNil(t, err)
}

func TestSuccessRevParseShort(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	mainRev, err := testRepo.sut.RevParseShort(git.DefaultBranch)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit[:10], mainRev)

	branchRev, err := testRepo.sut.RevParseShort(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, testRepo.thirdBranchCommit[:10], branchRev)

	tagRev, err := testRepo.sut.RevParseShort(testRepo.firstTagName)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit[:10], tagRev)

	tagRev, err = testRepo.sut.RevParseShort(testRepo.firstCommit)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit[:10], tagRev)
}

func TestSuccessRevParseTagShort(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	mainRev, err := testRepo.sut.RevParseTagShort(git.DefaultBranch)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit[:10], mainRev)

	branchRev, err := testRepo.sut.RevParseTagShort(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, testRepo.thirdBranchCommit[:10], branchRev)

	tagRev, err := testRepo.sut.RevParseTagShort(testRepo.firstTagName)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit[:10], tagRev)
}

func TestFailureRevParseShort(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	_, err := testRepo.sut.RevParseShort("wrong")
	require.NotNil(t, err)
}

func TestFailureRevParseTagShort(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	_, err := testRepo.sut.RevParseTagShort("wrong")
	require.NotNil(t, err)

	_, err = testRepo.sut.RevParseTagShort(testRepo.firstCommit)
	require.NotNil(t, err)
}

func TestSuccessPush(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Push(git.DefaultBranch)
	require.Nil(t, err)
}

func TestFailurePush(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.Push("wrong")
	require.NotNil(t, err)
}

func TestSuccessRemotify(t *testing.T) {
	newRemote := git.Remotify(git.DefaultBranch)
	require.Equal(t, git.DefaultRemote+"/"+git.DefaultBranch, newRemote)
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

	version, err := testRepo.sut.LatestTagForBranch(git.DefaultBranch)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstTagName, util.SemverToTagString(version))
}

func TestSuccessLatestTagForBranchRelease(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	version, err := testRepo.sut.LatestTagForBranch("release-1.17")
	require.Nil(t, err)
	require.Equal(t, testRepo.thirdTagName, util.SemverToTagString(version))
}

func TestFailureLatestTagForBranchInvalidBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	version, err := testRepo.sut.LatestTagForBranch("wrong-branch")
	require.NotNil(t, err)
	require.Equal(t, semver.Version{}, version)
}

func TestSuccessLatestPatchToPatch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	// This test case gets commits from v1.17.0 to v1.17.1
	result, err := testRepo.sut.LatestPatchToPatch(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit, result.StartSHA())
	require.Equal(t, testRepo.firstTagName, result.StartRev())
	require.Equal(t, testRepo.thirdTagName, result.EndRev())
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
	require.Equal(t, testRepo.secondBranchCommit, result.StartSHA())
	require.Equal(t, testRepo.thirdTagName, result.StartRev())
	require.Equal(t, nextMinorTag, result.EndRev())
}

func TestFailureLatestPatchToPatchWrongBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.LatestPatchToPatch("wrong-branch")
	require.NotNil(t, err)
	require.Equal(t, result, git.DiscoverResult{})
}

func TestSuccessLatestPatchToLatest(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	// This test case gets commits from v1.17.1 to head of release-1.17
	result, err := testRepo.sut.LatestPatchToLatest(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, testRepo.secondBranchCommit, result.StartSHA())
	require.Equal(t, testRepo.thirdTagName, result.StartRev())
	require.Equal(t, testRepo.thirdBranchCommit, result.EndSHA())
}

func TestSuccessDry(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	testRepo.sut.SetDry()

	err := testRepo.sut.Push(git.DefaultBranch)
	require.Nil(t, err)
}

func TestSuccessLatestReleaseBranchMergeBaseToLatest(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.LatestReleaseBranchMergeBaseToLatest()
	require.Nil(t, err)
	require.Equal(t, testRepo.firstCommit, result.StartSHA())
	require.Equal(t, testRepo.firstTagName, result.StartRev())
	require.Equal(t, testRepo.firstCommit, result.EndSHA())
	require.Equal(t, git.DefaultBranch, result.EndRev())
}

func TestFailureLatestReleaseBranchMergeBaseToLatestNoLatestTag(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.Nil(t, command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "tag", "-d", testRepo.firstTagName,
	).RunSuccess())

	result, err := testRepo.sut.LatestReleaseBranchMergeBaseToLatest()
	require.NotNil(t, err)
	require.Equal(t, result, git.DiscoverResult{})
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
	require.Equal(t, testRepo.firstCommit, result.StartSHA())
	require.Equal(t, testRepo.firstTagName, result.StartRev())
	require.Equal(t, nextMinorTag, result.EndRev())
}

func TestFailureLatestNonPatchFinalToMinor(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.LatestNonPatchFinalToMinor()
	require.NotNil(t, err)
	require.Equal(t, result, git.DiscoverResult{})
}

func TestTagsForBranchMain(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.TagsForBranch(git.DefaultBranch)
	require.Nil(t, err)
	require.Equal(t, []string{testRepo.firstTagName}, result)
}

func TestTagsForBranchOnBranch(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	result, err := testRepo.sut.TagsForBranch(testRepo.branchName)
	require.Nil(t, err)
	require.Equal(t, []string{
		testRepo.thirdTagName,
		testRepo.firstTagName,
		testRepo.secondTagName,
	}, result)
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

	require.Nil(t, os.WriteFile(
		testRepo.testFileName,
		[]byte("hello world"),
		os.FileMode(0o644),
	))
	res, err := command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "diff", "--name-only").Run()
	require.Nil(t, err)
	require.True(t, res.Success())
	require.Contains(t, res.Output(), filepath.Base(testRepo.testFileName))

	err = testRepo.sut.Checkout(git.DefaultBranch, testRepo.testFileName)
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

	f, err := os.CreateTemp(testRepo.sut.Dir(), "test")
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
	require.Equal(t, branch, testRepo.branchName)
}

func TestCurrentBranchMain(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)
	require.Nil(t, testRepo.sut.Checkout(git.DefaultBranch))

	branch, err := testRepo.sut.CurrentBranch()
	require.Nil(t, err)
	require.Equal(t, branch, git.DefaultBranch)
}

func TestRmSuccessForce(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)
	require.Nil(t, os.WriteFile(testRepo.testFileName,
		[]byte("test"), os.FileMode(0o755)),
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
	require.Equal(t, remotes[0].Name(), git.DefaultRemote)
	require.Len(t, remotes[0].URLs(), 1)
	require.Equal(t, remotes[0].URLs()[0], testRepo.dir)

	// Or via the API
	require.True(t, testRepo.sut.HasRemote("origin", testRepo.dir))

	// The added test remote
	require.Equal(t, remotes[1].Name(), "test")
	require.Len(t, remotes[1].URLs(), 1)

	url := git.GetRepoURL("owner", "repo", true)
	require.Equal(t, remotes[1].URLs()[0], url)

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
	require.Nil(t, os.WriteFile(testRepo.testFileName,
		[]byte("test"), os.FileMode(0o755)),
	)
	require.NotNil(t, testRepo.sut.Rm(false, testRepo.testFileName))
}

func TestOpenRepoSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	repo, err := git.OpenRepo(testRepo.sut.Dir())
	require.Nil(t, err)
	require.Equal(t, repo.Dir(), testRepo.sut.Dir())
}

func TestOpenRepoSuccessSearchGitDot(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	repo, err := git.OpenRepo(filepath.Join(testRepo.sut.Dir(), "not-existing"))
	require.Nil(t, err)
	require.Equal(t, repo.Dir(), testRepo.sut.Dir())
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

func TestPushToRemoteSuccessRemoteMain(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	err := testRepo.sut.PushToRemote(git.DefaultRemote, git.Remotify(git.DefaultBranch))
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

func TestIsDirtySuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	dirty, err := testRepo.sut.IsDirty()
	require.Nil(t, err)
	require.False(t, dirty)
}

func TestIsDirtyFailure(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.Nil(t, os.WriteFile(
		filepath.Join(testRepo.sut.Dir(), "any-file"),
		[]byte("test"), os.FileMode(0o755)),
	)

	dirty, err := testRepo.sut.IsDirty()
	require.Nil(t, err)
	require.True(t, dirty)
}

func TestSetURLSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	const remote = "https://exmaple.com"
	require.Nil(t, testRepo.sut.SetURL(git.DefaultRemote, remote))
	remotes, err := testRepo.sut.Remotes()
	require.Nil(t, err)
	require.Len(t, remotes, 1)
	require.Equal(t, git.DefaultRemote, remotes[0].Name())
	require.Len(t, remotes[0].URLs(), 1)
	require.Equal(t, remotes[0].URLs()[0], remote)
}

func TestSetURLFailureRemoteDoesNotExists(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	require.NotNil(t, testRepo.sut.SetURL("some-remote", ""))
}

func TestAllTags(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	tags, err := testRepo.sut.Tags()
	require.Nil(t, err)
	require.Len(t, tags, 3)
	require.Equal(t, tags[0], testRepo.secondTagName)
	require.Equal(t, tags[1], testRepo.firstTagName)
	require.Equal(t, tags[2], testRepo.thirdTagName)
}

func TestCommitEmptySuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	commitMessage := "This is an empty commit"
	require.Nil(t, testRepo.sut.CommitEmpty(commitMessage))
	res, err := command.NewWithWorkDir(
		testRepo.sut.Dir(), "git", "log", "-1",
	).Run()
	require.Nil(t, err)
	require.True(t, res.Success())
	require.Contains(t, res.Output(), commitMessage)
}

func TestTagSuccess(t *testing.T) {
	testRepo := newTestRepo(t)
	defer testRepo.cleanup(t)

	testTag := "testTag"
	require.Nil(t, testRepo.sut.Tag(testTag, "message"))
	tags, err := testRepo.sut.TagsForBranch(testRepo.branchName)
	require.Nil(t, err)
	require.Contains(t, tags, testTag)
}

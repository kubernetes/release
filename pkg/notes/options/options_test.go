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

package options

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	kgit "sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-sdk/github"
	"sigs.k8s.io/release-utils/command"
)

type testOptions struct {
	*Options
	testRepo *testRepo
}

type testRepo struct {
	sut                *kgit.Repo
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

func newTestOptions(t *testing.T) *testOptions {
	testRepo := newTestRepo(t)
	t.Setenv(github.TokenEnvKey, "token")
	return &testOptions{
		Options: &Options{
			DiscoverMode: RevisionDiscoveryModeNONE,
			StartSHA:     "0",
			EndSHA:       "0",
			Format:       FormatMarkdown,
			GoTemplate:   GoTemplateDefault,
			Pull:         true,
			gitCloneFn: func(string, string, string, bool) (*kgit.Repo, error) {
				return testRepo.sut, nil
			},
		},
		testRepo: testRepo,
	}
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
//
//   - commit `firstCommit` (tag: `firstTagName`, origin/master, origin/HEAD, master)
//     Author: John Doe <john@doe.org>
//
//     First commit
func newTestRepo(t *testing.T) *testRepo {
	// Setup the bare repo as base
	bareTempDir, err := os.MkdirTemp("", "k8s-test-bare-")
	require.NoError(t, err)

	bareRepo, err := git.PlainInit(bareTempDir, true)
	require.NoError(t, err)
	require.NotNil(t, bareRepo)

	// Clone from the bare to be able to add our test data
	cloneTempDir, err := os.MkdirTemp("", "k8s-test-clone-")
	require.NoError(t, err)
	cloneRepo, err := git.PlainInit(cloneTempDir, false)
	require.NoError(t, err)

	// Add the test data set
	const testFileName = "test-file"
	require.NoError(t, os.WriteFile(
		filepath.Join(cloneTempDir, testFileName),
		[]byte("test-content"),
		os.FileMode(0o644),
	))

	worktree, err := cloneRepo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add(testFileName)
	require.NoError(t, err)

	author := &object.Signature{
		Name:  "John Doe",
		Email: "john@doe.org",
		When:  time.Now(),
	}
	firstCommit, err := worktree.Commit("First commit", &git.CommitOptions{
		Author: author,
	})
	require.NoError(t, err)

	firstTagName := "v1.17.0"
	firstTagRef, err := cloneRepo.CreateTag(firstTagName, firstCommit,
		&git.CreateTagOptions{
			Tagger:  author,
			Message: firstTagName,
		},
	)
	require.NoError(t, err)

	// Create a test branch and a test commit on top
	branchName := "release-1.17"
	require.NoError(t, command.NewWithWorkDir(
		cloneTempDir, "git", "checkout", "-b", branchName,
	).RunSuccess())

	const branchTestFileName = "branch-test-file"
	require.NoError(t, os.WriteFile(
		filepath.Join(cloneTempDir, branchTestFileName),
		[]byte("test-content"),
		os.FileMode(0o644),
	))
	_, err = worktree.Add(branchTestFileName)
	require.NoError(t, err)

	firstBranchCommit, err := worktree.Commit("Second commit", &git.CommitOptions{
		Author: author,
		All:    true,
	})
	require.NoError(t, err)

	secondTagName := "v0.1.1"
	secondTagRef, err := cloneRepo.CreateTag(secondTagName, firstBranchCommit,
		&git.CreateTagOptions{
			Tagger:  author,
			Message: firstTagName,
		},
	)
	require.NoError(t, err)

	const secondBranchTestFileName = "branch-test-file-2"
	require.NoError(t, os.WriteFile(
		filepath.Join(cloneTempDir, secondBranchTestFileName),
		[]byte("test-content"),
		os.FileMode(0o644),
	))
	_, err = worktree.Add(secondBranchTestFileName)
	require.NoError(t, err)

	secondBranchCommit, err := worktree.Commit("Third commit", &git.CommitOptions{
		Author: author,
		All:    true,
	})
	require.NoError(t, err)

	thirdTagName := "v0.1.2"
	thirdTagRef, err := cloneRepo.CreateTag(thirdTagName, secondBranchCommit,
		&git.CreateTagOptions{
			Tagger:  author,
			Message: firstTagName,
		},
	)
	require.NoError(t, err)

	// Push the test content into the bare repo
	_, err = cloneRepo.CreateRemote(&config.RemoteConfig{
		Name: kgit.DefaultRemote,
		URLs: []string{bareTempDir},
	})
	require.NoError(t, err)
	require.NoError(t, cloneRepo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/*:refs/*"},
	}))

	require.NoError(t, os.RemoveAll(cloneTempDir))

	// Provide a system under test inside the test repo
	opts := &git.CloneOptions{}
	sut, err := kgit.CloneOrOpenRepo("", bareTempDir, false, false, opts)
	require.NoError(t, err)
	require.NoError(t, command.NewWithWorkDir(
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
	require.NoError(t, os.RemoveAll(r.dir))
	require.NoError(t, os.RemoveAll(r.sut.Dir()))
}

func TestNew(t *testing.T) {
	require.NotNil(t, New())
}

func TestValidateAndFinishSuccess(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	require.NoError(t, options.ValidateAndFinish())
	require.Equal(t, logrus.InfoLevel, logrus.StandardLogger().GetLevel())
}

func TestValidateAndFinishSuccessDebug(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.Debug = true
	require.NoError(t, options.ValidateAndFinish())
	require.Equal(t, logrus.DebugLevel, logrus.StandardLogger().GetLevel())
}

func TestValidateAndFinishFailureGithubTokenMissing(t *testing.T) {
	options := &Options{}
	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureStartShaAndRevWrong(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.StartSHA = ""
	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureEndShaAndRevWrong(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.EndSHA = ""
	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureClone(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.StartRev = options.testRepo.firstTagName
	options.gitCloneFn = func(string, string, string, bool) (*kgit.Repo, error) {
		return nil, errors.New("error")
	}
	options.StartSHA = ""
	options.EndSHA = ""

	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishSuccessStartRev(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.StartRev = options.testRepo.firstTagName
	require.NoError(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureStartRevNotExisting(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.StartRev = "invalid"
	options.StartSHA = ""

	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishSuccessEndRev(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.EndRev = options.testRepo.firstTagName
	require.NoError(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureEndRevNotExisting(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.EndRev = "invalid"
	options.EndSHA = ""

	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishSuccessDiscoveryModeMergeBaseToLatest(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	require.NoError(t, command.NewWithWorkDir(options.testRepo.sut.Dir(),
		"git", "checkout", "-b", options.testRepo.firstTagName).RunSuccess())

	options.DiscoverMode = RevisionDiscoveryModeMergeBaseToLatest
	require.NoError(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureDiscoveryModeMergeBaseToLatestNoTag(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	require.NoError(t, command.NewWithWorkDir(options.testRepo.sut.Dir(),
		"git", "tag", "-d", options.testRepo.firstTagName).RunSuccess())

	options.DiscoverMode = RevisionDiscoveryModeMergeBaseToLatest
	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureDiscoveryModeMergeBaseToLatestClone(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.DiscoverMode = RevisionDiscoveryModeMergeBaseToLatest
	options.gitCloneFn = func(string, string, string, bool) (*kgit.Repo, error) {
		return nil, errors.New("error")
	}
	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishSuccessDiscoveryModePatchToPatch(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	nextMinorTag := "v1.17.1"
	require.NoError(t, command.NewWithWorkDir(
		options.testRepo.sut.Dir(), "git", "tag", nextMinorTag,
	).RunSuccess())

	options.Branch = options.testRepo.branchName
	options.DiscoverMode = RevisionDiscoveryModePatchToPatch
	require.NoError(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureDiscoveryModePatchToPatchNoBranch(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	options.DiscoverMode = RevisionDiscoveryModePatchToPatch
	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureFormat(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	// Given
	options.Format = "wrong"

	// When
	require.Error(t, options.ValidateAndFinish())
}

func TestValidateAndFinishFailureGoTemplate(t *testing.T) {
	options := newTestOptions(t)
	defer options.testRepo.cleanup(t)

	// Given
	options.GoTemplate = "wrong"

	// When
	require.Error(t, options.ValidateAndFinish())
}

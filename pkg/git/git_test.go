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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/git/gitfakes"
)

func newSUT() (*git.Repo, *gitfakes.FakeWorktree) {
	repoMock := &gitfakes.FakeRepository{}
	worktreeMock := &gitfakes.FakeWorktree{}

	repo := &git.Repo{}
	repo.SetWorktree(worktreeMock)
	repo.SetInnerRepo(repoMock)

	return repo, worktreeMock
}

func TestCommit(t *testing.T) {
	repo, worktreeMock := newSUT()
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

		actual := git.GetDefaultKubernetesRepoURL()
		require.Equal(t, tc.expected, actual)
	}
}

// createTestRepository creates a test repo, cd into it and returns the path
func createTestRepository() (repoPath string, err error) {
	repoPath, err = ioutil.TempDir(os.TempDir(), "sigrelease-test-repo-*")
	if err != nil {
		return "", errors.Wrap(err, "creating a directory for test repository")
	}
	if err := os.Chdir(repoPath); err != nil {
		return "", errors.Wrap(err, "cd'ing into test repository")
	}
	out, err := exec.Command("git", "init").Output()
	if err != nil {
		return "", errors.Wrapf(err, "initializing test repository: %s", out)
	}
	return repoPath, nil
}

func TestGetUserName(t *testing.T) {
	const fakeUserName = "SIG Release Test User"
	currentDir, err := os.Getwd()
	require.Nil(t, err, "error reading the current directory")
	defer os.Chdir(currentDir) // nolint: errcheck

	// Create an empty repo and configure the users name to test
	repoPath, err := createTestRepository()
	require.Nil(t, err, "getting a test repo")

	// Call git to configure the user's name:
	_, err = exec.Command("git", "config", "user.name", fakeUserName).Output()
	require.Nil(t, err, fmt.Sprintf("configuring fake user email in %s", repoPath))

	testRepo, err := git.OpenRepo(repoPath)
	require.Nil(t, err, fmt.Sprintf("opening test repo in %s", repoPath))
	defer testRepo.Cleanup() // nolint: errcheck

	actual, err := git.GetUserName()
	require.Nil(t, err)
	require.Equal(t, fakeUserName, actual)
	require.NotEqual(t, fakeUserName, "")
}

func TestGetUserEmail(t *testing.T) {
	const fakeUserEmail = "kubernetes-test@example.com"
	currentDir, err := os.Getwd() // nolint: errcheck
	require.Nil(t, err, "error reading the current directory")
	defer os.Chdir(currentDir) // nolint: errcheck

	// Create an empty repo and configure the users name to test
	repoPath, err := createTestRepository()
	require.Nil(t, err, "getting a test repo")

	// Call git to configure the user's name:
	_, err = exec.Command("git", "config", "user.email", fakeUserEmail).Output()
	require.Nil(t, err, fmt.Sprintf("configuring fake user email in %s", repoPath))

	testRepo, err := git.OpenRepo(repoPath)
	require.Nil(t, err, fmt.Sprintf("opening test repo in %s", repoPath))
	defer testRepo.Cleanup() // nolint: errcheck

	// Do the actual call
	actual, err := git.GetUserEmail()
	require.Nil(t, err)
	require.Equal(t, fakeUserEmail, actual)
	require.NotEqual(t, fakeUserEmail, "")
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

		actual := git.GetKubernetesRepoURL(tc.org, tc.useSSH)
		require.Equal(t, tc.expected, actual)
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

		actual := git.GetRepoURL(tc.org, tc.repo, tc.useSSH)
		require.Equal(t, tc.expected, actual)
	}
}

func TestRemotify(t *testing.T) {
	testcases := []struct{ provided, expected string }{
		{provided: git.DefaultBranch, expected: git.DefaultRemote + "/" + git.DefaultBranch},
		{provided: "origin/ref", expected: "origin/ref"},
		{provided: "base/another_ref", expected: "base/another_ref"},
	}

	for _, tc := range testcases {
		require.Equal(t, git.Remotify(tc.provided), tc.expected)
	}
}

func TestIsDirtyMockSuccess(t *testing.T) {
	repo, _ := newSUT()

	dirty, err := repo.IsDirty()

	require.Nil(t, err)
	require.False(t, dirty)
}

func TestIsDirtyMockSuccessDirty(t *testing.T) {
	repo, worktreeMock := newSUT()
	worktreeMock.StatusReturns(gogit.Status{
		"file": &gogit.FileStatus{
			Worktree: gogit.Modified,
		},
	}, nil)

	dirty, err := repo.IsDirty()

	require.Nil(t, err)
	require.True(t, dirty)
}

func TestIsDirtyMockFailureWorktreeStatus(t *testing.T) {
	repo, worktreeMock := newSUT()
	worktreeMock.StatusReturns(gogit.Status{}, errors.New(""))

	dirty, err := repo.IsDirty()

	require.NotNil(t, err)
	require.False(t, dirty)
}

func TestParseRepoSlug(t *testing.T) {
	slugTests := []struct {
		caseName, repoSlug, orgName, repoName string
		isValid                               bool
	}{
		{
			caseName: "valid slug", repoSlug: "kubernetes/release",
			orgName: "kubernetes", repoName: "release", isValid: true,
		},

		{
			caseName: "slug with hyphens", repoSlug: "kubernetes/repo_with_underscores",
			orgName: "", repoName: "", isValid: false,
		},

		{
			caseName: "slug with dashes", repoSlug: "kubernetes-sigs/release-notes",
			orgName: "kubernetes-sigs", repoName: "release-notes", isValid: true,
		},

		{
			caseName: "slug with uppercase", repoSlug: "GoogleCloudPlatform/compute-image-tools",
			orgName: "GoogleCloudPlatform", repoName: "compute-image-tools", isValid: true,
		},

		{
			caseName: "slug with invalid chars", repoSlug: "kubern#etes/not.valid",
			orgName: "", repoName: "", isValid: false,
		},

		{
			caseName: "slug with extra slash", repoSlug: "kubernetes/not/valid",
			orgName: "", repoName: "", isValid: false,
		},

		{
			caseName: "slug with only org", repoSlug: "kubernetes",
			orgName: "kubernetes", repoName: "", isValid: true,
		},
	}

	for _, testCase := range slugTests {
		org, repo, err := git.ParseRepoSlug(testCase.repoSlug)
		if testCase.isValid {
			require.Nil(t, err, testCase.caseName)
		} else {
			require.NotNil(t, err, testCase.caseName)
		}
		require.Equal(t, testCase.orgName, org, testCase.caseName)
		require.Equal(t, testCase.repoName, repo, testCase.caseName)
	}
}

func TestRetryErrors(t *testing.T) {
	retryErrorStrings := []string{
		"dial tcp: lookup github.com on [::1]:53",
		"read udp [::1]:48087->[::1]:53",
		"read: connection refused",
	}

	nonRetryErrorStrings := []string{
		"could not list references on the remote repository",
		"error checking remote branch",
		"src refspec release-chorizo does not match",
	}

	for _, message := range retryErrorStrings {
		err := git.NewNetworkError(errors.New(message))
		require.True(t, err.CanRetry(), fmt.Sprintf("Checking retriable error '%s'", message))
	}

	for _, message := range nonRetryErrorStrings {
		err := git.NewNetworkError(errors.New(message))
		require.False(t, err.CanRetry(), fmt.Sprintf("Checking non-retriable error '%s'", message))
	}
}

func TestNetworkError(t *testing.T) {
	// Return a NetWorkError in a fun that returns a standard error
	err := func() error {
		return git.NewNetworkError(errors.New("This is a test error"))
	}()
	require.NotNil(t, err, "checking if NewNetWork error returns nil")
	require.NotEmpty(t, err.Error(), "checking if NetworkError returns a message")
	require.False(t, err.(git.NetworkError).CanRetry(), "checking if network error can be properly asserted")
}

func TestHasBranch(t *testing.T) {
	testBranchName := "git-package-test-branch"
	repoPath, err := createTestRepository()
	require.Nil(t, err, "getting a test repo")

	// Create a file and a test commit
	testfile := filepath.Join(repoPath, "README.md")
	err = ioutil.WriteFile(testfile, []byte("# WHY SIG-RELEASE ROCKS\n\n"), os.FileMode(0o644))
	require.Nil(t, err, "writing test file")

	err = command.NewWithWorkDir(repoPath, "git", "add", testfile).RunSuccess()
	require.Nil(t, err, fmt.Sprintf("adding test file in %s", repoPath))

	err = command.NewWithWorkDir(repoPath, "git", "commit", "-m", "adding test file").RunSuccess()
	require.Nil(t, err, "creating first commit")

	// Call git to configure the user's name:
	err = command.NewWithWorkDir(repoPath, "git", "branch", testBranchName).RunSuccess()
	require.Nil(t, err, fmt.Sprintf("configuring test branch in %s", repoPath))

	// Now, open the repo and test to see if branches are there
	testRepo, err := git.OpenRepo(repoPath)
	require.Nil(t, err, fmt.Sprintf("opening test repo in %s", repoPath))
	defer testRepo.Cleanup() // nolint: errcheck

	actual, err := testRepo.HasBranch(testBranchName)
	require.Nil(t, err)
	require.True(t, actual)

	actual, err = testRepo.HasBranch(git.DefaultBranch)
	require.Nil(t, err)
	require.True(t, actual)

	actual, err = testRepo.HasBranch("non-existing-branch")
	require.Nil(t, err)
	require.False(t, actual)
}

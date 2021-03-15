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

package release

import (
	"os"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/git"
	"sigs.k8s.io/release-utils/command"
)

func getTestGitObjectPusher() (pusher *GitObjectPusher, repoPath string, err error) {
	// Initialize a test repository for the test pusher
	repoPath, err = os.MkdirTemp("", "sigrelease-test-repo-*")
	if err != nil {
		return nil, "", errors.Wrap(err, "creating a directory for test repository")
	}

	if err := command.NewWithWorkDir(
		repoPath, "git", "init").RunSilentSuccess(); err != nil {
		os.RemoveAll(repoPath)
		return nil, repoPath, errors.Wrapf(err, "initializing test repository")
	}

	// Create root commit
	if err := command.NewWithWorkDir(
		repoPath, "git", "commit", "--allow-empty", "-m", "Root commit",
	).RunSilentSuccess(); err != nil {
		os.RemoveAll(repoPath)
		return nil, repoPath, errors.Wrapf(err, "creating first commit")
	}

	// Check if branch exists (in case initial branch is 'main' and we expect 'master')
	out, err := command.NewWithWorkDir(repoPath, "git", "branch").RunSuccessOutput()
	if err != nil {
		return nil, repoPath, errors.Wrap(err, "listing branches in test repo")
	}
	if !strings.Contains(out.Output(), git.DefaultBranch) {
		if err := command.NewWithWorkDir(
			repoPath, "git", "branch", git.DefaultBranch,
		).RunSilentSuccess(); err != nil {
			return nil, repoPath, errors.Wrap(err, "creating main branch")
		}
	}

	pusher, err = NewGitPusher(&GitObjectPusherOptions{RepoPath: repoPath})
	if err != nil {
		return nil, repoPath, errors.Wrap(err, "creating test git pusher")
	}

	return pusher, repoPath, nil
}

func TestCheckBranchName(t *testing.T) {
	ghp, repoPath, err := getTestGitObjectPusher()
	if repoPath != "" {
		defer os.RemoveAll(repoPath)
	}
	require.Nil(t, err)

	sampleBaranches := []struct {
		branchName string
		valid      bool
	}{
		{"release-1.20", true},     // Valid name
		{"release-chorizo", false}, // Invalid, not a semver
		{"1.20", false},            // Invalid, has to start with release
	}
	for _, testCase := range sampleBaranches {
		if testCase.valid {
			require.Nil(t, ghp.checkBranchName(testCase.branchName))
		} else {
			require.NotNil(t, ghp.checkBranchName(testCase.branchName))
		}
	}
}

func TestCheckTagName(t *testing.T) {
	ghp, repoPath, err := getTestGitObjectPusher()
	if repoPath != "" {
		defer os.RemoveAll(repoPath)
	}
	require.Nil(t, err)

	sampleTags := []struct {
		tagName string
		valid   bool
	}{
		{"v1.20.0-alpha.2", true}, // Valid
		{"myTag", false},          // Invalid, not a semver
		{"1.20", false},           // Invalid, incomplete
	}
	for _, testCase := range sampleTags {
		if testCase.valid {
			require.Nil(t, ghp.checkTagName(testCase.tagName))
		} else {
			require.NotNil(t, ghp.checkTagName(testCase.tagName))
		}
	}
}

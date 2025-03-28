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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/env"
)

const (
	pseudoReleaseBranch = "release-2.0"
	testCommitMessage   = `"This is my test commit"`
	testDataDir         = "testdata"
)

type sut struct {
	repo        *git.Repo
	baseDir     string
	bareCopyDir string
	tempDir     string
}

func newSUT(t *testing.T) *sut {
	// Bazel test runs have no $HOME set, which prevents git to find its
	// global .gitconfig. This means we're silently assuming that we're running
	// inside a container with the actual $HOME of `/root`.
	if !env.IsSet("HOME") {
		t.Setenv("HOME", "/root")
	}

	// A local k/k repo will be our test base
	tempDir := filepath.Join(os.TempDir(), "k8s-test")
	require.NoError(t, os.MkdirAll(tempDir, 0o755))

	// The bare repo which is the pseudo remote base
	bareDir := filepath.Join(tempDir, "bare")

	const url = "https://github.com/kubernetes/kubernetes"

	if _, err := os.Stat(bareDir); os.IsNotExist(err) {
		require.NoError(t,
			command.New("git", "clone", "--bare", url, bareDir).RunSuccess(),
		)
	}

	now := time.Now().Unix()
	bareCopyDir := filepath.Join(tempDir, fmt.Sprintf("bare-%d", now))
	require.NoError(t, command.New("cp", "-r", bareDir, bareCopyDir).RunSuccess())

	// The base repo where every test case is inherited
	baseDir := filepath.Join(tempDir, fmt.Sprintf("base-%d", now))

	// Clone the repo from the bare, which is safe to modify
	require.NoError(t,
		command.New("git", "clone", bareCopyDir, baseDir).RunSuccess(),
	)

	// Modify the bare repo with test content
	require.NoError(t,
		command.NewWithWorkDir(baseDir,
			"git", "checkout", "-b", pseudoReleaseBranch,
		).RunSuccess(),
	)
	require.NoError(t,
		command.NewWithWorkDir(baseDir,
			"git", "push", "-u", git.DefaultRemote, pseudoReleaseBranch,
		).RunSuccess(),
	)
	require.NoError(t,
		command.NewWithWorkDir(baseDir,
			"git", "checkout", git.DefaultBranch,
		).RunSuccess(),
	)
	require.NoError(t,
		command.NewWithWorkDir(baseDir,
			"git", "commit", "--allow-empty", "-m", testCommitMessage,
		).RunSuccess(),
	)
	require.NoError(t,
		command.NewWithWorkDir(baseDir,
			"git", "push",
		).RunSuccess(),
	)

	// The sut repo dir
	repoDir := filepath.Join(tempDir, fmt.Sprintf("test-%d", now))
	require.NoError(t, command.New("cp", "-r", baseDir, repoDir).RunSuccess())

	opts := &gogit.CloneOptions{}
	repo, err := git.CloneOrOpenRepo(repoDir, url, false, false, opts)
	require.NoError(t, err)

	// Adapt the settings
	return &sut{repo, baseDir, bareCopyDir, tempDir}
}

func (s *sut) cleanup(t *testing.T) {
	require.NoError(t, os.RemoveAll(s.repo.Dir()))
	require.NoError(t, os.RemoveAll(s.baseDir))
	require.NoError(t, os.RemoveAll(s.bareCopyDir))
}

func (s *sut) lastCommit(t *testing.T, branch string) string {
	res, err := command.NewWithWorkDir(s.repo.Dir(),
		"git", "log", "-1", branch).RunSilentSuccessOutput()
	require.NoError(t, err)

	return res.OutputTrimNL()
}

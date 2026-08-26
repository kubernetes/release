/*
Copyright 2026 The Kubernetes Authors.

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

package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/command"
)

func testGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	_, err := command.NewWithWorkDir(dir, "git", args...).RunSilentSuccessOutput()
	require.NoError(t, err)
}

func TestStripRemoteCredentials(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	repoPath := filepath.Join(workDir, "src", "k8s.io", "kubernetes")
	require.NoError(t, os.MkdirAll(repoPath, 0o755))
	testGit(t, repoPath, "init", "-q")
	testGit(t, repoPath, "remote", "add", "origin", "https://git:supersecret@github.com/kubernetes/kubernetes")
	testGit(t, repoPath, "remote", "add", "clean", "https://github.com/kubernetes/release")

	require.NoError(t, stripRemoteCredentials(filepath.Join(workDir, "src")))

	repo, err := git.OpenRepo(repoPath)
	require.NoError(t, err)
	remotes, err := repo.Remotes()
	require.NoError(t, err)
	require.Len(t, remotes, 2)

	urls := map[string]string{}

	for _, remote := range remotes {
		require.Len(t, remote.URLs(), 1)
		urls[remote.Name()] = remote.URLs()[0]
	}

	// The credentials are stripped, remotes without credentials are untouched
	require.Equal(t, "https://github.com/kubernetes/kubernetes", urls["origin"])
	require.Equal(t, "https://github.com/kubernetes/release", urls["clean"])

	// No repositories under the path is not an error
	require.NoError(t, stripRemoteCredentials(t.TempDir()))
}

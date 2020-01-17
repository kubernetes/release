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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/git"
)

func TestChangelogNoArgumentsOrFlags(t *testing.T) {
	err := runChangelog()
	require.NotNil(t, err)
}

type sut struct {
	repo    *git.Repo
	tempDir string
}

func newSUT(t *testing.T, replayDir string) *sut {
	// A local k/k repo will be our test base
	const testDataDir = "testdata"
	tempDir, err := ioutil.TempDir("", "k8s-test-changelog-")
	require.Nil(t, err)
	require.Nil(t, os.MkdirAll(tempDir, 0o755))

	// The base repo where every test is inherited
	baseDir := filepath.Join(tempDir, "k8s-test-changelog-base")
	const url = "https://github.com/kubernetes/kubernetes"
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		require.Nil(t, command.New("git", "clone", url, baseDir).RunSuccess())
	}

	// The sut repo dir
	repoDir := filepath.Join(
		tempDir, fmt.Sprintf("k8s-test-changelog-%d", time.Now().Unix()),
	)
	require.Nil(t, command.New("cp", "-r", baseDir, repoDir).RunSuccess())

	repo, err := git.CloneOrOpenRepo(repoDir, url, false)
	require.Nil(t, err)

	// Create mocked release tars
	tarsDir := filepath.Join(tempDir, "k8s-test-changelog-tars")
	if _, err := os.Stat(tarsDir); os.IsNotExist(err) {
		require.Nil(t, os.MkdirAll(tarsDir, 0o755))
		for _, file := range []string{
			"kubernetes-client-darwin-386.tar.gz",
			"kubernetes-client-darwin-amd64.tar.gz",
			"kubernetes-client-linux-386.tar.gz",
			"kubernetes-client-linux-amd64.tar.gz",
			"kubernetes-client-linux-arm.tar.gz",
			"kubernetes-client-linux-arm64.tar.gz",
			"kubernetes-client-linux-ppc64le.tar.gz",
			"kubernetes-client-linux-s390x.tar.gz",
			"kubernetes-client-windows-386.tar.gz",
			"kubernetes-client-windows-amd64.tar.gz",
			"kubernetes-node-linux-amd64.tar.gz",
			"kubernetes-node-linux-arm.tar.gz",
			"kubernetes-node-linux-arm64.tar.gz",
			"kubernetes-node-linux-ppc64le.tar.gz",
			"kubernetes-node-linux-s390x.tar.gz",
			"kubernetes-node-windows-amd64.tar.gz",
			"kubernetes-server-linux-amd64.tar.gz",
			"kubernetes-server-linux-arm.tar.gz",
			"kubernetes-server-linux-arm64.tar.gz",
			"kubernetes-server-linux-ppc64le.tar.gz",
			"kubernetes-server-linux-s390x.tar.gz",
			"kubernetes-src.tar.gz",
			"kubernetes.tar.gz",
		} {
			require.Nil(t, ioutil.WriteFile(
				filepath.Join(tarsDir, file), []byte(file), 0o755,
			))
		}
	}

	// Set the global options
	changelogOpts.tars = tarsDir
	changelogOpts.replayDir = filepath.Join(testDataDir, replayDir)
	rootOpts.nomock = true
	rootOpts.repoPath = repo.Dir()

	return &sut{repo, tempDir}
}

func (s *sut) cleanup(t *testing.T) {
	require.Nil(t, os.RemoveAll(s.repo.Dir()))
}

func TestNewPatchReleaseMock(t *testing.T) {
	// Given
	s := newSUT(t, "changelog-v1.16.3")
	defer s.cleanup(t)
	changelogOpts.tag = "v1.16.3"

	// When
	require.Nil(t, runChangelog())

	// Then
	// Verify local results
	verifyResults := func(r *git.Repo) {
		require.FileExists(t, "CHANGELOG-1.16.html")
		for _, branch := range []string{"release-1.16", git.Master} {
			// Switch to the test branch
			require.Nil(t, r.Checkout(branch))

			// Verify commit message
			res, err := command.NewWithWorkDir(r.Dir(), "git", "log", "-1").Run()
			require.Nil(t, err)
			require.True(t, res.Success())
			require.Contains(t, res.Output(), "Anago GCB <nobody@k8s.io>")
			require.Contains(t, res.Output(), "CHANGELOG-1.16.md for v1.16.3")

			// Verify changelog contents
			changelog, err := ioutil.ReadFile(
				filepath.Join(r.Dir(), "CHANGELOG-1.16.md"),
			)
			require.Nil(t, err)
			require.Contains(t, string(changelog), patchReleaseExpectedTOC)
			require.Contains(t, string(changelog), patchReleaseExpectedContent)
		}
	}
	verifyResults(s.repo)
}

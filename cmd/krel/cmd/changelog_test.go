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

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/changelog"
	"k8s.io/release/pkg/git"
)

func (s *sut) getChangelogOptions(tag string) *changelog.Options {
	return &changelog.Options{
		RepoPath:     s.repo.Dir(),
		ReplayDir:    filepath.Join(testDataDir, "changelog-"+tag),
		Tag:          tag,
		Tars:         ".",
		Branch:       git.DefaultBranch,
		CloneCVEMaps: false,
	}
}

func fileContains(t *testing.T, file, contains string) {
	require.FileExists(t, file)
	content, err := os.ReadFile(file)
	require.Nil(t, err)
	require.Contains(t, string(content), contains)
}

func TestChangelogNoArgumentsOrFlags(t *testing.T) {
	err := changelog.New(changelogOptions).Run()
	require.NotNil(t, err)
}

func TestNewPatchRelease(t *testing.T) { // nolint: dupl
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	releaseBranch := "release-1.16"
	co := s.getChangelogOptions("v1.16.3")
	co.Branch = releaseBranch
	co.Dependencies = true

	// When
	cl := changelog.New(co)
	require.Nil(t, cl.Run())

	// Then
	// Verify local results
	fileContains(t, "CHANGELOG-1.16.html", patchReleaseExpectedHTML)
	require.Nil(t, os.RemoveAll("CHANGELOG-1.16.html"))
	for _, x := range []struct {
		branch        string
		commitMessage string
	}{
		{releaseBranch, "Update CHANGELOG/CHANGELOG-1.16.md for v1.16.3"},
		{git.DefaultBranch, "Update directory for v1.16.3 release"},
	} {
		// Switch to the test branch
		require.Nil(t, s.repo.Checkout(x.branch))

		// Verify commit message
		lastCommit := s.lastCommit(t, x.branch)
		require.Contains(t, lastCommit, "Anago GCB <nobody@k8s.io>")
		require.Contains(t, lastCommit, x.commitMessage)

		// Verify changelog contents
		result, err := os.ReadFile(
			filepath.Join(s.repo.Dir(), changelog.RepoChangelogDir, "CHANGELOG-1.16.md"),
		)
		require.Nil(t, err)
		require.Contains(t, string(result), patchReleaseExpectedTOC)
		require.Contains(t, string(result), patchReleaseExpectedContent)
		require.Contains(t, string(result), patchReleaseDeps)
	}
}

func TestNewAlphaRelease(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	co := s.getChangelogOptions("v1.18.0-alpha.3")

	// When
	require.Nil(t, changelog.New(co).Run())

	// Then
	// Verify local results
	fileContains(t, "CHANGELOG-1.18.html", alphaReleaseExpectedHTMLHead)
	fileContains(t, "CHANGELOG-1.18.html", alphaReleaseExpectedHTMLBottom)
	require.Nil(t, os.RemoveAll("CHANGELOG-1.18.html"))

	// Verify commit message
	lastCommit := s.lastCommit(t, git.DefaultBranch)
	require.Contains(t, lastCommit, "Anago GCB <nobody@k8s.io>")
	require.Contains(t, lastCommit, "Update directory for v1.18.0-alpha.3 release")

	// Verify changelog contents
	result, err := os.ReadFile(
		filepath.Join(s.repo.Dir(), changelog.RepoChangelogDir, "CHANGELOG-1.18.md"),
	)
	require.Nil(t, err)
	require.Regexp(t, alphaReleaseExpectedTOC, string(result))
	require.Contains(t, string(result), alphaReleaseExpectedContent)
}

func TestNewAlpha1Release(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	co := s.getChangelogOptions("v1.19.0-alpha.1")

	// When
	require.Nil(t, changelog.New(co).Run())

	// Then
	// Verify local results
	fileContains(t, "CHANGELOG-1.19.html", alpha1ExpectedHTML)
	require.Nil(t, os.RemoveAll("CHANGELOG-1.19.html"))

	// Verify commit message
	lastCommit := s.lastCommit(t, git.DefaultBranch)
	require.Contains(t, lastCommit, "Anago GCB <nobody@k8s.io>")
	require.Contains(t, lastCommit, "Update directory for v1.19.0-alpha.1 release")

	// Verify changelog contents
	result, err := os.ReadFile(
		filepath.Join(s.repo.Dir(), changelog.RepoChangelogDir, "CHANGELOG-1.19.md"),
	)
	require.Nil(t, err)
	require.Regexp(t, alpha1ReleaseExpectedTOC, string(result))
}

func TestNewMinorRelease(t *testing.T) { // nolint: dupl
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	releaseBranch := "release-1.21"
	co := s.getChangelogOptions("v1.21.0")
	co.Branch = releaseBranch

	// Prepare repo
	changelogIter := func(callback func(string)) {
		for i := 18; i < 21; i++ {
			callback(filepath.Join(
				changelog.RepoChangelogDir, fmt.Sprintf("CHANGELOG-1.%d.md", i),
			))
		}
	}

	require.Nil(t, s.repo.Checkout(releaseBranch))
	changelogIter(func(filename string) {
		require.Nil(t,
			os.WriteFile(
				filepath.Join(s.repo.Dir(), filename),
				[]byte("Some content"),
				0644,
			),
		)
		require.Nil(t, s.repo.Add(filename))
	})
	require.Nil(t, s.repo.Commit("Adding other changelog files"))
	require.Nil(t, s.repo.Checkout(git.DefaultBranch))

	// When
	require.Nil(t, changelog.New(co).Run())

	// Then
	// Verify local results
	require.Nil(t, s.repo.Checkout(releaseBranch))
	changelogIter(func(filename string) {
		_, err := os.Stat(filename)
		require.True(t, os.IsNotExist(err))
	})

	fileContains(t, "CHANGELOG-1.21.html", minorReleaseExpectedHTML)
	require.Nil(t, os.RemoveAll("CHANGELOG-1.21.html"))
	for _, x := range []struct {
		branch        string
		commitMessage string
	}{
		{releaseBranch, "Update CHANGELOG/CHANGELOG-1.21.md for v1.21.0"},
		{git.DefaultBranch, "Update directory for v1.21.0 release"},
	} {
		// Switch to the test branch
		require.Nil(t, s.repo.Checkout(x.branch))

		// Verify commit message
		lastCommit := s.lastCommit(t, x.branch)
		require.Contains(t, lastCommit, "Anago GCB <nobody@k8s.io>")
		require.Contains(t, lastCommit, x.commitMessage)

		// Verify changelog contents
		result, err := os.ReadFile(
			filepath.Join(s.repo.Dir(), changelog.RepoChangelogDir, "CHANGELOG-1.21.md"),
		)
		require.Nil(t, err)
		require.Contains(t, string(result), minorReleaseExpectedTOC)
		require.Contains(t, string(result), minorReleaseExpectedContent)
	}
}

func TestNewRCRelease(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	releaseBranch := "release-1.16"
	co := s.getChangelogOptions("v1.16.0-rc.1")
	co.Branch = releaseBranch

	// When
	require.Nil(t, changelog.New(co).Run())

	// Then
	// Verify local results
	require.Nil(t, s.repo.Checkout(releaseBranch))

	result, err := os.ReadFile(
		filepath.Join(s.repo.Dir(), changelog.RepoChangelogDir, "CHANGELOG-1.16.md"),
	)
	require.Nil(t, err)
	require.Contains(t, string(result), rcReleaseExpectedTOC)
}

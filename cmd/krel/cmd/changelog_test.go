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

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/git"
)

func (s *sut) getChangelogOptions(tag string) *changelogOptions {
	return &changelogOptions{
		replayDir: filepath.Join(testDataDir, "changelog-"+tag),
		tag:       tag,
		tars:      ".",
		branch:    git.Master,
	}
}

func fileContains(t *testing.T, file, contains string) {
	require.FileExists(t, file)
	content, err := ioutil.ReadFile(file)
	require.Nil(t, err)
	require.Contains(t, string(content), contains)
	require.Nil(t, os.RemoveAll(file))
}

func TestChangelogNoArgumentsOrFlags(t *testing.T) {
	err := runChangelog(changelogOpts, rootOpts)
	require.NotNil(t, err)
}

func TestNewPatchRelease(t *testing.T) { // nolint: dupl
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	ro := s.getRootOptions()
	ro.nomock = true

	releaseBranch := "release-1.16"
	co := s.getChangelogOptions("v1.16.3")
	co.branch = releaseBranch

	// When
	require.Nil(t, runChangelog(co, ro))

	// Then
	// Verify local results
	fileContains(t, "CHANGELOG-1.16.html", patchReleaseExpectedHTML)
	for _, x := range []struct {
		branch        string
		commitMessage string
	}{
		{releaseBranch, "Update CHANGELOG/CHANGELOG-1.16.md for v1.16.3"},
		{git.Master, "Update directory for v1.16.3 release"},
	} {
		// Switch to the test branch
		require.Nil(t, s.repo.Checkout(x.branch))

		// Verify commit message
		lastCommit := s.lastCommit(t, x.branch)
		require.Contains(t, lastCommit, "Anago GCB <nobody@k8s.io>")
		require.Contains(t, lastCommit, x.commitMessage)

		// Verify changelog contents
		changelog, err := ioutil.ReadFile(
			filepath.Join(s.repo.Dir(), repoChangelogDir, "CHANGELOG-1.16.md"),
		)
		require.Nil(t, err)
		require.Contains(t, string(changelog), patchReleaseExpectedTOC)
		require.Contains(t, string(changelog), patchReleaseExpectedContent)
	}
}

func TestNewAlphaRelease(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)
	ro := s.getRootOptions()
	ro.nomock = true

	// When
	require.Nil(t, runChangelog(s.getChangelogOptions("v1.18.0-alpha.3"), ro))

	// Then
	// Verify local results
	fileContains(t, "CHANGELOG-1.18.html", alphaReleaseExpectedHTML)

	// Verify commit message
	lastCommit := s.lastCommit(t, git.Master)
	require.Contains(t, lastCommit, "Anago GCB <nobody@k8s.io>")
	require.Contains(t, lastCommit, "Update directory for v1.18.0-alpha.3 release")

	// Verify changelog contents
	changelog, err := ioutil.ReadFile(
		filepath.Join(s.repo.Dir(), repoChangelogDir, "CHANGELOG-1.18.md"),
	)
	require.Nil(t, err)
	require.Regexp(t, alphaReleaseExpectedTOC, string(changelog))
	require.Contains(t, string(changelog), alphaReleaseExpectedContent)
}

func TestNewMinorRelease(t *testing.T) { // nolint: dupl
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	ro := s.getRootOptions()
	ro.nomock = true

	releaseBranch := "release-1.17"
	co := s.getChangelogOptions("v1.17.0")
	co.branch = releaseBranch

	// When
	require.Nil(t, runChangelog(co, ro))

	// Then
	// Verify local results
	fileContains(t, "CHANGELOG-1.17.html", minorReleaseExpectedHTML)
	for _, x := range []struct {
		branch        string
		commitMessage string
	}{
		{releaseBranch, "Update CHANGELOG/CHANGELOG-1.17.md for v1.17.0"},
		{git.Master, "Update directory for v1.17.0 release"},
	} {
		// Switch to the test branch
		require.Nil(t, s.repo.Checkout(x.branch))

		// Verify commit message
		lastCommit := s.lastCommit(t, x.branch)
		require.Contains(t, lastCommit, "Anago GCB <nobody@k8s.io>")
		require.Contains(t, lastCommit, x.commitMessage)

		// Verify changelog contents
		changelog, err := ioutil.ReadFile(
			filepath.Join(s.repo.Dir(), repoChangelogDir, "CHANGELOG-1.17.md"),
		)
		require.Nil(t, err)
		require.Contains(t, string(changelog), minorReleaseExpectedTOC)
		require.Contains(t, string(changelog), minorReleaseExpectedContent)
	}
}

func TestNewRCRelease(t *testing.T) {
	// Given
	s := newSUT(t)
	defer s.cleanup(t)

	ro := s.getRootOptions()
	ro.nomock = true

	releaseBranch := "release-1.16"
	co := s.getChangelogOptions("v1.16.0-rc.1")
	co.branch = releaseBranch

	// Prepare repo
	changelogIter := func(callback func(string)) {
		for i := 0; i < 6; i++ {
			callback(filepath.Join(
				repoChangelogDir, fmt.Sprintf("CHANGELOG-1.1%d.md", i),
			))
		}
	}

	require.Nil(t, s.repo.Checkout(releaseBranch))
	changelogIter(func(filename string) {
		require.Nil(t,
			ioutil.WriteFile(
				filepath.Join(s.repo.Dir(), filename),
				[]byte("Some content"),
				0644,
			),
		)
		require.Nil(t, s.repo.Add(filename))
	})
	require.Nil(t, s.repo.Commit("Adding other changelog files"))
	require.Nil(t, s.repo.Checkout(git.Master))

	// When
	require.Nil(t, runChangelog(co, ro))

	// Then
	// Verify local results
	require.Nil(t, s.repo.Checkout(releaseBranch))
	changelogIter(func(filename string) {
		_, err := os.Stat(filename)
		require.True(t, os.IsNotExist(err))
	})
	changelog, err := ioutil.ReadFile(
		filepath.Join(s.repo.Dir(), repoChangelogDir, "CHANGELOG-1.16.md"),
	)
	require.Nil(t, err)
	require.Contains(t, string(changelog), rcReleaseExpectedTOC)
}

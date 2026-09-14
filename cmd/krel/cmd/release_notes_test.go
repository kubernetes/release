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

package cmd

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func setGitHubToken(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_TOKEN", "fake-token-for-testing")
}

func TestValidateRerunOpts_RequiredFlags(t *testing.T) {
	setGitHubToken(t)
	t.Run("missing draftPRSourceFork", func(t *testing.T) {
		opts := &releaseNotesOptions{
			tag:               "v1.32.0",
			draftPRSourceFork: "",
		}
		err := validateRerunOpts(opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--draft-pr-source-fork is required")
	})

	t.Run("missing tag", func(t *testing.T) {
		opts := &releaseNotesOptions{
			tag:               "",
			draftPRSourceFork: "myorg/sig-release",
		}
		err := validateRerunOpts(opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--tag is required")
	})
}

func TestValidateRerunOpts_ForkNormalization(t *testing.T) {
	setGitHubToken(t)
	t.Run("org-only source fork gets /sig-release appended", func(t *testing.T) {
		opts := &releaseNotesOptions{
			tag:               "v1.32.0",
			draftPRSourceFork: "myorg",
		}
		err := validateRerunOpts(opts)
		require.NoError(t, err)
		require.Equal(t, "myorg/sig-release", opts.draftPRSourceFork)
	})

	t.Run("full org/repo source fork unchanged", func(t *testing.T) {
		opts := &releaseNotesOptions{
			tag:               "v1.32.0",
			draftPRSourceFork: "myorg/myrepo",
		}
		err := validateRerunOpts(opts)
		require.NoError(t, err)
		require.Equal(t, "myorg/myrepo", opts.draftPRSourceFork)
	})

	t.Run("org-only push fork gets /sig-release appended", func(t *testing.T) {
		opts := &releaseNotesOptions{
			tag:               "v1.32.0",
			draftPRSourceFork: "myorg/sig-release",
			draftPRPushFork:   "pushorg",
		}
		err := validateRerunOpts(opts)
		require.NoError(t, err)
		require.Equal(t, "pushorg/sig-release", opts.draftPRPushFork)
	})
}

func TestValidateRerunOpts_BranchDefaults(t *testing.T) {
	setGitHubToken(t)
	t.Run("empty source branch defaults to release-notes-draft-<tag>", func(t *testing.T) {
		opts := &releaseNotesOptions{
			tag:                 "v1.32.0",
			draftPRSourceFork:   "myorg/sig-release",
			draftPRSourceBranch: "",
		}
		err := validateRerunOpts(opts)
		require.NoError(t, err)
		require.Equal(t, "release-notes-draft-v1.32.0", opts.draftPRSourceBranch)
	})

	t.Run("empty push branch defaults to source branch", func(t *testing.T) {
		opts := &releaseNotesOptions{
			tag:                 "v1.32.0",
			draftPRSourceFork:   "myorg/sig-release",
			draftPRSourceBranch: "my-branch",
			draftPRPushBranch:   "",
		}
		err := validateRerunOpts(opts)
		require.NoError(t, err)
		require.Equal(t, "my-branch", opts.draftPRPushBranch)
	})
}

func TestValidateRerunOpts_ValidOptions(t *testing.T) {
	setGitHubToken(t)

	opts := &releaseNotesOptions{
		tag:               "v1.32.0",
		draftPRSourceFork: "myorg/sig-release",
	}
	err := validateRerunOpts(opts)
	require.NoError(t, err)
}

func TestValidateRerunOpts_DynamicMapsWarning(t *testing.T) {
	setGitHubToken(t)

	tests := []struct {
		name     string
		tag      string
		expected string
	}{
		{"alpha tag", "v1.32.0-alpha.1", "releases/release-1.32/release-notes/maps"},
		{"beta tag", "v1.33.0-beta.0", "releases/release-1.33/release-notes/maps"},
		{"rc tag", "v1.36.0-rc.2", "releases/release-1.36/release-notes/maps"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := logrus.StandardLogger().Out

			var buf bytes.Buffer

			logrus.SetOutput(&buf)
			defer logrus.SetOutput(orig)

			opts := &releaseNotesOptions{
				tag:               tc.tag,
				draftPRSourceFork: "myorg/sig-release",
				mapProviders:      []string{},
			}
			err := validateRerunOpts(opts)
			require.NoError(t, err)
			require.Contains(t, buf.String(), tc.expected)
		})
	}
}

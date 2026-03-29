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
	"testing"

	"github.com/stretchr/testify/require"
)

// setGitHubToken sets a dummy GITHUB_TOKEN for Validate() which checks it early.
func setGitHubToken(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_TOKEN", "fake-token-for-testing")
}

func TestValidateRerun_MutualExclusion(t *testing.T) {
	setGitHubToken(t)

	tests := []struct {
		name            string
		createDraftPR   bool
		createWebsitePR bool
	}{
		{
			name:          "rerun with createDraftPR",
			createDraftPR: true,
		},
		{
			name:            "rerun with createWebsitePR",
			createWebsitePR: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := &releaseNotesOptions{
				rerun:             true,
				tag:               "v1.32.0",
				draftPRSourceFork: "myorg/sig-release",
				createDraftPR:     tc.createDraftPR,
				createWebsitePR:   tc.createWebsitePR,
			}
			releaseNotesOpts = opts
			err := opts.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), "cannot be used with")
		})
	}
}

func TestValidateRerun_RequiredFlags(t *testing.T) {
	setGitHubToken(t)

	t.Run("missing draftPRSourceFork", func(t *testing.T) {
		opts := &releaseNotesOptions{
			rerun:             true,
			tag:               "v1.32.0",
			draftPRSourceFork: "",
		}
		releaseNotesOpts = opts
		err := opts.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "--draft-pr-source-fork is required")
	})

	t.Run("missing tag", func(t *testing.T) {
		opts := &releaseNotesOptions{
			rerun:             true,
			tag:               "",
			draftPRSourceFork: "myorg/sig-release",
		}
		releaseNotesOpts = opts
		err := opts.Validate()
		require.Error(t, err)
		// tag validation happens before rerun validation via TagStringToSemver
		require.Error(t, err)
	})
}

func TestValidateRerun_ForkNormalization(t *testing.T) {
	setGitHubToken(t)

	t.Run("org-only source fork gets /sig-release appended", func(t *testing.T) {
		opts := &releaseNotesOptions{
			rerun:             true,
			tag:               "v1.32.0",
			draftPRSourceFork: "myorg",
		}
		releaseNotesOpts = opts
		err := opts.Validate()
		require.NoError(t, err)
		require.Equal(t, "myorg/sig-release", opts.draftPRSourceFork)
	})

	t.Run("full org/repo source fork unchanged", func(t *testing.T) {
		opts := &releaseNotesOptions{
			rerun:             true,
			tag:               "v1.32.0",
			draftPRSourceFork: "myorg/myrepo",
		}
		releaseNotesOpts = opts
		err := opts.Validate()
		require.NoError(t, err)
		require.Equal(t, "myorg/myrepo", opts.draftPRSourceFork)
	})

	t.Run("org-only push fork gets /sig-release appended", func(t *testing.T) {
		opts := &releaseNotesOptions{
			rerun:             true,
			tag:               "v1.32.0",
			draftPRSourceFork: "myorg/sig-release",
			draftPRPushFork:   "pushorg",
		}
		releaseNotesOpts = opts
		err := opts.Validate()
		require.NoError(t, err)
		require.Equal(t, "pushorg/sig-release", opts.draftPRPushFork)
	})
}

func TestValidateRerun_BranchDefaults(t *testing.T) {
	setGitHubToken(t)

	t.Run("empty source branch defaults to release-notes-draft-<tag>", func(t *testing.T) {
		opts := &releaseNotesOptions{
			rerun:               true,
			tag:                 "v1.32.0",
			draftPRSourceFork:   "myorg/sig-release",
			draftPRSourceBranch: "",
		}
		releaseNotesOpts = opts
		err := opts.Validate()
		require.NoError(t, err)
		require.Equal(t, "release-notes-draft-v1.32.0", opts.draftPRSourceBranch)
	})

	t.Run("empty push branch defaults to source branch", func(t *testing.T) {
		opts := &releaseNotesOptions{
			rerun:               true,
			tag:                 "v1.32.0",
			draftPRSourceFork:   "myorg/sig-release",
			draftPRSourceBranch: "my-branch",
			draftPRPushBranch:   "",
		}
		releaseNotesOpts = opts
		err := opts.Validate()
		require.NoError(t, err)
		require.Equal(t, "my-branch", opts.draftPRPushBranch)
	})
}

func TestValidateRerun_ValidOptions(t *testing.T) {
	setGitHubToken(t)

	opts := &releaseNotesOptions{
		rerun:             true,
		tag:               "v1.32.0",
		draftPRSourceFork: "myorg/sig-release",
	}
	releaseNotesOpts = opts
	err := opts.Validate()
	require.NoError(t, err)
}

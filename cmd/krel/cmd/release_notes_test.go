/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func commandInit() {
	rootCmd.ResetCommands()
	releaseNotesCmd.ResetCommands()
}

func initialize() *cobra.Command {
	var c = rootCmd
	Init()
	commandInit()
	return c
}

func TestRunReleaseNotes(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")

	for _, tc := range []struct {
		name        string
		args        []string
		shouldError bool
		output      string
		err         error
	}{
		{
			name:        "should show release-notes help",
			args:        []string{"release-notes"},
			shouldError: false,
			output:      "The 'release-notes' subcommand of krel has been developed to:",
			err:         nil,
		},
		{
			name:        "should fail since --fork flag was not specified",
			args:        strings.Split("release-notes --create-draft-pr", " "),
			shouldError: true,
			output:      "",
			err:         fmt.Errorf("validating command line options: cannot generate the Release Notes PR without --fork"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)
			buf := new(bytes.Buffer)

			c := initialize()
			c.AddCommand(releaseNotesCmd)

			c.SetOut(buf)
			c.SetErr(buf)
			c.SetArgs(tc.args)
			err := c.Execute()

			output := buf.String()

			if tc.shouldError {
				require.NotNil(t, err)
				assert.Contains(t, output, tc.output)
				assert.Contains(t, err.Error(), tc.err.Error())
			} else {
				require.Nil(t, err)
				assert.Contains(t, output, tc.output)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Validate if GITHUB_TOKEN is set
	err := releaseNotesOpts.Validate()
	require.NotNil(t, err)

	t.Setenv("GITHUB_TOKEN", "token")

	for _, tc := range []struct {
		provided    *releaseNotesOptions
		shouldError bool
	}{
		{ // success
			provided: &releaseNotesOptions{
				tag:             "v1.25.0-alpha.2",
				createDraftPR:   true,
				createWebsitePR: true,
				userFork:        "https://github.com/user-fork/kubernetes",
			},
			shouldError: false,
		},
		{ // invalid tag
			provided: &releaseNotesOptions{
				tag: "",
			},
			shouldError: true,
		},
		{ // invalid user fork
			provided: &releaseNotesOptions{
				tag:             "v1.25.0-alpha.2",
				createDraftPR:   true,
				createWebsitePR: true,
				userFork:        "",
			},
			shouldError: true,
		},
	} {
		err := tc.provided.Validate()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

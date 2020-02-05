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

package internal_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/patch/internal"
	"k8s.io/release/pkg/patch/internal/internalfakes"
	it "k8s.io/release/pkg/patch/internal/testing"
	"k8s.io/utils/exec"
)

func TestReleaseNoter(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		releaseToolsDir string
		k8sDir          string
		githubToken     string

		commandOutput []byte
		commandErr    error

		expectedCommandPath string
		expectedErr         string
		expectedOutput      string
	}{
		"happy path": {
			k8sDir:              "/some/dir/k8s",
			releaseToolsDir:     "/some/dir/release",
			githubToken:         "some github token",
			commandOutput:       []byte("some output"),
			expectedCommandPath: "/some/dir/release/relnotes",
			expectedOutput:      "some output",
		},
		"when the command returns an error, the error bubbles up": {
			commandErr:          fmt.Errorf("some random error"),
			expectedErr:         "some random error",
			expectedCommandPath: abs(t, "relnotes"),
		},
		"when the release dir is a relative path": {
			releaseToolsDir:     "../release",
			expectedCommandPath: abs(t, "../release/relnotes"),
		},
		"when the k8s dir is a relative path": {
			k8sDir:              "../k8s",
			expectedCommandPath: abs(t, "relnotes"),
		},
	}

	for name, tc := range tests {
		tc := tc

		it.Run(t, name, func(t *testing.T) {
			command := &internalfakes.FakeCmd{}
			command.OutputReturns(tc.commandOutput, tc.commandErr)

			rn := &internal.ReleaseNoter{
				K8sDir:          tc.k8sDir,
				ReleaseToolsDir: tc.releaseToolsDir,
				GithubToken:     tc.githubToken,
				CommandCreator: func(exe string, args ...string) exec.Cmd {
					require.Equal(t, "bash", exe)
					require.Contains(t, args[1], tc.expectedCommandPath)
					return command
				},
			}

			output, err := rn.GetMarkdown()
			it.CheckErrSub(t, err, tc.expectedErr)
			require.Equal(t, tc.expectedOutput, output, "output")
			require.Equal(t, tc.k8sDir, command.SetDirArgsForCall(0), "Command#SetDir arg")
			require.Contains(t, command.SetEnvArgsForCall(0), "GITHUB_TOKEN="+tc.githubToken)
		})
	}
}

func abs(t *testing.T, path string) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cannot determine current working directory")
	}
	return filepath.Join(cwd, path)
}

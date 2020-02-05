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
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/patch/internal"
	"k8s.io/release/pkg/patch/internal/internalfakes"
	it "k8s.io/release/pkg/patch/internal/testing"
	"k8s.io/utils/exec"
)

func TestWorkspace(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		k8sRepoPath string
		cmdOutput   []byte
		cmdErr      error

		expectedStatuses map[string]string
		expectedErrMsg   string
	}{
		"happy path": {
			k8sRepoPath: "some dir",
			cmdOutput: []byte("" +
				"\n" +
				"blipp blapp\n" +
				"\n" +
				"foo bar baz\n" +
				"zap   zip   \n" +
				"\n",
			),
			expectedStatuses: map[string]string{
				"zap":   "  zip   ", // not sure about that, should leading or trailing spaces be ignored?
				"foo":   "bar baz",
				"blipp": "blapp",
			},
		},
		"when the command errors, return the error": {
			cmdErr:         fmt.Errorf("some cmd error"),
			expectedErrMsg: "some cmd error",
		},
	}

	for name, tc := range tests {
		tc := tc

		it.Run(t, name, func(t *testing.T) {
			cmd := &internalfakes.FakeCmd{}
			cmd.OutputReturns(tc.cmdOutput, tc.cmdErr)

			w := &internal.Workspace{
				K8sRepoPath: tc.k8sRepoPath,
				CommandCreator: func(exe string, args ...string) exec.Cmd {
					require.Equalf(t, 0, len(args), "command args count")
					require.Equalf(t, "hack/print-workspace-status.sh", exe, "command executable")
					return cmd
				},
			}

			statuses, err := w.Status()
			it.CheckErrSub(t, err, tc.expectedErrMsg)
			require.Equal(t, 1, cmd.OutputCallCount(), "Command#Output call count")
			require.Equal(t, tc.k8sRepoPath, cmd.SetDirArgsForCall(0), "Command#SetDir args")
			require.Equal(t, tc.expectedStatuses, statuses)
		})
	}
}

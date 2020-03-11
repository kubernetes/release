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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/patch/internal"
	"k8s.io/release/pkg/patch/internal/internalfakes"
	it "k8s.io/release/pkg/testing"
	"k8s.io/utils/exec"
)

func TestFormatter(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		cmdOutput []byte
		cmdErr    error
		content   string
		title     string
		style     string

		expectedErr    string
		expectedOutput string
	}{
		"happy path": {
			content:        "some input",
			title:          "some title",
			style:          "some additional style",
			cmdOutput:      []byte("converted content"),
			expectedOutput: "converted content",
		},
		"when the command returns an error, the error bubbles up": {
			cmdErr:      fmt.Errorf("some cmd error"),
			expectedErr: "some cmd error",
		},
	}

	for name, tc := range tests {
		tc := tc

		it.Run(t, name, func(t *testing.T) {
			cmd := &internalfakes.FakeCmd{}
			cmd.OutputReturns(tc.cmdOutput, tc.cmdErr)

			f := &internal.Formatter{
				Style: tc.style,
				CommandCreator: func(exe string, args ...string) exec.Cmd {
					require.Equal(t, "bash", exe)
					require.Contains(t, args[1], "pandoc")
					return cmd
				},
			}

			output, err := f.MarkdownToHTML(tc.content, tc.title)
			it.CheckErrSub(t, err, tc.expectedErr)
			require.Equal(t, 1, cmd.OutputCallCount(), "Command#Output call count")

			require.Equal(t, tc.expectedOutput, output, "output")

			require.Contains(t, cmd.SetEnvArgsForCall(0), "STYLE="+tc.style)
			require.Contains(t, cmd.SetEnvArgsForCall(0), "TITLE="+tc.title)

			r := cmd.SetStdinArgsForCall(0)
			b, err := ioutil.ReadAll(r)
			require.NoError(t, err)
			require.EqualValues(t, string(b), tc.content)
		})
	}
}

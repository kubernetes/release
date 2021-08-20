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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFailRootCommand(t *testing.T) {
	err := rootCmd.Execute()
	require.Error(t, err)
}

const expectedPatchOut = `### Timeline

### 1.18

Next patch release is **1.18.4**

End of Life for **1.18** is **TBD**

| PATCH RELEASE | CHERRY PICK DEADLINE | TARGET DATE | NOTE |
|---------------|----------------------|-------------|------|
| 1.18.4        | 2020-06-12           | 2020-06-17  |      |
| 1.18.3        | 2020-05-15           | 2020-05-20  |      |
| 1.18.2        | 2020-04-13           | 2020-04-16  |      |

### 1.17

Next patch release is **1.17.7**

End of Life for **1.17** is **TBD**

| PATCH RELEASE | CHERRY PICK DEADLINE | TARGET DATE | NOTE |
|---------------|----------------------|-------------|------|
| 1.17.7        | 2020-06-12           | 2020-06-17  |      |
| 1.17.6        | 2020-05-15           | 2020-05-20  |      |

### 1.16

Next patch release is **1.16.11**

End of Life for **1.16** is **TBD**

| PATCH RELEASE | CHERRY PICK DEADLINE | TARGET DATE | NOTE |
|---------------|----------------------|-------------|------|
| 1.16.11       | 2020-06-12           | 2020-06-17  |      |
| 1.16.10       | 2020-05-15           | 2020-05-20  |      |
| 1.16.9        | 2020-04-13           | 2020-04-16  |      |
`

const expectedReleaseOut = `# Kubernetes 1.23

#### Links

* [This document](https://git.k8s.io/sig-release/releases/release-1.23/README.md)

* [Release Team](https://groups.google.com/a/kubernetes.io/g/release-team)

* [Meeting Minutes](http://bit.ly/k8s123-releasemtg)

#### Tracking docs

* [Enhancements Tracking Sheet](https://bit.ly/k8s123-enhancements)

* [Feature blog Tracking Sheet](TBD)

* [kubernetes/sig-release v1.23 milestone](https://github.com/kubernetes/kubernetes/milestone/56)

#### Guides

* [Targeting Issues and PRs to This Milestone](https://git.k8s.io/community/contributors/devel/sig-release/release.md)

* [Triaging and Escalating Test Failures](https://git.k8s.io/community/contributors/devel/sig-testing/testing.md#troubleshooting-a-failure)

## Timeline

|          **WHAT**           |      **WHO**      |      **WHEN**       | **WEEK** |  **CI SIGNAL**  |  |
|-----------------------------|-------------------|---------------------|----------|-----------------|--|
| Start of Release Cycle      | Lead              | Mon August 23, 2021 | week 1   | master-blocking |  |
| Start Enhancements Tracking | Enhancements Lead | Mon August 23, 2021 | week 1   |                 |  |
| 1.23.0-alpha.1 released     | Branch Manager    | Wed August 25, 2021 | week 1   |                 |  |
| Schedule finalized          | Lead              | Thu August 26, 2021 | week 1   |                 |  |
`

func TestRun(t *testing.T) {
	testcases := []struct {
		name    string
		options *options
		expect  func(error, string)
	}{
		{
			name: "should parse successfully-patch type",
			options: &options{
				configPath: "testdata/schedule.yaml",
				typeFile:   "patch",
			},
			expect: func(err error, out string) {
				// checks the error of run func call
				require.Nil(t, err)

				// compare the output generated with the expected
				outFile, errFile := os.ReadFile(out)
				require.NoError(t, errFile)
				require.Equal(t, string(outFile), expectedPatchOut)
			},
		},
		{
			name: "should parse successfully-release type",
			options: &options{
				configPath: "testdata/rel-schedule.yaml",
				typeFile:   "release",
			},
			expect: func(err error, out string) {
				// checks the error of run func call
				require.Nil(t, err)

				// compare the output generated with the expected
				outFile, errFile := os.ReadFile(out)
				require.NoError(t, errFile)
				require.Equal(t, string(outFile), expectedReleaseOut)
			},
		},
		{
			name: "should fail parsing",
			options: &options{
				configPath: "testdata/bad_schedule.yaml",
				typeFile:   "patch",
			},
			expect: func(err error, out string) {
				// checks the error of run func call
				require.NotNil(t, err)

				// should not create the output file
				_, errFile := os.Stat(out)
				require.True(t, os.IsNotExist(errFile))
			},
		},
		{
			name: "should fail parsing",
			options: &options{
				configPath: "testdata/bad_schedule.yaml",
				typeFile:   "release",
			},
			expect: func(err error, out string) {
				// checks the error of run func call
				require.NotNil(t, err)

				// should not create the output file
				_, errFile := os.Stat(out)
				require.True(t, os.IsNotExist(errFile))
			},
		},
	}

	for tcCount, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		tempDir, err := os.MkdirTemp("/tmp", "schedule-test")
		require.Nil(t, err)
		defer os.RemoveAll(tempDir)

		tc.options.outputFile = fmt.Sprintf("%s/output-%d.md", tempDir, tcCount)

		err = run(tc.options)
		tc.expect(err, tc.options.outputFile)
	}
}

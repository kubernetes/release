/*
Copyright 2021 The Kubernetes Authors.

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

package build

import (
	"testing"
)

func TestBuildDirFromRepoRoot(t *testing.T) {
	testCases := []struct {
		name             string
		instance         *Instance
		isBazel          bool
		expectedBuildDir string
	}{
		{
			name: "empty repoRoot",
			instance: &Instance{
				opts: &Options{
					Version: "fakeVersion",
				},
			},
			expectedBuildDir: "_output",
		},
		{
			name: "non-empty repoRoot, bazel",
			instance: &Instance{
				opts: &Options{
					Version:  "fakeVersion",
					RepoRoot: "/fake/repo/root",
				},
			},
			isBazel:          true,
			expectedBuildDir: "/fake/repo/root/bazel-bin/build",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			setupBuildDir(tc.instance, tc.isBazel)
			if tc.instance.opts.BuildDir != tc.expectedBuildDir {
				t.Errorf("buildDir mismatched, got: %v, want: %v", tc.instance.opts.BuildDir, tc.expectedBuildDir)
			}
		})
	}
}

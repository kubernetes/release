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
	t.Parallel()

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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			setupBuildDir(tc.instance)

			if tc.instance.opts.BuildDir != tc.expectedBuildDir {
				t.Errorf("buildDir mismatched, got: %v, want: %v", tc.instance.opts.BuildDir, tc.expectedBuildDir)
			}
		})
	}
}

func TestBucketRoot(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		opts     *Options
		expected string
	}{
		"defaults to the build type": {opts: &Options{CI: true}, expected: "ci"},
		"deprecated GCSRoot":         {opts: &Options{GCSRoot: "old"}, expected: "old"},
		"BucketRoot":                 {opts: &Options{BucketRoot: "new"}, expected: "new"},
		"BucketRoot over GCSRoot":    {opts: &Options{BucketRoot: "new", GCSRoot: "old"}, expected: "new"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := NewInstance(tc.opts).opts
			if opts.BucketRoot != tc.expected || opts.GCSRoot != tc.expected {
				t.Errorf("got BucketRoot %q and GCSRoot %q, want %q", opts.BucketRoot, opts.GCSRoot, tc.expected)
			}
		})
	}
}

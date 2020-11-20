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

package object_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/object"
)

var testGCS = object.NewGCS()

func TestGCSSetOptions(t *testing.T) {
	for _, tc := range []struct {
		name     string
		gcs      *object.GCS
		opt      bool
		expected bool
	}{
		{
			name:     "should be false",
			gcs:      testGCS,
			opt:      false,
			expected: false,
		},
	} {
		t.Logf("test case: %v", tc.name)

		testGCS.SetOptions(
			testGCS.WithConcurrent(tc.opt),
			testGCS.WithRecursive(tc.opt),
			testGCS.WithNoClobber(tc.opt),
			testGCS.WithAllowMissing(tc.opt),
		)

		require.Equal(t, tc.expected, testGCS.Concurrent())
		require.Equal(t, tc.expected, testGCS.Recursive())
		require.Equal(t, tc.expected, testGCS.NoClobber())
		require.Equal(t, tc.expected, testGCS.AllowMissing())
	}
}

// TODO: Add production use cases
func TestGetReleasePath(t *testing.T) {
	for _, tc := range []struct {
		bucket, gcsRoot, version string
		expected                 string
		fast                     bool
		shouldError              bool
	}{
		{ // default CI build
			bucket:      "k8s-release-dev",
			gcsRoot:     "ci",
			expected:    "gs://k8s-release-dev/ci",
			shouldError: false,
		},
		{ // fast CI build
			bucket:      "k8s-release-dev",
			gcsRoot:     "ci",
			version:     "",
			fast:        true,
			expected:    "gs://k8s-release-dev/ci/fast",
			shouldError: false,
		},
		{ // has version
			bucket:      "k8s-release-dev",
			gcsRoot:     "ci",
			version:     "42",
			fast:        true,
			expected:    "gs://k8s-release-dev/ci/fast/42",
			shouldError: false,
		},
	} {
		actual, err := testGCS.GetReleasePath(
			tc.bucket,
			tc.gcsRoot,
			tc.version,
			tc.fast,
		)

		require.Equal(t, tc.expected, actual)

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

// TODO: Add production use cases
func TestGetMarkerPath(t *testing.T) {
	for _, tc := range []struct {
		bucket, gcsRoot string
		expected        string
		shouldError     bool
	}{
		{ // default CI build
			bucket:      "k8s-release-dev",
			gcsRoot:     "ci",
			expected:    "gs://k8s-release-dev/ci",
			shouldError: false,
		},
	} {
		actual, err := testGCS.GetMarkerPath(
			tc.bucket,
			tc.gcsRoot,
		)

		require.Equal(t, tc.expected, actual)

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	for _, tc := range []struct {
		gcsPathParts []string
		expected     string
		shouldError  bool
	}{
		{ // empty parts
			gcsPathParts: []string{},
			expected:     "",
			shouldError:  true,
		},
		{ // one element, but empty string
			gcsPathParts: []string{
				"",
			},
			expected:    "",
			shouldError: true,
		},
		{ // multiple elements, all empty strings
			gcsPathParts: []string{
				"",
				"",
			},
			expected:    "",
			shouldError: true,
		},
		{ // strip `gs://` properly
			gcsPathParts: []string{
				"gs://foo",
			},
			expected:    "gs://foo",
			shouldError: false,
		},
		{ // strip `gs:/` properly
			gcsPathParts: []string{
				"gs://foo/bar",
			},
			expected:    "gs://foo/bar",
			shouldError: false,
		},
		{ // strip `/` properly
			gcsPathParts: []string{
				"/foo/bar",
			},
			expected:    "gs://foo/bar",
			shouldError: false,
		},
		{ // multiple parts
			gcsPathParts: []string{
				"foo",
				"bar",
			},
			expected:    "gs://foo/bar",
			shouldError: false,
		},
		{ // one of the non-zero parts already contains the `gs://` prefix
			gcsPathParts: []string{
				"k8s-release-dev",
				"gs://k8s-release-dev/ci-no-bootstrap/fast/v1.20.0-beta.1.655+d20e3246bade17",
			},
			expected:    "",
			shouldError: true,
		},
	} {
		actual, err := testGCS.NormalizePath(tc.gcsPathParts...)

		require.Equal(t, tc.expected, actual)

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestIsPathNormalized(t *testing.T) {
	for _, tc := range []struct {
		gcsPath  string
		expected bool
	}{
		{ // GCS path (%s) should be prefixed with `gs://`
			gcsPath:  "k8s-release-dev/g",
			expected: false,
		},
		{ // filepath.Join() caused an additional `gs:/` to be included in the path
			gcsPath:  "gs://k8s-release-dev/gs:/k8s-release-dev/ci-no-bootstrap/fast/v1.20.0-beta.1.655+d20e3246bade17",
			expected: false,
		},
		{ // fast CI build
			gcsPath:  "gs://k8s-release-dev/ci/fast",
			expected: true,
		},
	} {
		actual := testGCS.IsPathNormalized(tc.gcsPath)

		require.Equal(t, tc.expected, actual)
	}
}

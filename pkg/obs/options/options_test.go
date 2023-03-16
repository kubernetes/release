/*
Copyright 2023 The Kubernetes Authors.

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

package options

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSuccess(t *testing.T) {
	require.Nil(t, New().Validate())
}

func TestValidateFailureWrongPackage(t *testing.T) {
	sut := New()
	sut.Packages = []string{"wrong"}
	require.NotNil(t, sut.Validate())
}

func TestValidateFailureWrongChannel(t *testing.T) {
	sut := New()
	sut.Channel = "wrong"
	require.NotNil(t, sut.Validate())
}

func TestValidateFailureWrongArchitecture(t *testing.T) {
	sut := New()
	sut.Architectures = []string{"wrong"}
	require.NotNil(t, sut.Validate())
}

func TestValidateFailureNonExistingOutputDir(t *testing.T) {
	sut := New()
	sut.OutputDir = "non-existing-test-dir"
	require.NotNil(t, sut.Validate())
}

func TestIsSupportedSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		input    []string
		check    []string
		expected bool
	}{
		{
			name: "single input",
			input: []string{
				"kubelet",
			},
			check:    supportedPackages,
			expected: true,
		},
		{
			name: "multiple inputs",
			input: []string{
				"release",
				"testing",
			},
			check:    supportedChannels,
			expected: true,
		},
		{
			name:     "no inputs",
			input:    []string{},
			check:    supportedArchitectures,
			expected: true,
		},
	}

	for _, tc := range testcases {
		actual := isSupported(tc.input, tc.check)

		require.Equal(t, tc.expected, actual)
	}
}

func TestIsSupportedFailure(t *testing.T) {
	testcases := []struct {
		name     string
		input    []string
		check    []string
		expected bool
	}{
		{
			name: "some supported, some unsupported",
			input: []string{
				"fakearch",
				"amd64",
			},
			check:    supportedArchitectures,
			expected: true,
		},
	}

	for _, tc := range testcases {
		actual := isSupported(tc.input, tc.check)

		require.NotEqual(t, tc.expected, actual)
	}
}

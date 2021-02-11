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

package options

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	str := "test-string"
	slice := []string{str, str, str}

	sut := New()

	require.Equal(t, BuildDeb, sut.WithBuildType(BuildDeb).BuildType())
	require.Equal(t, str, sut.WithRevision(str).Revision())
	require.Equal(t, str, sut.WithKubeVersion(str).KubeVersion())
	require.Equal(t, str, sut.WithCNIVersion(str).CNIVersion())
	require.Equal(t, str, sut.WithCRIToolsVersion(str).CRIToolsVersion())
	require.Equal(t, slice, sut.WithPackages(slice...).Packages())
	require.Equal(t, slice, sut.WithChannels(slice...).Channels())
	require.Equal(t, slice, sut.WithArchitectures(slice...).Architectures())
	require.Equal(t, str, sut.WithReleaseDownloadLinkBase(str).ReleaseDownloadLinkBase())
	require.Equal(t, str, sut.WithTemplateDir(str).TemplateDir())
	require.Equal(t, true, sut.WithSpecOnly(true).SpecOnly())
}

func TestValidateSuccess(t *testing.T) {
	require.Nil(t, New().Validate())
}

func TestValidateFailureWrongPackage(t *testing.T) {
	require.NotNil(t, New().WithPackages("wrong").Validate())
}

func TestValidateFailureWrongChannel(t *testing.T) {
	require.NotNil(t, New().WithChannels("wrong").Validate())
}

func TestValidateFailureWrongArchitecture(t *testing.T) {
	require.NotNil(t, New().WithArchitectures("wrong").Validate())
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

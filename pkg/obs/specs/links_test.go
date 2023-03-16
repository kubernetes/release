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

package specs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/obs/specs"
)

func TestGetDownloadLinkBaseSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		kubeVersion string
		channel     specs.ChannelType
		expected    string
	}{
		{
			name:        "CI version",
			kubeVersion: "1.18.0-alpha.1.277+2099c00290d262",
			channel:     specs.ChannelNightly,
			expected:    "https://dl.k8s.io/ci/v1.18.0-alpha.1.277+2099c00290d262/bin/linux/",
		},
		{
			name:        "non-CI version",
			kubeVersion: "1.18.0-alpha.1",
			expected:    "https://dl.k8s.io/v1.18.0-alpha.1/bin/linux/",
		},
	}

	sut, _ := newSUT(nil)
	for _, tc := range testcases {
		actual, err := sut.GetDownloadLinkBase(tc.kubeVersion, tc.channel)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetDownloadLinkBaseFailure(t *testing.T) {
	sut, _ := newSUT(nil)
	_, err := sut.GetDownloadLinkBase("", "")
	require.NotNil(t, err)
}

func TestGetCIBuildsDownloadLinkBaseSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		kubeVersion string
		expected    string
	}{
		{
			name:        "CI version",
			kubeVersion: "1.18.0-alpha.1.277+2099c00290d262",
			expected:    "https://dl.k8s.io/ci/v1.18.0-alpha.1.277+2099c00290d262/bin/linux/",
		},
	}

	sut, _ := newSUT(nil)
	for _, tc := range testcases {
		actual, err := sut.GetCIBuildsDownloadLinkBase(tc.kubeVersion)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCIBuildsDownloadLinkBaseFailure(t *testing.T) {
	sut, _ := newSUT(nil)
	_, err := sut.GetCIBuildsDownloadLinkBase("")
	require.NotNil(t, err)
}

func TestGetDefaultReleaseDownloadLinkBaseSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		kubeVersion string
		expected    string
	}{
		{
			name:        "Release version",
			kubeVersion: "1.17.0",
			expected:    "https://dl.k8s.io/v1.17.0/bin/linux/",
		},
		{
			name:        "Pre-release version",
			kubeVersion: "1.18.0-alpha.1",
			expected:    "https://dl.k8s.io/v1.18.0-alpha.1/bin/linux/",
		},
	}

	for _, tc := range testcases {
		actual, err := specs.GetDefaultReleaseDownloadLinkBase(tc.kubeVersion)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetDefaultReleaseDownloadLinkBaseFailure(t *testing.T) {
	_, err := specs.GetDefaultReleaseDownloadLinkBase("")
	require.NotNil(t, err)
}

func TestGetCNIDownloadLinkSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		version  string
		arch     string
		expected string
	}{
		{
			name:     "minimum CNI version",
			version:  "0.8.6",
			arch:     "amd64",
			expected: "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.6/cni-plugins-linux-amd64-v0.8.6.tgz",
		},
		{
			name:     "minimum CNI version",
			version:  "v0.8.6",
			arch:     "amd64",
			expected: "https://storage.googleapis.com/k8s-artifacts-cni/release/v0.8.6/cni-plugins-linux-amd64-v0.8.6.tgz",
		},
	}

	for _, tc := range testcases {
		actual, err := specs.GetCNIDownloadLink(tc.version, tc.arch)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCNIDownloadLinkFailure(t *testing.T) {
	_, err := specs.GetCNIDownloadLink("badversion", "amd64")
	require.NotNil(t, err)
}

func TestGetCRIToolsDownloadLinkSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		version  string
		arch     string
		expected string
	}{
		{
			name:     "cri-tools 1.26.0",
			version:  "1.26.0",
			arch:     "amd64",
			expected: "https://storage.googleapis.com/k8s-artifacts-cri-tools/release/v1.26.0/crictl-v1.26.0-linux-amd64.tar.gz",
		},
		{
			name:     "cri-tools v1.26.0",
			version:  "v1.26.0",
			arch:     "amd64",
			expected: "https://storage.googleapis.com/k8s-artifacts-cri-tools/release/v1.26.0/crictl-v1.26.0-linux-amd64.tar.gz",
		},
	}

	for _, tc := range testcases {
		actual, err := specs.GetCRIToolsDownloadLink(tc.version, tc.arch)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCRIToolsDownloadLinkFailure(t *testing.T) {
	_, err := specs.GetCRIToolsDownloadLink("badversion", "amd64")
	require.NotNil(t, err)
}

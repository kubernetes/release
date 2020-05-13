/*
Copyright 2019 The Kubernetes Authors.

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

package kubepkg

import (
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/github/githubfakes"
	"k8s.io/release/pkg/release/releasefakes"
)

func newSUT() *Client {
	sut := New()

	githubMock := &githubfakes.FakeClient{}
	sut.github.SetClient(githubMock)

	versionMock := &releasefakes.FakeVersionClient{}
	sut.version.SetClient(versionMock)

	return sut
}

func TestGetPackageVersionSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		packageName string
		version     string
		kubeVersion string
		expected    string
	}{
		{
			name:        "Kubernetes version supplied",
			kubeVersion: "1.17.0",
			expected:    "1.17.0",
		},
		{
			name:        "Kubernetes version prefixed",
			kubeVersion: "v1.17.0",
			expected:    "1.17.0",
		},
		{
			name:     "Kubernetes version not supplied",
			expected: "",
		},
		{
			name:        "CNI version",
			packageName: "kubernetes-cni",
			version:     "0.8.3",
			kubeVersion: "1.17.0",
			expected:    "0.8.3",
		},
		{
			name:        "CRI tools version",
			packageName: "cri-tools",
			kubeVersion: "1.17.0",
			expected:    "1.17.0",
		},
	}

	sut := newSUT()
	for _, tc := range testcases {
		actual, err := sut.getPackageVersion(
			&PackageDefinition{
				Name:              tc.packageName,
				Version:           tc.version,
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetPackageVersionFailure(t *testing.T) {
	sut := newSUT()
	_, err := sut.getPackageVersion(nil)
	require.NotNil(t, err)
}

// TODO: Figure out how we want to test success of this function.
//       When channel type is provided, we return a func() (string, error), instead of (string, error).
//       Additionally, those functions have variable output depending on when we run the test cases.
func TestGetKubernetesVersionSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		version     string
		kubeVersion string
		channel     ChannelType
		expected    string
	}{
		{
			name:        "Kubernetes version supplied",
			kubeVersion: "1.17.0",
			expected:    "1.17.0",
		},
	}

	sut := newSUT()
	for _, tc := range testcases {
		actual, err := sut.getKubernetesVersion(
			&PackageDefinition{
				Version:           tc.version,
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetKubernetesVersionFailure(t *testing.T) {
	sut := newSUT()
	_, err := sut.getKubernetesVersion(nil)
	require.NotNil(t, err)
}

func TestGetCNIVersionSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		version     string
		kubeVersion string
		expected    string
	}{
		{
			name:        "CNI version supplied, Kubernetes version < 1.17",
			version:     "0.8.3",
			kubeVersion: "1.16.0",
			expected:    pre117CNIVersion,
		},
		{
			name:        "CNI version supplied, Kubernetes version >= 1.17",
			version:     "0.8.3",
			kubeVersion: "1.17.0",
			expected:    "0.8.3",
		},
		{
			name:        "CNI version not supplied",
			kubeVersion: "1.17.0",
			expected:    minimumCNIVersion,
		},
	}

	for _, tc := range testcases {
		actual, err := getCNIVersion(
			&PackageDefinition{
				Version:           tc.version,
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCNIVersionFailure(t *testing.T) {
	_, err := getCNIVersion(nil)
	require.NotNil(t, err)
}

func TestGetCRIToolsVersionSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		version     string
		kubeVersion string
		expected    string
	}{
		{
			name:     "User-supplied CRI tools version",
			version:  "1.17.0",
			expected: "1.17.0",
		},
		{
			name:        "Pre-release or CI Kubernetes version",
			kubeVersion: "1.18.0-alpha.1",
			expected:    "1.17.0",
		},
	}

	sut := newSUT()
	for _, tc := range testcases {
		actual, err := sut.getCRIToolsVersion(
			&PackageDefinition{
				Version:           tc.version,
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCRIToolsVersionFailure(t *testing.T) {
	sut := newSUT()
	_, err := sut.getCRIToolsVersion(nil)
	require.NotNil(t, err)
}

func TestGetDownloadLinkBaseSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		kubeVersion string
		channel     ChannelType
		expected    string
	}{
		{
			name:        "CI version",
			kubeVersion: "1.18.0-alpha.1.277+2099c00290d262",
			channel:     ChannelNightly,
			expected:    "https://dl.k8s.io/ci/v1.18.0-alpha.1.277+2099c00290d262",
		},
		{
			name:        "non-CI version",
			kubeVersion: "1.18.0-alpha.1",
			expected:    "https://dl.k8s.io/v1.18.0-alpha.1",
		},
	}

	sut := newSUT()
	for _, tc := range testcases {
		actual, err := sut.getDownloadLinkBase(
			&PackageDefinition{
				KubernetesVersion: tc.kubeVersion,
				Channel:           tc.channel,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetDownloadLinkBaseFailure(t *testing.T) {
	sut := newSUT()
	_, err := sut.getDownloadLinkBase(nil)
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
			expected:    "https://dl.k8s.io/ci/v1.18.0-alpha.1.277+2099c00290d262",
		},
	}

	sut := newSUT()
	for _, tc := range testcases {
		actual, err := sut.getCIBuildsDownloadLinkBase(
			&PackageDefinition{
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCIBuildsDownloadLinkBaseFailure(t *testing.T) {
	sut := newSUT()
	_, err := sut.getCIBuildsDownloadLinkBase(nil)
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
			expected:    "https://dl.k8s.io/v1.17.0",
		},
		{
			name:        "Pre-release version",
			kubeVersion: "1.18.0-alpha.1",
			expected:    "https://dl.k8s.io/v1.18.0-alpha.1",
		},
	}

	for _, tc := range testcases {
		actual, err := getDefaultReleaseDownloadLinkBase(
			&PackageDefinition{
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetDefaultReleaseDownloadLinkBaseFailure(t *testing.T) {
	_, err := getDefaultReleaseDownloadLinkBase(nil)
	require.NotNil(t, err)
}

func TestGetDependenciesSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		packageName string
		expected    map[string]string
	}{
		{
			name:        "get kubelet deps",
			packageName: "kubelet",
			expected: map[string]string{
				"kubernetes-cni": "0.7.5",
			},
		},
		{
			name:        "get kubeadm deps",
			packageName: "kubeadm",
			expected: map[string]string{
				"kubelet":        "1.13.0",
				"kubectl":        "1.13.0",
				"kubernetes-cni": "0.7.5",
				"cri-tools":      "1.13.0",
			},
		},
	}

	for _, tc := range testcases {
		actual, err := getDependencies(
			&PackageDefinition{
				Name: tc.packageName,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetDependenciesFailure(t *testing.T) {
	_, err := getDependencies(nil)
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
			name:     "CNI <= 0.7.5",
			version:  "0.7.5",
			arch:     "amd64",
			expected: "https://github.com/containernetworking/plugins/releases/download/v0.7.5/cni-plugins-amd64-v0.7.5.tgz",
		},
		{
			name:     "CNI > 0.8.3",
			version:  "0.8.3",
			arch:     "amd64",
			expected: "https://github.com/containernetworking/plugins/releases/download/v0.8.3/cni-plugins-linux-amd64-v0.8.3.tgz",
		},
	}

	for _, tc := range testcases {
		actual, err := getCNIDownloadLink(
			&PackageDefinition{
				Version: tc.version,
			},
			tc.arch,
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCNIDownloadLinkFailure(t *testing.T) {
	_, err := getCNIDownloadLink(nil, "amd64")
	require.NotNil(t, err)
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
			check:    SupportedPackages,
			expected: true,
		},
		{
			name: "multiple inputs",
			input: []string{
				"release",
				"testing",
			},
			check:    SupportedChannels,
			expected: true,
		},
		{
			name:     "no inputs",
			input:    []string{},
			check:    SupportedArchitectures,
			expected: true,
		},
	}

	for _, tc := range testcases {
		actual := IsSupported(tc.input, tc.check)

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
			check:    SupportedArchitectures,
			expected: true,
		},
	}

	for _, tc := range testcases {
		actual := IsSupported(tc.input, tc.check)

		require.NotEqual(t, tc.expected, actual)
	}
}

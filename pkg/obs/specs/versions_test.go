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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/obs/specs"
	"k8s.io/release/pkg/release"
)

func TestGetKubernetesVersionSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		kubeVersion string
		channel     specs.ChannelType
		expected    string
	}{
		{
			name:        "Kubernetes version supplied",
			kubeVersion: "1.17.0",
			expected:    "1.17.0",
		},
		{
			name:        "Kubernetes version not supplied, release channel",
			kubeVersion: "",
			channel:     specs.ChannelRelease,
			expected:    "1.17.0",
		},
		{
			name:        "Kubernetes version and channel not supplied",
			kubeVersion: "",
			channel:     specs.ChannelRelease,
			expected:    "1.17.0",
		},
		{
			name:        "Kubernetes version not supplied, testing channel",
			kubeVersion: "",
			channel:     specs.ChannelTesting,
			expected:    "1.17.0-rc.0",
		},
		{
			name:        "Kubernetes version not supplied, nightly channel",
			kubeVersion: "",
			channel:     specs.ChannelNightly,
			expected:    "1.18.0-alpha.1.277+2099c00290d262",
		},
	}

	sut, mock := newSUT(nil)
	mock.GetKubeVersionCalls(func(versionType release.VersionType) (string, error) {
		switch versionType {
		case release.VersionTypeStable:
			return "1.17.0", nil
		case release.VersionTypeStablePreRelease:
			return "1.17.0-rc.0", nil
		case release.VersionTypeCILatestCross:
			return "1.18.0-alpha.1.277+2099c00290d262", nil
		default:
			return "", fmt.Errorf("unsupported version type")
		}
	})

	for _, tc := range testcases {
		actual, err := sut.GetKubernetesVersion(tc.kubeVersion, tc.channel)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

// TODO(xmudrii): Figure out should with function fail with an non existing channel.
// func TestGetKubernetesVersionFailure(t *testing.T) {
// 	sut, mock := newSUT(nil)
// 	mock.GetKubeVersionCalls(func(versionType release.VersionType) (string, error) {
// 		switch versionType {
// 		case release.VersionTypeStable:
// 			return "1.17.0", nil
// 		case release.VersionTypeStablePreRelease:
// 			return "1.17.0-rc.0", nil
// 		case release.VersionTypeCILatestCross:
// 			return "1.18.0-alpha.1.277+2099c00290d262", nil
// 		default:
// 			return "", fmt.Errorf("unsupported version type")
// 		}
// 	})
// 	_, err := sut.GetKubernetesVersion("", "non-existing-channel")
// 	require.NotNil(t, err)
// }

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
			version:     "1.1.1",
			kubeVersion: "1.17.0",
			expected:    "1.1.1",
		},
		{
			name:        "CRI tools version",
			packageName: "cri-tools",
			kubeVersion: "1.26.0",
			expected:    "1.26.0",
		},
		{
			name:        "CNI version not supplied",
			packageName: "kubernetes-cni",
			version:     "",
			kubeVersion: "1.17.0",
			expected:    "1.1.1",
		},
		{
			name:        "CRI tools version not supplied",
			packageName: "cri-tools",
			version:     "",
			kubeVersion: "1.17.0",
			expected:    "1.26.0",
		},
	}

	sut, _ := newSUT(nil)
	for _, tc := range testcases {
		actual, err := sut.GetPackageVersion(
			&specs.PackageDefinition{
				Name:    tc.packageName,
				Version: tc.version,
			},
			tc.kubeVersion,
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetPackageVersionFailure(t *testing.T) {
	sut, _ := newSUT(nil)
	_, err := sut.GetPackageVersion(nil, "")
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
				"kubernetes-cni": "1.1.1",
			},
		},
		{
			name:        "get kubeadm deps",
			packageName: "kubeadm",
			expected: map[string]string{
				"kubelet":        "1.13.0",
				"kubectl":        "1.13.0",
				"kubernetes-cni": "1.1.1",
				"cri-tools":      "1.26.0",
			},
		},
	}

	for _, tc := range testcases {
		actual, err := specs.GetDependencies(
			&specs.PackageDefinition{
				Name: tc.packageName,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetDependenciesFailure(t *testing.T) {
	_, err := specs.GetDependencies(nil)
	require.NotNil(t, err)
}

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

package kubepkg_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/kubepkg"
	"k8s.io/release/pkg/kubepkg/kubepkgfakes"
	"k8s.io/release/pkg/kubepkg/options"
)

var err = errors.New("")

func newSUT(opts *options.Options) (*kubepkg.Client, *kubepkgfakes.FakeImpl) {
	if opts == nil {
		opts = options.New()
	}

	sut := kubepkg.New(opts)

	implMock := &kubepkgfakes.FakeImpl{}
	sut.SetImpl(implMock)

	return sut, implMock
}

func sutWithTemplateDir(
	t *testing.T, opts *options.Options, buildType options.BuildType,
) (sut *kubepkg.Client, cleanup func(), mock *kubepkgfakes.FakeImpl) {
	tempDir, err := os.MkdirTemp("", "kubepkg-test-")
	require.Nil(t, err)
	cleanup = func() { require.Nil(t, os.RemoveAll(tempDir)) }

	if opts == nil {
		opts = options.New()
	}
	opts = opts.
		WithTemplateDir(tempDir).
		WithBuildType(buildType).
		WithKubeVersion("v1.18.0")
	sut, mock = newSUT(opts)

	for _, dir := range opts.Packages() {
		pkgPath := filepath.Join(tempDir, string(buildType), dir)
		require.Nil(t, os.MkdirAll(pkgPath, 0755))
	}
	return sut, cleanup, mock
}

func TestConstructBuilds(t *testing.T) {
	sut, cleanup, _ := sutWithTemplateDir(t, nil, options.BuildRpm)
	defer cleanup()

	builds, err := sut.ConstructBuilds()
	require.Nil(t, err)
	require.NotEmpty(t, builds)
	require.Len(t, builds, 5)
	require.Len(t, builds[0].Definitions, 3)
	require.Equal(t, "kubelet", builds[0].Package)
	require.Equal(t, options.BuildRpm, builds[0].Type)
}

func TestWalkBuildsSuccessWithoutArchitectures(t *testing.T) {
	opts := options.New().WithArchitectures().WithSpecOnly(true)
	sut, cleanup, _ := sutWithTemplateDir(t, opts, options.BuildRpm)
	defer cleanup()

	builds, err := sut.ConstructBuilds()
	require.Nil(t, err)

	err = sut.WalkBuilds(builds)
	require.Nil(t, err)
}

func TestWalkBuildsSuccessRPM(t *testing.T) {
	sut, cleanup, _ := sutWithTemplateDir(t, nil, options.BuildRpm)
	defer cleanup()

	builds, err := sut.ConstructBuilds()
	require.Nil(t, err)

	err = sut.WalkBuilds(builds)
	require.Nil(t, err)
}

func TestWalkBuildsSuccessRPMSpecOnly(t *testing.T) {
	opts := options.New().WithSpecOnly(true)
	sut, cleanup, _ := sutWithTemplateDir(t, opts, options.BuildRpm)
	defer cleanup()

	builds, err := sut.ConstructBuilds()
	require.Nil(t, err)

	err = sut.WalkBuilds(builds)
	require.Nil(t, err)
}

func TestWalkBuildsSuccessDeb(t *testing.T) {
	sut, cleanup, _ := sutWithTemplateDir(t, nil, options.BuildDeb)
	defer cleanup()

	builds, err := sut.ConstructBuilds()
	require.Nil(t, err)

	err = sut.WalkBuilds(builds)
	require.Nil(t, err)
}

func TestWalkBuildsFailureReadFileFailed(t *testing.T) {
	sut, cleanup, mock := sutWithTemplateDir(t, nil, options.BuildDeb)
	mock.ReadFileReturns(nil, err)
	defer cleanup()

	builds, err := sut.ConstructBuilds()
	require.Nil(t, err)

	err = sut.WalkBuilds(builds)
	require.NotNil(t, err)
}

func TestWalkBuildsFailureWriteFileFailed(t *testing.T) {
	sut, cleanup, mock := sutWithTemplateDir(t, nil, options.BuildDeb)
	mock.WriteFileReturns(err)
	defer cleanup()

	builds, err := sut.ConstructBuilds()
	require.Nil(t, err)

	err = sut.WalkBuilds(builds)
	require.NotNil(t, err)
}

func TestWalkBuildsFailureDebDPKGFailed(t *testing.T) {
	sut, cleanup, mock := sutWithTemplateDir(t, nil, options.BuildDeb)
	mock.RunSuccessWithWorkDirReturns(err)
	defer cleanup()

	builds, err := sut.ConstructBuilds()
	require.Nil(t, err)

	err = sut.WalkBuilds(builds)
	require.NotNil(t, err)
}

func TestConstructBuildsFailedInvalidTemplateDir(t *testing.T) {
	sut, _ := newSUT(nil)
	builds, err := sut.ConstructBuilds()
	require.NotNil(t, err)
	require.Nil(t, builds)
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
			version:     "0.8.6",
			kubeVersion: "1.17.0",
			expected:    "0.8.6",
		},
		{
			name:        "CRI tools version",
			packageName: "cri-tools",
			kubeVersion: "1.17.0",
			expected:    "1.17.0",
		},
	}

	sut, _ := newSUT(nil)
	for _, tc := range testcases {
		actual, err := sut.GetPackageVersion(
			&kubepkg.PackageDefinition{
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
	sut, _ := newSUT(nil)
	_, err := sut.GetPackageVersion(nil)
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
		channel     kubepkg.ChannelType
		expected    string
	}{
		{
			name:        "Kubernetes version supplied",
			kubeVersion: "1.17.0",
			expected:    "1.17.0",
		},
	}

	sut, _ := newSUT(nil)
	for _, tc := range testcases {
		actual, err := sut.GetKubernetesVersion(
			&kubepkg.PackageDefinition{
				Version:           tc.version,
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetKubernetesVersionFailure(t *testing.T) {
	sut, _ := newSUT(nil)
	_, err := sut.GetKubernetesVersion(nil)
	require.NotNil(t, err)
}

func TestGetCNIVersionSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		cniVersion  string
		kubeVersion string
		expected    string
	}{
		{
			name:        "CNI version supplied",
			cniVersion:  "0.8.7",
			kubeVersion: "1.17.0",
			expected:    "0.8.7",
		},
		{
			name:        "CNI version not supplied",
			kubeVersion: "1.17.0",
			expected:    kubepkg.MinimumCNIVersion,
		},
	}

	for _, tc := range testcases {
		actual, err := kubepkg.GetCNIVersion(
			&kubepkg.PackageDefinition{
				CNIVersion:        tc.cniVersion,
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCNIVersionFailure(t *testing.T) {
	testcases := []struct {
		name       string
		packageDef *kubepkg.PackageDefinition
	}{
		{
			name:       "package definition is nil",
			packageDef: nil,
		},
		{
			name: "CNI version supplied less than minimum allowed CNI version",
			packageDef: &kubepkg.PackageDefinition{
				CNIVersion:        "0.8.3",
				KubernetesVersion: "1.17.0",
			},
		},
	}

	for _, tc := range testcases {
		_, err := kubepkg.GetCNIVersion(tc.packageDef)

		require.NotNil(t, err)
	}
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

	sut, _ := newSUT(nil)
	for _, tc := range testcases {
		actual, err := sut.GetCRIToolsVersion(
			&kubepkg.PackageDefinition{
				Version:           tc.version,
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCRIToolsVersionFailure(t *testing.T) {
	sut, _ := newSUT(nil)
	_, err := sut.GetCRIToolsVersion(nil)
	require.NotNil(t, err)
}

func TestGetDownloadLinkBaseSuccess(t *testing.T) {
	testcases := []struct {
		name        string
		kubeVersion string
		channel     kubepkg.ChannelType
		expected    string
	}{
		{
			name:        "CI version",
			kubeVersion: "1.18.0-alpha.1.277+2099c00290d262",
			channel:     kubepkg.ChannelNightly,
			expected:    "https://dl.k8s.io/ci/v1.18.0-alpha.1.277+2099c00290d262",
		},
		{
			name:        "non-CI version",
			kubeVersion: "1.18.0-alpha.1",
			expected:    "https://dl.k8s.io/v1.18.0-alpha.1",
		},
	}

	sut, _ := newSUT(nil)
	for _, tc := range testcases {
		actual, err := sut.GetDownloadLinkBase(
			&kubepkg.PackageDefinition{
				KubernetesVersion: tc.kubeVersion,
				Channel:           tc.channel,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetDownloadLinkBaseFailure(t *testing.T) {
	sut, _ := newSUT(nil)
	_, err := sut.GetDownloadLinkBase(nil)
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

	sut, _ := newSUT(nil)
	for _, tc := range testcases {
		actual, err := sut.GetCIBuildsDownloadLinkBase(
			&kubepkg.PackageDefinition{
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCIBuildsDownloadLinkBaseFailure(t *testing.T) {
	sut, _ := newSUT(nil)
	_, err := sut.GetCIBuildsDownloadLinkBase(nil)
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
		actual, err := kubepkg.GetDefaultReleaseDownloadLinkBase(
			&kubepkg.PackageDefinition{
				KubernetesVersion: tc.kubeVersion,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetDefaultReleaseDownloadLinkBaseFailure(t *testing.T) {
	_, err := kubepkg.GetDefaultReleaseDownloadLinkBase(nil)
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
				"kubernetes-cni": "0.8.6",
			},
		},
		{
			name:        "get kubeadm deps",
			packageName: "kubeadm",
			expected: map[string]string{
				"kubelet":        "1.13.0",
				"kubectl":        "1.13.0",
				"kubernetes-cni": "0.8.6",
				"cri-tools":      "1.13.0",
			},
		},
	}

	for _, tc := range testcases {
		actual, err := kubepkg.GetDependencies(
			&kubepkg.PackageDefinition{
				Name: tc.packageName,
			},
		)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetDependenciesFailure(t *testing.T) {
	_, err := kubepkg.GetDependencies(nil)
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
	}

	for _, tc := range testcases {
		actual, err := kubepkg.GetCNIDownloadLink(tc.version, tc.arch)

		require.Nil(t, err)
		require.Equal(t, tc.expected, actual)
	}
}

func TestGetCNIDownloadLinkFailure(t *testing.T) {
	_, err := kubepkg.GetCNIDownloadLink("badversion", "amd64")
	require.NotNil(t, err)
}

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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/obs/options"
	"k8s.io/release/pkg/obs/specs"
	"k8s.io/release/pkg/obs/specs/specsfakes"
)

func newSUT(opts *options.Options) (*specs.Client, *specsfakes.FakeImpl) {
	if opts == nil {
		opts = options.New()
	}

	sut := specs.New(opts)

	implMock := &specsfakes.FakeImpl{}
	sut.SetImpl(implMock)

	return sut, implMock
}

//nolint:unparam // TODO(xmudrii): we'll use this later.
func sutWithTemplateDir(
	t *testing.T, opts *options.Options, packageType options.PackageType,
) (sut *specs.Client, cleanup func(), mock *specsfakes.FakeImpl) {
	tempDir, err := os.MkdirTemp("", "obs-test-")
	require.Nil(t, err)
	cleanup = func() { require.Nil(t, os.RemoveAll(tempDir)) }

	if opts == nil {
		opts = options.New()
	}
	opts.TemplateDir = tempDir
	opts.Type = packageType
	opts.KubernetesVersion = "v1.18.0"
	sut, mock = newSUT(opts)

	for _, dir := range opts.Packages {
		pkgPath := filepath.Join(tempDir, string(packageType), dir)
		require.Nil(t, os.MkdirAll(pkgPath, 0o755))
	}
	return sut, cleanup, mock
}

func TestConstructPackageBuilder(t *testing.T) {
	sut, cleanup, _ := sutWithTemplateDir(t, nil, options.PackageRPM)
	defer cleanup()

	pkgBuilder, err := sut.ConstructPackageBuilder()
	require.Nil(t, err)
	require.NotNil(t, pkgBuilder.Architectures)
	require.NotEmpty(t, pkgBuilder.Architectures)
	require.NotEmpty(t, pkgBuilder.Channel)
	require.NotEmpty(t, pkgBuilder.DownloadLinkBase)
	require.NotEmpty(t, pkgBuilder.KubernetesVersion)
	require.NotEmpty(t, pkgBuilder.Type)
}

func TestConstructPackageDefinitions(t *testing.T) {
	sut, cleanup, _ := sutWithTemplateDir(t, nil, options.PackageRPM)
	defer cleanup()

	pkgBuilder, err := sut.ConstructPackageBuilder()
	require.Nil(t, err)
	err = sut.ConstructPackageDefinitions(pkgBuilder)
	require.Nil(t, err)
	require.NotNil(t, pkgBuilder.Definitions)
	require.NotEmpty(t, pkgBuilder.Definitions)
	require.Equal(t, 5, len(pkgBuilder.Definitions))
	require.Equal(t, "kubelet", pkgBuilder.Definitions[0].Name)
	require.Equal(t, "1.18.0", pkgBuilder.Definitions[0].Version)
}

func TestWalkBuildsSuccessRPM(t *testing.T) {
	sut, cleanup, _ := sutWithTemplateDir(t, nil, options.PackageRPM)
	defer cleanup()

	pkgBuilder, err := sut.ConstructPackageBuilder()
	require.Nil(t, err)

	err = sut.ConstructPackageDefinitions(pkgBuilder)
	require.Nil(t, err)

	err = sut.BuildSpecs(pkgBuilder)
	require.Nil(t, err)
}

// func TestWalkBuildsSuccessRPMSpecOnly(t *testing.T) {
// 	opts := options.New().WithSpecOnly(true)
// 	sut, cleanup, _ := sutWithTemplateDir(t, opts, options.PackageRpm)
// 	defer cleanup()

// 	builds, err := sut.ConstructBuilds()
// 	require.Nil(t, err)

// 	err = sut.WalkBuilds(builds)
// 	require.Nil(t, err)
// }

// func TestWalkBuildsSuccessDeb(t *testing.T) {
// 	sut, cleanup, _ := sutWithTemplateDir(t, nil, options.PackageDeb)
// 	defer cleanup()

// 	builds, err := sut.ConstructBuilds()
// 	require.Nil(t, err)

// 	err = sut.WalkBuilds(builds)
// 	require.Nil(t, err)
// }

// func TestWalkBuildsFailureReadFileFailed(t *testing.T) {
// 	sut, cleanup, mock := sutWithTemplateDir(t, nil, options.PackageDeb)
// 	mock.ReadFileReturns(nil, err)
// 	defer cleanup()

// 	builds, err := sut.ConstructBuilds()
// 	require.Nil(t, err)

// 	err = sut.WalkBuilds(builds)
// 	require.NotNil(t, err)
// }

// func TestWalkBuildsFailureWriteFileFailed(t *testing.T) {
// 	sut, cleanup, mock := sutWithTemplateDir(t, nil, options.PackageDeb)
// 	mock.WriteFileReturns(err)
// 	defer cleanup()

// 	builds, err := sut.ConstructBuilds()
// 	require.Nil(t, err)

// 	err = sut.WalkBuilds(builds)
// 	require.NotNil(t, err)
// }

// func TestWalkBuildsFailureDebDPKGFailed(t *testing.T) {
// 	sut, cleanup, mock := sutWithTemplateDir(t, nil, options.PackageDeb)
// 	mock.RunSuccessWithWorkDirReturns(err)
// 	defer cleanup()

// 	builds, err := sut.ConstructBuilds()
// 	require.Nil(t, err)

// 	err = sut.WalkBuilds(builds)
// 	require.NotNil(t, err)
// }

// func TestConstructBuildsFailedInvalidTemplateDir(t *testing.T) {
// 	sut, _ := newSUT(nil)
// 	builds, err := sut.ConstructBuilds()
// 	require.NotNil(t, err)
// 	require.Nil(t, builds)
// }

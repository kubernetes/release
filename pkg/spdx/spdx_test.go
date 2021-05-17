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

package spdx_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/spdx"
	"k8s.io/release/pkg/spdx/spdxfakes"
)

var (
	err      = errors.New("synthetic error")
	manifest = &spdx.ArchiveManifest{
		ConfigFilename: "9283479287498237498.json",
		RepoTags:       []string{"image-test:latest"},
		LayerFiles:     []string{"ksjdhfkjsdhfkjsdhf/layer.tar"},
	}
)

func TestPackageFromImageTarball(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*spdxfakes.FakeSpdxImplementation)
		shouldError bool
	}{
		{ // success
			prepare: func(mock *spdxfakes.FakeSpdxImplementation) {
				mock.ExtractTarballTmpReturns("/mock/path", nil)
				mock.ReadArchiveManifestReturns(manifest, nil)
				mock.PackageFromLayerTarBallReturns(&spdx.Package{Name: "test"}, nil)
			},
			shouldError: false,
		},
		{
			prepare: func(mock *spdxfakes.FakeSpdxImplementation) {
				mock.ReadArchiveManifestReturns(manifest, nil)
				mock.ExtractTarballTmpReturns("", err)
			},
			shouldError: true,
		},
		{
			prepare: func(mock *spdxfakes.FakeSpdxImplementation) {
				mock.ReadArchiveManifestReturns(nil, err)
			},
			shouldError: true,
		},
		{
			prepare: func(mock *spdxfakes.FakeSpdxImplementation) {
				mock.ReadArchiveManifestReturns(manifest, nil)
				mock.PackageFromLayerTarBallReturns(nil, err)
			},
			shouldError: true,
		},
	} {
		sut := spdx.NewSPDX()
		sut.Options().AnalyzeLayers = false
		mock := &spdxfakes.FakeSpdxImplementation{}
		tc.prepare(mock)
		sut.SetImplementation(mock)

		dir, err := sut.PackageFromImageTarball("mock.tar", &spdx.TarballOptions{})
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
			require.NotNil(t, dir)
		}
	}
}

func TestExtractTarballTmp(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*spdxfakes.FakeSpdxImplementation)
		shouldError bool
	}{
		{ // success
			prepare: func(mock *spdxfakes.FakeSpdxImplementation) {
				mock.ExtractTarballTmpReturns("/mock/path", nil)
			},
			shouldError: false,
		},
		{ // error
			prepare: func(mock *spdxfakes.FakeSpdxImplementation) {
				mock.ExtractTarballTmpReturns("/mock/path", err)
			},
			shouldError: true,
		},
	} {
		sut := spdx.NewSPDX()
		mock := &spdxfakes.FakeSpdxImplementation{}
		tc.prepare(mock)
		sut.SetImplementation(mock)

		path, err := sut.ExtractTarballTmp("/mock/path")
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.NotEmpty(t, path)
			require.Nil(t, err)
		}
	}
}

func TestPullImagesToArchive(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*spdxfakes.FakeSpdxImplementation)
		shouldError bool
	}{
		{ // success
			prepare: func(mock *spdxfakes.FakeSpdxImplementation) {
				mock.PullImagesToArchiveReturns(nil)
			},
			shouldError: false,
		},
		{ // success
			prepare: func(mock *spdxfakes.FakeSpdxImplementation) {
				mock.PullImagesToArchiveReturns(err)
			},
			shouldError: true,
		},
	} {
		sut := spdx.NewSPDX()
		sut.Options().AnalyzeLayers = false
		mock := &spdxfakes.FakeSpdxImplementation{}
		tc.prepare(mock)
		sut.SetImplementation(mock)

		err := sut.PullImagesToArchive("mock-image:latest", "/tmp")
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

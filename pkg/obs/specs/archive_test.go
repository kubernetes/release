/*
Copyright 2026 The Kubernetes Authors.

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
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/obs/specs"
	"k8s.io/release/pkg/obs/specs/specsfakes"
)

func TestBuildArtifactsArchive(t *testing.T) {
	testcases := []struct {
		name        string
		pkgDef      *specs.PackageDefinition
		shouldErr   bool
		archiveFile string
		prepare     func(mock *specsfakes.FakeImpl)
	}{
		{
			name:        "happy path, empty package",
			shouldErr:   false,
			archiveFile: "_.orig.tar.gz",
			pkgDef:      &specs.PackageDefinition{},
			prepare:     func(*specsfakes.FakeImpl) {},
		},
		{
			name:        "happy path with multiple variations (multiple architectures)",
			shouldErr:   false,
			archiveFile: "kubernetes-cni_1.2.3.orig.tar.gz",
			pkgDef: &specs.PackageDefinition{
				Name:           "kubernetes-cni",
				Version:        "1.2.3",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "amd64",
						Source:       "gs://",
					},
					{
						Architecture: "arm64",
						Source:       "gs://",
					},
				},
			},
			prepare: func(*specsfakes.FakeImpl) {},
		},
		{
			name:      "error if pkgDef is nil",
			shouldErr: true,
			pkgDef:    nil,
			prepare:   func(*specsfakes.FakeImpl) {},
		},
		{
			name:      "throw error when can't download from GCS",
			shouldErr: true,
			pkgDef: &specs.PackageDefinition{
				Name:           "cri-o",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "",
						Source:       "gs://",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.GCSCopyToLocalReturns(err)
			},
		},
		{
			name:      "throw error when can't extract from the archive",
			shouldErr: true,
			pkgDef: &specs.PackageDefinition{
				Name:           "cri-o",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "",
						Source:       "gs://",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.ExtractReturns(err)
			},
		},
		{
			name:      "throw error when can't remove the archive",
			shouldErr: true,
			pkgDef: &specs.PackageDefinition{
				Name:           "cri-o",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "",
						Source:       "gs://",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.RemoveFileReturns(err)
			},
		},
		{
			name:      "throw error when can't compress artifacts",
			shouldErr: true,
			pkgDef: &specs.PackageDefinition{
				Name:           "cri-o",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "",
						Source:       "gs://",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.CompressReturns(err)
			},
		},

		{
			name:      "throw error when MkdirAll fails",
			shouldErr: true,
			pkgDef: &specs.PackageDefinition{
				Name:           "cri-o",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "",
						Source:       "gs://",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.MkdirAllReturns(err)
			},
		},
		{
			name:      "throw error when can't RemoveAll after compressing",
			shouldErr: true,
			pkgDef: &specs.PackageDefinition{
				Name:           "kubernetes-cni",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "s390x",
						Source:       "gs://",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.RemoveAllReturns(err)
			},
		},
		{
			name:      "URL download with GetRequest error",
			shouldErr: true,
			pkgDef: &specs.PackageDefinition{
				Name:           "something-else",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "ppc64le",
						Source:       "https://example.com/artifact.tar",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.GetRequestReturns(nil, err)
			},
		},
		{
			name:      "URL download with non-200 status code",
			shouldErr: true,
			pkgDef: &specs.PackageDefinition{
				Name:           "something-else",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "arm64",
						Source:       "https://example.com/artifact.tar",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				resp := &http.Response{StatusCode: http.StatusNotFound}
				mock.GetRequestReturns(resp, nil)
			},
		},
		{
			name:      "URL download with CreateFile error",
			shouldErr: true,
			pkgDef: &specs.PackageDefinition{
				Name:           "something-else",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "ppc64le",
						Source:       "https://example.com/artifact.tar",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.CreateFileReturns(nil, err)
			},
		},
		{
			name:        "don't throw an error when download directory exists",
			shouldErr:   false,
			archiveFile: "kubernetes-cni_0.0.2.orig.tar.gz",
			pkgDef: &specs.PackageDefinition{
				Name:           "kubernetes-cni",
				Version:        "0.0.2",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "amd64",
						Source:       "gs://",
					},
				},
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.MkdirAllReturns(err)
				mock.IsExistReturns(true)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			options := &specs.Options{}
			sut := specs.New(options)

			mock := &specsfakes.FakeImpl{}

			tc.prepare(mock)

			sut.SetImpl(mock)

			err := sut.BuildArtifactsArchive(tc.pkgDef)

			if tc.shouldErr {
				require.Error(t, err)
			} else {
				// check archive destination
				if mock.CompressCallCount() > 0 {
					actualArchiveFile, _, _ := mock.CompressArgsForCall(0)
					expectedArchiveFile := filepath.Join(tc.pkgDef.SpecOutputPath, tc.archiveFile)
					require.Equal(t, expectedArchiveFile, actualArchiveFile)
				}

				require.NoError(t, err)
			}
		})
	}
}

func TestBuildArtifactsArchiveDownloadPaths(t *testing.T) {
	testcases := []struct {
		name             string
		pkgDef           *specs.PackageDefinition
		destinationPaths []string
	}{
		{
			name:             "check download path for kubernetes-cni",
			destinationPaths: []string{"kubernetes-cni/kubernetes-cni.tar.gz"},
			pkgDef: &specs.PackageDefinition{
				Name:           "kubernetes-cni",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "",
						Source:       "gs://",
					},
				},
			},
		},
		{
			name:             "check download path for cri-tools",
			destinationPaths: []string{"cri-tools/cri-tools.tar.gz"},
			pkgDef: &specs.PackageDefinition{
				Name:           "cri-tools",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "",
						Source:       "gs://",
					},
				},
			},
		},
		{
			name:             "check download path for cri-o",
			destinationPaths: []string{"cri-o/cri-o.tar.gz"},
			pkgDef: &specs.PackageDefinition{
				Name:           "cri-o",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "",
						Source:       "gs://",
					},
				},
			},
		},
		{
			name:             "check download path for all other packages",
			destinationPaths: []string{"some-other-package/some-other-package"},
			pkgDef: &specs.PackageDefinition{
				Name:           "some-other-package",
				Version:        "0.0.1",
				SpecOutputPath: t.TempDir(),
				Variations: []specs.PackageVariation{
					{
						Architecture: "",
						Source:       "gs://",
					},
				},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			options := &specs.Options{}
			sut := specs.New(options)

			mock := &specsfakes.FakeImpl{}
			sut.SetImpl(mock)

			err := sut.BuildArtifactsArchive(tc.pkgDef)
			require.NoError(t, err)

			for i, expectedPath := range tc.destinationPaths {
				_, actualDestinationPath := mock.GCSCopyToLocalArgsForCall(i)
				require.Equal(t, filepath.Join(tc.pkgDef.SpecOutputPath, expectedPath), actualDestinationPath)
			}
		})
	}
}

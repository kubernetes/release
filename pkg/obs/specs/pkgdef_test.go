/*
Copyright 2025 The Kubernetes Authors.

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
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/obs/metadata"
	"k8s.io/release/pkg/obs/specs"
	"k8s.io/release/pkg/obs/specs/specsfakes"
)

func TestConstructPackageDefinition(t *testing.T) {
	testcases := []struct {
		prepare   func(mock *specsfakes.FakeImpl)
		options   *specs.Options
		channel   string
		version   string
		shouldErr bool
	}{
		{ // happy path for core package with version specified
			shouldErr: false,
			channel:   "release",
			version:   "1.0.0",
			options: &specs.Options{
				Package:          "kubeadm",
				Version:          "v1.0.0",
				SpecTemplatePath: newSpecPath(t),
				SpecOutputPath:   t.TempDir(),
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mockVersion(mock, "1.0.0")
				mock.LoadPackageMetadataReturns(createKubeadmMetadata(), nil)
			},
		},
		{ // specifying non-parsable version should cause
			//  failure for core packages
			shouldErr: true,
			options: &specs.Options{
				Package:          "kubeadm",
				Version:          "not-parsable",
				SpecTemplatePath: newSpecPath(t),
				SpecOutputPath:   t.TempDir(),
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.TrimTagPrefixReturns("not-parsable")

				version, _ := semver.New("0.0.0")
				mock.TagStringToSemverReturns(*version, err)
			},
		},
		{
			// happy path for core package with channel specified
			shouldErr: false,
			channel:   "nightly",
			version:   "1.0.0",
			options: &specs.Options{
				Package:          "kubeadm",
				Channel:          "nightly",
				SpecTemplatePath: newSpecPath(t),
				SpecOutputPath:   t.TempDir(),
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mockVersion(mock, "1.0.0")
				mock.LoadPackageMetadataReturns(createKubeadmMetadata(), nil)
			},
		},
		{
			// can't get a version for core package with channel specified
			shouldErr: true,
			channel:   "nightly",
			version:   "1.0.0",
			options: &specs.Options{
				Package:          "kubeadm",
				Channel:          "nightly",
				SpecTemplatePath: newSpecPath(t),
				SpecOutputPath:   t.TempDir(),
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.GetKubeVersionReturns("1.0.0", err)
				mock.LoadPackageMetadataReturns(createKubeadmMetadata(), nil)
			},
		},
		{ // happy path for cri-tools package
			shouldErr: false,
			version:   "2.0.0",
			options: &specs.Options{
				Package:          "cri-tools",
				SpecTemplatePath: newSpecPath(t),
				SpecOutputPath:   t.TempDir(),
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mockVersion(mock, "2.0.0")
				mock.LoadPackageMetadataReturns(createCRIToolsMetadata(), nil)
				mock.HeadRequestReturns(
					&http.Response{
						StatusCode: http.StatusOK,
						Request:    &http.Request{URL: &url.URL{Scheme: "https", Host: "example.com", Path: "/cri-tools"}},
						Body:       io.NopCloser(strings.NewReader("response body")),
					}, nil)
			},
		},
		{ // HTTP head request error should cause failure
			// for cri-tools package
			shouldErr: true,
			version:   "2.0.0",
			options: &specs.Options{
				Package:          "cri-tools",
				SpecTemplatePath: newSpecPath(t),
				SpecOutputPath:   t.TempDir(),
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mockVersion(mock, "2.0.0")
				mock.LoadPackageMetadataReturns(createCRIToolsMetadata(), nil)
				mock.HeadRequestReturns(&http.Response{}, err)
			},
		},
		{ // happy path for kubernetes-cni package
			shouldErr: false,
			version:   "1.0.0",
			options: &specs.Options{
				Package:          "kubernetes-cni",
				SpecTemplatePath: newSpecPath(t),
				SpecOutputPath:   t.TempDir(),
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mockVersion(mock, "1.0.0")
				mock.LoadPackageMetadataReturns(createCNIMetadata(), nil)
				mock.HeadRequestReturns(
					&http.Response{
						StatusCode: http.StatusOK,
						Request:    &http.Request{URL: &url.URL{Scheme: "https", Host: "example.com", Path: "/kubernetes-cni"}},
						Body:       io.NopCloser(strings.NewReader("response body")),
					}, nil)
			},
		},
		{ // HTTP head request error should cause failure
			// for kubernetes-cni package
			shouldErr: true,
			version:   "1.0.0",
			options: &specs.Options{
				Package:          "kubernetes-cni",
				SpecTemplatePath: newSpecPath(t),
				SpecOutputPath:   t.TempDir(),
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mockVersion(mock, "1.0.0")
				mock.LoadPackageMetadataReturns(createCNIMetadata(), nil)
				mock.HeadRequestReturns(&http.Response{}, err)
			},
		},
	}
	for _, tc := range testcases {
		createTestMetadataFile(t, tc.options.SpecTemplatePath, "")
		sut := specs.New(tc.options)

		mock := &specsfakes.FakeImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		pkgDef, err := sut.ConstructPackageDefinition()

		if tc.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotNil(t, pkgDef)
			require.Equal(t, tc.options.Package, pkgDef.Name)
			require.Equal(t, tc.channel, pkgDef.Channel)
			require.Equal(t, tc.version, pkgDef.Version)
		}
	}
}

func mockVersion(mock *specsfakes.FakeImpl, version string) {
	mock.GetKubeVersionReturns(version, nil)
	mock.TrimTagPrefixReturns(version)
	semverVersion, _ := semver.New(version)
	mock.TagStringToSemverReturns(*semverVersion, nil)
}

func createTestMetadataFile(t *testing.T, dir, content string) string {
	metadataFile := filepath.Join(dir, "metadata.yaml")
	require.NoError(t, os.WriteFile(metadataFile, []byte(content), 0o600))

	return metadataFile
}

func createKubeadmMetadata() metadata.PackageMetadataList {
	return metadata.PackageMetadataList{
		"kubeadm": {
			{
				VersionConstraint: ">=1.0.0",
			},
		},
	}
}

func createCRIToolsMetadata() metadata.PackageMetadataList {
	return metadata.PackageMetadataList{
		"cri-tools": {
			{
				VersionConstraint: ">=2.0.0",
			},
		},
	}
}

func createCNIMetadata() metadata.PackageMetadataList {
	return metadata.PackageMetadataList{
		"kubernetes-cni": {
			{
				VersionConstraint: ">=1.0.0",
			},
		},
	}
}

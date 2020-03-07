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

package document

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/notes/internal"
	"k8s.io/release/pkg/release"
)

func TestFileMetadata(t *testing.T) {
	// Given
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	for _, file := range []string{
		"kubernetes-client-darwin-386.tar.gz",
		"kubernetes-client-darwin-amd64.tar.gz",
		"kubernetes-client-linux-386.tar.gz",
		"kubernetes-client-linux-amd64.tar.gz",
		"kubernetes-client-linux-arm.tar.gz",
		"kubernetes-client-linux-arm64.tar.gz",
		"kubernetes-client-linux-ppc64le.tar.gz",
		"kubernetes-client-linux-s390x.tar.gz",
		"kubernetes-client-windows-386.tar.gz",
		"kubernetes-client-windows-amd64.tar.gz",
		"kubernetes-node-linux-amd64.tar.gz",
		"kubernetes-node-linux-arm.tar.gz",
		"kubernetes-node-linux-arm64.tar.gz",
		"kubernetes-node-linux-ppc64le.tar.gz",
		"kubernetes-node-linux-s390x.tar.gz",
		"kubernetes-node-windows-amd64.tar.gz",
		"kubernetes-server-linux-amd64.tar.gz",
		"kubernetes-server-linux-arm.tar.gz",
		"kubernetes-server-linux-arm64.tar.gz",
		"kubernetes-server-linux-ppc64le.tar.gz",
		"kubernetes-server-linux-s390x.tar.gz",
		"kubernetes-src.tar.gz",
		"kubernetes.tar.gz",
	} {
		require.Nil(t, ioutil.WriteFile(
			filepath.Join(dir, file), []byte{1, 2, 3}, os.FileMode(0644),
		))
	}

	metadata, err := fetchMetadata(dir, "http://test.com", "test-release")
	require.Nil(t, err)

	expected := &FileMetadata{
		Source: []File{
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes.tar.gz", URL: "http://test.com/test-release/kubernetes.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-src.tar.gz", URL: "http://test.com/test-release/kubernetes-src.tar.gz"},
		},
		Client: []File{
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-darwin-386.tar.gz", URL: "http://test.com/test-release/kubernetes-client-darwin-386.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-darwin-amd64.tar.gz", URL: "http://test.com/test-release/kubernetes-client-darwin-amd64.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-linux-386.tar.gz", URL: "http://test.com/test-release/kubernetes-client-linux-386.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-linux-amd64.tar.gz", URL: "http://test.com/test-release/kubernetes-client-linux-amd64.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-linux-arm.tar.gz", URL: "http://test.com/test-release/kubernetes-client-linux-arm.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-linux-arm64.tar.gz", URL: "http://test.com/test-release/kubernetes-client-linux-arm64.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-linux-ppc64le.tar.gz", URL: "http://test.com/test-release/kubernetes-client-linux-ppc64le.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-linux-s390x.tar.gz", URL: "http://test.com/test-release/kubernetes-client-linux-s390x.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-windows-386.tar.gz", URL: "http://test.com/test-release/kubernetes-client-windows-386.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-client-windows-amd64.tar.gz", URL: "http://test.com/test-release/kubernetes-client-windows-amd64.tar.gz"},
		},
		Server: []File{
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-server-linux-amd64.tar.gz", URL: "http://test.com/test-release/kubernetes-server-linux-amd64.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-server-linux-arm.tar.gz", URL: "http://test.com/test-release/kubernetes-server-linux-arm.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-server-linux-arm64.tar.gz", URL: "http://test.com/test-release/kubernetes-server-linux-arm64.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-server-linux-ppc64le.tar.gz", URL: "http://test.com/test-release/kubernetes-server-linux-ppc64le.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-server-linux-s390x.tar.gz", URL: "http://test.com/test-release/kubernetes-server-linux-s390x.tar.gz"},
		},
		Node: []File{
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-node-linux-amd64.tar.gz", URL: "http://test.com/test-release/kubernetes-node-linux-amd64.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-node-linux-arm.tar.gz", URL: "http://test.com/test-release/kubernetes-node-linux-arm.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-node-linux-arm64.tar.gz", URL: "http://test.com/test-release/kubernetes-node-linux-arm64.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-node-linux-ppc64le.tar.gz", URL: "http://test.com/test-release/kubernetes-node-linux-ppc64le.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-node-linux-s390x.tar.gz", URL: "http://test.com/test-release/kubernetes-node-linux-s390x.tar.gz"},
			{Checksum: "27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29", Name: "kubernetes-node-windows-amd64.tar.gz", URL: "http://test.com/test-release/kubernetes-node-windows-amd64.tar.gz"},
		},
	}
	require.Nil(t, pretty.Diff(metadata, expected))
}

func TestRenderMarkdownTemplate(t *testing.T) {
	// Given
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	for _, file := range []string{
		"kubernetes-client-darwin-386.tar.gz",
		"kubernetes-client-darwin-amd64.tar.gz",
		"kubernetes-client-linux-386.tar.gz",
		"kubernetes-client-linux-amd64.tar.gz",
		"kubernetes-client-linux-arm.tar.gz",
		"kubernetes-client-linux-arm64.tar.gz",
		"kubernetes-client-linux-ppc64le.tar.gz",
		"kubernetes-client-linux-s390x.tar.gz",
		"kubernetes-client-windows-386.tar.gz",
		"kubernetes-client-windows-amd64.tar.gz",
		"kubernetes-node-linux-amd64.tar.gz",
		"kubernetes-node-linux-arm.tar.gz",
		"kubernetes-node-linux-arm64.tar.gz",
		"kubernetes-node-linux-ppc64le.tar.gz",
		"kubernetes-node-linux-s390x.tar.gz",
		"kubernetes-node-windows-amd64.tar.gz",
		"kubernetes-server-linux-amd64.tar.gz",
		"kubernetes-server-linux-arm.tar.gz",
		"kubernetes-server-linux-arm64.tar.gz",
		"kubernetes-server-linux-ppc64le.tar.gz",
		"kubernetes-server-linux-s390x.tar.gz",
		"kubernetes-src.tar.gz",
		"kubernetes.tar.gz",
	} {
		require.Nil(t, ioutil.WriteFile(
			filepath.Join(dir, file), []byte{1, 2, 3}, os.FileMode(0644),
		))
	}

	doc := Document{
		NotesWithActionRequired: []string{"If an API changes and no one documented it, did it really happen?"},
		NotesByKind: NotesByKind{
			KindAPIChange:       []string{"This might make people sad...or happy."},
			KindBug:             []string{"This will likely get you promoted."},
			KindCleanup:         []string{"This usually does not get you promoted but it should."},
			KindDeprecation:     []string{"Delorted."},
			KindDesign:          []string{"Change the world."},
			KindDocumentation:   []string{"There was  a library in Alexandria, Egypt once."},
			KindFailingTest:     []string{"Please run presubmit test!"},
			KindFeature:         []string{"This will get you promoted."},
			KindFlake:           []string{"This *should* get you promoted."},
			KindBugCleanupFlake: []string{"This should definitely get you promoted."},
			KindUncategorized:   []string{"Someone somewhere did the world a great justice."},
		},
		PreviousRevision: "v1.16.0",
		CurrentRevision:  "v1.16.1",
	}

	// When
	goldenFile, err := bazel.Runfile(filepath.Join("testdata", "document.md.golden"))
	require.Nil(t, err, "Locating runfiles are you using bazel test?")

	b, err := ioutil.ReadFile(goldenFile)
	require.Nil(t, err, "Reading golden file %q", goldenFile)
	expected := string(b)

	got, err := doc.RenderMarkdownTemplate(release.ProductionBucket, dir, internal.DefaultReleaseNotesTemplate)
	require.Nil(t, err, "Rendering document")

	// Then
	require.Nil(t, pretty.Diff(expected, got), "Unexpected diff")
}

func TestCreateDownloadsTable(t *testing.T) {
	// Given
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	for _, file := range []string{
		"kubernetes-client-darwin-386.tar.gz",
		"kubernetes-client-darwin-amd64.tar.gz",
		"kubernetes-client-linux-386.tar.gz",
		"kubernetes-client-linux-amd64.tar.gz",
		"kubernetes-client-linux-arm.tar.gz",
		"kubernetes-client-linux-arm64.tar.gz",
		"kubernetes-client-linux-ppc64le.tar.gz",
		"kubernetes-client-linux-s390x.tar.gz",
		"kubernetes-client-windows-386.tar.gz",
		"kubernetes-client-windows-amd64.tar.gz",
		"kubernetes-node-linux-amd64.tar.gz",
		"kubernetes-node-linux-arm.tar.gz",
		"kubernetes-node-linux-arm64.tar.gz",
		"kubernetes-node-linux-ppc64le.tar.gz",
		"kubernetes-node-linux-s390x.tar.gz",
		"kubernetes-node-windows-amd64.tar.gz",
		"kubernetes-server-linux-amd64.tar.gz",
		"kubernetes-server-linux-arm.tar.gz",
		"kubernetes-server-linux-arm64.tar.gz",
		"kubernetes-server-linux-ppc64le.tar.gz",
		"kubernetes-server-linux-s390x.tar.gz",
		"kubernetes-src.tar.gz",
		"kubernetes.tar.gz",
	} {
		require.Nil(t, ioutil.WriteFile(
			filepath.Join(dir, file), []byte{1, 2, 3}, os.FileMode(0644),
		))
	}

	// When
	output := &strings.Builder{}
	require.Nil(t, CreateDownloadsTable(
		output, release.ProductionBucket, dir, "v1.16.0", "v1.16.1",
	))

	// Then
	require.Equal(t, `# v1.16.1

[Documentation](https://docs.k8s.io)

## Downloads for v1.16.1

filename | sha512 hash
-------- | -----------
[kubernetes.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-src.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-src.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`

### Client Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-client-darwin-386.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-darwin-386.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-client-darwin-amd64.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-darwin-amd64.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-client-linux-386.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-linux-386.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-client-linux-amd64.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-linux-amd64.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-client-linux-arm.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-linux-arm.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-client-linux-arm64.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-linux-arm64.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-client-linux-ppc64le.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-linux-ppc64le.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-client-linux-s390x.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-linux-s390x.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-client-windows-386.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-windows-386.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-client-windows-amd64.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-client-windows-amd64.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`

### Server Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-server-linux-amd64.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-server-linux-amd64.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-server-linux-arm.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-server-linux-arm.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-server-linux-arm64.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-server-linux-arm64.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-server-linux-ppc64le.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-server-linux-ppc64le.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-server-linux-s390x.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-server-linux-s390x.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`

### Node Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-node-linux-amd64.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-node-linux-amd64.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-node-linux-arm.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-node-linux-arm.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-node-linux-arm64.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-node-linux-arm64.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-node-linux-ppc64le.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-node-linux-ppc64le.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-node-linux-s390x.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-node-linux-s390x.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`
[kubernetes-node-windows-amd64.tar.gz](https://dl.k8s.io/v1.16.1/kubernetes-node-windows-amd64.tar.gz) | `+"`"+`27864cc5219a951a7a6e52b8c8dddf6981d098da1658d96258c870b2c88dfbcb51841aea172a28bafa6a79731165584677066045c959ed0f9929688d04defc29`+"`"+`

## Changelog since v1.16.0

`, output.String())
}

func TestSortKinds(t *testing.T) {
	input := NotesByKind{
		"cleanup":                       nil,
		"api-change":                    nil,
		"deprecation":                   nil,
		"documentation":                 nil,
		"Other (Bug, Cleanup or Flake)": nil,
		"failing-test":                  nil,
		"design":                        nil,
		"flake":                         nil,
		"bug":                           nil,
		"feature":                       nil,
		"Uncategorized":                 nil,
	}
	res := sortKinds(input)
	require.Equal(t, res, kindPriority)
}

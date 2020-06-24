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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/options"
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
	require.Equal(t, metadata, expected)
}

func TestDocument_RenderMarkdownTemplateFailure(t *testing.T) {
	tests := []struct {
		name             string
		templateSpec     string
		templateContents string
		templateExist    bool
	}{
		{
			"given template exist but is empty",
			"go-template:empty.tmpl",
			"",
			true,
		},
		{
			"given bad template spec",
			"wrong-prefix:template.tmpl",
			"",
			true,
		},
		{
			"given bad template contents",
			"go-template:bad.tmpl",
			"# This template will not parse: {{}",
			true,
		},
		{
			"given non-existent template",
			"go-template:non-exist.tmpl",
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "")
			require.Nil(t, err)
			defer os.RemoveAll(dir)

			if tt.templateExist {
				fileName := strings.Split(tt.templateSpec, ":")[1]
				p := filepath.Join(dir, fileName)
				require.Nil(t, ioutil.WriteFile(p, []byte(tt.templateContents), 0664))
			}

			doc := Document{}
			_, err = doc.RenderMarkdownTemplate("", "", tt.templateSpec)
			require.Error(t, err, "Unexpected success")
		})
	}
}

func TestCreateDownloadsTable(t *testing.T) {
	// Given
	dir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)
	setupTestDir(t, dir)

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

// setupTestDir adds basic test files to a given directory.
func setupTestDir(t *testing.T, dir string) {
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
}

func TestNew(t *testing.T) {
	type args struct {
		releaseNotes notes.ReleaseNotes
		history      notes.ReleaseNotesHistory
	}
	tests := []struct {
		name string
		args args
		want *Document
	}{
		{
			"notes with no kinds are uncategorized",
			args{
				notes.ReleaseNotes{
					0: makeReleaseNote("", "No one gave me a kind"),
				},
				notes.ReleaseNotesHistory{0},
			},
			&Document{
				NotesWithActionRequired: notes.Notes{},
				Notes: NoteCollection{
					NoteCategory{
						Kind:        notes.KindUncategorized,
						NoteEntries: &notes.Notes{"No one gave me a kind"},
					},
				},
			},
		},
		{
			"notes of same kind are lexicographically sorted.",
			args{
				notes.ReleaseNotes{
					0: makeReleaseNote(notes.KindDeprecation, "C"),
					1: makeReleaseNote(notes.KindDeprecation, "B"),
					2: makeReleaseNote(notes.KindDeprecation, "A"),
				},
				notes.ReleaseNotesHistory{0, 1, 2},
			},
			&Document{
				NotesWithActionRequired: notes.Notes{},
				Notes: NoteCollection{
					NoteCategory{
						Kind:        notes.KindDeprecation,
						NoteEntries: &notes.Notes{"A", "B", "C"},
					},
				},
			},
		},
		{
			"notes are sorted by kind priority",
			args{
				notes.ReleaseNotes{
					0: makeReleaseNote(notes.KindFeature, "C"),
					1: makeReleaseNote(notes.KindAPIChange, "B"),
					2: makeReleaseNote(notes.KindDeprecation, "A"),
				},
				notes.ReleaseNotesHistory{0, 1, 2},
			},
			&Document{
				NotesWithActionRequired: notes.Notes{},
				Notes: NoteCollection{
					NoteCategory{
						Kind:        notes.KindDeprecation,
						NoteEntries: &notes.Notes{"A"},
					},
					NoteCategory{
						Kind:        notes.KindAPIChange,
						NoteEntries: &notes.Notes{"B"},
					},
					NoteCategory{
						Kind:        notes.KindFeature,
						NoteEntries: &notes.Notes{"C"},
					},
				},
			},
		},
		{
			"strip unwanted prefixes",
			args{
				notes.ReleaseNotes{
					0: makeReleaseNote(notes.KindBug, "- single dash"),
					1: makeReleaseNote(notes.KindBug, "-- double dash"),
					2: makeReleaseNote(notes.KindBug, "* single star"),
					3: makeReleaseNote(notes.KindBug, "** double star"),
					4: makeReleaseNote(notes.KindBug, "- --someflag"),
				},
				notes.ReleaseNotesHistory{0, 1, 2, 3, 4},
			},
			&Document{
				NotesWithActionRequired: notes.Notes{},
				Notes: NoteCollection{
					NoteCategory{
						Kind: notes.KindBug,
						NoteEntries: &notes.Notes{
							"--someflag",
							"double dash",
							"double star",
							"single dash",
							"single star",
						},
					},
				},
			},
		},
		{
			"highest kind for duplicate note",
			args{
				notes.ReleaseNotes{
					0: &notes.ReleaseNote{
						Markdown:       "A duplicate note gets the highest priority kind found",
						Kinds:          []string{string(notes.KindAPIChange), string(notes.KindDeprecation)},
						DuplicateKind:  true,
						ActionRequired: false,
					}},
				notes.ReleaseNotesHistory{0},
			},
			&Document{
				NotesWithActionRequired: notes.Notes{},
				Notes: NoteCollection{
					NoteCategory{
						Kind:        notes.KindDeprecation,
						NoteEntries: &notes.Notes{"A duplicate note gets the highest priority kind found"},
					},
				},
			},
		},
		{
			"notes with action required get their own category",
			args{
				notes.ReleaseNotes{
					0: &notes.ReleaseNote{
						Markdown:       "This note should not appear as a regular note.",
						Kinds:          []string{string(notes.KindDeprecation)},
						DuplicateKind:  true,
						ActionRequired: false,
					}},
				notes.ReleaseNotesHistory{0},
			},
			&Document{
				NotesWithActionRequired: notes.Notes{},
				Notes: NoteCollection{
					NoteCategory{
						Kind:        notes.KindDeprecation,
						NoteEntries: &notes.Notes{"This note should not appear as a regular note."},
					},
				},
			},
		},
		{
			"notes mapping to a single kind",
			args{
				notes.ReleaseNotes{
					0: makeReleaseNote(notes.KindCleanup, "PR#1"),
					1: makeReleaseNote(notes.KindFlake, "PR#2"),
				},
				notes.ReleaseNotesHistory{0, 1},
			},
			&Document{
				NotesWithActionRequired: notes.Notes{},
				Notes: NoteCollection{
					NoteCategory{
						Kind:        notes.KindOther,
						NoteEntries: &notes.Notes{"PR#1", "PR#2"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.releaseNotes, tt.args.history, "", "")
			require.NoError(t, err)
			require.Equal(t, got, tt.want, "Unexpected return.")
		})
	}
}

func TestDocument_RenderMarkdownTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateSpec   string
		userTemplate   bool
		hasDownloads   bool
		wantGoldenFile string
	}{
		{
			"render default template and downloads",
			options.FormatSpecDefaultGoTemplate,
			false,
			true,
			"document.md.golden",
		},
		{
			"render default template and no downloads",
			options.FormatSpecDefaultGoTemplate,
			false,
			false,
			"document_without_downloads.md.golden",
		},
		{
			"render user-specified template and downloads",
			"go-template:user-template.tmpl",
			true,
			true,
			"document.md.golden",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			testNotes := notes.ReleaseNotes{
				0:  makeReleaseNote(notes.KindDeprecation, "Deprecation #1."),
				1:  makeReleaseNote(notes.KindBug, "Bugfix."),
				2:  makeReleaseNote(notes.KindCleanup, "Clean up."),
				3:  makeReleaseNote(notes.KindDesign, "Design change."),
				4:  makeReleaseNote(notes.KindDocumentation, "Update docs."),
				5:  makeReleaseNote(notes.KindFailingTest, "Fix a failing test."),
				6:  makeReleaseNote(notes.KindFeature, "A feature."),
				7:  makeReleaseNote(notes.KindFlake, "Fix a flakey test."),
				8:  makeReleaseNote("", "Uncategorized note."),
				9:  makeReleaseNote(notes.KindBug, "- This note was prepended with a dash (-) initially."),
				10: makeReleaseNote(notes.KindBug, "* This note was prepended with a star (*) initially."),
			}
			duplicate := makeReleaseNote(notes.KindDeprecation, "This note is duplicated across SIGs.")
			duplicate.Kinds = append(duplicate.Kinds, string(notes.KindBug))
			duplicate.DuplicateKind = true

			actionNeeded := makeReleaseNote(notes.KindAPIChange, "Action required note.")
			actionNeeded.ActionRequired = true
			testNotes[11] = duplicate
			testNotes[12] = actionNeeded

			doc, err := New(testNotes, makeReleaseNoteHistory(testNotes), "v1.16.0", "v1.16.1")
			require.NoError(t, err, "Creating test document.")

			templateSpec := tt.templateSpec
			var dir string
			if tt.hasDownloads || tt.userTemplate {
				dir, err = ioutil.TempDir("", "")
				require.NoError(t, err, "Creating tmpDir")
				defer os.RemoveAll(dir)

				setupTestDir(t, dir)

				// This helps exercise reading a user template from disk.
				if tt.userTemplate {
					// Write out the default template to simulate reading an actual template.
					p := filepath.Join(dir, strings.Split(tt.templateSpec, ":")[1])
					templateSpec = fmt.Sprintf("go-template:%s", p)

					require.NoError(
						t,
						ioutil.WriteFile(p, []byte(defaultReleaseNotesTemplate), 0664),
						"Writing user specified template.")
				}
			}

			// When
			got, err := doc.RenderMarkdownTemplate(release.ProductionBucket, dir, templateSpec)

			// Then
			require.NoError(t, err, "Unexpected error.")
			expected := readFile(t, filepath.Join("testdata", tt.wantGoldenFile))
			require.Equal(t, expected, got)
		})
	}
}

func makeReleaseNote(kind notes.Kind, markdown string) *notes.ReleaseNote {
	n := &notes.ReleaseNote{Markdown: markdown}
	if kind != "" {
		n.Kinds = []string{string(kind)}
	}
	return n
}

func makeReleaseNoteHistory(n notes.ReleaseNotes) notes.ReleaseNotesHistory {
	var r notes.ReleaseNotesHistory
	for i := 0; i < len(n); i++ {
		r = append(r, i)
	}
	return r
}

func readFile(t *testing.T, path string) string {
	goldenFile, err := bazel.Runfile(path)
	require.NoError(t, err, "Locating runfiles are you using bazel test?")

	b, err := ioutil.ReadFile(goldenFile)
	require.NoError(t, err, "Reading file %q", path)
	return string(b)
}

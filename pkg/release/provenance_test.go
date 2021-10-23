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

package release

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/release-sdk/object"
)

func TestGetBuildSubjects(t *testing.T) {
	// create a test dir with stuff
	dir, err := os.MkdirTemp("", "provenance-test-")
	require.Nil(t, err)
	defer os.RemoveAll(dir)
	require.Nil(t, os.Mkdir(filepath.Join(dir, ImagesPath), os.FileMode(0o755)))

	testFiles := map[string]struct {
		data     string
		checksum string
	}{
		"README.md": {
			data:     "123",
			checksum: "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
		},
		"LICENSE": {
			data:     "Copyright © 2000",
			checksum: "c8377d9ac397f096eb81aa8b409945b8f228976bc8eb7f4de32d222018163534",
		},
	}

	for filename, testData := range testFiles {
		require.Nil(t, os.WriteFile(filepath.Join(dir, ImagesPath, filename), []byte(testData.data), os.FileMode(0o644)))
	}

	impl := defaultProvenanceReaderImpl{}
	opts := ProvenanceReaderOptions{
		Bucket:       "test-bucket",
		BuildVersion: "v0.0158987987",
		WorkspaceDir: dir,
	}
	version := "v1.0"
	subjects, err := impl.GetBuildSubjects(&opts, dir, version)
	require.Nil(t, err)
	require.Equal(t, len(testFiles), len(subjects))

	// Check the files have the bucket prefix:
	gcsPath := filepath.Join(opts.Bucket, StagePath, opts.BuildVersion)

	// Check subjects match in their data and rewritten path
	for _, sub := range subjects {
		filename := filepath.Base(sub.Name)
		require.NotEmpty(t, filename)
		_, ok := testFiles[filename]
		require.True(t, ok)
		require.Equal(t, testFiles[filename].checksum, sub.Digest["sha256"])
		require.True(t, strings.HasPrefix(sub.Name, object.GcsPrefix+gcsPath), filename)
		require.True(t, strings.HasPrefix(sub.Name, object.GcsPrefix+filepath.Join(gcsPath, version)), filename)
	}
}

func TestGetStagingSubjects(t *testing.T) {
	// create a test dir with stuff
	dir, err := os.MkdirTemp("", "provenance-test-")
	require.Nil(t, err)
	defer os.RemoveAll(dir)
	require.Nil(t, os.Mkdir(filepath.Join(dir, "second"), os.FileMode(0o755)))

	testFiles := map[string]struct {
		data     string
		checksum string
	}{
		SourcesTar: {
			data:     "123",
			checksum: "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
		},
		"second/LICENSE": {
			data:     "Copyright © 2000",
			checksum: "c8377d9ac397f096eb81aa8b409945b8f228976bc8eb7f4de32d222018163534",
		},
	}

	for filename, testData := range testFiles {
		require.Nil(t, os.WriteFile(filepath.Join(dir, filename), []byte(testData.data), os.FileMode(0o644)))
	}

	impl := defaultProvenanceReaderImpl{}
	opts := ProvenanceReaderOptions{
		Bucket:       "test-bucket",
		BuildVersion: "v0.0158987987",
		WorkspaceDir: dir,
	}
	subjects, err := impl.GetStagingSubjects(&opts, dir)
	require.Nil(t, err)
	require.Equal(t, len(testFiles), len(subjects))

	// Check the files have the bucket prefix:
	gcsPath := filepath.Join(opts.Bucket, StagePath, opts.BuildVersion)

	// Al subjects should appear in the provenance data, we have an exception
	// with the sources tar which always goes to the top of the staging dir
	for _, sub := range subjects {
		filename := filepath.Base(sub.Name)
		require.NotEmpty(t, filename)
		_, ok := testFiles[filepath.Join("second", filename)]
		if filename == SourcesTar {
			require.False(t, ok)
			require.Equal(t, sub.Name, object.GcsPrefix+filepath.Join(gcsPath, SourcesTar))
		} else {
			require.True(t, ok)
			require.True(t, strings.HasPrefix(sub.Name, object.GcsPrefix+gcsPath), filename)
		}
	}
}

/*
Copyright 2020 The Kubernetes Authors.

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
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/object"
)

const fictionalTestBucketName = "kubernetes-test-name"

func TestCopyReleaseLogsToWorkDir(t *testing.T) {
	// create two files to  simulate logs
	tmp1, err := os.CreateTemp("", "test-copylogs-")
	require.Nil(t, err)
	defer os.Remove(tmp1.Name())
	tmp2, err := os.CreateTemp("", "test-copylogs-")
	require.Nil(t, err)
	defer os.Remove(tmp2.Name())

	// Create a target directory to copy the files
	destDir, err := os.MkdirTemp("", "test-move-logs")
	require.Nil(t, err, "creating test dir")
	defer os.RemoveAll(destDir)

	// Put an oauth token in the first log
	content1 := "7aa33bd2186c40840c4c2df321241e241def98ca:x-oauth-basic\n"
	require.Nil(t, os.WriteFile(
		tmp1.Name(), []byte(content1), os.FileMode(0o644),
	))

	// Write a regular text in the second
	content2 := "Hello world!\n"
	require.Nil(t, os.WriteFile(
		tmp2.Name(), []byte(content2), os.FileMode(0o644),
	))

	// Create the implementation
	impl := &defaultArchiverImpl{}
	// Copy the log files to the mock directory
	err = impl.CopyReleaseLogs([]string{tmp1.Name(), tmp2.Name()}, destDir, "")
	require.Nil(t, err)

	// Reopoen the files to check them
	movedData1, err := os.ReadFile(filepath.Join(destDir, filepath.Base(tmp1.Name())))
	require.Nil(t, err)

	movedData2, err := os.ReadFile(filepath.Join(destDir, filepath.Base(tmp2.Name())))
	require.Nil(t, err)

	// The first file should be sanitized. Should differ
	require.NotEmpty(t, movedData1)
	require.NotEqual(t, content1, string(movedData1))

	// The second file should be the same.
	require.Equal(t, content2, string(movedData2))
}

func TestArchiveBucketPath(t *testing.T) {
	// By default, the bucket name is empty, resulting in an empty
	// GCS URI. This is to avoid getting a wrong but valid bucket location
	DefaultArchiveOptions := &ArchiverOptions{}
	require.Empty(t, DefaultArchiveOptions.ArchiveBucketPath())

	// Using a bucket name should return a valid URI
	opts := ArchiverOptions{
		// Here, we test without "gs://", the gcs package should
		// normalize the location with or without
		Bucket:       fictionalTestBucketName,
		PrimeVersion: "v1.20.0-beta.2",
	}
	require.Equal(t,
		object.GcsPrefix+filepath.Join(fictionalTestBucketName, archiveBucketPath, archiveDirPrefix+opts.PrimeVersion),
		opts.ArchiveBucketPath(),
	)
}

func TestValidateOpts(t *testing.T) {
	testOpts := &ArchiverOptions{}

	// An empty options struct should not validate:
	require.NotNil(t, testOpts.Validate())

	// Create a fake directory to test
	dir, err := os.MkdirTemp("", "archiver-opts-test-")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	// With complete values, still should not validate as most
	// directories do not exist
	testOpts.Bucket = "kubernetes-test-name"
	testOpts.PrimeVersion = "v1.20.0-beta.1"
	testOpts.BuildVersion = "v1.20.0-beta.0.80+cdfd82733af78c"
	testOpts.ReleaseBuildDir = filepath.Join(dir, testOpts.BuildVersion)
	require.NotNil(t, testOpts.Validate())

	// Creating a test log (/workdir/tmp) should still not
	// validate, build dir is missing
	tmplog, err := os.CreateTemp("", "anago-test-log-")
	require.Nil(t, err)
	testOpts.LogFile = tmplog.Name()
	require.NotNil(t, testOpts.Validate())

	// Finally create the build dir and we're done
	require.Nil(t, os.Mkdir(testOpts.ReleaseBuildDir, os.FileMode(0o755)))

	// Should succeed
	require.Nil(t, testOpts.Validate())
}

func TestGetLogFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-get-anago-logs")
	require.Nil(t, err, "creating test dir")
	defer os.RemoveAll(dir)

	goodFiles := []string{"anago.log", "anago-stage.log", "anago.log.1", "anago-stage.log.1"}
	allFiles := goodFiles
	allFiles = append(allFiles, "test1.txt", "other.log", "anago.txt")
	for _, fileName := range allFiles {
		require.Nil(t, os.WriteFile(filepath.Join(dir, fileName), []byte("test"), os.FileMode(0o644)))
	}
	impl := &defaultArchiverImpl{}
	foundPaths, err := impl.GetLogFiles(dir)
	foundFilenames := []string{}
	for _, logFile := range foundPaths {
		foundFilenames = append(foundFilenames, filepath.Base(logFile))
	}
	require.Nil(t, err)
	require.ElementsMatch(t, goodFiles, foundFilenames)
}

func TestDeleteStalePasswordFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-delete-stale-pw-")
	require.Nil(t, err, "creating test dir")
	defer os.RemoveAll(dir)

	// Create temporary files
	require.Nil(t, os.WriteFile(filepath.Join(dir, "test.txt"), []byte("Hello World"), os.FileMode(0o644)))
	require.Nil(t, os.WriteFile(filepath.Join(dir, "rsyncd.password"), []byte("Hello World"), os.FileMode(0o644)))

	// Run the pass clear
	impl := &defaultArchiverImpl{}
	err = impl.DeleteStalePasswordFiles(dir)
	require.Nil(t, err)

	// Check that the pw file is gone while the other files remained
	require.FileExists(t, filepath.Join(dir, "test.txt"))
	require.NoFileExists(t, filepath.Join(dir, "rsyncd.password"))
}

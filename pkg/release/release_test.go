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

package release

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultToolRepoURLSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		useSSH   bool
		expected string
	}{
		{
			name:     "default HTTPS",
			expected: "https://github.com/kubernetes/release",
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		actual, err := GetDefaultToolRepoURL()
		assert.Equal(t, tc.expected, actual)
		assert.Nil(t, err)
	}
}

func TestGetToolRepoURLSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		org      string
		repo     string
		useSSH   bool
		expected string
	}{
		{
			name:     "default HTTPS",
			expected: "https://github.com/kubernetes/release",
		},
		{
			name:     "ssh with custom org",
			org:      "fake-org",
			useSSH:   true,
			expected: "git@github.com:fake-org/release",
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)

		actual, err := GetToolRepoURL(tc.org, tc.repo, tc.useSSH)
		assert.Equal(t, tc.expected, actual)
		assert.Nil(t, err)
	}
}

func TestGetToolBranchSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "default branch",
			expected: "master",
		},
		{
			name:     "custom branch",
			branch:   "tool-branch",
			expected: "tool-branch",
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)
		os.Setenv("TOOL_BRANCH", tc.branch)

		actual := GetToolBranch()
		assert.Equal(t, tc.expected, actual)
	}
}

func TestBuiltWithBazel(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	bazelTmpDir, err := ioutil.TempDir("", "bazel")
	require.Nil(t, err)
	dockerTmpDir, err := ioutil.TempDir("", "docker")
	require.Nil(t, err)

	release := "kubernetes"

	// Build directories.
	require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, bazelBuildPath), os.ModePerm))
	require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, dockerBuildPath), os.ModePerm))
	require.Nil(t, os.MkdirAll(filepath.Join(bazelTmpDir, bazelBuildPath), os.ModePerm))
	require.Nil(t, os.MkdirAll(filepath.Join(dockerTmpDir, dockerBuildPath), os.ModePerm))

	// Create test files.
	baseBazelFile := filepath.Join(baseTmpDir, "bazel-bin/build/release-tars/kubernetes.tar.gz")
	require.Nil(t, ioutil.WriteFile(
		baseBazelFile,
		[]byte("test"),
		os.FileMode(0644),
	))
	bazelFile := filepath.Join(bazelTmpDir, "bazel-bin/build/release-tars/kubernetes.tar.gz")
	require.Nil(t, ioutil.WriteFile(
		bazelFile,
		[]byte("test"),
		os.FileMode(0644),
	))

	time.Sleep(1 * time.Second)

	baseDockerFile := filepath.Join(baseTmpDir, "_output/release-tars/kubernetes.tar.gz")
	require.Nil(t, ioutil.WriteFile(
		baseDockerFile,
		[]byte("test"),
		os.FileMode(0644),
	))
	dockerFile := filepath.Join(dockerTmpDir, "_output/release-tars/1.1.1.tar.gz")
	require.Nil(t, ioutil.WriteFile(
		dockerFile,
		[]byte("test"),
		os.FileMode(0644),
	))

	defer cleanupTmps(t, baseTmpDir, bazelTmpDir, dockerTmpDir)

	type args struct {
		path    string
		release string
	}
	type want struct {
		r   bool
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"DockerMoreRecent": {
			args: args{
				path:    baseTmpDir,
				release: release,
			},
			want: want{
				r:   false,
				err: nil,
			},
		},
		"DockerOnly": {
			args: args{
				path:    baseTmpDir,
				release: release,
			},
			want: want{
				r:   false,
				err: nil,
			},
		},
		"BazelOnly": {
			args: args{
				path:    bazelTmpDir,
				release: release,
			},
			want: want{
				r:   true,
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res, err := BuiltWithBazel(tc.args.path, tc.args.release)
			require.Equal(t, tc.want.err, err)
			require.Equal(t, tc.want.r, res)
		})
	}
}

func TestReadBazelVersion(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	version := "1.1.1"

	// Build directories.
	require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, "bazel-genfiles"), os.ModePerm))

	bazelVersionFile := filepath.Join(baseTmpDir, bazelVersionPath)
	require.Nil(t, ioutil.WriteFile(
		bazelVersionFile,
		[]byte(version),
		os.FileMode(0644),
	))

	defer cleanupTmps(t, baseTmpDir)

	type want struct {
		r    string
		rErr bool
	}
	cases := map[string]struct {
		path string
		want want
	}{
		"ReadVersion": {
			path: baseTmpDir,
			want: want{
				r:    version,
				rErr: false,
			},
		},
		"ReadVersionError": {
			path: "notadir",
			want: want{
				rErr: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res, err := ReadBazelVersion(tc.path)
			require.Equal(t, tc.want.rErr, err != nil)
			require.Equal(t, tc.want.r, res)
		})
	}
}

func TestReadDockerVersion(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	release := "kubernetes"
	version := "1.1.1"

	// Build directories.
	require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, dockerBuildPath), os.ModePerm))

	var b bytes.Buffer

	// Create a zip archive.
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	require.Nil(t, tw.WriteHeader(&tar.Header{
		Name: filepath.Join(release, dockerVersionPath),
		Size: int64(len(version)),
	}))
	_, err = tw.Write([]byte(version))
	require.Nil(t, err)
	require.Nil(t, gz.Close())
	require.Nil(t, tw.Close())
	require.Nil(t, ioutil.WriteFile(
		filepath.Join(baseTmpDir, dockerBuildPath, "kubernetes.tar.gz"),
		b.Bytes(),
		os.FileMode(0644),
	))

	defer cleanupTmps(t, baseTmpDir)

	type args struct {
		path        string
		releaseKind string
	}
	type want struct {
		r    string
		rErr bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"ReadVersion": {
			args: args{
				path:        baseTmpDir,
				releaseKind: release,
			},
			want: want{
				r:    version,
				rErr: false,
			},
		},
		"NoVersionFile": {
			args: args{
				path:        baseTmpDir,
				releaseKind: "notarelease",
			},
			want: want{
				rErr: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res, err := ReadDockerizedVersion(tc.args.path, tc.args.releaseKind)
			require.Equal(t, tc.want.rErr, err != nil)
			require.Equal(t, tc.want.r, res)
		})
	}
}

func TestIsValidReleaseBuild(t *testing.T) {
	type want struct {
		r    bool
		rErr bool
	}
	cases := map[string]struct {
		build string
		want  want
	}{
		"ValidRelease": {
			build: "v1.17.6",
			want: want{
				r:    true,
				rErr: false,
			},
		},
		"ValidReleaseBuild": {
			build: "v1.17.6.abcde",
			want: want{
				r:    true,
				rErr: false,
			},
		},
		"ValidReleaseDirty": {
			build: "v1.17.6.abcde-dirty",
			want: want{
				r:    true,
				rErr: false,
			},
		},
		"NotValidRelease": {
			build: "1.1.1",
			want: want{
				r:    false,
				rErr: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res, err := IsValidReleaseBuild(tc.build)
			require.Equal(t, tc.want.rErr, err != nil)
			require.Equal(t, tc.want.r, res)
		})
	}
}

func TestIsDirtyBuild(t *testing.T) {
	cases := map[string]struct {
		build string
		want  bool
	}{
		"Dirty": {
			build: "v1.17.6-dirty",
			want:  true,
		},
		"NotDirty": {
			build: "v1.17.6.abcde",
			want:  false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res := IsDirtyBuild(tc.build)
			require.Equal(t, tc.want, res)
		})
	}
}

func TestGetKubeVersionSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Release URL",
			url:      "https://dl.k8s.io/release/stable-1.13.txt",
			expected: "1.13.12",
		},
		{
			name:     "CI URL",
			url:      "https://dl.k8s.io/ci/latest-1.14.txt",
			expected: "1.14.11-beta.1.2+c8b135d0b49c44",
		},
	}

	for _, tc := range testcases {
		actual, err := GetKubeVersion(tc.url)

		if err != nil {
			t.Fatalf("did not expect an error: %v", err)
		}

		assert.Equal(t, tc.expected, actual)
	}
}

func TestGetKubeVersionFailure(t *testing.T) {
	testcases := []struct {
		name string
		url  string
	}{
		{
			name: "Empty URL string",
			url:  "",
		},
		{
			name: "Bad URL",
			url:  "https://fake.url",
		},
	}

	for _, tc := range testcases {
		_, err := GetKubeVersion(tc.url)

		require.Error(t, err)
	}
}

func cleanupTmps(t *testing.T, dir ...string) {
	for _, each := range dir {
		require.Nil(t, os.RemoveAll(each))
	}
}

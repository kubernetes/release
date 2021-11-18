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
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"sigs.k8s.io/release-sdk/git"
)

func TestGetToolRefSuccess(t *testing.T) {
	testcases := []struct {
		name     string
		ref      string
		expected string
	}{
		{
			name:     "default branch",
			expected: git.DefaultBranch,
		},
		{
			name:     "custom branch",
			ref:      "tool-branch",
			expected: "tool-branch",
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)
		require.Nil(t, os.Setenv("TOOL_REF", tc.ref))

		actual := GetToolRef()
		require.Equal(t, tc.expected, actual)
	}
}

func TestBuiltWithBazel(t *testing.T) {
	baseTmpDir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	bazelTmpDir, err := os.MkdirTemp("", "bazel")
	require.Nil(t, err)
	dockerTmpDir, err := os.MkdirTemp("", "docker")
	require.Nil(t, err)

	// Build directories.
	require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, BazelBuildDir, ReleaseTarsPath), os.ModePerm))
	require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, BuildDir, ReleaseTarsPath), os.ModePerm))
	require.Nil(t, os.MkdirAll(filepath.Join(bazelTmpDir, BazelBuildDir, ReleaseTarsPath), os.ModePerm))
	require.Nil(t, os.MkdirAll(filepath.Join(dockerTmpDir, BuildDir, ReleaseTarsPath), os.ModePerm))

	// Create test files.
	baseBazelFile := filepath.Join(baseTmpDir, "bazel-bin/build/release-tars/kubernetes.tar.gz")
	require.Nil(t, os.WriteFile(
		baseBazelFile,
		[]byte("test"),
		os.FileMode(0o644),
	))
	bazelFile := filepath.Join(bazelTmpDir, "bazel-bin/build/release-tars/kubernetes.tar.gz")
	require.Nil(t, os.WriteFile(
		bazelFile,
		[]byte("test"),
		os.FileMode(0o644),
	))

	time.Sleep(1 * time.Second)

	baseDockerFile := filepath.Join(
		baseTmpDir, BuildDir, "release-tars/kubernetes.tar.gz",
	)
	require.Nil(t, os.WriteFile(
		baseDockerFile,
		[]byte("test"),
		os.FileMode(0o644),
	))
	dockerFile := filepath.Join(
		dockerTmpDir, BuildDir, "release-tars/kubernetes.tar.gz",
	)
	require.Nil(t, os.WriteFile(
		dockerFile,
		[]byte("test"),
		os.FileMode(0o644),
	))

	defer cleanupTmps(t, baseTmpDir, bazelTmpDir, dockerTmpDir)

	type args struct {
		path string
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
				path: baseTmpDir,
			},
			want: want{
				r:   false,
				err: nil,
			},
		},
		"DockerOnly": {
			args: args{
				path: dockerTmpDir,
			},
			want: want{
				r:   false,
				err: nil,
			},
		},
		"BazelOnly": {
			args: args{
				path: bazelTmpDir,
			},
			want: want{
				r:   true,
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res, err := BuiltWithBazel(tc.args.path)
			require.Equal(t, tc.want.err, err)
			require.Equal(t, tc.want.r, res)
		})
	}
}

func TestReadBazelVersion(t *testing.T) {
	const version = "1.1.1"

	type want struct {
		r    string
		rErr bool
	}

	cases := map[string]struct {
		outdir string
		want   want
	}{
		"ReadLegacyVersion": {
			outdir: "bazel-genfiles",
			want: want{
				r:    version,
				rErr: false,
			},
		},
		"ReadVersion": {
			outdir: "bazel-bin",
			want: want{
				r:    version,
				rErr: false,
			},
		},
		"ReadVersionError": {
			outdir: "bazel-random",
			want: want{
				rErr: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			baseTmpDir, err := os.MkdirTemp("", "")
			if err != nil {
				t.Fatalf("unable to create temp dir: %v", err)
			}
			defer cleanupTmps(t, baseTmpDir)

			// Build directories.
			require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, tc.outdir), os.ModePerm))

			bazelVersionFile := filepath.Join(baseTmpDir, tc.outdir, "version")
			require.Nil(t, os.WriteFile(
				bazelVersionFile,
				[]byte(version),
				os.FileMode(0o644),
			))

			res, err := ReadBazelVersion(baseTmpDir)
			require.Equal(t, tc.want.rErr, err != nil)
			require.Equal(t, tc.want.r, res)
		})
	}
}

func TestReadDockerVersion(t *testing.T) {
	baseTmpDir, err := os.MkdirTemp("", "ahhh")
	require.Nil(t, err)

	release := "kubernetes"
	version := "1.1.1"
	versionBytes := []byte("1.1.1\n")

	// Build directories.
	require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, BuildDir, ReleaseTarsPath, release), os.ModePerm))

	var b bytes.Buffer

	// Create version file
	err = os.WriteFile(filepath.Join(baseTmpDir, BuildDir, ReleaseTarsPath, "kubernetes", "version"), versionBytes, os.FileMode(0o644))
	require.Nil(t, err)

	// Create a zip archive.
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	require.Nil(t, tw.WriteHeader(&tar.Header{
		Name: "kubernetes/version",
		Size: int64(len(versionBytes)),
	}))
	versionFile, err := os.Open(filepath.Join(baseTmpDir, BuildDir, ReleaseTarsPath, "kubernetes", "version"))
	require.Nil(t, err)
	_, err = io.Copy(tw, versionFile)
	require.Nil(t, err)
	require.Nil(t, tw.Close())
	require.Nil(t, gz.Close())
	require.Nil(t, os.WriteFile(
		filepath.Join(baseTmpDir, BuildDir, ReleaseTarsPath, KubernetesTar),
		b.Bytes(),
		os.FileMode(0o644),
	))

	defer cleanupTmps(t, baseTmpDir)

	type args struct {
		path string
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
				path: baseTmpDir,
			},
			want: want{
				r:    version,
				rErr: false,
			},
		},
		"ReadVersionError": {
			args: args{
				path: "notadir",
			},
			want: want{
				rErr: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res, err := ReadDockerizedVersion(tc.args.path)
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
		"InvalidCopiedBuild": {
			// This trimmed tag often gets copied when double clicking
			// on the CloudBuild console:
			build: "v1.22.0-alpha.0.787+e6",
			want: want{
				r:    false,
				rErr: false,
			},
		},
		"ValidCopiedBuild": {
			// Full tag from previous test case:
			build: "v1.22.0-alpha.0.787+e6b4fa381152d6",
			want: want{
				r:    true,
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

func cleanupTmps(t *testing.T, dir ...string) {
	for _, each := range dir {
		require.Nil(t, os.RemoveAll(each))
	}
}

func TestURLPrefixForBucket(t *testing.T) {
	for _, bucket := range []string{"bucket", "", ProductionBucket} {
		res := URLPrefixForBucket(bucket)
		parsed, err := url.Parse(res)
		require.Nil(t, err)
		require.NotNil(t, parsed)
	}
}

func TestCopyBinaries(t *testing.T) {
	for _, tc := range []struct {
		prepare  func() (rootPath string, cleanup func())
		validate func(error, string)
	}{
		{ // success client
			prepare: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "test-copy-binaries-")
				require.Nil(t, err)

				binDir := filepath.Join(
					tempDir, "client/linux-amd64/kubernetes/client/bin",
				)
				require.Nil(t, os.MkdirAll(binDir, os.FileMode(0o755)))

				for _, f := range []string{"1", "2", "3"} {
					_, err = os.Create(filepath.Join(binDir, f))
					require.Nil(t, err)
				}

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error, testDir string) {
				require.Nil(t, err)

				binDir := filepath.Join(testDir, "bin/linux/amd64")
				require.FileExists(t, filepath.Join(binDir, "1"))
				require.FileExists(t, filepath.Join(binDir, "2"))
				require.FileExists(t, filepath.Join(binDir, "3"))
				dirContent, err := os.ReadDir(binDir)
				require.Nil(t, err)
				require.Len(t, dirContent, 3)
			},
		},
		{ // success client skip non-dir
			prepare: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux-amd64/kubernetes/client/bin",
				), os.FileMode(0o755)))

				_, err = os.Create(filepath.Join(tempDir, "client/some-file"))
				require.Nil(t, err)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error, _ string) { require.Nil(t, err) },
		},
		{ // success server
			prepare: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux-amd64/kubernetes/client/bin",
				), os.FileMode(0o755)))
				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "server/linux-amd64/kubernetes/server/bin",
				), os.FileMode(0o755)))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error, _ string) { require.Nil(t, err) },
		},
		{ // success node
			prepare: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux-amd64/kubernetes/client/bin",
				), os.FileMode(0o755)))
				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "node/linux-amd64/kubernetes/node/bin",
				), os.FileMode(0o755)))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error, _ string) { require.Nil(t, err) },
		},
		{ // failure wrong server dir
			prepare: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux-amd64/kubernetes/client/bin",
				), os.FileMode(0o755)))
				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "server/linux-amd64/kubernetes/wrong/bin",
				), os.FileMode(0o755)))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error, _ string) { require.NotNil(t, err) },
		},
		{ // failure wrong node dir
			prepare: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux-amd64/kubernetes/client/bin",
				), os.FileMode(0o755)))
				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "node/linux-amd64/kubernetes/wrong/bin",
				), os.FileMode(0o755)))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error, _ string) { require.NotNil(t, err) },
		},
		{ // empty dirs should error
			prepare:  func() (string, func()) { return "", func() {} },
			validate: func(err error, _ string) { require.NotNil(t, err) },
		},
	} {
		// Given
		rootPath, cleanup := tc.prepare()
		stageDir := filepath.Join(rootPath, StagePath)

		// When
		err := CopyBinaries(rootPath, stageDir)

		// Then
		tc.validate(err, stageDir)
		cleanup()
	}
}

func TestWriteChecksums(t *testing.T) {
	for _, tc := range []struct {
		prepare  func() (rootPath string, cleanup func())
		validate func(err error, rootPath string)
	}{
		{ // success
			prepare: func() (rootPath string, cleanup func()) {
				tempDir, err := os.MkdirTemp("", "write-checksum-test-")
				require.Nil(t, err)

				for i, v := range []byte{1, 2, 4, 8, 16, 32, 64, 128} {
					require.Nil(t, os.WriteFile(
						filepath.Join(tempDir, fmt.Sprintf("%d", i)),
						[]byte{v}, os.FileMode(0o644),
					))
				}

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error, rootPath string) {
				require.Nil(t, err)

				for digest, shas := range map[int][]string{
					256: {
						"4bf5122f344554c53bde2ebb8cd2b7e3d1600ad631c385a5d7cce23c7785459a",
						"dbc1b4c900ffe48d575b5da5c638040125f65db0fe3e24494b76ea986457d986",
						"e52d9c508c502347344d8c07ad91cbd6068afc75ff6292f062a09ca381c89e71",
						"beead77994cf573341ec17b58bbf7eb34d2711c993c1d976b128b3188dc1829a",
						"c555eab45d08845ae9f10d452a99bfcb06f74a50b988fe7e48dd323789b88ee3",
						"36a9e7f1c95b82ffb99743e0c5c4ce95d83c9a430aac59f84ef3cbfab6145068",
						"c3641f8544d7c02f3580b07c0f9887f0c6a27ff5ab1d4a3e29caf197cfc299ae",
						"76be8b528d0075f7aae98d6fa57a6d3c83ae480a8469e668d7b0af968995ac71",
					},
					512: {
						"7b54b66836c1fbdd13d2441d9e1434dc62ca677fb68f5fe66a464baadecdbd00576f8d6b5ac3bcc80844b7d50b1cc6603444bbe7cfcf8fc0aa1ee3c636d9e339",
						"fab848c9b657a853ee37c09cbfdd149d0b3807b191dde9b623ccd95281dd18705b48c89b1503903845bba5753945351fe6b454852760f73529cf01ca8f69dcca",
						"b5b8c725507b5b13158e020d96fe4cfbf6d774e09161e2b599b8f35ae31f16e395825edef8aa69ad304ef80fed9baa0580d247cd84e57a2ae239aec90d2d5869",
						"f65a6bf8f40b01b87757cde53483d057e1442f3bd67d495d2047b7f7c329e0572e88c18808426706af3b8df2915ca3d527ad49597f211cf89e475a07c901312b",
						"dc3fee1c29fe441f11008464c18d074dc987dbe02831a4e06c1c4769e4bfce5e78e50f13d786389a577afb2563e306b5d079187e4eccb962e12a5f6c16f62a2e",
						"f90ddd77e400dfe6a3fcf479b00b1ee29e7015c5bb8cd70f5f15b4886cc339275ff553fc8a053f8ddc7324f45168cffaf81f8c3ac93996f6536eef38e5e40768",
						"e97b9cc0c1e22c66bff31f6c457c2b95b9f9af955c8a098e043734df7439031fd1c6748a139d99077eb2db5f3d98a0e9d05b6606e3d4010ec107a52cd7e43359",
						"dfe8ef54110b3324d3b889035c95cfb80c92704614bf76f17546ad4f4b08218a630e16da7df34766a975b3bb85b01df9e99a4ec0a1d0ec3de6bed7b7a40b2f10",
					},
				} {
					for i, expectedSha := range shas {
						shaSums, err := os.ReadFile(filepath.Join(
							rootPath, fmt.Sprintf("SHA%dSUMS", digest),
						))
						require.Nil(t, err)
						require.Contains(t, string(shaSums), expectedSha)

						sha, err := os.ReadFile(filepath.Join(
							rootPath, fmt.Sprintf("%d.sha%d", i, digest),
						))
						require.Nil(t, err)
						require.Equal(t, expectedSha, string(sha))
					}
				}
			},
		},

		{ // success no content
			prepare: func() (rootPath string, cleanup func()) {
				tempDir, err := os.MkdirTemp("", "write-checksum-test-")
				require.Nil(t, err)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error, _ string) {
				require.Nil(t, err)
			},
		},
		{ // failure dir not existing
			prepare: func() (rootPath string, cleanup func()) {
				tempDir, err := os.MkdirTemp("", "write-checksum-test-")
				require.Nil(t, err)
				require.Nil(t, os.RemoveAll(tempDir))

				return tempDir, func() {}
			},
			validate: func(err error, _ string) {
				require.NotNil(t, err)
			},
		},
	} {
		// Given
		rootPath, cleanup := tc.prepare()

		// When
		err := WriteChecksums(rootPath)

		// Then
		tc.validate(err, rootPath)
		cleanup()
	}
}

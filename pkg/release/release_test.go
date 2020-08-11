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
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/git"
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

		actual := GetDefaultToolRepoURL()
		require.Equal(t, tc.expected, actual)
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

		actual := GetToolRepoURL(tc.org, tc.repo, tc.useSSH)
		require.Equal(t, tc.expected, actual)
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
			expected: git.DefaultBranch,
		},
		{
			name:     "custom branch",
			branch:   "tool-branch",
			expected: "tool-branch",
		},
	}

	for _, tc := range testcases {
		t.Logf("Test case: %s", tc.name)
		require.Nil(t, os.Setenv("TOOL_BRANCH", tc.branch))

		actual := GetToolBranch()
		require.Equal(t, tc.expected, actual)
	}
}

func TestBuiltWithBazel(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	bazelTmpDir, err := ioutil.TempDir("", "bazel")
	require.Nil(t, err)
	dockerTmpDir, err := ioutil.TempDir("", "docker")
	require.Nil(t, err)

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
	dockerFile := filepath.Join(dockerTmpDir, "_output/release-tars/kubernetes.tar.gz")
	require.Nil(t, ioutil.WriteFile(
		dockerFile,
		[]byte("test"),
		os.FileMode(0644),
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
			baseTmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unable to create temp dir: %v", err)
			}
			defer cleanupTmps(t, baseTmpDir)

			// Build directories.
			require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, tc.outdir), os.ModePerm))

			bazelVersionFile := filepath.Join(baseTmpDir, tc.outdir, "version")
			require.Nil(t, ioutil.WriteFile(
				bazelVersionFile,
				[]byte(version),
				os.FileMode(0644),
			))

			res, err := ReadBazelVersion(baseTmpDir)
			require.Equal(t, tc.want.rErr, err != nil)
			require.Equal(t, tc.want.r, res)
		})
	}
}

func TestReadDockerVersion(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "ahhh")
	require.Nil(t, err)

	release := "kubernetes"
	version := "1.1.1"
	versionBytes := []byte("1.1.1\n")

	// Build directories.
	require.Nil(t, os.MkdirAll(filepath.Join(baseTmpDir, dockerBuildPath, release), os.ModePerm))

	var b bytes.Buffer

	// Create version file
	err = ioutil.WriteFile(filepath.Join(baseTmpDir, dockerBuildPath, dockerVersionPath), versionBytes, os.FileMode(0644))
	require.Nil(t, err)

	// Create a zip archive.
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	require.Nil(t, tw.WriteHeader(&tar.Header{
		Name: dockerVersionPath,
		Size: int64(len(versionBytes)),
	}))
	versionFile, err := os.Open(filepath.Join(baseTmpDir, dockerBuildPath, dockerVersionPath))
	require.Nil(t, err)
	_, err = io.Copy(tw, versionFile)
	require.Nil(t, err)
	require.Nil(t, tw.Close())
	require.Nil(t, gz.Close())
	require.Nil(t, ioutil.WriteFile(
		filepath.Join(baseTmpDir, dockerBuildPath, "kubernetes.tar.gz"),
		b.Bytes(),
		os.FileMode(0644),
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

func TestGetKubecrossVersionSuccess(t *testing.T) {
	_, err := GetKubecrossVersion("release-1.15")
	require.Nil(t, err)
}

func TestGetKubecrossVersionSuccessOneNotExisting(t *testing.T) {
	_, err := GetKubecrossVersion("not-existing", "release-1.15")
	require.Nil(t, err)
}

func TestGetKubecrossVersionFailureNotExisting(t *testing.T) {
	_, err := GetKubecrossVersion("not-existing")
	require.NotNil(t, err)
}

func TestGetKubecrossVersionFailureEmpty(t *testing.T) {
	_, err := GetKubecrossVersion()
	require.NotNil(t, err)
}

func writeTestManifests(t *testing.T) (mockPath string) {
	tmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	// Sample manifests
	var mockManifests = map[string]string{
		"amd64": `[{"Config":"3da02591bd93f4db77a2ee5fb83f28315bb034657447168cfa1ce6161a446873.json","RepoTags":["k8s.gcr.io/kube-apiserver-amd64:v1.19.0-rc.4","gcr.io/k8s-staging-kubernetes/kube-apiserver-amd64:v1.19.0-rc.4"],"Layers":["513643face35501b7b23d0c580bc9abea0d881b2ecc50cb9cb28f4ae58419552/layer.tar","b3b0bad90dd3a6fc642439a93726fcf3028505d1be034327a6d86fe357c3ea50/layer.tar","fcd29bdb829b1c4cf3bbdec0f44a5475d5b9877f225d39c3c22fa1902326261c/layer.tar"]}]`,
		"arm64": `[{"Config":"c47b48fe11f383c6e4f30fea7dbf507329d326e94524f8989328d3028a6bf5f5.json","RepoTags":["k8s.gcr.io/kube-apiserver-arm64:v1.19.0-rc.4","gcr.io/k8s-staging-kubernetes/kube-apiserver-arm64:v1.19.0-rc.4"],"Layers":["b3b0bad90dd3a6fc642439a93726fcf3028505d1be034327a6d86fe357c3ea50/layer.tar","db36fea64d6ee6553972b5fbae343c5fcd7cba44445db22e2ba5471045595372/layer.tar"]}]`,
	}

	// Prepare test environment
	for arch, manifest := range mockManifests {
		// Create the mock image directory
		require.Nil(t, os.MkdirAll(filepath.Join(tmpDir, ImagesPath, arch), os.ModePerm))

		// Create the fake image tar
		var b bytes.Buffer
		tw := tar.NewWriter(&b)
		require.Nil(t, tw.WriteHeader(&tar.Header{
			Name: "manifest.json",
			Size: int64(len(manifest)),
		}))
		_, err = fmt.Fprint(tw, manifest)
		require.Nil(t, err)
		require.Nil(t, tw.Close())

		require.Nil(t, ioutil.WriteFile(
			filepath.Join(filepath.Join(tmpDir, ImagesPath, arch), "kube-apiserver.tar"),
			b.Bytes(),
			os.FileMode(0o644),
		), "Failed writing mock tarfile")
	}
	return tmpDir
}

func TestGetImageTags(t *testing.T) {
	mockDir := writeTestManifests(t)
	require.NotEmpty(t, mockDir)
	defer cleanupTmps(t, mockDir)

	// Now, call the release lib and try to extract the tags
	tagList, err := GetImageTags(mockDir)
	require.Nil(t, err)

	for arch, tags := range tagList {
		require.Equal(t, fmt.Sprintf("k8s.gcr.io/kube-apiserver-%s:v1.19.0-rc.4", arch), tags[0])
		require.Equal(t, fmt.Sprintf("gcr.io/k8s-staging-kubernetes/kube-apiserver-%s:v1.19.0-rc.4", arch), tags[1])
	}
}

func TestGetTarManifest(t *testing.T) {
	mockDir := writeTestManifests(t)
	require.NotEmpty(t, mockDir)
	defer cleanupTmps(t, mockDir)

	// Read the mock manifests and check we are reporting the data
	finfos, err := ioutil.ReadDir(filepath.Join(mockDir, ImagesPath))
	require.Nil(t, err, "reading mock monifests directory")
	for _, finfo := range finfos {
		manifest, err := GetTarManifest(filepath.Join(mockDir, ImagesPath, finfo.Name(), "kube-apiserver.tar"))
		require.Nil(t, err)
		switch finfo.Name() {
		case "amd64":
			require.Equal(t, "3da02591bd93f4db77a2ee5fb83f28315bb034657447168cfa1ce6161a446873.json", manifest.Config)
			require.Equal(t, 3, len(manifest.Layers), "checking number of layers read")
		case "arm64":
			require.Equal(t, "c47b48fe11f383c6e4f30fea7dbf507329d326e94524f8989328d3028a6bf5f5.json", manifest.Config)
			require.Equal(t, 2, len(manifest.Layers), "checking number of layers read")
		}
		require.Equal(t, fmt.Sprintf("k8s.gcr.io/kube-apiserver-%s:v1.19.0-rc.4", finfo.Name()), manifest.RepoTags[0])
		require.Equal(t, fmt.Sprintf("gcr.io/k8s-staging-kubernetes/kube-apiserver-%s:v1.19.0-rc.4", finfo.Name()), manifest.RepoTags[1])
	}
}

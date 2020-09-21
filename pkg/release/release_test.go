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

// Writes a couple of test doccker images
func writeTestImages(t *testing.T, testIDs map[string]string) (mockPath string) {
	tmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	// Sample manifests
	var mockManifests = map[string]string{
		"amd64": fmt.Sprintf(`[{"Config":"%s.json","RepoTags":["k8s.gcr.io/kube-apiserver-amd64:v1.19.0-rc.4","gcr.io/k8s-staging-kubernetes/kube-apiserver-amd64:v1.19.0-rc.4"],"Layers":["%s/layer.tar"]}]`, testIDs["config"], testIDs["layer"]),
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
			Mode: int64(os.FileMode(0o644)),
		}))
		_, err = fmt.Fprint(tw, manifest)
		require.Nil(t, err, "writing config file into mock image")
		require.Nil(t, err)

		// Let's do a more complete mock image for deeper tests
		if arch == "amd64" {
			// This image has a config file
			configFile := `{"architecture":"amd64","config":{"Hostname":"","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt"],"Cmd":null,"Image":"sha256:9548f17b268a45da87ea8cf5d2258f587dec9fad955b0c719861242f0a34c5b2","Volumes":null,"WorkingDir":"/","Entrypoint":["/go-runner"],"OnBuild":null,"Labels":{"description":"go based runner for distroless scenarios","maintainers":"Kubernetes Authors"}},"container_config":{"Hostname":"","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt"],"Cmd":["/bin/sh","-c","#(nop) COPY file:e745d995f524b7efd3899831de234e4598a16d667f0befa724fa0585c878c1a8 in /usr/local/bin/kube-apiserver "],"Image":"sha256:9548f17b268a45da87ea8cf5d2258f587dec9fad955b0c719861242f0a34c5b2","Volumes":null,"WorkingDir":"/","Entrypoint":["/go-runner"],"OnBuild":null,"Labels":{"description":"go based runner for distroless scenarios","maintainers":"Kubernetes Authors"}},"created":"2020-09-09T11:59:00.064230207Z","docker_version":"19.03.8","history":[{"created":"1970-01-01T00:00:00Z","author":"Bazel","created_by":"bazel build ..."},{"created":"2020-08-21T17:22:11.318097036Z","created_by":"LABEL maintainers=Kubernetes Authors","comment":"buildkit.dockerfile.v0","empty_layer":true},{"created":"2020-08-21T17:22:11.318097036Z","created_by":"LABEL description=go based runner for distroless scenarios","comment":"buildkit.dockerfile.v0","empty_layer":true},{"created":"2020-08-21T17:22:11.318097036Z","created_by":"WORKDIR /","comment":"buildkit.dockerfile.v0","empty_layer":true},{"created":"2020-08-21T17:22:11.318097036Z","created_by":"COPY /workspace/go-runner . # buildkit","comment":"buildkit.dockerfile.v0"},{"created":"2020-08-21T17:22:11.318097036Z","created_by":"ENTRYPOINT [\"/go-runner\"]","comment":"buildkit.dockerfile.v0","empty_layer":true},{"created":"2020-09-09T11:59:00.064230207Z","created_by":"/bin/sh -c #(nop) COPY file:e745d995f524b7efd3899831de234e4598a16d667f0befa724fa0585c878c1a8 in /usr/local/bin/kube-apiserver "}],"os":"linux","rootfs":{"type":"layers","diff_ids":["sha256:513643face35501b7b23d0c580bc9abea0d881b2ecc50cb9cb28f4ae58419552"]}}`
			require.Nil(t, tw.WriteHeader(&tar.Header{
				Name: testIDs["config"] + ".json",
				Size: int64(len(configFile)),
				Mode: int64(os.FileMode(0o644)),
			}))
			_, err = fmt.Fprint(tw, configFile)
			require.Nil(t, err, "wrinting config to image")

			// And also, let's write a layer to the image
			var c bytes.Buffer
			layerTar := tar.NewWriter(&c)
			msg := "# All your base are belong to us\n"
			require.Nil(t, layerTar.WriteHeader(&tar.Header{
				Name: "/README.md",
				Size: int64(len(msg)),
			}))
			_, err = fmt.Fprint(layerTar, msg)
			require.Nil(t, err, "writing layer file into mock image")
			require.Nil(t, err)

			// Add the layer to the image tar
			require.Nil(t, tw.WriteHeader(&tar.Header{
				Name: testIDs["layer"] + "/layer.tar",
				Size: int64(c.Len()),
				Mode: int64(os.FileMode(0o644)),
			}))
			_, err = c.WriteTo(tw)
			require.Nilf(t, err, "%s writing layer tar into image", err)
		}
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
	mockDir := writeTestImages(t, map[string]string{
		"config": "3da02591bd93f4db77a2ee5fb83f28315bb034657447168cfa1ce6161a446873",
		"layer":  "513643face35501b7b23d0c580bc9abea0d881b2ecc50cb9cb28f4ae58419552"})
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
	mockDir := writeTestImages(t, map[string]string{
		"config": "3da02591bd93f4db77a2ee5fb83f28315bb034657447168cfa1ce6161a446873",
		"layer":  "513643face35501b7b23d0c580bc9abea0d881b2ecc50cb9cb28f4ae58419552"})
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
			require.Equal(t, 1, len(manifest.Layers), "checking number of layers read")
		case "arm64":
			require.Equal(t, "c47b48fe11f383c6e4f30fea7dbf507329d326e94524f8989328d3028a6bf5f5.json", manifest.Config)
			require.Equal(t, 2, len(manifest.Layers), "checking number of layers read")
		}
		require.Equal(t, fmt.Sprintf("k8s.gcr.io/kube-apiserver-%s:v1.19.0-rc.4", finfo.Name()), manifest.RepoTags[0])
		require.Equal(t, fmt.Sprintf("gcr.io/k8s-staging-kubernetes/kube-apiserver-%s:v1.19.0-rc.4", finfo.Name()), manifest.RepoTags[1])
	}
}

func TestGetOCIManifest(t *testing.T) {
	configid := "3da02591bd93f4db77a2ee5fb83f28315bb034657447168cfa1ce6161a446873"
	layerid := "513643face35501b7b23d0c580bc9abea0d881b2ecc50cb9cb28f4ae58419552"

	// Note: These values are derived from the mocks created in writeTestImages()
	manifestSize := 2485
	layerSize := 545

	mockDir := writeTestImages(t, map[string]string{"config": configid, "layer": layerid})
	require.DirExists(t, mockDir)
	defer cleanupTmps(t, mockDir)

	// We test with the mock amd64 image, which is a bit more complete
	ocimanifest, err := GetOCIManifest(filepath.Join(mockDir, ImagesPath, "amd64", "kube-apiserver.tar"))
	require.Nil(t, err, "getting manifest from tarred image")
	require.NotEmpty(t, ocimanifest)
	require.EqualValues(t, manifestSize, ocimanifest.Config.Size)

	require.Equal(t, 1, len(ocimanifest.Layers))
	require.EqualValues(t, layerSize, ocimanifest.Layers[0].Size)
	require.Equal(t, "application/vnd.docker.image.rootfs.diff.tar.gzip", ocimanifest.Layers[0].MediaType)
}

func TestCopyBinaries(t *testing.T) {
	cwd, err := os.Getwd()
	require.Nil(t, err)

	testDir, err := ioutil.TempDir("", "test-copy-binaries-")
	require.Nil(t, err)
	require.Nil(t, os.Chdir(testDir))
	defer os.RemoveAll(testDir)

	for _, tc := range []struct {
		prepare  func() (rootPath string, cleanup func())
		validate func(error)
	}{
		{ // success client
			prepare: func() (string, func()) {
				tempDir, err := ioutil.TempDir("", "test-copy-binaries-")
				require.Nil(t, err)

				binDir := filepath.Join(
					tempDir, "client/linux/kubernetes/client/bin",
				)
				require.Nil(t, os.MkdirAll(binDir, os.FileMode(0o644)))

				for _, f := range []string{"1", "2", "3"} {
					_, err = os.Create(filepath.Join(binDir, f))
					require.Nil(t, err)
				}

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error) {
				require.Nil(t, err)

				binDir := filepath.Join(testDir, "bin/linux")
				require.FileExists(t, filepath.Join(binDir, "1"))
				require.FileExists(t, filepath.Join(binDir, "2"))
				require.FileExists(t, filepath.Join(binDir, "3"))
				dirContent, err := ioutil.ReadDir(binDir)
				require.Nil(t, err)
				require.Len(t, dirContent, 3)
			},
		},
		{ // success client skip non-dir
			prepare: func() (string, func()) {
				tempDir, err := ioutil.TempDir("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux/kubernetes/client/bin",
				), os.FileMode(0o755)))

				_, err = os.Create(filepath.Join(tempDir, "client/some-file"))
				require.Nil(t, err)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error) { require.Nil(t, err) },
		},
		{ // success server
			prepare: func() (string, func()) {
				tempDir, err := ioutil.TempDir("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux/kubernetes/client/bin",
				), os.FileMode(0o755)))
				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "server/linux/kubernetes/server/bin",
				), os.FileMode(0o755)))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error) { require.Nil(t, err) },
		},
		{ // success node
			prepare: func() (string, func()) {
				tempDir, err := ioutil.TempDir("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux/kubernetes/client/bin",
				), os.FileMode(0o755)))
				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "node/linux/kubernetes/node/bin",
				), os.FileMode(0o755)))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error) { require.Nil(t, err) },
		},
		{ // failure wrong server dir
			prepare: func() (string, func()) {
				tempDir, err := ioutil.TempDir("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux/kubernetes/client/bin",
				), os.FileMode(0o755)))
				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "server/linux/kubernetes/wrong/bin",
				), os.FileMode(0o755)))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error) { require.NotNil(t, err) },
		},
		{ // failure wrong node dir
			prepare: func() (string, func()) {
				tempDir, err := ioutil.TempDir("", "test-copy-binaries-")
				require.Nil(t, err)

				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "client/linux/kubernetes/client/bin",
				), os.FileMode(0o755)))
				require.Nil(t, os.MkdirAll(filepath.Join(
					tempDir, "node/linux/kubernetes/wrong/bin",
				), os.FileMode(0o755)))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			validate: func(err error) { require.NotNil(t, err) },
		},
		{ // empty dirs should error
			prepare:  func() (string, func()) { return "", func() {} },
			validate: func(err error) { require.NotNil(t, err) },
		},
	} {
		// Given
		rootPath, cleanup := tc.prepare()

		// When
		err := CopyBinaries(rootPath)

		// Then
		tc.validate(err)
		cleanup()
	}

	require.Nil(t, os.Chdir(cwd))
}

func TestWriteChecksums(t *testing.T) {
	for _, tc := range []struct {
		prepare  func() (rootPath string, cleanup func())
		validate func(err error, rootPath string)
	}{
		{ // success
			prepare: func() (rootPath string, cleanup func()) {
				tempDir, err := ioutil.TempDir("", "write-checksum-test-")
				require.Nil(t, err)

				for i, v := range []byte{1, 2, 4, 8, 16, 32, 64, 128} {
					require.Nil(t, ioutil.WriteFile(
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
						shaSums, err := ioutil.ReadFile(filepath.Join(
							rootPath, fmt.Sprintf("SHA%dSUMS", digest),
						))
						require.Nil(t, err)
						require.Contains(t, string(shaSums), expectedSha)

						sha, err := ioutil.ReadFile(filepath.Join(
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
				tempDir, err := ioutil.TempDir("", "write-checksum-test-")
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
				tempDir, err := ioutil.TempDir("", "write-checksum-test-")
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

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

package spdx

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/release-utils/util"
)

func TestUnitExtractTarballTmp(t *testing.T) {
	tar := writeTestTarball(t)
	require.NotNil(t, tar)
	defer os.Remove(tar.Name())

	sut := NewSPDX()
	_, err := sut.ExtractTarballTmp("lsdjkflskdjfl")
	require.NotNil(t, err)
	dir, err := sut.ExtractTarballTmp(tar.Name())
	require.Nil(t, err, "extracting file")
	defer os.RemoveAll(dir)
	require.True(t, util.Exists(filepath.Join(dir, "/text.txt")), "checking directory")
	require.True(t, util.Exists(filepath.Join(dir, "/subdir/text.txt")), "checking subdirectory")
	require.True(t, util.Exists(dir), "checking directory")

	// Check files
}

func TestReadArchiveManifest(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "sample-manifest-*.json")
	require.Nil(t, err)
	defer os.Remove(f.Name())
	require.Nil(t, os.WriteFile(
		f.Name(), []byte(sampleManifest), os.FileMode(0o644),
	), "writing test manifest file")

	sut := spdxDefaultImplementation{}
	_, err = sut.ReadArchiveManifest("laksdjlakjsdlkjsd")
	require.NotNil(t, err)
	manifest, err := sut.ReadArchiveManifest(f.Name())
	require.Nil(t, err)
	require.Equal(
		t, "386bcf5c63de46c7066c42d4ae1c38af0689836e88fed37d1dca2d484b343cf5.json",
		manifest.ConfigFilename,
	)
	require.Equal(t, 1, len(manifest.RepoTags))
	require.Equal(t, "k8s.gcr.io/kube-apiserver-amd64:v1.22.0-alpha.1", manifest.RepoTags[0])
	require.Equal(t, 3, len(manifest.LayerFiles))
	for i, fname := range []string{
		"23e140cb8e03a12cba4ac571d9a7143cf5e2e9b72de3b33ce3243b4f7ad6a188/layer.tar",
		"48dd73ececdf0f52a174ad33a469145824713bd2b73c6257ce1ba8502003ad4e/layer.tar",
		"d397673d78556210baa112013c960cb95a3fd452e5c4a2ead2b26e5a458cd87f/layer.tar",
	} {
		require.Equal(t, fname, manifest.LayerFiles[i])
	}
}

func TestPackageFromLayerTarBall(t *testing.T) {
	tar := writeTestTarball(t)
	require.NotNil(t, tar)
	defer os.Remove(tar.Name())

	sut := spdxDefaultImplementation{}
	_, err := sut.PackageFromLayerTarBall("lsdkjflksdjflk", &TarballOptions{})
	require.NotNil(t, err)
	pkg, err := sut.PackageFromLayerTarBall(tar.Name(), &TarballOptions{})
	require.Nil(t, err)
	require.NotNil(t, pkg)

	require.NotNil(t, pkg.Checksum)
	_, ok := pkg.Checksum["SHA256"]
	require.True(t, ok, "checking if sha256 checksum is set")
	_, ok = pkg.Checksum["SHA512"]
	require.True(t, ok, "checking if sha512 checksum is set")
	require.Equal(t, "5e75826e1baf84d5c5b26cc8fc3744f560ef0288c767f1cbc160124733fdc50e", pkg.Checksum["SHA256"])
	require.Equal(t, "f3b48a64a3d9db36fff10a9752dea6271725ddf125baf7026cdf09a2c352d9ff4effadb75da31e4310bc1b2513be441c86488b69d689353128f703563846c97e", pkg.Checksum["SHA512"])
}

func writeTestTarball(t *testing.T) *os.File {
	// Create a testdire
	tar, err := os.CreateTemp(os.TempDir(), "test-tar-*.tar.gz")
	require.Nil(t, err)

	tardata, err := base64.StdEncoding.DecodeString(testTar)
	require.Nil(t, err)

	reader := bytes.NewReader(tardata)
	zipreader, err := gzip.NewReader(reader)
	require.Nil(t, err)

	bindata, err := ioutil.ReadAll(zipreader)
	require.Nil(t, err)

	require.Nil(t, os.WriteFile(
		tar.Name(), bindata, os.FileMode(0o644)), "writing test tar file",
	)
	return tar
}

var testTar string = `H4sICPIFo2AAA2hlbGxvLnRhcgDt1EsKwjAUBdCMXUXcQPuS5rMFwaEraDGgUFpIE3D5puAPRYuD
VNR7Jm+QQh7c3hQly44Sa/U4hdV0O8+YUKTJkLRCMhKk0zHX+VdjLA6h9pyz6Ju66198N3H+pYpy
iM1273P+Bm/lX4mUvyQlkP8cLvkHdwhFOIQMd4wBG6Oe5y/1Xf6VNhXjlGGXB3+e/yY2O9e2PV/H
xvnOBTcsF59eCmZT5Cz+yXT/5bX/pMb3P030fw4rlB8AAAAAAAAAAAAAAAAA4CccAXRRwL4AKAAA
`

var sampleManifest = `[{"Config":"386bcf5c63de46c7066c42d4ae1c38af0689836e88fed37d1dca2d484b343cf5.json","RepoTags":["k8s.gcr.io/kube-apiserver-amd64:v1.22.0-alpha.1"],"Layers":["23e140cb8e03a12cba4ac571d9a7143cf5e2e9b72de3b33ce3243b4f7ad6a188/layer.tar","48dd73ececdf0f52a174ad33a469145824713bd2b73c6257ce1ba8502003ad4e/layer.tar","d397673d78556210baa112013c960cb95a3fd452e5c4a2ead2b26e5a458cd87f/layer.tar"]}]
`

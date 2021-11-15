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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/release-utils/util"
)

func TestBuildIDString(t *testing.T) {
	cases := []struct {
		seeds    []string
		expected string
	}{
		{[]string{"1234"}, "1234"},
		{[]string{"abc"}, "abc"},
		{[]string{"ABC"}, "ABC"},
		{[]string{"ABC", "123"}, "ABC-123"},
		{[]string{"Hello:bye", "123"}, "Hello-bye-123"},
		{[]string{"Hello^bye", "123"}, "Hellobye-123"},
		{[]string{"Hello:bye", "123", "&^%&$"}, "Hello-bye-123"},
	}
	for _, tc := range cases {
		require.Equal(t, tc.expected, buildIDString(tc.seeds...))
	}

	// If we do not pass any seeds, func should return an UUID
	// which is 36 chars long
	require.Len(t, buildIDString(), 36)

	// Same thing for only invalid chars
	require.Len(t, buildIDString("&^$&^%"), 36)
}

func TestUnitExtractTarballTmp(t *testing.T) {
	tarFile := writeTestTarball(t)
	require.NotNil(t, tarFile)
	defer os.Remove(tarFile.Name())

	sut := NewSPDX()
	_, err := sut.ExtractTarballTmp("lsdjkflskdjfl")
	require.NotNil(t, err)
	dir, err := sut.ExtractTarballTmp(tarFile.Name())
	require.Nil(t, err, "extracting file")
	defer os.RemoveAll(dir)

	require.True(t, util.Exists(filepath.Join(dir, "/text.txt")), "checking directory")
	require.True(t, util.Exists(filepath.Join(dir, "/subdir/text.txt")), "checking subdirectory")
	require.True(t, util.Exists(dir), "checking directory")
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

func TestPackageFromTarball(t *testing.T) {
	tarFile := writeTestTarball(t)
	require.NotNil(t, tarFile)
	defer os.Remove(tarFile.Name())

	sut := spdxDefaultImplementation{}
	_, err := sut.PackageFromTarball(&Options{}, &TarballOptions{}, "lsdkjflksdjflk")
	require.NotNil(t, err)
	pkg, err := sut.PackageFromTarball(&Options{}, &TarballOptions{}, tarFile.Name())
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

func TestExternalDocRef(t *testing.T) {
	cases := []struct {
		DocRef    ExternalDocumentRef
		StringVal string
	}{
		{ExternalDocumentRef{ID: "", URI: "", Checksums: map[string]string{}}, ""},
		{ExternalDocumentRef{ID: "", URI: "http://example.com/", Checksums: map[string]string{"SHA256": "d3b53860aa08e5c7ea868629800eaf78856f6ef3bcd4a2f8c5c865b75f6837c8"}}, ""},
		{ExternalDocumentRef{ID: "test-id", URI: "", Checksums: map[string]string{"SHA256": "d3b53860aa08e5c7ea868629800eaf78856f6ef3bcd4a2f8c5c865b75f6837c8"}}, ""},
		{ExternalDocumentRef{ID: "test-id", URI: "http://example.com/", Checksums: map[string]string{}}, ""},
		{
			ExternalDocumentRef{
				ID: "test-id", URI: "http://example.com/", Checksums: map[string]string{"SHA256": "d3b53860aa08e5c7ea868629800eaf78856f6ef3bcd4a2f8c5c865b75f6837c8"},
			},
			"DocumentRef-test-id http://example.com/ SHA256: d3b53860aa08e5c7ea868629800eaf78856f6ef3bcd4a2f8c5c865b75f6837c8",
		},
	}
	for _, tc := range cases {
		require.Equal(t, tc.StringVal, tc.DocRef.String())
	}
}

func TestExtDocReadSourceFile(t *testing.T) {
	// Create a known testfile
	f, err := os.CreateTemp("", "")
	require.Nil(t, err)
	require.Nil(t, os.WriteFile(f.Name(), []byte("Hellow World"), os.FileMode(0o644)))
	defer os.Remove(f.Name())

	ed := ExternalDocumentRef{}
	require.NotNil(t, ed.ReadSourceFile("/kjfhg/skjdfkjh"))
	require.Nil(t, ed.ReadSourceFile(f.Name()))
	require.NotNil(t, ed.Checksums)
	require.Equal(t, len(ed.Checksums), 1)
	require.Equal(t, "5f341d31f6b6a8b15bc4e6704830bf37f99511d1", ed.Checksums["SHA1"])
}

func writeTestTarball(t *testing.T) *os.File {
	// Create a testdir
	tarFile, err := os.CreateTemp(os.TempDir(), "test-tar-*.tar.gz")
	require.Nil(t, err)

	tardata, err := base64.StdEncoding.DecodeString(testTar)
	require.Nil(t, err)

	reader := bytes.NewReader(tardata)
	zipreader, err := gzip.NewReader(reader)
	require.Nil(t, err)

	bindata, err := ioutil.ReadAll(zipreader)
	require.Nil(t, err)

	require.Nil(t, os.WriteFile(
		tarFile.Name(), bindata, os.FileMode(0o644)), "writing test tar file",
	)
	return tarFile
}

func TestRelationshipRender(t *testing.T) {
	host := NewPackage()
	host.BuildID("TestHost")
	peer := NewFile()
	peer.BuildID("TestPeer")
	dummyref := "SPDXRef-File-6c0c16be41af1064ee8fd2328b17a0a778dd5e52"

	cases := []struct {
		Rel      Relationship
		MustErr  bool
		Rendered string
	}{
		{
			// Relationships with a full peer object have to render
			Relationship{FullRender: false, Type: DEPENDS_ON, Peer: peer},
			false, fmt.Sprintf("Relationship: %s DEPENDS_ON %s\n", host.SPDXID(), peer.SPDXID()),
		},
		{
			// Relationships with a remote reference
			Relationship{FullRender: false, Type: DEPENDS_ON, Peer: peer, PeerExtReference: "Remote"},
			false, fmt.Sprintf("Relationship: %s DEPENDS_ON DocumentRef-Remote:%s\n", host.SPDXID(), peer.SPDXID()),
		},
		{
			// Relationships without a full object, but
			// with a set reference must render
			Relationship{FullRender: false, PeerReference: dummyref, Type: DEPENDS_ON},
			false, fmt.Sprintf("Relationship: %s DEPENDS_ON %s\n", host.SPDXID(), dummyref),
		},
		{
			// Relationships without a object and without a set reference
			// must return an error
			Relationship{FullRender: false, Type: DEPENDS_ON}, true, "",
		},
		{
			// Relationships with a peer object withouth id should err
			Relationship{FullRender: false, Peer: &File{}, Type: DEPENDS_ON}, true, "",
		},
		{
			// Relationships with only a a peer reference that should render
			// in full should err
			Relationship{FullRender: true, PeerReference: dummyref, Type: DEPENDS_ON}, true, "",
		},
		{
			// Relationships without a type should err
			Relationship{FullRender: false, PeerReference: dummyref}, true, "",
		},
	}

	for _, tc := range cases {
		res, err := tc.Rel.Render(host)
		if tc.MustErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
			require.Equal(t, tc.Rendered, res)
		}
	}

	// Full rednering should not be the same as non full render
	nonFullRender, err := cases[0].Rel.Render(host)
	require.Nil(t, err)
	cases[0].Rel.FullRender = true
	fullRender, err := cases[0].Rel.Render(host)
	require.Nil(t, err)
	require.NotEqual(t, nonFullRender, fullRender)

	// Finally, rendering with a host objectwithout an ID should err
	_, err = cases[0].Rel.Render(&File{})
	require.NotNil(t, err)
}

var testTar = `H4sICPIFo2AAA2hlbGxvLnRhcgDt1EsKwjAUBdCMXUXcQPuS5rMFwaEraDGgUFpIE3D5puAPRYuD
VNR7Jm+QQh7c3hQly44Sa/U4hdV0O8+YUKTJkLRCMhKk0zHX+VdjLA6h9pyz6Ju66198N3H+pYpy
iM1273P+Bm/lX4mUvyQlkP8cLvkHdwhFOIQMd4wBG6Oe5y/1Xf6VNhXjlGGXB3+e/yY2O9e2PV/H
xvnOBTcsF59eCmZT5Cz+yXT/5bX/pMb3P030fw4rlB8AAAAAAAAAAAAAAAAA4CccAXRRwL4AKAAA
`

var sampleManifest = `[{"Config":"386bcf5c63de46c7066c42d4ae1c38af0689836e88fed37d1dca2d484b343cf5.json","RepoTags":["k8s.gcr.io/kube-apiserver-amd64:v1.22.0-alpha.1"],"Layers":["23e140cb8e03a12cba4ac571d9a7143cf5e2e9b72de3b33ce3243b4f7ad6a188/layer.tar","48dd73ececdf0f52a174ad33a469145824713bd2b73c6257ce1ba8502003ad4e/layer.tar","d397673d78556210baa112013c960cb95a3fd452e5c4a2ead2b26e5a458cd87f/layer.tar"]}]
`

func TestGetImageReferences(t *testing.T) {
	references, err := getImageReferences("k8s.gcr.io/kube-apiserver:v1.23.0-alpha.3")
	images := map[string]struct {
		arch string
		os   string
	}{
		"k8s.gcr.io/kube-apiserver@sha256:a82ca097e824f99bfb2b5107aa9c427633f9afb82afd002d59204f39ef81ae70": {"amd64", "linux"},
		"k8s.gcr.io/kube-apiserver@sha256:2a11e07f916b5982d9a6e3bbf5defd66ad50359c00b33862552063beb6981aec": {"arm", "linux"},
		"k8s.gcr.io/kube-apiserver@sha256:18f97b8c1c9b7b35dea7ba122d86e23066ce347aa8bb75b7346fed3f79d0ea21": {"arm64", "linux"},
		"k8s.gcr.io/kube-apiserver@sha256:1a61b61491042e2b1e659c4d57d426d01d9467fb381404bff029be4d00ead519": {"ppc64le", "linux"},
		"k8s.gcr.io/kube-apiserver@sha256:3e98f1591a5052791eec71d3c5f5d0fa913140992cb9e1d19fd80a158305c2ff": {"s390x", "linux"},
	}
	require.NoError(t, err)
	// This image should have 5 architectures
	require.Len(t, references, 5)
	for _, refData := range references {
		_, ok := images[refData.Digest]
		require.True(t, ok, fmt.Sprintf("Image not found %s", refData.Digest))
		require.Equal(t, images[refData.Digest].os, refData.OS)
		require.Equal(t, images[refData.Digest].arch, refData.Arch)
	}

	// Test a sha reference. This is the linux/ppc64le image
	singleRef := "k8s.gcr.io/kube-apiserver@sha256:1a61b61491042e2b1e659c4d57d426d01d9467fb381404bff029be4d00ead519"
	references, err = getImageReferences(singleRef)
	require.NoError(t, err)
	require.Len(t, references, 1)
	require.Equal(t, singleRef, references[0].Digest)

	// Tag with a single image. Image 1.0 is a single image
	references, err = getImageReferences("k8s.gcr.io/pause:1.0")
	require.NoError(t, err)
	require.Len(t, references, 1)
	require.Equal(t, "k8s.gcr.io/pause@sha256:a78c2d6208eff9b672de43f880093100050983047b7b0afe0217d3656e1b0d5f", references[0].Digest)
}

func TestPullImagesToArchive(t *testing.T) {
	impl := spdxDefaultImplementation{}

	// First. If the tag does not represent an image, expect an error
	_, err := impl.PullImagesToArchive("k8s.gcr.io/pause:0.0", "/tmp")
	require.Error(t, err)

	// Create a temp workdir
	dir, err := os.MkdirTemp("", "extract-image-")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// The pause 1.0 image is a single image
	images, err := impl.PullImagesToArchive("k8s.gcr.io/pause:1.0", dir)
	require.NoError(t, err)
	require.Len(t, images, 1)
	require.FileExists(t, filepath.Join(dir, "a78c2d6208eff9b672de43f880093100050983047b7b0afe0217d3656e1b0d5f.tar"))

	foundFiles := []string{}
	expectedFiles := []string{
		"sha256:350b164e7ae1dcddeffadd65c76226c9b6dc5553f5179153fb0e36b78f2a5e06",
		"a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4.tar.gz",
		"4964c72cd0245a7f77da38425dc98b472b2699ba6c49d5a9221fb32b972bc06b.tar.gz",
		"manifest.json",
	}
	tarFile, err := os.Open(filepath.Join(dir, "a78c2d6208eff9b672de43f880093100050983047b7b0afe0217d3656e1b0d5f.tar"))
	require.NoError(t, err)
	defer tarFile.Close()
	tarReader := tar.NewReader(tarFile)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		if header == nil {
			break
		}
		foundFiles = append(foundFiles, header.Name)
	}
	require.Equal(t, expectedFiles, foundFiles)
}

func TestGetDirectoryTree(t *testing.T) {
	dir, err := os.MkdirTemp("", "dir-tree-")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a directory. This same array elements should be read back
	// from the temporary directory.
	files := []string{"test.txt", "sub1/test2.txt", "sub2/subsub/test3.txt"}
	path := ""
	for _, f := range files {
		path = filepath.Join(dir, f)
		dir := filepath.Dir(path)
		require.NoError(t, os.MkdirAll(dir, os.FileMode(0o0755)))
		require.NoError(t, os.WriteFile(path, []byte("test"), os.FileMode(0o644)))
	}

	impl := spdxDefaultImplementation{}
	readFiles, err := impl.GetDirectoryTree(dir)
	require.NoError(t, err)
	// Now, compare contents of th array is the same
	require.ElementsMatch(t, files, readFiles)
}

func TestIgnorePatterns(t *testing.T) {
	dir, err := os.MkdirTemp("", "dir-tree-")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	impl := spdxDefaultImplementation{}

	// First, a dir without a gitignore should return no patterns, but not err
	p, err := impl.IgnorePatterns(dir, []string{}, false)
	require.NoError(t, err)
	require.Len(t, p, 0)

	// If we pass an extra pattern, we should get it back
	p, err = impl.IgnorePatterns(dir, []string{".vscode"}, false)
	require.NoError(t, err)
	require.Len(t, p, 1)

	// Now put a gitignore and read it back
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ".gitignore"),
		[]byte("# NFS\n.nfs*\n\n# OSX leaves these everywhere on SMB shares\n._*\n\n# OSX trash\n.DS_Store\n"),
		os.FileMode(0o755),
	))
	p, err = impl.IgnorePatterns(dir, nil, false)
	require.NoError(t, err)
	require.Len(t, p, 4)
}

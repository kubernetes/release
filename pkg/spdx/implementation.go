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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"archive/tar"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/release-utils/util"
)

//counterfeiter:generate . spdxImplementation

type spdxImplementation interface {
	ExtractTarballTmp(string) (string, error)
	ReadArchiveManifest(string) (*ArchiveManifest, error)
	PullImagesToArchive(string, string) error
	PackageFromLayerTarBall(string, *TarballOptions) (*Package, error)
}

type spdxDefaultImplementation struct{}

// ExtractTarballTmp extracts a tarball to a temporary directory
func (di *spdxDefaultImplementation) ExtractTarballTmp(tarPath string) (tmpDir string, err error) {
	tmpDir, err = os.MkdirTemp(os.TempDir(), "spdx-tar-extract-")
	if err != nil {
		return tmpDir, errors.Wrap(err, "creating temporary directory for tar extraction")
	}

	// Open the tar file
	f, err := os.Open(tarPath)
	if err != nil {
		return tmpDir, errors.Wrap(err, "opening tarball")
	}

	tr := tar.NewReader(f)
	numFiles := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return tmpDir, errors.Wrap(err, "reading the image tarfile")
		}

		if hdr.FileInfo().IsDir() {
			continue
		}

		if strings.HasPrefix(filepath.Base(hdr.FileInfo().Name()), ".wh") {
			logrus.Info("Skipping extraction of whiteout file")
			continue
		}

		if err := os.MkdirAll(
			filepath.Join(tmpDir, filepath.Dir(hdr.Name)), os.FileMode(0o755),
		); err != nil {
			return tmpDir, errors.Wrap(err, "creating image directory structure")
		}

		targetFile := filepath.Join(tmpDir, hdr.Name)
		f, err := os.Create(targetFile)
		if err != nil {
			return tmpDir, errors.Wrap(err, "creating image layer file")
		}
		defer f.Close()

		if _, err := io.Copy(f, tr); err != nil {
			return tmpDir, errors.Wrap(err, "extracting image data")
		}
		numFiles++
	}
	logrus.Infof("Successfully extracted %d files from image tarball %s", numFiles, tarPath)
	return tmpDir, err
}

// readArchiveManifest extracts the manifest json from an image tar
//    archive and returns the data as a struct
func (di *spdxDefaultImplementation) ReadArchiveManifest(manifestPath string) (manifest *ArchiveManifest, err error) {
	// Check that we have the archive manifest.json file
	if !util.Exists(manifestPath) {
		return manifest, errors.New("unable to find manifest file " + manifestPath)
	}

	// Parse the json file
	manifestData := []ArchiveManifest{}
	manifestJSON, err := os.ReadFile(manifestPath)
	if err != nil {
		return manifest, errors.Wrap(err, "unable to read from tarfile")
	}
	if err := json.Unmarshal(manifestJSON, &manifestData); err != nil {
		fmt.Println(string(manifestJSON))
		return manifest, errors.Wrap(err, "unmarshalling image manifest")
	}
	return &manifestData[0], nil
}

// PullImagesToArchive takes an image reference (a tag or a digest)
// and writes it into a docker tar archive in path
func (di *spdxDefaultImplementation) PullImagesToArchive(referenceString, path string) error {
	// Parse the string to get a reference (tag or digest)
	ref, err := name.ParseReference(referenceString)
	if err != nil {
		return errors.Wrapf(err, "parsing reference %s", referenceString)
	}

	// Build an image from the reference
	img, err := remote.Image(ref)
	if err != nil {
		return errors.Wrap(err, "getting image")
	}

	// This algo comes from crane:
	// Try to cast the reference as a tag:
	tag, ok := ref.(name.Tag)
	// if it fails
	if !ok {
		// .. and it is a digest
		d, ok := ref.(name.Digest)
		if !ok {
			return fmt.Errorf("reference is not a tag or digest")
		}
		// We add a mock tag
		tag = d.Repository.Tag("from-digest") // Append digest here?
	}

	return tarball.MultiWriteToFile(path, map[name.Tag]v1.Image{tag: img})
}

// PackageFromLayerTarBall builds a SPDX package from an image
//  tarball
func (di *spdxDefaultImplementation) PackageFromLayerTarBall(
	layerFile string, opts *TarballOptions,
) (*Package, error) {
	logrus.Infof("Generating SPDX package from layer in %s", layerFile)

	pkg := NewPackage()
	pkg.options.WorkDir = opts.ExtractDir
	if err := pkg.ReadSourceFile(filepath.Join(opts.ExtractDir, layerFile)); err != nil {
		return nil, errors.Wrap(err, "reading source file")
	}
	// Build the pkg name from its internal path
	h := sha1.New()
	if _, err := h.Write([]byte(layerFile)); err != nil {
		return nil, errors.Wrap(err, "hashing file path")
	}
	pkg.Name = fmt.Sprintf("%x", h.Sum(nil))

	return pkg, nil
}

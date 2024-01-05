/*
Copyright 2023 The Kubernetes Authors.

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

package specs

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/consts"
)

var obsArchitectures = map[string]string{
	"amd64":   "x86_64",
	"arm64":   "aarch64",
	"ppc64le": "ppc64le",
	"s390x":   "s390x",
}

// BuildArtifactsArchive downloads and archives artifacts from the given
// package source for all selected architectures. This archive is used as
// a source for artifacts by OpenBuildService when building the package.
func (s *Specs) BuildArtifactsArchive(pkgDef *PackageDefinition) error {
	if pkgDef == nil {
		return errors.New("package definition cannot be nil")
	}

	logrus.Infof("Downloading artifacts for %s %s...", pkgDef.Name, pkgDef.Version)

	for _, pkgVar := range pkgDef.Variations {
		arch := obsArchitectures[pkgVar.Architecture]
		logrus.Infof("Downloading %s %s (%s)...", pkgDef.Name, pkgDef.Version, arch)

		dlRootPath := filepath.Join(pkgDef.SpecOutputPath, pkgDef.Name, arch)
		err := s.impl.MkdirAll(dlRootPath, os.FileMode(0o755))
		if err != nil {
			if !s.impl.IsExist(err) {
				return fmt.Errorf("creating directory to download %s: %w", pkgDef.Name, err)
			}
		}

		logrus.Debugf("Saving downloaded artifacts to temporary location %s...", dlRootPath)

		var dlPath string
		var dlTarGz bool
		switch pkgDef.Name {
		case consts.PackageKubernetesCNI:
			dlPath = filepath.Join(dlRootPath, "kubernetes-cni.tar.gz")
			dlTarGz = true
		case consts.PackageCRITools:
			dlPath = filepath.Join(dlRootPath, "cri-tools.tar.gz")
			dlTarGz = true
		case consts.PackageCRIO:
			dlPath = filepath.Join(dlRootPath, "cri-o.tar.gz")
			dlTarGz = true
		default:
			dlPath = filepath.Join(dlRootPath, pkgDef.Name)
		}

		if err := s.DownloadArtifact(pkgVar.Source, dlPath, dlTarGz); err != nil {
			return fmt.Errorf("downloading artifacts: %w", err)
		}

		logrus.Infof("Successfully downloaded %s %s (%s).", pkgDef.Name, pkgDef.Version, arch)
	}

	logrus.Infof("Download completed successfully for %s %s.", pkgDef.Name, pkgDef.Version)
	logrus.Infof("Archiving artifacts for %s %s...", pkgDef.Name, pkgDef.Version)

	archiveSrc := filepath.Join(pkgDef.SpecOutputPath, pkgDef.Name)
	archiveDst := filepath.Join(pkgDef.SpecOutputPath, fmt.Sprintf("%s_%s.orig.tar.gz", pkgDef.Name, pkgDef.RPMVersion()))

	if err := s.impl.Compress(archiveDst, archiveSrc); err != nil {
		return fmt.Errorf("creating archive: %w", err)
	}
	if err := s.impl.RemoveAll(archiveSrc); err != nil {
		return fmt.Errorf("cleaning up archive source: %w", err)
	}

	logrus.Infof("Successfully archived binaries for %s %s to %s!", pkgDef.Name, pkgDef.Version, archiveDst)

	return nil
}

// DownloadArtifact is a wrapper function that runs appropriate download
// function depending if the package source URL scheme is gs:// or https://
func (s *Specs) DownloadArtifact(sourcePath, destPath string, extractTgz bool) error {
	if strings.HasPrefix(sourcePath, "gs://") {
		return s.DownloadArtifactFromGCS(sourcePath, destPath, extractTgz)
	}

	return s.DownloadArtifactFromURL(sourcePath, destPath, extractTgz)
}

// DownloadArtifactFromGCS downloads the artifact from the given GCS bucket.
func (s *Specs) DownloadArtifactFromGCS(sourcePath, destPath string, extractTgz bool) error {
	if err := s.impl.GCSCopyToLocal(sourcePath, destPath); err != nil {
		return fmt.Errorf("copying file to archive: %w", err)
	}

	if extractTgz {
		if err := s.impl.Extract(destPath, filepath.Dir(destPath)); err != nil {
			return fmt.Errorf("extracting .tar.gz archive: %w", err)
		}
		if err := s.impl.RemoveFile(destPath); err != nil {
			return fmt.Errorf("removing extracted archive: %w", err)
		}
	}

	return nil
}

// downloadArtifactFromGCS downloads the artifact from the given URL.
func (s *Specs) DownloadArtifactFromURL(downloadURL, destPath string, extractTgz bool) error {
	out, err := s.impl.CreateFile(destPath)
	if err != nil {
		return fmt.Errorf("creating download destination file: %w", err)
	}
	defer out.Close()

	resp, err := s.impl.GetRequest(downloadURL)
	if err != nil {
		return fmt.Errorf("downloading artifact: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading artifact: status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("writing downloaded artifact: %w", err)
	}

	if extractTgz {
		if err := s.impl.Extract(destPath, filepath.Dir(destPath)); err != nil {
			return fmt.Errorf("extracting .tar.gz archive: %w", err)
		}
		if err := s.impl.RemoveFile(destPath); err != nil {
			return fmt.Errorf("removing extracted archive: %w", err)
		}
	}

	return nil
}

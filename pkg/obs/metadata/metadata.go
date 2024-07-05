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

package metadata

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"sigs.k8s.io/yaml"
)

// PackageMetadata is a struct that contains the following information about a package:
// - URL from which to download artifacts needed to build the package
// - Indicator if artifacts are packed in a .tar.gz archive
// - Dependencies needed to install the package
// Package's metadata is versioned based on the given version constraint.
type PackageMetadata struct {
	// VersionConstraint is a semver range that defines the version of the package for which the metadata is valid.
	VersionConstraint string `json:"versionConstraint"`
	// SourceURLTemplate is a template for the URL from which to download artifacts needed to build the package.
	SourceURLTemplate string `json:"sourceURLTemplate"`
	// SourceTarGz is an indicator if artifacts are packed in a .tar.gz archive.
	SourceTarGz bool `json:"sourceTarGz"`
	// Dependencies is a list of dependencies needed to install the package.
	Dependencies []PackageDependency `json:"dependencies,omitempty"`
}

// PackageDependency is a struct that defines a single runtime dependency.
type PackageDependency struct {
	// Name is the name of the package.
	Name string `json:"name"`
	// VersionConstraint is a version constraint that's embedded in the spec file.
	VersionConstraint string `json:"versionConstraint"`
}

// PackageMetadataList is a map that represents metadata.yaml file.
type PackageMetadataList map[string][]PackageMetadata

// LoadPackageMetadata loads metadata.yaml file from the given path.
func LoadPackageMetadata(path string) (PackageMetadataList, error) {
	if path == "" {
		return nil, errors.New("path cannot be empty")
	}

	logrus.Infof("Loading metadata from %s...", path)

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading metadata file %q: %w", path, err)
	}

	var pkgDeps PackageMetadataList
	if err := yaml.Unmarshal(b, &pkgDeps); err != nil {
		return nil, fmt.Errorf("unmarshalling metadata file %q: %w", path, err)
	}

	logrus.Infof("Found metadata for %d packages.", len(pkgDeps))

	return pkgDeps, nil
}

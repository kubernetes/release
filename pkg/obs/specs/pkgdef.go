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
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	template "github.com/google/safetext/yamltemplate"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/consts"
	"k8s.io/release/pkg/obs/metadata"
)

// PackageDefinition represents a concrete package and stores package's name,
// version, and metadata.
type PackageDefinition struct {
	Name       string
	Version    string
	Revision   string
	Channel    string
	Metadata   *metadata.PackageMetadata
	Variations []PackageVariation

	SpecTemplatePath string
	SpecOutputPath   string
}

// PackageVariation is a variation of the same package. Variation currently
// represents a different architecture and source for the given architecture.
type PackageVariation struct {
	Architecture string
	Source       string
}

// RPMVersion returns version that's escaped to be a valid RPM package
// version. This function currently replaces "-" with "~" as described in
// the following document:
// https://docs.fedoraproject.org/en-US/packaging-guidelines/Versioning/#_handling_non_sorting_versions_with_tilde_dot_and_caret
func (p *PackageDefinition) RPMVersion() string {
	return strings.ReplaceAll(p.Version, "-", "~")
}

// ConstructPackageDefinition creates a new instance of PackageDefinition based
// on provided options.
func (s *Specs) ConstructPackageDefinition() (*PackageDefinition, error) {
	logrus.Infof("Constructing package definition for %s %s...", s.options.Package, s.options.Version)

	var err error

	pkgDef := &PackageDefinition{
		Name:       s.options.Package,
		Version:    s.options.Version,
		Revision:   s.options.Revision,
		Channel:    s.options.Channel,
		Variations: []PackageVariation{},

		SpecTemplatePath: s.options.SpecTemplatePath,
		SpecOutputPath:   s.options.SpecOutputPath,
	}

	logrus.Infof("Writing output to %s", pkgDef.SpecOutputPath)

	// If Kubernetes version is provided, ensure that it is correct and determine the channel based on it.
	// Otherwise, try to automatically determine the Kubernetes version based on provided channel.
	if consts.IsCoreKubernetesPackage(pkgDef.Name) && pkgDef.Version != "" {
		pkgDef.Channel, err = s.GetKubernetesChannelForVersion(pkgDef.Version)
		if err != nil {
			return nil, fmt.Errorf("getting kubernetes channel: %w", err)
		}
	} else if consts.IsCoreKubernetesPackage(pkgDef.Name) && pkgDef.Version == "" {
		pkgDef.Version, err = s.GetKubernetesVersionForChannel(pkgDef.Channel)
		if err != nil {
			return nil, fmt.Errorf("getting kubernetes version: %w", err)
		}
	}

	pkgDef.Version = s.impl.TrimTagPrefix(pkgDef.Version)
	// For cases where a CI build version of Kubernetes is retrieved, replace instances
	// of "+" with "-", so that we build with a valid Debian package version.
	pkgDef.Version = strings.Replace(pkgDef.Version, "+", "-", 1)

	logrus.Infof("Using %s version %s/%s", pkgDef.Name, pkgDef.Channel, pkgDef.Version)

	// Get package metadata for the given package. Metadata includes information about package source and dependencies.
	pkgDef.Metadata, err = s.GetPackageMetadata(pkgDef.SpecTemplatePath, pkgDef.Name, pkgDef.Version)
	if err != nil {
		return nil, fmt.Errorf("getting metadata for %q: %w", pkgDef.Name, err)
	}

	// Create variation of the package for each architecture.
	for _, arch := range s.options.Architectures {
		sourceURL, err := s.GetPackageSource(pkgDef.Metadata.SourceURLTemplate, s.options.PackageSourceBase, pkgDef.Name, pkgDef.Version, arch, pkgDef.Channel)
		if err != nil {
			return nil, fmt.Errorf("getting package source download link: %w", err)
		}
		pkgVar := PackageVariation{
			Architecture: arch,
			Source:       sourceURL,
		}
		pkgDef.Variations = append(pkgDef.Variations, pkgVar)
	}

	logrus.Infof("Successfully constructed package definition for %s %s!", pkgDef.Name, pkgDef.Version)

	return pkgDef, nil
}

// GetPackageMetadata gets metadata for the given package.
// Metadata includes information about package source and dependencies,
// and is stored in a YAML manifest.
func (s *Specs) GetPackageMetadata(templateDir, packageName, packageVersion string) (*metadata.PackageMetadata, error) {
	m, err := s.impl.LoadPackageMetadata(filepath.Join(templateDir, "metadata.yaml"))
	if err != nil {
		return nil, fmt.Errorf("getting metadata for %s: %w", packageName, err)
	}

	deps, err := s.GetMetadataWithVersionConstraint(packageName, packageVersion, m[packageName])
	if err != nil {
		return nil, fmt.Errorf("parsing metadata for %s: %w", packageName, err)
	}

	return deps, nil
}

// GetMetadataWithVersionConstraint parses metadata and takes metadata that
// matches the given version constraint.
func (s *Specs) GetMetadataWithVersionConstraint(packageName, packageVersion string, constraintedMetadata []metadata.PackageMetadata) (*metadata.PackageMetadata, error) {
	for _, m := range constraintedMetadata {
		r, err := semver.ParseRange(m.VersionConstraint)
		if err != nil {
			return nil, fmt.Errorf("parsing semver range for package %s: %w", packageName, err)
		}
		kubeSemVer, err := s.impl.TagStringToSemver(packageVersion)
		if err != nil {
			return nil, fmt.Errorf("parsing package version %s: %w", packageVersion, err)
		}

		if r(kubeSemVer) {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("package %s is not defined in metadata.yaml file", packageName)
}

// GetPackageSource gets the download link for artifacts for the given package.
// This function runs template on sourceURLTemplate defined in the metadata manifest.
func (s *Specs) GetPackageSource(templateBaseURL, baseURL, packageName, packageVersion, packageArch, channel string) (string, error) {
	data := struct {
		BaseURL        string
		PackageName    string
		PackageVersion string
		Architecture   string
		Channel        string
	}{
		BaseURL:        baseURL,
		PackageName:    packageName,
		PackageVersion: packageVersion,
		Architecture:   packageArch,
		Channel:        channel,
	}

	tpl, err := template.New("").Funcs(
		template.FuncMap{
			"KubernetesURL": s.GetKubernetesDownloadLink(channel, baseURL, packageName, packageVersion, packageArch),
		},
	).Parse(templateBaseURL)
	if err != nil {
		return "", fmt.Errorf("getting download link base: creating template: %w", err)
	}

	var outBuf bytes.Buffer
	if err := tpl.Execute(&outBuf, data); err != nil {
		return "", fmt.Errorf("getting download link base: executing template: %w", err)
	}

	return outBuf.String(), nil
}

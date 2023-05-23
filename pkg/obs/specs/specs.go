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
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/obs/metadata"
	"k8s.io/release/pkg/obs/options"
	"k8s.io/release/pkg/release"
	khttp "sigs.k8s.io/release-utils/http"
	"sigs.k8s.io/release-utils/util"
)

type Client struct {
	options *options.Options
	impl    Impl
}

func New(o *options.Options) *Client {
	return &Client{
		options: o,
		impl:    &impl{},
	}
}

func (c *Client) SetImpl(impl Impl) {
	c.impl = impl
}

type impl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Impl
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt specsfakes/fake_impl.go > specsfakes/_fake_impl.go && mv specsfakes/_fake_impl.go specsfakes/fake_impl.go"
type Impl interface {
	GetKubeVersion(versionType release.VersionType) (string, error)
	GetRequest(url string) (*http.Response, error)
}

func (i *impl) GetKubeVersion(versionType release.VersionType) (string, error) {
	return release.NewVersion().GetKubeVersion(versionType)
}

func (i *impl) GetRequest(url string) (*http.Response, error) {
	// TODO(xmudrii): Make timeout configurable.
	return khttp.NewAgent().WithTimeout(3 * time.Minute).GetRequest(url)
}

// PackageDefinition represents a concrete package and stores package's name, version, and metadata.
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

// PackageVariation is a variation of the same package. Variation currently represents a different architecture and
// source for the given architecture.
type PackageVariation struct {
	Architecture string
	Source       string
}

// ConstructPackageDefinition creates a new instance of PackageDefinition based on provided options.
func (c *Client) ConstructPackageDefinition() (*PackageDefinition, error) {
	logrus.Infof("Constructing package definition for %s %s...", c.options.Package, c.options.Version)

	var err error

	pkgDef := &PackageDefinition{
		Name:       c.options.Package,
		Version:    c.options.Version,
		Revision:   c.options.Revision,
		Channel:    c.options.Channel,
		Variations: []PackageVariation{},

		SpecTemplatePath: c.options.SpecTemplatePath,
		SpecOutputPath:   c.options.SpecOutputPath,
	}

	// Check if input and output directories exist.
	if _, err := os.Stat(pkgDef.SpecTemplatePath); err != nil {
		return nil, fmt.Errorf("finding package template dir: %w", err)
	}
	if _, err := os.Stat(pkgDef.SpecOutputPath); err != nil {
		return nil, fmt.Errorf("finding package output dir: %w", err)
	}

	logrus.Infof("Writing output to %s", pkgDef.SpecOutputPath)

	// If Kubernetes version is provided, ensure that it is correct and determine the channel based on it.
	// Otherwise, try to automatically determine the Kubernetes version based on provided channel.
	if isCoreKubernetesPackage(pkgDef.Name) && pkgDef.Version != "" {
		pkgDef.Channel, err = getKubernetesChannelForVersion(pkgDef.Version)
		if err != nil {
			return nil, fmt.Errorf("getting kubernetes channel: %w", err)
		}
	} else if isCoreKubernetesPackage(pkgDef.Name) && pkgDef.Version == "" {
		pkgDef.Version, err = c.getKubernetesVersionForChannel(pkgDef.Channel)
		if err != nil {
			return nil, fmt.Errorf("getting kubernetes version: %w", err)
		}
	}

	pkgDef.Version = util.TrimTagPrefix(pkgDef.Version)
	// For cases where a CI build version of Kubernetes is retrieved, replace instances
	// of "+" with "-", so that we build with a valid Debian package version.
	pkgDef.Version = strings.Replace(pkgDef.Version, "+", "-", 1)

	logrus.Infof("Using %s version %s/%s", pkgDef.Name, pkgDef.Channel, pkgDef.Version)

	// Get package metadata for the given package. Metadata includes information about package source and dependencies.
	pkgDef.Metadata, err = getPackageMetadata(pkgDef.SpecTemplatePath, pkgDef.Name, pkgDef.Version)
	if err != nil {
		return nil, fmt.Errorf("getting metadata for %q: %w", pkgDef.Name, err)
	}

	// Create variation of the package for each architecture.
	for _, arch := range c.options.Architectures {
		sourceURL, err := c.getPackageSource(pkgDef.Metadata.SourceURLTemplate, c.options.PackageSourceBase, pkgDef.Name, pkgDef.Version, arch, pkgDef.Channel)
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

// getPackageMetadata gets metadata for the given package.
// Metadata includes information about package source and dependencies, and is stored in a YAML manifest.
func getPackageMetadata(templateDir, packageName, packageVersion string) (*metadata.PackageMetadata, error) {
	m, err := metadata.LoadPackageMetadata(filepath.Join(templateDir, "metadata.yaml"))
	if err != nil {
		return nil, fmt.Errorf("getting metadata for %s: %w", packageName, err)
	}

	deps, err := getMetadataWithVersionConstraint(packageName, packageVersion, m[packageName])
	if err != nil {
		return nil, fmt.Errorf("parsing metadata for %s: %w", packageName, err)
	}

	return deps, nil
}

// getMetadataWithVersionConstraint parses metadata and takes metadata that matches the given version constraint.
func getMetadataWithVersionConstraint(packageName, packageVersion string, constraintedMetadata []metadata.PackageMetadata) (*metadata.PackageMetadata, error) {
	for _, m := range constraintedMetadata {
		r, err := semver.ParseRange(m.VersionConstraint)
		if err != nil {
			return nil, fmt.Errorf("parsing semver range for package %s: %w", packageName, err)
		}
		kubeSemVer, err := util.TagStringToSemver(packageVersion)
		if err != nil {
			return nil, fmt.Errorf("parsing package version %s: %w", packageVersion, err)
		}

		if r(kubeSemVer) {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("package %s is not defined in metadata.yaml file", packageName)
}

// getPackageSource gets the download link for artifacts for the given package.
// This function runs template on sourceURLTemplate defined in the metadata manifest.
func (c *Client) getPackageSource(
	templateBaseURL,
	baseURL,
	packageName,
	packageVersion,
	packageArch,
	channel string) (string, error) {
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
			"KubernetesURL": c.getKubernetesDownloadLink(channel, baseURL, packageName, packageVersion, packageArch),
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

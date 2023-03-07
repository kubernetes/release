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
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/obs/options"
	"k8s.io/release/pkg/release"
	khttp "sigs.k8s.io/release-utils/http"
	"sigs.k8s.io/release-utils/util"
)

type ChannelType string

const (
	ChannelRelease ChannelType = "release"
	ChannelTesting ChannelType = "testing"
	ChannelNightly ChannelType = "nightly"
)

const (
	minimumKubernetesVersion = "1.13.0"
	kubernetesCNIPackage     = "kubernetes-cni"
	criToolsPackage          = "cri-tools"
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

// PackageBuilder holds information about packages that we want to build.
// That includes type (deb or rpm), channel (stable/pre-releases/nightly),
// Kubernetes version, spec sources, and more.
// This struct also holds definitions for all packages.
type PackageBuilder struct {
	Type          options.PackageType
	Channel       ChannelType
	Architectures []string

	KubernetesVersion string
	TemplateDir       string
	OutputDir         string

	Definitions []*PackageDefinition

	DownloadLinkBase string
}

// PackageDefinition represents a concrete package. We store information such
// as its name, version, revision, and more.
type PackageDefinition struct {
	Name         string
	Version      string
	Revision     string
	Dependencies map[string]string
}

// ConstructPackageBuilder creates a new instance of PackageBuilder while
// populating all required information.
func (c *Client) ConstructPackageBuilder() (*PackageBuilder, error) {
	logrus.Infof("Constructing package builder...")
	var err error

	pb := &PackageBuilder{
		Type:          c.options.Type,
		Channel:       ChannelType(c.options.Channel),
		Architectures: c.options.Architectures,
		OutputDir:     c.options.OutputDir,

		KubernetesVersion: c.options.KubernetesVersion,

		Definitions: []*PackageDefinition{},

		DownloadLinkBase: c.options.ReleaseDownloadLinkBase,
	}

	// TODO: Get package directory for any version once package definitions are broken out
	pb.TemplateDir = filepath.Join(c.options.TemplateDir, string(c.options.Type))
	if _, err := os.Stat(pb.TemplateDir); err != nil {
		return nil, fmt.Errorf("finding package template dir: %w", err)
	}

	// If output directory is not provided, create a temporary one.
	if pb.OutputDir == "" {
		pb.OutputDir, err = os.MkdirTemp("", "obs-")
		if err != nil {
			return nil, fmt.Errorf("creating temporary dir: %w", err)
		}
		c.options.OutputDir = pb.OutputDir
	}

	logrus.Infof("Writing output to %s", pb.OutputDir)

	// If Kubernetes version is provided, ensure that it is correct and determine
	// the channel based on it. Otherwise, try to automatically determine the
	// Kubernetes version based on provided channel.
	if pb.KubernetesVersion != "" {
		logrus.Infof("Checking if user-supplied Kubernetes version is a valid semver...")
		kubeSemver, err := util.TagStringToSemver(pb.KubernetesVersion)
		if err != nil {
			return nil, fmt.Errorf("user-supplied kubernetes version is not valid semver: %w", err)
		}

		kubeVersionString := kubeSemver.String()
		kubeVersionParts := strings.Split(kubeVersionString, ".")

		switch {
		case len(kubeVersionParts) > 4:
			logrus.Info("User-supplied Kubernetes version is a CI version")
			logrus.Info("Setting channel to nightly")
			pb.Channel = ChannelNightly
		case len(kubeVersionParts) == 4:
			logrus.Info("User-supplied Kubernetes version is a pre-release version")
			logrus.Info("Setting channel to testing")
			pb.Channel = ChannelTesting
		default:
			logrus.Info("User-supplied Kubernetes version is a release version")
			logrus.Info("Setting channel to release")
			pb.Channel = ChannelRelease
		}
	}

	// This function determines the Kubernetes version if it's not provided.
	pb.KubernetesVersion, err = c.GetKubernetesVersion(pb.KubernetesVersion, pb.Channel)
	if err != nil {
		return nil, fmt.Errorf("getting kubernetes version: %w", err)
	}

	logrus.Infof("Using Kubernetes version %s", pb.KubernetesVersion)

	// This function gets download link base depending on what channel are
	// we using.
	pb.DownloadLinkBase, err = c.GetDownloadLinkBase(pb.KubernetesVersion, pb.Channel)
	if err != nil {
		return nil, fmt.Errorf("getting kubernetes download link base: %w", err)
	}

	logrus.Infof("Kubernetes download link base: %s", pb.DownloadLinkBase)

	// For cases where a CI build version of Kubernetes is retrieved, replace instances
	// of "+" with "-", so that we build with a valid Debian package version.
	pb.KubernetesVersion = strings.Replace(pb.KubernetesVersion, "+", "-", 1)

	logrus.Infof("Successfully constructed package builder!")

	return pb, nil
}

// ConstructPackageDefinitions creates instances of PackageDefinition based
// on what packages we selected to build.
func (c *Client) ConstructPackageDefinitions(pkgBuilder *PackageBuilder) error {
	logrus.Infof("Constructing package definitions...")
	var err error

	if pkgBuilder == nil {
		return errors.New("package builder cannot be nil")
	}

	for _, pkg := range c.options.Packages {
		pkgDef := &PackageDefinition{
			Name:         pkg,
			Revision:     c.options.Revision,
			Dependencies: map[string]string{},
		}

		// Determine the package version.
		pkgDef.Version, err = c.GetPackageVersion(pkgDef, pkgBuilder.KubernetesVersion)
		if err != nil {
			return fmt.Errorf("getting package %q version: %w", pkg, err)
		}

		logrus.Infof("%s package version: %s", pkgDef.Name, pkgDef.Version)

		// Determine dependencies for given package.
		pkgDef.Dependencies, err = GetDependencies(pkgDef)
		if err != nil {
			return fmt.Errorf("getting dependencies for %q: %w", pkg, err)
		}

		pkgBuilder.Definitions = append(pkgBuilder.Definitions, pkgDef)
	}

	logrus.Infof("Successfully constructed package definitions!")
	return nil
}

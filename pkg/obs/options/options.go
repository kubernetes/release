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

package options

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-utils/util"
)

// Options defines options for the package building process.
type Options struct {
	// Package is name of the package to build.
	Package string

	// Version is the package version.
	// For kubelet, kubeadm, kubectl, this is Kubernetes version.
	// For cri-tools, this is cri-tools version.
	// For kubernetes-cni, this is cni-plugins version.
	Version string

	// Revision is the package revision.
	Revision string

	// Channel is a release Channel that we're building packages for.
	// This can be one of: release, prerelease, nightly.
	Channel string

	// Architectures to download binaries for.
	// This can be one of: amd64, arm64, ppc64le, s390x.
	Architectures []string

	// PackageSourceBase is the base URL to download artifacts from.
	PackageSourceBase string

	// SpecTemplatePath is a path to a directory with spec template files.
	SpecTemplatePath string

	// SpecOutputPath is a path to a directory where to save spec files and archives.
	SpecOutputPath string

	// SpecOnly generates only spec files.
	SpecOnly bool
}

const (
	PackageCRITools      string = "cri-tools"
	PackageKubeadm       string = "kubeadm"
	PackageKubectl       string = "kubectl"
	PackageKubelet       string = "kubelet"
	PackageKubernetesCNI string = "kubernetes-cni"
)

const (
	ChannelTypeRelease    string = "release"
	ChannelTypePrerelease string = "prerelease"
	ChannelTypeNightly    string = "nightly"
)

const (
	ArchitectureAMD64 string = "amd64"
	ArchitectureARM64 string = "arm64"
	ArchitecturePPC64 string = "ppc64le"
	ArchitectureS390X string = "s390x"
)

var (
	supportedPackages = []string{
		PackageCRITools,
		PackageKubeadm,
		PackageKubectl,
		PackageKubelet,
		PackageKubernetesCNI,
	}
	supportedChannels = []string{
		ChannelTypeRelease,
		ChannelTypePrerelease,
		ChannelTypeNightly,
	}
	supportedArchitectures = []string{
		ArchitectureAMD64,
		ArchitectureARM64,
		ArchitecturePPC64,
		ArchitectureS390X,
	}
	// TODO(xmudrii): Remove this.
	defaultTemplateDir = filepath.Join(templateRootDir, "latest")
)

const (
	templateRootDir = "cmd/krel/templates/"

	DefaultReleaseDownloadLinkBase  = "gs://kubernetes-release/release"
	DefaultCNIDownloadLinkBase      = "gs://k8s-artifacts-cni/release"
	DefaultCRIToolsDownloadLinkBase = "gs://k8s-artifacts-cri-tools/release"

	defaultChannel  = ChannelTypeRelease
	defaultRevision = "0"
)

func New() *Options {
	return &Options{
		Revision:         defaultRevision,
		Channel:          defaultChannel,
		Architectures:    supportedArchitectures,
		SpecTemplatePath: defaultTemplateDir,
	}
}

// Validate verifies if all set options are valid
func (o *Options) Validate() error {
	if ok := isSupported("package", []string{o.Package}, supportedPackages); !ok {
		return fmt.Errorf("selected package is not supported")
	}
	if ok := isSupported("channel", []string{o.Channel}, supportedChannels); !ok {
		return fmt.Errorf("selected channel is not supported")
	}
	if ok := isSupported("architectures", o.Architectures, supportedArchitectures); !ok {
		return fmt.Errorf("architectures selection is not supported")
	}
	if o.Revision == "" {
		return fmt.Errorf("revision is required")
	}

	if o.SpecOutputPath != "" {
		if _, err := os.Stat(o.SpecOutputPath); err != nil {
			return fmt.Errorf("output dir doesn't exist")
		}
	}

	// Replace the "+" with a "-" to make it semver-compliant
	o.Version = util.TrimTagPrefix(o.Version)

	return nil
}

func isSupported(field string, input, expected []string) bool {
	notSupported := []string{}

	supported := false
	for _, i := range input {
		supported = false
		for _, j := range expected {
			if i == j {
				supported = true
				break
			}
		}

		if !supported {
			notSupported = append(notSupported, i)
		}
	}

	if len(notSupported) > 0 {
		logrus.Infof(
			"Flag %s has an unsupported option: %v", field, notSupported,
		)
		return false
	}

	return true
}

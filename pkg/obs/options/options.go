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
	"errors"
	"os"
	"path/filepath"
	"reflect"

	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-utils/util"
)

type Options struct {
	// Type currently can be only "rpm".
	Type PackageType

	// KubernetesVersion is Kubernetes version to build packages for.
	// If empty, krel will automatically take the latest version for
	// the selected channel.
	KubernetesVersion string

	// Packages to build.
	Packages []string

	// Architectures to build for.
	Architectures []string

	// Channel is a release Channel that we're building packages for
	// (e.g. release, testing, nightly).
	Channel string

	// Revision to use for building pac
	Revision string

	// CNIVersion is cni-plugins version to use for kubernetes-cni package.
	CNIVersion string

	// CRIToolsVersion is cri-tools version to use for the package.
	CRIToolsVersion string

	// ReleaseDownloadLinkBase is base URL for dl.k8s.io.
	ReleaseDownloadLinkBase string

	// TemplateDir is a path to a directory with spec template files.
	TemplateDir string

	// OutputDir is a path to a directory where to save spec files and archives.
	OutputDir string

	// SpecOnly generates only spec files.
	SpecOnly bool
}

var (
	supportedPackages = []string{
		"kubelet", "kubectl", "kubeadm", "kubernetes-cni", "cri-tools",
	}
	supportedChannels = []string{
		"release", "testing", "nightly",
	}
	supportedArchitectures = []string{
		"amd64", "arm", "arm64", "ppc64le", "s390x",
	}
	defaultTemplateDir = filepath.Join(templateRootDir, "latest")
)

const (
	templateRootDir = "cmd/krel/templates/"

	DefaultReleaseDownloadLinkBase = "https://dl.k8s.io"
	defaultChannel                 = "release"
	defaultRevision                = "0"
)

type PackageType string

const (
	PackageRPM PackageType = "rpm"
)

func New() *Options {
	return &Options{
		Type:                    PackageRPM,
		Revision:                defaultRevision,
		Packages:                supportedPackages,
		Channel:                 defaultChannel,
		Architectures:           supportedArchitectures,
		ReleaseDownloadLinkBase: DefaultReleaseDownloadLinkBase,
		TemplateDir:             defaultTemplateDir,
	}
}

// Validate verifies if all set options are valid
func (o *Options) Validate() error {
	if ok := isSupported(o.Packages, supportedPackages); !ok {
		return errors.New("package selections are not supported")
	}
	if ok := isSupported([]string{o.Channel}, supportedChannels); !ok {
		return errors.New("channel selections are not supported")
	}
	if ok := isSupported(o.Architectures, supportedArchitectures); !ok {
		return errors.New("architectures selections are not supported")
	}
	if o.OutputDir != "" {
		if _, err := os.Stat(o.OutputDir); err != nil {
			return errors.New("output dir doesn't exist")
		}
	}

	// Replace the "+" with a "-" to make it semver-compliant
	o.KubernetesVersion = util.TrimTagPrefix(o.KubernetesVersion)

	return nil
}

func isSupported(input, expected []string) bool {
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
			logrus.Infof(
				"Adding %q (type: %v) to not supported", i, reflect.TypeOf(i),
			)
			notSupported = append(notSupported, i)
		}
	}

	if len(notSupported) > 0 {
		logrus.Infof(
			"The following options are not supported: %v", notSupported,
		)
		return false
	}

	return true
}

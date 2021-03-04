/*
Copyright 2020 The Kubernetes Authors.

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
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-utils/util"
)

type Options struct {
	buildType BuildType

	revision        string
	kubeVersion     string
	cniVersion      string
	criToolsVersion string

	packages      []string
	channels      []string
	architectures []string

	releaseDownloadLinkBase string

	templateDir string
	specOnly    bool
}

type BuildType string

const (
	BuildDeb BuildType = "deb"
	BuildRpm BuildType = "rpm"
	BuildAll BuildType = "all"

	DefaultReleaseDownloadLinkBase = "https://dl.k8s.io"

	defaultRevision = "0"
	templateRootDir = "templates"
)

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
	latestTemplateDir = filepath.Join(templateRootDir, "latest")
)

func New() *Options {
	return &Options{
		revision:                defaultRevision,
		packages:                supportedPackages,
		channels:                supportedChannels,
		architectures:           supportedArchitectures,
		releaseDownloadLinkBase: DefaultReleaseDownloadLinkBase,
		templateDir:             latestTemplateDir,
	}
}

func (o *Options) WithBuildType(buildType BuildType) *Options {
	o.buildType = buildType
	return o
}

func (o *Options) WithRevision(revision string) *Options {
	o.revision = revision
	return o
}

func (o *Options) WithKubeVersion(kubeVersion string) *Options {
	o.kubeVersion = kubeVersion
	return o
}

func (o *Options) WithCNIVersion(cniVersion string) *Options {
	o.cniVersion = cniVersion
	return o
}

func (o *Options) WithCRIToolsVersion(criToolsVersion string) *Options {
	o.criToolsVersion = criToolsVersion
	return o
}

func (o *Options) WithPackages(packages ...string) *Options {
	o.packages = packages
	return o
}

func (o *Options) WithChannels(channels ...string) *Options {
	o.channels = channels
	return o
}

func (o *Options) WithArchitectures(architectures ...string) *Options {
	o.architectures = architectures
	return o
}

func (o *Options) WithReleaseDownloadLinkBase(releaseDownloadLinkBase string) *Options {
	o.releaseDownloadLinkBase = releaseDownloadLinkBase
	return o
}

func (o *Options) WithTemplateDir(templateDir string) *Options {
	o.templateDir = templateDir
	return o
}

func (o *Options) WithSpecOnly(specOnly bool) *Options {
	o.specOnly = specOnly
	return o
}

func (o *Options) BuildType() BuildType {
	return o.buildType
}

func (o *Options) Revision() string {
	return o.revision
}

func (o *Options) KubeVersion() string {
	return o.kubeVersion
}

func (o *Options) CNIVersion() string {
	return o.cniVersion
}

func (o *Options) CRIToolsVersion() string {
	return o.criToolsVersion
}

func (o *Options) Packages() []string {
	return o.packages
}

func (o *Options) Channels() []string {
	return o.channels
}

func (o *Options) Architectures() []string {
	return o.architectures
}

func (o *Options) ReleaseDownloadLinkBase() string {
	return o.releaseDownloadLinkBase
}

func (o *Options) TemplateDir() string {
	return o.templateDir
}

func (o *Options) SpecOnly() bool {
	return o.specOnly
}

// Validate verifies if all set options are valid
func (o *Options) Validate() error {
	if ok := isSupported(o.packages, supportedPackages); !ok {
		return errors.New("package selections are not supported")
	}
	if ok := isSupported(o.channels, supportedChannels); !ok {
		return errors.New("channel selections are not supported")
	}
	if ok := isSupported(o.architectures, supportedArchitectures); !ok {
		return errors.New("architectures selections are not supported")
	}

	// Replace the "+" with a "-" to make it semver-compliant
	o.kubeVersion = util.TrimTagPrefix(o.kubeVersion)

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

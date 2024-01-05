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
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/consts"
	"sigs.k8s.io/release-utils/util"
)

// Options defines options for generating specs and artifacts archive for
// the given package.
type Options struct {
	// Package is name of the package to generate specs and artifacts archive for.
	Package string

	// Version is the package version.
	// For kubelet, kubeadm, kubectl, this is Kubernetes version.
	// For cri-tools, kubernetes-cni and cri-o, this is their corresponding version.
	Version string

	// Revision is the package revision.
	Revision string

	// Architectures to download binaries for.
	// This can be one of: amd64, arm64, ppc64le, s390x.
	Architectures []string

	// Channel is a release Channel that we're building packages for.
	// It's used to determine the Kubernetes/package version if version is not
	// explicitly provided.
	// This can be one of: release, prerelease, nightly.
	// Omit for non-core Kubernetes packages.
	Channel string

	// PackageSourceBase is the base URL to download artifacts from.
	// Can be https:// or gs:// URL.
	PackageSourceBase string

	// SpecTemplatePath is a path to a directory with spec template files.
	SpecTemplatePath string

	// SpecOutputPath is a path to a directory where to save spec files and
	// archives. The directory must exist before running the command.
	SpecOutputPath string

	// SpecOnly generates only spec files without the artifacts archive.
	SpecOnly bool
}

// DefaultOptions returns a new Options instance.
func DefaultOptions() *Options {
	return &Options{
		Revision: consts.DefaultRevision,
		Architectures: []string{
			consts.ArchitectureAMD64,
			consts.ArchitectureARM64,
			consts.ArchitecturePPC64,
			consts.ArchitectureS390X,
		},
		Channel:          consts.ChannelTypeRelease,
		SpecOutputPath:   ".",
		SpecTemplatePath: consts.DefaultSpecTemplatePath,
	}
}

// String returns a string representation for the `Options` type.
func (o *Options) String() string {
	return fmt.Sprintf(
		"Package: %v, Version: %s-%s, Architectures: %s, Templates: %s, Output: %q",
		o.Package, o.Version, o.Revision, o.Architectures, o.SpecTemplatePath, o.SpecOutputPath,
	)
}

// Validate verifies if all parameters in the `Options` instance are valid.
func (o *Options) Validate() error {
	if _, err := os.Stat(filepath.Join(o.SpecTemplatePath, o.Package)); err != nil {
		return fmt.Errorf("specs for package %s doesn't exist", o.Package)
	}

	if o.Version == "" && o.Channel == "" {
		return errors.New("one of version or channel is required")
	}
	if o.Channel != "" {
		if ok := consts.IsSupported("channel", []string{o.Channel}, consts.SupportedChannels); !ok {
			return errors.New("selected channel is not supported")
		}
	}

	if o.Revision == "" {
		return errors.New("revision is required")
	}

	if ok := consts.IsSupported("architectures", o.Architectures, consts.SupportedArchitectures); !ok {
		return errors.New("architectures selection is not supported")
	}

	if _, err := os.Stat(o.SpecTemplatePath); err != nil {
		return errors.New("templates dir doesn't exist")
	}
	if _, err := os.Stat(o.SpecOutputPath); err != nil {
		return errors.New("output dir doesn't exist")
	}

	// Replace the "+" with a "-" to make it semver-compliant
	o.Version = util.TrimTagPrefix(o.Version)

	return nil
}

type Specs struct {
	options *Options
	impl
}

func New(opts *Options) *Specs {
	return &Specs{
		options: opts,
		impl:    &defaultImpl{},
	}
}

func (s *Specs) SetImpl(impl impl) {
	s.impl = impl
}

func (s *Specs) Run() error {
	pkgDef, err := s.ConstructPackageDefinition()
	if err != nil {
		return fmt.Errorf("constructing package definition: %w", err)
	}

	if err = s.BuildSpecs(pkgDef, s.options.SpecOnly); err != nil {
		return fmt.Errorf("building specs: %w", err)
	}

	if !s.options.SpecOnly {
		if err = s.BuildArtifactsArchive(pkgDef); err != nil {
			return fmt.Errorf("building artifacts archive: %w", err)
		}
	} else {
		logrus.Infof("Specs only option enabled, skipping artifacts archive for %s", s.options.Package)
	}

	return nil
}

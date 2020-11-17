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

package build

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"k8s.io/release/pkg/gcp/gcs"
	"k8s.io/release/pkg/release"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

var DefaultExtraVersionMarkers = []string{}

//counterfeiter:generate . client
type Client interface {
	// Primary build implementation
	Build() error
	Push() error

	// TODO: Consider adding functions for:
	//       - setting options (*Options, filesystem, object store, registry)
	//       - validating options
	//
	// TODO: Needed for anago
	//       Uncomment these as:
	//       - pkg/anago starts implement them
	//       - pkg/build/ci can support them
	//
	//       Let's also remove anything that isn't explicitly needed to build.
	//
	// CheckReleaseBucket() error
	// StageLocalArtifacts() error
	// PushReleaseArtifacts(srcPath, gcsPath string) error
	// PushContainerImages() error
	// CopyStagedFromGCS(stagedBucket, buildVersion string) error
	// StageLocalSourceTree(workDir, buildVersion string) error
}

type DefaultClient struct {
	opts *Options
	fs   afero.Fs

	// TODO: Uncomment once these are implemented
	// objectStore object.Store   // new krel pkg
	// registry    image.Registry // new krel pkg
}

func New(opts *Options) *DefaultClient {
	return &DefaultClient{
		opts: opts,
		fs:   opts.Fs,

		// TODO: Uncomment once these are implemented
		// objectStore: opts.Store,
		// registry:    opts.Registry,
	}
}

// Options are the main options to pass to `Client`.
type Options struct {
	// Filesystem interface
	Fs afero.Fs

	// Specify an alternate bucket for pushes (normally 'devel' or 'ci').
	Bucket string

	// Specify an alternate build directory. Will be automatically determined
	// if not set.
	BuildDir string

	// Used to make determinations on where to push artifacts
	// Can be overridden using `GCSRoot`.
	//
	// May be one of: 'devel', 'ci', 'release'
	BuildType string

	// If set, push docker images to specified registry/project.
	Registry string

	// Comma separated list which can be used to upload additional version
	// files to GCS. The path is relative and is append to a GCS path. (--ci
	// only).
	ExtraVersionMarkers []string

	// The top-level GCS directory builds will be released to.
	// If specified, it will override BuildType.
	//
	// When unset:
	//   - BuildType: "ci"
	//   - final path: gs://<bucket>/ci
	//
	// When set:
	//   - BuildType: "ci"
	//   - GCSRoot: "new-root"
	//   - final path: gs://<bucket>/new-root
	//
	// This option exists to handle the now-deprecated GCSSuffix option, which
	// was not plumbed through
	GCSRoot string

	// Version to be used. Usually automatically discovered, but it can be
	// used to overwrite this behavior.
	Version string

	// Append suffix to version name if set.
	VersionSuffix string

	// Used to configure options for CI builds.
	CI bool

	// Configure docker client for gcr.io authentication to allow communication
	// with non-public registries.
	ConfigureDocker bool

	// Specifies a fast build (linux/amd64 only).
	Fast bool

	// Do not exit error if the build already exists on the GCS path.
	AllowDup bool

	// Do not update the latest file.
	NoUpdateLatest bool

	// Do not mark published bits on GCS as publicly readable.
	PrivateBucket bool

	// Validate that the remove image digests exists, needs `skopeo` in
	// `$PATH`.
	ValidateRemoteImageDigests bool
}

// TODO: Refactor so that version is not required as a parameter
func (d *DefaultClient) GetGCSBuildPath(version string) (string, error) {
	if d.opts.Bucket == "" {
		d.SetBucket()
	}

	buildPath, err := gcs.GetReleasePath(
		d.opts.Bucket,
		d.opts.GCSRoot,
		version,
		d.opts.Fast,
	)
	if err != nil {
		return "", errors.Wrap(err, "get GCS release path")
	}

	return buildPath, nil
}

func (d *DefaultClient) SetBucket() {
	bucket := d.opts.Bucket
	if d.opts.Bucket == "" {
		if d.opts.CI {
			// TODO: Remove this once all CI and release jobs run on K8s Infra
			bucket = release.CIBucketLegacy
		}
	}

	d.opts.Bucket = bucket

	logrus.Infof("Bucket has been set to %s", d.opts.Bucket)
}

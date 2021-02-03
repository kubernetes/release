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

	"k8s.io/release/pkg/object"
	"k8s.io/release/pkg/release"
)

var DefaultExtraVersionMarkers = []string{}

// Instance is the main structure for creating and pushing builds.
type Instance struct {
	opts     *Options
	objStore object.GCS
}

// NewInstance can be used to create a new build `Instance`.
// TODO: Prefer functional options here instead
func NewInstance(opts *Options) *Instance {
	instance := &Instance{
		opts:     opts,
		objStore: *object.NewGCS(),
	}

	instance.setBuildType()
	instance.setBucket()
	instance.setGCSRoot()

	return instance
}

// Options are the main options to pass to `Instance`.
type Options struct {
	// Specify an alternate bucket for pushes (normally 'devel' or 'ci').
	Bucket string

	// Specify an alternate build directory (relative to RepoRoot). Will be automatically determined
	// if not set.
	BuildDir string

	// Used to make determinations on where to push artifacts
	// Can be overridden using `GCSRoot`.
	//
	// May be one of: 'devel', 'ci', 'release'
	BuildType string

	// If set, push docker images to specified registry/project.
	Registry string

	// Absolute path to the kubernetes repository root directory
	// defaults to current working directory
	RepoRoot string

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

	// Stage additional files defined by `ExtraGcpStageFiles` and
	// `ExtraWindowsStageFiles`, otherwise they will be skipped.
	StageExtraFiles bool
}

// TODO: Refactor so that version is not required as a parameter
func (bi *Instance) getGCSBuildPath(version string) (string, error) {
	if bi.opts.Bucket == "" {
		bi.setBucket()
	}

	buildPath, err := bi.objStore.GetReleasePath(
		bi.opts.Bucket,
		bi.opts.GCSRoot,
		version,
		bi.opts.Fast,
	)
	if err != nil {
		return "", errors.Wrap(err, "get GCS release path")
	}

	return buildPath, nil
}

func (bi *Instance) setBucket() {
	bucket := bi.opts.Bucket
	if bi.opts.Bucket == "" {
		if bi.opts.CI {
			// TODO: Remove this once all CI and release jobs run on K8s Infra
			bucket = release.CIBucketLegacy
		}
	}

	bi.opts.Bucket = bucket

	logrus.Infof("Bucket has been set to %s", bi.opts.Bucket)
}

func (bi *Instance) setBuildType() {
	buildType := "devel"
	if bi.opts.CI {
		buildType = "ci"
	}

	bi.opts.BuildType = buildType

	logrus.Infof("Build type has been set to %s", bi.opts.BuildType)
}

func (bi *Instance) setGCSRoot() {
	if bi.opts.GCSRoot == "" {
		bi.opts.GCSRoot = bi.opts.BuildType
	}

	logrus.Infof("GCS root has been set to %s", bi.opts.GCSRoot)
}

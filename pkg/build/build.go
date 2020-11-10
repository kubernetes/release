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
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var DefaultExtraVersionMarkers = []string{}

// Instance is the main structure for creating and pushing builds.
type Instance struct {
	opts *Options
}

// NewInstance can be used to create a new build `Instance`.
func NewInstance(opts *Options) *Instance {
	instance := &Instance{opts}
	instance.setBuildType()

	return instance
}

// Options are the main options to pass to `Instance`.
type Options struct {
	// Specify an alternate bucket for pushes (normally 'devel' or 'ci').
	Bucket string

	// Specify an alternate build directory. Will be automatically determined
	// if not set.
	BuildDir string

	// Used to make determinations on where to push artifacts
	// May be one of: 'devel', 'ci', 'release'
	BuildType string

	// If set, push docker images to specified registry/project.
	DockerRegistry string

	// Comma separated list which can be used to upload additional version
	// files to GCS. The path is relative and is append to a GCS path. (--ci
	// only).
	ExtraVersionMarkers []string

	// Specify a suffix to append to the upload destination on GCS.
	GCSSuffix string

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

// TODO: Support "release" buildType
func (bi *Instance) getGCSBuildPath(version string) string {
	gcsDest := bi.opts.BuildType

	if bi.opts.GCSSuffix != "" {
		gcsDest += "-" + bi.opts.GCSSuffix
	}

	if bi.opts.Fast {
		gcsDest = filepath.Join(gcsDest, "fast")
	}
	gcsDest = filepath.Join(gcsDest, version)
	logrus.Infof("GCS destination is %s", gcsDest)

	return gcsDest
}

func (bi *Instance) setBuildType() {
	buildType := "devel"
	if bi.opts.CI {
		buildType = "ci"
	}

	bi.opts.BuildType = buildType

	logrus.Infof("Build type is %s", bi.opts.BuildType)
}

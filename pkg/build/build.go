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

	"k8s.io/release/pkg/gcp/gcs"
	"k8s.io/release/pkg/release"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

var DefaultExtraVersionMarkers = []string{}

// Client is the main structure for creating and pushing builds.
type Client struct {
	opts *Options
	impl
}

// New can be used to create a new build `Client`.
func New(opts *Options) *Client {
	return &Client{opts, &defaultImpl{}}
}

// SetImpl can be used to set the internal implementation.
func (c *Client) SetOptions(opts *Options) {
	c.opts = opts
}

// SetImpl can be used to set the internal implementation.
func (c *Client) SetImpl(impl impl) {
	c.impl = impl
}

type defaultImpl struct{}

//counterfeiter:generate . impl
type impl interface {
	Push() error
	SetBucket()
}

func (d *defaultImpl) SetBucket() {
}

func (d *defaultImpl) Push() error {
	return nil
}

// Options are the main options to pass to `Client`.
type Options struct {
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
func (c *Client) GetGCSBuildPath(version string) (string, error) {
	if c.opts.Bucket == "" {
		c.SetBucket()
	}

	buildPath, err := gcs.GetReleasePath(
		c.opts.Bucket,
		c.opts.GCSRoot,
		version,
		c.opts.Fast,
	)
	if err != nil {
		return "", errors.Wrap(err, "get GCS release path")
	}

	return buildPath, nil
}

func (c *Client) SetBucket() {
	bucket := c.opts.Bucket
	if c.opts.Bucket == "" {
		if c.opts.CI {
			// TODO: Remove this once all CI and release jobs run on K8s Infra
			bucket = release.CIBucketLegacy
		}
	}

	c.opts.Bucket = bucket

	logrus.Infof("Bucket has been set to %s", c.opts.Bucket)
}

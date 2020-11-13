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

package anago

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/build"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

// releaseClient is a client for release a previously staged release.
//counterfeiter:generate . releaseClient
type releaseClient interface {
	// Validate if the provided `ReleaseOptions` are correctly set.
	ValidateOptions() error

	// CheckPrerequisites verifies that a valid GITHUB_TOKEN environment
	// variable is set. It also checks for the existence and version of
	// required packages and if the correct Google Cloud project is set. A
	// basic hardware check will ensure that enough disk space is available,
	// too.
	CheckPrerequisites() error

	// SetBuildCandidate discovers the release branch, parent branch (if
	// available) and build version for this release.
	SetBuildCandidate() error

	// GenerateReleaseVersion discovers the next versions to be released.
	GenerateReleaseVersion(parentBranch string) (*release.Versions, error)

	// PrepareWorkspace verifies that the working directory is in the desired
	// state. This means that the staged sources will be downloaded from the
	// bucket which should contain a copy of the repository.
	PrepareWorkspace() error

	// PushArtifacts pushes the generated artifacts to the release bucket and
	// Google Container Registry for the specified release `versions`.
	PushArtifacts(versions []string) error

	// PushGitObjects pushes the new tags and branches to the repository remote
	// on GitHub.
	PushGitObjects() error

	// CreateAnnouncement creates the release announcement mail and update the
	// GitHub release page to contain the artifacts and their checksums.
	CreateAnnouncement() error

	// Archive copies the release process logs to a bucket and sets private
	// permissions on it.
	Archive() error
}

// DefaultRelease is the default staging implementation used in production.
type DefaultRelease struct {
	impl    releaseImpl
	options *ReleaseOptions
}

// NewDefaultRelease creates a new defaultRelease instance.
func NewDefaultRelease(options *ReleaseOptions) *DefaultRelease {
	return &DefaultRelease{&defaultReleaseImpl{}, options}
}

// SetClient can be used to set the internal release implementation.
func (d *DefaultRelease) SetClient(impl releaseImpl) {
	d.impl = impl
}

// defaultReleaseImpl is the default internal release client implementation.
type defaultReleaseImpl struct{}

// releaseImpl is the implementation of the release client.
//counterfeiter:generate . releaseImpl
type releaseImpl interface {
	PrepareWorkspaceRelease(buildVersion, bucket string) error
	GenerateReleaseVersion(
		releaseType, version, branch string, branchFromMaster bool,
	) (*release.Versions, error)
	CheckReleaseBucket(options *build.Options) error
	CopyStagedFromGCS(
		options *build.Options, stagedBucket, buildVersion string,
	) error
	ValidateImages(registry, version, buildPath string) error
	PublishVersion(
		buildType, version, buildDir, bucket, gcsSuffix string,
		versionMarkers []string,
		privateBucket, fast bool,
	) error
}

func (d *defaultReleaseImpl) PrepareWorkspaceRelease(
	buildVersion, bucket string,
) error {
	if err := release.PrepareWorkspaceRelease(
		gitRoot, buildVersion, bucket,
	); err != nil {
		return err
	}
	return os.Chdir(gitRoot)
}

func (d *defaultReleaseImpl) GenerateReleaseVersion(
	releaseType, version, branch string, branchFromMaster bool,
) (*release.Versions, error) {
	return release.GenerateReleaseVersion(
		releaseType, version, branch, branchFromMaster,
	)
}

func (d *defaultReleaseImpl) CheckReleaseBucket(
	options *build.Options,
) error {
	return build.NewInstance(options).CheckReleaseBucket()
}

func (d *defaultReleaseImpl) CopyStagedFromGCS(
	options *build.Options, stagedBucket, buildVersion string,
) error {
	return build.NewInstance(options).
		CopyStagedFromGCS(stagedBucket, buildVersion)
}

func (d *defaultReleaseImpl) ValidateImages(
	registry, version, buildPath string,
) error {
	return release.NewImages().Validate(registry, version, buildPath)
}

func (d *defaultReleaseImpl) PublishVersion(
	buildType, version, buildDir, bucket, gcsSuffix string,
	versionMarkers []string,
	privateBucket, fast bool,
) error {
	return release.
		NewPublisher().
		PublishVersion("release", version, buildDir, bucket, gcsSuffix, nil, false, false)
}

func (d *DefaultRelease) ValidateOptions() error {
	return d.options.Validate()
}

func (d *DefaultRelease) CheckPrerequisites() error { return nil }

func (d *DefaultRelease) SetBuildCandidate() error { return nil }

func (d *DefaultRelease) GenerateReleaseVersion(
	parentBranch string,
) (*release.Versions, error) {
	return d.impl.GenerateReleaseVersion(
		d.options.ReleaseType,
		d.options.BuildVersion,
		d.options.ReleaseBranch,
		parentBranch == git.DefaultBranch,
	)
}

func (d *DefaultRelease) PrepareWorkspace() error {
	if err := d.impl.PrepareWorkspaceRelease(
		d.options.BuildVersion, d.options.Bucket(),
	); err != nil {
		return errors.Wrap(err, "prepare workspace")
	}
	return nil
}

func (d *DefaultRelease) PushArtifacts(versions []string) error {
	for _, version := range versions {
		logrus.Infof("Pushing artifacts for version %s", version)
		buildDir := filepath.Join(
			gitRoot, fmt.Sprintf("%s-%s", release.BuildDir, version),
		)
		bucket := d.options.Bucket()
		containerRegistry := d.options.ContainerRegistry()
		pushBuildOptions := &build.Options{
			Bucket:                     bucket,
			BuildDir:                   buildDir,
			Registry:                   containerRegistry,
			Version:                    version,
			AllowDup:                   true,
			ValidateRemoteImageDigests: true,
		}
		if err := d.impl.CheckReleaseBucket(pushBuildOptions); err != nil {
			return errors.Wrap(err, "check release bucket access")
		}

		if err := d.impl.CopyStagedFromGCS(
			pushBuildOptions, bucket, d.options.BuildVersion,
		); err != nil {
			return errors.Wrap(err, "copy staged from GCS")
		}

		// In an official nomock release, we want to ensure that container
		// images have been promoted from staging to production, so we do the
		// image manifest validation against production instead of staging.
		targetRegistry := containerRegistry
		if targetRegistry == release.GCRIOPathStaging {
			targetRegistry = release.GCRIOPathProd
		}

		// Image promotion has been done on nomock stage, verify that the
		// images are available.
		if err := d.impl.ValidateImages(
			targetRegistry, version, buildDir,
		); err != nil {
			return errors.Wrap(err, "validate container images")
		}

		if err := d.impl.PublishVersion(
			"release", version, buildDir, bucket, "", nil, false, false,
		); err != nil {
			return errors.Wrap(err, "publish release")
		}
	}
	return nil
}

func (d *DefaultRelease) PushGitObjects() error { return nil }

func (d *DefaultRelease) CreateAnnouncement() error { return nil }

func (d *DefaultRelease) Archive() error { return nil }

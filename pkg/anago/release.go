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

	"k8s.io/release/pkg/announce"
	"k8s.io/release/pkg/build"
	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/util"
)

// releaseClient is a client for release a previously staged release.
//counterfeiter:generate . releaseClient
type releaseClient interface {
	// Submit can be used to submit a Google Cloud Build (GCB) job.
	Submit() error

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
	GenerateReleaseVersion() error

	// PrepareWorkspace verifies that the working directory is in the desired
	// state. This means that the staged sources will be downloaded from the
	// bucket which should contain a copy of the repository.
	PrepareWorkspace() error

	// PushArtifacts pushes the generated artifacts to the release bucket and
	// Google Container Registry for the specified release `versions`.
	PushArtifacts() error

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
	state   *ReleaseState
}

// NewDefaultRelease creates a new defaultRelease instance.
func NewDefaultRelease(options *ReleaseOptions) *DefaultRelease {
	return &DefaultRelease{&defaultReleaseImpl{}, options, nil}
}

// SetImpl can be used to set the internal release implementation.
func (d *DefaultRelease) SetImpl(impl releaseImpl) {
	d.impl = impl
}

// SetState fixes the current state. Mainly used for passing
// arbitrary values during testing
func (d *DefaultRelease) SetState(state *ReleaseState) {
	d.state = state
}

// defaultReleaseImpl is the default internal release client implementation.
type defaultReleaseImpl struct{}

// releaseImpl is the implementation of the release client.
//counterfeiter:generate . releaseImpl
type releaseImpl interface {
	Submit(options *gcb.Options) error
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
		buildType, version, buildDir, bucket, gcsRoot string,
		versionMarkers []string,
		privateBucket, fast bool,
	) error
	CreateAnnouncement(
		options *announce.Options,
	) error
	PushTags(pusher *release.GitObjectPusher, tagList []string) error
	PushBranches(pusher *release.GitObjectPusher, branchList []string) error
	PushMainBranch(pusher *release.GitObjectPusher) error
	NewGitPusher(opts *release.GitObjectPusherOptions) (*release.GitObjectPusher, error)
}

func (d *defaultReleaseImpl) Submit(options *gcb.Options) error {
	return gcb.New(options).Submit()
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
	buildType, version, buildDir, bucket, gcsRoot string,
	versionMarkers []string,
	privateBucket, fast bool,
) error {
	return release.
		NewPublisher().
		PublishVersion("release", version, buildDir, bucket, gcsRoot, nil, false, false)
}

func (d *DefaultRelease) Submit() error {
	options := gcb.NewDefaultOptions()
	options.Release = true
	options.NoMock = d.options.NoMock
	options.Branch = d.options.ReleaseBranch
	options.ReleaseType = d.options.ReleaseType
	options.BuildVersion = d.options.BuildVersion
	options.NoAnago = true
	return d.impl.Submit(options)
}

func (d *defaultReleaseImpl) CreateAnnouncement(options *announce.Options) error {
	// Create the announcement
	return announce.CreateForRelease(options)
}

func (d *defaultReleaseImpl) PushTags(
	pusher *release.GitObjectPusher, tagList []string,
) error {
	return pusher.PushTags(tagList)
}

func (d *defaultReleaseImpl) PushBranches(
	pusher *release.GitObjectPusher, branchList []string,
) error {
	return pusher.PushBranches(branchList)
}

func (d *defaultReleaseImpl) PushMainBranch(pusher *release.GitObjectPusher) error {
	if err := pusher.PushMain(); err != nil {
		return errors.Wrap(err, "pushing changes in main branch")
	}
	return nil
}

// NewGitPusher returns a new instance of the git pusher to reuse
func (d *defaultReleaseImpl) NewGitPusher(
	opts *release.GitObjectPusherOptions,
) (pusher *release.GitObjectPusher, err error) {
	pusher, err = release.NewGitPusher(opts)
	if err != nil {
		return nil, errors.Wrap(err, "creating new git object pusher")
	}
	return pusher, nil
}

func (d *DefaultRelease) ValidateOptions() error {
	// Call options, validate. The validation returns the initial
	// state of the release process
	state, err := d.options.Validate()
	if err != nil {
		return errors.Wrap(err, "validating options")
	}
	d.state = state
	return nil
}

func (d *DefaultRelease) CheckPrerequisites() error { return nil }

func (d *DefaultRelease) SetBuildCandidate() error {
	// TODO: finish the implementation
	// d.state.createReleaseBranch = true
	return nil
}

func (d *DefaultRelease) GenerateReleaseVersion() error {
	versions, err := d.impl.GenerateReleaseVersion(
		d.options.ReleaseType,
		d.options.BuildVersion,
		d.options.ReleaseBranch,
		d.state.createReleaseBranch,
	)
	if err != nil {
		return errors.Wrap(err, "generating versions for release")
	}
	// Set the versions object in the state
	d.state.versions = versions
	return nil
}

func (d *DefaultRelease) PrepareWorkspace() error {
	if err := d.impl.PrepareWorkspaceRelease(
		d.options.BuildVersion, d.options.Bucket(),
	); err != nil {
		return errors.Wrap(err, "prepare workspace")
	}
	return nil
}

func (d *DefaultRelease) PushArtifacts() error {
	for _, version := range d.state.versions.Ordered() {
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
			"release", version, buildDir, bucket, "release", nil, false, false,
		); err != nil {
			return errors.Wrap(err, "publish release")
		}
	}
	return nil
}

// PushGitObjects uploads to the remote repository the release's tags and branches.
// Internally, this function calls the release implementation's PushTags,
// PushBranches and PushMainBranch methods
func (d *DefaultRelease) PushGitObjects() error {
	// Build the git object pusher
	pusher, err := d.impl.NewGitPusher(
		&release.GitObjectPusherOptions{
			DryRun: !d.options.NoMock,
			// MaxRetries: options.maxRetries,
			RepoPath: gitRoot,
		})
	if err != nil {
		return errors.Wrap(err, "getting git pusher from the release implementation")
	}

	// The list of tags to be pushed to the remote repository.
	// These come from the versions object created during
	// GenerateReleaseVersion()
	if err := d.impl.PushTags(pusher, d.state.versions.Ordered()); err != nil {
		return errors.Wrap(err, "pushing release tags")
	}

	// Determine which branches have to be pushed, except main
	// which gets pushed at the end by itself
	branchList := []string{}
	if d.options.ReleaseBranch != git.DefaultBranch {
		branchList = append(branchList, d.options.ReleaseBranch)
	}

	// Call the release imprementation PushBranches() method
	if err := d.impl.PushBranches(pusher, branchList); err != nil {
		return errors.Wrap(err, "pushing branches to the remote repository")
	}

	// For files created on master with new branches and
	// for $CHANGELOG_FILEPATH, update the main branch
	if err := d.impl.PushMainBranch(pusher); err != nil {
		return errors.Wrap(err, "pushing changes in main branch")
	}

	logrus.Infof(
		"Git objects push complete (%d branches, %d tags & main branch)",
		len(d.state.versions.Ordered()), len(branchList),
	)
	return nil
}

// CreateAnnouncement creates the announcement.html file
func (d *DefaultRelease) CreateAnnouncement() error {
	// Buld the announcement options set
	announceOpts := announce.NewOptions()
	announceOpts.WithWorkDir(gitRoot)
	announceOpts.WithTag(util.SemverToTagString(d.state.semverBuildVersion))
	announceOpts.WithBranch(d.options.ReleaseBranch)
	announceOpts.WithChangelogPath(
		filepath.Join(
			gitRoot, fmt.Sprintf("/CHANGELOG/CHANGELOG-%d.%d.md",
				d.state.semverBuildVersion.Major, d.state.semverBuildVersion.Minor,
			),
		),
	)
	announceOpts.WithChangelogHTML(
		filepath.Join(workspaceDir, "/src/release-notes.html"),
	)

	if err := d.impl.CreateAnnouncement(announceOpts); err != nil {
		return errors.Wrap(err, "creating the announcement")
	}
	return nil
}

func (d *DefaultRelease) Archive() error { return nil }

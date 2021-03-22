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

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/announce"
	"k8s.io/release/pkg/build"
	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/object"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-utils/log"
	"sigs.k8s.io/release-utils/util"
)

// releaseClient is a client for release a previously staged release.
//counterfeiter:generate . releaseClient
type releaseClient interface {
	// Submit can be used to submit a Google Cloud Build (GCB) job.
	Submit(stream bool) error

	// InitState initializes the default internal state.
	InitState()

	// InitLogFile sets up the log file target.
	InitLogFile() error

	// Validate if the provided `ReleaseOptions` are correctly set.
	ValidateOptions() error

	// CheckPrerequisites verifies that a valid GITHUB_TOKEN environment
	// variable is set. It also checks for the existence and version of
	// required packages and if the correct Google Cloud project is set. A
	// basic hardware check will ensure that enough disk space is available,
	// too.
	CheckPrerequisites() error

	// CheckReleaseBranchState discovers if the provided release branch has to
	// be created.
	CheckReleaseBranchState() error

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

	// UpdateGitHubPage updates the GitHub release page to with the source code
	// and release information.
	UpdateGitHubPage() error

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
	ToFile(fileName string) error
	CheckPrerequisites() error
	BranchNeedsCreation(
		branch, releaseType string, buildVersion semver.Version,
	) (bool, error)
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
	UpdateGitHubPage(options *announce.GitHubPageOptions) error
	PushTags(pusher *release.GitObjectPusher, tagList []string) error
	PushBranches(pusher *release.GitObjectPusher, branchList []string) error
	PushMainBranch(pusher *release.GitObjectPusher) error
	NewGitPusher(opts *release.GitObjectPusherOptions) (*release.GitObjectPusher, error)
	ArchiveRelease(options *release.ArchiverOptions) error
	NormalizePath(store object.Store, pathParts ...string) (string, error)
	CopyToRemote(store object.Store, src, gcsPath string) error
	PublishReleaseNotesIndex(
		gcsIndexRootPath, gcsReleaseNotesPath, version string,
	) error
	CreatePubBotBranchIssue(string) error
}

func (d *defaultReleaseImpl) Submit(options *gcb.Options) error {
	return gcb.New(options).Submit()
}

func (d *defaultReleaseImpl) ToFile(fileName string) error {
	return log.ToFile(fileName)
}

func (d *defaultReleaseImpl) CheckPrerequisites() error {
	return release.NewPrerequisitesChecker().Run(workspaceDir)
}

func (d *defaultReleaseImpl) BranchNeedsCreation(
	branch, releaseType string, buildVersion semver.Version,
) (bool, error) {
	return release.NewBranchChecker().NeedsCreation(
		branch, releaseType, buildVersion,
	)
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

func (d *DefaultRelease) Submit(stream bool) error {
	options := gcb.NewDefaultOptions()
	options.Stream = stream
	options.Release = true
	options.NoMock = d.options.NoMock
	options.Branch = d.options.ReleaseBranch
	options.ReleaseType = d.options.ReleaseType
	options.BuildVersion = d.options.BuildVersion
	return d.impl.Submit(options)
}

func (d *DefaultRelease) InitState() {
	d.state = &ReleaseState{DefaultState()}
}

func (d *DefaultRelease) InitLogFile() error {
	logrus.SetFormatter(
		&logrus.TextFormatter{FullTimestamp: true, ForceColors: true},
	)
	logFile := filepath.Join(os.TempDir(), "release.log")
	if err := d.impl.ToFile(logFile); err != nil {
		return errors.Wrap(err, "setup log file")
	}
	d.state.logFile = logFile
	logrus.Infof("Additionally logging to file %s", d.state.logFile)
	return nil
}

func (d *defaultReleaseImpl) CreateAnnouncement(options *announce.Options) error {
	// Create the announcement
	return announce.CreateForRelease(options)
}

func (d *defaultReleaseImpl) ArchiveRelease(options *release.ArchiverOptions) error {
	// Create a new release archiver
	return release.NewArchiver(options).ArchiveRelease()
}

func (d *defaultReleaseImpl) UpdateGitHubPage(options *announce.GitHubPageOptions) error {
	return announce.UpdateGitHubPage(options)
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

func (d *defaultReleaseImpl) NormalizePath(
	store object.Store, pathParts ...string,
) (string, error) {
	return store.NormalizePath(pathParts...)
}

func (d *defaultReleaseImpl) CopyToRemote(
	store object.Store, src, gcsPath string,
) error {
	return store.CopyToRemote(src, gcsPath)
}

func (d *defaultReleaseImpl) PublishReleaseNotesIndex(
	gcsIndexRootPath, gcsReleaseNotesPath, version string,
) error {
	return release.NewPublisher().PublishReleaseNotesIndex(
		gcsIndexRootPath, gcsReleaseNotesPath, version,
	)
}

func (d *defaultReleaseImpl) CreatePubBotBranchIssue(branchName string) error {
	return release.CreatePubBotBranchIssue(branchName)
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
	if err := d.options.Validate(d.state.State); err != nil {
		return errors.Wrap(err, "validating options")
	}
	return nil
}

func (d *DefaultRelease) CheckPrerequisites() error {
	return d.impl.CheckPrerequisites()
}

func (d *DefaultRelease) CheckReleaseBranchState() error {
	createReleaseBranch, err := d.impl.BranchNeedsCreation(
		d.options.ReleaseBranch,
		d.options.ReleaseType,
		d.state.semverBuildVersion,
	)
	if err != nil {
		return errors.Wrap(err, "check if release branch needs creation")
	}
	d.state.createReleaseBranch = createReleaseBranch
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
	const gcsRoot = "release"

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
			"release", version, buildDir, bucket, gcsRoot, nil, false, false,
		); err != nil {
			return errors.Wrap(err, "publish release")
		}
	}

	logrus.Info("Publishing release notes JSON")
	objStore := object.NewGCS()
	objStore.SetOptions(objStore.WithNoClobber(false))
	gcsReleaseRootPath, err := d.impl.NormalizePath(
		objStore, d.options.Bucket(), gcsRoot,
	)
	if err != nil {
		return errors.Wrap(err, "get GCS release root path")
	}

	gcsReleaseNotesPath := gcsReleaseRootPath + fmt.Sprintf(
		"/%s/release-notes.json", d.state.versions.Prime(),
	)

	if err := d.impl.CopyToRemote(
		objStore,
		releaseNotesJSONFile,
		gcsReleaseNotesPath,
	); err != nil {
		return errors.Wrap(err, "copy release notes to bucket")
	}

	logrus.Info("Publishing updated release notes index")
	if err := d.impl.PublishReleaseNotesIndex(
		gcsReleaseRootPath, gcsReleaseNotesPath, d.state.versions.Prime(),
	); err != nil {
		return errors.Wrap(err, "publish release notes index")
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

	// Workdir is where the announce files will be saved
	announceOpts.WithWorkDir(filepath.Join(workspaceDir, "/src"))

	// Get a semver from the prime tag
	primeSemver, err := util.TagStringToSemver(d.state.versions.Prime())
	if err != nil {
		return errors.Wrap(err, "parsing prime version into semver")
	}

	// The main tag we are releasing
	announceOpts.WithTag(d.state.versions.Prime())

	// Path to the changelog in the k/k repo (used to build a link to it)
	announceOpts.WithChangelogPath(
		fmt.Sprintf("CHANGELOG/CHANGELOG-%d.%d.md", primeSemver.Major, primeSemver.Minor),
	)

	// Pass the file path as a string to the annoucement options
	announceOpts.WithChangelogFile(releaseNotesHTMLFile)

	// Run the annoucement creation
	if err := d.impl.CreateAnnouncement(announceOpts); err != nil {
		return errors.Wrap(err, "creating the announcement")
	}

	// Check if we are releasing is the initial minor (eg 1.20.0),
	// and we are working on a release-M.m branch
	if primeSemver.Patch == 0 && len(primeSemver.Pre) == 0 &&
		d.options.ReleaseBranch != git.DefaultBranch {
		if d.options.NoMock {
			// Create the publishing bot issue
			if err := d.impl.CreatePubBotBranchIssue(d.options.ReleaseBranch); err != nil {
				// If it fails, log the error, but do not treat it
				// as fatal to avoid breaking the release process:
				logrus.Warn("Failed to create Publishing Bot Issue")
				logrus.Error(err)
			}
		} else {
			logrus.Info("Not creating publishing bot issue in mock release")
		}
	}
	return nil
}

// UpdateGitHubPage Update the GitHub release page, uploading the
// source code
func (d *DefaultRelease) UpdateGitHubPage() error {
	// Array of assets to be published in the release page
	assetList := []string{
		// Build the path to the kubernetes tar file:
		filepath.Join(
			gitRoot, // /workspace/src/k8s.io/kubernetes
			fmt.Sprintf("%s-%s", release.BuildDir, d.state.versions.Prime()), // _output-v1.20.0-beta.3/
			release.GCSStagePath,     // gcs-stage/
			d.state.versions.Prime(), // v1.20.0-beta.3
			release.KubernetesTar,    // kubernetes.tar.gz
		) + ":Kubernetes Source Code",
	}

	// URL to the changelog:
	changelogURL := fmt.Sprintf(
		"https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-%d.%d.md",
		d.state.semverBuildVersion.Major,
		d.state.semverBuildVersion.Minor,
	)

	// Build the options set for the GitHub page
	ghPageOpts := &announce.GitHubPageOptions{
		AssetFiles:            assetList,
		Tag:                   d.state.versions.Prime(),
		NoMock:                d.options.NoMock,
		UpdateIfReleaseExists: true,
		Name:                  "Kubernetes " + d.state.versions.Prime(),
		Draft:                 false,
		Owner:                 git.DefaultGithubOrg,
		Repo:                  git.DefaultGithubRepo,
		// PageTemplate: ,     // If we use a custom template, define it here
		Substitutions: map[string]string{
			"intro": "See [kubernetes-announce@](https://groups.google.com/forum/#!forum/kubernetes-announce). " +
				fmt.Sprintf("Additional binary downloads are linked in the [CHANGELOG](%s).", changelogURL),
			"changelog": changelogURL,
		},
	}
	// Update the release page (or simply output it during mock)
	if err := d.impl.UpdateGitHubPage(ghPageOpts); err != nil {
		return errors.Wrap(err, "updating GitHub release page")
	}
	return nil
}

// Archive stores the release artifact in a bucket along with
// its logs for long term conservation
func (d *DefaultRelease) Archive() error {
	// Create a new options set for the release archiver
	archiverOptions := &release.ArchiverOptions{
		ReleaseBuildDir: filepath.Join(workspaceDir, "src"),
		LogFile:         d.state.logFile,
		BuildVersion:    d.options.BuildVersion,
		PrimeVersion:    d.state.versions.Prime(),
		Bucket:          d.options.Bucket(),
	}

	if err := d.impl.ArchiveRelease(archiverOptions); err != nil {
		return errors.Wrap(err, "running the release archival process")
	}

	args := ""
	if d.options.NoMock {
		args += " --nomock"
	}
	args += " --tag=" + d.state.versions.Prime()

	logrus.Infof(
		"To announce this release, run:\n\n$ krel announce send%s", args,
	)

	return nil
}

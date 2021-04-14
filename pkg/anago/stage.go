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
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/build"
	"k8s.io/release/pkg/changelog"
	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-utils/log"
)

// stageClient is a client for staging releases.
//counterfeiter:generate . stageClient
type stageClient interface {
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
	// state. This means that the build directory is cleaned up and the checked
	// out repository is in a clean state.
	PrepareWorkspace() error

	// TagRepository creates all necessary git objects by tagging the
	// repository for the provided `versions` the main version `versionPrime`
	// and the `parentBranch`.
	TagRepository() error

	// Build runs 'make cross-in-a-container' by using the latest kubecross
	// container image. This step also build all necessary release tarballs.
	Build() error

	// GenerateChangelog builds the CHANGELOG-x.y.md file and commits it
	// into the local repository.
	GenerateChangelog() error

	// StageArtifacts copies the build artifacts to a Google Cloud Bucket.
	StageArtifacts() error
}

// DefaultStage is the default staging implementation used in production.
type DefaultStage struct {
	impl    stageImpl
	options *StageOptions
	state   *StageState
}

// NewDefaultStage creates a new defaultStage instance.
func NewDefaultStage(options *StageOptions) *DefaultStage {
	return &DefaultStage{&defaultStageImpl{}, options, nil}
}

// SetImpl can be used to set the internal stage implementation.
func (d *DefaultStage) SetImpl(impl stageImpl) {
	d.impl = impl
}

// SetState fixes the current state. Mainly used for passing
// arbitrary values during testing
func (d *DefaultStage) SetState(state *StageState) {
	d.state = state
}

// State returns the internal state.
func (d *DefaultStage) State() *StageState {
	return d.state
}

// defaultStageImpl is the default internal stage client implementation.
type defaultStageImpl struct{}

// stageImpl is the implementation of the stage client.
//counterfeiter:generate . stageImpl
type stageImpl interface {
	Submit(options *gcb.Options) error
	ToFile(fileName string) error
	CheckPrerequisites() error
	BranchNeedsCreation(
		branch, releaseType string, buildVersion semver.Version,
	) (bool, error)
	PrepareWorkspaceStage() error
	GenerateReleaseVersion(
		releaseType, version, branch string, branchFromMaster bool,
	) (*release.Versions, error)
	OpenRepo(repoPath string) (*git.Repo, error)
	RevParse(repo *git.Repo, rev string) (string, error)
	RevParseTag(repo *git.Repo, rev string) (string, error)
	Checkout(repo *git.Repo, rev string, args ...string) error
	CurrentBranch(repo *git.Repo) (string, error)
	CommitEmpty(repo *git.Repo, msg string) error
	Tag(repo *git.Repo, name, message string) error
	CheckReleaseBucket(options *build.Options) error
	DockerHubLogin() error
	MakeCross(version string) error
	GenerateChangelog(options *changelog.Options) error
	StageLocalSourceTree(
		options *build.Options, workDir, buildVersion string,
	) error
	StageLocalArtifacts(options *build.Options) error
	PushReleaseArtifacts(
		options *build.Options, srcPath, gcsPath string,
	) error
	PushContainerImages(options *build.Options) error
}

func (d *defaultStageImpl) Submit(options *gcb.Options) error {
	return gcb.New(options).Submit()
}

func (d *defaultStageImpl) ToFile(fileName string) error {
	return log.ToFile(fileName)
}

func (d *defaultStageImpl) CheckPrerequisites() error {
	return release.NewPrerequisitesChecker().Run(workspaceDir)
}

func (d *defaultStageImpl) BranchNeedsCreation(
	branch, releaseType string, buildVersion semver.Version,
) (bool, error) {
	return release.NewBranchChecker().NeedsCreation(
		branch, releaseType, buildVersion,
	)
}

func (d *defaultStageImpl) PrepareWorkspaceStage() error {
	if err := release.PrepareWorkspaceStage(gitRoot); err != nil {
		return err
	}
	return os.Chdir(gitRoot)
}

func (d *defaultStageImpl) GenerateReleaseVersion(
	releaseType, version, branch string, branchFromMaster bool,
) (*release.Versions, error) {
	return release.GenerateReleaseVersion(
		releaseType, version, branch, branchFromMaster,
	)
}

func (d *defaultStageImpl) OpenRepo(repoPath string) (*git.Repo, error) {
	return git.OpenRepo(repoPath)
}

func (d *defaultStageImpl) RevParse(repo *git.Repo, rev string) (string, error) {
	return repo.RevParse(rev)
}

func (d *defaultStageImpl) RevParseTag(repo *git.Repo, rev string) (string, error) {
	return repo.RevParseTag(rev)
}

func (d *defaultStageImpl) Checkout(repo *git.Repo, rev string, args ...string) error {
	return repo.Checkout(rev, args...)
}

func (d *defaultStageImpl) CurrentBranch(repo *git.Repo) (string, error) {
	return repo.CurrentBranch()
}

func (d *defaultStageImpl) CommitEmpty(repo *git.Repo, msg string) error {
	return repo.CommitEmpty(msg)
}

func (d *defaultStageImpl) Tag(repo *git.Repo, name, message string) error {
	return repo.Tag(name, message)
}

func (d *defaultStageImpl) MakeCross(version string) error {
	return build.NewMake().MakeCross(version)
}

func (d *defaultStageImpl) DockerHubLogin() error {
	return release.DockerHubLogin()
}

func (d *defaultStageImpl) GenerateChangelog(options *changelog.Options) error {
	return changelog.New(options).Run()
}

func (d *defaultStageImpl) CheckReleaseBucket(
	options *build.Options,
) error {
	return build.NewInstance(options).CheckReleaseBucket()
}

func (d *defaultStageImpl) StageLocalSourceTree(
	options *build.Options, workDir, buildVersion string,
) error {
	return build.NewInstance(options).StageLocalSourceTree(workDir, buildVersion)
}

func (d *defaultStageImpl) StageLocalArtifacts(
	options *build.Options,
) error {
	return build.NewInstance(options).StageLocalArtifacts()
}

func (d *defaultStageImpl) PushReleaseArtifacts(
	options *build.Options, srcPath, gcsPath string,
) error {
	return build.NewInstance(options).PushReleaseArtifacts(srcPath, gcsPath)
}

func (d *defaultStageImpl) PushContainerImages(
	options *build.Options,
) error {
	return build.NewInstance(options).PushContainerImages()
}

func (d *DefaultStage) Submit(stream bool) error {
	options := gcb.NewDefaultOptions()
	options.Stream = stream
	options.Stage = true
	options.NoMock = d.options.NoMock
	options.Branch = d.options.ReleaseBranch
	options.ReleaseType = d.options.ReleaseType
	return d.impl.Submit(options)
}

func (d *DefaultStage) InitLogFile() error {
	logrus.SetFormatter(
		&logrus.TextFormatter{FullTimestamp: true, ForceColors: true},
	)
	logFile := filepath.Join(os.TempDir(), "stage.log")
	if err := d.impl.ToFile(logFile); err != nil {
		return errors.Wrap(err, "setup log file")
	}
	d.state.logFile = logFile
	logrus.Infof("Additionally logging to file %s", d.state.logFile)
	return nil
}

func (d *DefaultStage) InitState() {
	d.state = &StageState{DefaultState()}
}

func (d *DefaultStage) ValidateOptions() error {
	if err := d.options.Validate(d.state.State); err != nil {
		return errors.Wrap(err, "validating options")
	}
	return nil
}

func (d *DefaultStage) CheckPrerequisites() error {
	return d.impl.CheckPrerequisites()
}

func (d *DefaultStage) CheckReleaseBranchState() error {
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

func (d *DefaultStage) GenerateReleaseVersion() error {
	versions, err := d.impl.GenerateReleaseVersion(
		d.options.ReleaseType,
		d.options.BuildVersion,
		d.options.ReleaseBranch,
		d.state.createReleaseBranch,
	)
	if err != nil {
		return errors.Wrap(err, "generating release versions for stage")
	}
	// Set the versions on the state
	d.state.versions = versions
	return nil
}

func (d *DefaultStage) PrepareWorkspace() error {
	if err := d.impl.PrepareWorkspaceStage(); err != nil {
		return errors.Wrap(err, "prepare workspace")
	}
	return nil
}

func (d *DefaultStage) TagRepository() error {
	repo, err := d.impl.OpenRepo(gitRoot)
	if err != nil {
		return errors.Wrap(err, "open Kubernetes repository")
	}

	for _, version := range d.state.versions.Ordered() {
		logrus.Infof("Preparing version %s", version)

		// Ensure that the tag not already exists
		if _, err := d.impl.RevParseTag(repo, version); err == nil {
			return errors.Errorf("tag %s already exists", version)
		}

		// Usually the build version contains a commit we can reference. If
		// not, because the build version is exactly a tag, then we fallback to
		// that tag.
		commit := d.options.BuildVersion
		if len(d.state.semverBuildVersion.Build) > 0 {
			commit = d.state.semverBuildVersion.Build[0]
		}

		if d.state.createReleaseBranch {
			logrus.Infof("Creating release branch %s", d.options.ReleaseBranch)

			if version == d.state.versions.Prime() {
				logrus.Infof("Version %s is the prime version", version)
				logrus.Infof(
					"Creating release branch %s from commit %s",
					d.options.ReleaseBranch, commit,
				)
				if err := d.impl.Checkout(
					repo, "-b", d.options.ReleaseBranch, commit,
				); err != nil {
					return errors.Wrap(err, "create new release branch")
				}
			} else {
				logrus.Infof(
					"Version %s is not the prime, checking out %s branch",
					version, git.DefaultBranch,
				)
				if err := d.impl.Checkout(repo, git.DefaultBranch); err != nil {
					return errors.Wrapf(err, "checkout %s branch", git.DefaultBranch)
				}
			}
		} else {
			logrus.Infof("Checking out branch %s", d.options.ReleaseBranch)
			if err := d.impl.Checkout(repo, d.options.ReleaseBranch); err != nil {
				return errors.Wrapf(err, "checking out branch %s", d.options.ReleaseBranch)
			}
		}

		// `branch == ""` in case we checked out a commit directly, which is
		// then in detached head state.
		branch, err := d.impl.CurrentBranch(repo)
		if err != nil {
			return errors.Wrap(err, "get current branch")
		}
		if branch != "" {
			logrus.Infof("Current branch is %s", branch)
		}

		// For release branches, we create an empty release commit to avoid
		// potential ambiguous 'git describe' logic between the official
		// release, 'x.y.z' and the next beta of that release branch,
		// 'x.y.(z+1)-beta.0'.
		//
		// We avoid doing this empty release commit on 'master', as:
		//   - there is a potential for branch conflicts as upstream/master
		//     moves ahead
		//   - we're checking out a git ref, as opposed to a branch, which
		//     means the tag will detached from 'upstream/master'
		//
		// A side-effect of the tag being detached from 'master' is the primary
		// build job (ci-kubernetes-build) will build as the previous alpha,
		// instead of the assumed tag. This causes the next anago run against
		// 'master' to fail due to an old build version.
		//
		// Example: 'v1.18.0-alpha.2.663+df908c3aad70be'
		//          (should instead be:
		//			 'v1.18.0-alpha.3.<commits-since-tag>+<commit-ish>')
		//
		// ref:
		//   - https://github.com/kubernetes/release/issues/1020
		//   - https://github.com/kubernetes/release/pull/1030
		//   - https://github.com/kubernetes/release/issues/1080
		//   - https://github.com/kubernetes/kubernetes/pull/88074

		// When tagging a release branch, always create an empty commit:
		if strings.HasPrefix(branch, "release-") {
			logrus.Infof("Creating empty release commit for tag %s", version)
			if err := d.impl.CommitEmpty(
				repo,
				fmt.Sprintf("Release commit for Kubernetes %s", version),
			); err != nil {
				return errors.Wrap(err, "create empty release commit")
			}
		}

		// If we are on master/main we do not create an empty commit,
		// but we detach the head at the specified commit to avoid having
		// commits merged between the BuildVersion commit and the tag:
		if branch != "" && !strings.HasPrefix(branch, "release-") {
			logrus.Infof("Detaching HEAD at commit %s to create tag %s", commit, version)
			if err := d.impl.Checkout(repo, commit); err != nil {
				return errors.Wrap(err, "checkout release commit")
			}
		}

		// Tag the repository:
		logrus.Infof("Tagging version %s", version)
		if err := d.impl.Tag(
			repo,
			version,
			fmt.Sprintf(
				"Kubernetes %s release %s", d.options.ReleaseType, version,
			),
		); err != nil {
			return errors.Wrap(err, "tag version")
		}

		// if we are working on master/main at this point, we are in
		// detached HEAD state. So we checkout the branch again.
		// The next stage (build) will checkout the branch it needs, but
		// let's not end this step with a detached HEAD
		if branch != "" && !strings.HasPrefix(branch, "release-") {
			logrus.Infof("Checking out %s to reattach HEAD", d.options.ReleaseBranch)
			if err := d.impl.Checkout(repo, d.options.ReleaseBranch); err != nil {
				return errors.Wrapf(err, "checking out branch %s", d.options.ReleaseBranch)
			}
		}
	}
	return nil
}

func (d *DefaultStage) Build() error {
	// Log in to Docker Hub to avoid getting rate limited
	if err := d.impl.DockerHubLogin(); err != nil {
		return errors.Wrap(err, "loging into Docker Hub")
	}

	// Call MakeCross for each of the versions we are building
	for _, version := range d.state.versions.Ordered() {
		if err := d.impl.MakeCross(version); err != nil {
			return errors.Wrap(err, "build artifacts")
		}
	}
	return nil
}

func (d *DefaultStage) GenerateChangelog() error {
	branch := d.options.ReleaseBranch
	if d.state.createReleaseBranch {
		branch = git.DefaultBranch
	}
	return d.impl.GenerateChangelog(&changelog.Options{
		RepoPath:     gitRoot,
		Tag:          d.state.versions.Prime(),
		Branch:       branch,
		Bucket:       d.options.Bucket(),
		HTMLFile:     releaseNotesHTMLFile,
		JSONFile:     releaseNotesJSONFile,
		Dependencies: true,
		CloneCVEMaps: true,
		Tars: filepath.Join(
			gitRoot,
			fmt.Sprintf("%s-%s", release.BuildDir, d.state.versions.Prime()),
			release.ReleaseTarsPath,
		),
	})
}

func (d *DefaultStage) StageArtifacts() error {
	for _, version := range d.state.versions.Ordered() {
		logrus.Infof("Staging artifacts for version %s", version)
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

		// Stage the local source tree
		if err := d.impl.StageLocalSourceTree(
			pushBuildOptions,
			workspaceDir,
			d.options.BuildVersion,
		); err != nil {
			return errors.Wrap(err, "staging local source tree")
		}

		// Stage local artifacts and write checksums
		if err := d.impl.StageLocalArtifacts(pushBuildOptions); err != nil {
			return errors.Wrap(err, "staging local artifacts")
		}
		gcsPath := filepath.Join(
			d.options.Bucket(), "stage", d.options.BuildVersion, version,
		)

		// Push gcs-stage to GCS
		if err := d.impl.PushReleaseArtifacts(
			pushBuildOptions,
			filepath.Join(buildDir, release.GCSStagePath, version),
			filepath.Join(gcsPath, release.GCSStagePath, version),
		); err != nil {
			return errors.Wrap(err, "pushing release artifacts")
		}

		// Push container release-images to GCS
		if err := d.impl.PushReleaseArtifacts(
			pushBuildOptions,
			filepath.Join(buildDir, release.ImagesPath),
			filepath.Join(gcsPath, release.ImagesPath),
		); err != nil {
			return errors.Wrap(err, "pushing release artifacts")
		}

		// Push container images into registry
		if err := d.impl.PushContainerImages(pushBuildOptions); err != nil {
			return errors.Wrap(err, "pushing container images")
		}
	}

	args := ""
	if d.options.NoMock {
		args += " --nomock"
	}
	if d.options.ReleaseType != DefaultOptions().ReleaseType {
		args += " --type=" + d.options.ReleaseType
	}
	if d.options.ReleaseBranch != DefaultOptions().ReleaseBranch {
		args += " --branch=" + d.options.ReleaseBranch
	}
	args += " --build-version=" + d.options.BuildVersion

	logrus.Infof(
		"To release this staged build, run:\n\n$ krel release%s", args,
	)
	return nil
}

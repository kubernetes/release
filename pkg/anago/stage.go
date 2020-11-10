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
	"k8s.io/release/pkg/changelog"
	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

// stageClient is a client for staging releases.
//counterfeiter:generate . stageClient
type stageClient interface {
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
	// state. This means that the build directory is cleaned up and the checked
	// out repository is in a clean state.
	PrepareWorkspace() error

	// TagRepository creates all necessary git objects by tagging the
	// repository for the provided `versions`.
	TagRepository(versions []string) error

	// Build runs 'make cross-in-a-container' by using the latest kubecross
	// container image. This step also build all necessary release tarballs.
	Build(versions []string) error

	// GenerateChangelog builds the CHANGELOG-x.y.md file and commits it
	// into the local repository.
	GenerateChangelog(version, parentBranch string) error

	// StageArtifacts copies the build artifacts to a Google Cloud Bucket.
	StageArtifacts(versions []string) error
}

// DefaultStage is the default staging implementation used in production.
type DefaultStage struct {
	impl    stageImpl
	options *StageOptions
}

// NewDefaultStage creates a new defaultStage instance.
func NewDefaultStage(options *StageOptions) *DefaultStage {
	return &DefaultStage{&defaultStageImpl{}, options}
}

// SetClient can be used to set the internal stage implementation.
func (d *DefaultStage) SetClient(impl stageImpl) {
	d.impl = impl
}

// defaultStageImpl is the default internal stage client implementation.
type defaultStageImpl struct{}

// stageImpl is the implementation of the stage client.
//counterfeiter:generate . stageImpl
type stageImpl interface {
	PrepareWorkspaceStage() error
	GenerateReleaseVersion(
		releaseType, version, branch string, branchFromMaster bool,
	) (*release.Versions, error)
	CheckReleaseBucket(options *build.Options) error
	Tag(releaseType, version string) error
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

func (d *defaultStageImpl) GenerateReleaseVersion(
	releaseType, version, branch string, branchFromMaster bool,
) (*release.Versions, error) {
	return release.GenerateReleaseVersion(
		releaseType, version, branch, branchFromMaster,
	)
}

func (d *defaultStageImpl) PrepareWorkspaceStage() error {
	if err := release.PrepareWorkspaceStage(gitRoot); err != nil {
		return err
	}
	return os.Chdir(gitRoot)
}

// TODO: see TagRepository()
func (d *defaultStageImpl) Tag(releaseType, version string) error {
	if err := command.New(
		"git", "config", "--global", "user.email", "nobody@k8s.io",
	).RunSuccess(); err != nil {
		return err
	}
	if err := command.New(
		"git", "config", "--global", "user.name", "Anago GCB",
	).RunSuccess(); err != nil {
		return err
	}
	return command.New(
		"git", "tag", "-a", "-m",
		fmt.Sprintf("Kubernetes %s release %s", releaseType, version),
		version,
	).RunSuccess()
}

func (d *defaultStageImpl) MakeCross(version string) error {
	return build.NewMake().MakeCross(version)
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

func (d *DefaultStage) ValidateOptions() error {
	return d.options.Validate()
}

func (d *DefaultStage) CheckPrerequisites() error { return nil }

func (d *DefaultStage) SetBuildCandidate() error { return nil }

func (d *DefaultStage) GenerateReleaseVersion(
	parentBranch string,
) (*release.Versions, error) {
	return d.impl.GenerateReleaseVersion(
		d.options.ReleaseType,
		d.options.BuildVersion,
		d.options.ReleaseBranch,
		parentBranch == git.DefaultBranch,
	)
}

func (d *DefaultStage) PrepareWorkspace() error {
	if err := d.impl.PrepareWorkspaceStage(); err != nil {
		return errors.Wrap(err, "prepare workspace")
	}
	return nil
}

func (d *DefaultStage) TagRepository(versions []string) error {
	// TODO: this is a workaround to make `Build work`, it needs to have the
	// same implementation like anago:prepare_tree.
	for _, version := range versions {
		if err := d.impl.Tag(d.options.ReleaseType, version); err != nil {
			return errors.Wrap(err, "tag version")
		}
	}
	return nil
}

func (d *DefaultStage) Build(versions []string) error {
	for _, version := range versions {
		if err := d.impl.MakeCross(version); err != nil {
			return errors.Wrap(err, "build artifacts")
		}
	}
	return nil
}

func (d *DefaultStage) GenerateChangelog(version, parentBranch string) error {
	branch := d.options.ReleaseBranch
	if parentBranch != "" {
		branch = parentBranch
	}
	return d.impl.GenerateChangelog(&changelog.Options{
		RepoPath:     gitRoot,
		Tag:          version,
		Branch:       branch,
		Bucket:       d.options.Bucket(),
		HTMLFile:     filepath.Join(workspaceDir, "src/release-notes.html"),
		Dependencies: true,
		Tars: filepath.Join(
			gitRoot,
			fmt.Sprintf("%s-%s", release.BuildDir, version),
			release.ReleaseTarsPath,
		),
	})
}

func (d *DefaultStage) StageArtifacts(versions []string) error {
	for _, version := range versions {
		logrus.Infof("Staging artifacts for version %s", version)
		buildDir := filepath.Join(
			gitRoot, fmt.Sprintf("%s-%s", release.BuildDir, version),
		)
		bucket := d.options.Bucket()
		containerRegistry := d.options.ContainerRegistry()
		pushBuildOptions := &build.Options{
			Bucket:                     bucket,
			BuildDir:                   buildDir,
			DockerRegistry:             containerRegistry,
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
		gcsPath := filepath.Join("stage", d.options.BuildVersion, version)

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

	noMockFlag := ""
	if d.options.NoMock {
		noMockFlag = "--nomock"
	}

	logrus.Infof(
		"To release this staged build, run:\n\n"+
			"$ krel gcbmgr --no-anago --release "+
			"--type %s "+
			"--branch %s "+
			"--build-version=%s %s",
		d.options.ReleaseType,
		d.options.ReleaseBranch,
		d.options.BuildVersion,
		noMockFlag,
	)
	return nil
}

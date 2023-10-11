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

package obs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/obs/specs"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-sdk/osc"
	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/util"
)

// stageClient is a client for staging releases.
//
//counterfeiter:generate . stageClient
type stageClient interface {
	// Submit can be used to submit a Google Cloud Build (GCB) job.
	Submit(stream bool) error

	// InitState initializes the default internal state.
	InitState()

	// Validate if the provided `StageOptions` are correctly set.
	ValidateOptions() error

	// InitOBSRoot initializes the OBS root directory.
	InitOBSRoot() error

	// CheckPrerequisites verifies that a valid OBS_PASSWORD environment
	// variable is set. It also checks for the existence and version of
	// required packages and if the correct Google Cloud project is set. A
	// basic hardware check will ensure that enough disk space is available,
	// too.
	CheckPrerequisites() error

	// CheckReleaseBranchState discovers if the provided release branch has to
	// be created. This is used to correctly determine release versions that
	// packages are built for.
	CheckReleaseBranchState() error

	// GenerateReleaseVersion discovers the next versions to be released.
	GenerateReleaseVersion() error

	// GeneratePackageVersion discovers the package version.
	GeneratePackageVersion()

	// GenerateOBSProject discovers the OBS project name for the release.
	GenerateOBSProject() error

	// CheckoutOBSProject checkouts the OBS project in the provided working
	// directory.
	CheckoutOBSProject() error

	// GeneratePackageArtifacts generates spec file and archive with binaries
	// for the given package.
	GeneratePackageArtifacts() error

	// Push pushes the package (spec file and archive) to OBS which triggers
	// the build.
	Push() error

	// Wait waits for the OBS build results to succeed.
	Wait() error
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
//
//counterfeiter:generate . stageImpl
type stageImpl interface {
	Submit(options *gcb.Options) error
	CheckPrerequisites(workspaceDir string) error
	MkdirAll(path string) error
	RemovePackageFiles(path string) error
	GenerateReleaseVersion(
		releaseType, version, branch string, branchFromMaster bool,
	) (*release.Versions, error)
	BranchNeedsCreation(
		branch, releaseType string, buildVersion semver.Version,
	) (bool, error)
	GenerateSpecsAndArtifacts(options *specs.Options) error
	CreateOBSConfigFile(username, password string) error
	CheckoutProject(workspaceDir, project string) error
	AddRemoveChanges(workspaceDir, project, packageName string) error
	CommitChanges(workspaceDir, project, packageName, message string) error
	Wait(project, packageName string) error
}

func (d *defaultStageImpl) Submit(options *gcb.Options) error {
	return gcb.New(options).Submit()
}

func (d *defaultStageImpl) CheckPrerequisites(workspaceDir string) error {
	return NewPrerequisitesChecker().Run(workspaceDir)
}

func (d *defaultStageImpl) MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// RemoveNonHiddenFiles removes everything in the package directory except
// `.osc` directory which contains the package metadata.
func (d *defaultStageImpl) RemovePackageFiles(path string) error {
	return filepath.Walk(path, func(fullPath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Make sure we don't delete the root directory
		if path == fullPath {
			return nil
		}

		if filepath.Base(fullPath) == ".osc" {
			return fs.SkipDir
		}

		logrus.Infof("Removing path: %s", fullPath)
		return os.RemoveAll(fullPath)
	})
}

func (d *defaultStageImpl) BranchNeedsCreation(
	branch, releaseType string, buildVersion semver.Version,
) (bool, error) {
	return release.NewBranchChecker().NeedsCreation(
		branch, releaseType, buildVersion,
	)
}

func (d *defaultStageImpl) GenerateReleaseVersion(
	releaseType, version, branch string, branchFromMaster bool,
) (*release.Versions, error) {
	return release.GenerateReleaseVersion(
		releaseType, version, branch, branchFromMaster,
	)
}

// GenerateSpecsAndArtifacts creates spec file and artifacts archive for the
// given package (`krel obs specs`).
func (d *defaultStageImpl) GenerateSpecsAndArtifacts(options *specs.Options) error {
	return specs.New(options).Run()
}

// CreateOBSConfigFile creates `~/.oscrc` file which contains the OBS API URL
// and credentials for the k8s-release-bot user.
func (d *defaultStageImpl) CreateOBSConfigFile(username, password string) error {
	return osc.CreateOSCConfigFile(obsAPIURL, username, password)
}

// CheckoutProject runs `osc checkout` in the project directory.
func (d *defaultStageImpl) CheckoutProject(workspaceDir, project string) error {
	// TODO(xmudrii) - followup: figure out how to stream output.
	return osc.OSC(filepath.Join(workspaceDir, obsRoot), "checkout", project)
}

// AddRemovePackage run `osc addremove` in the project directory.
func (d *defaultStageImpl) AddRemoveChanges(workspaceDir, project, packageName string) error {
	// TODO(xmudrii) - followup: figure out how to stream output.
	return osc.OSC(filepath.Join(workspaceDir, obsRoot, project, packageName), "addremove")
}

// CommitChanges runs `osc commit` in the package directory.
func (d *defaultStageImpl) CommitChanges(workspaceDir, project, packageName, message string) error {
	// TODO(xmudrii) - followup: figure out how to stream output.
	return osc.OSC(filepath.Join(workspaceDir, obsRoot, project, packageName), "commit", "-m", message)
}

// Wait runs `osc results -w` for the package.
func (d *defaultStageImpl) Wait(project, packageName string) error {
	return command.New(osc.OSCExecutable, "results", fmt.Sprintf("%s/%s", project, packageName), "-w").RunSuccess()
}

func (d *DefaultStage) Submit(stream bool) error {
	options := gcb.NewDefaultOptions()

	options.Stream = stream
	options.OBSStage = true
	options.NoMock = d.options.NoMock
	// Required to determine kube-cross version
	options.Branch = d.options.ReleaseBranch

	options.SpecTemplatePath = d.options.SpecTemplatePath
	options.Packages = d.options.Packages
	options.Architectures = d.options.Architectures
	options.Version = d.options.Version
	options.OBSProject = d.options.Project
	options.PackageSource = d.options.PackageSource
	options.OBSWait = d.options.Wait

	return d.impl.Submit(options)
}

func (d *DefaultStage) InitState() {
	d.state = &StageState{DefaultState()}
}

// InitOBSRoot creates the OBS root directory and the OBS config file.
func (d *DefaultStage) InitOBSRoot() error {
	password := os.Getenv(OBSPasswordKey)
	if password == "" {
		return fmt.Errorf("%s environment variable not set", OBSPasswordKey)
	}

	username := os.Getenv(OBSUsernameKey)
	if username == "" {
		username = obsK8sUsername
	}

	if err := d.impl.CreateOBSConfigFile(username, password); err != nil {
		return fmt.Errorf("creating obs config file: %w", err)
	}

	return d.impl.MkdirAll(filepath.Join(d.options.Workspace, obsRoot))
}

// ValidateOptions validates the stage options.
func (d *DefaultStage) ValidateOptions() error {
	if err := d.options.Validate(d.state.State, false); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}
	return nil
}

// CheckPrerequisites checks if all prerequisites for the stage process
// are met.
func (d *DefaultStage) CheckPrerequisites() error {
	return d.impl.CheckPrerequisites(d.options.Workspace)
}

func (d *DefaultStage) CheckReleaseBranchState() error {
	if !d.state.corePackages {
		logrus.Info("Skipping checking release branch state because non-core package is being built.")

		return nil
	}

	createReleaseBranch, err := d.impl.BranchNeedsCreation(
		d.options.ReleaseBranch,
		d.options.ReleaseType,
		d.state.semverBuildVersion,
	)
	if err != nil {
		return fmt.Errorf("check if release branch needs creation: %w", err)
	}
	d.state.createReleaseBranch = createReleaseBranch
	return nil
}

func (d *DefaultStage) GenerateReleaseVersion() error {
	if !d.state.corePackages {
		logrus.Info("Skipping generating release version because non-core package is being built.")

		return nil
	}

	versions, err := d.impl.GenerateReleaseVersion(
		d.options.ReleaseType,
		d.options.BuildVersion,
		d.options.ReleaseBranch,
		d.state.createReleaseBranch,
	)
	if err != nil {
		return fmt.Errorf("generating release versions for stage: %w", err)
	}
	// Set the versions on the state
	d.state.versions = versions
	return nil
}

// GeneratePackageVersion generates the package version for the release.
// Uses the version from the options if set, otherwise uses the prime version.
func (d *DefaultStage) GeneratePackageVersion() {
	if d.options.Version != "" {
		d.state.packageVersion = d.options.Version
	} else {
		// TODO(xmudrii): We probably want to build non prime versions as well?
		d.state.packageVersion = util.TrimTagPrefix(d.state.versions.Prime())
	}

	logrus.Infof("Using package version: %s", d.state.packageVersion)
}

// GenerateOBSProject generates the OBS project name for the release.
// Uses the project from the options if set, otherwise generates the project
// name based on the release type.
func (d *DefaultStage) GenerateOBSProject() error {
	if d.options.Project != "" {
		logrus.Infof("Using provided OBS project: %s", d.options.Project)
		d.state.obsProject = d.options.Project

		return nil
	}

	primeSemver, err := util.TagStringToSemver(d.state.versions.Prime())
	if err != nil {
		return fmt.Errorf("parsing prime version as semver: %w", err)
	}

	namespace := OBSNamespacePrerelease
	if d.options.ReleaseType == release.ReleaseTypeOfficial {
		namespace = OBSNamespaceStable
	}

	d.state.obsProject = fmt.Sprintf("%s:core:%s:v%d.%d:build", OBSKubernetesProject, namespace, primeSemver.Major, primeSemver.Minor)

	logrus.Infof("Using OBS project: %s", d.state.obsProject)

	return nil
}

// CheckoutOBSProject checks out the OBS project.
func (d *DefaultStage) CheckoutOBSProject() error {
	if err := d.impl.CheckoutProject(d.options.Workspace, d.state.obsProject); err != nil {
		return fmt.Errorf("checking out obs project: %w", err)
	}

	return nil
}

// GeneratePackageArtifacts generates the spec file and artifacts archive
// for packages that are built.
func (d *DefaultStage) GeneratePackageArtifacts() error {
	for _, pkg := range d.options.Packages {
		opts := specs.DefaultOptions()
		opts.Package = pkg
		opts.Version = d.state.packageVersion
		opts.Architectures = d.options.Architectures
		opts.PackageSourceBase = d.options.PackageSource
		if d.state.corePackages {
			opts.PackageSourceBase = fmt.Sprintf("gs://%s/stage/%s/%s/gcs-stage", d.options.Bucket(), d.options.BuildVersion, d.state.versions.Prime())
		}
		opts.SpecTemplatePath = d.options.SpecTemplatePath
		opts.SpecOutputPath = filepath.Join(d.options.Workspace, obsRoot, d.state.obsProject, pkg)

		if err := d.impl.RemovePackageFiles(opts.SpecOutputPath); err != nil {
			return fmt.Errorf("cleaning up package %s directory: %w", pkg, err)
		}

		if err := d.impl.GenerateSpecsAndArtifacts(opts); err != nil {
			return fmt.Errorf("building specs and artifacts for %s: %w", pkg, err)
		}
	}

	return nil
}

// Push pushes changes to OpenBuildService which triggers the build.
func (d *DefaultStage) Push() error {
	if !d.options.NoMock {
		logrus.Info("Running stage in mock, skipping pushing changes to OBS.")

		return nil
	}

	for _, pkg := range d.options.Packages {
		if err := d.impl.AddRemoveChanges(d.options.Workspace, d.state.obsProject, pkg); err != nil {
			return fmt.Errorf("adding/removing package files: %w", err)
		}

		if err := d.impl.CommitChanges(d.options.Workspace, d.state.obsProject, pkg, d.state.packageVersion); err != nil {
			return fmt.Errorf("committing packages: %w", err)
		}
	}

	return nil
}

// Wait waits for the OBS build results to succeed.
func (d *DefaultStage) Wait() error {
	if !d.options.Wait {
		logrus.Info("Will not wait for the OBS build results")
		return nil
	}

	if !d.options.NoMock {
		logrus.Info("Running stage in mock, skipping waiting for OBS")
		return nil
	}

	const retries = 3
	for _, pkg := range d.options.Packages {
		var tryError error

		for try := 0; try < retries; try++ {
			logrus.Infof("Waiting for package: %s (try %d)", pkg, try)

			tryError = d.impl.Wait(d.state.obsProject, pkg)
			if tryError == nil {
				break
			}

			logrus.Errorf("Unable to wait for package %s: %v", pkg, tryError)
		}

		if tryError != nil {
			return fmt.Errorf("wait for package %s: %w", pkg, tryError)
		}
	}

	return nil
}

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
	"os"
	"path/filepath"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-sdk/osc"
	"sigs.k8s.io/release-utils/helpers"

	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/release"
)

// releaseClient is a client for release a previously staged release.
//
//counterfeiter:generate . releaseClient
type releaseClient interface {
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

	// GenerateOBSProject discovers the OBS project name for the release.
	GenerateOBSProject() error

	// CheckoutOBSProject checkouts the OBS project in the provided working
	// directory.
	CheckoutOBSProject() error

	// ReleasePackages publishes successful builds into the maintenance
	// project.
	ReleasePackages() error
}

// DefaultRelease is the default release implementation used in production.
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
// arbitrary values during testing.
func (d *DefaultRelease) SetState(state *ReleaseState) {
	d.state = state
}

// defaultReleaseImpl is the default internal release client implementation.
type defaultReleaseImpl struct{}

// releaseImpl is the implementation of the release client.
//
//counterfeiter:generate . releaseImpl
type releaseImpl interface {
	Submit(options *gcb.Options) error
	CheckPrerequisites(workspaceDir string) error
	MkdirAll(path string) error
	GenerateReleaseVersion(
		releaseType, version, branch string, branchFromMaster bool,
	) (*release.Versions, error)
	BranchNeedsCreation(
		branch, releaseType string, buildVersion semver.Version,
	) (bool, error)
	CreateOBSConfigFile(username, password string) error
	CheckoutProject(workspaceDir, project string) error
	ReleasePackage(workspaceDir, project, packageName string) error
}

func (d *defaultReleaseImpl) Submit(options *gcb.Options) error {
	return gcb.New(options).Submit()
}

func (d *defaultReleaseImpl) CheckPrerequisites(workspaceDir string) error {
	return NewPrerequisitesChecker().Run(workspaceDir)
}

func (d *defaultReleaseImpl) MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func (d *defaultReleaseImpl) BranchNeedsCreation(
	branch, releaseType string, buildVersion semver.Version,
) (bool, error) {
	return release.NewBranchChecker().NeedsCreation(
		branch, releaseType, buildVersion,
	)
}

func (d *defaultReleaseImpl) GenerateReleaseVersion(
	releaseType, version, branch string, branchFromMaster bool,
) (*release.Versions, error) {
	return release.GenerateReleaseVersion(
		releaseType, version, branch, branchFromMaster,
	)
}

// CreateOBSConfigFile creates `~/.oscrc` file which contains the OBS API URL
// and credentials for the k8s-release-bot user.
func (d *defaultReleaseImpl) CreateOBSConfigFile(username, password string) error {
	return osc.CreateOSCConfigFile(obsAPIURL, username, password)
}

// CheckoutProject runs `osc checkout` in the project directory.
func (d *defaultReleaseImpl) CheckoutProject(workspaceDir, project string) error {
	// TODO(xmudrii) - followup: figure out how to stream output.
	return osc.OSC(filepath.Join(workspaceDir, obsRoot), "checkout", project)
}

// ReleasePackage runs `osc release` in the package directory.
func (d *defaultReleaseImpl) ReleasePackage(workspaceDir, project, packageName string) error {
	return osc.OSC(filepath.Join(workspaceDir, obsRoot, project, packageName), "release")
}

func (d *DefaultRelease) Submit(stream bool) error {
	options := gcb.NewDefaultOptions()

	options.Stream = stream
	options.OBSRelease = true
	options.NoMock = d.options.NoMock
	// Required to determine kube-cross version
	options.Branch = d.options.ReleaseBranch
	options.Packages = d.options.Packages
	options.OBSProject = d.options.Project

	return d.impl.Submit(options)
}

func (d *DefaultRelease) InitState() {
	d.state = &ReleaseState{
		State: DefaultState(),
	}
}

// InitOBSRoot creates the OBS root directory and the OBS config file.
func (d *DefaultRelease) InitOBSRoot() error {
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

// ValidateOptions validates the release options.
func (d *DefaultRelease) ValidateOptions() error {
	if err := d.options.Validate(d.state.State, false); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}

	return nil
}

// CheckPrerequisites checks if all prerequisites for the release process
// are met.
func (d *DefaultRelease) CheckPrerequisites() error {
	return d.impl.CheckPrerequisites(d.options.Workspace)
}

func (d *DefaultRelease) CheckReleaseBranchState() error {
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

func (d *DefaultRelease) GenerateReleaseVersion() error {
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
		return fmt.Errorf("generating release versions for release: %w", err)
	}
	// Set the versions on the state
	d.state.versions = versions

	return nil
}

// GenerateOBSProject generates the OBS project name for the release.
// Uses the project from the options if set, otherwise generates the project
// name based on the release type.
func (d *DefaultRelease) GenerateOBSProject() error {
	if d.options.Project != "" {
		logrus.Infof("Using provided OBS project: %s", d.options.Project)
		d.state.obsProject = d.options.Project

		return nil
	}

	primeSemver, err := helpers.TagStringToSemver(d.state.versions.Prime())
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
func (d *DefaultRelease) CheckoutOBSProject() error {
	if err := d.impl.CheckoutProject(d.options.Workspace, d.state.obsProject); err != nil {
		return fmt.Errorf("checking out obs project: %w", err)
	}

	return nil
}

// ReleasePackage publishes successful builds into the maintenance
// project.
func (d *DefaultRelease) ReleasePackages() error {
	if !d.options.NoMock {
		logrus.Info("Running release in mock mode, skipping releasing packages.")

		return nil
	}

	for _, pkg := range d.options.Packages {
		if err := d.impl.ReleasePackage(d.options.Workspace, d.state.obsProject, pkg); err != nil {
			return fmt.Errorf("releasing package %s from project %s: %w", pkg, d.state.obsProject, err)
		}
	}

	return nil
}

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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/consts"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/log"
	"sigs.k8s.io/release-utils/util"
	"sigs.k8s.io/release-utils/version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt obsfakes/fake_prerequisites_checker_impl.go > obsfakes/_fake_prerequisites_checker_impl.go && mv obsfakes/_fake_prerequisites_checker_impl.go obsfakes/fake_prerequisites_checker_impl.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt obsfakes/fake_stage_client.go > obsfakes/_fake_stage_client.go && mv obsfakes/_fake_stage_client.go obsfakes/fake_stage_client.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt obsfakes/fake_stage_impl.go > obsfakes/_fake_stage_impl.go && mv obsfakes/_fake_stage_impl.go obsfakes/fake_stage_impl.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt obsfakes/fake_release_client.go > obsfakes/_fake_release_client.go && mv obsfakes/_fake_release_client.go obsfakes/fake_release_client.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt obsfakes/fake_release_impl.go > obsfakes/_fake_release_impl.go && mv obsfakes/_fake_release_impl.go obsfakes/fake_release_impl.go"
const (
	// OBSKubernetesProject is name of the organization/project on openSUSE's
	// OBS instance where packages are built and published.
	OBSKubernetesProject = "isv:kubernetes"

	// OBSNamespaceStable is part of the subproject name that's used for stable
	// packages.
	OBSNamespaceStable = "stable"
	// OBSNamespaceStable is part of the subproject name that's used for
	// prerelease (alpha, beta, rc) packages.
	OBSNamespacePrerelease = "prerelease"

	// workspaceDir is the global directory where the stage and release process
	// happens.
	defaultWorkspaceDir = "/workspace"

	// defaultSpecTemplatePath is path inside Google Cloud Build where package
	// specs for kubeadm, kubectl, and kubelet are located.
	defaultSpecTemplatePath = defaultWorkspaceDir + "/go/src/k8s.io/release/" + consts.DefaultSpecTemplatePath

	// obsRoot is path inside Google Cloud Build where OBS project and packages
	// are checked out.
	obsRoot = "/src/obs"

	// obsAPIURL is the URL of openSUSE's OpenBuildService instance.
	obsAPIURL = "https://api.opensuse.org"

	// obsK8sUsername is username for Kubernetes Release Bot account.
	obsK8sUsername = "k8s-release-bot"

	// OBSPasswordKey is name of the environment variable with password for
	// Kubernetes Release Bot account.
	OBSPasswordKey = "OBS_PASSWORD"

	// OBSUsernameKey is name of the environment variable containing the
	// username for the OBS account. If empty, obsK8sUsername will be used.
	OBSUsernameKey = "OBS_USERNAME"
)

// Options are settings which will be used by `StageOptions` as well as
// `ReleaseOptions`.
type Options struct {
	// Workspace is the root workspace.
	Workspace string

	// Run the whole process in non-mocked mode. Which means that it doesn't
	// push specs and artifacts to OpenBuildService.
	NoMock bool

	// SpecTemplatePath is path to a directory with spec template files.
	SpecTemplatePath string

	// Packages that should be built and published to OpenBuildService.
	Packages []string

	// Architectures for which packages should be built.
	Architectures []string

	/**
	* Parameters used for core packages.
	* Core packages are: kubeadm, kubectl, kubelet, cri-tools, kubernetes-cni.
	* The API is same as for "krel stage" for consistency reasons.
	**/

	// The release type for which packages are built for.
	// Can be either `alpha`, `beta`, `rc` or `official`.
	// Mutually exclusive with `Version`, `Project` and `Source`.
	ReleaseType string

	// The release branch for which the release should be built.
	// Can be `master`/`main` or any `release-x.y` branch.
	// Mutually exclusive with `Version`, `Project` and `Source`.
	ReleaseBranch string

	// The build version to be released. Has to be specified in the format:
	// `vX.Y.Z-[alpha|beta|rc].N.C+SHA`
	// Mutually exclusive with `Version`, `Project` and `Source`.
	BuildVersion string

	/**
	* Parameters used for non-core packages.
	* Non-core packages are: cri-o
	**/

	// Version of packages to build. Same version is used for all provided
	// packages.
	// Mutually exclusive with `ReleaseType`, `ReleaseBranch`, and
	// `BuildVersion`.
	Version string

	// Project is name of the OBS project where packages are built.
	// Mutually exclusive with `ReleaseType`, `ReleaseBranch`, and
	// `BuildVersion`.
	Project string

	// PackageSource is https:// or gs:// URL where to download binaries for
	// packages from.
	// Mutually exclusive with `ReleaseType`, `ReleaseBranch`, and
	// `BuildVersion`.
	PackageSource string

	// Wait can be used to wait for the OBS build results.
	Wait bool
}

// DefaultOptions returns a new `Options` instance.
func DefaultOptions() *Options {
	return &Options{
		Packages: []string{
			consts.PackageKubeadm,
			consts.PackageKubectl,
			consts.PackageKubelet,
		},
		Architectures: []string{
			consts.ArchitectureAMD64,
			consts.ArchitectureARM64,
			consts.ArchitecturePPC64,
			consts.ArchitectureS390X,
		},
		SpecTemplatePath: defaultSpecTemplatePath,
		Workspace:        defaultWorkspaceDir,
	}
}

// String returns a string representation for the `Options` type.
func (o *Options) String() string {
	return fmt.Sprintf(
		"NoMock: %v, Packages: %s, BuildVersion: %q, Version: %s, Project: %q, ",
		o.NoMock, o.Packages, o.BuildVersion, o.Version, o.Project,
	)
}

// Validate if the options are correctly set.
func (o *Options) Validate(submit bool) error {
	logrus.Infof("Validating generic options: %s", o.String())

	// TODO(xmudrii) - followup: consider a better way to handle this.
	if !submit {
		// Ensure provided SpecTemplatePath exists.
		if _, err := os.Stat(o.SpecTemplatePath); err != nil {
			return fmt.Errorf("invalid spec template path: %w", err)
		}

		for _, pkg := range o.Packages {
			if _, err := os.Stat(filepath.Join(o.SpecTemplatePath, pkg)); err != nil {
				return fmt.Errorf("specs for package %s doesn't exist", pkg)
			}
		}
	}

	// Ensure provided architectures are supported.
	if !consts.IsSupported("architectures", o.Architectures, consts.SupportedArchitectures) {
		return errors.New("provided architectures are not supported")
	}

	var (
		foundK8sOption    bool
		foundManualOption bool
	)

	// Ensure exclusive mutual exclusivity of options is respected.
	if o.ReleaseType != "" || o.ReleaseBranch != "" || o.BuildVersion != "" {
		foundK8sOption = true
	}
	if o.Project != "" || o.PackageSource != "" || o.Version != "" {
		foundManualOption = true
	}

	if foundK8sOption && foundManualOption {
		return errors.New("kubernetes and manual options are mutually exclusive")
	}
	if !foundK8sOption && !foundManualOption {
		return errors.New("one of kubernetes or manual options are required")
	}

	// Validate other options depending on release type.
	if foundK8sOption {
		if o.ReleaseType != release.ReleaseTypeAlpha &&
			o.ReleaseType != release.ReleaseTypeBeta &&
			o.ReleaseType != release.ReleaseTypeRC &&
			o.ReleaseType != release.ReleaseTypeOfficial {
			return fmt.Errorf("invalid release type: %s", o.ReleaseType)
		}

		if !git.IsReleaseBranch(o.ReleaseBranch) {
			return fmt.Errorf("invalid release branch: %s", o.ReleaseBranch)
		}

		if o.BuildVersion == "" {
			return errors.New("build version is required")
		}
	} else if foundManualOption {
		if o.Project == "" {
			return errors.New("project is required")
		}
	}

	return nil
}

// ValidateBuildVersion validates the provided build version.
func (o *Options) ValidateBuildVersion(state *State) error {
	// Verify the build version is correct:
	correct, err := release.IsValidReleaseBuild(o.BuildVersion)
	if err != nil {
		return fmt.Errorf("checking for a valid build version: %w", err)
	}
	if !correct {
		return errors.New("invalid BuildVersion specified")
	}

	semverBuildVersion, err := util.TagStringToSemver(o.BuildVersion)
	if err != nil {
		return fmt.Errorf("invalid build version: %s: %w", o.BuildVersion, err)
	}
	state.semverBuildVersion = semverBuildVersion
	return nil
}

// Bucket returns the Google Cloud Bucket for these `Options`.
func (o *Options) Bucket() string {
	if o.NoMock {
		return release.ProductionBucket
	}
	return release.TestBucket
}

// State holds all inferred and calculated values from the stage/release
// process, it's state mutates as each step es executed
type State struct {
	// corePackages is an indicator if we're releasing core/Kubernetes
	// packages. This is set upon validation.
	corePackages bool

	// obsProject is parametrized OBS project name.
	// For Kubernetes packages, this is autogenerated based on given release
	// type, branch, and build version.
	// For non-Kubernetes packages, this is the same as `Project`.
	// This is set after GenerateOBSProject()
	obsProject string

	// packageVersion is parametrized package version.
	// For Kubernetes packages, this is autogenerated based on given release
	// type, branch, and build version.
	// For non-Kubernetes packages, this is the same as `Project`.
	// This is set after GeneratePackageVersion()
	packageVersion string

	// semverBuildVersion is the parsed build version which is set after the
	// validation.
	semverBuildVersion semver.Version

	// The release versions generated after GenerateReleaseVersion()
	versions *release.Versions

	// Indicates if creating a release branch is needed. This parameter is
	// used when determining release versions
	createReleaseBranch bool

	// startTime is the time when stage/release starts
	startTime time.Time
}

// DefaultState returns a new empty State
func DefaultState() *State {
	// The default state is empty, it will be initialized after ValidateOptions()
	// runs in Stage/Release. It will change as the stage/release processes move forward
	return &State{
		startTime: time.Now(),
	}
}

// StageState holds the stage process state
type StageState struct {
	*State
}

// DefaultStageState create a new default `StageState`.
func DefaultStageState() *StageState {
	return &StageState{
		State: DefaultState(),
	}
}

// StageOptions contains the options for running `Stage`.
type StageOptions struct {
	*Options
}

// DefaultStageOptions create a new default `StageOptions`.
func DefaultStageOptions() *StageOptions {
	return &StageOptions{
		Options: DefaultOptions(),
	}
}

// String returns a string representation for the `StageOptions` type.
func (s *StageOptions) String() string {
	return s.Options.String()
}

// Validate validates the stage options.
func (s *StageOptions) Validate(state *State, submit bool) error {
	if err := s.Options.Validate(submit); err != nil {
		return fmt.Errorf("validating generic options: %w", err)
	}

	if s.Options.BuildVersion != "" {
		if err := s.Options.ValidateBuildVersion(state); err != nil {
			return errors.New("validating build version")
		}
	} else if s.Options.Version == "" {
		// Version is required only for stage,
		// release step only publishes what has been already built.
		return errors.New("version is required")
	}

	if s.Options.ReleaseType != "" || s.Options.ReleaseBranch != "" || s.Options.BuildVersion != "" {
		state.corePackages = true
	}

	return nil
}

// Stage is the structure to be used for staging packages.
type Stage struct {
	client stageClient
}

// NewStage creates a new `Stage` instance.
func NewStage(options *StageOptions) *Stage {
	return &Stage{NewDefaultStage(options)}
}

// SetClient can be used to set the internal stage client.
func (s *Stage) SetClient(client stageClient) {
	s.client = client
}

// Submit can be used to submit a staging Google Cloud Build (GCB) job.
func (s *Stage) Submit(stream bool) error {
	logrus.Info("Submitting OBS stage GCB job")
	if err := s.client.Submit(stream); err != nil {
		return fmt.Errorf("submit obs stage job: %w", err)
	}
	return nil
}

// Run for the `Stage` struct prepares a release and pushes specs and archives
// to OpenBuildService.
func (s *Stage) Run() error {
	s.client.InitState()

	logger := log.NewStepLogger(11)
	v := version.GetVersionInfo()
	logger.Infof("Using krel version: %s", v.GitVersion)

	logger.WithStep().Info("Validating options")
	if err := s.client.ValidateOptions(); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}

	logger.WithStep().Info("Initializing OBS root and config")
	if err := s.client.InitOBSRoot(); err != nil {
		return fmt.Errorf("initializing obs root: %w", err)
	}

	logger.WithStep().Info("Checking prerequisites")
	if err := s.client.CheckPrerequisites(); err != nil {
		return fmt.Errorf("check prerequisites: %w", err)
	}

	logger.WithStep().Info("Checking release branch state")
	if err := s.client.CheckReleaseBranchState(); err != nil {
		return fmt.Errorf("checking release branch state: %w", err)
	}

	logger.WithStep().Info("Generating release version")
	if err := s.client.GenerateReleaseVersion(); err != nil {
		return fmt.Errorf("generating release version: %w", err)
	}

	logger.WithStep().Info("Generating package version")
	s.client.GeneratePackageVersion()

	logger.WithStep().Info("Generating OBS project name")
	if err := s.client.GenerateOBSProject(); err != nil {
		return fmt.Errorf("generating obs project name: %w", err)
	}

	logger.WithStep().Info("Checking out OBS project")
	if err := s.client.CheckoutOBSProject(); err != nil {
		return fmt.Errorf("checking out obs project: %w", err)
	}

	logger.WithStep().Info("Generating spec files and artifact archives")
	if err := s.client.GeneratePackageArtifacts(); err != nil {
		return fmt.Errorf("generating package artifacts: %w", err)
	}

	logger.WithStep().Info("Pushing packages to OBS")
	if err := s.client.Push(); err != nil {
		return fmt.Errorf("pushing packages to obs: %w", err)
	}

	logger.WithStep().Info("Waiting for OBS build results if required")
	if err := s.client.Wait(); err != nil {
		return fmt.Errorf("wait for OBS build results: %w", err)
	}

	return nil
}

// ReleaseState holds the release process state
type ReleaseState struct {
	*State
}

// DefaultReleaseState create a new default `ReleaseOptions`.
func DefaultReleaseState() *ReleaseState {
	return &ReleaseState{
		State: DefaultState(),
	}
}

// ReleaseOptions contains the options for running `Release`.
type ReleaseOptions struct {
	*Options
}

// DefaultReleaseOptions create a new default `ReleaseOptions`.
func DefaultReleaseOptions() *ReleaseOptions {
	return &ReleaseOptions{
		Options: DefaultOptions(),
	}
}

// String returns a string representation for the `ReleaseOptions` type.
func (r *ReleaseOptions) String() string {
	return r.Options.String()
}

// Validate if the options are correctly set.
func (r *ReleaseOptions) Validate(state *State, submit bool) error {
	if err := r.Options.Validate(submit); err != nil {
		return fmt.Errorf("validating generic options: %w", err)
	}

	if r.Options.BuildVersion != "" {
		if err := r.Options.ValidateBuildVersion(state); err != nil {
			return fmt.Errorf("validating build version: %w", err)
		}
	}

	if r.Options.Version != "" {
		return errors.New("specifying version is not supported for release")
	}

	if r.Options.ReleaseType != "" || r.Options.ReleaseBranch != "" || r.Options.BuildVersion != "" {
		state.corePackages = true
	}

	return nil
}

// Release is the structure to be used for releasing staged OBS builds.
type Release struct {
	client releaseClient
}

// NewRelease creates a new `Release` instance.
func NewRelease(options *ReleaseOptions) *Release {
	return &Release{NewDefaultRelease(options)}
}

// SetClient can be used to set the internal stage client.
func (r *Release) SetClient(client releaseClient) {
	r.client = client
}

// Submit can be used to submit a releasing Google Cloud Build (GCB) job.
func (r *Release) Submit(stream bool) error {
	logrus.Info("Submitting release GCB job")
	if err := r.client.Submit(stream); err != nil {
		return fmt.Errorf("submit release job: %w", err)
	}
	return nil
}

// Run for `Release` struct finishes a previously staged release.
func (r *Release) Run() error {
	r.client.InitState()

	logger := log.NewStepLogger(8)
	v := version.GetVersionInfo()
	logger.Infof("Using krel version: %s", v.GitVersion)

	logger.WithStep().Info("Validating options")
	if err := r.client.ValidateOptions(); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}

	logger.WithStep().Info("Initializing OBS root and config")
	if err := r.client.InitOBSRoot(); err != nil {
		return fmt.Errorf("initializing obs root: %w", err)
	}

	logger.WithStep().Info("Checking prerequisites")
	if err := r.client.CheckPrerequisites(); err != nil {
		return fmt.Errorf("check prerequisites: %w", err)
	}

	logger.WithStep().Info("Checking release branch state")
	if err := r.client.CheckReleaseBranchState(); err != nil {
		return fmt.Errorf("checking release branch state: %w", err)
	}

	logger.WithStep().Info("Generating release version")
	if err := r.client.GenerateReleaseVersion(); err != nil {
		return fmt.Errorf("generating release version: %w", err)
	}

	logger.WithStep().Info("Generating OBS project name")
	if err := r.client.GenerateOBSProject(); err != nil {
		return fmt.Errorf("generating obs project name: %w", err)
	}

	logger.WithStep().Info("Checking out OBS project")
	if err := r.client.CheckoutOBSProject(); err != nil {
		return fmt.Errorf("checking out obs project: %w", err)
	}

	logger.WithStep().Info("Releasing packages to OBS")
	if err := r.client.ReleasePackages(); err != nil {
		return fmt.Errorf("releasing packages: %w", err)
	}

	return nil
}

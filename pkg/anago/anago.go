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
	"errors"
	"fmt"
	"time"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/log"
	"sigs.k8s.io/release-utils/util"
	"sigs.k8s.io/release-utils/version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt anagofakes/fake_release_client.go > anagofakes/_fake_release_client.go && mv anagofakes/_fake_release_client.go anagofakes/fake_release_client.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt anagofakes/fake_release_impl.go > anagofakes/_fake_release_impl.go && mv anagofakes/_fake_release_impl.go anagofakes/fake_release_impl.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt anagofakes/fake_stage_client.go > anagofakes/_fake_stage_client.go && mv anagofakes/_fake_stage_client.go anagofakes/fake_stage_client.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt anagofakes/fake_stage_impl.go > anagofakes/_fake_stage_impl.go && mv anagofakes/_fake_stage_impl.go anagofakes/fake_stage_impl.go"
const (
	// workspaceDir is the global directory where the stage and release process
	// happens.
	workspaceDir = "/workspace"

	// gitRoot is the local repository root of k/k.
	gitRoot = workspaceDir + "/src/k8s.io/kubernetes"

	// releaseNotesHTMLFile is the name of the release notes in HTML
	releaseNotesHTMLFile = workspaceDir + "/src/release-notes.html"

	// releaseNotesJSONFile is the file containing the release notes in json format
	releaseNotesJSONFile = workspaceDir + "/src/release-notes.json"

	// The default license for all artifacts
	LicenseIdentifier = "Apache-2.0"
)

// Options are settings which will be used by `StageOptions` as well as
// `ReleaseOptions`.
type Options struct {
	// Run the whole process in non-mocked mode. Which means that it uses
	// production remote locations for storing artifacts and modifying git
	// repositories.
	NoMock bool

	// The release type which should be produced. Can be either `alpha`,
	// `beta`, `rc` or `official`.
	ReleaseType string

	// The release branch for which the release should be build. Can be
	// `master` or any `release-x.y` branch.
	ReleaseBranch string

	// The build version to be released. Has to be specified in the format:
	// `vX.Y.Z-[alpha|beta|rc].N.C+SHA`
	BuildVersion string
}

// DefaultOptions returns a new Options instance.
func DefaultOptions() *Options {
	return &Options{
		ReleaseType:   release.ReleaseTypeAlpha,
		ReleaseBranch: git.DefaultBranch,
	}
}

// String returns a string representation for the `ReleaseOptions` type.
func (o *Options) String() string {
	return fmt.Sprintf(
		"NoMock: %v, ReleaseType: %q, BuildVersion: %q, ReleaseBranch: %q",
		o.NoMock, o.ReleaseType, o.BuildVersion, o.ReleaseBranch,
	)
}

// Validate if the options are correctly set.
func (o *Options) Validate() error {
	logrus.Infof("Validating generic options: %s", o.String())

	if o.ReleaseType != release.ReleaseTypeAlpha &&
		o.ReleaseType != release.ReleaseTypeBeta &&
		o.ReleaseType != release.ReleaseTypeRC &&
		o.ReleaseType != release.ReleaseTypeOfficial {
		return fmt.Errorf("invalid release type: %s", o.ReleaseType)
	}

	if !git.IsReleaseBranch(o.ReleaseBranch) {
		return fmt.Errorf("invalid release branch: %s", o.ReleaseBranch)
	}

	return nil
}

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

// ContainerRegistry returns the container registry for these `Options`.
func (o *Options) ContainerRegistry() string {
	if o.NoMock {
		return release.GCRIOPathStaging
	}
	return release.GCRIOPathMock
}

// State holds all inferred and calculated values from the release process
// it's state mutates as each step es executed
type State struct {
	// logFile is the internal logging file target.
	logFile string

	// semverBuildVersion is the parsed build version which is set after the
	// validation.
	semverBuildVersion semver.Version

	// The release versions generated after GenerateReleaseVersion()
	versions *release.Versions

	// Indicates if we're going to create a new release branch.
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

func (s *State) SetCreateReleaseBranch(createReleaseBranch bool) {
	s.createReleaseBranch = createReleaseBranch
}

func (s *State) SetVersions(versions *release.Versions) {
	s.versions = versions
}

// StageState holds the release process state
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

// Validate if the options are correctly set.
func (s *StageOptions) Validate(state *State) error {
	if err := s.Options.Validate(); err != nil {
		return fmt.Errorf("validating generic options: %w", err)
	}

	// build version is optional for staging, but if provided we should
	// validate it.
	if s.Options.BuildVersion != "" {
		if err := s.Options.ValidateBuildVersion(state); err != nil {
			return errors.New("validating build version")
		}
	}

	return nil
}

// Stage is the structure to be used for staging releases.
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
	logrus.Info("Submitting stage GCB job")
	if err := s.client.Submit(stream); err != nil {
		return fmt.Errorf("submit stage job: %w", err)
	}
	return nil
}

// Run for the `Stage` struct prepares a release and puts the results on a
// staging bucket.
func (s *Stage) Run() error {
	s.client.InitState()

	if err := s.client.InitLogFile(); err != nil {
		return fmt.Errorf("init log file: %w", err)
	}

	logger := log.NewStepLogger(12)
	v := version.GetVersionInfo()
	logger.Infof("Using krel version: %s", v.GitVersion)

	logger.WithStep().Info("Validating options")
	if err := s.client.ValidateOptions(); err != nil {
		return fmt.Errorf("validate options: %w", err)
	}

	logger.WithStep().Info("Checking prerequisites")
	if err := s.client.CheckPrerequisites(); err != nil {
		return fmt.Errorf("check prerequisites: %w", err)
	}

	logger.WithStep().Info("Checking release branch state")
	if err := s.client.CheckReleaseBranchState(); err != nil {
		return fmt.Errorf("check release branch state: %w", err)
	}

	logger.WithStep().Info("Generating release version")
	if err := s.client.GenerateReleaseVersion(); err != nil {
		return fmt.Errorf("generate release version: %w", err)
	}

	logger.WithStep().Info("Preparing workspace")
	if err := s.client.PrepareWorkspace(); err != nil {
		return fmt.Errorf("prepare workspace: %w", err)
	}

	logger.WithStep().Info("Tagging repository")
	if err := s.client.TagRepository(); err != nil {
		return fmt.Errorf("tag repository: %w", err)
	}

	logger.WithStep().Info("Building release")
	if err := s.client.Build(); err != nil {
		return fmt.Errorf("build release: %w", err)
	}

	logger.WithStep().Info("Generating changelog")
	if err := s.client.GenerateChangelog(); err != nil {
		return fmt.Errorf("generate changelog: %w", err)
	}

	logger.WithStep().Info("Verifying artifacts")
	if err := s.client.VerifyArtifacts(); err != nil {
		return fmt.Errorf("verifying artifacts: %w", err)
	}

	logger.WithStep().Info("Generating bill of materials")
	if err := s.client.GenerateBillOfMaterials(); err != nil {
		return fmt.Errorf("generating sbom: %w", err)
	}

	logger.WithStep().Info("Staging artifacts")
	if err := s.client.StageArtifacts(); err != nil {
		return fmt.Errorf("stage release artifacts: %w", err)
	}

	logger.Info("Stage done")
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
func (r *ReleaseOptions) Validate(state *State) error {
	if err := r.Options.Validate(); err != nil {
		return fmt.Errorf("validating generic options: %w", err)
	}
	if err := r.Options.ValidateBuildVersion(state); err != nil {
		return fmt.Errorf("validating build version: %w", err)
	}
	return nil
}

// Release is the structure to be used for releasing staged releases.
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

	if err := r.client.InitLogFile(); err != nil {
		return fmt.Errorf("init log file: %w", err)
	}

	logger := log.NewStepLogger(12)
	v := version.GetVersionInfo()
	logger.Infof("Using krel version: %s", v.GitVersion)

	logger.WithStep().Info("Validating options")
	if err := r.client.ValidateOptions(); err != nil {
		return fmt.Errorf("validate options: %w", err)
	}

	logger.WithStep().Info("Checking prerequisites")
	if err := r.client.CheckPrerequisites(); err != nil {
		return fmt.Errorf("check prerequisites: %w", err)
	}

	logger.WithStep().Info("Checking release branch state")
	if err := r.client.CheckReleaseBranchState(); err != nil {
		return fmt.Errorf("check release branch state: %w", err)
	}

	logger.WithStep().Info("Generating release version")
	if err := r.client.GenerateReleaseVersion(); err != nil {
		return fmt.Errorf("generate release version: %w", err)
	}

	logger.WithStep().Info("Preparing workspace")
	if err := r.client.PrepareWorkspace(); err != nil {
		return fmt.Errorf("prepare workspace: %w", err)
	}

	logger.WithStep().Info("Checking artifacts provenance")
	if err := r.client.CheckProvenance(); err != nil {
		// For now, we only notify provenance errors as not to treat
		// them as fatal while we finish testing SLSA compliance.
		logrus.Warnf("Unable to check provenance attestation: %v", err)
	}

	logger.WithStep().Info("Pushing artifacts")
	if err := r.client.PushArtifacts(); err != nil {
		return fmt.Errorf("push artifacts: %w", err)
	}

	logger.WithStep().Info("Pushing git objects")
	if err := r.client.PushGitObjects(); err != nil {
		return fmt.Errorf("push git objects: %w", err)
	}

	logger.WithStep().Info("Creating announcement")
	if err := r.client.CreateAnnouncement(); err != nil {
		return fmt.Errorf("create announcement: %w", err)
	}

	logger.WithStep().Info("Updating GitHub release page")
	if err := r.client.UpdateGitHubPage(); err != nil {
		return fmt.Errorf("updating github page: %w", err)
	}

	logger.WithStep().Info("Archiving release")
	if err := r.client.Archive(); err != nil {
		return fmt.Errorf("archive release: %w", err)
	}

	logger.Info("Release done")
	return nil
}

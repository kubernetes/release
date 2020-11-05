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

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

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
	if o.ReleaseType != release.ReleaseTypeAlpha &&
		o.ReleaseType != release.ReleaseTypeBeta &&
		o.ReleaseType != release.ReleaseTypeRC &&
		o.ReleaseType != release.ReleaseTypeOfficial {
		return errors.Errorf("invalid release type: %s", o.ReleaseType)
	}

	if !git.IsReleaseBranch(o.ReleaseBranch) {
		return errors.Errorf("invalid release branch: %s", o.ReleaseBranch)
	}

	if _, err := semver.Parse(o.BuildVersion); err != nil {
		return errors.Wrapf(err, "invalid build version: %s", o.BuildVersion)
	}
	return nil
}

// StageOptions contains the options for running `Stage`.
type StageOptions struct {
	*Options
}

// DefaultStageOptions createa a new default `StageOptions`.
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
func (s *StageOptions) Validate() error {
	return s.Options.Validate()
}

// Stage is the structure to be used for staging releases.
type Stage struct {
	client  stageClient
	options *StageOptions
}

// NewStage creates a new `Stage` instance.
func NewStage(options *StageOptions) *Stage {
	return &Stage{
		client:  NewDefaultStage(),
		options: options,
	}
}

// SetClient can be used to set the internal stage client.
func (s *Stage) SetClient(client stageClient) {
	s.client = client
}

// Run for the `Stage` struct prepares a release and puts the results on a
// staging bucket.
func (s *Stage) Run() error {
	logrus.Infof("Running stage with options: %s", s.options.String())

	logrus.Info("Validating provided options")
	if err := s.client.ValidateOptions(s.options); err != nil {
		return errors.Wrap(err, "validate options")
	}

	logrus.Info("Checking prerequisites")
	if err := s.client.CheckPrerequisites(); err != nil {
		return errors.Wrap(err, "check prerequisites")
	}

	logrus.Info("Setting build candidate")
	if err := s.client.SetBuildCandidate(); err != nil {
		return errors.Wrap(err, "set build candidate")
	}

	logrus.Info("Preparing workspace")
	if err := s.client.PrepareWorkspace(); err != nil {
		return errors.Wrap(err, "prepare workspace")
	}

	logrus.Info("Building release")
	if err := s.client.Build(); err != nil {
		return errors.Wrap(err, "build release")
	}

	logrus.Info("Generating release notes")
	if err := s.client.GenerateReleaseNotes(); err != nil {
		return errors.Wrap(err, "generate release notes")
	}

	logrus.Info("Staging artifacts")
	if err := s.client.StageArtifacts(); err != nil {
		return errors.Wrap(err, "stage release artifacts")
	}

	logrus.Info("Stage done")
	return nil
}

// ReleaseOptions contains the options for running `Release`.
type ReleaseOptions struct {
	*Options
}

// DefaultReleaseOptions createa a new default `ReleaseOptions`.
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
func (r *ReleaseOptions) Validate() error {
	return r.Options.Validate()
}

// Release is the structure to be used for releasing staged releases.
type Release struct {
	client  releaseClient
	options *ReleaseOptions
}

// NewRelease creates a new `Release` instance.
func NewRelease(options *ReleaseOptions) *Release {
	return &Release{
		client:  NewDefaultRelease(),
		options: options,
	}
}

// SetClient can be used to set the internal stage client.
func (r *Release) SetClient(client releaseClient) {
	r.client = client
}

// Run for for `Release` struct finishes a previously staged release.
func (r *Release) Run() error {
	logrus.Infof("Running release with options: %s", r.options.String())

	logrus.Info("Validating provided options")
	if err := r.client.ValidateOptions(r.options); err != nil {
		return errors.Wrap(err, "validate options")
	}

	logrus.Info("Checking prerequisites")
	if err := r.client.CheckPrerequisites(); err != nil {
		return errors.Wrap(err, "check prerequisites")
	}

	logrus.Info("Setting build candidate")
	if err := r.client.SetBuildCandidate(); err != nil {
		return errors.Wrap(err, "set build candidate")
	}

	logrus.Info("Preparing workspace")
	if err := r.client.PrepareWorkspace(); err != nil {
		return errors.Wrap(err, "prepare workspace")
	}

	logrus.Info("Pushing artifacts")
	if err := r.client.PushArtifacts(); err != nil {
		return errors.Wrap(err, "push artifacts")
	}

	logrus.Info("Pushing git Objects")
	if err := r.client.PushGitObjects(); err != nil {
		return errors.Wrap(err, "push git objects")
	}

	logrus.Info("Creating announcement")
	if err := r.client.CreateAnnouncement(); err != nil {
		return errors.Wrap(err, "create announcement")
	}

	logrus.Info("Archiving release")
	if err := r.client.Archive(); err != nil {
		return errors.Wrap(err, "archive release")
	}

	logrus.Info("Release done")
	return nil
}

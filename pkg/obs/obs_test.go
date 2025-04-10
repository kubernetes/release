/*
Copyright 2025 The Kubernetes Authors.

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

package obs_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"sigs.k8s.io/release-sdk/git"

	"k8s.io/release/pkg/consts"
	"k8s.io/release/pkg/obs"
	"k8s.io/release/pkg/obs/obsfakes"
	"k8s.io/release/pkg/release"
)

func TestValidateOptions(t *testing.T) {
	for _, tc := range []struct {
		provided    *obs.Options
		shouldError bool
		submit      bool
	}{
		{ // success, k8s option provided
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeAlpha,
				ReleaseBranch: git.DefaultBranch,
				BuildVersion:  "1.20.0-alpha.1",
			},
			submit:      true,
			shouldError: false,
		},
		{ // success, manual option provided
			provided: &obs.Options{
				Project: "foo",
			},
			submit:      true,
			shouldError: false,
		},
		{ // invalid, neither k8s or manual option provided
			provided:    &obs.Options{},
			submit:      true,
			shouldError: true,
		},
		{ // invalid, both k8s and manual option provided
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeAlpha,
				ReleaseBranch: git.DefaultBranch,
				BuildVersion:  "1.20.0-alpha.1",
				Project:       "foo",
			},
			submit:      true,
			shouldError: true,
		},
		{ // success, spec template path provided when not submitting
			provided: &obs.Options{
				SpecTemplatePath: newSpecPath(t, "/path/to/spec"),
				Project:          "foo",
			},
			submit:      false,
			shouldError: false,
		},
		{ // invalid, spec template path not provided when not submitting
			provided: &obs.Options{
				Project: "foo",
			},
			submit:      false,
			shouldError: true,
		},
		{ // invalid, specs for package foo doesn't exist
			provided: &obs.Options{
				SpecTemplatePath: newSpecPath(t, ""),
				Packages:         []string{"kubeadm"},
				Project:          "foo",
			},
			submit:      false,
			shouldError: true,
		},
		{ // success, specs for package foo provided
			provided: &obs.Options{
				SpecTemplatePath: newSpecPath(t, "kubeadm"),
				Packages:         []string{"kubeadm"},
				Project:          "foo",
			},
			submit:      false,
			shouldError: false,
		},
		{ // success, provided architectures are supported
			provided: &obs.Options{
				Architectures: []string{
					consts.ArchitectureAMD64,
					consts.ArchitectureARM64,
					consts.ArchitecturePPC64,
					consts.ArchitectureS390X,
				},
				Project: "foo",
			},
			submit:      true,
			shouldError: false,
		},
		{ // success, provided architectures are set to none
			// TODO: is it valid version?
			provided: &obs.Options{
				Architectures: []string{},
				Project:       "foo",
			},
			submit:      true,
			shouldError: false,
		},
		{ // invalid provided architectures
			provided: &obs.Options{
				Architectures: []string{consts.ArchitectureRISCV},
				Project:       "foo",
			},
			submit:      true,
			shouldError: true,
		},
		{ // invalid release type
			provided: &obs.Options{
				ReleaseType:   "invalid-release-type",
				ReleaseBranch: git.DefaultBranch,
				BuildVersion:  "1.20.0-alpha.1",
				Project:       "foo",
			},
			submit:      true,
			shouldError: true,
		},
		{ // success, beta release type is supported
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeBeta,
				ReleaseBranch: git.DefaultBranch,
				BuildVersion:  "1.20.0-beta.1",
			},
			submit:      true,
			shouldError: false,
		},
		{ // success, alpha release type is supported
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeAlpha,
				ReleaseBranch: git.DefaultBranch,
				BuildVersion:  "1.20.0-alpha.1",
			},
			submit:      true,
			shouldError: false,
		},
		{ // success, rc release type is supported
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeRC,
				ReleaseBranch: git.DefaultBranch,
				BuildVersion:  "1.20.0-rc.1",
			},
			submit:      true,
			shouldError: false,
		},
		{ // success, official release type is supported
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeOfficial,
				ReleaseBranch: "release-1.20",
				BuildVersion:  "1.20.0-official.1",
			},
			submit:      true,
			shouldError: false,
		},
		{ // success, release branch is supported
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeOfficial,
				ReleaseBranch: "master",
				BuildVersion:  "1.20.0-official.1",
			},
			submit:      true,
			shouldError: false,
		},
		{ // success, release branch is supported
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeOfficial,
				ReleaseBranch: "main",
				BuildVersion:  "1.20.0-official.1",
			},
			submit:      true,
			shouldError: false,
		},
		{ // invalid, release branch is not supported
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeOfficial,
				ReleaseBranch: "invalid-branch-name",
				BuildVersion:  "1.20.0-official.1",
			},
			submit:      true,
			shouldError: true,
		},
		{ // invalid, release branch is missing
			provided: &obs.Options{
				ReleaseType:  release.ReleaseTypeOfficial,
				BuildVersion: "1.20.0-official.1",
			},
			submit:      true,
			shouldError: true,
		},
		{ // invalid, build version is not specified
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeOfficial,
				ReleaseBranch: "invalid-branch-name",
			},
			submit:      true,
			shouldError: true,
		},
		{ // invalid, release type is not specified
			provided: &obs.Options{
				ReleaseBranch: "invalid-branch-name",
				BuildVersion:  "1.20.0-official.1",
			},
			submit:      true,
			shouldError: true,
		},
		{ // invalid, package source is mutually exclusive `ReleaseType`,
			// `ReleaseBranch` and `BuildVersion`
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeOfficial,
				ReleaseBranch: "invalid-branch-name",
				BuildVersion:  "1.20.0-official.1",
				PackageSource: "https://package",
			},
			submit:      true,
			shouldError: true,
		},
		{ // valid, package source and project provided
			provided: &obs.Options{
				PackageSource: "https://package",
				Project:       "foo",
			},
			submit:      true,
			shouldError: false,
		},
		{ // invalid, version is mutually exclusive `ReleaseType`,
			// `ReleaseBranch` and `BuildVersion`
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeOfficial,
				ReleaseBranch: "invalid-branch-name",
				BuildVersion:  "1.20.0-official.1",
				Version:       "1.3",
			},
			submit:      true,
			shouldError: true,
		},
	} {
		err := tc.provided.Validate(tc.submit)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestValidateBuildVersion(t *testing.T) {
	for _, tc := range []struct {
		provided    *obs.Options
		shouldError bool
	}{
		{ // success
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeAlpha,
				ReleaseBranch: git.DefaultBranch,
				BuildVersion:  "v1.20.0-beta.1.203+8f6ffb24df9896",
			},
			shouldError: false,
		},
		{ // invalid build version
			provided: &obs.Options{
				ReleaseType:   release.ReleaseTypeAlpha,
				ReleaseBranch: git.DefaultBranch,
				BuildVersion:  "invalid",
			},
			shouldError: true,
		},
	} {
		state := obs.DefaultState()

		err := tc.provided.ValidateBuildVersion(state)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestStagingOptionsValidate(t *testing.T) {
	for _, tc := range []struct {
		provided    *obs.StageOptions
		shouldError bool
		submit      bool
	}{
		{ // valid k8s build option should validate
			provided: &obs.StageOptions{
				&obs.Options{
					ReleaseType:   release.ReleaseTypeAlpha,
					ReleaseBranch: git.DefaultBranch,
					BuildVersion:  "v1.20.0-beta.1.203+8f6ffb24df9896",
				},
			},
			submit:      true,
			shouldError: false,
		},
		{ // valid manual build version should validate
			provided: &obs.StageOptions{
				&obs.Options{
					Project: "foo",
					Version: "1.20.0",
				},
			},
			submit:      true,
			shouldError: false,
		},
		{ // invalid build version should not validate
			provided: &obs.StageOptions{
				&obs.Options{
					ReleaseType:   release.ReleaseTypeAlpha,
					ReleaseBranch: git.DefaultBranch,
					BuildVersion:  "invalid",
				},
			},
			submit:      true,
			shouldError: true,
		},
		{ // version is required for stage, invalid if missing
			provided: &obs.StageOptions{
				&obs.Options{
					Project: "foo",
				},
			},
			submit:      true,
			shouldError: true,
		},
	} {
		state := obs.DefaultState()

		err := tc.provided.Validate(state, tc.submit)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestReleaseOptionsValidate(t *testing.T) {
	for _, tc := range []struct {
		provided    *obs.ReleaseOptions
		shouldError bool
		submit      bool
	}{
		{ // valid k8s build option should validate
			provided: &obs.ReleaseOptions{
				&obs.Options{
					ReleaseType:   release.ReleaseTypeBeta,
					ReleaseBranch: git.DefaultBranch,
					BuildVersion:  "v1.20.0-beta.1.203+8f6ffb24df9897",
				},
			},
			submit:      true,
			shouldError: false,
		},
		{ // invalid build version should not validate
			provided: &obs.ReleaseOptions{
				&obs.Options{
					ReleaseType:   release.ReleaseTypeBeta,
					ReleaseBranch: git.DefaultBranch,
					BuildVersion:  "invalid",
				},
			},
			submit:      true,
			shouldError: true,
		},
		{ // invalid manual build, version shouldn't be specified
			// for the release
			provided: &obs.ReleaseOptions{
				&obs.Options{
					Project: "foo",
					Version: "1.21.0",
				},
			},
			submit:      true,
			shouldError: true,
		},
		{ // valid manual build, version shouldn't be specified
			// for the release
			provided: &obs.ReleaseOptions{
				&obs.Options{
					Project: "foo",
				},
			},
			submit:      true,
			shouldError: false,
		},
	} {
		state := obs.DefaultState()

		err := tc.provided.Validate(state, tc.submit)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func mockGenerateReleaseVersionStage(mock *obsfakes.FakeStageClient) {
	mock.GenerateReleaseVersionReturns(nil)
}

func mockGenerateReleaseVersionRelease(mock *obsfakes.FakeReleaseClient) {
	mock.GenerateReleaseVersionReturns(nil)
}

func TestRunStage(t *testing.T) {
	err := errors.New("error")

	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageClient)
		shouldError bool
	}{
		{ // success
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.GenerateReleaseVersionReturns(nil)
			},
			shouldError: false,
		},
		{ // ValidateOptions fails
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.ValidateOptionsReturns(err)
			},
			shouldError: true,
		},
		{ // Initializing OBS root and config fails
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.InitOBSRootReturns(err)
			},
			shouldError: true,
		},
		{ // CheckPrerequisites fails
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.CheckPrerequisitesReturns(err)
			},
			shouldError: true,
		},
		{ // CheckReleaseBranchState fails
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.CheckReleaseBranchStateReturns(err)
			},
			shouldError: true,
		},
		{ // GenerateReleaseVersion fails
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.GenerateReleaseVersionReturns(err)
			},
			shouldError: true,
		},
		{ // Generating OBS project name
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.GenerateOBSProjectReturns(err)
			},
			shouldError: true,
		},
		{ // Checking out OBS project
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.CheckoutOBSProjectReturns(err)
			},
			shouldError: true,
		},
		{ // Generating spec files and artifact archives
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.GeneratePackageArtifactsReturns(err)
			},
			shouldError: true,
		},
		{ // Pushing packages to OBS
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.PushReturns(err)
			},
			shouldError: true,
		},
		{ // wait for OBS build results
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.WaitReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultStageOptions()
		sut := obs.NewStage(opts)
		mock := &obsfakes.FakeStageClient{}
		tc.prepare(mock)
		sut.SetClient(mock)

		err := sut.Run()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestRunRelease(t *testing.T) {
	err := errors.New("error")

	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeReleaseClient)
		shouldError bool
	}{
		{ // success
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.GenerateReleaseVersionReturns(nil)
			},
			shouldError: false,
		},
		{ // ValidateOptions fails
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.ValidateOptionsReturns(err)
			},
			shouldError: true,
		},
		{ // Initializing OBS root and config fails
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.InitOBSRootReturns(err)
			},
			shouldError: true,
		},
		{ // CheckPrerequisites fails
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.CheckPrerequisitesReturns(err)
			},
			shouldError: true,
		},
		{ // CheckReleaseBranchState fails
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.CheckReleaseBranchStateReturns(err)
			},
			shouldError: true,
		},
		{ // GenerateReleaseVersion fails
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.GenerateReleaseVersionReturns(err)
			},
			shouldError: true,
		},
		{ // Generating OBS project name
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.GenerateOBSProjectReturns(err)
			},
			shouldError: true,
		},
		{ // Checking out OBS project
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.CheckoutOBSProjectReturns(err)
			},
			shouldError: true,
		},
		{ // Releasing packages to OBS
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.ReleasePackagesReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultReleaseOptions()
		sut := obs.NewRelease(opts)
		mock := &obsfakes.FakeReleaseClient{}
		tc.prepare(mock)
		sut.SetClient(mock)

		err := sut.Run()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestSubmitStage(t *testing.T) {
	err := errors.New("error")

	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageClient)
		shouldError bool
	}{
		{ // success
			prepare: func(mock *obsfakes.FakeStageClient) {
				mockGenerateReleaseVersionStage(mock)
			},
			shouldError: false,
		},
		{ // Submit fails
			prepare: func(mock *obsfakes.FakeStageClient) {
				mock.SubmitReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultStageOptions()
		sut := obs.NewStage(opts)
		mock := &obsfakes.FakeStageClient{}
		tc.prepare(mock)
		sut.SetClient(mock)

		err := sut.Submit(false)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestSubmitRelease(t *testing.T) {
	err := errors.New("error")

	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeReleaseClient)
		shouldError bool
	}{
		{ // success
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mockGenerateReleaseVersionRelease(mock)
			},
			shouldError: false,
		},
		{ // Submit fails
			prepare: func(mock *obsfakes.FakeReleaseClient) {
				mock.SubmitReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultReleaseOptions()
		sut := obs.NewRelease(opts)
		mock := &obsfakes.FakeReleaseClient{}
		tc.prepare(mock)
		sut.SetClient(mock)

		err := sut.Submit(false)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func newSpecPath(t *testing.T, file string) string {
	tempDir := t.TempDir()

	require.NoError(t, os.MkdirAll(
		filepath.Join(tempDir, file),
		os.FileMode(0o755),
	))

	return tempDir
}

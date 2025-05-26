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

	"k8s.io/release/pkg/obs"
	"k8s.io/release/pkg/obs/obsfakes"
)

var err = errors.New("error")

func TestInitOBSRoot(t *testing.T) {
	testcases := []struct {
		username  string
		password  string
		shouldErr bool
	}{
		{ // no password
			shouldErr: true,
		},
		{ // success
			password:  "bar",
			shouldErr: false,
		},
		{ // success
			username:  "foo",
			password:  "bar",
			shouldErr: false,
		},
	}

	for _, tc := range testcases {
		t.Setenv(obs.OBSPasswordKey, tc.password)
		t.Setenv(obs.OBSUsernameKey, tc.username)

		mock := &obsfakes.FakeStageImpl{}
		sut := obs.NewDefaultStage(obs.DefaultStageOptions())
		sut.SetImpl(mock)

		err := sut.InitOBSRoot()

		if tc.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, 1, mock.CreateOBSConfigFileCallCount())
		}
	}
}

func TestSubmit(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageImpl)
		shouldError bool
	}{
		{ // Success
			prepare:     func(*obsfakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // Submit fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.SubmitReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultStageOptions()
		sut := obs.NewDefaultStage(opts)

		mock := &obsfakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.Submit(false)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestCheckPrerequisitesStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageImpl)
		shouldError bool
	}{
		{ // Success
			prepare:     func(*obsfakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // CheckPrerequisites fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.CheckPrerequisitesReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultStageOptions()
		sut := obs.NewDefaultStage(opts)

		mock := &obsfakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.CheckPrerequisites()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestCheckReleaseBranchStateStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*obsfakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // BranchNeedsCreation fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.BranchNeedsCreationReturns(false, err)
			},
			shouldError: true,
		},
	} {
		opts := preconfigureStageOptions(t)
		sut := generateTestingStageState(t, opts)

		mock := &obsfakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.CheckReleaseBranchState()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestGenerateReleaseVersionStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*obsfakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // GenerateReleaseVersion fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.GenerateReleaseVersionReturns(nil, err)
			},
			shouldError: true,
		},
	} {
		opts := preconfigureStageOptions(t)
		sut := generateTestingStageState(t, opts)

		mock := &obsfakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.GenerateReleaseVersion()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestCheckoutOBSProjectStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*obsfakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // CheckoutProject fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.CheckoutProjectReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultStageOptions()
		sut := obs.NewDefaultStage(opts)

		sut.SetState(obs.DefaultStageState())

		mock := &obsfakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.CheckoutOBSProject()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestGeneratePackageArtifactsStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*obsfakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // RemovePackageFiles fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.RemovePackageFilesReturns(err)
			},
			shouldError: true,
		},
		{ // GenerateSpecsAndArtifacts fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.GenerateSpecsAndArtifactsReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultStageOptions()
		sut := obs.NewDefaultStage(opts)

		sut.SetState(obs.DefaultStageState())

		mock := &obsfakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.GeneratePackageArtifacts()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestPushStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*obsfakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // AddRemoveChanges fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.AddRemoveChangesReturns(err)
			},
			shouldError: true,
		},
		{ // CommitChanges fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.CommitChangesReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultStageOptions()
		opts.NoMock = true
		sut := obs.NewDefaultStage(opts)
		sut.SetState(obs.DefaultStageState())

		mock := &obsfakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.Push()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestWaitStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*obsfakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // Wait always fails
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.WaitReturns(err)
			},
			shouldError: true,
		},
		{ // Wait fails once
			prepare: func(mock *obsfakes.FakeStageImpl) {
				mock.WaitReturnsOnCall(0, err)
			},
			shouldError: false,
		},
	} {
		opts := obs.DefaultStageOptions()
		opts.NoMock = true
		opts.Wait = true
		sut := obs.NewDefaultStage(opts)
		sut.SetState(obs.DefaultStageState())

		mock := &obsfakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.Wait()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func preconfigureStageOptions(t *testing.T) *obs.StageOptions {
	opts := obs.DefaultStageOptions()
	opts.ReleaseType = "alpha"
	opts.ReleaseBranch = "release-1.20"
	opts.BuildVersion = "v1.20.0"
	opts.SpecTemplatePath = newSpecPathWithPackages(t, "/path/to/spec", opts.Packages)

	return opts
}

func generateTestingStageState(t *testing.T, opts *obs.StageOptions) *obs.DefaultStage {
	sut := obs.NewDefaultStage(opts)
	sut.SetState(obs.DefaultStageState())

	err := sut.ValidateOptions()
	require.NoError(t, err)

	return sut
}

func newSpecPathWithPackages(t *testing.T, file string, packages []string) string {
	tempDir := t.TempDir()

	require.NoError(t, os.MkdirAll(
		filepath.Join(tempDir, file),
		os.FileMode(0o755),
	))

	for _, pkg := range packages {
		require.NoError(t, os.MkdirAll(
			filepath.Join(tempDir, pkg),
			os.FileMode(0o755),
		))
	}

	return tempDir
}

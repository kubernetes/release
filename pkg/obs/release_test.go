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
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/obs"
	"k8s.io/release/pkg/obs/obsfakes"
)

func TestInitOBSRootRelease(t *testing.T) {
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

		mock := &obsfakes.FakeReleaseImpl{}
		sut := obs.NewDefaultRelease(obs.DefaultReleaseOptions())
		sut.SetImpl(mock)

		err := sut.InitOBSRoot()

		if tc.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, 1, mock.CreateOBSConfigFileCallCount())
			require.Equal(t, 1, mock.MkdirAllCallCount())
		}
	}
}

func TestCheckPrerequisitesRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // Success
			prepare:     func(*obsfakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // CheckPrerequisites fails
			prepare: func(mock *obsfakes.FakeReleaseImpl) {
				mock.CheckPrerequisitesReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultReleaseOptions()
		sut := obs.NewDefaultRelease(opts)

		mock := &obsfakes.FakeReleaseImpl{}
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

func TestCheckReleaseBranchStateRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*obsfakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // BranchNeedsCreation fails
			prepare: func(mock *obsfakes.FakeReleaseImpl) {
				mock.BranchNeedsCreationReturns(false, err)
			},
			shouldError: true,
		},
	} {
		opts := preconfigureReleaseOptions(t)
		sut := generateTestingReleaseState(t, opts)

		mock := &obsfakes.FakeReleaseImpl{}
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

func TestGenerateReleaseVersionRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*obsfakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // GenerateReleaseVersion fails
			prepare: func(mock *obsfakes.FakeReleaseImpl) {
				mock.GenerateReleaseVersionReturns(nil, err)
			},
			shouldError: true,
		},
	} {
		opts := preconfigureReleaseOptions(t)
		sut := generateTestingReleaseState(t, opts)

		mock := &obsfakes.FakeReleaseImpl{}
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

func TestCheckoutOBSProjectRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*obsfakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // CheckoutProject fails
			prepare: func(mock *obsfakes.FakeReleaseImpl) {
				mock.CheckoutProjectReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := obs.DefaultReleaseOptions()
		sut := obs.NewDefaultRelease(opts)

		sut.SetState(obs.DefaultReleaseState())

		mock := &obsfakes.FakeReleaseImpl{}
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

func TestReleasePackagesRelease(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*obsfakes.FakeReleaseImpl)
		shouldError bool
	}{
		{ // Success
			prepare:     func(*obsfakes.FakeReleaseImpl) {},
			shouldError: false,
		},
		{ // CheckPrerequisites fails
			prepare: func(mock *obsfakes.FakeReleaseImpl) {
				mock.ReleasePackageReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := preconfigureReleaseOptions(t)
		opts.NoMock = true
		sut := generateTestingReleaseState(t, opts)

		mock := &obsfakes.FakeReleaseImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.ReleasePackages()
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, 3, mock.ReleasePackageCallCount())
		}
	}
}

func preconfigureReleaseOptions(t *testing.T) *obs.ReleaseOptions {
	opts := obs.DefaultReleaseOptions()
	opts.ReleaseType = "alpha"
	opts.ReleaseBranch = "release-1.20"
	opts.BuildVersion = "v1.20.0"
	opts.SpecTemplatePath = newSpecPathWithPackages(t, "/path/to/spec", opts.Packages)

	return opts
}

func generateTestingReleaseState(t *testing.T, opts *obs.ReleaseOptions) *obs.DefaultRelease {
	sut := obs.NewDefaultRelease(opts)
	sut.SetState(obs.DefaultReleaseState())

	err := sut.ValidateOptions()
	require.NoError(t, err)

	return sut
}

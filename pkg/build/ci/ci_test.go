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

package ci_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/build/ci"
	"k8s.io/release/pkg/build/ci/cifakes"
)

var err = errors.New("error")

//nolint
var testVersion string = "v1.20.0-beta.1.655+d20e3246bade17"

//nolint
var testGCSPaths = []string{
	"gs://k8s-release-dev/ci/v1.20.0-beta.1.655+d20e3246bade17",
	"gs://k8s-release-dev/ci/v1.20.0-beta.1.655+d20e3246bade17/kubernetes.tar.gz",
	"gs://k8s-release-dev/ci/v1.20.0-beta.1.655+d20e3246bade17/bin",
}

type testStateParameters struct {
	version *string
}

func generateTestingState(params *testStateParameters) *ci.State {
	state := ci.DefaultState()

	// TODO: Populate logic
	if params.version != nil { //nolint: staticcheck
	}

	return state
}

//nolint
func chdirKubernetes() {
	// TODO: Do this a little bit more elegantly
	wd, _ := os.Getwd()
	gopath := os.Getenv("GOPATH")
	kubernetesDir := filepath.Join(gopath, "/src/k8s.io/kubernetes")

	defer os.Chdir(wd)

	os.Chdir(kubernetesDir)
}

func TestBuild(t *testing.T) {
	for _, tc := range []struct {
		name        string
		prepare     func(*cifakes.FakeClient)
		shouldError bool
	}{
		{
			name: "success",
			prepare: func(mock *cifakes.FakeClient) {
				mock.IsKubernetesRepoReturnsOnCall(0, true, nil)
			},
			shouldError: false,
		},

		{
			name: "IsKubernetesRepo fails",
			prepare: func(mock *cifakes.FakeClient) {
				mock.IsKubernetesRepoReturns(true, nil)
			},
			shouldError: true,
		},
		{
			name: "IsKubernetesRepo fails",
			prepare: func(mock *cifakes.FakeClient) {
				mock.IsKubernetesRepoReturns(false, err)
			},
			shouldError: true,
		},
		{
			name: "IsKubernetesRepo fails",
			prepare: func(mock *cifakes.FakeClient) {
				mock.IsKubernetesRepoStub = func() (bool, error) {
					return true, nil
				}
			},
			shouldError: false,
		},
	} {
		t.Logf("test case: %v", tc.name)
		sut := ci.NewDefaultBuild()

		// TODO: Populate logic
		sut.SetState(
			generateTestingState(&testStateParameters{}),
		)

		mock := &cifakes.FakeClient{}

		tc.prepare(mock)
		sut.SetClient(mock)

		err := sut.Build()

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestIsKubernetesRepo(t *testing.T) {
	for _, tc := range []struct {
		name        string
		prepare     func(*cifakes.FakeClient)
		chdirK8s    bool
		shouldError bool
	}{
		{
			name: "success",
			prepare: func(mock *cifakes.FakeClient) {
			},
			chdirK8s:    true,
			shouldError: false,
		},
		{
			name: "IsKubernetesRepo fails",
			prepare: func(mock *cifakes.FakeClient) {
			},
			chdirK8s:    false,
			shouldError: true,
		},
	} {
		t.Logf("test case: %v", tc.name)
		sut := ci.NewDefaultBuild()

		// TODO: Populate logic
		sut.SetState(
			generateTestingState(&testStateParameters{}),
		)

		mock := &cifakes.FakeClient{}

		if tc.chdirK8s {
			chdirKubernetes()
		}

		tc.prepare(mock)
		sut.SetClient(mock)

		isK8sRepo, err := sut.IsKubernetesRepo()

		require.Equal(t, tc.chdirK8s, isK8sRepo)

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestCheckBuildExists(t *testing.T) {
	for _, tc := range []struct {
		name        string
		prepare     func(*cifakes.FakeClient)
		shouldError bool
	}{
		{
			name: "success",
			prepare: func(mock *cifakes.FakeClient) {
				/*
					mock.GetWorkspaceVersionReturns(testVersion, nil)
					mock.GetGCSBuildPathsReturns(testGCSPaths, nil)
					mock.GCSPathsExistReturns(true, nil)
					mock.ImagesExistReturns(true, nil)
				*/
				//mock.GetGCSBuildPathsReturns([]string{}, nil)
			},
			// TODO: WRONG
			shouldError: false,
		},
		{
			name: "GetWorkspaceVersion fails with err",
			prepare: func(mock *cifakes.FakeClient) {
				mock.GetWorkspaceVersionReturns("", err)
			},
			shouldError: true,
		},
		{
			name: "GetWorkspaceVersion fails on empty version",
			prepare: func(mock *cifakes.FakeClient) {
				mock.GetWorkspaceVersionReturns("", nil)
			},
			shouldError: true,
		},
		{
			name: "GetGCSBuildPaths fails with err",
			prepare: func(mock *cifakes.FakeClient) {
				mock.GetGCSBuildPathsReturns([]string{}, err)
			},
			shouldError: true,
		},
		{
			name: "ImagesExist fails with err",
			prepare: func(mock *cifakes.FakeClient) {
				mock.ImagesExistReturns(false, err)
			},
			shouldError: true,
		},
	} {
		t.Logf("test case: %v", tc.name)
		sut := ci.NewDefaultBuild()

		// TODO: Populate logic
		sut.SetState(
			generateTestingState(&testStateParameters{}),
		)

		mock := &cifakes.FakeClient{}

		tc.prepare(mock)
		sut.SetClient(mock)

		_, err := sut.CheckBuildExists()

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

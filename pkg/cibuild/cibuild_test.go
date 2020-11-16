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

package cibuild_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/cibuild"
	"k8s.io/release/pkg/cibuild/cibuildfakes"
)

var err = errors.New("error")

var testVersionTag string = "v1.20.0"

type testStateParameters struct {
}

func generateTestingState(params *testStateParameters) *cibuild.State {
	state := cibuild.DefaultState()

	// TODO: Populate logic

	return state
}

func TestBuild(t *testing.T) {
	for _, tc := range []struct {
		prepare          func(*cibuildfakes.FakeImpl)
		isKubernetesRepo bool
		shouldError      bool
	}{ /*
			{ // success
				prepare: func(mock *cibuildfakes.FakeImpl) {
					mock.IsKubernetesRepoReturnsOnCall(0, true, nil)
				},
				//isKubernetesRepo: true,
				shouldError: false,
			},
		*/
		{ // IsKubernetesRepo fails
			prepare: func(mock *cibuildfakes.FakeImpl) {
				mock.IsKubernetesRepoReturns(true, nil)
			},
			shouldError: false,
		}, /*
			{ // IsKubernetesRepo fails
				prepare: func(mock *cibuildfakes.FakeImpl) {
					mock.IsKubernetesRepoReturns(false, err)
				},
				shouldError: true,
			},
			{ // IsKubernetesRepo fails
				prepare: func(mock *cibuildfakes.FakeImpl) {
					mock.IsKubernetesRepoStub = func() (bool, error) {
						return true, nil
					}
				},
				shouldError: false,
			},*/
	} {
		sut := cibuild.NewDefaultBuild()

		// TODO: Populate logic
		sut.SetState(
			generateTestingState(&testStateParameters{}),
		)

		mock := &cibuildfakes.FakeImpl{}

		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.Build()

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

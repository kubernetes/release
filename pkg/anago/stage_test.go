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

// nolint: dupl
package anago_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/anago"
	"k8s.io/release/pkg/anago/anagofakes"
)

func TestCheckPrerequisitesStage(t *testing.T) {
	opts := anago.DefaultStageOptions()
	sut := anago.NewDefaultStage(opts)
	mock := &anagofakes.FakeStageImpl{}
	sut.SetClient(mock)
	require.Nil(t, sut.CheckPrerequisites())
}

func TestPrepareWorkspaceStage(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*anagofakes.FakeStageImpl)
		shouldError bool
	}{
		{ // success
			prepare:     func(*anagofakes.FakeStageImpl) {},
			shouldError: false,
		},
		{ // PrepareWorkspaceStage fails
			prepare: func(mock *anagofakes.FakeStageImpl) {
				mock.PrepareWorkspaceStageReturns(err)
			},
			shouldError: true,
		},
	} {
		opts := anago.DefaultStageOptions()
		sut := anago.NewDefaultStage(opts)
		mock := &anagofakes.FakeStageImpl{}
		tc.prepare(mock)
		sut.SetClient(mock)
		err := sut.PrepareWorkspace()
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

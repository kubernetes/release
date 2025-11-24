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

package specs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/obs/specs/specsfakes"
)

var err = errors.New("error")

func TestBuildSpecs(t *testing.T) {
	testcases := []struct {
		createPkgDef func(t *testing.T) *PackageDefinition
		prepare      func(mock *specsfakes.FakeImpl)
		verifyWrites func(*testing.T, *specsfakes.FakeImpl)
		shouldErr    bool
	}{
		{ // happy path
			createPkgDef: func(t *testing.T) *PackageDefinition {
				return &PackageDefinition{
					Name:             "kubeadm",
					SpecTemplatePath: "testdata/templates",
					SpecOutputPath:   t.TempDir(),
				}
			},
			prepare:      func(mock *specsfakes.FakeImpl) {},
			verifyWrites: func(t *testing.T, fi *specsfakes.FakeImpl) {},
			shouldErr:    false,
		},
		{ // package definition is nil
			createPkgDef: func(t *testing.T) *PackageDefinition {
				return nil
			},
			prepare:      func(mock *specsfakes.FakeImpl) {},
			verifyWrites: func(t *testing.T, fi *specsfakes.FakeImpl) {},
			shouldErr:    true,
		},
		{ // spec template path doesn't exist
			createPkgDef: func(t *testing.T) *PackageDefinition {
				return &PackageDefinition{
					Name:             "kubeadm",
					SpecTemplatePath: "does_not_exist",
				}
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.StatReturns(nil, err)
			},
			verifyWrites: func(t *testing.T, fi *specsfakes.FakeImpl) {},
			shouldErr:    true,
		},
		{ // error during walking the template directory
			createPkgDef: func(t *testing.T) *PackageDefinition {
				return &PackageDefinition{
					Name:             "kubeadm",
					SpecTemplatePath: "testdata/templates",
					SpecOutputPath:   t.TempDir(),
				}
			},
			prepare: func(mock *specsfakes.FakeImpl) {
				mock.WalkReturns(err)
			},
			verifyWrites: func(t *testing.T, fi *specsfakes.FakeImpl) {},
			shouldErr:    true,
		},
	}

	for _, tc := range testcases {
		sut := &Specs{}

		pkgDef := tc.createPkgDef(t)
		mock := &specsfakes.FakeImpl{}

		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.BuildSpecs(pkgDef, false)
		if tc.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			tc.verifyWrites(t, mock)
		}
	}
}

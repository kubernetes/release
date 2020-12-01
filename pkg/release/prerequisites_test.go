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

package release_test

import (
	"errors"
	"testing"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
)

func TestCheckPrerequisites(t *testing.T) {
	err := errors.New("error")
	for _, tc := range []struct {
		prepare   func(*releasefakes.FakePrerequisitesCheckerImpl)
		shouldErr bool
	}{
		{ // success
			prepare: func(mock *releasefakes.FakePrerequisitesCheckerImpl) {
				mock.CommandAvailableReturns(true)
				mock.DockerVersionReturns("19.03.13", nil)
				mock.IsEnvSetReturns(true)
				mock.UsageReturns(
					&disk.UsageStat{Free: 101 * 1024 * 1024 * 1024}, nil,
				)
			},
			shouldErr: false,
		},
		{ // failure CommandAvailable
			prepare: func(mock *releasefakes.FakePrerequisitesCheckerImpl) {
				mock.CommandAvailableReturns(false)
			},
			shouldErr: true,
		},
		{ // failure DockerVersion
			prepare: func(mock *releasefakes.FakePrerequisitesCheckerImpl) {
				mock.CommandAvailableReturns(true)
				mock.DockerVersionReturns("", err)
			},
			shouldErr: true,
		},
		{ // failure wrong DockerVersion
			prepare: func(mock *releasefakes.FakePrerequisitesCheckerImpl) {
				mock.CommandAvailableReturns(true)
				mock.DockerVersionReturns("18.03.00", nil)
			},
			shouldErr: true,
		},
		{ // failure GCloudOutput
			prepare: func(mock *releasefakes.FakePrerequisitesCheckerImpl) {
				mock.CommandAvailableReturns(true)
				mock.DockerVersionReturns("19.03.13", nil)
				mock.GCloudOutputReturns("", err)
			},
			shouldErr: true,
		},
		{ // failure IsEnvSet
			prepare: func(mock *releasefakes.FakePrerequisitesCheckerImpl) {
				mock.CommandAvailableReturns(true)
				mock.DockerVersionReturns("19.03.13", nil)
				mock.IsEnvSetReturns(false)
			},
			shouldErr: true,
		},
		{ // failure Usage
			prepare: func(mock *releasefakes.FakePrerequisitesCheckerImpl) {
				mock.CommandAvailableReturns(true)
				mock.DockerVersionReturns("19.03.13", nil)
				mock.UsageReturns(nil, err)
			},
			shouldErr: true,
		},
		{ // failure not enough disk space
			prepare: func(mock *releasefakes.FakePrerequisitesCheckerImpl) {
				mock.CommandAvailableReturns(true)
				mock.DockerVersionReturns("19.03.13", nil)
				mock.UsageReturns(&disk.UsageStat{Free: 100}, nil)
			},
			shouldErr: true,
		},
		{ // failure ConfigureGlobalDefaultUserAndEmail
			prepare: func(mock *releasefakes.FakePrerequisitesCheckerImpl) {
				mock.CommandAvailableReturns(true)
				mock.DockerVersionReturns("19.03.13", nil)
				mock.IsEnvSetReturns(true)
				mock.UsageReturns(
					&disk.UsageStat{Free: 101 * 1024 * 1024 * 1024}, nil,
				)
				mock.ConfigureGlobalDefaultUserAndEmailReturns(err)
			},
			shouldErr: true,
		},
	} {
		mock := &releasefakes.FakePrerequisitesCheckerImpl{}
		sut := release.NewPrerequisitesChecker()
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.Run("")
		if tc.shouldErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

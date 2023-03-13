/*
Copyright 2022 The Kubernetes Authors.

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

package packages_test

import (
	"errors"
	"io"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/packages"
	"k8s.io/release/pkg/packages/packagesfakes"
)

var errTest = errors.New("")

func TestBuild(t *testing.T) {}

func TestRelease(t *testing.T) {
	t.Parallel()
	logrus.SetOutput(io.Discard)

	for _, tc := range []struct {
		name    string
		prepare func(*packagesfakes.FakeImpl)
		assert  func(error)
	}{
		{
			name:    "success",
			prepare: func(mock *packagesfakes.FakeImpl) {},
			assert:  func(err error) { require.Nil(t, err) },
		},
		{
			name: "success, but failure on Chdir restore",
			prepare: func(mock *packagesfakes.FakeImpl) {
				mock.ChdirReturnsOnCall(1, errTest)
			},
			assert: func(err error) { require.Nil(t, err) },
		},
		{
			name: "failure on Getwd",
			prepare: func(mock *packagesfakes.FakeImpl) {
				mock.GetwdReturns("", errTest)
			},
			assert: func(err error) { require.NotNil(t, err) },
		},
		{
			name: "failure on Chdir",
			prepare: func(mock *packagesfakes.FakeImpl) {
				mock.ChdirReturns(errTest)
			},
			assert: func(err error) { require.NotNil(t, err) },
		},
		{
			name: "failure on TagStringToSemver",
			prepare: func(mock *packagesfakes.FakeImpl) {
				mock.TagStringToSemverReturns(semver.Version{}, errTest)
			},
			assert: func(err error) { require.NotNil(t, err) },
		},

		{
			name: "failure on RunCommand",
			prepare: func(mock *packagesfakes.FakeImpl) {
				mock.RunCommandReturns(errTest)
			},
			assert: func(err error) { require.NotNil(t, err) },
		},
		{
			name: "failure on NormalizePath",
			prepare: func(mock *packagesfakes.FakeImpl) {
				mock.NormalizePathReturns("", errTest)
			},
			assert: func(err error) { require.NotNil(t, err) },
		},
		{
			name: "failure on CopyToRemote 0",
			prepare: func(mock *packagesfakes.FakeImpl) {
				mock.CopyToRemoteReturnsOnCall(0, errTest)
			},
			assert: func(err error) { require.NotNil(t, err) },
		},
		{
			name: "failure on CopyToRemote 1",
			prepare: func(mock *packagesfakes.FakeImpl) {
				mock.CopyToRemoteReturnsOnCall(1, errTest)
			},
			assert: func(err error) { require.NotNil(t, err) },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			t.Parallel()

			mock := &packagesfakes.FakeImpl{}
			tc.prepare(mock)

			sut := packages.New("v0.0.0")
			sut.SetImpl(mock)

			err := sut.Release()
			tc.assert(err)
		})
	}
}

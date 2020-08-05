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

package cmd_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/release/cmd/krel/cmd"
	"k8s.io/release/cmd/krel/cmd/cmdfakes"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

func TestFindGreenBuild(t *testing.T) {
	testErr := errors.New("err")
	for _, tc := range []struct {
		prepareMock   func(*cmdfakes.FakeReleaseClient)
		opts          cmd.FindGreenBuildOptions
		shouldSucceed bool
	}{
		{ // success
			prepareMock: func(mock *cmdfakes.FakeReleaseClient) {
				mock.SetReleaseVersionReturns(&release.Versions{}, nil)
			},
			shouldSucceed: true,
		},
		{ // wrong release type on master branch
			opts: cmd.FindGreenBuildOptions{
				Branch:      git.DefaultBranch,
				ReleaseType: release.ReleaseTypeRC,
			},
			shouldSucceed: false,
		},
		{ // SetReleaseVersion fails
			prepareMock: func(mock *cmdfakes.FakeReleaseClient) {
				mock.SetReleaseVersionReturns(nil, testErr)
			},
			shouldSucceed: false,
		},
		{ // SetBuildVersion fails
			prepareMock: func(mock *cmdfakes.FakeReleaseClient) {
				mock.SetBuildVersionReturns("", testErr)
			},
			shouldSucceed: false,
		},
	} {
		sut := cmd.NewFindGreenBuild()
		releaseMock := &cmdfakes.FakeReleaseClient{}
		if tc.prepareMock != nil {
			tc.prepareMock(releaseMock)
		}
		sut.SetClient(releaseMock)

		err := sut.Run(&tc.opts)
		if tc.shouldSucceed {
			require.Nil(t, err)
		} else {
			require.NotNil(t, err)
		}
	}
}

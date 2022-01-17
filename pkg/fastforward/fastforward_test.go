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

package fastforward

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/fastforward/fastforwardfakes"
)

var errTest = errors.New("test")

func TestRun(t *testing.T) {
	t.Parallel()

	const branch = "release-x.y"

	for _, tc := range []struct {
		prepare func(*fastforwardfakes.FakeImpl) *Options
		assert  func(error)
	}{
		{ // success
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.AskReturns("", true, nil)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.Nil(t, err)
			},
		},
		{ // success NonInteractive
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				return &Options{Branch: branch, NonInteractive: true}
			},
			assert: func(err error) {
				require.Nil(t, err)
			},
		},
		{ // success restore failed
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoCheckoutReturnsOnCall(1, errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.Nil(t, err)
			},
		},
		{ // success cleanup failed
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoCleanupReturns(errTest)
				return &Options{Branch: branch, Cleanup: true}
			},
			assert: func(err error) {
				require.Nil(t, err)
			},
		},
		{ // success no release branch provided
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				return &Options{NonInteractive: true}
			},
			assert: func(err error) {
				require.Nil(t, err)
			},
		},
		{ // success no fast forward required
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.RepoHasRemoteTagReturns(true, nil)
				return &Options{}
			},
			assert: func(err error) {
				require.Nil(t, err)
			},
		},
		{ // success submit
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				return &Options{Submit: true}
			},
			assert: func(err error) {
				require.Nil(t, err)
			},
		},
		{ // success token
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.EnvDefaultReturns("token")
				return &Options{Branch: branch, NonInteractive: true}
			},
			assert: func(err error) {
				require.Nil(t, err)
			},
		},
		{ // failure with token on RepoSetURL
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.EnvDefaultReturns("token")
				mock.IsDefaultK8sUpstreamReturns(true)
				mock.RepoSetURLReturns(errTest)
				return &Options{Branch: branch, NonInteractive: true}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure with token on CloneOrOpenGitHubRepo
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.EnvDefaultReturns("token")
				mock.CloneOrOpenGitHubRepoReturns(nil, errTest)
				return &Options{Branch: branch, NonInteractive: true}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on Submit
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.SubmitReturns(errTest)
				return &Options{Submit: true}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on RepoLatestReleaseBranch
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.RepoLatestReleaseBranchReturns("", errTest)
				return &Options{}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on RepoHasRemoteTag
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.RepoHasRemoteTagReturns(false, errTest)
				return &Options{}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on CloneOrOpenDefaultGitHubRepoSSH
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.CloneOrOpenDefaultGitHubRepoSSHReturns(nil, errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure not a rrlease branch
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(false)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on RepoHasRemoteBranch
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(false, errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure no remote branch
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(false, nil)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on RepoCurrentBranch
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoCurrentBranchReturns("", errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on checkout
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoCheckoutReturnsOnCall(0, errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on RepoMergeBase
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoMergeBaseReturns("", errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on first RepoDescribe
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoDescribeReturnsOnCall(0, "", errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on second RepoDescribe
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoDescribeReturnsOnCall(1, "", errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure main tag != merge base tag
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoDescribeReturnsOnCall(0, "test", nil)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on first RepoHead
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoHeadReturnsOnCall(0, "", errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on RepoMerge
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoMergeReturns(errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on second RepoHead
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.RepoHeadReturnsOnCall(1, "", errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on Ask
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.AskReturns("", false, errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
		{ // failure on RepoPush
			prepare: func(mock *fastforwardfakes.FakeImpl) *Options {
				mock.IsReleaseBranchReturns(true)
				mock.RepoHasRemoteBranchReturns(true, nil)
				mock.AskReturns("", true, nil)
				mock.RepoPushReturns(errTest)
				return &Options{Branch: branch}
			},
			assert: func(err error) {
				require.NotNil(t, err)
			},
		},
	} {
		mock := &fastforwardfakes.FakeImpl{}
		options := tc.prepare(mock)

		sut := New(options)
		sut.impl = mock

		err := sut.Run()
		tc.assert(err)
	}
}

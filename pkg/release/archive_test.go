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

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
)

func TestArchiveRelease(t *testing.T) {
	err := errors.New("Synthetic error")
	for _, tc := range []struct {
		prepare   func(*releasefakes.FakeArchiverImpl)
		shouldErr bool
	}{
		{ // Default: Success
			prepare:   func(mock *releasefakes.FakeArchiverImpl) {},
			shouldErr: false,
		},
		{ // failure Validate options fails
			prepare: func(mock *releasefakes.FakeArchiverImpl) {
				mock.ValidateOptionsReturns(err)
			},
			shouldErr: true,
		},
		{ // failure CopyReleaseLogsReturns errors
			prepare: func(mock *releasefakes.FakeArchiverImpl) {
				mock.CopyReleaseLogsReturns(err)
			},
			shouldErr: true,
		},
		{ // failure CopyReleaseToBucket fails
			prepare: func(mock *releasefakes.FakeArchiverImpl) {
				mock.CopyReleaseToBucketReturns(err)
			},
			shouldErr: true,
		},
		{ // failure CopyReleaseToBucket fails
			prepare: func(mock *releasefakes.FakeArchiverImpl) {
				mock.MakeFilesPrivateReturns(err)
			},
			shouldErr: true,
		},
		{ // CleanStagedBuilds fails
			prepare: func(mock *releasefakes.FakeArchiverImpl) {
				mock.CleanStagedBuildsReturns(err)
			},
			shouldErr: true,
		},
	} {
		mock := &releasefakes.FakeArchiverImpl{}
		sut := release.NewArchiver(&release.ArchiverOptions{})
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.ArchiveRelease()
		if tc.shouldErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

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
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
)

func TestPublishVersion(t *testing.T) {
	const (
		testVersion      = "v1.20.0-alpha.1.66+d19aec8bf1c8ca"
		olderTestVersion = "v1.20.0-alpha.0.22+00000000000000"
	)

	mockVersionMarkers := func(mock *releasefakes.FakePublisherClient) {
		mock.GSUtilOutputReturnsOnCall(0, olderTestVersion, nil)
		mock.GSUtilOutputReturnsOnCall(1, testVersion, nil)
		mock.GSUtilOutputReturnsOnCall(2, olderTestVersion, nil)
		mock.GSUtilOutputReturnsOnCall(3, testVersion, nil)
		mock.GSUtilOutputReturnsOnCall(4, olderTestVersion, nil)
	}

	for _, tc := range []struct {
		bucket  string
		gcsRoot string
		version string
		prepare func(
			*releasefakes.FakePublisherClient,
		) (buildDir string, cleanup func())
		privateBucket bool
		fast          bool
		shouldError   bool
	}{
		{ // success update fast
			bucket:  release.ProductionBucket,
			gcsRoot: "release",
			version: testVersion,
			fast:    true,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-version-test-")
				require.Nil(t, err)

				mock.GSUtilOutputReturnsOnCall(0, olderTestVersion, nil)
				mock.GSUtilOutputReturnsOnCall(1, testVersion, nil)
				mock.GetURLResponseReturns(testVersion, nil)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // failure GetMarkerPath
			bucket:  release.ProductionBucket,
			version: testVersion,
			fast:    true,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-version-test-")
				require.Nil(t, err)

				mock.GetMarkerPathReturns("", err)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // success update on private bucket
			bucket:        release.ProductionBucket,
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: true,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-version-test-")
				require.Nil(t, err)

				mockVersionMarkers(mock)
				mock.GSUtilOutputReturnsOnCall(5, testVersion, nil)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // failure update on private bucket
			bucket:        release.ProductionBucket,
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: true,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-version-test-")
				require.Nil(t, err)

				mockVersionMarkers(mock)
				mock.GSUtilOutputReturnsOnCall(5, "", errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure update on private bucket wrong content
			bucket:        release.ProductionBucket,
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: true,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-version-test-")
				require.Nil(t, err)

				mockVersionMarkers(mock)
				mock.GSUtilOutputReturnsOnCall(5, "wrong", nil)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // success update non private bucket
			bucket:        "k8s-another-bucket",
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: false,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-version-test-")
				require.Nil(t, err)

				mockVersionMarkers(mock)
				mock.GetURLResponseReturns(testVersion, nil)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // failure update non private bucket url response failed
			bucket:        "k8s-another-bucket",
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: false,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-version-test-")
				require.Nil(t, err)
				mockVersionMarkers(mock)
				mock.GetURLResponseReturns("", errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure release files do not exist
			bucket:        release.ProductionBucket,
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: false,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-version-test-")
				require.Nil(t, err)

				mock.GSUtilReturnsOnCall(0, errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure no semver version
			bucket:  release.ProductionBucket,
			gcsRoot: "release",
			version: "wrong",
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-version-test-")
				require.Nil(t, err)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
	} {
		sut := release.NewPublisher()
		clientMock := &releasefakes.FakePublisherClient{}
		sut.SetClient(clientMock)
		buildDir, cleanup := tc.prepare(clientMock)

		err := sut.PublishVersion(
			"release", tc.version, buildDir, tc.bucket, tc.gcsRoot,
			nil, tc.privateBucket, tc.fast,
		)
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
		cleanup()
	}
}

func TestPublishReleaseNotesIndex(t *testing.T) {
	err := errors.New("")
	for _, tc := range []struct {
		prepare     func(*releasefakes.FakePublisherClient)
		shouldError bool
	}{
		{ // success not existing
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.TempFileCalls(os.CreateTemp)
			},
			shouldError: false,
		},
		{ // success existing
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.TempFileCalls(os.CreateTemp)
				mock.GSUtilStatusReturns(true, nil)
			},
			shouldError: false,
		},
		{ // failure CopyToRemote
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.TempFileCalls(os.CreateTemp)
				mock.CopyToRemoteReturns(err)
			},
			shouldError: true,
		},
		{ // failure TempFile
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.TempFileReturns(nil, err)
			},
			shouldError: true,
		},
		{ // failure Marshal
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.MarshalReturns(nil, err)
			},
			shouldError: true,
		},
		{ // failure Unmarshal
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.GSUtilStatusReturns(true, nil)
				mock.UnmarshalReturns(err)
			},
			shouldError: true,
		},
		{ // failure ReadFile
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.GSUtilStatusReturns(true, nil)
				mock.ReadFileReturns(nil, err)
			},
			shouldError: true,
		},
		{ // failure CopyToLocal
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.GSUtilStatusReturns(true, nil)
				mock.CopyToLocalReturns(err)
			},
			shouldError: true,
		},
		{ // failure TempDir
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.GSUtilStatusReturns(true, nil)
				mock.TempDirReturns("", err)
			},
			shouldError: true,
		},
		{ // failure GSUtilStatus
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.GSUtilStatusReturns(false, err)
			},
			shouldError: true,
		},
		{ // failure NormalizePath 0
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.NormalizePathReturnsOnCall(0, "", err)
			},
			shouldError: true,
		},
		{ // failure NormalizePath 1
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.NormalizePathReturnsOnCall(1, "", err)
			},
			shouldError: true,
		},
	} {
		sut := release.NewPublisher()
		clientMock := &releasefakes.FakePublisherClient{}
		sut.SetClient(clientMock)
		tc.prepare(clientMock)

		err := sut.PublishReleaseNotesIndex(
			"", "", "",
		)
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

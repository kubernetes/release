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
	"io/ioutil"
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

	mockVersioMarkers := func(mock *releasefakes.FakePublisherClient) {
		mock.GSUtilOutputReturnsOnCall(0, olderTestVersion, nil)
		mock.GSUtilOutputReturnsOnCall(1, testVersion, nil)
		mock.GSUtilOutputReturnsOnCall(2, olderTestVersion, nil)
		mock.GSUtilOutputReturnsOnCall(3, testVersion, nil)
		mock.GSUtilOutputReturnsOnCall(4, olderTestVersion, nil)
	}

	for _, tc := range []struct {
		bucket  string
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
			version: testVersion,
			fast:    true,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := ioutil.TempDir("", "publish-version-test-")
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
		{ // success update on private bucket
			bucket:        release.ProductionBucket,
			version:       testVersion,
			privateBucket: true,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := ioutil.TempDir("", "publish-version-test-")
				require.Nil(t, err)

				mockVersioMarkers(mock)
				mock.GSUtilOutputReturnsOnCall(5, testVersion, nil)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // failure update on private bucket
			bucket:        release.ProductionBucket,
			version:       testVersion,
			privateBucket: true,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := ioutil.TempDir("", "publish-version-test-")
				require.Nil(t, err)

				mockVersioMarkers(mock)
				mock.GSUtilOutputReturnsOnCall(5, "", errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure update on private bucket wrong content
			bucket:        release.ProductionBucket,
			version:       testVersion,
			privateBucket: true,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := ioutil.TempDir("", "publish-version-test-")
				require.Nil(t, err)

				mockVersioMarkers(mock)
				mock.GSUtilOutputReturnsOnCall(5, "wrong", nil)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // success update non private bucket
			bucket:        "k8s-another-bucket",
			version:       testVersion,
			privateBucket: false,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := ioutil.TempDir("", "publish-version-test-")
				require.Nil(t, err)

				mockVersioMarkers(mock)
				mock.GetURLResponseReturns(testVersion, nil)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // failure update non private bucket url response failed
			bucket:        "k8s-another-bucket",
			version:       testVersion,
			privateBucket: false,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := ioutil.TempDir("", "publish-version-test-")
				require.Nil(t, err)

				mockVersioMarkers(mock)
				mock.GetURLResponseReturns("", errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure release files do not exist
			bucket:        release.ProductionBucket,
			version:       testVersion,
			privateBucket: false,
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := ioutil.TempDir("", "publish-version-test-")
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
			version: "wrong",
			prepare: func(mock *releasefakes.FakePublisherClient) (string, func()) {
				tempDir, err := ioutil.TempDir("", "publish-version-test-")
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
			"release", tc.version, buildDir, tc.bucket,
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

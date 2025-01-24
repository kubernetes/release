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
	"os"
	"testing"

	"github.com/blang/semver/v4"
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
		bucket        string
		gcsRoot       string
		version       string
		prepare       func(*releasefakes.FakePublisherClient)
		privateBucket bool
		fast          bool
		shouldError   bool
	}{
		{ // success update fast
			bucket:  release.ProductionBucket,
			gcsRoot: "release",
			version: testVersion,
			fast:    true,
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.GSUtilOutputReturnsOnCall(0, olderTestVersion, nil)
				mock.GSUtilOutputReturnsOnCall(1, testVersion, nil)
				mock.GetURLResponseReturns(testVersion, nil)
			},
			shouldError: false,
		},
		{ // failure GetMarkerPath
			bucket:      release.ProductionBucket,
			version:     testVersion,
			fast:        true,
			shouldError: true,
		},
		{ // success update on private bucket
			bucket:        release.ProductionBucket,
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: true,
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mockVersionMarkers(mock)
				mock.GSUtilOutputReturnsOnCall(5, testVersion, nil)
			},
			shouldError: false,
		},
		{ // failure update on private bucket
			bucket:        release.ProductionBucket,
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: true,
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mockVersionMarkers(mock)
				mock.GSUtilOutputReturnsOnCall(5, "", errors.New(""))
			},
			shouldError: true,
		},
		{ // failure update on private bucket wrong content
			bucket:        release.ProductionBucket,
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: true,
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mockVersionMarkers(mock)
				mock.GSUtilOutputReturnsOnCall(5, "wrong", nil)
			},
			shouldError: true,
		},
		{ // success update non private bucket
			bucket:        "k8s-another-bucket",
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: false,
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mockVersionMarkers(mock)
				mock.GetURLResponseReturns(testVersion, nil)
			},
			shouldError: false,
		},
		{ // failure update non private bucket url response failed
			bucket:        "k8s-another-bucket",
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: false,
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mockVersionMarkers(mock)
				mock.GetURLResponseReturns("", errors.New(""))
			},
			shouldError: true,
		},
		{ // failure release files do not exist
			bucket:        release.ProductionBucket,
			gcsRoot:       "release",
			version:       testVersion,
			privateBucket: false,
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.GSUtilReturnsOnCall(0, errors.New(""))
			},
			shouldError: true,
		},
		{ // failure no semver version
			bucket:      release.ProductionBucket,
			gcsRoot:     "release",
			version:     "wrong",
			shouldError: true,
		},
	} {
		sut := release.NewPublisher()
		clientMock := &releasefakes.FakePublisherClient{}
		sut.SetClient(clientMock)
		if tc.prepare != nil {
			tc.prepare(clientMock)
		}

		err := sut.PublishVersion(
			"release", tc.version, t.TempDir(), tc.bucket, tc.gcsRoot,
			nil, tc.privateBucket, tc.fast,
		)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
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
		{ // failure NormalizePath
			prepare: func(mock *releasefakes.FakePublisherClient) {
				mock.NormalizePathReturnsOnCall(0, "", err)
			},
			shouldError: true,
		},
	} {
		sut := release.NewPublisher()
		clientMock := &releasefakes.FakePublisherClient{}
		sut.SetClient(clientMock)
		tc.prepare(clientMock)

		err := sut.PublishReleaseNotesIndex(
			"gs://foo-bar/release", "gs://foo-bar/release/v1.2.3/index.json", "v1.2.3",
		)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestIsUpToDate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, oldVersion, newVersion string
		expected                     bool
	}{
		{
			name:       "Final version after RC",
			oldVersion: "1.30.0-rc.2.10+00000000000000",
			newVersion: "1.30.0-11+00000000000000",
			expected:   false,
		},
		{
			name:       "More commits",
			oldVersion: "1.30.0-10+00000000000000",
			newVersion: "1.30.0-11+00000000000000",
			expected:   false,
		},
		{
			name:       "Newer release",
			oldVersion: "1.29.0-0+00000000000000",
			newVersion: "1.29.1-0+00000000000000",
			expected:   false,
		},
		{
			name:       "Counter reset after RC",
			oldVersion: "1.30.0-rc.2.10+00000000000000",
			newVersion: "1.30.0-1+00000000000000",
			expected:   false,
		},
		{
			name:       "Patch after newer RC (artificial corner case)",
			oldVersion: "1.29.0-rc.0.20+00000000000000",
			newVersion: "1.28.1-2+00000000000000",
			expected:   true,
		},
	} {
		oldVersion := semver.MustParse(tc.oldVersion)
		newVersion := semver.MustParse(tc.newVersion)
		expected := tc.expected

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res := release.IsUpToDate(oldVersion, newVersion)
			require.Equal(t, expected, res)
		})
	}
}

func TestFixPublicReleaseNotesURL(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		input, expected string
	}{
		"should not affect correct URL": {
			input:    "https://dl.k8s.io/release/v1.32.0-beta.0/release-notes.json",
			expected: "https://dl.k8s.io/release/v1.32.0-beta.0/release-notes.json",
		},
		"should fix URL referring to production bucket": {
			input:    "gs://767373bbdcb8270361b96548387bf2a9ad0d48758c35/release/v1.29.11/release-notes.json",
			expected: "https://dl.k8s.io/release/v1.29.11/release-notes.json",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := release.FixPublicReleaseNotesURL(tc.input)
			require.Equal(t, tc.expected, res)
		})
	}
}

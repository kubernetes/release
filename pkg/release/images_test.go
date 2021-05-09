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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
)

func TestPublish(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*releasefakes.FakeCommandClient) (buildPath string, cleanup func())
		shouldError bool
	}{
		{ // success
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // success skipping wrong dirs/files
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				// arch is not a directory, should be just skipped
				require.Nil(t, os.WriteFile(
					filepath.Join(tempDir, release.ImagesPath, "wrong"),
					[]byte{}, os.FileMode(0o644),
				))

				// image is no tarball, should be just skipped
				require.Nil(t, os.WriteFile(
					filepath.Join(tempDir, release.ImagesPath, "amd64", "no-tar"),
					[]byte{}, os.FileMode(0o644),
				))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // success no images
			prepare: func(*releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // failure on docker load
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteReturnsOnCall(0, errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure on docker tag
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteReturnsOnCall(1, errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure on docker push
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteReturnsOnCall(2, errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure on docker rmi
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteReturnsOnCall(3, errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure on docker manifest create
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteReturnsOnCall(36, errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure on docker manifest annotate
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteReturnsOnCall(37, errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure on docker manifest push
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteReturnsOnCall(40, errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure get repo tag from tarball
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.RepoTagFromTarballReturnsOnCall(3, "", errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure wrong repo tag from tarball
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.RepoTagFromTarballReturnsOnCall(3, "wrong-tag", nil)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure no images-path
			prepare: func(*releasefakes.FakeCommandClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-test-")
				require.Nil(t, err)
				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
	} {
		sut := release.NewImages()
		clientMock := &releasefakes.FakeCommandClient{}
		sut.SetClient(clientMock)
		buildPath, cleanup := tc.prepare(clientMock)

		err := sut.Publish(release.GCRIOPathProd, "v1.18.9", buildPath)
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
		cleanup()
	}
}

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*releasefakes.FakeCommandClient) (buildPath string, cleanup func())
		shouldError bool
	}{
		{ // success
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteOutputReturns("digest", nil)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // failure on crane call
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteOutputReturnsOnCall(1, "", errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure no digest
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure on remote digest retrieval
			prepare: func(mock *releasefakes.FakeCommandClient) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteOutputReturns("", errors.New(""))

				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure no images-path
			prepare: func(*releasefakes.FakeCommandClient) (string, func()) {
				tempDir, err := os.MkdirTemp("", "publish-test-")
				require.Nil(t, err)
				return tempDir, func() {
					require.Nil(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
	} {
		sut := release.NewImages()
		clientMock := &releasefakes.FakeCommandClient{}
		sut.SetClient(clientMock)
		buildPath, cleanup := tc.prepare(clientMock)

		err := sut.Validate(release.GCRIOPathStaging, "v1.18.9", buildPath)
		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
		cleanup()
	}
}

func newImagesPath(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "publish-test-")
	require.Nil(t, err)

	require.Nil(t, os.MkdirAll(
		filepath.Join(tempDir, release.ImagesPath),
		os.FileMode(0o755),
	))

	return tempDir
}

func prepareImages(t *testing.T, tempDir string, mock *releasefakes.FakeCommandClient) {
	c := 0
	for _, arch := range []string{"amd64", "arm", "arm64"} {
		archPath := filepath.Join(tempDir, release.ImagesPath, arch)
		require.Nil(t, os.MkdirAll(archPath, os.FileMode(0o755)))

		for _, image := range []string{
			"conformance-amd64.tar", "kube-apiserver.tar", "kube-proxy.tar",
		} {
			require.Nil(t, os.WriteFile(
				filepath.Join(archPath, image),
				[]byte{}, os.FileMode(0o644),
			))
			mock.RepoTagFromTarballReturnsOnCall(
				c,
				fmt.Sprintf(
					"k8s.gcr.io/%s:v1.18.9",
					strings.TrimSuffix(image, ".tar"),
				),
				nil,
			)
			c++
		}
	}
}

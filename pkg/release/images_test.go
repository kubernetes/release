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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/release-utils/http"

	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
)

func TestPublish(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		prepare     func(*releasefakes.FakeImageImpl) (buildPath string, cleanup func())
		shouldError bool
	}{
		{
			name: "success",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{
			name: "success skipping wrong dirs/files",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				// arch is not a directory, should be just skipped
				require.NoError(t, os.WriteFile(
					filepath.Join(tempDir, release.ImagesPath, "wrong"),
					[]byte{}, os.FileMode(0o644),
				))

				// image is no tarball, should be just skipped
				require.NoError(t, os.WriteFile(
					filepath.Join(tempDir, release.ImagesPath, "amd64", "no-tar"),
					[]byte{}, os.FileMode(0o644),
				))

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{
			name: "success no images",
			prepare: func(*releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{
			name: "failure on docker load",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteCalls(func(_ string, args ...string) error {
					if len(args) > 0 && args[0] == "load" {
						return errors.New("")
					}

					return nil
				})

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure on docker tag",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteCalls(func(_ string, args ...string) error {
					if len(args) > 0 && args[0] == "tag" {
						return errors.New("")
					}

					return nil
				})

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure on docker push",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteCalls(func(_ string, args ...string) error {
					if len(args) > 0 && args[0] == "push" {
						return errors.New("")
					}

					return nil
				})

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure on docker rmi",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteCalls(func(_ string, args ...string) error {
					if len(args) > 0 && args[0] == "rmi" {
						return errors.New("")
					}

					return nil
				})

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure on docker manifest create",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)
				failOnManifestSubcommand(mock, "create")

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure on docker manifest annotate",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)
				failOnManifestSubcommand(mock, "annotate")

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure on docker manifest push",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)
				failOnManifestSubcommand(mock, "push")

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure get repo tag from tarball",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.RepoTagFromTarballReturnsOnCall(3, "", errors.New(""))

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure wrong repo tag from tarball",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.RepoTagFromTarballReturnsOnCall(3, "wrong-tag", nil)

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure no images-path",
			prepare: func(*releasefakes.FakeImageImpl) (string, func()) {
				return t.TempDir(), func() {}
			},
			shouldError: true,
		},
		{
			name: "failure on sign image",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.SignImageReturns(errors.New(""))

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{
			name: "failure on sign manifest",
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				// 4 arches × 3 images = 12 arch-specific sign calls (indices 0-11),
				// then manifest list signing starts at index 12.
				mock.SignImageReturnsOnCall(12, errors.New(""))

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
	} {
		prepare := tc.prepare
		shouldError := tc.shouldError

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sut := release.NewImages()
			clientMock := &releasefakes.FakeImageImpl{}
			sut.SetImpl(clientMock)

			buildPath, cleanup := prepare(clientMock)
			defer cleanup()

			err := sut.Publish(release.GCRIOPathProd, "v1.18.9", buildPath)

			if shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*releasefakes.FakeImageImpl) (buildPath string, cleanup func())
		shouldError bool
	}{
		{ // success
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteOutputReturns("digest", nil)

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: false,
		},
		{ // failure on crane call
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteOutputReturnsOnCall(1, "", errors.New(""))

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure no digest
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure on remote digest retrieval
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteOutputReturns("", errors.New(""))

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure no images-path
			prepare: func(*releasefakes.FakeImageImpl) (string, func()) {
				return t.TempDir(), func() {}
			},
			shouldError: true,
		},
		{ // failure on signature verify of image
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.VerifyImageReturns(errors.New(""))

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
		{ // failure on signature verify of manifest
			prepare: func(mock *releasefakes.FakeImageImpl) (string, func()) {
				tempDir := newImagesPath(t)
				prepareImages(t, tempDir, mock)

				mock.ExecuteOutputReturns("digest", nil)
				mock.VerifyImageReturnsOnCall(10, errors.New(""))

				return tempDir, func() {
					require.NoError(t, os.RemoveAll(tempDir))
				}
			},
			shouldError: true,
		},
	} {
		sut := release.NewImages()
		clientMock := &releasefakes.FakeImageImpl{}
		sut.SetImpl(clientMock)
		buildPath, cleanup := tc.prepare(clientMock)
		version, _ := fetchStableMarker()
		// The staging registry has a 90d retention policy, so fetch a recent version of k/k
		err := sut.Validate(release.GCRIOPathStaging, version, buildPath)
		if tc.shouldError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		cleanup()
	}
}

func newImagesPath(t *testing.T) string {
	tempDir := t.TempDir()

	require.NoError(t, os.MkdirAll(
		filepath.Join(tempDir, release.ImagesPath),
		os.FileMode(0o755),
	))

	return tempDir
}

func prepareImages(t *testing.T, tempDir string, mock *releasefakes.FakeImageImpl) {
	c := 0

	for _, arch := range []string{"amd64", "s390x", "ppc64le", "arm64"} {
		archPath := filepath.Join(tempDir, release.ImagesPath, arch)
		require.NoError(t, os.MkdirAll(archPath, os.FileMode(0o755)))

		for _, image := range []string{
			"conformance-amd64.tar", "kube-apiserver.tar", "kube-proxy.tar",
		} {
			require.NoError(t, os.WriteFile(
				filepath.Join(archPath, image),
				[]byte{}, os.FileMode(0o644),
			))

			version, _ := fetchStableMarker()
			mock.RepoTagFromTarballReturnsOnCall(
				c,
				fmt.Sprintf(
					"registry.k8s.io/%s:%s",
					strings.TrimSuffix(image, ".tar"),
					version,
				),
				nil,
			)

			c++
		}
	}
}

func failOnManifestSubcommand(mock *releasefakes.FakeImageImpl, subcmd string) {
	mock.ExecuteCalls(func(_ string, args ...string) error {
		if len(args) >= 2 && args[0] == "manifest" && args[1] == subcmd {
			return errors.New("")
		}

		return nil
	})
}

func fetchStableMarker() (string, error) {
	content, err := http.NewAgent().WithTimeout(30 * time.Second).Get("https://dl.k8s.io/release/stable.txt")
	if err != nil {
		fmt.Println("Error:", err)

		return "", err
	}

	return string(content), nil
}

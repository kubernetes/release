/*
Copyright 2019 The Kubernetes Authors.

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

package release

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/util"
)

const (
	// gcbmgr/anago defaults
	DefaultReleaseToolRepo   = "https://github.com/kubernetes/release"
	DefaultReleaseToolBranch = "master"
	DefaultProject           = "kubernetes-release-test"
	DefaultDiskSize          = "300"
	BucketPrefix             = "kubernetes-release-"

	versionReleaseRE  = `v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[a-zA-Z0-9]+)*\.*(0|[1-9][0-9]*)?`
	versionBuildRE    = `([0-9]{1,})\+([0-9a-f]{5,40})`
	versionDirtyRE    = `(-dirty)`
	dockerBuildPath   = "_output/release-tars"
	bazelBuildPath    = "bazel-bin/build/release-tars"
	bazelVersionPath  = "bazel-genfiles/version"
	dockerVersionPath = "version"
	tarballExtension  = ".tar.gz"

	// GCSStagePath is the directory where release artifacts are staged before
	// push to GCS.
	GCSStagePath = "gcs-stage"

	// ReleaseStagePath is the directory where releases are staged.
	ReleaseStagePath = "release-stage"

	// GCEPath is the directory where GCE scripts are created.
	GCEPath = "release-stage/full/kubernetes/cluster/gce"

	// GCIPath is the path for the container optimized OS for GCP.
	GCIPath = "release-stage/full/kubernetes/cluster/gce/gci"

	// ReleaseTarsPath is the directory where release artifacts are created.
	ReleaseTarsPath = "release-tars"

	// WindowsLocalPath is the directory where Windows GCE scripts are created.
	WindowsLocalPath = "release-stage/full/kubernetes/cluster/gce/windows"

	// WindowsGCSPath is the directory where Windoes GCE scripts are staged
	// before push to GCS.
	WindowsGCSPath = "gcs-stage/extra/gce/windows"
)

// BuiltWithBazel determines whether the most recent release was built with Bazel.
func BuiltWithBazel(path, releaseKind string) (bool, error) {
	tar := releaseKind + tarballExtension
	bazelBuild := filepath.Join(path, bazelBuildPath, tar)
	dockerBuild := filepath.Join(path, dockerBuildPath, tar)
	return util.MoreRecent(bazelBuild, dockerBuild)
}

// ReadBazelVersion reads the version from a Bazel build.
func ReadBazelVersion(path string) (string, error) {
	version, err := ioutil.ReadFile(filepath.Join(path, bazelVersionPath))
	return string(version), err
}

// ReadDockerizedVersion reads the version from a Dockerized build.
func ReadDockerizedVersion(path, releaseKind string) (string, error) {
	tar := releaseKind + tarballExtension
	dockerTarball := filepath.Join(path, dockerBuildPath, tar)
	versionFile := filepath.Join(releaseKind, dockerVersionPath)
	reader, err := util.ReadFileFromGzippedTar(dockerTarball, versionFile)
	if err != nil {
		return "", err
	}
	file, err := ioutil.ReadAll(reader)
	return strings.TrimSpace(string(file)), err
}

// IsValidReleaseBuild checks if build version is valid for release.
func IsValidReleaseBuild(build string) (bool, error) {
	return regexp.MatchString("("+versionReleaseRE+`(\.`+versionBuildRE+")?"+versionDirtyRE+"?)", build)
}

// IsDirtyBuild checks if build version is dirty.
func IsDirtyBuild(build string) bool {
	return strings.Contains(build, "dirty")
}

// GetKubecrossVersion returns the current kube-cross container version.
// Replaces release::kubecross_version
func GetKubecrossVersion(branches ...string) (string, error) {
	var version string

	for _, branch := range branches {
		logrus.Infof("Trying to get the kube-cross version for %s...", branch)

		versionURL := fmt.Sprintf("https://raw.githubusercontent.com/kubernetes/kubernetes/%s/build/build-image/cross/VERSION", branch)

		resp, httpErr := http.Get(versionURL)
		if httpErr != nil {
			return "", errors.Wrapf(httpErr, "an error occurred GET-ing %s", versionURL)
		}

		defer resp.Body.Close()
		body, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			return "", errors.Wrapf(ioErr, "could not handle the response body for %s", versionURL)
		}

		version = strings.TrimSpace(string(body))

		if version != "" {
			logrus.Infof("Found the following kube-cross version: %s", version)
			return version, nil
		}
	}

	return "", errors.New("kube-cross version should not be empty; cannot continue")
}

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
	"io/ioutil"
	"regexp"
	"strings"

	"k8s.io/release/pkg/util"
)

const (
	versionReleaseRE  = `v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[a-zA-Z0-9]+)*\.*(0|[1-9][0-9]*)?`
	versionDotZeroRE  = `v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.0$`
	versionBuildRE    = `([0-9]{1,})\+([0-9a-f]{5,40})`
	versionDirtyRE    = `(-dirty)`
	dockerBuildPath   = "/_output/release-tars/"
	bazelBuildPath    = "/bazel-bin/build/release-tars/"
	bazelVersionPath  = "/bazel-genfiles/version"
	dockerVersionPath = "/version"
	tarballExtension  = ".tar.gz"
)

// BuiltWithBazel determines whether the most recent release was built with Bazel.
func BuiltWithBazel(path string, releaseKind string) (bool, error) {
	bazelBuild := path + bazelBuildPath + releaseKind + tarballExtension
	dockerBuild := path + dockerBuildPath + releaseKind + tarballExtension
	return util.MoreRecent(bazelBuild, dockerBuild)
}

// ReadBazelVersion reads the version from a Bazel build.
func ReadBazelVersion(path string) (string, error) {
	version, err := ioutil.ReadFile(path + bazelVersionPath)
	return string(version), err
}

// ReadDockerizedVersion reads the version from a Dockerized
func ReadDockerizedVersion(path, releaseKind string) (string, error) {
	dockerTarball := path + dockerBuildPath + releaseKind + tarballExtension
	return util.UntarAndRead(dockerTarball, releaseKind+dockerVersionPath)
}

// IsValidReleaseBuild checks if build version is valid for release.
func IsValidReleaseBuild(build string) (bool, error) {
	return regexp.MatchString("("+versionReleaseRE+`(\.`+versionBuildRE+")?"+versionDirtyRE+"?)", build)
}

// IsDirtyBuild checks if build version is dirty.
func IsDirtyBuild(build string) bool {
	return strings.Contains(build, "dirty")
}

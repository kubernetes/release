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

package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/gcp/auth"
	"k8s.io/release/pkg/release"
)

// Build starts a Kubernetes build with the options defined in the build
// `Instance`.
func (bi *Instance) Build() error {
	workingDir := bi.opts.RepoRoot
	if workingDir == "" {
		var dirErr error
		workingDir, dirErr = os.Getwd()
		if dirErr != nil {
			return errors.Wrapf(dirErr, "getting working directory")
		}
	}
	logrus.Infof("Current working directory: %s", workingDir)

	workingDirRelative := filepath.Base(workingDir)
	expectedDirRelative := "kubernetes"

	if workingDirRelative != expectedDirRelative {
		return errors.Errorf(
			"Build was executed from the %s directory but must be run from the %s directory. Exiting...",
			workingDirRelative,
			expectedDirRelative,
		)
	}

	buildExists, buildExistsErr := bi.checkBuildExists()
	if buildExistsErr != nil {
		return errors.Wrapf(buildExistsErr, "checking if build exists")
	}

	if buildExists {
		return nil
	}

	logrus.Infof("Proceeding with a new build...")

	// TODO: Should this be configurable?
	envErr := os.Setenv("KUBE_RELEASE_RUN_TESTS", "n")
	if envErr != nil {
		return errors.Wrapf(envErr, "setting 'KUBE_RELEASE_RUN_TESTS' to 'n'")
	}

	// Configure docker client for gcr.io authentication to allow communication
	// with non-public registries.
	if bi.opts.ConfigureDocker {
		if configureErr := auth.ConfigureDocker(); configureErr != nil {
			return errors.Wrapf(configureErr, "configuring docker auth")
		}
	}

	if cleanErr := command.New(
		"make",
		"clean",
	).RunSuccess(); cleanErr != nil {
		return errors.Wrapf(cleanErr, "running make clean")
	}

	// Create a Kubernetes build
	releaseType := "release"
	if bi.opts.Fast {
		releaseType = "quick-release"
	}

	if buildErr := command.New(
		"make",
		releaseType,
	).RunSuccess(); buildErr != nil {
		return errors.Wrapf(buildErr, fmt.Sprintf("running make %s", releaseType))
	}

	// Pushing the build
	return bi.Push()
}

// checkBuildExists check if the target build exists in the specified GCS
// bucket. This allows us to speed up build jobs by not duplicating builds.
func (bi *Instance) checkBuildExists() (bool, error) {
	version, getVersionErr := release.GetWorkspaceVersion()
	if getVersionErr != nil {
		return false, errors.Wrap(getVersionErr, "getting workspace version")
	}

	bi.opts.Version = version
	if bi.opts.Version == "" {
		logrus.Infof("Failed to get a build version from the workspace")
		return false, nil
	}

	gcsBuildRoot, gcsBuildRootErr := bi.getGCSBuildPath(bi.opts.Version)
	if gcsBuildRootErr != nil {
		return false, errors.Wrap(gcsBuildRootErr, "get GCS build root")
	}

	kubernetesTar, kubernetesTarErr := bi.objStore.NormalizePath(gcsBuildRoot, release.KubernetesTar)
	if kubernetesTarErr != nil {
		return false, errors.Wrap(kubernetesTarErr, "get tarball path")
	}

	binPath, binPathErr := bi.objStore.NormalizePath(gcsBuildRoot, "bin")
	if binPathErr != nil {
		return false, errors.Wrap(binPathErr, "get binary path")
	}

	gcsBuildPaths := []string{
		gcsBuildRoot,
		kubernetesTar,
		binPath,
	}

	// TODO: Do we need to handle the errors more effectively?
	existErrors := []error{}
	for _, path := range gcsBuildPaths {
		logrus.Infof("Checking if GCS build path (%s) exists", path)
		exists, existErr := bi.objStore.PathExists(path)
		if existErr != nil || !exists {
			existErrors = append(existErrors, existErr)
		}
	}

	images := release.NewImages()
	imagesExist, imagesExistErr := images.Exists(bi.opts.Registry, version, bi.opts.Fast)
	if imagesExistErr != nil {
		existErrors = append(existErrors, imagesExistErr)
	}

	if imagesExist && len(existErrors) == 0 {
		logrus.Infof("Build already exists. Exiting...")
		return true, nil
	}

	logrus.Infof("The following error(s) occurred while looking for a build: %v", existErrors)

	return false, nil
}

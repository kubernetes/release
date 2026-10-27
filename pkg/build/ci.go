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

	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-utils/command"

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
			return fmt.Errorf("getting working directory: %w", dirErr)
		}
	}

	logrus.Infof("Current working directory: %s", workingDir)

	workingDirRelative := filepath.Base(workingDir)
	expectedDirRelative := "kubernetes"

	if workingDirRelative != expectedDirRelative {
		return fmt.Errorf(
			"build was executed from the %s directory but must be run from the %s directory, exiting",
			workingDirRelative,
			expectedDirRelative,
		)
	}

	if err := bi.checkS3Options(); err != nil {
		return err
	}

	buildExists, buildExistsErr := bi.checkBuildExists()
	if buildExistsErr != nil {
		return fmt.Errorf("checking if build exists: %w", buildExistsErr)
	}

	if buildExists {
		return nil
	}

	logrus.Infof("Proceeding with a new build...")

	// Configure docker client for gcr.io authentication to allow communication
	// with non-public registries.
	if bi.opts.ConfigureDocker {
		if configureErr := auth.ConfigureDocker(); configureErr != nil {
			return fmt.Errorf("configuring docker auth: %w", configureErr)
		}
	}

	if err := bi.makeRelease(); err != nil {
		return err
	}

	// Pushing the build
	return bi.Push()
}

// checkBuildExists check if the target build exists in the specified GCS
// bucket. This allows us to speed up build jobs by not duplicating builds.
func (bi *Instance) checkBuildExists() (bool, error) {
	version, getVersionErr := release.GetWorkspaceVersion()
	if getVersionErr != nil {
		return false, fmt.Errorf("getting workspace version: %w", getVersionErr)
	}

	bi.opts.Version = version
	if bi.opts.Version == "" {
		logrus.Infof("Failed to get a build version from the workspace")

		return false, nil
	}

	var (
		foundItems  int
		existErrors []error
	)

	if bi.isS3() {
		foundItems, existErrors = bi.s3BuildPathsExist(version)
	} else {
		var err error

		foundItems, existErrors, err = bi.gcsBuildPathsExist(version)
		if err != nil {
			return false, err
		}
	}

	images := release.NewImages()

	imagesExist, imagesExistErr := images.Exists(bi.opts.Registry, version, bi.opts.Fast)
	if imagesExistErr != nil {
		existErrors = append(existErrors, imagesExistErr)
	}
	// we are expecting atleast 3 items to be found; /version folder, kubernetes.tgz and /version/bin folder
	// if bi.opts.AllowDup is false, we want to return this function as true
	if imagesExist && len(existErrors) == 0 && foundItems >= 3 && !bi.opts.AllowDup {
		logrus.Infof("Build already exists. Exiting...")

		return true, nil
	}

	logrus.Infof("The following error(s) occurred while looking for a build: %v", existErrors)

	return false, nil
}

func (bi *Instance) gcsBuildPathsExist(version string) (foundItems int, existErrors []error, err error) {
	gcsBuildRoot, gcsBuildRootErr := bi.getGCSBuildPath(version)
	if gcsBuildRootErr != nil {
		return 0, nil, fmt.Errorf("get GCS build root: %w", gcsBuildRootErr)
	}

	kubernetesTar, kubernetesTarErr := bi.objStore.NormalizePath(gcsBuildRoot, release.KubernetesTar)
	if kubernetesTarErr != nil {
		return 0, nil, fmt.Errorf("get tarball path: %w", kubernetesTarErr)
	}

	binPath, binPathErr := bi.objStore.NormalizePath(gcsBuildRoot, "bin")
	if binPathErr != nil {
		return 0, nil, fmt.Errorf("get binary path: %w", binPathErr)
	}

	gcsBuildPaths := []string{
		gcsBuildRoot,
		kubernetesTar,
		binPath,
	}

	// TODO: Do we need to handle the errors more effectively?
	for _, path := range gcsBuildPaths {
		logrus.Infof("Checking if GCS build path (%s) exists", path)

		exists, existErr := bi.objStore.PathExists(path)
		if existErr != nil || !exists {
			existErrors = append(existErrors, existErr)
		}

		foundItems++
	}

	return foundItems, existErrors, nil
}

// makeRelease runs the Kubernetes release build in RepoRoot, or in the current
// working directory if RepoRoot is not set.
func (bi *Instance) makeRelease() error {
	if err := command.NewWithWorkDir(
		bi.opts.RepoRoot, "make", "clean",
	).RunSuccess(); err != nil {
		return fmt.Errorf("running make clean: %w", err)
	}

	releaseType := "release"
	if bi.opts.Fast {
		releaseType = "quick-release"
	}

	cmd := command.NewWithWorkDir(bi.opts.RepoRoot, "make", releaseType).
		Env("KUBE_RELEASE_RUN_TESTS=n")
	if bi.opts.KubeBuildPlatforms != "" {
		cmd.Env("KUBE_BUILD_PLATFORMS=" + bi.opts.KubeBuildPlatforms)
	}

	if err := cmd.RunSuccess(); err != nil {
		return fmt.Errorf("running make %s: %w", releaseType, err)
	}

	return nil
}

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

package cibuild

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/build"
	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/gcp/auth"
	"k8s.io/release/pkg/gcp/gcs"
	"k8s.io/release/pkg/release"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// State holds all inferred and calculated values from the release process
// it's state mutates as each step es executed
type State struct {
}

// DefaultState returns a new empty State
func DefaultState() *State {
	// The default state is empty, it will be initialized after ValidateOptions()
	// runs in CI. It will change as the CI build processes move forward
	return &State{}
}

// TODO: Populate some appropriate setters here
/*
func (s *State) SetFoo(foo string) {
	s.foo = foo
}
*/

// Instance contains the build `Instance` for running `CIBuild`.
type Instance struct {
	*build.Instance
}

// DefaultInstance returns a new Instance instance.
func DefaultInstance() *Instance {
	return &Instance{
		Instance: &build.Instance{},
	}
}

// Options contains the options for running `CIBuild`.
type Options struct {
	*build.Options
}

// DefaultOptions returns a new Options instance.
func DefaultOptions() *Options {
	return &Options{
		Options: &build.Options{
			BuildType: "ci",
			Bucket:    release.CIBucketK8sInfra,
		},
	}
}

// DefaultBuild is the default build implementation used in CI.
type DefaultBuild struct {
	impl     impl
	instance *Instance
	opts     *Options
	state    *State
}

// NewDefaultBuild creates a new defaultBuild instance.
func NewDefaultBuild() *DefaultBuild {
	return &DefaultBuild{
		&defaultBuildImpl{},
		DefaultInstance(),
		DefaultOptions(),
		nil,
	}
}

// SetImpl can be used to set the internal ciBuild implementation.
func (d *DefaultBuild) SetImpl(impl impl) {
	d.impl = impl
}

// SetState fixes the current state. Mainly used for passing
// arbitrary values during testing
func (d *DefaultBuild) SetState(state *State) {
	d.state = state
}

// defaultBuildImpl is the default internal ciBuild client implementation.
type defaultBuildImpl struct{}

//counterfeiter:generate . impl
type impl interface {
	Build(opts *build.Options) error
	IsKubernetesRepo() (bool, error)
	CheckBuildExists() (bool, error)
	GetWorkspaceVersion() (string, error)
	GetGCSBuildPaths(opts *Options) ([]string, error)
	GCSPathsExist(gcsBuildPaths []string) (bool, []error)
	ImagesExist(opts *Options) (bool, error)
	SetRunTestsEnvVar() error
	ConfigureDockerAuth(opts *Options) error
	MakeClean() error
	SetReleaseType(opts *Options) string
	MakeRelease(releaseType string) error
	Push(opts *Options) error
}

// TODO: Populate logic
func (d *defaultBuildImpl) Build(opts *build.Options) error {
	return build.NewInstance(opts).Push()
}

// TODO: Populate logic
func (d *defaultBuildImpl) IsKubernetesRepo() (bool, error) {
	workingDir, dirErr := os.Getwd()
	if dirErr != nil {
		return false, errors.Wrapf(dirErr, "getting working directory")
	}

	logrus.Infof("Current working directory: %s", workingDir)

	workingDirRelative := filepath.Base(workingDir)
	expectedDirRelative := "kubernetes"

	if workingDirRelative != expectedDirRelative {
		return false, errors.Errorf(
			"Build was executed from the %s directory but must be run from the %s directory. Exiting...",
			workingDirRelative,
			expectedDirRelative,
		)
	}

	return true, nil
}

// TODO: Populate logic
func (d *defaultBuildImpl) CheckBuildExists() (bool, error) {
	return false, nil
}

func (d *defaultBuildImpl) GetWorkspaceVersion() (string, error) {
	return release.GetWorkspaceVersion()
}

func (d *defaultBuildImpl) GetGCSBuildPaths(opts *Options) ([]string, error) {
	buildInstance := build.NewInstance(opts.Options)
	gcsBuildPaths := []string{}

	gcsBuildRoot, gcsBuildRootErr := buildInstance.GetGCSBuildPath(opts.Version)
	if gcsBuildRootErr != nil {
		return gcsBuildPaths, errors.Wrap(gcsBuildRootErr, "get GCS build root")
	}

	kubernetesTar, kubernetesTarErr := gcs.NormalizeGCSPath(gcsBuildRoot, release.KubernetesTar)
	if kubernetesTarErr != nil {
		return gcsBuildPaths, errors.Wrap(kubernetesTarErr, "get tarball path")
	}

	binPath, binPathErr := gcs.NormalizeGCSPath(gcsBuildRoot, "bin")
	if binPathErr != nil {
		return gcsBuildPaths, errors.Wrap(binPathErr, "get binary path")
	}

	gcsBuildPaths = []string{
		gcsBuildRoot,
		kubernetesTar,
		binPath,
	}

	return gcsBuildPaths, nil
}

func (d *defaultBuildImpl) GCSPathsExist(gcsBuildPaths []string) (bool, []error) {
	// TODO: Do we need to handle the errors more effectively?
	existErrors := []error{}
	for _, path := range gcsBuildPaths {
		logrus.Infof("Checking if GCS build path (%s) exists", path)
		exists, err := gcs.PathExists(path)
		if err != nil || !exists {
			existErrors = append(existErrors, err)
		}
	}

	if len(existErrors) != 0 {
		return false, existErrors
	}

	return true, existErrors
}

func (d *defaultBuildImpl) ImagesExist(opts *Options) (bool, error) {
	images := release.NewImages()
	return images.Exists(opts.Registry, opts.Version, opts.Fast)
}

// TODO: Populate logic
func (d *defaultBuildImpl) SetRunTestsEnvVar() error {
	// TODO: Should this be configurable?
	envErr := os.Setenv("KUBE_RELEASE_RUN_TESTS", "n")
	if envErr != nil {
		return errors.Wrapf(envErr, "setting 'KUBE_RELEASE_RUN_TESTS' to 'n'")
	}

	// TODO: Populate logic

	return nil
}

func (d *defaultBuildImpl) ConfigureDockerAuth(opts *Options) error {
	// Configure docker client for gcr.io authentication to allow communication
	// with non-public registries.
	if opts.ConfigureDocker {
		if err := auth.ConfigureDocker(); err != nil {
			return errors.Wrapf(err, "configuring docker auth")
		}
	}

	return nil
}

func (d *defaultBuildImpl) MakeClean() error {
	if err := command.New(
		"make",
		"clean",
	).RunSuccess(); err != nil {
		return errors.Wrapf(err, "running make clean")
	}

	return nil
}

// TODO: Populate logic
func (d *defaultBuildImpl) SetReleaseType(opts *Options) string {
	releaseType := "release"
	if opts.Fast {
		releaseType = "quick-release"
	}

	return releaseType
}

func (d *defaultBuildImpl) MakeRelease(releaseType string) error {
	if err := command.New(
		"make",
		releaseType,
	).RunSuccess(); err != nil {
		return errors.Wrapf(err, fmt.Sprintf("running make %s", releaseType))
	}

	return nil
}

func (d *defaultBuildImpl) Push(opts *Options) error {
	return build.NewInstance(opts.Options).Push()
}

// Build starts a Kubernetes build with the options defined in the build
// `Instance`.
func (d *DefaultBuild) Build() error {
	isK8sRepo, isK8sRepoErr := d.impl.IsKubernetesRepo()
	if isK8sRepoErr != nil {
		return errors.Wrap(isK8sRepoErr, "checking if in a kubernetes repo")
	}
	if !isK8sRepo {
		return errors.New("Builds can only be started from a kubernetes repo")
	}

	buildExists, buildExistsErr := d.impl.CheckBuildExists()
	if buildExistsErr != nil {
		return errors.Wrap(buildExistsErr, "checking if build exists")
	}

	if buildExists {
		return nil
	}

	logrus.Infof("Proceeding with a new build...")

	envErr := d.impl.SetRunTestsEnvVar()
	if envErr != nil {
		return errors.Wrap(envErr, "setting `KUBE_RELEASE_RUN_TESTS`")
	}

	if configureDockerAuthErr := d.impl.ConfigureDockerAuth(d.opts); configureDockerAuthErr != nil {
		return errors.Wrap(configureDockerAuthErr, "configuring docker auth")
	}

	if cleanErr := d.impl.MakeClean(); cleanErr != nil {
		return errors.Wrapf(cleanErr, "running make clean")
	}

	// Create a Kubernetes build
	releaseType := d.impl.SetReleaseType(d.opts)

	if buildErr := d.impl.MakeRelease(releaseType); buildErr != nil {
		return errors.Wrapf(buildErr, fmt.Sprintf("running make %s", releaseType))
	}

	// Pushing the build
	return d.impl.Push(d.opts)
}

// IsKubernetesRepo validates if the current working directory is a Kubernetes
// repo before starting the build.
func (d *DefaultBuild) IsKubernetesRepo() (bool, error) {
	return d.impl.IsKubernetesRepo()
}

// CheckBuildExists check if the target build exists in the specified GCS
// bucket. This allows us to speed up build jobs by not duplicating builds.
func (d *DefaultBuild) CheckBuildExists() (bool, error) {
	version, getVersionErr := d.impl.GetWorkspaceVersion()
	if getVersionErr != nil {
		return false, errors.Wrap(getVersionErr, "getting workspace version")
	}

	d.opts.Version = version
	if d.opts.Version == "" {
		return false, errors.New("there may be something wrong with the method for retrieving the workspace version")
	}

	gcsBuildPaths, gcsBuildPathsErr := d.impl.GetGCSBuildPaths(d.opts)
	if gcsBuildPathsErr != nil {
		return false, errors.Wrap(gcsBuildPathsErr, "getting GCS build paths")
	}

	buildExists, existErrors := d.impl.GCSPathsExist(gcsBuildPaths)
	imagesExist, imagesExistErr := d.impl.ImagesExist(d.opts)
	if imagesExistErr != nil {
		existErrors = append(existErrors, imagesExistErr)
	}

	if buildExists && imagesExist {
		logrus.Infof("Build already exists. Exiting...")
		return true, nil
	}

	logrus.Infof("The following error(s) occurred while looking for a build: %v", existErrors)

	return false, nil
}

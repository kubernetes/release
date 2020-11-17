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

package ci

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

// Client is the main structure for creating and pushing builds.
type Client struct {
	client
	opts  *Options
	state *State
}

// NewBuild creates a new `Build` instance.
// TODO: Needs a client to support the non-default option
func New(opts *Options) *Client {
	c := &Client{}
	c.SetOptions(opts)
	c.SetState(DefaultState())

	return c
}

// SetOptions can be used to set the internal ciBuild implementation.
func (c *Client) SetOptions(opts *Options) {
	c.opts = opts
}

// SetState fixes the current state. Mainly used for passing
// arbitrary values during testing
func (c *Client) SetState(state *State) {
	c.state = state
}

// DefaultClient returns a new Client instance.
func DefaultClient() *Client {
	return New(DefaultOptions())
}

// DefaultBuild is the default build implementation used in CI.
type DefaultBuild struct {
	*Client
}

// NewDefaultBuild creates a new defaultBuild instance.
func NewDefaultBuild() *DefaultBuild {
	return &DefaultBuild{DefaultClient()}
}

// Options contains the options for running `CIBuild`.
type Options struct {
	*build.Options
}

// DefaultOptions returns a new Options instance.
func DefaultOptions() *Options {
	return &Options{
		&build.Options{
			BuildType:                  "ci",
			CI:                         true,
			Bucket:                     release.CIBucketK8sInfra,
			GCSRoot:                    "ci",
			AllowDup:                   true,
			BuildDir:                   release.BuildDir,
			Registry:                   release.GCRIOPathCI,
			ExtraVersionMarkers:        build.DefaultExtraVersionMarkers,
			ConfigureDocker:            false,
			NoUpdateLatest:             false,
			ValidateRemoteImageDigests: true,
		},
	}
}

// State holds all inferred and calculated values from the release process
// it's state mutates as each step es executed
type State struct {
	buildExists bool
}

// DefaultState returns a new empty State
func DefaultState() *State {
	// The default state is empty, it will be initialized after ValidateOptions()
	// runs in CI. It will change as the CI build processes move forward
	return &State{}
}

//counterfeiter:generate . client
type client interface {
	// Primary code path
	Build() error
	Push() error

	// Internal code path
	SetBucket()
	IsKubernetesRepo() (bool, error)
	CheckBuildExists() (bool, error)
	GetWorkspaceVersion() (string, error)
	GetGCSBuildPaths() ([]string, error)
	GCSPathsExist(gcsBuildPaths []string) (bool, []error)
	ImagesExist() (bool, error)
	SetRunTestsEnvVar() error
	ConfigureDockerAuth() error
	MakeClean() error
	SetReleaseType() string
	MakeRelease(releaseType string) error
}

func (c *Client) SetBucket() {
	bucket := c.opts.Bucket
	if c.opts.Bucket == "" {
		if c.opts.CI {
			// TODO: Remove this once all CI and release jobs run on K8s Infra
			bucket = release.CIBucketLegacy
		}
	}

	c.opts.Bucket = bucket

	logrus.Infof("Bucket has been set to %s", c.opts.Bucket)
}

func (c *Client) SetBuildType() {
	buildType := "devel"
	if c.opts.CI {
		buildType = "ci"
	}

	c.opts.BuildType = buildType

	logrus.Infof("Build type has been set to %s", c.opts.BuildType)
}

func (c *Client) SetGCSRoot() {
	if c.opts.GCSRoot == "" {
		c.opts.GCSRoot = c.opts.BuildType
	}

	logrus.Infof("GCS root has been set to %s", c.opts.GCSRoot)
}

func (c *Client) Push() error {
	buildClient := build.New(c.opts.Options)
	buildClient.SetImpl(c.client)

	return buildClient.Push()
}

// String returns a string representation for the `ReleaseOptions` type.
func (o *Options) String() string {
	return fmt.Sprintf(
		"BuildType: %v, Bucket: %v, Registry: %v, AllowDup: %v, ValidateRemoteImageDigests: %v",
		o.BuildType, o.Bucket, o.Registry, o.AllowDup, o.ValidateRemoteImageDigests,
	)
}

// Validate if the options are correctly set.
// TODO: Populate logic
func (o *Options) Validate() (*State, error) {
	logrus.Infof("Validating generic options: %s", o.String())
	state := DefaultState()

	if o.BuildType != "ci" {
		return nil, errors.Errorf("invalid release type: %s", o.BuildType)
	}

	if o.Bucket != release.CIBucketK8sInfra &&
		o.Bucket != release.CIBucketLegacy {
		return nil, errors.Errorf("invalid GCS bucket: %s", o.Bucket)
	}

	/*
		// TODO: Adjust this value once SetBuildCandidate is done
		state.parentBranch = ""

		semverBuildVersion, err := util.TagStringToSemver(o.BuildVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid build version: %s", o.BuildVersion)
		}
		if len(semverBuildVersion.Build) == 0 {
			return nil, errors.Errorf("build version does not contain build commit")
		}
		state.semverBuildVersion = semverBuildVersion
	*/

	return state, nil
}

// SetClient can be used to set the internal ciBuild implementation.
func (d *DefaultBuild) SetClient(client client) {
	d.client = client
}

// SetState fixes the current state. Mainly used for passing
// arbitrary values during testing
func (d *DefaultBuild) SetState(state *State) {
	d.state = state
}

// Build starts a Kubernetes build with the options defined in the build
// `Client`.
func (c *Client) Build() error {
	isK8sRepo, isK8sRepoErr := c.IsKubernetesRepo()
	if isK8sRepoErr != nil {
		return errors.Wrap(isK8sRepoErr, "checking if in a kubernetes repo")
	}
	if !isK8sRepo {
		return errors.New("Builds can only be started from a kubernetes repo")
	}

	buildExists, buildExistsErr := c.CheckBuildExists()
	if buildExistsErr != nil {
		return errors.Wrap(buildExistsErr, "checking if build exists")
	}

	if buildExists {
		return nil
	}

	logrus.Infof("Proceeding with a new build..")

	envErr := c.SetRunTestsEnvVar()
	if envErr != nil {
		return errors.Wrap(envErr, "setting `KUBE_RELEASE_RUN_TESTS`")
	}

	if configureDockerAuthErr := c.ConfigureDockerAuth(); configureDockerAuthErr != nil {
		return errors.Wrap(configureDockerAuthErr, "configuring docker auth")
	}

	if cleanErr := c.MakeClean(); cleanErr != nil {
		return errors.Wrapf(cleanErr, "running make clean")
	}

	// Create a Kubernetes build
	releaseType := c.SetReleaseType()

	if buildErr := c.MakeRelease(releaseType); buildErr != nil {
		return errors.Wrapf(buildErr, fmt.Sprintf("running make %s", releaseType))
	}

	// Pushing the build
	return c.Push()
}

func (c *Client) ValidateOptions() error {
	// Call options, validate. The validation returns the initial
	// state of the build process
	state, err := c.opts.Validate()
	if err != nil {
		return errors.Wrap(err, "validating options")
	}
	c.state = state
	return nil
}

// TODO: Populate logic
func (c *Client) IsKubernetesRepo() (bool, error) {
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

// CheckBuildExists check if the target build exists in the specified GCS
// bucket. This allows us to speed up build jobs by not duplicating builds.
func (c *Client) CheckBuildExists() (bool, error) {
	version, getVersionErr := c.GetWorkspaceVersion()
	if getVersionErr != nil {
		return false, errors.Wrap(getVersionErr, "getting workspace version")
	}

	c.opts.Version = version
	if c.opts.Version == "" {
		return false, errors.New("there may be something wrong with the method for retrieving the workspace version")
	}

	gcsBuildPaths, gcsBuildPathsErr := c.GetGCSBuildPaths()
	if gcsBuildPathsErr != nil {
		return false, errors.Wrap(gcsBuildPathsErr, "getting GCS build paths")
	}

	buildExists, existErrors := c.GCSPathsExist(gcsBuildPaths)
	imagesExist, imagesExistErr := c.ImagesExist()
	if imagesExistErr != nil {
		existErrors = append(existErrors, imagesExistErr)
	}

	if buildExists && imagesExist {
		logrus.Infof("Build already exists. Exiting...")
		c.state.buildExists = true
		return true, nil
	}

	logrus.Infof("The following error(s) occurred while looking for a build: %v", existErrors)

	c.state.buildExists = false
	return false, nil
}

// TODO: Populate logic
func (c *Client) SetRunTestsEnvVar() error {
	// TODO: Should this be configurable?
	envErr := os.Setenv("KUBE_RELEASE_RUN_TESTS", "n")
	if envErr != nil {
		return errors.Wrapf(envErr, "setting 'KUBE_RELEASE_RUN_TESTS' to 'n'")
	}

	// TODO: Populate logic

	return nil
}

func (c *Client) ConfigureDockerAuth() error {
	// Configure docker client for gcr.io authentication to allow communication
	// with non-public registries.
	if c.opts.ConfigureDocker {
		if err := auth.ConfigureDocker(); err != nil {
			return errors.Wrapf(err, "configuring docker auth")
		}
	}

	return nil
}

func (c *Client) MakeClean() error {
	if err := command.New(
		"make",
		"clean",
	).RunSuccess(); err != nil {
		return errors.Wrapf(err, "running make clean")
	}

	return nil
}

// TODO: Populate logic
func (c *Client) SetReleaseType() string {
	releaseType := "release"
	if c.opts.Fast {
		releaseType = "quick-release"
	}

	return releaseType
}

func (c *Client) MakeRelease(releaseType string) error {
	if err := command.New(
		"make",
		releaseType,
	).RunSuccess(); err != nil {
		return errors.Wrapf(err, fmt.Sprintf("running make %s", releaseType))
	}

	return nil
}

func (c *Client) GetWorkspaceVersion() (string, error) {
	return release.GetWorkspaceVersion()
}

func (c *Client) GetGCSBuildPaths() ([]string, error) {
	buildClient := build.New(c.opts.Options)
	gcsBuildPaths := []string{}

	gcsBuildRoot, gcsBuildRootErr := buildClient.GetGCSBuildPath(c.opts.Version)
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

func (c *Client) GCSPathsExist(gcsBuildPaths []string) (bool, []error) {
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

func (c *Client) ImagesExist() (bool, error) {
	images := release.NewImages()
	return images.Exists(
		c.opts.Registry,
		c.opts.Version,
		c.opts.Fast,
	)
}

// Internal

// defaultImpl is the default internal ciBuild client implementation.
type defaultImpl struct{}

// TODO: Populate logic
func (d *defaultImpl) Build(opts *build.Options) error {
	return nil
}

// TODO: Populate logic
func (d *defaultImpl) CheckBuildExists() (bool, error) {
	return false, nil
}

func (d *defaultImpl) Push() error {
	return nil
}

// IsKubernetesRepo validates if the current working directory is a Kubernetes
// repo before starting the build.
func (d *defaultImpl) IsKubernetesRepo() (bool, error) {
	return true, nil
}

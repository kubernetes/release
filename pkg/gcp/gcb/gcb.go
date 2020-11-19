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

package gcb

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/gcp"
	"k8s.io/release/pkg/gcp/auth"
	"k8s.io/release/pkg/gcp/build"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/util"
)

// GCB is the main structure of this package.
type GCB struct {
	options        *Options
	repoClient     Repository
	versionClient  Version
	listJobsClient ListJobs
}

// New creates a new `*GCB` instance.
func New(options *Options) *GCB {
	return &GCB{
		repoClient:     release.NewRepo(),
		versionClient:  release.NewVersion(),
		listJobsClient: &defaultListJobsClient{},
		options:        options,
	}
}

// SetRepoClient can be used to set the internal `Repository` client.
func (g *GCB) SetRepoClient(client Repository) {
	g.repoClient = client
}

// SetVersionClient can be used to set the internal `Version` client.
func (g *GCB) SetVersionClient(client Version) {
	g.versionClient = client
}

// SetListJobsClient can be used to set the internal `ListJobs` client.
func (g *GCB) SetListJobsClient(client ListJobs) {
	g.listJobsClient = client
}

type Options struct {
	build.Options
	NoMock       bool
	Stage        bool
	Release      bool
	Stream       bool
	BuildAtHead  bool
	NoAnago      bool
	Branch       string
	ReleaseType  string
	BuildVersion string
	GcpUser      string
	LogLevel     string
	LastJobs     int64
}

// NewDefaultOptions returns a new default `*Options` instance.
func NewDefaultOptions() *Options {
	return &Options{
		LogLevel: logrus.StandardLogger().GetLevel().String(),
		Options:  *build.NewDefaultOptions(),
	}
}

//counterfeiter:generate . Repository
type Repository interface {
	Open() error
	CheckState(string, string, string, bool) error
	GetTag() (string, error)
}

//counterfeiter:generate . Version
type Version interface {
	GetKubeVersionForBranch(release.VersionType, string) (string, error)
}

//counterfeiter:generate . ListJobs
type ListJobs interface {
	ListJobs(project string, lastJobs int64) error
}

type defaultListJobsClient struct{}

func (d *defaultListJobsClient) ListJobs(project string, lastJobs int64) error {
	return build.ListJobs(project, lastJobs)
}

// Validate checks if the Options are valid.
func (o *Options) Validate() error {
	if o.Stage && o.Release {
		return errors.New("cannot specify both the 'stage' and 'release' flag; resubmit with only one build type selected")
	}

	if o.Branch == git.DefaultBranch {
		if o.ReleaseType == release.ReleaseTypeRC || o.ReleaseType == release.ReleaseTypeOfficial {
			return errors.Errorf("cannot cut a release candidate or an official release from %s", git.DefaultBranch)
		}
	} else {
		if o.ReleaseType == release.ReleaseTypeAlpha || o.ReleaseType == release.ReleaseTypeBeta {
			return errors.New("cannot cut an alpha or beta release from a release branch")
		}
	}

	if o.BuildVersion != "" && o.BuildAtHead {
		return errors.New("cannot specify both the 'build-version' and 'build-at-head' flag; resubmit with only one build option selected")
	}

	if o.BuildAtHead && o.Release {
		return errors.New("cannot specify both the 'build-at-head' flag together with the 'release' flag; resubmit with a 'build-version' flag set")
	}

	return nil
}

// Submit is the main method responsible for submitting release jobs to GCB.
func (g *GCB) Submit() error {
	if err := g.options.Validate(); err != nil {
		return errors.Wrap(err, "validating GCB options")
	}

	toolOrg := release.GetToolOrg()
	toolRepo := release.GetToolRepo()
	toolBranch := release.GetToolBranch()

	if err := gcp.PreCheck(); err != nil {
		return errors.Wrap(err, "pre-checking for GCP package usage")
	}

	if err := g.repoClient.Open(); err != nil {
		return errors.Wrap(err, "open release repo")
	}

	if err := g.repoClient.CheckState(toolOrg, toolRepo, toolBranch, g.options.NoMock); err != nil {
		return errors.Wrap(err, "verifying repository state")
	}

	logrus.Infof("Running GCB with the following options: %+v", g.options)

	g.options.NoSource = true
	g.options.DiskSize = release.DefaultDiskSize

	g.options.Async = true

	if g.options.Stream {
		g.options.Async = false
	}

	gcbSubs, gcbSubsErr := g.SetGCBSubstitutions(toolOrg, toolRepo, toolBranch)
	if gcbSubs == nil || gcbSubsErr != nil {
		return gcbSubsErr
	}

	if g.options.NoMock {
		// TODO: Consider a '--yes' flag so we can mock this
		_, nomockSubmit, askErr := util.Ask(
			fmt.Sprintf("Really submit a --nomock release job against the %s branch? (yes/no)", g.options.Branch),
			"yes",
			3,
		)
		if askErr != nil {
			return askErr
		}

		if nomockSubmit {
			gcbSubs["NOMOCK_TAG"] = "nomock"
			gcbSubs["NOMOCK"] = fmt.Sprintf("--%s", gcbSubs["NOMOCK_TAG"])
		}
	} else {
		// TODO: Remove once cloudbuild.yaml doesn't strictly require vars to be set.
		gcbSubs["NOMOCK_TAG"] = ""
		gcbSubs["NOMOCK"] = ""

		bucketPrefix := release.BucketPrefix

		userBucket := fmt.Sprintf("%s%s", bucketPrefix, gcbSubs["GCP_USER_TAG"])
		userBucketSetErr := os.Setenv("USER_BUCKET", userBucket)
		if userBucketSetErr != nil {
			return userBucketSetErr
		}

		testBucket := fmt.Sprintf("%s%s", bucketPrefix, "gcb")
		testBucketSetErr := os.Setenv("BUCKET", testBucket)
		if testBucketSetErr != nil {
			return testBucketSetErr
		}
	}

	toolRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	logrus.Info("Listing GCB substitutions prior to build submission...")
	for k, v := range gcbSubs {
		logrus.Infof("%s: %s", k, v)
	}

	var jobType string
	switch {
	// TODO: Consider a '--validate' flag to validate the GCB config without submitting
	case g.options.Stage:
		jobType = "stage"
	case g.options.Release:
		jobType = "release"
	default:
		return g.listJobs(g.options.Project, g.options.LastJobs)
	}

	// Use dedicated job types for krel-based executions
	if g.options.NoAnago {
		delete(gcbSubs, "BUILD_AT_HEAD")
		gcbSubs["LOG_LEVEL"] = g.options.LogLevel
		jobType += "-krel"
	}

	g.options.ConfigDir = filepath.Join(toolRoot, "gcb", jobType)
	prepareBuildErr := build.PrepareBuilds(&g.options.Options)
	if prepareBuildErr != nil {
		return prepareBuildErr
	}

	// TODO: Need actual values
	var jobName, uploaded string

	version, err := g.repoClient.GetTag()
	if err != nil {
		return errors.Wrap(err, "getting current tag")
	}

	return build.RunSingleJob(
		&g.options.Options, jobName, uploaded, version, gcbSubs,
	)
}

// SetGCBSubstitutions takes a set of `Options` and returns a map of GCB
// substitutions.
func (g *GCB) SetGCBSubstitutions(toolOrg, toolRepo, toolBranch string) (map[string]string, error) {
	gcbSubs := map[string]string{}

	gcbSubs["TOOL_ORG"] = toolOrg
	gcbSubs["TOOL_REPO"] = toolRepo
	gcbSubs["TOOL_BRANCH"] = toolBranch

	gcpUser := g.options.GcpUser
	if gcpUser == "" {
		var gcpUserErr error
		gcpUser, gcpUserErr = auth.GetCurrentGCPUser()
		if gcpUserErr != nil {
			return gcbSubs, gcpUserErr
		}
	} else {
		// TODO: Consider removing this once the 'gcloud auth' is testable in CI
		gcpUser = auth.NormalizeGCPUser(gcpUser)
	}

	gcbSubs["GCP_USER_TAG"] = gcpUser

	gcbSubs["TYPE"] = g.options.ReleaseType
	gcbSubs["TYPE_TAG"] = g.options.ReleaseType

	gcbSubs["RELEASE_BRANCH"] = g.options.Branch

	buildVersion := g.options.BuildVersion
	if g.options.Release && buildVersion == "" {
		return gcbSubs, errors.New("Build version must be specified when sending a release GCB run")
	}

	if g.options.Stage && g.options.BuildAtHead {
		hash, err := git.LSRemoteExec(git.GetDefaultKubernetesRepoURL(), "rev-parse", g.options.Branch)
		if err != nil {
			return gcbSubs, errors.New("failed to execute the rev-parse")
		}

		fields := strings.Fields(hash)
		if len(fields) < 1 {
			return gcbSubs, errors.Errorf("unexpected output: %s", hash)
		}

		buildVersion = fields[0]
		gcbSubs["BUILD_AT_HEAD"] = buildVersion
	}

	if buildVersion == "" {
		var versionErr error
		buildVersion, versionErr = g.versionClient.GetKubeVersionForBranch(
			release.VersionTypeCILatest, g.options.Branch,
		)
		if versionErr != nil {
			return gcbSubs, versionErr
		}

		if g.options.Stage {
			gcbSubs["BUILD_AT_HEAD"] = ""
		}
	}

	buildpoint := buildVersion
	buildpoint = strings.ReplaceAll(buildpoint, "+", "-")
	gcbSubs["BUILD_POINT"] = buildpoint
	gcbSubs["BUILDVERSION"] = buildVersion

	kubecrossBranches := []string{
		g.options.Branch,
		git.DefaultBranch,
	}

	kubecrossVersion, kubecrossVersionErr := release.GetKubecrossVersion(kubecrossBranches...)
	if kubecrossVersionErr != nil {
		return gcbSubs, kubecrossVersionErr
	}
	gcbSubs["KUBE_CROSS_VERSION"] = kubecrossVersion

	v, err := util.TagStringToSemver(buildpoint)
	if err != nil {
		return gcbSubs, errors.Errorf("Failed to parse the build point %s", buildpoint)
	}

	gcbSubs["MAJOR_VERSION_TAG"] = strconv.FormatUint(v.Major, 10)
	gcbSubs["MINOR_VERSION_TAG"] = strconv.FormatUint(v.Minor, 10)

	patch := fmt.Sprintf("%d", v.Patch)
	if g.options.ReleaseType != release.ReleaseTypeOfficial && len(v.Pre) > 0 {
		// if the release we will build is the same in the current build point then we increment
		// otherwise we are building the next type so set to 0
		if v.Pre[0].String() == g.options.ReleaseType {
			patch = fmt.Sprintf("%d-%s.%d", v.Patch, g.options.ReleaseType, v.Pre[1].VersionNum+1)
		} else if g.options.ReleaseType == release.ReleaseTypeRC && v.Pre[0].String() != release.ReleaseTypeRC {
			// Now if is RC we are building and is the first time we set to 1 since the 0 is bypassed
			patch = fmt.Sprintf("%d-%s.1", v.Patch, g.options.ReleaseType)
		} else {
			patch = fmt.Sprintf("%d-%s.0", v.Patch, g.options.ReleaseType)
		}
	}
	gcbSubs["PATCH_VERSION_TAG"] = patch
	gcbSubs["KUBERNETES_VERSION_TAG"] = fmt.Sprintf("%d.%d.%s", v.Major, v.Minor, patch)

	return gcbSubs, nil
}

// listJobs lists recent GCB jobs run in the specified project
func (g *GCB) listJobs(project string, lastJobs int64) error {
	if lastJobs < 0 {
		logrus.Infof("--list-jobs was set to a negative number, defaulting to 5")
		lastJobs = 5
	}

	logrus.Infof("Listing last %d GCB jobs:", lastJobs)
	return g.listJobsClient.ListJobs(project, lastJobs)
}

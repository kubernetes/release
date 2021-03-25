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

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/gcp"
	"k8s.io/release/pkg/gcp/auth"
	"k8s.io/release/pkg/gcp/build"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/kubecross"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-utils/util"
)

// GCB is the main structure of this package.
type GCB struct {
	options        *Options
	repoClient     Repository
	versionClient  Version
	listJobsClient ListJobs
	releaseClient  Release
}

// New creates a new `*GCB` instance.
func New(options *Options) *GCB {
	return &GCB{
		repoClient:     release.NewRepo(),
		versionClient:  release.NewVersion(),
		listJobsClient: &defaultListJobsClient{},
		releaseClient:  &defaultReleaseClient{},
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

// SetReleaseClient can be used to set the internal `Release` client.
func (g *GCB) SetReleaseClient(client Release) {
	g.releaseClient = client
}

type Options struct {
	build.Options
	NoMock       bool
	Stage        bool
	Release      bool
	Stream       bool
	BuildAtHead  bool
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

//counterfeiter:generate . Release
type Release interface {
	NeedsCreation(
		branch, releaseType string, buildVersion semver.Version,
	) (createReleaseBranch bool, err error)
	GenerateReleaseVersion(
		releaseType, version, branch string, branchFromMaster bool,
	) (*release.Versions, error)
}

type defaultReleaseClient struct{}

func (*defaultReleaseClient) NeedsCreation(
	branch, releaseType string, buildVersion semver.Version,
) (createReleaseBranch bool, err error) {
	return release.NewBranchChecker().NeedsCreation(
		branch, releaseType, buildVersion,
	)
}

func (*defaultReleaseClient) GenerateReleaseVersion(
	releaseType, version, branch string, branchFromMaster bool,
) (*release.Versions, error) {
	return release.GenerateReleaseVersion(
		releaseType, version, branch, branchFromMaster,
	)
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
	toolRef := release.GetToolRef()

	if err := gcp.PreCheck(); err != nil {
		return errors.Wrap(err, "pre-checking for GCP package usage")
	}

	if err := g.repoClient.Open(); err != nil {
		return errors.Wrap(err, "open release repo")
	}

	if err := g.repoClient.CheckState(toolOrg, toolRepo, toolRef, g.options.NoMock); err != nil {
		return errors.Wrap(err, "verifying repository state")
	}

	logrus.Infof("Running GCB with the following options: %+v", g.options)

	g.options.NoSource = true
	g.options.DiskSize = release.DefaultDiskSize

	g.options.Async = true

	if g.options.Stream {
		g.options.Async = false
	}

	gcbSubs, gcbSubsErr := g.SetGCBSubstitutions(toolOrg, toolRepo, toolRef)
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

	gcbSubs["LOG_LEVEL"] = g.options.LogLevel

	g.options.ConfigDir = filepath.Join(toolRoot, "gcb", jobType)
	prepareBuildErr := build.PrepareBuilds(&g.options.Options)
	if prepareBuildErr != nil {
		return prepareBuildErr
	}

	version, err := g.repoClient.GetTag()
	if err != nil {
		return errors.Wrap(err, "getting current tag")
	}

	return errors.Wrap(
		build.RunSingleJob(&g.options.Options, "", "", version, gcbSubs),
		"run GCB job",
	)
}

// SetGCBSubstitutions takes a set of `Options` and returns a map of GCB
// substitutions.
func (g *GCB) SetGCBSubstitutions(toolOrg, toolRepo, toolRef string) (map[string]string, error) {
	gcbSubs := map[string]string{}

	gcbSubs["TOOL_ORG"] = toolOrg
	gcbSubs["TOOL_REPO"] = toolRepo
	gcbSubs["TOOL_REF"] = toolRef

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
			return gcbSubs, errors.Wrapf(
				err, "execute rev-parse for branch %s", g.options.Branch,
			)
		}

		fields := strings.Fields(hash)
		if len(fields) < 1 {
			return gcbSubs, errors.Errorf("unexpected output: %s", hash)
		}

		buildVersion = fields[0]
	}

	if buildVersion == "" {
		var versionErr error
		buildVersion, versionErr = g.versionClient.GetKubeVersionForBranch(
			release.VersionTypeCILatest, g.options.Branch,
		)
		if versionErr != nil {
			return gcbSubs, versionErr
		}
	}

	gcbSubs["BUILDVERSION"] = buildVersion

	kc := kubecross.New()
	kcVersionBranch, err := kc.ForBranch(g.options.Branch)
	if err != nil {
		// If the kubecross version is not set, we will get a 404 from GitHub.
		// In that case, we do not err but use the latest version (unless we're on main branch)
		if g.options.Branch == git.DefaultBranch || !strings.Contains(err.Error(), "404") {
			return gcbSubs, errors.Wrap(err, "retrieve kube-cross version")
		}
		logrus.Infof("KubeCross version not set for %s, falling back to latest", g.options.Branch)
	}

	kcVersionLatest := kcVersionBranch
	if g.options.Branch != git.DefaultBranch {
		kcVersionLatest, err = kc.Latest()
		if err != nil {
			return gcbSubs, errors.Wrap(err, "retrieve latest kube-cross version")
		}

		// if kcVersionBranch is empty, the branch does not exist yet, we use
		// the latest kubecross version
		if kcVersionBranch == "" {
			kcVersionBranch = kcVersionLatest
		}
	}
	gcbSubs["KUBE_CROSS_VERSION"] = kcVersionBranch
	gcbSubs["KUBE_CROSS_VERSION_LATEST"] = kcVersionLatest

	buildVersionSemver, err := util.TagStringToSemver(buildVersion)
	if err != nil {
		return gcbSubs, errors.Wrap(err, "parse build version")
	}

	createBranch, err := g.releaseClient.NeedsCreation(
		g.options.Branch, g.options.ReleaseType, buildVersionSemver,
	)
	if err != nil {
		return nil, errors.Wrap(err, "check if branch needs to be created")
	}
	versions, err := g.releaseClient.GenerateReleaseVersion(
		g.options.ReleaseType, buildVersion,
		g.options.Branch, createBranch,
	)
	if err != nil {
		return nil, errors.Wrap(err, "generate release version")
	}
	primeSemver, err := util.TagStringToSemver(versions.Prime())
	if err != nil {
		return gcbSubs, errors.Wrap(err, "parse prime version")
	}

	gcbSubs["MAJOR_VERSION_TAG"] = strconv.FormatUint(primeSemver.Major, 10)
	gcbSubs["MINOR_VERSION_TAG"] = strconv.FormatUint(primeSemver.Minor, 10)
	gcbSubs["PATCH_VERSION_TAG"] = strconv.FormatUint(primeSemver.Patch, 10)
	gcbSubs["KUBERNETES_VERSION_TAG"] = primeSemver.String()

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

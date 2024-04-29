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
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt gcbfakes/fake_history_impl.go > gcbfakes/_fake_history_impl.go  && mv gcbfakes/_fake_history_impl.go gcbfakes/fake_history_impl.go"
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt gcbfakes/fake_list_jobs.go > gcbfakes/_fake_list_jobs.go  && mv gcbfakes/_fake_list_jobs.go gcbfakes/fake_list_jobs.go"
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt gcbfakes/fake_release.go > gcbfakes/_fake_release.go  && mv gcbfakes/_fake_release.go gcbfakes/fake_release.go"
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt gcbfakes/fake_repository.go > gcbfakes/_fake_repository.go  && mv gcbfakes/_fake_repository.go gcbfakes/fake_repository.go"
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt gcbfakes/fake_version.go > gcbfakes/_fake_version.go  && mv gcbfakes/_fake_version.go gcbfakes/fake_version.go"
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"
	gogit "github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"

	"k8s.io/release/gcb"
	"k8s.io/release/pkg/gcp/auth"
	"k8s.io/release/pkg/gcp/build"
	"k8s.io/release/pkg/kubecross"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-sdk/gcli"
	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/util"
	utilsversion "sigs.k8s.io/release-utils/version"
)

// StringSliceSeparator is the separator used for passing string slices as GCB
// substitutions.
const StringSliceSeparator = "..."

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

	// NonInteractive does not ask any questions if set to true.
	NonInteractive bool

	NoMock        bool
	Stage         bool
	Release       bool
	FastForward   bool
	Stream        bool
	BuildAtHead   bool
	Branch        string
	ReleaseType   string
	BuildVersion  string
	GcpUser       string
	LogLevel      string
	CustomK8SRepo string
	CustomK8sOrg  string
	LastJobs      int64

	// OpenBuildService parameters
	OBSStage         bool
	OBSRelease       bool
	SpecTemplatePath string
	Packages         []string
	Version          string
	Architectures    []string
	OBSProject       string
	PackageSource    string
	OBSWait          bool
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
			return fmt.Errorf("cannot cut a release candidate or an official release from %s", git.DefaultBranch)
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
		return fmt.Errorf("validating GCB options: %w", err)
	}

	toolOrg := release.GetToolOrg()
	toolRepo := release.GetToolRepo()
	toolRef := release.GetToolRef()

	if err := gcli.PreCheck(); err != nil {
		return fmt.Errorf("pre-checking for GCP package usage: %w", err)
	}

	var jobType string
	switch {
	// TODO: Consider a '--validate' flag to validate the GCB config without submitting
	case g.options.Stage:
		jobType = gcb.JobTypeStage
	case g.options.Release:
		jobType = gcb.JobTypeRelease
	case g.options.FastForward:
		jobType = gcb.JobTypeFastForward
	case g.options.OBSStage:
		jobType = gcb.JobTypeObsStage
	case g.options.OBSRelease:
		jobType = gcb.JobTypeObsRelease
	default:
		return g.listJobs(g.options.Project, g.options.LastJobs)
	}

	version := utilsversion.GetVersionInfo().GitVersion
	if err := g.repoClient.Open(); errors.Is(err, gogit.ErrRepositoryNotExists) {
		// Use the embedded cloudbuild files
		configDir, err := gcb.New().DirForJobType(jobType)
		if err != nil {
			return fmt.Errorf("get cloudbuild dir for job type: %w", err)
		}

		g.options.ConfigDir = configDir
		defer os.RemoveAll(configDir)
	} else if err != nil {
		// Any other error
		return fmt.Errorf("open release repo: %w", err)
	} else {
		// Using the local k/release repository
		if err := g.repoClient.CheckState(toolOrg, toolRepo, toolRef, g.options.NoMock); err != nil {
			return fmt.Errorf("verifying repository state: %w", err)
		}

		toolRoot, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get tool root: %w", err)
		}

		g.options.ConfigDir = filepath.Join(toolRoot, "gcb", jobType)

		version, err = g.repoClient.GetTag()
		if err != nil {
			return fmt.Errorf("getting current tag: %w", err)
		}
	}

	logrus.Infof("Running GCB with the following options: %+v", g.options)

	g.options.NoSource = true
	g.options.DiskSize = release.DefaultDiskSize

	g.options.Async = true

	if g.options.Stream {
		g.options.Async = false
	}

	// build the GCS bucket string to be used to sign all the artifacts
	bucketPrefix := release.BucketPrefix
	gcsBucket := "gs://" + bucketPrefix
	if g.options.NoMock {
		gcsBucket = strings.TrimSuffix(gcsBucket, "-")
	} else {
		gcsBucket = fmt.Sprintf("%s%s", gcsBucket, "gcb")
	}

	gcbSubs, gcbSubsErr := g.SetGCBSubstitutions(toolOrg, toolRepo, toolRef, gcsBucket)
	if gcbSubs == nil || gcbSubsErr != nil {
		return gcbSubsErr
	}

	if g.options.NoMock {
		submit := true

		if !g.options.NonInteractive {
			var err error
			_, submit, err = util.Ask(
				fmt.Sprintf("Really submit a --nomock release job against the %s branch? (yes/no)", g.options.Branch),
				"yes",
				3,
			)
			if err != nil {
				return err
			}
		}

		if submit {
			gcbSubs["NOMOCK_TAG"] = "nomock"
			gcbSubs["NOMOCK"] = "--" + gcbSubs["NOMOCK_TAG"]
		}
	} else {
		// TODO: Remove once cloudbuild.yaml doesn't strictly require vars to be set.
		gcbSubs["NOMOCK_TAG"] = ""
		gcbSubs["NOMOCK"] = ""

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

	logrus.Info("Listing GCB substitutions prior to build submission...")
	for k, v := range gcbSubs {
		logrus.Infof("%s: %s", k, v)
	}

	gcbSubs["LOG_LEVEL"] = g.options.LogLevel

	prepareBuildErr := build.PrepareBuilds(&g.options.Options)
	if prepareBuildErr != nil {
		return prepareBuildErr
	}

	if err := build.RunSingleJob(&g.options.Options, "", "", version, gcbSubs); err != nil {
		return fmt.Errorf("run GCB job: %w", err)
	}

	return nil
}

// SetGCBSubstitutions takes a set of `Options` and returns a map of GCB
// substitutions.
func (g *GCB) SetGCBSubstitutions(toolOrg, toolRepo, toolRef, gcsBucket string) (map[string]string, error) {
	gcbSubs := map[string]string{}

	gcbSubs["TOOL_ORG"] = toolOrg
	gcbSubs["TOOL_REPO"] = toolRepo
	gcbSubs["TOOL_REF"] = toolRef

	gcbSubs["K8S_ORG"] = release.GetK8sOrg()
	if g.options.CustomK8sOrg != "" {
		gcbSubs["K8S_ORG"] = g.options.CustomK8sOrg
	}

	gcbSubs["K8S_REPO"] = release.GetK8sRepo()
	if g.options.CustomK8SRepo != "" {
		gcbSubs["K8S_REPO"] = g.options.CustomK8SRepo
	}

	gcbSubs["K8S_REF"] = release.GetK8sRef()

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

	kc := kubecross.New()
	kcVersionBranch, err := kc.ForBranch(g.options.Branch)
	if err != nil {
		// If the kubecross version is not set, we will get a 404 from GitHub.
		// In that case, we do not err but use the latest version (unless we're on main branch)
		if g.options.Branch == git.DefaultBranch || !strings.Contains(err.Error(), "404") {
			return gcbSubs, fmt.Errorf("retrieve kube-cross version: %w", err)
		}
		logrus.Infof("KubeCross version not set for %s, falling back to latest", g.options.Branch)
	}

	if g.options.Branch != git.DefaultBranch {
		kcVersionLatest, err := kc.Latest()
		if err != nil {
			return gcbSubs, fmt.Errorf("retrieve latest kube-cross version: %w", err)
		}

		// if kcVersionBranch is empty, the branch does not exist yet, we use
		// the latest kubecross version
		if kcVersionBranch == "" {
			kcVersionBranch = kcVersionLatest
		}
	}
	gcbSubs["KUBE_CROSS_VERSION"] = kcVersionBranch

	switch {
	case g.options.OBSStage:
		gcbSubs["SPEC_TEMPLATE_PATH"] = g.options.SpecTemplatePath
		gcbSubs["PACKAGES"] = strings.Join(g.options.Packages, StringSliceSeparator)
		gcbSubs["ARCHITECTURES"] = strings.Join(g.options.Architectures, StringSliceSeparator)
		gcbSubs["VERSION"] = g.options.Version
		gcbSubs["OBS_PROJECT"] = g.options.OBSProject
		gcbSubs["OBS_PROJECT_TAG"] = strings.ReplaceAll(g.options.OBSProject, ":", "-")
		gcbSubs["PACKAGE_SOURCE"] = g.options.PackageSource
		gcbSubs["WAIT"] = strconv.FormatBool(g.options.OBSWait)

		// Stop here when doing OBS stage
		return gcbSubs, nil
	case g.options.OBSRelease:
		gcbSubs["PACKAGES"] = strings.Join(g.options.Packages, StringSliceSeparator)
		gcbSubs["OBS_PROJECT"] = g.options.OBSProject
		gcbSubs["OBS_PROJECT_TAG"] = strings.ReplaceAll(g.options.OBSProject, ":", "-")

		// Stop here when doing OBS release
		return gcbSubs, nil
	case g.options.FastForward:
		// Stop here when doing a fast-forward
		return gcbSubs, nil
	}

	buildVersion := g.options.BuildVersion
	if g.options.Release && buildVersion == "" {
		return gcbSubs, errors.New("build version must be specified when sending a release GCB run")
	}

	if g.options.Stage && g.options.BuildAtHead {
		hash, err := git.LSRemoteExec(git.GetDefaultKubernetesRepoURL(), "rev-parse", g.options.Branch)
		if err != nil {
			return gcbSubs, fmt.Errorf("execute rev-parse for branch %s: %w", g.options.Branch, err)
		}

		fields := strings.Fields(hash)
		if len(fields) < 1 {
			return gcbSubs, fmt.Errorf("unexpected output: %s", hash)
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

	buildVersionSemver, err := util.TagStringToSemver(buildVersion)
	if err != nil {
		return gcbSubs, fmt.Errorf("parse build version: %w", err)
	}

	createBranch, err := g.releaseClient.NeedsCreation(
		g.options.Branch, g.options.ReleaseType, buildVersionSemver,
	)
	if err != nil {
		return nil, fmt.Errorf("check if branch needs to be created: %w", err)
	}
	versions, err := g.releaseClient.GenerateReleaseVersion(
		g.options.ReleaseType, buildVersion,
		g.options.Branch, createBranch,
	)
	if err != nil {
		return nil, fmt.Errorf("generate release version: %w", err)
	}
	primeSemver, err := util.TagStringToSemver(versions.Prime())
	if err != nil {
		return gcbSubs, fmt.Errorf("parse prime version: %w", err)
	}

	gcbSubs["MAJOR_VERSION_TAG"] = strconv.FormatUint(primeSemver.Major, 10)
	gcbSubs["MINOR_VERSION_TAG"] = strconv.FormatUint(primeSemver.Minor, 10)
	gcbSubs["PATCH_VERSION_TAG"] = strconv.FormatUint(primeSemver.Patch, 10)
	gcbSubs["KUBERNETES_VERSION_TAG"] = primeSemver.String()

	if g.options.Release {
		gcbSubs["KUBERNETES_GCS_BUCKET"] = fmt.Sprintf("%s/stage/%s/%s/gcs-stage/%s", gcsBucket, buildVersion, versions.Prime(), versions.Prime())
	}

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

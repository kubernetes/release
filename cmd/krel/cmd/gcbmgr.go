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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/gcp"
	"k8s.io/release/pkg/gcp/auth"
	"k8s.io/release/pkg/gcp/build"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/util"
)

type GcbmgrOptions struct {
	Stage        bool
	Release      bool
	Stream       bool
	BuildAtHead  bool
	NoAnago      bool
	Branch       string
	ReleaseType  string
	BuildVersion string
	GcpUser      string
	LastJobs     int64
	Repo         Repository
	Version      Version
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Repository
type Repository interface {
	Open() error
	CheckState(string, string, string) error
	GetTag() (string, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Version
type Version interface {
	GetKubeVersionForBranch(release.VersionType, string) (string, error)
}

var (
	gcbmgrOpts = &GcbmgrOptions{}
	buildOpts  = &build.Options{}
)

// gcbmgrCmd is a krel subcommand which invokes RunGcbmgr()
var gcbmgrCmd = &cobra.Command{
	Use:           "gcbmgr",
	Short:         "Run gcbmgr",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunGcbmgr(gcbmgrOpts)
	},
}

func init() {
	// Submit types
	gcbmgrCmd.PersistentFlags().BoolVar(
		&gcbmgrOpts.Stage,
		"stage",
		false,
		"submit a stage run to GCB",
	)
	gcbmgrCmd.PersistentFlags().BoolVar(
		&gcbmgrOpts.Release,
		"release",
		false,
		"submit a release run to GCB",
	)

	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.Branch,
		"branch",
		git.DefaultBranch,
		"branch to run the specified GCB run against",
	)

	// Release types
	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.ReleaseType,
		"type",
		release.ReleaseTypeAlpha,
		fmt.Sprintf("release type, must be one of: '%s'",
			strings.Join([]string{
				release.ReleaseTypeAlpha,
				release.ReleaseTypeBeta,
				release.ReleaseTypeRC,
				release.ReleaseTypeOfficial,
			}, "', '"),
		),
	)

	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.BuildVersion,
		"build-version",
		"",
		fmt.Sprintf("the build version to be used. "+
			"Can be empty for `stage` releases, where it gets automatically "+
			"inferred by %q and the provided target branch.",
			release.VersionTypeCILatest,
		),
	)
	gcbmgrCmd.PersistentFlags().BoolVar(
		&gcbmgrOpts.NoAnago,
		"no-anago",
		false,
		"do not use anago in favor of the native golang implementation",
	)

	gcbmgrCmd.PersistentFlags().BoolVar(
		&gcbmgrOpts.BuildAtHead,
		"build-at-head",
		false,
		"the build version to be used when was to use the lastest in the branch.",
	)

	// gcloud options
	gcbmgrCmd.PersistentFlags().StringVar(
		&buildOpts.Project,
		"project",
		release.DefaultKubernetesStagingProject,
		"GCP project to run GCB in",
	)
	gcbmgrCmd.PersistentFlags().BoolVar(
		&gcbmgrOpts.Stream,
		"stream",
		false,
		"if specified, GCB will run synchronously, tailing its logs to stdout",
	)
	gcbmgrCmd.PersistentFlags().StringVar(
		&buildOpts.CloudbuildFile,
		"gcb-config",
		build.DefaultCloudbuildFile,
		"if specified, this will be used as the name of the Google Cloud Build config file",
	)
	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.GcpUser,
		"gcp-user",
		"",
		"if specified, this will be used as the GCP_USER_TAG",
	)

	gcbmgrCmd.PersistentFlags().Int64Var(
		&gcbmgrOpts.LastJobs,
		"list-jobs",
		5,
		"list the last N build jobs in the project",
	)

	gcbmgrOpts.Repo = release.NewRepo()
	gcbmgrOpts.Version = release.NewVersion()
	rootCmd.AddCommand(gcbmgrCmd)
}

// RunGcbmgr is the function invoked by 'krel gcbmgr', responsible for
// submitting release jobs to GCB
func RunGcbmgr(opts *GcbmgrOptions) error {
	if err := opts.Validate(); err != nil {
		return errors.Wrap(err, "validating gcbmgr options")
	}

	toolOrg := release.GetToolOrg()
	toolRepo := release.GetToolRepo()
	toolBranch := release.GetToolBranch()

	if err := gcp.PreCheck(); err != nil {
		return errors.Wrap(err, "pre-checking for GCP package usage")
	}

	if err := opts.Repo.Open(); err != nil {
		return errors.Wrap(err, "open release repo")
	}

	if err := opts.Repo.CheckState(toolOrg, toolRepo, toolBranch); err != nil {
		return errors.Wrap(err, "verifying repository state")
	}

	logrus.Infof("Running gcbmgr with the following options: %+v", opts)
	logrus.Infof("Build options: %v", *buildOpts)

	buildOpts.NoSource = true
	buildOpts.DiskSize = release.DefaultDiskSize

	buildOpts.Async = true

	if opts.Stream {
		buildOpts.Async = false
	}

	gcbSubs, gcbSubsErr := SetGCBSubstitutions(opts, toolOrg, toolRepo, toolBranch)
	if gcbSubs == nil || gcbSubsErr != nil {
		return gcbSubsErr
	}

	if rootOpts.nomock {
		// TODO: Consider a '--yes' flag so we can mock this
		_, nomockSubmit, askErr := util.Ask(
			fmt.Sprintf("Really submit a --nomock release job against the %s branch? (yes/no)", opts.Branch),
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
	case opts.Stage:
		jobType = "stage"
	case opts.Release:
		jobType = "release"
	default:
		return listJobs(buildOpts.Project, opts.LastJobs)
	}

	// Use dedicated job types for krel-based executions
	if opts.NoAnago {
		delete(gcbSubs, "BUILD_AT_HEAD")
		gcbSubs["LOG_LEVEL"] = rootOpts.logLevel
		jobType += "-krel"
	}

	buildOpts.ConfigDir = filepath.Join(toolRoot, "gcb", jobType)
	prepareBuildErr := build.PrepareBuilds(buildOpts)
	if prepareBuildErr != nil {
		return prepareBuildErr
	}

	// TODO: Need actual values
	var jobName, uploaded string

	version, err := opts.Repo.GetTag()
	if err != nil {
		return errors.Wrap(err, "getting current tag")
	}

	return build.RunSingleJob(buildOpts, jobName, uploaded, version, gcbSubs)
}

// SetGCBSubstitutions takes a set of GcbmgrOptions and returns a map of GCB substitutions
func SetGCBSubstitutions(o *GcbmgrOptions, toolOrg, toolRepo, toolBranch string) (map[string]string, error) {
	gcbSubs := map[string]string{}

	gcbSubs["TOOL_ORG"] = toolOrg
	gcbSubs["TOOL_REPO"] = toolRepo
	gcbSubs["TOOL_BRANCH"] = toolBranch

	gcpUser := o.GcpUser
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

	gcbSubs["TYPE"] = o.ReleaseType
	gcbSubs["TYPE_TAG"] = o.ReleaseType

	gcbSubs["RELEASE_BRANCH"] = o.Branch

	buildVersion := o.BuildVersion
	if o.Release && buildVersion == "" {
		return gcbSubs, errors.New("Build version must be specified when sending a release GCB run")
	}

	if o.Stage && o.BuildAtHead {
		hash, err := git.LSRemoteExec(git.GetDefaultKubernetesRepoURL(), "rev-parse", o.Branch)
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
		buildVersion, versionErr = o.Version.GetKubeVersionForBranch(
			release.VersionTypeCILatest, o.Branch,
		)
		if versionErr != nil {
			return gcbSubs, versionErr
		}

		if o.Stage {
			gcbSubs["BUILD_AT_HEAD"] = ""
		}
	}

	buildpoint := buildVersion
	buildpoint = strings.ReplaceAll(buildpoint, "+", "-")
	gcbSubs["BUILD_POINT"] = buildpoint
	gcbSubs["BUILDVERSION"] = buildVersion

	kubecrossBranches := []string{
		o.Branch,
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
	if o.ReleaseType != release.ReleaseTypeOfficial && len(v.Pre) > 0 {
		// if the release we will build is the same in the current build point then we increment
		// otherwise we are building the next type so set to 0
		if v.Pre[0].String() == o.ReleaseType {
			patch = fmt.Sprintf("%d-%s.%d", v.Patch, o.ReleaseType, v.Pre[1].VersionNum+1)
		} else if o.ReleaseType == release.ReleaseTypeRC && v.Pre[0].String() != release.ReleaseTypeRC {
			// Now if is RC we are building and is the first time we set to 1 since the 0 is bypassed
			patch = fmt.Sprintf("%d-%s.1", v.Patch, o.ReleaseType)
		} else {
			patch = fmt.Sprintf("%d-%s.0", v.Patch, o.ReleaseType)
		}
	}
	gcbSubs["PATCH_VERSION_TAG"] = patch
	gcbSubs["KUBERNETES_VERSION_TAG"] = fmt.Sprintf("%d.%d.%s", v.Major, v.Minor, patch)

	return gcbSubs, nil
}

var BuildListJobs = build.ListJobs

// listJobs lists recent GCB jobs run in the specified project
func listJobs(project string, lastJobs int64) error {
	if lastJobs < 0 {
		logrus.Infof("--list-jobs was set to a negative number, defaulting to 5")
		lastJobs = 5
	}

	logrus.Infof("Listing last %d GCB jobs:", lastJobs)
	return BuildListJobs(project, lastJobs)
}

func (o *GcbmgrOptions) Validate() error {
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

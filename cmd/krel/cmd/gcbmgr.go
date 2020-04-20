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
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/command"
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

	requiredPackages = []string{
		// "bsdmainutils",
	}

	// TODO: Do we really need this if we use the Google Cloud SDK instead?
	requiredCommands = []string{
		"gcloud",
		"git",
		"gsutil",
		"jq",
	}
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
		"Submit a stage run to GCB",
	)
	gcbmgrCmd.PersistentFlags().BoolVar(
		&gcbmgrOpts.Release,
		"release",
		false,
		"Submit a release run to GCB",
	)

	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.Branch,
		"branch",
		git.Master,
		"Branch to run the specified GCB run against",
	)

	// Release types
	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.ReleaseType,
		"type",
		"prerelease",
		"Release type (must be one of: 'prerelease', 'rc', 'official')",
	)

	// TODO: Remove default once find_green_build logic exists
	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.BuildVersion,
		"build-version",
		"",
		"Build version",
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
		"If specified, GCB will run synchronously, tailing its' logs to stdout",
	)
	gcbmgrCmd.PersistentFlags().StringVar(
		&buildOpts.CloudbuildFile,
		"gcb-config",
		build.DefaultCloudbuildFile,
		"If provided, this will be used as the name of the Google Cloud Build config file.",
	)
	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.GcpUser,
		"gcp-user",
		"",
		"If provided, this will be used as the GCP_USER_TAG.",
	)

	gcbmgrCmd.PersistentFlags().Int64Var(
		&gcbmgrOpts.LastJobs,
		"list-jobs",
		5,
		"List the last x build jobs in the project. Default to 5.",
	)

	gcbmgrOpts.Repo = release.NewRepo()
	gcbmgrOpts.Version = release.NewVersion()
	rootCmd.AddCommand(gcbmgrCmd)
}

// RunGcbmgr is the function invoked by 'krel gcbmgr', responsible for
// submitting release jobs to GCB
func RunGcbmgr(opts *GcbmgrOptions) error {
	logrus.Info("Checking for required packages")
	ok, err := util.PackagesAvailable(requiredPackages...)
	if err != nil {
		return errors.Wrap(err, "unable to verify if packages are available")
	}
	if !ok {
		return errors.New("packages required to run gcbmgr are not present")
	}

	logrus.Info("Checking for required commands")
	if cmdAvailable := command.Available(requiredCommands...); !cmdAvailable {
		return errors.New("binaries required to run gcbmgr are not present")
	}

	toolOrg := release.GetToolOrg()
	toolRepo := release.GetToolRepo()
	toolBranch := release.GetToolBranch()

	if err := opts.Repo.Open(); err != nil {
		return errors.Wrap(err, "open release repo")
	}

	if err := opts.Repo.CheckState(toolOrg, toolRepo, toolBranch); err != nil {
		return errors.Wrap(err, "verifying repository state")
	}

	logrus.Infof("Running gcbmgr with the following options: %+v", opts)
	logrus.Infof("Build options: %v", *buildOpts)

	if opts.Stage && opts.Release {
		return errors.New("cannot specify both the 'stage' and 'release' flag; resubmit with only one build type selected")
	}

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
			fmt.Sprintf("Really submit a --nomock release job against the %s branch?", opts.Branch),
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
		buildOpts.DiskSize = "100"
	default:
		return listJobs(buildOpts.Project, opts.LastJobs)
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

	// TODO: The naming for these env vars is clumsy/confusing, but we're bound by anago right now.
	releaseType := o.ReleaseType
	switch releaseType {
	case "official":
		gcbSubs["OFFICIAL_TAG"] = releaseType
		gcbSubs["OFFICIAL"] = fmt.Sprintf("--%s", releaseType)

		// TODO: Remove once cloudbuild.yaml doesn't strictly require vars to be set.
		gcbSubs["RC_TAG"] = ""
		gcbSubs["RC"] = ""
	case "rc":
		gcbSubs["RC_TAG"] = releaseType
		gcbSubs["RC"] = fmt.Sprintf("--%s", releaseType)

		// TODO: Remove once cloudbuild.yaml doesn't strictly require vars to be set.
		gcbSubs["OFFICIAL_TAG"] = ""
		gcbSubs["OFFICIAL"] = ""
	case "prerelease":
		// TODO: Remove once cloudbuild.yaml doesn't strictly require vars to be set.
		gcbSubs["OFFICIAL_TAG"] = ""
		gcbSubs["OFFICIAL"] = ""
		gcbSubs["RC_TAG"] = ""
		gcbSubs["RC"] = ""
	}

	gcbSubs["RELEASE_BRANCH"] = o.Branch

	if o.Stage {
		// TODO: Remove once we remove support for --built-at-head.
		gcbSubs["BUILD_AT_HEAD"] = ""
	}

	buildVersion := o.BuildVersion
	if buildVersion == "" {
		if o.Release {
			return gcbSubs, errors.New("Build version must be specified when sending a release GCB run")
		}

		var versionErr error
		buildVersion, versionErr = o.Version.GetKubeVersionForBranch(
			release.VersionTypeCILatest, o.Branch,
		)
		if versionErr != nil {
			return gcbSubs, versionErr
		}
	}

	buildpoint := buildVersion
	buildpoint = strings.ReplaceAll(buildpoint, "+", "-")
	gcbSubs["BUILD_POINT"] = buildpoint

	// TODO: Add conditionals for find_green_build
	buildVersion = fmt.Sprintf("--buildversion=%s", buildVersion)
	gcbSubs["BUILDVERSION"] = buildVersion

	kubecrossBranches := []string{
		o.Branch,
		git.Master,
	}

	kubecrossVersion, kubecrossVersionErr := release.GetKubecrossVersion(kubecrossBranches...)
	if kubecrossVersionErr != nil {
		return gcbSubs, kubecrossVersionErr
	}
	gcbSubs["KUBE_CROSS_VERSION"] = kubecrossVersion

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

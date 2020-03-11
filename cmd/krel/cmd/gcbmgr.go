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

	"k8s.io/release/pkg/gcp/auth"
	"k8s.io/release/pkg/gcp/build"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/util"
)

type gcbmgrOptions struct {
	stage        bool
	release      bool
	stream       bool
	branch       string
	releaseType  string
	buildVersion string
	gcpUser      string
}

var (
	gcbmgrOpts = &gcbmgrOptions{}
	buildOpts  = &build.Options{}

	// TODO: Commenting these packages/commands out since they fail in CI.
	//       These can be fixed by changing the CI test image to one that includes the packages.
	//nolint:gocritic
	/*
		requiredPackages = []string{
			"jq",
			"git",
			"bsdmainutils",
		}

		// TODO: Do we really need this if we use the Google Cloud SDK instead?
		requiredCommands = []string{
			"gsutil",
			"gcloud",
		}
	*/
)

// gcbmgrCmd is a krel subcommand which invokes runGcbmgr()
var gcbmgrCmd = &cobra.Command{
	Use:           "gcbmgr",
	Short:         "Run gcbmgr",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGcbmgr()
	},
}

func init() {
	// Submit types
	gcbmgrCmd.PersistentFlags().BoolVar(
		&gcbmgrOpts.stage,
		"stage",
		false,
		"Submit a stage run to GCB",
	)
	gcbmgrCmd.PersistentFlags().BoolVar(
		&gcbmgrOpts.release,
		"release",
		false,
		"Submit a release run to GCB",
	)

	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.branch,
		"branch",
		"",
		"Branch to run the specified GCB run against",
	)

	// Release types
	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.releaseType,
		"type",
		"prerelease",
		"Release type (must be one of: 'prerelease', 'rc', 'official')",
	)

	// TODO: Remove default once find_green_build logic exists
	gcbmgrCmd.PersistentFlags().StringVar(
		&gcbmgrOpts.buildVersion,
		"build-version",
		"",
		"Build version",
	)

	// gcloud options
	gcbmgrCmd.PersistentFlags().StringVar(
		&buildOpts.Project,
		"project",
		release.DefaultProject,
		"Branch to run the specified GCB run against",
	)
	gcbmgrCmd.PersistentFlags().BoolVar(
		&gcbmgrOpts.stream,
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
		&gcbmgrOpts.gcpUser,
		"gcp-user",
		"",
		"If provided, this will be used as the GCP_USER_TAG.",
	)

	rootCmd.AddCommand(gcbmgrCmd)
}

// runGcbmgr is the function invoked by 'krel gcbmgr', responsible for submitting release jobs to GCB
func runGcbmgr() error {
	// TODO: Commenting these checks out since they fail in CI.
	//       These can be fixed by changing the CI test image to one that includes the packages.
	//nolint:gocritic
	/*
		logrus.Info("Checking for required packages...")
		pkgAvailable, pkgAvailableErr := util.PackagesAvailable(requiredPackages...)
		if pkgAvailableErr != nil {
			return pkgAvailableErr
		}
		if !pkgAvailable {
			return errors.New("packages required to run gcbmgr are not present; cannot continue")
		}

		logrus.Info("Checking for required commands...")
		if cmdAvailable := command.Available(requiredCommands...); !cmdAvailable {
			return errors.New("binaries required to run gcbmgr are not present; cannot continue")
		}
	*/

	// TODO: Add gitlib::repo_state check

	logrus.Infof("Running gcbmgr with the following options: %v", *gcbmgrOpts)
	logrus.Infof("Build options: %v", *buildOpts)

	if gcbmgrOpts.stage && gcbmgrOpts.release {
		return errors.New("cannot specify both the 'stage' and 'release' flag; resubmit with only one build type selected")
	}

	buildOpts.NoSource = true
	buildOpts.DiskSize = release.DefaultDiskSize

	buildOpts.Async = true

	if gcbmgrOpts.stream {
		buildOpts.Async = false
	}

	gcbSubs, gcbSubsErr := setGCBSubstitutions(gcbmgrOpts)
	if gcbSubs == nil || gcbSubsErr != nil {
		return gcbSubsErr
	}

	if rootOpts.nomock {
		// TODO: Consider a '--yes' flag so we can mock this
		_, nomockSubmit, askErr := util.Ask(
			"Really submit a --nomock release job against the $RELEASE_BRANCH branch?",
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
	case gcbmgrOpts.stage:
		jobType = "stage"
	case gcbmgrOpts.release:
		jobType = "release"
		buildOpts.DiskSize = "100"
	default:
		return listJobs()
	}

	buildOpts.ConfigDir = filepath.Join(toolRoot, "gcb", jobType)
	prepareBuildErr := build.PrepareBuilds(buildOpts)
	if prepareBuildErr != nil {
		return prepareBuildErr
	}

	// TODO: Need actual values
	var jobName, uploaded string

	version, err := git.GetTag()
	if err != nil {
		return err
	}

	return build.RunSingleJob(buildOpts, jobName, uploaded, version, gcbSubs)
}

// setGCBSubstitutions takes a set of gcbmgrOptions and returns a map of GCB substitutions
func setGCBSubstitutions(o *gcbmgrOptions) (map[string]string, error) {
	gcbSubs := map[string]string{}

	toolOrg := release.GetToolOrg()
	gcbSubs["TOOL_ORG"] = toolOrg

	toolRepo := release.GetToolRepo()
	gcbSubs["TOOL_REPO"] = toolRepo

	toolBranch := release.GetToolBranch()
	gcbSubs["TOOL_BRANCH"] = toolBranch

	gcpUser := o.gcpUser
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
	releaseType := o.releaseType
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

	branch := o.branch
	if branch == "" {
		return gcbSubs, errors.New("Release branch must be set to continue")
	}

	gcbSubs["RELEASE_BRANCH"] = branch

	if o.stage {
		// TODO: Remove once we remove support for --built-at-head.
		gcbSubs["BUILD_AT_HEAD"] = ""
	}

	buildVersion := o.buildVersion
	if buildVersion == "" {
		if o.release {
			return gcbSubs, errors.New("Build version must be specified when sending a release GCB run")
		}

		var versionErr error
		buildVersion, versionErr = release.GetKubeVersionForBranch(
			release.VersionTypeCILatest, o.branch,
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
		branch,
		"master",
	}

	kubecrossVersion, kubecrossVersionErr := release.GetKubecrossVersion(kubecrossBranches...)
	if kubecrossVersionErr != nil {
		return gcbSubs, kubecrossVersionErr
	}
	gcbSubs["KUBE_CROSS_VERSION"] = kubecrossVersion

	return gcbSubs, nil
}

// listJobs lists recent GCB jobs run in the specified project
func listJobs() error {
	logrus.Info("Listing GCB jobs is not currently supported.")

	// TODO: Add job listing logic
	// logrus.Info("Listing recent GCB jobs...")
	return nil
}

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
}

var (
	gcbmgrOpts = &gcbmgrOptions{}
	buildOpts  = &build.Options{}

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
)

const (
	// TODO: This should maybe be in pkg/release
	defaultReleaseToolRepo   = "https://github.com/kubernetes/release"
	defaultReleaseToolBranch = "master"
	defaultProject           = "kubernetes-release-test"
	defaultDiskSize          = "300"

	bucketPrefix = "kubernetes-release-"
)

// gcbmgrCmd is the command when calling `krel version`
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
		"FAKE+BUILD+POINT",
		"Build version",
	)

	// gcloud options
	gcbmgrCmd.PersistentFlags().StringVar(
		&buildOpts.Project,
		"project",
		defaultProject,
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

	rootCmd.AddCommand(gcbmgrCmd)
}

func runGcbmgr() error {
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

	// TODO: Add gitlib::repo_state check

	logrus.Infof("Running gcbmgr with the following options: %v", *gcbmgrOpts)
	logrus.Infof("Build options: %v", *buildOpts)

	buildOpts.NoSource = true
	buildOpts.DiskSize = defaultDiskSize

	if gcbmgrOpts.stream {
		buildOpts.Async = false
	} else {
		buildOpts.Async = true
	}

	if gcbmgrOpts.stage && gcbmgrOpts.release {
		logrus.Fatal("Cannot specify both the 'stage' and 'release' flag. Please resubmit with only one of those flags selected.")
	}

	gcbSubs, gcbSubsErr := setGCBSubstitutions()
	if gcbSubs == nil || gcbSubsErr != nil {
		return gcbSubsErr
	}

	if rootOpts.nomock {
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
	case gcbmgrOpts.stage && gcbmgrOpts.release:
		return errors.New("The '--stage' and '--release' flags cannot be used together")
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
	version := "FAKEVERSION"

	return build.RunSingleJob(buildOpts, jobName, uploaded, version, gcbSubs)
}

func setGCBSubstitutions() (map[string]string, error) {
	gcbSubs := map[string]string{}

	releaseToolRepo := os.Getenv("RELEASE_TOOL_REPO")
	if releaseToolRepo == "" {
		releaseToolRepo = defaultReleaseToolRepo
	}

	releaseToolBranch := os.Getenv("RELEASE_TOOL_BRANCH")
	if releaseToolBranch == "" {
		releaseToolBranch = defaultReleaseToolBranch
	}

	gcbSubs["RELEASE_TOOL_REPO"] = releaseToolRepo
	gcbSubs["RELEASE_TOOL_BRANCH"] = releaseToolBranch

	gcpUser, gcpUserErr := auth.GetCurrentGCPUser()
	if gcpUserErr != nil {
		return gcbSubs, gcpUserErr
	}

	gcbSubs["GCP_USER_TAG"] = gcpUser

	// TODO: The naming for these env vars is clumsy/confusing, but we're bound by anago right now.
	releaseType := gcbmgrOpts.releaseType
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

	// TODO: Remove once we remove support for --built-at-head.
	gcbSubs["BUILD_AT_HEAD"] = ""

	buildpoint := gcbmgrOpts.buildVersion
	buildpoint = strings.ReplaceAll(buildpoint, "+", "-")
	gcbSubs["BUILD_POINT"] = buildpoint

	// TODO: Add conditionals for find_green_build
	buildVersion := gcbmgrOpts.buildVersion
	buildVersion = fmt.Sprintf("--buildversion=%s", buildVersion)
	gcbSubs["BUILDVERSION"] = buildVersion

	branch := gcbmgrOpts.branch

	if branch != "" {
		gcbSubs["RELEASE_BRANCH"] = branch
	} else {
		return gcbSubs, errors.New("Release branch must be set to continue")
	}

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

func listJobs() error {
	logrus.Info("Listing GCB jobs is not currently supported.")

	// TODO: Add job listing logic
	// logrus.Info("Listing recent GCB jobs...")
	return nil
}

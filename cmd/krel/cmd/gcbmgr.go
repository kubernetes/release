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
)

const (
	// TODO: This should maybe be in pkg/release
	defaultReleaseToolRepo   = "https://github.com/kubernetes/release"
	defaultReleaseToolBranch = "master"
	defaultProject           = "kubernetes-release-test"
	//nolint
	defaultDiskSize = "300"

	bucketPrefix = "kubernetes-release-"
)

// gcbmgrCmd is the command when calling `krel version`
var gcbmgrCmd = &cobra.Command{
	Use:           "gcbmgr",
	Short:         "Run gcbmgr",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE:       initLogging,
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

	rootCmd.AddCommand(gcbmgrCmd)
}

func runGcbmgr() error {
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
	if gcbSubsErr != nil {
		return gcbSubsErr
	}

	if rootOpts.nomock {
		_, nomockSubmit, askErr := util.Ask(
			"Really submit a --nomock release job against the $RELEASE_BRANCH branch",
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

	switch {
	case gcbmgrOpts.stage:
		return submitStage(toolRoot, gcbSubs)
	case gcbmgrOpts.release:
		return submitRelease()
	default:
		return listJobs()
	}
}

func submitStage(toolRoot string, substitutions map[string]string) error {
	logrus.Infof("Submitting a stage to GCB")

	buildOpts.CloudbuildFile = filepath.Join(toolRoot, "gcb/stage/cloudbuild.yaml")

	// TODO: Need actual values
	var jobName, uploaded string
	version := "FAKEVERSION"

	return build.RunSingleJob(buildOpts, jobName, uploaded, version, substitutions)
}

// TODO: Populate logic once we're happy with the flow for the submitStage() function.
func submitRelease() error {
	logrus.Infof("Submitting a release to GCB")

	buildOpts.DiskSize = "100"

	//nolint:gocritic
	return nil // build.RunSingleJob(buildOpts, jobName, uploaded, version, subs)
}

func setGCBSubstitutions() (map[string]string, error) {
	gcbSubs := make(map[string]string)

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

	// TODO: Need to find out if command.Execute supports capturing the command output
	gcpUser := "FAKEUSER"
	//nolint
	/*
		gcpUser := command.Execute(
			"gcloud",
			"auth",
			"list",
			"--filter=status:ACTIVE",
			`--format="value(account)"`,
		)
		if gcpUser != nil {
			return nil, gcpUserErr
		}
	*/

	gcpUser = strings.ReplaceAll(gcpUser, "@", "-at-")
	gcpUser = strings.ReplaceAll(gcpUser, ".", "-")
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

	if gcbmgrOpts.branch != "" {
		gcbSubs["RELEASE_BRANCH"] = gcbmgrOpts.branch
	} else {
		return nil, errors.New("Release branch must be set to continue")
	}

	// TODO: Ensure release.GetKubecrossVersion() isn't hardcoded.
	gcbSubs["KUBE_CROSS_VERSION"] = release.GetKubecrossVersion()

	return gcbSubs, nil
}

func listJobs() error {
	logrus.Info("Listing GCB jobs is not currently supported.")

	// TODO: Add job listing logic
	// logrus.Info("Listing recent GCB jobs...")
	return nil
}

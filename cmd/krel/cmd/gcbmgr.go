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
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/release/pkg/gcp/build"
	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

var gcbmgrOpts = &gcb.Options{}

// gcbmgrCmd is a krel subcommand which invokes RunGcbmgr()
var gcbmgrCmd = &cobra.Command{
	Use:           "gcbmgr",
	Short:         "Run gcbmgr",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		gcbmgrOpts.NoMock = rootOpts.nomock
		gcbmgrOpts.LogLevel = rootOpts.logLevel
		return gcb.New(gcbmgrOpts).Submit()
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
		&gcbmgrOpts.Project,
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
		&gcbmgrOpts.CloudbuildFile,
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

	rootCmd.AddCommand(gcbmgrCmd)
}

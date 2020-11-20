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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/anago"
	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/release"
)

// stageCmd represents the subcommand for `krel stage`
var stageCmd = &cobra.Command{
	Use:   "stage",
	Short: "Stage a new Kubernetes version",
	Long: fmt.Sprintf(`krel stage

This subcommand is the first of two necessary steps for cutting Kubernetes
releases. It is intended to be run by users who want to submit a Google Cloud
Build (GCB) job which does:

1. Check Prerequisites: Verify that a valid %s environment variable is set. It
   also checks for the existence and version of required packages and if
   the correct Google Cloud project is set. A basic hardware check will ensure
   that enough disk space is available, too.

2. Set Build Candidate: Discovers the release branch, parent branch (if
   available) and build version for this release.

3. Prepare Workspace: Verifies that the working directory is in the desired
   state. This means that the build directory "%s" is cleaned up and the
   checked out repository is in a clean state.

4. Build: Run 'make cross-in-a-container' by using the latest kubecross
   container image. This step also build all necessary release tarballs.

5. Generate release notes: Generate the CHANGELOG-x.y.md file and commit it
   into the local working repository.

6. Stage: Copies the build artifacts to a Google Cloud Bucket.
`, github.TokenEnvKey, release.BuildDir),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStage(stageOptions)
	},
}

var (
	stageOptions = anago.DefaultStageOptions()
	submitJob    = true
	stream       = false
)

const (
	buildVersionFlag = "build-version"
	submitJobFlag    = "submit"
	streamFlag       = "stream"
)

func init() {
	stageCmd.PersistentFlags().
		StringVar(
			&stageOptions.ReleaseType,
			"type",
			stageOptions.ReleaseType,
			fmt.Sprintf("The release type, must be one of: '%s'",
				strings.Join([]string{
					release.ReleaseTypeAlpha,
					release.ReleaseTypeBeta,
					release.ReleaseTypeRC,
					release.ReleaseTypeOfficial,
				}, "', '"),
			))

	stageCmd.PersistentFlags().
		StringVar(
			&stageOptions.ReleaseBranch,
			"branch",
			stageOptions.ReleaseBranch,
			"The release branch for which the release should be build",
		)

	stageCmd.PersistentFlags().
		StringVar(
			&stageOptions.BuildVersion,
			buildVersionFlag,
			"",
			"The build version to be released.",
		)

	stageCmd.PersistentFlags().
		BoolVar(
			&submitJob,
			submitJobFlag,
			true,
			"Submit a Google Cloud Build job",
		)

	stageCmd.PersistentFlags().
		BoolVar(
			&stream,
			streamFlag,
			false,
			"Run the Google Cloud Build job synchronously",
		)

	for _, flag := range []string{buildVersionFlag, submitJobFlag} {
		if err := stageCmd.PersistentFlags().MarkHidden(flag); err != nil {
			logrus.Fatal(err)
		}
	}

	rootCmd.AddCommand(stageCmd)
}

func runStage(options *anago.StageOptions) error {
	options.NoMock = rootOpts.nomock
	stage := anago.NewStage(options)
	if submitJob {
		return stage.Submit(stream)
	}
	return stage.Run()
}

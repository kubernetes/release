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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/anago"
	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/release"
)

// releaseCmd represents the subcommand for `krel release`
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release a staged Kubernetes version",
	Long: fmt.Sprintf(`krel release

This subcommand is the second of two necessary steps for cutting Kubernetes
releases. It is intended to be run by users who want to submit a Google Cloud
Build (GCB) job, which does:

1. Check Prerequisites: Verify that a valid %s environment variable is set. It
   also checks for the existence and version of required packages and if
   the correct Google Cloud project is set. A basic hardware check will ensure
   that enough disk space is available, too.

2. Set Build Candidate: Discovers the release branch, parent branch (if
   available) and build version for this release.

3. Prepare Workspace: Verifies that the working directory is in the desired
   state. This means that the staged sources will be downloaded from the bucket
   which should contain a copy of the repository.

4. Push Artifacts: Pushes the generated artifacts to the release bucket and
   Google Container Registry.

5. Push Git Objects: Pushes the new tags and branches to the repository remote
   on GitHub.

6. Announce: Create the release announcement mail and update the GitHub release
   page to contain the artifacts and their checksums.

7. Archive: Copies the release process logs to a bucket and sets private
   permissions on it.
`, github.TokenEnvKey),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRelease(releaseOptions)
	},
}

var releaseOptions = anago.DefaultReleaseOptions()

func init() {
	releaseCmd.PersistentFlags().
		StringVar(
			&releaseOptions.ReleaseType,
			"type",
			releaseOptions.ReleaseType,
			fmt.Sprintf("The release type, must be one of: '%s'",
				strings.Join([]string{
					release.ReleaseTypeAlpha,
					release.ReleaseTypeBeta,
					release.ReleaseTypeRC,
					release.ReleaseTypeOfficial,
				}, "', '"),
			))

	releaseCmd.PersistentFlags().
		StringVar(
			&releaseOptions.ReleaseBranch,
			"branch",
			releaseOptions.ReleaseBranch,
			"The release branch for which the release should be build",
		)

	releaseCmd.PersistentFlags().
		StringVar(
			&releaseOptions.BuildVersion,
			buildVersionFlag,
			"",
			"The build version to be released.",
		)

	releaseCmd.PersistentFlags().
		BoolVar(
			&submitJob,
			submitJobFlag,
			true,
			"Submit a Google Cloud Build job",
		)

	releaseCmd.PersistentFlags().
		BoolVar(
			&stream,
			streamFlag,
			false,
			"Run the Google Cloud Build job synchronously",
		)

	if err := releaseCmd.PersistentFlags().MarkHidden(submitJobFlag); err != nil {
		logrus.Fatal(err)
	}

	if err := releaseCmd.MarkPersistentFlagRequired(buildVersionFlag); err != nil {
		logrus.Fatal(err)
	}

	rootCmd.AddCommand(releaseCmd)
}

func runRelease(options *anago.ReleaseOptions) error {
	options.NoMock = rootOpts.nomock
	rel := anago.NewRelease(options)

	if submitJob {
		// Perform a local check of the specified options
		// before launching a Cloud Build job:
		if err := options.Validate(&anago.State{}); err != nil {
			return errors.Wrap(err, "prechecking release options")
		}
		return rel.Submit(stream)
	}
	return rel.Run()
}

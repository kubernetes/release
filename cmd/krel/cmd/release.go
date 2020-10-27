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

	"github.com/spf13/cobra"

	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/release"
)

// releaseCmd represents the subcommand for `krel release`
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release a staged Kubernetes version",
	Long: fmt.Sprintf(`krel release

This subcommand is the second of two necessary steps for cutting Kubernetes
releases. It is intended to be run only in Google Cloud Build (GCB) and after a
successful 'krel stage'. The following steps are involved in the process:

1. Check Prerequisites: Verify that a valid %s environment variable is set. It
   also checks for the existence and version of required packages and if
   the correct Google Cloud project is set. A basic hardware check will ensure
   that enough disk space is available, too.

2. Set Build Candidate: Discovers the release branch, parent branch (if
   available) and build version for this release.

3. Prepare Workspace: Verifies that the working directory is in the desired
   state. This means that the build directory "%s" is cleaned up and the
   checked out repository is in a clean state.

4. Push Artifacts: Pushes the generated artifacts to the release bucket and
   Google Container Registry.

5. Push Git Objects: Pushes the new tags and branches to the repository remote
   on GitHub.

6. Announce: Create the release announcement mail and update the GitHub release
   page to contain the artifacts and their checksums.

7. Archive: Copies the release process logs to a bucket and sets private
   permissions on it.

`, github.TokenEnvKey, release.BuildDir),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRelease(releaseOpts, rootOpts)
	},
}

type releaseOptions struct{}

var releaseOpts = &releaseOptions{}

func init() {
	rootCmd.AddCommand(releaseCmd)
}

func runRelease(opts *releaseOptions, rootOpts *rootOptions) error {
	return nil
}

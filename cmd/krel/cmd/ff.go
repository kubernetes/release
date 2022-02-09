/*
Copyright 2019 The Kubernetes Authors.

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

	"github.com/spf13/cobra"

	"k8s.io/release/pkg/fastforward"
	"k8s.io/release/pkg/release"
	kgit "sigs.k8s.io/release-sdk/git"
)

var ffOpts = &fastforward.Options{}

// ffCmd represents the base command when called without any subcommands
var ffCmd = &cobra.Command{
	Use:     "fast-forward --branch <release-branch> [--ref <main-ref>] [--nomock] [--cleanup]",
	Short:   "Fast forward a Kubernetes release branch",
	Aliases: []string{"ff"},
	Long: fmt.Sprintf(`fast-forward fast forwards a branch to a specified git object (defaults to %s).

krel fast-forward pre-checks that the provided branch to be forwarded is an
actual 'release-x.y' branch and that the branch exists remotely. If that is not
the case, krel fast-forward will fail.

If no branch is provided, then krel will try to find the latest upstream k/k
release branch. If this release branch already contains a final minor release,
then krel ff will do nothing at all.

After that preflight-check, the release branch will be checked out and krel
verifies that the latest merge base tag is the same for the main and the
release branch. This means that only the latest release branch can be fast
forwarded.

krel merges the provided ref into the release branch and asks for a final
confirmation if the push should really happen. The push will only be executed
as real push if the '--nomock' flag is specified.

If --non-interactive is set to true, then krel will not require any user
interaction.  This mode is mainly made for CI purposes.

If --submit is set to true, then krel fast-forward will run by submitting a new
Google Cloud Build job.
`, kgit.Remotify(kgit.DefaultBranch)),
	Example:       "krel fast-forward --branch release-1.17 --ref origin/master --cleanup",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ffOpts.NoMock = rootOpts.nomock
		return fastforward.New(ffOpts).Run()
	},
}

func init() {
	ffCmd.PersistentFlags().StringVar(&ffOpts.RepoPath, "repo", filepath.Join(os.TempDir(), "k8s"), "the local path to the repository to be used")
	ffCmd.PersistentFlags().StringVar(&ffOpts.Branch, "branch", "", "branch")
	ffCmd.PersistentFlags().StringVar(&ffOpts.MainRef, "ref", kgit.Remotify(kgit.DefaultBranch), "ref on the main branch")
	ffCmd.PersistentFlags().StringVar(&ffOpts.GCPProjectID, "project-id", release.DefaultRelengStagingTestProject, "Google Cloud Porject to use to submit the job")
	ffCmd.PersistentFlags().BoolVar(&ffOpts.Cleanup, "cleanup", false, "cleanup the repository after the run")
	ffCmd.PersistentFlags().BoolVar(&ffOpts.NonInteractive, "non-interactive", false, "do not require any user interaction")
	ffCmd.PersistentFlags().BoolVar(&ffOpts.Submit, "submit", false, "run inside of Google Cloud Build by submitting a new job")

	rootCmd.AddCommand(ffCmd)
}

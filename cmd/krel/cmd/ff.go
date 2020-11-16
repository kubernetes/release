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
	kgit "k8s.io/release/pkg/git"
)

var ffOpts = &fastforward.Options{}

// ffCmd represents the base command when called without any subcommands
var ffCmd = &cobra.Command{
	Use:   "ff --branch <release-branch> [--ref <main-ref>] [--nomock] [--cleanup]",
	Short: "Fast forward a Kubernetes release branch",
	Long: fmt.Sprintf(`ff fast forwards a branch to a specified git object (defaults to %s).

krel ff pre-checks that the local branch to be forwarded is an actual
'release-x.y' branch and that the branch exists remotely. If that is not the
case, krel ff will fail.

After that preflight-check, the release branch will be checked out and krel
verifies that the latest merge base tag is the same for the main and the
release branch. This means that only the latest release branch can be fast
forwarded.

krel merges the provided ref into the release branch and asks for a final
confirmation if the push should really happen. The push will only be executed
as real push if the '--nomock' flag is specified.
`, kgit.Remotify(kgit.DefaultBranch)),
	Example:       "krel ff --branch release-1.17 --ref origin/master --cleanup",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ffOpts.NoMock = rootOpts.nomock
		return fastforward.Run(ffOpts)
	},
}

func init() {
	ffCmd.PersistentFlags().StringVar(&ffOpts.RepoPath, "repo", filepath.Join(os.TempDir(), "k8s"), "the local path to the repository to be used")
	ffCmd.PersistentFlags().StringVar(&ffOpts.Branch, "branch", "", "branch")
	ffCmd.PersistentFlags().StringVar(&ffOpts.MainRef, "ref", kgit.Remotify(kgit.DefaultBranch), "ref on the main branch")
	ffCmd.PersistentFlags().BoolVar(&ffOpts.Cleanup, "cleanup", false, "cleanup the repository after the run")

	rootCmd.AddCommand(ffCmd)
}

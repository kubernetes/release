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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/changelog"
	"k8s.io/release/pkg/github"
)

// changelogCmd represents the subcommand for `krel changelog`
var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Automate the lifecycle of CHANGELOG-x.y.{md,html} files in a k/k repository",
	Long: fmt.Sprintf(`krel changelog

To let this tool work, please point '--repo' to a local copy of the target k/k
repository. This local checkout will be modified during the run of 'krel
changelog' and should contain all changes from the remote location. Beside this,
a valid %s=<TOKEN> environment variable has to be exported to let the generation
of the release notes work.

The 'changelog' subcommand of 'krel' does the following things by utilizing
the golang based 'release-notes' tool:

1. Generate the release notes for either a patch or a new minor release. Minor
   releases can be alpha, beta or rc’s, too.
   a) Create a new CHANGELOG-x.y.md file if not existing.
   b) Correctly prepend the generated notes to the existing CHANGELOG-x.y.md
      file if already existing. This also includes the modification of the
	  table of contents.

2. Convert the markdown release notes into a HTML equivalent on purpose of
   sending it by mail to the announce list. The HTML file will be dropped into
   the current working directly as 'CHANGELOG-x.y.html'. Sending the
   announcement is done by another subcommand of 'krel', not 'changelog'.

3. Commit the modified CHANGELOG-x.y.md into the main branch as well as the
   corresponding release-branch of kubernetes/kubernetes. The release branch
   will be pruned from all other CHANGELOG-*.md files which do not belong to
   this release branch.
`, github.TokenEnvKey),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return changelog.New(changelogOptions).Run()
	},
}

var changelogOptions = &changelog.Options{}

func init() {
	changelogCmd.PersistentFlags().StringVar(&changelogOptions.RepoPath, "repo", filepath.Join(os.TempDir(), "k8s"), "the local path to the repository to be used")
	changelogCmd.PersistentFlags().StringVar(&changelogOptions.Bucket, "bucket", "kubernetes-release", "Specify gs bucket to point to in generated notes")
	changelogCmd.PersistentFlags().StringVar(&changelogOptions.Tag, "tag", "", "The version tag of the release, for example v1.17.0-rc.1")
	changelogCmd.PersistentFlags().StringVar(&changelogOptions.Branch, "branch", "", "The branch to be used. Will be automatically inherited by the tag if not set.")
	changelogCmd.PersistentFlags().StringVar(&changelogOptions.Tars, "tars", ".", "Directory of tars to SHA512 sum for display")
	changelogCmd.PersistentFlags().StringVar(&changelogOptions.HTMLFile, "html-file", "", "The target html file to be written. If empty, then it will be CHANGELOG-x.y.html in the current path.")
	changelogCmd.PersistentFlags().StringVar(&changelogOptions.RecordDir, "record", "", "Record the API into a directory")
	changelogCmd.PersistentFlags().StringVar(&changelogOptions.ReplayDir, "replay", "", "Replay a previously recorded API from a directory")
	changelogCmd.PersistentFlags().BoolVar(&changelogOptions.Dependencies, "dependencies", true, "Add dependency report")

	if err := changelogCmd.MarkPersistentFlagRequired("tag"); err != nil {
		logrus.Fatalf("Unable to %v", err)
	}

	rootCmd.AddCommand(changelogCmd)
}

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

import "github.com/spf13/cobra"

// changelogCmd represents the subcommand for `krel changelog`
var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "changelog maintains the lifecycle of CHANGELOG-x.y.md files",
	Long: `krel changelog

The 'changelog' subcommand of 'krel' does the following things by utilizing
the golang based 'release-notes' tool:

1. Generate the release notes for either the patch or the new minor release
   a) Createa a new CHANGELOG-x.y.md file if we're working on a minor release
   b) Correctly prepend the generated notes to the existing CHANGELOG-x.y.md
      file if it’s a patch release. This also includes the table of contents.

2. Push the modified CHANGELOG-x.y.md into the master branch of
   kubernetes/kubernetes
   a) Push the release notes to the 'release-x.y' branch as well if it’s a
      patch release

3. Convert the markdown release notes into a HTML equivalent on purpose of
   sending it by mail to the announce list. Sending the announcement is done
   by another subcommand of 'krel', not "changelog'.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE:       initLogging,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangelog()
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(changelogCmd)
}

func runChangelog() error {
	return nil
}

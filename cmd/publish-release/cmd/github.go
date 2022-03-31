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
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/release/pkg/announce"
)

// releaseNotesCmd represents the subcommand for `krel release-notes`
var githubPageCmd = &cobra.Command{
	Use:   "github",
	Short: "Updates the github page of a release",
	Long: `publish-release github

This command updates the GitHub release page for a given tag. It will
update the page using a built in template or you can update it using
a custom template.

Before updating the page, the tag has to exist already on github.

To publish the page, --nomock has to be defined. Otherwise, the rendered
page will be printed to stdout and the program will exit.

CUSTOM TEMPLATES
================
You can define a custom golang template to use in your release page. Your
template can contain string substitutions and you can define those using 
the --substitution flag:

  --substitution="releaseTheme:Accentuate the Paw-sitive"
  --substitution="releaseLogo:accentuate-the-pawsitive.png"

ASSET FILES
===========
This command supports uploading release assets to the github page. You
can add asset files with the --asset flag:

  --asset=_output/kubernetes-1.18.2-2.fc33.x86_64.rpm

You can also specify a label for the assets by appending it with a colon
to the asset file:

  --asset="_output/kubernetes-1.18.2-2.fc33.x86_64.rpm:RPM Package for amd64"

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Run the PR creation function
		return runGithubPage(ghPageOpts)
	},
}

type githubPageCmdLineOptions struct {
	noupdate         bool
	draft            bool
	sbom             bool
	name             string
	repo             string
	template         string
	repoPath         string
	ReleaseNotesFile string
	substitutions    []string
	assets           []string
}

var ghPageOpts = &githubPageCmdLineOptions{}

func init() {
	githubPageCmd.PersistentFlags().StringVarP(
		&ghPageOpts.repo,
		"repo",
		"r",
		"",
		"repository slug containing the release page",
	)
	githubPageCmd.PersistentFlags().StringVar(
		&ghPageOpts.template,
		"template",
		"",
		"path to a custom page template",
	)
	githubPageCmd.PersistentFlags().StringVarP(
		&ghPageOpts.name,
		"name",
		"n",
		"",
		"name for the release",
	)
	githubPageCmd.PersistentFlags().StringSliceVarP(
		&ghPageOpts.assets,
		"asset",
		"a",
		[]string{},
		"Path to asset file for the release. Can be specified multiple times.",
	)
	githubPageCmd.PersistentFlags().StringSliceVarP(
		&ghPageOpts.substitutions,
		"substitution",
		"s",
		[]string{},
		"String substitution for the page template",
	)
	githubPageCmd.PersistentFlags().BoolVar(
		&ghPageOpts.noupdate,
		"noupdate",
		false,
		"Fail if the release already exists",
	)
	githubPageCmd.PersistentFlags().BoolVar(
		&ghPageOpts.draft,
		"draft",
		false,
		"Mark the release as a draft in GitHub so you can finish editing and publish it manually.",
	)
	githubPageCmd.PersistentFlags().BoolVar(
		&ghPageOpts.sbom,
		"sbom",
		true,
		"Generate an SPDX bill of materials and attach it to the release",
	)

	githubPageCmd.PersistentFlags().StringVar(
		&ghPageOpts.repoPath,
		"repo-path",
		".",
		"Path to the source code repository",
	)

	githubPageCmd.PersistentFlags().StringVar(
		&ghPageOpts.ReleaseNotesFile,
		"release-notes-file",
		"",
		"Path to a release notes markdown file to include in the release",
	)

	for _, f := range []string{"template", "asset"} {
		if err := githubPageCmd.MarkPersistentFlagFilename(f); err != nil {
			logrus.Error(err)
		}
	}

	if err := githubPageCmd.MarkPersistentFlagRequired("repo"); err != nil {
		logrus.Error(err)
	}

	rootCmd.AddCommand(githubPageCmd)
}

func getAssetsFromStrings(assetStrings []string) []announce.Asset {
	r := []announce.Asset{}
	for _, s := range assetStrings {
		parts := strings.Split(s, ":")
		l := ""
		if len(parts) > 1 {
			l = parts[1]
		}
		r = append(r, announce.Asset{
			Path:     filepath.Base(parts[0]),
			ReadFrom: parts[0],
			Label:    l,
		})
	}
	return r
}

func runGithubPage(opts *githubPageCmdLineOptions) (err error) {
	// Generate the release SBOM
	assets := getAssetsFromStrings(opts.assets)
	sbom := ""
	if opts.sbom {
		// Generate the assets file
		sbom, err = announce.GenerateReleaseSBOM(&announce.SBOMOptions{
			ReleaseName:   opts.name,
			Repo:          opts.repo,
			RepoDirectory: opts.repoPath,
			Assets:        assets,
			Tag:           commandLineOpts.tag,
		})
		if err != nil {
			return errors.Wrap(err, "generating sbom")
		}
		opts.assets = append(opts.assets, sbom+":SPDX Software Bill of Materials (SBOM)")
		// Delete the temporary sbom  when we're done
		if commandLineOpts.nomock {
			defer os.Remove(sbom)
		}
	}

	// Build the release page options
	announceOpts := announce.GitHubPageOptions{
		AssetFiles:            opts.assets,
		Tag:                   commandLineOpts.tag,
		NoMock:                commandLineOpts.nomock,
		UpdateIfReleaseExists: !opts.noupdate,
		Name:                  opts.name,
		Draft:                 opts.draft,
		ReleaseNotesFile:      opts.ReleaseNotesFile,
	}

	// Assign the repository data
	if err := announceOpts.SetRepository(opts.repo); err != nil {
		return errors.Wrap(err, "assigning the repository slug")
	}

	// Assign the substitutions
	if err := announceOpts.ParseSubstitutions(opts.substitutions); err != nil {
		return errors.Wrap(err, "parsing template substitutions")
	}

	// Read the csutom template data
	if err := announceOpts.ReadTemplate(opts.template); err != nil {
		return errors.Wrap(err, "reading the template file")
	}

	// Validate the options
	if err := announceOpts.Validate(); err != nil {
		return errors.Wrap(err, "validating options")
	}

	// Run the update process
	return announce.UpdateGitHubPage(&announceOpts)
}

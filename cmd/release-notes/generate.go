/*
Copyright 2023 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/env"

	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/options"
	"k8s.io/release/pkg/release"
)

func addGenerateFlags(subcommand *cobra.Command) {
	// githubBaseURL contains the github base URL.
	subcommand.PersistentFlags().StringVar(
		&opts.GithubBaseURL,
		"github-base-url",
		env.Default("GITHUB_BASE_URL", ""),
		"Base URL of github",
	)

	// githubUploadURL contains the github upload URL.
	subcommand.PersistentFlags().StringVar(
		&opts.GithubUploadURL,
		"github-upload-url",
		env.Default("GITHUB_UPLOAD_URL", ""),
		"Upload URL of github",
	)

	// githubOrg contains name of github organization that holds the repo to scrape.
	subcommand.PersistentFlags().StringVar(
		&opts.GithubOrg,
		"org",
		env.Default("ORG", notes.DefaultOrg),
		"Name of github organization",
	)

	// githubRepo contains name of github repository to scrape.
	subcommand.PersistentFlags().StringVar(
		&opts.GithubRepo,
		"repo",
		env.Default("REPO", notes.DefaultRepo),
		"Name of github repository",
	)

	// output contains the path on the filesystem to where the resultant
	// release notes should be printed.
	subcommand.PersistentFlags().StringVar(
		&releaseNotesOpts.outputFile,
		"output",
		env.Default("OUTPUT", ""),
		"The path to the where the release notes will be printed",
	)

	// branch is which branch to scrape.
	subcommand.PersistentFlags().StringVar(
		&opts.Branch,
		"branch",
		env.Default("BRANCH", git.DefaultBranch),
		fmt.Sprintf("Select which branch to scrape. Defaults to `%s`", git.DefaultBranch),
	)

	// startSHA contains the commit SHA where the release note generation
	// begins.
	subcommand.PersistentFlags().StringVar(
		&opts.StartSHA,
		"start-sha",
		env.Default("START_SHA", ""),
		"The commit hash to start at",
	)

	// endSHA contains the commit SHA where the release note generation ends.
	subcommand.PersistentFlags().StringVar(
		&opts.EndSHA,
		"end-sha",
		env.Default("END_SHA", ""),
		"The commit hash to end at",
	)

	// startRev contains any valid git object where the release note generation
	// begins. Can be used as alternative to start-sha.
	subcommand.PersistentFlags().StringVar(
		&opts.StartRev,
		"start-rev",
		env.Default("START_REV", ""),
		"The git revision to start at. Can be used as alternative to start-sha.",
	)

	// endRev contains any valid git object where the release note generation
	// ends. Can be used as alternative to start-sha.
	subcommand.PersistentFlags().StringVar(
		&opts.EndRev,
		"end-rev",
		env.Default("END_REV", ""),
		"The git revision to end at. Can be used as alternative to end-sha.",
	)

	// SkipFirstCommit skips the first commit if StartRev is being used. This
	// is useful if StartRev is a tag which should not be included in the
	// release notes.
	subcommand.PersistentFlags().BoolVarP(
		&opts.SkipFirstCommit,
		"skip-first-commit",
		"s",
		env.IsSet("SKIP_FIRST_COMMIT"),
		"Skip the first commit if --start-rev is being used. This is useful if the --start-rev is a tag which should not be included in the release notes.",
	)

	// repoPath contains the path to a local Kubernetes repository to avoid the
	// delay during git clone
	subcommand.PersistentFlags().StringVar(
		&opts.RepoPath,
		"repo-path",
		env.Default("REPO_PATH", ""),
		"Path to a local Kubernetes repository, used only for tag discovery.",
	)

	// format is the output format to produce the notes in.
	subcommand.PersistentFlags().StringVar(
		&opts.Format,
		"format",
		env.Default("FORMAT", options.FormatMarkdown),
		fmt.Sprintf("The format for notes output (options: %s)",
			options.FormatJSON+", "+options.FormatMarkdown,
		),
	)

	// go-template is the go template to be used when the format is markdown
	subcommand.PersistentFlags().StringVar(
		&opts.GoTemplate,
		"go-template",
		env.Default("GO_TEMPLATE", options.GoTemplateDefault),
		fmt.Sprintf("The go template to be used if --format=markdown (options: %s)",
			strings.Join([]string{
				options.GoTemplateDefault,
				options.GoTemplateInline + "<template>",
				options.GoTemplatePrefix + "<file.template>",
			}, ", "),
		),
	)

	subcommand.PersistentFlags().BoolVar(
		&opts.AddMarkdownLinks,
		"markdown-links",
		env.IsSet("MARKDOWN_LINKS"),
		"Add links for PRs and authors are added in the markdown format",
	)

	subcommand.PersistentFlags().StringVar(
		&opts.RequiredAuthor,
		"required-author",
		env.Default("REQUIRED_AUTHOR", "k8s-ci-robot"),
		"Only commits from this GitHub user are considered. Set to empty string to include all users",
	)

	subcommand.PersistentFlags().BoolVar(
		&opts.Debug,
		"debug",
		env.IsSet("DEBUG"),
		"Enable debug logging",
	)

	subcommand.PersistentFlags().StringVar(
		&opts.DiscoverMode,
		"discover",
		env.Default("DISCOVER", options.RevisionDiscoveryModeNONE),
		fmt.Sprintf("The revision discovery mode for automatic revision retrieval (options: %s)",
			strings.Join([]string{
				options.RevisionDiscoveryModeNONE,
				options.RevisionDiscoveryModeMergeBaseToLatest,
				options.RevisionDiscoveryModePatchToPatch,
				options.RevisionDiscoveryModeMinorToMinor,
			}, ", "),
		),
	)

	subcommand.PersistentFlags().StringVar(
		&opts.ReleaseBucket,
		"release-bucket",
		env.Default("RELEASE_BUCKET", release.ProductionBucket),
		"Specify gs bucket to point to in generated notes",
	)

	subcommand.PersistentFlags().StringVar(
		&opts.ReleaseTars,
		"release-tars",
		env.Default("RELEASE_TARS", ""),
		"Directory of tars to sha512 sum for display",
	)

	subcommand.PersistentFlags().BoolVar(
		&releaseNotesOpts.tableOfContents,
		"toc",
		env.IsSet("TOC"),
		"Enable the rendering of the table of contents",
	)

	subcommand.PersistentFlags().StringVar(
		&opts.RecordDir,
		"record",
		env.Default("RECORD", ""),
		"Record the API into a directory",
	)

	subcommand.PersistentFlags().StringVar(
		&opts.ReplayDir,
		"replay",
		env.Default("REPLAY", ""),
		"Replay a previously recorded API from a directory",
	)

	subcommand.PersistentFlags().BoolVar(
		&releaseNotesOpts.dependencies,
		"dependencies",
		true,
		"Add dependency report",
	)

	subcommand.PersistentFlags().StringSliceVarP(
		&opts.MapProviderStrings,
		"maps-from",
		"m",
		[]string{},
		"specify a location to recursively look for release notes *.y[a]ml file mappings",
	)
}

// addGenerate adds the generate subcomand to the main release notes cobra cmd.
func addGenerate(parent *cobra.Command) {
	// Create the cobra command
	generateCmd := &cobra.Command{
		Short:         "Generate release notes from GitHub pull request data (default)",
		Use:           "generate",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseNotes, err := notes.GatherReleaseNotes(opts)
			if err != nil {
				return fmt.Errorf("gathering release notes: %w", err)
			}

			return WriteReleaseNotes(releaseNotes)
		},
		PreRunE: func(*cobra.Command, []string) error {
			return opts.ValidateAndFinish()
		},
	}

	addGenerateFlags(generateCmd)
	parent.AddCommand(generateCmd)
}

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
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/options"
	"sigs.k8s.io/release-utils/env"
)

const (
	pullRequestGuidance = "The pull request(s) specified do not have a valid release note.\n\n" +
		"Make sure the PRs have a ```release-note block (see help) in the PR body with markdown text\n" +
		"clearly specifying the change being introduced in your commits.\n\n" +
		"If you want to skip the release notes check, you can type NONE in the release notes block\n" +
		"or add a \"release-notes-none\" (without quotes) to your PR. This will\n" +
		"tell the release notes checker to allow PRs without a note.\n\n" +
		"For more information see:\n" +
		"https://github.com/kubernetes/release/tree/master/cmd/release-notes\n\n\n"
	bkTicks = "```"
)

type checkPROptions struct {
	options.Options
	PullRequests []int
}

func (o *checkPROptions) ValidateAndFinish() error {
	var lenErr, prNrErr, orgErr, repoErr error
	if len(o.PullRequests) == 0 {
		lenErr = fmt.Errorf("no pull requests numbers specified")
	}

	for _, n := range o.PullRequests {
		if n == 0 {
			prNrErr = fmt.Errorf("invalid pull request number (must be an integer larger than 0)")
			break
		}
	}

	if o.GithubOrg == "" {
		orgErr = fmt.Errorf("no GitHub organization specified")
	}

	if o.GithubRepo == "" {
		orgErr = fmt.Errorf("no GitHub repository specified")
	}

	return errors.Join(
		lenErr, prNrErr, orgErr, repoErr,
	)
}

var checkPROpts *checkPROptions

func addCheckPRFlags(subcommand *cobra.Command) {
	// githubBaseURL contains the github base URL.
	subcommand.PersistentFlags().StringVar(
		&checkPROpts.GithubBaseURL,
		"github-base-url",
		env.Default("GITHUB_BASE_URL", ""),
		"Base URL of github",
	)

	// githubOrg contains name of github organization that holds the repo to scrape.
	subcommand.PersistentFlags().StringVar(
		&checkPROpts.GithubOrg,
		"org",
		env.Default("ORG", notes.DefaultOrg),
		"Name of github organization",
	)

	// githubRepo contains name of github repository to scrape.
	subcommand.PersistentFlags().StringVar(
		&checkPROpts.GithubRepo,
		"repo",
		env.Default("REPO", notes.DefaultRepo),
		"Name of github repository",
	)

	// Debug output
	subcommand.PersistentFlags().BoolVar(
		&checkPROpts.Debug,
		"debug",
		env.IsSet("DEBUG"),
		"Enable debug logging",
	)

	subcommand.PersistentFlags().IntSliceVar(
		&checkPROpts.PullRequests,
		"pr",
		[]int{},
		"pull request number(s) to check",
	)
}

func addCheckPR(parent *cobra.Command) {
	checkPROpts = &checkPROptions{
		Options: options.Options{},
	}

	checkprCmd := &cobra.Command{
		Short: "Checks a pull request on GitHub to ensure it has a release note",
		Long: `release-notes check checks one or more PRs to ensure they contain
a valid release note. It is a subcommand designed to run in a postsubmit job to
block PRs missing a release note.

release-notes check will retrieve pull request data using the GitHub API and
look for a valid release notes block in the PR body. For example:

` + bkTicks + `release-note
Fixed a bug to make my software even more awesome
` + bkTicks + `

When enforcing release notes in PRs, we recommend adding the release-notes block
to your PR templates. See how Kubernetes does it here:
https://github.com/kubernetes/release/blob/d546da8a2ec580ea4c024637234cc976a6ba398a/.github/PULL_REQUEST_TEMPLATE.md?plain=1#L45-L57

If you want to skip the release notes check, you can create a new "release-notes-none"
label to your PR or type "NONE" in the notes block:

` + bkTicks + `release-note
NONE
` + bkTicks + `

Either of these will instruct the release note checked to allow a PR without a
valid note.

To generate release notes from these blocks, use release-notes generate.


		`,
		Use:           "check",
		SilenceUsage:  false,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := notes.NewGatherer(context.Background(), &options.Options{
				GithubBaseURL: checkPROpts.GithubBaseURL,
				GithubOrg:     checkPROpts.GithubOrg,
				GithubRepo:    checkPROpts.GithubRepo,
			})
			if err != nil {
				return fmt.Errorf("creating notes gatherer: %w", err)
			}

			errs := []error{}

			for _, prNr := range checkPROpts.PullRequests {
				_, err := g.ReleaseNoteForPullRequest(prNr)
				if err != nil {
					errs = append(errs, fmt.Errorf("checking notes for PR #%d: %w", prNr, err))
				}
			}

			if len(errs) > 0 {
				fmt.Fprintf(os.Stderr, "\nError Checking Release Notes:\n\n"+pullRequestGuidance)
				return errors.Join(errs...)
			}

			return nil
		},
		PreRunE: func(*cobra.Command, []string) error {
			return checkPROpts.ValidateAndFinish()
		},
	}

	addCheckPRFlags(checkprCmd)
	parent.AddCommand(checkprCmd)
}

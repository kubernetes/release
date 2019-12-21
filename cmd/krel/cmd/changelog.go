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
	"context"

	"github.com/google/go-github/v28/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/notes"
)

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

type changelogOptions struct {
	branch string
	bucket string
	tars   string
	token  string
}

var changelogOpts = &changelogOptions{}

func init() {
	cobra.OnInitialize(initConfig)

	const (
		tarsFlag  = "tars"
		tokenFlag = "token"
	)
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.branch, "branch", git.Master, "The target release branch. Leave it default for non-patch releases.")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.bucket, "bucket", "kubernetes-release", "Specify gs bucket to point to in generated notes")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.tars, tarsFlag, "", "Directory of tars to sha512 sum for display")
	changelogCmd.PersistentFlags().StringVarP(&changelogOpts.token, tokenFlag, "t", "", "GitHub token for release notes retrieval")

	if err := changelogCmd.MarkPersistentFlagRequired(tokenFlag); err != nil {
		logrus.Fatal(err)
	}
	if err := changelogCmd.MarkPersistentFlagRequired(tarsFlag); err != nil {
		logrus.Fatal(err)
	}

	rootCmd.AddCommand(changelogCmd)
}

func runChangelog() (err error) {
	branch := git.Master
	revisionDiscoveryMode := notes.RevisionDiscoveryModeMinorToMinor

	if changelogOpts.branch != branch {
		if !git.IsReleaseBranch(changelogOpts.branch) {
			return errors.Wrapf(err, "Branch %q is no release branch", changelogOpts.branch)
		}
		branch = changelogOpts.branch
		revisionDiscoveryMode = notes.RevisionDiscoveryModePatchToPatch
	}
	logrus.Infof("Using branch %q", branch)
	logrus.Infof("Using discovery mode %q", revisionDiscoveryMode)

	notesOptions := notes.NewOptions()
	notesOptions.Branch = branch
	notesOptions.DiscoverMode = revisionDiscoveryMode
	notesOptions.GithubOrg = git.DefaultGithubOrg
	notesOptions.GithubRepo = git.DefaultGithubRepo
	notesOptions.GithubToken = changelogOpts.token
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.ReleaseBucket = changelogOpts.bucket
	notesOptions.ReleaseTars = changelogOpts.tars
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	if err := notesOptions.ValidateAndFinish(); err != nil {
		return err
	}

	// Create the GitHub API client
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: changelogOpts.token},
	))
	githubClient := github.NewClient(httpClient)

	// Fetch a list of fully-contextualized release notes
	logrus.Info("fetching all commits. This might take a while...")

	gatherer := &notes.Gatherer{
		Client:  notes.WrapGithubClient(githubClient),
		Context: ctx,
		Org:     git.DefaultGithubOrg,
		Repo:    git.DefaultGithubRepo,
	}
	releaseNotes, history, err := gatherer.ListReleaseNotes(
		branch, notesOptions.StartSHA, notesOptions.EndSHA, "", "",
	)
	if err != nil {
		return errors.Wrapf(err, "listing release notes")
	}

	// Create the markdown
	doc, err := notes.CreateDocument(releaseNotes, history)
	if err != nil {
		return errors.Wrapf(err, "creating release note document")
	}

	// TODO: mangle the documents into the target files
	logrus.Infof("doc: %v", doc)

	return nil
}

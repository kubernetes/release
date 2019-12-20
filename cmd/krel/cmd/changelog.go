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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/v28/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/util"
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
	branch    string
	bucket    string
	tars      string
	token     string
	outputDir string
}

var changelogOpts = &changelogOptions{}

const (
	tocStart = "<!-- BEGIN MUNGE: GENERATED_TOC -->"
	tocEnd   = "<!-- END MUNGE: GENERATED_TOC -->"
)

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
	changelogCmd.PersistentFlags().StringVarP(&changelogOpts.outputDir, "output", "o", os.TempDir(), "Output directory for the generated files")

	if err := changelogCmd.MarkPersistentFlagRequired(tokenFlag); err != nil {
		logrus.Fatal(err)
	}
	if err := changelogCmd.MarkPersistentFlagRequired(tarsFlag); err != nil {
		logrus.Fatal(err)
	}

	rootCmd.AddCommand(changelogCmd)
}

func runChangelog() (err error) {
	// TODO: remote lookup of 'final release notes draft',
	// including error handling if not found

	markdown, version, err := createReleaseNotes()
	if err != nil {
		return err
	}
	toc, err := notes.GenerateTOC(markdown)
	if err != nil {
		return err
	}

	if err := writeMarkdown(toc, markdown, version); err != nil {
		return err
	}

	// TODO: HTML output
	// TODO: Pushing changes
	return nil
}

func createReleaseNotes() (markdown, version string, err error) {
	revisionDiscoveryMode := notes.RevisionDiscoveryModeMinorToMinor

	if changelogOpts.branch != git.Master {
		if !git.IsReleaseBranch(changelogOpts.branch) {
			return "", "", errors.Wrapf(
				err, "Branch %q is no release branch", changelogOpts.branch,
			)
		}
		revisionDiscoveryMode = notes.RevisionDiscoveryModePatchToPatch
	}
	logrus.Infof("Using branch %q", changelogOpts.branch)
	logrus.Infof("Using discovery mode %q", revisionDiscoveryMode)

	notesOptions := notes.NewOptions()
	notesOptions.Branch = changelogOpts.branch
	notesOptions.DiscoverMode = revisionDiscoveryMode
	notesOptions.GithubOrg = git.DefaultGithubOrg
	notesOptions.GithubRepo = git.DefaultGithubRepo
	notesOptions.GithubToken = changelogOpts.token
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.ReleaseBucket = changelogOpts.bucket
	notesOptions.ReleaseTars = changelogOpts.tars
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	if err := notesOptions.ValidateAndFinish(); err != nil {
		return "", "", err
	}

	// Create the GitHub API client
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: changelogOpts.token},
	))
	githubClient := github.NewClient(httpClient)

	gatherer := &notes.Gatherer{
		Client:  notes.WrapGithubClient(githubClient),
		Context: ctx,
		Org:     git.DefaultGithubOrg,
		Repo:    git.DefaultGithubRepo,
	}
	releaseNotes, history, err := gatherer.ListReleaseNotes(
		changelogOpts.branch,
		notesOptions.StartSHA,
		notesOptions.EndSHA,
		"", "",
	)
	if err != nil {
		return "", "", errors.Wrapf(err, "listing release notes")
	}

	// Create the markdown
	doc, err := notes.CreateDocument(releaseNotes, history)
	if err != nil {
		return "", "", errors.Wrapf(err, "creating release note document")
	}

	markdown, err = notes.RenderMarkdown(
		doc, changelogOpts.bucket, changelogOpts.tars,
		notesOptions.StartRev, notesOptions.EndRev,
	)
	if err != nil {
		return "", "", errors.Wrapf(
			err, "rendering release notes to markdown",
		)
	}

	return markdown, util.TrimTagPrefix(notesOptions.StartRev), nil
}

func writeMarkdown(toc, markdown, version string) error {
	changelogPath, err := changelogFilename(version)
	if err != nil {
		return err
	}

	writeFile := func(t, m string) error {
		return ioutil.WriteFile(
			changelogPath, []byte(strings.Join(
				[]string{addTocMarkers(t), strings.TrimSpace(m)}, "\n",
			)), 0o644,
		)
	}

	// No changelog exists, simply write the content to a new one
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		logrus.Infof("Changelog %q does not exist, creating it", changelogPath)
		return writeFile(toc, markdown)
	}

	// Changelog seems to exist, prepend the notes and re-generate the TOC
	logrus.Infof("Adding new content to changelog file %q ", changelogPath)
	content, err := ioutil.ReadFile(changelogPath)
	if err != nil {
		return err
	}

	tocEndIndex := bytes.Index(content, []byte(tocEnd))
	if tocEndIndex < 0 {
		return errors.Errorf(
			"unable to find table of contents end marker `%s` in %q",
			tocEnd, changelogPath,
		)
	}

	mergedMarkdown := fmt.Sprintf(
		"%s\n%s",
		strings.TrimSpace(markdown),
		string(content[(len(tocEnd)+tocEndIndex):]),
	)
	mergedTOC, err := notes.GenerateTOC(mergedMarkdown)
	if err != nil {
		return err
	}
	return writeFile(mergedTOC, mergedMarkdown)
}

func changelogFilename(version string) (string, error) {
	v, err := semver.Parse(version)
	if err != nil {
		return "", err
	}
	return filepath.Join(
		changelogOpts.outputDir,
		fmt.Sprintf("CHANGELOG-%d.%d.md", v.Major, v.Minor),
	), nil
}

func addTocMarkers(toc string) string {
	return fmt.Sprintf("%s\n\n%s\n%s\n", tocStart, toc, tocEnd)
}

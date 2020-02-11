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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/options"
	"k8s.io/release/pkg/util"
)

// releaseNotesCmd represents the subcommand for `krel release-notes`
var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "release-notes is the tool of choice for the Release Notes subteam of SIG Release",
	Long: fmt.Sprintf(`krel release-notes

The 'release-notes' subcommand of krel has been developed to:

1. Generate the release notes for the provided tag for commits on the master
   branch. We always use the master branch because a release branch
   gets fast-forwarded until we hit the first release candidate (rc). This is
   also the reason why we select the first 'v1.xx.0-rc.1' as start tag for
   the notes generation.

2. Put the generated notes into a release notes draft markdown document and
   create a GitHub pull request targeting to update the file:
   https://github.com/kubernetes/sig-release/blob/master/releases/release-1.xx/release-notes-draft.md

3. Put the generated notes into a JSON file and create a GitHub pull request
   to update the website https://relnotes.k8s.io.

To use the tool, please set the %v environment variable which needs write
permissions to your fork of k/sig-release and k-sigs/release-notes.`,
		options.TokenKey),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReleaseNotes()
	},
}

type releaseNotesOptions struct {
	tag string
}

type releaseNotesResult struct {
	markdown string
	json     string
}

var releaseNotesOpts = &releaseNotesOptions{}

func init() {
	releaseNotesCmd.PersistentFlags().StringVarP(
		&releaseNotesOpts.tag,
		"tag",
		"t",
		"",
		"version tag for the notes",
	)

	rootCmd.AddCommand(releaseNotesCmd)
}

func runReleaseNotes() (err error) {
	var tag string
	if releaseNotesOpts.tag == "" {
		tag, err = tryToFindLatestMinorTag()
		if err != nil {
			return errors.Wrapf(err, "unable to find latest minor tag")
		}
	} else {
		tag = releaseNotesOpts.tag
	}

	s, err := util.TagStringToSemver(tag)
	if err != nil {
		return errors.Wrapf(err, "no valid tag: %v", tag)
	}
	start := util.SemverToTagString(semver.Version{
		Major: s.Major,
		Minor: s.Minor - 1,
		Patch: 0,
		Pre:   []semver.PRVersion{{VersionStr: "rc.1"}},
	})
	logrus.Infof("Using start tag %v", start)
	logrus.Infof("Using end tag %v", tag)

	_, err = releaseNotesFrom(start)
	if err != nil {
		return errors.Wrapf(err, "generating release notes")
	}

	//TODO: implement PR creation for k-sigs/release-notes and k/sig-release
	return nil
}

// tryToFindLatestMinorTag looks-up the default k/k remote to find the latest
// non final version
func tryToFindLatestMinorTag() (string, error) {
	status, err := command.New(
		"git", "ls-remote", "--sort=v:refname",
		"--tags", git.DefaultGithubRepoURL,
	).
		Pipe("grep", "-Eo", "v[0-9].[0-9]+.0-.*.[0-9]$").
		Pipe("tail", "-1").
		RunSilent()

	if err != nil {
		return "", errors.Wrapf(err, "git ls-remote command failed")
	}

	if !status.Success() {
		return "", errors.Errorf(
			"git ls-remote command not successful: %v", status.Error(),
		)
	}

	return strings.TrimSpace(status.Output()), nil
}

func releaseNotesFrom(startTag string) (*releaseNotesResult, error) {
	logrus.Info("Generating release notes")

	notesOptions := options.New()
	notesOptions.Branch = git.Master
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.StartRev = startTag
	notesOptions.EndRev = releaseNotesOpts.tag
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return nil, err
	}

	// Fetch the notes
	gatherer := notes.NewGatherer(context.Background(), notesOptions)
	releaseNotes, history, err := gatherer.ListReleaseNotes()
	if err != nil {
		return nil, errors.Wrapf(err, "listing release notes")
	}

	doc, err := notes.CreateDocument(releaseNotes, history)
	if err != nil {
		return nil, errors.Wrapf(err, "creating release note document")
	}

	// Create the markdown
	markdown, err := doc.RenderMarkdown(
		"", "", notesOptions.StartRev, notesOptions.EndRev,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err, "rendering release notes to markdown",
		)
	}

	// Create the JSON
	j, err := json.Marshal(releaseNotes)
	if err != nil {
		return nil, errors.Wrapf(err, "generating release notes JSON")
	}

	return &releaseNotesResult{markdown: markdown, json: string(j)}, nil
}

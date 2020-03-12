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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/document"
	"k8s.io/release/pkg/notes/options"
	"k8s.io/release/pkg/util"
)

const (
	// draftFilename filename for the release notes draft
	draftFilename = "release-notes-draft.md"
)

// releaseNotesCmd represents the subcommand for `krel release-notes`
var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "The subcommand of choice for the Release Notes subteam of SIG Release",
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
		github.TokenEnvKey),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReleaseNotes()
	},
}

type releaseNotesOptions struct {
	tag                string
	draftOrg           string
	draftRepo          string
	createDraftPR      bool
	outputDir          string
	sigreleaseForkPath string
	Format             string
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

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.draftOrg,
		"draft-org",
		"",
		"a Github organization owner of the fork of k/sig-release where the Release Notes Draft PR will be created",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.draftRepo,
		"draft-repo",
		git.DefaultGithubReleaseRepo,
		"the name of the fork of k/sig-release, the Release Notes Draft PR will be created from this repository",
	)

	releaseNotesCmd.PersistentFlags().BoolVar(
		&releaseNotesOpts.createDraftPR,
		"create-draft-pr",
		false,
		"create the Release Notes Draft PR. --draft-org and --draft-repo must be set along with this option",
	)

	releaseNotesCmd.PersistentFlags().StringVarP(
		&releaseNotesOpts.outputDir,
		"output-dir",
		"o",
		".",
		"output a copy of the release notes to this directory",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.Format,
		"format",
		util.EnvDefault("FORMAT", "markdown"),
		"The format for notes output (options: markdown, json)",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.sigreleaseForkPath,
		"sigrelease-fork-path",
		filepath.Join(os.TempDir(), "k8s-sigrelease"),
		"output a copy of the release notes to this directory",
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
		releaseNotesOpts.tag = tag
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

	if releaseNotesOpts.createDraftPR {
		if err = validateDraftPROptions(); err != nil {
			return errors.Wrap(err, "validating PR command line options")
		}
	}

	result, err := releaseNotesFrom(start)
	if err != nil {
		return errors.Wrapf(err, "generating release notes")
	}

	// Create RN draft PR
	if releaseNotesOpts.createDraftPR {
		if err := createDraftPR(tag, result); err != nil {
			return errors.Wrap(err, "failed to create release notes draft PR")
		}
		return nil
	}

	switch releaseNotesOpts.Format {
	case "json":
		err = ioutil.WriteFile(filepath.Join(releaseNotesOpts.outputDir, "release-notes.json"), []byte(result.json), 0644)
		if err != nil {
			return errors.Wrap(err, "writing release notes JSON file")
		}
	case "markdown":
		err = ioutil.WriteFile(filepath.Join(releaseNotesOpts.outputDir, "release-notes.md"), []byte(result.json), 0644)
		if err != nil {
			return errors.Wrap(err, "writing release notes markdown file")
		}
	default:
		return errors.Errorf("%q is an unsupported format", releaseNotesOpts.Format)
	}

	// TODO: implement PR creation for k-sigs/release-notes
	return nil
}

// validateDraftPROptions checks if we have all needed parameters to create the Release Notes PR
func validateDraftPROptions() error {
	if releaseNotesOpts.createDraftPR {
		// Check if --draft-org is set
		if releaseNotesOpts.draftOrg == "" {
			logrus.Warn("cannot generate the Release Notes draft PR without --draft-org")
		}

		// Check if --draft-repo is set
		if releaseNotesOpts.draftRepo == "" {
			logrus.Warn("cannot generate the Release Notes draft PR without --draft-repo")
		}

		if releaseNotesOpts.draftOrg == "" || releaseNotesOpts.draftRepo == "" {
			return errors.New("To generate the release notes draft you must define both --draft-org and --draft-repo")
		}
	}
	return nil
}

// createDraftPR pushes the release notes draft to the users fork
func createDraftPR(tag string, result *releaseNotesResult) error {
	s, err := util.TagStringToSemver(tag)
	if err != nil {
		return errors.Wrapf(err, "no valid tag: %v", tag)
	}

	branchname := "release-notes-draft-" + tag

	// checkout kubernetes/sig-release
	sigReleaseRepo, err := git.CloneOrOpenGitHubRepo(releaseNotesOpts.sigreleaseForkPath, git.DefaultGithubOrg, git.DefaultGithubReleaseRepo, true)
	if err != nil {
		return errors.Wrap(err, "cloning k/sig-release")
	}

	// test if the fork remote is already existing
	const remote = "userfork"
	url, err := git.GetRepoURL(
		releaseNotesOpts.draftOrg, releaseNotesOpts.draftRepo, true,
	)
	if err != nil {
		return errors.Wrap(err, "unable to get repository URL")
	}
	if sigReleaseRepo.HasRemote(remote, url) {
		logrus.Infof(
			"Using already existing remote %v (%v) in repository",
			remote, url,
		)
	} else {
		// add the user's fork as a remote
		err = sigReleaseRepo.AddRemote(remote, releaseNotesOpts.draftOrg, releaseNotesOpts.draftRepo)
		if err != nil {
			return errors.Wrap(err, "adding users fork as remote repository")
		}
	}

	// verify the branch doesn't already exist on the user's fork
	err = sigReleaseRepo.HasRemoteBranch(branchname)
	if err == nil {
		return errors.Errorf("remote repo already has a branch named %s", branchname)
	}

	// checkout the new branch
	err = sigReleaseRepo.Checkout("-b", branchname)
	if err != nil {
		return errors.Wrapf(err, "creating new branch %s", branchname)
	}

	// generate the notes
	targetdir := filepath.Join(sigReleaseRepo.Dir(), "releases", fmt.Sprintf("release-%d.%d", s.Major, s.Minor))
	logrus.Debugf("Release notes markdown will be written to %s", targetdir)
	err = ioutil.WriteFile(filepath.Join(targetdir, draftFilename), []byte(result.markdown), 0644)
	if err != nil {
		return errors.Wrapf(err, "writing release notes draft")
	}

	// commit the results
	err = sigReleaseRepo.Add(filepath.Join("releases", fmt.Sprintf("release-%d.%d", s.Major, s.Minor), draftFilename))
	if err != nil {
		return errors.Wrap(err, "adding release notes draft to staging area")
	}

	err = sigReleaseRepo.Commit("Release Notes draft for k/k " + tag)
	if err != nil {
		return errors.Wrapf(err, "Error creating commit in %s/%s", releaseNotesOpts.draftOrg, releaseNotesOpts.draftRepo)
	}

	// push to fork
	logrus.Infof("Pushing release notes draft to %s/%s", releaseNotesOpts.draftOrg, releaseNotesOpts.draftRepo)
	err = sigReleaseRepo.PushToRemote(remote, branchname)
	if err != nil {
		return errors.Wrapf(err, "pushing changes to %s/%s", releaseNotesOpts.draftOrg, releaseNotesOpts.draftRepo)
	}

	// TODO: Call github API and create PR against k/sig-release
	return nil
}

// tryToFindLatestMinorTag looks-up the default k/k remote to find the latest
// non final version
func tryToFindLatestMinorTag() (string, error) {
	url, err := git.GetDefaultKubernetesRepoURL()
	if err != nil {
		return "", err
	}

	status, err := command.New(
		"git", "ls-remote", "--sort=v:refname",
		"--tags", url,
	).
		Pipe("grep", "-Eo", "v[0-9].[0-9]+.0-.*.[0-9]$").
		Pipe("tail", "-1").
		RunSilentSuccessOutput()

	if err != nil {
		return "", err
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

	doc, err := document.CreateDocument(releaseNotes, history)
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

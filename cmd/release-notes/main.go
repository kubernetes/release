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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v28/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/notes"
)

type options struct {
	githubToken    string
	githubOrg      string
	githubRepo     string
	output         string
	branch         string
	startSHA       string
	endSHA         string
	startRev       string
	endRev         string
	repoPath       string
	releaseVersion string
	format         string
	requiredAuthor string
	debug          bool
	discoverMode   string
	releaseBucket  string
	releaseTars    string
}

var opts = &options{}

const (
	revisionDiscoveryModeNONE          = "none"
	revisionDiscoveryModeMinorToLatest = "minor-to-latest"
)

var cmd = &cobra.Command{
	Short:         "release-notes - The Kubernetes Release Notes Generator",
	Use:           "release-notes",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE:       validateOptions,
	RunE:          run,
}

func init() {
	// githubToken contains a personal GitHub access token. This is used to
	// scrape the commits of the Kubernetes repo.
	cmd.PersistentFlags().StringVar(
		&opts.githubToken,
		"github-token",
		envDefault("GITHUB_TOKEN", ""),
		"A personal GitHub access token (required)",
	)

	// githubOrg contains name of github organization that holds the repo to scrape.
	cmd.PersistentFlags().StringVar(
		&opts.githubOrg,
		"github-org",
		envDefault("GITHUB_ORG", "kubernetes"),
		"Name of github organization",
	)

	// githubRepo contains name of github repository to scrape.
	cmd.PersistentFlags().StringVar(
		&opts.githubRepo,
		"github-repo",
		envDefault("GITHUB_REPO", "kubernetes"),
		"Name of github repository",
	)

	// output contains the path on the filesystem to where the resultant
	// release notes should be printed.
	cmd.PersistentFlags().StringVar(
		&opts.output,
		"output",
		envDefault("OUTPUT", ""),
		"The path to the where the release notes will be printed",
	)

	// branch is which branch to scrape.
	cmd.PersistentFlags().StringVar(
		&opts.branch,
		"branch",
		envDefault("BRANCH", "master"),
		"Select which branch to scrape. Defaults to `master`",
	)

	// startSHA contains the commit SHA where the release note generation
	// begins.
	cmd.PersistentFlags().StringVar(
		&opts.startSHA,
		"start-sha",
		envDefault("START_SHA", ""),
		"The commit hash to start at",
	)

	// endSHA contains the commit SHA where the release note generation ends.
	cmd.PersistentFlags().StringVar(
		&opts.endSHA,
		"end-sha",
		envDefault("END_SHA", ""),
		"The commit hash to end at",
	)

	// startRev contains any valid git object where the release note generation
	// begins. Can be used as alternative to start-sha.
	cmd.PersistentFlags().StringVar(
		&opts.startRev,
		"start-rev",
		envDefault("START_REV", ""),
		"The git revision to start at. Can be used as alternative to start-sha.",
	)

	// endRev contains any valid git object where the release note generation
	// ends. Can be used as alternative to start-sha.
	cmd.PersistentFlags().StringVar(
		&opts.endRev,
		"end-rev",
		envDefault("END_REV", ""),
		"The git revision to end at. Can be used as alternative to end-sha.",
	)

	// repoPath contains the path to a local Kubernetes repository to avoid the
	// delay during git clone
	cmd.PersistentFlags().StringVar(
		&opts.repoPath,
		"repo-path",
		envDefault("REPO_PATH", filepath.Join(os.TempDir(), "k8s-repo")),
		"Path to a local Kubernetes repository, used only for tag discovery.",
	)

	// releaseVersion is the version number you want to tag the notes with.
	cmd.PersistentFlags().StringVar(
		&opts.releaseVersion,
		"release-version",
		envDefault("RELEASE_VERSION", ""),
		"Which release version to tag the entries as.",
	)

	// format is the output format to produce the notes in.
	cmd.PersistentFlags().StringVar(
		&opts.format,
		"format",
		envDefault("FORMAT", "markdown"),
		"The format for notes output (options: markdown, json)",
	)

	cmd.PersistentFlags().StringVar(
		&opts.requiredAuthor,
		"requiredAuthor",
		envDefault("REQUIRED_AUTHOR", "k8s-ci-robot"),
		"Only commits from this GitHub user are considered. Set to empty string to include all users",
	)

	cmd.PersistentFlags().BoolVar(
		&opts.debug,
		"debug",
		isEnvSet("DEBUG"),
		"Enable debug logging",
	)

	cmd.PersistentFlags().StringVar(
		&opts.discoverMode,
		"discover",
		envDefault("DISCOVER", revisionDiscoveryModeNONE),
		fmt.Sprintf("The revision discovery mode for automatic revision retrieval (options: %s)",
			strings.Join([]string{
				revisionDiscoveryModeNONE,
				revisionDiscoveryModeMinorToLatest,
			}, ", "),
		),
	)

	cmd.PersistentFlags().StringVar(
		&opts.releaseBucket,
		"release-bucket",
		envDefault("RELEASE_BUCKET", "kubernetes-release"),
		"Specify gs bucket to point to in generated notes",
	)

	cmd.PersistentFlags().StringVar(
		&opts.releaseTars,
		"release-tars",
		envDefault("RELEASE_TARS", ""),
		"Directory of tars to sha512 sum for display",
	)
}

func envDefault(key, def string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return value
}

func isEnvSet(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func GetReleaseNotes() (notes.ReleaseNotes, notes.ReleaseNotesHistory, error) {
	// Create the GitHub API client
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: opts.githubToken},
	))
	githubClient := github.NewClient(httpClient)

	// Fetch a list of fully-contextualized release notes
	logrus.Info("fetching all commits. This might take a while...")

	apiOptions := []notes.GitHubAPIOption{notes.WithContext(ctx)}
	if opts.githubOrg != "" {
		apiOptions = append(apiOptions, notes.WithOrg(opts.githubOrg))
	}
	if opts.githubRepo != "" {
		apiOptions = append(apiOptions, notes.WithRepo(opts.githubRepo))
	}

	notesClient := notes.WrapGithubClient(githubClient)
	releaseNotes, history, err := notes.ListReleaseNotes(
		notesClient, opts.branch, opts.startSHA, opts.endSHA,
		opts.requiredAuthor, opts.releaseVersion, apiOptions...)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "listing release notes")
	}

	return releaseNotes, history, nil
}

func WriteReleaseNotes(releaseNotes notes.ReleaseNotes, history notes.ReleaseNotesHistory) (err error) {
	logrus.Info("got the commits, performing rendering")

	// Open a handle to the file which will contain the release notes output
	var output *os.File
	var existingNotes notes.ReleaseNotes

	if opts.output != "" {
		output, err = os.OpenFile(opts.output, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return errors.Wrapf(err, "opening the supplied output file")
		}
	} else {
		output, err = ioutil.TempFile("", "release-notes-")
		if err != nil {
			return errors.Wrapf(err, "creating a temporary file to write the release notes to")
		}
	}

	// Contextualized release notes can be printed in a variety of formats
	switch opts.format {
	case "json":
		byteValue, err := ioutil.ReadAll(output)
		if err != nil {
			return err
		}

		if len(byteValue) > 0 {
			if err := json.Unmarshal(byteValue, &existingNotes); err != nil {
				return errors.Wrapf(err, "unmarshalling existing notes")
			}
		}

		if len(existingNotes) > 0 {
			if err := output.Truncate(0); err != nil {
				return err
			}
			if _, err := output.Seek(0, 0); err != nil {
				return err
			}

			for i := 0; i < len(existingNotes); i++ {
				_, ok := releaseNotes[existingNotes[i].PrNumber]
				if !ok {
					releaseNotes[existingNotes[i].PrNumber] = existingNotes[i]
				}
			}
		}

		enc := json.NewEncoder(output)
		enc.SetIndent("", "  ")
		if err := enc.Encode(releaseNotes); err != nil {
			return errors.Wrapf(err, "encoding JSON output")
		}
	case "markdown":
		doc, err := notes.CreateDocument(releaseNotes, history)
		if err != nil {
			return errors.Wrapf(err, "creating release note document")
		}

		if err := notes.RenderMarkdown(
			output, doc, opts.releaseBucket,
			opts.releaseTars, opts.startRev, opts.endRev,
		); err != nil {
			return errors.Wrapf(err, "rendering release note document to markdown")
		}

	default:
		return errors.Errorf("%q is an unsupported format", opts.format)
	}

	logrus.
		WithField("path", output.Name()).
		WithField("format", opts.format).
		Info("release notes written to file")
	return nil
}

func validateOptions(*cobra.Command, []string) error {
	// The GitHub Token is required.
	if opts.githubToken == "" {
		return errors.New("GitHub token must be set via -github-token or $GITHUB_TOKEN")
	}

	// Check if we want to automatically discover the revisions
	if opts.discoverMode == revisionDiscoveryModeMinorToLatest {
		repo, err := git.CloneOrOpenGitHubRepo(
			opts.repoPath,
			opts.githubOrg,
			opts.githubRepo,
			false,
		)
		if err != nil {
			return err
		}
		start, end, err := repo.LatestNonPatchFinalToLatest()
		if err != nil {
			return err
		}
		opts.startSHA = start
		opts.endSHA = end
		logrus.Infof("discovered start SHA %s", start)
		logrus.Infof("discovered end SHA %s", end)
	}

	// The start SHA is required.
	if opts.startSHA == "" && opts.startRev == "" {
		return errors.New("the starting commit hash must be set via -start-sha, $START_SHA, -start-rev or $START_REV")
	}

	// The end SHA is required.
	if opts.endSHA == "" && opts.endRev == "" {
		return errors.New("the ending commit hash must be set via -end-sha, $END_SHA, -end-rev or $END_REV")
	}

	// Check if we have to parse a revision
	if opts.startRev != "" || opts.endRev != "" {
		logrus.Info("cloning/updating repository to discover start or end sha")
		repo, err := git.CloneOrOpenGitHubRepo(
			opts.repoPath,
			opts.githubOrg,
			opts.githubRepo,
			false,
		)
		if err != nil {
			return err
		}
		if opts.startRev != "" {
			sha, err := repo.RevParse(opts.startRev)
			if err != nil {
				return err
			}
			logrus.Infof("using found start SHA: %s", sha)
			opts.startSHA = sha
		}
		if opts.endRev != "" {
			sha, err := repo.RevParse(opts.endRev)
			if err != nil {
				return err
			}
			logrus.Infof("using found end SHA: %s", sha)
			opts.endSHA = sha
		}
	}

	// Add appropriate log filtering
	if opts.debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	return nil
}

func run(*cobra.Command, []string) error {
	releaseNotes, history, err := GetReleaseNotes()
	if err != nil {
		return errors.Wrapf(err, "retrieving release notes")
	}

	return WriteReleaseNotes(releaseNotes, history)
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	if err := cmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

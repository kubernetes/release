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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v28/github"
	"github.com/kolide/kit/env"
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
	logger         log.Logger
}

const (
	revisionDiscoveryModeNONE          = "none"
	revisionDiscoveryModeMinorToLatest = "minor-to-latest"
)

func (o *options) BindFlags() *flag.FlagSet {
	flags := flag.NewFlagSet("release-notes", flag.ContinueOnError)
	// githubToken contains a personal GitHub access token. This is used to
	// scrape the commits of the Kubernetes repo.
	flags.StringVar(
		&o.githubToken,
		"github-token",
		env.String("GITHUB_TOKEN", ""),
		"A personal GitHub access token (required)",
	)

	// githubOrg contains name of github organization that holds the repo to scrape.
	flags.StringVar(
		&o.githubOrg,
		"github-org",
		env.String("GITHUB_ORG", "kubernetes"),
		"Name of github organization",
	)

	// githubRepo contains name of github repository to scrape.
	flags.StringVar(
		&o.githubRepo,
		"github-repo",
		env.String("GITHUB_REPO", "kubernetes"),
		"Name of github repository",
	)

	// output contains the path on the filesystem to where the resultant
	// release notes should be printed.
	flags.StringVar(
		&o.output,
		"output",
		env.String("OUTPUT", ""),
		"The path to the where the release notes will be printed",
	)

	// branch is which branch to scrape.
	flags.StringVar(
		&o.branch,
		"branch",
		env.String("BRANCH", "master"),
		"Select which branch to scrape. Defaults to `master`",
	)

	// startSHA contains the commit SHA where the release note generation
	// begins.
	flags.StringVar(
		&o.startSHA,
		"start-sha",
		env.String("START_SHA", ""),
		"The commit hash to start at",
	)

	// endSHA contains the commit SHA where the release note generation ends.
	flags.StringVar(
		&o.endSHA,
		"end-sha",
		env.String("END_SHA", ""),
		"The commit hash to end at",
	)

	// startRev contains any valid git object where the release note generation
	// begins. Can be used as alternative to start-sha.
	flags.StringVar(
		&o.startRev,
		"start-rev",
		env.String("START_REV", ""),
		"The git revision to start at. Can be used as alternative to start-sha.",
	)

	// endRev contains any valid git object where the release note generation
	// ends. Can be used as alternative to start-sha.
	flags.StringVar(
		&o.endRev,
		"end-rev",
		env.String("END_REV", ""),
		"The git revision to end at. Can be used as alternative to end-sha.",
	)

	// repoPath contains the path to a local Kubernetes repository to avoid the
	// delay during git clone
	flags.StringVar(
		&o.repoPath,
		"repo-path",
		env.String("REPO_PATH", ""),
		"Path to a the local Kubernetes repository",
	)

	// releaseVersion is the version number you want to tag the notes with.
	flags.StringVar(
		&o.releaseVersion,
		"release-version",
		env.String("RELEASE_VERSION", ""),
		"Which release version to tag the entries as.",
	)

	// format is the output format to produce the notes in.
	flags.StringVar(
		&o.format,
		"format",
		env.String("FORMAT", "markdown"),
		"The format for notes output (options: markdown, json)",
	)

	flags.StringVar(
		&o.requiredAuthor,
		"requiredAuthor",
		env.String("REQUIRED_AUTHOR", "k8s-ci-robot"),
		"Only commits from this GitHub user are considered. Set to empty string to include all users",
	)

	flags.BoolVar(
		&o.debug,
		"debug",
		env.Bool("DEBUG", false),
		"Enable debug logging",
	)

	flags.StringVar(
		&o.discoverMode,
		"discover",
		env.String("DISCOVER", revisionDiscoveryModeNONE),
		fmt.Sprintf("The revision discovery mode for automatic revision retrieval (options: %s, %s)",
			revisionDiscoveryModeNONE,
			revisionDiscoveryModeMinorToLatest),
	)

	return flags
}

func (o *options) GetReleaseNotes() (notes.ReleaseNotes, notes.ReleaseNotesHistory, error) {
	// Create the GitHub API client
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: o.githubToken},
	))
	githubClient := github.NewClient(httpClient)

	// Fetch a list of fully-contextualized release notes
	level.Info(o.logger).Log("msg", "fetching all commits. this might take a while...")

	opts := []notes.GitHubAPIOption{notes.WithContext(ctx)}
	if o.githubOrg != "" {
		opts = append(opts, notes.WithOrg(o.githubOrg))
	}
	if o.githubRepo != "" {
		opts = append(opts, notes.WithRepo(o.githubRepo))
	}

	releaseNotes, history, err := notes.ListReleaseNotes(
		githubClient, o.logger, o.branch, o.startSHA, o.endSHA,
		o.requiredAuthor, o.releaseVersion, opts...)
	if err != nil {
		level.Error(o.logger).Log("msg", "error generating release notes", "err", err)
		return nil, nil, err
	}

	return releaseNotes, history, nil
}

func (o *options) WriteReleaseNotes(releaseNotes notes.ReleaseNotes, history notes.ReleaseNotesHistory) error {
	level.Info(o.logger).Log("msg", "got the commits, performing rendering")

	// Open a handle to the file which will contain the release notes output
	var output *os.File
	var err error
	var existingNotes notes.ReleaseNotes

	if o.output != "" {
		output, err = os.OpenFile(o.output, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			level.Error(o.logger).Log("msg", "error opening the supplied output file", "err", err)
			return err
		}
	} else {
		output, err = ioutil.TempFile("", "release-notes-")
		if err != nil {
			level.Error(o.logger).Log("msg", "error creating a temporary file to write the release notes to", "err", err)
			return err
		}
	}

	// Contextualized release notes can be printed in a variety of formats
	switch o.format {
	case "json":
		byteValue, _ := ioutil.ReadAll(output)

		if len(byteValue) > 0 {
			if err := json.Unmarshal(byteValue, &existingNotes); err != nil {
				level.Error(o.logger).Log("msg", "error unmarshalling existing notes", "err", err)
				return err
			}
		}

		if len(existingNotes) > 0 {
			output.Truncate(0)
			output.Seek(0, 0)

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
			level.Error(o.logger).Log("msg", "error encoding JSON output", "err", err)
			os.Exit(1)
		}
	case "markdown":
		doc, err := notes.CreateDocument(releaseNotes, history)
		if err != nil {
			level.Error(o.logger).Log("msg", "error creating release note document", "err", err)
			return err
		}

		if err := notes.RenderMarkdown(doc, output); err != nil {
			level.Error(o.logger).Log("msg", "error rendering release note document to markdown", "err", err)
			return err
		}

	default:
		errString := fmt.Sprintf("%q is an unsupported format", o.format)
		level.Error(o.logger).Log("msg", errString)
		return errors.New(errString)
	}

	level.Info(o.logger).Log(
		"msg", "release notes written to file",
		"path", output.Name(),
		"format", o.format,
	)
	return nil
}

func parseOptions(args []string, logger log.Logger) (*options, error) {
	opts := &options{}
	flags := opts.BindFlags()

	// Parse the args.
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	// The GitHub Token is required.
	if opts.githubToken == "" {
		return nil, errors.New("GitHub token must be set via -github-token or $GITHUB_TOKEN")
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
			return nil, err
		}
		start, end, err := repo.LatestNonPatchFinalToLatest()
		if err != nil {
			return nil, err
		}
		opts.startSHA = start
		opts.endSHA = end
		level.Info(logger).Log("msg", "discovered start SHA "+start)
		level.Info(logger).Log("msg", "discovered end SHA "+end)
	}

	// The start SHA is required.
	if opts.startSHA == "" && opts.startRev == "" {
		return nil, errors.New("the starting commit hash must be set via -start-sha, $START_SHA, -start-rev or $START_REV")
	}

	// The end SHA is required.
	if opts.endSHA == "" && opts.endRev == "" {
		return nil, errors.New("the ending commit hash must be set via -end-sha, $END_SHA, -end-rev or $END_REV")
	}

	// Check if we have to parse a revision
	if opts.startRev != "" || opts.endRev != "" {
		level.Info(logger).Log("msg", "cloning/updating repository to discover start or end sha")
		repo, err := git.CloneOrOpenGitHubRepo(
			opts.repoPath,
			opts.githubOrg,
			opts.githubRepo,
			false,
		)
		if err != nil {
			return nil, err
		}
		if opts.startRev != "" {
			sha, err := repo.RevParse(opts.startRev)
			if err != nil {
				return nil, err
			}
			level.Info(logger).Log("msg", "using found start SHA: "+sha)
			opts.startSHA = sha
		}
		if opts.endRev != "" {
			sha, err := repo.RevParse(opts.endRev)
			if err != nil {
				return nil, err
			}
			level.Info(logger).Log("msg", "using found end SHA: "+sha)
			opts.endSHA = sha
		}
	}

	// Add appropriate log filtering
	if opts.debug {
		logger = level.NewFilter(logger, level.AllowDebug())
	} else {
		logger = level.NewFilter(logger, level.AllowInfo())
	}
	logger = log.With(logger, "timestamp", log.DefaultTimestamp, "caller", log.DefaultCaller)
	opts.logger = logger

	return opts, nil
}

func run(logger log.Logger, args []string) error {
	// Parse the CLI options and enforce required defaults
	opts, err := parseOptions(args, logger)
	if err != nil {
		level.Error(logger).Log("msg", "error parsing options", "err", err)
		return err
	}
	logger = opts.logger

	// get the release notes
	releaseNotes, history, err := opts.GetReleaseNotes()
	if err != nil {
		return err
	}

	err = opts.WriteReleaseNotes(releaseNotes, history)
	if err != nil {
		level.Error(logger).Log("msg", "error writing to file", "err", err)
		return err
	}

	return nil
}

func main() {
	// Use the go-kit structured logger for logging. To learn more about structured
	// logging see: https://github.com/go-kit/kit/tree/master/log#structured-logging
	// https://godoc.org/github.com/go-kit/kit/log/level
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	if err := run(logger, os.Args[1:]); err != nil {
		os.Exit(-1)
	}
}

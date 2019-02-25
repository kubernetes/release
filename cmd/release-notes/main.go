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
	"github.com/google/go-github/github"
	"github.com/kolide/kit/env"
	"github.com/kubernetes/release/pkg/notes"
	"golang.org/x/oauth2"
)

type options struct {
	githubToken string
	output      string
	startSHA    string
	endSHA      string
	relVer      string
	format      string
}

func parseOptions(args []string) (*options, error) {
	flagset := flag.NewFlagSet("release-notes", flag.ExitOnError)
	var (
		// flGitHubToken contains a personal GitHub access token. This is used to
		// scrape the commits of the Kubernetes repo.
		flGitHubToken = flagset.String(
			"github-token",
			env.String("GITHUB_TOKEN", ""),
			"A personal GitHub access token (required)",
		)

		// flOutput contains the path on the filesystem to where the resultant
		// release notes should be printed.
		flOutput = flagset.String(
			"output",
			env.String("OUTPUT", ""),
			"The path to the where the release notes will be printed",
		)

		// flStartSHA contains the commit SHA where the release note generation
		// begins.
		flStartSHA = flagset.String(
			"start-sha",
			env.String("START_SHA", ""),
			"The commit hash to start at",
		)

		// flEndSHA contains the commit SHA where the release note generation ends.
		flEndSHA = flagset.String(
			"end-sha",
			env.String("END_SHA", ""),
			"The commit hash to end at",
		)

		// flRelVer contains the commit SHA where the release note generation ends.
		flRelVer = flagset.String(
			"release-version",
			env.String("RELEASE_VERSION", ""),
			"The release version to generate notes for. e.g. `1.14`",
		)

		// flFormat is the output format to produce the notes in.
		flFormat = flagset.String(
			"format",
			env.String("FORMAT", "markdown"),
			"The format for notes output (options: markdown, json)",
		)
	)

	// Parse the args.
	if err := flagset.Parse(args); err != nil {
		return nil, err
	}

	// The GitHub Token is required.
	if *flGitHubToken == "" {
		return nil, errors.New("GitHub token must be set via -github-token or $GITHUB_TOKEN")
	}

	// The start SHA is required.
	if *flStartSHA == "" {
		return nil, errors.New("The starting commit hash must be set via -start-sha or $START_SHA")
	}

	// The end SHA is required.
	if *flEndSHA == "" {
		return nil, errors.New("The ending commit hash must be set via -end-sha or $END_SHA")
	}

	return &options{
		githubToken: *flGitHubToken,
		output:      *flOutput,
		startSHA:    *flStartSHA,
		endSHA:      *flEndSHA,
		relVer:      *flRelVer,
		format:      *flFormat,
	}, nil
}

func main() {
	// Use the go-kit structured logger for logging. To learn more about structured
	// logging see: https://github.com/go-kit/kit/tree/master/log#structured-logging
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewInjector(logger, level.DebugValue())

	// Parse the CLI options and enforce required defaults
	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		level.Error(logger).Log("msg", "error parsing options", "err", err)
		os.Exit(1)
	}

	// Create the GitHub API client
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: opts.githubToken},
	))
	githubClient := github.NewClient(httpClient)

	// Fetch a list of fully-contextualized release notes
	level.Info(logger).Log("msg", "fetching all commits. this might take a while...")
	releaseNotes, err := notes.ListReleaseNotes(githubClient, logger, opts.startSHA, opts.endSHA, opts.relVer, notes.WithContext(ctx))
	if err != nil {
		level.Error(logger).Log("msg", "error generating release notes", "err", err)
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "got the commits, performing rendering")

	// Open a handle to the file which will contain the release notes output
	var output *os.File
	if opts.output != "" {
		output, err = os.Open(opts.output)
		if err != nil {
			level.Error(logger).Log("msg", "error opening the supplied output file", "err", err)
			os.Exit(1)
		}
	} else {
		output, err = ioutil.TempFile("", "release-notes-")
		if err != nil {
			level.Error(logger).Log("msg", "error creating a temporary file to write the release notes to", "err", err)
			os.Exit(1)
		}
	}

	// Contextualized release notes can be printed in a variety of formats
	switch opts.format {
	case "json":
		enc := json.NewEncoder(output)
		enc.SetIndent("", "  ")
		if err := enc.Encode(releaseNotes); err != nil {
			level.Error(logger).Log("msg", "error encoding JSON output", "err", err)
			os.Exit(1)
		}
	case "markdown":
		doc, err := notes.CreateDocument(releaseNotes)
		if err != nil {
			level.Error(logger).Log("msg", "error creating release note document", "err", err)
			os.Exit(1)
		}

		if err := notes.RenderMarkdown(doc, output); err != nil {
			level.Error(logger).Log("msg", "error rendering release note document to markdown", "err", err)
			os.Exit(1)
		}

	default:
		level.Error(logger).Log("msg", fmt.Sprintf("%q is an unsupported format", opts.format))
		os.Exit(1)
	}

	level.Info(logger).Log(
		"msg", "release notes written to file",
		"path", output.Name(),
		"format", opts.format,
	)
}

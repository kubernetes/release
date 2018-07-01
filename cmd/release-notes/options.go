package main

import (
	"errors"
	"flag"

	"github.com/kolide/kit/env"
)

type options struct {
	githubToken string
	output      string
	startSHA    string
	endSHA      string
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

		// flFormat is the output format to produce the notes in.
		flFormat = flagset.String(
			"format",
			env.String("FORMAT", "json"),
			"The format for notes output (options: json)",
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
		format:      *flFormat,
	}, nil
}

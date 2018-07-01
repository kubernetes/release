package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/github"
	"github.com/kolide/kit/env"
	"golang.org/x/oauth2"

	"k8s.io/release/pkg/githubutil"
)

func main() {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewInjector(logger, level.DebugValue())
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

		flStartSHA = flagset.String(
			"start-sha",
			env.String("START_SHA", ""),
			"The commit hash to start at",
		)

		flEndSHA = flagset.String(
			"end-sha",
			env.String("END_SHA", ""),
			"The commit hash to end at",
		)
	)
	if err := flagset.Parse(os.Args[1:]); err != nil {
		fmt.Println("Error parsing flags:", err)
		os.Exit(1)
	}

	// The GitHub Token is required.
	if *flGitHubToken == "" {
		level.Error(logger).Log("msg", "GitHub token must be set via -github-token or $GITHUB_TOKEN")
		os.Exit(1)
	}

	if *flStartSHA == "" {
		level.Error(logger).Log("msg", "The starting commit hash must be set via -start-sha or $START_SHA")
		os.Exit(1)
	}

	if *flEndSHA == "" {
		level.Error(logger).Log("msg", "The ending commit hash must be set via -end-sha or $END_SHA")
		os.Exit(1)
	}

	// Open a handle to the file which will contain the release notes output
	var output *os.File
	var err error
	if *flOutput != "" {
		output, err = os.Open(*flOutput)
		if err != nil {
			level.Error(logger).Log("msg", "error opening the supplied output file", "err", err)
			os.Exit(1)
		}
	} else {
		// TODO(marpaia): change the second argument of this function invocation to:
		// "release-notes-*.md" after Go 1.11: https://github.com/golang/go/issues/4896
		output, err = ioutil.TempFile("", "release-notes-*")
		if err != nil {
			level.Error(logger).Log("msg", "error creating a temporary file to write the release notes to", "err", err)
			os.Exit(1)
		}
	}

	level.Debug(logger).Log("msg", "successfully opened file", "name", output.Name())

	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *flGitHubToken},
	))
	githubClient := github.NewClient(httpClient)

	notes, err := githubutil.ListReleaseNotes(githubClient, *flStartSHA, *flEndSHA, githubutil.WithContext(ctx))
	if err != nil {
		level.Error(logger).Log("msg", "error release notes", "err", err)
		os.Exit(1)
	}

	level.Debug(logger).Log("msg", "found release notes", "count", len(notes))

	for _, note := range notes {
		fmt.Println("==============================================")
		spew.Dump(note)
	}
}

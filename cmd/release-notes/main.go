package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"k8s.io/release/pkg/notes"
)

func main() {
	// use the go-kit structured logger for logging. To learn more about structured
	// logging see: https://github.com/go-kit/kit/tree/master/log#structured-logging
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewInjector(logger, level.DebugValue())

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

	// Fetch a list of fully-contextualized release notes.
	level.Info(logger).Log("msg", "fetching all commits. this might take a while...")
	notes, err := notes.ListReleaseNotes(githubClient, opts.startSHA, opts.endSHA, notes.WithContext(ctx))
	if err != nil {
		level.Error(logger).Log("msg", "error release notes", "err", err)
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

	// Contextualized release notes can be printed in a variety of formats. Right
	// now only JSON is supported, but a markdown format would be nice as well.
	switch opts.format {
	case "json":
		enc := json.NewEncoder(output)
		enc.SetIndent("", "  ")
		if err := enc.Encode(notes); err != nil {
			level.Error(logger).Log("msg", "error encoding JSON output", "err", err)
			os.Exit(1)
		}

		level.Info(logger).Log("msg", "release notes JSON written to file", "path", output.Name())
	default:
		level.Error(logger).Log("msg", fmt.Sprintf("%q is an unsupported format", opts.format))
		os.Exit(1)
	}
}

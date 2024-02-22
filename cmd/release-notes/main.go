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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/document"
	"k8s.io/release/pkg/notes/options"
	"sigs.k8s.io/mdtoc/pkg/mdtoc"
	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/log"
	"sigs.k8s.io/release-utils/version"
)

type releaseNotesOptions struct {
	outputFile      string
	tableOfContents bool
	dependencies    bool
}

var (
	releaseNotesOpts = &releaseNotesOptions{}
	opts             = options.New()
)

func WriteReleaseNotes(releaseNotes *notes.ReleaseNotes) (err error) {
	logrus.Infof(
		"Got %d release notes, performing rendering",
		len(releaseNotes.History()),
	)

	var (
		// Open a handle to the file which will contain the release notes output
		output        *os.File
		existingNotes notes.ReleaseNotesByPR
	)

	if releaseNotesOpts.outputFile != "" {
		output, err = os.OpenFile(releaseNotesOpts.outputFile, os.O_RDWR|os.O_CREATE, os.FileMode(0o644))
		if err != nil {
			return fmt.Errorf("opening the supplied output file: %w", err)
		}
	} else {
		output, err = os.CreateTemp("", "release-notes-")
		if err != nil {
			return fmt.Errorf("creating a temporary file to write the release notes to: %w", err)
		}
	}

	// Contextualized release notes can be printed in a variety of formats
	if opts.Format == options.FormatJSON {
		byteValue, err := io.ReadAll(output)
		if err != nil {
			return err
		}

		if len(byteValue) > 0 {
			if err := json.Unmarshal(byteValue, &existingNotes); err != nil {
				return fmt.Errorf("unmarshalling existing notes: %w", err)
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
				pr := existingNotes[i].PrNumber
				if releaseNotes.Get(pr) == nil {
					releaseNotes.Set(pr, existingNotes[i])
				}
			}
		}

		enc := json.NewEncoder(output)
		enc.SetIndent("", "  ")
		if err := enc.Encode(releaseNotes.ByPR()); err != nil {
			return fmt.Errorf("encoding JSON output: %w", err)
		}
	} else {
		doc, err := document.New(releaseNotes, opts.StartRev, opts.EndRev)
		if err != nil {
			return fmt.Errorf("creating release note document: %w", err)
		}

		markdown, err := doc.RenderMarkdownTemplate(opts.ReleaseBucket, opts.ReleaseTars, "", opts.GoTemplate)
		if err != nil {
			return fmt.Errorf("rendering release note document with template: %w", err)
		}

		const nl = "\n"
		if releaseNotesOpts.dependencies {
			if opts.StartSHA == opts.EndSHA {
				logrus.Info("Skipping dependency report because start and end SHA are the same")
			} else {
				url := git.GetRepoURL(opts.GithubOrg, opts.GithubRepo, false)
				deps, err := notes.NewDependencies().ChangesForURL(
					url, opts.StartSHA, opts.EndSHA,
				)
				if err != nil {
					return fmt.Errorf("generating dependency report: %w", err)
				}
				markdown += strings.Repeat(nl, 2) + deps
			}
		}

		if releaseNotesOpts.tableOfContents {
			toc, err := mdtoc.GenerateTOC([]byte(markdown), mdtoc.Options{
				Dryrun:     false,
				SkipPrefix: false,
				MaxDepth:   mdtoc.MaxHeaderDepth,
			})
			if err != nil {
				return fmt.Errorf("generating table of contents: %w", err)
			}
			markdown = toc + nl + markdown
		}

		if _, err := output.WriteString(markdown); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
	}

	logrus.Infof("Release notes written to file: %s", output.Name())
	return nil
}

// hackDefaultSubcommand is a utility function that hacks the "generate"
// subcommand as default to avoid breaking compatibility with previoud
// versions of release-notes.
func hackDefaultSubcommand(cmd *cobra.Command) {
	if len(os.Args) > 1 {
		if os.Args[1] == "completion" {
			return
		}

		// Check if the first arg corresponds to a registered subcommand
		for _, command := range cmd.Commands() {
			if command.Use == os.Args[1] {
				return
			}
		}
	}

	logrus.Warn("No subcommand specified, running \"generate\" ")
	os.Args = append([]string{os.Args[0], "generate"}, os.Args[1:]...)
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	logrus.AddHook(log.NewFilenameHook())

	cmd := &cobra.Command{
		Short:         "release-notes - The Kubernetes Release Notes Generator",
		Use:           "release-notes",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	addGenerate(cmd)
	addCheckPR(cmd)

	cmd.AddCommand(version.WithFont("slant"))

	hackDefaultSubcommand(cmd)

	if err := cmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

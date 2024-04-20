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
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
	"k8s.io/release/pkg/announce/github"
	"k8s.io/release/pkg/announce/sbom"
)

// releaseNotesCmd represents the subcommand for `krel release-notes`
var githubPageCmd = &cobra.Command{
	Use:   "github",
	Short: "Updates the github page of a release",
	Long: `publish-release github

This command updates the GitHub release page for a given tag. It will
update the page using a built in template or you can update it using
a custom template.

Before updating the page, the tag has to exist already on github.

To publish the page, --nomock has to be defined. Otherwise, the rendered
page will be printed to stdout and the program will exit.

CUSTOM TEMPLATES
================
You can define a custom golang template to use in your release page. Your
template can contain string substitutions and you can define those using 
the --substitution flag:

  --substitution="releaseTheme:Accentuate the Paw-sitive"
  --substitution="releaseLogo:accentuate-the-pawsitive.png"

ASSET FILES
===========
This command supports uploading release assets to the github page. You
can add asset files with the --asset flag:

  --asset=_output/kubernetes-1.18.2-2.fc33.x86_64.rpm

You can also specify a label for the assets by appending it with a colon
to the asset file:

  --asset="_output/kubernetes-1.18.2-2.fc33.x86_64.rpm:RPM Package for amd64"

Assets can be read from Google Cloud buckets using ambient credentials.
Simply point the asset flag to an object in a bucket instead of a file path:

  --asset="gs://kubernetes-release/release/v1.25.1/bin/linux/amd64/kubectl"

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Run the PR creation function
		return runGithubPage(ghPageOpts)
	},
}

type githubPageCmdLineOptions struct {
	noupdate         bool
	draft            bool
	sbom             bool
	sbomFormat       string
	name             string
	repo             string
	template         string
	repoPath         string
	ReleaseNotesFile string
	substitutions    []string
	assets           []string
}

var ghPageOpts = &githubPageCmdLineOptions{}

func init() {
	githubPageCmd.PersistentFlags().StringVarP(
		&ghPageOpts.repo,
		"repo",
		"r",
		"",
		"repository slug containing the release page",
	)
	githubPageCmd.PersistentFlags().StringVar(
		&ghPageOpts.template,
		"template",
		"",
		"path to a custom page template",
	)
	githubPageCmd.PersistentFlags().StringVarP(
		&ghPageOpts.name,
		"name",
		"n",
		"",
		"name for the release",
	)
	githubPageCmd.PersistentFlags().StringSliceVarP(
		&ghPageOpts.assets,
		"asset",
		"a",
		[]string{},
		"Path to asset file for the release. Can be specified multiple times.",
	)
	githubPageCmd.PersistentFlags().StringSliceVarP(
		&ghPageOpts.substitutions,
		"substitution",
		"s",
		[]string{},
		"String substitution for the page template",
	)
	githubPageCmd.PersistentFlags().BoolVar(
		&ghPageOpts.noupdate,
		"noupdate",
		false,
		"Fail if the release already exists",
	)
	githubPageCmd.PersistentFlags().BoolVar(
		&ghPageOpts.draft,
		"draft",
		false,
		"Mark the release as a draft in GitHub so you can finish editing and publish it manually.",
	)
	githubPageCmd.PersistentFlags().BoolVar(
		&ghPageOpts.sbom,
		"sbom",
		true,
		"Generate an SPDX bill of materials and attach it to the release",
	)
	githubPageCmd.PersistentFlags().StringVar(
		&ghPageOpts.sbomFormat,
		"sbom-format",
		string(sbom.FormatJSON),
		"format to use for the SBOM [json|tag-value]",
	)
	githubPageCmd.PersistentFlags().StringVar(
		&ghPageOpts.repoPath,
		"repo-path",
		".",
		"Path to the source code repository",
	)
	githubPageCmd.PersistentFlags().StringVar(
		&ghPageOpts.ReleaseNotesFile,
		"release-notes-file",
		"",
		"Path to a release notes markdown file to include in the release",
	)

	for _, f := range []string{"template", "asset"} {
		if err := githubPageCmd.MarkPersistentFlagFilename(f); err != nil {
			logrus.Error(err)
		}
	}

	if err := githubPageCmd.MarkPersistentFlagRequired("repo"); err != nil {
		logrus.Error(err)
	}

	rootCmd.AddCommand(githubPageCmd)
}

func getAssetsFromStrings(assetStrings []string) ([]sbom.Asset, error) {
	r := []sbom.Asset{}
	var isBucket bool
	for _, s := range assetStrings {
		isBucket = false
		if strings.HasPrefix(s, "gs:") {
			s = strings.TrimPrefix(s, "gs:")
			isBucket = true
		}
		parts := strings.Split(s, ":")
		l := ""
		if len(parts) > 1 {
			l = parts[1]
		}

		if isBucket {
			path, err := processRemoteAsset("gs:" + parts[0])
			if err != nil {
				return nil, fmt.Errorf("downloading remote asset: %w", err)
			}
			parts[0] = path
		}
		r = append(r, sbom.Asset{
			Path:     filepath.Base(parts[0]),
			ReadFrom: parts[0],
			Label:    l,
		})
	}

	return r, nil
}

// processRemoteAsset gets an object from a bucket and gets it ready for upload
// as an asset of the github release
func processRemoteAsset(urlString string) (path string, err error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return path, fmt.Errorf("parsing URL: %w", err)
	}
	if u.Scheme != "gs" {
		return path, errors.New("only GCS objects are supported at this time")
	}

	filename := filepath.Base(u.Path)
	if filename == "" {
		return path, errors.New("unable to parse filename from path")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return path, fmt.Errorf("creating storage client: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "publish-release-asset-")
	if err != nil {
		return path, fmt.Errorf("creating temp directory: %w", err)
	}

	rc, err := client.Bucket(u.Hostname()).Object(strings.TrimPrefix(u.Path, "/")).NewReader(ctx)
	if err != nil {
		return path, fmt.Errorf("creating bucket reader: %w", err)
	}
	defer rc.Close()

	f, err := os.Create(filepath.Join(tmpDir, filename))
	if err != nil {
		return path, fmt.Errorf("creating temporary file: %w", err)
	}

	if _, err := io.Copy(f, rc); err != nil {
		return path, fmt.Errorf("copying data: %w", err)
	}

	if err = f.Close(); err != nil {
		return path, fmt.Errorf("closing file: %w", err)
	}

	return filepath.Join(tmpDir, filename), nil
}

func runGithubPage(opts *githubPageCmdLineOptions) (err error) {
	// Generate the release SBOM
	assets, err := getAssetsFromStrings(opts.assets)
	if err != nil {
		return fmt.Errorf("getting assets: %w", err)
	}
	sbomStr := ""
	if opts.sbom {
		// Generate the assets file
		sbomStr, err = sbom.NewSBOM(&sbom.Options{
			ReleaseName:   opts.name,
			Repo:          opts.repo,
			RepoDirectory: opts.repoPath,
			Assets:        assets,
			Tag:           commandLineOpts.tag,
			Format:        sbom.SBOMFormat(opts.sbomFormat),
		}).Generate()
		if err != nil {
			return fmt.Errorf("generating sbom: %w", err)
		}
		opts.assets = append(opts.assets, sbomStr+":SPDX Software Bill of Materials (SBOM)")
		// Delete the temporary sbom  when we're done
		if commandLineOpts.nomock {
			defer os.Remove(sbomStr)
		}
	}

	newAssets := make([]string, len(assets)+1)
	for i, a := range assets {
		newAssets[i] = a.ReadFrom
	}

	// add sbom to the path to upload
	newAssets[len(assets)] = sbomStr

	// Build the release page options
	ghOpts := github.Options{
		AssetFiles:            newAssets,
		Tag:                   commandLineOpts.tag,
		NoMock:                commandLineOpts.nomock,
		UpdateIfReleaseExists: !opts.noupdate,
		Name:                  opts.name,
		Draft:                 opts.draft,
		ReleaseNotesFile:      opts.ReleaseNotesFile,
	}

	// Assign the repository data
	if err := ghOpts.SetRepository(opts.repo); err != nil {
		return fmt.Errorf("assigning the repository slug: %w", err)
	}

	// Assign the substitutions
	if err := ghOpts.ParseSubstitutions(opts.substitutions); err != nil {
		return fmt.Errorf("parsing template substitutions: %w", err)
	}

	// Read the csutom template data
	if err := ghOpts.ReadTemplate(opts.template); err != nil {
		return fmt.Errorf("reading the template file: %w", err)
	}

	// Validate the options
	if err := ghOpts.Validate(); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}

	// Run the update process
	return github.NewGitHub(&ghOpts).UpdateGitHubPage()
}

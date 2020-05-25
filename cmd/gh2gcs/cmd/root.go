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
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/gcp"
	"k8s.io/release/pkg/gh2gcs"
	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/log"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "gh2gcs --org kubernetes --repo release [--tag v0.0.0]",
	Short:             "gh2gcs uploads GitHub releases to Google Cloud Storage",
	Example:           "gh2gcs --org kubernetes --repo release --tag v0.0.0",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
	RunE: func(*cobra.Command, []string) error {
		return run(opts)
	},
}

type options struct {
	org                string
	repo               string
	tag                string
	includePrereleases bool
	bucket             string
	releaseDir         string
	outputDir          string
	logLevel           string
}

var opts = &options{}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&opts.org,
		"org",
		// TODO: Remove test value
		"containernetworking",
		// TODO: Improve usage text
		"org",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.repo,
		"repo",
		// TODO: Remove test value
		"plugins",
		// TODO: Improve usage text
		"repo",
	)

	// TODO: This should be a string array to accept multiple tags
	rootCmd.PersistentFlags().StringVar(
		&opts.tag,
		"tag",
		"",
		// TODO: Improve usage text
		"tag",
	)

	rootCmd.PersistentFlags().BoolVar(
		&opts.includePrereleases,
		"include-prereleases",
		false,
		// TODO: Improve usage text
		"include-prereleases",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.bucket,
		"bucket",
		// TODO: Remove test value
		"k8s-staging-release-test",
		// TODO: Improve usage text
		"bucket",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.releaseDir,
		"release-dir",
		// TODO: Remove test value
		"augustus/release",
		// TODO: Improve usage text
		"release-dir",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.outputDir,
		"output-dir",
		"",
		// TODO: Improve usage text
		"output-dir",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.logLevel,
		"log-level",
		"info",
		"the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace'",
	)
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(opts.logLevel)
}

func run(opts *options) error {
	if err := gcp.PreCheck(); err != nil {
		return errors.Wrap(err, "pre-checking for GCP package usage")
	}

	if opts.outputDir == "" {
		tmpDir, err := ioutil.TempDir("", "gh2gcs")
		if err != nil {
			return err
		}

		opts.outputDir = tmpDir
	}

	// Create a real GitHub API client
	gh := github.New()

	// TODO: Support downloading releases via yaml config
	uploadConfig := &gh2gcs.Config{}
	releaseConfig := &gh2gcs.ReleaseConfig{
		Org:        opts.org,
		Repo:       opts.repo,
		Tags:       []string{},
		GCSBucket:  opts.bucket,
		ReleaseDir: opts.releaseDir,
	}

	if opts.tag != "" {
		releaseConfig.Tags = append(releaseConfig.Tags, opts.tag)
	} else {
		releaseTags, err := gh.GetReleaseTags(opts.org, opts.repo, opts.includePrereleases)
		if err != nil {
			return err
		}

		releaseConfig.Tags = releaseTags
	}

	uploadConfig.ReleaseConfigs = append(uploadConfig.ReleaseConfigs, *releaseConfig)

	for _, rc := range uploadConfig.ReleaseConfigs {
		if err := gh2gcs.DownloadReleases(&rc, gh, opts.outputDir); err != nil {
			return err
		}

		if err := gh2gcs.Upload(&rc, gh, opts.outputDir); err != nil {
			return err
		}
	}

	return nil
}

// Validate verifies if all set options are valid
func (o *options) Validate() error {
	// TODO: Add validation logic for options
	return nil
}

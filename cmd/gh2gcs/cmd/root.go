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
	Use:               "gh2gcs --org kubernetes --repo release --bucket <bucket> --release-dir <release-dir> [--tag v0.0.0] [--include-prereleases] [--output-dir <temp-dir>]",
	Short:             "gh2gcs uploads GitHub releases to Google Cloud Storage",
	Example:           "gh2gcs --org kubernetes --repo release --bucket k8s-staging-release-test --release-dir release --tag v0.0.0",
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
		"GitHub org/user",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.repo,
		"repo",
		// TODO: Remove test value
		"plugins",
		"GitHub repo",
	)

	// TODO: This should be a string array to accept multiple tags
	rootCmd.PersistentFlags().StringVar(
		&opts.tag,
		"tag",
		"",
		"release tag to upload to GCS",
	)

	rootCmd.PersistentFlags().BoolVar(
		&opts.includePrereleases,
		"include-prereleases",
		// TODO: Need to wire this
		false,
		"specifies whether prerelease assets should be uploaded to GCS",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.bucket,
		"bucket",
		// TODO: Remove test value
		"k8s-staging-release-test",
		"GCS bucket to upload to",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.releaseDir,
		"release-dir",
		// TODO: Remove test value
		"augustus/release",
		"directory to upload to within the specified GCS bucket",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.outputDir,
		"output-dir",
		"",
		"local directory for releases to be downloaded to",
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
	if err := opts.SetAndValidate(); err != nil {
		return errors.Wrap(err, "validating gh2gcs options")
	}

	if err := gcp.PreCheck(); err != nil {
		return errors.Wrap(err, "pre-checking for GCP package usage")
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
			return errors.Wrap(err, "getting release tags")
		}

		releaseConfig.Tags = releaseTags
	}

	uploadConfig.ReleaseConfigs = append(uploadConfig.ReleaseConfigs, *releaseConfig)

	for _, rc := range uploadConfig.ReleaseConfigs {
		if err := gh2gcs.DownloadReleases(&rc, gh, opts.outputDir); err != nil {
			return errors.Wrap(err, "downloading release assets")
		}

		if err := gh2gcs.Upload(&rc, gh, opts.outputDir); err != nil {
			return errors.Wrap(err, "uploading release assets to GCS")
		}
	}

	return nil
}

// SetAndValidate sets some default options and verifies if options are valid
func (o *options) SetAndValidate() error {
	logrus.Info("Validating gh2gcs options...")

	// TODO: Temp dir should cleanup after itself
	if o.outputDir == "" {
		tmpDir, err := ioutil.TempDir("", "gh2gcs")
		if err != nil {
			return errors.Wrap(err, "creating temp directory")
		}

		o.outputDir = tmpDir
	}

	return nil
}

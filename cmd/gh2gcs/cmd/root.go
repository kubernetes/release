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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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
	org         string
	repo        string
	tag         string
	gcsEndpoint string
	outputDir   string
	logLevel    string
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
		"org",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.repo,
		"repo",
		// TODO: Remove test value
		"plugins",
		"repo",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.tag,
		"tag",
		// TODO: Remove test value
		"",
		"tag",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.gcsEndpoint,
		"gcs-endpoint",
		// TODO: Remove test value
		"k8s-staging-release-test/augustus/cni",
		"org",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.outputDir,
		"output-dir",
		"",
		"org",
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
	// TODO: Support single release or multi-release scenarios
	if err := gh.DownloadReleaseAssets(opts.org, opts.repo, opts.tag, opts.outputDir); err != nil {
		return err
	}

	// TODO: Add GCS upload logic

	return nil
}

// Validate verifies if all set options are valid
func (o *options) Validate() error {
	// TODO: Add validation logic for options
	return nil
}

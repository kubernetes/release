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

package cmd

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/gcp/build"
	"sigs.k8s.io/release-utils/log"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "gcbuilder",
	Short:             "gcbuilder",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

type rootOptions struct {
	logLevel string
}

var (
	rootOpts  = &rootOptions{}
	buildOpts = &build.Options{}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&buildOpts.ConfigDir, "config-dir", ".", "Configuration directory")
	rootCmd.PersistentFlags().StringVar(&buildOpts.BuildDir, "build-dir", "", "If provided, this directory will be uploaded as the source for the Google Cloud Build run.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.CloudbuildFile, "gcb-config", build.DefaultCloudbuildFile, "If provided, this will be used as the name of the Google Cloud Build config file.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.LogDir, "log-dir", "", "If provided, build logs will be sent to files in this directory instead of to stdout/stderr.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.ScratchBucket, "scratch-bucket", "", "The complete GCS path for Cloud Build to store scratch files (sources, logs).")
	rootCmd.PersistentFlags().StringVar(&buildOpts.Project, "project", "", "If specified, use a non-default GCP project.")
	rootCmd.PersistentFlags().BoolVar(&buildOpts.AllowDirty, "allow-dirty", false, "If true, allow pushing dirty builds.")
	rootCmd.PersistentFlags().BoolVar(&buildOpts.NoSource, "no-source", false, "If true, no source will be uploaded with this build.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.Variant, "variant", "", "If specified, build only the given variant. An error if no variants are defined.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.EnvPassthrough, "env-passthrough", "", "Comma-separated list of specified environment variables to be passed to GCB as substitutions with an _ prefix. If the variable doesn't exist, the substitution will exist but be empty.")
	rootCmd.PersistentFlags().StringVar(&rootOpts.logLevel, "log-level", "info", fmt.Sprintf("the logging verbosity, either %s", log.LevelNames()))

	buildOpts.ConfigDir = strings.TrimSuffix(buildOpts.ConfigDir, "/")
}

func run() error {
	prepareBuildErr := build.PrepareBuilds(buildOpts)
	if prepareBuildErr != nil {
		return prepareBuildErr
	}

	buildErrors := build.RunBuildJobs(buildOpts)
	if len(buildErrors) != 0 {
		logrus.Fatalf("Failed to run some build jobs: %v", buildErrors)
	}
	logrus.Info("Finished.")

	return nil
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(rootOpts.logLevel)
}

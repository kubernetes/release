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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/gcp/build"
	"k8s.io/release/pkg/log"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "gcbuilder",
	Short:             "gcbuilder",
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
	rootCmd.PersistentFlags().StringVar(&buildOpts.ConfigDir, "config-dir", "", "Configuration directory")
	rootCmd.PersistentFlags().StringVar(&buildOpts.BuildDir, "build-dir", "", "If provided, this directory will be uploaded as the source for the Google Cloud Build run.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.CloudbuildFile, "gcb-config", "cloudbuild.yaml", "If provided, this will be used as the name of the Google Cloud Build config file.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.LogDir, "log-dir", "", "If provided, build logs will be sent to files in this directory instead of to stdout/stderr.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.ScratchBucket, "scratch-bucket", "", "The complete GCS path for Cloud Build to store scratch files (sources, logs).")
	rootCmd.PersistentFlags().StringVar(&buildOpts.Project, "project", "", "If specified, use a non-default GCP project.")
	rootCmd.PersistentFlags().BoolVar(&buildOpts.AllowDirty, "allow-dirty", false, "If true, allow pushing dirty builds.")
	rootCmd.PersistentFlags().BoolVar(&buildOpts.NoSource, "no-source", false, "If true, no source will be uploaded with this build.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.Variant, "variant", "", "If specified, build only the given variant. An error if no variants are defined.")
	rootCmd.PersistentFlags().StringVar(&buildOpts.EnvPassthrough, "env-passthrough", "", "Comma-separated list of specified environment variables to be passed to GCB as substitutions with an _ prefix. If the variable doesn't exist, the substitution will exist but be empty.")
	rootCmd.PersistentFlags().StringVar(&rootOpts.logLevel, "log-level", "info", "the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace'")

	buildOpts.ConfigDir = strings.TrimSuffix(buildOpts.ConfigDir, "/")
}

// TODO: Clean up error handling
func run() error {
	if buildOpts.ConfigDir == "" {
		logrus.Fatal("expected a config directory to be provided")
	}

	if bazelWorkspace := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); bazelWorkspace != "" {
		if err := os.Chdir(bazelWorkspace); err != nil {
			logrus.Fatalf("Failed to chdir to bazel workspace (%s): %v", bazelWorkspace, err)
		}
	}

	if buildOpts.BuildDir == "" {
		buildOpts.BuildDir = buildOpts.ConfigDir
	}

	logrus.Infof("Build directory: %s\n", buildOpts.BuildDir)

	// Canonicalize the config directory to be an absolute path.
	// As we're about to cd into the build directory, we need a consistent way to reference the config files
	// when the config directory is not the same as the build directory.
	absConfigDir, absErr := filepath.Abs(buildOpts.ConfigDir)
	if absErr != nil {
		logrus.Fatalf("Could not resolve absolute path for config directory: %v", absErr)
	}

	buildOpts.ConfigDir = absConfigDir
	buildOpts.CloudbuildFile = path.Join(buildOpts.ConfigDir, buildOpts.CloudbuildFile)

	configDirErr := buildOpts.ValidateConfigDir()
	if configDirErr != nil {
		logrus.Fatalf("Could not validate config directory: %v", configDirErr)
	}

	logrus.Infof("Config directory: %s\n", buildOpts.ConfigDir)

	logrus.Infof("cd-ing to build directory: %s\n", buildOpts.BuildDir)
	if err := os.Chdir(buildOpts.BuildDir); err != nil {
		logrus.Fatalf("Failed to chdir to build directory (%s): %v", buildOpts.BuildDir, err)
	}

	errors := build.RunBuildJobs(buildOpts)
	if len(errors) != 0 {
		logrus.Fatalf("Failed to run some build jobs: %v", errors)
	}
	logrus.Info("Finished.")

	return nil
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(rootOpts.logLevel)
}

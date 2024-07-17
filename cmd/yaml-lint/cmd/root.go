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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/release/pkg/gcp/build"
	"sigs.k8s.io/release-utils/log"
)

var rootCmd = &cobra.Command{
	Short: "yaml-lint → A tool for linting yaml directories",
	Long: `yaml-lint → A tool for linting yaml directories

This tool lets software developers lint yaml directories.

`,
	Use:               "yaml-lint",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

type commandLineOptions struct {
	directory string
}

var commandLineOpts = &commandLineOptions{}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&commandLineOpts.directory,
		"directory",
		"d",
		"",
		"directory path where yaml files to be linted",
	)
}

// Execute builds the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	prepareBuildErr := validateYAMLFilesInDirectory(commandLineOpts.directory)
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

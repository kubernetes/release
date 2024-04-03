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
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/release-utils/log"
	"sigs.k8s.io/yaml"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "schedule-builder --config-path path/to/schedule.yaml --type <release>/or/<patch>[--output-file <filename.md>]",
	Short:             "schedule-builder generate a human readable format of the Kubernetes release schedule",
	Example:           "schedule-builder --config-path /home/user/kubernetes/sig-release/releases/schedule.yaml --type release",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
	RunE: func(*cobra.Command, []string) error {
		return run(opts)
	},
}

type options struct {
	configPath string
	outputFile string
	logLevel   string
	typeFile   string
}

var opts = &options{}

const (
	configPathFlag = "config-path"
	outputFileFlag = "output-file"
	typeFlag       = "type"
	typePatch      = "patch"
	typeRelease    = "release"
)

var requiredFlags = []string{
	configPathFlag,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&opts.configPath,
		configPathFlag,
		"c",
		"",
		"path where can find the schedule.yaml file",
	)

	rootCmd.PersistentFlags().StringVarP(
		&opts.outputFile,
		outputFileFlag,
		"o",
		"",
		"name of the file that save the schedule to. If not set, it will just output to the stdout",
	)

	rootCmd.PersistentFlags().StringVarP(
		&opts.logLevel,
		"log-level",
		"l",
		"info",
		fmt.Sprintf("the logging verbosity, either %s", log.LevelNames()),
	)

	rootCmd.PersistentFlags().StringVarP(
		&opts.typeFile,
		typeFlag,
		"t",
		"patch",
		fmt.Sprintf("type of file to be produced - release cycle schedule or patch schedule. To be set to '%s' or '%s' and respective yaml needs to be supplied with '--%s'", typeRelease, typePatch, configPathFlag),
	)

	for _, flag := range requiredFlags {
		if err := rootCmd.MarkPersistentFlagRequired(flag); err != nil {
			logrus.Fatal(err)
		}
	}
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(opts.logLevel)
}

func run(opts *options) error {
	if err := opts.SetAndValidate(); err != nil {
		return fmt.Errorf("validating schedule-path options: %w", err)
	}

	logrus.Infof("Reading the schedule file %s...", opts.configPath)
	data, err := os.ReadFile(opts.configPath)
	if err != nil {
		return fmt.Errorf("failed to read the file: %w", err)
	}

	var (
		patchSchedule   PatchSchedule
		releaseSchedule ReleaseSchedule
		scheduleOut     string
	)

	logrus.Info("Parsing the schedule...")

	switch opts.typeFile {
	case "patch":
		if err := yaml.UnmarshalStrict(data, &patchSchedule); err != nil {
			return fmt.Errorf("failed to decode the file: %w", err)
		}

		logrus.Info("Generating the markdown output...")
		scheduleOut = parseSchedule(patchSchedule)

	case "release":
		if err := yaml.UnmarshalStrict(data, &releaseSchedule); err != nil {
			return fmt.Errorf("failed to decode the file: %w", err)
		}

		logrus.Info("Generating the markdown output...")
		scheduleOut = parseReleaseSchedule(releaseSchedule)

	default:
		return fmt.Errorf("type must be either %q or %q", typeRelease, typePatch)
	}

	if opts.outputFile != "" {
		logrus.Infof("Saving schedule to a file %s.", opts.outputFile)
		//nolint:gosec // TODO(gosec): G306: Expect WriteFile permissions to be
		// 0600 or less
		err := os.WriteFile(opts.outputFile, []byte(scheduleOut), 0o644)
		if err != nil {
			return fmt.Errorf("failed to save schedule to the file: %w", err)
		}
		logrus.Info("File saved")
	}

	return nil
}

// SetAndValidate sets some default options and verifies if options are valid
func (o *options) SetAndValidate() error {
	logrus.Info("Validating schedule-path options...")

	if o.configPath == "" {
		return fmt.Errorf("need to set the '--%s' flag", configPathFlag)
	}

	return nil
}

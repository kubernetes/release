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
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/release-utils/log"
	"sigs.k8s.io/release-utils/version"
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
	configPath    string
	eolConfigPath string
	outputFile    string
	logLevel      string
	typeFile      string
	update        bool
	version       bool
}

var opts = &options{}

const (
	configPathFlag    = "config-path"
	eolConfigPathFlag = "eol-config-path"
	outputFileFlag    = "output-file"
	typeFlag          = "type"
	updateFlag        = "update"
	versionFlag       = "version"
	typePatch         = "patch"
	typeRelease       = "release"
)

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
		&opts.eolConfigPath,
		eolConfigPathFlag,
		"e",
		"",
		"path where can find the eol.yaml file for updating end of life releases",
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
		"the logging verbosity, either "+log.LevelNames(),
	)

	rootCmd.PersistentFlags().StringVarP(
		&opts.typeFile,
		typeFlag,
		"t",
		"patch",
		fmt.Sprintf("type of file to be produced - release cycle schedule or patch schedule. To be set to '%s' or '%s' and respective yaml needs to be supplied with '--%s'", typeRelease, typePatch, configPathFlag),
	)

	rootCmd.PersistentFlags().BoolVarP(
		&opts.update,
		updateFlag,
		"u",
		false,
		fmt.Sprintf("update the '--%s' based on the latest available data (or date). Right now only supported if '--%s' is set to '%s'", configPathFlag, typeFlag, typePatch),
	)

	rootCmd.PersistentFlags().BoolVarP(
		&opts.version,
		versionFlag,
		"v",
		false,
		"print the version and exit",
	)
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(opts.logLevel)
}

func run(opts *options) error {
	if opts.version {
		info := version.GetVersionInfo()
		fmt.Print(info.String())
		return nil
	}

	if err := opts.SetAndValidate(); err != nil {
		return fmt.Errorf("validating options: %w", err)
	}

	logrus.Infof("Reading schedule file: %s", opts.configPath)
	data, err := os.ReadFile(opts.configPath)
	if err != nil {
		return fmt.Errorf("failed to read the file: %w", err)
	}

	var (
		patchSchedule   PatchSchedule
		releaseSchedule ReleaseSchedule
		eolBranches     EolBranches
		scheduleOut     string
	)

	logrus.Info("Parsing schedule")

	switch opts.typeFile {
	case typePatch:
		if err := yaml.UnmarshalStrict(data, &patchSchedule); err != nil {
			return fmt.Errorf("failed to decode patch schedule: %w", err)
		}

		if opts.eolConfigPath != "" {
			data, err := os.ReadFile(opts.eolConfigPath)
			if err != nil {
				return fmt.Errorf("failed to read end of life config path: %w", err)
			}

			if err := yaml.UnmarshalStrict(data, &eolBranches); err != nil {
				return fmt.Errorf("failed to decode end of life branches: %w", err)
			}
		}

		if opts.update {
			logrus.Info("Updating schedule")
			if err := updatePatchSchedule(
				time.Now(),
				patchSchedule,
				eolBranches,
				opts.configPath,
				opts.eolConfigPath,
			); err != nil {
				return fmt.Errorf("update patch schedule: %w", err)
			}
		} else {
			logrus.Infof("Generating markdown output for type %q", typePatch)
			scheduleOut = parsePatchSchedule(patchSchedule)
			println(scheduleOut)
		}

	case typeRelease:
		if err := yaml.UnmarshalStrict(data, &releaseSchedule); err != nil {
			return fmt.Errorf("failed to decode the file: %w", err)
		}

		logrus.Infof("Generating markdown output for type %q", typeRelease)
		scheduleOut = parseReleaseSchedule(releaseSchedule)
		println(scheduleOut)

	default:
		return fmt.Errorf("type must be either %q or %q", typeRelease, typePatch)
	}

	if opts.outputFile != "" && scheduleOut != "" {
		logrus.Infof("Saving schedule to file: %s", opts.outputFile)
		//nolint:gosec // TODO(gosec): G306: Expect WriteFile permissions to be
		// 0600 or less
		if err := os.WriteFile(opts.outputFile, []byte(scheduleOut), 0o644); err != nil {
			return fmt.Errorf("failed to save schedule to the file: %w", err)
		}
		logrus.Info("File saved")
	}

	return nil
}

// SetAndValidate sets some default options and verifies if options are valid
func (o *options) SetAndValidate() error {
	logrus.Info("Validating options")

	if o.configPath == "" {
		return fmt.Errorf("need to set the '--%s' flag", configPathFlag)
	}

	if o.update && o.typeFile != typePatch {
		return fmt.Errorf("'--%s' is only supported for '--%s=%s', not '%s'", updateFlag, typeFlag, typePatch, o.typeFile)
	}

	return nil
}

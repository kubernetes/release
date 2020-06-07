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
	"io/ioutil"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/log"
	"sigs.k8s.io/yaml"
)

// PatchSchedule main struct to hold the schedules
type PatchSchedule struct {
	Schedules []Schedule `yaml:"schedules"`
}

// PreviousPatches struct to define the old pacth schedules
type PreviousPatches struct {
	Release            string `yaml:"release"`
	CherryPickDeadline string `yaml:"cherryPickDeadline"`
	TargetDate         string `yaml:"targetDate"`
}

// Schedule struct to define the release schedule for a specific version
type Schedule struct {
	Release            string            `yaml:"release"`
	Next               string            `yaml:"next"`
	CherryPickDeadline string            `yaml:"cherryPickDeadline"`
	TargetDate         string            `yaml:"targetDate"`
	EndOfLifeDate      string            `yaml:"endOfLifeDate"`
	PreviousPatches    []PreviousPatches `yaml:"previousPatches"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "schedule-builder --config-path path/to/schedule.yaml [--output-file <filename.md>]",
	Short:             "schedule-builder generate a humam readable format of the Kubernetes release schedule",
	Example:           "schedule-builder --config-path /home/user/kubernetes/sig-release/releases/schedule.yaml",
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
}

var opts = &options{}

const (
	configPathFlag = "config-path"
	outputFileFlag = "output-file"
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
	rootCmd.PersistentFlags().StringVar(
		&opts.configPath,
		configPathFlag,
		"",
		"path where can find the schedule.yaml file",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.outputFile,
		outputFileFlag,
		"",
		"name of the file that save the schedule to. If not set it will just output to the stdout.",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.logLevel,
		"log-level",
		"info",
		"the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace'",
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
		return errors.Wrap(err, "validating schedule-path options")
	}

	logrus.Infof("Reading the schedule file %s...", opts.configPath)
	data, err := ioutil.ReadFile(opts.configPath)
	if err != nil {
		return errors.Wrap(err, "failed to read the file")
	}

	var patchSchedule PatchSchedule

	logrus.Info("Parsing the schedule...")
	err = yaml.UnmarshalStrict(data, &patchSchedule)
	if err != nil {
		return errors.Wrap(err, "failed to decode the file")
	}

	logrus.Info("Generating the markdown output...")

	output := []string{}
	output = append(output, "### Timeline\n")
	for _, releaseSchedule := range patchSchedule.Schedules {
		output = append(output, fmt.Sprintf("### %s\n", releaseSchedule.Release),
			fmt.Sprintf("Next patch release is **%s**\n", releaseSchedule.Next),
			fmt.Sprintf("End of Life for **%s** is **%s**\n", releaseSchedule.Release, releaseSchedule.EndOfLifeDate))

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"Patch Release", "Cherry Pick Deadline", "Target Date"})
		table.Append([]string{strings.TrimSpace(releaseSchedule.Next), strings.TrimSpace(releaseSchedule.CherryPickDeadline), strings.TrimSpace(releaseSchedule.TargetDate)})

		for _, previous := range releaseSchedule.PreviousPatches {
			table.Append([]string{strings.TrimSpace(previous.Release), strings.TrimSpace(previous.CherryPickDeadline), strings.TrimSpace(previous.TargetDate)})
		}
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
		table.Render()

		output = append(output, tableString.String())
	}

	scheduleOut := strings.Join(output, "\n")

	logrus.Info("Schedule parsed")
	println(scheduleOut)

	if opts.outputFile != "" {
		logrus.Infof("Saving schedule to a file %s.", opts.outputFile)
		err := ioutil.WriteFile(opts.outputFile, []byte(scheduleOut), 0644)
		if err != nil {
			return errors.Wrap(err, "failed to save schedule to the file")
		}
		logrus.Info("File saved")
	}

	return nil
}

// SetAndValidate sets some default options and verifies if options are valid
func (o *options) SetAndValidate() error {
	logrus.Info("Validating schedule-path options...")

	if o.configPath == "" {
		return errors.Errorf("need to set the config-path")
	}

	return nil
}

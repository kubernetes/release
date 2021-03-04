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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sigs.k8s.io/release-utils/log"
)

var rootCmd = &cobra.Command{
	Short: "publish-release → A tool for announcing software releases",
	Long: `publish-release → A tool for announcing software releases

This tool lets software developers announce new software releases.

`,
	Use:               "publish-release",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
}

type commandLineOptions struct {
	nomock   bool
	logLevel string
	tag      string
}

var commandLineOpts = &commandLineOptions{}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&commandLineOpts.tag,
		"tag",
		"t",
		"",
		"tag for the release to be used",
	)
	rootCmd.PersistentFlags().BoolVar(
		&commandLineOpts.nomock,
		"nomock",
		false,
		"run in no mock mode, otherwise only prints to stdout",
	)
	rootCmd.PersistentFlags().StringVar(
		&commandLineOpts.logLevel,
		"log-level",
		"info",
		fmt.Sprintf("the logging verbosity, either %s", log.LevelNames()),
	)
}

// Execute builds the command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(commandLineOpts.logLevel)
}

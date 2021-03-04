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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/release-utils/log"
)

const (
	nl = "\n"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "krel",
	Long: `krel - The Kubernetes Release toolbox

krel is the new golang based tool for managing releases. Target of krel is to
provide a toolkit for managing the different steps needed for creating
Kubernetes Releases. This includes manually executed tasks, like the generation
of the release notes during the release cycle, as well as automated tasks like
pushing the Kubernetes release artifacts to the Google Cloud Storage.

Each subcommand should contain its own self describing help output which
clarifies its purpose.`,
	PersistentPreRunE: initLogging,
}

type rootOptions struct {
	nomock   bool
	logLevel string
}

var rootOpts = &rootOptions{}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(
		&rootOpts.nomock,
		"nomock",
		false,
		"run the command to target the production environment",
	)

	rootCmd.PersistentFlags().StringVar(
		&rootOpts.logLevel,
		"log-level",
		"info",
		fmt.Sprintf("the logging verbosity, either %s", log.LevelNames()),
	)
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(rootOpts.logLevel)
}

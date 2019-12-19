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
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "krel",
	Short:   "krel",
	PreRunE: initLogging,
}

type rootOptions struct {
	nomock   bool
	cleanup  bool
	repoPath string
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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVar(&rootOpts.nomock, "nomock", false, "nomock flag")
	rootCmd.PersistentFlags().BoolVar(&rootOpts.cleanup, "cleanup", false, "cleanup flag")
	rootCmd.PersistentFlags().StringVar(&rootOpts.repoPath, "repo", filepath.Join(os.TempDir(), "k8s"), "the local path to the repository to be used")
	rootCmd.PersistentFlags().StringVar(&rootOpts.logLevel, "log-level", "info", "the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace'")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}

func initLogging(*cobra.Command, []string) error {
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	lvl, err := logrus.ParseLevel(rootOpts.logLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	logrus.Debugf("Using log level %q", lvl)
	return nil
}

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

	"k8s.io/release/cmd/kpromo/cmd/manifest"
	"k8s.io/release/cmd/kpromo/cmd/run"
	"sigs.k8s.io/release-utils/log"
)

// rootCmd represents the base command when called without any subcommands
// TODO: Update command description
var rootCmd = &cobra.Command{
	Use: "kpromo",
	Long: `kpromo - Kubernetes project artifact promoter

kpromo is a tool responsible for artifact promotion.

It has two operation modes:
- "run" - Execute a file promotion (formerly "promobot-files") (image promotion coming soon)
- "manifest" - Generate/modify a file manifest to target for promotion (image support coming soon)

Expectations:
- "kpromo run" should only be run in auditable environments
- "kpromo manifest" should primarily be run by contributors 

Each subcommand should contain its own self-describing help output which
clarifies its purpose.`,
	PersistentPreRunE: initLogging,
}

type rootOptions struct {
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
	rootCmd.PersistentFlags().StringVar(
		&rootOpts.logLevel,
		"log-level",
		"info",
		fmt.Sprintf("the logging verbosity, either %s", log.LevelNames()),
	)

	rootCmd.AddCommand(run.RunCmd)
	rootCmd.AddCommand(manifest.ManifestCmd)
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(rootOpts.logLevel)
}

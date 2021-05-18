/*
Copyright 2021 The Kubernetes Authors.

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
	"bytes"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"

	"sigs.k8s.io/release-utils/log"
)

const (
	progName             = "rnd"
	optionsUsageHeader   = "\nConfigurable environment variables with their default values: (required fields marked with (*))"
	optionsUsageTemplate = `{{range .}}
  {{if usage_required .}}(*) {{else}}    {{end}}{{usage_key .}}={{usage_default .}}{{end}}`
)

type rootOptions struct {
	LogLevel string `split_words:"true" default:"info"`
}

var rootOpts = &rootOptions{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   progName,
	Short: "rnd: Release Notes Daemon",
	Long: `rnd: release notes (daemon) with continuous integration in mind

rnd was designed with container-first ergonomics, since
the program expects to be run unattended, such as through CI systems`,
	PersistentPreRunE: func(*cobra.Command, []string) error {
		return processRootFlags()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + optionsUsageHeader + rootCmdOptionsUsage())
}

func rootCmdOptionsUsage() string {
	var optionsUsage bytes.Buffer
	if err := envconfig.Usagef(progName, rootOpts, &optionsUsage, optionsUsageTemplate); err != nil {
		panic(err)
	}
	return optionsUsage.String()
}

func processRootFlags() error {
	err := envconfig.Process(progName, rootOpts)
	if err != nil {
		return err
	}

	err = log.SetupGlobalLogger(rootOpts.LogLevel)
	if err != nil {
		return err
	}

	return nil
}

/*
Copyright 2022 The Kubernetes Authors.

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
	"time"

	"github.com/spf13/cobra"
)

const (
	verboseFlag    = "verbose"
	timeoutFlag    = "timeout"
	maxWorkersFlag = "max-workers"
)

type signOptions struct {
	verbose bool
	timeout time.Duration
	// The amount of maximum workers for parallel executions.
	// Defaults to 100.
	maxWorkers uint
}

var singOpts = &signOptions{}

// signCmd represents the subcommand for `krel sign`
var signCmd = &cobra.Command{
	Use:           "sign",
	Short:         "sign images and blobs",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	signCmd.PersistentFlags().BoolVar(
		&singOpts.verbose,
		verboseFlag,
		false,
		"can be used to enable a higher log verbosity",
	)

	signCmd.PersistentFlags().DurationVarP(
		&singOpts.timeout,
		timeoutFlag,
		"t",
		3*time.Minute,
		"is the default timeout for network operations",
	)

	signCmd.PersistentFlags().UintVar(
		&singOpts.maxWorkers,
		maxWorkersFlag,
		100,
		"The amount of maximum workers for parallel executions",
	)

	rootCmd.AddCommand(signCmd)
}

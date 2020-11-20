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
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/gcp/gcb"
)

var historyOpts = gcb.NewHistoryOptions()

// historyCmd is a krel subcommand which generates information about the
// command that the operator ran for a specific release cut.
var historyCmd = &cobra.Command{
	Use:           "history --branch release-1.19 --date-from 2020-06-18 [--date-to 2020-06-19]",
	Short:         "Run history to build a list of commands that ran when cutting a specific Kubernetes release",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return gcb.NewHistory(historyOpts).Run()
	},
}

func init() {
	historyCmd.PersistentFlags().StringVar(
		&historyOpts.Branch,
		"branch",
		historyOpts.Branch,
		"The release branch for which the release should be build",
	)

	historyCmd.PersistentFlags().StringVar(
		&historyOpts.DateFrom,
		"date-from",
		historyOpts.DateFrom,
		"Get the jobs starting from a specific date.",
	)

	historyCmd.PersistentFlags().StringVar(
		&historyOpts.DateTo,
		"date-to",
		historyOpts.DateTo,
		"Get the jobs ending from a specific date.",
	)

	rootCmd.AddCommand(historyCmd)
}

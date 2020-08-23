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
)

const (
	tagFlag       = "tag"
	printOnlyFlag = "print-only"
)

type announceOptions struct {
	tag       string
	printOnly bool
}

var announceOpts = &announceOptions{}

// announceCmd represents the subcommand for `krel announce`
var announceCmd = &cobra.Command{
	Use:           "announce",
	Short:         "Build and announce Kubernetes releases",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	announceCmd.PersistentFlags().StringVarP(
		&announceOpts.tag,
		tagFlag,
		"t",
		"",
		"built tag to be announced, will be used for fetching the "+
			"announcement from the google cloud bucket and to create the annoucements file",
	)

	announceCmd.PersistentFlags().BoolVarP(
		&announceOpts.printOnly,
		printOnlyFlag,
		"p",
		false,
		"print the mail contents without sending it",
	)

	rootCmd.AddCommand(announceCmd)
}

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
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/release/pkg/spdx"
)

var outlineOpts = &spdx.DrawingOptions{}

var documentCmd = &cobra.Command{
	Short: "bom document → Work with SPDX documents",
	Long: `bom document → Work with SPDX documents",


`,
	Use:               "document",
	SilenceUsage:      false,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
	/*
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("You should only specify one file")
			}
			doc, err := spdx.OpenDoc(args[0])
			if err != nil {
				return errors.Wrap(err, "opening doc")
			}
			output, err := doc.Outline()
			if err != nil {
				return errors.Wrap(err, "generating document outline")
			}
			fmt.Println(output)
			return nil
		},
	*/
	// doc.

}

var outlineCmd = &cobra.Command{
	Short: "bom document outline → Draw structure of a SPDX document",
	Long: `bom document outline → Draw structure of a SPDX document",


`,
	Use:               "outline",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("You should only specify one file")
		}
		doc, err := spdx.OpenDoc(args[0])
		if err != nil {
			return errors.Wrap(err, "opening doc")
		}
		output, err := doc.Outline(outlineOpts)
		if err != nil {
			return errors.Wrap(err, "generating document outline")
		}
		fmt.Println(output)
		return nil
	},
}

func init() {
	outlineCmd.PersistentFlags().IntVarP(
		&outlineOpts.Recursion,
		"depth",
		"d",
		9999,
		"recursion level",
	)

	outlineCmd.PersistentFlags().BoolVar(
		&outlineOpts.OnlyIDs,
		"spdx-ids",
		false,
		"use SPDX identifiers in tree nodes instead of names",
	)

	documentCmd.AddCommand(outlineCmd)
}

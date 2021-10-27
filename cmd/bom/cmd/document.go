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
}

var outlineCmd = &cobra.Command{
	Short: "bom document outline → Draw structure of a SPDX document",
	Long: `bom document outline → Draw structure of a SPDX document",

This subcommand draws a tree-like outline to help the user visualize 
the structure of the bom. Even when an SBOM represents a graph structure, 
drawing a tree helps a lot to understand what is contained in the document.

You can define a level of depth to limit the expansion of the entities.
For example set --depth=1 to only visualize only the files and packages
attached directly to the root of the document. 

bom will try to add useful information to the oultine but, if needed, you can
set the --spdx-ids to only output the IDs of the entities.

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
		fmt.Println(spdx.Banner())
		fmt.Println(output)
		return nil
	},
}

func init() {
	outlineCmd.PersistentFlags().IntVarP(
		&outlineOpts.Recursion,
		"depth",
		"d",
		-1,
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

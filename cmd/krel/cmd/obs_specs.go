/*
Copyright 2023 The Kubernetes Authors.

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

	"k8s.io/release/pkg/obs/specs"
)

var specsOpts = specs.DefaultOptions()

// obsSpecsCmd represents the subcommand for `krel obs specs`
var obsSpecsCmd = &cobra.Command{
	Use:           "specs",
	Short:         "generate specs and artifacts archive",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerateOBSSpecs(specsOpts)
	},
}

func init() {
	obsSpecsCmd.PersistentFlags().StringVar(
		&specsOpts.Package,
		"package",
		specsOpts.Package,
		"package to create specs and archives for",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&specsOpts.Version,
		"version",
		specsOpts.Version,
		"package version",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&specsOpts.Revision,
		"revision",
		specsOpts.Revision,
		"package revision",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&specsOpts.Channel,
		"channel",
		specsOpts.Channel,
		"channel to create specs for",
	)

	obsSpecsCmd.PersistentFlags().StringSliceVar(
		&specsOpts.Architectures,
		"architectures",
		specsOpts.Architectures,
		"architectures to download binaries for when creating the archive",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&specsOpts.PackageSourceBase,
		"package-source",
		specsOpts.PackageSourceBase,
		"source to download artifacts for package",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&specsOpts.SpecTemplatePath,
		"template-dir",
		specsOpts.SpecTemplatePath,
		"template directory containing spec files",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&specsOpts.SpecOutputPath,
		"output",
		specsOpts.SpecOutputPath,
		"output directory to store specs and archives",
	)

	obsSpecsCmd.PersistentFlags().BoolVar(
		&specsOpts.SpecOnly,
		"spec-only",
		specsOpts.SpecOnly,
		"only create specs without downloading binaries and creating archives",
	)

	obsCmd.AddCommand(obsSpecsCmd)
}

func runGenerateOBSSpecs(opts *specs.Options) (err error) {
	logrus.Debugf("Using options: %s", opts.String())

	if err := opts.Validate(); err != nil {
		return fmt.Errorf("running krel obs: %w", err)
	}

	if err := specs.New(opts).Run(); err != nil {
		return fmt.Errorf("running krel obs: %w", err)
	}

	return nil
}

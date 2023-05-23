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

	"k8s.io/release/pkg/obs/options"
	"k8s.io/release/pkg/obs/specs"
)

var obsOpts = options.New()

// obsSpecsCmd represents the subcommand for `krel obs specs`
var obsSpecsCmd = &cobra.Command{
	Use:           "specs",
	Short:         "generate specs and binary archives",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerateOBSSpecs(obsOpts)
	},
}

func init() {
	obsSpecsCmd.PersistentFlags().StringVar(
		&obsOpts.Package,
		"package",
		obsOpts.Package,
		"package to create specs and archives for",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&obsOpts.Channel,
		"channel",
		obsOpts.Channel,
		"channel to create specs for",
	)

	obsSpecsCmd.PersistentFlags().StringSliceVar(
		&obsOpts.Architectures,
		"arch",
		obsOpts.Architectures,
		"architectures to download binaries for when creating the archive",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&obsOpts.Version,
		"version",
		obsOpts.Version,
		"package version",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&obsOpts.Revision,
		"revision",
		obsOpts.Revision,
		"package revision",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&obsOpts.PackageSourceBase,
		"package-source",
		obsOpts.PackageSourceBase,
		"source to download artifacts for package",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&obsOpts.SpecTemplatePath,
		"template-dir",
		obsOpts.SpecTemplatePath,
		"template directory containing spec files",
	)

	obsSpecsCmd.PersistentFlags().StringVar(
		&obsOpts.SpecOutputPath,
		"output",
		obsOpts.SpecOutputPath,
		"output directory to store specs and archives",
	)

	obsSpecsCmd.PersistentFlags().BoolVar(
		&obsOpts.SpecOnly,
		"spec-only",
		obsOpts.SpecOnly,
		"only create specs without downloading binaries and creating archives",
	)

	obsCmd.AddCommand(obsSpecsCmd)
}

func runGenerateOBSSpecs(opts *options.Options) (err error) {
	logrus.Debugf("Using options: %+v", opts)

	if err := opts.Validate(); err != nil {
		return fmt.Errorf("running krel obs: %w", err)
	}

	client := specs.New(opts)

	pkgDef, err := client.ConstructPackageDefinition()
	if err != nil {
		return fmt.Errorf("running krel obs: %w", err)
	}

	if err = client.BuildSpecs(pkgDef, opts.SpecOnly); err != nil {
		return fmt.Errorf("running krel obs: %w", err)
	}

	if !opts.SpecOnly {
		if err = client.BuildArtifactsArchive(pkgDef); err != nil {
			return fmt.Errorf("running krel obs: %w", err)
		}
	}

	logrus.Infof("krel obs done, files available in %s", opts.SpecOutputPath)

	return nil
}

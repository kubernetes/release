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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/obs"
	"k8s.io/release/pkg/release"
)

// obsReleaseCmd represents the subcommand for `krel obs release`.
var obsReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release packages for a new Kubernetes version or Kubernetes subproject",
	Long: fmt.Sprintf(`krel obs release

This subcommand is the second of two necessary steps for building and publishing
Debian and RPM packages. This subcommand works in two modes: publish core
Kubernetes packages (kubeadm, kubelet, kubectl, kubernetes-cni, cri-tools) or
publish packages for a Kubernetes subproject. The first mode is directly
embedded in "krel release" and in which case this command is not supposed to be
used standalone. The second mode is intended to be run by users who want to
submit a Google Cloud Build (GCB) job which does:

1. Check Prerequisites: Verify that a valid %s environment variable is set.
   A basic hardware check will ensure that enough disk space is available, too.

2. Initialize OBS root and configuration: creates the directory structure
   needed for working with OBS and a configuration file which contains
   credentials to authenticate with OBS.

3. Run "osc release" for each package: this triggers a new OBS job that's
   going to publish successful builds to the configured maintenance project.
   Configuration to which project the packages should be published is done via
   OBS UI.
`, obs.OBSPasswordKey),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOBSRelease(obsReleaseOptions)
	},
}

var obsReleaseOptions = obs.DefaultReleaseOptions()

func init() {
	obsReleaseCmd.PersistentFlags().
		StringVar(
			&obsReleaseOptions.Workspace,
			"workspace",
			obsReleaseOptions.Workspace,
			"Workspace directory for running krel obs",
		)

	obsReleaseCmd.PersistentFlags().
		StringVar(
			&obsReleaseOptions.ReleaseType,
			"type",
			obsReleaseOptions.ReleaseType,
			fmt.Sprintf("The release type, must be one of: '%s'",
				strings.Join([]string{
					release.ReleaseTypeAlpha,
					release.ReleaseTypeBeta,
					release.ReleaseTypeRC,
					release.ReleaseTypeOfficial,
				}, "', '"),
			))

	obsReleaseCmd.PersistentFlags().
		StringVar(
			&obsReleaseOptions.ReleaseBranch,
			"branch",
			obsReleaseOptions.ReleaseBranch,
			"The release branch for which the release should be build",
		)

	obsReleaseCmd.PersistentFlags().
		StringVar(
			&obsReleaseOptions.BuildVersion,
			buildVersionFlag,
			"",
			"The build version to be released.",
		)

	obsReleaseCmd.PersistentFlags().
		StringArrayVar(
			&obsReleaseOptions.Packages,
			obsPackagesFlag,
			obsReleaseOptions.Packages,
			"List of packages to build",
		)

	obsReleaseCmd.PersistentFlags().
		StringVar(
			&obsReleaseOptions.SpecTemplatePath,
			obsSpecTemplatePathFlag,
			obsReleaseOptions.SpecTemplatePath,
			"Path to a directory containing templates for specs",
		)

	obsReleaseCmd.PersistentFlags().
		StringVar(
			&obsReleaseOptions.Project,
			obsProjectFlag,
			obsReleaseOptions.Project,
			"OBS project where to publish artifacts/packages",
		)

	obsReleaseCmd.PersistentFlags().
		BoolVar(
			&submitJob,
			submitJobFlag,
			submitJob,
			"Submit build to GCS",
		)

	obsReleaseCmd.PersistentFlags().
		BoolVar(
			&stream,
			streamFlag,
			false,
			"Run the Google Cloud Build job synchronously",
		)

	for _, flag := range []string{buildVersionFlag, submitJobFlag} {
		if err := obsReleaseCmd.PersistentFlags().MarkHidden(flag); err != nil {
			logrus.Fatal(err)
		}
	}

	obsCmd.AddCommand(obsReleaseCmd)
}

func runOBSRelease(options *obs.ReleaseOptions) error {
	options.NoMock = rootOpts.nomock
	obsRelease := obs.NewRelease(options)
	if submitJob {
		// Perform a local check of the specified options before launching a
		// Cloud Build job:
		if err := options.Validate(&obs.State{}, true); err != nil {
			return fmt.Errorf("prechecking release options: %w", err)
		}
		return obsRelease.Submit(stream)
	}
	return obsRelease.Run()
}

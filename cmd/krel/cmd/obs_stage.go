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

	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/obs"
	"k8s.io/release/pkg/release"
)

// obsStageCmd represents the subcommand for `krel obs stage`
var obsStageCmd = &cobra.Command{
	Use:   "stage",
	Short: "Stage packages for a new Kubernetes version or Kubernetes subproject",
	Long: fmt.Sprintf(`krel obs stage

This subcommand is the first of two necessary steps for building and publishing
Debian and RPM packages. This subcommand works in two modes: publish core
Kubernetes packages (kubeadm, kubelet, kubectl, kubernetes-cni, cri-tools) or
publish packages for a Kubernetes subproject. The first mode is directly
embedded in "krel stage" and in which case this command is not supposed to be
used standalone. The second mode is intended to be run by users who want to
submit a Google Cloud Build (GCB) job which does:

1. Check Prerequisites: Verify that a valid %s environment variable is set. It
   also checks for the existence of required spec files. A basic hardware check
   will ensure that enough disk space is available, too.

2. Initialize OBS root and configuration: creates the directory structure
   needed for working with OBS and a configuration file which contains
   credentials to authenticate with OBS.

3. Generate specs and artifacts archive: given specs templates are executed to
   fill information such as version and dependencies. Binaries needed to build
   the package are downloaded and archived.

4. Push artifacts to OBS: generated specs and artifacts archive are pushed to
   the given OBS project/package.
`, obs.OBSPasswordKey),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOBSStage(obsStageOptions)
	},
}

var obsStageOptions = obs.DefaultStageOptions()

const (
	obsBuildVersionFlag     = "build-version"
	obsPackagesFlag         = "packages"
	obsArchitecturesFlag    = "architectures"
	obsSpecTemplatePathFlag = "template-dir"
	obsVersionFlag          = "version"
	obsProjectFlag          = "project"
	obsSourceFlag           = "source"
	obsWaitFlag             = "wait"
)

func init() {
	obsStageCmd.PersistentFlags().
		StringVar(
			&obsStageOptions.Workspace,
			"workspace",
			obsStageOptions.Workspace,
			"Workspace directory for running krel obs",
		)

	obsStageCmd.PersistentFlags().
		StringVar(
			&obsStageOptions.ReleaseType,
			"type",
			obsStageOptions.ReleaseType,
			fmt.Sprintf("The release type, must be one of: '%s'",
				strings.Join([]string{
					release.ReleaseTypeAlpha,
					release.ReleaseTypeBeta,
					release.ReleaseTypeRC,
					release.ReleaseTypeOfficial,
				}, "', '"),
			))

	obsStageCmd.PersistentFlags().
		StringVar(
			&obsStageOptions.ReleaseBranch,
			"branch",
			obsStageOptions.ReleaseBranch,
			"The release branch for which the release should be build",
		)

	obsStageCmd.PersistentFlags().
		StringVar(
			&obsStageOptions.BuildVersion,
			buildVersionFlag,
			"",
			"The build version to be released.",
		)

	obsStageCmd.PersistentFlags().
		StringVar(
			&obsStageOptions.SpecTemplatePath,
			obsSpecTemplatePathFlag,
			obsStageOptions.SpecTemplatePath,
			"Path to a directory containing templates for specs",
		)

	obsStageCmd.PersistentFlags().
		StringSliceVar(
			&obsStageOptions.Packages,
			obsPackagesFlag,
			obsStageOptions.Packages,
			"List of packages to build",
		)

	obsStageCmd.PersistentFlags().
		StringSliceVar(
			&obsStageOptions.Architectures,
			obsArchitecturesFlag,
			obsStageOptions.Architectures,
			"List of architectures to build",
		)

	obsStageCmd.PersistentFlags().
		StringVar(
			&obsStageOptions.Version,
			obsVersionFlag,
			obsStageOptions.Version,
			"Desired version of packages",
		)

	obsStageCmd.PersistentFlags().
		StringVar(
			&obsStageOptions.Project,
			obsProjectFlag,
			obsStageOptions.Project,
			"OBS project where to publish artifacts/packages",
		)

	obsStageCmd.PersistentFlags().
		StringVar(
			&obsStageOptions.PackageSource,
			obsSourceFlag,
			obsStageOptions.PackageSource,
			"HTTPS or GS URL to be used when downloading binaries",
		)

	obsStageCmd.PersistentFlags().
		BoolVar(
			&submitJob,
			submitJobFlag,
			submitJob,
			"Submit build to GCS",
		)

	obsStageCmd.PersistentFlags().
		BoolVar(
			&stream,
			streamFlag,
			false,
			"Run the Google Cloud Build job synchronously",
		)

	obsStageCmd.PersistentFlags().
		BoolVar(
			&obsStageOptions.Wait,
			obsWaitFlag,
			true,
			"Wait for the OBS build results to succeed",
		)

	for _, flag := range []string{buildVersionFlag, submitJobFlag} {
		if err := obsStageCmd.PersistentFlags().MarkHidden(flag); err != nil {
			logrus.Fatal(err)
		}
	}

	obsCmd.AddCommand(obsStageCmd)
}

func runOBSStage(options *obs.StageOptions) error {
	options.NoMock = rootOpts.nomock

	// Allow submitting packages and architectures separated by the string
	// slice separator. This allows us to pass the GCB substitution, which
	// already uses comma as default separator.
	// We cannot use the GCB separator substitution (see `gcloud topic
	// escaping`) because GCB complains that a 'build tag must match format
	// "^[\\w][\\w.-]{0,127}$"'.
	archSplit := strings.Split(options.Architectures[0], gcb.StringSliceSeparator)
	if len(archSplit) > 1 {
		options.Architectures = archSplit
	}

	packageSplit := strings.Split(options.Packages[0], gcb.StringSliceSeparator)
	if len(packageSplit) > 1 {
		options.Packages = packageSplit
	}

	stage := obs.NewStage(options)
	if submitJob {
		// Perform a local check of the specified options before launching a
		// Cloud Build job:
		if err := options.Validate(&obs.State{}, true); err != nil {
			return fmt.Errorf("prechecking stage options: %w", err)
		}
		return stage.Submit(stream)
	}
	return stage.Run()
}

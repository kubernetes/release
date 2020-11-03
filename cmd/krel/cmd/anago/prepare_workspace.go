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

package anago

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/release/pkg/release"
)

// prepareWorkspaceCmd represents the subcommand for `krel anago prepare-workspace`
var prepareWorkspaceCmd = &cobra.Command{
	Use:           "prepare-workspace",
	Short:         "Prepare the workspace for cutting releases",
	Long:          "krel anago prepare-workspace",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.Wrap(
			runPrepareWorkspace(prepareWorkspaceOpts),
			"run krel anago prepare-workspace",
		)
	},
}

type prepareWorkspaceOptions struct {
	stage        bool
	directory    string
	bucket       string
	buildVersion string
}

var prepareWorkspaceOpts = &prepareWorkspaceOptions{}

func init() {
	prepareWorkspaceCmd.PersistentFlags().BoolVar(
		&prepareWorkspaceOpts.stage,
		"stage",
		false,
		"run in stage mode, otherwise release",
	)

	prepareWorkspaceCmd.PersistentFlags().StringVar(
		&prepareWorkspaceOpts.directory,
		"directory",
		"",
		"root directory of the target workspace",
	)

	prepareWorkspaceCmd.PersistentFlags().StringVar(
		&prepareWorkspaceOpts.bucket,
		"bucket",
		"",
		"GCS bucket to be used",
	)

	prepareWorkspaceCmd.PersistentFlags().StringVar(
		&prepareWorkspaceOpts.buildVersion,
		"build-version",
		"",
		"build version for the release",
	)

	AnagoCmd.AddCommand(prepareWorkspaceCmd)
}

func runPrepareWorkspace(opts *prepareWorkspaceOptions) error {
	if opts.stage {
		return release.PrepareWorkspaceStage(opts.directory)
	}
	return release.PrepareWorkspaceRelease(
		opts.directory, opts.buildVersion, opts.bucket,
	)
}

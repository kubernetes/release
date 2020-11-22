/*
Copyright 2019 The Kubernetes Authors.

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

package run

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/promobot"
)

// filesCmd represents the subcommand for `kpromo run files`
var filesCmd = &cobra.Command{
	Use:           "files",
	Short:         "Promote files from a staging object store to production",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.Wrap(runFilePromotion(filesOpts), "run `kpromo run files`")
	},
}

var filesOpts = &promobot.PromoteFilesOptions{}

func init() {
	// TODO: Move this into a default options function in pkg/promobot
	filesOpts.PopulateDefaults()

	filesCmd.PersistentFlags().StringVar(
		&filesOpts.FilestoresPath,
		"filestores",
		filesOpts.FilestoresPath,
		"path to the `filestores` promoter manifest",
	)

	filesCmd.PersistentFlags().StringVar(
		&filesOpts.FilesPath,
		"files",
		filesOpts.FilesPath,
		"path to the `files` manifest",
	)

	// TODO: Consider moving this to the root command
	filesCmd.PersistentFlags().BoolVar(
		&filesOpts.DryRun,
		"dry-run",
		filesOpts.DryRun,
		"test run promotion without modifying any filestore",
	)

	filesCmd.PersistentFlags().BoolVar(
		&filesOpts.UseServiceAccount,
		"use-service-account",
		filesOpts.UseServiceAccount,
		"allow service account usage with gcloud calls",
	)

	// TODO: Consider moving this into a validation function
	// nolint: errcheck
	filesCmd.MarkPersistentFlagRequired("filestores")
	// nolint: errcheck
	filesCmd.MarkPersistentFlagRequired("files")

	RunCmd.AddCommand(filesCmd)
}

func runFilePromotion(opts *promobot.PromoteFilesOptions) error {
	ctx := context.Background()

	return promobot.RunPromoteFiles(ctx, *opts)
}

// TODO: Validate options

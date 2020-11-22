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

package manifest

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/promobot"
	"sigs.k8s.io/yaml"
)

// filesCmd represents the subcommand for `kpromo manifest files`
var filesCmd = &cobra.Command{
	Use:           "files",
	Short:         "Promote files from a staging object store to production",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.Wrap(runFileManifest(filesOpts), "run `kpromo manifest files`")
	},
}

var filesOpts = &promobot.GenerateManifestOptions{}

func init() {
	// TODO: Move this into a default options function in pkg/promobot
	filesOpts.PopulateDefaults()

	filesCmd.PersistentFlags().StringVar(
		&filesOpts.BaseDir,
		"src",
		filesOpts.BaseDir,
		"the base directory to copy from",
	)

	filesCmd.PersistentFlags().StringVar(
		&filesOpts.Prefix,
		"prefix",
		filesOpts.Prefix,
		"only export files starting with the provided prefix",
	)

	// TODO: Consider moving this into a validation function
	// nolint: errcheck
	filesCmd.MarkPersistentFlagRequired("src")

	ManifestCmd.AddCommand(filesCmd)
}

func runFileManifest(opts *promobot.GenerateManifestOptions) error {
	ctx := context.Background()

	src, err := filepath.Abs(opts.BaseDir)
	if err != nil {
		return errors.Wrapf(err, "resolving %q to absolute path", src)
	}

	opts.BaseDir = src

	manifest, err := promobot.GenerateManifest(ctx, *opts)
	if err != nil {
		return err
	}

	manifestYAML, err := yaml.Marshal(manifest)
	if err != nil {
		return errors.Wrap(err, "serializing manifest")
	}

	if _, err := os.Stdout.Write(manifestYAML); err != nil {
		return err
	}

	return nil
}

// TODO: Validate options

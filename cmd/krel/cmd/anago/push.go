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
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/release"
)

// pushCmd represents the subcommand for `krel anago push`
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push release artifacts into the Google Cloud",
	Long: `krel anago push

This subcommand can be used to push the release artifacts to the Google Cloud. 
It's only indented to be used from anago, which means the command might be
removed in future releases again when anago goes end of life.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.Wrap(runPush(pushOpts, version), "run krel anago push")
	},
}

var (
	pushOpts               = &release.PushBuildOptions{}
	version                string
	localGCSStagePath      string
	localReleaseImagesPath string
)

func init() {
	pushCmd.PersistentFlags().StringVar(
		&version,
		"version",
		"",
		"version to be used",
	)

	pushCmd.PersistentFlags().StringVar(
		&pushOpts.BuildDir,
		"build-dir",
		"",
		"build artifact directory of the release",
	)

	pushCmd.PersistentFlags().StringVar(
		&pushOpts.Bucket,
		"bucket",
		"",
		"GCS bucket to be used",
	)

	pushCmd.PersistentFlags().StringVar(
		&pushOpts.GCSSuffix,
		"gcs-suffix",
		"",
		"Specify a suffix to append to the upload destination on GCS",
	)

	pushCmd.PersistentFlags().StringVar(
		&localGCSStagePath,
		"local-gcs-stage-path",
		"",
		"Path to the local gcs-stage artifacts",
	)

	pushCmd.PersistentFlags().StringVar(
		&localReleaseImagesPath,
		"local-release-images-path",
		"",
		"Path to the local release-images artifacts",
	)

	pushOpts.AllowDup = true

	AnagoCmd.AddCommand(pushCmd)
}

func runPush(opts *release.PushBuildOptions, version string) error {
	pushBuild := release.NewPushBuild(opts)

	// Stage local artifacts and write checksums
	if err := pushBuild.StageLocalArtifacts(version); err != nil {
		return errors.Wrap(err, "staging local artifacts")
	}

	// Push gcs-stage to GCS
	if err := pushBuild.PushReleaseArtifacts(
		localGCSStagePath,
		filepath.Join(opts.GCSSuffix, release.GCSStagePath, version),
	); err != nil {
		return errors.Wrap(err, "pushing release artifacts")
	}

	// Push container release-images to GCS
	if err := pushBuild.PushReleaseArtifacts(
		localReleaseImagesPath,
		filepath.Join(opts.GCSSuffix, release.ImagesPath),
	); err != nil {
		return errors.Wrap(err, "pushing release artifacts")
	}

	return nil
}

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

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/build"
	"k8s.io/release/pkg/release"
)

const pushCmdDescription = `
Used for pushing developer builds and Jenkins' continuous builds.

Developer pushes simply run as they do pushing to devel/ on GCS.`

const pushCmdExample = `
krel push [--noupdatelatest] [--ci] [--bucket=<GCS bucket>] [--private-bucket]

Scenarios:

krel push                                   - Do a developer push
krel push --ci                              - Do a CI push
krel push --bucket=kubernetes-release-$USER - Do a developer push to kubernetes-release-$USER`

var pushBuildOpts = &build.Options{}

var pushBuildCmd = &cobra.Command{
	Use:           "push",
	Short:         "Push Kubernetes release artifacts to Google Cloud Storage (GCS)",
	Long:          pushCmdDescription,
	Example:       pushCmdExample,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runPushBuild(pushBuildOpts); err != nil {
			return errors.Wrap(err, "Failed to run:")
		}

		return nil
	},
}

func init() {
	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.AllowDup,
		"allow-dup",
		false,
		"Do not exit error if the build already exists on the gcs path",
	)

	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.CI,
		"ci",
		false,
		"Used when called from Jenkins (for ci runs)",
	)

	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.NoUpdateLatest,
		"noupdatelatest",
		false,
		"Do not update the latest file",
	)

	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.PrivateBucket,
		"private-bucket",
		false,
		"Do not mark published bits on GCS as publicly readable",
	)

	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.Bucket,
		"bucket",
		"devel",
		"Specify an alternate bucket for pushes (normally 'devel' or 'ci')",
	)

	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.BuildDir,
		"buildDir",
		release.BuildDir,
		fmt.Sprintf(
			"Specify an alternate build directory (defaults to '%s')",
			release.BuildDir,
		),
	)

	// TODO: Switch to "--registry" once CI no longer uses it
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.Registry,
		"registry",
		"",
		"If set, push docker images to specified registry/project",
	)

	pushBuildCmd.PersistentFlags().StringSliceVar(
		&pushBuildOpts.ExtraVersionMarkers,
		"extra-version-markers",
		build.DefaultExtraVersionMarkers,
		"Comma separated list which can be used to upload additional version files to GCS. The path is relative and is append to a GCS path. (--ci only)",
	)

	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.GCSRoot,
		"gcs-root",
		"",
		"Specify an alternate GCS path to push artifacts to",
	)

	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.VersionSuffix,
		"version-suffix",
		"",
		"Append suffix to version name if set",
	)

	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.Fast,
		"fast",
		false,
		"Specifies a fast build (linux/amd64 only)",
	)

	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.ValidateRemoteImageDigests,
		"validate-images",
		false,
		"Validate that the remote image digests exists",
	)

	rootCmd.AddCommand(pushBuildCmd)
}

func runPushBuild(opts *build.Options) error {
	return build.NewInstance(opts).Push()
}

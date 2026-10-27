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

	"github.com/spf13/cobra"

	"k8s.io/release/pkg/build"
	"k8s.io/release/pkg/release"
)

const ciBuildCmdDescription = `
Build Kubernetes in CI and push artifacts.

Developer pushes simply run as they do pushing to devel/ on GCS.`

const ciBuildExample = `
krel ci-build [--noupdatelatest] [--bucket=<bucket>] [--bucket-root=<bucket root>]

Scenarios:

krel ci-build --allow-dup --fast --registry=gcr.io/foo           - Run a fast build and push images to gcr.io/foo
krel ci-build --bucket cool-bucket --bucket-root new-bucket-root - Push to gs://cool-bucket/new-bucket-root
krel ci-build --bucket s3://cool-bucket                          - Push to s3://cool-bucket/ci`

var ciBuildOpts = &build.Options{}

// ciBuildSign is the inverse of `build.Options.NoSign`, exposed as `--sign`.
var ciBuildSign bool

var ciBuildCmd = &cobra.Command{
	Use:           "ci-build",
	Short:         "Build Kubernetes in CI and push release artifacts to Google Cloud Storage (GCS) or Amazon S3",
	Long:          ciBuildCmdDescription,
	Example:       ciBuildExample,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runCIBuild(ciBuildOpts); err != nil {
			return fmt.Errorf("failed to run: %w", err)
		}

		return nil
	},
}

func init() {
	// Build options
	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.Fast,
		"fast",
		false,
		"Specifies a fast build (linux/amd64 only)",
	)

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.ConfigureDocker,
		"configure-docker",
		false,
		"Configure docker client for gcr.io authentication to allow communication with non-public registries",
	)

	// Push options

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.AllowDup,
		"allow-dup",
		false,
		"Do not exit error if the build already exists on the gcs path",
	)

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.NoUpdateLatest,
		"noupdatelatest",
		false,
		"Do not update the latest file",
	)

	// TODO: Configure a default const here
	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.Bucket,
		"bucket",
		"",
		"Specify an alternate bucket for pushes (normally 'devel' or 'ci'), or an s3://<bucket>[/<prefix>] URL to push to Amazon S3, which stores every '+' in the object keys as a space",
	)

	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.BuildDir,
		"buildDir",
		release.BuildDir,
		fmt.Sprintf(
			"Specify an alternate build directory (defaults to '%s')",
			release.BuildDir,
		),
	)

	// TODO: Switch to "--registry" once CI no longer uses it
	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.Registry,
		"registry",
		"",
		"If set, push docker images to specified registry/project",
	)

	ciBuildCmd.PersistentFlags().StringSliceVar(
		&ciBuildOpts.ExtraVersionMarkers,
		"extra-version-markers",
		build.DefaultExtraVersionMarkers,
		"Comma separated list which can be used to upload additional version files to GCS. The path is relative and is append to a GCS path. (--ci only)",
	)

	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.BucketRoot,
		"bucket-root",
		"",
		"Specify an alternate path in the bucket to push artifacts to",
	)

	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.GCSRoot, //nolint:staticcheck // --gcs-root is deprecated, too
		"gcs-root",
		"",
		"Specify an alternate GCS path to push artifacts to",
	)
	_ = ciBuildCmd.PersistentFlags().MarkDeprecated("gcs-root", "This flag is deprecated and will be removed in v0.24.0. Use --bucket-root instead.")

	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.VersionSuffix,
		"version-suffix",
		"",
		"Append suffix to version name if set",
	)

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.ValidateRemoteImageDigests,
		"validate-images",
		false,
		"Validate that the remote image digests exists",
	)

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildSign,
		"sign",
		true,
		"Sign the pushed container images, use --sign=false to skip the signing step",
	)

	rootCmd.AddCommand(ciBuildCmd)
}

func runCIBuild(opts *build.Options) error {
	opts.CI = true
	opts.NoSign = !ciBuildSign

	return build.NewInstance(opts).Build()
}

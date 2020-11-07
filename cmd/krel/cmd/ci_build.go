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

const ciBuildCmdDescription = `
Used for pushing developer builds and Jenkins' continuous builds.

Developer pushes simply run as they do pushing to devel/ on GCS.`

const ciBuildExample = `
ci-build [--noupdatelatest] [--ci] [--bucket=<GCS bucket>] [--private-bucket]

Scenarios:

krel push                                   - Do a developer push
krel push --ci                              - Do a CI push
krel push --bucket=kubernetes-release-$USER - Do a developer push to kubernetes-release-$USER`

var ciBuildOpts = &release.PushBuildOptions{}

var ciBuildCmd = &cobra.Command{
	Use:           "ci-build",
	Short:         "Push Kubernetes release artifacts to Google Cloud Storage (GCS)",
	Long:          ciBuildCmdDescription,
	Example:       ciBuildExample,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runCIBuild(ciBuildOpts); err != nil {
			return errors.Wrap(err, "Failed to run:")
		}

		return nil
	},
}

func init() {
	// Build options
	/*
		PARSER.add_argument(
				'--fast', action='store_true', help='Specifies a fast build')
		PARSER.add_argument(
				'--register-gcloud-helper', action='store_true',
				help='Register gcloud as docker credentials helper')
	*/

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.Fast,
		"fast",
		false,
		"Specifies a fast build (linux/amd64 only)",
	)

	// Push options
	/*
		PARSER.add_argument(
				'--release', help='Upload binaries to the specified gs:// path')
		PARSER.add_argument(
				'--suffix', help='Append suffix to the upload path if set')
		PARSER.add_argument(
				'--registry', help='Push images to the specified docker registry')
		PARSER.add_argument(
				'--extra-version-markers', help='Additional version file uploads to')
		PARSER.add_argument(
				'--fast', action='store_true', help='Specifies a fast build')
		PARSER.add_argument(
				'--skip-update-latest', action='store_true', help='Do not update the latest file')
	*/

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.AllowDup,
		"allow-dup",
		false,
		"Do not exit error if the build already exists on the gcs path",
	)

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.CI,
		"ci",
		false,
		"Used when called from Jenkins (for ci runs)",
	)

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.NoUpdateLatest,
		"noupdatelatest",
		false,
		"Do not update the latest file",
	)

	ciBuildCmd.PersistentFlags().BoolVar(
		&ciBuildOpts.PrivateBucket,
		"private-bucket",
		false,
		"Do not mark published bits on GCS as publicly readable",
	)

	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.Bucket,
		"bucket",
		"devel",
		"Specify an alternate bucket for pushes (normally 'devel' or 'ci')",
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

	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.DockerRegistry,
		"docker-registry",
		"",
		"If set, push docker images to specified registry/project",
	)

	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.ExtraVersionMarkers,
		"extra-version-markers",
		"",
		"Comma separated list which can be used to upload additional version files to GCS. The path is relative and is append to a GCS path. (--ci only)",
	)

	ciBuildCmd.PersistentFlags().StringVar(
		&ciBuildOpts.GCSSuffix,
		"gcs-suffix",
		"",
		"Specify a suffix to append to the upload destination on GCS",
	)

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
		"Validate that the remove image digests exists, needs `skopeo` in `$PATH`",
	)

	rootCmd.AddCommand(ciBuildCmd)
}

func runCIBuild(opts *release.PushBuildOptions) error {
	return build.NewBuild(opts).Build()
}

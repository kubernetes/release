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
	"log"

	"github.com/spf13/cobra"
)

const description = `
Used for pushing developer builds and Jenkins' continuous builds.

Developer pushes simply run as they do pushing to devel/ on GCS.
In --ci mode, 'push' runs in mock mode by default. Use --nomock to do
a real push.

Federation values are just passed through as exported global vars still
due to the fact that we're still leveraging the existing federation
interface in kubernetes proper.

push                       - Do a developer push
push --nomock --federation --ci
                           - Do a (non-mocked) CI push with federation
push --bucket=kubernetes-release-$USER
                           - Do a developer push to kubernetes-release-$USER`

type pushBuildOptions struct {
	bucket           string
	dockerRegistry   string
	extraPublishFile string
	gcsSuffix        string
	releaseKind      string
	releaseType      string
	versionSuffix    string
	allowDup         bool
	ci               bool
	federation       bool
	noUpdateLatest   bool
	privateBucket    bool
}

var pushBuildOpts = &pushBuildOptions{}

var pushBuildCmd = &cobra.Command{
	Use:     "push [--federation] [--noupdatelatest] [--ci] [--bucket=<GS bucket>] [--private-bucket]",
	Short:   "push kubernetes release artifacts to GCS",
	Example: description,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runPushBuild(pushBuildOpts); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.allowDup,
		"allow-dup",
		false,
		"Do not exit error if the build already exists on the gcs path",
	)
	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.ci,
		"ci",
		false,
		"Used when called from Jenkins (for ci runs)",
	)
	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.federation,
		"federation",
		false,
		"Enable FEDERATION push",
	)
	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.noUpdateLatest,
		"noupdatelatest",
		false,
		"Do not update the latest file",
	)
	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.privateBucket,
		"private-bucket",
		false, "Do not mark published bits on GCS as publicly readable",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.bucket,
		"bucket",
		"devel",
		"Specify an alternate bucket for pushes (normally 'devel' or 'ci')",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.dockerRegistry,
		"docker-registry",
		"",
		"If set, push docker images to specified registry/project",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.extraPublishFile,
		"extra-publish-file",
		"",
		"Used when need to upload additional version file to GCS. The path is relative and is append to a GCS path. (--ci only)",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.gcsSuffix,
		"gcs-suffix",
		"",
		"Specify a suffix to append to the upload destination on GCS",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.releaseKind,
		"release-kind",
		"devel",
		"Specify an alternate bucket for pushes (normally 'devel' or 'ci')",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.releaseType,
		"release-type",
		"kubernetes",
		"Kind of release to push to GCS. Supported values are kubernetes (default) or federation",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.versionSuffix,
		"version-suffix",
		"",
		"Append suffix to version name if set",
	)

	rootCmd.AddCommand(pushBuildCmd)
}

func runPushBuild(opts *pushBuildOptions) error {
	log.Println("unimplemented")
	return nil
}

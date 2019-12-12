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
	"context"
	"os"
	"os/user"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/release"
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
		"kubernetes",
		"Kind of release to push to GCS. Supported values are kubernetes (default) or federation",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.releaseType,
		"release-type",
		"devel",
		"Specify an alternate bucket for pushes (normally 'devel' or 'ci')",
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
	var latest string
	releaseKind := opts.releaseKind

	// Check if latest build uses bazel
	dir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Unable to get working directory")
	}

	isBazel, err := release.BuiltWithBazel(dir, releaseKind)
	if err != nil {
		return errors.Wrap(err, "Unable to identify if release built with Bazel")
	}

	if isBazel {
		logrus.Info("Using Bazel build version")
		version, err := release.ReadBazelVersion(dir)
		if err != nil {
			return errors.Wrap(err, "Unable to read Bazel build version")
		}
		latest = version
	} else {
		logrus.Info("Using Dockerized build version")
		version, err := release.ReadDockerizedVersion(dir, releaseKind)
		if err != nil {
			return errors.Wrap(err, "Unable to read Dockerized build version")
		}
		latest = version
	}

	logrus.Infof("Found build version: %s", latest)

	valid, err := release.IsValidReleaseBuild(latest)
	if err != nil {
		return errors.Wrap(err, "Unable to determine if release build version is valid")
	}
	if !valid {
		return errors.Errorf("Build version %s is not valid for release", latest)
	}

	if opts.ci && release.IsDirtyBuild(latest) {
		return errors.New(`Refusing to push dirty build with --ci flag given.\n
			CI builds should always be performed from clean commits`)
	}

	if opts.versionSuffix != "" {
		latest += "-" + opts.versionSuffix
	}

	gcsDest := opts.releaseType

	// TODO: is this how we want to handle gcs dest args?
	if opts.ci {
		gcsDest = "ci"
	}

	gcsDest += opts.gcsSuffix

	releaseBucket := opts.bucket
	if !rootOpts.nomock {
		u, err := user.Current()
		if err != nil {
			return errors.Wrap(err, "Unable to identify current user")
		}

		releaseBucket += "-" + u.Username
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return errors.Wrap(err, "error fetching gcloud credentials... try running \"gcloud auth application-default login\"")
	}

	bucket := client.Bucket(releaseBucket)
	if bucket == nil {
		return errors.Errorf("unable to identify specified bucket for artifacts: %s", releaseBucket)
	}

	// Check if bucket exists.
	if _, err = bucket.Attrs(context.Background()); err != nil {
		return errors.Wrap(err, "Unable to find release artifact bucket")
	}

	return nil
}

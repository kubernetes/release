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
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/util"
)

const description = `
Used for pushing developer builds and Jenkins' continuous builds.

Developer pushes simply run as they do pushing to devel/ on GCS.
In --ci mode, 'push' runs in mock mode by default. Use --nomock to do
a real push.

push                       - Do a developer push
push --nomock --ci
                           - Do a (non-mocked) CI push
push --bucket=kubernetes-release-$USER
                           - Do a developer push to kubernetes-release-$USER`

type pushBuildOptions struct {
	bucket           string
	buildDir         string
	dockerRegistry   string
	extraPublishFile string
	gcsSuffix        string
	releaseType      string
	versionSuffix    string
	allowDup         bool
	ci               bool
	noUpdateLatest   bool
	privateBucket    bool
}

var pushBuildOpts = &pushBuildOptions{}

var pushBuildCmd = &cobra.Command{
	Use:           "push [--noupdatelatest] [--ci] [--bucket=<GS bucket>] [--private-bucket]",
	Short:         "Push Kubernetes release artifacts to Google Cloud Storage (GCS)",
	Example:       description,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runPushBuild(pushBuildOpts); err != nil {
			return errors.Wrap(err, "Failed to run:")
		}

		return nil
	},
}

type stageFile struct {
	srcPath  string
	dstPath  string
	required bool
}

var gcpStageFiles = []stageFile{
	{
		srcPath:  filepath.Join(release.GCEPath, "configure-vm.sh"),
		dstPath:  filepath.Join(release.GCSStagePath, "extra/gce"),
		required: false,
	},
	{
		srcPath:  filepath.Join(release.GCIPath, "node.yaml"),
		dstPath:  filepath.Join(release.GCSStagePath, "extra/gce"),
		required: true,
	},
	{
		srcPath:  filepath.Join(release.GCIPath, "master.yaml"),
		dstPath:  filepath.Join(release.GCSStagePath, "extra/gce"),
		required: true,
	},
	{
		srcPath:  filepath.Join(release.GCIPath, "configure.sh"),
		dstPath:  filepath.Join(release.GCSStagePath, "extra/gce"),
		required: true,
	},
	{
		srcPath:  filepath.Join(release.GCIPath, "shutdown.sh"),
		dstPath:  filepath.Join(release.GCSStagePath, "extra/gce"),
		required: false,
	},
}

var windowsStageFiles = []stageFile{
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "configure.ps1"),
		dstPath:  release.WindowsGCSPath,
		required: true,
	},
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "common.psm1"),
		dstPath:  release.WindowsGCSPath,
		required: true,
	},
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "k8s-node-setup.psm1"),
		dstPath:  release.WindowsGCSPath,
		required: true,
	},
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "testonly/install-ssh.psm1"),
		dstPath:  release.WindowsGCSPath,
		required: true,
	},
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "testonly/user-profile.psm1"),
		dstPath:  release.WindowsGCSPath,
		required: true,
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
		&pushBuildOpts.noUpdateLatest,
		"noupdatelatest",
		false,
		"Do not update the latest file",
	)
	pushBuildCmd.PersistentFlags().BoolVar(
		&pushBuildOpts.privateBucket,
		"private-bucket",
		false,
		"Do not mark published bits on GCS as publicly readable",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.bucket,
		"bucket",
		"devel",
		"Specify an alternate bucket for pushes (normally 'devel' or 'ci')",
	)
	pushBuildCmd.PersistentFlags().StringVar(
		&pushBuildOpts.buildDir,
		"buildDir",
		"_output",
		"Specify an alternate build directory (defaults to '_output')",
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

	// Check if latest build uses bazel
	dir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "get working directory")
	}

	isBazel, err := release.BuiltWithBazel(dir)
	if err != nil {
		return errors.Wrap(err, "identify if release built with Bazel")
	}

	if isBazel {
		logrus.Info("Using Bazel build version")
		version, err := release.ReadBazelVersion(dir)
		if err != nil {
			return errors.Wrap(err, "read Bazel build version")
		}
		latest = version
	} else {
		logrus.Info("Using Dockerized build version")
		version, err := release.ReadDockerizedVersion(dir)
		if err != nil {
			return errors.Wrap(err, "read Dockerized build version")
		}
		latest = version
	}

	logrus.Infof("Found build version: %s", latest)

	valid, err := release.IsValidReleaseBuild(latest)
	if err != nil {
		return errors.Wrap(err, "determine if release build version is valid")
	}
	if !valid {
		return errors.Errorf("build version %s is not valid for release", latest)
	}

	if opts.ci && release.IsDirtyBuild(latest) {
		return errors.New(`refusing to push dirty build with --ci flag given. CI builds should always be performed from clean commits`)
	}

	if opts.versionSuffix != "" {
		latest += "-" + opts.versionSuffix
	}

	logrus.Infof("Latest version is %s", latest)

	gcsDest := opts.releaseType

	// TODO: is this how we want to handle gcs dest args?
	if opts.ci {
		gcsDest = "ci"
	}

	gcsDest += opts.gcsSuffix

	logrus.Infof("GCS destination is %s", gcsDest)

	releaseBucket := opts.bucket
	if rootOpts.nomock {
		logrus.Infof("Running a *REAL* push with bucket %s", releaseBucket)
	} else {
		u, err := user.Current()
		if err != nil {
			return errors.Wrap(err, "identify current user")
		}

		releaseBucket += "-" + u.Username
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return errors.Wrap(err, "fetching gcloud credentials... try running \"gcloud auth application-default login\"")
	}

	bucket := client.Bucket(releaseBucket)
	if bucket == nil {
		return errors.Errorf("identify specified bucket for artifacts: %s", releaseBucket)
	}

	// Check if bucket exists and user has permissions
	requiredGCSPerms := []string{"storage.objects.create"}
	perms, err := bucket.IAM().TestPermissions(context.Background(), requiredGCSPerms)
	if err != nil {
		return errors.Wrap(err, "find release artifact bucket")
	}
	if len(perms) != 1 {
		return errors.Errorf("GCP user must have at least %s permissions on bucket %s", requiredGCSPerms, releaseBucket)
	}

	buildDir := buildOpts.BuildDir
	if err = util.RemoveAndReplaceDir(filepath.Join(buildDir, release.GCSStagePath)); err != nil {
		return errors.Wrap(err, "remove and replace GCS staging directory")
	}

	// Copy release tarballs to local GCS staging directory for push
	if err = util.CopyDirContentsLocal(filepath.Join(buildDir, release.ReleaseTarsPath), filepath.Join(buildDir, release.GCSStagePath)); err != nil {
		return errors.Wrap(err, "copy source directory into destination")
	}

	// Copy helpful GCP scripts to local GCS staging directory for push
	for _, file := range gcpStageFiles {
		if err := util.CopyFileLocal(filepath.Join(buildDir, file.srcPath), filepath.Join(buildDir, file.dstPath), file.required); err != nil {
			return errors.Wrap(err, "copy GCP stage files")
		}
	}

	// Copy helpful Windows scripts to local GCS staging directory for push
	for _, file := range windowsStageFiles {
		if err := util.CopyFileLocal(filepath.Join(buildDir, file.srcPath), filepath.Join(buildDir, file.dstPath), file.required); err != nil {
			return errors.Wrap(err, "copy Windows stage files")
		}
	}

	// TODO
	// Prepare naked binaries
	// Write checksums
	// Push Docker images
	// Push artifacts to release bucket is --ci

	return nil
}

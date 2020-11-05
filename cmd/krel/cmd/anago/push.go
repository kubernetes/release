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
		return errors.Wrap(runPush(pushOpts), "run krel anago push")
	},
}

var (
	pushOpts     = &release.PushBuildOptions{}
	runStage     bool
	runRelease   bool
	buildVersion string
)

func init() {
	pushCmd.PersistentFlags().BoolVar(
		&runStage,
		"stage",
		false,
		"run in stage mode",
	)

	pushCmd.PersistentFlags().BoolVar(
		&runRelease,
		"release",
		false,
		"run in release mode",
	)

	pushCmd.PersistentFlags().StringVar(
		&pushOpts.Version,
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
		&pushOpts.DockerRegistry,
		"container-registry",
		"",
		"Container image registry to be used",
	)

	pushCmd.PersistentFlags().StringVar(
		&buildVersion,
		"build-version",
		"",
		"Build version from Jenkins (only used when --release specified)",
	)

	pushOpts.AllowDup = true
	pushOpts.ValidateRemoteImageDigests = true

	AnagoCmd.AddCommand(pushCmd)
}

func runPush(opts *release.PushBuildOptions) error {
	pushBuild := release.NewPushBuild(opts)
	if err := pushBuild.CheckReleaseBucket(); err != nil {
		return errors.Wrap(err, "check release bucket access")
	}

	if runStage {
		return runPushStage(pushBuild, opts)
	} else if runRelease {
		return runPushRelease(pushBuild, opts)
	}

	return errors.New("neither --stage nor --release provided")
}

func runPushStage(
	pushBuild *release.PushBuild,
	opts *release.PushBuildOptions,
) error {
	// Stage the local source tree
	if err := pushBuild.StageLocalSourceTree(buildVersion); err != nil {
		return errors.Wrap(err, "staging local source tree")
	}

	// Stage local artifacts and write checksums
	if err := pushBuild.StageLocalArtifacts(); err != nil {
		return errors.Wrap(err, "staging local artifacts")
	}
	gcsPath := filepath.Join("stage", buildVersion, opts.Version)

	// Push gcs-stage to GCS
	if err := pushBuild.PushReleaseArtifacts(
		filepath.Join(opts.BuildDir, release.GCSStagePath, opts.Version),
		filepath.Join(gcsPath, release.GCSStagePath, opts.Version),
	); err != nil {
		return errors.Wrap(err, "pushing release artifacts")
	}

	// Push container release-images to GCS
	if err := pushBuild.PushReleaseArtifacts(
		filepath.Join(opts.BuildDir, release.ImagesPath),
		filepath.Join(gcsPath, release.ImagesPath),
	); err != nil {
		return errors.Wrap(err, "pushing release artifacts")
	}

	// Push container images into registry
	if err := pushBuild.PushContainerImages(); err != nil {
		return errors.Wrap(err, "pushing container images")
	}

	return nil
}

func runPushRelease(
	pushBuild *release.PushBuild,
	opts *release.PushBuildOptions,
) error {
	if err := pushBuild.CopyStagedFromGCS(opts.Bucket, buildVersion); err != nil {
		return errors.Wrap(err, "copy staged from GCS")
	}

	// In an official nomock release, we want to ensure that container images
	// have been promoted from staging to production, so we do the image
	// manifest validation against production instead of staging.
	targetRegistry := opts.DockerRegistry
	if targetRegistry == release.GCRIOPathStaging {
		targetRegistry = release.GCRIOPathProd
	}
	// Image promotion has been done on nomock stage, verify that the images
	// are available.
	if err := release.NewImages().Validate(
		targetRegistry, opts.Version, opts.BuildDir,
	); err != nil {
		return errors.Wrap(err, "validate container images")
	}

	if err := release.NewPublisher().PublishVersion(
		"release", opts.Version, opts.BuildDir, opts.Bucket, nil, false, false,
	); err != nil {
		return errors.Wrap(err, "publish release")
	}
	return nil
}

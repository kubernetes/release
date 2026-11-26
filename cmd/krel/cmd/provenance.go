/*
Copyright 2026 The Kubernetes Authors.

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
	"errors"

	"github.com/spf13/cobra"

	"k8s.io/release/pkg/release"
)

var provenanceOpts = release.DefaultSigningBuildOptions()

// provenanceCmd represents the subcommand for `krel provenance`.
var provenanceCmd = &cobra.Command{
	Use:   "provenance --build-id <stage build ID>",
	Short: "Write and sign the provenance of a stage build from outside of it",
	Long: `krel provenance --build-id <stage build ID>

Writes the SLSA provenance of a stage build as the signing build of SLSA
build level 3, see gcb/provenance/cloudbuild.yaml: it reads the record of the
stage build from Cloud Build, downloads the staged artifacts, computes the
digests of the staged files and images itself, and signs SLSA v1 provenance
about them with the builder ` + release.SigningBuildBuilder + `
and the build type ` + release.SigningBuildType + `,
as the service account set with --signer, by impersonating it. It writes the
bundles next to the provenance of the stage itself, which it doesn't touch:

  stage/<build version>/` + release.SigningBuildProvenanceFilename + `        the staged files
  stage/<build version>/` + release.SigningBuildImageProvenanceFilename + `  the staged images, also attached to them
  stage/<build version>/<version>/` + release.SigningBuildFinalProvenanceFilename + `  the files of the release

It only signs builds of the stage trigger set with --stage-trigger-id, or
--mock-stage-trigger-id for the mock signer, that ran as
` + release.StageServiceAccount + `.
It fails if a staged file was written, or a staged image uploaded, outside
of the run of the stage build. A bundle that exists already, verifies as the
signer's and holds the same statement is kept, so that running it again for
a stage build signs nothing new. The production signer only signs nomock
stages in the production bucket and registry, the mock signer only stages in
the mock ones.

With --dry-run, it writes the unsigned statements to --output-dir and
changes nothing remote.`,
	Example: `  # Write the unsigned statements of a stage build locally:
  krel provenance --build-id 54ef86d9-5a89-4e0e-a008-10904a0f5a34 --dry-run

  # Sign the provenance of a mock stage:
  krel provenance --build-id 54ef86d9-5a89-4e0e-a008-10904a0f5a34 --signer ` + release.SigningBuildMockSigner,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(*cobra.Command, []string) error {
		if provenanceOpts.BuildID == "" {
			return errors.New("--build-id is required")
		}

		return release.NewSigningBuild(provenanceOpts).Run()
	},
}

func init() {
	flags := provenanceCmd.PersistentFlags()

	flags.StringVar(&provenanceOpts.BuildID, "build-id", "", "Cloud Build ID of the stage build")
	flags.StringVar(&provenanceOpts.Project, "project", provenanceOpts.Project, "project the stage build ran in")
	flags.StringVar(&provenanceOpts.Region, "region", provenanceOpts.Region, "region the stage build ran in")
	flags.StringVar(&provenanceOpts.Signer, "signer", provenanceOpts.Signer, "service account to sign the provenance as")
	flags.StringVar(
		&provenanceOpts.StageTriggerID, "stage-trigger-id", "",
		"ID of the Cloud Build trigger of production stages, the only builds the production signer signs",
	)
	flags.StringVar(
		&provenanceOpts.MockStageTriggerID, "mock-stage-trigger-id", "",
		"ID of the Cloud Build trigger of mock stages, the only builds the mock signer signs",
	)
	flags.StringVar(
		&provenanceOpts.ScratchDirectory, "scratch-dir", provenanceOpts.ScratchDirectory,
		"directory to download the staged artifacts to",
	)
	flags.BoolVar(&provenanceOpts.DryRun, "dry-run", false, "write the unsigned statements to --output-dir only")
	flags.StringVar(
		&provenanceOpts.OutputDirectory, "output-dir", provenanceOpts.OutputDirectory,
		"directory for the statements of --dry-run",
	)

	rootCmd.AddCommand(provenanceCmd)
}

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
	"bytes"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/attestation"
)

const serviceAccountFileFlag = "service-account-file"

type signAttestationOptions struct {
	outputPath         string
	serviceAccountFile string
	// serviceAccountJSON is the key data read from the environment
	serviceAccountJSON string
}

var signAttestationOpts = &signAttestationOptions{}

// signAttestationCmd represents the subcommand for `krel sign attestation`.
var signAttestationCmd = &cobra.Command{
	Use:   "attestation statement.json",
	Short: "Sign an in-toto statement into a sigstore bundle",
	Long: `krel sign attestation attestation.json

Signs an in-toto statement using sigstore and writes the resulting bundle
(DSSE envelope, Fulcio certificate and transparency log proofs) to stdout
or to the file set with --` + outputPathFlag + `.

By default the statement is signed with the ambient identity provider.

To sign with an explicit identity, pass a Google Cloud service account key
file with --` + serviceAccountFileFlag + ` or set the contents of the key in
the ` + attestation.ServiceAccountEnvKey + ` environment variable (the flag
takes precedence). The signer is then locked to that service account: the
certificate is only requested with its identity and signing fails if that is
not possible, it never falls back to the ambient credentials.`,

	Example: `  # Sign an attestation using the ambient GCP credentials:
  krel sign attestation provenance.json > provenance.json.sigstore.json

  # Sign using a service account key:
  krel sign attestation --service-account-file key.json --output-path provenance.sigstore.json provenance.json`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(_ *cobra.Command, args []string) error {
		if key, isSet := os.LookupEnv(attestation.ServiceAccountEnvKey); isSet {
			signAttestationOpts.serviceAccountJSON = key
		}

		return runSignAttestation(singOpts, signAttestationOpts, args[0])
	},
}

func init() {
	signAttestationCmd.PersistentFlags().StringVar(
		&signAttestationOpts.outputPath,
		outputPathFlag,
		"",
		"write the signed bundle to a file instead of stdout",
	)

	signAttestationCmd.PersistentFlags().StringVar(
		&signAttestationOpts.serviceAccountFile,
		serviceAccountFileFlag,
		"",
		"path to a Google service account key (defaults to $"+attestation.ServiceAccountEnvKey+" or the ambient credentials)",
	)

	signCmd.AddCommand(signAttestationCmd)
}

func runSignAttestation(signOpts *signOptions, opts *signAttestationOptions, statementPath string) error {
	signerOpts := attestation.DefaultSignerOptions()
	signerOpts.ServiceAccountFile = opts.serviceAccountFile
	signerOpts.ServiceAccountJSON = []byte(opts.serviceAccountJSON)
	signerOpts.Timeout = signOpts.timeout

	// We will now sign the bundle in memory to avoid writing until
	// we know signing succeeded
	var bundle bytes.Buffer
	if err := attestation.NewSigner(signerOpts).SignFile(statementPath, &bundle); err != nil {
		return fmt.Errorf("signing attestation: %w", err)
	}

	if opts.outputPath == "" {
		if _, err := bundle.WriteTo(os.Stdout); err != nil {
			return fmt.Errorf("writing bundle to stdout: %w", err)
		}

		return nil
	}

	if err := os.WriteFile(opts.outputPath, bundle.Bytes(), 0o644); err != nil { //nolint:gosec // bundles are public
		return fmt.Errorf("writing bundle to %s: %w", opts.outputPath, err)
	}

	logrus.Infof("Signed bundle written to %s", opts.outputPath)

	return nil
}

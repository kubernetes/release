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

const (
	serviceAccountFileFlag = "service-account-file"
	inPlaceFlag            = "in-place"
)

type signAttestationOptions struct {
	outputPath         string
	inPlace            bool
	serviceAccountFile string
	// serviceAccountJSON is the key data read from the environment
	serviceAccountJSON string
}

var signAttestationOpts = &signAttestationOptions{}

// signAttestationCmd represents the subcommand for `krel sign attestation`.
var signAttestationCmd = &cobra.Command{
	Use:   "attestation statement.json [--in-place statement.json...]",
	Short: "Sign an in-toto statement into a sigstore bundle",
	Long: `krel sign attestation attestation.json [--in-place statement.json...]

Signs an in-toto statement using sigstore and writes the resulting bundle
(DSSE envelope, Fulcio certificate and transparency log proofs) to stdout
or to the file set with --` + outputPathFlag + `. Both the statement and the
output path can be local files or objects in Google Cloud Storage
(gs://bucket/path/statement.json).

With --` + inPlaceFlag + `, the signed bundle replaces the statement file (or
gs:// object) itself. Several statements can then be signed at once, all in
the same signing session, reusing the identity and the Fulcio certificate.
--` + outputPathFlag + ` can be combined with --` + inPlaceFlag + ` to
additionally write a copy of the bundle, but only for a single statement.
Files that are already signed (sigstore bundles or DSSE envelopes) are
rejected.

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
  krel sign attestation --service-account-file key.json --output-path provenance.sigstore.json provenance.json

  # Sign a staged provenance stored in a bucket:
  krel sign attestation gs://k8s-release-dev/stage/v1.36.0-alpha.1.10+abcdef/provenance.json

  # Sign several statements in place, replacing the originals with the bundles:
  krel sign attestation --in-place gs://bucket/stage/build/provenance.json sbom.intoto.json`,
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(_ *cobra.Command, args []string) error {
		if key, isSet := os.LookupEnv(attestation.ServiceAccountEnvKey); isSet {
			signAttestationOpts.serviceAccountJSON = key
		}

		return runSignAttestation(singOpts, signAttestationOpts, args)
	},
}

func init() {
	signAttestationCmd.PersistentFlags().StringVar(
		&signAttestationOpts.outputPath,
		outputPathFlag,
		"",
		"write the signed bundle to a file or gs:// object instead of stdout",
	)

	signAttestationCmd.PersistentFlags().BoolVar(
		&signAttestationOpts.inPlace,
		inPlaceFlag,
		false,
		"replace each statement file or gs:// object with its signed bundle",
	)

	signAttestationCmd.PersistentFlags().StringVar(
		&signAttestationOpts.serviceAccountFile,
		serviceAccountFileFlag,
		"",
		"path to a Google service account key (defaults to $"+attestation.ServiceAccountEnvKey+" or the ambient credentials)",
	)

	signCmd.AddCommand(signAttestationCmd)
}

func runSignAttestation(signOpts *signOptions, opts *signAttestationOptions, statements []string) error {
	if err := validateSignAttestationArgs(opts, statements); err != nil {
		return err
	}

	signerOpts := attestation.DefaultSignerOptions()
	signerOpts.ServiceAccountFile = opts.serviceAccountFile
	signerOpts.ServiceAccountJSON = []byte(opts.serviceAccountJSON)
	signerOpts.Timeout = signOpts.timeout

	signer := attestation.NewSigner(signerOpts)

	if opts.inPlace {
		return signInPlace(signer, opts, statements)
	}

	// We will now sign the bundle in memory to avoid writing until
	// we know signing succeeded
	var bundle bytes.Buffer

	if err := signer.SignFile(statements[0], &bundle); err != nil {
		return fmt.Errorf("signing attestation: %w", err)
	}

	if opts.outputPath == "" {
		if _, err := bundle.WriteTo(os.Stdout); err != nil {
			return fmt.Errorf("writing bundle to stdout: %w", err)
		}

		return nil
	}

	if err := signer.WriteFile(opts.outputPath, bundle.Bytes()); err != nil {
		return fmt.Errorf("writing bundle: %w", err)
	}

	logrus.Infof("Signed bundle written to %s", opts.outputPath)

	return nil
}

// signInPlace signs all statements in one session and replaces each of them
// with its resulting bundle. Nothing is written unless all statements are signed.
func signInPlace(signer *attestation.Signer, opts *signAttestationOptions, statements []string) error {
	signed, err := signer.SignFiles(statements)
	if err != nil {
		return fmt.Errorf("signing attestations: %w", err)
	}

	for _, statement := range signed {
		var bundle bytes.Buffer
		if err := signer.WriteBundle(statement.Bundle, &bundle); err != nil {
			return fmt.Errorf("serializing bundle of %s: %w", statement.Path, err)
		}

		if err := signer.WriteFile(statement.Path, bundle.Bytes()); err != nil {
			return fmt.Errorf("replacing statement with its bundle: %w", err)
		}

		logrus.Infof("Signed %s in place", statement.Path)

		if opts.outputPath != "" {
			if err := signer.WriteFile(opts.outputPath, bundle.Bytes()); err != nil {
				return fmt.Errorf("writing bundle copy: %w", err)
			}

			logrus.Infof("Signed bundle written to %s", opts.outputPath)
		}
	}

	return nil
}

// validateSignAttestationArgs checks the combination of flags and statements.
func validateSignAttestationArgs(opts *signAttestationOptions, statements []string) error {
	if len(statements) > 1 {
		if !opts.inPlace {
			return fmt.Errorf("signing more than one statement requires --%s", inPlaceFlag)
		}

		if opts.outputPath != "" {
			return fmt.Errorf("--%s can only be used when signing a single statement", outputPathFlag)
		}
	}

	return nil
}

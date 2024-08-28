/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nozzle/throttler"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/release-sdk/gcli"
	"sigs.k8s.io/release-sdk/object"
	"sigs.k8s.io/release-sdk/sign"
)

const (
	outputPathFlag           = "output-path"
	privateKeyPathFlag       = "private-key-path"
	publicKeyPathFlag        = "public-key-path"
	certIdentityFlag         = "certificate-identity"
	certIdentityRegexpFlag   = "certificate-identity-regexp"
	certOidcIssuerFlag       = "certificate-oidc-issuer"
	certOidcIssuerRegexpFlag = "certificate-oidc-issuer-regexp"
	sigExt                   = ".sig"
	certExt                  = ".cert"
)

type signBlobOptions struct {
	outputPath string

	privateKeyPath string
	publicKeyPath  string

	certOidcIssuer       string
	certOidcIssuerRegexp string
	certIdentity         string
	certIdentityRegexp   string
}

type signingBundle struct {
	destinationPathToCopy string
	fileToSign            string
	fileLocalLocation     string
}

var signBlobOpts = &signBlobOptions{}

// signBlobCmd represents the subcommand for `krel sign blobs`.
var signBlobCmd = &cobra.Command{
	Use:           "blobs",
	Short:         "Sign blobs",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSignBlobs(singOpts, signBlobOpts, args)
	},
}

func init() {
	signBlobCmd.PersistentFlags().StringVarP(
		&signBlobOpts.outputPath,
		outputPathFlag,
		"",
		"",
		"write the certificate and signatures to a file in the set path",
	)

	signBlobCmd.PersistentFlags().StringVarP(
		&signBlobOpts.privateKeyPath,
		privateKeyPathFlag,
		"",
		"",
		"path for the cosign private key",
	)

	signBlobCmd.PersistentFlags().StringVarP(
		&signBlobOpts.publicKeyPath,
		publicKeyPathFlag,
		"",
		"",
		"path for the cosign public key",
	)

	signBlobCmd.PersistentFlags().StringVarP(
		&signBlobOpts.certIdentity,
		certIdentityFlag,
		"",
		"",
		"The identity expected in a valid Fulcio certificate. Valid values include email address, DNS names, IP addresses, and URIs. Either --certificate-identity or --certificate-identity-regexp must be set for keyless flows.",
	)

	signBlobCmd.PersistentFlags().StringVarP(
		&signBlobOpts.certIdentityRegexp,
		certIdentityRegexpFlag,
		"",
		"",
		"A regular expression alternative to --certificate-identity. Accepts the Go regular expression syntax described at https://golang.org/s/re2syntax. Either --certificate-identity or --certificate-identity-regexp must be set for keyless flows.",
	)

	signBlobCmd.PersistentFlags().StringVarP(
		&signBlobOpts.certOidcIssuer,
		certOidcIssuerFlag,
		"",
		"",
		"The OIDC issuer expected in a valid Fulcio certificate, e.g. https://token.actions.githubusercontent.com or https://oauth2.sigstore.dev/auth. Either --certificate-oidc-issuer or --certificate-oidc-issuer-regexp must be set for keyless flows.",
	)

	signBlobCmd.PersistentFlags().StringVarP(
		&signBlobOpts.certOidcIssuerRegexp,
		certOidcIssuerRegexpFlag,
		"",
		"",
		"A regular expression alternative to --certificate-oidc-issuer. Accepts the Go regular expression syntax described at https://golang.org/s/re2syntax. Either --certificate-oidc-issuer or --certificate-oidc-issuer-regexp must be set for keyless flows.",
	)

	signCmd.AddCommand(signBlobCmd)
}

func runSignBlobs(signOpts *signOptions, signBlobOpts *signBlobOptions, args []string) (err error) {
	if err := validateSignBlobsArgs(args); err != nil {
		return fmt.Errorf("blobs to be signed does not exist: %w", err)
	}

	var tempDir string
	defer func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	}()

	var bundle []signingBundle
	isGCSBucket := false
	if strings.HasPrefix(args[0], object.GcsPrefix) {
		// GCS Bucket remote location
		isGCSBucket = true

		tempDir, err = os.MkdirTemp("", "release-sign-blobs-")
		if err != nil {
			return fmt.Errorf("creating a temporary directory to save the files to be signed: %w", err)
		}

		logrus.Infof("Getting a list of files to be signed from %s", args[0])
		output, err := gcli.GSUtilOutput("ls", "-R", args[0])
		if err != nil {
			return fmt.Errorf("listing bucket contents: %w", err)
		}

		gcsClient := object.NewGCS()
		for _, file := range strings.Fields(output) {
			if strings.HasSuffix(file, ".sha256") || strings.HasSuffix(file, ".sha512") ||
				strings.HasSuffix(file, ":") || strings.HasSuffix(file, ".docker_tag") ||
				strings.Contains(file, "SHA256SUMS") || strings.Contains(file, "SHA512SUMS") ||
				strings.Contains(file, "README") || strings.Contains(file, "Makefile") ||
				strings.HasSuffix(file, certExt) || strings.HasSuffix(file, sigExt) || strings.HasSuffix(file, ".pem") {
				continue
			}

			destinationPath := strings.TrimPrefix(file, object.GcsPrefix)
			localPath := filepath.Join(tempDir, filepath.Dir(destinationPath), filepath.Base(destinationPath))
			if err := gcsClient.CopyToLocal(file, localPath); err != nil {
				return fmt.Errorf("copying file to sign: %w", err)
			}

			bundle = append(bundle, signingBundle{
				destinationPathToCopy: filepath.Dir(destinationPath),
				fileToSign:            filepath.Base(destinationPath),
				fileLocalLocation:     localPath,
			})
		}
	} else {
		// Local files
		for _, arg := range args {
			bundle = append(bundle, signingBundle{
				fileLocalLocation: arg,
				fileToSign:        filepath.Base(arg),
			})
		}
	}

	t := throttler.New(int(signOpts.maxWorkers), len(bundle)) //nolint:gosec // overflow is highly unlikely
	for _, fileBundle := range bundle {
		go func(fileBundle signingBundle) {
			logrus.Infof("Signing %s...", fileBundle.fileToSign)
			signerOpts := sign.Default()
			signerOpts.Verbose = signOpts.verbose
			signerOpts.Timeout = signOpts.timeout
			signerOpts.PrivateKeyPath = signBlobOpts.privateKeyPath
			signerOpts.PublicKeyPath = signBlobOpts.publicKeyPath

			signerOpts.OutputCertificatePath = fmt.Sprintf("%s/%s%s", signBlobOpts.outputPath, fileBundle.fileToSign, certExt)
			signerOpts.OutputSignaturePath = fmt.Sprintf("%s/%s%s", signBlobOpts.outputPath, fileBundle.fileToSign, sigExt)
			if signBlobOpts.outputPath == "" {
				signerOpts.OutputCertificatePath = fmt.Sprintf("%s%s", fileBundle.fileLocalLocation, certExt)
				signerOpts.OutputSignaturePath = fmt.Sprintf("%s%s", fileBundle.fileLocalLocation, sigExt)
			}

			signerOpts.CertIdentity = signBlobOpts.certIdentity
			signerOpts.CertIdentityRegexp = signBlobOpts.certIdentityRegexp
			signerOpts.CertOidcIssuer = signBlobOpts.certOidcIssuer
			signerOpts.CertOidcIssuerRegexp = signBlobOpts.certOidcIssuerRegexp

			signer := sign.New(signerOpts)
			if _, err := signer.SignFile(fileBundle.fileLocalLocation); err != nil {
				t.Done(fmt.Errorf("signing the file %s: %w", fileBundle.fileLocalLocation, err))
				return
			}
			t.Done(nil)
		}(fileBundle)

		if t.Throttle() > 0 {
			break
		}
	}
	if err := t.Err(); err != nil {
		return fmt.Errorf("signing the blobs: %w", err)
	}

	if isGCSBucket {
		logrus.Info("Copying Certificates and Signatures back to the bucket...")
		for _, fileBundle := range bundle {
			certFiles := fmt.Sprintf("%s/%s%s", signBlobOpts.outputPath, fileBundle.fileToSign, certExt)
			signFiles := fmt.Sprintf("%s/%s%s", signBlobOpts.outputPath, fileBundle.fileToSign, sigExt)
			if signBlobOpts.outputPath == "" {
				certFiles = fmt.Sprintf("%s%s", fileBundle.fileLocalLocation, certExt)
				signFiles = fmt.Sprintf("%s%s", fileBundle.fileLocalLocation, sigExt)
			}

			logrus.Infof("Copying %s and %s...", certFiles, signFiles)
			if _, err := gcli.GSUtilOutput(
				"cp", certFiles, signFiles, fmt.Sprintf("%s%s", object.GcsPrefix, fileBundle.destinationPathToCopy),
			); err != nil {
				return fmt.Errorf("copying certificates and signatures to the bucket: %w", err)
			}
		}
	}

	logrus.Info("Done")
	return nil
}

func validateSignBlobsArgs(args []string) error {
	if len(args) < 1 {
		return errors.New("missing set files or gcs bucket")
	}

	if len(args) > 1 {
		tempArgs := strings.Join(args, ",")
		if strings.Count(tempArgs, object.GcsPrefix) > 0 {
			return errors.New("only one GCS Bucket is allowed and/or cannot mix with local files")
		}
	}

	if strings.HasPrefix(args[0], object.GcsPrefix) {
		return nil
	}

	for _, file := range args {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("blob %s does not exist", file)
		}
	}

	return nil
}

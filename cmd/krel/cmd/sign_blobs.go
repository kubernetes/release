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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

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

	// signaturePath and certificatePath name the signature and certificate
	// of a previous run, if both exist.
	signaturePath   string
	certificatePath string
}

// fileVerifier verifies the signature and certificate of a file.
type fileVerifier interface {
	VerifyFile(path string, ignoreTLog bool) (*sign.SignedObject, error)
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

		objects := strings.Fields(output)

		listed := make(map[string]struct{}, len(objects))
		for _, file := range objects {
			listed[file] = struct{}{}
		}

		for _, file := range objects {
			if !isSignable(file) {
				continue
			}

			destinationPath := strings.TrimPrefix(file, object.GcsPrefix)

			localPath := filepath.Join(tempDir, filepath.Dir(destinationPath), filepath.Base(destinationPath))
			if err := gcsClient.CopyToLocal(file, localPath); err != nil {
				return fmt.Errorf("copying file to sign: %w", err)
			}

			fileBundle := signingBundle{
				destinationPathToCopy: filepath.Dir(destinationPath),
				fileToSign:            filepath.Base(destinationPath),
				fileLocalLocation:     localPath,
			}

			// Download the signature of a previous run, so that it gets
			// verified instead of signing the file again.
			if canVerify(signBlobOpts) && hasSignatureObjects(listed, file) {
				fileBundle.signaturePath = localPath + sigExt
				fileBundle.certificatePath = localPath + certExt

				if _, err := gcli.GSUtilOutput(
					"cp", file+sigExt, file+certExt, filepath.Dir(localPath)+string(filepath.Separator),
				); err != nil {
					return fmt.Errorf("copying existing signature of %s: %w", file, err)
				}
			}

			bundle = append(bundle, fileBundle)
		}
	} else {
		// Local files
		for _, arg := range args {
			fileBundle := signingBundle{
				fileLocalLocation: arg,
				fileToSign:        filepath.Base(arg),
			}

			// A previous run wrote its signature where this one writes it.
			signaturePath, certificatePath := outputPaths(signBlobOpts, arg)
			if fileExists(signaturePath) && fileExists(certificatePath) {
				fileBundle.signaturePath = signaturePath
				fileBundle.certificatePath = certificatePath
			}

			bundle = append(bundle, fileBundle)
		}
	}

	// Files whose existing signature verifies are not signed again, so
	// that a rerun keeps the signatures of a previous run. A file that
	// changed since, for example because the stage ran again, gets a new
	// signature, as its old one doesn't verify anymore.
	signed := make([]bool, len(bundle))

	g := new(errgroup.Group)
	g.SetLimit(int(signOpts.maxWorkers)) //nolint:gosec // overflow is highly unlikely

	for i, fileBundle := range bundle {
		g.Go(func() error {
			if !needsSigning(&fileBundle, signBlobOpts, func(signaturePath, certificatePath string) fileVerifier {
				return newBlobVerifier(signOpts, signBlobOpts, signaturePath, certificatePath)
			}) {
				logrus.Infof("Skipping %s, it is signed already", fileBundle.fileToSign)

				return keepSignature(signBlobOpts, &fileBundle)
			}

			logrus.Infof("Signing %s...", fileBundle.fileToSign)

			signerOpts := sign.Default()
			signerOpts.Verbose = signOpts.verbose
			signerOpts.Timeout = signOpts.timeout
			signerOpts.PrivateKeyPath = signBlobOpts.privateKeyPath
			signerOpts.PublicKeyPath = signBlobOpts.publicKeyPath

			signerOpts.OutputSignaturePath, signerOpts.OutputCertificatePath = outputPaths(signBlobOpts, fileBundle.fileLocalLocation)

			signerOpts.CertIdentity = signBlobOpts.certIdentity
			signerOpts.CertIdentityRegexp = signBlobOpts.certIdentityRegexp
			signerOpts.CertOidcIssuer = signBlobOpts.certOidcIssuer
			signerOpts.CertOidcIssuerRegexp = signBlobOpts.certOidcIssuerRegexp

			signer := sign.New(signerOpts)
			if _, err := signer.SignFile(fileBundle.fileLocalLocation); err != nil {
				return fmt.Errorf("signing the file %s: %w", fileBundle.fileLocalLocation, err)
			}

			signed[i] = true

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("signing the blobs: %w", err)
	}

	if isGCSBucket {
		logrus.Info("Copying Certificates and Signatures back to the bucket...")

		for _, fileBundle := range signedBundles(bundle, signed) {
			signFiles, certFiles := outputPaths(signBlobOpts, fileBundle.fileLocalLocation)

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

// isSignable reports whether a file listed in the bucket gets signed.
func isSignable(file string) bool {
	for _, suffix := range []string{".sha256", ".sha512", ":", ".docker_tag", certExt, sigExt, ".pem"} {
		if strings.HasSuffix(file, suffix) {
			return false
		}
	}

	for _, name := range []string{"SHA256SUMS", "SHA512SUMS", "README", "Makefile"} {
		if strings.Contains(file, name) {
			return false
		}
	}

	return true
}

// hasSignatureObjects reports whether the listed objects contain a
// signature and a certificate for the file.
func hasSignatureObjects(listed map[string]struct{}, file string) bool {
	_, hasSignature := listed[file+sigExt]
	_, hasCertificate := listed[file+certExt]

	return hasSignature && hasCertificate
}

// fileExists reports whether the path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// canVerify reports whether an existing signature can be verified: only
// keyless signatures are, against the identity and issuer of the options.
// Signing with a key writes no certificate, so those files are always
// signed again.
func canVerify(opts *signBlobOptions) bool {
	if opts.privateKeyPath != "" {
		return false
	}

	return (opts.certIdentity != "" || opts.certIdentityRegexp != "") &&
		(opts.certOidcIssuer != "" || opts.certOidcIssuerRegexp != "")
}

// outputPaths returns where the signature and certificate of the file are
// written: next to it, or in the output path.
func outputPaths(opts *signBlobOptions, path string) (signaturePath, certificatePath string) {
	if opts.outputPath == "" {
		return path + sigExt, path + certExt
	}

	return filepath.Join(opts.outputPath, filepath.Base(path)+sigExt), filepath.Join(opts.outputPath, filepath.Base(path)+certExt)
}

// newBlobVerifier returns a verifier for the signature and certificate of a
// previous run of the file.
func newBlobVerifier(signOpts *signOptions, opts *signBlobOptions, signaturePath, certificatePath string) *sign.Signer {
	verifierOpts := sign.Default()
	verifierOpts.Verbose = signOpts.verbose
	verifierOpts.Timeout = signOpts.timeout
	verifierOpts.OutputSignaturePath = signaturePath
	verifierOpts.OutputCertificatePath = certificatePath
	verifierOpts.CertIdentity = opts.certIdentity
	verifierOpts.CertIdentityRegexp = opts.certIdentityRegexp
	verifierOpts.CertOidcIssuer = opts.certOidcIssuer
	verifierOpts.CertOidcIssuerRegexp = opts.certOidcIssuerRegexp

	return sign.New(verifierOpts)
}

// needsSigning reports whether the file has to be signed: unless the
// signature and certificate of a previous run exist and verify.
func needsSigning(
	fileBundle *signingBundle, opts *signBlobOptions, newVerifier func(signaturePath, certificatePath string) fileVerifier,
) bool {
	if fileBundle.signaturePath == "" || !canVerify(opts) {
		return true
	}

	return !signatureVerifies(newVerifier(fileBundle.signaturePath, fileBundle.certificatePath), fileBundle.fileLocalLocation)
}

// keepSignature puts the verified signature and certificate of a previous
// run where this run writes them, so that the output path holds the
// signatures of all files, also the skipped ones.
func keepSignature(opts *signBlobOptions, fileBundle *signingBundle) error {
	signaturePath, certificatePath := outputPaths(opts, fileBundle.fileLocalLocation)

	for src, dst := range map[string]string{
		fileBundle.signaturePath:   signaturePath,
		fileBundle.certificatePath: certificatePath,
	} {
		if src == dst {
			continue
		}

		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("reading the existing signature of %s: %w", fileBundle.fileToSign, err)
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("creating the output path: %w", err)
		}

		if err := os.WriteFile(dst, data, 0o644); err != nil { //nolint:gosec // signatures and certificates are public
			return fmt.Errorf("writing the existing signature of %s: %w", fileBundle.fileToSign, err)
		}
	}

	return nil
}

// signedBundles returns the files that were signed in this run.
func signedBundles(bundle []signingBundle, signed []bool) []signingBundle {
	res := make([]signingBundle, 0, len(bundle))

	for i := range bundle {
		if signed[i] {
			res = append(res, bundle[i])
		}
	}

	return res
}

// signatureVerifies reports whether the existing signature of the file
// verifies for its current content, including its transparency log entry.
// Any error, also a temporary one like an unavailable transparency log,
// leads to signing the file again: that is always correct and only costs
// time, while failing would block the release. Errors can't be told apart
// reliably either, Rekor rejects the search for a changed file with a bad
// request instead of an empty result.
func signatureVerifies(verifier fileVerifier, path string) bool {
	signed, err := verifier.VerifyFile(path, false)
	if err != nil {
		logrus.Warnf("Existing signature of %s does not verify, signing it again: %v", path, err)

		return false
	}

	return signed != nil
}

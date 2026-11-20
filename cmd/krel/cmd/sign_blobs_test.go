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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"sigs.k8s.io/release-sdk/sign"
	"sigs.k8s.io/release-sdk/sign/signfakes"
)

func TestIsSignable(t *testing.T) {
	t.Parallel()

	for file, signable := range map[string]bool{
		"gs://bucket/stage/v1.38.0/kubernetes.tar.gz":           true,
		"gs://bucket/stage/v1.38.0/bin/linux/amd64/kubectl":     true,
		"gs://bucket/stage/v1.38.0/kubernetes-release.spdx":     true,
		"gs://bucket/stage/v1.38.0/kubernetes.tar.gz.sig":       false,
		"gs://bucket/stage/v1.38.0/kubernetes.tar.gz.cert":      false,
		"gs://bucket/stage/v1.38.0/kubernetes.tar.gz.sha256":    false,
		"gs://bucket/stage/v1.38.0/SHA512SUMS":                  false,
		"gs://bucket/stage/v1.38.0/bin/linux/amd64:":            false,
		"gs://bucket/stage/v1.38.0/bin/linux/amd64/kubectl.pem": false,
	} {
		require.Equal(t, signable, isSignable(file), file)
	}
}

func TestHasSignatureObjects(t *testing.T) {
	t.Parallel()

	const file = "gs://bucket/stage/v1.38.0/kubernetes.tar.gz"

	listed := func(objects ...string) map[string]struct{} {
		res := map[string]struct{}{}
		for _, o := range objects {
			res[o] = struct{}{}
		}

		return res
	}

	require.True(t, hasSignatureObjects(listed(file, file+sigExt, file+certExt), file))
	require.False(t, hasSignatureObjects(listed(file, file+sigExt), file), "a signature without certificate")
	require.False(t, hasSignatureObjects(listed(file, file+certExt), file), "a certificate without signature")
	require.False(t, hasSignatureObjects(listed(file), file))
}

func TestOutputPaths(t *testing.T) {
	t.Parallel()

	sig, cert := outputPaths(&signBlobOptions{}, "/tmp/stage/kubectl")
	require.Equal(t, "/tmp/stage/kubectl.sig", sig)
	require.Equal(t, "/tmp/stage/kubectl.cert", cert)

	sig, cert = outputPaths(&signBlobOptions{outputPath: "/tmp/out"}, "/tmp/stage/kubectl")
	require.Equal(t, "/tmp/out/kubectl.sig", sig)
	require.Equal(t, "/tmp/out/kubectl.cert", cert)
}

func TestKeepSignature(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	downloaded := filepath.Join(dir, "stage", "kubectl")
	require.NoError(t, os.MkdirAll(filepath.Dir(downloaded), 0o755))
	require.NoError(t, os.WriteFile(downloaded+sigExt, []byte("signature"), 0o600))
	require.NoError(t, os.WriteFile(downloaded+certExt, []byte("certificate"), 0o600))

	fileBundle := &signingBundle{
		fileToSign:        "kubectl",
		fileLocalLocation: downloaded,
		signaturePath:     downloaded + sigExt,
		certificatePath:   downloaded + certExt,
	}

	// Without an output path, the signature stays where it is.
	require.NoError(t, keepSignature(&signBlobOptions{}, fileBundle))

	// With an output path, the skipped file gets its signature there too.
	out := filepath.Join(dir, "out")
	require.NoError(t, keepSignature(&signBlobOptions{outputPath: out}, fileBundle))

	sig, err := os.ReadFile(filepath.Join(out, "kubectl"+sigExt))
	require.NoError(t, err)
	require.Equal(t, "signature", string(sig))

	cert, err := os.ReadFile(filepath.Join(out, "kubectl"+certExt))
	require.NoError(t, err)
	require.Equal(t, "certificate", string(cert))
}

func TestCanVerify(t *testing.T) {
	t.Parallel()

	require.True(t, canVerify(&signBlobOptions{certIdentity: "krel-staging", certOidcIssuer: "https://accounts.google.com"}))
	require.True(t, canVerify(&signBlobOptions{certIdentityRegexp: "krel-.*", certOidcIssuerRegexp: ".*"}))
	require.False(t, canVerify(&signBlobOptions{certIdentity: "krel-staging"}), "no issuer")
	require.False(t, canVerify(&signBlobOptions{certOidcIssuer: "https://accounts.google.com"}), "no identity")
	require.False(t, canVerify(&signBlobOptions{}))
	require.False(t, canVerify(&signBlobOptions{
		privateKeyPath: "key", publicKeyPath: "key.pub", certIdentity: "krel-staging", certOidcIssuer: "https://accounts.google.com",
	}), "signing with a key writes no certificate")
}

type fakeVerifier struct {
	signed *sign.SignedObject
	err    error
	calls  int
}

func (f *fakeVerifier) VerifyFile(string, bool) (*sign.SignedObject, error) {
	f.calls++

	return f.signed, f.err
}

func TestNeedsSigning(t *testing.T) {
	t.Parallel()

	keyless := &signBlobOptions{certIdentity: "krel-staging", certOidcIssuer: "https://accounts.google.com"}
	previous := signingBundle{fileLocalLocation: "kubectl", signaturePath: "kubectl.sig", certificatePath: "kubectl.cert"}

	for _, tc := range []struct {
		name          string
		fileBundle    signingBundle
		opts          *signBlobOptions
		verifier      *fakeVerifier
		expectedSign  bool
		expectedCalls int
	}{
		{
			name:          "a valid signature of a previous run",
			fileBundle:    previous,
			opts:          keyless,
			verifier:      &fakeVerifier{signed: &sign.SignedObject{}},
			expectedSign:  false,
			expectedCalls: 1,
		},
		{
			name:          "no signature of a previous run",
			fileBundle:    signingBundle{fileLocalLocation: "kubectl"},
			opts:          keyless,
			verifier:      &fakeVerifier{signed: &sign.SignedObject{}},
			expectedSign:  true,
			expectedCalls: 0,
		},
		{
			name:          "no identity to verify against",
			fileBundle:    previous,
			opts:          &signBlobOptions{},
			verifier:      &fakeVerifier{signed: &sign.SignedObject{}},
			expectedSign:  true,
			expectedCalls: 0,
		},
		{
			name:          "the content changed since",
			fileBundle:    previous,
			opts:          keyless,
			verifier:      &fakeVerifier{},
			expectedSign:  true,
			expectedCalls: 1,
		},
		{
			name:          "the signature does not verify",
			fileBundle:    previous,
			opts:          keyless,
			verifier:      &fakeVerifier{err: errors.New("no matching CertificateIdentity found")},
			expectedSign:  true,
			expectedCalls: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var signaturePath, certificatePath string

			require.Equal(t, tc.expectedSign, needsSigning(&tc.fileBundle, tc.opts, func(sig, cert string) fileVerifier {
				signaturePath, certificatePath = sig, cert

				return tc.verifier
			}))
			require.Equal(t, tc.expectedCalls, tc.verifier.calls)

			if tc.expectedCalls > 0 {
				require.Equal(t, tc.fileBundle.signaturePath, signaturePath)
				require.Equal(t, tc.fileBundle.certificatePath, certificatePath)
			}
		})
	}
}

func TestSignedBundles(t *testing.T) {
	t.Parallel()

	bundle := []signingBundle{{fileToSign: "kubectl"}, {fileToSign: "kubeadm"}, {fileToSign: "kubelet"}}

	require.Equal(t,
		[]signingBundle{{fileToSign: "kubectl"}, {fileToSign: "kubelet"}},
		signedBundles(bundle, []bool{true, false, true}),
	)
	require.Empty(t, signedBundles(bundle, []bool{false, false, false}))
}

func TestSignatureVerifies(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		hasSignature  bool
		verifyErr     error
		expectedValid bool
	}{
		{
			name:          "the existing signature verifies",
			hasSignature:  true,
			expectedValid: true,
		},
		{
			name:         "no signature next to the file",
			hasSignature: false,
		},
		{
			// A file that changed since it was signed, for example by
			// running the stage again, has no transparency log entry for
			// its new digest.
			name:         "no transparency log entry for the current content",
			hasSignature: true,
			verifyErr:    sign.ErrTlogEntryNotFound,
		},
		{
			name:         "signed by another identity",
			hasSignature: true,
			verifyErr:    errors.New("no matching CertificateIdentity found"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "kubectl")
			require.NoError(t, os.WriteFile(path, []byte("kubectl"), 0o600))

			impl := &signfakes.FakeImpl{}
			impl.FileExistsReturns(tc.hasSignature)
			impl.VerifyFileInternalReturns(tc.verifyErr)

			verifier := newBlobVerifier(
				&signOptions{},
				&signBlobOptions{
					certIdentity: "krel-staging", certOidcIssuer: "https://accounts.google.com", publicKeyPath: "unused.pub",
				},
				path+sigExt, path+certExt,
			)
			verifier.SetImpl(impl)

			require.Equal(t, tc.expectedValid, signatureVerifies(verifier, path))

			if tc.hasSignature {
				// The signature next to the file is verified, with its
				// transparency log entry.
				require.Equal(t, 1, impl.VerifyFileInternalCallCount())

				_, opts, _, withTLog := impl.VerifyFileInternalArgsForCall(0)
				require.Equal(t, path+sigExt, opts.OutputSignaturePath)
				require.Equal(t, path+certExt, opts.OutputCertificatePath)
				require.Equal(t, "krel-staging", opts.CertIdentity)
				require.Empty(t, opts.PublicKeyPath, "a keyless signature is verified against its certificate")
				require.True(t, withTLog)
			}
		})
	}
}

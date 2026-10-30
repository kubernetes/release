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
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/policylabs/signer"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protodsse "github.com/sigstore/protobuf-specs/gen/pb-go/dsse"
	sbundle "github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/attestation"
	"k8s.io/release/pkg/attestation/attestationfakes"
)

func TestValidateSignAttestationArgs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		opts       *signAttestationOptions
		statements []string
		shouldErr  bool
	}{
		{name: "single statement", opts: &signAttestationOptions{}, statements: []string{"a.json"}},
		{
			name:       "single statement with output path",
			opts:       &signAttestationOptions{outputPath: "out.json"},
			statements: []string{"a.json"},
		},
		{
			name:       "single statement in place with output path",
			opts:       &signAttestationOptions{inPlace: true, outputPath: "out.json"},
			statements: []string{"a.json"},
		},
		{
			name:       "several statements in place",
			opts:       &signAttestationOptions{inPlace: true},
			statements: []string{"a.json", "gs://bucket/b.json"},
		},
		{
			name:       "several statements without in place",
			opts:       &signAttestationOptions{},
			statements: []string{"a.json", "b.json"},
			shouldErr:  true,
		},
		{
			name:       "several statements in place with output path",
			opts:       &signAttestationOptions{inPlace: true, outputPath: "out.json"},
			statements: []string{"a.json", "b.json"},
			shouldErr:  true,
		},
		{
			name: "attaching one of several statements",
			opts: &signAttestationOptions{
				inPlace: true, attachToImages: []string{"b.json"}, imageRepository: "gcr.io/k8s-staging-kubernetes",
			},
			statements: []string{"a.json", "b.json"},
		},
		{
			name:       "attaching a statement that isn't signed",
			opts:       &signAttestationOptions{attachToImages: []string{"b.json"}, imageRepository: "gcr.io/k8s-staging-kubernetes"},
			statements: []string{"a.json"},
			shouldErr:  true,
		},
		{
			name:       "attaching without a repository",
			opts:       &signAttestationOptions{attachToImages: []string{"a.json"}},
			statements: []string{"a.json"},
			shouldErr:  true,
		},
		{
			name:       "a repository without attaching",
			opts:       &signAttestationOptions{imageRepository: "gcr.io/k8s-staging-kubernetes"},
			statements: []string{"a.json"},
			shouldErr:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateSignAttestationArgs(tc.opts, tc.statements)
			if tc.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRunSignAttestation(t *testing.T) {
	t.Parallel()

	signOpts := &signOptions{timeout: time.Second}

	t.Run("missing statement", func(t *testing.T) {
		t.Parallel()

		err := runSignAttestation(
			signOpts, &signAttestationOptions{}, []string{filepath.Join(t.TempDir(), "missing.json")},
		)
		require.ErrorContains(t, err, "reading statement")
	})

	t.Run("output file is not touched when signing fails", func(t *testing.T) {
		t.Parallel()

		outputPath := filepath.Join(t.TempDir(), "out.json")

		err := runSignAttestation(
			signOpts,
			&signAttestationOptions{outputPath: outputPath},
			[]string{filepath.Join(t.TempDir(), "missing.json")},
		)
		require.ErrorContains(t, err, "signing attestation")
		require.NoFileExists(t, outputPath)
	})

	t.Run("in place refuses an already signed file and leaves it untouched", func(t *testing.T) {
		t.Parallel()

		original := []byte(`{"mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json", "dsseEnvelope": {}}`)

		bundlePath := filepath.Join(t.TempDir(), "provenance.json")
		require.NoError(t, os.WriteFile(bundlePath, original, 0o600))

		err := runSignAttestation(signOpts, &signAttestationOptions{inPlace: true}, []string{bundlePath})
		require.ErrorIs(t, err, attestation.ErrAlreadySigned)

		data, err := os.ReadFile(bundlePath)
		require.NoError(t, err)
		require.Equal(t, original, data)
	})

	t.Run("several statements without in place", func(t *testing.T) {
		t.Parallel()

		err := runSignAttestation(signOpts, &signAttestationOptions{}, []string{"a.json", "b.json"})
		require.ErrorContains(t, err, "--in-place")
	})

	t.Run("invalid service account key from the environment", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		statement := filepath.Join(dir, "statement.json")
		require.NoError(t, os.WriteFile(statement, []byte(`{
			"_type": "https://in-toto.io/Statement/v1",
			"subject": [{"name": "a", "digest": {"sha256": "0e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7"}}],
			"predicateType": "https://example.com/test", "predicate": {}
		}`), 0o600))

		err := runSignAttestation(
			signOpts, &signAttestationOptions{serviceAccountJSON: `{"type": "authorized_user"}`}, []string{statement},
		)
		require.ErrorContains(t, err, "not a service account key")
	})

	t.Run("invalid service account to impersonate", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		statement := filepath.Join(dir, "statement.json")
		require.NoError(t, os.WriteFile(statement, []byte(`{
			"_type": "https://in-toto.io/Statement/v1",
			"subject": [{"name": "a", "digest": {"sha256": "0e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7"}}],
			"predicateType": "https://example.com/test", "predicate": {}
		}`), 0o600))

		err := runSignAttestation(
			signOpts, &signAttestationOptions{impersonateServiceAccount: "not-an-email"}, []string{statement},
		)
		require.ErrorContains(t, err, "not a valid service account email")
	})
}

// imageStatementJSON is a statement about one staged image.
const imageStatementJSON = `{
	"_type": "https://in-toto.io/Statement/v1",
	"subject": [{"name": "gcr.io/k8s-staging-kubernetes/kube-apiserver", "digest": {"sha256": "0e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7"}}],
	"predicateType": "https://slsa.dev/provenance/v1", "predicate": {}
}`

func TestSignAttestationsAttachesBeforeWriting(t *testing.T) {
	t.Parallel()

	const (
		provenance      = "gs://bucket/stage/v1.36.0/provenance.json"
		imageProvenance = "gs://bucket/stage/v1.36.0/image-provenance.json"
	)

	for _, tc := range []struct {
		name      string
		attachErr error
		calls     []string
	}{
		{
			name: "attaches the image provenance, then replaces both statements",
			calls: []string{
				"attach gcr.io/k8s-staging-kubernetes/kube-apiserver@sha256:0e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7",
				"write " + provenance,
				"write " + imageProvenance,
			},
		},
		{
			name:      "a failed attach leaves the statements in the bucket untouched",
			attachErr: errors.New("registry down"),
			calls: []string{
				"attach gcr.io/k8s-staging-kubernetes/kube-apiserver@sha256:0e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				mu    sync.Mutex
				calls []string
			)

			record := func(call string) {
				mu.Lock()
				defer mu.Unlock()

				calls = append(calls, call)
			}

			signerImpl := &attestationfakes.FakeSignerImplementation{}
			signerImpl.NewSignerReturns(signer.NewSigner())
			signerImpl.ReadStatementReturns([]byte(imageStatementJSON), nil)
			signerImpl.SignStatementCalls(func(_ *signer.Signer, data []byte) (*sbundle.Bundle, error) {
				return &sbundle.Bundle{Bundle: &protobundle.Bundle{
					Content: &protobundle.Bundle_DsseEnvelope{DsseEnvelope: &protodsse.Envelope{Payload: data}},
				}}, nil
			})
			signerImpl.WriteBundleCalls(func(_ *sbundle.Bundle, w io.Writer) error {
				_, err := w.Write([]byte("bundle"))

				return err
			})
			signerImpl.WriteFileCalls(func(path string, _ []byte) error {
				record("write " + path)

				return nil
			})

			attacherImpl := &attestationfakes.FakeAttacherImplementation{}
			attacherImpl.AttachBundleCalls(func(_ context.Context, ref name.Digest, _ []byte, _ string) error {
				record("attach " + ref.String())

				return tc.attachErr
			})

			sut := attestation.NewSigner(attestation.DefaultSignerOptions())
			sut.SetImpl(signerImpl)

			attacher := attestation.NewAttacher(&attestation.AttacherOptions{Repository: "gcr.io/k8s-staging-kubernetes"})
			attacher.SetImpl(attacherImpl)

			// Only the image provenance is attached, though both statements
			// name the same image here.
			err := signAttestations(sut, attacher, &signAttestationOptions{
				inPlace:         true,
				attachToImages:  []string{imageProvenance},
				imageRepository: "gcr.io/k8s-staging-kubernetes",
			}, []string{provenance, imageProvenance})

			if tc.attachErr != nil {
				require.ErrorIs(t, err, tc.attachErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.calls, calls)
		})
	}
}

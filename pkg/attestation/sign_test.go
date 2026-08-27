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

package attestation

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/carabiner-dev/signer"
	sbundle "github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore/pkg/oauthflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/attestation/attestationfakes"
)

const testStatement = `{
  "_type": "https://in-toto.io/Statement/v1",
  "subject": [{"name": "kubernetes.tar.gz", "digest": {"sha256": "0e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7"}}],
  "predicateType": "https://slsa.dev/provenance/v1",
  "predicate": {"buildDefinition": {"buildType": "https://git.k8s.io/release/docs/krel/buildtypes/v1"}}
}`

var errTest = errors.New("synthetic error")

func TestSignFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		opts       *SignerOptions
		prepare    func(*attestationfakes.FakeSignerImplementation)
		shouldErr  bool
		wantOutput string
		wantSAToks int
	}{
		{
			name: "success with ambient credentials",
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.WriteBundleCalls(func(_ *signer.Signer, _ *sbundle.Bundle, w io.Writer) error {
					_, err := w.Write([]byte("bundle"))

					return err
				})
			},
			wantOutput: "bundle",
		},
		{
			name: "success with service account key",
			opts: &SignerOptions{ServiceAccountFile: "key.json"},
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.ServiceAccountTokenReturns(&oauthflow.OIDCIDToken{Subject: "sa@example.com"}, nil)
				mock.WriteBundleCalls(func(_ *signer.Signer, _ *sbundle.Bundle, w io.Writer) error {
					_, err := w.Write([]byte("bundle"))

					return err
				})
			},
			wantOutput: "bundle",
			wantSAToks: 1,
		},
		{
			name: "success with service account JSON",
			opts: &SignerOptions{ServiceAccountJSON: []byte("service-account-key-data")},
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.ServiceAccountTokenReturns(&oauthflow.OIDCIDToken{Subject: "sa@example.com"}, nil)
				mock.WriteBundleCalls(func(_ *signer.Signer, _ *sbundle.Bundle, w io.Writer) error {
					_, err := w.Write([]byte("bundle"))

					return err
				})
			},
			wantOutput: "bundle",
			wantSAToks: 1,
		},
		{
			name: "ReadStatement fails",
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.ReadStatementReturns(nil, errTest)
			},
			shouldErr: true,
		},
		{
			name: "ServiceAccountToken fails",
			opts: &SignerOptions{ServiceAccountFile: "key.json"},
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.ServiceAccountTokenReturns(nil, errTest)
			},
			shouldErr:  true,
			wantSAToks: 1,
		},
		{
			name: "SignStatement fails",
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.SignStatementReturns(nil, errTest)
			},
			shouldErr: true,
		},
		{
			name: "WriteBundle fails",
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.WriteBundleReturns(errTest)
			},
			shouldErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &attestationfakes.FakeSignerImplementation{}
			mock.NewSignerReturns(signer.NewSigner())
			tc.prepare(mock)

			sut := NewSigner(tc.opts)
			sut.impl = mock

			var out bytes.Buffer

			err := sut.SignFile("statement.json", &out)
			if tc.shouldErr {
				require.Error(t, err)
				require.ErrorIs(t, err, errTest)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantOutput, out.String())
			}

			require.Equal(t, tc.wantSAToks, mock.ServiceAccountTokenCallCount())

			if tc.wantSAToks > 0 {
				wantKeyData := tc.opts.ServiceAccountJSON
				_, keyFile, keyData, audience := mock.ServiceAccountTokenArgsForCall(0)
				require.Equal(t, tc.opts.ServiceAccountFile, keyFile)
				require.Equal(t, wantKeyData, keyData)
				require.NotEmpty(t, audience, "audience must be the sigstore OIDC client ID")
			}
		})
	}
}

func TestSignStatementLocksToken(t *testing.T) {
	t.Parallel()

	// The signer must be pinned to the token so it cannot fall back to any
	// other credential. We verify the pinning without signing (which would
	// need network access) by checking the options the signer ends up with.
	sgnr := signer.NewSigner()
	require.Nil(t, sgnr.Options.Token)
	require.False(t, sgnr.Options.DisableSTS)

	token := &oauthflow.OIDCIDToken{RawString: "not-a-token", Subject: "sa@example.com"}
	sgnr.Options.Token = token
	sgnr.Options.DisableSTS = true

	creds := sgnr.Options.BuildSigstoreCredentials()
	require.Same(t, token, creds.Token)
	require.True(t, creds.DisableSTS)
}

// fakeObjectStore is an objectStore that serves a fixed object and records
// what gets uploaded.
type fakeObjectStore struct {
	data     []byte
	err      error
	gcsPath  string
	dst      string
	src      string
	uploaded []byte
}

func (f *fakeObjectStore) CopyToLocal(gcsPath, dst string) error {
	f.gcsPath, f.dst = gcsPath, dst
	if f.err != nil {
		return f.err
	}

	return os.WriteFile(dst, f.data, 0o600)
}

func (f *fakeObjectStore) CopyToRemote(src, gcsPath string) error {
	f.src, f.gcsPath = src, gcsPath
	if f.err != nil {
		return f.err
	}

	data, err := os.ReadFile(src)
	f.uploaded = data

	return err
}

func TestWriteFile(t *testing.T) {
	t.Parallel()

	const gcsPath = "gs://bucket/stage/v1.36.0-alpha.1/provenance.json.sigstore.json"

	want := []byte("bundle")

	t.Run("local file", func(t *testing.T) {
		t.Parallel()

		filePath := filepath.Join(t.TempDir(), "bundle.json")
		store := &fakeObjectStore{}

		require.NoError(t, (&defaultSignerImpl{gcs: store}).WriteFile(filePath, want))

		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, want, data)
		require.Empty(t, store.gcsPath, "local files must not touch GCS")
	})

	t.Run("upload to GCS", func(t *testing.T) {
		t.Parallel()

		store := &fakeObjectStore{}

		require.NoError(t, (&defaultSignerImpl{gcs: store}).WriteFile(gcsPath, want))
		require.Equal(t, gcsPath, store.gcsPath)
		require.Equal(t, want, store.uploaded)
		require.Equal(t, "provenance.json.sigstore.json", filepath.Base(store.src))
		require.NoFileExists(t, store.src, "temporary upload must be cleaned up")
	})

	t.Run("upload fails", func(t *testing.T) {
		t.Parallel()

		store := &fakeObjectStore{err: errTest}

		require.ErrorIs(t, (&defaultSignerImpl{gcs: store}).WriteFile(gcsPath, want), errTest)
	})
}

func TestReadStatementFromGCS(t *testing.T) {
	t.Parallel()

	const gcsPath = "gs://bucket/stage/v1.36.0-alpha.1/provenance.json"

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		store := &fakeObjectStore{data: []byte(testStatement)}

		want := []byte(testStatement)

		data, err := (&defaultSignerImpl{gcs: store}).ReadStatement(gcsPath)
		require.NoError(t, err)
		require.Equal(t, want, data)
		require.Equal(t, gcsPath, store.gcsPath)
		require.Equal(t, "provenance.json", filepath.Base(store.dst))
		require.NoFileExists(t, store.dst, "temporary download must be cleaned up")
	})

	t.Run("download fails", func(t *testing.T) {
		t.Parallel()

		store := &fakeObjectStore{err: errTest}

		data, err := (&defaultSignerImpl{gcs: store}).ReadStatement(gcsPath)
		require.ErrorIs(t, err, errTest)
		require.Nil(t, data)
	})

	t.Run("object is not a statement", func(t *testing.T) {
		t.Parallel()

		store := &fakeObjectStore{data: []byte("not a statement")}

		data, err := (&defaultSignerImpl{gcs: store}).ReadStatement(gcsPath)
		require.Error(t, err)
		require.Nil(t, data)
	})
}

func TestReadStatement(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		content   string
		missing   bool
		shouldErr bool
	}{
		{name: "valid statement", content: testStatement},
		{name: "missing file", missing: true, shouldErr: true},
		{name: "not json", content: "this is not json", shouldErr: true},
		{name: "not a statement", content: `{"foo": "bar"}`, shouldErr: true},
		{
			name: "no subjects",
			content: `{"_type": "https://in-toto.io/Statement/v1", "subject": [],
			"predicateType": "https://slsa.dev/provenance/v1", "predicate": {}}`,
			shouldErr: true,
		},
		{
			name:      "no predicate type",
			content:   `{"_type": "https://in-toto.io/Statement/v1", "subject": [{"name": "a"}], "predicate": {}}`,
			shouldErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "statement.json")
			if !tc.missing {
				require.NoError(t, os.WriteFile(path, []byte(tc.content), 0o600))
			}

			data, err := (&defaultSignerImpl{}).ReadStatement(path)
			if tc.shouldErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			// The statement bytes are signed verbatim, they must not be altered.
			require.Equal(t, tc.content, string(data))
		})
	}
}

// fakeJWT returns an unsigned JWT with the given subject claim, enough for
// the provider to extract the identity from the token endpoint response.
func fakeJWT(t *testing.T, subject string) string {
	t.Helper()

	claims, err := json.Marshal(map[string]any{"sub": subject, "email": subject, "email_verified": true})
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`)) +
		"." + base64.RawURLEncoding.EncodeToString(claims) + ".c2ln"
}

// writeServiceAccountKey writes a service account key file whose token
// endpoint is tokenURI.
func writeServiceAccountKey(t *testing.T, credType, tokenURI string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "key.json")
	require.NoError(t, os.WriteFile(path, serviceAccountKey(t, credType, tokenURI), 0o600))

	return path
}

// serviceAccountKey returns the JSON of a service account key whose token
// endpoint is tokenURI.
func serviceAccountKey(t *testing.T, credType, tokenURI string) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Google issues service account keys in PKCS#8 format.
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	keyData, err := json.Marshal(map[string]string{
		"type":           credType,
		"client_email":   "signer@example.iam.gserviceaccount.com",
		"private_key_id": "key-id",
		"private_key":    string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})),
		"token_uri":      tokenURI,
	})
	require.NoError(t, err)

	return keyData
}

func TestServiceAccountToken(t *testing.T) {
	t.Parallel()

	const subject = "signer@example.iam.gserviceaccount.com"

	newTokenServer := func(t *testing.T, status int, body string) *httptest.Server {
		t.Helper()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.NoError(t, r.ParseForm())
			assert.Equal(t, "urn:ietf:params:oauth:grant-type:jwt-bearer", r.Form.Get("grant_type"))
			assert.NotEmpty(t, r.Form.Get("assertion"))

			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		}))
		t.Cleanup(srv.Close)

		return srv
	}

	t.Run("success with key file", func(t *testing.T) {
		t.Parallel()

		srv := newTokenServer(t, http.StatusOK, `{"id_token": "`+fakeJWT(t, subject)+`"}`)
		opts := &SignerOptions{ServiceAccountFile: writeServiceAccountKey(t, "service_account", srv.URL)}

		token, err := (&defaultSignerImpl{}).ServiceAccountToken(
			context.Background(), opts.ServiceAccountFile, opts.ServiceAccountJSON, "sigstore",
		)
		require.NoError(t, err)
		require.NotNil(t, token)
		require.Equal(t, subject, token.Subject)
		require.NotEmpty(t, token.RawString)
	})

	t.Run("success with key JSON", func(t *testing.T) {
		t.Parallel()

		srv := newTokenServer(t, http.StatusOK, `{"id_token": "`+fakeJWT(t, subject)+`"}`)
		opts := &SignerOptions{ServiceAccountJSON: serviceAccountKey(t, "service_account", srv.URL)}

		token, err := (&defaultSignerImpl{}).ServiceAccountToken(
			context.Background(), opts.ServiceAccountFile, opts.ServiceAccountJSON, "sigstore",
		)
		require.NoError(t, err)
		require.NotNil(t, token)
		require.Equal(t, subject, token.Subject)
	})

	t.Run("key file takes precedence over key JSON", func(t *testing.T) {
		t.Parallel()

		srv := newTokenServer(t, http.StatusOK, `{"id_token": "`+fakeJWT(t, subject)+`"}`)
		opts := &SignerOptions{
			ServiceAccountFile: writeServiceAccountKey(t, "service_account", srv.URL),
			ServiceAccountJSON: []byte("not even json"),
		}

		token, err := (&defaultSignerImpl{}).ServiceAccountToken(
			context.Background(), opts.ServiceAccountFile, opts.ServiceAccountJSON, "sigstore",
		)
		require.NoError(t, err)
		require.NotNil(t, token)
	})

	t.Run("no key configured", func(t *testing.T) {
		t.Parallel()

		token, err := (&defaultSignerImpl{}).ServiceAccountToken(context.Background(), "", nil, "sigstore")
		require.Error(t, err)
		require.Nil(t, token)
	})

	t.Run("key JSON is not a service account", func(t *testing.T) {
		t.Parallel()

		opts := &SignerOptions{ServiceAccountJSON: serviceAccountKey(t, "authorized_user", "http://127.0.0.1:1")}

		token, err := (&defaultSignerImpl{}).ServiceAccountToken(
			context.Background(), opts.ServiceAccountFile, opts.ServiceAccountJSON, "sigstore",
		)
		require.Error(t, err)
		require.Nil(t, token)
	})

	t.Run("token endpoint rejects the key", func(t *testing.T) {
		t.Parallel()

		srv := newTokenServer(t, http.StatusUnauthorized, `{"error": "invalid_grant"}`)
		opts := &SignerOptions{ServiceAccountFile: writeServiceAccountKey(t, "service_account", srv.URL)}

		token, err := (&defaultSignerImpl{}).ServiceAccountToken(
			context.Background(), opts.ServiceAccountFile, opts.ServiceAccountJSON, "sigstore",
		)
		require.Error(t, err)
		require.Nil(t, token)
	})

	t.Run("token endpoint returns no token", func(t *testing.T) {
		t.Parallel()

		srv := newTokenServer(t, http.StatusOK, `{}`)
		opts := &SignerOptions{ServiceAccountFile: writeServiceAccountKey(t, "service_account", srv.URL)}

		token, err := (&defaultSignerImpl{}).ServiceAccountToken(
			context.Background(), opts.ServiceAccountFile, opts.ServiceAccountJSON, "sigstore",
		)
		require.Error(t, err)
		require.Nil(t, token)
	})

	t.Run("key file is not a service account", func(t *testing.T) {
		t.Parallel()

		opts := &SignerOptions{ServiceAccountFile: writeServiceAccountKey(t, "authorized_user", "http://127.0.0.1:1")}

		token, err := (&defaultSignerImpl{}).ServiceAccountToken(
			context.Background(), opts.ServiceAccountFile, opts.ServiceAccountJSON, "sigstore",
		)
		require.Error(t, err)
		require.Nil(t, token)
	})

	t.Run("key file does not exist", func(t *testing.T) {
		t.Parallel()

		opts := &SignerOptions{ServiceAccountFile: filepath.Join(t.TempDir(), "missing.json")}

		token, err := (&defaultSignerImpl{}).ServiceAccountToken(
			context.Background(), opts.ServiceAccountFile, opts.ServiceAccountJSON, "sigstore",
		)
		require.Error(t, err)
		require.Nil(t, token)
	})
}

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

// Package attestation signs in-toto statements into sigstore bundles.
package attestation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/carabiner-dev/signer"
	"github.com/carabiner-dev/signer/sts/providers/gcp"
	intoto "github.com/in-toto/attestation/go/v1"
	sbundle "github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore/pkg/oauthflow"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"

	"sigs.k8s.io/release-sdk/object"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt attestationfakes/fake_signer_implementation.go > attestationfakes/_fake_signer_implementation.go && mv attestationfakes/_fake_signer_implementation.go attestationfakes/fake_signer_implementation.go"

// ServiceAccountEnvKey is the name of the environment variable that can
// carry the JSON contents of the Google service account key to sign with.
const ServiceAccountEnvKey = "KREL_SIGNING_SERVICE_ACCOUNT_KEY"

// ErrNoIdentity is returned when the service account key did not yield an
// identity token.
var ErrNoIdentity = errors.New("service account key did not produce an identity token")

// SignerOptions configures the attestation Signer.
type SignerOptions struct {
	// ServiceAccountFile is the path to a Google service account key (JSON).
	// When set, the statement is signed exclusively with the identity of that
	// service account: the signer will not fall back to any other credential
	// if obtaining a certificate with it fails. When neither this nor
	// ServiceAccountJSON is set, the signer tries the ambient credentials of
	// the environment (the GCP metadata server or
	// GOOGLE_APPLICATION_CREDENTIALS, GitHub Actions, GitLab CI).
	ServiceAccountFile string

	// ServiceAccountJSON is the contents of a Google service account key. It
	// locks the signer to the service account exactly like ServiceAccountFile
	// does, which takes precedence when both are set.
	ServiceAccountJSON []byte

	// Timeout bounds the identity token exchange with Google when signing
	// with a service account key.
	Timeout time.Duration
}

// DefaultSignerOptions returns the default signer options.
func DefaultSignerOptions() *SignerOptions {
	return &SignerOptions{
		Timeout: 3 * time.Minute,
	}
}

// hasServiceAccount returns true when a service account key was configured,
// either as a file or as its JSON contents.
func (o *SignerOptions) hasServiceAccount() bool {
	return o.ServiceAccountFile != "" || len(o.ServiceAccountJSON) > 0
}

// Signer signs in-toto statements using a Google Cloud identity and wraps
// the results in sigstore bundles (DSSE envelope + Fulcio certificate +
// Rekor and timestamp proofs).
type Signer struct {
	options *SignerOptions
	impl    signerImplementation
}

// NewSigner returns a new Signer configured with opts. A nil opts uses
// DefaultSignerOptions.
func NewSigner(opts *SignerOptions) *Signer {
	if opts == nil {
		opts = DefaultSignerOptions()
	}

	return &Signer{
		options: opts,
		impl:    &defaultSignerImpl{},
	}
}

//counterfeiter:generate . signerImplementation
type signerImplementation interface {
	ReadStatement(statementPath string) ([]byte, error)
	NewSigner() *signer.Signer
	ServiceAccountToken(ctx context.Context, keyFile string, keyJSON []byte, audience string) (*oauthflow.OIDCIDToken, error)
	SignStatement(sgnr *signer.Signer, token *oauthflow.OIDCIDToken, data []byte) (*sbundle.Bundle, error)
	WriteBundle(sgnr *signer.Signer, bndl *sbundle.Bundle, w io.Writer) error
	WriteFile(filePath string, data []byte) error
}

// SignFile reads the in-toto statement stored in path, which can be a local
// file or an object in Google Cloud Storage (gs://bucket/path), signs it and
// writes the resulting sigstore bundle to w.
func (s *Signer) SignFile(statementPath string, w io.Writer) error {
	data, err := s.impl.ReadStatement(statementPath)
	if err != nil {
		return fmt.Errorf("reading statement: %w", err)
	}

	sgnr := s.impl.NewSigner()
	defer sgnr.Close()

	// When a service account key is set, the identity token is minted here
	// from the key and pinned into the signer. Otherwise token is left nil and
	// the signer runs its own ambient credential discovery.
	var token *oauthflow.OIDCIDToken

	if s.options.hasServiceAccount() {
		ctx, cancel := context.WithTimeout(context.Background(), s.options.Timeout)
		defer cancel()

		// The token audience must match the client ID the sigstore instance
		// expects, otherwise Fulcio will reject it.
		token, err = s.impl.ServiceAccountToken(
			ctx, s.options.ServiceAccountFile, s.options.ServiceAccountJSON, sgnr.Options.OIDCConfig.ClientID,
		)
		if err != nil {
			return fmt.Errorf("obtaining identity from service account key: %w", err)
		}

		logrus.Infof("Signing statement %s as %s", statementPath, token.Subject)
	} else {
		logrus.Infof("Signing statement %s with the ambient credentials", statementPath)
	}

	bndl, err := s.impl.SignStatement(sgnr, token, data)
	if err != nil {
		return fmt.Errorf("signing statement: %w", err)
	}

	if err := s.impl.WriteBundle(sgnr, bndl, w); err != nil {
		return fmt.Errorf("writing bundle: %w", err)
	}

	return nil
}

// WriteFile writes data to filePath, a local file or an object in Google
// Cloud Storage (gs://bucket/path) which is overwritten if it exists.
func (s *Signer) WriteFile(filePath string, data []byte) error {
	if err := s.impl.WriteFile(filePath, data); err != nil {
		return fmt.Errorf("writing %s: %w", filePath, err)
	}

	return nil
}

// objectStore abstracts the Google Cloud Storage operations the signer needs.
type objectStore interface {
	CopyToLocal(gcsPath, dst string) error
	CopyToRemote(src, gcsPath string) error
}

type defaultSignerImpl struct {
	// gcs is the object store used to fetch statements from Google Cloud
	// Storage. Defaults to a GCS client when nil.
	gcs objectStore
}

// ReadStatement reads the statement in path, a local file or an object in
// Google Cloud Storage, and ensures it contains a valid in-toto statement
// before it gets signed.
func (di *defaultSignerImpl) ReadStatement(statementPath string) ([]byte, error) {
	data, err := di.readFile(statementPath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	statement := &intoto.Statement{}
	if err := protojson.Unmarshal(data, statement); err != nil {
		return nil, fmt.Errorf("parsing in-toto statement: %w", err)
	}

	if err := statement.Validate(); err != nil {
		return nil, fmt.Errorf("invalid in-toto statement: %w", err)
	}

	logrus.Infof(
		"Loaded %s statement with %d subjects", statement.GetPredicateType(), len(statement.GetSubject()),
	)

	return data, nil
}

// NewSigner creates a signer targeting the default (public good) sigstore
// instance.
func (*defaultSignerImpl) NewSigner() *signer.Signer {
	return signer.NewSigner()
}

// ServiceAccountToken obtains an OIDC token for the given audience from
// Google Cloud using a service account key, either the file in keyFile or
// the key contents in keyJSON (the file takes precedence). The identity of
// the host is never used, even if exchanging the key fails.
func (*defaultSignerImpl) ServiceAccountToken(
	ctx context.Context, keyFile string, keyJSON []byte, audience string,
) (*oauthflow.OIDCIDToken, error) {
	var key gcp.Option

	switch {
	case keyFile != "":
		logrus.Infof("Obtaining identity token from service account key %s", keyFile)

		key = gcp.WithServiceAccountFile(keyFile)
	case len(keyJSON) > 0:
		logrus.Info("Obtaining identity token from service account key data")

		key = gcp.WithServiceAccountJSON(keyJSON)
	default:
		return nil, errors.New("no service account key configured")
	}

	provider, err := gcp.New(key, gcp.WithAmbientCredentials(false))
	if err != nil {
		return nil, fmt.Errorf("creating GCP identity provider: %w", err)
	}

	token, err := provider.Provide(ctx, audience)
	if err != nil {
		return nil, fmt.Errorf("obtaining identity token: %w", err)
	}

	if token == nil {
		return nil, ErrNoIdentity
	}

	return token, nil
}

// SignStatement signs the statement data and wraps the signed envelope in a
// sigstore bundle. If token is not nil, the signer is locked to it: the
// Fulcio certificate can only be obtained with that identity and the signer
// will not try its own credential discovery. A nil token leaves the signer's
// ambient credential providers in charge of finding an identity. The caller
// owns the signer and is responsible for closing it.
func (*defaultSignerImpl) SignStatement(
	sgnr *signer.Signer, token *oauthflow.OIDCIDToken, data []byte,
) (*sbundle.Bundle, error) {
	if token != nil {
		sgnr.Options.Token = token
		sgnr.Options.DisableSTS = true
	}

	return sgnr.SignStatementBundle(data)
}

// WriteBundle marshals the bundle as JSON into w.
func (*defaultSignerImpl) WriteBundle(sgnr *signer.Signer, bndl *sbundle.Bundle, w io.Writer) error {
	return sgnr.WriteBundle(bndl, w)
}

// WriteFile writes data to a local file or, when filePath is a gs:// URL,
// uploads it to Google Cloud Storage, replacing the object if it exists.
func (di *defaultSignerImpl) WriteFile(filePath string, data []byte) error {
	if !strings.HasPrefix(filePath, object.GcsPrefix) {
		return os.WriteFile(filePath, data, 0o644) //nolint:gosec // bundles are public
	}

	tmpDir, err := os.MkdirTemp("", "krel-sign-attestation-")
	if err != nil {
		return fmt.Errorf("creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, path.Base(filePath))
	if err := os.WriteFile(localPath, data, 0o600); err != nil {
		return fmt.Errorf("writing temporary file: %w", err)
	}

	if di.gcs == nil {
		di.gcs = newGCSClient()
	}

	if err := di.gcs.CopyToRemote(localPath, filePath); err != nil {
		return fmt.Errorf("uploading to %s: %w", filePath, err)
	}

	return nil
}

// readFile returns the contents of a local file or, when filePath is a
// gs:// URL, of the object it points to in Google Cloud Storage.
func (di *defaultSignerImpl) readFile(filePath string) ([]byte, error) {
	if !strings.HasPrefix(filePath, object.GcsPrefix) {
		return os.ReadFile(filePath)
	}

	tmpDir, err := os.MkdirTemp("", "krel-sign-attestation-")
	if err != nil {
		return nil, fmt.Errorf("creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if di.gcs == nil {
		di.gcs = newGCSClient()
	}

	localPath := filepath.Join(tmpDir, path.Base(filePath))
	if err := di.gcs.CopyToLocal(filePath, localPath); err != nil {
		return nil, fmt.Errorf("downloading %s: %w", filePath, err)
	}

	return os.ReadFile(localPath)
}

// newGCSClient returns a GCS client configured to copy single objects.
func newGCSClient() *object.GCS {
	gcs := object.NewGCS()
	gcs.SetOptions(
		gcs.WithConcurrent(false),
		gcs.WithRecursive(false),
		gcs.WithNoClobber(false),
	)

	return gcs
}

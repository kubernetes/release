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
	"encoding/json"
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

var (
	// ErrNoIdentity is returned when the service account key did not yield an
	// identity token.
	ErrNoIdentity = errors.New("service account key did not produce an identity token")

	// ErrAlreadySigned is returned when the file to sign is already a signed
	// artifact (a sigstore bundle or a DSSE envelope) instead of a statement.
	ErrAlreadySigned = errors.New("file is already signed")
)

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

	// ImpersonateServiceAccount is the email of a Google service account to
	// sign as by impersonating it through the IAM Credentials API. The
	// credential calling the API is the service account key, when one is
	// set, or the ambient Google Cloud credential of the host otherwise; it
	// needs roles/iam.serviceAccountTokenCreator on the impersonated account.
	// Like the keys, it locks the signer to that identity: signing fails if
	// impersonation does, it never falls back to another credential.
	ImpersonateServiceAccount string

	// Timeout bounds the identity token exchange with Google when signing
	// with a pinned identity.
	Timeout time.Duration
}

// DefaultSignerOptions returns the default signer options.
func DefaultSignerOptions() *SignerOptions {
	return &SignerOptions{
		Timeout: 3 * time.Minute,
	}
}

// hasPinnedIdentity returns true when the signer must sign with a specific
// identity: a service account key (file or JSON contents) or a service
// account to impersonate.
func (o *SignerOptions) hasPinnedIdentity() bool {
	return o.ServiceAccountFile != "" || len(o.ServiceAccountJSON) > 0 || o.ImpersonateServiceAccount != ""
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
	IdentityToken(ctx context.Context, keyFile string, keyJSON []byte, impersonate, audience string) (*oauthflow.OIDCIDToken, error)
	SignStatement(sgnr *signer.Signer, data []byte) (*sbundle.Bundle, error)
	WriteBundle(bndl *sbundle.Bundle, w io.Writer) error
	WriteFile(filePath string, data []byte) error
}

// SignedStatement is the result of signing one statement.
type SignedStatement struct {
	// Path of the file the statement was read from.
	Path string
	// Bundle is the sigstore bundle wrapping the signed statement.
	Bundle *sbundle.Bundle
}

// SignFile reads the in-toto statement stored in path, which can be a local
// file or an object in Google Cloud Storage (gs://bucket/path), signs it and
// writes the resulting sigstore bundle to w.
func (s *Signer) SignFile(statementPath string, w io.Writer) error {
	signed, err := s.SignFiles([]string{statementPath})
	if err != nil {
		return err
	}

	return s.WriteBundle(signed[0].Bundle, w)
}

// SignFiles reads the in-toto statements stored in paths (local files or
// gs:// objects) and signs them, returning the signed statements in the
// same order. All statements are read and validated before anything is
// signed and they all share the same signing session: the identity token is
// obtained once and the Fulcio certificate is reused for every signature
// while it remains valid.
func (s *Signer) SignFiles(paths []string) ([]*SignedStatement, error) {
	if len(paths) == 0 {
		return nil, errors.New("no statements to sign")
	}

	type statement struct {
		path string
		data []byte
	}

	statements := make([]statement, 0, len(paths))

	for _, statementPath := range paths {
		data, err := s.impl.ReadStatement(statementPath)
		if err != nil {
			return nil, fmt.Errorf("reading statement %s: %w", statementPath, err)
		}

		statements = append(statements, statement{path: statementPath, data: data})
	}

	sgnr := s.impl.NewSigner()
	defer sgnr.Close()

	// When an identity is pinned (a service account key or an account to
	// impersonate), the identity token is minted here and pinned into the
	// signer, which is locked to it: the Fulcio certificate can only be
	// obtained with that identity and the signer will not try its own
	// credential discovery. Otherwise, the signer's ambient credential
	// providers are in charge of finding one.
	if s.options.hasPinnedIdentity() {
		ctx, cancel := context.WithTimeout(context.Background(), s.options.Timeout)
		defer cancel()

		// The token audience must match the client ID the sigstore instance
		// expects, otherwise Fulcio will reject it.
		token, err := s.impl.IdentityToken(
			ctx, s.options.ServiceAccountFile, s.options.ServiceAccountJSON,
			s.options.ImpersonateServiceAccount, sgnr.Options.OIDCConfig.ClientID,
		)
		if err != nil {
			return nil, fmt.Errorf("obtaining the signing identity: %w", err)
		}

		sgnr.Options.Token = token
		sgnr.Options.DisableSTS = true

		logrus.Infof("Signing %d statement(s) as %s", len(statements), token.Subject)
	} else {
		logrus.Infof("Signing %d statement(s) with the ambient credentials", len(statements))
	}

	signed := make([]*SignedStatement, 0, len(statements))

	for _, statement := range statements {
		logrus.Infof("Signing statement %s", statement.path)

		bndl, err := s.impl.SignStatement(sgnr, statement.data)
		if err != nil {
			return nil, fmt.Errorf("signing statement %s: %w", statement.path, err)
		}

		signed = append(signed, &SignedStatement{Path: statement.path, Bundle: bndl})
	}

	return signed, nil
}

// WriteBundle writes the sigstore bundle as JSON to w.
func (s *Signer) WriteBundle(bndl *sbundle.Bundle, w io.Writer) error {
	if err := s.impl.WriteBundle(bndl, w); err != nil {
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

	if isSigned(data) {
		return nil, ErrAlreadySigned
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

// IdentityToken obtains an OIDC token for the given audience from Google
// Cloud for a pinned identity: a service account key, either the file in
// keyFile or the key contents in keyJSON (the file takes precedence), and/or
// a service account to impersonate.
//
// If a key is set, the identity of the host is never used, even if exchanging
// the key fails.
//
// With impersonation, the key or the ambient credential of the host authenticates
// the IAM Credentials API call, minting the token for the impersonated account.
// Equally, any failure is an error and the provider never falls back to signing
// with the caller's own identity.
func (*defaultSignerImpl) IdentityToken(
	ctx context.Context, keyFile string, keyJSON []byte, impersonate, audience string,
) (*oauthflow.OIDCIDToken, error) {
	providerOpts := []gcp.Option{}

	switch {
	case keyFile != "":
		logrus.Infof("Obtaining identity token from service account key %s", keyFile)

		providerOpts = append(providerOpts, gcp.WithServiceAccountFile(keyFile), gcp.WithAmbientCredentials(false))
	case len(keyJSON) > 0:
		logrus.Info("Obtaining identity token from service account key data")

		providerOpts = append(providerOpts, gcp.WithServiceAccountJSON(keyJSON), gcp.WithAmbientCredentials(false))
	}

	if impersonate != "" {
		logrus.Infof("Impersonating service account %s", impersonate)

		providerOpts = append(providerOpts, gcp.WithImpersonation(impersonate))
	}

	if len(providerOpts) == 0 {
		return nil, errors.New("no service account key or account to impersonate configured")
	}

	provider, err := gcp.New(providerOpts...)
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

// SignStatement signs the statement data with sgnr and wraps the signed
// envelope in a sigstore bundle. Repeated calls on the same signer reuse its
// identity and Fulcio certificate. The caller owns the signer and is
// responsible for closing it.
func (*defaultSignerImpl) SignStatement(sgnr *signer.Signer, data []byte) (*sbundle.Bundle, error) {
	return sgnr.SignStatementBundle(data)
}

// WriteBundle marshals the bundle as JSON into w.
func (*defaultSignerImpl) WriteBundle(bndl *sbundle.Bundle, w io.Writer) error {
	bundleJSON, err := protojson.Marshal(bndl)
	if err != nil {
		return fmt.Errorf("marshaling bundle: %w", err)
	}

	if _, err := w.Write(bundleJSON); err != nil {
		return fmt.Errorf("writing bundle: %w", err)
	}

	return nil
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

// isSigned returns true if data looks like an already signed artifact, that
// is a sigstore bundle or a DSSE envelope, rather than a bare statement.
func isSigned(data []byte) bool {
	probe := struct {
		MediaType    string          `json:"mediaType"`
		DSSEEnvelope json.RawMessage `json:"dsseEnvelope"`
		PayloadType  string          `json:"payloadType"`
		Signatures   json.RawMessage `json:"signatures"`
	}{}

	if err := json.Unmarshal(data, &probe); err != nil {
		return false
	}

	return strings.HasPrefix(probe.MediaType, "application/vnd.dev.sigstore.bundle") ||
		len(probe.DSSEEnvelope) > 0 ||
		(probe.PayloadType != "" && len(probe.Signatures) > 0)
}

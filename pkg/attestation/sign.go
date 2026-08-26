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
	"time"

	"github.com/carabiner-dev/signer"
	"github.com/carabiner-dev/signer/sts/providers/gcp"
	intoto "github.com/in-toto/attestation/go/v1"
	sbundle "github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore/pkg/oauthflow"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt attestationfakes/fake_signer_implementation.go > attestationfakes/_fake_signer_implementation.go && mv attestationfakes/_fake_signer_implementation.go attestationfakes/fake_signer_implementation.go"

// ErrNoIdentity is returned when the service account key did not yield an
// identity token.
var ErrNoIdentity = errors.New("service account key did not produce an identity token")

// SignerOptions configures the attestation Signer.
type SignerOptions struct {
	// ServiceAccountFile is the path to a Google service account key (JSON).
	// When set, the statement is signed exclusively with the identity of that
	// service account: the signer will not fall back to any other credential
	// if obtaining a certificate with it fails. When empty, the signer tries
	// the ambient credentials of the environment (the GCP metadata server or
	// GOOGLE_APPLICATION_CREDENTIALS, GitHub Actions, GitLab CI).
	ServiceAccountFile string

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
	ReadStatement(path string) ([]byte, error)
	NewSigner() *signer.Signer
	ServiceAccountToken(ctx context.Context, keyFile, audience string) (*oauthflow.OIDCIDToken, error)
	SignStatement(sgnr *signer.Signer, token *oauthflow.OIDCIDToken, data []byte) (*sbundle.Bundle, error)
	WriteBundle(sgnr *signer.Signer, bndl *sbundle.Bundle, w io.Writer) error
}

// SignFile reads the in-toto statement stored in path, signs it and writes
// the resulting sigstore bundle to w.
func (s *Signer) SignFile(path string, w io.Writer) error {
	data, err := s.impl.ReadStatement(path)
	if err != nil {
		return fmt.Errorf("reading statement: %w", err)
	}

	sgnr := s.impl.NewSigner()

	// When a service account key is set, the identity token is minted here
	// from the key and pinned into the signer. Otherwise token is left nil and
	// the signer runs its own ambient credential discovery.
	var token *oauthflow.OIDCIDToken

	if s.options.ServiceAccountFile != "" {
		ctx, cancel := context.WithTimeout(context.Background(), s.options.Timeout)
		defer cancel()

		// The token audience must match the client ID the sigstore instance
		// expects, otherwise Fulcio will reject it.
		token, err = s.impl.ServiceAccountToken(
			ctx, s.options.ServiceAccountFile, sgnr.Options.OIDCConfig.ClientID,
		)
		if err != nil {
			return fmt.Errorf("obtaining identity from service account key: %w", err)
		}

		logrus.Infof("Signing statement %s as %s", path, token.Subject)
	} else {
		logrus.Infof("Signing statement %s with the ambient credentials", path)
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

type defaultSignerImpl struct{}

// ReadStatement reads the file in path and ensures it contains a valid
// in-toto statement before it gets signed.
func (*defaultSignerImpl) ReadStatement(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
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
// Google Cloud using the service account key in keyFile. The identity of the
// host is never used, even if exchanging the key fails.
func (*defaultSignerImpl) ServiceAccountToken(
	ctx context.Context, keyFile, audience string,
) (*oauthflow.OIDCIDToken, error) {
	logrus.Infof("Obtaining identity token from service account key %s", keyFile)

	provider, err := gcp.New(
		gcp.WithServiceAccountFile(keyFile),
		gcp.WithAmbientCredentials(false),
	)
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
// ambient credential providers in charge of finding an identity.
func (*defaultSignerImpl) SignStatement(
	sgnr *signer.Signer, token *oauthflow.OIDCIDToken, data []byte,
) (*sbundle.Bundle, error) {
	if token != nil {
		sgnr.Options.Token = token
		sgnr.Options.DisableSTS = true
	}

	defer sgnr.Close()

	return sgnr.SignStatementBundle(data)
}

// WriteBundle marshals the bundle as JSON into w.
func (*defaultSignerImpl) WriteBundle(sgnr *signer.Signer, bndl *sbundle.Bundle, w io.Writer) error {
	return sgnr.WriteBundle(bndl, w)
}

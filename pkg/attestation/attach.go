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
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	intoto "github.com/in-toto/attestation/go/v1"
	ociremote "github.com/sigstore/cosign/v3/pkg/oci/remote"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"

	"sigs.k8s.io/release-sdk/object"
)

// predicateTypeAnnotation is the annotation of sigstore bundle referrers that
// names the predicate type of the attestation they carry.
const predicateTypeAnnotation = "dev.sigstore.bundle.predicateType"

// AttachToImages attaches the bundle of a signed statement as an OCI referrer
// to each container image the statement is about, the way cosign attaches
// attestations in the sigstore bundle format. Subjects named by their
// registry repository, including the registry, with a SHA-256 digest are
// images; others, like the gs:// objects of the staged release artifacts,
// are skipped. An image that
// already has an attestation of the same predicate type is left alone, so
// that attaching again adds nothing. It returns the number of bundles
// attached.
func (s *Signer) AttachToImages(signed *SignedStatement) (int, error) {
	if signed == nil || signed.Bundle == nil || signed.Bundle.Bundle == nil {
		return 0, errors.New("no signed bundle to attach")
	}

	envelope := signed.Bundle.GetDsseEnvelope()
	if envelope == nil {
		return 0, fmt.Errorf("the bundle of %s holds no DSSE envelope", signed.Path)
	}

	statement := &intoto.Statement{}
	if err := protojson.Unmarshal(envelope.GetPayload(), statement); err != nil {
		return 0, fmt.Errorf("parsing the statement of %s: %w", signed.Path, err)
	}

	var bundle bytes.Buffer
	if err := s.WriteBundle(signed.Bundle, &bundle); err != nil {
		return 0, err
	}

	attached := 0

	for _, subject := range statement.GetSubject() {
		ref, ok := imageReference(subject)
		if !ok {
			continue
		}

		exists, err := s.impl.HasReferrer(ref, statement.GetPredicateType())
		if err != nil {
			return attached, fmt.Errorf("looking up the attestations of %s: %w", ref, err)
		}

		if exists {
			logrus.Infof("%s already has a %s attestation", ref, statement.GetPredicateType())

			continue
		}

		logrus.Infof("Attaching the %s attestation to %s", statement.GetPredicateType(), ref)

		if err := s.impl.AttachBundle(ref, bundle.Bytes(), statement.GetPredicateType()); err != nil {
			return attached, fmt.Errorf("attaching the attestation to %s: %w", ref, err)
		}

		attached++
	}

	return attached, nil
}

// imageReference returns the digest reference of the container image a
// subject names, if it is one.
func imageReference(subject *intoto.ResourceDescriptor) (string, bool) {
	digest := subject.GetDigest()[intoto.AlgorithmSHA256.String()]
	if digest == "" || strings.HasPrefix(subject.GetName(), object.GcsPrefix) {
		return "", false
	}

	// Strict validation requires an explicit registry, so that names of
	// other subjects, like relative file paths, don't default to Docker Hub.
	ref, err := name.NewDigest(subject.GetName()+"@sha256:"+digest, name.StrictValidation)
	if err != nil {
		return "", false
	}

	return ref.String(), true
}

// keychain resolves the registry credentials: the docker configuration with
// its credential helpers, and the Google Cloud credentials of the host.
var keychain = authn.NewMultiKeychain(authn.DefaultKeychain, google.Keychain)

// HasReferrer returns whether the image has a sigstore bundle referrer with
// an attestation of the predicate type.
func (*defaultSignerImpl) HasReferrer(ref, predicateType string) (bool, error) {
	digest, err := name.NewDigest(ref)
	if err != nil {
		return false, fmt.Errorf("parsing %s: %w", ref, err)
	}

	index, err := remote.Referrers(digest, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		return false, fmt.Errorf("listing the referrers of %s: %w", ref, err)
	}

	manifest, err := index.IndexManifest()
	if err != nil {
		return false, fmt.Errorf("reading the referrers of %s: %w", ref, err)
	}

	for i := range manifest.Manifests {
		if manifest.Manifests[i].Annotations[predicateTypeAnnotation] == predicateType {
			return true, nil
		}
	}

	return false, nil
}

// AttachBundle attaches the sigstore bundle as an OCI referrer to the image.
func (*defaultSignerImpl) AttachBundle(ref string, bundle []byte, predicateType string) error {
	digest, err := name.NewDigest(ref)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", ref, err)
	}

	return ociremote.WriteAttestationNewBundleFormat(
		digest, bundle, predicateType,
		ociremote.WithRemoteOptions(remote.WithAuthFromKeychain(keychain)),
	)
}

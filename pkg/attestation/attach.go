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
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	intoto "github.com/in-toto/attestation/go/v1"
	ociremote "github.com/sigstore/cosign/v3/pkg/oci/remote"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"

	"sigs.k8s.io/release-sdk/object"
)

// defaultAttachTimeout bounds the registry requests for one image.
const defaultAttachTimeout = 2 * time.Minute

// AttacherOptions configures where and how bundles get attached.
type AttacherOptions struct {
	// Repository is the repository the images have to be in, or below, for
	// example gcr.io/k8s-staging-kubernetes. A subject that is neither a
	// gs:// object nor such an image fails attaching, so that a statement
	// can't make the attacher write to other repositories.
	Repository string

	// Timeout bounds the registry requests for one image.
	Timeout time.Duration
}

// Attacher attaches signed statements to the container images they are
// about.
type Attacher struct {
	options *AttacherOptions
	impl    attacherImplementation
}

// NewAttacher returns an attacher for the options.
func NewAttacher(opts *AttacherOptions) *Attacher {
	options := *opts
	if options.Timeout == 0 {
		options.Timeout = defaultAttachTimeout
	}

	return &Attacher{options: &options, impl: &defaultAttacherImpl{}}
}

// SetImpl replaces the registry implementation, for testing.
func (a *Attacher) SetImpl(impl attacherImplementation) {
	a.impl = impl
}

//counterfeiter:generate . attacherImplementation
type attacherImplementation interface {
	HasBundle(ctx context.Context, ref name.Digest, predicateType string, payload []byte) (bool, error)
	AttachBundle(ctx context.Context, ref name.Digest, bundle []byte, predicateType string) error
}

// AttachToImages attaches the bundle, the serialized sigstore bundle of the
// signed statement, as an OCI referrer to each container image the statement
// is about, the way cosign attaches attestations in the sigstore bundle
// format. The subjects have to be gs:// objects, which are skipped, or
// images below the repository of the options with a SHA-256 digest. An image
// that already has a bundle with the same statement is left alone, so that
// attaching again adds nothing. It returns the number of bundles attached.
func (a *Attacher) AttachToImages(signed *SignedStatement, bundle []byte) (int, error) {
	if signed == nil || signed.Bundle == nil || signed.Bundle.Bundle == nil {
		return 0, errors.New("no signed bundle to attach")
	}

	if a.options.Repository == "" {
		return 0, errors.New("no repository to attach to")
	}

	envelope := signed.Bundle.GetDsseEnvelope()
	if envelope == nil {
		return 0, fmt.Errorf("the bundle of %s holds no DSSE envelope", signed.Path)
	}

	payload := envelope.GetPayload()

	statement := &intoto.Statement{}
	if err := protojson.Unmarshal(payload, statement); err != nil {
		return 0, fmt.Errorf("parsing the statement of %s: %w", signed.Path, err)
	}

	// All subjects are checked before anything is written.
	refs := []name.Digest{}

	for _, subject := range statement.GetSubject() {
		if strings.HasPrefix(subject.GetName(), object.GcsPrefix) {
			continue
		}

		ref, err := imageReference(subject, a.options.Repository)
		if err != nil {
			return 0, fmt.Errorf("subject of %s: %w", signed.Path, err)
		}

		refs = append(refs, ref)
	}

	attached := 0

	for _, ref := range refs {
		done, err := a.attach(ref, bundle, statement.GetPredicateType(), payload)
		if err != nil {
			return attached, err
		}

		if done {
			attached++
		}
	}

	return attached, nil
}

// attach attaches the bundle to one image, unless it has a bundle with the
// same statement already, and returns whether it did.
func (a *Attacher) attach(ref name.Digest, bundle []byte, predicateType string, payload []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.options.Timeout)
	defer cancel()

	exists, err := a.impl.HasBundle(ctx, ref, predicateType, payload)
	if err != nil {
		return false, fmt.Errorf("looking up the attestations of %s: %w", ref, err)
	}

	if exists {
		logrus.Infof("%s already has this %s attestation", ref, predicateType)

		return false, nil
	}

	logrus.Infof("Attaching the %s attestation to %s", predicateType, ref)

	if err := a.impl.AttachBundle(ctx, ref, bundle, predicateType); err != nil {
		return false, fmt.Errorf("attaching the attestation to %s: %w", ref, err)
	}

	return true, nil
}

// imageReference returns the digest reference of the container image a
// subject names, which has to be in the repository or below it.
func imageReference(subject *intoto.ResourceDescriptor, repository string) (name.Digest, error) {
	digest := subject.GetDigest()[intoto.AlgorithmSHA256.String()]
	if digest == "" {
		return name.Digest{}, fmt.Errorf("%s has no SHA-256 digest", subject.GetName())
	}

	// Strict validation requires an explicit registry, so that names of
	// other subjects, like relative file paths, don't default to Docker Hub.
	ref, err := name.NewDigest(subject.GetName()+"@sha256:"+digest, name.StrictValidation)
	if err != nil {
		return name.Digest{}, fmt.Errorf("%s is not an image: %w", subject.GetName(), err)
	}

	// A name like v1.36.0/kubernetes.tar.gz parses as an image of the
	// registry v1.36.0 too, so the repository decides.
	if repo := ref.Context().Name(); repo != repository && !strings.HasPrefix(repo, repository+"/") {
		return name.Digest{}, fmt.Errorf("%s is not an image below %s", subject.GetName(), repository)
	}

	return ref, nil
}

// defaultAttacherImpl attaches through the registry. The credentials come
// from the Google Cloud credentials of the host only, which cover gcr.io and
// Artifact Registry, and not from the docker configuration or its credential
// helpers.
type defaultAttacherImpl struct{}

// HasBundle returns whether the image has a sigstore bundle referrer of the
// predicate type whose DSSE envelope carries the payload.
func (*defaultAttacherImpl) HasBundle(ctx context.Context, ref name.Digest, predicateType string, payload []byte) (bool, error) {
	opts := []remote.Option{remote.WithContext(ctx), remote.WithAuthFromKeychain(google.Keychain)}

	index, err := remote.Referrers(ref, opts...)
	if err != nil {
		return false, fmt.Errorf("listing the referrers: %w", err)
	}

	manifest, err := index.IndexManifest()
	if err != nil {
		return false, fmt.Errorf("reading the referrers: %w", err)
	}

	for i := range manifest.Manifests {
		if manifest.Manifests[i].Annotations[ociremote.BundlePredicateType] != predicateType {
			continue
		}

		same, err := bundleCarries(ref.Context().Digest(manifest.Manifests[i].Digest.String()), payload, opts)
		if err != nil {
			return false, err
		}

		if same {
			return true, nil
		}
	}

	return false, nil
}

// bundleCarries returns whether the bundle in the layer of the referrer has
// a DSSE envelope with the payload.
func bundleCarries(referrer name.Digest, payload []byte, opts []remote.Option) (bool, error) {
	img, err := remote.Image(referrer, opts...)
	if err != nil {
		return false, fmt.Errorf("reading the referrer %s: %w", referrer, err)
	}

	layers, err := img.Layers()
	if err != nil {
		return false, fmt.Errorf("reading the layers of %s: %w", referrer, err)
	}

	for _, layer := range layers {
		reader, err := layer.Uncompressed()
		if err != nil {
			return false, fmt.Errorf("reading a layer of %s: %w", referrer, err)
		}

		data, err := io.ReadAll(reader)
		_ = reader.Close()

		if err != nil {
			return false, fmt.Errorf("reading a layer of %s: %w", referrer, err)
		}

		// Only the statement matters here, so the bundle is parsed without
		// validating the rest of it. A layer that isn't a bundle isn't ours.
		existing := &protobundle.Bundle{}
		if err := protojson.Unmarshal(data, existing); err != nil {
			continue
		}

		if envelope := existing.GetDsseEnvelope(); envelope != nil && bytes.Equal(envelope.GetPayload(), payload) {
			return true, nil
		}
	}

	return false, nil
}

// AttachBundle attaches the sigstore bundle as an OCI referrer to the image.
func (*defaultAttacherImpl) AttachBundle(ctx context.Context, ref name.Digest, bundle []byte, predicateType string) error {
	return ociremote.WriteAttestationNewBundleFormat(
		ref, bundle, predicateType,
		ociremote.WithRemoteOptions(remote.WithContext(ctx), remote.WithAuthFromKeychain(google.Keychain)),
	)
}

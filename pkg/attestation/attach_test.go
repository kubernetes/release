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
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protodsse "github.com/sigstore/protobuf-specs/gen/pb-go/dsse"
	sbundle "github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/attestation/attestationfakes"
)

const imageStatement = `{
  "_type": "https://in-toto.io/Statement/v1",
  "subject": [
    {"name": "gs://bucket/stage/v1.36.0/kubernetes.tar.gz", "digest": {"sha256": "` + digestA + `"}},
    {"name": "gcr.io/k8s-staging-kubernetes/kube-apiserver", "digest": {"sha256": "` + digestA + `"}},
    {"name": "gcr.io/k8s-staging-kubernetes/kube-apiserver-amd64", "digest": {"sha256": "` + digestB + `"}},
    {"name": "gcr.io/k8s-staging-kubernetes/kube-apiserver", "digest": {"sha256": "` + digestB + `"}},
    {"name": "gcr.io/k8s-staging-kubernetes/kube-proxy", "digest": {"sha512": "` + digestA + digestA + `"}},
    {"name": "Not A Repository", "digest": {"sha256": "` + digestA + `"}},
    {"name": "release-images/amd64/kube-apiserver.tar", "digest": {"sha256": "` + digestA + `"}}
  ],
  "predicateType": "https://slsa.dev/provenance/v1",
  "predicate": {"buildDefinition": {"buildType": "https://git.k8s.io/release/docs/krel/buildtypes/v1"}}
}`

const (
	digestA = "0e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7"
	digestB = "1e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7"
)

func signedImageStatement(payload string) *SignedStatement {
	return &SignedStatement{
		Path: "gs://bucket/stage/v1.36.0/image-provenance.json",
		Bundle: &sbundle.Bundle{Bundle: &protobundle.Bundle{
			MediaType: "application/vnd.dev.sigstore.bundle.v0.3+json",
			Content: &protobundle.Bundle_DsseEnvelope{DsseEnvelope: &protodsse.Envelope{
				Payload:     []byte(payload),
				PayloadType: "application/vnd.in-toto+json",
			}},
		}},
	}
}

func TestAttachToImages(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		signed   *SignedStatement
		prepare  func(*attestationfakes.FakeSignerImplementation)
		attached []string
		err      string
	}{
		{
			name:   "attaches to the image subjects only",
			signed: signedImageStatement(imageStatement),
			attached: []string{
				"gcr.io/k8s-staging-kubernetes/kube-apiserver@sha256:" + digestA,
				"gcr.io/k8s-staging-kubernetes/kube-apiserver-amd64@sha256:" + digestB,
				"gcr.io/k8s-staging-kubernetes/kube-apiserver@sha256:" + digestB,
			},
		},
		{
			name:   "skips images that have the attestation already",
			signed: signedImageStatement(imageStatement),
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.HasReferrerCalls(func(ref, _ string) (bool, error) {
					return strings.Contains(ref, "-amd64@"), nil
				})
			},
			attached: []string{
				"gcr.io/k8s-staging-kubernetes/kube-apiserver@sha256:" + digestA,
				"gcr.io/k8s-staging-kubernetes/kube-apiserver@sha256:" + digestB,
			},
		},
		{
			name:   "lookup fails",
			signed: signedImageStatement(imageStatement),
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.HasReferrerReturns(false, errTest)
			},
			err: "looking up the attestations",
		},
		{
			name:   "attaching fails",
			signed: signedImageStatement(imageStatement),
			prepare: func(mock *attestationfakes.FakeSignerImplementation) {
				mock.AttachBundleReturns(errTest)
			},
			err: "attaching the attestation",
		},
		{
			name:   "no envelope",
			signed: &SignedStatement{Bundle: &sbundle.Bundle{Bundle: &protobundle.Bundle{}}},
			err:    "no DSSE envelope",
		},
		{
			name:   "no bundle",
			signed: &SignedStatement{},
			err:    "no signed bundle",
		},
		{
			name:   "malformed statement",
			signed: signedImageStatement("{"),
			err:    "parsing the statement",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &attestationfakes.FakeSignerImplementation{}
			mock.WriteBundleCalls(func(_ *sbundle.Bundle, w io.Writer) error {
				_, err := w.Write([]byte("bundle"))

				return err
			})

			if tc.prepare != nil {
				tc.prepare(mock)
			}

			sut := NewSigner(nil)
			sut.impl = mock

			attached, err := sut.AttachToImages(tc.signed)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tc.attached), attached)
			require.Equal(t, len(tc.attached), mock.AttachBundleCallCount())

			for i, ref := range tc.attached {
				gotRef, bundle, predicateType := mock.AttachBundleArgsForCall(i)
				require.Equal(t, ref, gotRef)
				require.Equal(t, []byte("bundle"), bundle)
				require.Equal(t, "https://slsa.dev/provenance/v1", predicateType)
			}
		})
	}
}

func TestAttachBundleToRegistry(t *testing.T) {
	t.Parallel()

	for _, referrersAPI := range []bool{true, false} {
		t.Run(fmt.Sprintf("referrers API %t", referrersAPI), func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(registry.New(registry.WithReferrersSupport(referrersAPI)))
			t.Cleanup(server.Close)

			img, err := random.Image(1024, 1)
			require.NoError(t, err)

			repo := strings.TrimPrefix(server.URL, "http://") + "/kube-apiserver"
			tag, err := name.NewTag(repo + ":v1.36.0")
			require.NoError(t, err)
			require.NoError(t, remote.Write(tag, img))

			digest, err := img.Digest()
			require.NoError(t, err)

			ref := repo + "@" + digest.String()
			impl := &defaultSignerImpl{}
			predicateType := "https://slsa.dev/provenance/v1"

			exists, err := impl.HasReferrer(ref, predicateType)
			require.NoError(t, err)
			require.False(t, exists)

			require.NoError(t, impl.AttachBundle(ref, []byte(`{"mediaType":"application/vnd.dev.sigstore.bundle.v0.3+json"}`), predicateType))

			// Found by the predicate type annotation, which HasReferrer keys
			// on. The image promoter finds the bundle by its artifact type
			// or layer media type instead, the annotation names its
			// predicate type.
			exists, err = impl.HasReferrer(ref, predicateType)
			require.NoError(t, err)
			require.True(t, exists)

			exists, err = impl.HasReferrer(ref, "https://spdx.dev/Document")
			require.NoError(t, err)
			require.False(t, exists)
		})
	}
}

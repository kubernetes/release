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
	"context"
	"fmt"
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
	"google.golang.org/protobuf/encoding/protojson"

	"k8s.io/release/pkg/attestation/attestationfakes"
)

const (
	digestA = "0e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7"
	digestB = "1e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7"

	stagingRepository = "gcr.io/k8s-staging-kubernetes"
	provenanceType    = "https://slsa.dev/provenance/v1"
)

// imageStatement returns a statement with the subjects of the staged images
// and a release artifact in the bucket, plus the extra subjects.
func imageStatement(extra ...string) string {
	subjects := append([]string{
		`{"name": "gs://bucket/stage/v1.36.0/kubernetes.tar.gz", "digest": {"sha256": "` + digestA + `"}}`,
		`{"name": "gcr.io/k8s-staging-kubernetes/kube-apiserver", "digest": {"sha256": "` + digestA + `"}}`,
		`{"name": "gcr.io/k8s-staging-kubernetes/kube-apiserver-amd64", "digest": {"sha256": "` + digestB + `"}}`,
		`{"name": "gcr.io/k8s-staging-kubernetes/mock/kube-apiserver", "digest": {"sha256": "` + digestB + `"}}`,
	}, extra...)

	return `{
  "_type": "https://in-toto.io/Statement/v1",
  "subject": [` + strings.Join(subjects, ",") + `],
  "predicateType": "` + provenanceType + `",
  "predicate": {"buildDefinition": {"buildType": "https://git.k8s.io/release/docs/krel/buildtypes/v1"}}
}`
}

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
		name       string
		signed     *SignedStatement
		repository string
		prepare    func(*attestationfakes.FakeAttacherImplementation)
		attached   []string
		err        string
	}{
		{
			name:   "attaches to the image subjects below the repository",
			signed: signedImageStatement(imageStatement()),
			attached: []string{
				"gcr.io/k8s-staging-kubernetes/kube-apiserver@sha256:" + digestA,
				"gcr.io/k8s-staging-kubernetes/kube-apiserver-amd64@sha256:" + digestB,
				"gcr.io/k8s-staging-kubernetes/mock/kube-apiserver@sha256:" + digestB,
			},
		},
		{
			name:   "skips images that have the same statement already",
			signed: signedImageStatement(imageStatement()),
			prepare: func(mock *attestationfakes.FakeAttacherImplementation) {
				mock.HasBundleCalls(func(_ context.Context, ref name.Digest, _ string, _ []byte) (bool, error) {
					return strings.Contains(ref.String(), "-amd64@"), nil
				})
			},
			attached: []string{
				"gcr.io/k8s-staging-kubernetes/kube-apiserver@sha256:" + digestA,
				"gcr.io/k8s-staging-kubernetes/mock/kube-apiserver@sha256:" + digestB,
			},
		},
		{
			name: "a file path that parses as an image of another registry",
			signed: signedImageStatement(imageStatement(
				`{"name": "v1.36.0-alpha.1/kubernetes.tar.gz", "digest": {"sha256": "` + digestA + `"}}`,
			)),
			err: "is not an image below gcr.io/k8s-staging-kubernetes",
		},
		{
			name: "an image of another registry",
			signed: signedImageStatement(imageStatement(
				`{"name": "docker.io/k8sreleng/kube-apiserver", "digest": {"sha256": "` + digestA + `"}}`,
			)),
			err: "is not an image below",
		},
		{
			name: "an image of a repository with the same prefix",
			signed: signedImageStatement(imageStatement(
				`{"name": "gcr.io/k8s-staging-kubernetes-other/kube-apiserver", "digest": {"sha256": "` + digestA + `"}}`,
			)),
			err: "is not an image below",
		},
		{
			name: "a name without a registry",
			signed: signedImageStatement(imageStatement(
				`{"name": "Not A Repository", "digest": {"sha256": "` + digestA + `"}}`,
			)),
			err: "is not an image",
		},
		{
			name: "an image without a SHA-256 digest",
			signed: signedImageStatement(imageStatement(
				`{"name": "gcr.io/k8s-staging-kubernetes/kube-proxy", "digest": {"sha512": "` + digestA + digestA + `"}}`,
			)),
			err: "has no SHA-256 digest",
		},
		{
			name:   "lookup fails",
			signed: signedImageStatement(imageStatement()),
			prepare: func(mock *attestationfakes.FakeAttacherImplementation) {
				mock.HasBundleReturns(false, errTest)
			},
			err: "looking up the attestations",
		},
		{
			name:   "attaching fails",
			signed: signedImageStatement(imageStatement()),
			prepare: func(mock *attestationfakes.FakeAttacherImplementation) {
				mock.AttachBundleReturns(errTest)
			},
			err: "attaching the attestation",
		},
		{
			name:       "no repository",
			signed:     signedImageStatement(imageStatement()),
			repository: "-",
			err:        "no repository",
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

			mock := &attestationfakes.FakeAttacherImplementation{}
			if tc.prepare != nil {
				tc.prepare(mock)
			}

			repository := stagingRepository
			if tc.repository == "-" {
				repository = ""
			}

			sut := NewAttacher(&AttacherOptions{Repository: repository})
			sut.SetImpl(mock)

			attached, err := sut.AttachToImages(tc.signed, []byte("bundle"))
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)

				if !strings.Contains(tc.name, "fails") {
					require.Zero(t, mock.HasBundleCallCount(), "nothing is looked up before all subjects are valid")
					require.Zero(t, mock.AttachBundleCallCount(), "nothing is written before all subjects are valid")
				}

				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tc.attached), attached)
			require.Equal(t, len(tc.attached), mock.AttachBundleCallCount())

			for i := range mock.HasBundleCallCount() {
				_, _, predicateType, payload := mock.HasBundleArgsForCall(i)
				require.Equal(t, provenanceType, predicateType)
				require.Equal(t, tc.signed.Bundle.GetDsseEnvelope().GetPayload(), payload, "the statement decides")
			}

			for i, ref := range tc.attached {
				_, gotRef, bundle, predicateType := mock.AttachBundleArgsForCall(i)
				require.Equal(t, ref, gotRef.String())
				require.Equal(t, []byte("bundle"), bundle)
				require.Equal(t, provenanceType, predicateType)
			}
		})
	}
}

// testBundle returns a serialized sigstore bundle of the statement.
func testBundle(t *testing.T, statement string) []byte {
	t.Helper()

	data, err := protojson.Marshal(signedImageStatement(statement).Bundle.Bundle)
	require.NoError(t, err)

	return data
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

			ref, err := name.NewDigest(repo + "@" + digest.String())
			require.NoError(t, err)

			ctx := t.Context()
			impl := &defaultAttacherImpl{}
			ours := imageStatement()
			other := imageStatement(`{"name": "gcr.io/k8s-staging-kubernetes/pause", "digest": {"sha256": "` + digestA + `"}}`)

			exists, err := impl.HasBundle(ctx, ref, provenanceType, []byte(ours))
			require.NoError(t, err)
			require.False(t, exists)

			// A referrer with the predicate type annotation, but another
			// statement, like one planted before the sign step, doesn't
			// count.
			require.NoError(t, impl.AttachBundle(ctx, ref, testBundle(t, other), provenanceType))

			exists, err = impl.HasBundle(ctx, ref, provenanceType, []byte(ours))
			require.NoError(t, err)
			require.False(t, exists)

			require.NoError(t, impl.AttachBundle(ctx, ref, testBundle(t, ours), provenanceType))

			exists, err = impl.HasBundle(ctx, ref, provenanceType, []byte(ours))
			require.NoError(t, err)
			require.True(t, exists)

			exists, err = impl.HasBundle(ctx, ref, "https://spdx.dev/Document", []byte(ours))
			require.NoError(t, err)
			require.False(t, exists)
		})
	}
}

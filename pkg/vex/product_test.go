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

package vex

import (
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	openvex "github.com/openvex/go-vex/pkg/vex"
	"github.com/stretchr/testify/require"
)

const testDigest = "d1045e5c6d2f016797d22143eba7502e1bb712a4681836a7c35763a9c192dd70"

func TestImageProduct(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		reference string
		want      string
	}{
		{
			name:      "image of the production registry",
			reference: "registry.k8s.io/kube-apiserver@sha256:" + testDigest,
			want:      "pkg:oci/kube-apiserver@sha256:" + testDigest,
		},
		{
			name:      "image of a staging repository",
			reference: "gcr.io/k8s-staging-kubernetes/kube-apiserver@sha256:" + testDigest,
			want:      "pkg:oci/kube-apiserver@sha256:" + testDigest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ref, err := name.NewDigest(tc.reference)
			require.NoError(t, err)

			product, err := ImageProduct(ref)
			require.NoError(t, err)
			require.Equal(t, tc.want, product.ID)
			require.Empty(t, product.Hashes)
		})
	}
}

func TestImageProductRejectsNamelessImage(t *testing.T) {
	t.Parallel()

	// The repository of a parsed reference can end in a slash, which would
	// leave the purl without an image name.
	ref, err := name.NewDigest("gcr.io/k8s-staging-kubernetes/kube-apiserver/@sha256:" + testDigest)
	require.NoError(t, err)

	_, err = ImageProduct(ref)
	require.Error(t, err)
}

func TestBinaryProduct(t *testing.T) {
	t.Parallel()

	product, err := BinaryProduct("sha256:" + strings.ToUpper(testDigest))
	require.NoError(t, err)
	require.Empty(t, product.ID)
	require.Equal(t, map[openvex.Algorithm]openvex.Hash{
		openvex.SHA256: openvex.Hash(testDigest),
	}, product.Hashes)
}

func TestBinaryProductRejectsInvalidDigest(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		digest string
	}{
		{name: "too short", digest: "abc"},
		{name: "not hexadecimal", digest: strings.Repeat("z", sha256HexLength)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := BinaryProduct(tc.digest)
			require.Error(t, err)
		})
	}
}

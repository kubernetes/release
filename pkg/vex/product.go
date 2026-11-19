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
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	openvex "github.com/openvex/go-vex/pkg/vex"
	packageurl "github.com/package-url/packageurl-go"
)

const sha256HexLength = 64

// ImageProduct identifies an OCI image by its repository and manifest digest.
// The purl takes the last fragment of the repository as its name and the
// repository_url qualifier carries the full repository:
// https://github.com/package-url/purl-spec/blob/main/docs/types/definitions/oci-definition.md
func ImageProduct(ref name.Digest) (openvex.Component, error) {
	repository := ref.Context().Name()

	imageName := repository[strings.LastIndex(repository, "/")+1:]
	if imageName == "" {
		return openvex.Component{}, fmt.Errorf("image reference %s has no name", ref.Name())
	}

	purl := packageurl.NewPackageURL(
		packageurl.TypeOCI,
		"", imageName, ref.DigestStr(),
		packageurl.Qualifiers{},
		"",
	)

	return openvex.Component{ID: purl.String()}, nil
}

// BinaryProduct identifies a release binary by its SHA-256 digest.
func BinaryProduct(digest string) (openvex.Component, error) {
	digest, err := normalizeSHA256(digest)
	if err != nil {
		return openvex.Component{}, err
	}

	return openvex.Component{
		Hashes: map[openvex.Algorithm]openvex.Hash{
			openvex.SHA256: openvex.Hash(digest),
		},
	}, nil
}

func normalizeSHA256(digest string) (string, error) {
	digest = strings.TrimSpace(digest)

	digest = strings.TrimPrefix(digest, "sha256:")
	if len(digest) != sha256HexLength {
		return "", fmt.Errorf("sha-256 digest must contain %d hexadecimal characters", sha256HexLength)
	}

	if _, err := hex.DecodeString(digest); err != nil {
		return "", fmt.Errorf("invalid sha-256 digest: %w", err)
	}

	return strings.ToLower(digest), nil
}

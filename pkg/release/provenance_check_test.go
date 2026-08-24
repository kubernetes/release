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

package release

import (
	"os"
	"path/filepath"
	"testing"

	intoto "github.com/in-toto/attestation/go/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"sigs.k8s.io/release-sdk/object"
	"sigs.k8s.io/release-utils/hash"
)

func TestProcessAttestationAndCheckProvenance(t *testing.T) {
	t.Parallel()

	const (
		bucket       = "test-bucket"
		buildVersion = "v1.36.0-alpha.1.10+abcdef0123456"
	)

	// A local copy of the staged artifacts as downloaded from the bucket
	stageDir := t.TempDir()
	artifact := filepath.Join(buildVersion, "gcs-stage", "v1.36.0-alpha.1", "kubernetes.tar.gz")
	require.NoError(t, os.MkdirAll(filepath.Join(stageDir, filepath.Dir(artifact)), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(stageDir, artifact), []byte("artifact data"), 0o600))

	sha256Sum, err := hash.SHA256ForFile(filepath.Join(stageDir, artifact))
	require.NoError(t, err)
	sha512Sum, err := hash.SHA512ForFile(filepath.Join(stageDir, artifact))
	require.NoError(t, err)

	// The provenance attestation generated during the stage run
	statement := &intoto.Statement{
		Type:          intoto.StatementTypeUri,
		PredicateType: "https://slsa.dev/provenance/v1",
		Subject: []*intoto.ResourceDescriptor{{
			Name:   object.GcsPrefix + filepath.Join(bucket, StagePath, artifact),
			Digest: map[string]string{"sha256": sha256Sum, "sha512": sha512Sum},
		}},
	}
	data, err := protojson.Marshal(statement)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(stageDir, buildVersion), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(stageDir, buildVersion, ProvenanceFilename), data, 0o600,
	))

	opts := &ProvenanceCheckerOptions{
		StageBucket:    bucket,
		StageDirectory: stageDir,
	}
	impl := &defaultProvenanceCheckerImpl{}

	// The attestation parses and the subject paths lose the bucket prefix
	parsed, err := impl.processAttestation(opts, buildVersion)
	require.NoError(t, err)
	require.Len(t, parsed.GetSubject(), 1)
	// The stripped name keeps a leading separator, absorbed by
	// filepath.Join when the local copies are checked
	require.Equal(t, "/"+artifact, parsed.GetSubject()[0].GetName())

	// The artifact hashes verify
	require.NoError(t, impl.checkProvenance(opts, parsed))

	// Tampering with the artifact makes the check fail
	require.NoError(t, os.WriteFile(filepath.Join(stageDir, artifact), []byte("tampered"), 0o600))
	require.Error(t, impl.checkProvenance(opts, parsed))

	// A subject without a supported digest fails
	parsed.Subject[0].Digest = map[string]string{"md5": "abc"}
	require.Error(t, impl.checkProvenance(opts, parsed))

	// A missing artifact fails
	parsed.Subject[0].Digest = map[string]string{"sha256": sha256Sum}
	parsed.Subject[0].Name = "does-not-exist.tar.gz"
	require.Error(t, impl.checkProvenance(opts, parsed))
}

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
	"strings"
	"testing"

	intoto "github.com/in-toto/attestation/go/v1"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protodsse "github.com/sigstore/protobuf-specs/gen/pb-go/dsse"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"sigs.k8s.io/bom/pkg/spdx"
	"sigs.k8s.io/release-sdk/object"
	"sigs.k8s.io/release-utils/hash"
)

// signedProvenance wraps the statement in a sigstore bundle like the stage
// run does with krel sign attestation --in-place. The signature is not
// verified, so it doesn't have to be valid.
func signedProvenance(t *testing.T, statement []byte, payloadType string) []byte {
	t.Helper()

	data, err := protojson.Marshal(&protobundle.Bundle{
		MediaType: "application/vnd.dev.sigstore.bundle.v0.3+json",
		Content: &protobundle.Bundle_DsseEnvelope{DsseEnvelope: &protodsse.Envelope{
			Payload:     statement,
			PayloadType: payloadType,
			Signatures:  []*protodsse.Signature{{Sig: []byte("signature")}},
		}},
	})
	require.NoError(t, err)

	return data
}

func TestProcessAttestationAndCheckProvenance(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		signed bool
	}{
		{"bare statement", false},
		{"sigstore bundle", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testProcessAttestationAndCheckProvenance(t, tc.signed)
		})
	}
}

func testProcessAttestationAndCheckProvenance(t *testing.T, signed bool) {
	t.Helper()

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

	if signed {
		data = signedProvenance(t, data, inTotoPayloadType)
	}

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

	// A subject without any digest fails
	parsed.Subject[0].Digest = nil
	require.Error(t, impl.checkProvenance(opts, parsed))

	// Malformed digest values fail
	for _, digest := range []string{"not-a-hash", sha256Sum + "ff", "FF" + sha256Sum[2:]} {
		parsed.Subject[0].Digest = map[string]string{"sha256": digest}
		require.Error(t, impl.checkProvenance(opts, parsed), digest)
	}

	// A missing artifact fails
	parsed.Subject[0].Digest = map[string]string{"sha256": sha256Sum}
	parsed.Subject[0].Name = "does-not-exist.tar.gz"
	require.Error(t, impl.checkProvenance(opts, parsed))
}

// TestGenerateFinalAttestation runs on its own: the final attestation is
// written to the temporary directory of the process.
func TestGenerateFinalAttestation(t *testing.T) { //nolint:paralleltest // sets TMPDIR
	const (
		bucket  = "test-bucket"
		version = "v1.36.0"
	)

	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)

	// A release SBOM with one binary, built and written like the stage run
	// does in GenerateVersionArtifactsBOM
	binary := filepath.Join(t.TempDir(), "kubectl")
	require.NoError(t, os.WriteFile(binary, []byte("kubectl binary"), 0o600))

	file := spdx.NewFile()
	require.NoError(t, file.ReadSourceFile(binary))
	file.Name = filepath.Join("bin", "linux", "amd64", "kubectl")
	file.FileName = file.Name
	doc := spdx.NewDocument()
	doc.Name = "Kubernetes Release " + version
	doc.Namespace = "https://sbom.k8s.io/" + version + "/release"
	require.NoError(t, doc.AddFile(file))
	sbom := filepath.Join(t.TempDir(), "kubernetes-release.spdx")
	require.NoError(t, doc.Write(sbom))

	// The SLSA v1 provenance of the stage run, signed in place
	predicate, err := structpb.NewStruct(map[string]any{
		"buildDefinition": map[string]any{"buildType": "https://k8s.io/release/stage"},
	})
	require.NoError(t, err)
	stageStatement, err := protojson.Marshal(&intoto.Statement{
		Type:          intoto.StatementTypeUri,
		PredicateType: "https://slsa.dev/provenance/v1",
		Subject: []*intoto.ResourceDescriptor{{
			Name:   "gs://test-bucket/stage/kubernetes.tar.gz",
			Digest: map[string]string{"sha256": strings.Repeat("b", 64)},
		}},
		Predicate: predicate,
	})
	require.NoError(t, err)

	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"bare statement", stageStatement},
		{"sigstore bundle", signedProvenance(t, stageStatement, inTotoPayloadType)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stageProvenance := filepath.Join(t.TempDir(), ProvenanceFilename)
			require.NoError(t, os.WriteFile(stageProvenance, tc.data, 0o600))

			impl := &defaultProvenanceCheckerImpl{}
			require.NoError(t, impl.generateFinalAttestation(
				&ProvenanceCheckerOptions{StageBucket: bucket}, sbom, stageProvenance, version,
			))

			data, err := os.ReadFile(filepath.Join(tmpDir, "provenance-"+version+".json"))
			require.NoError(t, err)

			final := &intoto.Statement{}
			require.NoError(t, protojson.Unmarshal(data, final))

			require.Equal(t, intoto.StatementTypeUri, final.GetType())
			require.Equal(t, "https://slsa.dev/provenance/v1", final.GetPredicateType())
			require.Equal(t,
				"https://k8s.io/release/stage",
				final.GetPredicate().GetFields()["buildDefinition"].GetStructValue().GetFields()["buildType"].GetStringValue(),
			)

			// The subjects are the files of the SBOM in the release bucket
			require.NotEmpty(t, final.GetSubject())

			for _, sub := range final.GetSubject() {
				require.True(t,
					strings.HasPrefix(sub.GetName(), "gs://"+bucket+"/release/"+version+"/"), sub.GetName(),
				)
			}

			require.Contains(t, subjectNames(final), "gs://"+bucket+"/release/"+version+"/bin/linux/amd64/kubectl")
		})
	}
}

func subjectNames(s *intoto.Statement) []string {
	names := make([]string, 0, len(s.GetSubject()))
	for _, sub := range s.GetSubject() {
		names = append(names, sub.GetName())
	}

	return names
}

func TestReadStageStatementInvalidBundle(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		data []byte
	}{
		{
			name: "another payload type",
			data: signedProvenance(t, []byte("{}"), "application/octet-stream"),
		},
		{
			name: "no DSSE envelope",
			data: []byte(`{"mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json"}`),
		},
		{
			name: "malformed bundle",
			data: []byte(`{"mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json", "dsseEnvelope": 42}`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), ProvenanceFilename)
			require.NoError(t, os.WriteFile(path, tc.data, 0o600))
			_, err := readStageStatement(path)
			require.Error(t, err)
		})
	}
}

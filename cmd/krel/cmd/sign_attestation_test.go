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

package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunSignAttestation(t *testing.T) {
	t.Parallel()

	signOpts := &signOptions{timeout: time.Second}

	t.Run("missing statement", func(t *testing.T) {
		t.Parallel()

		err := runSignAttestation(
			signOpts, &signAttestationOptions{}, filepath.Join(t.TempDir(), "missing.json"),
		)
		require.Error(t, err)
	})

	t.Run("invalid service account key from the environment", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		statement := filepath.Join(dir, "statement.json")
		require.NoError(t, os.WriteFile(statement, []byte(`{
			"_type": "https://in-toto.io/Statement/v1",
			"subject": [{"name": "a", "digest": {"sha256": "0e8a8b6f7c6cf3b0f2f2b6c2d1a4f4b3c2e1d0f9a8b7c6d5e4f3a2b1c0d9e8f7"}}],
			"predicateType": "https://example.com/test", "predicate": {}
		}`), 0o600))

		err := runSignAttestation(
			signOpts, &signAttestationOptions{serviceAccountJSON: `{"type": "authorized_user"}`}, statement,
		)
		require.ErrorContains(t, err, "not a service account key")
	})

	t.Run("output file is not touched when signing fails", func(t *testing.T) {
		t.Parallel()

		outputPath := filepath.Join(t.TempDir(), "out.json")

		err := runSignAttestation(
			signOpts,
			&signAttestationOptions{outputPath: outputPath},
			filepath.Join(t.TempDir(), "missing.json"),
		)
		require.ErrorContains(t, err, "signing attestation")
		require.NoFileExists(t, outputPath)
	})
}

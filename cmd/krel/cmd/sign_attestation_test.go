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

	t.Run("output path cannot be created", func(t *testing.T) {
		t.Parallel()

		err := runSignAttestation(
			signOpts,
			&signAttestationOptions{outputPath: filepath.Join(t.TempDir(), "no", "such", "dir", "out.json")},
			filepath.Join(t.TempDir(), "missing.json"),
		)
		require.ErrorContains(t, err, "creating output file")
	})
}

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

package anago

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/attestation"
)

type fakeProvenanceSigner struct {
	paths []string
	err   error
}

func (f *fakeProvenanceSigner) SignFilesInPlace(paths []string) ([]*attestation.SignedStatement, error) {
	f.paths = paths

	return nil, f.err
}

// TestSignProvenance checks which final attestations SignProvenance signs and
// as which account, without reaching out to sigstore: the attestations are
// read from the temporary directory of the process.
func TestSignProvenance(t *testing.T) { //nolint:paralleltest // sets TMPDIR and the signing account
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)

	// v1.36.0 has a final attestation, v1.36.1 none
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "provenance-v1.36.0.json"), []byte("{}"), 0o600))

	var gotOpts *attestation.SignerOptions

	signer := &fakeProvenanceSigner{}
	impl := &defaultReleaseImpl{newProvenanceSigner: func(opts *attestation.SignerOptions) provenanceSigner {
		gotOpts = opts

		return signer
	}}

	// Without the signing account nothing is signed
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_NAME", "")
	require.Error(t, impl.SignProvenance([]string{"v1.36.0"}))
	require.Nil(t, signer.paths)

	t.Setenv("GOOGLE_SERVICE_ACCOUNT_NAME", "krel-staging@k8s-releng-prod.iam.gserviceaccount.com")
	require.NoError(t, impl.SignProvenance([]string{"v1.36.0"}))
	require.Equal(t, []string{filepath.Join(tmpDir, "provenance-v1.36.0.json")}, signer.paths)
	require.Equal(t, "krel-staging@k8s-releng-prod.iam.gserviceaccount.com", gotOpts.ImpersonateServiceAccount)

	// A missing final attestation fails before anything is signed
	signer.paths = nil

	require.Error(t, impl.SignProvenance([]string{"v1.36.0", "v1.36.1"}))
	require.Nil(t, signer.paths)

	// A signing failure fails
	signer.err = errors.New("signing failed")

	require.Error(t, impl.SignProvenance([]string{"v1.36.0"}))
}

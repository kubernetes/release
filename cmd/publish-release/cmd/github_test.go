/*
Copyright 2022 The Kubernetes Authors.

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

	"github.com/stretchr/testify/require"
)

func TestProcessRemoteAsset(t *testing.T) {
	// Remove the fefault app creds temporarily
	const gacVar = "GOOGLE_APPLICATION_CREDENTIALS"

	prev := os.Getenv(gacVar)
	defer t.Setenv(gacVar, prev)
	t.Setenv(gacVar, "")

	files := []string{}
	defer func() {
		for _, f := range files {
			os.RemoveAll(f)
		}
	}()

	path, err := processRemoteAsset("gs://kubernetes-release/release/v1.25.1/kubernetes.tar.gz.sha512")
	require.NoError(t, err)
	require.FileExists(t, path)
	require.Equal(t, "kubernetes.tar.gz.sha512", filepath.Base(path))
	files = append(files, path)

	// Non existent object should fail
	_, err = processRemoteAsset("gs://kubernetes-release/release/v1.25.1/0000000")
	require.Error(t, err)
}

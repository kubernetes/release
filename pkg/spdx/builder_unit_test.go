/*
Copyright 2021 The Kubernetes Authors.

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

package spdx

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var testConfig = `---
namespace: http://www.example.com/
license: Apache-2.0
name: bom-test
creator:
    person: Kubernetes Release Managers (release-managers@kubernetes.io)
    tool: bom
artifacts:
    - type: directory
      source: .
      license: Apache-2.0
      gomodules: true
    - type: file
      source: ./SECURITY.md
    - type: image
      source: k8s.gcr.io/kube-apiserver:v1.22.0-alpha.2
    - type: docker-archive
      source: tmp/sample-images/kube-apiserver.tar
`

func TestYAMLParse(t *testing.T) {
	opts := &DocGenerateOptions{}
	impl := defaultDocBuilderImpl{}
	f, err := os.CreateTemp("", "*.yaml")
	require.Nil(t, err)
	defer os.Remove(f.Name())
	require.Nil(t, os.WriteFile(f.Name(), []byte(testConfig), os.FileMode(0o644)))

	require.Nil(t, impl.ReadYamlConfiguration(f.Name(), opts))

	require.Equal(t, 1, len(opts.Images))
	require.Equal(t, 1, len(opts.Files))
	require.Equal(t, 1, len(opts.Tarballs))
	require.Equal(t, 1, len(opts.Directories))

	require.Equal(t, "./SECURITY.md", opts.Files[0])
	require.Equal(t, "k8s.gcr.io/kube-apiserver:v1.22.0-alpha.2", opts.Images[0])
	require.Equal(t, ".", opts.Directories[0])
	require.Equal(t, "tmp/sample-images/kube-apiserver.tar", opts.Tarballs[0])

	require.Equal(t, "Kubernetes Release Managers (release-managers@kubernetes.io)", opts.CreatorPerson)
	require.Equal(t, "http://www.example.com/", opts.Namespace)
	require.Equal(t, "bom-test", opts.Name)
	require.Equal(t, "Apache-2.0", opts.License)
}

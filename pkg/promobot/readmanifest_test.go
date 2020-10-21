/*
Copyright 2020 The Kubernetes Authors.

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

package promobot_test

import (
	"testing"

	"k8s.io/release/pkg/promobot"
	"sigs.k8s.io/yaml"
)

func TestReadManifests(t *testing.T) {
	grid := []struct {
		Expected string
		Options  promobot.PromoteFilesOptions
	}{
		{
			Expected: "testdata/manifests/onefiles/expected.yaml",
			Options: promobot.PromoteFilesOptions{
				FilestoresPath: "testdata/manifests/onefiles/filestores.yaml",
				FilesPath:      "testdata/manifests/onefiles/files.yaml",
			},
		},
		{
			Expected: "testdata/manifests/manyfiles/expected.yaml",
			Options: promobot.PromoteFilesOptions{
				FilestoresPath: "testdata/manifests/manyfiles/filestores.yaml",
				FilesPath:      "testdata/manifests/manyfiles/files/",
			},
		},
	}

	for _, g := range grid {
		g := g // avoid closure go-tcha
		t.Run(g.Expected, func(t *testing.T) {
			manifest, err := promobot.ReadManifest(g.Options)
			if err != nil {
				t.Fatalf("failed to read manifest: %v", err)
			}

			manifestYAML, err := yaml.Marshal(manifest)
			if err != nil {
				t.Fatalf("error serializing manifest: %v", err)
			}

			AssertMatchesFile(t, string(manifestYAML), g.Expected)
		})
	}
}

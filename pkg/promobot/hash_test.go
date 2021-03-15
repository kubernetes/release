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
	"context"
	"os"
	"testing"

	"k8s.io/release/pkg/promobot"
	"k8s.io/utils/diff"
	"sigs.k8s.io/yaml"
)

func TestHash(t *testing.T) {
	ctx := context.Background()

	var opt promobot.GenerateManifestOptions
	opt.PopulateDefaults()

	opt.BaseDir = "testdata/files"

	manifest, err := promobot.GenerateManifest(ctx, opt)
	if err != nil {
		t.Fatalf("failed to generate manifest: %v", err)
	}

	manifestYAML, err := yaml.Marshal(manifest)
	if err != nil {
		t.Fatalf("error serializing manifest: %v", err)
	}

	AssertMatchesFile(t, string(manifestYAML), "testdata/files-manifest.yaml")
}

// AssertMatchesFile verifies that the contents of p match actual.
//
//  We break this out into a file because we also support the
//  UPDATE_EXPECTED_OUTPUT magic env var. When that env var is
//  set, we will write the actual output to the expected file, which
//  is very handy when making bigger changes.  The intention of these
//  tests is to make the changes explicit, particularly in code
//  review, not to force manual updates.
func AssertMatchesFile(t *testing.T, actual, p string) {
	b, err := os.ReadFile(p)
	if err != nil {
		if os.Getenv("UPDATE_EXPECTED_OUTPUT") == "" {
			t.Fatalf("error reading file %q: %v", p, err)
		}
	}

	expected := string(b)

	if actual != expected {
		if os.Getenv("UPDATE_EXPECTED_OUTPUT") != "" {
			if err := os.WriteFile(p, []byte(actual), 0644); err != nil {
				t.Fatalf("error writing file %q: %v", p, err)
			}
		}
		t.Errorf("actual did not match expected; diff=%s", diff.StringDiff(actual, expected))
	}
}

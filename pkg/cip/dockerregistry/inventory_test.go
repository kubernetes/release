/*
Copyright 2019 The Kubernetes Authors.

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

package inventory_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	cr "github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/require"

	reg "k8s.io/release/pkg/cip/dockerregistry"
	"k8s.io/release/pkg/cip/json"
	"k8s.io/release/pkg/cip/stream"
)

type ParseJSONStreamResult struct {
	jsons json.Objects
	err   error
}

func TestReadJSONStream(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput ParseJSONStreamResult
	}{
		{
			name:  "Blank input stream",
			input: `[]`,
			expectedOutput: ParseJSONStreamResult{
				json.Objects{},
				nil,
			},
		},
		// The order of the maps matters.
		{
			name: "Simple case",
			input: `[
  {
    "name": "gcr.io/louhi-gke-k8s/addon-resizer"
  },
  {
    "name": "gcr.io/louhi-gke-k8s/pause"
  }
]`,
			expectedOutput: ParseJSONStreamResult{
				json.Objects{
					{"name": "gcr.io/louhi-gke-k8s/addon-resizer"},
					{"name": "gcr.io/louhi-gke-k8s/pause"},
				},
				nil,
			},
		},
		// The order of the maps matters.
		{
			"Expected failure: missing closing brace",
			`[
  {
    "name": "gcr.io/louhi-gke-k8s/addon-resizer"
  ,
]`,
			ParseJSONStreamResult{
				nil,
				errors.New("yaml: line 4: did not find expected node content"),
			},
		},
	}

	// Test only the JSON unmarshalling logic.
	for _, test := range tests {
		var sr stream.Fake
		sr.Bytes = []byte(test.input)
		stdout, _, err := sr.Produce()
		require.Nil(t, err)

		jsons, err := json.Consume(stdout)
		defer sr.Close()

		if test.expectedOutput.err != nil {
			require.NotNil(t, err)
			require.Error(t, err, test.expectedOutput.err)
		} else {
			require.Nil(t, err)
		}

		require.Equal(t, jsons, test.expectedOutput.jsons)
	}
}

func TestParseRegistryManifest(t *testing.T) {
	// TODO: Create a function to convert an Manifest to a YAML
	// representation, and vice-versa.
	//
	// TODO: Use property-based testing to test the fidelity of the conversion
	// (marshaling/unmarshaling) functions.
	tests := []struct {
		name           string
		input          string
		expectedOutput reg.Manifest
		expectedError  error
	}{
		{
			"Empty manifest (invalid)",
			``,
			reg.Manifest{},
			fmt.Errorf(`'registries' field cannot be empty`),
		},
		{
			"Stub manifest (`images` field is empty)",
			`registries:
- name: gcr.io/bar
  service-account: foobar@google-containers.iam.gserviceaccount.com
- name: gcr.io/foo
  service-account: src@google-containers.iam.gserviceaccount.com
  src: true
images: []
`,
			reg.Manifest{
				Registries: []reg.RegistryContext{
					{
						Name:           "gcr.io/bar",
						ServiceAccount: "foobar@google-containers.iam.gserviceaccount.com",
					},
					{
						Name:           "gcr.io/foo",
						ServiceAccount: "src@google-containers.iam.gserviceaccount.com",
						Src:            true,
					},
				},

				Images: []reg.Image{},
			},
			nil,
		},
		{
			"Basic manifest",
			`registries:
- name: gcr.io/bar
  service-account: foobar@google-containers.iam.gserviceaccount.com
- name: gcr.io/foo
  service-account: src@google-containers.iam.gserviceaccount.com
  src: true
images:
- name: agave
  dmap:
    "sha256:aab34c5841987a1b133388fa9f27e7960c4b1307e2f9147dca407ba26af48a54": ["latest"]
- name: banana
  dmap:
    "sha256:07353f7b26327f0d933515a22b1de587b040d3d85c464ea299c1b9f242529326": [ "1.8.3" ]  # Branches: ['master']
`,
			reg.Manifest{
				Registries: []reg.RegistryContext{
					{
						Name:           "gcr.io/bar",
						ServiceAccount: "foobar@google-containers.iam.gserviceaccount.com",
					},
					{
						Name:           "gcr.io/foo",
						ServiceAccount: "src@google-containers.iam.gserviceaccount.com",
						Src:            true,
					},
				},

				Images: []reg.Image{
					{
						ImageName: "agave",
						Dmap: reg.DigestTags{
							"sha256:aab34c5841987a1b133388fa9f27e7960c4b1307e2f9147dca407ba26af48a54": {"latest"},
						},
					},
					{
						ImageName: "banana",
						Dmap: reg.DigestTags{
							"sha256:07353f7b26327f0d933515a22b1de587b040d3d85c464ea299c1b9f242529326": {"1.8.3"},
						},
					},
				},
			},
			nil,
		},
		{
			"Missing src registry in registries (invalid)",
			`registries:
- name: gcr.io/bar
  service-account: foobar@google-containers.iam.gserviceaccount.com
- name: gcr.io/foo
  service-account: src@google-containers.iam.gserviceaccount.com
images:
- name: agave
  dmap:
    "sha256:aab34c5841987a1b133388fa9f27e7960c4b1307e2f9147dca407ba26af48a54": ["latest"]
- name: banana
  dmap:
    "sha256:07353f7b26327f0d933515a22b1de587b040d3d85c464ea299c1b9f242529326": [ "1.8.3" ]  # Branches: ['master']
`,
			reg.Manifest{},
			fmt.Errorf("source registry must be set"),
		},
	}

	// Test only the JSON unmarshalling logic.
	for _, test := range tests {
		b := []byte(test.input)

		imageManifest, err := reg.ParseManifestYAML(b)
		if test.expectedError != nil {
			require.NotNil(t, err)
			require.Error(t, err, test.expectedError)
		} else {
			require.Nil(t, err)
			require.Equal(t, imageManifest, test.expectedOutput)
		}
	}
}

func TestParseThinManifestsFromDir(t *testing.T) {
	pwd := bazelTestPath("TestParseThinManifestsFromDir")

	tests := []struct {
		name string
		// "input" is folder name, relative to the location of this source file.
		input              string
		expectedOutput     []reg.Manifest
		expectedParseError error
	}{
		{
			"No manifests found (invalid)",
			"empty",
			[]reg.Manifest{},
			&os.PathError{
				Op:   "stat",
				Path: filepath.Join(pwd, "empty/images"),
				Err:  fmt.Errorf("no such file or directory"),
			},
		},
		{
			"Singleton (single manifest)",
			"singleton",
			[]reg.Manifest{
				{
					Registries: []reg.RegistryContext{
						{
							Name:           "gcr.io/foo-staging",
							ServiceAccount: "sa@robot.com",
							Src:            true,
						},
						{
							Name:           "us.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "eu.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "asia.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
					},
					Images: []reg.Image{
						{
							ImageName: "foo-controller",
							Dmap: reg.DigestTags{
								"sha256:c3d310f4741b3642497da8826e0986db5e02afc9777a2b8e668c8e41034128c1": {"1.0"},
							},
						},
					},
					Filepath: "manifests/a/promoter-manifest.yaml",
				},
			},
			nil,
		},
		{
			"Multiple (with 'rebase')",
			"multiple-rebases",
			[]reg.Manifest{
				{
					Registries: []reg.RegistryContext{
						{
							Name:           "gcr.io/foo-staging",
							ServiceAccount: "sa@robot.com",
							Src:            true,
						},
						{
							Name:           "us.gcr.io/some-prod/foo",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "eu.gcr.io/some-prod/foo",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "asia.gcr.io/some-prod/foo",
							ServiceAccount: "sa@robot.com",
						},
					},
					Images: []reg.Image{
						{
							ImageName: "foo-controller",
							Dmap: reg.DigestTags{
								"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {"1.0"},
							},
						},
					},
					Filepath: "manifests/a/promoter-manifest.yaml",
				},
				{
					Registries: []reg.RegistryContext{
						{
							Name:           "gcr.io/bar-staging",
							ServiceAccount: "sa@robot.com",
							Src:            true,
						},
						{
							Name:           "us.gcr.io/some-prod/bar",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "eu.gcr.io/some-prod/bar",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "asia.gcr.io/some-prod/bar",
							ServiceAccount: "sa@robot.com",
						},
					},
					Images: []reg.Image{
						{
							ImageName: "bar-controller",
							Dmap: reg.DigestTags{
								"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb": {"1.0"},
							},
						},
					},
					Filepath: "manifests/b/promoter-manifest.yaml",
				},
			},
			nil,
		},
		{
			"Basic (multiple thin manifests)",
			"basic-thin",
			[]reg.Manifest{
				{
					Registries: []reg.RegistryContext{
						{
							Name:           "gcr.io/foo-staging",
							ServiceAccount: "sa@robot.com",
							Src:            true,
						},
						{
							Name:           "us.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "eu.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "asia.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
					},
					Images: []reg.Image{
						{
							ImageName: "foo-controller",
							Dmap: reg.DigestTags{
								"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {"1.0"},
							},
						},
					},
					Filepath: "manifests/a/promoter-manifest.yaml",
				},
				{
					Registries: []reg.RegistryContext{
						{
							Name:           "gcr.io/bar-staging",
							ServiceAccount: "sa@robot.com",
							Src:            true,
						},
						{
							Name:           "us.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "eu.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "asia.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
					},
					Images: []reg.Image{
						{
							ImageName: "bar-controller",
							Dmap: reg.DigestTags{
								"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb": {"1.0"},
							},
						},
					},
					Filepath: "manifests/b/promoter-manifest.yaml",
				},
				{
					Registries: []reg.RegistryContext{
						{
							Name:           "gcr.io/cat-staging",
							ServiceAccount: "sa@robot.com",
							Src:            true,
						},
						{
							Name:           "us.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "eu.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "asia.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
					},
					Images: []reg.Image{
						{
							ImageName: "cat-controller",
							Dmap: reg.DigestTags{
								"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc": {"1.0"},
							},
						},
					},
					Filepath: "manifests/c/promoter-manifest.yaml",
				},
				{
					Registries: []reg.RegistryContext{
						{
							Name:           "gcr.io/qux-staging",
							ServiceAccount: "sa@robot.com",
							Src:            true,
						},
						{
							Name:           "us.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "eu.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
						{
							Name:           "asia.gcr.io/some-prod",
							ServiceAccount: "sa@robot.com",
						},
					},
					Images: []reg.Image{
						{
							ImageName: "qux-controller",
							Dmap: reg.DigestTags{
								"sha256:0000000000000000000000000000000000000000000000000000000000000000": {"1.0"},
							},
						},
					},
					Filepath: "manifests/d/promoter-manifest.yaml",
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		fixtureDir := bazelTestPath("TestParseThinManifestsFromDir", test.input)

		// Fixup expected filepaths to match bazel's testing directory.
		expectedModified := test.expectedOutput[:0]
		for _, mfest := range test.expectedOutput {
			mfest.Filepath = filepath.Join(fixtureDir, mfest.Filepath)

			// SA4010: this result of append is never used, except maybe in other appends
			// nolint: staticcheck
			expectedModified = append(expectedModified, mfest)
		}

		got, errParse := reg.ParseThinManifestsFromDir(fixtureDir)
		if test.expectedParseError != nil {
			require.NotNil(t, errParse)
			require.Error(t, errParse, test.expectedParseError)
			continue
		}

		// Clear private fields (redundant data) that are calculated on-the-fly
		// (it's too verbose to include them here; besides, it's not what we're
		// testing).
		gotModified := got[:0]
		for _, mfest := range got {
			mfest.SrcRegistry = nil
			gotModified = append(gotModified, mfest)
		}

		require.Nil(t, errParse)
		require.Equal(t, gotModified, test.expectedOutput)
	}
}

func TestValidateThinManifestsFromDir(t *testing.T) {
	shouldBeValid := []string{
		"singleton",
		"multiple-rebases",
		"overlapping-src-registries",
		"overlapping-destination-vertices-same-digest",
		"malformed-directory-tree-structure-bad-prefix-is-ignored",
	}

	pwd := bazelTestPath("TestValidateThinManifestsFromDir")

	for _, testInput := range shouldBeValid {
		fixtureDir := filepath.Join(pwd, "valid", testInput)

		mfests, errParse := reg.ParseThinManifestsFromDir(fixtureDir)
		require.Nil(t, errParse)

		_, edgeErr := reg.ToPromotionEdges(mfests)
		require.Nil(t, edgeErr)
	}

	shouldBeInvalid := []struct {
		dirName            string
		expectedParseError error
		expectedEdgeError  error
	}{
		{
			"empty",
			&os.PathError{
				Op:   "stat",
				Path: filepath.Join(pwd, "invalid/empty/images"),
				Err:  fmt.Errorf("no such file or directory"),
			},
			nil,
		},
		{
			"overlapping-destination-vertices-different-digest",
			nil,
			fmt.Errorf("overlapping edges detected"),
		},
		{
			"malformed-directory-tree-structure",
			fmt.Errorf(
				"corresponding file %q does not exist",
				filepath.Join(pwd, "invalid/malformed-directory-tree-structure/images/b/images.yaml"),
			),
			nil,
		},
		{
			"malformed-directory-tree-structure-nested",
			fmt.Errorf(
				"unexpected manifest path %q",
				filepath.Join(pwd, "invalid/malformed-directory-tree-structure-nested/manifests/b/c/promoter-manifest.yaml"),
			),
			nil,
		},
	}

	for _, test := range shouldBeInvalid {
		fixtureDir := bazelTestPath("TestValidateThinManifestsFromDir", "invalid", test.dirName)

		// It could be that a manifest, taken individually, failed on its own,
		// before we even get to ValidateThinManifestsFromDir(). So handle these
		// cases as well.
		mfests, errParse := reg.ParseThinManifestsFromDir(fixtureDir)
		if test.expectedParseError != nil {
			require.NotNil(t, errParse)
			require.Error(t, errParse, test.expectedParseError)
		} else {
			require.Nil(t, errParse)
		}

		_, edgeErr := reg.ToPromotionEdges(mfests)
		if test.expectedEdgeError != nil {
			require.NotNil(t, edgeErr)
			require.Error(t, edgeErr, test.expectedEdgeError)
		} else {
			require.Nil(t, edgeErr)
		}
	}
}

func TestParseImageDigest(t *testing.T) {
	shouldBeValid := []string{
		`sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef`,
		`sha256:0000000000000000000000000000000000000000000000000000000000000000`,
		`sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`,
		`sha256:3243f6a8885a308d313198a2e03707344a4093822299f31d0082efa98ec4e6c8`,
	}

	for _, testInput := range shouldBeValid {
		d := reg.Digest(testInput)
		require.NotEmpty(t, d)

		err := reg.ValidateDigest(d)
		require.Nil(t, err)
	}

	shouldBeInvalid := []string{
		// Empty.
		``,
		// Too short.
		`sha256:0`,
		// Too long.
		`sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef1`,
		// Invalid character 'x'.
		`sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdex`,
		// No prefix 'sha256'.
		`0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef`,
	}

	for _, testInput := range shouldBeInvalid {
		d := reg.Digest(testInput)
		err := reg.ValidateDigest(d)
		require.NotNil(t, err)
	}
}

func TestParseImageTag(t *testing.T) {
	shouldBeValid := []string{
		`a`,
		`_`,
		`latest`,
		`_latest`,
		// Awkward, but valid.
		`_____----hello........`,
		// Longest tag is 128 chars.
		`this-is-exactly-128-chars-this-is-exactly-128-chars-this-is-exactly-128-chars-this-is-exactly-128-chars-this-is-exactly-128-char`,
	}

	for _, testInput := range shouldBeValid {
		tag := reg.Tag(testInput)
		err := reg.ValidateTag(tag)
		require.Nil(t, err)
	}

	shouldBeInvalid := []string{
		// Empty.
		``,
		// Does not begin with an ASCII word character.
		`.`,
		// Does not begin with an ASCII word character.
		`-`,
		// Unicode not allowed.
		`안녕`,
		// No spaces allowed.
		`a b`,
		// Too long (>128 ASCII chars).
		`this-is-longer-than-128-chars-this-is-longer-than-128-chars-this-is-longer-than-128-chars-this-is-longer-than-128-chars-this-is-l`,
	}

	for _, testInput := range shouldBeInvalid {
		tag := reg.Tag(testInput)
		err := reg.ValidateTag(tag)
		require.NotNil(t, err)
	}
}

func TestValidateRegistryImagePath(t *testing.T) {
	shouldBeValid := []string{
		`gcr.io/foo/bar`,
		`k8s.gcr.io/foo`,
		`staging-k8s.gcr.io/foo`,
		`staging-k8s.gcr.io/foo/bar/nested/path/image`,
	}

	for _, testInput := range shouldBeValid {
		rip := reg.RegistryImagePath(testInput)
		require.NotEmpty(t, rip)

		err := reg.ValidateRegistryImagePath(rip)
		require.Nil(t, err)
	}

	shouldBeInvalid := []string{
		// Empty.
		``,
		// No dot.
		`gcrio`,
		// Too many dots.
		`gcr..io`,
		// Leading dot.
		`.gcr.io`,
		// Trailing dot.
		`gcr.io.`,
		// Too many slashes.
		`gcr.io//foo`,
		// Leading slash.
		`/gcr.io`,
		// Trailing slash (1).
		`gcr.io/`,
		// Trailing slash (2).
		`gcr.io/foo/`,
	}

	for _, testInput := range shouldBeInvalid {
		rip := reg.RegistryImagePath(testInput)
		err := reg.ValidateRegistryImagePath(rip)
		require.NotNil(t, err)
	}
}

func TestSplitRegistryImagePath(t *testing.T) {
	knownRegistryNames := []reg.RegistryName{
		`gcr.io/foo`,
		`us.gcr.io/foo`,
		`k8s.gcr.io`,
		`eu.gcr.io/foo/d`,
	}

	tests := []struct {
		name                 string
		input                reg.RegistryImagePath
		expectedRegistryName reg.RegistryName
		expectedImageName    reg.ImageName
		expectedErr          error
	}{
		{
			`basic gcr.io`,
			`gcr.io/foo/a/b/c`,
			`gcr.io/foo`,
			`a/b/c`,
			nil,
		},
		{
			`regional GCR`,
			`us.gcr.io/foo/a/b/c`,
			`us.gcr.io/foo`,
			`a/b/c`,
			nil,
		},
		{
			`regional GCR (extra level of nesting)`,
			`eu.gcr.io/foo/d/e/f`,
			`eu.gcr.io/foo/d`,
			`e/f`,
			nil,
		},
		{
			`vanity GCR`,
			`k8s.gcr.io/a/b/c`,
			`k8s.gcr.io`,
			`a/b/c`,
			nil,
		},
	}
	for _, test := range tests {
		rName, iName, err := reg.SplitRegistryImagePath(test.input, knownRegistryNames)
		if test.expectedErr != nil {
			require.NotNil(t, err)
			require.Error(t, err, test.expectedErr)
		} else {
			require.Nil(t, err)
		}

		require.Equal(t, rName, test.expectedRegistryName)
		require.Equal(t, iName, test.expectedImageName)
	}
}

func TestSplitByKnownRegistries(t *testing.T) {
	knownRegistryNames := []reg.RegistryName{
		// See
		// https://github.com/kubernetes-sigs/k8s-container-image-promoter/issues/188.
		`us.gcr.io/k8s-artifacts-prod/kube-state-metrics`,
		`us.gcr.io/k8s-artifacts-prod/metrics-server`,
		`us.gcr.io/k8s-artifacts-prod`,
	}
	knownRegistryContexts := make([]reg.RegistryContext, 0)
	for _, knownRegistryName := range knownRegistryNames {
		rc := reg.RegistryContext{}
		rc.Name = knownRegistryName
		knownRegistryContexts = append(knownRegistryContexts, rc)
	}

	tests := []struct {
		name                 string
		input                reg.RegistryName
		expectedRegistryName reg.RegistryName
		expectedImageName    reg.ImageName
		expectedErr          error
	}{
		{
			`image at toplevel root path`,
			`us.gcr.io/k8s-artifacts-prod/kube-state-metrics`,
			`us.gcr.io/k8s-artifacts-prod`,
			`kube-state-metrics`,
			nil,
		},
		{
			`unclean split (known repo cuts into middle of image name)`,
			`us.gcr.io/k8s-artifacts-prod/metrics-server-amd64`,
			`us.gcr.io/k8s-artifacts-prod`,
			`metrics-server-amd64`,
			nil,
		},
	}
	for _, test := range tests {
		rootReg, imageName, err := reg.SplitByKnownRegistries(test.input, knownRegistryContexts)
		if test.expectedErr != nil {
			require.NotNil(t, err)
			require.Error(t, err, test.expectedErr)
		} else {
			require.Nil(t, err)
		}

		require.Equal(t, rootReg, test.expectedRegistryName)
		require.Equal(t, imageName, test.expectedImageName)
	}
}

func TestCommandGeneration(t *testing.T) {
	destRC := reg.RegistryContext{
		Name:           "gcr.io/foo",
		ServiceAccount: "robot",
	}

	var (
		srcRegName    reg.RegistryName = "gcr.io/bar"
		srcImageName  reg.ImageName    = "baz"
		destImageName reg.ImageName    = "baz"
		digest        reg.Digest       = "sha256:000"
		tag           reg.Tag          = "1.0"
		tp            reg.TagOp
	)

	t.Run(
		"GetDeleteCmd",
		func(t *testing.T) {
			got := reg.GetDeleteCmd(
				destRC,
				true,
				destImageName,
				digest,
				false)

			expected := []string{
				"gcloud",
				"--account=robot",
				"container",
				"images",
				"delete",
				reg.ToFQIN(destRC.Name, destImageName, digest),
				"--format=json",
			}

			require.Equal(t, got, expected)

			got = reg.GetDeleteCmd(
				destRC,
				false,
				destImageName,
				digest,
				false,
			)

			expected = []string{
				"gcloud",
				"container",
				"images",
				"delete",
				reg.ToFQIN(destRC.Name, destImageName, digest),
				"--format=json",
			}

			require.Equal(t, got, expected)
		},
	)

	t.Run(
		"GetWriteCmd (Delete)",
		func(t *testing.T) {
			tp = reg.Delete

			got := reg.GetWriteCmd(
				destRC,
				true,
				srcRegName,
				srcImageName,
				destImageName,
				digest,
				tag,
				tp,
			)

			expected := []string{
				"gcloud",
				"--account=robot",
				"--quiet",
				"container",
				"images",
				"untag",
				reg.ToPQIN(destRC.Name, destImageName, tag),
			}

			require.Equal(t, got, expected)

			got = reg.GetWriteCmd(
				destRC,
				false,
				srcRegName,
				srcImageName,
				destImageName,
				digest,
				tag,
				tp,
			)

			expected = []string{
				"gcloud",
				"--quiet",
				"container",
				"images",
				"untag",
				reg.ToPQIN(destRC.Name, destImageName, tag),
			}

			require.Equal(t, got, expected)
		},
	)
}

// TestReadRegistries tests reading images and tags from a registry.
func TestReadRegistries(t *testing.T) {
	const fakeRegName reg.RegistryName = "gcr.io/foo"

	tests := []struct {
		name           string
		input          map[string]string
		expectedOutput reg.RegInvImage
	}{
		{
			"Only toplevel repos (no child repos)",
			map[string]string{
				"gcr.io/foo": `{
  "child": [
    "addon-resizer",
    "pause"
  ],
  "manifest": {},
  "name": "foo",
  "tags": []
}`,
				"gcr.io/foo/addon-resizer": `{
  "child": [],
  "manifest": {
    "sha256:b5b2d91319f049143806baeacc886f82f621e9a2550df856b11b5c22db4570a7": {
      "imageSizeBytes": "12875324",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [
        "latest"
      ],
      "timeCreatedMs": "1501774217070",
      "timeUploadedMs": "1552917295327"
    },
    "sha256:0519a83e8f217e33dd06fe7a7347444cfda5e2e29cf52aaa24755999cb104a4d": {
      "imageSizeBytes": "12875324",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [
        "1.0"
      ],
      "timeCreatedMs": "1501774217070",
      "timeUploadedMs": "1552917295327"
    }
  },
  "name": "foo/addon-resizer",
  "tags": [
    "latest",
    "1.0"
  ]
}`,
				"gcr.io/foo/pause": `{
  "child": [],
  "manifest": {
    "sha256:06fdf10aae2eeeac5a82c213e4693f82ab05b3b09b820fce95a7cac0bbdad534": {
      "imageSizeBytes": "12875324",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [
        "v1.2.3"
      ],
      "timeCreatedMs": "1501774217070",
      "timeUploadedMs": "1552917295327"
    }
  },
  "name": "foo/pause",
  "tags": [
    "v1.2.3"
  ]
}`,
			},
			reg.RegInvImage{
				"addon-resizer": {
					"sha256:b5b2d91319f049143806baeacc886f82f621e9a2550df856b11b5c22db4570a7": {"latest"},
					"sha256:0519a83e8f217e33dd06fe7a7347444cfda5e2e29cf52aaa24755999cb104a4d": {"1.0"},
				},
				"pause": {
					"sha256:06fdf10aae2eeeac5a82c213e4693f82ab05b3b09b820fce95a7cac0bbdad534": {"v1.2.3"},
				},
			},
		},
		{
			"Recursive repos (child repos)",
			map[string]string{
				"gcr.io/foo": `{
  "child": [
    "addon-resizer",
    "pause"
  ],
  "manifest": {},
  "name": "foo",
  "tags": []
}`,
				"gcr.io/foo/addon-resizer": `{
  "child": [],
  "manifest": {
    "sha256:b5b2d91319f049143806baeacc886f82f621e9a2550df856b11b5c22db4570a7": {
      "imageSizeBytes": "12875324",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [
        "latest"
      ],
      "timeCreatedMs": "1501774217070",
      "timeUploadedMs": "1552917295327"
    },
    "sha256:0519a83e8f217e33dd06fe7a7347444cfda5e2e29cf52aaa24755999cb104a4d": {
      "imageSizeBytes": "12875324",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [
        "1.0"
      ],
      "timeCreatedMs": "1501774217070",
      "timeUploadedMs": "1552917295327"
    }
  },
  "name": "foo/addon-resizer",
  "tags": [
    "latest",
    "1.0"
  ]
}`,
				"gcr.io/foo/pause": `{
  "child": [
    "childLevel1"
  ],
  "manifest": {
    "sha256:06fdf10aae2eeeac5a82c213e4693f82ab05b3b09b820fce95a7cac0bbdad534": {
      "imageSizeBytes": "12875324",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [
        "v1.2.3"
      ],
      "timeCreatedMs": "1501774217070",
      "timeUploadedMs": "1552917295327"
    }
  },
  "name": "foo/pause",
  "tags": [
    "v1.2.3"
  ]
}`,
				"gcr.io/foo/pause/childLevel1": `{
  "child": [
    "childLevel2"
  ],
  "manifest": {
    "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {
      "imageSizeBytes": "12875324",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [
        "aaa"
      ],
      "timeCreatedMs": "1501774217070",
      "timeUploadedMs": "1552917295327"
    }
  },
  "name": "foo/pause/childLevel1",
  "tags": [
    "aaa"
  ]
}`,
				"gcr.io/foo/pause/childLevel1/childLevel2": `{
  "child": [],
  "manifest": {
    "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff": {
      "imageSizeBytes": "12875324",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [
        "fff"
      ],
      "timeCreatedMs": "1501774217070",
      "timeUploadedMs": "1552917295327"
    }
  },
  "name": "foo/pause/childLevel1/childLevel2",
  "tags": [
    "fff"
  ]
}`,
			},
			reg.RegInvImage{
				"addon-resizer": {
					"sha256:b5b2d91319f049143806baeacc886f82f621e9a2550df856b11b5c22db4570a7": {"latest"},
					"sha256:0519a83e8f217e33dd06fe7a7347444cfda5e2e29cf52aaa24755999cb104a4d": {"1.0"},
				},
				"pause": {
					"sha256:06fdf10aae2eeeac5a82c213e4693f82ab05b3b09b820fce95a7cac0bbdad534": {"v1.2.3"},
				},
				"pause/childLevel1": {
					"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {"aaa"},
				},
				"pause/childLevel1/childLevel2": reg.DigestTags{
					"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff": {"fff"},
				},
			},
		},
	}

	for _, test := range tests {
		// Destination registry is a placeholder, because ReadImageNames acts on
		// 2 registries (src and dest) at once.
		rcs := []reg.RegistryContext{
			{
				Name:           fakeRegName,
				ServiceAccount: "robot",
			},
		}

		sc := reg.SyncContext{
			RegistryContexts: rcs,
			Inv:              map[reg.RegistryName]reg.RegInvImage{fakeRegName: nil},
			DigestMediaType:  make(reg.DigestMediaType),
			DigestImageSize:  make(reg.DigestImageSize),
		}

		// test is used to pin the "test" variable from the outer "range"
		// scope (see scopelint).
		test := test
		mkFakeStream1 := func(sc *reg.SyncContext, rc reg.RegistryContext) stream.Producer {
			var sr stream.Fake

			_, domain, repoPath := reg.GetTokenKeyDomainRepoPath(rc.Name)
			fakeHTTPBody, ok := test.input[domain+"/"+repoPath]
			if !ok {
				require.False(t, ok)
			}

			sr.Bytes = []byte(fakeHTTPBody)
			return &sr
		}

		sc.ReadRegistries(rcs, true, mkFakeStream1)
		got := sc.Inv[fakeRegName]
		require.Equal(t, got, test.expectedOutput)
	}
}

// TestReadGManifestLists tests reading ManifestList information from GCR.
func TestReadGManifestLists(t *testing.T) {
	const fakeRegName reg.RegistryName = "gcr.io/foo"

	tests := []struct {
		name           string
		input          map[string]string
		expectedOutput reg.ParentDigest
	}{
		{
			"Basic example",
			map[string]string{
				"gcr.io/foo/someImage": `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
   "manifests": [
      {
         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
         "size": 739,
         "digest": "sha256:0bd88bcba94f800715fca33ffc4bde430646a7c797237313cbccdcdef9f80f2d",
         "platform": {
            "architecture": "amd64",
            "os": "linux"
         }
      },
      {
         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
         "size": 739,
         "digest": "sha256:0ad4f92011b2fa5de88a6e6a2d8b97f38371246021c974760e5fc54b9b7069e5",
         "platform": {
            "architecture": "s390x",
            "os": "linux"
         }
      }
   ]
}`,
			},
			reg.ParentDigest{
				"sha256:0bd88bcba94f800715fca33ffc4bde430646a7c797237313cbccdcdef9f80f2d": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
				"sha256:0ad4f92011b2fa5de88a6e6a2d8b97f38371246021c974760e5fc54b9b7069e5": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
		},
	}

	for _, test := range tests {
		// Destination registry is a placeholder, because ReadImageNames acts on
		// 2 registries (src and dest) at once.
		rcs := []reg.RegistryContext{
			{
				Name:           fakeRegName,
				ServiceAccount: "robot",
			},
		}
		sc := reg.SyncContext{
			RegistryContexts: rcs,
			Inv: map[reg.RegistryName]reg.RegInvImage{
				"gcr.io/foo": {
					"someImage": reg.DigestTags{
						"sha256:0000000000000000000000000000000000000000000000000000000000000000": {"1.0"},
					},
				},
			},
			DigestMediaType: reg.DigestMediaType{
				"sha256:0000000000000000000000000000000000000000000000000000000000000000": cr.DockerManifestList,
			},
			DigestImageSize: make(reg.DigestImageSize),
			ParentDigest:    make(reg.ParentDigest),
		}

		// test is used to pin the "test" variable from the outer "range"
		// scope (see scopelint).
		test := test
		mkFakeStream1 := func(sc *reg.SyncContext, gmlc *reg.GCRManifestListContext) stream.Producer {
			var sr stream.Fake

			_, domain, repoPath := reg.GetTokenKeyDomainRepoPath(gmlc.RegistryContext.Name)
			fakeHTTPBody, ok := test.input[domain+"/"+repoPath+"/"+string(gmlc.ImageName)]
			if !ok {
				require.False(t, ok)
			}

			sr.Bytes = []byte(fakeHTTPBody)
			return &sr
		}

		sc.ReadGCRManifestLists(mkFakeStream1)
		got := sc.ParentDigest
		require.Equal(t, got, test.expectedOutput)
	}
}

func TestGetTokenKeyDomainRepoPath(t *testing.T) {
	type TokenKeyDomainRepoPath [3]string

	tests := []struct {
		name     string
		input    reg.RegistryName
		expected TokenKeyDomainRepoPath
	}{
		{
			"basic",
			"gcr.io/foo/bar",
			[3]string{"gcr.io/foo", "gcr.io", "foo/bar"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(
			test.name,
			func(t *testing.T) {
				tokenKey, domain, repoPath := reg.GetTokenKeyDomainRepoPath(test.input)

				require.Equal(t, tokenKey, test.expected[0])
				require.Equal(t, domain, test.expected[1])
				require.Equal(t, repoPath, test.expected[2])
			},
		)
	}
}

func TestSetManipulationsRegistryInventories(t *testing.T) {
	tests := []struct {
		name           string
		input1         reg.RegInvImage
		input2         reg.RegInvImage
		op             func(a, b reg.RegInvImage) reg.RegInvImage
		expectedOutput reg.RegInvImage
	}{
		{
			"Set Minus",
			reg.RegInvImage{
				"foo": {
					"sha256:abc": {"1.0", "latest"},
				},
				"bar": {
					"sha256:def": {"0.9"},
				},
			},
			reg.RegInvImage{
				"foo": {
					"sha256:abc": {"1.0", "latest"},
				},
				"bar": {
					"sha256:def": {"0.9"},
				},
			},
			reg.RegInvImage.Minus,
			reg.RegInvImage{},
		},
		{
			"Set Union",
			reg.RegInvImage{
				"foo": {
					"sha256:abc": {"1.0", "latest"},
				},
				"bar": {
					"sha256:def": {"0.9"},
				},
			},
			reg.RegInvImage{
				"apple": {
					"sha256:abc": {"1.0", "latest"},
				},
				"banana": {
					"sha256:def": {"0.9"},
				},
			},
			reg.RegInvImage.Union,
			reg.RegInvImage{
				"foo": {
					"sha256:abc": {"1.0", "latest"},
				},
				"bar": {
					"sha256:def": {"0.9"},
				},
				"apple": {
					"sha256:abc": {"1.0", "latest"},
				},
				"banana": {
					"sha256:def": {"0.9"},
				},
			},
		},
	}

	for _, test := range tests {
		got := test.op(test.input1, test.input2)
		require.Equal(t, got, test.expectedOutput)
	}
}

func TestSetManipulationsTags(t *testing.T) {
	tests := []struct {
		name           string
		input1         reg.TagSlice
		input2         reg.TagSlice
		op             func(a, b reg.TagSlice) reg.TagSet
		expectedOutput reg.TagSet
	}{
		{
			"Set Minus (both blank)",
			reg.TagSlice{},
			reg.TagSlice{},
			reg.TagSlice.Minus,
			reg.TagSet{},
		},
		{
			"Set Minus (first blank)",
			reg.TagSlice{},
			reg.TagSlice{"a"},
			reg.TagSlice.Minus,
			reg.TagSet{},
		},
		{
			"Set Minus (second blank)",
			reg.TagSlice{"a", "b"},
			reg.TagSlice{},
			reg.TagSlice.Minus,
			reg.TagSet{"a": nil, "b": nil},
		},
		{
			"Set Minus",
			reg.TagSlice{"a", "b"},
			reg.TagSlice{"b"},
			reg.TagSlice.Minus,
			reg.TagSet{"a": nil},
		},
		{
			"Set Union (both blank)",
			reg.TagSlice{},
			reg.TagSlice{},
			reg.TagSlice.Union,
			reg.TagSet{},
		},
		{
			"Set Union (first blank)",
			reg.TagSlice{},
			reg.TagSlice{"a"},
			reg.TagSlice.Union,
			reg.TagSet{"a": nil},
		},
		{
			"Set Union (second blank)",
			reg.TagSlice{"a"},
			reg.TagSlice{},
			reg.TagSlice.Union,
			reg.TagSet{"a": nil},
		},
		{
			"Set Union",
			reg.TagSlice{"a", "c"},
			reg.TagSlice{"b", "d"},
			reg.TagSlice.Union,
			reg.TagSet{"a": nil, "b": nil, "c": nil, "d": nil},
		},
		{
			"Set Intersection (no intersection)",
			reg.TagSlice{"a"},
			reg.TagSlice{"b"},
			reg.TagSlice.Intersection,
			reg.TagSet{},
		},
		{
			"Set Intersection (some intersection)",
			reg.TagSlice{"a", "b"},
			reg.TagSlice{"b", "c"},
			reg.TagSlice.Intersection,
			reg.TagSet{"b": nil},
		},
	}

	for _, test := range tests {
		got := test.op(test.input1, test.input2)
		require.Equal(t, got, test.expectedOutput)
	}
}

func TestSetManipulationsRegInvImageTag(t *testing.T) {
	tests := []struct {
		name           string
		input1         reg.RegInvImageTag
		input2         reg.RegInvImageTag
		op             func(a, b reg.RegInvImageTag) reg.RegInvImageTag
		expectedOutput reg.RegInvImageTag
	}{
		{
			"Set Minus (both blank)",
			reg.RegInvImageTag{},
			reg.RegInvImageTag{},
			reg.RegInvImageTag.Minus,
			reg.RegInvImageTag{},
		},
		{
			"Set Minus (first blank)",
			reg.RegInvImageTag{},
			reg.RegInvImageTag{
				reg.ImageTag{ImageName: "pear", Tag: "latest"}: "123",
			},
			reg.RegInvImageTag.Minus,
			reg.RegInvImageTag{},
		},
		{
			"Set Minus (second blank)",
			reg.RegInvImageTag{
				reg.ImageTag{
					ImageName: "pear",
					Tag:       "latest",
				}: "123",
			},
			reg.RegInvImageTag{},
			reg.RegInvImageTag.Minus,
			reg.RegInvImageTag{
				reg.ImageTag{
					ImageName: "pear",
					Tag:       "latest",
				}: "123",
			},
		},
		{
			"Set Intersection (both blank)",
			reg.RegInvImageTag{},
			reg.RegInvImageTag{},
			reg.RegInvImageTag.Intersection,
			reg.RegInvImageTag{},
		},
		{
			"Set Intersection (first blank)",
			reg.RegInvImageTag{},
			reg.RegInvImageTag{
				reg.ImageTag{ImageName: "pear", Tag: "latest"}: "123",
			},
			reg.RegInvImageTag.Intersection,
			reg.RegInvImageTag{},
		},
		{
			"Set Intersection (second blank)",
			reg.RegInvImageTag{
				reg.ImageTag{ImageName: "pear", Tag: "latest"}: "123",
			},
			reg.RegInvImageTag{},
			reg.RegInvImageTag.Intersection,
			reg.RegInvImageTag{},
		},
		{
			"Set Intersection (no intersection)",
			reg.RegInvImageTag{
				reg.ImageTag{ImageName: "pear", Tag: "latest"}: "123",
			},
			reg.RegInvImageTag{
				reg.ImageTag{ImageName: "pear", Tag: "1.0"}: "123",
			},
			reg.RegInvImageTag.Intersection,
			reg.RegInvImageTag{},
		},
		{
			"Set Intersection (some intersection)",
			reg.RegInvImageTag{
				reg.ImageTag{ImageName: "pear", Tag: "latest"}: "this-is-kept",
				reg.ImageTag{ImageName: "pear", Tag: "1.0"}:    "123",
			},
			reg.RegInvImageTag{
				reg.ImageTag{ImageName: "pear", Tag: "latest"}: "this-is-lost",
				reg.ImageTag{ImageName: "foo", Tag: "2.0"}:     "def",
			},
			// The intersection code throws out the second value, because it
			// treats a Map as a Set (and doesn't care about preserving
			// information for the key's value).
			reg.RegInvImageTag.Intersection,
			reg.RegInvImageTag{
				reg.ImageTag{ImageName: "pear", Tag: "latest"}: "this-is-kept",
			},
		},
	}

	for _, test := range tests {
		got := test.op(test.input1, test.input2)
		require.Equal(t, got, test.expectedOutput)
	}
}

func TestToPromotionEdges(t *testing.T) {
	srcRegName := reg.RegistryName("gcr.io/foo")
	destRegName := reg.RegistryName("gcr.io/bar")
	destRegName2 := reg.RegistryName("gcr.io/cat")
	destRC := reg.RegistryContext{
		Name:           destRegName,
		ServiceAccount: "robot",
	}
	destRC2 := reg.RegistryContext{
		Name:           destRegName2,
		ServiceAccount: "robot",
	}
	srcRC := reg.RegistryContext{
		Name:           srcRegName,
		ServiceAccount: "robot",
		Src:            true,
	}
	registries1 := []reg.RegistryContext{destRC, srcRC}
	registries2 := []reg.RegistryContext{destRC, srcRC, destRC2}

	sc := reg.SyncContext{
		Inv: reg.MasterInventory{
			"gcr.io/foo": reg.RegInvImage{
				"a": {
					"sha256:000": {"0.9"},
				},
				"c": {
					"sha256:222": {"2.0"},
					"sha256:333": {"3.0"},
				},
			},
			"gcr.io/bar": {
				"a": {
					"sha256:000": {"0.9"},
				},
				"b": {
					"sha256:111": {},
				},
				"c": {
					"sha256:222": {"2.0"},
					"sha256:333": {"3.0"},
				},
			},
			"gcr.io/cat": {
				"a": {
					"sha256:000": {"0.9"},
				},
				"c": {
					"sha256:222": {"2.0"},
					"sha256:333": {"3.0"},
				},
			},
		},
	}

	tests := []struct {
		name                  string
		input                 []reg.Manifest
		expectedInitial       map[reg.PromotionEdge]interface{}
		expectedInitialErr    error
		expectedFiltered      map[reg.PromotionEdge]interface{}
		expectedFilteredClean bool
	}{
		{
			"Basic case (1 new edge; already promoted)",
			[]reg.Manifest{
				{
					Registries: registries1,
					Images: []reg.Image{
						{
							ImageName: "a",
							Dmap: reg.DigestTags{
								"sha256:000": {"0.9"},
							},
						},
					},
					SrcRegistry: &srcRC,
				},
			},
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
				}: nil,
			},
			nil,
			make(map[reg.PromotionEdge]interface{}),
			true,
		},
		{
			"Basic case (2 new edges; already promoted)",
			[]reg.Manifest{
				{
					Registries: registries2,
					Images: []reg.Image{
						{
							ImageName: "a",
							Dmap: reg.DigestTags{
								"sha256:000": {"0.9"},
							},
						},
					},
					SrcRegistry: &srcRC,
				},
			},
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
				}: nil,
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC2,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
				}: nil,
			},
			nil,
			make(map[reg.PromotionEdge]interface{}),
			true,
		},
		{
			"Tag move (tag swap image c:2.0 and c:3.0)",
			[]reg.Manifest{
				{
					Registries: registries2,
					Images: []reg.Image{
						{
							ImageName: "c",
							Dmap: reg.DigestTags{
								"sha256:222": {"3.0"},
								"sha256:333": {"2.0"},
							},
						},
					},
					SrcRegistry: &srcRC,
				},
			},
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "c",
						Tag:       "2.0",
					},
					Digest:      "sha256:333",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "c",
						Tag:       "2.0",
					},
				}: nil,
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "c",
						Tag:       "3.0",
					},
					Digest:      "sha256:222",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "c",
						Tag:       "3.0",
					},
				}: nil,
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "c",
						Tag:       "2.0",
					},
					Digest:      "sha256:333",
					DstRegistry: destRC2,
					DstImageTag: reg.ImageTag{
						ImageName: "c",
						Tag:       "2.0",
					},
				}: nil,
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "c",
						Tag:       "3.0",
					},
					Digest:      "sha256:222",
					DstRegistry: destRC2,
					DstImageTag: reg.ImageTag{
						ImageName: "c",
						Tag:       "3.0",
					},
				}: nil,
			},
			nil,
			make(map[reg.PromotionEdge]interface{}),
			false,
		},
	}

	for _, test := range tests {
		// Finalize Manifests.
		for i := range test.input {
			require.Nil(t, test.input[i].Finalize())
		}

		got, gotErr := reg.ToPromotionEdges(test.input)
		if test.expectedInitialErr != nil {
			require.NotNil(t, gotErr)
			require.Error(t, gotErr, test.expectedInitialErr)
		}
		require.Equal(t, got, test.expectedInitial)

		got, gotClean := sc.GetPromotionCandidates(got)
		require.Equal(t, got, test.expectedFiltered)
		require.Equal(t, gotClean, test.expectedFilteredClean)
	}
}

func TestCheckOverlappingEdges(t *testing.T) {
	srcRegName := reg.RegistryName("gcr.io/foo")
	destRegName := reg.RegistryName("gcr.io/bar")
	destRC := reg.RegistryContext{
		Name:           destRegName,
		ServiceAccount: "robot",
	}
	srcRC := reg.RegistryContext{
		Name:           srcRegName,
		ServiceAccount: "robot",
		Src:            true,
	}

	tests := []struct {
		name        string
		input       map[reg.PromotionEdge]interface{}
		expected    map[reg.PromotionEdge]interface{}
		expectedErr error
	}{
		{
			"Basic case (0 edges)",
			make(map[reg.PromotionEdge]interface{}),
			make(map[reg.PromotionEdge]interface{}),
			nil,
		},
		{
			"Basic case (singleton edge, no overlapping edges)",
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
				}: nil,
			},
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
				}: nil,
			},
			nil,
		},
		{ // nolint: dupl
			"Basic case (two edges, no overlapping edges)",
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
				}: nil,
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "b",
						Tag:       "0.9",
					},
					Digest:      "sha256:111",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "b",
						Tag:       "0.9",
					},
				}: nil,
			},
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
				}: nil,
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "b",
						Tag:       "0.9",
					},
					Digest:      "sha256:111",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "b",
						Tag:       "0.9",
					},
				}: nil,
			},
			nil,
		},
		{
			"Basic case (two edges, overlapped)",
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
				}: nil,
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "b",
						Tag:       "0.9",
					},
					Digest:      "sha256:111",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
				}: nil,
			},
			nil,
			fmt.Errorf("overlapping edges detected"),
		},
		{ // nolint: dupl
			"Basic case (two tagless edges (different digests, same PQIN), no overlap)",
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "",
					},
				}: nil,
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "b",
						Tag:       "0.9",
					},
					Digest:      "sha256:111",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "",
					},
				}: nil,
			},
			map[reg.PromotionEdge]interface{}{
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "0.9",
					},
					Digest:      "sha256:000",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "",
					},
				}: nil,
				{
					SrcRegistry: srcRC,
					SrcImageTag: reg.ImageTag{
						ImageName: "b",
						Tag:       "0.9",
					},
					Digest:      "sha256:111",
					DstRegistry: destRC,
					DstImageTag: reg.ImageTag{
						ImageName: "a",
						Tag:       "",
					},
				}: nil,
			},
			nil,
		},
	}

	for _, test := range tests {
		got, gotErr := reg.CheckOverlappingEdges(test.input)
		if test.expectedErr != nil {
			require.NotNil(t, gotErr)
			require.Error(t, gotErr, test.expectedErr)
		}
		require.Equal(t, got, test.expected)
	}
}

type FakeCheckAlwaysSucceed struct{}

func (c *FakeCheckAlwaysSucceed) Run() error {
	return nil
}

type FakeCheckAlwaysFail struct{}

func (c *FakeCheckAlwaysFail) Run() error {
	return fmt.Errorf("there was an error in the pull request check")
}

func TestRunChecks(t *testing.T) {
	sc := reg.SyncContext{}

	tests := []struct {
		name     string
		checks   []reg.PreCheck
		expected error
	}{
		{
			"Checking pull request with successful checks",
			[]reg.PreCheck{
				&FakeCheckAlwaysSucceed{},
			},
			nil,
		},
		{
			"Checking pull request with unsuccessful checks",
			[]reg.PreCheck{
				&FakeCheckAlwaysFail{},
			},
			fmt.Errorf("1 error(s) encountered during the prechecks"),
		},
		{
			"Checking pull request with successful and unsuccessful checks",
			[]reg.PreCheck{
				&FakeCheckAlwaysSucceed{},
				&FakeCheckAlwaysFail{},
				&FakeCheckAlwaysFail{},
			},
			fmt.Errorf("2 error(s) encountered during the prechecks"),
		},
	}

	for _, test := range tests {
		got := sc.RunChecks(test.checks)
		require.Equal(t, got, test.expected)
	}
}

// TestPromotion is the most important test as it simulates the main job of the
// promoter.
func TestPromotion(t *testing.T) {
	// CapturedRequests is like a bitmap. We clear off bits (delete keys) for
	// each request that we see that got generated. Then it's just a matter of
	// ensuring that the map is empty. If it is not empty, we can just show what
	// it looks like (basically a list of all requests that did not get
	// generated).
	//
	// We could make it even more "powerful" by storing a histogram instead of a
	// set. Then we can check that all requests were generated exactly 1 time.
	srcRegName := reg.RegistryName("gcr.io/foo")
	destRegName := reg.RegistryName("gcr.io/bar")
	destRegName2 := reg.RegistryName("gcr.io/cat")
	destRC := reg.RegistryContext{
		Name:           destRegName,
		ServiceAccount: "robot",
	}
	destRC2 := reg.RegistryContext{
		Name:           destRegName2,
		ServiceAccount: "robot",
	}
	srcRC := reg.RegistryContext{
		Name:           srcRegName,
		ServiceAccount: "robot",
		Src:            true,
	}
	registries := []reg.RegistryContext{destRC, srcRC, destRC2}

	registriesRebase := []reg.RegistryContext{
		{
			Name:           reg.RegistryName("us.gcr.io/dog/some/subdir/path/foo"),
			ServiceAccount: "robot",
		},
		srcRC,
	}

	tests := []struct {
		name                  string
		inputM                reg.Manifest
		inputSc               reg.SyncContext
		badReads              []reg.RegistryName
		expectedReqs          reg.CapturedRequests
		expectedFilteredClean bool
	}{
		{
			// TODO: Use quickcheck to ensure certain properties.
			"No promotion",
			reg.Manifest{},
			reg.SyncContext{},
			nil,
			reg.CapturedRequests{},
			true,
		},
		{
			"No promotion; tag is already promoted",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"0.9"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
					"gcr.io/bar": {
						"a": {
							"sha256:000": {"0.9"},
						},
						"b": {
							"sha256:111": {},
						},
					},
					"gcr.io/cat": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{},
			true,
		},
		{
			"No promotion; network errors reading from src registry for all images",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"0.9"},
						},
					},
					{
						ImageName: "b",
						Dmap: reg.DigestTags{
							"sha256:111": {"0.9"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							"sha256:000": {"0.9"},
						},
						"b": {
							"sha256:111": {"0.9"},
						},
					},
				},
				InvIgnore: []reg.ImageName{},
			},
			[]reg.RegistryName{"gcr.io/foo/a", "gcr.io/foo/b", "gcr.io/foo/c"},
			reg.CapturedRequests{},
			true,
		},
		{
			"Promote 1 tag; image digest does not exist in dest",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"0.9"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
					"gcr.io/bar": {
						"b": {
							"sha256:111": {},
						},
					},
					"gcr.io/cat": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{
				reg.PromotionRequest{
					TagOp:          reg.Add,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[0].Name,
					ServiceAccount: registries[0].ServiceAccount,
					ImageNameSrc:   "a",
					ImageNameDest:  "a",
					Digest:         "sha256:000",
					Tag:            "0.9",
				}: 1,
			},
			true,
		},
		{
			"Promote 1 tag; image already exists in dest, but digest does not",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"0.9"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
					"gcr.io/bar": {
						"a": {
							"sha256:111": {},
						},
					},
					"gcr.io/cat": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{
				reg.PromotionRequest{
					TagOp:          reg.Add,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[0].Name,
					ServiceAccount: registries[0].ServiceAccount,
					ImageNameSrc:   "a",
					ImageNameDest:  "a",
					Digest:         "sha256:000",
					Tag:            "0.9",
				}: 1,
			},
			true,
		},
		{
			"Promote 1 tag; tag already exists in dest but is pointing to a different digest (move tag)",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							// sha256:bad is a bad image uploaded by a
							// compromised account. "good" is a good tag that is
							// already known and used for this image "a" (and in
							// both gcr.io/bar and gcr.io/cat, point to a known
							// good digest, 600d.).
							"sha256:bad": {"good"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							// Malicious image.
							"sha256:bad": {"some-other-tag"},
						},
					},
					"gcr.io/bar": {
						"a": {
							"sha256:bad":  {"some-other-tag"},
							"sha256:600d": {"good"},
						},
					},
					"gcr.io/cat": {
						"a": {
							"sha256:bad":  {"some-other-tag"},
							"sha256:600d": {"good"},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{},
			false,
		},
		{
			"Promote 1 tag as a 'rebase'",
			reg.Manifest{
				Registries: registriesRebase,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"0.9"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
					"us.gcr.io/dog/some/subdir/path": {
						"a": {
							"sha256:111": {"0.8"},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{
				reg.PromotionRequest{
					TagOp:          reg.Add,
					RegistrySrc:    srcRegName,
					RegistryDest:   registriesRebase[0].Name,
					ServiceAccount: registriesRebase[0].ServiceAccount,
					ImageNameSrc:   "a",
					ImageNameDest:  "a",
					Digest:         "sha256:000",
					Tag:            "0.9",
				}: 1,
			},
			true,
		},
		{
			"Promote 1 digest (tagless promotion)",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							"sha256:000": {},
						},
					},
					"gcr.io/bar": {
						"a": {
							// "bar" already has it
							"sha256:000": {},
						},
					},
					"gcr.io/cat": {
						"c": {
							"sha256:222": {},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{
				reg.PromotionRequest{
					TagOp:          reg.Add,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[2].Name,
					ServiceAccount: registries[2].ServiceAccount,
					ImageNameSrc:   "a",
					ImageNameDest:  "a",
					Digest:         "sha256:000",
					Tag:            "",
				}: 1,
			},
			true,
		},
		{
			"NOP; dest has extra tag, but NOP because -delete-extra-tags NOT specified",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"0.9"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
					"gcr.io/bar": {
						"a": {
							"sha256:000": {"0.9", "extra-tag"},
						},
					},
					"gcr.io/cat": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{},
			true,
		},
		{
			"NOP (src registry does not have any of the images we want to promote)",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"missing-from-src"},
							"sha256:333": {"0.8"},
						},
					},
					{
						ImageName: "b",
						Dmap: reg.DigestTags{
							"sha256:bbb": {"also-missing"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"c": {
							"sha256:000": {"0.9"},
						},
						"d": {
							"sha256:bbb": {"1.0"},
						},
					},
					"gcr.io/bar": {
						"a": {
							"sha256:333": {"0.8"},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{},
			true,
		},
		{
			"Add 1 tag for 2 registries",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"0.9", "1.0"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
					"gcr.io/bar": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
					"gcr.io/cat": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{
				reg.PromotionRequest{
					TagOp:          reg.Add,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[0].Name,
					ServiceAccount: registries[0].ServiceAccount,
					ImageNameSrc:   "a",
					ImageNameDest:  "a",
					Digest:         "sha256:000",
					Tag:            "1.0",
				}: 1,
				reg.PromotionRequest{
					TagOp:          reg.Add,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[2].Name,
					ServiceAccount: registries[2].ServiceAccount,
					ImageNameSrc:   "a",
					ImageNameDest:  "a",
					Digest:         "sha256:000",
					Tag:            "1.0",
				}: 1,
			},
			true,
		},
		{
			"Add 1 tag for 1 registry",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"0.9", "1.0"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
					"gcr.io/bar": {
						"a": {
							"sha256:000": {"0.9"},
						},
					},
					"gcr.io/cat": {
						"a": {
							"sha256:000": {
								"0.9", "1.0", "extra-tag",
							},
						},
					},
				},
			},
			nil,
			reg.CapturedRequests{
				reg.PromotionRequest{
					TagOp:          reg.Add,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[0].Name,
					ServiceAccount: registries[0].ServiceAccount,
					ImageNameSrc:   "a",
					ImageNameDest:  "a",
					Digest:         "sha256:000",
					Tag:            "1.0",
				}: 1,
			},
			true,
		},
	}

	// captured is sort of a "global variable" because processRequestFake
	// closes over it.
	captured := make(reg.CapturedRequests)
	processRequestFake := reg.MkRequestCapturer(&captured)

	nopStream := func(
		srcRegistry reg.RegistryName,
		srcImageName reg.ImageName,
		rc reg.RegistryContext,
		destImageName reg.ImageName,
		digest reg.Digest,
		tag reg.Tag,
		tp reg.TagOp,
	) stream.Producer {
		// We don't even need a stream producer, because we are not creating
		// subprocesses that generate JSON or any other output; the vanilla
		// "mkReq" in Promote() already stores all the info we need to check.
		return nil
	}

	for _, test := range tests {
		// Reset captured for each test.
		captured = make(reg.CapturedRequests)
		srcReg, err := reg.GetSrcRegistry(registries)
		require.Nil(t, err)

		test.inputSc.SrcRegistry = srcReg

		// Simulate bad network conditions.
		if test.badReads != nil {
			for _, badRead := range test.badReads {
				test.inputSc.IgnoreFromPromotion(badRead)
			}
		}

		edges, err := reg.ToPromotionEdges([]reg.Manifest{test.inputM})
		require.Nil(t, err)

		filteredEdges, gotClean := test.inputSc.FilterPromotionEdges(
			edges,
			false)
		require.Equal(t, gotClean, test.expectedFilteredClean)

		require.Nil(t, test.inputSc.Promote(
			filteredEdges,
			nopStream,
			&processRequestFake,
		),
		)

		require.Equal(t, captured, test.expectedReqs)
	}
}

func TestExecRequests(t *testing.T) {
	sc := reg.SyncContext{}

	destRC := reg.RegistryContext{
		Name:           reg.RegistryName("gcr.io/bar"),
		ServiceAccount: "robot",
	}

	destRC2 := reg.RegistryContext{
		Name:           reg.RegistryName("gcr.io/cat"),
		ServiceAccount: "robot",
	}

	srcRC := reg.RegistryContext{
		Name:           reg.RegistryName("gcr.io/foo"),
		ServiceAccount: "robot",
		Src:            true,
	}

	registries := []reg.RegistryContext{destRC, srcRC, destRC2}

	nopStream := func(
		srcRegistry reg.RegistryName,
		srcImageName reg.ImageName,
		rc reg.RegistryContext,
		destImageName reg.ImageName,
		digest reg.Digest,
		tag reg.Tag,
		tp reg.TagOp,
	) stream.Producer {
		return nil
	}

	edges, err := reg.ToPromotionEdges(
		[]reg.Manifest{
			{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"0.9"},
						},
					},
				},
				SrcRegistry: &srcRC,
			},
		},
	)
	require.Nil(t, err)

	populateRequests := reg.MKPopulateRequestsForPromotionEdges(
		edges,
		nopStream,
	)

	var processRequestSuccess reg.ProcessRequest = func(
		sc *reg.SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- reg.RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex) {
		for req := range reqs {
			reqRes := reg.RequestResult{Context: req}
			requestResults <- reqRes
		}
	}

	var processRequestError reg.ProcessRequest = func(
		sc *reg.SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- reg.RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex) {
		for req := range reqs {
			reqRes := reg.RequestResult{Context: req}
			errors := make(reg.Errors, 0)
			errors = append(errors, reg.Error{
				Context: "Running TestExecRequests",
				Error:   fmt.Errorf("This request results in an error"),
			})
			reqRes.Errors = errors
			requestResults <- reqRes
		}
	}

	tests := []struct {
		name             string
		processRequestFn reg.ProcessRequest
		expected         error
	}{
		{
			"Error tracking for successful promotion",
			processRequestSuccess,
			nil,
		},
		{
			"Error tracking for promotion with errors",
			processRequestError,
			fmt.Errorf("Encountered an error while executing requests"),
		},
	}

	for _, test := range tests {
		got := sc.ExecRequests(populateRequests, test.processRequestFn)
		require.Equal(t, got, test.expected)
	}
}

func TestGarbageCollection(t *testing.T) {
	srcRegName := reg.RegistryName("gcr.io/foo")
	destRegName := reg.RegistryName("gcr.io/bar")
	destRegName2 := reg.RegistryName("gcr.io/cat")
	registries := []reg.RegistryContext{
		{
			Name:           srcRegName,
			ServiceAccount: "robot",
			Src:            true,
		},
		{
			Name:           destRegName,
			ServiceAccount: "robot",
		},
		{
			Name:           destRegName2,
			ServiceAccount: "robot",
		},
	}

	tests := []struct {
		name         string
		inputM       reg.Manifest
		inputSc      reg.SyncContext
		expectedReqs reg.CapturedRequests
	}{
		{
			"No garbage collection (no empty digests)",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"missing-from-src"},
							"sha256:333": {"0.8"},
						},
					},
					{
						ImageName: "b",
						Dmap: reg.DigestTags{
							"sha256:bbb": {"also-missing"},
						},
					},
				},
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"c": {
							"sha256:000": {"0.9"},
						},
						"d": {
							"sha256:bbb": {"1.0"},
						},
					},
					"gcr.io/bar": {
						"a": {
							"sha256:333": {"0.8"},
						},
					},
				},
			},
			reg.CapturedRequests{},
		},
		{
			"Simple garbage collection (delete ALL images in dest that are untagged))",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"missing-from-src"},
							"sha256:333": {"0.8"},
						},
					},
					{
						ImageName: "b",
						Dmap: reg.DigestTags{
							"sha256:bbb": {"also-missing"},
						},
					},
				},
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/foo": {
						"c": {
							"sha256:000": nil,
						},
						"d": {
							"sha256:bbb": nil,
						},
					},
					"gcr.io/bar": {
						"a": {
							// NOTE: this is skipping the first step where we
							// delete extra tags away (-delete-extra-tags).
							"sha256:111": nil,
						},
						"z": {
							"sha256:000": nil,
						},
					},
				},
			},
			reg.CapturedRequests{
				reg.PromotionRequest{
					TagOp:          reg.Delete,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[1].Name,
					ServiceAccount: registries[1].ServiceAccount,
					ImageNameSrc:   "",
					ImageNameDest:  "a",
					Digest:         "sha256:111",
					Tag:            "",
				}: 1,
				reg.PromotionRequest{
					TagOp:          reg.Delete,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[1].Name,
					ServiceAccount: registries[1].ServiceAccount,
					ImageNameSrc:   "",
					ImageNameDest:  "z",
					Digest:         "sha256:000",
					Tag:            "",
				}: 1,
			},
		},
	}

	captured := make(reg.CapturedRequests)

	var processRequestFake reg.ProcessRequest = func(
		sc *reg.SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- reg.RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex,
	) {
		for req := range reqs {
			// TODO: Why are we not checking errors here?
			// nolint: errcheck
			pr := req.RequestParams.(reg.PromotionRequest)

			mutex.Lock()
			captured[pr]++
			mutex.Unlock()
			requestResults <- reg.RequestResult{}
		}
	}

	for _, test := range tests {
		// Reset captured for each test.
		captured = make(reg.CapturedRequests)
		nopStream := func(
			destRC reg.RegistryContext,
			imageName reg.ImageName,
			digest reg.Digest) stream.Producer {
			return nil
		}
		srcReg, err := reg.GetSrcRegistry(registries)
		require.Nil(t, err)

		test.inputSc.SrcRegistry = srcReg
		test.inputSc.GarbageCollect(test.inputM, nopStream, &processRequestFake)

		require.Equal(t, captured, test.expectedReqs)
	}
}

func TestGarbageCollectionMulti(t *testing.T) {
	srcRegName := reg.RegistryName("gcr.io/src")
	destRegName1 := reg.RegistryName("gcr.io/dest1")
	destRegName2 := reg.RegistryName("gcr.io/dest2")

	destRC := reg.RegistryContext{
		Name:           destRegName1,
		ServiceAccount: "robotDest1",
	}

	destRC2 := reg.RegistryContext{
		Name:           destRegName2,
		ServiceAccount: "robotDest2",
	}

	srcRC := reg.RegistryContext{
		Name:           srcRegName,
		ServiceAccount: "robotSrc",
		Src:            true,
	}

	registries := []reg.RegistryContext{srcRC, destRC, destRC2}
	tests := []struct {
		name         string
		inputM       reg.Manifest
		inputSc      reg.SyncContext
		expectedReqs reg.CapturedRequests
	}{
		{
			"Simple garbage collection (delete ALL images in all dests that are untagged))",
			reg.Manifest{
				Registries: registries,
				Images: []reg.Image{
					{
						ImageName: "a",
						Dmap: reg.DigestTags{
							"sha256:000": {"missing-from-src"},
							"sha256:333": {"0.8"},
						},
					},
					{
						ImageName: "b",
						Dmap: reg.DigestTags{
							"sha256:bbb": {"also-missing"},
						},
					},
				},
			},
			reg.SyncContext{
				Inv: reg.MasterInventory{
					"gcr.io/src": {
						"c": {
							"sha256:000": nil,
						},
						"d": {
							"sha256:bbb": nil,
						},
					},
					"gcr.io/dest1": {
						"a": {
							"sha256:111": nil,
						},
						"z": {
							"sha256:222": nil,
						},
					},
					"gcr.io/dest2": {
						"a": {
							"sha256:123": nil,
						},
						"b": {
							"sha256:444": nil,
						},
					},
				},
			},
			reg.CapturedRequests{
				reg.PromotionRequest{
					TagOp:          reg.Delete,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[1].Name,
					ServiceAccount: registries[1].ServiceAccount,
					ImageNameSrc:   "",
					ImageNameDest:  "a",
					Digest:         "sha256:111",
					Tag:            "",
				}: 1,
				reg.PromotionRequest{
					TagOp:          reg.Delete,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[1].Name,
					ServiceAccount: registries[1].ServiceAccount,
					ImageNameSrc:   "",
					ImageNameDest:  "z",
					Digest:         "sha256:222",
					Tag:            "",
				}: 1,
				reg.PromotionRequest{
					TagOp:          reg.Delete,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[2].Name,
					ServiceAccount: registries[2].ServiceAccount,
					ImageNameSrc:   "",
					ImageNameDest:  "a",
					Digest:         "sha256:123",
					Tag:            "",
				}: 1,
				reg.PromotionRequest{
					TagOp:          reg.Delete,
					RegistrySrc:    srcRegName,
					RegistryDest:   registries[2].Name,
					ServiceAccount: registries[2].ServiceAccount,
					ImageNameSrc:   "",
					ImageNameDest:  "b",
					Digest:         "sha256:444",
					Tag:            "",
				}: 1,
			},
		},
	}

	captured := make(reg.CapturedRequests)

	var processRequestFake reg.ProcessRequest = func(
		sc *reg.SyncContext,
		reqs chan stream.ExternalRequest,
		requestResults chan<- reg.RequestResult,
		wg *sync.WaitGroup,
		mutex *sync.Mutex,
	) {
		for req := range reqs {
			// TODO: Why are we not checking errors here?
			// nolint: errcheck
			pr := req.RequestParams.(reg.PromotionRequest)
			mutex.Lock()
			captured[pr]++
			mutex.Unlock()
			requestResults <- reg.RequestResult{}
		}
	}

	for _, test := range tests {
		// Reset captured for each test.
		captured = make(reg.CapturedRequests)
		nopStream := func(
			destRC reg.RegistryContext,
			imageName reg.ImageName,
			digest reg.Digest,
		) stream.Producer {
			return nil
		}

		srcReg, err := reg.GetSrcRegistry(registries)
		require.Nil(t, err)

		test.inputSc.SrcRegistry = srcReg
		test.inputSc.GarbageCollect(test.inputM, nopStream, &processRequestFake)

		require.Equal(t, captured, test.expectedReqs)
	}
}

func TestSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		input    reg.RegInvImage
		expected string
	}{
		{
			"Basic",
			reg.RegInvImage{
				"foo": {
					"sha256:111": {"one"},
					"sha256:fff": {"0.9", "0.5"},
					"sha256:abc": {"0.3", "0.2"},
				},
				"bar": {
					"sha256:000": {"0.8", "0.5", "0.9"},
				},
			},
			`- name: bar
  dmap:
    "sha256:000": ["0.5", "0.8", "0.9"]
- name: foo
  dmap:
    "sha256:111": ["one"]
    "sha256:abc": ["0.2", "0.3"]
    "sha256:fff": ["0.5", "0.9"]
`,
		},
	}

	for _, test := range tests {
		gotYAML := test.input.ToYAML(reg.YamlMarshalingOpts{})
		require.YAMLEq(t, gotYAML, test.expected)
	}
}

func TestParseContainerParts(t *testing.T) {
	type ContainerParts struct {
		registry   string
		repository string
		err        error
	}

	shouldBeValid := []struct {
		input    string
		expected ContainerParts
	}{
		{
			"gcr.io/google-containers/foo",
			ContainerParts{
				"gcr.io/google-containers",
				"foo",
				nil,
			},
		},
		{
			"us.gcr.io/google-containers/foo",
			ContainerParts{
				"us.gcr.io/google-containers",
				"foo",
				nil,
			},
		},
		{
			"us.gcr.io/google-containers/foo/bar",
			ContainerParts{
				"us.gcr.io/google-containers",
				"foo/bar",
				nil,
			},
		},
		{
			"k8s.gcr.io/a/b/c",
			ContainerParts{
				"k8s.gcr.io",
				"a/b/c",
				nil,
			},
		},
	}

	for _, test := range shouldBeValid {
		registry, repository, err := reg.ParseContainerParts(test.input)
		got := ContainerParts{
			registry,
			repository,
			err,
		}

		require.Equal(t, got, test.expected)
	}

	shouldBeInvalid := []struct {
		input    string
		expected ContainerParts
	}{
		{
			// Blank string.
			"",
			ContainerParts{
				"",
				"",
				fmt.Errorf("invalid string '%s'", ""),
			},
		},
		{
			// Bare domain..
			"gcr.io",
			ContainerParts{
				"",
				"",
				fmt.Errorf("invalid string '%s'", "gcr.io"),
			},
		},
		{
			// Another top-level name (missing image name).
			"gcr.io/google-containers",
			ContainerParts{
				"",
				"",
				fmt.Errorf("invalid string '%s'", "gcr.io/google-containers"),
			},
		},
		{
			// Naked vanity domain (missing image name).
			"k8s.gcr.io",
			ContainerParts{
				"",
				"",
				fmt.Errorf("invalid string '%s'", "k8s.gcr.io"),
			},
		},
		{
			// Double slash.
			"k8s.gcr.io//a/b",
			ContainerParts{
				"",
				"",
				fmt.Errorf("invalid string '%s'", "k8s.gcr.io//a/b"),
			},
		},
		{
			// Trailing slash.
			"k8s.gcr.io/a/b/",
			ContainerParts{
				"",
				"",
				fmt.Errorf("invalid string '%s'", "k8s.gcr.io/a/b/"),
			},
		},
	}

	for _, test := range shouldBeInvalid {
		registry, repository, err := reg.ParseContainerParts(test.input)
		got := ContainerParts{
			registry,
			repository,
			err,
		}

		require.Equal(t, got, test.expected)
	}
}

func TestMatch(t *testing.T) {
	inputMfest := reg.Manifest{
		Registries: []reg.RegistryContext{
			{
				Name:           "gcr.io/foo-staging",
				ServiceAccount: "sa@robot.com",
				Src:            true,
			},
			{
				Name:           "us.gcr.io/some-prod",
				ServiceAccount: "sa@robot.com",
			},
			{
				Name:           "eu.gcr.io/some-prod",
				ServiceAccount: "sa@robot.com",
			},
			{
				Name:           "asia.gcr.io/some-prod",
				ServiceAccount: "sa@robot.com",
			},
		},
		Images: []reg.Image{
			{
				ImageName: "foo-controller",
				Dmap: reg.DigestTags{
					"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {"1.0"},
				},
			},
		},
		Filepath: "a/promoter-manifest.yaml",
	}

	tests := []struct {
		name          string
		mfest         reg.Manifest
		gcrPayload    reg.GCRPubSubPayload
		expectedMatch reg.GcrPayloadMatch
	}{
		{
			"INSERT message contains both Digest and Tag",
			inputMfest,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/some-prod/foo-controller@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				PQIN:   "us.gcr.io/some-prod/foo-controller:1.0",
			},
			reg.GcrPayloadMatch{
				PathMatch:   true,
				DigestMatch: true,
				TagMatch:    true,
			},
		},
		{
			"INSERT message only contains Digest",
			inputMfest,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/some-prod/foo-controller@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			reg.GcrPayloadMatch{
				PathMatch:   true,
				DigestMatch: true,
			},
		},
		{
			"INSERT's digest is not in Manifest (digest mismatch, but path matched)",
			inputMfest,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/some-prod/foo-controller@sha256:000",
			},
			reg.GcrPayloadMatch{
				PathMatch: true,
			},
		},
		{
			"INSERT's digest is not in Manifest (neither digest nor tag match, but path matched)",
			inputMfest,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/some-prod/foo-controller@sha256:000",
				PQIN:   "us.gcr.io/some-prod/foo-controller:1.0",
			},
			reg.GcrPayloadMatch{
				PathMatch: true,
			},
		},
		{
			"INSERT's digest is not in Manifest (tag specified, but tag mismatch)",
			inputMfest,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/some-prod/foo-controller@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				PQIN:   "us.gcr.io/some-prod/foo-controller:white-powder",
			},
			reg.GcrPayloadMatch{
				PathMatch:   true,
				DigestMatch: true,
				TagMismatch: true,
			},
		},
	}

	for _, test := range tests {
		require.Nil(t, test.gcrPayload.PopulateExtraFields())
		got := test.gcrPayload.Match(&test.mfest)

		require.Equal(t, got, test.expectedMatch)
	}
}

func TestPopulateExtraFields(t *testing.T) {
	shouldBeValid := []struct {
		name     string
		input    reg.GCRPubSubPayload
		expected reg.GCRPubSubPayload
	}{
		{
			"basic",
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "k8s.gcr.io/subproject/foo@sha256:000",
				PQIN:   "k8s.gcr.io/subproject/foo:1.0",
				Path:   "",
				Digest: "",
				Tag:    "",
			},
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "k8s.gcr.io/subproject/foo@sha256:000",
				PQIN:   "k8s.gcr.io/subproject/foo:1.0",
				Path:   "k8s.gcr.io/subproject/foo",
				Digest: "sha256:000",
				Tag:    "1.0",
			},
		},
		{
			"only FQIN",
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "k8s.gcr.io/subproject/foo@sha256:000",
				PQIN:   "",
				Path:   "",
				Digest: "",
				Tag:    "",
			},
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "k8s.gcr.io/subproject/foo@sha256:000",
				PQIN:   "",
				Path:   "k8s.gcr.io/subproject/foo",
				Digest: "sha256:000",
				Tag:    "",
			},
		},
		{
			"only PQIN",
			reg.GCRPubSubPayload{
				Action: "DELETE",
				FQIN:   "",
				PQIN:   "k8s.gcr.io/subproject/foo:1.0",
				Path:   "",
				Digest: "",
				Tag:    "",
			},
			reg.GCRPubSubPayload{
				Action: "DELETE",
				FQIN:   "",
				PQIN:   "k8s.gcr.io/subproject/foo:1.0",
				Path:   "k8s.gcr.io/subproject/foo",
				Digest: "",
				Tag:    "1.0",
			},
		},
	}

	for _, test := range shouldBeValid {
		require.Nil(t, test.input.PopulateExtraFields())

		got := test.input
		require.Equal(t, got, test.expected)
	}

	shouldBeInvalid := []struct {
		name     string
		input    reg.GCRPubSubPayload
		expected error
	}{
		{
			"FQIN missing @-sign",
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "k8s.gcr.io/subproject/foosha256:000",
				PQIN:   "k8s.gcr.io/subproject/foo:1.0",
				Path:   "",
				Digest: "",
				Tag:    "",
			},
			fmt.Errorf("invalid FQIN: k8s.gcr.io/subproject/foosha256:000"),
		},
		{
			"PQIN missing colon",
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "k8s.gcr.io/subproject/foo@sha256:000",
				PQIN:   "k8s.gcr.io/subproject/foo1.0",
				Path:   "",
				Digest: "",
				Tag:    "",
			},
			fmt.Errorf("invalid PQIN: k8s.gcr.io/subproject/foo1.0"),
		},
	}

	for _, test := range shouldBeInvalid {
		err := test.input.PopulateExtraFields()
		require.NotNil(t, err)
		require.Error(t, err, test.expected)
	}
}

// Helper functions.

func bazelTestPath(testName string, paths ...string) string {
	prefix := []string{
		os.Getenv("PWD"),
		"inventory_test",
		testName,
	}

	return filepath.Join(append(prefix, paths...)...)
}

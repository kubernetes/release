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

package inventory_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	reg "k8s.io/release/pkg/cip/dockerregistry"
)

func TestFindManifest(t *testing.T) {
	pwd := bazelTestPath("TestFindManifest")
	srcRC := reg.RegistryContext{
		Name:           "gcr.io/foo-staging",
		ServiceAccount: "sa@robot.com",
		Src:            true,
	}

	tests := []struct {
		// name is folder name
		name             string
		input            reg.GrowManifestOptions
		expectedManifest reg.Manifest
		expectedErr      error
	}{
		{
			"empty",
			reg.GrowManifestOptions{
				BaseDir:     filepath.Join(pwd, "empty"),
				StagingRepo: "gcr.io/foo",
			},
			reg.Manifest{},
			&os.PathError{
				Op:   "stat",
				Path: filepath.Join(pwd, "empty/images"),
				Err:  fmt.Errorf("no such file or directory"),
			},
		},
		{
			"singleton",
			reg.GrowManifestOptions{
				BaseDir:     filepath.Join(pwd, "singleton"),
				StagingRepo: "gcr.io/foo-staging",
			},
			reg.Manifest{
				Registries: []reg.RegistryContext{
					srcRC,
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
				Filepath: filepath.Join(pwd, "singleton/manifests/a/promoter-manifest.yaml"),
			},
			nil,
		},
		{
			"singleton (unrecognized staging repo)",
			reg.GrowManifestOptions{
				BaseDir:     filepath.Join(pwd, "singleton"),
				StagingRepo: "gcr.io/nonsense-staging",
			},
			reg.Manifest{},
			fmt.Errorf("could not find Manifest for %q", "gcr.io/nonsense-staging"),
		},
	}

	for _, test := range tests {
		gotManifest, gotErr := reg.FindManifest(&test.input)
		if test.expectedErr != nil {
			require.NotNil(t, gotErr)
			require.Error(t, gotErr, test.expectedErr.Error())
		} else {
			require.Nil(t, gotErr)
		}

		// Clean up gotManifest for purposes of comparing against expected
		// results. Namely, clear out the SrcRegistry pointer because this will
		// always be different.
		gotManifest.SrcRegistry = nil

		require.Equal(t, gotManifest, test.expectedManifest)
	}
}

func TestApplyFilters(t *testing.T) {
	tests := []struct {
		// name is folder name
		name         string
		inputOptions reg.GrowManifestOptions
		inputRii     reg.RegInvImage
		expectedRii  reg.RegInvImage
		expectedErr  error
	}{
		{
			"empty rii",
			reg.GrowManifestOptions{},
			reg.RegInvImage{},
			reg.RegInvImage{},
			nil,
		},
		{
			"no filters --- same as input",
			reg.GrowManifestOptions{},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"2.0"},
				},
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"2.0"},
				},
			},
			nil,
		},
		{
			"remove 'latest' tag by default, even if no filters",
			reg.GrowManifestOptions{},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"latest", "2.0"},
				},
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"2.0"},
				},
			},
			nil,
		},
		{
			"filter on image name only",
			reg.GrowManifestOptions{
				FilterImage: "bar",
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"latest", "2.0"},
				},
				"bar": {
					"sha256:111": {"latest", "1.0"},
				},
			},
			reg.RegInvImage{
				"bar": {
					"sha256:111": {"1.0"},
				},
			},
			nil,
		},
		{
			"filter on tag only",
			reg.GrowManifestOptions{
				FilterTag: "1.0",
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"latest", "2.0"},
				},
				"bar": {
					"sha256:111": {"latest", "1.0"},
				},
			},
			reg.RegInvImage{
				"bar": {
					"sha256:111": {"1.0"},
				},
			},
			nil,
		},
		{
			"filter on 'latest' tag",
			reg.GrowManifestOptions{
				FilterTag: "latest",
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"latest", "2.0"},
				},
				"bar": {
					"sha256:111": {"latest", "1.0"},
				},
			},
			reg.RegInvImage{},
			xerrors.New("no images survived filtering; double-check your --filter_* flag(s) for typos"),
		},
		{
			"filter on digest",
			reg.GrowManifestOptions{
				FilterDigest: "sha256:222",
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"latest", "2.0"},
					"sha256:222": {"3.0"},
				},
				"bar": {
					"sha256:111": {"latest", "1.0"},
				},
			},
			reg.RegInvImage{
				"foo": {
					"sha256:222": {"3.0"},
				},
			},
			nil,
		},
		{
			"filter on shared tag (multiple images share same tag)",
			reg.GrowManifestOptions{
				FilterTag: "1.2.3",
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"latest", "1.2.3"},
					"sha256:222": {"3.0"},
				},
				"bar": {
					"sha256:111": {"latest", "1.2.3"},
				},
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"1.2.3"},
				},
				"bar": {
					"sha256:111": {"1.2.3"},
				},
			},
			nil,
		},
		{
			"filter on shared tag and image name (multiple images share same tag)",
			reg.GrowManifestOptions{
				FilterImage: "foo",
				FilterTag:   "1.2.3",
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"latest", "1.2.3"},
					"sha256:222": {"3.0"},
				},
				"bar": {
					"sha256:111": {"latest", "1.2.3"},
				},
			},
			reg.RegInvImage{
				"foo": {
					"sha256:000": {"1.2.3"},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		gotRii, gotErr := reg.ApplyFilters(&test.inputOptions, test.inputRii)
		if test.expectedErr != nil {
			require.NotNil(t, gotErr)
			require.Error(t, gotErr, test.expectedErr.Error())
		} else {
			require.Nil(t, gotErr)
		}

		require.Equal(t, gotRii, test.expectedRii)
	}
}

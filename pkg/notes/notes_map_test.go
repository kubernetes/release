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

package notes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewProviderFromInitString(t *testing.T) {
	testCases := []struct {
		initString   string
		returnsError bool
	}{
		{initString: "maps/testdata/unit/", returnsError: false},
		{initString: "/this/shoud/not/really.exist/as/a/d33rect0ree", returnsError: true},
		{initString: "gs://bucket-name/map/path/", returnsError: true},
		{initString: "github://kubernetes/sig-release/maps", returnsError: true},
	}
	for _, testCase := range testCases {
		provider, err := NewProviderFromInitString(testCase.initString)
		if !testCase.returnsError {
			require.Nil(t, err)
			require.NotNil(t, provider)
		} else {
			require.NotNil(t, err)
		}
	}
}

func TestParseReleaseNotesMap(t *testing.T) {
	maps, err := ParseReleaseNotesMap("maps/testdata/unit/maps.yaml")
	require.Nil(t, err)
	require.GreaterOrEqual(t, 6, len(*maps))

	maps, err = ParseReleaseNotesMap("maps/testdata/fullmap.yaml")
	require.Nil(t, err)
	require.GreaterOrEqual(t, 4, len(*maps))
}

func TestGetMapsForPR(t *testing.T) {
	provider, err := NewProviderFromInitString("maps/testdata")
	require.Nil(t, err)

	maps, err := provider.GetMapsForPR(95000)
	require.Nil(t, err)
	require.GreaterOrEqual(t, 6, len(maps))

	maps, err = provider.GetMapsForPR(123)
	require.Nil(t, err)
	require.GreaterOrEqual(t, 4, len(maps))
}

func TestReleaseNotesMapIntegrity(t *testing.T) {
	maps, err := ParseReleaseNotesMap("maps/testdata/fullmap.yaml")
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(*maps), 1)

	// The first map in the test file contains a full map
	testMap := (*maps)[0]

	// Map metadata
	require.Equal(t, 123, testMap.PR)
	require.Equal(t, "1a89038915fe77d73bf7c9cfa8f2ce123a464c82", testMap.Commit)

	// Map release note elements. All are defined, so none should be nil
	require.NotNil(t, testMap.ReleaseNote.Text)
	require.Equal(t, "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", *testMap.ReleaseNote.Text)

	require.NotNil(t, *testMap.ReleaseNote.Author)
	require.Equal(t, "kubernetes-ci-robot", *testMap.ReleaseNote.Author)

	require.NotNil(t, *testMap.ReleaseNote.Areas)
	require.ElementsMatch(t, []string{"testarea"}, *testMap.ReleaseNote.Areas)

	require.NotNil(t, *testMap.ReleaseNote.Kinds)
	require.ElementsMatch(t, []string{"documentation"}, *testMap.ReleaseNote.Kinds)

	require.NotNil(t, *testMap.ReleaseNote.SIGs)
	require.ElementsMatch(t, []string{"api-machinery"}, *testMap.ReleaseNote.SIGs)

	require.NotNil(t, *testMap.ReleaseNote.Feature)
	require.Equal(t, true, *testMap.ReleaseNote.Feature)

	require.NotNil(t, *testMap.ReleaseNote.ActionRequired)
	require.Equal(t, false, *testMap.ReleaseNote.ActionRequired)
}

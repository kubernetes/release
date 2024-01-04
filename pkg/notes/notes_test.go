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
	"context"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/notes/options"

	kgithub "sigs.k8s.io/release-sdk/github"
)

const (
	mdSep            = "```"
	docsBlock        = mdSep + "docs"
	releaseNoteBlock = mdSep + "release-note"
)

func githubClient(t *testing.T) (kgithub.Client, context.Context) {
	_, tokenSet := os.LookupEnv(kgithub.TokenEnvKey)
	if !tokenSet {
		t.Skipf("%s environment variable is not set", kgithub.TokenEnvKey)
	}

	c := kgithub.New()
	return c.Client(), context.Background()
}

func TestStripActionRequired(t *testing.T) {
	notes := []string{
		"[action required] The note text",
		"[ACTION REQUIRED] The note text",
		"[AcTiOn ReQuIrEd] The note text",
	}

	for _, note := range notes {
		require.Equal(t, "The note text", stripActionRequired(note))
	}
}

func TestStripStar(t *testing.T) {
	notes := []string{
		"* The note text",
	}

	for _, note := range notes {
		require.Equal(t, "The note text", stripStar(note))
	}
}

func TestReleaseNoteParsing(t *testing.T) {
	c, ctx := githubClient(t)
	commitsWithNote := []string{
		"973dcd0c1a2555a6726aed8248ca816c9771253f",
		"27e5971c11cfcda703a39ed670a565f0f3564713",
	}
	gatherer := NewGathererWithClient(ctx, c)

	for _, sha := range commitsWithNote {
		fmt.Println(sha)
		commit, _, err := c.GetRepoCommit(ctx, "kubernetes", "kubernetes", sha)
		require.NoError(t, err)
		prs, err := gatherer.prsFromCommit(commit)
		require.NoError(t, err)
		_, err = gatherer.ReleaseNoteFromCommit(&Result{commit: commit, pullRequest: prs[0]})
		require.NoError(t, err)
	}
}

func TestDocumentationFromString(t *testing.T) {
	const (
		description1         = "Some description (#123), see :::"
		expectedDescription1 = "Some description (#123), see"
		url1                 = "http://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/kubectl-staging.md"
		description2         = "Another description -- "
		expectedDescription2 = "Another description"
		url2                 = "http://www.myurl.com/docs"
	)
	// multi line without prefix
	result := DocumentationFromString(
		fmt.Sprintf(docsBlock+"\r\n%s%s\r\n%s%s\r\n"+mdSep,
			description1, url1,
			description2, url2,
		),
	)
	require.Equal(t, 2, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)
	require.Equal(t, expectedDescription2, result[1].Description)
	require.Equal(t, url2, result[1].URL)

	// multi line without carriage return
	result = DocumentationFromString(
		fmt.Sprintf(docsBlock+"\n%s%s\n%s%s\n"+mdSep,
			description1, url1,
			description2, url2,
		),
	)
	require.Equal(t, 2, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)
	require.Equal(t, DocTypeKEP, result[0].Type)
	require.Equal(t, expectedDescription2, result[1].Description)
	require.Equal(t, url2, result[1].URL)
	require.Equal(t, DocTypeExternal, result[1].Type)

	// multi line with prefixes
	result = DocumentationFromString(
		fmt.Sprintf(docsBlock+"\r\n - %s%s\r\n * %s%s\r\n"+mdSep,
			description1, url1,
			description2, url2,
		),
	)
	require.Equal(t, 2, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)
	require.Equal(t, DocTypeKEP, result[0].Type)
	require.Equal(t, expectedDescription2, result[1].Description)
	require.Equal(t, url2, result[1].URL)
	require.Equal(t, DocTypeExternal, result[1].Type)

	// single line without star/dash prefix
	result = DocumentationFromString(
		fmt.Sprintf(docsBlock+"\r\n%s%s\r\n"+mdSep, description1, url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)

	// single line with star prefix
	result = DocumentationFromString(
		fmt.Sprintf(docsBlock+"\r\n * %s%s\r\n"+mdSep, description1, url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)

	// single line with dash prefix
	result = DocumentationFromString(
		fmt.Sprintf(docsBlock+"\r\n - %s%s\r\n"+mdSep, description1, url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)

	// single line without carriage return
	result = DocumentationFromString(
		fmt.Sprintf(docsBlock+"\n%s%s\n"+mdSep, description1, url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)

	// single line with empty description
	result = DocumentationFromString(
		fmt.Sprintf(docsBlock+"\n%s\n"+mdSep, url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, "", result[0].Description)
	require.Equal(t, url1, result[0].URL)
}

func TestClassifyURL(t *testing.T) {
	// A KEP
	u, err := url.Parse("http://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/kubectl-staging.md")
	require.Equal(t, err, nil)
	result := classifyURL(u)
	require.Equal(t, result, DocTypeKEP)

	// An official documentation
	u, err = url.Parse("https://kubernetes.io/docs/concepts/#kubernetes-objects")
	require.Equal(t, err, nil)
	result = classifyURL(u)
	require.Equal(t, result, DocTypeOfficial)

	// An external documentation
	u, err = url.Parse("https://google.com/")
	require.Equal(t, err, nil)
	result = classifyURL(u)
	require.Equal(t, result, DocTypeExternal)
}

func TestGetPRNumberFromCommitMessage(t *testing.T) {
	testCases := []struct {
		name             string
		commitMessage    string
		expectedPRNumber int
	}{
		{
			name: "Get PR number from merged PR",
			commitMessage: `Merge pull request #76030 from andrewsykim/e2e-legacyscheme

    test/e2e: replace legacy scheme with client-go scheme`,
			expectedPRNumber: 76030,
		},
		{
			name:             "Get PR number from squash merged PR",
			commitMessage:    "Add swapoff to centos so kubelet starts (#504)",
			expectedPRNumber: 504,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prs, err := prsNumForCommitFromMessage(tc.commitMessage)
			if err != nil {
				t.Fatalf("Expected no error to occur but got %v", err)
			}
			if prs[0] != tc.expectedPRNumber {
				t.Errorf("Expected PR number to be %d but was %d", tc.expectedPRNumber, prs[0])
			}
		})
	}
}

func TestPrettySIG(t *testing.T) {
	cases := map[string]string{
		"scheduling":        "Scheduling",
		"cluster-lifecycle": "Cluster Lifecycle",
		"cli":               "CLI",
		"aws":               "AWS",
		"api-machinery":     "API Machinery",
		"vsphere":           "vSphere",
		"openstack":         "OpenStack",
	}

	for input, expected := range cases {
		require.Equal(t, expected, (prettySIG(input)))
	}
}

func TestNoteTextFromString(t *testing.T) {
	noteBlock := func(note string) string {
		return releaseNoteBlock + "\n" + note + "\n" + mdSep
	}
	for _, tc := range []struct {
		input  string
		expect func(string, error)
	}{
		{
			noteBlock("test"),
			func(res string, err error) {
				require.Nil(t, err)
				require.Equal(t, "test", res)
			},
		},
		{
			noteBlock("test\ntest\ntest"),
			func(res string, err error) {
				require.Nil(t, err)
				require.Equal(t, "test\ntest\ntest", res)
			},
		},
		{
			noteBlock("Action Required: test"),
			func(res string, err error) {
				require.Nil(t, err)
				require.Equal(t, "test", res)
			},
		},
		{
			noteBlock(
				"- item\n  item\n  item",
			),
			func(res string, err error) {
				require.Nil(t, err)
				require.Equal(t, "item\nitem\nitem", res)
			},
		},
		{
			noteBlock(
				"- item\n  item\n- item\n  item",
			),
			func(res string, err error) {
				require.Nil(t, err)
				require.Equal(t, "item\nitem\n- item\n  item", res)
			},
		},
	} {
		tc.expect(noteTextFromString(tc.input))
	}
}

func TestMatchesExcludeFilter(t *testing.T) {
	for _, tc := range []struct {
		input         string
		shouldExclude bool
	}{
		{
			input:         "some test input",
			shouldExclude: false,
		},
		{
			input:         releaseNoteBlock + "\nnone\n" + mdSep,
			shouldExclude: true,
		},
		{
			input:         releaseNoteBlock + "\nn/a\n" + mdSep,
			shouldExclude: true,
		},
		{
			input:         releaseNoteBlock + "\nNA\n" + mdSep,
			shouldExclude: true,
		},
		{
			input:         releaseNoteBlock + "\nthis none should\n" + mdSep,
			shouldExclude: false,
		},
		{
			input: `@kubernetes/sig-auth-pr-reviews 
/milestone v1.19
/priority important-longterm
/kind cleanup
/kind deprecation

xref: #81126
xref: #81152
xref: https://github.com/kubernetes/website/pull/19630

` + mdSep + `release-note
Action Required: Support for basic authentication via the --basic-auth-file flag has been removed.  Users should migrate to --token-auth-file for similar functionality.
` + mdSep + `

` + mdSep + `docs
Removed "Static Password File" section from https://kubernetes.io/docs/reference/access-authn-authz/authentication/#static-password-file
https://github.com/kubernetes/website/pull/19630
` + mdSep,
			shouldExclude: false,
		},
	} {
		res := MatchesExcludeFilter(tc.input)
		require.Equal(t, tc.shouldExclude, res)
	}
}

func TestApplyMap(t *testing.T) {
	makeNewNote := func() ReleaseNote {
		return ReleaseNote{
			Commit:   "078b355da3cf56668ca1a8a5e36f2b3b52ff1bd8",
			Text:     "[ACTION REQUIRED] scheduler alpha metrics binding_duration_seconds and scheduling_algorithm_preemption_evaluation_seconds are deprecated, Both of those metrics are now covered as part of framework_extension_point_duration_seconds, the former as a PostFilter the latter and a Bind plugin. The plan is to remove both in 1.21",
			Markdown: "[ACTION REQUIRED] scheduler alpha metrics binding_duration_seconds and scheduling_algorithm_preemption_evaluation_seconds are deprecated, Both of those metrics are now covered as part of framework_extension_point_duration_seconds, the former as a PostFilter the latter and a Bind plugin. The plan is to remove both in 1.21",
			// Documentation:  documentation,
			Author:         "arghya88",
			AuthorURL:      "https://github.com/arghya88",
			PrURL:          "https://github.com/kubernetes/kubernetes/pull/95001",
			PrNumber:       95000,
			SIGs:           []string{"instrumentation", "scheduling"},
			Kinds:          []string{"deprecation", "feature"},
			Areas:          []string{},
			Feature:        true,
			Duplicate:      false,
			DuplicateKind:  false,
			ActionRequired: true,
		}
	}

	reflectedOriginalNote := reflect.ValueOf(makeNewNote())

	mp, err := NewProviderFromInitString("maps/testdata/unit/")
	require.Nil(t, err)

	// Read the maps from out test directory
	testMaps, err := mp.GetMapsForPR(95000)
	require.Nil(t, err)
	lastProp := ""

	for _, testMap := range testMaps {
		testNote := makeNewNote()

		// Check that the map application does note return error
		require.Nil(t, testNote.ApplyMap(testMap, false))

		reflectedNote := reflect.ValueOf(testNote)

		// Read the test case from the map datafields
		testcase, ok := testNote.DataFields["testcase"]
		require.True(t, ok, "found map test without testcase")

		// Read the property this test case checks
		property, ok := testcase.(map[interface{}]interface{})["property"].(string)
		require.True(t, ok)
		require.NotEmpty(t, property)
		require.NotEmpty(t, property, "testcase found without property")
		require.NotEqual(t, lastProp, property)
		lastProp = property

		// Grab the original value to ensure we're changing it
		originalVal := reflect.Indirect(reflectedOriginalNote).FieldByName(property)

		// Factor the test name
		testName, ok := testcase.(map[interface{}]interface{})["name"].(string)
		require.True(t, ok)

		switch expectedValue := testcase.(map[interface{}]interface{})["expected"].(type) {
		case bool:
			actualVal := reflect.Indirect(reflectedNote).FieldByName(property).Bool()
			require.Equalf(t, expectedValue, actualVal, "Failed %s", testName)
			require.NotEqual(t, expectedValue, originalVal.Bool(), "Failed %s", testName)
		case string:
			// Handle string test cases
			actualVal := reflect.Indirect(reflectedNote).FieldByName(property).String()
			require.Equalf(t, expectedValue, actualVal, "Failed %s", testName)
			require.NotEqualf(t, expectedValue, originalVal.String(), "Failed %s", testName)
		// Handle string slice cases
		case []interface{}:
			actualVal := reflect.Indirect(reflectedNote).FieldByName(property)
			actualSlice := make([]string, 0)
			for i := 0; i < actualVal.Len(); i++ {
				actualSlice = append(actualSlice, actualVal.Index(i).String())
			}
			require.ElementsMatchf(t, expectedValue, actualSlice, "Failed %s", testName)
		default:
			require.FailNowf(
				t, "Unknown case", "Unable to handle case for %s %T",
				property, testcase.(map[interface{}]interface{})["expected"],
			)
		}
	}
}

func TestDashify(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		input, expected string
	}{
		{ // modify list
			input:    "* test",
			expected: "- test",
		},
		{ // modify list
			input: `
				* list item
				  * sub list item
				  * sub list item
			`,
			expected: `
				- list item
				  - sub list item
				  - sub list item
			`,
		},
		{ // no substitution
			input: `
				This is some plain **bold** text.
				**And bold at the beginning of a line**.
			`,
			expected: `
				This is some plain **bold** text.
				**And bold at the beginning of a line**.
			`,
		},
	} {
		result := dashify(tc.input)
		require.Equal(t, tc.expected, result)
	}
}

func TestCapitalizeString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		input, expected string
	}{
		{ // capitalize English
			input:    "hello, world!",
			expected: "Hello, world!",
		},
		{ // capitalize Russian
			input:    "привет, мир!",
			expected: "Привет, мир!",
		},
		{ // no capitalization, English
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{ // no capitalization, Korean
			input:    "안녕, 세상아!",
			expected: "안녕, 세상아!",
		},
	} {
		result := capitalizeString(tc.input)
		require.Equal(t, tc.expected, result)
	}
}

func TestReleaseNoteForPullRequest(t *testing.T) {
	g, err := NewGatherer(context.Background(), &options.Options{
		GithubBaseURL: kgithub.GitHubURL,
		GithubOrg:     DefaultOrg,
		GithubRepo:    "release",
	})
	require.NoError(t, err)
	for _, tc := range []struct {
		name         string
		prNr         int
		expectedNote string
		notPublish   bool
		shouldErr    bool
	}{
		{
			name:         "Normal Release Note",
			prNr:         3378,
			expectedNote: "Fixed wrong amount of logger steps for `krel obs`.",
			notPublish:   false,
			shouldErr:    false,
		},
		{
			name:         "tagged release-note-none",
			prNr:         3398,
			expectedNote: "",
			notPublish:   true,
			shouldErr:    false,
		},
		{
			name:       "NONE release note",
			prNr:       3401,
			notPublish: true,
			shouldErr:  false,
		},
		{
			name:      "Not a PR",
			prNr:      3403,
			shouldErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			note, err := g.ReleaseNoteForPullRequest(tc.prNr)
			if tc.shouldErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, note)
			if tc.notPublish {
				require.True(t, note.DoNotPublish)
			}
			require.Equal(t, tc.expectedNote, note.Markdown)
		})
	}
}

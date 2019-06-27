package notes

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func githubClient(t *testing.T) *github.Client {
	token, tokenSet := os.LookupEnv("GITHUB_TOKEN")
	if !tokenSet {
		t.Skip("GITHUB_TOKEN is not set")
	}

	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	return github.NewClient(httpClient)
}

func TestConfigFromOpts(t *testing.T) {
	// fake config with an override for the org
	c := configFromOpts(
		WithOrg("marpaia"),
	)

	// test the override works
	require.Equal(t, "marpaia", c.org)

	// test the default value
	require.Equal(t, "kubernetes", c.repo)
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
	client := githubClient(t)
	commitsWithNote := []string{
		"973dcd0c1a2555a6726aed8248ca816c9771253f",
		"27e5971c11cfcda703a39ed670a565f0f3564713",
	}
	ctx := context.Background()

	for _, sha := range commitsWithNote {
		fmt.Println(sha)
		commit, _, err := client.Repositories.GetCommit(ctx, "kubernetes", "kubernetes", sha)
		require.NoError(t, err)
		_, err = ReleaseNoteFromCommit(commit, client, "0.1")
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
		fmt.Sprintf("```docs\r\n%s%s\r\n%s%s\r\n```",
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
		fmt.Sprintf("```docs\n%s%s\n%s%s\n```",
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
		fmt.Sprintf("```docs\r\n - %s%s\r\n * %s%s\r\n```",
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
		fmt.Sprintf("```docs\r\n%s%s\r\n```", description1, url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)

	// single line with star prefix
	result = DocumentationFromString(
		fmt.Sprintf("```docs\r\n * %s%s\r\n```", description1, url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)

	// single line with dash prefix
	result = DocumentationFromString(
		fmt.Sprintf("```docs\r\n - %s%s\r\n```", description1, url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)

	// single line without carriage return
	result = DocumentationFromString(
		fmt.Sprintf("```docs\n%s%s\n```", description1, url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, expectedDescription1, result[0].Description)
	require.Equal(t, url1, result[0].URL)

	// single line with empty description
	result = DocumentationFromString(
		fmt.Sprintf("```docs\n%s\n```", url1),
	)
	require.Equal(t, 1, len(result))
	require.Equal(t, "", result[0].Description)
	require.Equal(t, url1, result[0].URL)
}

func TestClassifyURL(t *testing.T) {
	// A KEP
	url, err := url.Parse("http://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/kubectl-staging.md")
	require.Equal(t, err, nil)
	result := classifyURL(url)
	require.Equal(t, result, DocTypeKEP)

	// An official documentation
	url, err = url.Parse("https://kubernetes.io/docs/concepts/#kubernetes-objects")
	require.Equal(t, err, nil)
	result = classifyURL(url)
	require.Equal(t, result, DocTypeOfficial)

	// An external documentation
	url, err = url.Parse("https://google.com/")
	require.Equal(t, err, nil)
	result = classifyURL(url)
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
			actualPRNumber, err := getPRNumberFromCommitMessage(tc.commitMessage)
			if err != nil {
				t.Fatalf("Expected no error to occur but got %v", err)
			}
			if actualPRNumber != tc.expectedPRNumber {
				t.Errorf("Expected PR number to be %d but was %d", tc.expectedPRNumber, actualPRNumber)
			}
		})
	}

}

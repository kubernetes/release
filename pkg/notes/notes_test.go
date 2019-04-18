package notes

import (
	"context"
	"fmt"
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

func TestNoteTextFromString(t *testing.T) {
	// multi line
	result, _ := NoteTextFromString("```release-note\r\ntest\r\ntest\r\n```")
	require.Equal(t, "test\ntest", result)

	// single line
	result, _ = NoteTextFromString("```release-note\r\ntest\r\n```")
	require.Equal(t, "test", result)

	// multi line, without carriage return
	result, _ = NoteTextFromString("```release-note\ntest\ntest\n```")
	require.Equal(t, "test\ntest", result)

	// single line, without carriage return
	result, _ = NoteTextFromString("```release-note\ntest\n```")
	require.Equal(t, "test", result)
}

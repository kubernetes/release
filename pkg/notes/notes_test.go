package notes

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

const (
	v1_10_0        = "fc32d2f3698e36b93322a3465f63a14e9f0eaead"
	v1_11_0_alpha1 = "4761788b2afa42a4573a6794902eb93fe666d5c5"
	v1_11_0_alpha2 = "ed9b25c90241b2b8a1fa10b96381c57f99ca952a"
	v1_11_0_beta1  = "4e3b2843df571c3b80c834d7c23bc6da1a22aab8"
	v1_11_0_beta2  = "be2cfcf9e44b5162a294e977329d6c8194748c4e"
	v1_11_0_rc1    = "8745ea56e3f1f3ad20050c1762eb6ba6f7786675"
	v1_11_0_rc2    = "d0a17cb4bbdf608559f257a76acfaa9acb054903"
	v1_11_0_rc3    = "931fc3b3aef9d679436978529fc7065d75352671"
	v1_11_0        = "91e7b4fd31fcd3d5f436da26c980becec37ceefe"
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

func TestGitHubAPIOperations(t *testing.T) {
	client := githubClient(t)

	// there were 48 commits between v1.11.0-rc3 and v1.11.0
	commits, err := ListCommits(client, v1_11_0_rc3, v1_11_0)
	require.NoError(t, err)
	require.Len(t, commits, 55)

	// there were 4 commits with release notes between v1.11.0-rc3 and v1.11.0
	commits, err = ListCommitsWithNotes(client, v1_11_0_rc3, v1_11_0)
	require.NoError(t, err)
	require.Len(t, commits, 4)

	for _, commit := range commits {
		// the string "release-note" must be in the commit message
		require.Contains(t, *commit.Commit.Message, "release-note")

		// each commit must have an associated PR
		pr, err := PRFromCommit(client, commit)
		require.NoError(t, err)

		// the PR must have labels
		require.NotEmpty(t, pr.Labels)

		// the commit must produce a release note
		note, err := ReleaseNoteFromCommit(commit, client)
		require.NoError(t, err)
		require.NotContains(t, note.Text, "\r")
	}
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

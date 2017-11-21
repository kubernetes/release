package util

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLastReleases(t *testing.T) {
	tables := []struct {
		owner       string
		repo        string
		lastRelease map[string]string
	}{
		// NOTE: Github webpage doesn't show correct release order. Use Github
		// API to get the releases.
		{"kubernetes", "kubernetes", map[string]string{
			"master":      "v1.8.0",
			"release-1.9": "v1.9.0-alpha.1",
			"release-1.8": "v1.8.0",
			"release-1.7": "v1.7.7",
			"release-1.6": "v1.6.11",
			"release-1.5": "v1.5.8",
		}},
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	c := NewClient(githubToken)

	for _, table := range tables {
		r, err := c.LastReleases(table.owner, table.repo)
		if err != nil {
			t.Errorf("%v %v: Unexpected error: %v", table.owner, table.repo, err)
		}
		for k, v := range table.lastRelease {
			if r[k] != v {
				t.Errorf("%v %v %v: Last release was incorrect, want: %v, got: %v",
					table.owner, table.repo, k, v, r[k])
			}
		}
	}
}

func TestListAllReleases(t *testing.T) {
	tables := []struct {
		owner       string
		repo        string
		numReleases int
	}{
		// NOTE: Github webpage doesn't show number of releases directly. Use Github
		// API to get the numbers.
		{"kubernetes", "kubernetes", 191},
		{"kubernetes", "helm", 30},
		{"kubernetes", "dashboard", 23},
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	c := NewClient(githubToken)

	for _, table := range tables {
		r, err := c.ListAllReleases(table.owner, table.repo)
		if err != nil {
			t.Errorf("%v %v: Unexpected error: %v", table.owner, table.repo, err)
		}
		if len(r) != table.numReleases {
			t.Errorf("%v %v: Number of releases was incorrect, want: %d, got: %d",
				table.owner, table.repo, table.numReleases, len(r))
		}
	}
}

func TestListAllTags(t *testing.T) {
	tables := []struct {
		owner   string
		repo    string
		numTags int
	}{
		{"kubernetes", "kubernetes", 295},
		{"kubernetes", "helm", 35},
		{"roycaihw", "kubernetes", 267},
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	c := NewClient(githubToken)

	for _, table := range tables {
		tags, err := c.ListAllTags(table.owner, table.repo)
		if err != nil {
			t.Errorf("%v %v: Unexpected error: %v", table.owner, table.repo, err)
		}
		if len(tags) != table.numTags {
			t.Errorf("%v %v: Number of tags was incorrect, want: %d, got: %d",
				table.owner, table.repo, table.numTags, len(tags))
		}
	}
}

func TestListAllIssues(t *testing.T) {
	tables := []struct {
		owner     string
		repo      string
		numIssues int
	}{
		// NOTE: including open and closed issues and PRs.
		{"kubernetes", "features", 164 + 67 + 1 + 253},
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	c := NewClient(githubToken)

	for _, table := range tables {
		i, err := c.ListAllIssues(table.owner, table.repo)
		if err != nil {
			t.Errorf("%v %v: Unexpected error: %v", table.owner, table.repo, err)
		}
		if len(i) != table.numIssues {
			t.Errorf("%v %v: Number of issues was incorrect, want: %d, got: %d",
				table.owner, table.repo, table.numIssues, len(i))
		}
	}
}

func TestListAllCommits(t *testing.T) {
	te, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", "2017-09-28 20:17:32 +0000 UTC")
	ts, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", "2017-03-30 20:44:26 +0000 UTC")

	tables := []struct {
		owner      string
		repo       string
		branch     string
		start      time.Time
		end        time.Time
		numCommits int
	}{
		{"kubernetes", "features", "58315cc33c51f8f4d05364d80f0b66f5d980bad7", time.Time{}, time.Time{}, 705},
		{"kubernetes", "features", "", time.Time{}, time.Time{}, 705},
		{"kubernetes", "helm", "release-v1.2.1", time.Time{}, time.Time{}, 373},
		{"kubernetes", "kubectl", "master", time.Time{}, time.Time{}, 9},
		{"kubernetes", "kubectl", "master", ts, te, 7},
	}
	githubToken := os.Getenv("GITHUB_TOKEN")
	c := NewClient(githubToken)

	for _, table := range tables {
		commits, err := c.ListAllCommits(table.owner, table.repo, table.branch, table.start, table.end)
		if err != nil {
			t.Errorf("%v %v %v: Unexpected error: %v", table.owner, table.repo, table.branch, err)
		}
		if len(commits) != table.numCommits {
			t.Errorf("%v %v %v: Number of commits was incorrect, want: %d, got: %d",
				table.owner, table.repo, table.branch, table.numCommits, len(commits))
		}
	}
}

func TestGetCommitDate(t *testing.T) {
	tables := []struct {
		owner     string
		repo      string
		tagCommit string
		date      string
		exist     bool
	}{
		{"kubernetes", "helm", "v2.6.0", "2017-08-16 18:56:09 +0000 UTC", true},
		{"kubernetes", "helm", "018ef2426f4932b2d8b9a772176acb548810a222", "2017-10-03 05:14:25 +0000 UTC", true},
		{"kubernetes", "helm", "018ef2426f4932b2d8b9a772176acb548810a221", "", false},
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	c := NewClient(githubToken)
	tags, _ := c.ListAllTags("kubernetes", "helm")

	for _, table := range tables {
		d, err := c.GetCommitDate(table.owner, table.repo, table.tagCommit, tags)
		var ok bool
		if err == nil {
			ok = true
		}
		if table.exist != ok {
			t.Errorf("%v: Existence check failed, want: %v, got: %v", table.tagCommit, table.exist, ok)
		}
		if table.exist && table.date != d.String() {
			t.Errorf("%v: Date was incorrect, want: %v, got: %v", table.tagCommit, table.date, d.String())
		}
	}
}

func TestAddQuerySearchIssues(t *testing.T) {
	tables := []struct {
		q   map[string]string
		num int
	}{
		{map[string]string{
			"repo": "kubernetes/kubernetes",
			"is":   "open",
			"type": "pr",
			"base": "release-1.7",
		}, 9},
		{map[string]string{
			"repo": "kubernetes/kubernetes",
			"is":   "open",
			"type": "pr",
			"base": "release-1.5",
		}, 2},
		{map[string]string{
			"repo":  "kubernetes/kubernetes",
			"type":  "pr",
			"label": "release-note",
		}, 3259},
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	c := NewClient(githubToken)

	for _, table := range tables {
		var query []string
		for k, v := range table.q {
			query = AddQuery(query, k, v)
		}
		result, err := c.SearchIssues(strings.Join(query, " "))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result) != table.num {
			t.Errorf("Result number was incorrect, want: %d, got %d", table.num, len(result))
		}
	}
}

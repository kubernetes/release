// Copyright 2017 The Kubernetes Authors All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

const (
	// GithubRawURL is the url prefix for getting raw github user content.
	GithubRawURL = "https://raw.githubusercontent.com/"
)

// GithubClient wraps github client with methods in this file.
type GithubClient struct {
	client *github.Client
	token  string
}

// ReadToken reads Github token from input file.
func ReadToken(filename string) (string, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(dat)), nil
}

// NewClient sets up a new github client with input assess token.
func NewClient(githubToken string) *GithubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GithubClient{github.NewClient(tc), githubToken}
}

// LastReleases looks up the list of releases on github and puts the last release per branch
// into a branch-indexed dictionary.
func (g GithubClient) LastReleases(owner, repo string) (map[string]string, error) {
	lastRelease := make(map[string]string)

	r, err := g.ListAllReleases(owner, repo)
	if err != nil {
		return nil, err
	}

	for _, release := range r {
		// Skip draft releases
		if *release.Draft {
			continue
		}
		// Alpha release goes only on master branch
		if strings.Contains(*release.TagName, "-alpha") && lastRelease["master"] == "" {
			lastRelease["master"] = *release.TagName
		} else {
			re, _ := regexp.Compile("v([0-9]+\\.[0-9]+)\\.([0-9]+(-.+)?)")
			version := re.FindStringSubmatch(*release.TagName)

			if version != nil {
				// Lastest vx.y.0 release goes on both master and release-vx.y branch
				if version[2] == "0" && lastRelease["master"] == "" {
					lastRelease["master"] = *release.TagName
				}

				branchName := "release-" + version[1]
				if lastRelease[branchName] == "" {
					lastRelease[branchName] = *release.TagName
				}
			}
		}
	}

	return lastRelease, nil
}

// ListAllReleases lists all releases for given owner and repo.
func (g GithubClient) ListAllReleases(owner, repo string) ([]*github.RepositoryRelease, error) {
	lo := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	releases, resp, err := g.client.Repositories.ListReleases(context.Background(), owner, repo, lo)
	if err != nil {
		return nil, err
	}
	lo.Page++

	for lo.Page <= resp.LastPage {
		re, _, err := g.client.Repositories.ListReleases(context.Background(), owner, repo, lo)
		if err != nil {
			return nil, err
		}
		for _, r := range re {
			releases = append(releases, r)
		}
		lo.Page++
	}
	return releases, nil
}

// ListAllIssues lists all issues and PRs for given owner and repo.
func (g GithubClient) ListAllIssues(owner, repo string) ([]*github.Issue, error) {
	// Because gathering all issues from large Github repo is time-consuming, we add a progress bar
	// rendering for more user-helpful output.
	log.Printf("Gathering all issues from Github for %s/%s. This may take a while...", owner, repo)

	start := time.Now().Round(time.Second)

	lo := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}
	ilo := &github.IssueListByRepoOptions{
		State:       "all",
		ListOptions: *lo,
	}

	issues, resp, err := g.client.Issues.ListByRepo(context.Background(), owner, repo, ilo)
	if err != nil {
		return nil, err
	}
	RenderProgressBar(ilo.ListOptions.Page, resp.LastPage, time.Now().Round(time.Second).Sub(start).String(), true)

	ilo.ListOptions.Page++

	for ilo.ListOptions.Page <= resp.LastPage {
		is, _, err := g.client.Issues.ListByRepo(context.Background(), owner, repo, ilo)
		if err != nil {
			// New line following the progress bar
			fmt.Print("\n")
			return nil, err
		}
		RenderProgressBar(ilo.ListOptions.Page, resp.LastPage, time.Now().Round(time.Second).Sub(start).String(), false)

		for _, i := range is {
			issues = append(issues, i)
		}
		ilo.ListOptions.Page++
	}
	// New line following the progress bar
	fmt.Print("\n")
	log.Print("All issues fetched.")
	return issues, nil
}

// ListAllTags lists all tags for given owner and repo.
func (g GithubClient) ListAllTags(owner, repo string) ([]*github.RepositoryTag, error) {
	lo := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	tags, resp, err := g.client.Repositories.ListTags(context.Background(), owner, repo, lo)
	if err != nil {
		return nil, err
	}
	lo.Page++

	for lo.Page <= resp.LastPage {
		ta, _, err := g.client.Repositories.ListTags(context.Background(), owner, repo, lo)
		if err != nil {
			return nil, err
		}
		for _, t := range ta {
			tags = append(tags, t)
		}
		lo.Page++
	}
	return tags, nil
}

// ListAllCommits lists all commits for given owner, repo, branch and time range.
func (g GithubClient) ListAllCommits(owner, repo, branch string, start, end time.Time) ([]*github.RepositoryCommit, error) {
	lo := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	clo := &github.CommitsListOptions{
		SHA:         branch,
		Since:       start,
		Until:       end,
		ListOptions: *lo,
	}

	commits, resp, err := g.client.Repositories.ListCommits(context.Background(), owner, repo, clo)
	if err != nil {
		return nil, err
	}
	clo.ListOptions.Page++

	for clo.ListOptions.Page <= resp.LastPage {
		co, _, err := g.client.Repositories.ListCommits(context.Background(), owner, repo, clo)
		if err != nil {
			return nil, err
		}
		for _, commit := range co {
			commits = append(commits, commit)
		}
		clo.ListOptions.Page++
	}
	return commits, nil
}

// GetCommitDate gets commit time for given tag/commit, provided with repository tags and commits.
// The function returns non-nil error if input tag/commit cannot be found in the repository.
func (g GithubClient) GetCommitDate(owner, repo, tagCommit string, tags []*github.RepositoryTag) (time.Time, error) {
	sha := tagCommit
	// If input string is a tag, convert it into SHA
	for _, t := range tags {
		if tagCommit == *t.Name {
			sha = *t.Commit.SHA
			break
		}
	}
	commit, _, err := g.client.Git.GetCommit(context.Background(), owner, repo, sha)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get commit date for SHA %s (original tag/commit %s): %v", sha, tagCommit, err)
	}
	return *commit.Committer.Date, nil
}

// HasLabel checks if input github issue contains input label.
func HasLabel(i *github.Issue, label string) bool {
	for _, l := range i.Labels {
		if *l.Name == label {
			return true
		}
	}

	return false
}

// SearchIssues gets all issues matching search query.
// NOTE: Github Search API has tight rate limit (30 requests per minute) and only returns the first 1,000 results.
// The function waits if it hits the rate limit, and reconstruct the search query with "created:<=YYYY-MM-DD" to
// search for issues out of the first 1,000 results.
func (g GithubClient) SearchIssues(query string) ([]github.Issue, error) {
	issues := make([]github.Issue, 0)
	issuesGot := make(map[int]bool)
	lastDateGot := ""
	totalIssueNumber := 0

	lo := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}
	// In order to search for >1,000 result, we use created time as a metric of searching progress. Therefore we
	// enforce "Sort:created" and "Order:desc".
	so := &github.SearchOptions{
		Sort:        "created",
		ListOptions: *lo,
	}

	for {
		r, _, err := g.client.Search.Issues(context.Background(), query, so)
		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				log.Printf("Hitting Github search API rate limit, sleeping for 30 seconds... error message: %v", err)
				time.Sleep(30 * time.Second)
				continue
			}
			return nil, err
		}
		totalIssueNumber = *r.Total
		break
	}

	for len(issues) < totalIssueNumber {
		q := query + lastDateGot
		// Get total number of pages in resp.LastPage
		result, resp, err := g.client.Search.Issues(context.Background(), q, so)
		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				log.Printf("Hitting Github search API rate limit, sleeping for 30 seconds... error message: %v", err)
				time.Sleep(30 * time.Second)
				continue
			}
			return nil, err
		}
		for _, i := range result.Issues {
			if issuesGot[*i.Number] == false {
				issues = append(issues, i)
				issuesGot[*i.Number] = true
				lastDateGot = fmt.Sprintf(" created:<=%s", i.CreatedAt.Format("2006-01-02"))
			}
		}
		so.ListOptions.Page++

		for so.ListOptions.Page <= resp.LastPage {
			result, _, err = g.client.Search.Issues(context.Background(), q, so)
			if err != nil {
				if _, ok := err.(*github.RateLimitError); ok {
					log.Printf("Hitting Github search API rate limit, sleeping for 30 seconds... error message: %v", err)
					time.Sleep(30 * time.Second)
					continue
				}
				return nil, err
			}
			for _, i := range result.Issues {
				if issuesGot[*i.Number] == false {
					issues = append(issues, i)
					issuesGot[*i.Number] = true
					lastDateGot = fmt.Sprintf(" created:<=%s", i.CreatedAt.Format("2006-01-02"))
				}
			}
			so.ListOptions.Page++
		}
		// Reset page number
		so.ListOptions.Page = 1
	}
	return issues, nil
}

// AddQuery forms a Github query by appending new query parts to input query
func AddQuery(query []string, queryParts ...string) []string {
	if len(queryParts) < 2 {
		log.Printf("not enough parts to form a query: %v", queryParts)
		return query
	}
	for _, part := range queryParts {
		if part == "" {
			return query
		}
	}

	return append(query, fmt.Sprintf("%s:%s", queryParts[0], strings.Join(queryParts[1:], "")))
}

// GetBranch is a wrapper of Github GetBranch function.
func (g GithubClient) GetBranch(ctx context.Context, owner, repo, branch string) (*github.Branch, *github.Response, error) {
	return g.client.Repositories.GetBranch(ctx, owner, repo, branch)
}

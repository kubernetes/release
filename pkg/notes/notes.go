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

package notes

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

// ReleaseNote is the type that represents the total sum of all the information
// we've gathered about a single release note.
type ReleaseNote struct {
	// Commit is the SHA of the commit which is the source of this note. This is
	// also effectively a unique ID for release notes.
	Commit string `json:"commit"`

	// Text is the actual content of the release note
	Text string `json:"text"`

	// Markdown is the markdown formatted note
	Markdown string `json:"markdown"`

	// Author is the GitHub username of the commit author
	Author string `json:"author"`

	// AuthorUrl is the GitHub URL of the commit author
	AuthorUrl string `json:"author_url"`

	// PrUrl is a URL to the PR
	PrUrl string `json:"pr_url"`

	// PrNumber is the number of the PR
	PrNumber int `json:"pr_number"`

	// Areas is a list of the labels beginning with area/
	Areas []string `json:"areas,omitempty"`

	// Kinds is a list of the labels beginning with kind/
	Kinds []string `json:"kinds,omitempty"`

	// SIGs is a list of the labels beginning with sig/
	SIGs []string `json:"sigs,omitempty"`

	// ActionRequired indicates whether or not the release-note-action-required
	// label was set on the PR
	ActionRequired bool `json:"action_required,omitempty"`
}

// githubApiOption is a type which allows for the expression of API configuration
// via the "functional option" pattern.
// For more information on this pattern, see the following blog post:
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type githubApiOption func(*githubApiConfig)

// githubApiConfig is a configuration struct that is used to express optional
// configuration for GitHub API requests
type githubApiConfig struct {
	ctx    context.Context
	org    string
	repo   string
	branch string
}

// WithContext allows the caller to inject a context into GitHub API requests
func WithContext(ctx context.Context) githubApiOption {
	return func(c *githubApiConfig) {
		c.ctx = ctx
	}
}

// WithOrg allows the caller to override the GitHub organization for the API
// request. By default, it is usually "kubernetes".
func WithOrg(org string) githubApiOption {
	return func(c *githubApiConfig) {
		c.org = org
	}
}

// WithRepo allows the caller to override the GitHub repo for the API
// request. By default, it is usually "kubernetes".
func WithRepo(repo string) githubApiOption {
	return func(c *githubApiConfig) {
		c.repo = repo
	}
}

// WithBranch allows the caller to override the repo branch for the API
// request. By default, it is usually "master".
func WithBranch(branch string) githubApiOption {
	return func(c *githubApiConfig) {
		c.branch = branch
	}
}

// ListReleaseNotes produces a list of fully contextualized release notes
// starting from a given commit SHA and ending at starting a given commit SHA.
func ListReleaseNotes(
	client *github.Client,
	logger log.Logger,
	start,
	end string,
	opts ...githubApiOption,
) ([]*ReleaseNote, error) {
	commits, err := ListCommitsWithNotes(client, logger, start, end, opts...)
	if err != nil {
		return nil, err
	}

	notes := []*ReleaseNote{}
	for _, commit := range commits {
		note, err := ReleaseNoteFromCommit(commit, client, opts...)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// NoteTextFromString returns the text of the release note given a string which
// may contain the commit message, the PR description, etc.
// This is generally the content inside the ```release-note ``` stanza.
func NoteTextFromString(s string) (string, error) {
	exp := regexp.MustCompile("```release-note\\r\\n(?P<note>.+)")
	match := exp.FindStringSubmatch(s)
	if len(match) == 0 {
		return "", errors.New("no matches found when parsing note text from commit string")
	}
	result := map[string]string{}
	for i, name := range exp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	note := strings.TrimRight(result["note"], "\r")
	note = stripActionRequired(note)
	note = stripStar(note)
	return note, nil
}

// ReleaseNoteFromCommit produces a full contextualized release note given a
// GitHub commit API resource.
func ReleaseNoteFromCommit(commit *github.RepositoryCommit, client *github.Client, opts ...githubApiOption) (*ReleaseNote, error) {
	pr, err := PRFromCommit(client, commit, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing release note from commit %s", commit.GetSHA())
	}

	text, err := NoteTextFromString(pr.GetBody())
	if err != nil {
		return nil, err
	}

	author := pr.GetUser().GetLogin()
	authorUrl := fmt.Sprintf("https://github.com/%s", author)
	prUrl := fmt.Sprintf("https://github.com/kubernetes/kubernetes/pull/%d", pr.GetNumber())
	markdown := fmt.Sprintf("%s ([#%d](%s), [@%s](%s))", text, pr.GetNumber(), prUrl, author, authorUrl)

	return &ReleaseNote{
		Commit:         commit.GetSHA(),
		Text:           text,
		Markdown:       markdown,
		Author:         author,
		AuthorUrl:      authorUrl,
		PrUrl:          prUrl,
		PrNumber:       pr.GetNumber(),
		SIGs:           LabelsWithPrefix(pr, "sig"),
		Kinds:          LabelsWithPrefix(pr, "kind"),
		Areas:          LabelsWithPrefix(pr, "area"),
		ActionRequired: IsActionRequired(pr),
	}, nil
}

// ListCommits lists all commits starting from a given commit SHA and ending at
// a given commit SHA.
func ListCommits(client *github.Client, start, end string, opts ...githubApiOption) ([]*github.RepositoryCommit, error) {
	c := configFromOpts(opts...)

	startCommit, _, err := client.Git.GetCommit(c.ctx, c.org, c.repo, start)
	if err != nil {
		return nil, err
	}

	endCommit, _, err := client.Git.GetCommit(c.ctx, c.org, c.repo, end)
	if err != nil {
		return nil, err
	}

	clo := &github.CommitsListOptions{
		SHA:   c.branch,
		Since: *startCommit.Committer.Date,
		Until: *endCommit.Committer.Date,
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	commits, resp, err := client.Repositories.ListCommits(c.ctx, c.org, c.repo, clo)
	if err != nil {
		return nil, err
	}
	clo.ListOptions.Page++

	for clo.ListOptions.Page <= resp.LastPage {
		commitPage, _, err := client.Repositories.ListCommits(c.ctx, c.org, c.repo, clo)
		if err != nil {
			return nil, err
		}
		for _, commit := range commitPage {
			commits = append(commits, commit)
		}
		clo.ListOptions.Page++
	}

	return commits, nil
}

// ListCommitsWithNotes list commits that have release notes starting from a
// given commit SHA and ending at a given commit SHA. This function is similar
// to ListCommits except that only commits with tagged release notes are
// returned.
func ListCommitsWithNotes(
	client *github.Client,
	logger log.Logger,
	start,
	end string,
	opts ...githubApiOption,
) ([]*github.RepositoryCommit, error) {
	commits, err := ListCommits(client, start, end, opts...)
	if err != nil {
		return nil, err
	}

	// exclusionFilters is a list of regular expressions that match commits that
	// do NOT contain release notes. Notably, this is all of the variations of
	// "release note none" that appear in the commit log.
	exclusionFilters := []string{
		"```release-note\\r\\nNONE",
		"```release-note\\r\\n\\s+NONE",
		"```release-note\\r\\nNONE",
		"```release-note\\r\\n\"NONE\"",
		"```release-note\\r\\nNone",
		"```release-note\\r\\nnone",
		"```release-note\\r\\nN/A",
		"```release-note\\r\\n\\r\\n```",
		"```release-note\\r\\n```",
	}

	// We "filter" the commits and "exclude" any commits that match the patterns
	// by setting the "include" parameter of this function to false
	commits, err = filterCommits(client, logger, commits, exclusionFilters, false, opts...)
	if err != nil {
		return nil, err
	}

	// Similarly, now that the known not-release-notes are filtered out, we can
	// use some patterns to find actual release notes.
	inclusionFilters := []string{
		"```release-note\\r\\n",
	}

	// We "filter" the commits and only include commits that match the patterns
	// by setting the "include" parameter of this function to true
	commits, err = filterCommits(client, logger, commits, inclusionFilters, true, opts...)
	if err != nil {
		return nil, err
	}

	// This final list contains all commits that contained a release notes stanza
	// and were not some variation of "none
	return commits, nil
}

// PRFromCommit return an API Pull Request struct given a commit struct. This is
// useful for going from a commit log to the PR (which contains useful info such
// as labels).
func PRFromCommit(client *github.Client, commit *github.RepositoryCommit, opts ...githubApiOption) (*github.PullRequest, error) {
	c := configFromOpts(opts...)

	// Thankfully k8s-merge-robot commits the PR number consistently. If this ever
	// stops being true, this definitely won't work anymore.
	exp := regexp.MustCompile(`Merge pull request #(?P<number>\d+)`)
	match := exp.FindStringSubmatch(*commit.Commit.Message)
	if len(match) == 0 {
		return nil, errors.New("no matches found when parsing PR from commit")
	}
	result := map[string]string{}
	for i, name := range exp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	number, err := strconv.Atoi(result["number"])
	if err != nil {
		return nil, err
	}

	// Given the PR number that we've now converted to an integer, get the PR from
	// the API
	pr, _, err := client.PullRequests.Get(c.ctx, c.org, c.repo, number)
	return pr, err
}

// LabelsWithPrefix is a helper for fetching all labels on a PR that start with
// a given string. This pattern is used often in the k/k repo and we can take
// advantage of this to contextualize release note generation with the kind, sig,
// area, etc labels.
func LabelsWithPrefix(pr *github.PullRequest, prefix string) []string {
	labels := []string{}
	for _, label := range pr.Labels {
		if strings.HasPrefix(*label.Name, prefix) {
			labels = append(labels, strings.TrimPrefix(*label.Name, prefix+"/"))
		}
	}
	return labels
}

// IsActionRequired indicates whether or not the release-note-action-required
// label was set on the PR.
func IsActionRequired(pr *github.PullRequest) bool {
	for _, label := range pr.Labels {
		if *label.Name == "release-note-action-required" {
			return true
		}
	}
	return false
}

// filterCommits is a helper that allows you to filter a set of commits by
// applying a set of regular expressions over the commit messages. If include is
// true, only commits that match at least one expression are returned. If include
// is false, only commits that match 0 of the expressions are returned.
func filterCommits(
	client *github.Client,
	logger log.Logger,
	commits []*github.RepositoryCommit,
	filters []string,
	include bool,
	opts ...githubApiOption,
) ([]*github.RepositoryCommit, error) {
	filteredCommits := []*github.RepositoryCommit{}
	for _, commit := range commits {
		body := commit.GetCommit().GetMessage()
		if commit.GetAuthor().GetLogin() == "k8s-merge-robot" {
			pr, err := PRFromCommit(client, commit, opts...)
			if err != nil {
				level.Info(logger).Log(
					"msg", "error getting PR from k8s-merge-robot commit",
					"err", err,
					"sha", commit.GetSHA(),
				)
				continue
			}
			body = pr.GetBody()
		}

		skip := false
		for _, filter := range filters {
			match, err := regexp.MatchString(filter, body)
			if err != nil {
				return nil, err
			}
			if match && !include || !match && include {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		filteredCommits = append(filteredCommits, commit)
	}

	return filteredCommits, nil
}

// configFromOpts is an internal helper for turning a set of functional options
// into a populated *githubApiConfig struct with consistent defaults.
func configFromOpts(opts ...githubApiOption) *githubApiConfig {
	c := &githubApiConfig{
		ctx:    context.Background(),
		org:    "kubernetes",
		repo:   "kubernetes",
		branch: "master",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func stripActionRequired(note string) string {
	expressions := []string{
		`(?i)\[action required\]\s`,
		`(?i)action required:\s`,
	}

	for _, exp := range expressions {
		re := regexp.MustCompile(exp)
		note = re.ReplaceAllString(note, "")
	}

	return note
}

func stripStar(note string) string {
	re := regexp.MustCompile(`(?i)\*\s`)
	return re.ReplaceAllString(note, "")
}

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

package githubutil

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
)

type ReleaseNote struct {
	Text           string
	Areas          []string
	Kinds          []string
	SIGs           []string
	ActionRequired bool
}

type githubApiOption func(*githubApiConfig)

type githubApiConfig struct {
	ctx    context.Context
	org    string
	repo   string
	branch string
}

func WithContext(ctx context.Context) githubApiOption {
	return func(c *githubApiConfig) {
		c.ctx = ctx
	}
}

func WithOrg(org string) githubApiOption {
	return func(c *githubApiConfig) {
		c.org = org
	}
}

func WithRepo(repo string) githubApiOption {
	return func(c *githubApiConfig) {
		c.repo = repo
	}
}

func WithBranch(branch string) githubApiOption {
	return func(c *githubApiConfig) {
		c.branch = branch
	}
}

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

func ListCommitsWithNotes(client *github.Client, start, end string, opts ...githubApiOption) ([]*github.RepositoryCommit, error) {
	commits, err := ListCommits(client, start, end, opts...)
	if err != nil {
		return nil, err
	}

	exclusionFilters := []string{
		"```release-note\\r\\nNONE",
		"```release-note\\r\\n\"NONE\"",
		"```release-note\\r\\nNone",
		"```release-note\\r\\nnone",
		"```release-note\\r\\nN/A",
		"```release-note\\r\\n\\r\\n```",
		"```release-note\\r\\n```",
	}

	commits, err = filterCommits(commits, exclusionFilters, false)
	if err != nil {
		return nil, err
	}

	inclusionFilters := []string{
		"```release-note\\r\\n",
	}
	commits, err = filterCommits(commits, inclusionFilters, true)
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func PRFromCommit(client *github.Client, commit *github.RepositoryCommit, opts ...githubApiOption) (*github.PullRequest, error) {
	c := configFromOpts(opts...)

	exp := regexp.MustCompile(`Merge pull request #(?P<number>\d+)`)
	match := exp.FindStringSubmatch(*commit.Commit.Message)
	if len(match) == 0 {
		return nil, errors.New("no matches found")
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

	pr, _, err := client.PullRequests.Get(c.ctx, c.org, c.repo, number)
	return pr, err
}

func NoteFromCommit(commit *github.RepositoryCommit) (string, error) {
	exp := regexp.MustCompile("```release-note\\r\\n(?P<note>.+)")
	match := exp.FindStringSubmatch(*commit.Commit.Message)
	if len(match) == 0 {
		return "", errors.New("no matches found")
	}
	result := map[string]string{}
	for i, name := range exp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result["note"], nil
}

func LabelsWithPrefix(pr *github.PullRequest, prefix string) []string {
	labels := []string{}
	for _, label := range pr.Labels {
		if strings.HasPrefix(*label.Name, prefix) {
			labels = append(labels, strings.TrimPrefix(*label.Name, prefix+"/"))
		}
	}
	return labels
}

func SIGsFromPR(pr *github.PullRequest) []string {
	sigs := []string{}
	for _, label := range pr.Labels {
		if strings.HasPrefix(*label.Name, "sig/") {
			sigs = append(sigs, strings.TrimPrefix(*label.Name, "sig/"))
		}
	}
	return sigs
}

func KindsFromPR(pr *github.PullRequest) []string {
	kinds := []string{}
	for _, label := range pr.Labels {
		if strings.HasPrefix(*label.Name, "kind/") {
			kinds = append(kinds, strings.TrimPrefix(*label.Name, "kind/"))
		}
	}
	return kinds
}

func IsActionRequired(pr *github.PullRequest) bool {
	for _, label := range pr.Labels {
		if *label.Name == "release-note-action-required" {
			return true
		}
	}
	return false
}

func ListReleaseNotes(client *github.Client, start, end string, opts ...githubApiOption) ([]*ReleaseNote, error) {
	commits, err := ListCommitsWithNotes(client, start, end, opts...)
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

func ReleaseNoteFromCommit(commit *github.RepositoryCommit, client *github.Client, opts ...githubApiOption) (*ReleaseNote, error) {
	pr, err := PRFromCommit(client, commit, opts...)
	if err != nil {
		return nil, err
	}

	text, err := NoteFromCommit(commit)
	if err != nil {
		return nil, err
	}

	return &ReleaseNote{
		Text:           text,
		SIGs:           LabelsWithPrefix(pr, "sig"),
		Kinds:          LabelsWithPrefix(pr, "kind"),
		Areas:          LabelsWithPrefix(pr, "area"),
		ActionRequired: IsActionRequired(pr),
	}, nil
}

func filterCommits(commits []*github.RepositoryCommit, filters []string, include bool) ([]*github.RepositoryCommit, error) {
	filteredCommits := []*github.RepositoryCommit{}
	for _, commit := range commits {
		skip := false
		for _, filter := range filters {
			match, err := regexp.MatchString(filter, *commit.Commit.Message)
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

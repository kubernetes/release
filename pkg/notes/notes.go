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
	"bufio"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v27/github"
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

	// Docs is additional documentation for the release note
	Documentation []*Documentation `json:"documentation,omitempty"`

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

	// Indicates whether or not a note will appear as a new feature
	Feature bool `json:"feature,omitempty"`

	// Indicates whether or not a note is duplicated across SIGs
	Duplicate bool `json:"duplicate,omitempty"`

	// ActionRequired indicates whether or not the release-note-action-required
	// label was set on the PR
	ActionRequired bool `json:"action_required,omitempty"`

	// Tags each note with a release version if specified
	// If not specified, omitted
	ReleaseVersion string `json:"release_version,omitempty"`
}

type Documentation struct {
	// A description about the documentation
	Description string `json:"description,omitempty"`

	// The url to be linked
	URL string `json:"url"`

	// Clssifies the link as something special, like a KEP
	Type DocType `json:"type"`
}

type DocType string

const (
	DocTypeExternal DocType = "external"
	DocTypeKEP      DocType = "KEP"
	DocTypeOfficial DocType = "official"
)

// ReleaseNoteList is a map of PR numbers referencing notes.
// To avoid needless loops, we need to be able to reference things by PR
// When we have to merge old and new entries, we want to be able to override
// the old entries with the new ones efficiently.
type ReleaseNoteList map[int]*ReleaseNote

// GithubApiOption is a type which allows for the expression of API configuration
// via the "functional option" pattern.
// For more information on this pattern, see the following blog post:
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type GithubApiOption func(*githubApiConfig)

// githubApiConfig is a configuration struct that is used to express optional
// configuration for GitHub API requests
type githubApiConfig struct {
	ctx    context.Context
	org    string
	repo   string
	branch string
}

// WithContext allows the caller to inject a context into GitHub API requests
func WithContext(ctx context.Context) GithubApiOption {
	return func(c *githubApiConfig) {
		c.ctx = ctx
	}
}

// WithOrg allows the caller to override the GitHub organization for the API
// request. By default, it is usually "kubernetes".
func WithOrg(org string) GithubApiOption {
	return func(c *githubApiConfig) {
		c.org = org
	}
}

// WithRepo allows the caller to override the GitHub repo for the API
// request. By default, it is usually "kubernetes".
func WithRepo(repo string) GithubApiOption {
	return func(c *githubApiConfig) {
		c.repo = repo
	}
}

// WithBranch allows the caller to override the repo branch for the API
// request. By default, it is usually "master".
func WithBranch(branch string) GithubApiOption {
	return func(c *githubApiConfig) {
		c.branch = branch
	}
}

// ListReleaseNotes produces a list of fully contextualized release notes
// starting from a given commit SHA and ending at starting a given commit SHA.
func ListReleaseNotes(
	client *github.Client,
	logger log.Logger,
	branch,
	start,
	end,
	requiredAuthor,
	relVer string,
	opts ...GithubApiOption,
) (ReleaseNoteList, error) {
	commits, err := ListCommitsWithNotes(client, logger, branch, start, end, opts...)
	if err != nil {
		return nil, err
	}

	dedupeCache := map[string]struct{}{}
	notes := make(ReleaseNoteList)
	for _, commit := range commits {

		if requiredAuthor != "" {
			if commit.GetAuthor().GetLogin() != requiredAuthor {
				continue
			}
		}

		note, err := ReleaseNoteFromCommit(commit, client, relVer, opts...)
		if err != nil {
			level.Error(logger).Log(
				"err", err,
				"msg", "error getting the release note from commit while listing release notes",
				"sha", commit.GetSHA(),
			)
			continue
		}

		// exclusionFilters is a list of regular expressions that match notes text that
		// are deemed to have no content and should NOT be added to release notes.
		exclusionFilters := []string{
			"^(?i)(none|n/a)$", // 'none' or 'n/a' case insensitive
		}
		excluded := false
		for _, filter := range exclusionFilters {
			match, err := regexp.MatchString(filter, strings.ToUpper(strings.TrimSpace(note.Text)))
			if err != nil {
				return nil, err
			}
			if match {
				excluded = true
				level.Debug(logger).Log(
					"msg", "Excluding notes that are deemed to have no content based on filter, and should NOT be added to release notes.",
					"sha", commit.GetSHA(),
					"func", "ListReleaseNotes",
					"filter", filter,
				)
				break
			}
		}
		if excluded {
			continue
		}

		if _, ok := dedupeCache[note.Text]; !ok {
			notes[note.PrNumber] = note
			dedupeCache[note.Text] = struct{}{}
		}
	}

	return notes, nil
}

// NoteTextFromString returns the text of the release note given a string which
// may contain the commit message, the PR description, etc.
// This is generally the content inside the ```release-note ``` stanza.
func NoteTextFromString(s string) (string, error) {
	exps := []*regexp.Regexp{
		// (?s) is needed for '.' to be matching on newlines, by default that's disabled
		// we need to match ungreedy 'U', because after the notes a `docs` block can occur
		regexp.MustCompile("(?sU)```release-note\\r\\n(?P<note>.+)\\r\\n```"),
		regexp.MustCompile("(?sU)```dev-release-note\\r\\n(?P<note>.+)"),
		regexp.MustCompile("(?sU)```\\r\\n(?P<note>.+)\\r\\n```"),
		regexp.MustCompile("(?sU)```release-note\n(?P<note>.+)\n```"),
	}

	for _, exp := range exps {
		match := exp.FindStringSubmatch(s)
		if len(match) == 0 {
			continue
		}
		result := map[string]string{}
		for i, name := range exp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		note := strings.ReplaceAll(result["note"], "#", "&#35;")
		note = strings.ReplaceAll(note, "\r", "")
		note = stripActionRequired(note)
		note = dashify(note)
		note = strings.TrimSpace(note)
		return note, nil
	}

	return "", errors.New("no matches found when parsing note text from commit string")
}

func DocumentationFromString(s string) []*Documentation {
	regex := regexp.MustCompile("(?s)```docs[\\r]?\\n(?P<text>.+)[\\r]?\\n```")
	match := regex.FindStringSubmatch(s)

	if len(match) < 1 {
		// Nothing found, but we don't require it
		return nil
	}

	result := []*Documentation{}
	text := match[1]
	text = stripStar(text)
	text = stripDash(text)

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		const httpPrefix = "http"
		s := strings.SplitN(scanner.Text(), httpPrefix, 2)
		if len(s) != 2 {
			continue
		}
		description := strings.TrimRight(strings.TrimSpace(s[0]), " :-")
		urlString := httpPrefix + strings.TrimSpace(s[1])

		// Validate the URL
		parsedURL, err := url.Parse(urlString)
		if err != nil {
			continue
		}

		result = append(result, &Documentation{
			Description: description,
			URL:         urlString,
			Type:        classifyURL(parsedURL),
		})
	}

	return result
}

// classifyURL returns the correct DocType for the given url
func classifyURL(url *url.URL) DocType {
	// Kubernetes Enhancement Proposals (KEPs)
	if strings.Contains(url.Host, "github.com") &&
		strings.Contains(url.Path, "/kubernetes/enhancements/") {
		return DocTypeKEP
	}

	// Official documentation
	if strings.Contains(url.Host, "kubernetes.io") &&
		strings.Contains(url.Path, "/docs/") {
		return DocTypeOfficial
	}

	return DocTypeExternal
}

// ReleaseNoteFromCommit produces a full contextualized release note given a
// GitHub commit API resource.
func ReleaseNoteFromCommit(commit *github.RepositoryCommit, client *github.Client, relVer string, opts ...GithubApiOption) (*ReleaseNote, error) {
	c := configFromOpts(opts...)

	pr, err := PRFromCommit(client, commit, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing release note from commit %s", commit.GetSHA())
	}

	prBody := pr.GetBody()
	text, err := NoteTextFromString(prBody)
	if err != nil {
		return nil, err
	}
	documentation := DocumentationFromString(prBody)

	author := pr.GetUser().GetLogin()
	authorUrl := fmt.Sprintf("https://github.com/%s", author)
	prUrl := fmt.Sprintf("https://github.com/%s/%s/pull/%d", c.org, c.repo, pr.GetNumber())
	IsFeature := HasString(LabelsWithPrefix(pr, "kind"), "feature")
	IsDuplicate := false
	sigsListPretty := prettifySigList(LabelsWithPrefix(pr, "sig"))
	noteSuffix := ""

	if IsActionRequired(pr) || IsFeature {
		if sigsListPretty != "" {
			noteSuffix = fmt.Sprintf("Courtesy of %s", sigsListPretty)
		}
	} else if len(LabelsWithPrefix(pr, "sig")) > 1 {
		IsDuplicate = true
	}

	indented := strings.ReplaceAll(text, "\n", "\n  ")
	markdown := fmt.Sprintf("%s ([#%d](%s), [@%s](%s))",
		indented, pr.GetNumber(), prUrl, author, authorUrl)

	if noteSuffix != "" {
		markdown = fmt.Sprintf("%s\n\n  %s", markdown, noteSuffix)
	}

	return &ReleaseNote{
		Commit:         commit.GetSHA(),
		Text:           text,
		Markdown:       markdown,
		Documentation:  documentation,
		Author:         author,
		AuthorUrl:      authorUrl,
		PrUrl:          prUrl,
		PrNumber:       pr.GetNumber(),
		SIGs:           LabelsWithPrefix(pr, "sig"),
		Kinds:          LabelsWithPrefix(pr, "kind"),
		Areas:          LabelsWithPrefix(pr, "area"),
		Feature:        IsFeature,
		Duplicate:      IsDuplicate,
		ActionRequired: IsActionRequired(pr),
		ReleaseVersion: relVer,
	}, nil
}

// ListCommits lists all commits starting from a given commit SHA and ending at
// a given commit SHA.
func ListCommits(client *github.Client, branch, start, end string, opts ...GithubApiOption) ([]*github.RepositoryCommit, error) {
	c := configFromOpts(opts...)

	c.branch = branch

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
		commits = append(commits, commitPage...)
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
	branch,
	start,
	end string,
	opts ...GithubApiOption,
) ([]*github.RepositoryCommit, error) {
	filteredCommits := []*github.RepositoryCommit{}

	commits, err := ListCommits(client, branch, start, end, opts...)
	if err != nil {
		return nil, err
	}

	for i, commit := range commits {

		level.Debug(logger).Log("msg", "################################################")
		level.Info(logger).Log("msg", fmt.Sprintf("[%d/%d - %0.2f%%]", i+1, len(commits), (float64(i+1)/float64(len(commits)))*100.0 ))
		level.Debug(logger).Log(
			"msg", "Processing commit",
			"func", "ListCommitsWithNotes",
			"sha", commit.GetSHA(),
		)

		pr, err := PRFromCommit(client, commit, opts...)
		if err != nil {
			if err.Error() == "no matches found when parsing PR from commit" {
				level.Debug(logger).Log(
					"msg", fmt.Sprintf("No matches found when parsing PR from commit sha '%s'.", commit.GetSHA()),
					"func", "ListCommitsWithNotes",
				)
				continue
			}
		}

		level.Debug(logger).Log(
			"msg", fmt.Sprintf("Obtaining PR associated with commit sha '%s'.", commit.GetSHA()),
			"func", "ListCommitsWithNotes",
			"pr no", pr.GetNumber(),
			"pr body", pr.GetBody(),
		)

		// exclusionFilters is a list of regular expressions that match commits that
		// do NOT contain release notes. Notably, this is all of the variations of
		// "release note none" that appear in the commit log.
		exclusionFilters := []string{

			// 'none','n/a','na' case insensitive with optional trailing
			// whitespace, wrapped in ``` with/without release-note identifier
			// the 'none','n/a','na' can also optionally be wrapped in quotes ' or "
			"(?i)```(release-note[s]?\\s*)?('|\")?(none|n/a|na)?('|\")?\\s*```",

			// This filter is too aggressive within the PR body and picks up matches unrelated to release notes
			// 'none' or 'n/a' case insensitive wrapped optionally with whitespace
			// "(?i)\\s*(none|n/a)\\s*",

			// simple '/release-note-none' tag
			"/release-note-none",
		}

		excluded := false

		for _, filter := range exclusionFilters {
			match, err := regexp.MatchString(filter, pr.GetBody())
			if err != nil {
				return nil, err
			}
			if match {
				excluded = true
				level.Debug(logger).Log(
					"msg", "Excluding notes for PR based on the exclusion filter.",
					"func", "ListCommitsWithNotes",
					"filter", filter,
				)
				break
			}
		}

		if excluded {
			continue
		}

		// Similarly, now that the known not-release-notes are filtered out, we can
		// use some patterns to find actual release notes.
		inclusionFilters := []string{
			"release-note",
			"Does this PR introduce a user-facing change?",
		}

		for _, filter := range inclusionFilters {
			match, err := regexp.MatchString(filter, pr.GetBody())
			if err != nil {
				return nil, err
			}
			if match {
				filteredCommits = append(filteredCommits, commit)
				level.Debug(logger).Log(
					"msg", "Including notes for PR based on the inclusion filter.",
					"func", "ListCommitsWithNotes",
					"filter", filter,
				)
			}
		}
	}

	return filteredCommits, nil
}

// PRFromCommit return an API Pull Request struct given a commit struct. This is
// useful for going from a commit log to the PR (which contains useful info such
// as labels).
func PRFromCommit(client *github.Client, commit *github.RepositoryCommit, opts ...GithubApiOption) (*github.PullRequest, error) {
	c := configFromOpts(opts...)

	number, err := getPRNumberFromCommitMessage(*commit.Commit.Message)
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
	opts ...GithubApiOption,
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
func configFromOpts(opts ...GithubApiOption) *githubApiConfig {
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

func stripDash(note string) string {
	re := regexp.MustCompile(`(?i)\-\s`)
	return re.ReplaceAllString(note, "")
}

func dashify(note string) string {
	return strings.ReplaceAll(note, "* ", "- ")
}

func HasString(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func getPRNumberFromCommitMessage(commitMessage string) (int, error) {
	// Thankfully k8s-merge-robot commits the PR number consistently. If this ever
	// stops being true, this definitely won't work anymore.
	exp := regexp.MustCompile(`automated-cherry-pick-of-#(?P<number>\d+)`)
	match := exp.FindStringSubmatch(commitMessage)
	if len(match) == 0 {
		exp = regexp.MustCompile(`Merge pull request #(?P<number>\d+)`)
		match = exp.FindStringSubmatch(commitMessage)
		if len(match) == 0 {
			// If the PR was squash merged, the regexp is different
			match = regexp.MustCompile(`\(#(?P<number>\d+)\)`).FindStringSubmatch(commitMessage)
			if len(match) == 0 {
				return 0, errors.New("no matches found when parsing PR from commit")
			}
		}
	}
	result := map[string]string{}
	for i, name := range exp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	number, err := strconv.Atoi(result["number"])
	if err != nil {
		return 0, err
	}
	return number, nil
}

/*
Copyright 2019 The Kubernetes Authors.

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
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/github/githubfakes"
)

func TestMain(m *testing.M) {
	// logrus, shut up
	logrus.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func TestListCommits(t *testing.T) {
	const always = -1

	zeroTime := &time.Time{}

	type listCommitsReturnsList map[int]struct {
		rc []*github.RepositoryCommit
		r  *github.Response
		e  error
	}
	type getCommitReturnsList map[int]struct {
		c *github.Commit
		r *github.Response
		e error
	}

	tests := map[string]struct {
		// branch, start, end are the args to call the `ListCommits` method with
		branch, start, end string

		// getCommitReturns is a list of mock returns for a call to `GetCommit`,
		// per method call count
		getCommitReturns getCommitReturnsList
		// getCommitArgValidator is a function that gets called for all every
		// `GetCommit` method with the original calls arguments
		getCommitArgValidator func(*testing.T, int, context.Context, string, string, string)
		// listCommitsReturns is a list of mock returns for a call to
		// `ListCommits`, per method call count
		listCommitsReturns listCommitsReturnsList
		// listCommitsArgValidator is a function that gets called for all every
		// `ListCommits` method with the original calls arguments
		listCommitsArgValidator func(*testing.T, int, context.Context, string, string, *github.CommitsListOptions)

		// expectedErrMsg is the error message we expect from the `ListCommits`
		// call
		expectedErrMsg string
		// expectedGetCommitCallCount is the number of calls to the `GetCommit`
		// method we expect
		expectedGetCommitCallCount int
		// expectedListCommitsMinCallCount, expectedListCommitsMaxCallCount is the number
		// of calls to the `ListCommits` method call we expect at least or at most
		expectedListCommitsMinCallCount int
		expectedListCommitsMaxCallCount int
		// expectedCommitCount is the number of commits we expect the `ListCommits`
		// method should return
		expectedCommitCount int
	}{
		"happy path": {
			branch: "the-branch", start: "the-start", end: "the-end",
			getCommitReturns: getCommitReturnsList{always: {
				c: &github.Commit{Committer: &github.CommitAuthor{Date: zeroTime}},
			}},
			listCommitsReturns: listCommitsReturnsList{always: {
				rc: []*github.RepositoryCommit{{}, {}}, // we create 2 commits per page
				r:  response(200, 100),
			}},
			getCommitArgValidator: func(t *testing.T, callCount int, ctx context.Context, org, repo, rev string) {
				checkOrgRepo(t, org, repo)
				if a, e := rev, "the-start"; callCount == 0 && a != e {
					t.Errorf("Expected to be called with revision '%s' on first call, got: '%s'", e, a)
				}
				if a, e := rev, "the-end"; callCount == 1 && a != e {
					t.Errorf("Expected to be called with revision '%s' on second call, got: '%s'", e, a)
				}
			},
			listCommitsArgValidator: func(t *testing.T, callCount int, ctx context.Context, org, repo string, clo *github.CommitsListOptions) {
				checkOrgRepo(t, org, repo)
				if page, min, max := clo.ListOptions.Page, 1, 100; page < min || page > max {
					t.Errorf("Expected page number to be between %d and %d, got: %d", min, max, page)
				}
				if a, e := clo.SHA, "the-branch"; a != e {
					t.Errorf("Expected SHA to be the branch '%s', got: '%s'", e, a)
				}
			},
			expectedGetCommitCallCount:      2,
			expectedListCommitsMinCallCount: 100,
			expectedListCommitsMaxCallCount: 100,
			expectedCommitCount:             200,
		},
		"returns no results, no further pages": {
			getCommitReturns: getCommitReturnsList{always: {
				c: &github.Commit{Committer: &github.CommitAuthor{Date: zeroTime}},
			}},
			listCommitsReturns: listCommitsReturnsList{always: {
				rc: []*github.RepositoryCommit{}, // we create 2 commits per page
				r:  response(200, 0),
			}},
			expectedGetCommitCallCount:      2,
			expectedListCommitsMaxCallCount: 1,
		},
		"http error on GetCommit(...)": {
			getCommitReturns: getCommitReturnsList{always: {
				e: fmt.Errorf("some err on GetCommit"),
			}},
			expectedGetCommitCallCount: 1,
			expectedErrMsg:             "some err on GetCommit",
		},
		"http error on 2nd GetCommit(...)": {
			getCommitReturns: getCommitReturnsList{
				always: {
					c: &github.Commit{Committer: &github.CommitAuthor{Date: zeroTime}},
				},
				1: {
					e: fmt.Errorf("some err on 2nd GetCommit"),
				},
			},
			expectedGetCommitCallCount: 2,
			expectedErrMsg:             "some err on 2nd GetCommit",
		},
		"http error on ListCommit(...)": {
			getCommitReturns: getCommitReturnsList{always: {
				c: &github.Commit{Committer: &github.CommitAuthor{Date: zeroTime}},
			}},
			listCommitsReturns: listCommitsReturnsList{always: {
				e: fmt.Errorf("some err on ListCommits"),
			}},
			expectedGetCommitCallCount:      2,
			expectedListCommitsMaxCallCount: 1,
			expectedErrMsg:                  "some err on ListCommits",
		},
		`http error on "random" ListCommit(...)`: { // random in this case means 3 ;)
			getCommitReturns: getCommitReturnsList{always: {
				c: &github.Commit{Committer: &github.CommitAuthor{Date: zeroTime}},
			}},
			listCommitsReturns: listCommitsReturnsList{
				always: {
					rc: []*github.RepositoryCommit{{}, {}}, // we create 2 commits per page
					r:  response(200, 100),
				},
				2: {
					e: fmt.Errorf("some err on a random ListCommits call"),
				},
			},
			expectedGetCommitCallCount:      2,
			expectedListCommitsMinCallCount: 3,
			expectedListCommitsMaxCallCount: 21, // This depends on how much requests we actually allow in parallel
			expectedErrMsg:                  "some err on a random ListCommits call",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			client := &githubfakes.FakeClient{}

			for i, returns := range tc.getCommitReturns {
				if i == always {
					client.GetCommitReturns(returns.c, returns.r, returns.e)
				} else {
					client.GetCommitReturnsOnCall(i, returns.c, returns.r, returns.e)
				}
			}

			for i, returns := range tc.listCommitsReturns {
				if i == always {
					client.ListCommitsReturns(returns.rc, returns.r, returns.e)
				} else {
					client.ListCommitsReturnsOnCall(i, returns.rc, returns.r, returns.e)
				}
			}

			gatherer := NewGathererWithClient(context.Background(), client)
			commits, err := gatherer.listCommits(tc.branch, tc.start, tc.end)

			checkErrMsg(t, err, tc.expectedErrMsg)

			checkCallCount(t, "GetCommits(...)", tc.expectedGetCommitCallCount, client.GetCommitCallCount())

			if min, max, a := tc.expectedListCommitsMinCallCount, tc.expectedListCommitsMaxCallCount, client.ListCommitsCallCount(); a < min || a > max {
				t.Errorf("Expected ListCommits(...) to be called between %d and %d times, got called %d times", min, max, a)
			}

			if a, e := len(commits), tc.expectedCommitCount; a != e {
				t.Errorf("Expected to get %d commits, got: %d", e, a)
			}

			if val := tc.getCommitArgValidator; val != nil {
				for i := 0; i < client.GetCommitCallCount(); i++ {
					ctx, org, repo, rev := client.GetCommitArgsForCall(i)
					val(t, i, ctx, org, repo, rev)
				}
			}
			if val := tc.listCommitsArgValidator; val != nil {
				for i := 0; i < client.ListCommitsCallCount(); i++ {
					ctx, org, repo, clo := client.ListCommitsArgsForCall(i)
					val(t, i, ctx, org, repo, clo)
				}
			}
		})
	}
}

func TestGatherNotes(t *testing.T) {
	type getPullRequestStub func(context.Context, string, string, int) (*github.PullRequest, *github.Response, error)
	type listPullRequestsWithCommitStub func(context.Context, string, string, string, *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)

	tests := map[string]struct {
		// listPullRequestsWithCommitStubber is a function that needs to return
		// another function which can be used as a stand-in for the
		// ListPullRequestsWithCommit method on the Client.
		// It can be used to check the arguments ListPullRequestsWithCommit is
		// called with and to inject faked return data.
		listPullRequestsWithCommitStubber func(*testing.T) listPullRequestsWithCommitStub
		// getPullRequestStubber is a function that needs to return another
		// function which can be used as a stand-in for the GetPullRequest method
		// on the Client.
		// It can be used to check the arguments GetPullRequest is
		// called with and to inject faked return data.
		getPullRequestStubber func(*testing.T) getPullRequestStub

		// commitList is the list of commits the gatherNotes is acting on
		commitList []*github.RepositoryCommit

		// expectedErrMsg is the error message the method is expected to return
		expectedErrMsg string
		// expectedListPullRequestsWithCommitCallCount is the expected call count
		// of the method ListPullRequestsWithCommit
		expectedListPullRequestsWithCommitCallCount int
		// expectedGetPullRequestCallCount is the expected call count of the method
		// GetPullRequest
		expectedGetPullRequestCallCount int

		// resultChecker is a function which gets called with the result of
		// gatherNotes, giving users the option to validate the returned
		// data
		resultsChecker func(*testing.T, []*Result)
	}{
		"empty commit list": {
			// Does not call anything
		},
		"when no PR number can be parsed from the commit message, we try to get a PR by SHA": {
			commitList: []*github.RepositoryCommit{
				repoCommit("some-random-sha", "some-random-commit-msg"),
			},
			listPullRequestsWithCommitStubber: func(t *testing.T) listPullRequestsWithCommitStub {
				return func(_ context.Context, org, repo, sha string, _ *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
					checkOrgRepo(t, org, repo)
					if e, a := "some-random-sha", sha; e != a {
						t.Errorf("Expected ListPullRequestsWithCommit(...) to be called for SHA '%s', have been called for '%s'", e, a)
					}
					return nil, &github.Response{}, nil
				}
			},
			expectedListPullRequestsWithCommitCallCount: 1,
		},
		"when commit messages hold PR numbers, we use those and don't query to get a list of PRs for a SHA": {
			commitList: []*github.RepositoryCommit{
				repoCommit("123", "there is the message Merge pull request #123 somewhere in the middle"),
				repoCommit("124", "some automated-cherry-pick-of-#124 can be found too"),
				repoCommit("125", "and lastly in parens (#125) yeah"),
				repoCommit("126", `all three together
					some Merge pull request #126 and
					another automated-cherry-pick-of-#127 with
					a thing (#128) in parens`),
			},
			getPullRequestStubber: func(t *testing.T) getPullRequestStub {
				seenPRs := newIntsRecorder(123, 124, 125, 126, 127, 128)

				return func(_ context.Context, org, repo string, prID int) (*github.PullRequest, *github.Response, error) {
					checkOrgRepo(t, org, repo)
					if err := seenPRs.Mark(prID); err != nil {
						t.Errorf("In GetPullRequest: %w", err)
					}
					return nil, nil, nil
				}
			},
			expectedGetPullRequestCallCount: 6,
		},
		"when GetPullRequest(...) returns an error": {
			commitList: []*github.RepositoryCommit{repoCommit("some-sha", "some #123 thing")},
			listPullRequestsWithCommitStubber: func(t *testing.T) listPullRequestsWithCommitStub {
				return func(_ context.Context, _, _, _ string, _ *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
					return nil, nil, fmt.Errorf("some-error-from-get-pull-request")
				}
			},
			expectedListPullRequestsWithCommitCallCount: 1,
			expectedErrMsg: "some-error-from-get-pull-request",
		},
		"when ListPullRequestsWithCommit(...) returns an error": {
			commitList: []*github.RepositoryCommit{repoCommit("some-sha", "some-msg")},
			listPullRequestsWithCommitStubber: func(t *testing.T) listPullRequestsWithCommitStub {
				return func(_ context.Context, _, _, _ string, _ *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
					return nil, nil, fmt.Errorf("some-error-from-list-pull-requests-with-commit")
				}
			},
			expectedListPullRequestsWithCommitCallCount: 1,
			expectedErrMsg: "some-error-from-list-pull-requests-with-commit",
		},
		"when we get PRs they get filtered based on the content of the PR body": {
			commitList: manyRepoCommits(20),
			listPullRequestsWithCommitStubber: func(t *testing.T) listPullRequestsWithCommitStub {
				prsPerCall := [][]*github.PullRequest{
					// explicitly excluded
					{pullRequest(1, "something ```release-note N/a``` something")},
					{pullRequest(2, "something ```release-note Na``` something")},
					{pullRequest(3, "something ```release-note None ``` something")},
					{pullRequest(4, "something ```release-note 'None' ``` something")},
					{pullRequest(5, "something /release-note-none something")},
					// explicitly included
					{pullRequest(6, "something release-note something")},
					{pullRequest(7, "something Does this PR introduce a user-facing change something")},
					// multiple PRs
					{ // first does no match, second one matches, rest is ignored
						pullRequest(8, ""),
						pullRequest(9, " Does this PR introduce a user-facing change"),
						pullRequest(10, "does-not-matter--is-not-considered"),
					},
					// some other strange things
					{pullRequest(11, "Does this PR introduce a user-facing chang")}, // matches despite the missing 'e'
					{pullRequest(12, "release-note /release-note-none")},            // excluded, the exclusion filters take precedence
					{pullRequest(13, "```release-note NAAAAAAAAAA```")},             // included, does not match the N/A filter, but the 'release-note' check
					{pullRequest(14, "```release-note none something ```")},         // included, does not match the N/A filter, but the 'release-note' check
					{pullRequest(15, "release-noteNOOOO")},                          // included
				}
				var callCount int64 = -1

				return func(_ context.Context, _, _, _ string, _ *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
					callCount := int(atomic.AddInt64(&callCount, 1))
					if a, e := callCount+1, len(prsPerCall); a > e {
						return nil, &github.Response{}, nil
					}
					return prsPerCall[callCount], &github.Response{}, nil
				}
			},
			expectedListPullRequestsWithCommitCallCount: 20,
			resultsChecker: func(t *testing.T, results []*Result) {
				// there is not much we can check on the Result, as all the fields are
				// unexported
				expectedResultSize := 13
				if e, a := expectedResultSize, len(results); e != a {
					t.Errorf("Expected the result to be of size %d, got %d", e, a)
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			client := &githubfakes.FakeClient{}

			gatherer := NewGathererWithClient(context.Background(), client)
			if stubber := tc.listPullRequestsWithCommitStubber; stubber != nil {
				client.ListPullRequestsWithCommitStub = stubber(t)
			}
			if stubber := tc.getPullRequestStubber; stubber != nil {
				client.GetPullRequestStub = stubber(t)
			}

			results, err := gatherer.gatherNotes(tc.commitList)

			checkErrMsg(t, err, tc.expectedErrMsg)

			if checker := tc.resultsChecker; checker != nil {
				checker(t, results)
			}

			checkCallCount(t, "ListPullRequestsWithCommit(...)",
				tc.expectedListPullRequestsWithCommitCallCount, client.ListPullRequestsWithCommitCallCount(),
			)
			checkCallCount(t, "GetPullRequest(...)",
				tc.expectedGetPullRequestCallCount, client.GetPullRequestCallCount(),
			)
		})
	}
}

func pullRequest(id int, msg string) *github.PullRequest {
	return &github.PullRequest{
		Body:   strPtr(msg),
		Number: intPtr(id),
	}
}

func manyRepoCommits(nr int) []*github.RepositoryCommit {
	cs := make([]*github.RepositoryCommit, nr)

	for i := 1; i <= nr; i++ {
		cs[i-1] = repoCommit(fmt.Sprintf("commit-%d", i), fmt.Sprintf("commit-msg-%d", i))
	}

	return cs
}

func repoCommit(sha, commitMsg string) *github.RepositoryCommit {
	return &github.RepositoryCommit{
		SHA: strPtr(sha),
		Commit: &github.Commit{
			Message: strPtr(commitMsg),
		},
	}
}

// newIntsRecorder gives you an intsRecorder which you can use to keep track of
// elements you have already seen. You need to preload it with the elements you
// expect, it will return an error if you try to makr an element as seen which
// is not in the list of preloaded elements or that you have already marked.
// intsRecorder is goroutine safe, so you can use it to e.g. check if
// concurrent gofuncs don't run for the same thing.
func newIntsRecorder(ints ...int) *intsRecorder {
	l := make(map[int]bool, len(ints))

	for _, i := range ints {
		l[i] = false
	}

	return &intsRecorder{seen: l}
}

type intsRecorder struct {
	sync.Mutex
	seen map[int]bool
}

func (s *intsRecorder) Mark(what int) error {
	s.Lock()
	defer s.Unlock()

	seen, ok := s.seen[what]
	if !ok {
		return fmt.Errorf("Expected not to get a request to mark %d as seen", what)
	}
	if seen {
		return fmt.Errorf("Expected to mark %d as seen only once", what)
	}
	s.seen[what] = true
	return nil
}

func intPtr(i int) *int       { return &i }
func strPtr(s string) *string { return &s }

func checkCallCount(t *testing.T, what string, expected, actual int) {
	t.Helper()

	if expected != actual {
		t.Errorf("Expected %s to be called %d times, got called %d times", what, expected, actual)
	}
}

func checkOrgRepo(t *testing.T, org, repo string) {
	t.Helper()

	if org != git.DefaultGithubOrg {
		t.Errorf("Expected to be called with '%s' as an org, got: %s", git.DefaultGithubOrg, org)
	}
	if repo != git.DefaultGithubRepo {
		t.Errorf("Expected to be called with '%s' as a repo, got: %s", git.DefaultGithubRepo, repo)
	}
}

func checkErrMsg(t *testing.T, err error, expectedMsg string) {
	t.Helper()

	if expectedMsg == "" {
		if err != nil {
			t.Errorf("Expected no error, got: %#v", err)
		}
		return
	}

	if err == nil {
		t.Errorf("Expected error, but got none")
		return
	}

	if e, a := expectedMsg, err.Error(); !strings.Contains(a, e) {
		t.Errorf("Expected error to match '%s', got: '%s'", e, a)
	}
}

func response(statusCode, lastPage int) *github.Response {
	res := &github.Response{
		LastPage: lastPage,
		Response: &http.Response{
			StatusCode: statusCode,
		},
	}
	return res
}

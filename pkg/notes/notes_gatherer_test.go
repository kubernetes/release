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

package notes_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v28/github"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/notesfakes"
)

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
		// opts is the GithubAPIOptions to use when making API calls
		opts []notes.GitHubAPIOption
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
			opts:   []notes.GitHubAPIOption{notes.WithBranch("branch-from-opts-is-ignored")},
			branch: "the-branch", start: "the-start", end: "the-end",
			getCommitReturns: getCommitReturnsList{always: {
				c: &github.Commit{Committer: &github.CommitAuthor{Date: zeroTime}},
			}},
			listCommitsReturns: listCommitsReturnsList{always: {
				rc: []*github.RepositoryCommit{{}, {}}, // we create 2 commits per page
				r:  response(200, 100),
			}},
			getCommitArgValidator: func(t *testing.T, callCount int, ctx context.Context, org, repo, rev string) {
				checkOrgRepo(t, "kubernetes", "kubernetes", org, repo)
				if a, e := rev, "the-start"; callCount == 0 && a != e {
					t.Errorf("Expected to be called with revision '%s' on first call, got: '%s'", e, a)
				}
				if a, e := rev, "the-end"; callCount == 1 && a != e {
					t.Errorf("Expected to be called with revision '%s' on second call, got: '%s'", e, a)
				}
			},
			listCommitsArgValidator: func(t *testing.T, callCount int, ctx context.Context, org, repo string, clo *github.CommitsListOptions) {
				checkOrgRepo(t, "kubernetes", "kubernetes", org, repo)
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
			expectedListCommitsMaxCallCount: 21, // This depends on how much requests we actually allow in parrallel
			expectedErrMsg:                  "some err on a random ListCommits call",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			client := &notesfakes.FakeClient{}

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

			gatherer := notes.Gatherer{
				Client: client,
				Opts:   tc.opts,
			}

			commits, err := gatherer.ListCommits(tc.branch, tc.start, tc.end)

			checkErrMsg(t, err, tc.expectedErrMsg)

			if a, e := client.GetCommitCallCount(), tc.expectedGetCommitCallCount; a != e {
				t.Errorf("Expected GetCommits(...) to be called %d times, got called %d times", e, a)
			}

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

func checkOrgRepo(t *testing.T, expectedOrg, expectedRepo, actualOrg, actualRepo string) {
	t.Helper()

	if a, e := actualOrg, expectedOrg; e != a {
		t.Errorf("Expected to be called with '%s' as an org, got: %s", e, a)
	}
	if a, e := actualRepo, expectedRepo; e != a {
		t.Errorf("Expected to be called with '%s' as a repo, got: %s", e, a)
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

func response(statusCode int, lastPage int) *github.Response {
	res := &github.Response{
		LastPage: lastPage,
		Response: &http.Response{
			StatusCode: statusCode,
		},
	}
	return res
}

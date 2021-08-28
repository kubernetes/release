/*
Copyright 2020 The Kubernetes Authors.

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

package github_test

import (
	"errors"
	"fmt"
	"testing"

	gogithub "github.com/google/go-github/v37/github"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/github/githubfakes"
)

func newSUT() (*github.GitHub, *githubfakes.FakeClient) {
	client := &githubfakes.FakeClient{}
	sut := github.New()
	sut.SetClient(client)

	return sut, client
}

func TestLatestGitHubTagsPerBranchSuccessEmptyResult(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListTagsReturns(nil, &gogithub.Response{NextPage: 0}, nil)

	// When
	res, err := sut.LatestGitHubTagsPerBranch()

	// Then
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestLatestGitHubTagsPerBranchSuccessAlphaAfterMinor(t *testing.T) {
	// Given
	var (
		tag1 = "v1.18.0-alpha.2"
		tag2 = "v1.18.0"
	)
	sut, client := newSUT()
	client.ListTagsReturns([]*gogithub.RepositoryTag{
		{Name: &tag1},
		{Name: &tag2},
	}, &gogithub.Response{NextPage: 0}, nil)

	// When
	res, err := sut.LatestGitHubTagsPerBranch()

	// Then
	require.Nil(t, err)
	require.Len(t, res, 2)
	require.Equal(t, tag1, res[git.DefaultBranch])
	require.Equal(t, tag2, res["release-1.18"])
}

func TestLatestGitHubTagsPerBranchMultiplePages(t *testing.T) {
	// Given
	var (
		tag1 = "v1.18.0-alpha.2"
		tag2 = "v1.18.0"
	)
	sut, client := newSUT()
	client.ListTagsReturnsOnCall(0, []*gogithub.RepositoryTag{
		{Name: &tag1},
	}, &gogithub.Response{NextPage: 1}, nil)
	client.ListTagsReturnsOnCall(1, []*gogithub.RepositoryTag{
		{Name: &tag2},
	}, &gogithub.Response{NextPage: 0}, nil)

	// When
	res, err := sut.LatestGitHubTagsPerBranch()

	// Then
	require.Nil(t, err)
	require.Len(t, res, 2)
	require.Equal(t, tag1, res[git.DefaultBranch])
	require.Equal(t, tag2, res["release-1.18"])
}

func TestLatestGitHubTagsPerBranchSuccessMultipleForSameBranch(t *testing.T) {
	// Given
	var (
		tag1 = "v1.18.0-beta.0"
		tag2 = "v1.18.0-alpha.3"
		tag3 = "v1.15.2"
		tag4 = "v1.18.0-alpha.2"
		tag5 = "v1.16.3"
		tag6 = "v1.18.0-alpha.1"
		tag7 = "v1.13.0"
		tag8 = "v1.18.0-alpha.2"
	)
	sut, client := newSUT()
	client.ListTagsReturns([]*gogithub.RepositoryTag{
		{Name: &tag1},
		{Name: &tag2},
		{Name: &tag3},
		{Name: &tag4},
		{Name: &tag5},
		{Name: &tag6},
		{Name: &tag7},
		{Name: &tag8},
	}, &gogithub.Response{NextPage: 0}, nil)

	// When
	res, err := sut.LatestGitHubTagsPerBranch()

	// Then
	require.Nil(t, err)
	require.Len(t, res, 4)
	require.Equal(t, tag1, res[git.DefaultBranch])
	require.Empty(t, res["release-1.18"])
	require.Empty(t, res["release-1.17"])
	require.Equal(t, tag5, res["release-1.16"])
	require.Equal(t, tag3, res["release-1.15"])
	require.Empty(t, res["release-1.14"])
	require.Equal(t, tag7, res["release-1.13"])
}

func TestLatestGitHubTagsPerBranchSuccessPatchReleases(t *testing.T) {
	// Given
	var (
		tag1 = "v1.17.1"
		tag2 = "v1.16.2"
		tag3 = "v1.15.3"
	)
	sut, client := newSUT()
	client.ListTagsReturns([]*gogithub.RepositoryTag{
		{Name: &tag1},
		{Name: &tag2},
		{Name: &tag3},
	}, &gogithub.Response{NextPage: 0}, nil)

	// When
	res, err := sut.LatestGitHubTagsPerBranch()

	// Then
	require.Nil(t, err)
	require.Len(t, res, 4)
	require.Equal(t, tag1, res[git.DefaultBranch])
	require.Equal(t, tag1, res["release-1.17"])
	require.Equal(t, tag2, res["release-1.16"])
	require.Equal(t, tag3, res["release-1.15"])
	require.Empty(t, res["release-1.18"])
}

func TestLatestGitHubTagsPerBranchFailedOnList(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListTagsReturns(nil, nil, errors.New("error"))

	// When
	res, err := sut.LatestGitHubTagsPerBranch()

	// Then
	require.NotNil(t, err)
	require.Nil(t, res)
}

func TestLatestGitHubTagsPerBranchSkippedNonSemverTag(t *testing.T) {
	// Given
	tag1 := "not a semver tag"
	sut, client := newSUT()
	client.ListTagsReturns([]*gogithub.RepositoryTag{
		{Name: &tag1},
	}, &gogithub.Response{NextPage: 0}, nil)

	// When
	res, err := sut.LatestGitHubTagsPerBranch()

	// Then
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestReleasesSuccessEmpty(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListReleasesReturns([]*gogithub.RepositoryRelease{}, nil, nil)

	// When
	res, err := sut.Releases("", "", false)

	// Then
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestReleasesSuccessNoPreReleases(t *testing.T) {
	// Given
	var (
		tag1  = "v1.18.0"
		tag2  = "v1.17.0"
		tag3  = "v1.16.0"
		tag4  = "v1.15.0"
		aTrue = true
	)
	sut, client := newSUT()
	client.ListReleasesReturns([]*gogithub.RepositoryRelease{
		{TagName: &tag1},
		{TagName: &tag2},
		{TagName: &tag3, Prerelease: &aTrue},
		{TagName: &tag4},
	}, nil, nil)

	// When
	res, err := sut.Releases("", "", false)

	// Then
	require.Nil(t, err)
	require.Len(t, res, 3)
	require.Equal(t, tag1, res[0].GetTagName())
	require.Equal(t, tag2, res[1].GetTagName())
	require.Equal(t, tag4, res[2].GetTagName())
}

func TestReleasesSuccessWithPreReleases(t *testing.T) {
	// Given
	var (
		tag1  = "v1.18.0"
		tag2  = "v1.17.0"
		tag3  = "v1.16.0"
		tag4  = "v1.15.0"
		aTrue = true
	)
	sut, client := newSUT()
	client.ListReleasesReturns([]*gogithub.RepositoryRelease{
		{TagName: &tag1},
		{TagName: &tag2, Prerelease: &aTrue},
		{TagName: &tag3, Prerelease: &aTrue},
		{TagName: &tag4},
	}, nil, nil)

	// When
	res, err := sut.Releases("", "", true)

	// Then
	require.Nil(t, err)
	require.Len(t, res, 4)
	require.Equal(t, tag1, res[0].GetTagName())
	require.Equal(t, tag2, res[1].GetTagName())
	require.Equal(t, tag3, res[2].GetTagName())
	require.Equal(t, tag4, res[3].GetTagName())
}

func TestReleasesFailed(t *testing.T) {
	// Given
	sut, client := newSUT()
	client.ListReleasesReturns(nil, nil, errors.New("error"))

	// When
	res, err := sut.Releases("", "", false)

	// Then
	require.NotNil(t, err)
	require.Nil(t, res, nil)
}

func TestCreatePullRequest(t *testing.T) {
	// Given
	sut, client := newSUT()
	fakeID := int64(1234)
	client.CreatePullRequestReturns(&gogithub.PullRequest{ID: &fakeID}, nil)

	// When
	pr, err := sut.CreatePullRequest("kubernetes-fake-org", "kubernetes-fake-repo", git.DefaultBranch, "user:head-branch", "PR Title", "PR Body")

	// Then
	require.Nil(t, err)
	require.NotNil(t, pr, nil)
	require.Equal(t, fakeID, pr.GetID())
}

func TestGetMilestone(t *testing.T) {
	sut, client := newSUT()
	// Given
	searchTitle := "Target Milestone"
	otherTitle := "Another Milestone"
	fakeMstoneID := 9999

	client.ListMilestonesReturns(
		[]*gogithub.Milestone{
			{
				Title: &otherTitle,
			},
			{
				Number: &fakeMstoneID,
				Title:  &searchTitle,
			},
		},
		&gogithub.Response{NextPage: 0},
		nil,
	)

	// When
	for _, tc := range []struct {
		Title string
		Err   bool
	}{
		{Title: searchTitle},
		{Title: "Non existent"},
		{Title: "", Err: true},
	} {
		ms, exists, err := sut.GetMilestone("test", "test", tc.Title)

		// Then
		if searchTitle == tc.Title {
			require.True(t, exists)
			require.Equal(t, fakeMstoneID, ms.GetNumber())
			require.Equal(t, searchTitle, ms.GetTitle())
		} else {
			require.False(t, exists)
		}

		if tc.Err {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestGetRepository(t *testing.T) {
	// Given
	sut, client := newSUT()
	fakeRepositoryID := int64(54596517) // k/release
	kubernetesUserID := int64(13629408)
	kubernetesLogin := "kubernetes"
	repoName := "release"
	client.GetRepositoryReturns(&gogithub.Repository{
		ID:   &fakeRepositoryID,
		Name: &repoName,
		Owner: &gogithub.User{
			Login: &kubernetesLogin,
			ID:    &kubernetesUserID,
		},
	}, &gogithub.Response{NextPage: 0}, nil)

	// When
	repo, err := sut.GetRepository("kubernetes", "release")

	// Then
	require.Nil(t, err)
	require.NotNil(t, repo, nil)
	require.Equal(t, fakeRepositoryID, repo.GetID())
	require.Equal(t, kubernetesUserID, repo.GetOwner().GetID())
	require.Equal(t, kubernetesLogin, repo.GetOwner().GetLogin())
	require.Equal(t, repoName, repo.GetName())
}

func TestRepoIsForkOf(t *testing.T) {
	// Given
	sut, client := newSUT()

	forkOwner := "fork"
	parentOwner := "kubernetes"
	repoName := "forkedRepo"

	parentFullName := fmt.Sprintf("%s/%s", parentOwner, repoName)

	trueVal := true

	client.GetRepositoryReturns(&gogithub.Repository{
		Name: &repoName,
		Fork: &trueVal,
		Owner: &gogithub.User{
			Login: &forkOwner,
		},
		Parent: &gogithub.Repository{
			Name: &repoName,
			Owner: &gogithub.User{
				Login: &parentOwner,
			},
			FullName: &parentFullName,
		},
	}, &gogithub.Response{NextPage: 0}, nil)

	// When
	result, err := sut.RepoIsForkOf("fork", repoName, "kubernetes", repoName)

	// Then
	require.Nil(t, err)
	require.Equal(t, result, true)
}

func TestRepoIsNotForkOf(t *testing.T) {
	// Given
	sut, client := newSUT()

	forkOwner := "fork"
	parentOwner := "borg"
	repoName := "notForkedRepo"

	parentFullName := fmt.Sprintf("%s/%s", parentOwner, repoName)

	trueVal := true

	client.GetRepositoryReturns(&gogithub.Repository{
		Name: &repoName,
		Fork: &trueVal,
		Owner: &gogithub.User{
			Login: &forkOwner,
		},
		Parent: &gogithub.Repository{
			Name: &repoName,
			Owner: &gogithub.User{
				Login: &parentOwner,
			},
			FullName: &parentFullName,
		},
	}, &gogithub.Response{NextPage: 0}, nil)

	// When
	result, err := sut.RepoIsForkOf("fork", repoName, "kubernetes", repoName)

	// Then
	require.Nil(t, err)
	require.Equal(t, result, false)
}

func TestListBranches(t *testing.T) {
	// Given
	sut, client := newSUT()

	branch0 := git.DefaultBranch
	branch1 := "myfork"
	branch2 := "feature-branch"

	branches := []*gogithub.Branch{
		{
			Name: &branch0,
		},
		{
			Name: &branch1,
		},
		{
			Name: &branch2,
		},
	}

	client.ListBranchesReturns(branches, &gogithub.Response{NextPage: 0}, nil)

	// When
	result, err := sut.ListBranches("kubernetes", "kubernotia")

	// Then
	require.Nil(t, err)
	require.Len(t, result, 3)
	require.Equal(t, result[1].GetName(), branch1)
}

func TestCreateIssue(t *testing.T) {
	// Given
	sut, client := newSUT()
	fakeID := 100000
	title := "Test Issue"
	body := "Issue body text"
	opts := &github.NewIssueOptions{
		Assignees: []string{"k8s-ci-robot"},
		Milestone: "v1.21",
		State:     "open",
		Labels:    []string{"bug"},
	}
	issue := &gogithub.Issue{
		Number: &fakeID,
		State:  &opts.State,
		Title:  &title,
		Body:   &body,
	}

	for _, tcErr := range []error{errors.New("Test error"), nil} {
		// When
		client.CreateIssueReturns(issue, tcErr)
		newissue, err := sut.CreateIssue("kubernetes-fake-org", "kubernetes-fake-repo", title, body, opts)

		// Then
		if tcErr == nil {
			require.Nil(t, err)
			require.NotNil(t, newissue)
			require.Equal(t, fakeID, issue.GetNumber())
		} else {
			require.NotNil(t, err)
		}
	}
}

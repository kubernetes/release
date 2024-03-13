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

package fastforward

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-sdk/github"
)

// Options is the main structure for configuring a fast forward.
type Options struct {
	// GitHubOrg is the GitHub Organization to be used do the initial clone.
	GitHubOrg string

	// GitHubRepo is the GitHub Repository to be used do the initial clone.
	GitHubRepo string

	// Branch is the release branch to be fast forwarded.
	Branch string

	// MainRef is the git ref of the base branch.
	MainRef string

	// Submit can be used to run inside of a new Google Cloud Build job.
	Submit bool

	// NonInteractive does not ask any questions if set to true.
	NonInteractive bool

	// NoMock actually pushes the changes if set to true.
	NoMock bool

	// Cleanup the repository after the run if set to true.
	Cleanup bool

	// RepoPath is the local path to the repository to be used.
	RepoPath string

	// GCPProjectID is the GCP project to use to submit the job.
	GCPProjectID string
}

// FastForward is the main structure of this package.
type FastForward struct {
	impl
	options *Options
}

// New returns a new FastForward instance.
func New(opts *Options) *FastForward {
	return &FastForward{
		impl:    &defaultImpl{},
		options: opts,
	}
}

const pushUpstreamQuestion = `Are you ready to push the local branch fast-forward changes upstream?
Please only answer after you have validated the changes.`

// Run starts the FastForward.
func (f *FastForward) Run() (err error) {
	if f.options.Submit {
		if err := f.prepareToolRepo(); err != nil {
			return fmt.Errorf("prepare tool repo: %w", err)
		}
		logrus.Info("Submitting GCB job")
		options := gcb.NewDefaultOptions()
		options.FastForward = true
		options.NoMock = f.options.NoMock
		options.NonInteractive = f.options.NonInteractive
		options.Stream = true
		options.Project = f.options.GCPProjectID
		options.ScratchBucket = "gs://" + f.options.GCPProjectID + "-gcb"
		options.CustomK8SRepo = f.options.GitHubRepo
		options.CustomK8sOrg = f.options.GitHubOrg
		return f.Submit(options)
	}

	repo, err := f.prepareFastForwardRepo()
	if err != nil {
		return fmt.Errorf("prepare repository: %w", err)
	}

	if !f.options.NoMock {
		logrus.Info("Using dry mode, which does not modify any remote content")
		f.RepoSetDry(repo)
	}

	branch := f.options.Branch
	if branch == "" {
		logrus.Info("No release branch specified, finding the latest")
		branch, err = f.RepoLatestReleaseBranch(repo)
		if err != nil {
			return fmt.Errorf("finding latest release branch: %w", err)
		}
		logrus.Infof("Found latest release branch: %s", branch)

		notRequired, err := f.noFastForwardRequired(repo, branch)
		if err != nil {
			return fmt.Errorf("check if fast forward is required: %w", err)
		}
		if notRequired {
			logrus.Infof(
				"Fast forward not required because final tag already exists for latest release branch %s",
				branch,
			)
			return nil
		}
	} else {
		logrus.Infof("Checking if %q is a release branch", branch)
		if isReleaseBranch := f.IsReleaseBranch(branch); !isReleaseBranch {
			return fmt.Errorf("%s is not a release branch", branch)
		}

		logrus.Info("Checking if branch is available on the default remote")
		branchExists, err := f.RepoHasRemoteBranch(repo, branch)
		if err != nil {
			return fmt.Errorf("checking if branch exists on the default remote: %w", err)
		}
		if !branchExists {
			return errors.New("branch does not exist on the default remote")
		}
	}

	issues, err := f.ListIssues()
	if err != nil {
		return fmt.Errorf(
			"unable to list GitHub issues for %s/%s repo: %w",
			git.DefaultGithubOrg, git.DefaultGithubReleaseRepo, err)
	}
	title := fmt.Sprintf("Cut %s release", f.branchToVersion(branch))
	for _, issue := range issues {
		if issue.IsPullRequest() {
			continue
		}
		if issue.GetTitle() == title {
			logrus.Infof("Skipping fast forward: release cut issue is open: %s", issue.GetURL())
			return nil
		}
	}

	if f.options.Cleanup {
		defer func() {
			if err := f.RepoCleanup(repo); err != nil {
				logrus.Errorf("Repo cleanup failed: %v", err)
			}
		}()
	} else {
		// Restore the currently checked out branch afterwards
		currentBranch, err := f.RepoCurrentBranch(repo)
		if err != nil {
			return fmt.Errorf("unable to retrieve current branch: %w", err)
		}
		defer func() {
			if err := f.RepoCheckout(repo, currentBranch); err != nil {
				logrus.Errorf("Unable to restore branch %s: %v", currentBranch, err)
			}
		}()
	}

	logrus.Info("Checking out release branch")
	if err := f.RepoCheckout(repo, branch); err != nil {
		return fmt.Errorf("checking out branch %s: %w", branch, err)
	}

	logrus.Infof("Finding merge base between %q and %q", git.DefaultBranch, branch)
	mergeBase, err := f.RepoMergeBase(repo, git.DefaultBranch, branch)
	if err != nil {
		return fmt.Errorf("find merge base: %w", err)
	}

	// Verify the tags
	mainTag, err := f.RepoDescribe(
		repo,
		git.NewDescribeOptions().
			WithRevision(git.Remotify(git.DefaultBranch)).
			WithAbbrev(0).
			WithTags(),
	)
	if err != nil {
		return fmt.Errorf("describe latest main tag: %w", err)
	}
	mergeBaseTag, err := f.RepoDescribe(
		repo,
		git.NewDescribeOptions().
			WithRevision(mergeBase).
			WithAbbrev(0).
			WithTags(),
	)
	if err != nil {
		return fmt.Errorf("describe latest merge base tag: %w", err)
	}
	logrus.Infof("Merge base tag is: %s", mergeBaseTag)

	if mainTag != mergeBaseTag {
		return fmt.Errorf(
			"unable to fast forward: tag %q does not match %q",
			mainTag, mergeBaseTag,
		)
	}
	logrus.Infof("Verified that the latest tag on the main branch is the same as the merge base tag")

	releaseRev, err := f.RepoHead(repo)
	if err != nil {
		return fmt.Errorf("get release rev: %w", err)
	}
	logrus.Infof("Latest release branch revision is %s", releaseRev)

	logrus.Info("Configuring git user and email")
	if err := f.ConfigureGlobalDefaultUserAndEmail(); err != nil {
		return fmt.Errorf("configure git user and email: %w", err)
	}

	logrus.Info("Merging main branch changes into release branch")
	if err := f.RepoMerge(repo, f.options.MainRef); err != nil {
		return fmt.Errorf("merge main ref: %w", err)
	}

	headRev, err := f.RepoHead(repo)
	if err != nil {
		return fmt.Errorf("get HEAD rev: %w", err)
	}

	prepushMessage(f.RepoDir(repo), f.options.GitHubOrg, f.options.GitHubRepo, branch, f.options.MainRef, releaseRev, headRev)

	pushUpstream := f.options.NonInteractive
	if !pushUpstream {
		_, pushUpstream, err = f.Ask(pushUpstreamQuestion, "yes", 3)
		if err != nil {
			return fmt.Errorf("ask upstream question: %w", err)
		}
	}

	if pushUpstream {
		logrus.Infof("Pushing %s branch", branch)
		if err := f.RepoPush(repo, branch); err != nil {
			return fmt.Errorf("push to repo: %w", err)
		}
	}

	return nil
}

func prepushMessage(gitRoot, org, repo, branch, ref, releaseRev, headRev string) {
	fmt.Printf(`Go look around in %s to make sure things look okay before pushing…
	
	Check for files left uncommitted using:
	
		git status -s
	
	Validate the fast-forward commit using:
	
		git show
	
	Validate the changes pulled in from main branch using:
	
		git log %s..%s
	
	Once the branch fast-forward is complete, the diff will be available after push at:
	
		https://github.com/%s/%s/compare/%s...%s
	
	`,
		gitRoot,
		git.Remotify(branch),
		ref,
		org,
		repo,
		releaseRev,
		headRev,
	)
}

func (f *FastForward) noFastForwardRequired(repo *git.Repo, branch string) (bool, error) {
	version := f.branchToVersion(branch)

	tagExists, err := f.RepoHasRemoteTag(repo, version)
	if err != nil {
		return false, fmt.Errorf("finding remote tag %s: %w", version, err)
	}

	return tagExists, nil
}

func (f *FastForward) branchToVersion(branch string) string {
	return fmt.Sprintf("v%s.0", strings.TrimPrefix(branch, "release-"))
}

func (f *FastForward) prepareFastForwardRepo() (*git.Repo, error) {
	logrus.Infof("Preparing to %s/%s fast-forward from %s", f.options.GitHubOrg, f.options.GitHubRepo, f.options.MainRef)

	token := f.EnvDefault(github.TokenEnvKey, "")

	useSSH := true
	stringMsg := "using SSH"
	if token != "" {
		useSSH = false
		stringMsg = "using HTTPs"
	}

	logrus.Infof("Cloning repository %s/%s %s", f.options.GitHubOrg, f.options.GitHubRepo, stringMsg)
	repo, err := f.CloneOrOpenGitHubRepo(f.options.RepoPath, f.options.GitHubOrg, f.options.GitHubRepo, useSSH)
	if err != nil {
		return nil, fmt.Errorf("clone or open %s/%s GitHub repository: %w",
			f.options.GitHubOrg, f.options.GitHubRepo, err,
		)
	}

	if token != "" {
		logrus.Info("Found GitHub token, using it for repository interactions")
		if f.IsDefaultK8sUpstream() {
			if err := f.RepoSetURL(repo, git.DefaultRemote, (&url.URL{
				Scheme: "https",
				User:   url.UserPassword("git", token),
				Host:   "github.com",
				Path:   filepath.Join(f.options.GitHubOrg, f.options.GitHubRepo),
			}).String()); err != nil {
				return nil, fmt.Errorf("changing git remote of repository: %w", err)
			}
		} else {
			logrus.Info("Using non-default k8s upstream, doing no git modifications")
		}
	}

	return repo, nil
}

func (f *FastForward) prepareToolRepo() error {
	if f.Exists(".git") {
		return nil
	}

	logrus.Info("Not in a git repo, preparing k/release clone")

	tmpPath, err := f.MkdirTemp("", "k-release-")
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	if err := f.RemoveAll(tmpPath); err != nil {
		return fmt.Errorf("remove temp directory: %w", err)
	}
	if _, err := f.CloneOrOpenGitHubRepo(
		tmpPath,
		release.DefaultToolOrg,
		release.DefaultToolRepo,
		false,
	); err != nil {
		return fmt.Errorf("clone tool repository: %w", err)
	}
	if err := f.Chdir(tmpPath); err != nil {
		return fmt.Errorf("change directory: %w", err)
	}

	return nil
}

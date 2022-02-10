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
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-sdk/github"
)

// Options is the main structure for configuring a fast forward.
type Options struct {
	// Branch is the release branch to be fast forwarded.
	Branch string

	// MainRef is the git ref ot the base branch.
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
			return errors.Wrap(err, "prepare tool repo")
		}
		logrus.Info("Submitting GCB job")
		options := gcb.NewDefaultOptions()
		options.FastForward = true
		options.NoMock = f.options.NoMock
		options.Stream = true
		options.Project = f.options.GCPProjectID
		options.ScratchBucket = "gs://" + f.options.GCPProjectID + "-gcb"
		return f.Submit(options)
	}

	repo, err := f.prepareKubernetesRepo()
	if err != nil {
		return errors.Wrap(err, "prepare repository")
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
			return errors.Wrap(err, "finding latest release branch")
		}
		logrus.Infof("Found latest release branch: %s", branch)

		notRequired, err := f.noFastForwardRequired(repo, branch)
		if err != nil {
			return errors.Wrap(err, "check if fast forward is required")
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
			return errors.Errorf("%s is not a release branch", branch)
		}

		logrus.Info("Checking if branch is available on the default remote")
		branchExists, err := f.RepoHasRemoteBranch(repo, branch)
		if err != nil {
			return errors.Wrap(err, "checking if branch exists on the default remote")
		}
		if !branchExists {
			return errors.New("branch does not exist on the default remote")
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
			return errors.Wrap(err, "unable to retrieve current branch")
		}
		defer func() {
			if err := f.RepoCheckout(repo, currentBranch); err != nil {
				logrus.Errorf("Unable to restore branch %s: %v", currentBranch, err)
			}
		}()
	}

	logrus.Info("Checking out release branch")
	if err := f.RepoCheckout(repo, branch); err != nil {
		return errors.Wrapf(err, "checking out branch %s", branch)
	}

	logrus.Infof("Finding merge base between %q and %q", git.DefaultBranch, branch)
	mergeBase, err := f.RepoMergeBase(repo, git.DefaultBranch, branch)
	if err != nil {
		return errors.Wrap(err, "find merge base")
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
		return errors.Wrap(err, "describe latest main tag")
	}
	mergeBaseTag, err := f.RepoDescribe(
		repo,
		git.NewDescribeOptions().
			WithRevision(mergeBase).
			WithAbbrev(0).
			WithTags(),
	)
	if err != nil {
		return errors.Wrap(err, "describe latest merge base tag")
	}
	logrus.Infof("Merge base tag is: %s", mergeBaseTag)

	if mainTag != mergeBaseTag {
		return errors.Errorf(
			"unable to fast forward: tag %q does not match %q",
			mainTag, mergeBaseTag,
		)
	}
	logrus.Infof("Verified that the latest tag on the main branch is the same as the merge base tag")

	releaseRev, err := f.RepoHead(repo)
	if err != nil {
		return errors.Wrap(err, "get release rev")
	}
	logrus.Infof("Latest release branch revision is %s", releaseRev)

	logrus.Info("Merging main branch changes into release branch")
	if err := f.RepoMerge(repo, f.options.MainRef); err != nil {
		return errors.Wrap(err, "merge main ref")
	}

	headRev, err := f.RepoHead(repo)
	if err != nil {
		return errors.Wrap(err, "get HEAD rev")
	}

	prepushMessage(f.RepoDir(repo), branch, f.options.MainRef, releaseRev, headRev)

	pushUpstream := false
	if f.options.NonInteractive {
		pushUpstream = true
	} else {
		_, pushUpstream, err = f.Ask(pushUpstreamQuestion, "yes", 3)
		if err != nil {
			return errors.Wrap(err, "ask upstream question")
		}
	}

	if pushUpstream {
		logrus.Infof("Pushing %s branch", branch)
		if err := f.RepoPush(repo, branch); err != nil {
			return errors.Wrap(err, "push to repo")
		}
	}

	return nil
}

func prepushMessage(gitRoot, branch, ref, releaseRev, headRev string) {
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
		git.DefaultGithubOrg,
		git.DefaultGithubRepo,
		releaseRev,
		headRev,
	)
}

func (f *FastForward) noFastForwardRequired(repo *git.Repo, branch string) (bool, error) {
	version := fmt.Sprintf("v%s.0", strings.TrimPrefix(branch, "release-"))

	tagExists, err := f.RepoHasRemoteTag(repo, version)
	if err != nil {
		return false, errors.Wrapf(err, "finding remote tag %s", version)
	}

	return tagExists, nil
}

func (f *FastForward) prepareKubernetesRepo() (*git.Repo, error) {
	logrus.Infof("Preparing to fast-forward from %s", f.options.MainRef)

	token := f.EnvDefault(github.TokenEnvKey, "")
	if token != "" {
		logrus.Info("Found GitHub token, using it for repository interactions")
		k8sOrg := release.GetK8sOrg()
		k8sRepo := release.GetK8sRepo()

		logrus.Info("Cloning repository by using HTTPs")
		repo, err := f.CloneOrOpenGitHubRepo(f.options.RepoPath, k8sOrg, k8sRepo, false)
		if err != nil {
			return nil, errors.Wrap(err, "clone or open k/k GitHub repository")
		}

		if f.IsDefaultK8sUpstream() {
			if err := f.RepoSetURL(repo, git.DefaultRemote, (&url.URL{
				Scheme: "https",
				User:   url.UserPassword("git", token),
				Host:   "github.com",
				Path:   filepath.Join(git.DefaultGithubOrg, git.DefaultGithubRepo),
			}).String()); err != nil {
				return nil, errors.Wrap(err, "changing git remote of repository")
			}
		} else {
			logrus.Info("Using non-default k8s upstream, doing no git modifications")
		}

		return repo, nil
	}

	logrus.Info("Cloning repository by using SSH")
	repo, err := f.CloneOrOpenDefaultGitHubRepoSSH(f.options.RepoPath)
	if err != nil {
		return nil, errors.Wrap(err, "clone or open k/k GitHub repository")
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
		return errors.Wrap(err, "create temp directory")
	}
	if err := f.RemoveAll(tmpPath); err != nil {
		return errors.Wrap(err, "remove temp directory")
	}
	if _, err := f.CloneOrOpenGitHubRepo(
		tmpPath,
		release.DefaultToolOrg,
		release.DefaultToolRepo,
		false,
	); err != nil {
		return errors.Wrap(err, "clone tool repository")
	}
	if err := f.Chdir(tmpPath); err != nil {
		return errors.Wrap(err, "change directory")
	}
	return nil
}

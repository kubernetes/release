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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	kgit "k8s.io/release/pkg/git"
	"sigs.k8s.io/release-utils/util"
)

// Options is the main structure for configuring a fast forward.
type Options struct {
	Branch         string
	MainRef        string
	NonInteractive bool
	NoMock         bool
	Cleanup        bool
	RepoPath       string
	LogLevel       string
}

const pushUpstreamQuestion = `Are you ready to push the local branch fast-forward changes upstream?
Please only answer after you have validated the changes.`

func Run(opts *Options) error {
	branch := opts.Branch
	if branch == "" {
		return errors.New("please specify valid release branch")
	}

	logrus.Infof("Preparing to fast-forward %s onto the %s branch", opts.MainRef, branch)
	repo, err := kgit.CloneOrOpenDefaultGitHubRepoSSH(opts.RepoPath)
	if err != nil {
		return err
	}

	if !opts.NoMock {
		logrus.Info("Using dry mode, which does not modify any remote content")
		repo.SetDry()
	}

	logrus.Infof("Checking if %q is a release branch", branch)
	if isReleaseBranch := kgit.IsReleaseBranch(branch); !isReleaseBranch {
		return errors.Errorf("%s is not a release branch", branch)
	}

	logrus.Info("Checking if branch is available on the default remote")
	branchExists, err := repo.HasRemoteBranch(branch)
	if err != nil {
		return errors.Wrap(err, "checking if branch exists on the default remote")
	}
	if !branchExists {
		return errors.New("branch does not exist on the default remote")
	}

	if opts.Cleanup {
		defer repo.Cleanup() // nolint: errcheck
	} else {
		// Restore the currently checked out branch afterwards
		currentBranch, err := repo.CurrentBranch()
		if err != nil {
			return errors.Wrap(err, "unable to retrieve current branch")
		}
		defer func() {
			if err := repo.Checkout(currentBranch); err != nil {
				logrus.Errorf("Unable to restore branch %s: %v", currentBranch, err)
			}
		}()
	}

	logrus.Info("Checking out release branch")
	if err := repo.Checkout(branch); err != nil {
		return errors.Wrapf(err, "checking out branch %s", branch)
	}

	logrus.Infof("Finding merge base between %q and %q", kgit.DefaultBranch, branch)
	mergeBase, err := repo.MergeBase(kgit.DefaultBranch, branch)
	if err != nil {
		return err
	}

	// Verify the tags
	mainTag, err := repo.Describe(
		kgit.NewDescribeOptions().
			WithRevision(kgit.Remotify(kgit.DefaultBranch)).
			WithAbbrev(0).
			WithTags(),
	)
	if err != nil {
		return err
	}
	mergeBaseTag, err := repo.Describe(
		kgit.NewDescribeOptions().
			WithRevision(mergeBase).
			WithAbbrev(0).
			WithTags(),
	)
	if err != nil {
		return err
	}
	logrus.Infof("Merge base tag is: %s", mergeBaseTag)

	if mainTag != mergeBaseTag {
		return errors.Errorf(
			"unable to fast forward: tag %q does not match %q",
			mainTag, mergeBaseTag,
		)
	}
	logrus.Infof("Verified that the latest tag on the main branch is the same as the merge base tag")

	releaseRev, err := repo.Head()
	if err != nil {
		return err
	}
	logrus.Infof("Latest release branch revision is %s", releaseRev)

	logrus.Info("Merging main branch changes into release branch")
	if err := repo.Merge(opts.MainRef); err != nil {
		return err
	}

	headRev, err := repo.Head()
	if err != nil {
		return err
	}

	prepushMessage(repo.Dir(), branch, opts.MainRef, releaseRev, headRev)

	pushUpstream := false
	if opts.NonInteractive {
		pushUpstream = true
	} else {
		_, pushUpstream, err = util.Ask(pushUpstreamQuestion, "yes", 3)
		if err != nil {
			return err
		}
	}

	if pushUpstream {
		logrus.Infof("Pushing %s branch", branch)
		if err := repo.Push(branch); err != nil {
			return err
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
		kgit.Remotify(branch),
		ref,
		kgit.DefaultGithubOrg,
		kgit.DefaultGithubRepo,
		releaseRev[:11],
		headRev[:11],
	)
}

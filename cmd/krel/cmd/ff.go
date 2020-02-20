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

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	kgit "k8s.io/release/pkg/git"
	"k8s.io/release/pkg/util"
)

type ffOptions struct {
	branch         string
	masterRef      string
	nonInteractive bool
}

var ffOpts = &ffOptions{}

// ffCmd represents the base command when called without any subcommands
var ffCmd = &cobra.Command{
	Use:   "ff --branch <release-branch> [--ref <master-ref>] [--nomock] [--cleanup]",
	Short: "ff fast forwards a Kubernetes release branch",
	Long: fmt.Sprintf(`ff fast forwards a branch to a specified git object (defaults to %s).

krel ff pre-checks that the local branch to be forwarded is an actual
'release-x.y' branch and that the branch exists remotely. If that is not the
case, krel ff will fail.

After that preflight-check, the release branch will be checked out and krel
verifies that the latest merge base tag is the same for the master and the
release branch. This means that only the latest release branch can be fast
forwarded.

krel merges the provided ref into the release branch and asks for a final
confirmation if the push should really happen. The push will only be executed
as real push if the '--nomock' flag is specified.
`, kgit.Remotify(kgit.Master)),
	Example:       "krel ff --branch release-1.17 --ref origin/master --cleanup",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFf(ffOpts, rootOpts)
	},
}

const pushUpstreamQuestion = `Are you ready to push the local branch fast-forward changes upstream?
Please only answer after you have validated the changes.`

func init() {
	ffCmd.PersistentFlags().StringVar(&ffOpts.branch, "branch", "", "branch")
	ffCmd.PersistentFlags().StringVar(&ffOpts.masterRef, "ref", kgit.Remotify(kgit.Master), "ref on master")

	rootCmd.AddCommand(ffCmd)
}

func runFf(opts *ffOptions, rootOpts *rootOptions) error {
	branch := opts.branch
	if branch == "" {
		return errors.New("please specify valid release branch")
	}

	logrus.Infof("Preparing to fast-forward %s onto the %s branch", opts.masterRef, branch)
	repo, err := kgit.CloneOrOpenDefaultGitHubRepoSSH(rootOpts.repoPath)
	if err != nil {
		return err
	}

	if !rootOpts.nomock {
		logrus.Info("Using dry mode, which does not modify any remote content")
		repo.SetDry()
	}

	logrus.Infof("Checking if %q is a release branch", branch)
	if isReleaseBranch := kgit.IsReleaseBranch(branch); !isReleaseBranch {
		return errors.Errorf("%s is not a release branch", branch)
	}

	logrus.Info("Checking if branch is available on the default remote")
	if err := repo.HasRemoteBranch(branch); err != nil {
		return err
	}

	logrus.Info("Checking out release branch")
	if err := repo.Checkout(branch); err != nil {
		return err
	}

	cleanup := rootOpts.cleanup
	if cleanup {
		defer repo.Cleanup() // nolint: errcheck
	}

	logrus.Infof("Finding merge base between %q and %q", kgit.Master, branch)
	mergeBase, err := repo.MergeBase(kgit.Master, branch)
	if err != nil {
		return err
	}

	// Verify the tags
	masterTag, err := repo.DescribeTag(kgit.Remotify(kgit.Master))
	if err != nil {
		return err
	}
	mergeBaseTag, err := repo.DescribeTag(mergeBase)
	if err != nil {
		return err
	}
	logrus.Infof("Merge base tag is: %s", mergeBaseTag)

	if masterTag != mergeBaseTag {
		return errors.Errorf(
			"unable to fast forward: tag %q does not match %q",
			masterTag, mergeBaseTag,
		)
	}
	logrus.Infof("Verified that the latest master tag is the same as the merge base tag")

	releaseRev, err := repo.Head()
	if err != nil {
		return err
	}
	logrus.Infof("Latest release branch revision is %s", releaseRev)

	logrus.Info("Merging master changes into release branch")
	if err := repo.Merge(opts.masterRef); err != nil {
		return err
	}

	headRev, err := repo.Head()
	if err != nil {
		return err
	}

	prepushMessage(repo.Dir(), kgit.DefaultRemote, branch, kgit.DefaultGithubOrg, releaseRev, headRev)

	pushUpstream := false
	if opts.nonInteractive {
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

func prepushMessage(gitRoot, remote, branch, org, releaseRev, headRev string) {
	fmt.Printf(`Go look around in %s to make sure things look okay before pushing…

Check for files left uncommitted using:

	git status -s

Validate the fast-forward commit using:

	git show

Validate the changes pulled in from master using:

	git log %s/%s..HEAD

Once the branch fast-forward is complete, the diff will be available after push at:

	https://github.com/%s/%s/compare/%s...%s"

`, gitRoot, remote, branch, org, kgit.DefaultGithubRepo, releaseRev, headRev)
}

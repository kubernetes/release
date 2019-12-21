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
	branch    string
	masterRef string
	org       string
}

var ffOpts = &ffOptions{}

// ffCmd represents the base command when called without any subcommands
var ffCmd = &cobra.Command{
	Use:   "ff --branch <release-branch> [--ref <master-ref>] [--nomock] [--cleanup]",
	Short: "ff fast forwards a Kubernetes release branch",
	Long: `ff fast forwards a branch to a specified master object reference
(defaults to HEAD), and then prepares the branch as a Kubernetes release branch:

- Run hack/update-all.sh to ensure compliance of generated files`,
	Example:       "krel ff --branch release-1.17 --ref HEAD --cleanup",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE:       initLogging,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runFf(ffOpts)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	ffCmd.PersistentFlags().StringVar(&ffOpts.branch, "branch", "", "branch")
	ffCmd.PersistentFlags().StringVar(&ffOpts.masterRef, "ref", kgit.DefaultMasterRef, "ref on master")
	ffCmd.PersistentFlags().StringVar(&ffOpts.org, "org", kgit.DefaultGithubOrg, "org to run tool against")

	rootCmd.AddCommand(ffCmd)
}

func runFf(opts *ffOptions) error {
	branch := opts.branch
	if branch == "" {
		return errors.New("Please specify valid release branch")
	}
	masterRef := opts.masterRef
	remoteMaster := kgit.Remotify(kgit.Master)

	logrus.Infof("Preparing to fast-forward master@%s onto the %s branch", masterRef, branch)
	repo, err := kgit.CloneOrOpenDefaultGitHubRepoSSH(rootOpts.repoPath, opts.org)
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
	if err := repo.CheckoutBranch(branch); err != nil {
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
	masterTag, err := repo.DescribeTag(remoteMaster)
	if err != nil {
		return err
	}
	mergeBaseTag, err := repo.DescribeTag(mergeBase)
	if err != nil {
		return err
	}
	if masterTag != mergeBaseTag {
		return errors.Errorf(
			"Unable to fast forward: tag %q does not match %q",
			masterTag, mergeBaseTag,
		)
	}
	logrus.Infof("Last tag is: %s", masterTag)

	releaseRev, err := repo.Head()
	if err != nil {
		return err
	}
	logrus.Infof("Latest release branch revision is %s", releaseRev)

	logrus.Info("Merging master changes into release branch")
	if err := repo.Merge(remoteMaster); err != nil {
		return err
	}

	headRev, err := repo.Head()
	if err != nil {
		return err
	}

	prepushMessage(repo.Dir(), kgit.DefaultRemote, branch, opts.org, releaseRev, headRev)

	_, pushUpstream, err := util.Ask("Are you ready to push the local branch fast-forward changes upstream? Please only answer after you have validated the changes.", "yes", 3)
	if err != nil {
		return err
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
	fmt.Printf(`Go look around in %s to make sure things look okay before pushing...

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

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

package anago

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/regex"
)

// pushGitObjectsCmd is the krel push-git-objects subcommand
var pushGitObjectsCmd = &cobra.Command{
	Use:   "push-git-objects",
	Short: "Push branches and tags to the remote github repository",
	Long: `krel push-git-objects

NOTE: This subcommand is a temporary commando to be invoked from anago.

The purpose of krel push-git-objects is to push branches and tags to the
git remote repository of kubernetes. 
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runPushGitObjects(pushGitObjectsOpts); err != nil {
			return err
		}
		return nil
	},
}

type pushGitObjectsOptions struct {
	nomock        bool
	maxRetries    int
	releaseBranch string
	parentBranch  string
	repoPath      string
	tags          []string
}

var pushGitObjectsOpts = &pushGitObjectsOptions{}

func init() {
	pushGitObjectsCmd.PersistentFlags().StringVar(
		&pushGitObjectsOpts.releaseBranch,
		"release-branch",
		"",
		"branch name to push",
	)

	pushGitObjectsCmd.PersistentFlags().StringVar(
		&pushGitObjectsOpts.parentBranch,
		"parent-branch",
		"",
		fmt.Sprintf("parent branch, if different from %s", git.DefaultBranch),
	)

	pushGitObjectsCmd.PersistentFlags().BoolVar(
		&pushGitObjectsOpts.nomock,
		"nomock",
		false,
		"nomock flag",
	)

	pushGitObjectsCmd.PersistentFlags().StringSliceVarP(
		&pushGitObjectsOpts.tags,
		"tags",
		"t",
		[]string{},
		"list of tags to push",
	)

	pushGitObjectsCmd.PersistentFlags().StringVar(
		&pushGitObjectsOpts.repoPath,
		"repo",
		"",
		"the local path to the repository to be used",
	)

	pushGitObjectsCmd.PersistentFlags().IntVar(
		&pushGitObjectsOpts.maxRetries,
		"max-retries",
		10,
		"number of times to retry git operations if a recoverable error occurs",
	)

	AnagoCmd.AddCommand(pushGitObjectsCmd)
}

func runPushGitObjects(options *pushGitObjectsOptions) (err error) {
	if err := options.Validate(); err != nil {
		return errors.Wrap(err, "validating command line options")
	}
	// Create the git pusher object
	gitPusher, err := release.NewGitPusher(&release.GitObjectPusherOptions{
		DryRun:     !options.nomock,
		MaxRetries: options.maxRetries,
		RepoPath:   options.repoPath,
	})
	if err != nil {
		return errors.Wrap(err, "creating new git pusher object")
	}

	// # The real deal?
	nomockLabel := map[bool]string{true: "(nomock)", false: "(mock)"}

	// Tags are a range, push them all:
	logrus.Infof("Pushing %s %d tags", nomockLabel[options.nomock], len(options.tags))
	for _, tag := range options.tags {
		if err := gitPusher.PushTag(tag); err != nil {
			return err
		}
	}

	// if a release branch was specified, push it
	if isPushableBranch(options.releaseBranch) {
		logrus.Infof("Pushing %s %s branch:", nomockLabel[options.nomock], options.releaseBranch)
		if err := gitPusher.PushBranch(options.releaseBranch); err != nil {
			return errors.Wrapf(err, "pushing branch %s", options.releaseBranch)
		}

		// # Additionally push the parent branch if a branch of branch
		if isPushableBranch(options.parentBranch) {
			logrus.Infof("Pushing %s %s branch: ", nomockLabel[options.nomock], options.parentBranch)
			if err := gitPusher.PushBranch(options.parentBranch); err != nil {
				return errors.Wrapf(err, "pushing parent branch %s", options.parentBranch)
			}
		}
	}

	// For files created on master with new branches and
	// for $CHANGELOG_FILEPATH, update the master
	if err := gitPusher.PushMain(); err != nil {
		return errors.Wrap(err, "pushing changes in main branch")
	}

	logrus.Info("git objects push complete")
	return nil
}

// Validate checks if the passed options are correct
func (o *pushGitObjectsOptions) Validate() error {
	if len(o.tags) == 0 && o.releaseBranch == "" {
		return errors.New("to run push-git-objects, at least a branch or a tag has to be specified")
	}

	if o.parentBranch != "" && o.releaseBranch == "" {
		return errors.New("cannot specify a parent branch if no release branch is defined")
	}
	return nil
}

// isPushableBranch returns a bool indicating if the string in branchName
// corresponds to a branch that should be pushed
func isPushableBranch(branchName string) bool {
	if branchName == "" {
		return false
	}
	// Keeping "master" here to cover retroactively in case the default changes
	if branchName == git.DefaultBranch || branchName == "master" {
		return false
	}
	if regex.BranchRegex.Match([]byte(branchName)) {
		return true
	}
	return false
}

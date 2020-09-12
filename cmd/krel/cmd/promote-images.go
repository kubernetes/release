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

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/util"
)

const (
	k8sioRepo             = "k8s.io"
	k8sioManifestsPath    = "k8s.gcr.io"
	stagingRepo           = "k8s-staging-kubernetes"
	promotionBranchSuffix = "-image-promotion"
)

// promoteCommand is the krel subcommand to promote conainer images
var imagePromoteCommand = &cobra.Command{
	Use:   "promote-images",
	Short: "Starts an image promotion for a tag of kubernetes images",
	Long: `krel promote

The 'promote' subcommand of krel updates the image promoter manifests
ans creates a PR in kubernetes/k8s.io`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Run the PR creation function
		return runPromote(promoteOpts)
	},
}

type promoteOptions struct {
	userFork        string
	tag             string
	interactiveMode bool
}

func (o *promoteOptions) Validate() error {
	if o.tag == "" {
		return errors.New("cannot start promotion --tag is required")
	}
	if o.userFork == "" {
		return errors.New("cannot start promotion --fork is required")
	}

	// Check the fork slug
	if _, _, err := git.ParseRepoSlug(o.userFork); err != nil {
		return errors.Wrap(err, "checking user's fork")
	}

	// Verify we got a valid tag
	if _, err := util.TagStringToSemver(o.tag); err != nil {
		return errors.Wrapf(err, "verifying tag: %s", o.tag)
	}

	// Check that the GitHub token is set
	token, isset := os.LookupEnv(github.TokenEnvKey)
	if !isset || token == "" {
		return errors.New("cannot promote images if GitHub token is not set")
	}
	return nil
}

var promoteOpts = &promoteOptions{}

func init() {
	imagePromoteCommand.PersistentFlags().StringVarP(
		&promoteOpts.tag,
		"tag",
		"t",
		"",
		"version tag of the images we will promote",
	)

	imagePromoteCommand.PersistentFlags().StringVar(
		&promoteOpts.userFork,
		"fork",
		"",
		"the user's fork of kubernetes/k8s.io",
	)

	imagePromoteCommand.PersistentFlags().BoolVarP(
		&promoteOpts.interactiveMode,
		"interactive",
		"i",
		false,
		"interactive mode, asks before every step",
	)

	for _, flagName := range []string{"tag", "fork"} {
		if err := imagePromoteCommand.MarkPersistentFlagRequired(flagName); err != nil {
			logrus.Error(errors.Wrapf(err, "marking tag %s as required", flagName))
		}
	}

	rootCmd.AddCommand(imagePromoteCommand)
}

func runPromote(opts *promoteOptions) error {
	// Validate options
	branchname := opts.tag + promotionBranchSuffix

	// Check the cmd line opts
	if err := opts.Validate(); err != nil {
		return errors.Wrap(err, "checking command line options")
	}

	// Get the github org and repo from the fork slug
	userForkOrg, userForkRepo, err := git.ParseRepoSlug(opts.userFork)
	if err != nil {
		return errors.Wrap(err, "parsing user's fork")
	}
	if userForkRepo == "" {
		userForkRepo = k8sioRepo
	}

	// Check Environment
	gh := github.New()

	// Check for the cip-mm binary
	cipmm, err := exec.LookPath("cip-mm")
	if err != nil {
		return errors.Wrap(err, "while looking for cip-mm in your path")
	}

	// Verify the repository is a fork of k8s.io
	if err = verifyFork(
		branchname, userForkOrg, userForkRepo, git.DefaultGithubOrg, k8sioRepo,
	); err != nil {
		return errors.Wrapf(err, "while checking fork of %s/%s ", git.DefaultGithubOrg, k8sioRepo)
	}

	// Clone k8s.io
	repo, err := prepareFork(branchname, git.DefaultGithubOrg, k8sioRepo, userForkOrg, userForkRepo)
	if err != nil {
		return errors.Wrap(err, "while preparing k/k8s.io fork")
	}

	defer func() {
		if mustRun(opts, "Clean fork directory?") {
			err = repo.Cleanup()
		} else {
			logrus.Infof("All modified files will be left untouched in %s", repo.Dir())
		}
	}()

	// Run cip-mm
	if mustRun(opts, "Run cip-mm?") {
		if err := command.New(
			cipmm,
			fmt.Sprintf("--base_dir=%s", filepath.Join(repo.Dir(), k8sioManifestsPath)),
			fmt.Sprintf("--staging_repo=gcr.io/%s", stagingRepo),
			fmt.Sprintf("--filter_tag=%s", opts.tag),
		).RunSuccess(); err != nil {
			return errors.Wrap(err, "running cip-mm install in kubernetes-sigs/release-notes")
		}
	}

	// TODO: Either remove mock images or we fix them at the origin

	// TODO: verify that the manifest was actually modified

	// add the modified manifest to staging
	logrus.Debugf("Adding %s to staging area", filepath.Join(k8sioManifestsPath, "images", stagingRepo, "images.yaml"))
	if err := repo.Add(filepath.Join(k8sioManifestsPath, "images", stagingRepo, "images.yaml")); err != nil {
		return errors.Wrap(err, "adding image manifest to staging area")
	}

	commitMessage := fmt.Sprintf("releng: Image promotion for %s", opts.tag)

	// Commit files
	logrus.Debug("Creating commit")
	if err := repo.UserCommit(commitMessage); err != nil {
		return errors.Wrapf(err, "Error creating commit in %s/%s", git.DefaultGithubOrg, k8sioRepo)
	}

	// Push to fork
	if mustRun(opts, "Push changes to user's fork?") {
		logrus.Infof("Pushing manifest changes to %s/%s", userForkOrg, userForkRepo)
		if err := repo.PushToRemote(userForkName, branchname); err != nil {
			return errors.Wrapf(err, "pushing %s to %s/%s", userForkName, userForkOrg, userForkRepo)
		}
	}

	prBody := fmt.Sprintf("Image promotion for Kubernetes %s\n", opts.tag)
	prBody += "This is an automated PR generated from `krel The Kubernetes Release Toolbox`\n\n"

	// Create the Pull Request
	if mustRun(opts, "Create pull request?") {
		pr, err := gh.CreatePullRequest(
			git.DefaultGithubOrg, k8sioRepo, git.DefaultBranch,
			fmt.Sprintf("%s:%s", userForkOrg, branchname),
			commitMessage, prBody,
		)

		if err != nil {
			return errors.Wrap(err, "creating the pull request in k/k8s.io")
		}
		logrus.Infof(
			"Successfully created PR: %s%s/%s/pull/%d",
			github.GitHubURL, git.DefaultGithubOrg, k8sioRepo, pr.GetNumber(),
		)
	}

	// Success!
	return nil
}

// mustRun avoids running when a users chooses n in interactive mode
func mustRun(opts *promoteOptions, question string) bool {
	if !opts.interactiveMode {
		return true
	}
	_, success, err := util.Ask(fmt.Sprintf("%s (Y/n)", question), "y:Y:yes|n:N:no|y", 10)
	if err != nil {
		logrus.Error(err)
		if err.(util.UserInputError).IsCtrlC() {
			os.Exit(1)
		}
		return false
	}
	if success {
		return true
	}
	return false
}

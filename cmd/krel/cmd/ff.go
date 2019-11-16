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
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"k8s.io/release/pkg/command"
	kgit "k8s.io/release/pkg/git"
	"k8s.io/release/pkg/util"
)

var cfgFile string

const defaultMasterRef string = "HEAD"

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
	Example: "krel ff --branch release-1.16 39d0135e --ref HEAD --cleanup",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runFf(ffOpts); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	ffCmd.PersistentFlags().StringVar(&ffOpts.branch, "branch", "", "branch")
	ffCmd.PersistentFlags().StringVar(&ffOpts.masterRef, "ref", defaultMasterRef, "ref on master")
	ffCmd.PersistentFlags().StringVar(&ffOpts.org, "org", kgit.DefaultGithubOrg, "org to run tool against")

	rootCmd.AddCommand(ffCmd)
}

func runFf(opts *ffOptions) error {
	// TODO: Add usage/help
	// TODO: Set positional args
	// TODO: Fail on empty branch
	// TODO: Fail on GITHUB_TOKEN not set

	if !command.Available("git", "go", "make", "jq") {
		log.Fatalf("Unable to meet executable dependency requirements")
	}

	branch := opts.branch
	masterRef := opts.masterRef
	remote := kgit.DefaultRemote
	remoteMaster := fmt.Sprintf("%s/%s", remote, "master")

	log.Printf("Preparing to fast-forward master@%s onto the %s branch...\n", masterRef, branch)

	nomock := rootOpts.nomock
	dryRunFlag := "--dry-run"
	if nomock {
		// TODO: Set this to empty string once we're ready to turn on the tool
		log.Println("Running in no-mock mode!")
		dryRunFlag = "--dry-run"
	}

	// TODO: Refactor everything to use this repo instance
	repo, err := kgit.CloneOrOpenDefaultGitHubRepoSSH(rootOpts.repoPath)
	if err != nil {
		return err
	}

	isReleaseBranch := kgit.IsReleaseBranch(branch)
	if !isReleaseBranch {
		log.Fatalf("%s is not a release branch\n", branch)
	}

	if err := repo.HasRemoteBranch(branch); err != nil {
		return err
	}

	baseDir, err := ioutil.TempDir("", "ff-")
	if err != nil {
		return err
	}

	cleanup := rootOpts.cleanup
	if cleanup {
		defer cleanupTmpDir(baseDir)
	}

	workingDir := filepath.Join(baseDir, branch)
	log.Printf("%s", workingDir)

	os.Setenv("GOPATH", workingDir)
	log.Printf("GOPATH: %s", os.Getenv("GOPATH"))

	gitRoot := fmt.Sprintf("%s/src/k8s.io/kubernetes", workingDir)

	// TODO: nomock?
	if nomock {
		log.Printf("nomock mode (from within ff)\n")
	}

	// TODO: If workingDir exists, prompt user to delete
	// TODO: Tweak file permissions (dir + user rwx)
	workingDirErr := os.MkdirAll(workingDir, os.ModePerm)
	if workingDirErr != nil {
		return err
	}

	// TODO: Remove once SyncRepo works
	gitRoot = "/tmp/ff-465649664/release-1.16/src/k8s.io/kubernetes"

	syncErr := kgit.SyncRepo(kgit.KubernetesGitHubAuthURL, gitRoot)
	if syncErr != nil {
		return syncErr
	}

	chdirErr := os.Chdir(gitRoot)
	if chdirErr != nil {
		return chdirErr
	}

	mergebaseRepo, repoErr := git.PlainOpen(gitRoot)
	if repoErr != nil {
		return repoErr
	}

	mergeBase, err := kgit.GetMergeBase("master", branch, mergebaseRepo)
	if err != nil {
		return err
	}

	// TODO: Rewrite using go-git
	comparedCommits := []string{mergeBase, remoteMaster}
	var tags []string
	for _, commit := range comparedCommits {
		cmd := exec.Command("git", "describe", "--abbrev=0", "--tags", commit)
		cmd.Stdin = strings.NewReader("some input")
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("in all caps: %q\n", out.String())
		tags = append(tags, strings.TrimSuffix(out.String(), "\n"))
	}

	// TODO: This should return an error if it fails
	// TODO: Provide more information to debug here
	if tags[0] != tags[1] {
		log.Printf("%s did not match %s", tags[0], tags[1])
	}

	// TODO: Rewrite using go-git
	//       --dry-run appears to be unsupported for git push, so we shell out here.
	if err := command.Execute("git push -q --dry-run", kgit.KubernetesGitHubAuthURL); err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	// ... checking out to commit
	//Info("git checkout %s", commit)
	remoteHash, err := repo.ResolveRevision(plumbing.Revision(fmt.Sprintf("%s/%s", remote, branch)))
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash:   plumbing.NewHash(remoteHash.String()),
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	})

	if err != nil {
		return err
	}

	// TODO: Merge and update
	if err := command.Execute("git merge -X ours", remoteMaster); err != nil {
		return err
	}

	releaseRefName := remoteHash.String()
	releaseRev, err := kgit.RevParseShort(releaseRefName, gitRoot)
	if err != nil {
		return err
	}

	headRev, err := kgit.RevParseShort("HEAD", gitRoot)
	if err != nil {
		return err
	}

	log.Printf("%s", prepushMessage(gitRoot, remote, branch, kgit.KubernetesGitHubURL, releaseRev, headRev))

	_, pushUpstream, err := util.Ask("Are you ready to push the local branch fast-forward changes upstream? Please only answer after you have validated the changes.", "yes", 3)
	if err != nil {
		return err
	}

	if pushUpstream {
		log.Printf("Pushing %s %s branch upstream: ", dryRunFlag, branch)
		//git push $DRYRUN_FLAG origin $RELEASE_BRANCH:$RELEASE_BRANCH
		// TODO: Need to handle https and ssh auth sanely
		//fmt.Sprintf("%s:%s", branch, branch))
		if err := command.Execute("git push", dryRunFlag, remote, branch); err != nil {
			return err
		}
	}

	return nil
}

func prepushMessage(gitRoot, remote, branch, githubURL, releaseRev, headRev string) string {
	message := fmt.Sprintf(`Go look around in %s to make sure things look okay before pushing...

Check for files left uncommitted using:

	git status -s

Validate the fast-forward commit using:

	git show

Validate the changes pulled in from master using:

	git log %s/%s..HEAD

Once the branch fast-forward is complete, the diff will be available after push at:

	%s/compare/%s...%s"

`, gitRoot, remote, branch, githubURL, releaseRev, headRev)

	return message
}

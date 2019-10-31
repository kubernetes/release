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

package util

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const (
	branchRE = `master|release-([0-9]{1,})\.([0-9]{1,})(\.([0-9]{1,}))*$`
	// TODO: Use "kubernetes" as DefaultGithubOrg when this is ready to turn on
	DefaultGithubOrg      = "justaugustus"
	DefaultGithubAuthRoot = "git@github.com:"
	DefaultRemote         = "origin"
)

var (
	KubernetesGitHubURL     = fmt.Sprintf("https://github.com/%s/kubernetes", DefaultGithubOrg)
	KubernetesGitHubAuthURL = fmt.Sprintf("%s%s/kubernetes.git", DefaultGithubAuthRoot, DefaultGithubOrg)
)

func BranchExists(branch string) (bool, error) {
	log.Printf("Verifying %s branch exists on the remote...", branch)

	// Create the remote with repository URL
	rem := git.NewRemote(
		memory.NewStorage(),
		&config.RemoteConfig{
			Name: DefaultRemote,
			URLs: []string{KubernetesGitHubURL},
		},
	)

	log.Print("Fetching tags...")

	// We can then use every Remote functions to retrieve wanted information
	refs, err := rem.List(&git.ListOptions{})
	if err != nil {
		log.Printf("Could not list references on the remote repository.")
		return false, err
	}

	// Filters the references list and only keeps branches
	for _, ref := range refs {
		if ref.Name().IsBranch() {
			if ref.Name().Short() == branch {
				log.Printf("Found branch %s", ref.Name().Short())
				return true, nil
			}
		}
	}

	log.Printf("Could not find branch %s", branch)

	return false, nil
}

func IsReleaseBranch(branch string) bool {
	re := regexp.MustCompile(branchRE)
	if !re.MatchString(branch) {
		log.Fatalf("%s is not a release branch\n", branch)
		return false
	}

	return true
}

// TODO: Need to handle https and ssh auth sanely
func SyncRepo(repoURL, destination string) error {
	log.Printf("Syncing %s to %s:", repoURL, destination)

	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		chdirErr := os.Chdir(destination)
		if chdirErr != nil {
			return chdirErr
		}

		repo, repoErr := git.PlainOpen(destination)
		if repoErr != nil {
			return repoErr
		}

		w, err := repo.Worktree()
		if err != nil {
			return err
		}

		// ... checking out to commit
		//Info("git checkout %s", commit)
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName("master"),
		})
		if err != nil {
			return err
		}

		pullOpts := &git.PullOptions{RemoteName: DefaultRemote}
		err = w.Pull(pullOpts)
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return err
		}

		return nil
	}

	cloneOpts := &git.CloneOptions{
		URL: repoURL,
	}
	_, repoErr := git.PlainClone(destination, false, cloneOpts)
	if repoErr != nil {
		return repoErr
	}

	return nil
}

// RevParse parses a git revision and returns a SHA1 on success, otherwise an
// error.
func RevParse(rev, workDir string) (string, error) {
	repo, err := git.PlainOpen(workDir)
	if err != nil {
		return "", err
	}

	ref, err := repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", err
	}

	return ref.String(), nil
}

// RevParseShort parses a git revision and returns a SHA1 on success, otherwise an
// error.
func RevParseShort(rev, workDir string) (string, error) {
	fullRev, err := RevParse(rev, workDir)
	if err != nil {
		return "", err
	}

	shortRev := fullRev[:10]

	return shortRev, nil
}

func GetMergeBase(masterRefShort, releaseRefShort string, repo *git.Repository) (string, error) {
	masterRef := fmt.Sprintf("%s/%s", DefaultRemote, masterRefShort)
	releaseRef := fmt.Sprintf("%s/%s", DefaultRemote, releaseRefShort)

	log.Printf("masterRef: %s, releaseRef: %s", masterRef, releaseRef)

	commitRevs := []string{masterRef, releaseRef}
	var res []*object.Commit

	var hashes []*plumbing.Hash
	for _, rev := range commitRevs {
		hash, err := repo.ResolveRevision(plumbing.Revision(rev))
		if err != nil {
			return "", err
		}
		hashes = append(hashes, hash)
	}

	var commits []*object.Commit
	for _, hash := range hashes {
		commit, err := repo.CommitObject(*hash)
		if err != nil {
			return "", err
		}
		commits = append(commits, commit)
	}

	res, err := commits[0].MergeBase(commits[1])
	if err != nil {
		return "", err
	}

	if len(res) == 0 {
		return "", errors.Errorf("could not find a merge base between %s and %s", masterRefShort, releaseRefShort)
	}

	log.Printf("len commits %v\n", len(res))

	mergeBase := res[0].Hash.String()
	log.Printf("merge base is %s", mergeBase)

	return mergeBase, nil
}

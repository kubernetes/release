package rnd

import (
	"fmt"

	"github.com/blang/semver"

	"k8s.io/release/pkg/git"
)

func StorageWorkDir(releaseTagVersion semver.Version) string {
	return fmt.Sprintf("release-%d.%d", releaseTagVersion.Major, releaseTagVersion.Minor)
}

func CloneStorageRepo(repoSlug string) (*git.Repo, error) {
	repoOrg, repoName, err := git.ParseRepoSlug(repoSlug)
	if err != nil {
		return nil, err
	}

	storageGit, err := git.CleanCloneGitHubRepo(repoOrg, repoName, false)
	if err != nil {
		return nil, err
	}

	return storageGit, nil
}

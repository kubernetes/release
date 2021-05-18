/*
Copyright 2021 The Kubernetes Authors.

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

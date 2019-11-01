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

package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Wrapper type for a Kubernetes repository instance
type Repo struct {
	*git.Repository
	dir string
}

// CloneOrOpenRepo creates a temp directory containing the provided
// GitHub repository via owner and name.
//
// If a repoPath is given, then the function tries to update the repository.
//
// The function returns the repository if cloning or updating of the repository
// was successful, otherwise an error.
func CloneOrOpenRepo(repoPath, owner, name string) (*Repo, error) {
	targetDir := ""
	if repoPath != "" {
		_, err := os.Stat(repoPath)

		if err == nil {
			// The file or directory exists, just try to update the repo
			return updateRepo(repoPath)

		} else if os.IsNotExist(err) {
			// The directory does not exists, we still have to clone it
			targetDir = repoPath

		} else {
			// Something else bad happened
			return nil, err
		}

	} else {
		// No repoPath given, use a random temp dir instead
		t, err := ioutil.TempDir("", "k8s-")
		if err != nil {
			return nil, err
		}
		targetDir = t
	}

	r, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL:      fmt.Sprintf("https://github.com/%s/%s", owner, name),
		Progress: os.Stdout,
	})
	if err != nil {
		return nil, err
	}
	return &Repo{Repository: r, dir: targetDir}, nil
}

// updateRepo tries to open the provided repoPath and fetches the latest
// changed from the configured remote location
func updateRepo(repoPath string) (*Repo, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}
	err = r.Fetch(&git.FetchOptions{Progress: os.Stdout})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, err
	}
	return &Repo{Repository: r, dir: repoPath}, nil
}

func (r *Repo) Cleanup() error {
	log.Printf("Deleting %s...", r.dir)
	return os.RemoveAll(r.dir)
}

// RevParse parses a git revision and returns a SHA1 on success, otherwise an
// error.
func (r *Repo) RevParse(rev string) (string, error) {
	// Prefix all non-tags the default remote "origin"
	if isVersion, _ := regexp.MatchString(`v\d+\.\d+\.\d+.*`, rev); !isVersion {
		rev = "origin/" + rev
	}

	// Try to resolve the rev
	ref, err := r.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", err
	}

	return ref.String(), nil
}

// LatestNonPatchFinalToLatest tries to discover the start (latest v1.xx.0) and
// end (release-1.xx or master) revision inside the repository
func (r *Repo) LatestNonPatchFinalToLatest() (start, end string, err error) {
	// Find the last non patch version tag, then resolve its revision
	version, err := r.latestNonPatchFinalVersion()
	if err != nil {
		return "", "", err
	}
	versionTag := "v" + version.String()
	log.Printf("latest non patch version %s", versionTag)
	start, err = r.RevParse(versionTag)
	if err != nil {
		return "", "", err
	}

	// If a release branch exists for the next version, we use it. Otherwise we
	// fallback to the master branch.
	end, err = r.releaseBranchOrMasterRev(version.Major, version.Minor+1)
	if err != nil {
		return "", "", err
	}

	return start, end, nil
}

func (r *Repo) latestNonPatchFinalVersion() (semver.Version, error) {
	latestFinalTag := semver.Version{}

	tags, err := r.Tags()
	if err != nil {
		return latestFinalTag, err
	}

	found := false
	_ = tags.ForEach(func(t *plumbing.Reference) error {
		tag := strings.TrimPrefix(t.Name().Short(), "v")
		ver, err := semver.Make(tag)

		if err == nil {
			// We're searching for the latest, non patch final tag
			if ver.Patch == 0 && len(ver.Pre) == 0 {
				if ver.GT(latestFinalTag) {
					latestFinalTag = ver
					found = true
				}
			}
		}
		return nil
	})
	if !found {
		return latestFinalTag, fmt.Errorf("unable to find latest non patch release")
	}
	return latestFinalTag, nil
}

func (r *Repo) releaseBranchOrMasterRev(major, minor uint64) (rev string, err error) {
	relBranch := fmt.Sprintf("release-%d.%d", major, minor)
	rev, err = r.RevParse(relBranch)
	if err == nil {
		log.Printf("found release branch %s", relBranch)
		return rev, nil
	}

	rev, err = r.RevParse("master")
	if err == nil {
		log.Println("no release branch found, using master")
		return rev, nil
	}

	return "", err
}

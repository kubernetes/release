package notes

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Wrapper type for a Kubernetes repository instance
type KubernetesRepo struct{ *git.Repository }

// RevParse parses a git revision and returns a SHA1 on success, otherwise an
// error.
func (k *KubernetesRepo) RevParse(rev string) (string, error) {
	// Prefix all non-tags the default remote "origin"
	if isVersion, _ := regexp.MatchString(`v\d+\.\d+\.\d+.*`, rev); !isVersion {
		rev = "origin/" + rev
	}

	// Try to resolve the rev
	ref, err := k.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", err
	}

	return ref.String(), nil
}

// NewKubernetesRepo creates a temp directory containing the provided
// GitHub repository via owner and name.
//
// If a repoPath is given, then the function tries to update the repository.
//
// The funciton returns the repository if cloning or updating of the repository
// was successful, otherwise an error.
func NewKubernetesRepo(repoPath, owner, name string) (*KubernetesRepo, error) {
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
			// Something else bad happended
			return nil, err
		}

	} else {
		// No repoPath given, use a random temp dir instead
		t, err := ioutil.TempDir("", "release-notes")
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
	return &KubernetesRepo{r}, nil
}

// updateRepo tries to open the provided repoPath and fetches the latest
// changed from the configured remote location
func updateRepo(repoPath string) (*KubernetesRepo, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}
	err = r.Fetch(&git.FetchOptions{Progress: os.Stdout})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, err
	}
	return &KubernetesRepo{r}, nil
}

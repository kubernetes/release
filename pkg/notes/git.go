package notes

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

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

// CloneTempRepository creates a temp directory containing the provided
// GitHub repository via owner and name. It returns that directory if cloning
// of the repository was successful, otherwise an error.
func CloneTempRepository(owner, name string) (string, error) {
	dir, err := ioutil.TempDir("", "release-notes")
	if err != nil {
		return "", err
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:      fmt.Sprintf("https://github.com/%s/%s", owner, name),
		Progress: os.Stdout,
	})
	if err != nil {
		return "", err
	}

	return dir, nil
}

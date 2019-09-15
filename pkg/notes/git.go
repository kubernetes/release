package notes

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
// The function returns the repository if cloning or updating of the repository
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
			// Something else bad happened
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

func (k *KubernetesRepo) DiscoverRevs(logger log.Logger) (start, end string, err error) {
	// Find the last non patch version tag, then resolve its revision
	version, err := k.latestNonPatchFinalVersion()
	if err != nil {
		return "", "", err
	}
	versionTag := "v" + version.String()
	level.Info(logger).Log("msg", "latest non patch version "+versionTag)
	start, err = k.RevParse(versionTag)
	if err != nil {
		return "", "", err
	}

	// If a release branch exists for the next version, we use it. Otherwise we
	// fallback to the master branch.
	end, err = k.releaseBranchOrMasterRev(logger, version.Major, version.Minor+1)
	if err != nil {
		return "", "", err
	}

	return start, end, nil
}

func (k *KubernetesRepo) latestNonPatchFinalVersion() (semver.Version, error) {
	latestFinalTag := semver.Version{}

	tags, err := k.Tags()
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

func (k *KubernetesRepo) releaseBranchOrMasterRev(logger log.Logger, major, minor uint64) (rev string, err error) {
	relBranch := fmt.Sprintf("release-%d.%d", major, minor)
	rev, err = k.RevParse(relBranch)
	if err == nil {
		level.Info(logger).Log("msg", "found release branch "+relBranch)
		return rev, nil
	}

	rev, err = k.RevParse("master")
	if err == nil {
		level.Info(logger).Log("msg", "no release branch found, using master")
		return rev, nil
	}

	return "", err
}

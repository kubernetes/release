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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"

	"k8s.io/release/pkg/command"
)

const (
	DefaultGithubOrg  = "kubernetes"
	DefaultGithubRepo = "kubernetes"
	DefaultRemote     = "origin"
	DefaultMasterRef  = "HEAD"
	Master            = "master"

	branchRE              = `master|release-([0-9]{1,})\.([0-9]{1,})(\.([0-9]{1,}))*$`
	defaultGithubAuthRoot = "git@github.com:"
	gitExecutable         = "git"
	tagPrefix             = "v"
)

// DiscoverResult is the result of a revision discovery
type DiscoverResult struct {
	startSHA, startRev, endSHA, endRev string
}

// StartSHA returns the start SHA for the DiscoverResult
func (d *DiscoverResult) StartSHA() string {
	return d.startSHA
}

// StartRev returns the start revision for the DiscoverResult
func (d *DiscoverResult) StartRev() string {
	return d.startRev
}

// EndSHA returns the end SHA for the DiscoverResult
func (d *DiscoverResult) EndSHA() string {
	return d.endSHA
}

// EndRev returns the end revision for the DiscoverResult
func (d *DiscoverResult) EndRev() string {
	return d.endRev
}

// Wrapper type for a Kubernetes repository instance
type Repo struct {
	inner  *git.Repository
	auth   transport.AuthMethod
	dir    string
	dryRun bool
}

// Dir returns the directory where the repository is stored on disk
func (r *Repo) Dir() string {
	return r.dir
}

// Set the repo into dry run mode, which does not modify any remote locations
// at all.
func (r *Repo) SetDry() {
	r.dryRun = true
}

// CloneOrOpenDefaultGitHubRepoSSH clones the default Kubernets GitHub
// repository into the path or updates it.
func CloneOrOpenDefaultGitHubRepoSSH(path, owner string) (*Repo, error) {
	return CloneOrOpenGitHubRepo(path, owner, DefaultGithubRepo, true)
}

// CloneOrOpenGitHubRepo creates a temp directory containing the provided
// GitHub repository via the owner and repo. If useSSH is true, then it will
// clone the repository using the defaultGithubAuthRoot.
func CloneOrOpenGitHubRepo(path, owner, repo string, useSSH bool) (*Repo, error) {
	return CloneOrOpenRepo(
		path,
		func() string {
			slug := fmt.Sprintf("%s/%s", owner, repo)
			if useSSH {
				return defaultGithubAuthRoot + slug
			}
			return fmt.Sprintf("https://github.com/%s", slug)
		}(),
		useSSH,
	)
}

// CloneOrOpenRepo creates a temp directory containing the provided
// GitHub repository via the url.
//
// If a repoPath is given, then the function tries to update the repository.
//
// The function returns the repository if cloning or updating of the repository
// was successful, otherwise an error.
func CloneOrOpenRepo(repoPath, url string, useSSH bool) (*Repo, error) {
	// We still need the plain git executable for some methods
	if !command.Available(gitExecutable) {
		return nil, errors.New("git is needed to support all repository features")
	}

	logrus.Debugf("Using repository url %q", url)
	targetDir := ""
	if repoPath != "" {
		logrus.Debugf("Using existing repository path %q", repoPath)
		_, err := os.Stat(repoPath)

		if err == nil {
			// The file or directory exists, just try to update the repo
			return updateRepo(repoPath, useSSH)
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

	if _, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	}); err != nil {
		return nil, err
	}
	return updateRepo(targetDir, useSSH)
}

// updateRepo tries to open the provided repoPath and fetches the latest
// changed from the configured remote location
func updateRepo(repoPath string, useSSH bool) (*Repo, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	var auth transport.AuthMethod
	if useSSH {
		auth, err = ssh.NewPublicKeysFromFile(gitExecutable,
			filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "")
		if err != nil {
			return nil, err
		}
	}

	err = r.Fetch(&git.FetchOptions{
		Auth:     auth,
		Force:    true,
		Progress: os.Stdout,
		RefSpecs: []config.RefSpec{"refs/*:refs/*"},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, err
	}
	return &Repo{inner: r, auth: auth, dir: repoPath}, nil
}

func (r *Repo) Cleanup() error {
	logrus.Debugf("Deleting %s", r.dir)
	return os.RemoveAll(r.dir)
}

// RevParse parses a git revision and returns a SHA1 on success, otherwise an
// error.
func (r *Repo) RevParse(rev string) (string, error) {
	matched, err := regexp.MatchString(`v\d+\.\d+\.\d+.*`, rev)
	if err != nil {
		return "", err
	}
	if !matched {
		// Prefix all non-tags the default remote "origin"
		rev = Remotify(rev)
	}

	// Try to resolve the rev
	ref, err := r.inner.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", err
	}

	return ref.String(), nil
}

// RevParseShort parses a git revision and returns a SHA1 trimmed to the length
// 10 on success, otherwise an error.
func (r *Repo) RevParseShort(rev string) (string, error) {
	fullRev, err := r.RevParse(rev)
	if err != nil {
		return "", err
	}

	return fullRev[:10], nil
}

// LatestNonPatchFinalToLatest tries to discover the start (latest v1.xx.0) and
// end (release-1.xx or master) revision inside the repository
func (r *Repo) LatestNonPatchFinalToLatest() (DiscoverResult, error) {
	// Find the last non patch version tag, then resolve its revision
	versions, err := r.latestNonPatchFinalVersions()
	if err != nil {
		return DiscoverResult{}, err
	}
	version := versions[0]
	versionTag := addTagPrefix(version.String())
	logrus.Debugf("latest non patch version %s", versionTag)
	start, err := r.RevParse(versionTag)
	if err != nil {
		return DiscoverResult{}, err
	}

	// If a release branch exists for the next version, we use it. Otherwise we
	// fallback to the master branch.
	end, branch, err := r.releaseBranchOrMasterRev(version.Major, version.Minor+1)
	if err != nil {
		return DiscoverResult{}, err
	}

	return DiscoverResult{
		startSHA: start,
		startRev: versionTag,
		endSHA:   end,
		endRev:   branch,
	}, nil
}

func (r *Repo) LatestNonPatchFinalToMinor() (DiscoverResult, error) {
	// Find the last non patch version tag, then resolve its revision
	versions, err := r.latestNonPatchFinalVersions()
	if err != nil {
		return DiscoverResult{}, err
	}
	if len(versions) < 2 {
		return DiscoverResult{}, errors.New("unable to find two latest non patch versions")
	}

	latestVersion := versions[0]
	latestVersionTag := addTagPrefix(latestVersion.String())
	logrus.Debugf("latest non patch version %s", latestVersionTag)
	end, err := r.RevParse(latestVersionTag)
	if err != nil {
		return DiscoverResult{}, err
	}

	previousVersion := versions[1]
	previousVersionTag := addTagPrefix(previousVersion.String())
	logrus.Debugf("previous non patch version %s", previousVersionTag)
	start, err := r.RevParse(previousVersionTag)
	if err != nil {
		return DiscoverResult{}, err
	}

	return DiscoverResult{
		startSHA: start,
		startRev: previousVersionTag,
		endSHA:   end,
		endRev:   latestVersionTag,
	}, nil
}

func (r *Repo) latestNonPatchFinalVersions() ([]semver.Version, error) {
	latestVersions := []semver.Version{}

	tags, err := r.inner.Tags()
	if err != nil {
		return nil, err
	}

	_ = tags.ForEach(func(t *plumbing.Reference) error { // nolint: errcheck
		tag := trimTagPrefix(t.Name().Short())
		ver, err := semver.Make(tag)

		if err == nil {
			// We're searching for the latest, non patch final tag
			if ver.Patch == 0 && len(ver.Pre) == 0 {
				if len(latestVersions) == 0 || ver.GT(latestVersions[0]) {
					latestVersions = append([]semver.Version{ver}, latestVersions...)
				}
			}
		}
		return nil
	})
	if len(latestVersions) == 0 {
		return nil, fmt.Errorf("unable to find latest non patch release")
	}
	return latestVersions, nil
}

func (r *Repo) releaseBranchOrMasterRev(major, minor uint64) (sha, rev string, err error) {
	relBranch := fmt.Sprintf("release-%d.%d", major, minor)
	sha, err = r.RevParse(relBranch)
	if err == nil {
		logrus.Debugf("found release branch %s", relBranch)
		return sha, relBranch, nil
	}

	sha, err = r.RevParse(Master)
	if err == nil {
		logrus.Debug("no release branch found, using master")
		return sha, Master, nil
	}

	return "", "", err
}

// HasRemoteBranch takes a branch string and verifies that it exists
// on the default remote
func (r *Repo) HasRemoteBranch(branch string) error {
	logrus.Infof("Verifying %s branch exists on the remote", branch)

	remote, err := r.inner.Remote(DefaultRemote)
	if err != nil {
		return err
	}

	// We can then use every Remote functions to retrieve wanted information
	refs, err := remote.List(&git.ListOptions{Auth: r.auth})
	if err != nil {
		logrus.Warn("Could not list references on the remote repository.")
		return err
	}

	for _, ref := range refs {
		if ref.Name().IsBranch() {
			if ref.Name().Short() == branch {
				logrus.Infof("Found branch %s", ref.Name().Short())
				return nil
			}
		}
	}
	return errors.Errorf("branch %v not found", branch)
}

// CheckoutBranch can be used to switch to another branch
func (r *Repo) CheckoutBranch(name string) error {
	worktree, err := r.inner.Worktree()
	if err != nil {
		return err
	}

	return worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(name),
		Force:  true,
	})
}

func IsReleaseBranch(branch string) bool {
	re := regexp.MustCompile(branchRE)
	if !re.MatchString(branch) {
		logrus.Warnf("%s is not a release branch", branch)
		return false
	}

	return true
}

func (r *Repo) MergeBase(from, to string) (string, error) {
	masterRef := Remotify(from)
	releaseRef := Remotify(to)

	logrus.Debugf("masterRef: %s, releaseRef: %s", masterRef, releaseRef)

	commitRevs := []string{masterRef, releaseRef}
	var res []*object.Commit

	hashes := []*plumbing.Hash{}
	for _, rev := range commitRevs {
		hash, err := r.inner.ResolveRevision(plumbing.Revision(rev))
		if err != nil {
			return "", err
		}
		hashes = append(hashes, hash)
	}

	commits := []*object.Commit{}
	for _, hash := range hashes {
		commit, err := r.inner.CommitObject(*hash)
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
		return "", errors.Errorf("could not find a merge base between %s and %s", from, to)
	}

	mergeBase := res[0].Hash.String()
	logrus.Infof("merge base is %s", mergeBase)

	return mergeBase, nil
}

// Remotify returns the name prepended with the default remote
func Remotify(name string) string {
	return fmt.Sprintf("%s/%s", DefaultRemote, name)
}

// DescribeTag can be used to retrieve the latest tag for a provided revision
func (r *Repo) DescribeTag(rev string) (string, error) {
	// go git seems to have no implementation for `git describe`
	// which means that we fallback to a shell command for sake of
	// simplicity
	status, err := command.NewWithWorkDir(
		r.Dir(), gitExecutable, "describe", "--abbrev=0", "--tags", rev,
	).RunSilent()
	if err != nil {
		return "", err
	}
	if !status.Success() {
		return "", errors.Errorf("git describe command failed: %s", status.Error())
	}

	return strings.TrimSpace(status.Output()), nil
}

// Merge does a git merge into the current branch from the provided one
func (r *Repo) Merge(from string) error {
	return command.NewWithWorkDir(
		r.Dir(), gitExecutable, "merge", "-X", "ours", from,
	).RunSuccess()
}

// Push does push the specified branch to the default remote, but only if the
// repository is not in dry run mode
func (r *Repo) Push(remoteBranch string) error {
	args := []string{"push"}
	if r.dryRun {
		logrus.Infof("Won't push due to dry run repository")
		args = append(args, "--dry-run")
	}
	args = append(args, DefaultRemote, remoteBranch)

	return command.NewWithWorkDir(r.Dir(), gitExecutable, args...).RunSuccess()
}

// Head retrieves the current repository HEAD as a string
func (r *Repo) Head() (string, error) {
	ref, err := r.inner.Head()
	if err != nil {
		return "", err
	}
	return ref.Hash().String(), nil
}

// LatestPatchToPatch tries to discover the start (latest v1.x.[x-1]) and
// end (latest v1.x.x) revision inside the repository for the specified release
// branch.
func (r *Repo) LatestPatchToPatch(branch string) (DiscoverResult, error) {
	latestTag, err := r.latestTagForBranch(branch)
	if err != nil {
		return DiscoverResult{}, err
	}

	if len(latestTag.Pre) > 0 && latestTag.Patch > 0 {
		latestTag.Patch--
		latestTag.Pre = nil
	}

	if latestTag.Patch == 0 {
		return DiscoverResult{}, errors.Errorf(
			"found non-patch version %v as latest tag on branch %s",
			latestTag, branch,
		)
	}

	prevTag := semver.Version{
		Major: latestTag.Major,
		Minor: latestTag.Minor,
		Patch: latestTag.Patch - 1,
	}

	logrus.Debugf("parsing latest tag %s%v", tagPrefix, latestTag)
	latestVersionTag := addTagPrefix(latestTag.String())
	end, err := r.RevParse(latestVersionTag)
	if err != nil {
		return DiscoverResult{}, errors.Wrapf(err, "parsing version %v", latestTag)
	}

	logrus.Debugf("parsing previous tag %s%v", tagPrefix, prevTag)
	previousVersionTag := addTagPrefix(prevTag.String())
	start, err := r.RevParse(previousVersionTag)
	if err != nil {
		return DiscoverResult{}, errors.Wrapf(err, "parsing previous version %v", prevTag)
	}

	return DiscoverResult{
		startSHA: start,
		startRev: previousVersionTag,
		endSHA:   end,
		endRev:   latestVersionTag,
	}, nil
}

// latestTagForBranch returns the latest available semver tag for a given branch
func (r *Repo) latestTagForBranch(branch string) (*semver.Version, error) {
	status, err := command.NewWithWorkDir(
		r.Dir(), gitExecutable, "rev-list", branch, "--tags", "--max-count=1",
	).RunSilent()
	if err != nil {
		return nil, err
	}
	if !status.Success() {
		return nil, errors.Errorf("git rev-list command failed: %s", status.Error())
	}

	rev, err := r.DescribeTag(strings.TrimSpace(status.Output()))
	if err != nil {
		return nil, err
	}

	tag, err := semver.Make(trimTagPrefix(rev))
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

func addTagPrefix(tag string) string {
	return tagPrefix + tag
}

func trimTagPrefix(tag string) string {
	return strings.TrimPrefix(tag, tagPrefix)
}

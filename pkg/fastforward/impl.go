/*
Copyright 2022 The Kubernetes Authors.

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

package fastforward

import (
	"os"

	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/release"

	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/env"
	"sigs.k8s.io/release-utils/util"
)

type defaultImpl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . impl
type impl interface {
	CloneOrOpenDefaultGitHubRepoSSH(string) (*git.Repo, error)
	RepoSetDry(*git.Repo)
	IsReleaseBranch(string) bool
	RepoHasRemoteBranch(*git.Repo, string) (bool, error)
	RepoCleanup(*git.Repo) error
	RepoCurrentBranch(*git.Repo) (string, error)
	RepoCheckout(*git.Repo, string, ...string) error
	RepoMergeBase(*git.Repo, string, string) (string, error)
	RepoDescribe(*git.Repo, *git.DescribeOptions) (string, error)
	RepoHead(*git.Repo) (string, error)
	RepoMerge(*git.Repo, string) error
	RepoDir(*git.Repo) string
	Ask(string, string, int) (string, bool, error)
	RepoPush(*git.Repo, string) error
	RepoLatestReleaseBranch(*git.Repo) (string, error)
	RepoHasRemoteTag(*git.Repo, string) (bool, error)
	Submit(*gcb.Options) error
	EnvDefault(string, string) string
	CloneOrOpenGitHubRepo(string, string, string, bool) (*git.Repo, error)
	IsDefaultK8sUpstream() bool
	RepoSetURL(*git.Repo, string, string) error
	Chdir(string) error
	RemoveAll(string) error
	MkdirTemp(string, string) (string, error)
	Exists(string) bool
}

func (*defaultImpl) CloneOrOpenDefaultGitHubRepoSSH(repo string) (*git.Repo, error) {
	return git.CloneOrOpenDefaultGitHubRepoSSH(repo)
}

func (*defaultImpl) RepoSetDry(r *git.Repo) {
	r.SetDry()
}

func (*defaultImpl) IsReleaseBranch(branch string) bool {
	return git.IsReleaseBranch(branch)
}

func (*defaultImpl) RepoHasRemoteBranch(r *git.Repo, branch string) (branchExists bool, err error) {
	return r.HasRemoteBranch(branch)
}

func (*defaultImpl) RepoCleanup(r *git.Repo) error {
	return r.Cleanup()
}

func (*defaultImpl) RepoCurrentBranch(r *git.Repo) (string, error) {
	return r.CurrentBranch()
}

func (*defaultImpl) RepoCheckout(r *git.Repo, rev string, args ...string) error {
	return r.Checkout(rev, args...)
}

func (*defaultImpl) RepoMergeBase(r *git.Repo, from, to string) (string, error) {
	return r.MergeBase(from, to)
}

func (*defaultImpl) RepoDescribe(r *git.Repo, options *git.DescribeOptions) (string, error) {
	return r.Describe(options)
}

func (*defaultImpl) RepoHead(r *git.Repo) (string, error) {
	return r.Head()
}

func (*defaultImpl) RepoMerge(r *git.Repo, from string) error {
	return r.Merge(from)
}

func (*defaultImpl) RepoDir(r *git.Repo) string {
	return r.Dir()
}

func (*defaultImpl) Ask(q, e string, r int) (answer string, success bool, err error) {
	return util.Ask(q, e, r)
}

func (*defaultImpl) RepoPush(r *git.Repo, remoteBranch string) error {
	return r.Push(remoteBranch)
}

func (*defaultImpl) RepoLatestReleaseBranch(r *git.Repo) (string, error) {
	return r.LatestReleaseBranch()
}

func (*defaultImpl) RepoHasRemoteTag(r *git.Repo, tag string) (bool, error) {
	return r.HasRemoteTag(tag)
}

func (*defaultImpl) Submit(options *gcb.Options) error {
	return gcb.New(options).Submit()
}

func (*defaultImpl) EnvDefault(key, def string) string {
	return env.Default(key, def)
}

func (*defaultImpl) CloneOrOpenGitHubRepo(repoPath, owner, repo string, useSSH bool) (*git.Repo, error) {
	return git.CloneOrOpenGitHubRepo(repoPath, owner, repo, useSSH)
}

func (*defaultImpl) IsDefaultK8sUpstream() bool {
	return release.IsDefaultK8sUpstream()
}

func (*defaultImpl) RepoSetURL(r *git.Repo, remote, newURL string) error {
	return r.SetURL(remote, newURL)
}

func (*defaultImpl) Chdir(dir string) error {
	return os.Chdir(dir)
}

func (*defaultImpl) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (*defaultImpl) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

func (*defaultImpl) Exists(path string) bool {
	return util.Exists(path)
}

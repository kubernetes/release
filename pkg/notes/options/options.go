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

package options

import (
	"context"
	"os"

	"github.com/google/go-github/v28/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/notes/client"
)

type Options struct {
	GithubToken     string
	GithubOrg       string
	GithubRepo      string
	Output          string
	Branch          string
	StartSHA        string
	EndSHA          string
	StartRev        string
	EndRev          string
	RepoPath        string
	ReleaseVersion  string
	Format          string
	RequiredAuthor  string
	DiscoverMode    string
	ReleaseBucket   string
	ReleaseTars     string
	TableOfContents bool
	Debug           bool
	RecordDir       string
	ReplayDir       string
	gitCloneFn      func(string, string, string, bool) (*git.Repo, error)
}

type RevisionDiscoveryMode string

const (
	RevisionDiscoveryModeNONE              = "none"
	RevisionDiscoveryModeMergeBaseToLatest = "mergebase-to-latest"
	RevisionDiscoveryModePatchToPatch      = "patch-to-patch"
	RevisionDiscoveryModeMinorToMinor      = "minor-to-minor"
)

// New creates a new Options instance with the default values
func New() *Options {
	return &Options{
		DiscoverMode: RevisionDiscoveryModeNONE,
		GithubOrg:    git.DefaultGithubOrg,
		GithubRepo:   git.DefaultGithubRepo,
		gitCloneFn:   git.CloneOrOpenGitHubRepo,
	}
}

// ValidateAndFinish checks if the options are set in a consistent way and
// adapts them if necessary. It returns an error if options are set to invalid
// values.
func (o *Options) ValidateAndFinish() error {
	// Add appropriate log filtering
	if o.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if o.ReplayDir != "" && o.RecordDir != "" {
		return errors.New("please do not use record and replay together")
	}

	// Recover for replay if needed
	if o.ReplayDir != "" {
		logrus.Info("using replay mode")
		return nil
	}

	// The GitHub Token is required.
	if o.GithubToken == "" {
		return errors.New("GitHub token must be set via -github-token or $GITHUB_TOKEN")
	}

	// Check if we want to automatically discover the revisions
	if o.DiscoverMode != RevisionDiscoveryModeNONE {
		repo, err := o.gitCloneFn(
			o.RepoPath,
			o.GithubOrg,
			o.GithubRepo,
			false,
		)
		if err != nil {
			return err
		}

		var result git.DiscoverResult
		if o.DiscoverMode == RevisionDiscoveryModeMergeBaseToLatest {
			result, err = repo.LatestReleaseBranchMergeBaseToLatest()
		} else if o.DiscoverMode == RevisionDiscoveryModePatchToPatch {
			result, err = repo.LatestPatchToPatch(o.Branch)
		} else if o.DiscoverMode == RevisionDiscoveryModeMinorToMinor {
			result, err = repo.LatestNonPatchFinalToMinor()
		}
		if err != nil {
			return err
		}

		o.StartSHA = result.StartSHA()
		o.StartRev = result.StartRev()
		o.EndSHA = result.EndSHA()
		o.EndRev = result.EndRev()

		logrus.Infof("discovered start SHA %s", o.StartSHA)
		logrus.Infof("discovered end SHA %s", o.EndSHA)

		logrus.Infof("using start revision %s", o.StartRev)
		logrus.Infof("using end revision %s", o.EndRev)
	}

	// The start SHA or rev is required.
	if o.StartSHA == "" && o.StartRev == "" {
		return errors.New("the starting commit hash must be set via -start-sha, $START_SHA, -start-rev or $START_REV")
	}

	// The end SHA or rev is required.
	if o.EndSHA == "" && o.EndRev == "" {
		return errors.New("the ending commit hash must be set via -end-sha, $END_SHA, -end-rev or $END_REV")
	}

	// Check if we have to parse a revision
	if (o.StartRev != "" && o.StartSHA == "") || (o.EndRev != "" && o.EndSHA == "") {
		logrus.Info("cloning/updating repository to discover start or end sha")
		repo, err := o.gitCloneFn(
			o.RepoPath,
			o.GithubOrg,
			o.GithubRepo,
			false,
		)
		if err != nil {
			return err
		}
		if o.StartRev != "" {
			sha, err := repo.RevParse(o.StartRev)
			if err != nil {
				return err
			}
			logrus.Infof("using found start SHA: %s", sha)
			o.StartSHA = sha
		}
		if o.EndRev != "" {
			sha, err := repo.RevParse(o.EndRev)
			if err != nil {
				return err
			}
			logrus.Infof("using found end SHA: %s", sha)
			o.EndSHA = sha
		}
	}

	// Create the record dir
	if o.RecordDir != "" {
		logrus.Info("using record mode")
		if err := os.MkdirAll(o.RecordDir, 0o755); err != nil {
			return err
		}
	}

	return nil
}

// Client returns a Client to be used by the Gatherer. Depending on
// the provided options this is either a real client talking to the GitHub API,
// a Client which in addition records the responses from Github and stores them
// on disk, or a Client that replays those pre-recorded responses and does not
// talk to the GitHub API at all.
func (o *Options) Client() client.Client {
	if o.ReplayDir != "" {
		return client.NewReplayer(o.ReplayDir)
	}

	// Create a real GitHub API client
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: o.GithubToken},
	))
	c := client.New(github.NewClient(httpClient))

	if o.RecordDir != "" {
		return client.NewRecorder(c, o.RecordDir)
	}

	return c
}

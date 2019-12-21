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

package notes

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/git"
)

type Options struct {
	GithubToken    string
	GithubOrg      string
	GithubRepo     string
	Output         string
	Branch         string
	StartSHA       string
	EndSHA         string
	StartRev       string
	EndRev         string
	RepoPath       string
	ReleaseVersion string
	Format         string
	RequiredAuthor string
	Debug          bool
	DiscoverMode   string
	ReleaseBucket  string
	ReleaseTars    string
	gitCloneFn     func(string, string, string, bool) (*git.Repo, error)
}

type RevisionDiscoveryMode string

const (
	RevisionDiscoveryModeNONE          = "none"
	RevisionDiscoveryModeMinorToLatest = "minor-to-latest"
	RevisionDiscoveryModePatchToPatch  = "patch-to-patch"
	RevisionDiscoveryModeMinorToMinor  = "minor-to-minor"
)

// NewOptions creates a new Options instance with the default values
func NewOptions() *Options {
	return &Options{
		DiscoverMode: RevisionDiscoveryModeNONE,
		gitCloneFn:   git.CloneOrOpenGitHubRepo,
	}
}

// ValidateAndFinish checks if the options are set in a consistent way and
// adapts them if necessary. It returns an error if options are set to invalid
// values.
func (o *Options) ValidateAndFinish() error {
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
		if o.DiscoverMode == RevisionDiscoveryModeMinorToLatest {
			result, err = repo.LatestNonPatchFinalToLatest()
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

		logrus.Infof("discovered start %s (%s)", o.StartRev, o.StartSHA)
		logrus.Infof("discovered end %s (%s)", o.EndRev, o.EndSHA)
	}

	// The start SHA is required.
	if o.StartSHA == "" && o.StartRev == "" {
		return errors.New("the starting commit hash must be set via -start-sha, $START_SHA, -start-rev or $START_REV")
	}

	// The end SHA is required.
	if o.EndSHA == "" && o.EndRev == "" {
		return errors.New("the ending commit hash must be set via -end-sha, $END_SHA, -end-rev or $END_REV")
	}

	// Check if we have to parse a revision
	if o.StartRev != "" || o.EndRev != "" {
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

	// Add appropriate log filtering
	if o.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	return nil
}

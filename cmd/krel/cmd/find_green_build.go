/*
Copyright 2020 The Kubernetes Authors.

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

package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

// findGreenBuildCmd represents the subcommand for `krel find-green-build`
var findGreenBuildCmd = &cobra.Command{
	Use:   "find-green-build",
	Short: "Find a good Kubernetes build from Jenkins",
	Long: fmt.Sprintf(`krel find-green-build

This subcommand tries to receive the last successful build result for the
%s blocking build jobs. If the job has been found it will automatically
calculate the next build version based on the provided flags.

It is possible to exclude some tests suites by specifying them via the
'--exclude-suites' flag. The '--branch' flag should specify for which branch
we're planning to retrieve the latest build. Beside that, a '--type' has to be
provided. 

The subcommand also checks for sanity of those flags, for example it is only
possible to cut new alpha or beta releases on the %s branch.`,
		git.DefaultBranch, git.DefaultBranch,
	),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return NewFindGreenBuild().Run(findGreenBuildOpts)
	},
}

type FindGreenBuildOptions struct {
	Branch        string
	ExcludeSuites []string
	ReleaseType   string
}

var findGreenBuildOpts = &FindGreenBuildOptions{}

type FindGreenBuild struct {
	client releaseClient
}

//counterfeiter:generate . releaseClient
type releaseClient interface {
	SetBuildVersion(string, []string) (string, error)
	SetReleaseVersion(string, string, string) (*release.Versions, error)
}

type defaultReleaseClient struct{}

// NewFindGreenBuild createas a new FindGreenBuild instance
func NewFindGreenBuild() *FindGreenBuild {
	return &FindGreenBuild{&defaultReleaseClient{}}
}

// Set client can be used to set the internal release lib client
func (f *FindGreenBuild) SetClient(client releaseClient) {
	f.client = client
}

func (*defaultReleaseClient) SetBuildVersion(
	branch string, excludeSuites []string,
) (string, error) {
	return release.NewBuildVersionClient().SetBuildVersion(
		branch, excludeSuites,
	)
}

func (*defaultReleaseClient) SetReleaseVersion(
	releaseType, version, branch string,
) (*release.Versions, error) {
	return release.SetReleaseVersion(
		releaseType, version, branch, "",
	)
}

func init() {
	findGreenBuildCmd.PersistentFlags().StringVarP(
		&findGreenBuildOpts.Branch,
		"branch",
		"b",
		git.DefaultBranch,
		"The branch for which to find the green build",
	)

	findGreenBuildCmd.PersistentFlags().StringSliceVarP(
		&findGreenBuildOpts.ExcludeSuites,
		"exclude-suites",
		"e",
		nil,
		"List of CI suites regex to exclude from the go/nogo criteria",
	)

	findGreenBuildCmd.PersistentFlags().StringVarP(
		&findGreenBuildOpts.ReleaseType,
		"type",
		"t",
		release.ReleaseTypeAlpha,
		fmt.Sprintf("release type, must be one of: '%s'",
			strings.Join([]string{
				release.ReleaseTypeAlpha,
				release.ReleaseTypeBeta,
				release.ReleaseTypeRC,
				release.ReleaseTypeOfficial,
			}, "', '"),
		),
	)

	rootCmd.AddCommand(findGreenBuildCmd)
}

// Run starts the FindGreenBuild session
func (f *FindGreenBuild) Run(opts *FindGreenBuildOptions) error {
	if opts.Branch == git.DefaultBranch &&
		opts.ReleaseType != release.ReleaseTypeAlpha &&
		opts.ReleaseType != release.ReleaseTypeBeta {
		return errors.Errorf(
			"can only do alpha's or beta's on the %s branch, not %s's",
			git.DefaultBranch, opts.ReleaseType,
		)
	}

	logrus.Infof(
		"Searching for a build version on branch %s (excluded jobs: %v)",
		opts.Branch, opts.ExcludeSuites,
	)

	foundVersion, err := f.client.SetBuildVersion(
		opts.Branch, opts.ExcludeSuites,
	)
	if err != nil {
		return errors.Wrap(err, "set build version")
	}
	logrus.Infof("Found build version %s", foundVersion)

	logrus.Info("Evaluating next release versions")
	versions, err := f.client.SetReleaseVersion(
		opts.ReleaseType, foundVersion, opts.Branch,
	)
	if err != nil {
		return errors.Wrap(err, "set release version")
	}
	logrus.Infof("The next release version is: %s", versions.Prime())

	return nil
}

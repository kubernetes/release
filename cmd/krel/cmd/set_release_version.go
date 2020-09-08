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

	"k8s.io/release/pkg/release"
)

// setReleaseVersionCmd represents the subcommand for `krel set-release-version`
var setReleaseVersionCmd = &cobra.Command{
	Use:   "set-release-version",
	Short: "Set the release version to be sourced from bash",
	Long: `krel set-release-version

This subcommand can be used to set the release version from a bash command by
sourcing it's output. It's mainly indented to be used from anago, which means
the command might be removed in future releases again if anago is end of life.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := runSetReleaseVersion(setReleaseVersionOpts, anagoOpts)
		if err != nil {
			return err
		}
		fmt.Print(res)
		return nil
	},
}

type setReleaseVersionOptions struct {
	buildVersion string
	branch       string
	parentBranch string
}

var setReleaseVersionOpts = &setReleaseVersionOptions{}

func init() {
	setReleaseVersionCmd.PersistentFlags().StringVar(
		&setReleaseVersionOpts.buildVersion,
		"build-version",
		"",
		"build version to be used",
	)

	setReleaseVersionCmd.PersistentFlags().StringVar(
		&setReleaseVersionOpts.branch,
		"branch",
		"",
		"branch for which the version should be calculated",
	)

	setReleaseVersionCmd.PersistentFlags().StringVar(
		&setReleaseVersionOpts.parentBranch,
		"parent-branch",
		"",
		"the parent branch for the target branch",
	)

	for _, f := range []string{
		"build-version",
		"branch",
		"parent-branch",
	} {
		if err := setReleaseVersionCmd.MarkPersistentFlagRequired(f); err != nil {
			logrus.Fatalf("Unable to set %q flag as required: %v", f, err)
		}
	}

	anagoCmd.AddCommand(setReleaseVersionCmd)
}

func runSetReleaseVersion(opts *setReleaseVersionOptions, anagoOpts *release.Options) (string, error) {
	releaseVersion, err := release.SetReleaseVersion(
		anagoOpts.ReleaseType,
		opts.buildVersion,
		opts.branch,
		opts.parentBranch,
	)
	if err != nil {
		return "", errors.Wrap(err, "set release version")
	}

	res := strings.Builder{}
	res.WriteString("declare -Ag RELEASE_VERSION\n")
	res.WriteString("declare -ag ORDERED_RELEASE_KEYS\n")

	addKey := func(v string) {
		res.WriteString(fmt.Sprintf("ORDERED_RELEASE_KEYS+=(%q)\n", v))
	}

	addVersion := func(k, v string) {
		res.WriteString(fmt.Sprintf("RELEASE_VERSION[%s]=%q\n", k, v))
	}

	if releaseVersion.Official() != "" {
		const v = "official"
		addVersion(v, releaseVersion.Official())
		addKey(v)
	}
	if releaseVersion.RC() != "" {
		const v = "rc"
		addVersion(v, releaseVersion.RC())
		addKey(v)
	}
	if releaseVersion.Beta() != "" {
		const v = "beta"
		addVersion(v, releaseVersion.Beta())
		addKey(v)
	}
	if releaseVersion.Alpha() != "" {
		const v = "alpha"
		addVersion(v, releaseVersion.Alpha())
		addKey(v)
	}

	res.WriteString(fmt.Sprintf(
		"export RELEASE_VERSION_PRIME=%s\n", releaseVersion.Prime(),
	))
	return res.String(), nil
}

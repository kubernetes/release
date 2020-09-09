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

// generateReleaseVersionCmd represents the subcommand for `krel generate-release-version`
var generateReleaseVersionCmd = &cobra.Command{
	Use:   "generate-release-version",
	Short: "Generate the release version to be sourced from bash",
	Long: `krel generate-release-version

This subcommand can be used to generate the release version from a bash command
by sourcing it's output. It's mainly indented to be used from anago, which
means the command might be removed in future releases again if anago is end of
life.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := runGenerateReleaseVersion(generateReleaseVersionOpts, anagoOpts)
		if err != nil {
			return err
		}
		fmt.Print(res)
		return nil
	},
}

type generateReleaseVersionOptions struct {
	buildVersion string
	branch       string
	parentBranch string
}

var generateReleaseVersionOpts = &generateReleaseVersionOptions{}

func init() {
	generateReleaseVersionCmd.PersistentFlags().StringVar(
		&generateReleaseVersionOpts.buildVersion,
		"build-version",
		"",
		"build version to be used",
	)

	generateReleaseVersionCmd.PersistentFlags().StringVar(
		&generateReleaseVersionOpts.branch,
		"branch",
		"",
		"branch for which the version should be calculated",
	)

	generateReleaseVersionCmd.PersistentFlags().StringVar(
		&generateReleaseVersionOpts.parentBranch,
		"parent-branch",
		"",
		"the parent branch for the target branch",
	)

	for _, f := range []string{
		"build-version",
		"branch",
		"parent-branch",
	} {
		if err := generateReleaseVersionCmd.MarkPersistentFlagRequired(f); err != nil {
			logrus.Fatalf("Unable to set %q flag as required: %v", f, err)
		}
	}

	anagoCmd.AddCommand(generateReleaseVersionCmd)
}

func runGenerateReleaseVersion(opts *generateReleaseVersionOptions, anagoOpts *release.Options) (string, error) {
	releaseVersion, err := release.GenerateReleaseVersion(
		anagoOpts.ReleaseType,
		opts.buildVersion,
		opts.branch,
		opts.parentBranch == git.DefaultBranch,
	)
	if err != nil {
		return "", errors.Wrap(err, "generate release version")
	}

	res := strings.Builder{}
	res.WriteString("declare -Ag RELEASE_VERSION\n")
	res.WriteString("declare -ag ORDERED_RELEASE_KEYS\n")

	add := func(k, v string) {
		res.WriteString(fmt.Sprintf("RELEASE_VERSION[%s]=%q\n", k, v))
		res.WriteString(fmt.Sprintf("ORDERED_RELEASE_KEYS+=(%q)\n", k))
	}

	if releaseVersion.Official() != "" {
		add(release.ReleaseTypeOfficial, releaseVersion.Official())
	}
	if releaseVersion.RC() != "" {
		add(release.ReleaseTypeRC, releaseVersion.RC())
	}
	if releaseVersion.Beta() != "" {
		add(release.ReleaseTypeBeta, releaseVersion.Beta())
	}
	if releaseVersion.Alpha() != "" {
		add(release.ReleaseTypeAlpha, releaseVersion.Alpha())
	}

	res.WriteString(fmt.Sprintf(
		"export RELEASE_VERSION_PRIME=%s\n", releaseVersion.Prime(),
	))
	return res.String(), nil
}

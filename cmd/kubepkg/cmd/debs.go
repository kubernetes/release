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

package cmd

import (
	"github.com/spf13/cobra"

	kpkg "k8s.io/release/pkg/kubepkg"
	"k8s.io/release/pkg/util"
)

type debsOptions struct {
}

//nolint
// TODO: Determine if we need debOpts
var debsOpts = &debsOptions{}

// debsCmd represents the base command when called without any subcommands
var debsCmd = &cobra.Command{
	Use:           "debs [--arch <architectures>] [--channels <channels>]",
	Short:         "debs creates Debiam-based packages for Kubernetes components",
	Example:       "kubepkg debs --arch amd64 --channels nightly",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runDebs(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(debsCmd)
}

func runDebs() error {
	ro := rootOpts

	// Replace the "+" with a "-" to make it semver-compliant
	ro.kubeVersion = util.TrimTagPrefix(ro.kubeVersion)

	builds, err := kpkg.ConstructBuilds(ro.packages, ro.channels, ro.kubeVersion, ro.revision, ro.cniVersion, ro.criToolsVersion)
	if err != nil {
		return err
	}

	return kpkg.WalkBuilds(builds, ro.architectures)
}

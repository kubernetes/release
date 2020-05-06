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
	"github.com/spf13/cobra"

	kpkg "k8s.io/release/pkg/kubepkg"
)

type rpmsOptions struct {
}

// TODO: Determine if we need rpmsOpts
var rpmsOpts = &rpmsOptions{} // nolint: deadcode,varcheck,unused

// rpmsCmd represents the base command when called without any subcommands
var rpmsCmd = &cobra.Command{
	Use:           "rpms [--arch <architectures>] [--channels <channels>]",
	Short:         "rpms creates RPMs for Kubernetes components",
	Example:       "kubepkg rpms --arch amd64 --channels nightly",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(*cobra.Command, []string) error {
		return rootOpts.validate()
	},
	RunE: func(*cobra.Command, []string) error {
		return runRpms(rootOpts)
	},
}

func init() {
	rootCmd.AddCommand(rpmsCmd)
}

func runRpms(ro *rootOptions) error {
	builds, err := kpkg.ConstructBuilds("rpm", ro.packages, ro.channels, ro.kubeVersion, ro.revision, ro.cniVersion, ro.criToolsVersion, ro.templateDir)
	if err != nil {
		return err
	}
	return kpkg.WalkBuilds(builds, ro.architectures, ro.specOnly)
}

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
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/release/pkg/command"
)

type installToolsOptions struct {
	notFailOnOutdated bool
}

var installToolsOpts = &installToolsOptions{}

var installToolsCmd = &cobra.Command{
	Use:          "install-tools",
	Short:        "install-tools downloads and compiles necessary release tools",
	SilenceUsage: true,
	RunE: func(*cobra.Command, []string) error {
		return runInstallTools()
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	installToolsCmd.PersistentFlags().BoolVar(
		&installToolsOpts.notFailOnOutdated,
		"not-fail-on-outdated",
		false,
		"do not fail if go modules are outdated",
	)
	rootCmd.AddCommand(installToolsCmd)
}

var installTools = []string{
	"k8s.io/release/cmd/blocking-testgrid-tests",
}

func runInstallTools() error {
	setupEnv()

	if err := checkDeps(installToolsOpts.notFailOnOutdated); err != nil {
		return err
	}

	return goInstall(installTools...)
}

func setupEnv() {
	if gobin, isSet := os.LookupEnv("GOBIN"); isSet {
		os.Setenv("PATH", fmt.Sprintf("%s:%s", gobin, os.Getenv("PATH")))
	}
}

func checkDeps(notFailOnOutdated bool) error {
	goInstall("github.com/psampaz/go-mod-outdated")
	ciFlag := "-ci"
	if notFailOnOutdated {
		ciFlag = ""
	}
	return command.Execute(
		"go list -u -m -json all 2>/dev/null |",
		"go-mod-outdated -update -direct", ciFlag,
	)
}

func goInstall(tools ...string) error {
	for _, tool := range tools {
		log.Printf("Installing %q", tool)
		if err := command.Execute("go install", tool); err != nil {
			return err
		}
	}
	return nil
}

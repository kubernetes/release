/*
Copyright © The Kubernetes Authors

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
	"bytes"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

// ATTENTION: if you're modifying this struct, make sure you update the command help
type fetchOptions struct {
	SourceRepo  string `split_words:"true" required:"true" default:"github.com/kubernetes/kubernetes"`
	StorageRepo string `split_words:"true" required:"true" default:"github.com/wilsonehusin/k8s-release-notes-data"` // TODO: create new repository?
	StorageDir  string `split_words:"true" required:"true" default:"release-1.21/raw"`
}

var fetchOpts = &fetchOptions{}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetches the latest changes to store",
	Long: `rnd fetch -- Fetch latest release notes

This command will look up relevant pull requests in RND_SOURCE_REPO,
add / update the entries in RND_STORAGE_REPO under RND_STORAGE_DIR.`,
	PreRunE: func(*cobra.Command, []string) error {
		return processFetchFlags()
	},
	Run: func(*cobra.Command, []string) {
		cmd.Help()
	},
}

func init() {
	var optionsUsage bytes.Buffer
	if err := envconfig.Usagef(progName, fetchOpts, &optionsUsage, optionsUsageTemplate); err != nil {
		panic(err)
	}
	fetchCmd.SetUsageTemplate(fetchCmd.UsageTemplate() + optionsUsageHeader + optionsUsage.String() + rootCmdOptionsUsage())
	rootCmd.AddCommand(fetchCmd)
}

func processFetchFlags() error {
	if err := envconfig.Process(progName, fetchOpts); err != nil {
		return err
	}

	return nil
}

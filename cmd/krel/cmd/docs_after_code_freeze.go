/*
Copyright 2021 The Kubernetes Authors.

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
	"context"
	"errors"
	"os"

	"github.com/google/go-github/v33/github"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/docs"
)

var afterCodeCmd = &cobra.Command{
	Use:   "after-code-freeze",
	Short: "Arrange kubernetes website pull request after code freeze",
	Long: `The docs after-code-freeze checks if the kubernetes(k/k) pull request is merged
and if it isn't merged, closes the kubernetes website pull request and removes it from the current milestone`,
	Example: `# Create a Github personal access token and export it as a var
	export GITHUB_TOKEN=<token>

	# Run the docs command
	krel docs after-code-freeze --sheet-id=<sheet-id> --release=<release-cycle> 

	# Run the docs command for a different sheet name(default is docs)
	krel docs after-code-freeze --sheet-id=<sheet-id> --release=<release-cycle> --name=<sheet-name>
	`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          afterCodeRun,
}

func init() {
	automateCmd.AddCommand(afterCodeCmd)
}

func afterCodeRun(cmd *cobra.Command, args []string) error {
	ghToken := os.Getenv(tokenVar)
	if ghToken == "" {
		return errors.New("GITHUB_TOKEN environment variable not found")
	}

	readRange := docsOpts.sheetName + "!" + docsOpts.rangeCol
	prdata, err := getPrStruct(docsOpts.sheetId, readRange, docsOpts.colID)
	if err != nil {
		return err
	}

	glClient, err := docs.AuthGitHub(ghToken)
	if err != nil {
		return err
	}
	ctx := context.Background()

	for _, data := range prdata {
		if data.KubernetesPR == 0 {
			continue
		}

		kubernetesPR, _, err := glClient.PullRequests.Get(ctx, k8sOrg, k8sOrg, data.KubernetesPR)
		if err != nil {
			return err
		}

		if !kubernetesPR.GetMerged() {
			closed := "closed"
			reqData := github.IssueRequest{
				State:     &closed,
				Milestone: nil,
			}
			_, _, err := glClient.Issues.Edit(ctx, k8sOrg, websiteRepo, data.WebsitePR, &reqData)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

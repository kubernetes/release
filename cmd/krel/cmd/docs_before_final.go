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
	"fmt"

	"github.com/spf13/cobra"
	sdkGithub "sigs.k8s.io/release-sdk/github"

	"k8s.io/release/cmd/krel/cmd/templates"
	"k8s.io/release/pkg/docs"
)

var beforeFinalCmd = &cobra.Command{
	Use:   "before-final-deadline",
	Short: "Arrange docs PR after before the final docs deadline",
	Long: `The docs before-final-deadline checks if the kubernetes website pull request 
is merged and removes all the labels on the pull request. If it isn't merged yet, It adds comment to
remind the owner of the pull request of the upcoming deadline`,
	Example: `# Create a Github personal access token and export it as a var
	export GITHUB_TOKEN=<token>

	# Run the docs command
	krel docs before-final-deadline --sheet-id=<sheet-id> --release=<release-cycle> 

	# Run the docs command for a different sheet name(default is docs)
	krel docs before-final-deadline --sheet-id=<sheet-id> --release=<release-cycle> --name=<sheet-name>`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          beforeFinalRun,
}

func init() {
	automateCmd.AddCommand(beforeFinalCmd)
}

func beforeFinalRun(cmd *cobra.Command, args []string) error {
	readRange := docsOpts.sheetName + "!" + docsOpts.rangeCol
	prdata, err := getPrStruct(docsOpts.sheetID, readRange, docsOpts.colID)
	if err != nil {
		return err
	}

	ghClient := sdkGithub.New()
	ctx := context.Background()

	for _, data := range prdata {
		websitePR, _, err := ghClient.Client().GetPullRequest(ctx, k8sOrg, websiteRepo, data.KubernetesPR)
		if err != nil {
			return err
		}

		if websitePR.GetMerged() {
			// Do we want to close issues and add/ remove milestones
			// through comments(let the ci-robot make the API calls)?
			_, _, err := ghClient.Client().AddLabels(ctx, k8sOrg, websiteRepo, data.WebsitePR, []string{})
			if err != nil {
				return fmt.Errorf("error putting label on pull request: %s", err)
			}
			return nil
		}

		err = docs.CommentOnIssue(ctx, ghClient.Client(), websiteRepo, data.WebsitePR, templates.UpcomingDeadline, docs.MessageVar{Name: data.Assignee})
		if err != nil {
			return err
		}
	}

	return nil
}

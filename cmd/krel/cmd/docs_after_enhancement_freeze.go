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
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/go-github/v39/github"
	"github.com/spf13/cobra"

	"k8s.io/release/cmd/krel/cmd/templates"
	"k8s.io/release/pkg/docs"
	sdkGithub "sigs.k8s.io/release-sdk/github"
)

type afterFreezeOptions struct {
	milestoneID int
}

var afterFreezeOpts afterFreezeOptions

// announceCmd represents the subcommand for `krel announce`
var afterFreezeCmd = &cobra.Command{
	Use:   "after-enhancements-freeze",
	Short: "Arrange docs PR after enhancement freeze",
	Long: `The docs after-enhancemet-freeze checks 
1. if the kubernetes(k/k) pull request is merged else it put a "hold" label on the website pull request
2. Adds the website pull request to the correct milestone
3. Checks if the website pull request was made to the correct branch 
else it adds a comment to notify the owner of the pull request`,
	Example: `# Create a Github personal access token and export it as a var
	export GITHUB_TOKEN=<token>

	# Run the docs command
	krel docs after-enhancements-freeze --sheet-id=<sheet-id> --release=<release-cycle> 

	# Run the docs command for a different sheet name(default is docs)
	krel docs after-enhancements-freeze --sheet-id=<sheet-id> --release=<release-cycle> --name=<sheet-name>`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          afterFreezeRun,
}

func init() {
	afterFreezeCmd.Flags().IntVarP(
		&afterFreezeOpts.milestoneID,
		"milestone",
		"",
		0,
		"the number of the milestone in the website repository, can be found by looking at the url when on the milestone page")

	automateCmd.AddCommand(afterFreezeCmd)
}

func afterFreezeRun(_ *cobra.Command, _ []string) error {
	ghToken := os.Getenv(ghTokVar)
	if ghToken == "" {
		return errors.New("GITHUB_TOKEN environment variable not found")
	}

	baseBranch := fmt.Sprintf("dev-v%s", docsOpts.release)

	if docsOpts.sheetID == "" {
		return fmt.Errorf("please pass in the sheet id")
	}

	if len(docsOpts.colID) != 5 {
		return fmt.Errorf("please pass in column letter for kep, k/w pr, sig, assignee and k/k pr only")
	}

	ctx := context.Background()

	// TODO(SomtochiAma): Validate rangeCol
	readRange := docsOpts.sheetName + "!" + docsOpts.rangeCol
	issueMaps, err := getPrStruct(docsOpts.sheetID, readRange, docsOpts.colID)
	if err != nil {
		return fmt.Errorf("error getting data from sheets: %s", err)
	}

	ghClient := sdkGithub.New()

	for _, issuePr := range issueMaps {
		if issuePr.WebsitePR == 0 {
			continue
		}

		// Check if k/k pull request is merged
		kubernetesPr, _, err := ghClient.Client().GetPullRequest(ctx, k8sOrg, k8sOrg, issuePr.KubernetesPR)
		if err != nil {
			return err
		}

		// if it isn't merged, put '/hold' label on website pr
		if !kubernetesPr.GetMerged() {
			_, _, err := ghClient.Client().AddLabels(ctx, k8sOrg, websiteRepo, issuePr.WebsitePR, []string{"do-not-merge/hold"})
			if err != nil {
				return fmt.Errorf("error putting label on pull request: %s", err)
			}
		}

		websitePR, _, err := ghClient.Client().GetPullRequest(ctx, k8sOrg, websiteRepo, issuePr.WebsitePR)
		if err != nil {
			return err
		}

		// Add website pr to release milestone if not already present
		if websitePR.Milestone.GetNumber() != afterFreezeOpts.milestoneID {
			issue := &github.IssueRequest{
				Milestone: &afterFreezeOpts.milestoneID,
			}
			_, _, err := ghClient.Client().UpdateIssue(ctx, k8sOrg, websiteRepo, issuePr.WebsitePR, issue)
			if err != nil {
				return err
			}
		}

		// comment on PR if it made against wrong branch
		if websitePR.GetBase().GetRef() != baseBranch {
			msgVar := docs.MessageVar{
				Name:    issuePr.Assignee,
				Release: docsOpts.release,
			}

			err := docs.CommentOnIssue(ctx, ghClient.Client(), websiteRepo, issuePr.WebsitePR, templates.WrongBaseBranch, msgVar)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

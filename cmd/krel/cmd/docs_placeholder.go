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

	"github.com/spf13/cobra"
	sdkGithub "sigs.k8s.io/release-sdk/github"

	"k8s.io/release/cmd/krel/cmd/templates"
	"k8s.io/release/pkg/docs"
)

var placeHolderCmd = &cobra.Command{
	Use:   "before-placeholder",
	Short: "Comment on KEP without placeholder PR",
	Long: `The docs before-placeholder comments on the kubernetes enhancements issue 
if the placeholder issue hasn't been created before the placeholder'`,
	Example: `# Create a Github personal access token and export it as a var
	export GITHUB_TOKEN=<token>

	# Run the docs command
	krel docs before-placeholder --sheet-id=<sheet-id> --release=<release-cycle> 

	# Run the docs command for a different sheet name(default is docs)
	krel docs before-placeholder --sheet-id=<sheet-id> --release=<release-cycle> --name=<sheet-name>`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          beforePlaceholderRun,
}

func init() {
	automateCmd.AddCommand(placeHolderCmd)
}

func beforePlaceholderRun(cmd *cobra.Command, args []string) error {
	readRange := docsOpts.sheetName + "!" + docsOpts.rangeCol
	prdata, err := getPrStruct(docsOpts.sheetID, readRange, docsOpts.colID)
	if err != nil {
		return err
	}

	glClient := sdkGithub.New()

	ctx := context.Background()
	for _, pr := range prdata {
		// don't comment on kep issue if placeholder pr already exists
		if pr.WebsitePR != 0 {
			continue
		}

		msgVar := docs.MessageVar{
			Name:    pr.Assignee,
			Release: docsOpts.release,
		}

		err := docs.CommentOnIssue(ctx, glClient.Client(), "mars-dashboard", pr.KEP, templates.MissingPlaceholder, msgVar)
		if err != nil {
			// Do we want to return an error or just log it
			return err
		}
	}

	return nil
}

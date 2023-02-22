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
	"os"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"sigs.k8s.io/release-utils/env"
)

var githubCmd = &cobra.Command{
	Use:    "github",
	Short:  "Github report generator",
	Long:   "CI-Signal reporter that generates only a github report.",
	PreRun: setGithubConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunReport(cfg, &CIReporters{GithubReporter{}})
	},
}

// look for token in environment variables & create a github client
func setGithubConfig(cmd *cobra.Command, args []string) {
	cfg.GithubToken = env.Default("GITHUB_TOKEN", "")
	if cfg.GithubToken == "" {
		logrus.Fatal("Please specify your Github access token via the environment variable 'GITHUB_TOKEN' to generate a ci-report")
		os.Exit(1)
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	cfg.GithubClient = githubv4.NewClient(httpClient)
}

// GithubReporterName used to identify github reporter
var GithubReporterName CIReporterName = "github"

func init() {
	rootCmd.AddCommand(githubCmd)
}

//
// GithubReporter implementation
//

// GithubReporter github CIReporter implementation
type GithubReporter struct{}

// GetCIReporterHead implementation from CIReporter
func (r GithubReporter) GetCIReporterHead() CIReporterInfo {
	return CIReporterInfo{Name: GithubReporterName}
}

// CollectReportData implementation from CIReporter
func (r GithubReporter) CollectReportData(cfg *Config) ([]*CIReportRecord, error) {
	// set filter configuration
	denyListFilter := map[FilteredFieldName][]FilteredListVal{}
	allowListFilter := map[FilteredFieldName][]FilteredListVal{
		FilteredFieldName("view"): {"issue-tracking"},
	}
	if cfg.ShortReport {
		denyListFilter[FilteredFieldName("Status")] = []FilteredListVal{FilteredListVal("RESOLVED"), FilteredListVal("PASSING")}
	}
	if cfg.ReleaseVersion != "" {
		allowListFilter[FilteredFieldName("K8s Release")] = []FilteredListVal{FilteredListVal(cfg.ReleaseVersion)}
	}
	// request github projectboard data
	githubReportData, err := GetGithubReportData(*cfg, denyListFilter, allowListFilter)
	if err != nil {
		return nil, fmt.Errorf("getting GitHub report data: %w", err)
	}
	records := []*CIReportRecord{}

	for _, item := range githubReportData {
		// set the URL to the Issue- / PR- URL if set
		URL := ""
		if issueURL, ok := item.Fields[fieldName(IssueURLKey)]; ok {
			URL = string(issueURL)
		}
		if prURL, ok := item.Fields[fieldName(PullRequestURLKey)]; ok {
			URL = string(prURL)
		}
		// add a new record to the report
		records = append(records, &CIReportRecord{
			Title:            item.Title,
			TestgridBoard:    string(item.Fields[fieldName(TestgridBoardKey)]),
			URL:              URL,
			Status:           string(item.Fields[fieldName(StatusKey)]),
			StatusDetails:    string(item.Fields[fieldName(CiSignalMemberKey)]),
			CreatedTimestamp: string(item.Fields[fieldName(CreatedAtKey)]),
			UpdatedTimestamp: string(item.Fields[fieldName(UpdatedAtKey)]),
		})
	}
	return records, nil
}

//
// Helper functions to collect github data
//

// This can be seen in the project's URL or looked up in the API
const ciSignalProjectBoardNumber = 68
const kubernetesOrganizationName = "kubernetes"

type ciSignalProjectBoardKey string

const (
	// custom project board keys that get extracted via graphql
	IssueURLKey       = ciSignalProjectBoardKey("Issue URL")
	PullRequestURLKey = ciSignalProjectBoardKey("PullRequest URL")
	// project board column headers
	TestgridBoardKey       = ciSignalProjectBoardKey("Testgrid Board")
	SlackDiscussionLinkKey = ciSignalProjectBoardKey("Slack discussion link")
	StatusKey              = ciSignalProjectBoardKey("Status")
	CiSignalMemberKey      = ciSignalProjectBoardKey("CI Signal Member")
	CreatedAtKey           = ciSignalProjectBoardKey("Created At")
	UpdatedAtKey           = ciSignalProjectBoardKey("Updated At")
)

// This struct represents a graphql query
//
//	that is getting executed using the githubv4
//	graphql library: https://github.com/shurcooL/githubv4
//	for the GitHub graphql api, see: https://docs.github.com/en/issues/trying-out-the-new-projects-experience/using-the-api-to-manage-projects
//
// ENHANCEMENT: filter via request, see: https://dgraph.io/docs/graphql/queries/search-filtering/

type (
	TextValueFragment struct {
		Text  string
		Field struct {
			TextFieldNameFragment `graphql:"... on ProjectV2Field"`
		}
	}
	TextFieldNameFragment struct {
		Name string
	}
	SingleSelectFragment struct {
		Name  string
		Field struct {
			SingleSelectFieldNameFragment `graphql:"... on ProjectV2SingleSelectField"`
		}
	}
	SingleSelectFieldNameFragment struct {
		Name string
	}
)

type ciSignalProjectBoardGraphQLQuery struct {
	Organization struct {
		ProjectV2 struct {
			// Items board rows with content
			Items struct {
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
					StartCursor string
				}
				Nodes []struct {
					ID          string
					FieldValues struct {
						Nodes []struct {
							TextValueFragment    `graphql:"... on ProjectV2ItemFieldTextValue"`
							SingleSelectFragment `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
						}
					} `graphql:"fieldValues(first: 20)"`
					Content struct {
						Issue struct {
							URL   string
							Title string
						} `graphql:"... on Issue"`
						PullRequest struct {
							URL   string
							Title string
						} `graphql:"... on PullRequest"`
					}
				}
			} `graphql:"items(first: 100 after: $cursor)"`
		} `graphql:"projectV2(number: $projectBoardNum)"`
	} `graphql:"organization(login: $kubernetesOrganizationName)"`
}

type (
	fieldValue                  string
	fieldName                   string
	TransformedProjectBoardItem struct {
		ID     string
		Title  string
		Fields map[fieldName]fieldValue
	}

	// Types for project board filtering
	FilteredFieldName string
	FilteredListVal   string
)

// GetGithubReportData used to request the raw report data from github
func GetGithubReportData(cfg Config, denyListFieldFilter, allowListFieldFilter map[FilteredFieldName][]FilteredListVal) ([]*TransformedProjectBoardItem, error) {
	hasNextPage := true
	var cursor string
	var queryCiSignalProjectBoard ciSignalProjectBoardGraphQLQuery
	transformedProjectBoardItems := []*TransformedProjectBoardItem{}

	for hasNextPage {
		// lookup project board information
		variablesProjectBoardFields := map[string]interface{}{
			"projectBoardNum":            githubv4.Int(ciSignalProjectBoardNumber),
			"kubernetesOrganizationName": githubv4.String(kubernetesOrganizationName),
			"cursor":                     githubv4.String(cursor),
		}

		if err := cfg.GithubClient.Query(context.Background(), &queryCiSignalProjectBoard, variablesProjectBoardFields); err != nil {
			return nil, err
		}

		hasNextPage = queryCiSignalProjectBoard.Organization.ProjectV2.Items.PageInfo.HasNextPage
		cursor = queryCiSignalProjectBoard.Organization.ProjectV2.Items.PageInfo.EndCursor

		// until better graphql filtering, filter out allow and denylist in go
		for _, item := range queryCiSignalProjectBoard.Organization.ProjectV2.Items.Nodes {
			transFields := map[fieldName]fieldValue{}
			itemBlacklisted := false
			for _, field := range item.FieldValues.Nodes {
				// TODO: better graphql aliasing for these fragments?
				fieldVal := field.Text
				fieldN := field.TextValueFragment.Field.Name
				if fieldVal == "" {
					fieldVal = field.Name
					fieldN = field.SingleSelectFragment.Field.Name
				}

				// filter out deny listed values
				if denyListValues, filteredFieldFound := denyListFieldFilter[FilteredFieldName(fieldN)]; filteredFieldFound {
					// The field is a filtered field since it could be found in the fieldFilter map
					// 	check if the value of the field is blacklisted
					for _, bv := range denyListValues {
						if fieldVal == string(bv) {
							itemBlacklisted = true
							break
						}
					}
					if itemBlacklisted {
						break
					}
				}
				// filter for allow listed values
				if allowListValues, filteredFieldFound := allowListFieldFilter[FilteredFieldName(fieldN)]; filteredFieldFound {
					// The field is a filtered field since it could be found in the fieldFilter map
					// 	check if the value of the field is blacklisted
					for _, bv := range allowListValues {
						if fieldVal != string(bv) {
							itemBlacklisted = true
							break
						}
					}
					if itemBlacklisted {
						break
					}
				}

				if itemBlacklisted {
					break
				}

				transFields[fieldName(fieldN)] = fieldValue(fieldVal)

			}

			if itemBlacklisted {
				continue
			}

			// TODO: better graphql aliasing for these fragments?
			if item.Content.Issue.URL != "" {
				transFields[fieldName("Issue URL")] = fieldValue(item.Content.Issue.URL)
			}
			if item.Content.Issue.Title != "" {
				transFields[fieldName("Title")] = fieldValue(item.Content.Issue.Title)
			}
			if item.Content.PullRequest.URL != "" {
				transFields[fieldName("PullRequest URL")] = fieldValue(item.Content.PullRequest.URL)
			}
			if item.Content.PullRequest.Title != "" {
				transFields[fieldName("Title")] = fieldValue(item.Content.PullRequest.Title)
			}
			transformedProjectBoardItems = append(transformedProjectBoardItems, &TransformedProjectBoardItem{
				ID:     item.ID,
				Title:  string(transFields["Title"]),
				Fields: transFields,
			})
		}
	}
	return transformedProjectBoardItems, nil
}

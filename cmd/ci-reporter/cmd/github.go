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
	"encoding/json"
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

// This can be looked up using the API, see https://docs.github.com/en/issues/trying-out-the-new-projects-experience/using-the-api-to-manage-projects#finding-the-node-id-of-an-organization-project
const ciSignalProjectBoardID = "PN_kwDOAM_34M4AAThW"

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

// GitHubProjectBoardFieldSettings settings for a column of a github beta project board
// --> | Testgrid Board | -> { ID: XXX, Name: Testgrid Board, ... }
// This information is required to match the settings ID to the name since table entries ref. id
type GitHubProjectBoardFieldSettings struct {
	Width   int `json:"width"`
	Options []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		NameHTML string `json:"name_html"`
	} `json:"options"`
}

// This struct represents a graphql query
// 	that is getting executed using the githubv4
// 	graphql library: https://github.com/shurcooL/githubv4
// 	for the GitHub graphql api, see: https://docs.github.com/en/issues/trying-out-the-new-projects-experience/using-the-api-to-manage-projects
// ENHANCEMENT: filter via request, see: https://dgraph.io/docs/graphql/queries/search-filtering/
type ciSignalProjectBoardGraphQLQuery struct {
	Node struct {
		ProjectNext struct {
			// Fields information about the column headers of the project
			// --> | Title | Testgrid Board | Testgrid URL | UpdatedAt | ... |
			Fields struct {
				Nodes []struct {
					Name     string
					Settings string
				}
			} `graphql:"fields(first: 100)"`
			// Items board rows with content
			Items struct {
				Nodes []struct {
					ID          string
					Title       string
					FieldValues struct {
						Nodes []struct {
							Value        string
							ProjectField struct {
								Name string
							}
						}
					} `graphql:"fieldValues(first: 20)"`
					Content struct {
						Issue struct {
							URL string
						} `graphql:"... on Issue"`
						PullRequest struct {
							URL string
						} `graphql:"... on PullRequest"`
					}
				}
			} `graphql:"items(first: 100)"`
		} `graphql:"... on ProjectNext"`
	} `graphql:"node(id: $projectBoardID)"`
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
	// lookup project board information
	var queryCiSignalProjectBoard ciSignalProjectBoardGraphQLQuery
	variablesProjectBoardFields := map[string]interface{}{
		"projectBoardID": githubv4.ID(ciSignalProjectBoardID),
	}
	if err := cfg.GithubClient.Query(context.Background(), &queryCiSignalProjectBoard, variablesProjectBoardFields); err != nil {
		return nil, err
	}

	// projectBoardFieldIDs hold input IDs of the project board to replace all IDs with names
	// Example: The input "Testgrid Board" is of the type "select"
	// 	to enter a value on the project board you can select of defined values
	// 	every value gets an ID assigned, like this: "master-blocking" = 34u5h2l, "master-informing" = 438tz93
	// 	the information that is looked up on each row references the ID which is cryptic to read
	//
	// 	Received row information: { Testgrid Board: 34u5h2l, ... }
	// 	Transformed row information: { Testgrid Board: "master-blocking", ... }
	type (
		// verbose types
		projectBoardFieldID   string
		projectBoardFieldName string
	)
	projectBoardFieldIDs := map[projectBoardFieldID]projectBoardFieldName{}

	for _, field := range queryCiSignalProjectBoard.Node.ProjectNext.Fields.Nodes {
		var fieldSettings GitHubProjectBoardFieldSettings
		if err := json.Unmarshal([]byte(field.Settings), &fieldSettings); err != nil {
			return nil, err
		}
		for _, option := range fieldSettings.Options {
			projectBoardFieldIDs[projectBoardFieldID(option.ID)] = projectBoardFieldName(option.Name)
		}
	}

	transformedProjectBoardItems := []*TransformedProjectBoardItem{}
	for _, item := range queryCiSignalProjectBoard.Node.ProjectNext.Items.Nodes {
		transFields := map[fieldName]fieldValue{}
		itemBlacklisted := false
		for _, field := range item.FieldValues.Nodes {
			fieldVal := field.Value
			// To check if the field value is blacklisted
			// 	in the case of a ID stored in the field
			//	this must be replaced first with the projectBoardFieldIDs map
			if val, ok := projectBoardFieldIDs[projectBoardFieldID(field.Value)]; ok {
				// ID detected replace ID with Name
				fieldVal = string(val)
			}

			// filter out deny listed values
			if denyListValues, filteredFieldFound := denyListFieldFilter[FilteredFieldName(field.ProjectField.Name)]; filteredFieldFound {
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
			if allowListValues, filteredFieldFound := allowListFieldFilter[FilteredFieldName(field.ProjectField.Name)]; filteredFieldFound {
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
			transFields[fieldName(field.ProjectField.Name)] = fieldValue(fieldVal)
		}
		if itemBlacklisted {
			continue
		}
		if item.Content.Issue.URL != "" {
			transFields[fieldName("Issue URL")] = fieldValue(item.Content.Issue.URL)
		}
		if item.Content.PullRequest.URL != "" {
			transFields[fieldName("PullRequest URL")] = fieldValue(item.Content.PullRequest.URL)
		}
		transformedProjectBoardItems = append(transformedProjectBoardItems, &TransformedProjectBoardItem{
			ID:     item.ID,
			Title:  item.Title,
			Fields: transFields,
		})
	}

	return transformedProjectBoardItems, nil
}

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
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v34/github"
	"github.com/pkg/errors"
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

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.GithubToken})
	tc := oauth2.NewClient(ctx, ts)
	cfg.GithubClient = github.NewClient(tc)
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
	githubReportData, err := GetGithubReportData(*cfg)
	if err != nil {
		return nil, err
	}
	records := []*CIReportRecord{}

	for columnTitle, issues := range githubReportData {
		for _, issue := range issues {
			records = append(records, &CIReportRecord{
				ID:       fmt.Sprintf("%d", issue.ID),
				Title:    issue.Title,
				URL:      issue.URL,
				Category: string(columnTitle),
				Sigs:     issue.Sigs,
				// information not collected
				Status:           "",
				CreatedTimestamp: "",
			})
		}
	}
	return records, nil
}

//
// Helper functions to collect github data
//

// This regex is getting used to identify sig lables on github issues
var sigRegex = regexp.MustCompile(`sig/[a-zA-Z-]+`)

var (
	newColumn = GithubProjectBoardColumn{
		ColumnTitle: "New/Not Yet Started",
		ColumnID:    4212817,
	}
	underInvestigationColumn = GithubProjectBoardColumn{
		ColumnTitle: "In flight",
		ColumnID:    4212819,
	}
	observingColumn = GithubProjectBoardColumn{
		ColumnTitle: "New/Not Yet Started",
		ColumnID:    4212821,
	}
	resolvedColumn = GithubProjectBoardColumn{
		ColumnTitle: "Resolved",
		ColumnID:    6798858,
	}
)

type (
	// ColumnTitle title of a github project board column
	ColumnTitle string
	// ColumnID ID of a github project board column
	ColumnID int64
)

// GithubProjectBoardColumn specifies a github project board column
type GithubProjectBoardColumn struct {
	ColumnTitle ColumnTitle `json:"column_title"`
	ColumnID    ColumnID    `json:"column_id"`
}

// GithubReportData defines the github report data structure
type GithubReportData map[ColumnTitle][]IssueOverview

// Marshal used to marshal GithubReportData into bytes
func (d *GithubReportData) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

// IssueOverview defines the data types of a github issue in github report data
type IssueOverview struct {
	// URL github issue url
	URL string `json:"url"`
	// ID github issue id
	ID int64 `json:"id"`
	// Title github issue title
	Title string `json:"title"`
	// Sigs kubernetes sigs that are referenced via label
	Sigs []string `json:"sigs"`
}

type issueDetail struct {
	Number  int64          `json:"number"`
	HTMLURL string         `json:"html_url"`
	Title   string         `json:"title"`
	Labels  []github.Label `json:"labels,omitempty"`
}

// GetGithubReportData used to request the raw report data from github
func GetGithubReportData(cfg Config) (GithubReportData, error) {
	ciSignalProjectBoard := []GithubProjectBoardColumn{newColumn, underInvestigationColumn}

	// if the short flag is not set observingColumn & resolvedColumn will be added to the report
	if !cfg.ShortReport {
		ciSignalProjectBoard = append(ciSignalProjectBoard, observingColumn, resolvedColumn)
	}

	githubReportData := map[ColumnTitle][]IssueOverview{}
	for _, column := range ciSignalProjectBoard {
		cards, err := getCardsFromColumn(cfg, column.ColumnID)
		if err != nil {
			return nil, err
		}
		githubReportData[column.ColumnTitle] = cards
	}
	return githubReportData, nil
}

func getCardsFromColumn(cfg Config, cardsID ColumnID) ([]IssueOverview, error) {
	opt := &github.ProjectCardListOptions{}
	cards, _, err := cfg.GithubClient.Projects.ListProjectCards(context.Background(), int64(cardsID), opt)
	if err != nil {
		return nil, errors.Wrap(err, "querying cards")
	}

	issues := []IssueOverview{}
	for _, c := range cards {
		issueDetail, err := getIssueDetail(cfg, *c.ContentURL)
		if err != nil {
			return nil, err
		}

		overview := IssueOverview{
			URL:   issueDetail.HTMLURL,
			ID:    issueDetail.Number,
			Title: issueDetail.Title,
		}
		for _, v := range issueDetail.Labels {
			sig := sigRegex.FindString(*v.Name)
			if sig != "" {
				sig = strings.Replace(sig, "/", " ", 1)
				overview.Sigs = append(overview.Sigs, sig)
			}
		}
		issues = append(issues, overview)
	}
	return issues, nil
}

func getIssueDetail(cfg Config, url string) (*issueDetail, error) {
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating HTTP request")
	}
	// add authorization header to the req
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.GithubToken))

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "getting card details from GitHub")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading GitHub response data")
	}
	var result issueDetail
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal GitHub response data")
	}
	return &result, nil
}

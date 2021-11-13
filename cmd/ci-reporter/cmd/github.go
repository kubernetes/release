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
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/go-github/v34/github"
)

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
	ColumnTitle ColumnTitle
	ColumnID    ColumnID
}

// GithubReportData defines the github report data structure
type GithubReportData map[ColumnTitle][]IssueOverview

// IssueOverview defines the data types of a github issue in github report data
type IssueOverview struct {
	// URL github issue url
	URL string
	// ID github issue id
	ID int64
	// Title github issue title
	Title string
	// Sigs kubernetes sigs that are referenced via label
	Sigs []string
}

// GetGithubReportData used to request the raw report data from github
func GetGithubReportData(cfg ReporterConfig) (GithubReportData, error) {
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

// PrintGithubReportData used to print github report data out to the console
func PrintGithubReportData(reportData map[ColumnTitle][]IssueOverview) {
	for columnTitle, issuesInColumn := range reportData {
		if len(issuesInColumn) > 0 {
			fmt.Println("\n" + columnTitle)
			for _, issue := range issuesInColumn {
				fmt.Printf("#%d %v\n - %s\n - %s\n", issue.ID, issue.Sigs, issue.Title, issue.URL)
			}
		}
	}
	fmt.Print("\n\n")
}

func getCardsFromColumn(cfg ReporterConfig, cardsID ColumnID) ([]IssueOverview, error) {
	opt := &github.ProjectCardListOptions{}
	cards, _, err := cfg.GithubClient.Projects.ListProjectCards(context.Background(), int64(cardsID), opt)
	if err != nil {
		fmt.Printf("error when querying cards %v", err)
		return nil, err
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

type issueDetail struct {
	Number  int64          `json:"number"`
	HTMLURL string         `json:"html_url"`
	Title   string         `json:"title"`
	Labels  []github.Label `json:"labels,omitempty"`
}

func getIssueDetail(cfg ReporterConfig, url string) (*issueDetail, error) {
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// add authorization header to the req
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.GithubToken))

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%v", err)
		return nil, err
	}
	var result issueDetail
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v34/github"
	"golang.org/x/oauth2"
)

type requiredJob struct {
	OutputName string
	URLName    string
}

func main() {
	// parse flags
	isShortFlagSet := flag.Bool("short", false, "A short report for mails and slack")
	releaseVersion := flag.String("v", "", "Adds specific K8s release version to the report like 1.22")
	flag.Parse()

	// get environment variables
	githubAPIToken := os.Getenv("GITHUB_AUTH_TOKEN")
	if githubAPIToken == "" {
		fmt.Printf("Please provide GITHUB_AUTH_TOKEN env variable to be able to pull cards from the github board")
		os.Exit(1)
	}

	// create a new github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAPIToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// GitHub Report
	err := printGithubIssueReport(githubAPIToken, client, *isShortFlagSet)
	if err != nil {
		fmt.Printf("error when querying cards overview, exiting: %v\n", err)
		os.Exit(1)
	}

	// Testgrid Report
	printTestgridReport(*releaseVersion)
}

// This regex is getting used to identify sig lables on github issues
var sigRegex = regexp.MustCompile(`sig/[a-zA-Z-]+`)

type (
	columnTitle string
	columnID    int64
)

const (
	newColumn                columnID = 4212817
	underInvestigationColumn columnID = 4212819
	observingColumn          columnID = 4212821
	resolvedColumn           columnID = 6798858

	newColumnTitle       columnTitle = "New/Not Yet Started"
	inFLightColumnTitle  columnTitle = "In flight"
	observingColumnTitle columnTitle = "Observing"
	resolvedColumnTitle  columnTitle = "Resolved"
)

type issueOverview struct {
	url   string
	id    int64
	title string
	sig   []string
}

func printGithubIssueReport(token string, client *github.Client, setShort bool) error {
	ciSignalProjectBoard := map[columnTitle]columnID{
		newColumnTitle:      newColumn,
		inFLightColumnTitle: underInvestigationColumn,
	}

	// if the short flag is not set observingColumn & resolvedColumn will be added to the report
	if !setShort {
		ciSignalProjectBoard[observingColumnTitle] = observingColumn
		ciSignalProjectBoard[resolvedColumnTitle] = resolvedColumn
	}

	githubReportData := map[columnTitle][]issueOverview{}
	for columnTitle, columnID := range ciSignalProjectBoard {
		cards, err := getCardsFromColumn(columnID, client, token)
		if err != nil {
			return err
		}
		githubReportData[columnTitle] = cards
	}
	printGithubCards(githubReportData)
	return nil
}

func printGithubCards(reportData map[columnTitle][]issueOverview) {
	for columnTitle, issuesInColumn := range reportData {
		if len(issuesInColumn) > 0 {
			fmt.Println("\n" + columnTitle)
			for _, issue := range issuesInColumn {
				fmt.Printf("#%d %v\n - %s\n - %s\n", issue.id, issue.sig, issue.title, issue.url)
			}
		}
	}
	fmt.Print("\n\n")
}

func getCardsFromColumn(cardsID columnID, client *github.Client, token string) ([]issueOverview, error) {
	opt := &github.ProjectCardListOptions{}
	cards, _, err := client.Projects.ListProjectCards(context.Background(), int64(cardsID), opt)
	if err != nil {
		fmt.Printf("error when querying cards %v", err)
		return nil, err
	}

	issues := []issueOverview{}
	for _, c := range cards {
		issueDetail, err := getIssueDetail(*c.ContentURL, token)
		if err != nil {
			return nil, err
		}

		overview := issueOverview{
			url:   issueDetail.HTMLURL,
			id:    issueDetail.Number,
			title: cleanTitle(issueDetail.Title),
		}
		for _, v := range issueDetail.Labels {
			sig := sigRegex.FindString(*v.Name)
			if sig != "" {
				sig = strings.Replace(sig, "/", " ", 1)
				overview.sig = append(overview.sig, sig)
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

func getIssueDetail(url, authToken string) (*issueDetail, error) {
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// add authorization header to the req
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

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

func cleanTitle(title string) string {
	return strings.ReplaceAll(title, "[Failing Test]", "")
}

func printTestgridReport(version string) {
	requiredJobs := []requiredJob{
		{OutputName: "Master-Blocking", URLName: "sig-release-master-blocking"},
		{OutputName: "Master-Informing", URLName: "sig-release-master-informing"},
	}

	if version != "" {
		requiredJobs = append(requiredJobs, []requiredJob{
			{OutputName: version + "-blocking", URLName: "sig-release-" + version + "-blocking"},
			{OutputName: version + "-informing", URLName: "sig-release-" + version + "-informing"},
		}...)
	}

	result := make([]statistics, 0)
	for _, job := range requiredJobs {
		resp, err := http.Get(fmt.Sprintf("https://testgrid.k8s.io/%s/summary", job.URLName))
		if err != nil {
			fmt.Printf("%v", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%v", err)
			return
		}

		jobs, err := getStatsFromJSON(body)
		if err != nil {
			fmt.Printf("%v", err)
			return
		}

		statistics := getStatistics(jobs)
		statistics.name = job.OutputName
		result = append(result, statistics)
	}

	prettyPrint(result)
}

func prettyPrint(stats []statistics) {
	for _, stat := range stats {
		fmt.Printf("Failures in %s\n", stat.name)
		fmt.Printf("\t%d jobs total\n", stat.total)
		fmt.Printf("\t%d are passing\n", stat.passing)
		fmt.Printf("\t%d are flaking\n", stat.flaking)
		fmt.Printf("\t%d are failing\n", stat.failing)
		fmt.Printf("\t%d are stale\n", stat.stale)
		fmt.Print("\n\n")
	}
}

func getStatistics(jobs map[string]overview) statistics {
	result := statistics{}
	for _, v := range jobs {
		if v.OverallStatus == "PASSING" {
			result.passing++
		} else if v.OverallStatus == "FAILING" {
			result.failing++
		} else if v.OverallStatus == "FLAKY" {
			result.flaking++
		} else {
			result.stale++
		}
		result.total++
	}
	return result
}

type statistics struct {
	name    string
	total   int
	passing int
	flaking int
	failing int
	stale   int
}

type overview struct {
	OverallStatus string `json:"overall_status"`
}

func getStatsFromJSON(body []byte) (map[string]overview, error) {
	result := make(map[string]overview)
	err := json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

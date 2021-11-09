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
	"sort"
	"strings"

	"github.com/google/go-github/v34/github"
	"golang.org/x/oauth2"
)

type requiredJob struct {
	OutputName string
	URLName    string
}

func main() {
	boolPtr := flag.Bool("short", false, "a short report for mails and slack")
	flag.Parse()

	githubAPIToken := os.Getenv("GITHUB_AUTH_TOKEN")
	if githubAPIToken == "" {
		fmt.Printf("Please provide GITHUB_AUTH_TOKEN env variable to be able to pull cards from the github board")
		os.Exit(1)
	}

	releaseVersion := os.Getenv("RELEASE_VERSION")

	err := printCardsOverview(githubAPIToken, *boolPtr)
	if err != nil {
		fmt.Printf("error when querying cards overview, exiting: %v\n", err)
		os.Exit(1)
	}

	printJobsStatistics(releaseVersion)
}

const (
	newCards                = 4212817
	underInvestigationCards = 4212819
	observingCards          = 4212821
	ciSignalBoardProjectID  = 2093513
)

type issueOverview struct {
	url   string
	id    int64
	title string
	sig   string
}

func printCardsOverview(token string, setShort bool) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	newCardsOverview, err := getCardsFromColumn(newCards, client, token)
	if err != nil {
		return err
	}

	investigationCardsOverview, err := getCardsFromColumn(underInvestigationCards, client, token)
	if err != nil {
		return err
	}

	observingCardsOverview, err := getCardsFromColumn(observingCards, client, token)
	if err != nil {
		return err
	}

	resolvedCards, err := findResolvedCardsColumns(client)
	if err != nil {
		return err
	}

	resolvedCardsOverview, err := getCardsFromColumn(resolvedCards, client, token)
	if err != nil {
		return err
	}

	printCards(setShort, groupByCards(newCardsOverview), groupByCards(investigationCardsOverview), groupByCards(observingCardsOverview), groupByCards(resolvedCardsOverview))

	return nil
}

func findResolvedCardsColumns(client *github.Client) (int64, error) {
	opt := &github.ListOptions{}
	columns, _, err := client.Projects.ListProjectColumns(context.Background(), ciSignalBoardProjectID, opt)
	if err != nil {
		return 0, err
	}

	resolvedColumns := make([]*github.ProjectColumn, 0)
	for _, v := range columns {
		if v.Name != nil && *v.Name == "Resolved" {
			resolvedColumns = append(resolvedColumns, v)
		}
	}

	sort.Slice(resolvedColumns, func(i, j int) bool {
		return resolvedColumns[i].GetID() < resolvedColumns[j].GetID()
	})
	return resolvedColumns[0].GetID(), err
}

func printCards(shortReport bool, newCards, investigation, observing, resolved map[string][]*issueOverview) {
	fmt.Println("New/Not Yet Started")
	for k, v := range newCards {
		fmt.Printf("SIG %s\n", k)
		for _, i := range v {
			fmt.Printf("#%d %s %s\n", i.id, i.url, i.title)
		}
		fmt.Println()
	}

	fmt.Println("In flight")
	for k, v := range investigation {
		fmt.Printf("SIG %s\n", k)
		for _, i := range v {
			fmt.Printf("#%d %s %s\n", i.id, i.url, i.title)
		}
		fmt.Println()
	}

	if !shortReport {
		fmt.Println("Observing")
		for k, v := range observing {
			fmt.Printf("SIG %s\n", k)
			for _, i := range v {
				fmt.Printf("#%d %s %s\n", i.id, i.url, i.title)
			}
			fmt.Println()
		}

		fmt.Println("Resolved")
		for k, v := range resolved {
			fmt.Printf("SIG %s\n", k)
			for _, i := range v {
				fmt.Printf("#%d %s %s\n", i.id, i.url, i.title)
			}
			fmt.Println()
		}
	}
}

func groupByCards(issues []*issueOverview) map[string][]*issueOverview {
	result := make(map[string][]*issueOverview)
	for _, i := range issues {
		_, ok := result[i.sig]
		if !ok {
			result[i.sig] = make([]*issueOverview, 0)
		}
		result[i.sig] = append(result[i.sig], i)
	}
	return result
}

func getCardsFromColumn(cardsID int64, client *github.Client, token string) ([]*issueOverview, error) {
	opt := &github.ProjectCardListOptions{}
	cards, _, err := client.Projects.ListProjectCards(context.Background(), cardsID, opt)
	if err != nil {
		fmt.Printf("error when querying cards %v", err)
		return nil, err
	}

	issues := make([]*issueOverview, 0)
	for _, c := range cards {
		if c.ContentURL == nil {
			continue
		}

		issueURL := *c.ContentURL
		issueDetail, err := getIssueDetail(issueURL, token)
		if err != nil {
			return nil, err
		}

		overview := issueOverview{
			url:   issueDetail.HTMLURL,
			id:    issueDetail.Number,
			title: cleanTitle(issueDetail.Title),
		}
		for _, v := range issueDetail.Labels {
			if strings.Contains(*v.Name, "sig/") {
				overview.sig = strings.Title(strings.ReplaceAll(*v.Name, "sig/", ""))
				if strings.EqualFold(overview.sig, "cli") {
					overview.sig = strings.ToUpper(overview.sig)
				}
				if strings.EqualFold(overview.sig, "cluster-lifecycle") {
					overview.sig = strings.ToLower(overview.sig)
				}
				break
			}
		}
		issues = append(issues, &overview)
	}

	return issues, nil
}

type IssueDetail struct {
	Number  int64          `json:"number"`
	HTMLURL string         `json:"html_url"`
	Title   string         `json:"title"`
	Labels  []github.Label `json:"labels,omitempty"`
}

func getIssueDetail(url, authToken string) (*IssueDetail, error) {
	// Create a Bearer string by appending string access token
	bearer := "Bearer " + authToken

	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// add authorization header to the req
	req.Header.Add("Authorization", bearer)

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
	var result IssueDetail
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func cleanTitle(title string) string {
	return strings.ReplaceAll(title, "[Failing Test]", "")
}

func printJobsStatistics(version string) {
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
	for _, kubeJob := range requiredJobs {
		resp, err := http.Get(fmt.Sprintf("https://testgrid.k8s.io/%s/summary", kubeJob.URLName))
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
		statistics.name = kubeJob.OutputName
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

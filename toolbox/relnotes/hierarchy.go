package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	u "k8s.io/release/toolbox/util"
)

type dictSIG map[string]dictArea
type dictArea map[string]dictIssue
type dictIssue map[int]dictPR
type dictPR map[int]bool

func hierarchicalNoteLayout(f *os.File, dict dictSIG, prMap map[int]*github.Issue, g *u.GithubClient, owner, repo string) error {
	// Sort and iterate through sigs
	var keySigs []string
	for sig := range dict {
		keySigs = append(keySigs, sig)
	}
	sort.Strings(keySigs)
	for _, sig := range keySigs {
		areas := dict[sig]
		f.WriteString(fmt.Sprintf(" - %s\n\n", strings.Title(sig)))

		// Sort and iterate through areas
		var keyAreas []string
		for area := range areas {
			keyAreas = append(keyAreas, area)
		}
		sort.Strings(keyAreas)
		for _, area := range keyAreas {
			issues := areas[area]
			f.WriteString(fmt.Sprintf("    - %s\n\n", strings.Title(area)))

			for issueID, prs := range issues {
				if issueID >= 0 {
					issue, _, err := g.GetIssue(context.Background(), owner, repo, issueID)
					if err != nil {
						return err
					}
					f.WriteString(fmt.Sprintf("        - %s (#%d)\n\n", *issue.Title, issueID))
				} else {
					f.WriteString(fmt.Sprintf("        - NullIssue\n\n"))
				}
				for pr := range prs {
					f.WriteString(fmt.Sprintf("            * %s (#%d, @%s)\n", extractReleaseNoteFromPR(prMap[pr]), pr, *prMap[pr].User.Login))
				}
				f.WriteString("\n")
			}
		}
	}
	return nil
}

// createHierarchicalNote given release PRs and issue map, creates hierarchical release note
// map[SIG]map[Area]map[Issue]PR.
func createHierarchicalNote(prs []int, prMap map[int]*github.Issue, g *u.GithubClient, owner, repo string) (dictSIG, error) {
	var dict = dictSIG{}

	for _, pr := range prs {
		issues, err := extractFixedIssues(*prMap[pr].Body, owner, repo, g)
		if err != nil {
			return nil, fmt.Errorf("failed to extract fixed issues: %v", err)
		}
		if len(issues) == 0 {
			// In our design doc, the automation should enforce every release-note PR
			// with at least one issue. However it's observed that the rule is not
			// applied yet. Also old PRs may not link to an issue.
			//
			// To produce info-richer release note, we try to get SIG and Area label
			// from PRs which don't link to any issue.
			issues = append(issues, prMap[pr])
		}
		for _, issue := range issues {
			sigs := extractIssueSIGs(issue)
			area := extractIssueArea(issue)
			i := *issue.Number
			// For PRs that don't link to any issue, restore the nullIssue information
			if issue.IsPullRequest() {
				i = -1
			}
			if len(sigs) == 0 {
				setNoteDict(dict, "nullSig", area, i, pr)
				continue
			}
			for _, s := range sigs {
				setNoteDict(dict, s, area, i, pr)
			}
		}
	}

	return dict, nil
}

// setNoteDict sets the entry dict[sig][area][issue][pr] to be true, initializes nil maps along
// the way.
func setNoteDict(dict dictSIG, sig, area string, issue, pr int) {
	if dict[sig] == nil {
		dict[sig] = dictArea{}
	}
	if dict[sig][area] == nil {
		dict[sig][area] = dictIssue{}
	}
	if dict[sig][area][issue] == nil {
		dict[sig][area][issue] = dictPR{}
	}
	dict[sig][area][issue][pr] = true
}

// extractFixedIssues parses the fixed issues' id from PR body.
func extractFixedIssues(msg, owner, repo string, g *u.GithubClient) ([]*github.Issue, error) {
	var issueIDs = make([]int, 0)
	var issues = make([]*github.Issue, 0)
	// Captures "fixes #<issue id>" and "fixes: #<issue id>"
	re, _ := regexp.Compile("fixes:* #([0-9]+)")
	matches := re.FindAllStringSubmatch(strings.ToLower(msg), -1)
	for _, match := range matches {
		id, _ := strconv.Atoi(match[1])
		issueIDs = append(issueIDs, id)
	}

	// Captures "ref #<issue id>" and "ref: #<issue id>"
	re, _ = regexp.Compile("ref:* #([0-9]+)")
	matches = re.FindAllStringSubmatch(strings.ToLower(msg), -1)
	for _, match := range matches {
		id, _ := strconv.Atoi(match[1])
		issueIDs = append(issueIDs, id)
	}

	// Captures "fixes https://github.com/kubernetes/kubernetes/issues/<issue id>" and "fixes: https://github.com/kubernetes/kubernetes/issues/<issue id>"
	re, _ = regexp.Compile("fixes:* https://github.com/kubernetes/kubernetes/issues/([0-9]+)")
	matches = re.FindAllStringSubmatch(strings.ToLower(msg), -1)
	for _, match := range matches {
		id, _ := strconv.Atoi(match[1])
		issueIDs = append(issueIDs, id)
	}

	// Captures "ref https://github.com/kubernetes/kubernetes/issues/<issue id>" and "ref: https://github.com/kubernetes/kubernetes/issues/<issue id>"
	re, _ = regexp.Compile("ref:* https://github.com/kubernetes/kubernetes/issues/([0-9]+)")
	matches = re.FindAllStringSubmatch(strings.ToLower(msg), -1)
	for _, match := range matches {
		id, _ := strconv.Atoi(match[1])
		issueIDs = append(issueIDs, id)
	}

	// Extract issues
	for _, id := range issueIDs {
		issue, _, err := g.GetIssue(context.Background(), owner, repo, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get issue %d: %v", id, err)
		}
		// Get rid of PRs
		if issue.IsPullRequest() {
			continue
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// extractIssueSIGs gets the SIGs of the input issue (if there is any)
func extractIssueSIGs(i *github.Issue) []string {
	var sigs = make([]string, 0)
	for _, l := range i.Labels {
		if strings.HasPrefix(*l.Name, "sig/") {
			sigs = append(sigs, (*l.Name)[4:])
		}
	}

	return sigs
}

// extractIssueSIGs gets the Areas of the input issue and returns as a single string. If the issue
// doesn't have any Area label, the function returns "nullArea".
func extractIssueArea(i *github.Issue) string {
	var areas = make([]string, 0)
	for _, l := range i.Labels {
		if strings.HasPrefix(*l.Name, "area/") {
			areas = append(areas, (*l.Name)[5:])
		}
	}

	if len(areas) == 0 {
		return "nullArea"
	}

	return strings.Join(areas, " & ")
}

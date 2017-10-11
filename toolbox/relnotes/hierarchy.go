package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
)

type dictSIG map[string]dictArea
type dictArea map[string]dictIssue
type dictIssue map[int]dictPR
type dictPR map[int]bool

func hierarchicalNoteLayout(f *os.File, dict dictSIG, issueMap map[int]*github.Issue) {
	for sig, areas := range dict {
		f.WriteString(fmt.Sprintf(" - %s\n\n", strings.Title(sig)))
		for area, issues := range areas {
			f.WriteString(fmt.Sprintf("    - %s\n\n", strings.Title(area)))
			for issue, prs := range issues {
				if issue >= 0 {
					f.WriteString(fmt.Sprintf("        - %s (#%d)\n", *issueMap[issue].Title, issue))
				} else {
					f.WriteString(fmt.Sprintf("        - NullIssue\n"))
				}
				for pr := range prs {
					f.WriteString(fmt.Sprintf("            * %s (#%d, @%s)\n", extractReleaseNote(issueMap[pr]), pr, *issueMap[pr].User.Login))
				}
				f.WriteString("\n")
			}
		}
	}
}

// createHierarchicalNote given release PRs and issue map, creates hierarchical release note
// map[SIG]map[Area]map[Issue]PR.
func createHierarchicalNote(prs []int, issueMap map[int]*github.Issue) dictSIG {
	var dict = dictSIG{}

	for _, pr := range prs {
		issues := extractFixedIssues(*issueMap[pr].Body)
		if len(issues) == 0 {
			// In our design doc, the automation should enforce every release-note PR
			// with at least one issue. However it's observed that the rule is not
			// applied yet. Also old PRs may not link to an issue.
			//
			// To produce info-richer release note, we try to get SIG and Area label
			// from PRs which don't link to any issue.
			issues = append(issues, pr)
		}
		for _, i := range issues {
			sigs := extractIssueSIGs(issueMap[i])
			area := extractIssueArea(issueMap[i])
			// For PRs that don't link to any issue, restore the nullIssue information
			if issueMap[i].PullRequestLinks != nil {
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

	return dict
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
func extractFixedIssues(msg string) []int {
	var issues = make([]int, 0)
	re, _ := regexp.Compile("fixes #([0-9]+)")
	matches := re.FindAllStringSubmatch(strings.ToLower(msg), -1)
	for _, match := range matches {
		id, _ := strconv.Atoi(match[1])
		issues = append(issues, id)
	}

	return issues
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

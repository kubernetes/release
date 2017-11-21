// Copyright 2017 The Kubernetes Authors All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	u "k8s.io/release/toolbox/util"
)

const (
	k8sReleaseURLPrefix = "https://dl.k8s.io"
	verDotzero          = "dotzero"
)

var (
	// Flags
	// TODO: golang flags and parameters syntax
	branch           = flag.String("branch", "", "Specify a branch other than the current one")
	documentURL      = flag.String("doc-url", "https://docs.k8s.io", "Documentation URL displayed in release notes")
	exampleURLPrefix = flag.String("example-url-prefix", "https://releases.k8s.io/", "Example URL prefix displayed in release notes")
	full             = flag.Bool("full", false, "Force 'full' release format to show all sections of release notes. "+
		"(This is the *default* for new branch X.Y.0 notes)")
	githubToken   = flag.String("github-token", "", "The file that contains Github token. Must be specified, or set the GITHUB_TOKEN environment variable.")
	htmlFileName  = flag.String("html-file", "", "Produce a html version of the notes")
	htmlizeMD     = flag.Bool("htmlize-md", false, "Output markdown with html for PRs and contributors (for use in CHANGELOG.md)")
	mdFileName    = flag.String("markdown-file", "", "Specify an alt file to use to store notes")
	owner         = flag.String("owner", "kubernetes", "Github owner or organization")
	preview       = flag.Bool("preview", false, "Report additional branch statistics (used for reporting outside of releases)")
	quiet         = flag.Bool("quiet", false, "Don't display the notes when done")
	releaseBucket = flag.String("release-bucket", "kubernetes-release", "Specify Google Storage bucket to point to in generated notes (informational only)")
	releaseTars   = flag.String("release-tars", "", "Directory of tars to sha256 sum for display")
	repo          = flag.String("repo", "kubernetes", "Github repository")

	// Global
	branchHead      = ""
	branchVerSuffix = "" // e.g. branch: "release-1.8", branchVerSuffix: "-1.8"
)

// ReleaseInfo contains release related information to generate a release note.
// NOTE: the prMap only includes PRs with "release-note" label.
type ReleaseInfo struct {
	startTag, releaseTag     string
	prMap                    map[int]*github.Issue
	releasePRs               []int
	releaseActionRequiredPRs []int
}

func main() {
	// Initialization
	flag.Parse()
	branchRange := flag.Arg(0)
	startingTime := time.Now().Round(time.Second)

	log.Printf("Boolean flags: full: %v, htmlize-md: %v, preview: %v, quiet: %v", *full, *htmlizeMD, *preview, *quiet)
	log.Printf("Input branch range: %s", branchRange)

	if *branch == "" {
		// If branch isn't specified in flag, use current branch
		var err error
		*branch, err = u.GetCurrentBranch()
		if err != nil {
			log.Printf("failed to get current branch: %v", err)
			os.Exit(1)
		}
	}
	branchVerSuffix = strings.TrimPrefix(*branch, "release")
	log.Printf("Working branch: %s. Branch version suffix: %s.", *branch, branchVerSuffix)

	prFileName := fmt.Sprintf("/tmp/release-notes-%s-prnotes", *branch)
	if *mdFileName == "" {
		*mdFileName = fmt.Sprintf("/tmp/release-notes-%s.md", *branch)
	}
	log.Printf("Output markdown file path: %s", *mdFileName)
	if *htmlFileName != "" {
		log.Printf("Output HTML file path: %s", *htmlFileName)
	}

	if *githubToken == "" {
		// If githubToken isn't specified in flag, use the GITHUB_TOKEN environment variable
		*githubToken = os.Getenv("GITHUB_TOKEN")
	} else {
		token, err := u.ReadToken(*githubToken)
		if err != nil {
			log.Printf("failed to read Github token: %v", err)
			os.Exit(1)
		}
		*githubToken = token
	}
	// Github token must be provided to ensure great rate limit experience
	if *githubToken == "" {
		log.Print("Github token not provided. Exiting now...")
		os.Exit(1)
	}
	client := u.NewClient(*githubToken)

	// End of initialization

	// Gather release related information including startTag, releaseTag, prMap and releasePRs
	releaseInfo, err := gatherReleaseInfo(client, branchRange)
	if err != nil {
		log.Printf("failed to gather release related information: %v", err)
		os.Exit(1)
	}

	// Generating release note...
	log.Print("Generating release notes...")
	err = gatherPRNotes(prFileName, releaseInfo)
	if err != nil {
		log.Printf("failed to gather PR notes: %v", err)
		os.Exit(1)
	}

	// Start generating markdown file
	log.Print("Preparing layout...")
	err = generateMDFile(client, releaseInfo.releaseTag, prFileName)
	if err != nil {
		log.Printf("failed to generate markdown file: %v", err)
		os.Exit(1)
	}

	if *htmlizeMD && !u.IsVer(releaseInfo.releaseTag, verDotzero) {
		// HTML-ize markdown file
		// Make users and PRs linkable
		// Also, expand anchors (needed for email announce())
		projectGithubURL := fmt.Sprintf("https://github.com/%s/%s", *owner, *repo)
		_, err = u.Shell("sed", "-i", "-e", "s,#\\([0-9]\\{5\\,\\}\\),[#\\1]("+projectGithubURL+"/pull/\\1),g",
			"-e", "s,\\(#v[0-9]\\{3\\}-\\),"+projectGithubURL+"/blob/master/CHANGELOG"+branchVerSuffix+".md\\1,g",
			"-e", "s,@\\([a-zA-Z0-9-]*\\),[@\\1](https://github.com/\\1),g", *mdFileName)

		if err != nil {
			log.Printf("failed to htmlize markdown file: %v", err)
			os.Exit(1)
		}
	}

	if *preview && *owner == "kubernetes" && *repo == "kubernetes" {
		// If in preview mode, get the current CI job status
		// We do this after htmlizing because we don't want to update the
		// issues in the block of this section
		//
		// NOTE: this function is Kubernetes-specified and runs the find_green_build script under
		// kubernetes/release. Make sure you have the dependencies installed for find_green_build
		// before running this function.
		err = getCIJobStatus(*mdFileName, *branch, *htmlizeMD)
		if err != nil {
			log.Printf("failed to get CI status: %v", err)
			os.Exit(1)
		}
	}

	if *htmlFileName != "" {
		// If HTML file name is given, generate HTML release note
		err = createHTMLNote(*htmlFileName, *mdFileName)
		if err != nil {
			log.Printf("failed to generate HTML release note: %v", err)
			os.Exit(1)
		}
	}

	if !*quiet {
		// If --quiet flag is not specified, print the markdown release note to stdout
		log.Print("Displaying the markdown release note to stdout...")
		dat, err := ioutil.ReadFile(*mdFileName)
		if err != nil {
			log.Printf("failed to read markdown release note: %v", err)
			os.Exit(1)
		}
		fmt.Print(string(dat))
	}

	log.Printf("Successfully generated release note. Total running time: %s", time.Now().Round(time.Second).Sub(startingTime).String())

	return
}

func gatherReleaseInfo(g *u.GithubClient, branchRange string) (*ReleaseInfo, error) {
	var info ReleaseInfo
	log.Print("Gathering release commits from Github...")
	// Get release related commits on the release branch within release range
	releaseCommits, startTag, releaseTag, err := getReleaseCommits(g, *owner, *repo, *branch, branchRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get release commits for %s: %v", branchRange, err)
	}
	info.startTag = startTag
	info.releaseTag = releaseTag

	// Parse release related PR ids from the release commits
	commitPRs, err := parsePRFromCommit(releaseCommits)
	if err != nil {
		return nil, fmt.Errorf("failed to parse release commits: %v", err)
	}

	log.Print("Gathering \"release-note\" labelled PRs using Github search API. This may take a while...")
	var query []string
	query = u.AddQuery(query, "repo", *owner, "/", *repo)
	query = u.AddQuery(query, "type", "pr")
	query = u.AddQuery(query, "label", "release-note")
	releaseNotePRs, err := g.SearchIssues(strings.Join(query, " "))
	if err != nil {
		return nil, fmt.Errorf("failed to search release-note labelled PRs: %v", err)
	}
	log.Print("\"release-note\" labelled PRs gathered.")

	log.Print("Gathering \"release-note-action-required\" labelled PRs using Github search API.")
	query = nil
	query = u.AddQuery(query, "repo", *owner, "/", *repo)
	query = u.AddQuery(query, "type", "pr")
	query = u.AddQuery(query, "label", "release-note-action-required")
	releaseNoteActionRequiredPRs, err := g.SearchIssues(strings.Join(query, " "))
	if err != nil {
		return nil, fmt.Errorf("failed to search release-note-action-required labelled PRs: %v", err)
	}
	log.Print("\"release-note-action-required\" labelled PRs gathered.")

	info.prMap = make(map[int]*github.Issue)
	for _, i := range releaseNotePRs {
		ptr := new(github.Issue)
		*ptr = i
		info.prMap[*ptr.Number] = ptr
	}

	actionRequiredPRMap := make(map[int]*github.Issue)
	for _, i := range releaseNoteActionRequiredPRs {
		ptr := new(github.Issue)
		*ptr = i
		actionRequiredPRMap[*ptr.Number] = ptr
	}

	// Get release note PRs by examining release-note label on commit PRs
	info.releasePRs = make([]int, 0)
	for _, pr := range commitPRs {
		if info.prMap[pr] != nil {
			info.releasePRs = append(info.releasePRs, pr)
		}
		if actionRequiredPRMap[pr] != nil {
			info.releaseActionRequiredPRs = append(info.releaseActionRequiredPRs, pr)
		}
	}

	for k, v := range actionRequiredPRMap {
		info.prMap[k] = v
	}

	return &info, nil
}

func gatherPRNotes(prFileName string, info *ReleaseInfo) error {
	var result error
	prFile, err := os.Create(prFileName)
	if err != nil {
		return fmt.Errorf("failed to create release note file %s: %v", prFileName, err)
	}
	defer func() {
		if err = prFile.Close(); err != nil {
			result = fmt.Errorf("failed to close file %s, %v", prFileName, err)
		}
	}()

	// Bootstrap notes for minor (new branch) releases
	if *full || u.IsVer(info.releaseTag, verDotzero) {
		draftURL := fmt.Sprintf("%s%s/features/master/%s/release-notes-draft.md", u.GithubRawURL, *owner, *branch)
		changelogURL := fmt.Sprintf("%s%s/%s/master/CHANGELOG%s.md", u.GithubRawURL, *owner, *repo, branchVerSuffix)
		minorRelease(prFile, info.releaseTag, draftURL, changelogURL)
	} else {
		patchRelease(prFile, info)
	}
	return result
}

func generateMDFile(g *u.GithubClient, releaseTag, prFileName string) error {
	var result error
	mdFile, err := os.Create(*mdFileName)
	if err != nil {
		return fmt.Errorf("failed to create release note markdown file %s: %v", *mdFileName, err)
	}
	defer func() {
		if err = mdFile.Close(); err != nil {
			result = fmt.Errorf("failed to close file %s, %v", *mdFileName, err)
		}
	}()

	// Create markdown file body with documentation and example URLs from program flags
	exampleURL := fmt.Sprintf("%s%s/examples", *exampleURLPrefix, *branch)
	err = createBody(mdFile, releaseTag, *branch, *documentURL, exampleURL, *releaseTars)
	if err != nil {
		return fmt.Errorf("failed to create file body: %v", err)
	}

	// Copy (append) the pull request notes into the output markdown file
	dat, err := ioutil.ReadFile(prFileName)
	if err != nil {
		return fmt.Errorf("failed to copy from PR file to release note markdown file: %v", err)
	}
	mdFile.WriteString(string(dat))

	if *preview {
		// If in preview mode, get the pending PRs
		err = getPendingPRs(g, mdFile, *owner, *repo, *branch)
		if err != nil {
			return fmt.Errorf("failed to get pending PRs: %v", err)
		}
	}
	return result
}

// getPendingPRs gets pending PRs on given branch in the repo.
func getPendingPRs(g *u.GithubClient, f *os.File, owner, repo, branch string) error {
	log.Print("Getting pending PR status...")
	f.WriteString("-------\n")
	f.WriteString(fmt.Sprintf("## PENDING PRs on the %s branch\n", branch))

	if *htmlizeMD {
		f.WriteString("PR | Milestone | User | Date | Commit Message\n")
		f.WriteString("-- | --------- | ---- | ---- | --------------\n")
	}

	var query []string
	query = u.AddQuery(query, "repo", owner, "/", repo)
	query = u.AddQuery(query, "is", "open")
	query = u.AddQuery(query, "type", "pr")
	query = u.AddQuery(query, "base", branch)
	pendingPRs, err := g.SearchIssues(strings.Join(query, " "))
	if err != nil {
		return fmt.Errorf("failed to search pending PRs: %v", err)
	}

	for _, pr := range pendingPRs {
		var str string
		// escape '*' in commit messages so they don't mess up formatting
		msg := strings.Replace(*pr.Title, "*", "", -1)
		milestone := "null"
		if pr.Milestone != nil {
			milestone = *pr.Milestone.Title
		}
		if *htmlizeMD {
			str = fmt.Sprintf("#%-8d | %-4s | @%-10s| %s   | %s\n", *pr.Number, milestone, *pr.User.Login, pr.UpdatedAt.Format("Mon Jan  2 15:04:05 MST 2006"), msg)
		} else {
			str = fmt.Sprintf("#%-8d  %-4s  @%-10s %s    %s\n", *pr.Number, milestone, *pr.User.Login, pr.UpdatedAt.Format("Mon Jan  2 15:04:05 MST 2006"), msg)
		}
		f.WriteString(str)
	}
	f.WriteString("\n\n")
	return nil
}

// createHTMLNote generates HTML release note based on the input markdown release note.
func createHTMLNote(htmlFileName, mdFileName string) error {
	var result error
	log.Print("Generating HTML release note...")
	cssFileName := "/tmp/release_note_cssfile"
	cssFile, err := os.Create(cssFileName)
	if err != nil {
		return fmt.Errorf("failed to create css file %s: %v", cssFileName, err)
	}

	cssFile.WriteString("<style type=text/css> ")
	cssFile.WriteString("table,th,tr,td {border: 1px solid gray; ")
	cssFile.WriteString("border-collapse: collapse;padding: 5px;} ")
	cssFile.WriteString("</style>")
	// Here we manually close the css file instead of defer the close function,
	// because we need to use the css file for pandoc command below.
	// Writing to css file is a clear small logic so we don't separate it into
	// another function.
	if err = cssFile.Close(); err != nil {
		return fmt.Errorf("failed to close file %s, %v", cssFileName, err)
	}

	htmlStr, err := u.Shell("pandoc", "-H", cssFileName, "--from", "markdown_github", "--to", "html", mdFileName)
	if err != nil {
		return fmt.Errorf("failed to generate html content: %v", err)
	}

	htmlFile, err := os.Create(htmlFileName)
	if err != nil {
		return fmt.Errorf("failed to create html file: %v", err)
	}
	defer func() {
		if err = htmlFile.Close(); err != nil {
			result = fmt.Errorf("failed to close file %s, %v", htmlFileName, err)
		}
	}()

	htmlFile.WriteString(htmlStr)
	return result
}

// getCIJobStatus runs the script find_green_build and append CI job status to outputFile.
// NOTE: this function is Kubernetes-specified and runs the find_green_build script under
// kubernetes/release. Make sure you have the dependencies installed for find_green_build
// before running this function.
func getCIJobStatus(outputFile, branch string, htmlize bool) error {
	var result error
	log.Print("Getting CI job status (this may take a while)...")

	red := "<span style=\"color:red\">"
	green := "<span style=\"color:green\">"
	off := "</span>"

	if htmlize {
		red = "<FONT COLOR=RED>"
		green = "<FONT COLOR=GREEN>"
		off = "</FONT>"
	}

	var extraFlag string

	if strings.Contains(branch, "release-") {
		// If working on a release branch assume --official for the purpose of displaying
		// find_green_build output
		extraFlag = "--official"
	} else {
		// For master branch, limit the analysis to 30 primary ci jobs. This is necessary
		// due to the recently expanded blocking test list for master. The expanded test
		// list is often unable to find a complete passing set and find_green_build runs
		// unbounded for hours
		extraFlag = "--limit=30"
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			result = fmt.Errorf("failed to close file %s, %v", outputFile, err)
		}
	}()

	f.WriteString(fmt.Sprintf("## State of %s branch\n", branch))

	// Call script find_green_build to get CI job status
	content, err := u.Shell(os.Getenv("GOPATH")+"/src/k8s.io/release/find_green_build", "-v", extraFlag, branch)
	if err == nil {
		f.WriteString(fmt.Sprintf("%sGOOD TO GO!%s\n\n", green, off))
	} else {
		f.WriteString(fmt.Sprintf("%sNOT READY%s\n\n", red, off))
	}

	f.WriteString("### Details\n```\n")
	f.WriteString(content)
	f.WriteString("```\n")

	log.Print("CI job status fetched.")
	return result
}

// createBody creates the general documentation, example and downloads table body for the
// markdown file.
func createBody(f *os.File, releaseTag, branch, docURL, exampleURL, releaseTars string) error {
	var title string
	if *preview {
		title = "Branch "
	}

	if releaseTag == "HEAD" || releaseTag == branchHead {
		title += branch
	} else {
		title += releaseTag
	}

	if *preview {
		f.WriteString(fmt.Sprintf("**Release Note Preview - generated on %s**\n", time.Now().Format("Mon Jan  2 15:04:05 MST 2006")))
	}

	f.WriteString(fmt.Sprintf("\n# %s\n\n", title))
	f.WriteString(fmt.Sprintf("[Documentation](%s) & [Examples](%s)\n\n", docURL, exampleURL))

	if releaseTars != "" {
		f.WriteString(fmt.Sprintf("## Downloads for %s\n\n", title))
		tables := []struct {
			heading  string
			filename []string
		}{
			{"", []string{releaseTars + "/kubernetes.tar.gz", releaseTars + "/kubernetes-src.tar.gz"}},
			{"Client Binaries", []string{releaseTars + "/kubernetes-client*.tar.gz"}},
			{"Server Binaries", []string{releaseTars + "/kubernetes-server*.tar.gz"}},
			{"Node Binaries", []string{releaseTars + "/kubernetes-node*.tar.gz"}},
		}

		for _, table := range tables {
			err := createDownloadsTable(f, releaseTag, table.heading, table.filename...)
			if err != nil {
				return fmt.Errorf("failed to create downloads table: %v", err)
			}
		}
		f.WriteString("\n")
	}
	return nil
}

// createDownloadTable creates table of download link and sha256 hash for given file.
func createDownloadsTable(f *os.File, releaseTag, heading string, filename ...string) error {
	var urlPrefix string

	if *releaseBucket == "kubernetes-release" {
		urlPrefix = k8sReleaseURLPrefix
	} else {
		urlPrefix = fmt.Sprintf("https://storage.googleapis.com/%s/release", *releaseBucket)
	}

	if *releaseBucket == "" {
		log.Print("NOTE: empty Google Storage bucket specified. Please specify valid bucket using \"release-bucket\" flag.")
	}

	if heading != "" {
		f.WriteString(fmt.Sprintf("\n### %s\n", heading))
	}

	f.WriteString("\n")
	f.WriteString("filename | sha256 hash\n")
	f.WriteString("-------- | -----------\n")

	files := make([]string, 0)
	for _, name := range filename {
		fs, _ := filepath.Glob(name)
		for _, v := range fs {
			files = append(files, v)
		}
	}

	for _, file := range files {
		fn := filepath.Base(file)
		sha, err := u.GetSha256(file)
		if err != nil {
			return fmt.Errorf("failed to calc SHA256 of file %s: %v", file, err)
		}
		f.WriteString(fmt.Sprintf("[%s](%s/%s/%s) | `%s`\n", fn, urlPrefix, releaseTag, fn, sha))
	}
	return nil
}

// minorReleases performs a minor (vX.Y.0) release by fetching the release template and aggregate
// previous release in series.
func minorRelease(f *os.File, release, draftURL, changelogURL string) {
	// Check for draft and use it if available
	log.Printf("Checking if draft release notes exist for %s...", release)

	resp, err := http.Get(draftURL)
	if err == nil {
		defer resp.Body.Close()
	}

	if err == nil && resp.StatusCode == 200 {
		log.Print("Draft found - using for release notes...")
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			log.Printf("error during copy to file: %v", err)
			return
		}
		f.WriteString("\n")
	} else {
		log.Print("Failed to find draft - creating generic template... (error message/status code printed below)")
		if err != nil {
			log.Printf("Error message: %v", err)
		} else {
			log.Printf("Response status code: %d", resp.StatusCode)
		}
		f.WriteString("## Major Themes\n\n* TBD\n\n## Other notable improvements\n\n* TBD\n\n## Known Issues\n\n* TBD\n\n## Provider-specific Notes\n\n* TBD\n\n")
	}

	// Aggregate all previous release in series
	f.WriteString(fmt.Sprintf("### Previous Release Included in %s\n\n", release))

	// Regexp Example:
	// Assume the release tag is v1.7.0, this regexp matches "- [v1.7.0-" in
	//     "- [v1.7.0-rc.1](#v170-rc1)"
	//     "- [v1.7.0-beta.2](#v170-beta2)"
	//     "- [v1.7.0-alpha.3](#v170-alpha3)"
	reAnchor, _ := regexp.Compile(fmt.Sprintf("- \\[%s-", release))

	resp, err = http.Get(changelogURL)
	if err == nil {
		defer resp.Body.Close()
	}

	if err == nil && resp.StatusCode == 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		for _, line := range strings.Split(buf.String(), "\n") {
			if anchor := reAnchor.FindStringSubmatch(line); anchor != nil {
				f.WriteString(line + "\n")
			}
		}
		f.WriteString("\n")
	} else {
		log.Print("Failed to fetch past changelog for minor release - continuing... (error message/status code printed below)")
		if err != nil {
			log.Printf("Error message: %v", err)
		} else {
			log.Printf("Response status code: %d", resp.StatusCode)
		}
	}
}

// patchRelease performs a patch (vX.Y.Z) release by printing out all the related changes.
func patchRelease(f *os.File, info *ReleaseInfo) {
	// Release note for different labels
	f.WriteString(fmt.Sprintf("## Changelog since %s\n\n", info.startTag))

	if len(info.releaseActionRequiredPRs) > 0 {
		f.WriteString("### Action Required\n\n")
		for _, pr := range info.releaseActionRequiredPRs {
			f.WriteString(fmt.Sprintf("* %s (#%d, @%s)\n", extractReleaseNoteFromPR(info.prMap[pr]), pr, *info.prMap[pr].User.Login))
		}
		f.WriteString("\n")
	}

	if len(info.releasePRs) > 0 {
		f.WriteString("### Other notable changes\n\n")
		for _, pr := range info.releasePRs {
			f.WriteString(fmt.Sprintf("* %s (#%d, @%s)\n", extractReleaseNoteFromPR(info.prMap[pr]), pr, *info.prMap[pr].User.Login))
		}
		f.WriteString("\n")
	} else {
		f.WriteString("**No notable changes for this release**\n\n")
	}
}

// extractReleaseNoteFromPR tries to fetch release note from PR body, otherwise uses PR title.
func extractReleaseNoteFromPR(pr *github.Issue) string {
	// Regexp Example:
	// This regexp matches the release note section in Kubernetes pull request template:
	// https://github.com/kubernetes/kubernetes/blob/master/.github/PULL_REQUEST_TEMPLATE.md
	re, _ := regexp.Compile("```release-note\r\n(.+)\r\n```")
	if note := re.FindStringSubmatch(*pr.Body); note != nil {
		return note[1]
	}
	return *pr.Title
}

// determineRange examines a Git branch range in the format of [[startTag..]endTag], and
// determines a valid range. For example:
//
//     ""                       - last release to HEAD on the branch
//     "v1.1.4.."               - v1.1.4 to HEAD
//     "v1.1.4..v1.1.7"         - v1.1.4 to v1.1.7
//     "v1.1.7"                 - last release on the branch to v1.1.7
//
// NOTE: the input branch must be the corresponding release branch w.r.t. input range. For example:
//
//     Getting "v1.1.4..v1.1.7" on branch "release-1.1" makes sense
//     Getting "v1.1.4..v1.1.7" on branch "release-1.2" doesn't
func determineRange(g *u.GithubClient, owner, repo, branch, branchRange string) (startTag, releaseTag string, err error) {
	b, _, err := g.GetBranch(context.Background(), owner, repo, branch)
	if err != nil {
		return "", "", err
	}
	branchHead = *b.Commit.SHA

	lastRelease, err := g.LastReleases(owner, repo)
	if err != nil {
		return "", "", err
	}

	// If lastRelease[branch] is unset, attempt to get the last release from the parent branch
	// and then master
	if i := strings.LastIndex(branch, "."); lastRelease[branch] == "" && i != -1 {
		lastRelease[branch] = lastRelease[branch[:i]]
	}
	if lastRelease[branch] == "" {
		lastRelease[branch] = lastRelease["master"]
	}

	// Regexp Example:
	// This regexp matches the Git branch range in the format of [[startTag..]endTag]. For example:
	//
	//     ""
	//     "v1.1.4.."
	//     "v1.1.4..v1.1.7"
	//     "v1.1.7"
	re, _ := regexp.Compile("([v0-9.]*-*(alpha|beta|rc)*\\.*[0-9]*)\\.\\.([v0-9.]*-*(alpha|beta|rc)*\\.*[0-9]*)$")
	tags := re.FindStringSubmatch(branchRange)
	if tags != nil {
		startTag = tags[1]
		releaseTag = tags[3]
	} else {
		startTag = lastRelease[branch]
		releaseTag = branchHead
	}

	if startTag == "" {
		return "", "", fmt.Errorf("unable to set beginning of range automatically")
	}
	if releaseTag == "" {
		releaseTag = branchHead
	}

	return startTag, releaseTag, nil
}

// getReleaseCommits given a Git branch range in the format of [[startTag..]endTag], determines
// a valid range and returns all the commits on the branch in that range.
func getReleaseCommits(g *u.GithubClient, owner, repo, branch, branchRange string) ([]*github.RepositoryCommit, string, string, error) {
	// Get start and release tag/commit based on input branch range
	startTag, releaseTag, err := determineRange(g, owner, repo, branch, branchRange)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to determine branch range: %v", err)
	}

	// Get all tags in the repository
	tags, err := g.ListAllTags(owner, repo)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to fetch repo tags: %v", err)
	}

	// Get commits for specified branch and range
	tStart, err := g.GetCommitDate(owner, repo, startTag, tags)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get start commit date for %s: %v", startTag, err)
	}
	tEnd, err := g.GetCommitDate(owner, repo, releaseTag, tags)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get release commit date for %s: %v", releaseTag, err)
	}

	releaseCommits, err := g.ListAllCommits(owner, repo, branch, tStart, tEnd)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to fetch release repo commits: %v", err)
	}

	return releaseCommits, startTag, releaseTag, nil
}

// parsePRFromCommit goes through commit messages, and parse PR IDs for normal pull requests as
// well as cherry picks.
func parsePRFromCommit(commits []*github.RepositoryCommit) ([]int, error) {
	prs := make([]int, 0)
	prsMap := make(map[int]bool)

	// Regexp example:
	// This regexp matches (Note that it supports multiple-source cherry pick)
	//
	// "automated-cherry-pick-of-#12345-#23412-"
	// "automated-cherry-pick-of-#23791-"
	reCherry, _ := regexp.Compile("automated-cherry-pick-of-(#[0-9]+-){1,}")
	reCherryID, _ := regexp.Compile("#([0-9]+)-")
	reMerge, _ := regexp.Compile("^Merge pull request #([0-9]+) from")

	for _, c := range commits {
		// Deref all PRs back to master
		// Match cherry pick PRs first and then normal pull requests
		// Paying special attention to automated cherrypicks that could have multiple
		// sources
		if cpStr := reCherry.FindStringSubmatch(*c.Commit.Message); cpStr != nil {
			cpPRs := reCherryID.FindAllStringSubmatch(cpStr[0], -1)
			for _, pr := range cpPRs {
				id, err := strconv.Atoi(pr[1])
				if err != nil {
					return nil, err
				}
				if prsMap[id] == false {
					prs = append(prs, id)
					prsMap[id] = true
				}
			}
		} else if pr := reMerge.FindStringSubmatch(*c.Commit.Message); pr != nil {
			id, err := strconv.Atoi(pr[1])
			if err != nil {
				return nil, err
			}
			if prsMap[id] == false {
				prs = append(prs, id)
				prsMap[id] = true
			}
		}
	}

	return prs, nil
}

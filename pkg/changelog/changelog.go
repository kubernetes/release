/*
Copyright 2020 The Kubernetes Authors.

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

package changelog

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-sdk/github"
	"sigs.k8s.io/release-utils/util"

	"k8s.io/release/pkg/notes/options"
)

// Options are the main settings for generating the changelog.
type Options struct {
	RepoPath     string
	Tag          string
	Branch       string
	Bucket       string
	Tars         string
	Images       string
	HTMLFile     string
	JSONFile     string
	RecordDir    string
	ReplayDir    string
	CVEDataDir   string
	CloneCVEMaps bool
	Dependencies bool
}

// Changelog can be used to generate the changelog for a release.
type Changelog struct {
	options *Options
	impl
}

// New creates a new Changelog instance.
func New(opts *Options) *Changelog {
	return &Changelog{
		options: opts,
		impl:    &defaultImpl{},
	}
}

// SetImpl can be used to set the internal implementation.
func (c *Changelog) SetImpl(impl impl) {
	c.impl = impl
}

// Run starts the changelog generation.
func (c *Changelog) Run() error {
	tag, err := c.TagStringToSemver(c.options.Tag)
	if err != nil {
		return fmt.Errorf("parse tag %s: %w", c.options.Tag, err)
	}

	// Automatically set the branch to a release branch if not provided
	branch := c.options.Branch
	if branch == "" {
		branch = fmt.Sprintf("release-%d.%d", tag.Major, tag.Minor)
	}

	logrus.Infof("Using release branch %s", branch)

	logrus.Infof("Using local repository path %s", c.options.RepoPath)

	repo, err := c.OpenRepo(c.options.RepoPath)
	if err != nil {
		return fmt.Errorf("open expected k/k repository %q: %w", c.options.RepoPath, err)
	}

	if currentBranch, err := c.CurrentBranch(repo); err == nil {
		logrus.Infof("We're currently on branch: %s", currentBranch)
	}

	remoteBranch := git.Remotify(branch)

	head, err := c.RevParseTag(repo, remoteBranch)
	if err != nil {
		return fmt.Errorf("get latest branch commit: %w", err)
	}

	logrus.Infof("Found latest %s commit %s", remoteBranch, head)

	var markdown, jsonStr, startRev, endRev string

	if tag.Patch == 0 {
		if len(tag.Pre) == 0 { //nolint:gocritic // a switch case would not make it better
			// Still create the downloads table
			downloadsTable := &bytes.Buffer{}
			startTag := util.SemverToTagString(semver.Version{
				Major: tag.Major, Minor: tag.Minor - 1, Patch: 0,
			})

			startRev = startTag
			endRev = head

			if err := c.CreateDownloadsTable(
				downloadsTable, c.options.Bucket, c.options.Tars,
				c.options.Images, startRev, c.options.Tag,
			); err != nil {
				return fmt.Errorf("create downloads table: %w", err)
			}

			// New final minor versions should have remote release notes
			markdown, jsonStr, err = c.lookupRemoteReleaseNotes(branch)
			markdown = downloadsTable.String() + markdown
		} else if tag.Pre[0].String() == "alpha" && tag.Pre[1].VersionNum == 1 {
			// v1.x.0-alpha.1 releases use the previous minor as start commit.
			// Those are usually the first releases being cut on master after
			// the previous final has been released.
			startRev = util.SemverToTagString(semver.Version{
				Major: tag.Major, Minor: tag.Minor - 1, Patch: 0,
			})
			logrus.Infof("Using previous minor %s as start tag", startRev)

			// The end tag does not yet exist which means that we stick to
			// the current HEAD as end revision.
			endRev = head

			markdown, jsonStr, err = c.generateReleaseNotes(branch, startRev, endRev)
		} else {
			// New minor alpha, beta and rc releases get generated notes
			var latestTags github.TagsPerBranch
			if c.options.ReplayDir != "" {
				// Do not access the API on replay
				latestTags = github.TagsPerBranch{branch: c.options.Tag}
			} else {
				latestTags, err = c.LatestGitHubTagsPerBranch()
				if err != nil {
					return fmt.Errorf("get latest GitHub tags: %w", err)
				}
			}

			if startTag, ok := latestTags[branch]; ok {
				logrus.Infof("Found start tag %s", startTag)

				// The end tag does not yet exist which means that we stick to
				// the current HEAD as end revision.
				startRev = startTag
				endRev = head

				markdown, jsonStr, err = c.generateReleaseNotes(branch, startRev, endRev)
			} else {
				return fmt.Errorf(
					"no latest tag available for branch %s", branch,
				)
			}
		}
	} else {
		if c.options.CloneCVEMaps {
			cveDir, err := c.CloneCVEData()
			if err != nil {
				return fmt.Errorf("getting cve data maps: %w", err)
			}

			c.options.CVEDataDir = cveDir
		}

		// A patch version, let’s just use the previous patch
		startTag := util.SemverToTagString(semver.Version{
			Major: tag.Major, Minor: tag.Minor, Patch: tag.Patch - 1,
		})

		startRev = startTag
		endRev = head

		markdown, jsonStr, err = c.generateReleaseNotes(branch, startTag, endRev)
	}

	if err != nil {
		return fmt.Errorf("generate release notes: %w", err)
	}

	if c.options.Dependencies {
		logrus.Info("Generating dependency changes")

		deps, err := c.DependencyChanges(startRev, endRev)
		if err != nil {
			return fmt.Errorf("generate dependency changes: %w", err)
		}

		markdown += strings.Repeat(nl, 2) + deps
	}

	logrus.Info("Generating TOC")

	toc, err := c.GenerateTOC(markdown)
	if err != nil {
		return fmt.Errorf("generate table of contents: %w", err)
	}

	// Restore the currently checked out branch
	currentBranch, err := c.CurrentBranch(repo)
	if err != nil {
		return fmt.Errorf("get current branch: %w", err)
	}

	if currentBranch != "" {
		defer func() {
			if err := c.Checkout(repo, currentBranch); err != nil {
				logrus.Errorf("Restore branch %s: %v", currentBranch, err)
			}
		}()
	}

	logrus.Infof("Checking out %s branch", git.DefaultBranch)

	if err := c.Checkout(repo, git.DefaultBranch); err != nil {
		return fmt.Errorf("checkout %s branch: %w", git.DefaultBranch, err)
	}

	logrus.Info("Writing markdown")

	if err := c.writeMarkdown(repo, toc, markdown, tag); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}

	logrus.Info("Writing HTML")

	if err := c.writeHTML(tag, markdown); err != nil {
		return fmt.Errorf("write HTML: %w", err)
	}

	logrus.Info("Writing JSON")

	if err := c.writeJSON(tag, jsonStr); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	logrus.Info("Committing changes")

	if err := c.commitChanges(repo, branch, tag); err != nil {
		return fmt.Errorf("commit changes: %w", err)
	}

	return nil
}

func (c *Changelog) generateReleaseNotes(
	branch, startRev, endRev string,
) (markdown, jsonStr string, err error) {
	logrus.Info("Generating release notes")

	notesOptions := options.New()
	notesOptions.Branch = branch
	notesOptions.StartRev = startRev
	notesOptions.EndSHA = endRev
	notesOptions.RepoPath = c.options.RepoPath
	notesOptions.ReleaseBucket = c.options.Bucket
	notesOptions.ReleaseTars = c.options.Tars
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOptions.RecordDir = c.options.RecordDir
	notesOptions.ReplayDir = c.options.ReplayDir
	notesOptions.Pull = false
	notesOptions.AddMarkdownLinks = true

	if c.options.CVEDataDir != "" {
		notesOptions.MapProviderStrings = append(
			notesOptions.MapProviderStrings, c.options.CVEDataDir,
		)
	}

	if err := c.ValidateAndFinish(notesOptions); err != nil {
		return "", "", fmt.Errorf("validating notes options: %w", err)
	}

	releaseNotes, err := c.GatherReleaseNotes(notesOptions)
	if err != nil {
		return "", "", fmt.Errorf("gather release notes: %w", err)
	}

	doc, err := c.NewDocument(releaseNotes, startRev, c.options.Tag)
	if err != nil {
		return "", "", fmt.Errorf("create release note document: %w", err)
	}

	releaseNotesJSON, err := json.MarshalIndent(releaseNotes.ByPR(), "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("build release notes JSON: %w", err)
	}

	markdown, err = c.RenderMarkdownTemplate(
		doc, c.options.Bucket, c.options.Tars, c.options.Images,
		options.GoTemplateInline+releaseNotesTemplate,
	)
	if err != nil {
		return "", "", fmt.Errorf("render release notes to markdown: %w", err)
	}

	return markdown, string(releaseNotesJSON), nil
}

func (c *Changelog) writeMarkdown(
	repo *git.Repo, toc, markdown string, tag semver.Version,
) error {
	changelogPath := filepath.Join(
		c.RepoDir(repo),
		markdownChangelogFilename(tag),
	)
	writeFile := func(t, m string) error {
		return c.WriteFile(
			changelogPath,
			[]byte(addTocMarkers(t)+"\n"+strings.TrimSpace(m)),
			os.FileMode(0o644),
		)
	}

	// No changelog exists, simply write the content to a new one
	if _, err := c.Stat(changelogPath); os.IsNotExist(err) {
		logrus.Infof("Changelog %q does not exist, creating it", changelogPath)

		if err := c.adaptChangelogReadmeFile(repo, tag); err != nil {
			return fmt.Errorf("adapt changelog readme: %w", err)
		}

		return writeFile(toc, markdown)
	}

	// Changelog seems to exist, prepend the notes and re-generate the TOC
	logrus.Infof("Adding new content to changelog file %s ", changelogPath)

	content, err := c.ReadFile(changelogPath)
	if err != nil {
		return fmt.Errorf("read changelog file: %w", err)
	}

	tocEndIndex := bytes.Index(content, []byte(TocEnd))
	if tocEndIndex < 0 {
		return fmt.Errorf(
			"find table of contents end marker `%s` in %q",
			TocEnd, changelogPath,
		)
	}

	mergedMarkdown := fmt.Sprintf(
		"%s\n%s", markdown, string(content[(len(TocEnd)+tocEndIndex):]),
	)

	mergedTOC, err := c.GenerateTOC(mergedMarkdown)
	if err != nil {
		return fmt.Errorf("generate table of contents: %w", err)
	}

	if err := writeFile(mergedTOC, mergedMarkdown); err != nil {
		return fmt.Errorf("write merged markdown: %w", err)
	}

	return nil
}

func (c *Changelog) htmlChangelogFilename(tag semver.Version) string {
	if c.options.HTMLFile != "" {
		return c.options.HTMLFile
	}

	return changelogFilename(tag, "html")
}

func (c *Changelog) jsonChangelogFilename(tag semver.Version) string {
	if c.options.JSONFile != "" {
		return c.options.JSONFile
	}

	return changelogFilename(tag, "json")
}

func markdownChangelogReadme() string {
	return filepath.Join(RepoChangelogDir, "README.md")
}

func markdownChangelogFilename(tag semver.Version) string {
	return filepath.Join(RepoChangelogDir, changelogFilename(tag, "md"))
}

func changelogFilename(tag semver.Version, ext string) string {
	return fmt.Sprintf("CHANGELOG-%d.%d.%s", tag.Major, tag.Minor, ext)
}

func addTocMarkers(toc string) string {
	return fmt.Sprintf("%s\n\n%s\n%s\n", tocStart, toc, TocEnd)
}

func (c *Changelog) writeHTML(tag semver.Version, markdown string) error {
	content := &bytes.Buffer{}
	if err := c.MarkdownToHTML(markdown, content); err != nil {
		return fmt.Errorf("render HTML from markdown: %w", err)
	}

	t, err := c.ParseHTMLTemplate(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parse HTML template: %w", err)
	}

	output := bytes.Buffer{}
	if err := c.TemplateExecute(t, &output, struct {
		Title, Content string
	}{util.SemverToTagString(tag), content.String()}); err != nil {
		return fmt.Errorf("execute HTML template: %w", err)
	}

	absOutputPath, err := c.Abs(c.htmlChangelogFilename(tag))
	if err != nil {
		return fmt.Errorf("get absolute file path: %w", err)
	}

	logrus.Infof("Writing HTML file to %s", absOutputPath)

	if err := c.WriteFile(absOutputPath, output.Bytes(), os.FileMode(0o644)); err != nil {
		return fmt.Errorf("write template: %w", err)
	}

	return nil
}

func (c *Changelog) writeJSON(tag semver.Version, jsonStr string) error {
	absOutputPath, err := c.Abs(c.jsonChangelogFilename(tag))
	if err != nil {
		return fmt.Errorf("get absolute file path: %w", err)
	}

	logrus.Infof("Writing JSON file to %s", absOutputPath)

	if err := c.WriteFile(absOutputPath, []byte(jsonStr), os.FileMode(0o644)); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	return nil
}

func (c *Changelog) lookupRemoteReleaseNotes(
	branch string,
) (markdownStr, jsonStr string, err error) {
	logrus.Info("Assuming new minor release, fetching remote release notes")

	remoteBase := fmt.Sprintf(
		"https://raw.githubusercontent.com/kubernetes/sig-release/%s/"+
			"releases/%s/release-notes/", git.DefaultBranch, branch,
	)

	// Retrieve the markdown version
	remoteMarkdown := remoteBase + "release-notes-draft.md"

	markdownStr, err = c.GetURLResponse(remoteMarkdown)
	if err != nil {
		return "", "", fmt.Errorf("fetch release notes markdown from remote: %s: %w", remoteMarkdown, err)
	}

	logrus.Infof("Found remote release notes markdown on: %s", remoteMarkdown)

	// Retrieve the JSON version
	remoteJSON := remoteBase + "release-notes-draft.json"

	jsonStr, err = c.GetURLResponse(remoteJSON)
	if err != nil {
		logrus.Warnf(
			"Unable to fetch release notes JSON from remote %s: %v",
			remoteJSON, err,
		)
		// Fallback in case we're not able to retrieve a JSON draft.
		jsonStr = "{}"
	}

	logrus.Infof("Found remote release notes JSON on: %s", remoteJSON)

	return markdownStr, jsonStr, nil
}

func (c *Changelog) commitChanges(
	repo *git.Repo, branch string, tag semver.Version,
) error {
	// main branch modifications
	releaseChangelog := markdownChangelogFilename(tag)
	changelogReadme := markdownChangelogReadme()

	changelogFiles := []string{
		releaseChangelog,
		changelogReadme,
	}

	for _, filename := range changelogFiles {
		logrus.Infof("Adding %s to repository", filename)

		if err := c.Add(repo, filename); err != nil {
			return fmt.Errorf("add file %s to repository: %w", filename, err)
		}
	}

	logrus.Info("Committing changes to main branch in repository")

	if err := c.Commit(repo, fmt.Sprintf(
		"CHANGELOG: Update directory for %s release", util.SemverToTagString(tag),
	)); err != nil {
		return fmt.Errorf("committing changes into repository: %w", err)
	}

	if branch != git.DefaultBranch {
		logrus.Infof("Checking out %s branch", branch)
		// Release branch modifications
		if err := c.Checkout(repo, branch); err != nil {
			return fmt.Errorf("checking out release branch %s: %w", branch, err)
		}

		// Remove all other changelog files if we’re on the the first official release
		if tag.Patch == 0 && len(tag.Pre) == 0 {
			pattern := filepath.Join(RepoChangelogDir, "CHANGELOG-*.md")
			logrus.Infof("Removing unnecessary %s files", pattern)

			if err := c.Rm(repo, true, pattern); err != nil {
				return fmt.Errorf("removing %s files: %w", pattern, err)
			}
		}

		logrus.Info("Checking out changelog from main branch")

		if err := c.Checkout(
			repo, git.DefaultBranch, releaseChangelog,
		); err != nil {
			return fmt.Errorf("check out main branch changelog: %w", err)
		}

		logrus.Info("Committing changes to release branch in repository")

		if err := c.Commit(repo, fmt.Sprintf(
			"Update %s for %s", releaseChangelog, util.SemverToTagString(tag),
		)); err != nil {
			return fmt.Errorf("committing changes into repository: %w", err)
		}
	}

	return nil
}

func (c *Changelog) adaptChangelogReadmeFile(
	repo *git.Repo, tag semver.Version,
) error {
	targetFile := filepath.Join(repo.Dir(), RepoChangelogDir, "README.md")

	readme, err := c.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("read changelog README.md: %w", err)
	}

	cf := filepath.Base(markdownChangelogFilename(tag))

	const listPrefix = "- "

	changelogEntry := fmt.Sprintf("%s[%s](./%s)", listPrefix, cf, cf)
	scanner := bufio.NewScanner(bytes.NewReader(readme))

	res := []string{}
	inserted := false

	for scanner.Scan() {
		line := scanner.Text()
		if !inserted && strings.HasPrefix(line, listPrefix) {
			res = append(res, changelogEntry)
			inserted = true
		}

		res = append(res, line)
	}

	if err := c.WriteFile(
		targetFile, []byte(strings.Join(res, nl)+nl), os.FileMode(0o644)); err != nil {
		return fmt.Errorf("write changelog README.md: %w", err)
	}

	return nil
}

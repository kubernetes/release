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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/http"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/document"
	"k8s.io/release/pkg/notes/options"
	"k8s.io/release/pkg/util"
	"sigs.k8s.io/mdtoc/pkg/mdtoc"
)

const (
	// The default CHANGELOG directory inside the k/k repository.
	RepoChangelogDir = "CHANGELOG"

	nl                   = "\n"
	tocStart             = "<!-- BEGIN MUNGE: GENERATED_TOC -->"
	tocEnd               = "<!-- END MUNGE: GENERATED_TOC -->"
	releaseNotesTemplate = `
{{- $CurrentRevision := .CurrentRevision -}}
{{- $PreviousRevision := .PreviousRevision -}}
# {{$CurrentRevision}}

{{if .Downloads}}
## Downloads for {{$CurrentRevision}}

{{- with .Downloads.Source }}

### Source Code

filename | sha512 hash
-------- | -----------
{{range .}}[{{.Name}}]({{.URL}}) | {{.Checksum}}{{println}}{{end}}
{{end}}

{{- with .Downloads.Client -}}
### Client binaries

filename | sha512 hash
-------- | -----------
{{range .}}[{{.Name}}]({{.URL}}) | {{.Checksum}}{{println}}{{end}}
{{end}}

{{- with .Downloads.Server -}}
### Server binaries

filename | sha512 hash
-------- | -----------
{{range .}}[{{.Name}}]({{.URL}}) | {{.Checksum}}{{println}}{{end}}
{{end}}

{{- with .Downloads.Node -}}
### Node binaries

filename | sha512 hash
-------- | -----------
{{range .}}[{{.Name}}]({{.URL}}) | {{.Checksum}}{{println}}{{end}}
{{end -}}
{{- end -}}
## Changelog since {{$PreviousRevision}}

{{with .NotesWithActionRequired -}}
## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)

{{range .}} {{println "-" .}} {{end}}
{{end}}

{{- if .Notes -}}
## Changes by Kind
{{ range .Notes}}
### {{.Kind | prettyKind}}

{{range $note := .NoteEntries }}{{println "-" $note}}{{end}}
{{- end -}}
{{- end -}}
`

	htmlTemplate = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width" />
    <title>{{ .Title }}</title>
    <style type="text/css">
      table,
      th,
      tr,
      td {
        border: 1px solid gray;
        border-collapse: collapse;
        padding: 5px;
      }
    </style>
  </head>
  <body>
    {{ .Content }}
  </body>
</html>`
)

// Changelog can be used to generate the changelog for a release.
type Changelog struct {
	options      *Options
	dependencies *notes.Dependencies
}

// New creates a new Changelog instance.
func New(opts *Options) *Changelog {
	return &Changelog{
		options:      opts,
		dependencies: notes.NewDependencies(),
	}
}

// SetMoDiff can be used to set the dependency module differ.
func (c *Changelog) SetMoDiff(moDiff notes.MoDiff) {
	c.dependencies.SetMoDiff(moDiff)
}

// Options are the main settings for generating the changelog.
type Options struct {
	RepoPath     string
	Tag          string
	Branch       string
	Bucket       string
	Tars         string
	HTMLFile     string
	RecordDir    string
	ReplayDir    string
	Dependencies bool
}

// Run starts the changelog generation.
func (c *Changelog) Run() error {
	tag, err := util.TagStringToSemver(c.options.Tag)
	if err != nil {
		return errors.Wrapf(err, "parse tag %s", c.options.Tag)
	}

	// Automatically set the branch to a release branch if not provided
	branch := c.options.Branch
	if branch == "" {
		branch = fmt.Sprintf("release-%d.%d", tag.Major, tag.Minor)
	}
	logrus.Infof("Using release branch %s", branch)

	logrus.Infof("Using local repository path %s", c.options.RepoPath)
	repo, err := git.OpenRepo(c.options.RepoPath)
	if err != nil {
		return errors.Wrapf(err,
			"open expected k/k repository %q", c.options.RepoPath,
		)
	}
	if currentBranch, err := repo.CurrentBranch(); err == nil {
		logrus.Infof("We're currently on branch: %s", currentBranch)
	}

	remoteBranch := git.Remotify(branch)
	head, err := repo.RevParse(remoteBranch)
	if err != nil {
		return errors.Wrap(err, "get latest branch commit")
	}
	logrus.Infof("Found latest %s commit %s", remoteBranch, head)

	var markdown, startRev, endRev string
	if tag.Patch == 0 {
		if len(tag.Pre) == 0 {
			// Still create the downloads table
			downloadsTable := &bytes.Buffer{}
			startTag := util.SemverToTagString(semver.Version{
				Major: tag.Major, Minor: tag.Minor - 1, Patch: 0,
			})

			startRev = startTag
			endRev = head

			if err := document.CreateDownloadsTable(
				downloadsTable, c.options.Bucket, c.options.Tars,
				startRev, c.options.Tag,
			); err != nil {
				return errors.Wrapf(err, "create downloads table")
			}

			// New final minor versions should have remote release notes
			markdown, err = lookupRemoteReleaseNotes(branch)
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

			markdown, err = c.generateReleaseNotes(branch, startRev, endRev)
		} else {
			// New minor alpha, beta and rc releases get generated notes
			latestTags, tErr := github.New().LatestGitHubTagsPerBranch()
			if tErr != nil {
				return errors.Wrap(tErr, "get latest GitHub tags")
			}

			if startTag, ok := latestTags[branch]; ok {
				logrus.Infof("Found start tag %s", startTag)

				// The end tag does not yet exist which means that we stick to
				// the current HEAD as end revision.
				startRev = startTag
				endRev = head

				markdown, err = c.generateReleaseNotes(branch, startRev, endRev)
			} else {
				return errors.Errorf(
					"no latest tag available for branch %s", branch,
				)
			}
		}
	} else {
		// A patch version, let’s just use the previous patch
		startTag := util.SemverToTagString(semver.Version{
			Major: tag.Major, Minor: tag.Minor, Patch: tag.Patch - 1,
		})

		startRev = startTag
		endRev = head

		markdown, err = c.generateReleaseNotes(branch, startTag, endRev)
	}
	if err != nil {
		return err
	}

	logrus.Info("Generating TOC")
	toc, err := mdtoc.GenerateTOC([]byte(markdown))
	if err != nil {
		return err
	}

	if c.options.Dependencies {
		logrus.Info("Generating dependency changes")
		deps, err := c.dependencies.Changes(startRev, endRev)
		if err != nil {
			return err
		}
		markdown += strings.Repeat(nl, 2) + deps
	}

	// Restore the currently checked out branch
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		return err
	}
	if currentBranch != "" {
		defer func() {
			if err := repo.Checkout(currentBranch); err != nil {
				logrus.Errorf("Restore branch %s: %v", currentBranch, err)
			}
		}()
	}

	logrus.Infof("Checking out %s branch", git.DefaultBranch)
	if err := repo.Checkout(git.DefaultBranch); err != nil {
		return errors.Wrap(err, "checking out main branch")
	}

	logrus.Info("Writing markdown")
	if err := writeMarkdown(repo, toc, markdown, tag); err != nil {
		return err
	}

	logrus.Info("Writing HTML")
	if err := c.writeHTML(tag, markdown); err != nil {
		return err
	}

	logrus.Info("Committing changes")
	return commitChanges(repo, branch, tag)
}

func (c *Changelog) generateReleaseNotes(branch, startRev, endRev string) (string, error) {
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

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return "", errors.Wrap(err, "validating notes options")
	}

	doc, err := document.GatherReleaseNotesDocument(
		notesOptions, startRev, c.options.Tag,
	)
	if err != nil {
		return "", err
	}

	markdown, err := doc.RenderMarkdownTemplate(
		c.options.Bucket, c.options.Tars,
		options.GoTemplateInline+releaseNotesTemplate,
	)
	if err != nil {
		return "", errors.Wrapf(
			err, "rendering release notes to markdown",
		)
	}

	return markdown, nil
}

func writeMarkdown(repo *git.Repo, toc, markdown string, tag semver.Version) error {
	changelogPath := filepath.Join(repo.Dir(), markdownChangelogFilename(tag))
	writeFile := func(t, m string) error {
		return ioutil.WriteFile(
			changelogPath, []byte(strings.Join(
				[]string{addTocMarkers(t), strings.TrimSpace(m)}, "\n",
			)), os.FileMode(0644),
		)
	}

	// No changelog exists, simply write the content to a new one
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		logrus.Infof("Changelog %q does not exist, creating it", changelogPath)
		if err := adaptChangelogReadmeFile(repo, tag); err != nil {
			return errors.Wrap(err, "changelog readme adaption failed")
		}
		return writeFile(toc, markdown)
	}

	// Changelog seems to exist, prepend the notes and re-generate the TOC
	logrus.Infof("Adding new content to changelog file %s ", changelogPath)
	content, err := ioutil.ReadFile(changelogPath)
	if err != nil {
		return err
	}

	tocEndIndex := bytes.Index(content, []byte(tocEnd))
	if tocEndIndex < 0 {
		return errors.Errorf(
			"find table of contents end marker `%s` in %q",
			tocEnd, changelogPath,
		)
	}

	mergedMarkdown := fmt.Sprintf(
		"%s\n%s", markdown, string(content[(len(tocEnd)+tocEndIndex):]),
	)
	mergedTOC, err := mdtoc.GenerateTOC([]byte(mergedMarkdown))
	if err != nil {
		return err
	}
	return writeFile(mergedTOC, mergedMarkdown)
}

func (c *Changelog) htmlChangelogFilename(tag semver.Version) string {
	if c.options.HTMLFile != "" {
		return c.options.HTMLFile
	}
	return changelogFilename(tag, "html")
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
	return fmt.Sprintf("%s\n\n%s\n%s\n", tocStart, toc, tocEnd)
}

func (c *Changelog) writeHTML(tag semver.Version, markdown string) error {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	content := &bytes.Buffer{}
	if err := md.Convert([]byte(markdown), content); err != nil {
		return errors.Wrap(err, "render HTML from markdown")
	}

	t, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	output := bytes.Buffer{}
	if err := t.Execute(&output, struct {
		Title, Content string
	}{util.SemverToTagString(tag), content.String()}); err != nil {
		return err
	}

	absOutputPath, err := filepath.Abs(c.htmlChangelogFilename(tag))
	if err != nil {
		return err
	}
	logrus.Infof("Writing HTML file to %s", absOutputPath)
	return ioutil.WriteFile(absOutputPath, output.Bytes(), os.FileMode(0644))
}

func lookupRemoteReleaseNotes(branch string) (string, error) {
	logrus.Info("Assuming new minor release")

	remote := fmt.Sprintf(
		"https://raw.githubusercontent.com/kubernetes/sig-release/%s/"+
			"releases/%s/release-notes-draft.md", git.DefaultBranch, branch,
	)
	response, err := http.GetURLResponse(remote, false)
	if err != nil {
		return "", errors.Wrapf(err,
			"fetching release notes from remote: %s", remote,
		)
	}
	logrus.Infof("Found remote release notes on: %s", remote)
	return response, nil
}

func commitChanges(repo *git.Repo, branch string, tag semver.Version) error {
	// main branch modifications
	releaseChangelog := markdownChangelogFilename(tag)
	changelogReadme := markdownChangelogReadme()

	changelogFiles := []string{
		releaseChangelog,
		changelogReadme,
	}

	for _, filename := range changelogFiles {
		logrus.Infof("Adding %s to repository", filename)
		if err := repo.Add(filename); err != nil {
			return errors.Wrapf(err, "trying to add file %s to repository", filename)
		}
	}

	logrus.Info("Committing changes to main branch in repository")
	if err := repo.Commit(fmt.Sprintf(
		"CHANGELOG: Update directory for %s release", util.SemverToTagString(tag),
	)); err != nil {
		return errors.Wrap(err, "committing changes into repository")
	}

	if branch != git.DefaultBranch {
		logrus.Infof("Checking out %s branch", branch)
		// Release branch modifications
		if err := repo.Checkout(branch); err != nil {
			return errors.Wrapf(err, "checking out release branch %s", branch)
		}

		// Remove all other changelog files if we’re on the the first official release
		if tag.Patch == 0 && len(tag.Pre) == 0 {
			pattern := filepath.Join(RepoChangelogDir, "CHANGELOG-*.md")
			logrus.Infof("Removing unnecessary %s files", pattern)
			if err := repo.Rm(true, pattern); err != nil {
				return errors.Wrapf(err, "removing %s files", pattern)
			}
		}

		logrus.Info("Checking out changelog from main branch")
		if err := repo.Checkout(git.DefaultBranch, releaseChangelog); err != nil {
			return errors.Wrap(err, "check out main branch changelog")
		}

		logrus.Info("Committing changes to release branch in repository")
		if err := repo.Commit(fmt.Sprintf(
			"Update %s for %s", releaseChangelog, util.SemverToTagString(tag),
		)); err != nil {
			return errors.Wrap(err, "committing changes into repository")
		}
	}

	return nil
}

func adaptChangelogReadmeFile(repo *git.Repo, tag semver.Version) error {
	targetFile := filepath.Join(repo.Dir(), RepoChangelogDir, "README.md")
	readme, err := ioutil.ReadFile(targetFile)
	if err != nil {
		return errors.Wrap(err, "read changelog README.md")
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

	if err := ioutil.WriteFile(
		targetFile, []byte(strings.Join(res, nl)+nl), os.FileMode(0644)); err != nil {
		return errors.Wrap(err, "write changelog README.md")
	}
	return nil
}

/*
Copyright 2019 The Kubernetes Authors.

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
	"github.com/spf13/cobra"
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

// changelogCmd represents the subcommand for `krel changelog`
var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Automate the lifecycle of CHANGELOG-x.y.{md,html} files in a k/k repository",
	Long: fmt.Sprintf(`krel changelog

To let this tool work, please point '--repo' to a local copy of the target k/k
repository. This local checkout will be modified during the run of 'krel
changelog' and should contain all changes from the remote location. Beside this,
a valid %s=<TOKEN> environment variable has to be exported to let the generation
of the release notes work.

The 'changelog' subcommand of 'krel' does the following things by utilizing
the golang based 'release-notes' tool:

1. Generate the release notes for either a patch or a new minor release. Minor
   releases can be alpha, beta or rc’s, too.
   a) Create a new CHANGELOG-x.y.md file if not existing.
   b) Correctly prepend the generated notes to the existing CHANGELOG-x.y.md
      file if already existing. This also includes the modification of the
	  table of contents.

2. Convert the markdown release notes into a HTML equivalent on purpose of
   sending it by mail to the announce list. The HTML file will be dropped into
   the current working directly as 'CHANGELOG-x.y.html'. Sending the
   announcement is done by another subcommand of 'krel', not 'changelog'.

3. Commit the modified CHANGELOG-x.y.md into the main branch as well as the
   corresponding release-branch of kubernetes/kubernetes. The release branch
   will be pruned from all other CHANGELOG-*.md files which do not belong to
   this release branch.
`, github.TokenEnvKey),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return newChangelog().run(changelogOpts, rootOpts)
	},
}

type changelogOptions struct {
	tag          string
	branch       string
	bucket       string
	tars         string
	htmlFile     string
	recordDir    string
	replayDir    string
	dependencies bool
}

var changelogOpts = &changelogOptions{}

const (
	tocStart             = "<!-- BEGIN MUNGE: GENERATED_TOC -->"
	tocEnd               = "<!-- END MUNGE: GENERATED_TOC -->"
	repoChangelogDir     = "CHANGELOG"
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

type Changelog struct {
	dependencies *notes.Dependencies
}

func newChangelog() *Changelog {
	return &Changelog{
		dependencies: notes.NewDependencies(),
	}
}

func init() {
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.bucket, "bucket", "kubernetes-release", "Specify gs bucket to point to in generated notes")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.tag, "tag", "", "The version tag of the release, for example v1.17.0-rc.1")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.branch, "branch", "", "The branch to be used. Will be automatically inherited by the tag if not set.")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.tars, "tars", ".", "Directory of tars to SHA512 sum for display")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.htmlFile, "html-file", "", "The target html file to be written. If empty, then it will be CHANGELOG-x.y.html in the current path.")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.recordDir, "record", "", "Record the API into a directory")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.replayDir, "replay", "", "Replay a previously recorded API from a directory")
	changelogCmd.PersistentFlags().BoolVar(&changelogOpts.dependencies, "dependencies", true, "Add dependency report")

	if err := changelogCmd.MarkPersistentFlagRequired("tag"); err != nil {
		logrus.Fatalf("Unable to %v", err)
	}

	rootCmd.AddCommand(changelogCmd)
}

func (c *Changelog) run(opts *changelogOptions, rootOpts *rootOptions) error {
	tag, err := util.TagStringToSemver(opts.tag)
	if err != nil {
		return errors.Wrapf(err, "parse tag %s", opts.tag)
	}

	// Automatically set the branch to a release branch if not provided
	branch := opts.branch
	if branch == "" {
		branch = fmt.Sprintf("release-%d.%d", tag.Major, tag.Minor)
	}
	logrus.Infof("Using release branch %s", branch)

	logrus.Infof("Using local repository path %s", rootOpts.repoPath)
	repo, err := git.OpenRepo(rootOpts.repoPath)
	if err != nil {
		return errors.Wrapf(err,
			"open expected k/k repository %q", rootOpts.repoPath,
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
				downloadsTable, opts.bucket, opts.tars, startRev, opts.tag,
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

			markdown, err = generateReleaseNotes(opts, branch, startRev, endRev)
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

				markdown, err = generateReleaseNotes(opts, branch, startRev, endRev)
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

		markdown, err = generateReleaseNotes(opts, branch, startTag, endRev)
	}
	if err != nil {
		return err
	}

	logrus.Info("Generating TOC")
	toc, err := mdtoc.GenerateTOC([]byte(markdown))
	if err != nil {
		return err
	}

	if opts.dependencies {
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
	defer func() {
		if err := repo.Checkout(currentBranch); err != nil {
			logrus.Errorf("Restore branch %s: %v", currentBranch, err)
		}
	}()

	logrus.Infof("Checking out %s branch", git.DefaultBranch)
	if err := repo.Checkout(git.DefaultBranch); err != nil {
		return errors.Wrap(err, "checking out main branch")
	}

	logrus.Info("Writing markdown")
	if err := writeMarkdown(repo, toc, markdown, tag); err != nil {
		return err
	}

	logrus.Info("Writing HTML")
	if err := writeHTML(opts, tag, markdown); err != nil {
		return err
	}

	logrus.Info("Committing changes")
	return commitChanges(repo, branch, tag)
}

func generateReleaseNotes(opts *changelogOptions, branch, startRev, endRev string) (string, error) {
	logrus.Info("Generating release notes")

	notesOptions := options.New()
	notesOptions.Branch = branch
	notesOptions.StartRev = startRev
	notesOptions.EndSHA = endRev
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.ReleaseBucket = opts.bucket
	notesOptions.ReleaseTars = opts.tars
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOptions.RecordDir = opts.recordDir
	notesOptions.ReplayDir = opts.replayDir
	notesOptions.Pull = false

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return "", errors.Wrap(err, "validating notes options")
	}

	doc, err := document.GatherReleaseNotesDocument(
		notesOptions, startRev, opts.tag,
	)
	if err != nil {
		return "", err
	}

	markdown, err := doc.RenderMarkdownTemplate(
		opts.bucket, opts.tars,
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

func htmlChangelogFilename(opts *changelogOptions, tag semver.Version) string {
	if opts.htmlFile != "" {
		return opts.htmlFile
	}
	return changelogFilename(tag, "html")
}

func markdownChangelogReadme() string {
	return filepath.Join(repoChangelogDir, "README.md")
}

func markdownChangelogFilename(tag semver.Version) string {
	return filepath.Join(repoChangelogDir, changelogFilename(tag, "md"))
}

func changelogFilename(tag semver.Version, ext string) string {
	return fmt.Sprintf("CHANGELOG-%d.%d.%s", tag.Major, tag.Minor, ext)
}

func addTocMarkers(toc string) string {
	return fmt.Sprintf("%s\n\n%s\n%s\n", tocStart, toc, tocEnd)
}

func writeHTML(opts *changelogOptions, tag semver.Version, markdown string) error {
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

	absOutputPath, err := filepath.Abs(htmlChangelogFilename(opts, tag))
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
			pattern := filepath.Join(repoChangelogDir, "CHANGELOG-*.md")
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
	targetFile := filepath.Join(repo.Dir(), repoChangelogDir, "README.md")
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

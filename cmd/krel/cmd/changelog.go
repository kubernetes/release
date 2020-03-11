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
	"context"
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

3. Commit the modified CHANGELOG-x.y.md into the master branch as well as the
   corresponding release-branch of kubernetes/kubernetes. The release branch
   will be pruned from all other CHANGELOG-*.md files which do not belong to
   this release branch.
`, options.GitHubToken),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangelog(changelogOpts, rootOpts)
	},
}

type changelogOptions struct {
	tag       string
	branch    string
	bucket    string
	tars      string
	htmlFile  string
	recordDir string
	replayDir string
}

var changelogOpts = &changelogOptions{}

const (
	tocStart         = "<!-- BEGIN MUNGE: GENERATED_TOC -->"
	tocEnd           = "<!-- END MUNGE: GENERATED_TOC -->"
	repoChangelogDir = "CHANGELOG"
)

func init() {
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.bucket, "bucket", "kubernetes-release", "Specify gs bucket to point to in generated notes")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.tag, "tag", "", "The version tag of the release, for example v1.17.0-rc.1")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.branch, "branch", "", "The branch to be used. Will be automatically inherited by the tag if not set.")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.tars, "tars", ".", "Directory of tars to SHA512 sum for display")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.htmlFile, "html-file", "", "The target html file to be written. If empty, then it will be CHANGELOG-x.y.html in the current path.")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.recordDir, "record", "", "Record the API into a directory")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.replayDir, "replay", "", "Replay a previously recorded API from a directory")

	if err := changelogCmd.MarkPersistentFlagRequired("tag"); err != nil {
		logrus.Fatal(err)
	}

	rootCmd.AddCommand(changelogCmd)
}

func runChangelog(opts *changelogOptions, rootOpts *rootOptions) error {
	tag, err := util.TagStringToSemver(opts.tag)
	if err != nil {
		return err
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
			"unable to open expected k/k repository %q", rootOpts.repoPath,
		)
	}
	if currentBranch, err := repo.CurrentBranch(); err == nil {
		logrus.Infof("We're currently on branch: %s", currentBranch)
	}

	remoteBranch := git.Remotify(branch)
	head, err := repo.RevParse(remoteBranch)
	if err != nil {
		return errors.Wrapf(err, "unable to get latest branch commit")
	}
	logrus.Infof("Found latest %s commit %s", remoteBranch, head)

	var markdown string
	if tag.Patch == 0 {
		if len(tag.Pre) == 0 {
			// Still create the downloads table
			downloadsTable := &bytes.Buffer{}
			previousTag := util.SemverToTagString(semver.Version{
				Major: tag.Major, Minor: tag.Minor - 1, Patch: 0,
			})
			if err := document.CreateDownloadsTable(
				downloadsTable, opts.bucket, opts.tars, previousTag, opts.tag,
			); err != nil {
				return errors.Wrapf(err, "unable to create downloads table")
			}

			// New final minor versions should have remote release notes
			markdown, err = lookupRemoteReleaseNotes(branch)
			markdown = downloadsTable.String() + markdown
		} else {
			// New minor alphas, betas and rc get generated notes
			latestTags, tErr := github.New().LatestGitHubTagsPerBranch()
			if tErr != nil {
				return errors.Wrap(tErr, "unable to get latest GitHub tags")
			}

			if startTag, ok := latestTags[branch]; ok {
				logrus.Infof("Found start tag %s", startTag)
				markdown, err = generateReleaseNotes(opts, branch, startTag, head)
			} else {
				return errors.Errorf(
					"no latest tag available for branch %s", branch,
				)
			}
		}
	} else {
		// A patch version, let’s just use the previous patch
		start := util.SemverToTagString(semver.Version{
			Major: tag.Major, Minor: tag.Minor, Patch: tag.Patch - 1,
		})

		markdown, err = generateReleaseNotes(opts, branch, start, head)
	}
	if err != nil {
		return err
	}

	logrus.Info("Generating TOC")
	toc, err := notes.GenerateTOC(markdown)
	if err != nil {
		return err
	}

	// Restore the currently checked out branch
	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		return err
	}
	defer func() {
		if err := repo.Checkout(currentBranch); err != nil {
			logrus.Errorf("unable to restore branch %s: %v", currentBranch, err)
		}
	}()

	if err := repo.Checkout(git.Master); err != nil {
		return errors.Wrap(err, "checking out master branch")
	}

	if err := writeMarkdown(repo, toc, markdown, tag); err != nil {
		return err
	}

	if err := writeHTML(opts, tag, markdown); err != nil {
		return err
	}

	return commitChanges(repo, branch, tag)
}

func generateReleaseNotes(opts *changelogOptions, branch, startRev, endRev string) (string, error) {
	logrus.Info("Generating release notes")

	notesOptions := options.New()
	notesOptions.Branch = branch
	notesOptions.StartRev = startRev
	notesOptions.EndSHA = endRev
	notesOptions.EndRev = opts.tag
	notesOptions.GithubOrg = git.DefaultGithubOrg
	notesOptions.GithubRepo = git.DefaultGithubRepo
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.ReleaseBucket = opts.bucket
	notesOptions.ReleaseTars = opts.tars
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOptions.RecordDir = opts.recordDir
	notesOptions.ReplayDir = opts.replayDir
	notesOptions.Pull = false

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return "", err
	}

	gatherer := notes.NewGatherer(context.Background(), notesOptions)
	releaseNotes, history, err := gatherer.ListReleaseNotes()
	if err != nil {
		return "", errors.Wrapf(err, "listing release notes")
	}

	// Create the markdown
	doc, err := document.CreateDocument(releaseNotes, history)
	if err != nil {
		return "", errors.Wrapf(err, "creating release note document")
	}

	markdown, err := doc.RenderMarkdown(
		opts.bucket, opts.tars, notesOptions.StartRev, notesOptions.EndRev,
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
			"unable to find table of contents end marker `%s` in %q",
			tocEnd, changelogPath,
		)
	}

	mergedMarkdown := fmt.Sprintf(
		"%s\n%s", markdown, string(content[(len(tocEnd)+tocEndIndex):]),
	)
	mergedTOC, err := notes.GenerateTOC(mergedMarkdown)
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

const htmlTemplate = `<!DOCTYPE html>
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

func writeHTML(opts *changelogOptions, tag semver.Version, markdown string) error {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	content := &bytes.Buffer{}
	if err := md.Convert([]byte(markdown), content); err != nil {
		return errors.Wrap(err, "unable to render HTML from markdown")
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
	logrus.Infof("Writing single HTML to %s", absOutputPath)
	return ioutil.WriteFile(absOutputPath, output.Bytes(), os.FileMode(0644))
}

func lookupRemoteReleaseNotes(branch string) (string, error) {
	logrus.Info("Assuming new minor release")

	remote := fmt.Sprintf(
		"https://raw.githubusercontent.com/kubernetes/sig-release/master/"+
			"releases/%s/release-notes-draft.md", branch,
	)
	response, err := http.GetURLResponse(remote, false)
	if err != nil {
		return "", errors.Wrapf(err,
			"fetching release notes from remote: %s", remote,
		)
	}
	logrus.Info("Found release notes")
	return response, nil
}

func commitChanges(repo *git.Repo, branch string, tag semver.Version) error {
	// Master branch modifications
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

	logrus.Info("Committing changes to master branch in repository")
	if err := repo.Commit(fmt.Sprintf(
		"CHANGELOG: Update directory for %s release", util.SemverToTagString(tag),
	)); err != nil {
		return errors.Wrap(err, "committing changes into repository")
	}

	if branch != git.Master {
		// Release branch modifications
		if err := repo.Checkout(branch); err != nil {
			return errors.Wrapf(err, "checking out release branch %s", branch)
		}

		// Remove all other changelog files
		if err := repo.Rm(true, repoChangelogDir+"/CHANGELOG-*.md"); err != nil {
			return errors.Wrap(err, "unable to remove CHANGELOG-*.md files")
		}

		logrus.Info("Checking out changelog from master branch")
		if err := repo.Checkout(git.Master, releaseChangelog); err != nil {
			return errors.Wrap(err, "unable to check out master branch changelog")
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
		return errors.Wrap(err, "unable to read changelog README.md")
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

	const nl = "\n"
	if err := ioutil.WriteFile(
		targetFile, []byte(strings.Join(res, nl)+nl), os.FileMode(0644)); err != nil {
		return errors.Wrap(err, "unable to write changelog README.md")
	}
	return nil
}

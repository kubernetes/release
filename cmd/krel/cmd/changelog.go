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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/russross/blackfriday.v2"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/options"
	"k8s.io/release/pkg/util"
)

// changelogCmd represents the subcommand for `krel changelog`
var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "changelog maintains the lifecycle of CHANGELOG-x.y.{md,html} files",
	Long: `krel changelog

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
   announcement is done by another subcommand of 'krel', not "changelog'.

3. Commit the modified CHANGELOG-x.y.md into the master branch as well as the
   corresponding release-branch of kubernetes/kubernetes.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE:       initLogging,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangelog()
	},
}

type changelogOptions struct {
	tag    string
	bucket string
	tars   string
	token  string
}

var changelogOpts = &changelogOptions{}

const (
	tocStart = "<!-- BEGIN MUNGE: GENERATED_TOC -->"
	tocEnd   = "<!-- END MUNGE: GENERATED_TOC -->"
)

func init() {
	cobra.OnInitialize(initConfig)

	const (
		tagFlag   = "tag"
		tarsFlag  = "tars"
		tokenFlag = "token"
	)
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.bucket, "bucket", "kubernetes-release", "Specify gs bucket to point to in generated notes")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.tag, tagFlag, "", "The version tag of the release, for example v1.17.0-rc.1")
	changelogCmd.PersistentFlags().StringVar(&changelogOpts.tars, tarsFlag, "", "Directory of tars to sha512 sum for display")

	if err := changelogCmd.MarkPersistentFlagRequired(tagFlag); err != nil {
		logrus.Fatal(err)
	}
	if err := changelogCmd.MarkPersistentFlagRequired(tarsFlag); err != nil {
		logrus.Fatal(err)
	}

	rootCmd.AddCommand(changelogCmd)
}

func runChangelog() (err error) {
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		return errors.New("environment variable `GITHUB_TOKEN` is not set but needed for release notes generation")
	}
	changelogOpts.token = token

	tag, err := semver.Make(util.TrimTagPrefix(changelogOpts.tag))
	if err != nil {
		return err
	}
	branch := fmt.Sprintf("release-%d.%d", tag.Major, tag.Minor)
	logrus.Infof("Using release branch %s", branch)

	logrus.Infof("Using local repository path %s", rootOpts.repoPath)
	repo, err := git.CloneOrOpenDefaultGitHubRepoSSH(rootOpts.repoPath, git.DefaultGithubOrg)
	if err != nil {
		return err
	}

	var markdown string
	if tag.Patch == 0 {
		if len(tag.Pre) == 0 {
			// New final minor versions should have remote release notes
			markdown, err = lookupRemoteReleaseNotes(branch)
		} else {
			// New minor alphas, betas and rc get generated notes
			start, e := repo.PreviousTag(changelogOpts.tag, branch)
			if e != nil {
				return e
			}
			logrus.Infof("Found previous tag %s", start)
			markdown, err = generateReleaseNotes(branch, start, changelogOpts.tag)
		}
	} else {
		// A patch version, let’s just use the previous patch
		start := util.AddTagPrefix(semver.Version{
			Major: tag.Major, Minor: tag.Minor, Patch: tag.Patch - 1,
		}.String())

		markdown, err = generateReleaseNotes(branch, start, changelogOpts.tag)
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
		if err := repo.CheckoutBranch(currentBranch); err != nil {
			logrus.Errorf("unable to restore branch %s: %v", currentBranch, err)
		}
	}()

	if err := repo.CheckoutBranch(git.Master); err != nil {
		return errors.Wrap(err, "checking out master branch")
	}

	if err := writeMarkdown(repo, toc, markdown, tag); err != nil {
		return err
	}

	if err := writeHTML(tag, markdown); err != nil {
		return err
	}

	return commitChanges(repo, branch, tag)
}

func generateReleaseNotes(branch, startRev, endRev string) (string, error) {
	logrus.Info("Generating release notes")

	notesOptions := options.New()
	notesOptions.Branch = branch
	notesOptions.StartRev = startRev
	notesOptions.EndRev = endRev
	notesOptions.GithubOrg = git.DefaultGithubOrg
	notesOptions.GithubRepo = git.DefaultGithubRepo
	notesOptions.GithubToken = changelogOpts.token
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.ReleaseBucket = changelogOpts.bucket
	notesOptions.ReleaseTars = changelogOpts.tars
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return "", err
	}

	gatherer := notes.NewGatherer(context.Background(), notesOptions)
	releaseNotes, history, err := gatherer.ListReleaseNotes()
	if err != nil {
		return "", errors.Wrapf(err, "listing release notes")
	}

	// Create the markdown
	doc, err := notes.CreateDocument(releaseNotes, history)
	if err != nil {
		return "", errors.Wrapf(err, "creating release note document")
	}

	markdown, err := notes.RenderMarkdown(
		doc, changelogOpts.bucket, changelogOpts.tars,
		notesOptions.StartRev, notesOptions.EndRev,
	)
	if err != nil {
		return "", errors.Wrapf(
			err, "rendering release notes to markdown",
		)
	}

	return markdown, nil
}

func writeMarkdown(repo *git.Repo, toc, markdown string, tag semver.Version) error {
	changelogPath := markdownChangelogFilename(repo, tag)
	writeFile := func(t, m string) error {
		return ioutil.WriteFile(
			changelogPath, []byte(strings.Join(
				[]string{addTocMarkers(t), strings.TrimSpace(m)}, "\n",
			)), 0o644,
		)
	}

	// No changelog exists, simply write the content to a new one
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		logrus.Infof("Changelog %q does not exist, creating it", changelogPath)
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
		"%s\n%s",
		strings.TrimSpace(markdown),
		string(content[(len(tocEnd)+tocEndIndex):]),
	)
	mergedTOC, err := notes.GenerateTOC(mergedMarkdown)
	if err != nil {
		return err
	}
	return writeFile(mergedTOC, mergedMarkdown)
}

func htmlChangelogFilename(tag semver.Version) string {
	return changelogFilename(tag, "html")
}

func markdownChangelogFilename(repo *git.Repo, tag semver.Version) string {
	return filepath.Join(repo.Dir(), changelogFilename(tag, "md"))
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

func writeHTML(tag semver.Version, markdown string) error {
	content := blackfriday.Run([]byte(markdown))

	t, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	output := bytes.Buffer{}
	if err := t.Execute(&output, struct {
		Title, Content string
	}{util.AddTagPrefix(tag.String()), string(content)}); err != nil {
		return err
	}

	absOutputPath, err := filepath.Abs(htmlChangelogFilename(tag))
	if err != nil {
		return err
	}
	logrus.Infof("Writing single HTML to %s", absOutputPath)
	return ioutil.WriteFile(absOutputPath, output.Bytes(), 0o644)
}

func lookupRemoteReleaseNotes(branch string) (string, error) {
	logrus.Info("Assuming new minor release")

	remote := fmt.Sprintf(
		"https://raw.githubusercontent.com/kubernetes/sig-release/master/"+
			"releases/%s/release-notes-draft.md", branch,
	)
	resp, err := http.Get(remote)
	if err != nil {
		return "", errors.Wrapf(err,
			"fetching release notes from remote: %s", remote,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf(
			"remote release notes not found at: %s", remote,
		)
	}
	logrus.Info("Found release notes")

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func commitChanges(repo *git.Repo, branch string, tag semver.Version) error {
	// Master branch modifications
	filename := filepath.Base(markdownChangelogFilename(repo, tag))
	logrus.Infof("Adding %s to repository", filename)
	if err := repo.Add(filename); err != nil {
		return errors.Wrapf(err, "trying to add file %s to repository", filename)
	}

	logrus.Info("Committing changes to master branch in repository")
	if err := repo.Commit(fmt.Sprintf(
		"Add %s for %s", filename, util.AddTagPrefix(tag.String()),
	)); err != nil {
		return errors.Wrap(err, "committing changes into repository")
	}

	// Release branch modifications
	if err := repo.CheckoutBranch(branch); err != nil {
		return errors.Wrapf(err, "checking out release branch %s", branch)
	}

	logrus.Info("Checking out changelog from master branch")
	if err := repo.Checkout(git.Master, filename); err != nil {
		return errors.Wrap(err, "checking out master branch changelog")
	}

	logrus.Info("Committing changes to release branch in repository")
	if err := repo.Commit(fmt.Sprintf(
		"Update %s for %s", filename, util.AddTagPrefix(tag.String()),
	)); err != nil {
		return errors.Wrap(err, "committing changes into repository")
	}

	return nil
}

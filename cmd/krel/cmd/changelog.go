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
	"github.com/google/go-github/v28/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"gopkg.in/russross/blackfriday.v2"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/util"
)

// changelogCmd represents the subcommand for `krel changelog`
var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "changelog maintains the lifecycle of CHANGELOG-x.y.md files",
	Long: `krel changelog

The 'changelog' subcommand of 'krel' does the following things by utilizing
the golang based 'release-notes' tool:

1. Generate the release notes for either the patch or the new minor release
   a) Createa a new CHANGELOG-x.y.md file if we're working on a minor release
   b) Correctly prepend the generated notes to the existing CHANGELOG-x.y.md
      file if it’s a patch release. This also includes the table of contents.

2. Push the modified CHANGELOG-x.y.md into the master branch of
   kubernetes/kubernetes
   a) Push the release notes to the 'release-x.y' branch as well if it’s a
      patch release

3. Convert the markdown release notes into a HTML equivalent on purpose of
   sending it by mail to the announce list. Sending the announcement is done
   by another subcommand of 'krel', not "changelog'.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE:       initLogging,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangelog()
	},
}

type changelogOptions struct {
	tag       string
	bucket    string
	tars      string
	token     string
	outputDir string
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
	changelogCmd.PersistentFlags().StringVarP(&changelogOpts.token, tokenFlag, "t", "", "GitHub token for release notes retrieval")
	changelogCmd.PersistentFlags().StringVarP(&changelogOpts.outputDir, "output", "o", os.TempDir(), "Output directory for the generated files")

	if err := changelogCmd.MarkPersistentFlagRequired(tagFlag); err != nil {
		logrus.Fatal(err)
	}
	if err := changelogCmd.MarkPersistentFlagRequired(tarsFlag); err != nil {
		logrus.Fatal(err)
	}

	rootCmd.AddCommand(changelogCmd)
}

func runChangelog() (err error) {
	tag, err := semver.Make(util.TrimTagPrefix(changelogOpts.tag))
	if err != nil {
		return err
	}
	branch := fmt.Sprintf("release-%d.%d", tag.Major, tag.Minor)
	logrus.Infof("Using release branch %s", branch)

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

	toc, err := notes.GenerateTOC(markdown)
	if err != nil {
		return err
	}

	if err := writeMarkdown(toc, markdown, tag); err != nil {
		return err
	}

	if err := writeHTML(tag, markdown); err != nil {
		return err
	}

	// TODO: Push changes into repo
	return nil
}

func generateReleaseNotes(branch, startRev, endRev string) (string, error) {
	logrus.Info("Generating release notes")

	notesOptions := notes.NewOptions()
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

	// Create the GitHub API client
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: changelogOpts.token},
	))
	githubClient := github.NewClient(httpClient)

	gatherer := &notes.Gatherer{
		Client:  notes.WrapGithubClient(githubClient),
		Context: ctx,
		Org:     git.DefaultGithubOrg,
		Repo:    git.DefaultGithubRepo,
	}
	releaseNotes, history, err := gatherer.ListReleaseNotes(
		branch, notesOptions.StartSHA, notesOptions.EndSHA, "", "",
	)
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

func writeMarkdown(toc, markdown string, tag semver.Version) error {
	changelogPath := changelogFilename(tag, "md")
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
	logrus.Infof("Adding new content to changelog file %q ", changelogPath)
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

func changelogFilename(tag semver.Version, ext string) string {
	return filepath.Join(
		changelogOpts.outputDir,
		fmt.Sprintf("CHANGELOG-%d.%d.%s", tag.Major, tag.Minor, ext),
	)
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

	outputPath := changelogFilename(tag, "html")
	logrus.Infof("Writing single HTML to %s", outputPath)
	return ioutil.WriteFile(outputPath, output.Bytes(), 0o644)
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

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

package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/document"
	"k8s.io/release/pkg/notes/options"
	"k8s.io/release/pkg/util"
)

const (
	// draftFilename filename for the release notes draft
	draftFilename = "release-notes-draft.md"
	// defaultKubernetesSigsOrg GitHub org owner of the release-notes repo
	defaultKubernetesSigsOrg = "kubernetes-sigs"
	// defaultKubernetesSigsRepo relnotes.k8s.io repository name
	defaultKubernetesSigsRepo = "release-notes"
	// userForkName The name we will give to the user's remote when adding it to repos
	userForkName = "userfork"
	// assetsFilePath Path to the assets.ts file
	assetsFilePath = "src/environments/assets.ts"
)

// releaseNotesCmd represents the subcommand for `krel release-notes`
var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "The subcommand of choice for the Release Notes subteam of SIG Release",
	Long: fmt.Sprintf(`krel release-notes

The 'release-notes' subcommand of krel has been developed to:

1. Generate the release notes for the provided tag for commits on the master
   branch. We always use the master branch because a release branch
   gets fast-forwarded until we hit the first release candidate (rc). This is
   also the reason why we select the first 'v1.xx.0-rc.1' as start tag for
   the notes generation.

2. Put the generated notes into a release notes draft markdown document and
   create a GitHub pull request targeting to update the file:
   https://github.com/kubernetes/sig-release/blob/master/releases/release-1.xx/release-notes-draft.md

3. Put the generated notes into a JSON file and create a GitHub pull request
   to update the website https://relnotes.k8s.io.

To use the tool, please set the %v environment variable which needs write
permissions to your fork of k/sig-release and k-sigs/release-notes.`,
		github.TokenEnvKey),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReleaseNotes()
	},
}

type releaseNotesOptions struct {
	tag             string
	draftOrg        string
	draftRepo       string
	createDraftPR   bool
	createWebsitePR bool
	outputDir       string
	Format          string
	websiteOrg      string
	websiteRepo     string
}

type releaseNotesResult struct {
	markdown string
	json     string
}

var releaseNotesOpts = &releaseNotesOptions{}

func init() {
	releaseNotesCmd.PersistentFlags().StringVarP(
		&releaseNotesOpts.tag,
		"tag",
		"t",
		"",
		"version tag for the notes",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.draftOrg,
		"draft-org",
		"",
		"a Github organization owner of the fork of k/sig-release where the Release Notes Draft PR will be created",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.draftRepo,
		"draft-repo",
		git.DefaultGithubReleaseRepo,
		"the name of the fork of k/sig-release, the Release Notes Draft PR will be created from this repository",
	)

	releaseNotesCmd.PersistentFlags().BoolVar(
		&releaseNotesOpts.createDraftPR,
		"create-draft-pr",
		false,
		"create the Release Notes Draft PR. --draft-org and --draft-repo must be set along with this option",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.websiteOrg,
		"website-org",
		"",
		"a Github organization owner of the fork of kuberntets-sigs/release-notes where the Website PR will be created",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.websiteRepo,
		"website-repo",
		"release-notes",
		"the name of the fork of kuberntets-sigs/release-notes, the Release Notes Draft PR will be created from this repository",
	)

	releaseNotesCmd.PersistentFlags().BoolVar(
		&releaseNotesOpts.createWebsitePR,
		"create-website-pr",
		false,
		"generate the Releas Notes to a local fork of relnotes.k8s.io and create a PR.  --draft-org and --draft-repo must be set along with this option",
	)

	releaseNotesCmd.PersistentFlags().StringVarP(
		&releaseNotesOpts.outputDir,
		"output-dir",
		"o",
		".",
		"output a copy of the release notes to this directory",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.Format,
		"format",
		util.EnvDefault("FORMAT", "markdown"),
		"The format for notes output (options: markdown, json)",
	)

	rootCmd.AddCommand(releaseNotesCmd)
}

func runReleaseNotes() (err error) {
	var tag string
	if releaseNotesOpts.tag == "" {
		tag, err = tryToFindLatestMinorTag()
		if err != nil {
			return errors.Wrapf(err, "unable to find latest minor tag")
		}
		releaseNotesOpts.tag = tag
	} else {
		tag = releaseNotesOpts.tag
	}

	s, err := util.TagStringToSemver(tag)
	if err != nil {
		return errors.Wrapf(err, "no valid tag: %v", tag)
	}
	start := util.SemverToTagString(semver.Version{
		Major: s.Major,
		Minor: s.Minor - 1,
		Patch: 0,
		Pre:   []semver.PRVersion{{VersionStr: "rc.1"}},
	})

	var result *releaseNotesResult

	// Create the PR for relnotes.k8s.io
	if releaseNotesOpts.createWebsitePR {
		// Check cmd line options
		if err = validateWebsitePROptions(); err != nil {
			return errors.Wrap(err, "validating PR command line options")
		}

		// Generate the release notes for ust the current tag
		jsonStr, err := releaseNotesJSON(tag)
		if err != nil {
			return errors.Wrapf(err, "generating release notes")
		}

		// Run the website PR process
		if err := createWebsitePR(tag, jsonStr); err != nil {
			return errors.Wrapf(err, "generating releasenotes for tag %s", tag)
		}
		return nil
	}

	// Create the PR for the Release Notes Draft in k/sig-release
	if releaseNotesOpts.createDraftPR {
		// Check cmd line options
		if releaseNotesOpts.createDraftPR {
			if err = validateDraftPROptions(); err != nil {
				return errors.Wrap(err, "validating PR command line options")
			}
		}

		// Generate the notes for the current version
		result, err = releaseNotesFrom(start)
		if err != nil {
			return errors.Wrapf(err, "while generating the release notes for tag %s", start)
		}

		// Create the Draft PR Process
		if err := createDraftPR(tag, result); err != nil {
			return errors.Wrap(err, "failed to create release notes draft PR")
		}
		return nil
	}

	// Otherwise, generate the release notes to a file
	result, err = releaseNotesFrom(start)
	if err != nil {
		return errors.Wrap(err, "generating release notes to file")
	}

	switch releaseNotesOpts.Format {
	case "json":
		err = ioutil.WriteFile(filepath.Join(releaseNotesOpts.outputDir, "release-notes.json"), []byte(result.json), 0644)
		if err != nil {
			return errors.Wrap(err, "writing release notes JSON file")
		}
	case "markdown":
		err = ioutil.WriteFile(filepath.Join(releaseNotesOpts.outputDir, "release-notes.md"), []byte(result.json), 0644)
		if err != nil {
			return errors.Wrap(err, "writing release notes markdown file")
		}
	default:
		return errors.Errorf("%q is an unsupported format", releaseNotesOpts.Format)
	}

	return nil
}

// validateDraftPROptions checks if we have all needed parameters to create the Release Notes PR
func validateDraftPROptions() error {
	if releaseNotesOpts.createDraftPR {
		// Check if --create-website-pr is set
		if releaseNotesOpts.createWebsitePR {
			return errors.New("Cannot create release notes draft if --create-website-pr is set")
		}

		// Check if --draft-org is set
		if releaseNotesOpts.draftOrg == "" {
			logrus.Warn("cannot generate the Release Notes PR without --draft-org")
		}

		// Check if --draft-repo is set
		if releaseNotesOpts.draftRepo == "" {
			logrus.Warn("cannot generate the Release Notes PR without --draft-repo")
		}

		if releaseNotesOpts.draftOrg == "" || releaseNotesOpts.draftRepo == "" {
			return errors.New("To generate the release notes PR you must define both --draft-org and --draft-repo")
		}
	}

	return nil
}

// validateWebsitePROptions checks if we have all needed parameters to create the Release Notes PR
func validateWebsitePROptions() error {
	if releaseNotesOpts.createWebsitePR {
		// Check if --website-org is set
		if releaseNotesOpts.websiteOrg == "" {
			logrus.Warn("cannot generate the Website PR without --website-org")
		}

		// Check if --website-repo is set
		if releaseNotesOpts.websiteRepo == "" {
			logrus.Warn("cannot generate the Website PR without --website-repo")
		}

		if releaseNotesOpts.websiteOrg == "" || releaseNotesOpts.websiteRepo == "" {
			return errors.New("To generate the website PR you must define both --website-org and --website-repo")
		}
	}
	return nil
}

// createDraftPR pushes the release notes draft to the users fork
func createDraftPR(tag string, result *releaseNotesResult) (err error) {
	s, err := util.TagStringToSemver(tag)
	if err != nil {
		return errors.Wrapf(err, "no valid tag: %v", tag)
	}

	branchname := "release-notes-draft-" + tag

	// Prepare the fork of k/sig-release
	sigReleaseRepo, err := prepareFork(
		branchname,
		git.DefaultGithubOrg, git.DefaultGithubReleaseRepo,
		releaseNotesOpts.draftOrg, releaseNotesOpts.draftRepo,
	)
	if err != nil {
		return errors.Wrap(err, "preparing local fork of kubernetes/sig-release")
	}

	defer func() {
		err = sigReleaseRepo.Cleanup()
	}()

	// generate the notes
	targetdir := filepath.Join(sigReleaseRepo.Dir(), "releases", fmt.Sprintf("release-%d.%d", s.Major, s.Minor))
	logrus.Debugf("release notes markdown will be written to %v", targetdir)
	err = ioutil.WriteFile(filepath.Join(targetdir, draftFilename), []byte(result.markdown), 0644)
	if err != nil {
		return errors.Wrapf(err, "writing release notes draft")
	}

	// add the updated file
	if err := sigReleaseRepo.Add(filepath.Join("releases", fmt.Sprintf("release-%d.%d", s.Major, s.Minor), draftFilename)); err != nil {
		return errors.Wrap(err, "adding release notes draft to staging area")
	}

	// commit the changes
	if err := sigReleaseRepo.UserCommit("Release Notes draft for k/k " + tag); err != nil {
		return errors.Wrapf(err, "creating commit in %v/%v", releaseNotesOpts.draftOrg, releaseNotesOpts.draftRepo)
	}

	// push to the user's remote
	logrus.Infof("Pushing modified release notes draft to %v/%v", releaseNotesOpts.draftOrg, releaseNotesOpts.draftRepo)
	if err := sigReleaseRepo.PushToRemote(userForkName, branchname); err != nil {
		return errors.Wrapf(err, "pushing %v to remote", userForkName)
	}

	// TODO: Call github API and create PR against k/sig-release
	return nil
}

// prepareFork Prepare a branch a repo
func prepareFork(branchName, upstreamOrg, upstreamRepo, myOrg, myRepo string) (repo *git.Repo, err error) {
	// checkout the upstream repository
	logrus.Infof("cloning/updating repository %s/%s", upstreamOrg, upstreamRepo)

	repo, err = git.CleanCloneGitHubRepo(
		upstreamOrg, upstreamRepo, false,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "cloning %s/%s", upstreamOrg, upstreamRepo)
	}

	// test if the fork remote is already existing
	url := git.GetRepoURL(myOrg, myRepo, false)
	if repo.HasRemote(userForkName, url) {
		logrus.Infof(
			"Using already existing remote %v (%v) in repository",
			userForkName, url,
		)
	} else {
		// add the user's fork as a remote
		err = repo.AddRemote(userForkName, myOrg, myRepo)
		if err != nil {
			return nil, errors.Wrap(err, "adding user's fork as remote repository")
		}
	}

	// checkout the new branch
	err = repo.Checkout("-B", branchName)
	if err != nil {
		return nil, errors.Wrapf(err, "creating new branch %s", branchName)
	}

	return repo, nil
}

// addReferenceToAssetsFile adds a new entry in the assets.ts file in repoPath to include newJsonFile
func addReferenceToAssetsFile(repoPath, newJSONFile string) error {
	// Full  filesystem path to the assets.ts file
	assetsFullPath := filepath.Join(repoPath, assetsFilePath)

	file, err := os.Open(assetsFullPath)
	if err != nil {
		return errors.Wrap(err, "opening assets.ts to check for current version")
	}
	defer file.Close()

	logrus.Infof("Writing json reference to %s in %s", newJSONFile, assetsFullPath)

	scanner := bufio.NewScanner(file)
	var assetsBuffer bytes.Buffer
	assetsFileWasModified := false
	for scanner.Scan() {
		// Check if the assets file already has the json notes referenced:
		if strings.Contains(scanner.Text(), fmt.Sprintf("assets/%s", newJSONFile)) {
			return errors.New(fmt.Sprintf("assets.ts already has a reference to %s ", newJSONFile))
		}

		assetsBuffer.WriteString(scanner.Text())

		// Add the current version right after the array export
		if strings.Contains(scanner.Text(), "export const assets =") {
			assetsBuffer.WriteString(fmt.Sprintf("  'assets/%s',\n", newJSONFile))
			assetsFileWasModified = true
		}
	}

	// Return an error if the array decalra
	if !assetsFileWasModified {
		return errors.New("unable to modify assets file, could not find assets array declaration")
	}

	// write the modified assets.ts file
	if err := ioutil.WriteFile(assetsFullPath, assetsBuffer.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "writing assets.ts file")
	}

	return nil
}

// processJSONOutput Runs NPM prettier inside repoPath to format the JSON output
func processJSONOutput(repoPath string) error {
	npmpath, err := exec.LookPath("npm")
	if err != nil {
		return errors.Wrap(err, "while looking for npm in your path")
	}

	// run npm install
	logrus.Info("Installing npm modules, this can take a while")
	if err := command.NewWithWorkDir(repoPath, npmpath, "install").RunSuccess(); err != nil {
		return errors.Wrap(err, "running npm install in kubernetes-sigs/release-notes")
	}

	// run npm prettier
	logrus.Info("Running npm prettier...")
	if err := command.NewWithWorkDir(repoPath, npmpath, "run", "prettier").RunSuccess(); err != nil {
		return errors.Wrap(err, "running npm prettier in kubernetes-sigs/release-notes")
	}

	return nil
}

// createWebsitePR creates the JSON version of the release notes and pushes them to a user fork
func createWebsitePR(tag, jsonStr string) (err error) {
	_, err = util.TagStringToSemver(tag)
	if err != nil {
		return errors.Wrapf(err, "no valid tag: %v", tag)
	}

	jsonNotesFilename := fmt.Sprintf("release-notes-%s.json", tag[1:])
	branchname := "release-notes-json-" + tag

	// checkout kubernetes-sigs/release-notes
	k8sSigsRepo, err := prepareFork(
		branchname, defaultKubernetesSigsOrg,
		defaultKubernetesSigsRepo, releaseNotesOpts.websiteOrg, releaseNotesOpts.websiteRepo,
	)
	if err != nil {
		return errors.Wrap(err, "preparing local fork branch")
	}
	defer func() {
		err = k8sSigsRepo.Cleanup()
	}()

	// add a reference to the new json file in assets.ts
	if err := addReferenceToAssetsFile(k8sSigsRepo.Dir(), jsonNotesFilename); err != nil {
		return errors.Wrapf(err, "adding %s to assets file", jsonNotesFilename)
	}

	// generate the notes
	jsonNotesPath := filepath.Join("src", "assets", jsonNotesFilename)
	logrus.Debugf("Release notes json file will be written to %s", filepath.Join(k8sSigsRepo.Dir(), jsonNotesPath))
	err = ioutil.WriteFile(filepath.Join(k8sSigsRepo.Dir(), jsonNotesPath), []byte(jsonStr), 0644)
	if err != nil {
		return errors.Wrapf(err, "writing release notes json file")
	}

	// Run NPM prettier
	if err := processJSONOutput(k8sSigsRepo.Dir()); err != nil {
		return errors.Wrap(err, "while formatting release notes JSON files")
	}

	// add the modified files & commit the results
	if err := k8sSigsRepo.Add(jsonNotesPath); err != nil {
		return errors.Wrap(err, "adding release notes draft to staging area")
	}

	if err := k8sSigsRepo.Add(filepath.FromSlash(assetsFilePath)); err != nil {
		return errors.Wrap(err, "adding release notes draft to staging area")
	}

	if err := k8sSigsRepo.UserCommit(fmt.Sprintf("Patch relnotes.k8s.io with release %s", tag)); err != nil {
		return errors.Wrapf(err, "Error creating commit in %s/%s", releaseNotesOpts.websiteOrg, releaseNotesOpts.websiteRepo)
	}

	// push to the user's fork
	logrus.Infof("Pushing website changes to %s/%s", releaseNotesOpts.websiteOrg, releaseNotesOpts.websiteRepo)
	if err := k8sSigsRepo.PushToRemote(userForkName, branchname); err != nil {
		return errors.Wrapf(err, "pushing %v to %v/%v", userForkName, releaseNotesOpts.websiteOrg, releaseNotesOpts.websiteRepo)
	}

	// TODO: Call github API and create PR against k/sig-release
	return nil
}

// tryToFindPreviuousTag gets a release tag and returns the one before it
func tryToFindPreviuousTag(tag string) (string, error) {
	url := git.GetDefaultKubernetesRepoURL()
	status, err := command.New(
		"git", "ls-remote", "--sort=v:refname",
		"--tags", url,
	).
		Pipe("grep", "-Eo", "v[0-9].[0-9]+.[0-9]+[-a-z0-9\\.]*$").
		Pipe("grep", "-B1", tag).
		Pipe("head", "-1").
		RunSilentSuccessOutput()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(status.Output()), nil
}

// tryToFindLatestMinorTag looks-up the default k/k remote to find the latest
// non final version
func tryToFindLatestMinorTag() (string, error) {
	url := git.GetDefaultKubernetesRepoURL()
	status, err := command.New(
		"git", "ls-remote", "--sort=v:refname",
		"--tags", url,
	).
		Pipe("grep", "-Eo", "v[0-9].[0-9]+.0-.*.[0-9]$").
		Pipe("tail", "-1").
		RunSilentSuccessOutput()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(status.Output()), nil
}

// releaseNotesJSON generate the release notes for a specific tag and returns
// them as JSON blob
func releaseNotesJSON(tag string) (string, error) {
	logrus.Infof("Generating release notes for tag %s", tag)

	// TODO: Change this logic to get the tag from git.PreviousTag
	startTag, err := tryToFindPreviuousTag(tag)
	if err != nil {
		return "", errors.Wrap(err, "trying to get previous tag")
	}

	notesOptions := options.New()
	notesOptions.Branch = git.Master
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.StartRev = startTag
	notesOptions.EndRev = tag
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOptions.ReleaseVersion = util.TrimTagPrefix(tag)

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return "", err
	}

	logrus.Infof("Using start tag %v", startTag)
	logrus.Infof("Using end tag %v", tag)

	// Fetch the notes
	releaseNotes, history, err := notes.GatherReleaseNotes(notesOptions)
	if err != nil {
		return "", errors.Wrapf(err, "gathering release notes")
	}

	doc, err := document.New(
		releaseNotes, history, notesOptions.StartRev, notesOptions.EndRev,
	)
	if err != nil {
		return "", errors.Wrapf(err, "creating release note document")
	}
	doc.PreviousRevision = startTag
	doc.CurrentRevision = tag

	// Create the JSON
	j, err := json.Marshal(releaseNotes)
	if err != nil {
		return "", errors.Wrapf(err, "generating release notes JSON")
	}

	return string(j), nil
}

func releaseNotesFrom(startTag string) (*releaseNotesResult, error) {
	logrus.Info("Generating release notes")

	notesOptions := options.New()
	notesOptions.Branch = git.Master
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.StartRev = startTag
	notesOptions.EndRev = releaseNotesOpts.tag
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOptions.ReleaseVersion = util.TrimTagPrefix(releaseNotesOpts.tag)

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return nil, err
	}

	logrus.Infof("Using start tag %v", startTag)
	logrus.Infof("Using end tag %v", releaseNotesOpts.tag)

	// Fetch the notes
	releaseNotes, history, err := notes.GatherReleaseNotes(notesOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "gathering release notes")
	}

	doc, err := document.New(
		releaseNotes, history, notesOptions.StartRev, notesOptions.EndRev,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "creating release note document")
	}
	doc.PreviousRevision = startTag
	doc.CurrentRevision = releaseNotesOpts.tag

	// Create the markdown
	markdown, err := doc.RenderMarkdownTemplate(
		"", "", options.FormatSpecDefaultGoTemplate,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err, "rendering release notes to markdown",
		)
	}

	// Create the JSON
	j, err := json.Marshal(releaseNotes)
	if err != nil {
		return nil, errors.Wrapf(err, "generating release notes JSON")
	}

	return &releaseNotesResult{markdown: markdown, json: string(j)}, nil
}

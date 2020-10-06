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
	"time"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/editor"
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
	// serviceDirectory is where we keep the files used to generate the notes
	releaseNotesWorkDir = "release-notes"
	// mapsMainDirectory is where we will save the release notes maps
	mapsMainDirectory = "maps"
	// mapsCVEDirectory holds the maps with CVE data
	mapsCVEDirectory = "cve"
	// Directory where session editing files are stored
	mapsSessionDirectory = "sessions"
	// The themes directory holds the maps for the release notes major themes
	mapsThemesDirectory = "themes"
	// defaultKubernetesSigsOrg GitHub org owner of the release-notes repo
	defaultKubernetesSigsOrg = "kubernetes-sigs"
	// defaultKubernetesSigsRepo relnotes.k8s.io repository name
	defaultKubernetesSigsRepo = "release-notes"
	// userForkName The name we will give to the user's remote when adding it to repos
	userForkName = "userfork"
	// assetsFilePath Path to the assets.ts file
	assetsFilePath = "src/environments/assets.ts"
	// websiteBranchPrefix name of the website branch in the user's fork
	websiteBranchPrefix = "release-notes-json-"
	// draftBranchPrefix name of the draft branch in the user's fork
	draftBranchPrefix = "release-notes-draft-"
	// Editing instructions for notemaps
	mapEditingInstructions = `# This is the current map for this Pull Request.
# The original note content is commented out, if you need to
# change a field, remove the comment and change the value.
# To cancel, exit without changing anything or leave the file blank.
# Important! pr: and releasenote: have to be uncommented.
#`
)

// releaseNotesCmd represents the subcommand for `krel release-notes`
var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "The subcommand of choice for the Release Notes subteam of SIG Release",
	Long: fmt.Sprintf(`krel release-notes

The 'release-notes' subcommand of krel has been developed to:

1. Generate the release notes for the provided tag for commits on the main
   branch. We always use the main branch because a release branch
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// If none of the operation modes is defined, show the usage help and exit
		if !releaseNotesOpts.createDraftPR &&
			!releaseNotesOpts.createWebsitePR {
			if err := cmd.Help(); err != nil {
				return err
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Run the PR creation function
		return runReleaseNotes()
	},
}

type releaseNotesOptions struct {
	tag             string
	userFork        string
	createDraftPR   bool
	createWebsitePR bool
	dependencies    bool
	fixNotes        bool
	websiteRepo     string
	mapProviders    []string
	githubOrg       string
	draftRepo       string
}

type releaseNotesResult struct {
	markdown string
	json     string
}

// A datatype to record a notes repair session
type sessionData struct {
	UserEmail    string `json:"mail"`
	UserName     string `json:"name"`
	Date         int64  `json:"date"`
	PullRequests []struct {
		Number int    `json:"nr"`
		Hash   string `json:"hash"`
	} `json:"prs"`
	Path string `json:"-"`
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

	releaseNotesCmd.PersistentFlags().BoolVar(
		&releaseNotesOpts.createDraftPR,
		"create-draft-pr",
		false,
		"update the Release Notes draft and create a PR in k/sig-release",
	)

	releaseNotesCmd.PersistentFlags().BoolVar(
		&releaseNotesOpts.createWebsitePR,
		"create-website-pr",
		false,
		"patch the relnotes.k8s.io sources and generate a PR with the changes",
	)

	releaseNotesCmd.PersistentFlags().BoolVar(
		&releaseNotesOpts.dependencies,
		"dependencies",
		true,
		"add dependency report",
	)

	releaseNotesCmd.PersistentFlags().StringSliceVarP(
		&releaseNotesOpts.mapProviders,
		"maps-from",
		"m",
		[]string{},
		"specify a location to recursively look for release notes *.y[a]ml file mappings",
	)

	releaseNotesCmd.PersistentFlags().BoolVar(
		&releaseNotesOpts.fixNotes,
		"fix",
		false,
		"fix release notes",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.userFork,
		"fork",
		"",
		"the user's fork in the form org/repo. Used to submit Pull Requests for the website and draft",
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

	if releaseNotesOpts.userFork != "" {
		org, repo, err := git.ParseRepoSlug(releaseNotesOpts.userFork)
		if err != nil {
			return errors.Wrap(err, "parsing the user's fork")
		}
		releaseNotesOpts.githubOrg = org
		releaseNotesOpts.websiteRepo, releaseNotesOpts.draftRepo = repo, repo
		// If the slug did not have a repo, use the defaults
		if repo == "" {
			releaseNotesOpts.websiteRepo = defaultKubernetesSigsRepo
			releaseNotesOpts.draftRepo = git.DefaultGithubReleaseRepo
		}
	}

	// First, validate cmdline options
	if err := releaseNotesOpts.Validate(); err != nil {
		return errors.Wrap(err, "validating command line options")
	}

	// before running the generators, verify that the repositories are ready
	if releaseNotesOpts.createWebsitePR {
		if err = verifyFork(
			websiteBranchPrefix+tag,
			releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo,
			defaultKubernetesSigsOrg, defaultKubernetesSigsRepo,
		); err != nil {
			return errors.Wrapf(err, "while checking %s/%s fork", defaultKubernetesSigsOrg, defaultKubernetesSigsRepo)
		}
	}

	if releaseNotesOpts.createDraftPR {
		if err = verifyFork(
			draftBranchPrefix+tag,
			releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo,
			git.DefaultGithubOrg, git.DefaultGithubReleaseRepo,
		); err != nil {
			return errors.Wrapf(err, "while checking %s/%s fork", defaultKubernetesSigsOrg, git.DefaultGithubReleaseRepo)
		}
	}

	// Create the PR for relnotes.k8s.io
	if releaseNotesOpts.createWebsitePR {
		// Run the website PR process
		if err := createWebsitePR(tag); err != nil {
			return errors.Wrap(err, "creating website PR")
		}
	}

	// Create the PR for the Release Notes Draft in k/sig-release
	if releaseNotesOpts.createDraftPR {
		// Create the Draft PR Process
		if err := createDraftPR(tag); err != nil {
			return errors.Wrap(err, "creating Draft PR")
		}
	}

	if releaseNotesOpts.createDraftPR || releaseNotesOpts.createWebsitePR {
		logrus.Info("Release notes generation complete!")
	}

	return nil
}

// verifyFork does a pre-check of a fork to see if we can create a PR from it
func verifyFork(branchName, forkOwner, forkRepo, parentOwner, parentRepo string) error {
	logrus.Infof("Checking if a PR can be created from %s/%s", forkOwner, forkRepo)
	gh := github.New()

	// Check th PR
	isrepo, err := gh.RepoIsForkOf(
		forkOwner, forkRepo, parentOwner, parentRepo,
	)
	if err != nil {
		return errors.Wrapf(
			err, "while checking if repository is a fork of %s/%s",
			parentOwner, parentRepo,
		)
	}

	if !isrepo {
		return errors.Errorf(
			"cannot create PR, %s/%s is not a fork of %s/%s",
			forkOwner, forkRepo, parentOwner, parentRepo,
		)
	}

	// verify the branch does not previously exist
	branchExists, err := gh.BranchExists(
		forkOwner, forkRepo, branchName,
	)
	if err != nil {
		return errors.Wrap(err, "while checking if branch can be created")
	}

	if branchExists {
		return errors.Errorf(
			"a branch named %s already exists in %s/%s",
			branchName, forkOwner, forkRepo,
		)
	}
	return nil
}

// createDraftPR pushes the release notes draft to the users fork
func createDraftPR(tag string) (err error) {
	tagVersion, err := util.TagStringToSemver(tag)
	if err != nil {
		return errors.Wrapf(err, "reading tag: %s", tag)
	}

	// Release notes are built from the first RC in the previous
	// minor to encompass all changes received after Code Thaw,
	// the point where the last minor was forked.
	start := util.SemverToTagString(semver.Version{
		Major: tagVersion.Major,
		Minor: tagVersion.Minor - 1,
		Patch: 0,
		Pre:   []semver.PRVersion{{VersionStr: "rc.1"}},
	})

	gh := github.New()
	autoCreatePullRequest := true

	// Verify the repository
	isrepo, err := gh.RepoIsForkOf(
		releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo,
		git.DefaultGithubOrg, git.DefaultGithubReleaseRepo,
	)
	if err != nil {
		return errors.Wrapf(
			err, "while checking if repository is a fork of %s/%s",
			git.DefaultGithubOrg, git.DefaultGithubReleaseRepo,
		)
	}

	if !isrepo {
		return errors.New(
			fmt.Sprintf(
				"Cannot create PR, %s/%s is not a fork of %s/%s",
				releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo,
				git.DefaultGithubOrg, git.DefaultGithubReleaseRepo,
			),
		)
	}

	// Generate the notes for the current version
	releaseNotes, err := gatherNotesFrom(start)
	if err != nil {
		return errors.Wrapf(err, "while generating the release notes for tag %s", start)
	}

	branchname := draftBranchPrefix + tag

	// Prepare the fork of k/sig-release
	sigReleaseRepo, err := prepareFork(
		branchname,
		git.DefaultGithubOrg, git.DefaultGithubReleaseRepo,
		releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo,
	)
	if err != nil {
		return errors.Wrap(err, "preparing local fork of kubernetes/sig-release")
	}

	// The release path inside the repository
	releasePath := filepath.Join("releases", fmt.Sprintf("release-%d.%d", tagVersion.Major, tagVersion.Minor))

	// Check if the directory exists
	releaseDir := filepath.Join(sigReleaseRepo.Dir(), releasePath)
	if !util.Exists(releaseDir) {
		return errors.New(fmt.Sprintf("could not find release directory %s", releaseDir))
	}

	// If we got the --fix flag, start the fix flow
	if releaseNotesOpts.fixNotes {
		_, _, err = util.Ask("Press enter to start", "y:yes|n:no|y", 10)
		// In interactive mode, we will ask the user before sending the PR
		autoCreatePullRequest = false

		// createNotesWorkDir is idempotent, we can use it to verify the tree is complete
		if err := createNotesWorkDir(releaseDir); err != nil {
			return errors.Wrap(err, "creating working directory")
		}

		// Run the release notes fix flow
		err := fixReleaseNotes(filepath.Join(releaseDir, releaseNotesWorkDir), releaseNotes)
		if err != nil {
			return errors.Wrap(err, "while running release notes fix flow")
		}

		// Create the map provider to read the changes so far
		rnMapProvider, err := notes.NewProviderFromInitString(filepath.Join(releaseDir, releaseNotesWorkDir, mapsMainDirectory))
		if err != nil {
			return errors.Wrap(err, "creating release notes draft")
		}
		for _, note := range releaseNotes.ByPR() {
			maps, err := rnMapProvider.GetMapsForPR(note.PrNumber)
			if err != nil {
				return errors.Wrapf(err, "while getting maps for PR#%d", note.PrNumber)
			}
			for _, noteMap := range maps {
				if err := note.ApplyMap(noteMap); err != nil {
					return errors.Wrapf(err, "applying note maps to pr #%d", note.PrNumber)
				}
			}
		}
	}

	// Generate the results struct
	result, err := buildNotesResult(start, releaseNotes)
	if err != nil {
		return errors.Wrap(err, "building release notes results")
	}

	// generate the notes
	logrus.Debugf("Release notes markdown will be written to %s", releaseDir)
	err = ioutil.WriteFile(filepath.Join(releaseDir, draftFilename), []byte(result.markdown), 0644)
	if err != nil {
		return errors.Wrapf(err, "writing release notes draft")
	}
	logrus.Infof("Release Notes Draft written to %s", filepath.Join(releaseDir, draftFilename))

	// If we are in interactive mode, ask before continuing
	if !autoCreatePullRequest {
		_, autoCreatePullRequest, err = util.Ask("Create pull request with your changes? (y/n)", "y:yes|n:no", 10)
		if err != nil {
			return errors.Wrap(err, "while asking to create pull request")
		}
	}

	if !autoCreatePullRequest {
		fmt.Println("\nPull request has NOT been created. The changes were made to your local copy of k/sig-release.")
		fmt.Println("To complete the process, you will need to:")
		fmt.Println("  1. Review the changes in your local copy")
		fmt.Printf("  2. Push the changes to your fork (git push -u %s %s)\n", userForkName, branchname)
		fmt.Println("  3. Submit a pull request to k/sig-release")
		fmt.Println("\nYou can find your local copy here:")
		fmt.Println(sigReleaseRepo.Dir())
		fmt.Println(nl)
		logrus.Warn("Changes were made locally, user needs to perform manual push and create pull request.")
		return nil
	}

	defer func() {
		err = sigReleaseRepo.Cleanup()
	}()

	// add the updated draft
	if err := sigReleaseRepo.Add(filepath.Join(releasePath, draftFilename)); err != nil {
		return errors.Wrap(err, "adding release notes draft to staging area")
	}

	// List of directories we'll consider for the PR
	releaseDirectories := []struct{ Path, Name, Ext string }{
		{
			Path: filepath.Join(releasePath, releaseNotesWorkDir, mapsMainDirectory),
			Name: "release notes maps", Ext: "yaml",
		},
		{
			Path: filepath.Join(releasePath, releaseNotesWorkDir, mapsSessionDirectory),
			Name: "release notes session files", Ext: "json",
		},
		{
			Path: filepath.Join(releasePath, releaseNotesWorkDir, mapsCVEDirectory),
			Name: "release notes cve data", Ext: "yaml",
		},
		{
			Path: filepath.Join(releasePath, releaseNotesWorkDir, mapsThemesDirectory),
			Name: "release notes major theme files", Ext: "yaml",
		},
	}

	// Add to the PR all files that exist
	for _, dirData := range releaseDirectories {
		// add the updated maps
		if util.Exists(filepath.Join(sigReleaseRepo.Dir(), dirData.Path)) {
			// Check if there are any files to commit
			matches, err := filepath.Glob(filepath.Join(sigReleaseRepo.Dir(), dirData.Path, "*"+dirData.Ext))
			logrus.Debugf("Adding %d %s from %s to commit", len(matches), dirData.Name, dirData.Path)
			if err != nil {
				return errors.Wrapf(err, "checking for %s files in %s", dirData.Ext, dirData.Path)
			}
			if len(matches) > 0 {
				if err := sigReleaseRepo.Add(filepath.Join(dirData.Path, "*"+dirData.Ext)); err != nil {
					return errors.Wrapf(err, "adding %s to staging area", dirData.Name)
				}
			}
		} else {
			logrus.Debugf("Not adding %s files, directory %s not found", dirData.Name, dirData.Path)
		}
	}

	// add the generated draft
	if err := sigReleaseRepo.UserCommit("Release Notes draft for k/k " + tag); err != nil {
		return errors.Wrapf(err, "creating commit in %s/%s", releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo)
	}

	// push to the user's remote
	logrus.Infof("Pushing modified release notes draft to %s/%s", releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo)
	if err := sigReleaseRepo.PushToRemote(userForkName, branchname); err != nil {
		return errors.Wrapf(err, "pushing %s to remote", userForkName)
	}

	// Create a PR against k/sig-release using the github API
	// TODO: Maybe read and parse the PR template from sig-release?
	prBody := "**What type of PR is this?**\n"
	prBody += "/kind documentation\n\n"
	prBody += "**What this PR does / why we need it**:\n"
	prBody += fmt.Sprintf("This PR updates the Release Notes Draft to k/k %s\n\n", tag)
	prBody += "**Which issue(s) this PR fixes**:\n\n"
	prBody += "**Special notes for your reviewer**:\n"
	prBody += "This is an automated PR generated from `krel The Kubernetes Release Toolbox`\n\n"

	// Create the pull request
	logrus.Debugf(
		"PR params: org: %s, repo: %s, headBranch: %s baseBranch: %s",
		git.DefaultGithubOrg, git.DefaultGithubReleaseRepo, git.DefaultBranch,
		fmt.Sprintf("%s:%s", releaseNotesOpts.githubOrg, branchname),
	)

	// Create the PR
	pr, err := gh.CreatePullRequest(
		git.DefaultGithubOrg, git.DefaultGithubReleaseRepo, git.DefaultBranch,
		fmt.Sprintf("%s:%s", releaseNotesOpts.githubOrg, branchname),
		fmt.Sprintf("Update release notes draft to version %s", tag), prBody,
	)
	if err != nil {
		logrus.Warnf("An error has occurred while creating the pull request for %s", tag)
		logrus.Warn("While the PR failed, the release notes draft was generated and submitted to your fork")
		return errors.Wrap(err, "creating the pull request")
	}
	logrus.Infof(
		"Successfully created PR: %s%s/%s/pull/%d",
		github.GitHubURL, git.DefaultGithubOrg,
		git.DefaultGithubReleaseRepo, pr.GetNumber(),
	)
	logrus.Infof("Successfully created PR #%d", pr.GetNumber())

	return nil
}

// prepareFork Prepare a branch a repo
func prepareFork(branchName, upstreamOrg, upstreamRepo, myOrg, myRepo string) (repo *git.Repo, err error) {
	// checkout the upstream repository
	logrus.Infof("Cloning/updating repository %s/%s", upstreamOrg, upstreamRepo)

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
			"Using already existing remote %s (%s) in repository",
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
	fileIsReferenced := false
	for scanner.Scan() {
		// Check if the assets file already has the json notes referenced:
		if strings.Contains(scanner.Text(), fmt.Sprintf("assets/%s", newJSONFile)) {
			logrus.Warnf("File %s is already referenced in assets.ts", newJSONFile)
			fileIsReferenced = true
			break
		}

		assetsBuffer.WriteString(scanner.Text())

		// Add the current version right after the array export
		if strings.Contains(scanner.Text(), "export const assets =") {
			assetsBuffer.WriteString(fmt.Sprintf("  'assets/%s',\n", newJSONFile))
			assetsFileWasModified = true
		}
	}

	if fileIsReferenced {
		logrus.Infof("Not modifying assets.ts since it already has a reference to %s", newJSONFile)
		return nil
	}

	// Return an error if the array decalra
	if !assetsFileWasModified {
		return errors.New("unable to modify assets file, could not find assets array declaration")
	}

	// write the modified assets.ts file
	if err := ioutil.WriteFile(assetsFullPath, assetsBuffer.Bytes(), os.FileMode(0o644)); err != nil {
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
func createWebsitePR(tag string) error {
	_, err := util.TagStringToSemver(tag)
	if err != nil {
		return errors.Wrapf(err, "reading tag: %s", tag)
	}

	// Generate the release notes for ust the current tag
	jsonStr, err := releaseNotesJSON(tag)
	if err != nil {
		return errors.Wrapf(err, "generating release notes in JSON format")
	}

	jsonNotesFilename := fmt.Sprintf("release-notes-%s.json", tag[1:])
	branchname := websiteBranchPrefix + tag

	// checkout kubernetes-sigs/release-notes
	k8sSigsRepo, err := prepareFork(
		branchname, defaultKubernetesSigsOrg,
		defaultKubernetesSigsRepo, releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo,
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
		return errors.Wrapf(err, "Error creating commit in %s/%s", releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo)
	}

	// push to the user's fork
	logrus.Infof("Pushing website changes to %s/%s", releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo)
	if err := k8sSigsRepo.PushToRemote(userForkName, branchname); err != nil {
		return errors.Wrapf(err, "pushing %s to %s/%s", userForkName, releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo)
	}

	// Create a PR against k-sigs/release-notes using the github API
	gh := github.New()

	logrus.Debugf(
		"PR params: org: %s, repo: %s, headBranch: %s baseBranch: %s",
		defaultKubernetesSigsOrg, defaultKubernetesSigsRepo, git.DefaultBranch,
		fmt.Sprintf("%s:%s", releaseNotesOpts.githubOrg, branchname),
	)

	pr, err := gh.CreatePullRequest(
		defaultKubernetesSigsOrg, defaultKubernetesSigsRepo, git.DefaultBranch,
		fmt.Sprintf("%s:%s", releaseNotesOpts.githubOrg, branchname),
		fmt.Sprintf("Patch relnotes.k8s.io to release %s", tag),
		fmt.Sprintf("Automated patch to update relnotes.k8s.io to k/k version `%s` ", tag),
	)
	if err != nil {
		logrus.Warnf("An error has occurred while creating the pull request for %s", tag)
		logrus.Warn("While the PR failed, the release notes where generated and submitted to your fork")
		return errors.Wrap(err, "creating the pull request")
	}

	logrus.Infof(
		"Successfully created PR: %s%s/%s/pull/%d",
		github.GitHubURL, defaultKubernetesSigsOrg,
		defaultKubernetesSigsRepo, pr.GetNumber(),
	)
	return nil
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

	tagVersion, err := util.TagStringToSemver(tag)
	if err != nil {
		return "", errors.Wrap(err, "parsing semver from tag string")
	}

	branchName := git.DefaultBranch
	releaseBranch := fmt.Sprintf("release-%d.%d", tagVersion.Major, tagVersion.Minor)

	// Ensure we have a valid branch
	if !git.IsReleaseBranch(branchName) {
		return "", errors.New("Could not determine a release branch for tag")
	}

	// Preclone the repo to be able to read branches and tags
	logrus.Infof("Cloning %s/%s", git.DefaultGithubOrg, git.DefaultGithubRepo)
	repo, err := git.CloneOrOpenDefaultGitHubRepoSSH(rootOpts.repoPath)
	if err != nil {
		return "", errors.Wrap(err, "cloning default github repo")
	}

	// Chech if release branch already exists
	_, err = repo.RevParse(releaseBranch)
	if err == nil {
		logrus.Infof("Working on branch %s instead of %s", releaseBranch, git.DefaultBranch)
		branchName = releaseBranch
	} else {
		logrus.Infof("Release branch %s does not exist, working on %s", releaseBranch, git.DefaultBranch)
	}

	// If it's a patch release, we deduce the startTag manually:
	var startTag string
	if tagVersion.Patch > 0 {
		logrus.Debugf("Working on branch %s instead of %s", tag, git.DefaultBranch)
		startTag = fmt.Sprintf("v%d.%d.%d", tagVersion.Major, tagVersion.Minor, tagVersion.Patch-1)
	} else {
		startTag, err = repo.PreviousTag(tag, branchName)
		if err != nil {
			return "", errors.Wrap(err, "getting previous tag from branch")
		}
	}

	logrus.Infof("Using start tag %v", startTag)
	logrus.Infof("Using end tag %v", tag)

	notesOptions := options.New()
	notesOptions.Branch = branchName
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.StartRev = startTag
	notesOptions.EndRev = tag
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOptions.MapProviderStrings = releaseNotesOpts.mapProviders

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return "", err
	}

	// Fetch the notes
	releaseNotes, err := notes.GatherReleaseNotes(notesOptions)
	if err != nil {
		return "", errors.Wrapf(err, "gathering release notes")
	}

	doc, err := document.New(
		releaseNotes, notesOptions.StartRev, notesOptions.EndRev,
	)
	if err != nil {
		return "", errors.Wrapf(err, "creating release note document")
	}
	doc.PreviousRevision = startTag
	doc.CurrentRevision = tag

	// Create the JSON
	j, err := json.Marshal(releaseNotes.ByPR())
	if err != nil {
		return "", errors.Wrapf(err, "generating release notes JSON")
	}

	return string(j), nil
}

// gatherNotesFrom gathers all the release notes from the specified startTag up to --tag
func gatherNotesFrom(startTag string) (*notes.ReleaseNotes, error) {
	logrus.Infof("Gathering release notes from %s to %s", startTag, releaseNotesOpts.tag)

	notesOptions := options.New()
	notesOptions.Branch = git.DefaultBranch
	notesOptions.RepoPath = rootOpts.repoPath
	notesOptions.StartRev = startTag
	notesOptions.EndRev = releaseNotesOpts.tag
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOptions.MapProviderStrings = releaseNotesOpts.mapProviders

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return nil, err
	}

	logrus.Infof("Using start tag %v", startTag)
	logrus.Infof("Using end tag %v", releaseNotesOpts.tag)

	// Fetch the notes
	releaseNotes, err := notes.GatherReleaseNotes(notesOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "gathering release notes")
	}

	return releaseNotes, nil
}

func buildNotesResult(startTag string, releaseNotes *notes.ReleaseNotes) (*releaseNotesResult, error) {
	doc, err := document.New(
		releaseNotes, startTag, releaseNotesOpts.tag,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "creating release note document")
	}
	doc.PreviousRevision = startTag
	doc.CurrentRevision = releaseNotesOpts.tag

	// Create the markdown
	markdown, err := doc.RenderMarkdownTemplate(
		"", "", options.GoTemplateDefault,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err, "rendering release notes to markdown",
		)
	}

	// Add the dependency report if necessary
	if releaseNotesOpts.dependencies {
		logrus.Info("Generating dependency changes")
		deps, err := notes.NewDependencies().Changes(
			startTag, releaseNotesOpts.tag,
		)
		if err != nil {
			return nil, errors.Wrap(err, "creating dependency report")
		}
		markdown += strings.Repeat(nl, 2) + deps
	}

	// Create the JSON
	j, err := json.Marshal(releaseNotes.ByPR())
	if err != nil {
		return nil, errors.Wrapf(err, "generating release notes JSON")
	}

	return &releaseNotesResult{markdown: markdown, json: string(j)}, nil
}

// Validate checks if passed cmdline options are sane
func (o *releaseNotesOptions) Validate() error {
	// Check that we have a GitHub token set
	token, isset := os.LookupEnv(github.TokenEnvKey)
	if !isset || token == "" {
		return errors.New("Cannot generate release notes if GitHub token is not set")
	}

	// If a tag is defined, see if it is a valid semver tag
	_, err := util.TagStringToSemver(releaseNotesOpts.tag)
	if err != nil {
		return errors.Wrapf(err, "reading tag: %s", releaseNotesOpts.tag)
	}

	// Options for PR creation
	if o.createDraftPR || o.createWebsitePR {
		if o.userFork == "" {
			return errors.New("cannot generate the Release Notes PR without --fork")
		}
	}

	return nil
}

// Save the session to a file
func (sd *sessionData) Save() error {
	if sd.Date == 0 {
		return errors.New("unable to save session, date is note defined")
	}
	if sd.Path == "" {
		return errors.New("unable to save session, path is not defined")
	}

	jsonData, err := json.Marshal(sd)
	if err != nil {
		return errors.Wrap(err, "marshaling session data")
	}

	if err := ioutil.WriteFile(
		filepath.Join(sd.Path, fmt.Sprintf("maps-%d.json", sd.Date)),
		jsonData, os.FileMode(0o644)); err != nil {
		return errors.Wrap(err, "writing session data to disk")
	}
	return nil
}

// readFixSessions reads all the previous fixing data
func readFixSessions(sessionPath string) (pullRequestChecklist map[int]string, err error) {
	files, err := ioutil.ReadDir(sessionPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading working directory")
	}
	pullRequestList := make([]struct {
		Number int    `json:"nr"`
		Hash   string `json:"hash"`
	}, 0)

	// Look in the work dir for all json files
	for _, fileData := range files {
		currentSession := &sessionData{}
		if strings.HasSuffix(fileData.Name(), ".json") {
			logrus.Debugf("Reading session data from %s", fileData.Name())
			jsonData, err := ioutil.ReadFile(filepath.Join(sessionPath, fileData.Name()))
			if err != nil {
				return nil, errors.Wrapf(err, "reading session data from %s", fileData.Name())
			}
			if err := json.Unmarshal(jsonData, currentSession); err != nil {
				return nil, errors.Wrapf(err, "unmarshalling session data in %s", fileData.Name())
			}
			pullRequestList = append(pullRequestList, currentSession.PullRequests...)
		}
	}

	// Copy the PRs to a map for easy lookup
	pullRequestChecklist = map[int]string{}
	for _, pr := range pullRequestList {
		pullRequestChecklist[pr.Number] = pr.Hash
	}
	logrus.Infof("Read %d PR reviews from previous sessions", len(pullRequestList))
	return pullRequestChecklist, nil
}

// Do the fix process for the current tag
func fixReleaseNotes(workDir string, releaseNotes *notes.ReleaseNotes) error {
	// Get data to record the session
	userEmail, err := git.GetUserEmail()
	if err != nil {
		return errors.Wrap(err, "getting local user's email")
	}
	userName, err := git.GetUserName()
	if err != nil {
		return errors.Wrap(err, "getting local user's name")
	}

	// Check the workDir before going further
	if !util.Exists(workDir) {
		return errors.New("map directory does not exist")
	}

	// Create the new session struct
	session := &sessionData{
		UserEmail: userEmail,
		UserName:  userName,
		Date:      time.Now().UTC().Unix(),
		Path:      filepath.Join(workDir, mapsSessionDirectory),
	}

	// Read the list of all PRs we've processed so far
	pullRequestChecklist, err := readFixSessions(filepath.Join(workDir, mapsSessionDirectory))
	if err != nil {
		return errors.Wrapf(err, "reading previous session data")
	}

	// Greet the user with basic instructions
	greetingMessage := "\nWelcome to the Kubernetes Release Notes editing tool!\n\n"
	greetingMessage += "This tool will allow you to review and edit all the release\n"
	greetingMessage += "notes submitted by the Kubernetes contributors before publishing\n"
	greetingMessage += "the updated draft.\n\n"
	greetingMessage += "The flow will show each of the release notes that need to be\n"
	greetingMessage += "reviewed once and you can choose to edit it or not.\n\n"
	greetingMessage += "After you choose, it will be marked as reviewed and will not\n"
	greetingMessage += "be shown during the next sessions unless you choose to do a\n"
	greetingMessage += "full review of all notes.\n\n"
	greetingMessage += "You can hit Ctrl+C at any time to exit the review process\n"
	greetingMessage += "and submit the draft PR with the revisions made so far.\n\n"
	fmt.Print(greetingMessage)

	// Ask the user if they want to continue the last session o fix all notes
	continueFromLastSession := true
	if len(pullRequestChecklist) > 0 {
		_, continueFromLastSession, err = util.Ask("Would you like to continue from the last session? (Y/n)", "y:yes|n:no|y", 10)
	} else {
		_, _, err = util.Ask("Press enter to start editing", "y:yes|n:no|y", 10)
	}
	if err != nil {
		return errors.Wrap(err, "asking to retrieve last session")
	}

	// Bring up the provider
	provider, err := notes.NewProviderFromInitString(workDir)
	if err != nil {
		return errors.Wrap(err, "while getting map provider for current notes")
	}

	const (
		spacer = "    │ "
	)

	// Cycle all gathered release notes
	for pr, note := range releaseNotes.ByPR() {
		contentHash, err := note.ContentHash()
		if err != nil {
			return errors.Wrapf(err, "getting the content hash for PR#%d", pr)
		}
		// We'll skip editing if the Releas Note has been reviewed
		if _, ok := pullRequestChecklist[pr]; ok &&
			// and if we chose not to edit all
			continueFromLastSession &&
			// and if the not has not been modified in GutHub
			contentHash == pullRequestChecklist[pr] {
			logrus.Debugf("Pull Request %d already reviewed", pr)
			continue
		}
		title := fmt.Sprintf("Release Note for PR %d:", pr)
		fmt.Println(nl + title)
		fmt.Println(strings.Repeat("=", len(title)))
		fmt.Printf("Pull Request URL: %skubernetes/kubernetes/pull/%d%s", github.GitHubURL, pr, nl)
		noteMaps, err := provider.GetMapsForPR(pr)
		if err != nil {
			return errors.Wrapf(err, "while getting map for PR #%d", pr)
		}

		// Capture the original note values to compare
		originalNote := &notes.ReleaseNote{
			Text:           note.Text,
			Author:         note.Author,
			Areas:          note.Areas,
			Kinds:          note.Kinds,
			SIGs:           note.SIGs,
			Feature:        note.Feature,
			ActionRequired: note.ActionRequired,
			Documentation:  note.Documentation,
		}

		if noteMaps != nil {
			fmt.Println("✨ Note contents are modified with a map")
			for _, noteMap := range noteMaps {
				if err := note.ApplyMap(noteMap); err != nil {
					return errors.Wrapf(err, "applying notemap for PR #%d", pr)
				}
			}
		}

		fmt.Println(pointIfChanged("Author", note.Author, originalNote.Author), "@"+note.Author)
		fmt.Println(pointIfChanged("SIGs", note.SIGs, originalNote.SIGs), note.SIGs)
		fmt.Println(pointIfChanged("Kinds", note.Kinds, originalNote.Kinds), note.Kinds)
		fmt.Println(pointIfChanged("Areas", note.Areas, originalNote.Areas), note.Areas)
		fmt.Println(pointIfChanged("Feature", note.Feature, originalNote.Feature), note.Feature)
		fmt.Println(pointIfChanged("ActionRequired", note.ActionRequired, originalNote.ActionRequired), note.ActionRequired)
		// TODO: Implement note.Documentation

		// Wrap the note for better readability on the terminal
		fmt.Println(pointIfChanged("Text", note.Text, originalNote.Text))
		text := util.WrapText(note.Text, 80)
		fmt.Println(spacer + strings.ReplaceAll(text, nl, nl+spacer))

		_, choice, err := util.Ask(fmt.Sprintf("\n- Fix note for PR #%d? (y/N)", note.PrNumber), "y:yes|n:no|n", 10)
		if err != nil {
			// If the user cancelled with ctr+c exit and continue the PR flow
			if err.(util.UserInputError).IsCtrlC() {
				logrus.Info("Input cancelled, exiting edit flow >> PRESS ENTER TO CONTINUE")
				return nil
			}
			return errors.Wrap(err, "while asking to edit release note")
		}
		if choice {
			for {
				retry, err := editReleaseNote(pr, workDir, originalNote, note)
				if err == nil {
					break
				}
				// If it's a user error (like yaml error) we can try again
				if retry {
					_, retryEditingChoice, err := util.Ask(
						fmt.Sprintf("\n- An error occurred while editing PR #%d. Try again?", note.PrNumber),
						"y:yes|n:no", 10,
					)
					if err != nil {
						return errors.Wrap(err, "while asking to re-edit release note")
					}
					if !retryEditingChoice {
						return errors.Wrap(err, "editing release note map")
					}
				} else {
					return errors.Wrap(err, "while editing release note")
				}
			}
		}
		// Add this PR to the checklist:
		pullRequestChecklist[note.PrNumber] = contentHash
		session.PullRequests = append(session.PullRequests, struct {
			Number int    `json:"nr"`
			Hash   string `json:"hash"`
		}{
			Number: note.PrNumber,
			Hash:   contentHash,
		})
		if err := session.Save(); err != nil {
			return errors.Wrap(err, "while saving editing session data")
		}
	}
	return nil
}

// Check two values and print a prefix if they are different
func pointIfChanged(label string, var1, var2 interface{}) string {
	changed := false
	// Check if alues are string
	if _, ok := var1.(string); ok {
		if var1.(string) != var2.(string) {
			changed = true
		}
	}

	// Check if string slices
	if _, ok := var1.([]string); ok {
		if fmt.Sprint(var1) != fmt.Sprint(var2) {
			changed = true
		}
	}

	// Check if string slices
	if _, ok := var1.(bool); ok {
		if var1.(bool) != var2.(bool) {
			changed = true
		}
	}

	if changed {
		return fmt.Sprintf(" >> %s:", label)
	}
	return fmt.Sprintf("    %s:", label)
}

// editReleaseNote opens the user's editor for them to update the note.
//   In case of an editing error by the user, it returns shouldRetryEditing
//   set to true to retry editing.
func editReleaseNote(pr int, workDir string, originalNote, modifiedNote *notes.ReleaseNote) (shouldRetryEditing bool, err error) {
	// To edit the note, we will create a yaml file, with the changed fields
	// active and we'll add the unaltered fields commented for the user to review

	modifiedFields := &notes.ReleaseNotesMap{PR: pr}
	unalteredFields := &notes.ReleaseNotesMap{PR: pr}
	numChanges := 0

	if originalNote.Text == modifiedNote.Text {
		unalteredFields.ReleaseNote.Text = &originalNote.Text
	} else {
		modifiedFields.ReleaseNote.Text = &modifiedNote.Text
		numChanges++
	}

	if originalNote.Author == modifiedNote.Author {
		unalteredFields.ReleaseNote.Author = &originalNote.Author
	} else {
		modifiedFields.ReleaseNote.Author = &modifiedNote.Author
		numChanges++
	}

	if fmt.Sprint(originalNote.SIGs) == fmt.Sprint(modifiedNote.SIGs) {
		unalteredFields.ReleaseNote.SIGs = &originalNote.SIGs
	} else {
		modifiedFields.ReleaseNote.SIGs = &modifiedNote.SIGs
		numChanges++
	}

	if fmt.Sprint(originalNote.Kinds) == fmt.Sprint(modifiedNote.Kinds) {
		unalteredFields.ReleaseNote.Kinds = &originalNote.Kinds
	} else {
		modifiedFields.ReleaseNote.Kinds = &modifiedNote.Kinds
		numChanges++
	}

	if fmt.Sprint(originalNote.Areas) == fmt.Sprint(modifiedNote.Areas) {
		unalteredFields.ReleaseNote.Areas = &originalNote.Areas
	} else {
		modifiedFields.ReleaseNote.Areas = &modifiedNote.Areas
		numChanges++
	}

	if fmt.Sprint(originalNote.Feature) == fmt.Sprint(modifiedNote.Feature) {
		unalteredFields.ReleaseNote.Feature = &originalNote.Feature
	} else {
		modifiedFields.ReleaseNote.Feature = &modifiedNote.Feature
		numChanges++
	}

	if fmt.Sprint(originalNote.ActionRequired) == fmt.Sprint(modifiedNote.ActionRequired) {
		unalteredFields.ReleaseNote.ActionRequired = &originalNote.ActionRequired
	} else {
		modifiedFields.ReleaseNote.ActionRequired = &modifiedNote.ActionRequired
		numChanges++
	}

	// TODO: Implement after writing a documentation comparison func
	unalteredFields.ReleaseNote.Documentation = &originalNote.Documentation

	// Create the release note map for the editor:
	output := "---\n" + string(mapEditingInstructions) + "\n"

	if numChanges == 0 {
		// If there are no changes, present the user with the commented
		// map with the original values
		yamlCode, err := yaml.Marshal(&unalteredFields)
		if err != nil {
			return false, errors.Wrap(err, "marshalling release note to map")
		}
		output += "# " + strings.ReplaceAll(string(yamlCode), "\n", "\n# ")
	} else {
		// ... otherwise build a mixed map with the changes and the original
		// values commented out for reference
		yamlCode, err := yaml.Marshal(&modifiedFields)
		if err != nil {
			return false, errors.Wrap(err, "marshalling release note to map")
		}

		unalteredYAML, err := yaml.Marshal(&unalteredFields.ReleaseNote)
		if err != nil {
			return false, errors.Wrap(err, "marshalling release note to map")
		}
		output += string(yamlCode) + " # " + strings.ReplaceAll(string(unalteredYAML), "\n", "\n # ")
	}

	kubeEditor := editor.NewDefaultEditor([]string{"KUBE_EDITOR", "EDITOR"})
	changes, _, err := kubeEditor.LaunchTempFile("map", ".yaml", bytes.NewReader([]byte(output)))
	if err != nil {
		return false, errors.Wrap(err, "while launching editor")
	}

	// If the map was not modified, we don't make any changes
	if string(changes) == output || string(changes) == "" {
		logrus.Info("Release notes map was not modified")
		return false, nil
	}

	// Verify that the new yaml is valid and can be serialized back into a Map
	testMap := notes.ReleaseNotesMap{}
	err = yaml.Unmarshal(changes, &testMap)

	if err != nil {
		logrus.Error("The YAML code has errors")
		return true, errors.Wrap(err, "while verifying if changes are a valid map")
	}

	if testMap.PR == 0 {
		logrus.Error("The yaml code does not have a PR number")
		return true, errors.New("Invalid map: the YAML code did not have a PR number")
	}

	// Remarshall the newyaml to save only the new values
	newYAML, err := yaml.Marshal(testMap)
	if err != nil {
		return true, errors.Wrap(err, "while re-marshaling new map")
	}

	// Write the new map, removing the instructions
	mapPath := filepath.Join(workDir, mapsMainDirectory, fmt.Sprintf("pr-%d-map.yaml", pr))
	err = ioutil.WriteFile(mapPath, newYAML, os.FileMode(0o644))
	if err != nil {
		logrus.Errorf("Error writing map to %s: %s", mapPath, err)
		return true, errors.Wrap(err, "writing modified release note map")
	}

	return false, nil
}

// createNotesWorkDir creates the release notes working directory
func createNotesWorkDir(releaseDir string) error {
	// Check that the working tree is complete:
	for _, dirPath := range []string{
		filepath.Join(releaseDir, releaseNotesWorkDir),                       // Main work dir
		filepath.Join(releaseDir, releaseNotesWorkDir, mapsMainDirectory),    // Maps directory
		filepath.Join(releaseDir, releaseNotesWorkDir, mapsCVEDirectory),     // Maps for CVE data
		filepath.Join(releaseDir, releaseNotesWorkDir, mapsSessionDirectory), // Editing session files
		filepath.Join(releaseDir, releaseNotesWorkDir, mapsThemesDirectory),  // Major themes directory
	} {
		if !util.Exists(dirPath) {
			if err := os.Mkdir(dirPath, os.FileMode(0o755)); err != nil {
				return errors.Wrap(err, "creating working directory")
			}
		}
	}
	return nil
}

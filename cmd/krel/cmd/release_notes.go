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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	gogit "github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-sdk/github"
	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/editor"
	"sigs.k8s.io/release-utils/util"
	"sigs.k8s.io/yaml"

	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/notes/document"
	"k8s.io/release/pkg/notes/options"
)

const (
	// draftMarkdownFile filename for the release notes draft.
	draftMarkdownFile = "release-notes-draft.md"
	// draftJSONFile is the json version of the release notes.
	draftJSONFile = "release-notes-draft.json"
	// serviceDirectory is where we keep the files used to generate the notes.
	releaseNotesWorkDir = "release-notes"
	// mapsMainDirectory is where we will save the release notes maps.
	mapsMainDirectory = "maps"
	// mapsCVEDirectory holds the maps with CVE data.
	mapsCVEDirectory = "cve"
	// Directory where session editing files are stored.
	mapsSessionDirectory = "sessions"
	// The themes directory holds the maps for the release notes major themes.
	mapsThemesDirectory = "themes"
	// defaultKubernetesSigsOrg GitHub org owner of the release-notes repo.
	defaultKubernetesSigsOrg = "kubernetes-sigs"
	// defaultKubernetesSigsRepo relnotes.k8s.io repository name.
	defaultKubernetesSigsRepo = "release-notes"
	// userForkName The name we will give to the user's remote when adding it to repos.
	userForkName = github.UserForkName
	// assetsFilePath Path to the assets.ts file.
	assetsFilePath = "src/environments/assets.ts"
	// websiteBranchPrefix name of the website branch in the user's fork.
	websiteBranchPrefix = "release-notes-json-"
	// draftBranchPrefix name of the draft branch in the user's fork.
	draftBranchPrefix = "release-notes-draft-"
	// Editing instructions for notemaps.
	mapEditingInstructions = `# This is the current map for this Pull Request.
# The original note content is commented out, if you need to
# change a field, remove the comment and change the value.
# To cancel, exit without changing anything or leave the file blank.
# Important! pr: and releasenote: have to be uncommented.
#`
)

// releaseNotesCmd represents the subcommand for `krel release-notes`.
var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "The subcommand of choice for the Release Notes subteam of SIG Release",
	Long: fmt.Sprintf(`krel release-notes

The 'release-notes' subcommand of krel has been developed to:

1. Generate the release notes draft for the provided tag for commits on the main
   branch.

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
	createDraftPR      bool
	createWebsitePR    bool
	fixNotes           bool
	listReleaseNotesV2 bool
	interactiveMode    bool
	updateRepo         bool
	useSSH             bool
	repoPath           string
	tag                string
	userFork           string
	websiteRepo        string
	githubOrg          string
	draftRepo          string
	mapProviders       []string
}

type releaseNotesResult struct {
	markdown string
	json     string
}

// A datatype to record a notes repair session.
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
	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.repoPath,
		"repo",
		filepath.Join(os.TempDir(), "k8s"),
		"the local path to the repository to be used",
	)

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
		"[DEPRECATED] patch the relnotes.k8s.io sources and generate a PR with the changes",
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

	releaseNotesCmd.PersistentFlags().BoolVar(
		&releaseNotesOpts.listReleaseNotesV2,
		"list-v2",
		false,
		"enable experimental implementation to list commits (ListReleaseNotesV2)",
	)

	releaseNotesCmd.PersistentFlags().BoolVar(
		&releaseNotesOpts.interactiveMode,
		"interactiveMode",
		true,
		"interactive mode, ask before creating the PR",
	)

	releaseNotesCmd.PersistentFlags().BoolVarP(
		&releaseNotesOpts.useSSH,
		"use-ssh",
		"",
		true,
		"use ssh to clone the repository, if false will use https (default: true)",
	)

	releaseNotesCmd.PersistentFlags().BoolVarP(
		&releaseNotesOpts.updateRepo,
		"update-repo",
		"",
		true,
		"update the cloned repository to fetch any upstream change (default: true)",
	)

	rootCmd.AddCommand(releaseNotesCmd)
}

func runReleaseNotes() (err error) {
	var tag string
	if releaseNotesOpts.tag == "" {
		tag, err = tryToFindLatestMinorTag()
		if err != nil {
			return fmt.Errorf("unable to find latest minor tag: %w", err)
		}

		releaseNotesOpts.tag = tag
	} else {
		tag = releaseNotesOpts.tag
	}

	if releaseNotesOpts.userFork != "" {
		org, repo, err := git.ParseRepoSlug(releaseNotesOpts.userFork)
		if err != nil {
			return fmt.Errorf("parsing the user's fork: %w", err)
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
		return fmt.Errorf("validating command line options: %w", err)
	}

	// before running the generators, verify that the repositories are ready
	if releaseNotesOpts.createWebsitePR {
		if err := github.VerifyFork(
			websiteBranchPrefix+tag,
			releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo,
			defaultKubernetesSigsOrg, defaultKubernetesSigsRepo,
		); err != nil {
			return fmt.Errorf("while checking %s/%s fork: %w", defaultKubernetesSigsOrg, defaultKubernetesSigsRepo, err)
		}
	}

	if releaseNotesOpts.createDraftPR {
		if err := github.VerifyFork(
			draftBranchPrefix+tag,
			releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo,
			git.DefaultGithubOrg, git.DefaultGithubReleaseRepo,
		); err != nil {
			return fmt.Errorf("while checking %s/%s fork: %w", git.DefaultGithubOrg, git.DefaultGithubReleaseRepo, err)
		}
	}

	// Create the PR for relnotes.k8s.io
	if releaseNotesOpts.createWebsitePR && confirmWithUser(releaseNotesOpts, "Create website pull request?") {
		// Run the website PR process
		if err := createWebsitePR(releaseNotesOpts.repoPath, tag); err != nil {
			return fmt.Errorf("creating website PR: %w", err)
		}
	}

	// Create the PR for the Release Notes Draft in k/sig-release
	if releaseNotesOpts.createDraftPR && confirmWithUser(releaseNotesOpts, "Create draft pull request?") {
		// Create the Draft PR Process
		if err := createDraftPR(releaseNotesOpts.repoPath, tag); err != nil {
			return fmt.Errorf("creating Draft PR: %w", err)
		}
	}

	if releaseNotesOpts.createDraftPR || releaseNotesOpts.createWebsitePR {
		logrus.Info("Release notes generation complete!")
	}

	return nil
}

// createDraftPR pushes the release notes draft to the users fork.
func createDraftPR(repoPath, tag string) (err error) {
	tagVersion, err := util.TagStringToSemver(tag)
	if err != nil {
		return fmt.Errorf("reading tag: %s: %w", tag, err)
	}

	// From v1.20.0 on we use the previous minor as a starting tag
	// for the Release Notes draft because the branch is fast-rowarded now:
	start := util.SemverToTagString(semver.Version{
		Major: tagVersion.Major,
		Minor: tagVersion.Minor - 1,
		Patch: 0,
	})

	gh := github.New()

	autoCreatePullRequest := true

	// Verify the repository
	isrepo, err := gh.RepoIsForkOf(
		releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo,
		git.DefaultGithubOrg, git.DefaultGithubReleaseRepo,
	)
	if err != nil {
		return fmt.Errorf(
			"while checking if repository is a fork of %s/%s: %w",
			git.DefaultGithubOrg, git.DefaultGithubReleaseRepo, err,
		)
	}

	if !isrepo {
		return fmt.Errorf(
			"cannot create PR, %s/%s is not a fork of %s/%s",
			releaseNotesOpts.githubOrg,
			releaseNotesOpts.draftRepo,
			git.DefaultGithubOrg,
			git.DefaultGithubReleaseRepo,
		)
	}

	// Generate the notes for the current version
	releaseNotes, err := gatherNotesFrom(repoPath, start)
	if err != nil {
		return fmt.Errorf("while generating the release notes for tag %s: %w", start, err)
	}

	branchname := draftBranchPrefix + tag

	// Prepare the fork of k/sig-release
	opts := &gogit.CloneOptions{}

	sigReleaseRepo, err := github.PrepareFork(
		branchname,
		git.DefaultGithubOrg, git.DefaultGithubReleaseRepo,
		releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo,
		releaseNotesOpts.useSSH, releaseNotesOpts.updateRepo,
		opts,
	)
	if err != nil {
		return fmt.Errorf("preparing local fork of kubernetes/sig-release: %w", err)
	}

	// The release path inside the repository
	releasePath := filepath.Join("releases", fmt.Sprintf("release-%d.%d", tagVersion.Major, tagVersion.Minor))

	// Check if the directory exists
	releaseDir := filepath.Join(sigReleaseRepo.Dir(), releasePath)
	if !util.Exists(releaseDir) {
		return fmt.Errorf("could not find release directory %s", releaseDir)
	}

	// If we got the --fix flag, start the fix flow
	if releaseNotesOpts.fixNotes {
		_, _, err = util.Ask("Press enter to start", "y:yes|n:no|y", 10)
		// In interactive mode, we will ask the user before sending the PR
		autoCreatePullRequest = false

		// createNotesWorkDir is idempotent, we can use it to verify the tree is complete
		if err := createNotesWorkDir(releaseDir); err != nil {
			return fmt.Errorf("creating working directory: %w", err)
		}

		// Run the release notes fix flow
		err := fixReleaseNotes(filepath.Join(releaseDir, releaseNotesWorkDir), releaseNotes)
		if err != nil {
			return fmt.Errorf("while running release notes fix flow: %w", err)
		}

		// Create the map provider to read the changes so far
		rnMapProvider, err := notes.NewProviderFromInitString(filepath.Join(releaseDir, releaseNotesWorkDir, mapsMainDirectory))
		if err != nil {
			return fmt.Errorf("creating release notes draft: %w", err)
		}

		for _, note := range releaseNotes.ByPR() {
			maps, err := rnMapProvider.GetMapsForPR(note.PrNumber)
			if err != nil {
				return fmt.Errorf("while getting maps for PR #%d: %w", note.PrNumber, err)
			}

			for _, noteMap := range maps {
				if err := note.ApplyMap(noteMap, true); err != nil {
					return fmt.Errorf("applying note maps to PR #%d: %w", note.PrNumber, err)
				}
			}
		}
	}

	// Generate the results struct
	result, err := buildNotesResult(start, releaseNotes)
	if err != nil {
		return fmt.Errorf("building release notes results: %w", err)
	}

	// generate the notes files
	logrus.Debugf("Release notes draft files will be written to %s", releaseDir)

	// Write the markdown draft
	//nolint:gosec // TODO(gosec): G306: Expect WriteFile permissions to be
	// 0600 or less
	err = os.WriteFile(filepath.Join(releaseDir, releaseNotesWorkDir, draftMarkdownFile), []byte(result.markdown), 0o644)
	if err != nil {
		return fmt.Errorf("writing release notes draft: %w", err)
	}

	logrus.Infof("Release Notes Markdown Draft written to %s", filepath.Join(releaseDir, releaseNotesWorkDir, draftMarkdownFile))

	// Write the JSON file of the current notes
	//nolint:gosec // TODO(gosec): G306: Expect WriteFile permissions to be
	// 0600 or less
	err = os.WriteFile(filepath.Join(releaseDir, releaseNotesWorkDir, draftJSONFile), []byte(result.json), 0o644)
	if err != nil {
		return fmt.Errorf("writing release notes json file: %w", err)
	}

	logrus.Infof("Release Notes JSON version written to %s", filepath.Join(releaseDir, releaseNotesWorkDir, draftJSONFile))

	// If we are in interactive mode, ask before continuing
	if !autoCreatePullRequest {
		_, autoCreatePullRequest, err = util.Ask("Create pull request with your changes? (y/n)", "y:Y:yes|n:N:no|y", 10)
		if err != nil {
			return fmt.Errorf("while asking to create pull request: %w", err)
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
		fmt.Print(nl + nl)
		logrus.Warn("Changes were made locally, user needs to perform manual push and create pull request.")

		return nil
	}

	defer func() {
		if e := sigReleaseRepo.Cleanup(); e != nil {
			err = fmt.Errorf("cleaning temporary sig release clone: %w", e)
		}
	}()

	// Create the commit
	if err := createDraftCommit(
		sigReleaseRepo, releasePath, "Release Notes draft for k/k "+tag,
	); err != nil {
		return fmt.Errorf("creating release notes commit: %w", err)
	}

	// push to the user's remote
	logrus.Infof("Pushing modified release notes draft to %s/%s", releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo)

	if err := sigReleaseRepo.PushToRemote(userForkName, branchname); err != nil {
		return fmt.Errorf("pushing %s to remote: %w", userForkName, err)
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
		"Update release notes draft to version "+tag, prBody, false,
	)
	if err != nil {
		logrus.Warnf("An error has occurred while creating the pull request for %s", tag)
		logrus.Warn("While the PR failed, the release notes draft was generated and submitted to your fork")

		return fmt.Errorf("creating the pull request: %w", err)
	}

	logrus.Infof(
		"Successfully created PR: %s%s/%s/pull/%d",
		github.GitHubURL, git.DefaultGithubOrg,
		git.DefaultGithubReleaseRepo, pr.GetNumber(),
	)
	logrus.Infof("Successfully created PR #%d", pr.GetNumber())

	return err
}

// createDraftCommit creates the release notes commit in the temporary clone
// of the user's sig-release fork. It will include both versions of the draft,
// the maps and cve directories and the edit session files.
func createDraftCommit(repo *git.Repo, releasePath, commitMessage string) error {
	// add the updated draft
	if err := repo.Add(filepath.Join(releasePath, releaseNotesWorkDir, draftMarkdownFile)); err != nil {
		return fmt.Errorf("adding release notes draft to staging area: %w", err)
	}

	// add the json draft
	if err := repo.Add(filepath.Join(releasePath, releaseNotesWorkDir, draftJSONFile)); err != nil {
		return fmt.Errorf("adding release notes json to staging area: %w", err)
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
		if util.Exists(filepath.Join(repo.Dir(), dirData.Path)) {
			// Check if there are any files to commit
			matches, err := filepath.Glob(filepath.Join(repo.Dir(), dirData.Path, "*"+dirData.Ext))
			logrus.Debugf("Adding %d %s from %s to commit", len(matches), dirData.Name, dirData.Path)

			if err != nil {
				return fmt.Errorf("checking for %s files in %s: %w", dirData.Ext, dirData.Path, err)
			}

			if len(matches) > 0 {
				if err := repo.Add(filepath.Join(dirData.Path, "*"+dirData.Ext)); err != nil {
					return fmt.Errorf("adding %s to staging area: %w", dirData.Name, err)
				}
			}
		} else {
			logrus.Debugf("Not adding %s files, directory %s not found", dirData.Name, dirData.Path)
		}
	}

	// add the generated draft
	if err := repo.UserCommit(commitMessage); err != nil {
		return fmt.Errorf("creating commit in %s/%s: %w", releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo, err)
	}

	return nil
}

// addReferenceToAssetsFile adds a new entry in the assets.ts file in repoPath to include newJsonFile.
func addReferenceToAssetsFile(repoPath, newJSONFile string) error {
	// Full  filesystem path to the assets.ts file
	assetsFullPath := filepath.Join(repoPath, assetsFilePath)

	file, err := os.Open(assetsFullPath)
	if err != nil {
		return fmt.Errorf("opening assets.ts to check for current version: %w", err)
	}
	defer file.Close()

	logrus.Infof("Writing json reference to %s in %s", newJSONFile, assetsFullPath)

	scanner := bufio.NewScanner(file)

	var assetsBuffer bytes.Buffer

	assetsFileWasModified := false
	fileIsReferenced := false

	for scanner.Scan() {
		// Check if the assets file already has the json notes referenced:
		if strings.Contains(scanner.Text(), "assets/"+newJSONFile) {
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
	if err := os.WriteFile(assetsFullPath, assetsBuffer.Bytes(), os.FileMode(0o644)); err != nil {
		return fmt.Errorf("writing assets.ts file: %w", err)
	}

	return nil
}

// processJSONOutput Runs NPM prettier inside repoPath to format the JSON output.
func processJSONOutput(repoPath string) error {
	npmpath, err := exec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("while looking for npm in your path: %w", err)
	}

	// run npm install
	logrus.Info("Installing npm modules, this can take a while")

	if err := command.NewWithWorkDir(repoPath, npmpath, "install").RunSuccess(); err != nil {
		return fmt.Errorf("running npm install in kubernetes-sigs/release-notes: %w", err)
	}

	// run npm prettier
	logrus.Info("Running npm prettier...")

	if err := command.NewWithWorkDir(repoPath, npmpath, "run", "prettier").RunSuccess(); err != nil {
		return fmt.Errorf("running npm prettier in kubernetes-sigs/release-notes: %w", err)
	}

	return nil
}

// createWebsitePR creates the JSON version of the release notes and pushes them to a user fork.
func createWebsitePR(repoPath, tag string) (err error) {
	_, err = util.TagStringToSemver(tag)
	if err != nil {
		return fmt.Errorf("reading tag: %s: %w", tag, err)
	}

	// Generate the release notes for ust the current tag
	jsonStr, err := releaseNotesJSON(repoPath, tag)
	if err != nil {
		return fmt.Errorf("generating release notes in JSON format: %w", err)
	}

	jsonNotesFilename := fmt.Sprintf("release-notes-%s.json", tag[1:])
	branchname := websiteBranchPrefix + tag

	// checkout kubernetes-sigs/release-notes
	opts := &gogit.CloneOptions{}

	k8sSigsRepo, err := github.PrepareFork(
		branchname,
		defaultKubernetesSigsOrg, defaultKubernetesSigsRepo,
		releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo,
		releaseNotesOpts.useSSH, releaseNotesOpts.updateRepo,
		opts,
	)
	if err != nil {
		return fmt.Errorf("preparing local fork branch: %w", err)
	}

	defer func() {
		if e := k8sSigsRepo.Cleanup(); e != nil {
			err = fmt.Errorf("cleaning up k/sigs repo: %w", e)
		}
	}()

	// add a reference to the new json file in assets.ts
	if err := addReferenceToAssetsFile(k8sSigsRepo.Dir(), jsonNotesFilename); err != nil {
		return fmt.Errorf("adding %s to assets file: %w", jsonNotesFilename, err)
	}

	// generate the notes
	jsonNotesPath := filepath.Join("src", "assets", jsonNotesFilename)
	logrus.Debugf("Release notes json file will be written to %s", filepath.Join(k8sSigsRepo.Dir(), jsonNotesPath))

	if err := os.WriteFile(
		filepath.Join(k8sSigsRepo.Dir(), jsonNotesPath), []byte(jsonStr), os.FileMode(0o644),
	); err != nil {
		return fmt.Errorf("writing release notes json file: %w", err)
	}

	// Run NPM prettier
	if err := processJSONOutput(k8sSigsRepo.Dir()); err != nil {
		return fmt.Errorf("while formatting release notes JSON files: %w", err)
	}

	// add the modified files & commit the results
	if err := k8sSigsRepo.Add(jsonNotesPath); err != nil {
		return fmt.Errorf("adding release notes draft to staging area: %w", err)
	}

	if err := k8sSigsRepo.Add(filepath.FromSlash(assetsFilePath)); err != nil {
		return fmt.Errorf("adding release notes draft to staging area: %w", err)
	}

	if err := k8sSigsRepo.UserCommit("Patch relnotes.k8s.io with release " + tag); err != nil {
		return fmt.Errorf("error creating commit in %s/%s: %w", releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo, err)
	}

	// push to the user's fork
	logrus.Infof("Pushing website changes to %s/%s", releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo)

	if err := k8sSigsRepo.PushToRemote(userForkName, branchname); err != nil {
		return fmt.Errorf("pushing %s to %s/%s: %w", userForkName, releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo, err)
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
		"Patch relnotes.k8s.io to release "+tag,
		fmt.Sprintf("Automated patch to update relnotes.k8s.io to k/k version `%s` ", tag), false,
	)
	if err != nil {
		logrus.Warnf("An error has occurred while creating the pull request for %s", tag)
		logrus.Warn("While the PR failed, the release notes where generated and submitted to your fork")

		return fmt.Errorf("creating the pull request: %w", err)
	}

	logrus.Infof(
		"Successfully created PR: %s%s/%s/pull/%d",
		github.GitHubURL, defaultKubernetesSigsOrg,
		defaultKubernetesSigsRepo, pr.GetNumber(),
	)

	return err
}

// tryToFindLatestMinorTag looks-up the default k/k remote to find the latest
// non final version.
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
// them as JSON blob.
func releaseNotesJSON(repoPath, tag string) (jsonString string, err error) {
	logrus.Infof("Generating release notes for tag %s", tag)

	tagVersion, err := util.TagStringToSemver(tag)
	if err != nil {
		return "", fmt.Errorf("parsing semver from tag string: %w", err)
	}

	logrus.Info("Cloning kubernetes/sig-release to read mapping files")

	opts := &gogit.CloneOptions{}

	sigReleaseRepo, err := git.CleanCloneGitHubRepo(
		git.DefaultGithubOrg, git.DefaultGithubReleaseRepo, false, true, opts,
	)
	if err != nil {
		return "", fmt.Errorf("performing clone of k/sig-release: %w", err)
	}

	defer func() {
		if e := sigReleaseRepo.Cleanup(); e != nil {
			err = sigReleaseRepo.Cleanup()
		}
	}()

	branchName := git.DefaultBranch
	releaseBranch := fmt.Sprintf("release-%d.%d", tagVersion.Major, tagVersion.Minor)

	// Ensure we have a valid branch
	if !git.IsReleaseBranch(branchName) {
		return "", errors.New("could not determine a release branch for tag")
	}

	// Preclone the repo to be able to read branches and tags
	logrus.Infof("Cloning %s/%s", git.DefaultGithubOrg, git.DefaultGithubRepo)

	repo, err := git.CloneOrOpenGitHubRepo(repoPath, git.DefaultGithubOrg, git.DefaultGithubRepo, false)
	if err != nil {
		return "", fmt.Errorf("cloning default github repo: %w", err)
	}

	// Check if release branch already exists
	_, err = repo.RevParseTag(releaseBranch)
	if err == nil {
		logrus.Infof("Working on branch %s instead of %s", releaseBranch, git.DefaultBranch)
		branchName = releaseBranch
	} else {
		logrus.Infof("Release branch %s does not exist, working on %s", releaseBranch, git.DefaultBranch)
	}

	// Notes for patch releases are generated starting from the previous patch release:
	var startTag, tagChoice string
	if tagVersion.Patch > 0 {
		startTag = fmt.Sprintf("v%d.%d.%d", tagVersion.Major, tagVersion.Minor, tagVersion.Patch-1)
		tagChoice = "previous patch release"
	} else {
		// From 1.20 the notes for the first alpha start from the previous minor
		if len(tagVersion.Pre) == 2 && //nolint:gocritic // a switch case would not make it better
			tagVersion.Pre[0].String() == "alpha" &&
			tagVersion.Pre[1].VersionNum == 1 {
			startTag = util.SemverToTagString(semver.Version{
				Major: tagVersion.Major, Minor: tagVersion.Minor - 1, Patch: 0,
			})
			tagChoice = "previous minor version"
		} else if len(tagVersion.Pre) == 0 && tagVersion.Patch == 0 {
			// If we are writing the notes for the first minor version (eg 1.20.0)
			// we choose as the start tag also the previous minor
			startTag = util.SemverToTagString(semver.Version{
				Major: tagVersion.Major, Minor: tagVersion.Minor - 1, Patch: 0,
			})
			tagChoice = "previous minor version because we are in a new minor version"
		} else {
			// All others from the previous existing tag
			startTag, err = repo.PreviousTag(tag, branchName)
			if err != nil {
				return "", fmt.Errorf("getting previous tag from branch: %w", err)
			}

			tagChoice = "previous tag"
		}
	}

	logrus.Infof("Using start tag %v from %s", startTag, tagChoice)
	logrus.Infof("Using end tag %v", tag)

	notesOptions := options.New()
	notesOptions.Branch = branchName
	notesOptions.RepoPath = repoPath
	notesOptions.StartRev = startTag
	notesOptions.EndRev = tag
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOptions.MapProviderStrings = releaseNotesOpts.mapProviders
	notesOptions.AddMarkdownLinks = true

	// If the release for the tag we are using has a mapping directory,
	// add it to the mapProviders array to read the edits from the release team:
	mapsDir := filepath.Join(
		sigReleaseRepo.Dir(), "releases",
		fmt.Sprintf("release-%d.%d", tagVersion.Major, tagVersion.Minor),
		releaseNotesWorkDir, mapsMainDirectory,
	)
	if util.Exists(mapsDir) {
		logrus.Infof("Notes gatherer will read maps from %s", mapsDir)
		notesOptions.MapProviderStrings = append(notesOptions.MapProviderStrings, mapsDir)
	}

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return "", err
	}

	// Fetch the notes
	releaseNotes, err := notes.GatherReleaseNotes(notesOptions)
	if err != nil {
		return "", fmt.Errorf("gathering release notes: %w", err)
	}

	doc, err := document.New(
		releaseNotes, notesOptions.StartRev, notesOptions.EndRev,
	)
	if err != nil {
		return "", fmt.Errorf("creating release note document: %w", err)
	}

	doc.PreviousRevision = startTag
	doc.CurrentRevision = tag

	// Create the JSON
	j, err := json.Marshal(releaseNotes.ByPR())
	if err != nil {
		return "", fmt.Errorf("generating release notes JSON: %w", err)
	}

	return string(j), err
}

// gatherNotesFrom gathers all the release notes from the specified startTag up to --tag.
func gatherNotesFrom(repoPath, startTag string) (*notes.ReleaseNotes, error) {
	logrus.Infof("Gathering release notes from %s to %s", startTag, releaseNotesOpts.tag)

	notesOptions := options.New()
	notesOptions.Branch = git.DefaultBranch
	notesOptions.RepoPath = repoPath
	notesOptions.StartRev = startTag
	notesOptions.EndRev = releaseNotesOpts.tag
	notesOptions.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOptions.MapProviderStrings = releaseNotesOpts.mapProviders
	notesOptions.ListReleaseNotesV2 = releaseNotesOpts.listReleaseNotesV2
	notesOptions.AddMarkdownLinks = true

	if err := notesOptions.ValidateAndFinish(); err != nil {
		return nil, err
	}

	logrus.Infof("Using start tag %v", startTag)
	logrus.Infof("Using end tag %v", releaseNotesOpts.tag)

	// Fetch the notes
	releaseNotes, err := notes.GatherReleaseNotes(notesOptions)
	if err != nil {
		return nil, fmt.Errorf("gathering release notes: %w", err)
	}

	return releaseNotes, nil
}

func buildNotesResult(startTag string, releaseNotes *notes.ReleaseNotes) (*releaseNotesResult, error) {
	doc, err := document.New(
		releaseNotes, startTag, releaseNotesOpts.tag,
	)
	if err != nil {
		return nil, fmt.Errorf("creating release note document: %w", err)
	}

	doc.PreviousRevision = startTag
	doc.CurrentRevision = releaseNotesOpts.tag

	// Create the markdown
	markdown, err := doc.RenderMarkdownTemplate(
		"", "", "", options.GoTemplateDefault,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"rendering release notes to markdown: %w", err,
		)
	}

	// Create the JSON
	j, err := json.MarshalIndent(releaseNotes.ByPR(), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("generating release notes JSON: %w", err)
	}

	return &releaseNotesResult{markdown: markdown, json: string(j)}, nil
}

// Validate checks if passed cmdline options are sane.
func (o *releaseNotesOptions) Validate() error {
	// Check that we have a GitHub token set
	token, isset := os.LookupEnv(github.TokenEnvKey)
	if !isset || token == "" {
		return fmt.Errorf("cannot generate release notes if %s env variable is not set", github.TokenEnvKey)
	}

	// If a tag is defined, see if it is a valid semver tag
	_, err := util.TagStringToSemver(releaseNotesOpts.tag)
	if err != nil {
		return fmt.Errorf("reading tag: %s: %w", releaseNotesOpts.tag, err)
	}

	// Options for PR creation
	if o.createDraftPR || o.createWebsitePR {
		if o.userFork == "" {
			return errors.New("cannot generate the Release Notes PR without --fork")
		}
	}

	return nil
}

// Save the session to a file.
func (sd *sessionData) Save() error {
	if sd.Date == 0 {
		return errors.New("unable to save session, date is note defined")
	}

	if sd.Path == "" {
		return errors.New("unable to save session, path is not defined")
	}

	jsonData, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session data: %w", err)
	}

	if err := os.WriteFile(
		filepath.Join(sd.Path, fmt.Sprintf("maps-%d.json", sd.Date)),
		jsonData, os.FileMode(0o644)); err != nil {
		return fmt.Errorf("writing session data to disk: %w", err)
	}

	return nil
}

// readFixSessions reads all the previous fixing data.
func readFixSessions(sessionPath string) (pullRequestChecklist map[int]string, err error) {
	files, err := os.ReadDir(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("reading working directory: %w", err)
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

			jsonData, err := os.ReadFile(filepath.Join(sessionPath, fileData.Name()))
			if err != nil {
				return nil, fmt.Errorf("reading session data from %s: %w", fileData.Name(), err)
			}

			if err := json.Unmarshal(jsonData, currentSession); err != nil {
				return nil, fmt.Errorf("unmarshalling session data in %s: %w", fileData.Name(), err)
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

// Do the fix process for the current tag.
func fixReleaseNotes(workDir string, releaseNotes *notes.ReleaseNotes) error {
	// Get data to record the session
	userEmail, err := git.GetUserEmail()
	if err != nil {
		return fmt.Errorf("getting local user's email: %w", err)
	}

	userName, err := git.GetUserName()
	if err != nil {
		return fmt.Errorf("getting local user's name: %w", err)
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
		return fmt.Errorf("reading previous session data: %w", err)
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
		_, continueFromLastSession, err = util.Ask("Would you like to continue from the last session? (Y/n)", "y:Y:yes|n:N:no|y", 10)
	} else {
		_, _, err = util.Ask("Press enter to start editing", "y:Y:yes|n:N:no|y", 10)
	}

	if err != nil {
		return fmt.Errorf("asking to retrieve last session: %w", err)
	}

	// Bring up the provider
	provider, err := notes.NewProviderFromInitString(workDir)
	if err != nil {
		return fmt.Errorf("while getting map provider for current notes: %w", err)
	}

	const (
		spacer = "    │ "
	)

	// Cycle all gathered release notes
	for pr, note := range releaseNotes.ByPR() {
		contentHash, err := note.ContentHash()
		if err != nil {
			return fmt.Errorf("getting the content hash for PR#%d: %w", pr, err)
		}
		// We'll skip editing if the Release Note has been reviewed
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
			return fmt.Errorf("while getting map for PR #%d: %w", pr, err)
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
			DoNotPublish:   note.DoNotPublish,
			PRBody:         note.PRBody,
		}

		if noteMaps != nil {
			fmt.Println("✨ Note contents was previously modified with a map")

			for _, noteMap := range noteMaps {
				if err := note.ApplyMap(noteMap, true); err != nil {
					return fmt.Errorf("applying notemap for PR #%d: %w", pr, err)
				}
			}
		}

		fmt.Println(pointIfChanged("Author", note.Author, originalNote.Author), "@"+note.Author)
		fmt.Println(pointIfChanged("SIGs", note.SIGs, originalNote.SIGs), note.SIGs)
		fmt.Println(pointIfChanged("Kinds", note.Kinds, originalNote.Kinds), note.Kinds)
		fmt.Println(pointIfChanged("Areas", note.Areas, originalNote.Areas), note.Areas)
		fmt.Println(pointIfChanged("Feature", note.Feature, originalNote.Feature), note.Feature)
		fmt.Println(pointIfChanged("ActionRequired", note.ActionRequired, originalNote.ActionRequired), note.ActionRequired)
		fmt.Println(pointIfChanged("DoNotPublish", note.DoNotPublish, originalNote.DoNotPublish), note.DoNotPublish)
		// TODO: Implement note.Documentation

		// Wrap the note for better readability on the terminal
		fmt.Println(pointIfChanged("Text", note.Text, originalNote.Text))
		text := util.WrapText(note.Text, 80)
		fmt.Println(spacer + strings.ReplaceAll(text, nl, nl+spacer))

		_, choice, err := util.Ask(fmt.Sprintf("\n- Fix note for PR #%d? (y/N)", note.PrNumber), "y:Y:yes|n:N:no|n", 10)
		if err != nil {
			// If the user cancelled with ctr+c exit and continue the PR flow
			var userInputErr util.UserInputError
			if errors.As(err, &userInputErr) && userInputErr.IsCtrlC() {
				logrus.Info("Input cancelled, exiting edit flow")

				return nil
			}

			return fmt.Errorf("while asking to edit release note: %w", err)
		}

		noteReviewed := true

		if choice {
			for {
				retry, err := editReleaseNote(pr, workDir, originalNote, note)
				if err == nil {
					break
				}
				// If it's a user error (like yaml error) we can try again
				if retry {
					logrus.Error(err)

					_, retryEditingChoice, err := util.Ask(
						fmt.Sprintf("\n- An error occurred while editing PR #%d. Try again?", note.PrNumber),
						"y:yes|n:no", 10,
					)
					if err != nil {
						return fmt.Errorf("while asking to re-edit release note: %w", err)
					}
					// If user chooses not to fix the faulty yaml, do not mark as fixed
					if !retryEditingChoice {
						noteReviewed = false

						break
					}
				} else {
					return fmt.Errorf("while editing release note: %w", err)
				}
			}
		}
		// If the note was reviewed, add the PR to the session file:
		if noteReviewed {
			pullRequestChecklist[note.PrNumber] = contentHash

			session.PullRequests = append(session.PullRequests, struct {
				Number int    `json:"nr"`
				Hash   string `json:"hash"`
			}{
				Number: note.PrNumber,
				Hash:   contentHash,
			})
			if err := session.Save(); err != nil {
				return fmt.Errorf("while saving editing session data: %w", err)
			}
		}
	}

	return nil
}

// Check two values and print a prefix if they are different.
func pointIfChanged(label string, var1, var2 interface{}) string {
	changed := false
	// Check if alues are string
	var1String, ok1 := var1.(string)
	var2String, ok2 := var2.(string)

	if ok1 && ok2 {
		if var1String != var2String {
			changed = true
		}
	}

	// Check if string slices
	if _, ok := var1.([]string); ok {
		if fmt.Sprint(var1) != fmt.Sprint(var2) {
			changed = true
		}
	}

	// Check if bools
	var1Bool, ok1 := var1.(bool)
	var2Bool, ok2 := var2.(bool)

	if ok1 && ok2 {
		if var1Bool != var2Bool {
			changed = true
		}
	}

	if changed {
		return fmt.Sprintf(" >> %s:", label)
	}

	return fmt.Sprintf("    %s:", label)
}

// editReleaseNote opens the user's editor for them to update the note.
//
//	In case of an editing error by the user, it returns shouldRetryEditing
//	set to true to retry editing.
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

	if strconv.FormatBool(originalNote.Feature) == strconv.FormatBool(modifiedNote.Feature) {
		unalteredFields.ReleaseNote.Feature = &originalNote.Feature
	} else {
		modifiedFields.ReleaseNote.Feature = &modifiedNote.Feature
		numChanges++
	}

	if strconv.FormatBool(originalNote.ActionRequired) == strconv.FormatBool(modifiedNote.ActionRequired) {
		unalteredFields.ReleaseNote.ActionRequired = &originalNote.ActionRequired
	} else {
		modifiedFields.ReleaseNote.ActionRequired = &modifiedNote.ActionRequired
		numChanges++
	}

	if strconv.FormatBool(originalNote.DoNotPublish) == strconv.FormatBool(modifiedNote.DoNotPublish) {
		unalteredFields.ReleaseNote.DoNotPublish = &originalNote.DoNotPublish
	} else {
		modifiedFields.ReleaseNote.DoNotPublish = &modifiedNote.DoNotPublish
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
			return false, fmt.Errorf("marshalling release note to map: %w", err)
		}

		output += "# " + strings.ReplaceAll(string(yamlCode), "\n", "\n# ")
	} else {
		// ... otherwise build a mixed map with the changes and the original
		// values commented out for reference
		yamlCode, err := yaml.Marshal(&modifiedFields)
		if err != nil {
			return false, fmt.Errorf("marshalling release note to map: %w", err)
		}

		unalteredYAML, err := yaml.Marshal(&unalteredFields.ReleaseNote)
		if err != nil {
			return false, fmt.Errorf("marshalling release note to map: %w", err)
		}

		output += string(yamlCode) + " # " + strings.ReplaceAll(string(unalteredYAML), "\n", "\n # ")
	}

	kubeEditor := editor.NewDefaultEditor([]string{"KUBE_EDITOR", "EDITOR"})

	changes, tempFilePath, err := kubeEditor.LaunchTempFile("release-notes-map-", ".yaml", bytes.NewReader([]byte(output)))
	if err != nil {
		return false, fmt.Errorf("while launching editor: %w", err)
	}

	defer func() {
		// Cleanup the temporary map file
		if err := os.Remove(tempFilePath); err != nil {
			logrus.Warn("could not remove temporary mapfile")
		}
	}()

	// If the map was not modified, we don't make any changes
	if string(changes) == output || len(changes) == 0 {
		logrus.Info("Release notes map was not modified")

		return false, nil
	}

	// If the yaml file is blank, return non error
	lines := strings.Split(string(changes), "\n")
	re := regexp.MustCompile(`^\s*#|^\s*$`)
	blankFile := true

	for _, line := range lines {
		// If only only one line is not blank/comment
		if line != "---" && !re.MatchString(line) {
			blankFile = false

			break
		}
	}

	if blankFile {
		logrus.Info("YAML mapfile is blank, ignoring")

		return false, nil
	}

	// Verify that the new yaml is valid and can be serialized back into a Map
	testMap := notes.ReleaseNotesMap{}

	err = yaml.Unmarshal(changes, &testMap)
	if err != nil {
		logrus.Error("The YAML code has errors")

		return true, fmt.Errorf("while verifying if changes are a valid map: %w", err)
	}

	if testMap.PR == 0 {
		logrus.Error("The yaml code does not have a PR number")

		return true, errors.New("invalid map: the YAML code did not have a PR number")
	}

	testMap.PRBody = &originalNote.PRBody

	// Remarshall the newyaml to save only the new values
	newYAML, err := yaml.Marshal(testMap)
	if err != nil {
		return true, fmt.Errorf("while re-marshaling new map: %w", err)
	}

	// Write the new map, removing the instructions
	mapPath := filepath.Join(workDir, mapsMainDirectory, fmt.Sprintf("pr-%d-map.yaml", pr))

	err = os.WriteFile(mapPath, newYAML, os.FileMode(0o644))
	if err != nil {
		logrus.Errorf("Error writing map to %s: %s", mapPath, err)

		return true, fmt.Errorf("writing modified release note map: %w", err)
	}

	return false, nil
}

// createNotesWorkDir creates the release notes working directory.
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
				return fmt.Errorf("creating working directory: %w", err)
			}
		}
	}

	return nil
}

// confirmWithUser avoid running when a users chooses n in interactive mode.
func confirmWithUser(opts *releaseNotesOptions, question string) bool {
	if !opts.interactiveMode {
		return true
	}

	_, success, err := util.Ask(question+" (Y/n)", "y:Y:yes|n:N:no|y", 10)
	if err != nil {
		logrus.Error(err)

		var userInputErr util.UserInputError
		if errors.As(err, &userInputErr) && userInputErr.IsCtrlC() {
			os.Exit(1)
		}

		return false
	}

	if success {
		return true
	}

	return false
}

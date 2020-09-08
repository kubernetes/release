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
	// websiteBranchPrefix name of the website branch in the user's fork
	websiteBranchPrefix = "release-notes-json-"
	// draftBranchPrefix name of the draft branch in the user's fork
	draftBranchPrefix = "release-notes-draft-"
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
		if !releaseNotesOpts.createDraftPR && !releaseNotesOpts.createWebsitePR {
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
	githubOrg       string
	draftRepo       string
	createDraftPR   bool
	createWebsitePR bool
	dependencies    bool
	websiteRepo     string
	mapProviders    []string
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
		&releaseNotesOpts.githubOrg,
		"org",
		"",
		"a Github organization hosting the forks of k/sig-release or ksigs/release-notes",
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

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.websiteRepo,
		"website-repo",
		"release-notes",
		"ksigs/release-notes fork used when creating the website PR",
	)

	releaseNotesCmd.PersistentFlags().StringVar(
		&releaseNotesOpts.draftRepo,
		"draft-repo",
		git.DefaultGithubReleaseRepo,
		"k/sig-release fork used when creating the draft PR",
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
			return errors.Wrapf(err, "while checking %s/%s fork", defaultKubernetesSigsOrg, defaultKubernetesSigsRepo)
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
	s, err := util.TagStringToSemver(tag)
	if err != nil {
		return errors.Wrapf(err, "no valid tag: %v", tag)
	}

	// Release notes are built from the first RC in the previous
	// minor to encompass all changes received after Code Thaw,
	// the point where the last minor was forked.
	start := util.SemverToTagString(semver.Version{
		Major: s.Major,
		Minor: s.Minor - 1,
		Patch: 0,
		Pre:   []semver.PRVersion{{VersionStr: "rc.1"}},
	})

	gh := github.New()

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
	result, err := releaseNotesFrom(start)
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

	defer func() {
		err = sigReleaseRepo.Cleanup()
	}()

	// generate the notes
	targetdir := filepath.Join(sigReleaseRepo.Dir(), "releases", fmt.Sprintf("release-%d.%d", s.Major, s.Minor))
	logrus.Debugf("Release notes markdown will be written to %v", targetdir)
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
		return errors.Wrapf(err, "creating commit in %v/%v", releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo)
	}

	// push to the user's remote
	logrus.Infof("Pushing modified release notes draft to %v/%v", releaseNotesOpts.githubOrg, releaseNotesOpts.draftRepo)
	if err := sigReleaseRepo.PushToRemote(userForkName, branchname); err != nil {
		return errors.Wrapf(err, "pushing %v to remote", userForkName)
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
func createWebsitePR(tag string) error {
	_, err := util.TagStringToSemver(tag)
	if err != nil {
		return errors.Wrapf(err, "no valid tag: %v", tag)
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
		return errors.Wrapf(err, "pushing %v to %v/%v", userForkName, releaseNotesOpts.githubOrg, releaseNotesOpts.websiteRepo)
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

	logrus.Infof("Successfully created PR #%d", pr.GetNumber())
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
			return startTag, errors.Wrap(err, "getting previous tag from branch")
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

func releaseNotesFrom(startTag string) (*releaseNotesResult, error) {
	logrus.Info("Generating release notes")

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

	doc, err := document.New(
		releaseNotes, notesOptions.StartRev, notesOptions.EndRev,
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
			notesOptions.StartRev, notesOptions.EndRev,
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
		return errors.Wrapf(err, "no valid tag: %v", releaseNotesOpts.tag)
	}

	// Options for PR creation
	if releaseNotesOpts.createDraftPR || releaseNotesOpts.createWebsitePR {
		if releaseNotesOpts.githubOrg == "" {
			return errors.New("cannot generate the Release Notes PR without --org")
		}
	}

	return nil
}

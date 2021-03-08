/*
Copyright © The Kubernetes Authors

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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/notes"
	notesoptions "k8s.io/release/pkg/notes/options"
	"k8s.io/release/pkg/rnd"
	"k8s.io/release/pkg/util"
)

// ATTENTION: if you're modifying this struct, make sure you update the command help
type fetchOptions struct {
	GithubAuth     string `split_words:"true" required:"true"` // TODO: more generic field?
	SourceRepo     string `split_words:"true" required:"true" default:"kubernetes/kubernetes"`
	SourceRepoPath string `split_words:"true"`
	ReleaseTag     string `split_words:"true" required:"true"`
	StorageRepo    string `split_words:"true" required:"true" default:"wilsonehusin/k8s-release-notes-data"` // TODO: create new repository?
}

var fetchOpts = &fetchOptions{}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetches the latest changes to store",
	Long: `rnd fetch -- Fetch latest release notes

This command will look up relevant pull requests in RND_SOURCE_REPO,
add / update the entries in RND_STORAGE_REPO`,
	PreRunE: func(*cobra.Command, []string) error {
		return processFetchFlags()
	},
	Run: func(*cobra.Command, []string) {
		if err := rndFetch(); err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
	},
}

func init() {
	var optionsUsage bytes.Buffer
	if err := envconfig.Usagef(progName, fetchOpts, &optionsUsage, optionsUsageTemplate); err != nil {
		panic(err)
	}
	fetchCmd.SetUsageTemplate(fetchCmd.UsageTemplate() + optionsUsageHeader + optionsUsage.String() + rootCmdOptionsUsage())
	rootCmd.AddCommand(fetchCmd)
}

func processFetchFlags() error {
	if err := envconfig.Process(progName, fetchOpts); err != nil {
		return err
	}

	return nil
}

func rndFetch() error {
	releaseTagVersion, err := util.TagStringToSemver(fetchOpts.ReleaseTag)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"releaseTagVersion": releaseTagVersion,
	}).Debug("converted string tag to semver")

	startTag, err := getPreviousTag(&releaseTagVersion)
	if err != nil {
		return err
	}

	notesOpts := notesoptions.New()
	notesOpts.Branch = git.DefaultBranch
	notesOpts.StartRev = startTag
	notesOpts.EndRev = fetchOpts.ReleaseTag
	notesOpts.Debug = logrus.StandardLogger().Level >= logrus.DebugLevel
	notesOpts.ListReleaseNotesV2 = true
	notesOpts.RepoPath = fetchOpts.SourceRepoPath
	notesOpts.SetGithubToken(fetchOpts.GithubAuth)

	logrus.Trace("begin k/k validation")
	if err := notesOpts.ValidateAndFinish(); err != nil {
		return err
	}
	logrus.Trace("completed k/k validation")

	releaseNotes, err := notes.GatherReleaseNotes(notesOpts)
	if err != nil {
		return err
	}

	storageGit, err := rnd.CloneStorageRepo(fetchOpts.StorageRepo)
	if err != nil {
		return err
	}

	storagePath := filepath.Join(rnd.StorageWorkDir(releaseTagVersion), "fetched")
	if err = writeReleaseNotesToGitRepo(releaseNotes, storageGit, storagePath); err != nil {
		return err
	}

	hasChanges, err := storageGit.IsDirty()
	if err != nil {
		return err
	}
	if hasChanges {
		if err := storageGit.Add("."); err != nil {
			return err
		}
		if err := storageGit.Commit("Add release notes"); err != nil {
			return err
		}

		if err = storageGit.PushToRemote("origin", "main"); err != nil {
			return err
		}
	} else {
		logrus.Info("no new changes to be committed")
		return nil
	}
	return nil
}

func writeReleaseNotesToGitRepo(releaseNotes *notes.ReleaseNotes, storageGit *git.Repo, storagePath string) error {
	workDir := filepath.Join(storageGit.Dir(), storagePath)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return err
	}
	for prnum, releaseNote := range releaseNotes.ByPR() {
		notePath := filepath.Join(workDir, fmt.Sprintf("%d.yaml", prnum))
		logrus.WithFields(logrus.Fields{
			"filepath": notePath,
		}).Debug("writing")
		noteContent, err := yaml.Marshal(releaseNote)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(notePath, noteContent, os.FileMode(0o644)); err != nil {
			return err
		}
	}
	return nil
}

// copy pasta'd from cmd/krel/cmd/release_notes.go with modifications
func getPreviousTag(releaseTagVersion *semver.Version) (string, error) {
	var previousTag, tagStrategy string
	if releaseTagVersion.Patch > 0 {
		previousTag = fmt.Sprintf("v%d.%d.%d", releaseTagVersion.Major, releaseTagVersion.Minor, releaseTagVersion.Patch-1)
		tagStrategy = "previous patch release"
	} else {
		if len(releaseTagVersion.Pre) == 2 && releaseTagVersion.Patch == 0 {
			// pre-releases
			previousTag = util.SemverToTagString(semver.Version{
				Major: releaseTagVersion.Major, Minor: releaseTagVersion.Minor - 1, Patch: 0,
			})
			tagStrategy = "previous minor release for pre-release"
		} else if len(releaseTagVersion.Pre) == 0 && releaseTagVersion.Patch == 0 {
			// full minor release
			previousTag = util.SemverToTagString(semver.Version{
				Major: releaseTagVersion.Major, Minor: releaseTagVersion.Minor - 1, Patch: 0,
			})
			tagStrategy = "previous minor release for new minor"
		} else {
			return "", fmt.Errorf("unable to determine previous tag")
		}
	}
	logrus.WithFields(logrus.Fields{
		"previousTag": previousTag,
		"tagStrategy": tagStrategy,
	}).Info("getting previous tag")
	return previousTag, nil
}

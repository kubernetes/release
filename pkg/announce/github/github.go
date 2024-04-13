/*
Copyright 2024 The Kubernetes Authors.

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

package github

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/release-sdk/github"
	"sigs.k8s.io/release-utils/hash"
	"sigs.k8s.io/release-utils/util"
)

// UpdateGitHubPage updates a github page with data from the release
func UpdateGitHubPage(opts *Options) (err error) {
	token := os.Getenv(github.TokenEnvKey)
	if token == "" {
		return errors.New("cannot update release page without a GitHub token")
	}

	gh := github.New()
	releaseVerb := "Posting"
	semver, err := util.TagStringToSemver(opts.Tag)
	if err != nil {
		return fmt.Errorf("parsing semver from tag: %w", err)
	}

	// Determine if this is a prerelase
	// // [[ "$FLAGS_type" == official ]] && prerelease="false"
	isPrerelease := false
	if len(semver.Pre) > 0 {
		isPrerelease = true
	}

	// Process the specified assets
	releaseAssets, err := processAssetFiles(opts.AssetFiles)
	if err != nil {
		return fmt.Errorf("processing the asset file list: %w", err)
	}

	// Substitution struct for the template
	subs := struct {
		Substitutions map[string]string
		Assets        []map[string]string
	}{
		Substitutions: opts.Substitutions,
		Assets:        releaseAssets,
	}

	// If we have a release notes file defined and set a substitution
	// entry for its contents
	if opts.ReleaseNotesFile != "" {
		rnData, err := os.ReadFile(opts.ReleaseNotesFile)
		if err != nil {
			return fmt.Errorf("reading release notes file: %w", err)
		}
		subs.Substitutions["ReleaseNotes"] = string(rnData)
	}

	// Open the template file (if a custom)
	templateText := ghPageBody
	if opts.PageTemplate != "" {
		logrus.Debugf("Using custom page template %s", opts.PageTemplate)
		templateText = opts.PageTemplate
	}
	// Parse the template we will use to build the release page
	tmpl, err := template.New("GitHubPage").Parse(templateText)
	if err != nil {
		return fmt.Errorf("parsing github page template: %w", err)
	}

	// Run the template to verify the output.
	output := new(bytes.Buffer)
	err = tmpl.Execute(output, subs)
	if err != nil {
		return fmt.Errorf("executing page template: %w", err)
	}

	// If we are in mock, we write it to stdout and exit. All checks
	// performed to the repo are skipped as the tag may not exist yet.
	if !opts.NoMock {
		logrus.Info("Mock mode, outputting the release page")
		_, err := os.Stdout.Write(output.Bytes())
		if err != nil {
			return fmt.Errorf("writing github page to stdout: %w", err)
		}
		return nil
	}

	// Check to see that a tag exists.
	// non-draft release posts to github create a tag.  We don't want to
	// create any tags on the repo this way. The tag should already exist
	// as a result of the release process.
	tagFound, err := gh.TagExists(opts.Owner, opts.Repo, opts.Tag)
	if err != nil {
		return fmt.Errorf("checking if the tag already exists in GitHub: %w", err)
	}
	if !tagFound {
		logrus.Warnf("The %s tag doesn't exist yet on GitHub.", opts.Tag)
		logrus.Warnf("That can't be good.")
		logrus.Warnf("We certainly cannot publish a release without a tag.")
		return errors.New("tag not found while trying to publish release page")
	}

	// Get the release we are looking for. We only need to fetch prereleases
	// if the release is a prerelease. If we don't filter them out, comparing
	// e.g. v1.30.0-rc.0 with v1.29.4 will incorrectly determine that v1.29.4
	// is *not* the latest (stable) release.
	releases, err := gh.Releases(opts.Owner, opts.Repo, isPrerelease)
	if err != nil {
		return fmt.Errorf("listing the repositories releases: %w", err)
	}

	// Does the release exist yet and should it be marked as latest?
	var releaseID int64
	commitish := ""
	// No pre-release should ever be marked as "latest"
	markAsLatest := !isPrerelease

	for _, release := range releases {
		if release.GetTagName() == opts.Tag {
			releaseID = release.GetID()
			commitish = release.GetTargetCommitish()
		} else if markAsLatest {
			// If this release is not identical to the one being cut right now,
			// we will check if *our* release is lower than the release from the loop.
			// If the first page of releases does not include a release that is
			// greater than *our* release, we can assume that ours will be the
			// latest release right now.
			releaseSemver, err := util.TagStringToSemver(release.GetTagName())
			if err != nil {
				return fmt.Errorf("parsing existing release tags as semver: %w", err)
			}

			if semver.LE(releaseSemver) {
				markAsLatest = false
			}
		}
	}

	if releaseID != 0 {
		logrus.Warnf("The %s is already published on github.", opts.Tag)
		if !opts.UpdateIfReleaseExists {
			return errors.New("release " + opts.Tag + " already exists. Left intact")
		}
		logrus.Infof("Using release id %d to update existing release.", releaseID)
		releaseVerb = "Updating"
	}

	// Post release data
	logrus.Infof("%s the %s release on github...", releaseVerb, opts.Tag)

	ghOpts := &github.UpdateReleasePageOptions{
		Name:       &opts.Name,
		Body:       ptr.To(output.String()),
		Draft:      &opts.Draft,
		Prerelease: &isPrerelease,
		Latest:     ptr.To(markAsLatest),
	}

	// Call GitHub to set the release page
	release, err := gh.UpdateReleasePageWithOptions(
		opts.Owner, opts.Repo, releaseID,
		opts.Tag, commitish,
		ghOpts,
	)
	if err != nil {
		return fmt.Errorf("updating the release on GitHub: %w", err)
	}

	// Releases often take a bit of time to show up in the API
	// after creating the page. If the release does not appear
	// in the API right away , sleep 3 secs and retry 3 times.
	for checkAttempts := 3; checkAttempts >= 0; checkAttempts-- {
		releaseFound := false
		releases, err = gh.Releases(opts.Owner, opts.Repo, true)
		if err != nil {
			return fmt.Errorf("listing releases in repository: %w", err)
		}
		// Check if the page shows up in the API
		for _, testRelease := range releases {
			if testRelease.GetID() == release.GetID() {
				releaseFound = true
				break
			}
		}
		if releaseFound {
			break
		}

		if checkAttempts == 0 {
			return errors.New("release not found, even when call to github was successful")
		}
		logrus.Info("Release page not yet returned by the GitHub API, sleeping and retrying")
		time.Sleep(3 * time.Second)
	}

	// Delete any assets reviously uploaded
	if err := deleteReleaseAssets(gh, opts.Owner, opts.Repo, release.GetID()); err != nil {
		return fmt.Errorf("deleting the existing release assets: %w", err)
	}

	// publish binary
	for _, assetData := range releaseAssets {
		logrus.Infof("Uploading %s as release asset", assetData["realpath"])
		asset, err := gh.UploadReleaseAsset(opts.Owner, opts.Repo, release.GetID(), assetData["rawpath"])
		if err != nil {
			return fmt.Errorf("uploading %s to the release: %w", assetData["realpath"], err)
		}
		logrus.Info("Successfully uploaded asset #", asset.GetID())
	}
	logrus.Infof("Release %s published on GitHub", opts.Tag)
	return nil
}

// processAssetFiles reads the command line strings and returns
// a map holding the needed info from the asset files
func processAssetFiles(assetFiles []string) (releaseAssets []map[string]string, err error) {
	// Check all asset files and get their hashes
	for _, path := range assetFiles {
		assetData := map[string]string{
			"rawpath": path,
			"name":    "",
		}
		// Check if asset path has a label
		if strings.Contains(path, ":") {
			p := strings.SplitN(path, ":", 2)
			if len(p) == 2 {
				path = p[0]
				assetData["name"] = p[1]
			}
		}

		logrus.Debugf("Checking asset file %s", path)

		// Verify path exists
		if !util.Exists(path) {
			return nil, errors.New("unable to render release page, asset file does not exist")
		}

		assetData["realpath"] = path
		assetData["filename"] = filepath.Base(path)

		fileHashes, err := getFileHashes(path)
		if err != nil {
			return nil, fmt.Errorf("getting the hashes: %w", err)
		}

		assetData["sha512"] = fileHashes["512"]
		assetData["sha256"] = fileHashes["256"]

		releaseAssets = append(releaseAssets, assetData)
	}
	return releaseAssets, nil
}

func deleteReleaseAssets(gh *github.GitHub, owner, repo string, releaseID int64) error {
	// If the release already contains assets, delete them to match
	// the new uploads we are sending
	currentAssets, err := gh.ListReleaseAssets(owner, repo, releaseID)
	if err != nil {
		return fmt.Errorf("while checking if the release already has assets: %w", err)
	}
	if len(currentAssets) == 0 {
		logrus.Info("No assets found in release")
		return nil
	}

	logrus.Warnf("Deleting %d release assets to upload the latest files", len(currentAssets))
	for _, asset := range currentAssets {
		logrus.Infof("Deleting %s", asset.GetName())
		if err := gh.DeleteReleaseAsset(owner, repo, asset.GetID()); err != nil {
			return fmt.Errorf("deleting existing release assets: %w", err)
		}
	}
	return nil
}

// getFileHashes obtains a file's sha256 and 512
func getFileHashes(path string) (hashes map[string]string, err error) {
	sha256, err := hash.SHA256ForFile(path)
	if err != nil {
		return nil, fmt.Errorf("get sha256: %w", err)
	}

	sha512, err := hash.SHA512ForFile(path)
	if err != nil {
		return nil, fmt.Errorf("get sha512: %w", err)
	}

	return map[string]string{"256": sha256, "512": sha512}, nil
}

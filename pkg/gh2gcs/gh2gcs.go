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

package gh2gcs

import (
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/gcp/gcs"
	"k8s.io/release/pkg/github"
)

// Config contains a slice of `ReleaseConfig` to be used when unmarshalling a
// yaml config containing multiple repository configs.
type Config struct {
	ReleaseConfigs []ReleaseConfig
}

// ReleaseConfig contains source (GitHub) and destination (GCS) information
// to perform a copy/upload operation using gh2gcs.
type ReleaseConfig struct {
	Org        string
	Repo       string
	Tags       []string
	GCSBucket  string
	ReleaseDir string
}

// DownloadReleases downloads release assets to a local directory
// Assets to download are derived from the tags specified in `ReleaseConfig`.
func DownloadReleases(releaseCfg *ReleaseConfig, ghClient *github.GitHub, outputDir string) error {
	tags := releaseCfg.Tags
	tagsString := strings.Join(tags, ", ")

	logrus.Infof(
		"Downloading assets for the following %s/%s release tags: %s",
		releaseCfg.Org,
		releaseCfg.Repo,
		tagsString,
	)
	if err := ghClient.DownloadReleaseAssets(releaseCfg.Org, releaseCfg.Repo, tags, outputDir); err != nil {
		return err
	}

	return nil
}

// Upload copies a set of release assets from local directory to GCS
// Assets to upload are derived from the tags specified in `ReleaseConfig`.
func Upload(releaseCfg *ReleaseConfig, ghClient *github.GitHub, outputDir string) error {
	uploadBase := filepath.Join(outputDir, releaseCfg.Org, releaseCfg.Repo)
	gcsPath := filepath.Join(releaseCfg.GCSBucket, releaseCfg.ReleaseDir)

	tags := releaseCfg.Tags
	for _, tag := range tags {
		srcDir := filepath.Join(uploadBase, tag)
		if err := gcs.CopyToGCS(srcDir, gcsPath, true, true); err != nil {
			return err
		}
	}

	return nil
}

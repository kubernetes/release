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
	"k8s.io/utils/pointer"
)

// Config contains a slice of `ReleaseConfig` to be used when unmarshalling a
// yaml config containing multiple repository configs.
type Config struct {
	ReleaseConfigs []ReleaseConfig `yaml:"releaseConfigs"`
}

// ReleaseConfig contains source (GitHub) and destination (GCS) information
// to perform a copy/upload operation using gh2gcs.
type ReleaseConfig struct {
	Org                string       `yaml:"org"`
	Repo               string       `yaml:"repo"`
	Tags               []string     `yaml:"tags"`
	IncludePrereleases bool         `yaml:"includePrereleases"`
	GCSBucket          string       `yaml:"gcsBucket"`
	ReleaseDir         string       `yaml:"releaseDir"`
	GCSCopyOptions     *gcs.Options `yaml:"gcsCopyOptions"`
}

// DefaultGCSCopyOptions have the default options for the GCS copy action
var DefaultGCSCopyOptions = &gcs.Options{
	Concurrent:   pointer.BoolPtr(true),
	Recursive:    pointer.BoolPtr(true),
	NoClobber:    pointer.BoolPtr(true),
	AllowMissing: pointer.BoolPtr(true),
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
		if err := gcs.CopyToGCS(srcDir, gcsPath, releaseCfg.GCSCopyOptions); err != nil {
			return err
		}
	}

	return nil
}

// CheckGCSCopyOptions checks if the user set any config or we need to set the default config
func CheckGCSCopyOptions(copyOptions *gcs.Options) *gcs.Options {
	// set the GCS Copy options to default values
	if copyOptions == nil {
		return DefaultGCSCopyOptions
	}

	if copyOptions.AllowMissing == nil {
		copyOptions.AllowMissing = pointer.BoolPtr(true)
	}

	if copyOptions.Concurrent == nil {
		copyOptions.Concurrent = pointer.BoolPtr(true)
	}

	if copyOptions.NoClobber == nil {
		copyOptions.NoClobber = pointer.BoolPtr(true)
	}

	if copyOptions.Recursive == nil {
		copyOptions.Recursive = pointer.BoolPtr(true)
	}

	return copyOptions
}

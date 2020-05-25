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

	"k8s.io/release/pkg/gcp/gcs"
	"k8s.io/release/pkg/github"
)

type Config struct {
	ReleaseConfigs []ReleaseConfig
}

type ReleaseConfig struct {
	Org        string
	Repo       string
	Tags       []string
	GCSBucket  string
	ReleaseDir string
}

func DownloadReleases(releaseCfg *ReleaseConfig, ghClient *github.GitHub, outputDir string) error {
	tags := releaseCfg.Tags
	if err := ghClient.DownloadReleaseAssets(releaseCfg.Org, releaseCfg.Repo, tags, outputDir); err != nil {
		return err
	}

	return nil
}

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

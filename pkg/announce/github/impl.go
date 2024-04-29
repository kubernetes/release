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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	githubsdk "sigs.k8s.io/release-sdk/github"
	"sigs.k8s.io/release-utils/hash"
	"sigs.k8s.io/release-utils/util"
)

type defaultImpl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . impl
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt githubfakes/fake_impl.go > githubfakes/_fake_impl.go && mv githubfakes/_fake_impl.go githubfakes/fake_impl.go"

type impl interface {
	github() *githubsdk.GitHub
	processAssetFiles(assetFiles []string) (releaseAssets []map[string]string, err error)
}

func (i *defaultImpl) github() *githubsdk.GitHub {
	return githubsdk.New()
}

// processAssetFiles reads the command line strings and returns
// a map holding the needed info from the asset files
func (i *defaultImpl) processAssetFiles(assetFiles []string) (releaseAssets []map[string]string, err error) {
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

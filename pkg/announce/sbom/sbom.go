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

package sbom

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"sigs.k8s.io/bom/pkg/serialize"
	"sigs.k8s.io/bom/pkg/spdx"
	"sigs.k8s.io/release-sdk/github"
)

// GenerateReleaseSBOM creates an SBOM describing the release
func GenerateReleaseSBOM(opts *Options) (string, error) {
	// Create a temporary file to write the sbom
	dir, err := os.MkdirTemp("", "project-sbom-")
	if err != nil {
		return "", fmt.Errorf("creating temporary directory to write sbom: %w", err)
	}

	sbomFile := filepath.Join(dir, sbomFileName)
	logrus.Infof("SBOM will be temporarily written to %s", sbomFile)

	builder := spdx.NewDocBuilder()
	builderOpts := &spdx.DocGenerateOptions{
		ProcessGoModules: true,
		ScanLicenses:     true,
		Name:             opts.ReleaseName,
		Namespace:        github.GitHubURL + opts.Repo + "@" + opts.Tag,
		Directories:      []string{opts.RepoDirectory},
	}

	doc, err := builder.Generate(builderOpts)
	if err != nil {
		return "", fmt.Errorf("generating initial SBOM: %w", err)
	}

	// Add the download location and version to the first
	// SPDX package (which represents the repo)
	for t := range doc.Packages {
		doc.Packages[t].Version = opts.Tag
		doc.Packages[t].DownloadLocation = "git+" + github.GitHubURL + opts.Repo + "@" + opts.Tag
		break
	}

	// List all artifacts and add them
	spdxClient := spdx.NewSPDX()
	for _, f := range opts.Assets {
		logrus.Infof("Adding file %s to SBOM", f.Path)
		spdxFile, err := spdxClient.FileFromPath(f.ReadFrom)
		if err != nil {
			return "", fmt.Errorf("adding %s to SBOM: %w", f.ReadFrom, err)
		}
		spdxFile.Name = f.Path
		spdxFile.BuildID() // This is a boog in the spdx pkg, we have to call manually
		spdxFile.DownloadLocation = github.GitHubURL + filepath.Join(
			opts.Repo, assetDownloadPath, opts.Tag, f.Path,
		)
		if err := doc.AddFile(spdxFile); err != nil {
			return "", fmt.Errorf("adding %s as SPDX file to SBOM: %w", f.ReadFrom, err)
		}
	}

	var renderer serialize.Serializer
	switch opts.Format {
	case FormatJSON:
		renderer = &serialize.JSON{}
	case FormatTagValue:
		renderer = &serialize.TagValue{}
	default:
		return "", fmt.Errorf("invalid SBOM format, must be one of %s, %s", FormatJSON, FormatTagValue)
	}

	markup, err := renderer.Serialize(doc)
	if err != nil {
		return "", fmt.Errorf("serializing sbom: %w", err)
	}

	if err := os.WriteFile(sbomFile, []byte(markup), 0o600); err != nil {
		return "", fmt.Errorf("writing sbom to disk: %w", err)
	}

	return sbomFile, nil
}

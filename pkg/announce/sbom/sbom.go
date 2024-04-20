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
	"path/filepath"

	"github.com/sirupsen/logrus"

	"sigs.k8s.io/bom/pkg/serialize"
	"sigs.k8s.io/bom/pkg/spdx"
	"sigs.k8s.io/release-sdk/github"
)

type SBOM struct {
	options *Options
	impl
}

// NewGitHub returns a new GitHub instance.
func NewSBOM(opts *Options) *SBOM {
	return &SBOM{
		impl:    &defaultImpl{},
		options: opts,
	}
}

// SetImplementation sets the implementation to handle file operations and SPDX.
func (s *SBOM) SetImplementation(i impl) {
	s.impl = i
}

// Generate creates an SBOM describing the release.
func (s *SBOM) Generate() (string, error) {
	// Create a temporary file to write the sbom
	sbomFile, err := s.impl.tmpFile()
	if err != nil {
		return "", fmt.Errorf("setting up temporary file for SBOM: %w", err)
	}
	logrus.Infof("SBOM will be temporarily written to %s", sbomFile)

	builder := s.impl.docBuilder()
	builderOpts := &spdx.DocGenerateOptions{
		ProcessGoModules: true,
		ScanLicenses:     true,
		Name:             s.options.ReleaseName,
		Namespace:        github.GitHubURL + s.options.Repo + "@" + s.options.Tag,
		Directories:      []string{s.options.RepoDirectory},
	}

	doc, err := builder.Generate(builderOpts)
	if err != nil {
		return "", fmt.Errorf("generating initial SBOM: %w", err)
	}

	// Add the download location and version to the first
	// SPDX package (which represents the repo)
	for t := range doc.Packages {
		doc.Packages[t].Version = s.options.Tag
		doc.Packages[t].DownloadLocation = "git+" + github.GitHubURL + s.options.Repo + "@" + s.options.Tag
		break
	}

	// List all artifacts and add them
	spdxClient := s.impl.spdxClient()
	for _, f := range s.options.Assets {
		logrus.Infof("Adding file %s to SBOM", f.Path)
		spdxFile, err := spdxClient.FileFromPath(f.ReadFrom)
		if err != nil {
			return "", fmt.Errorf("adding %s to SBOM: %w", f.ReadFrom, err)
		}
		spdxFile.Name = f.Path
		spdxFile.BuildID() // This is a boog in the spdx pkg, we have to call manually
		spdxFile.DownloadLocation = github.GitHubURL + filepath.Join(
			s.options.Repo, assetDownloadPath, s.options.Tag, f.Path,
		)
		if err := doc.AddFile(spdxFile); err != nil {
			return "", fmt.Errorf("adding %s as SPDX file to SBOM: %w", f.ReadFrom, err)
		}
	}

	var renderer serialize.Serializer
	switch s.options.Format {
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

	if err := s.impl.writeFile(sbomFile, []byte(markup)); err != nil {
		return "", fmt.Errorf("writing sbom to disk: %w", err)
	}

	return sbomFile, nil
}

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
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blang/semver/v4"

	"sigs.k8s.io/bom/pkg/bom"
	"sigs.k8s.io/bom/pkg/license"
	"sigs.k8s.io/bom/pkg/spdx"
)

type defaultImpl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . impl
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt sbomfakes/fake_impl.go > sbomfakes/_fake_impl.go && mv sbomfakes/_fake_impl.go sbomfakes/fake_impl.go"

type impl interface {
	tmpFile() (string, error)
	generateDocument(options *bom.GenerateOptions) (*spdx.Document, error)
	spdxClient() *spdx.SPDX
	writeFile(file string, data []byte) error
}

func (i *defaultImpl) tmpFile() (string, error) {
	// Create a temporary file to write the sbom
	dir, err := os.MkdirTemp("", "project-sbom-")
	if err != nil {
		return "", fmt.Errorf("creating temporary directory to write sbom: %w", err)
	}

	return filepath.Join(dir, sbomFileName), nil
}

// generateDocument generates an SBOM and converts it to the SPDX object
// model, completing it the same way bom does when writing SPDX documents.
func (i *defaultImpl) generateDocument(options *bom.GenerateOptions) (*spdx.Document, error) {
	pdoc, err := bom.Generate(context.Background(), options)
	if err != nil {
		return nil, fmt.Errorf("generating SBOM: %w", err)
	}

	doc, err := spdx.FromProtobom(pdoc)
	if err != nil {
		return nil, fmt.Errorf("converting SBOM to SPDX: %w", err)
	}

	listVersion, err := semver.ParseTolerant(license.DefaultCatalogOpts.Version)
	if err != nil {
		return nil, fmt.Errorf("parsing license list version: %w", err)
	}

	doc.LicenseListVersion = fmt.Sprintf("%d.%d", listVersion.Major, listVersion.Minor)

	if doc.Creator.Organization == "" {
		doc.Creator.Organization = sbomOrganization
	}

	return doc, nil
}

func (i *defaultImpl) spdxClient() *spdx.SPDX {
	return spdx.NewSPDX()
}

func (i *defaultImpl) writeFile(file string, data []byte) error {
	return os.WriteFile(file, data, 0o600)
}

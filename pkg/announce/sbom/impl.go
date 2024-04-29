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

	"sigs.k8s.io/bom/pkg/spdx"
)

type defaultImpl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . impl
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt sbomfakes/fake_impl.go > sbomfakes/_fake_impl.go && mv sbomfakes/_fake_impl.go sbomfakes/fake_impl.go"

type impl interface {
	tmpFile() (string, error)
	docBuilder() *spdx.DocBuilder
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

func (i *defaultImpl) docBuilder() *spdx.DocBuilder {
	return spdx.NewDocBuilder()
}

func (i *defaultImpl) spdxClient() *spdx.SPDX {
	return spdx.NewSPDX()
}

func (i *defaultImpl) writeFile(file string, data []byte) error {
	return os.WriteFile(file, data, 0o600)
}

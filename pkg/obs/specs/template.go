/*
Copyright 2023 The Kubernetes Authors.

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

package specs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

type work struct {
	src    string
	dst    string
	t      *template.Template
	info   os.FileInfo
	pkgDef *PackageDefinition
}

// BuildSpecs creates spec file based on provided package definition.
func (c *Client) BuildSpecs(pkgDef *PackageDefinition, specOnly bool) (err error) {
	if pkgDef == nil {
		return fmt.Errorf("package definition cannot be nil")
	}

	workItems := []work{}

	tplDir := filepath.Join(pkgDef.SpecTemplatePath, pkgDef.Name)
	if _, err := os.Stat(tplDir); err != nil {
		return fmt.Errorf("building specs for %s: finding package template dir: %w", pkgDef.Name, err)
	}

	if err := filepath.Walk(tplDir, func(templateFile string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		specFile := filepath.Join(pkgDef.SpecOutputPath, pkgDef.Name, templateFile[len(tplDir):])

		if specFile == pkgDef.SpecOutputPath {
			return nil
		}

		if f.IsDir() {
			return os.Mkdir(specFile, f.Mode())
		}
		if filepath.Ext(templateFile) == ".spec" {
			// Spec is intentionally saved outside package dir, which is later on archived
			specFile = filepath.Join(pkgDef.SpecOutputPath, templateFile[len(tplDir):])
		} else if specOnly && filepath.Ext(templateFile) != ".spec" {
			return nil
		}

		t, err := template.
			New("").
			Option("missingkey=error").
			ParseFiles(templateFile)
		if err != nil {
			return err
		}

		workItems = append(workItems, work{
			src:    templateFile,
			dst:    specFile,
			t:      t.Templates()[0],
			info:   f,
			pkgDef: pkgDef,
		})

		return nil
	}); err != nil {
		return err
	}

	for _, item := range workItems {
		buf := bytes.Buffer{}
		if err := item.t.Execute(&buf, item.pkgDef); err != nil {
			return fmt.Errorf("executing template for %s: %w", item.src, err)
		}

		if err := os.WriteFile(
			item.dst, buf.Bytes(), item.info.Mode(),
		); err != nil {
			return fmt.Errorf("writing file %s: %w", item.dst, err)
		}
	}

	logrus.Info("Package specs have successfully been built!")

	return nil
}

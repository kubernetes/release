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
	"errors"
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

// BuildSpecs creates spec files based on provided templates.
func (c *Client) BuildSpecs(pkgBuilder *PackageBuilder) (err error) {
	if pkgBuilder == nil {
		return errors.New("package builder cannot be nil")
	}

	workItems := []work{}

	for _, pkg := range pkgBuilder.Definitions {
		tplDir := filepath.Join(pkgBuilder.TemplateDir, pkg.Name)
		if _, err := os.Stat(tplDir); err != nil {
			return fmt.Errorf("finding package template dir: %w", err)
		}

		if err := filepath.Walk(tplDir, func(templateFile string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			specFile := filepath.Join(pkgBuilder.OutputDir, pkg.Name, templateFile[len(tplDir):])

			if specFile == pkgBuilder.OutputDir {
				return nil
			}

			if f.IsDir() {
				return os.Mkdir(specFile, f.Mode())
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
				pkgDef: pkg,
			})

			return nil
		}); err != nil {
			return err
		}
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

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

package kubepkg

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type work struct {
	src  string
	dst  string
	t    *template.Template
	info os.FileInfo
}

func buildSpecs(bc *buildConfig, specDir string) (workItems []work, err error) {
	if err := filepath.Walk(bc.TemplateDir, func(templateFile string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		specFile := filepath.Join(specDir, templateFile[len(bc.TemplateDir):])
		if specFile == specDir {
			return nil
		}
		if f.IsDir() {
			return os.Mkdir(specFile, f.Mode())
		}
		t, err := template.
			New("").
			Funcs(builtins).
			Option("missingkey=error").
			ParseFiles(templateFile)
		if err != nil {
			return err
		}
		workItems = append(workItems, work{
			src:  templateFile,
			dst:  specFile,
			t:    t.Templates()[0],
			info: f,
		})

		return nil
	}); err != nil {
		return nil, err
	}

	for _, item := range workItems {
		buf := bytes.Buffer{}
		if err := item.t.Execute(&buf, bc); err != nil {
			return nil, errors.Wrapf(err, "executing template for %s", item.src)
		}

		if err := os.WriteFile(
			item.dst, buf.Bytes(), item.info.Mode(),
		); err != nil {
			return nil, errors.Wrapf(err, "writing file %s", item.dst)
		}
	}

	logrus.Info("Package specs have successfully been built")

	return workItems, nil
}

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
	"os"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

type work struct {
	src  string
	dst  string
	t    *template.Template
	info os.FileInfo
}

func buildTemplate(c cfg, srcdir, dstdir string) ([]work, error) {
	var w []work

	switch c.Type {
	case BuildDeb:
		if err := filepath.Walk(srcdir, func(srcfile string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			dstfile := filepath.Join(dstdir, srcfile[len(srcdir):])
			if dstfile == dstdir {
				return nil
			}
			if f.IsDir() {
				return os.Mkdir(dstfile, f.Mode())
			}
			t, err := template.
				New("").
				Funcs(builtins).
				Option("missingkey=error").
				ParseFiles(srcfile)
			if err != nil {
				return err
			}
			w = append(w, work{
				src:  srcfile,
				dst:  dstfile,
				t:    t.Templates()[0],
				info: f,
			})

			return nil
		}); err != nil {
			return nil, err
		}

		for _, w := range w {
			if err := func() error {
				f, err := os.OpenFile(w.dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0)
				if err != nil {
					return err
				}
				defer f.Close()

				if err := w.t.Execute(f, c); err != nil {
					return err
				}
				if err := os.Chmod(w.dst, w.info.Mode()); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return nil, err
			}
		}
	case BuildRpm:
		logrus.Fatal("Templating rpm specs via kubepkg is not currently supported")
	}

	return w, nil
}

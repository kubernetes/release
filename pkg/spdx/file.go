/*
Copyright 2021 The Kubernetes Authors.

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

package spdx

import (
	"bytes"
	"crypto/sha1"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/release-utils/hash"
	"sigs.k8s.io/release-utils/util"
)

var fileTemplate = `{{ if .Name }}FileName: {{ .Name }}
{{ end -}}
{{ if .ID }}SPDXID: {{ .ID }}
{{ end -}}
{{- if .Checksum -}}
{{- range $key, $value := .Checksum -}}
{{ if . }}FileChecksum: {{ $key }}: {{ $value }}
{{ end -}}
{{- end -}}
{{- end -}}
LicenseConcluded: {{ if .LicenseConcluded }}{{ .LicenseConcluded }}{{ else }}NOASSERTION{{ end }}
LicenseInfoInFile: {{ if .LicenseInfoInFile }}LicenseInfoInFile: {{ .LicenseInfoInFile }}{{ else }}NOASSERTION{{ end }}
FileCopyrightText: {{ if .CopyrightText }}<text>{{ .CopyrightText }}
</text>{{ else }}NOASSERTION{{ end }}

`

// File abstracts a file contained in a package
type File struct {
	Name              string // string /Makefile
	FileName          string // Name of the file
	ID                string // SPDXRef-Makefile
	LicenseConcluded  string // GPL-3.0-or-later
	LicenseInfoInFile string // GPL-3.0-or-later
	CopyrightText     string // NOASSERTION
	SourceFile        string // Source file to read from (not part of the spec)
	Checksum          map[string]string

	options *FileOptions // Options
}

func NewFile() (f *File) {
	f = &File{
		options: &FileOptions{},
	}
	return f
}

func (f *File) Options() *FileOptions {
	return f.options
}

// FileOptions
type FileOptions struct {
	WorkDir string
}

// ReadChecksums receives a path to a file and calculates its checksums
func (f *File) ReadChecksums(filePath string) error {
	if f.Checksum == nil {
		f.Checksum = map[string]string{}
	}
	file, err := os.Open(filePath)
	if err != nil {
		return errors.Wrap(err, "opening file for reading: "+filePath)
	}
	defer file.Close()
	// TODO: Make this line like the others once this PR is
	// included in a k-sigs/release-util release:
	// https://github.com/kubernetes-sigs/release-utils/pull/16
	s1, err := hash.ForFile(filePath, sha1.New())
	if err != nil {
		return errors.Wrap(err, "getting sha1 sum for file")
	}
	s256, err := hash.SHA256ForFile(filePath)
	if err != nil {
		return errors.Wrap(err, "getting file checksums")
	}
	s512, err := hash.SHA512ForFile(filePath)
	if err != nil {
		return errors.Wrap(err, "getting file checksums")
	}

	f.Checksum = map[string]string{
		"SHA1":   s1,
		"SHA256": s256,
		"SHA512": s512,
	}
	return nil
}

// Render renders the document fragment of a file
func (f *File) Render() (docFragment string, err error) {
	var buf bytes.Buffer
	tmpl, err := template.New("file").Parse(fileTemplate)
	if err != nil {
		return "", errors.Wrap(err, "parsing file template")
	}

	// Run the template to verify the output.
	if err := tmpl.Execute(&buf, f); err != nil {
		return "", errors.Wrap(err, "executing spdx file template")
	}

	docFragment = buf.String()
	return docFragment, nil
}

// ReadSourceFile reads the source file for the package and populates
//  the fields derived from it (Checksums and FileName)
func (f *File) ReadSourceFile(path string) error {
	if !util.Exists(path) {
		return errors.New("unable to find package source file")
	}

	if err := f.ReadChecksums(path); err != nil {
		return errors.Wrap(err, "reading file checksums")
	}

	f.SourceFile = path
	f.Name = strings.TrimPrefix(
		path, f.Options().WorkDir+string(filepath.Separator),
	)
	f.ID = "SPDXRef-File-" + f.Checksum["SHA256"][0:15]
	logrus.Infof("Added file %s as %s", f.Name, f.ID)
	return nil
}

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
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/release-utils/hash"
	"sigs.k8s.io/release-utils/util"
)

var packageTemplate = `##### Package: {{ .Name }}

{{ if .Name }}PackageName: {{ .Name }}
{{ end -}}
{{ if .ID }}SPDXID: {{ .ID }}
{{ end -}}
{{- if .Checksum -}}
{{- range $key, $value := .Checksum -}}
{{ if . }}PackageChecksum: {{ $key }}: {{ $value }}
{{ end -}}
{{- end -}}
{{- end -}}
PackageDownloadLocation: {{ if .DownloadLocation }}{{ .DownloadLocation }}{{ else }}NONE{{ end }}
FilesAnalyzed: {{ .FilesAnalyzed }}
{{ if .VerificationCode }}PackageVerificationCode: {{ .VerificationCode }}
{{ end -}}
PackageLicenseConcluded: {{ if .LicenseConcluded }}{{ .LicenseConcluded }}{{ else }}NOASSERTION{{ end }}
{{ if .FileName }}PackageFileName: {{ .FileName }}
{{ end -}}
{{ if .LicenseInfoFromFiles }}PackageLicenseInfoFromFiles: {{ .LicenseInfoFromFiles }}
{{ end -}}
{{ if .Version }}PackageVersion: {{ .Version }}
{{ end -}}
PackageLicenseDeclared: {{ if .LicenseDeclared }}{{ .LicenseDeclared }}{{ else }}NOASSERTION{{ end }}
PackageCopyrightText: {{ if .CopyrightText }}<text>{{ .CopyrightText }}
</text>{{ else }}NOASSERTION{{ end }}
`

// Package groups a set of files
type Package struct {
	FilesAnalyzed        bool   // true
	Name                 string // hello-go-src
	ID                   string // SPDXRef-Package-hello-go-src
	DownloadLocation     string // git@github.com:swinslow/spdx-examples.git#example6/content/src
	VerificationCode     string // 6486e016b01e9ec8a76998cefd0705144d869234
	LicenseConcluded     string // LicenseID o NOASSERTION
	LicenseInfoFromFiles string // GPL-3.0-or-later
	LicenseDeclared      string // GPL-3.0-or-later
	LicenseComments      string // record any relevant background information or analysis that went in to arriving at the Concluded License
	CopyrightText        string // string NOASSERTION
	Version              string // Package version
	FileName             string // Name of the package
	SourceFile           string // Source file for the package (taball for images, rpm, deb, etc)

	// Supplier: the actual distribution source for the package/directory
	Supplier struct {
		Person       string // person name and optional (<email>)
		Organization string // organization name and optional (<email>)
	}
	// Originator: For example, the SPDX file identifies the package glibc and Red Hat as the Package Supplier,
	// but the Free Software Foundation is the Package Originator.
	Originator struct {
		Person       string // person name and optional (<email>)
		Organization string // organization name and optional (<email>)
	}
	// Subpackages contained
	Packages map[string]*Package // Sub packages conatined in this pkg
	Files    map[string]*File    // List of files
	Checksum map[string]string   // Checksum of the package

	options *PackageOptions // Options
}

func NewPackage() (p *Package) {
	p = &Package{
		options: &PackageOptions{},
	}
	return p
}

type PackageOptions struct {
	WorkDir string // Working directory to read files from
}

func (p *Package) Options() *PackageOptions {
	return p.options
}

// ReadSourceFile reads the source file for the package and populates
//  the package fields derived from it (Checksums and FileName)
func (p *Package) ReadSourceFile(path string) error {
	if !util.Exists(path) {
		return errors.New("unable to find package source file")
	}
	s256, err := hash.SHA256ForFile(path)
	if err != nil {
		return errors.Wrap(err, "getting source file sha256")
	}
	s512, err := hash.SHA512ForFile(path)
	if err != nil {
		return errors.Wrap(err, "getting source file sha512")
	}
	p.Checksum = map[string]string{
		"SHA256": s256,
		"SHA512": s512,
	}
	p.SourceFile = path
	p.FileName = strings.TrimPrefix(path, p.Options().WorkDir+string(filepath.Separator))
	return nil
}

// AddFile adds a file contained in the package
func (p *Package) AddFile(file *File) error {
	if p.Files == nil {
		p.Files = map[string]*File{}
	}
	// If file does not have an ID, we try to build one
	// by hashing the file name
	if file.ID == "" {
		if file.Name == "" {
			return errors.New("unable to generate file ID, filename not set")
		}
		if p.Name == "" {
			return errors.New("unable to generate file ID, filename not set")
		}
		h := sha1.New()
		if _, err := h.Write([]byte(p.Name + ":" + file.Name)); err != nil {
			return errors.Wrap(err, "getting sha1 of filename")
		}
		file.ID = "SPDXRef-File-" + fmt.Sprintf("%x", h.Sum(nil))
	}
	p.Files[file.ID] = file
	return nil
}

// AddPackage adds a new subpackage to a package
func (p *Package) AddPackage(pkg *Package) error {
	if p.Packages == nil {
		p.Packages = map[string]*Package{}
	}
	if pkg.ID == "" {
		// If we so not have an ID but have a name generate it fro there
		reg := regexp.MustCompile("[^a-zA-Z0-9-]+")
		id := reg.ReplaceAllString(pkg.Name, "")
		if id != "" {
			pkg.ID = "SPDXRef-Package-" + id
		}
	}
	if pkg.ID == "" {
		return errors.New("package name is needed to add a new package")
	}
	if _, ok := p.Packages[pkg.ID]; ok {
		return errors.New("a package named " + pkg.ID + " already exists in the document")
	}

	p.Packages[pkg.ID] = pkg
	return nil
}

// Render renders the document fragment of the package
func (p *Package) Render() (docFragment string, err error) {
	var buf bytes.Buffer
	tmpl, err := template.New("package").Parse(packageTemplate)
	if err != nil {
		return "", errors.Wrap(err, "parsing package template")
	}

	// If files were analyzed, calculate the verification
	if p.FilesAnalyzed {
		if len(p.Files) == 0 {
			return docFragment, errors.New("unable to get package verification code, package has no files")
		}
		shaList := []string{}
		for _, f := range p.Files {
			if f.Checksum == nil {
				return docFragment, errors.New("unable to render package, file has no checksums")
			}
			if _, ok := f.Checksum["SHA1"]; !ok {
				return docFragment, errors.New("unable to render package, files were analyzed but some do not have sha1 checksum")
			}
			shaList = append(shaList, f.Checksum["SHA1"])
		}
		sort.Strings(shaList)
		h := sha1.New()
		if _, err := h.Write([]byte(strings.Join(shaList, ""))); err != nil {
			return docFragment, errors.Wrap(err, "getting sha1 verification of files")
		}
		p.VerificationCode = fmt.Sprintf("%x", h.Sum(nil))
	}

	// Run the template to verify the output.
	if err := tmpl.Execute(&buf, p); err != nil {
		return "", errors.Wrap(err, "executing spdx package template")
	}

	docFragment = buf.String()

	for _, f := range p.Files {
		fileFragment, err := f.Render()
		if err != nil {
			return "", errors.Wrap(err, "rendering file "+f.Name)
		}
		docFragment += fileFragment
		docFragment += fmt.Sprintf("Relationship: %s CONTAINS %s\n\n", p.ID, f.ID)
	}

	// Print the contained sub packages
	if p.Packages != nil {
		for _, pkg := range p.Packages {
			pkgDoc, err := pkg.Render()
			if err != nil {
				return "", errors.Wrap(err, "rendering pkg "+pkg.Name)
			}

			docFragment += pkgDoc
			docFragment += fmt.Sprintf("Relationship: %s CONTAINS %s\n\n", p.ID, pkg.ID)
		}
	}
	return docFragment, nil
}

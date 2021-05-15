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
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/release-utils/util"
)

func NewDocBuilder() *DocBuilder {
	db := &DocBuilder{
		options: &defaultDocBuilderOpts,
		impl:    defaultDocBuilderImpl{},
	}
	return db
}

// DocBuilder is a tool to write spdx manifests
type DocBuilder struct {
	options *DocBuilderOptions
	impl    DocBuilderImplementation
}

// Generate creates anew SPDX document describing the artifacts specified in the options
func (db *DocBuilder) Generate(genopts *DocGenerateOptions) (*Document, error) {
	// Create the SPDX document
	doc, err := db.impl.GenerateDoc(db.options, genopts)
	if err != nil {
		return nil, errors.Wrap(err, "creating SPDX document")
	}

	// If we have a specified output file, write it
	if genopts.OutputFile == "" {
		return doc, nil
	}

	return doc, errors.Wrapf(
		db.impl.WriteDoc(doc, genopts.OutputFile),
		"writing doc to %s", genopts.OutputFile,
	)
}

type DocGenerateOptions struct {
	Tarballs      []string // A slice of tar paths
	Files         []string // A slice of naked files to include in the bom
	Images        []string // A slice of docker images
	OutputFile    string   // Output location
	Namespace     string   // Namespace for the document (a unique URI)
	AnalyseLayers bool     // A flag that controls if deep layer analysis should be performed
}

func (o *DocGenerateOptions) Validate() error {
	if len(o.Tarballs) == 0 && len(o.Files) == 0 && len(o.Images) == 0 {
		return errors.New(
			"To build a document at least an image, tarball or a file has to be specified",
		)
	}
	return nil
}

type DocBuilderOptions struct {
	WorkDir string // Working directory (defaults to a tmp dir)
}

var defaultDocBuilderOpts = DocBuilderOptions{
	WorkDir: filepath.Join(os.TempDir(), "spdx-docbuilder"),
}

type DocBuilderImplementation interface {
	GenerateDoc(*DocBuilderOptions, *DocGenerateOptions) (*Document, error)
	WriteDoc(*Document, string) error
}

// defaultDocBuilderImpl is the default implementation for the
// SPDX document builder
type defaultDocBuilderImpl struct{}

// Generate generates a document
func (builder defaultDocBuilderImpl) GenerateDoc(
	opts *DocBuilderOptions, genopts *DocGenerateOptions,
) (doc *Document, err error) {
	if err := genopts.Validate(); err != nil {
		return nil, errors.Wrap(err, "checking build options")
	}

	spdx := NewSPDX()
	spdx.options.AnalyzeLayers = genopts.AnalyseLayers

	if !util.Exists(opts.WorkDir) {
		if err := os.MkdirAll(opts.WorkDir, os.FileMode(0o755)); err != nil {
			return nil, errors.Wrap(err, "creating builder worskpace dir")
		}
	}

	tmpdir, err := os.MkdirTemp(opts.WorkDir, "doc-build-")
	if err != nil {
		return nil, errors.Wrapf(err, "creating temporary workdir in %s", opts.WorkDir)
	}
	defer os.RemoveAll(tmpdir)

	// Create the new document
	doc = NewDocument()
	doc.Namespace = genopts.Namespace

	if genopts.Namespace == "" {
		logrus.Warn("Document namespace is empty, a mock URI will be supplied but the doc will not be valid")
		doc.Namespace = "http://example.com/"
	}

	for _, i := range genopts.Images {
		logrus.Infof("Processing image: %s", i)
		tararchive := filepath.Join(tmpdir, uuid.New().String()+".tar")
		if err := spdx.PullImagesToArchive(i, tararchive); err != nil {
			return nil, errors.Wrapf(err, "writing image %s to file", i)
		}
		p, err := spdx.PackageFromImageTarball(tararchive, &TarballOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "generating tarball package")
		}
		ref, err := name.ParseReference(i)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing image reference %q", i)
		}

		// Grab the package data from wither the tag or, if it's a digest,
		// from parsing the digest
		tag, ok := ref.(name.Tag)
		if ok {
			p.Name = tag.RepositoryStr()
			p.DownloadLocation = tag.Name()
			p.Version = tag.Identifier()
		} else {
			dgst, ok := ref.(name.Digest)
			if ok {
				p.Version = dgst.DigestStr()
				p.Name = dgst.RepositoryStr()
				p.DownloadLocation = dgst.Name()
			}
		}
		if err := doc.AddPackage(p); err != nil {
			return nil, errors.Wrap(err, "adding package to document")
		}
	}

	for _, tb := range genopts.Tarballs {
		logrus.Infof("Processing tarball %s", tb)
		p, err := spdx.PackageFromImageTarball(tb, &TarballOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "generating tarball package")
		}
		if err := doc.AddPackage(p); err != nil {
			return nil, errors.Wrap(err, "adding package to document")
		}
	}

	for _, f := range genopts.Files {
		logrus.Infof("Processing file %s", f)
		f, err := spdx.FileFromPath(f)
		if err != nil {
			return nil, errors.Wrap(err, "adding file")
		}
		if err := doc.AddFile(f); err != nil {
			return nil, errors.Wrap(err, "adding file to document")
		}
	}
	return doc, nil
}

// WriteDoc renders the document to a file
func (builder defaultDocBuilderImpl) WriteDoc(doc *Document, path string) error {
	markup, err := doc.Render()
	if err != nil {
		return errors.Wrap(err, "generating document markup")
	}
	logrus.Infof("writing document to %s", path)
	return errors.Wrap(
		os.WriteFile(path, []byte(markup), os.FileMode(0o644)),
		"writing document markup to file",
	)
}

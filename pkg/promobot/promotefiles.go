/*
Copyright 2019 The Kubernetes Authors.

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

// nolint[lll]
package promobot

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	api "k8s.io/release/pkg/api/files"
	"k8s.io/release/pkg/filepromoter"
)

// PromoteFilesOptions holds the flag-values for a file promotion
type PromoteFilesOptions struct {
	// FilestoresPath is the path to the manifest file containing the filestores section
	FilestoresPath string

	// FilesPath specifies a path to manifest files containing the files section.
	FilesPath string

	// DryRun (if set) will not perform operations, but print them instead
	DryRun bool

	// UseServiceAccount must be true, for service accounts to be used
	// This gives some protection against a hostile manifest.
	UseServiceAccount bool

	// Out is the destination for "normal" output (such as dry-run)
	Out io.Writer
}

// PopulateDefaults sets the default values for PromoteFilesOptions
func (o *PromoteFilesOptions) PopulateDefaults() {
	o.DryRun = true
	o.UseServiceAccount = false
	o.Out = os.Stdout
}

// RunPromoteFiles executes a file promotion command
// nolint[gocyclo]
func RunPromoteFiles(ctx context.Context, options PromoteFilesOptions) error {
	manifest, err := ReadManifest(options)
	if err != nil {
		return err
	}

	if options.DryRun {
		fmt.Fprintf(
			options.Out,
			"********** START (DRY RUN) **********\n")
	} else {
		fmt.Fprintf(
			options.Out,
			"********** START **********\n")
	}

	promoter := &filepromoter.ManifestPromoter{
		Manifest:          manifest,
		UseServiceAccount: options.UseServiceAccount,
	}

	ops, err := promoter.BuildOperations(ctx)
	if err != nil {
		return fmt.Errorf(
			"error building operations: %v",
			err)
	}

	// So that we can support future parallel execution, an error
	// in one operation does not prevent us attempting the
	// remaining operations
	var errors []error
	for _, op := range ops {
		if _, err := fmt.Fprintf(options.Out, "%v\n", op); err != nil {
			errors = append(errors, fmt.Errorf(
				"error writing to output: %v", err))
		}

		if !options.DryRun {
			if err := op.Run(ctx); err != nil {
				logrus.Warnf("error copying file: %v", err)
				errors = append(errors, err)
			}
		}
	}

	if len(errors) != 0 {
		fmt.Fprintf(
			options.Out,
			"********** FINISHED WITH ERRORS **********\n")
		for _, err := range errors {
			fmt.Fprintf(options.Out, "%v\n", err)
		}

		return errors[0]
	}

	if options.DryRun {
		fmt.Fprintf(
			options.Out,
			"********** FINISHED (DRY RUN) **********\n")
	} else {
		fmt.Fprintf(
			options.Out,
			"********** FINISHED **********\n")
	}

	return nil
}

// ReadManifest reads a manifest.
func ReadManifest(options PromoteFilesOptions) (*api.Manifest, error) {
	merged := &api.Manifest{}

	filestores, err := readFilestores(options.FilestoresPath)
	if err != nil {
		return nil, err
	}
	merged.Filestores = filestores

	files, err := readFiles(options.FilesPath)
	if err != nil {
		return nil, err
	}
	merged.Files = files

	// Validate the merged manifest
	if err := merged.Validate(); err != nil {
		return nil, fmt.Errorf("error validating merged manifest: %v", err)
	}

	return merged, nil
}

// readFilestores reads a filestores manifest
func readFilestores(p string) ([]api.Filestore, error) {
	if p == "" {
		return nil, fmt.Errorf("FilestoresPath is required")
	}

	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest %q: %v", p, err)
	}

	manifest, err := api.ParseManifest(b)
	if err != nil {
		return nil, fmt.Errorf("error parsing manifest %q: %v", p, err)
	}

	if len(manifest.Files) != 0 {
		return nil, xerrors.Errorf(
			"files should not be present in filestore manifest %q",
			p)
	}

	return manifest.Filestores, nil
}

// readFiles reads and merges the file manifests from the file or directory filesPath
func readFiles(filesPath string) ([]api.File, error) {
	// We first list and sort the paths, for a consistent ordering
	var paths []string
	err := filepath.Walk(filesPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		paths = append(paths, p)
		return nil
	})
	if err != nil {
		return nil, xerrors.Errorf("error listing file manifests: %w", err)
	}

	sort.Strings(paths)

	var files []api.File
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, xerrors.Errorf("error reading file %q: %w", p, err)
		}

		manifest, err := api.ParseManifest(b)
		if err != nil {
			return nil, xerrors.Errorf("error parsing manifest %q: %v", p, err)
		}

		if len(manifest.Filestores) != 0 {
			return nil, xerrors.Errorf("filestores should not be present in manifest %q", p)
		}

		files = append(files, manifest.Files...)
	}

	return files, nil
}

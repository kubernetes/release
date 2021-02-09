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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

package license

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	licenseclassifier "github.com/google/licenseclassifier/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/release/v1/pkg/util"
)

const (
	licenseFilanameRe    = `(?i).*license.*`
	defaultCacheSubDir   = "cache"
	defaultLicenseSubDir = "licenses"
)

// Reader is an object that finds and interprets license files
type Reader struct {
	impl    ReaderImplementation
	Options *ReaderOptions
}

// SetImplementation sets the implementation that the license reader will use
func (r *Reader) SetImplementation(i ReaderImplementation) error {
	r.impl = i
	return errors.Wrap(
		r.impl.Initialize(r.Options),
		"initializing the reader implementation",
	)
}

// NewReader returns a license reader with the default options
func NewReader() (*Reader, error) {
	return NewReaderWithOptions(DefaultReaderOptions)
}

// NewReaderWithOptions returns a new license reader with the specified options
func NewReaderWithOptions(opts *ReaderOptions) (r *Reader, err error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "validating reader options")
	}
	r = &Reader{
		Options: opts,
	}

	if err := r.SetImplementation(&ReaderDefaultImpl{}); err != nil {
		return nil, errors.Wrap(err, "setting the reader implementation")
	}

	return r, nil
}

// ReaderOptions are the optional settings for the license reader
type ReaderOptions struct {
	ConfidenceThreshold float64 // Miniumum confidence to consider a license detected
	WorkDir             string  // Directory where the reader will store its data
	CacheDir            string  // Optional directory where the reader will store its downloads cache
	LicenseDir          string  // Optional dir to store and read the SPDX licenses from
}

// Validate checks the options to verify the are sane
func (ro *ReaderOptions) Validate() error {
	// if there is no working dir, create one
	if ro.WorkDir == "" {
		dir, err := ioutil.TempDir(os.TempDir(), "license-reader-")
		if err != nil {
			return errors.Wrap(err, "creating working dir")
		}
		ro.WorkDir = dir
		// Otherwise, check it exists
	} else if _, err := os.Stat(ro.WorkDir); err != nil {
		return errors.Wrap(err, "checking working directory")
	}

	// Check the cache directory
	if !util.Exists(ro.CacheDir) {
		if ro.CacheDir == "" {
			if err := os.MkdirAll(
				filepath.Join(ro.WorkDir, defaultCacheSubDir), os.FileMode(0o755),
			); err != nil {
				return errors.Wrap(err, "creating cache directory")
			}
		} else {
			return errors.New("specified cache directory does not exist")
		}
	}

	// Check the licenses directory
	if !util.Exists(ro.LicenseDir) {
		if ro.LicenseDir == "" {
			if err := os.MkdirAll(
				filepath.Join(ro.WorkDir, defaultLicenseSubDir), os.FileMode(0o755),
			); err != nil {
				return errors.Wrap(err, "creating licenses directory")
			}
		} else {
			return errors.New("specified licenses directory does not exist")
		}
	}

	// TODO check dirs
	return nil
}

// CachePath return the full path to the downloads cache
func (ro *ReaderOptions) CachePath() string {
	if ro.CacheDir != "" {
		return ro.CacheDir
	}

	return filepath.Join(ro.WorkDir, defaultCacheSubDir)
}

// LicensesPath return the full path to the downloads cache
func (ro *ReaderOptions) LicensesPath() string {
	if ro.LicenseDir != "" {
		return ro.LicenseDir
	}

	return filepath.Join(ro.WorkDir, defaultLicenseSubDir)
}

// DefaultReaderOptions is the default set of options for the classifier
var DefaultReaderOptions = &ReaderOptions{
	ConfidenceThreshold: 0.9,
}

// ReadLicenses returns an array of all licenses found in the specified path
func (r *Reader) ReadLicenses(path string) (licenseList []ClassifyResult, unknownPaths []string, err error) {
	licenseFiles, err := r.impl.FindLicenseFiles(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "searching for license files")
	}

	licenseList, unknownPaths, err = r.impl.ClassifyLicenseFiles(licenseFiles)
	if err != nil {
		return nil, nil, errors.Wrap(err, "classifying found licenses")
	}
	return licenseList, unknownPaths, nil
}

// ClassifyResult abstracts the data resulting from a file classification
type ClassifyResult struct {
	File    string
	License *SPDXLicense
}

//counterfeiter:generate . ReaderImplementation

// ReaderImplementation implements the basic lifecycle of a license reader:
// initializes -> finds license files to scan -> classifies them to a SPDX license
type ReaderImplementation interface {
	Initialize(*ReaderOptions) error
	ClassifyLicenseFiles([]string) ([]ClassifyResult, []string, error)
	ClassifyFile(string) (string, []string, error)
	FindLicenseFiles(string) ([]string, error)
}

// ReaderDefaultImpl the default license reader imlementation, uses
// Google's cicense classifier
type ReaderDefaultImpl struct {
	lc   *licenseclassifier.Classifier
	spdx *SPDX
}

// Initialize checks the options and creates the needed objects
func (d *ReaderDefaultImpl) Initialize(opts *ReaderOptions) error {
	// Validate our options before startin
	if err := opts.Validate(); err != nil {
		return errors.Wrap(err, "validating the license reader options")
	}

	// Create the implementation's SPDX object
	spdxopts := DefaultSPDXOpts
	spdxopts.CacheDir = opts.CachePath()
	spdx, err := NewSPDXWithOptions(spdxopts)
	if err != nil {
		return errors.Wrap(err, "creating SPDX object")
	}
	d.spdx = spdx

	if err := d.spdx.LoadLicenses(); err != nil {
		return errors.Wrap(err, "loading licenses")
	}

	// Write the licenses to disk as th classifier will need them
	if err := spdx.WriteLicensesAsText(opts.LicensesPath()); err != nil {
		return errors.Wrap(err, "writing license data to disk")
	}

	// Create the implementation's classifier
	d.lc = licenseclassifier.NewClassifier(opts.ConfidenceThreshold)
	return errors.Wrap(d.lc.LoadLicenses(opts.LicensesPath()), "loading licenses at init")
}

// Classifier returns the license classifier
func (d *ReaderDefaultImpl) Classifier() *licenseclassifier.Classifier {
	return d.lc
}

// SPDX returns the reader's SPDX object
func (d *ReaderDefaultImpl) SPDX() *SPDX {
	return d.spdx
}

// ClassifyFile takes a file path and returns the most probable license tag
func (d *ReaderDefaultImpl) ClassifyFile(path string) (licenseTag string, moreTags []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return licenseTag, nil, errors.Wrap(err, "opening file for analysis")
	}
	defer file.Close()

	// Get the classsification
	matches, err := d.Classifier().MatchFrom(file)
	if len(matches) == 0 {
		logrus.Warn("File does not match a known license: " + path)
	}
	var highestConf float64
	moreTags = []string{}
	for _, match := range matches {
		if match.Confidence > highestConf {
			highestConf = match.Confidence
			licenseTag = match.Name
			moreTags = append(moreTags, match.Name)
		}
	}
	return licenseTag, []string{}, nil
}

// ClassifyLicenseFiles takes a list of paths and tries to find return all licenses found in it
func (d *ReaderDefaultImpl) ClassifyLicenseFiles(paths []string) (
	licenseList []ClassifyResult, unrecognizedPaths []string, err error) {
	// Run the files through the clasifier
	for _, f := range paths {
		label, _, err := d.ClassifyFile(f)
		if err != nil {
			return nil, unrecognizedPaths, errors.Wrap(err, "classifying file")
		}
		if label == "" {
			unrecognizedPaths = append(unrecognizedPaths, f)
			continue
		}
		// Get the license corresponding to the ID label
		license := d.spdx.GetLicense(label)
		if license == nil {
			return nil, unrecognizedPaths,
				errors.New(fmt.Sprintf("ID does not correspond to a valid license: %s", label))
		}
		// Apend to the return results
		licenseList = append(licenseList, ClassifyResult{f, license})
	}
	logrus.Infof(
		"License classifier recognized %d/%d (%d%%) os the files",
		len(licenseList), len(paths), (len(licenseList)/len(paths))*100,
	)
	return licenseList, unrecognizedPaths, nil
}

// FindLicenseFiles will scan a directory and return files that may be licenses
func (d *ReaderDefaultImpl) FindLicenseFiles(path string) ([]string, error) {
	logrus.Infof("Scanning %s for license files", path)
	licenseList := []string{}
	re := regexp.MustCompile(licenseFilanameRe)
	if err := filepath.Walk(path,
		func(path string, finfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Directories are ignored
			if finfo.IsDir() {
				return nil
			}

			// No go source files are considered
			if filepath.Ext(path) == ".go" {
				return nil
			}
			// Check if tehe file matches the license regexp
			if re.MatchString(filepath.Base(path)) {
				licenseList = append(licenseList, path)
			}
			return nil
		}); err != nil {
		return nil, errors.Wrap(err, "scanning the directory for license files")
	}
	logrus.Infof("%d license files found in directory", len(licenseList))
	return licenseList, nil
}

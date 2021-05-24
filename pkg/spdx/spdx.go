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
	"regexp"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-utils/util"
)

const (
	defaultDocumentAuthor   = "Kubernetes Release Managers (release-managers@kubernetes.io)"
	archiveManifestFilename = "manifest.json"
	spdxTempDir             = "spdx"
	spdxLicenseData         = spdxTempDir + "/licenses"
	spdxLicenseDlCache      = spdxTempDir + "/downloadCache"
	gitIgnoreFile           = ".gitignore"
	validNameCharsRe        = `[^a-zA-Z0-9-]+`
)

type SPDX struct {
	impl    spdxImplementation
	options *Options
}

func NewSPDX() *SPDX {
	return &SPDX{
		impl:    &spdxDefaultImplementation{},
		options: &defaultSPDXOptions,
	}
}

func (spdx *SPDX) SetImplementation(impl spdxImplementation) {
	spdx.impl = impl
}

type Options struct {
	AnalyzeLayers    bool
	NoGitignore      bool     // Do not read exclusions from gitignore file
	ProcessGoModules bool     // If true, spdx will check if dirs are go modules and analize the packages
	LicenseCacheDir  string   // Directory to cache SPDX license information
	IgnorePatterns   []string // Patterns to ignore when scanning file
}

func (spdx *SPDX) Options() *Options {
	return spdx.options
}

var defaultSPDXOptions = Options{
	LicenseCacheDir:  filepath.Join(os.TempDir(), spdxLicenseDlCache),
	AnalyzeLayers:    true,
	ProcessGoModules: true,
	IgnorePatterns:   []string{},
}

type ArchiveManifest struct {
	ConfigFilename string   `json:"Config"`
	RepoTags       []string `json:"RepoTags"`
	LayerFiles     []string `json:"Layers"`
}

// ImageOptions set of options for processing tar files
type TarballOptions struct {
	ExtractDir string // Directory where the docker tar archive will be extracted
}

// PackageFromDirectory indexes all files in a directory and builds a
//  SPDX package describing its contents
func (spdx *SPDX) PackageFromDirectory(dirPath string) (pkg *Package, err error) {
	fileList, err := spdx.impl.GetDirectoryTree(dirPath)
	if err != nil {
		return nil, errors.Wrap(err, "building directory tree")
	}
	reader, err := spdx.impl.LicenseReader(spdx.Options())
	if err != nil {
		return nil, errors.Wrap(err, "creating license reader")
	}
	licenseTag := ""
	lic, err := spdx.impl.GetDirectoryLicense(reader, dirPath, spdx.Options())
	if err != nil {
		return nil, errors.Wrap(err, "scanning directory for licenses")
	}
	if lic == nil {
		logrus.Warn(err, "Licenseclassifier could not find a license for directory")
	} else {
		licenseTag = lic.LicenseID
	}

	// Build a list of patterns from those found in the .gitignore file and
	// posssibly others passed in the options:
	patterns, err := spdx.impl.IgnorePatterns(
		dirPath, spdx.Options().IgnorePatterns, spdx.Options().NoGitignore,
	)
	if err != nil {
		return nil, errors.Wrap(err, "building ignore patterns list")
	}

	// Apply the ignore patterns to the list of files
	fileList = spdx.impl.ApplyIgnorePatterns(fileList, patterns)

	pkg = NewPackage()
	pkg.FilesAnalyzed = true
	pkg.Name = filepath.Base(dirPath)
	// If the package file will result in an empty ID, generate one
	reg := regexp.MustCompile(validNameCharsRe)
	if reg.ReplaceAllString(pkg.Name, "") == "" {
		pkg.Name = uuid.NewString()
	}
	pkg.LicenseConcluded = licenseTag

	// todo: parallellize
	for _, path := range fileList {
		f := NewFile()
		f.Name = path
		f.FileName = path
		f.SourceFile = filepath.Join(dirPath, path)
		lic, err := reader.LicenseFromFile(f.SourceFile)
		if err != nil {
			return nil, errors.Wrap(err, "scanning file for license")
		}
		if lic != nil {
			f.LicenseInfoInFile = lic.LicenseID
		} else {
			f.LicenseInfoInFile = "NONE"
		}
		f.LicenseConcluded = licenseTag
		if err := f.ReadSourceFile(filepath.Join(dirPath, path)); err != nil {
			return nil, errors.Wrap(err, "checksumming file")
		}
		if err := pkg.AddFile(f); err != nil {
			return nil, errors.Wrapf(err, "adding %s as file to the spdx package", path)
		}
	}

	if util.Exists(filepath.Join(dirPath, GoModFileName)) && spdx.Options().ProcessGoModules {
		logrus.Info("Directory contains a go module. Scanning go packages")
		deps, err := spdx.impl.GetGoDependencies(dirPath, true)
		if err != nil {
			return nil, errors.Wrap(err, "scanning go packages")
		}
		for _, dep := range deps {
			if err := pkg.AddDependency(dep); err != nil {
				return nil, errors.Wrap(err, "adding go dependency")
			}
		}
	}

	// Add files into the package
	return pkg, nil
}

// PackageFromImageTarball returns a SPDX package from a tarball
func (spdx *SPDX) PackageFromImageTarball(
	tarPath string, opts *TarballOptions,
) (imagePackage *Package, err error) {
	logrus.Infof("Generating SPDX package from image tarball %s", tarPath)

	// Extract all files from tarfile
	opts.ExtractDir, err = spdx.impl.ExtractTarballTmp(tarPath)
	if err != nil {
		return nil, errors.Wrap(err, "extracting tarball to temp dir")
	}
	defer os.RemoveAll(opts.ExtractDir)

	// Read the archive manifest json:
	manifest, err := spdx.impl.ReadArchiveManifest(
		filepath.Join(opts.ExtractDir, archiveManifestFilename),
	)
	if err != nil {
		return nil, errors.Wrap(err, "while reading docker archive manifest")
	}

	if len(manifest.RepoTags) == 0 {
		return nil, errors.New("No RepoTags found in manifest")
	}

	if manifest.RepoTags[0] == "" {
		return nil, errors.New(
			"unable to add tar archive, manifest does not have a RepoTags entry",
		)
	}

	logrus.Infof("Package describes %s image", manifest.RepoTags[0])

	// Create the new SPDX package
	imagePackage = NewPackage()
	imagePackage.Options().WorkDir = opts.ExtractDir
	imagePackage.Name = manifest.RepoTags[0]

	logrus.Infof("Image manifest lists %d layers", len(manifest.LayerFiles))

	// Cycle all the layers from the manifest and add them as packages
	for _, layerFile := range manifest.LayerFiles {
		// Generate a package from a layer
		pkg, err := spdx.impl.PackageFromLayerTarBall(layerFile, opts)
		if err != nil {
			return nil, errors.Wrap(err, "building package from layer")
		}

		// If the option is enabled, scan the container layers
		if spdx.options.AnalyzeLayers {
			if err := spdx.AnalyzeImageLayer(filepath.Join(opts.ExtractDir, layerFile), pkg); err != nil {
				return nil, errors.Wrap(err, "scanning layer "+pkg.ID)
			}
		} else {
			logrus.Info("Not performing deep image analysis (opts.AnalyzeLayers = false)")
		}

		// Add the layer package to the image package
		if err := imagePackage.AddPackage(pkg); err != nil {
			return nil, errors.Wrap(err, "adding layer to image package")
		}
	}

	// return the finished package
	return imagePackage, nil
}

// FileFromPath creates a File object from a path
func (spdx *SPDX) FileFromPath(filePath string) (*File, error) {
	if !util.Exists(filePath) {
		return nil, errors.New("file does not exist")
	}
	f := NewFile()
	if err := f.ReadSourceFile(filePath); err != nil {
		return nil, errors.Wrap(err, "creating file from path")
	}
	return f, nil
}

// AnalyzeLayer uses the collection of image analyzers to see if
//  it matches a known image from which a spdx package can be
//  enriched with more information
func (spdx *SPDX) AnalyzeImageLayer(layerPath string, pkg *Package) error {
	return NewImageAnalyzer().AnalyzeLayer(layerPath, pkg)
}

// ExtractTarballTmp extracts a tarball to a temp file
func (spdx *SPDX) ExtractTarballTmp(tarPath string) (tmpDir string, err error) {
	return spdx.impl.ExtractTarballTmp(tarPath)
}

// PullImagesToArchive
func (spdx *SPDX) PullImagesToArchive(reference, path string) error {
	return spdx.impl.PullImagesToArchive(reference, path)
}

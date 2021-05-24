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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/vcs"
	"k8s.io/release/pkg/license"
	"sigs.k8s.io/release-utils/util"
)

const (
	downloadDir   = spdxTempDir + "/gomod-scanner"
	GoModFileName = "go.mod"
)

// NewGoModule returns a new go module from the specified path
func NewGoModuleFromPath(path string) (*GoModule, error) {
	mod := NewGoModule()
	mod.opts.Path = path
	if err := mod.Open(); err != nil {
		return nil, errors.Wrap(err, "opening new module path")
	}
	return mod, nil
}

func NewGoModule() *GoModule {
	return &GoModule{
		opts: &GoModuleOptions{},
		impl: &GoModDefaultImpl{},
	}
}

// GoModule abstracts the go module data of a project
type GoModule struct {
	impl     GoModImplementation
	GoMod    *modfile.File
	opts     *GoModuleOptions // Options
	Packages []*GoPackage     // maps of package download locations
}

type GoModuleOptions struct {
	Path string // Path to the dir where go.mod resides
}

// GoPackage basic pkg data we need
type GoPackage struct {
	ImportPath string
	Revision   string
	LocalDir   string
	LicenseID  string
}

// SPDXPackage builds a spdx package from the go package data
func (pkg *GoPackage) ToSPDXPackage() (*Package, error) {
	repo, err := vcs.RepoRootForImportPath(pkg.ImportPath, true)
	if err != nil {
		return nil, errors.Wrap(err, "building repository from package import path")
	}
	spdxPackage := NewPackage()
	spdxPackage.Name = pkg.ImportPath
	spdxPackage.DownloadLocation = repo.Repo
	spdxPackage.LicenseConcluded = pkg.LicenseID
	spdxPackage.Version = pkg.Revision
	return spdxPackage, nil
}

type GoModImplementation interface {
	OpenModule(*GoModuleOptions) (*modfile.File, error)
	BuildPackageList(*modfile.File) ([]*GoPackage, error)
	DownloadPackage(*GoPackage, *GoModuleOptions, bool) error
	RemoveDownloads([]*GoPackage) error
	LicenseReader() (*license.Reader, error)
	ScanPackageLicense(*GoPackage, *license.Reader, *GoModuleOptions) error
}

// Initializes a go module from the specified path
func (mod *GoModule) Open() error {
	gomod, err := mod.impl.OpenModule(mod.opts)
	if err != nil {
		return errors.Wrap(err, "opening module")
	}
	mod.GoMod = gomod

	// Build the package list
	pkgs, err := mod.impl.BuildPackageList(mod.GoMod)
	if err != nil {
		return errors.Wrap(err, "building module package list")
	}
	mod.Packages = pkgs
	return nil
}

// RemoveDownloads cleans all downloads
func (mod *GoModule) RemoveDownloads() error {
	return mod.impl.RemoveDownloads(mod.Packages)
}

// DownloadPackages downloads all the module's packages to the local disk
func (mod *GoModule) DownloadPackages() error {
	logrus.Infof("Downloading source code for %d packages", len(mod.Packages))
	if mod.Packages == nil {
		return errors.New("Unable to download packages, package list is nil")
	}

	for _, pkg := range mod.Packages {
		if err := mod.impl.DownloadPackage(pkg, mod.opts, true); err != nil {
			return err
		}
	}
	return nil
}

// ScanLicenses scans the licenses and populats the fields
func (mod *GoModule) ScanLicenses() error {
	if mod.Packages == nil {
		return errors.New("Unable to scan lincese files, package list is nil")
	}

	reader, err := mod.impl.LicenseReader()
	if err != nil {
		return errors.Wrap(err, "creating license scanner")
	}

	// Do a quick re-check for missing downloads
	// todo: paralelize this. urgently.
	for _, pkg := range mod.Packages {
		// Call download with no force in case local data is missing
		if err := mod.impl.DownloadPackage(pkg, mod.opts, false); err != nil {
			// If we're unable to download the module we dont treat it as
			// fatal, package will remain without license info but we go
			// on scanning the rest of the packages.
			logrus.Error(err)
			continue
		}

		if err := mod.impl.ScanPackageLicense(pkg, reader, mod.opts); err != nil {
			return errors.Wrapf(err, "scanning package %s for licensing info", pkg.ImportPath)
		}
	}

	return nil
}

type GoModDefaultImpl struct {
	licenseReader *license.Reader
}

// OpenModule opens the go,mod file for the module and parses it
func (di *GoModDefaultImpl) OpenModule(opts *GoModuleOptions) (*modfile.File, error) {
	modData, err := os.ReadFile(filepath.Join(opts.Path, GoModFileName))
	if err != nil {
		return nil, errors.Wrap(err, "reading module's go.mod file")
	}
	gomod, err := modfile.ParseLax("file", modData, nil)
	if err != nil {
		return nil, errors.Wrap(err, "reading go.mod")
	}
	logrus.Infof(
		"Parsed go.mod file for %s, found %d packages",
		gomod.Module.Mod.Path,
		len(gomod.Require),
	)
	return gomod, nil
}

// BuildPackageList builds a slice of packages to assign to the module
func (di *GoModDefaultImpl) BuildPackageList(gomod *modfile.File) ([]*GoPackage, error) {
	pkgs := []*GoPackage{}
	for _, req := range gomod.Require {
		pkgs = append(pkgs, &GoPackage{
			ImportPath: req.Mod.Path,
			Revision:   req.Mod.Version,
		})
	}
	return pkgs, nil
}

// DownloadPackage takes a pkg, downloads it from its src and sets
//  the download dir in the LocalDir field
func (di *GoModDefaultImpl) DownloadPackage(pkg *GoPackage, opts *GoModuleOptions, force bool) error {
	logrus.Infof("Downloading package %s@%s", pkg.ImportPath, pkg.Revision)
	repo, err := vcs.RepoRootForImportPath(pkg.ImportPath, true)
	if err != nil {
		return errors.Wrapf(err, "Fetching package %s from %s", pkg.ImportPath, repo.Repo)
	}

	if pkg.LocalDir != "" && util.Exists(pkg.LocalDir) && !force {
		logrus.Infof("Not downloading %s as it already has local data", pkg.ImportPath)
		return nil
	}

	if !util.Exists(filepath.Join(os.TempDir(), downloadDir)) {
		if err := os.MkdirAll(
			filepath.Join(os.TempDir(), downloadDir), os.FileMode(0o755),
		); err != nil {
			return errors.Wrap(err, "creating parent tmpdir")
		}
	}

	// Create tempdir
	tmpDir, err := os.MkdirTemp(filepath.Join(os.TempDir(), downloadDir), "package-download-")
	if err != nil {
		return errors.Wrap(err, "creating temporary dir")
	}
	// Create a clone of the module repo at the revision
	rev := pkg.Revision
	m := regexp.MustCompile(`v\d+\.\d+\.\d+-[0-9.]+-([a-f0-9]+)`).FindStringSubmatch(pkg.Revision)
	if len(m) > 1 {
		rev = m[1]
		logrus.Infof("Using commit %s as revision for download", rev)
	}
	if err := repo.VCS.CreateAtRev(tmpDir, repo.Repo, rev); err != nil {
		return errors.Wrapf(err, "creating local clone of %s", repo.Repo)
	}

	logrus.Infof("Go Package %s (rev %s) downloaded to %s", pkg.ImportPath, pkg.Revision, tmpDir)
	pkg.LocalDir = tmpDir
	return nil
}

// RemoveDownloads takes a list of packages and remove its downloads
func (di *GoModDefaultImpl) RemoveDownloads(packageList []*GoPackage) error {
	for _, pkg := range packageList {
		if pkg.ImportPath != "" && util.Exists(pkg.LocalDir) {
			if err := os.RemoveAll(pkg.ImportPath); err != nil {
				return errors.Wrap(err, "removing package data")
			}
		}
	}
	return nil
}

// LicenseReader returns a license reader
func (di *GoModDefaultImpl) LicenseReader() (*license.Reader, error) {
	if di.licenseReader == nil {
		opts := license.DefaultReaderOptions
		opts.CacheDir = filepath.Join(os.TempDir(), spdxLicenseDlCache)
		opts.LicenseDir = filepath.Join(os.TempDir(), spdxLicenseData)
		if !util.Exists(opts.CacheDir) {
			if err := os.MkdirAll(opts.CacheDir, os.FileMode(0o755)); err != nil {
				return nil, errors.Wrap(err, "creating dir")
			}
		}
		reader, err := license.NewReaderWithOptions(opts)
		if err != nil {
			return nil, errors.Wrap(err, "creating reader")
		}

		di.licenseReader = reader
	}
	return di.licenseReader, nil
}

// ScanPackageLicense scans a package for licensing info
func (di *GoModDefaultImpl) ScanPackageLicense(
	pkg *GoPackage, reader *license.Reader, opts *GoModuleOptions) error {
	licenselist, _, err := reader.ReadLicenses(pkg.LocalDir)
	if err != nil {
		return errors.Wrapf(err, "scanning package %s for licensing information", pkg.ImportPath)
	}

	if len(licenselist) > 1 {
		logrus.Warnf("Package %s has %d licenses, picking the first", pkg.ImportPath, len(licenselist))
	}

	if len(licenselist) != 0 {
		logrus.Infof(
			"Package %s license is %s", pkg.ImportPath,
			licenselist[0].License.LicenseID,
		)
		pkg.LicenseID = licenselist[0].License.LicenseID
	} else {
		logrus.Infof("Could not find licensing information for package %s", pkg.ImportPath)
	}
	return nil
}

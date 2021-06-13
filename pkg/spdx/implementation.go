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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"archive/tar"
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/license"
	"sigs.k8s.io/release-utils/util"
)

//counterfeiter:generate . spdxImplementation

type spdxImplementation interface {
	ExtractTarballTmp(string) (string, error)
	ReadArchiveManifest(string) (*ArchiveManifest, error)
	PullImagesToArchive(string, string) error
	PackageFromLayerTarBall(string, *TarballOptions) (*Package, error)
	GetDirectoryTree(string) ([]string, error)
	IgnorePatterns(string, []string, bool) ([]gitignore.Pattern, error)
	ApplyIgnorePatterns([]string, []gitignore.Pattern) []string
	GetGoDependencies(string, *Options) ([]*Package, error)
	GetDirectoryLicense(*license.Reader, string, *Options) (*license.License, error)
	LicenseReader(*Options) (*license.Reader, error)
}

type spdxDefaultImplementation struct{}

// ExtractTarballTmp extracts a tarball to a temporary directory
func (di *spdxDefaultImplementation) ExtractTarballTmp(tarPath string) (tmpDir string, err error) {
	tmpDir, err = os.MkdirTemp(os.TempDir(), "spdx-tar-extract-")
	if err != nil {
		return tmpDir, errors.Wrap(err, "creating temporary directory for tar extraction")
	}

	// Open the tar file
	f, err := os.Open(tarPath)
	if err != nil {
		return tmpDir, errors.Wrap(err, "opening tarball")
	}

	tr := tar.NewReader(f)
	numFiles := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return tmpDir, errors.Wrap(err, "reading the image tarfile")
		}

		if hdr.FileInfo().IsDir() {
			continue
		}

		if strings.HasPrefix(filepath.Base(hdr.FileInfo().Name()), ".wh") {
			logrus.Info("Skipping extraction of whiteout file")
			continue
		}

		if err := os.MkdirAll(
			filepath.Join(tmpDir, filepath.Dir(hdr.Name)), os.FileMode(0o755),
		); err != nil {
			return tmpDir, errors.Wrap(err, "creating image directory structure")
		}

		targetFile := filepath.Join(tmpDir, hdr.Name)
		f, err := os.Create(targetFile)
		if err != nil {
			return tmpDir, errors.Wrap(err, "creating image layer file")
		}
		defer f.Close()

		if _, err := io.Copy(f, tr); err != nil {
			return tmpDir, errors.Wrap(err, "extracting image data")
		}
		numFiles++
	}
	logrus.Infof("Successfully extracted %d files from image tarball %s", numFiles, tarPath)
	return tmpDir, err
}

// readArchiveManifest extracts the manifest json from an image tar
//    archive and returns the data as a struct
func (di *spdxDefaultImplementation) ReadArchiveManifest(manifestPath string) (manifest *ArchiveManifest, err error) {
	// Check that we have the archive manifest.json file
	if !util.Exists(manifestPath) {
		return manifest, errors.New("unable to find manifest file " + manifestPath)
	}

	// Parse the json file
	manifestData := []ArchiveManifest{}
	manifestJSON, err := os.ReadFile(manifestPath)
	if err != nil {
		return manifest, errors.Wrap(err, "unable to read from tarfile")
	}
	if err := json.Unmarshal(manifestJSON, &manifestData); err != nil {
		fmt.Println(string(manifestJSON))
		return manifest, errors.Wrap(err, "unmarshalling image manifest")
	}
	return &manifestData[0], nil
}

// PullImagesToArchive takes an image reference (a tag or a digest)
// and writes it into a docker tar archive in path
func (di *spdxDefaultImplementation) PullImagesToArchive(referenceString, path string) error {
	// Parse the string to get a reference (tag or digest)
	ref, err := name.ParseReference(referenceString)
	if err != nil {
		return errors.Wrapf(err, "parsing reference %s", referenceString)
	}

	// Build an image from the reference
	img, err := remote.Image(ref)
	if err != nil {
		return errors.Wrap(err, "getting image")
	}

	// This algo comes from crane:
	// Try to cast the reference as a tag:
	tag, ok := ref.(name.Tag)
	// if it fails
	if !ok {
		// .. and it is a digest
		d, ok := ref.(name.Digest)
		if !ok {
			return fmt.Errorf("reference is not a tag or digest")
		}
		// We add a mock tag
		tag = d.Repository.Tag("from-digest") // Append digest here?
	}

	return tarball.MultiWriteToFile(path, map[name.Tag]v1.Image{tag: img})
}

// PackageFromLayerTarBall builds a SPDX package from an image
//  tarball
func (di *spdxDefaultImplementation) PackageFromLayerTarBall(
	layerFile string, opts *TarballOptions,
) (*Package, error) {
	logrus.Infof("Generating SPDX package from layer in %s", layerFile)

	pkg := NewPackage()
	pkg.options.WorkDir = opts.ExtractDir
	if err := pkg.ReadSourceFile(filepath.Join(opts.ExtractDir, layerFile)); err != nil {
		return nil, errors.Wrap(err, "reading source file")
	}
	// Build the pkg name from its internal path
	h := sha1.New()
	if _, err := h.Write([]byte(layerFile)); err != nil {
		return nil, errors.Wrap(err, "hashing file path")
	}
	pkg.Name = fmt.Sprintf("%x", h.Sum(nil))

	return pkg, nil
}

// GetDirectoryTree traverses a directory and return a slice of strings with all files
func (di *spdxDefaultImplementation) GetDirectoryTree(dirPath string) ([]string, error) {
	fileList := []string{}

	if err := fs.WalkDir(os.DirFS(dirPath), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if d.Type() == os.ModeSymlink {
			return nil
		}

		fileList = append(fileList, path)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "buiding directory tree")
	}
	return fileList, nil
}

// IgnorePatterns return a list of gitignore patterns
func (di *spdxDefaultImplementation) IgnorePatterns(
	dirPath string, extraPatterns []string, skipGitIgnore bool,
) ([]gitignore.Pattern, error) {
	patterns := []gitignore.Pattern{}
	for _, s := range extraPatterns {
		patterns = append(patterns, gitignore.ParsePattern(s, nil))
	}

	if skipGitIgnore {
		logrus.Debug("Not using patterns in .gitignore")
		return patterns, nil
	}

	if util.Exists(filepath.Join(dirPath, gitIgnoreFile)) {
		f, err := os.Open(filepath.Join(dirPath, gitIgnoreFile))
		if err != nil {
			return nil, errors.Wrap(err, "opening gitignore file")
		}
		defer f.Close()

		// When using .gitignore files, we alwas add the .git directory
		// to match git's behavior
		patterns = append(patterns, gitignore.ParsePattern(".git/", nil))

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			s := scanner.Text()
			if !strings.HasPrefix(s, "#") && len(strings.TrimSpace(s)) > 0 {
				logrus.Debugf("Loaded .gitignore pattern: >>%s<<", s)
				patterns = append(patterns, gitignore.ParsePattern(s, nil))
			}
		}
	}

	logrus.Debugf(
		"Loaded %d patterns from .gitignore (+ %d extra) at root of directory", len(patterns), len(extraPatterns),
	)
	return patterns, nil
}

// ApplyIgnorePatterns applies the gitignore patterns to a list of files, removing matched
func (di *spdxDefaultImplementation) ApplyIgnorePatterns(
	fileList []string, patterns []gitignore.Pattern,
) (filteredList []string) {
	// We will return a new file list
	filteredList = []string{}

	// Build the new gitignore matcher
	matcher := gitignore.NewMatcher(patterns)

	// Cycle all files, removing those matched:
	for _, file := range fileList {
		if matcher.Match(strings.Split(file, string(filepath.Separator)), false) {
			logrus.Debugf("File ignored by .gitignore: %s", file)
		} else {
			filteredList = append(filteredList, file)
		}
	}
	return filteredList
}

// GetGoDependencies opens a Go module and directory and returns the
// dependencies as SPDX packages.
func (di *spdxDefaultImplementation) GetGoDependencies(
	path string, opts *Options,
) (spdxPackages []*Package, err error) {
	// Open the directory as a go module:
	mod, err := NewGoModuleFromPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "creating a mod from the specified path")
	}
	mod.Options().OnlyDirectDeps = opts.OnlyDirectDeps
	mod.Options().ScanLicenses = opts.ScanLicenses

	// Open the module
	if err := mod.Open(); err != nil {
		return nil, errors.Wrap(err, "opening new module path")
	}

	defer func() { err = mod.RemoveDownloads() }()

	if opts.ScanLicenses {
		if err := mod.ScanLicenses(); err != nil {
			return nil, errors.Wrap(err, "scanning go module licenses")
		}
	}

	spdxPackages = []*Package{}
	for _, goPkg := range mod.Packages {
		spdxPkg, err := goPkg.ToSPDXPackage()
		if err != nil {
			return nil, errors.Wrap(err, "converting go module to spdx package")
		}
		spdxPackages = append(spdxPackages, spdxPkg)
	}

	return spdxPackages, err
}

func (di *spdxDefaultImplementation) LicenseReader(spdxOpts *Options) (*license.Reader, error) {
	opts := license.DefaultReaderOptions
	opts.CacheDir = spdxOpts.LicenseCacheDir
	opts.LicenseDir = spdxOpts.LicenseData
	// Create the new reader
	reader, err := license.NewReaderWithOptions(opts)
	if err != nil {
		return nil, errors.Wrap(err, "creating reusable license reader")
	}
	return reader, nil
}

// GetDirectoryLicense takes a path and scans
// the files in it to determine licensins information
func (di *spdxDefaultImplementation) GetDirectoryLicense(
	reader *license.Reader, path string, spdxOpts *Options,
) (*license.License, error) {
	// Perhaps here we should take into account thre results
	licenselist, _, err := reader.ReadLicenses(path)
	if err != nil {
		return nil, errors.Wrap(err, "scanning directory for licensing information")
	}
	if len(licenselist) == 0 {
		logrus.Warn("No license info could be determined, SPDX SBOM will not be valid")
		return nil, nil
	}
	logrus.Infof("Determined license %s for directory %s", licenselist[0].License.LicenseID, path)

	if len(licenselist) > 1 {
		logrus.Warnf("Found %d licenses in directory, picking the first", len(licenselist))
	}
	return licenselist[0].License, nil
}

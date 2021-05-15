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

package license

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NewSPDX returns a SPDX object with the default options
func NewSPDX() (spdx *SPDX, err error) {
	return NewSPDXWithOptions(DefaultSPDXOpts)
}

// NewSPDXWithOptions returns a SPDX object with the specified options
func NewSPDXWithOptions(opts *SPDXOptions) (spdx *SPDX, err error) {
	// Create the license Downloader
	doptions := DefaultDownloaderOpts
	doptions.CacheDir = opts.CacheDir
	downloader, err := NewDownloaderWithOptions(doptions)
	if err != nil {
		return nil, errors.Wrap(err, "creating downloader")
	}
	spdx = &SPDX{
		Downloader: downloader,
		Options:    DefaultSPDXOpts,
	}
	if err := spdx.Options.Validate(); err != nil {
		return nil, err
	}
	return spdx, nil
}

// SPDX is an objec to interact with licenses and manifest creation
type SPDX struct {
	Downloader *Downloader      // License Downloader
	Licenses   *SPDXLicenseList // List of licenses
	Options    *SPDXOptions     // SPDX Options
}

// SPDXOptions are the spdx settings
type SPDXOptions struct {
	CacheDir string
}

// Validate checks the spdx options
func (o *SPDXOptions) Validate() error {
	return nil
}

// DefaultSPDXOpts are the predetermined settings. License and cache directories
// are in the temporary OS directory and are created if the do not exist
var DefaultSPDXOpts = &SPDXOptions{}

// SPDXLicenseList abstracts the list of licenses published by SPDX.org
type SPDXLicenseList struct {
	sync.RWMutex
	Version           string                 `json:"licenseListVersion"`
	ReleaseDateString string                 `json:"releaseDate "`
	LicenseData       []SPDXLicenseListEntry `json:"licenses"`
	Licenses          map[string]*SPDXLicense
}

// Add appends a license to the license list
func (list *SPDXLicenseList) Add(license *SPDXLicense) {
	list.Lock()
	defer list.Unlock()
	if list.Licenses == nil {
		list.Licenses = map[string]*SPDXLicense{}
	}
	list.Licenses[license.LicenseID] = license
}

// SPDXLicense is a license described in JSON
type SPDXLicense struct {
	IsDeprecatedLicenseID         bool     `json:"isDeprecatedLicenseId"`
	IsFsfLibre                    bool     `json:"isFsfLibre"`
	IsOsiApproved                 bool     `json:"isOsiApproved"`
	LicenseText                   string   `json:"licenseText"`
	StandardLicenseHeaderTemplate string   `json:"standardLicenseHeaderTemplate"`
	StandardLicenseTemplate       string   `json:"standardLicenseTemplate"`
	Name                          string   `json:"name"`
	LicenseID                     string   `json:"licenseId"`
	StandardLicenseHeader         string   `json:"standardLicenseHeader"`
	SeeAlso                       []string `json:"seeAlso"`
}

// WriteText writes the SPDX license text to a text file
func (license *SPDXLicense) WriteText(filePath string) error {
	return errors.Wrap(
		os.WriteFile(
			filePath, []byte(license.LicenseText), os.FileMode(0o644),
		), "while writing license to text file",
	)
}

// SPDXLicenseListEntry a license entry in the list
type SPDXLicenseListEntry struct {
	IsOsiApproved   bool     `json:"isOsiApproved"`
	IsDeprectaed    bool     `json:"isDeprecatedLicenseId"`
	Reference       string   `json:"reference"`
	DetailsURL      string   `json:"detailsUrl"`
	ReferenceNumber string   `json:"referenceNumber"`
	Name            string   `json:"name"`
	LicenseID       string   `json:"licenseId"`
	SeeAlso         []string `json:"seeAlso"`
}

// LoadLicenses reads the license data from the downloader
func (spdx *SPDX) LoadLicenses() error {
	logrus.Info("Loading license data from downloader")
	licenses, err := spdx.Downloader.GetLicenses()
	if err != nil {
		return errors.Wrap(err, "getting licenses from downloader")
	}
	spdx.Licenses = licenses
	logrus.Infof("SPDX: Got %d licenses from downloader", len(licenses.Licenses))
	return nil
}

// WriteLicensesAsText writes the SPDX license collection to text files
func (spdx *SPDX) WriteLicensesAsText(targetDir string) error {
	logrus.Info("Writing SPDX licenses to " + targetDir)
	if spdx.Licenses.Licenses == nil {
		return errors.New("unable to write licenses, they have not been loaded yet")
	}
	for _, l := range spdx.Licenses.Licenses {
		if err := l.WriteText(filepath.Join(targetDir, l.LicenseID+".txt")); err != nil {
			return errors.Wrapf(err, "while writing license %s", l.LicenseID)
		}
	}
	return nil
}

// GetLicense returns a license struct from its SPDX ID label
func (spdx *SPDX) GetLicense(label string) *SPDXLicense {
	if lic, ok := spdx.Licenses.Licenses[label]; ok {
		return lic
	}
	logrus.Warn("Label %s is not an ID of a known license " + label)
	return nil
}

// ParseSPDXLicense parses a SPDX license from its JSON source
func ParseSPDXLicense(licenseJSON []byte) (license *SPDXLicense, err error) {
	license = &SPDXLicense{}
	if err := json.Unmarshal(licenseJSON, license); err != nil {
		return nil, errors.Wrap(err, "parsing SPDX licence")
	}
	return license, nil
}

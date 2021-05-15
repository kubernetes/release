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

package license_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/license"

	"k8s.io/release/pkg/license/licensefakes"
)

const testFullLicense = `
{
  "isDeprecatedLicenseId": false,
  "isFsfLibre": true,
  "licenseText": "Apache License\nVersion 2.0, January 2004\nhttp://www.apache.org/licenses/\n\nTERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION",
  "name": "Apache License 2.0",
  "licenseComments": "This license was released January 2004",
  "licenseId": "Apache-2.0",
  "standardLicenseHeader": "Copyright [yyyy] [name of copyright owner]\n\nLicensed under the Apache License, Version 2.0 (the \"License\");\n\nyou may not use this file except in compliance with the License.\n\nYou may obtain a copy of the License at\n\nhttp://www.apache.org/licenses/LICENSE-2.0\n\nUnless required by applicable law or agreed to in writing, software\n\ndistributed under the License is distributed on an \"AS IS\" BASIS,\n\nWITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n\nSee the License for the specific language governing permissions and\n\nlimitations under the License.",
  "crossRef": [{"isLive": true,"isValid": true,"isWayBackLink": false,"match": "true","url": "http://www.apache.org/licenses/LICENSE-2.0","order": 0,"timestamp": "2020-11-25 - 21:56:49"}],
  "seeAlso": [
    "http://www.apache.org/licenses/LICENSE-2.0",
    "https://opensource.org/licenses/Apache-2.0"
  ],
  "isOsiApproved": true
}
`

func TestISCatalogLoadLicenses(t *testing.T) {
	downloader := &license.Downloader{}
	// Create a SPDX to test
	spdx := &license.Catalog{
		Downloader: downloader,
		Options:    license.DefaultCatalogOpts,
	}

	for _, tc := range []struct {
		mustFail        bool
		dnLoaderReturns *license.List
		dnLoaderError   error
	}{
		{true, nil, errors.New("Some download error")},
		{false, &license.List{}, nil},
	} {
		impl := licensefakes.FakeDownloaderImplementation{}
		impl.GetLicensesReturns(tc.dnLoaderReturns, tc.dnLoaderError)
		downloader.SetImplementation(&impl)

		if tc.mustFail {
			require.NotNil(t, spdx.LoadLicenses())
		} else {
			require.Nil(t, spdx.LoadLicenses())
		}
	}
}

func TestUSPDXWriteLicensesAsText(t *testing.T) {
	testLicenseID := "test-license"
	downloader := &license.Downloader{}
	impl := licensefakes.FakeDownloaderImplementation{}
	impl.GetLicensesReturns(&license.List{
		Licenses: map[string]*license.License{
			testLicenseID: {LicenseID: testLicenseID, LicenseText: "Test"},
		},
	}, nil)
	downloader.SetImplementation(&impl)

	// Create a SPDX to test
	spdx := &license.Catalog{
		Downloader: downloader,
		Options:    license.DefaultCatalogOpts,
	}

	// Get the licenses from the fke downloader
	require.Nil(t, spdx.LoadLicenses())

	// Create a test directory
	tempdir, err := os.MkdirTemp("", "spdx-test-")
	require.Nil(t, err)
	defer func() { require.Nil(t, os.RemoveAll(tempdir)) }()

	// Check the call works:
	require.Nil(t, spdx.WriteLicensesAsText(tempdir))

	// Check we have one file
	require.Nil(t, CheckFileExists(t, filepath.Join(tempdir, testLicenseID+".txt")))
}

func TestUSPDXGetLicense(t *testing.T) {
	testLicenseID := "test-license"
	testLicenseContent := "Test license content"
	catalog := license.Catalog{
		Downloader: &license.Downloader{},
		List: &license.List{
			Licenses: map[string]*license.License{
				testLicenseID: {LicenseID: testLicenseID, LicenseText: testLicenseContent},
			},
		},
		Options: &license.CatalogOptions{},
	}

	testTicense := catalog.GetLicense(testLicenseID)
	require.NotNil(t, testTicense)
	require.Equal(t, testTicense.LicenseID, testLicenseID)
	require.Equal(t, testTicense.LicenseText, testLicenseContent)

	testTicense = catalog.GetLicense("invalid-license-id")
	require.Nil(t, testTicense)
}

func TestUSPDXLicenseListAdd(t *testing.T) {
	// Create a sample license
	licenseList := &license.List{}
	testLicense := &license.License{LicenseID: "test-license", LicenseText: "test text"}
	// Use the Add method to add it to the collection
	licenseList.Add(testLicense)
	// Retrieve the data from the struct
	l, exists := licenseList.Licenses[testLicense.LicenseID]
	// Verify
	require.NotNil(t, l)
	require.True(t, exists)
}

func TestISetImplementation(t *testing.T) {
	// Initialize runs the
	reader := license.Reader{}

	// Create a moxk implementation
	impl := licensefakes.FakeReaderImplementation{}

	// Initialization fails
	impl.InitializeReturns(nil)
	require.Nil(t, reader.SetImplementation(&impl))

	// Initialization works
	impl.InitializeReturns(errors.New("Mock init error"))
	require.NotNil(t, reader.SetImplementation(&impl))
}

// CheckFileExists checks if a file exists and is not empty
func CheckFileExists(t *testing.T, path string) error {
	finfo, err := os.Stat(path)
	require.Nil(t, err)
	require.False(t, finfo.IsDir())
	require.Greater(t, finfo.Size(), int64(0))
	return nil
}

func TestULicenseWriteText(t *testing.T) {
	testLicense := license.License{
		LicenseText: "Test license text",
		LicenseID:   "test-license",
	}

	path := filepath.Join(os.TempDir(), testLicense.LicenseID+"txt")

	// Write the license text to a file
	require.Nil(t, testLicense.WriteText(path))
	defer func() { os.Remove(path) }()

	require.Nil(t, CheckFileExists(t, path))
}

func TestParseSPDXLicense(t *testing.T) {
	testsLicense, err := license.ParseLicense([]byte(testFullLicense))
	require.Nil(t, err)
	require.NotNil(t, testsLicense)

	// Check one or two bits of the example json
	require.Equal(t, testsLicense.Name, "Apache License 2.0")
	require.Equal(t, testsLicense.LicenseID, "Apache-2.0")
}

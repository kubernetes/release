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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testDirPrefix   = "license-test-"
	testFullLicense = `
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
)

func TestCacheData(t *testing.T) {
	tempdir, err := os.MkdirTemp("", testDirPrefix)
	require.Nil(t, err)
	defer func() { require.Nil(t, os.RemoveAll(tempdir)) }()

	impl := DefaultDownloaderImpl{
		Options: &DownloaderOptions{
			EnableCache:       true,
			CacheDir:          tempdir,
			parallelDownloads: 0,
		},
	}
	// Get some testing data
	testData := []byte("Testing 1,2,3")
	testURL := "http://example.com/"

	// Test storing the data
	require.Nil(t, impl.cacheData(testURL, testData))

	// Now test getting the data back
	cachedData, err := impl.getCachedData(testURL)
	require.Nil(t, err)
	require.Equal(t, testData, cachedData)
}

func TestFindLicenseFiles(t *testing.T) {
	files := []string{
		"LICENSE", "LICENSE.txt", "LICENSE-APACHE2", "APACHE2-LICENSE", "license",
		"license.go", "README.md",
	}

	tempdir, err := os.MkdirTemp("", testDirPrefix)
	require.Nil(t, err)
	defer func() { require.Nil(t, os.RemoveAll(tempdir)) }()

	require.Nil(t, os.MkdirAll(filepath.Join(tempdir, "/some/sub/dir"), os.FileMode(0o755)))
	fileData := []byte("bleh")
	for _, sub := range []string{"", "/some/sub/dir"} {
		for _, filename := range files {
			require.Nil(t, os.WriteFile(
				filepath.Join(tempdir, sub, filename), fileData, os.FileMode(0o644),
			))
		}
	}

	impl := ReaderDefaultImpl{}
	res, err := impl.FindLicenseFiles(tempdir)
	require.Nil(t, err)
	require.Equal(t, 10, len(res), fmt.Sprintf("%+v", res))
	require.NotContains(t, res, filepath.Join(tempdir, "license.go"))
	require.NotContains(t, res, filepath.Join(tempdir, "README.md"))
}

func TestGetLicenseFromURL(t *testing.T) {
	tempdir, err := os.MkdirTemp("", testDirPrefix)
	require.Nil(t, err)
	defer func() { require.Nil(t, os.RemoveAll(tempdir)) }()

	testURL := "http://www.example.com"

	// Create a default implementation with caching enabled
	impl := DefaultDownloaderImpl{
		Options: &DownloaderOptions{
			EnableCache:       true,
			CacheDir:          tempdir,
			parallelDownloads: 1,
		},
	}

	// First, cache the data artificially
	require.Nil(t, impl.cacheData(testURL, []byte(testFullLicense)))

	// Now, get tej license from the URL, should be cached
	l, err := impl.getLicenseFromURL(testURL)
	require.Nil(t, err)
	require.NotNil(t, l)
	require.Equal(t, l.LicenseID, "Apache-2.0")
}

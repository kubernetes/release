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

package hash_test

import (
	"crypto/sha1"
	"crypto/sha256"
	"hash"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	kHash "k8s.io/release/v1/pkg/hash"
)

func TestSHA512ForFile(t *testing.T) {
	for _, tc := range []struct {
		prepare     func() string
		expected    string
		shouldError bool
	}{
		{ // success
			prepare: func() string {
				f, err := ioutil.TempFile("", "")
				require.Nil(t, err)

				_, err = f.WriteString("test")
				require.Nil(t, err)

				return f.Name()
			},
			expected: "ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f88" +
				"19a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc" +
				"5fa9ad8e6f57f50028a8ff",
			shouldError: false,
		},
		{ // error open file
			prepare:     func() string { return "" },
			shouldError: true,
		},
	} {
		filename := tc.prepare()

		res, err := kHash.SHA512ForFile(filename)

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
			require.Equal(t, tc.expected, res)
		}
	}
}

func TestSHA256ForFile(t *testing.T) {
	for _, tc := range []struct {
		prepare     func() string
		expected    string
		shouldError bool
	}{
		{ // success
			prepare: func() string {
				f, err := ioutil.TempFile("", "")
				require.Nil(t, err)

				_, err = f.WriteString("test")
				require.Nil(t, err)

				return f.Name()
			},
			expected:    "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
			shouldError: false,
		},
		{ // error open file
			prepare:     func() string { return "" },
			shouldError: true,
		},
	} {
		filename := tc.prepare()

		res, err := kHash.SHA256ForFile(filename)

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
			require.Equal(t, tc.expected, res)
		}
	}
}

func TestForFile(t *testing.T) {
	for _, tc := range []struct {
		prepare     func() (string, hash.Hash)
		expected    string
		shouldError bool
	}{
		{ // success
			prepare: func() (string, hash.Hash) {
				f, err := ioutil.TempFile("", "")
				require.Nil(t, err)

				_, err = f.WriteString("test")
				require.Nil(t, err)

				return f.Name(), sha1.New()
			},
			expected:    "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
			shouldError: false,
		},
		{ // error hasher is nil
			prepare: func() (string, hash.Hash) {
				return "", nil
			},
			shouldError: true,
		},
		{ // error file does not exist is nil
			prepare: func() (string, hash.Hash) {
				return "", sha256.New()
			},
			shouldError: true,
		},
	} {
		filename, hasher := tc.prepare()

		res, err := kHash.ForFile(filename, hasher)

		if tc.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
			require.Equal(t, tc.expected, res)
		}
	}
}

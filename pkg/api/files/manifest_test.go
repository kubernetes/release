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

package files_test

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/release/pkg/api/files"
)

func TestValidateFilestores(t *testing.T) {
	tests := []struct {
		filestores    []files.Filestore
		expectedError string
	}{
		{
			// Filestores are required
			filestores:    []files.Filestore{},
			expectedError: "filestore must be specified",
		},
		{
			// Filestores are required
			filestores:    nil,
			expectedError: "filestore must be specified",
		},
		{
			filestores: []files.Filestore{
				{Src: true, Base: "gs://src"},
			},
			expectedError: "no destination filestores found",
		},
		{
			filestores: []files.Filestore{
				{Base: "gs://dest1"},
			},
			expectedError: "source filestore not found",
		},
		{
			filestores: []files.Filestore{
				{Src: true, Base: "gs://src1"},
				{Src: true, Base: "gs://src2"},
			},
			expectedError: "found multiple source filestores",
		},
		{
			filestores: []files.Filestore{
				{Src: true, Base: "gs://src"},
				{Base: "gs://dest1"},
				{Base: "gs://dest2"},
			},
		},
		{
			filestores: []files.Filestore{
				{Src: true},
				{Base: "gs://dest"},
			},
			expectedError: "filestore did not have base set",
		},
		{
			filestores: []files.Filestore{
				{Src: true, Base: "gs://src"},
				{Base: "s3://dest"},
			},
			expectedError: "unsupported scheme in base",
		},
	}
	for _, test := range tests {
		err := files.ValidateFilestores(test.filestores)
		checkErrorMatchesExpected(t, err, test.expectedError)
	}
}

func TestValidateFiles(t *testing.T) {
	oksha := "4f2f040fa2bfe9bea64911a2a756e8a1727a8bfd757c5e031631a6e699fcf246"

	tests := []struct {
		files         []files.File
		expectedError string
	}{
		{
			// Files are required
			files:         []files.File{},
			expectedError: "file must be specified",
		},
		{
			// Files are required
			files:         nil,
			expectedError: "file must be specified",
		},
		{
			files: []files.File{
				{Name: "foo", SHA256: oksha},
			},
		},
		{
			files: []files.File{
				{SHA256: oksha},
			},
			expectedError: "name is required for file",
		},
		{
			files: []files.File{
				{Name: "foo", SHA256: "bad"},
			},
			expectedError: "sha256 was not valid (not hex)",
		},
		{
			files: []files.File{
				{Name: "foo"},
			},
			expectedError: "sha256 is required",
		},
		{
			files: []files.File{
				{Name: "foo", SHA256: "abcd"},
			},
			expectedError: "sha256 was not valid (bad length)",
		},
	}
	for _, test := range tests {
		err := files.ValidateFiles(test.files)
		checkErrorMatchesExpected(t, err, test.expectedError)
	}
}

func checkErrorMatchesExpected(t *testing.T, err error, expected string) {
	if err != nil && expected == "" {
		t.Errorf("unexpected error: %v", err)
	}
	if err != nil && expected != "" {
		actual := fmt.Sprintf("%v", err)
		if !strings.Contains(actual, expected) {
			t.Errorf("error %q did not contain expected %q", err, expected)
		}
	}
	if err == nil && expected != "" {
		t.Errorf("expected error %q", expected)
	}
}

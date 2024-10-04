/*
Copyright 2024 The Kubernetes Authors.

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

package cmd

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunValidateReleaseNotes(t *testing.T) {
	testDataPath := "testdata/validation-data"

	// Valid YAML returns no error
	err := runValidateReleaseNotes(filepath.Join(testDataPath, "valid.yaml"))
	assert.NoError(t, err, "Expected no error for valid YAML file")

	// Try a non-existent path
	err = runValidateReleaseNotes("nonexistent/path")
	assert.Error(t, err, "Expected error for non-existent path")
	assert.Contains(t, err.Error(), "does not exist", "Error should be about non-existent path")

	// Missing punctuation YAML returns error
	err = runValidateReleaseNotes(filepath.Join(testDataPath, "missing-punctuation.yaml"))
	assert.Error(t, err, "Expected error for missing punctuation YAML file")
	assert.Contains(t, err.Error(), "field does not end with valid punctuation", "Error should be about missing punctuation")

	// Try invalid yaml starting with "`"
	err = runValidateReleaseNotes(filepath.Join(testDataPath, "invalid-yaml-start.yaml"))
	assert.Error(t, err, "Expected error for invalid yaml")
	assert.Contains(t, err.Error(), "YAML unmarshaling testdata/validation-data/invalid-yaml-start", "Error should be about invalid yaml")

	// Try invalid multi line yaml
	err = runValidateReleaseNotes(filepath.Join(testDataPath, "invalid-multi-line.yaml"))
	assert.Error(t, err, "Expected error for invalid yaml")
	assert.Contains(t, err.Error(), "YAML unmarshaling testdata/validation-data/invalid-multi-line.yaml", "Error should be about invalid yaml")

	// Try invalid indent
	err = runValidateReleaseNotes(filepath.Join(testDataPath, "invalid-indent.yaml"))
	assert.Error(t, err, "Expected error for invalid yaml")
	assert.Contains(t, err.Error(), "YAML unmarshaling testdata/validation-data/invalid-indent.yaml", "Error should be about invalid yaml")
}

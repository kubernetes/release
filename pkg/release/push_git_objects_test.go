/*
Copyright 2020 The Kubernetes Authors.

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

package release

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckBranchName(t *testing.T) {
	ghp, err := NewGitPusher(&GitObjectPusherOptions{})
	require.Nil(t, err)

	sampleBaranches := []struct {
		branchName string
		valid      bool
	}{
		{"release-1.20", true},     // Valid name
		{"release-chorizo", false}, // Invalid, not a semver
		{"1.20", false},            // Invalid, has to start with release
	}
	for _, testCase := range sampleBaranches {
		if testCase.valid {
			require.Nil(t, ghp.checkBranchName(testCase.branchName))
		} else {
			require.NotNil(t, ghp.checkBranchName(testCase.branchName))
		}
	}
}

func TestCheckTagName(t *testing.T) {
	ghp, err := NewGitPusher(&GitObjectPusherOptions{})
	require.Nil(t, err)

	sampleTags := []struct {
		tagName string
		valid   bool
	}{
		{"v1.20.0-alpha.2", true}, // Valid
		{"myTag", false},          // Invalid, not a semver
		{"1.20", false},           // Invalid, incomplete
	}
	for _, testCase := range sampleTags {
		if testCase.valid {
			require.Nil(t, ghp.checkTagName(testCase.tagName))
		} else {
			require.NotNil(t, ghp.checkTagName(testCase.tagName))
		}
	}
}

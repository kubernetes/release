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

package git

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDescribeOptionsEmpty(t *testing.T) {
	sut := NewDescribeOptions()
	require.NotNil(t, sut)
	require.Nil(t, sut.toArgs())
}

func TestDescribeOptionsArguments(t *testing.T) {
	sut := NewDescribeOptions().
		WithTags().
		WithDirty().
		WithAbbrev(3).
		WithAlways().
		WithRevision("rev")
	require.NotNil(t, sut)
	require.Equal(t, []string{
		"--tags",
		"--dirty",
		"--always",
		"--abbrev=3",
		"rev",
	}, sut.toArgs())
}

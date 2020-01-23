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

package testing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func CheckErr(t *testing.T, err error, expectedMsg string) {
	t.Helper()
	if expectedMsg == "" {
		require.NoError(t, err)
		return
	}
	require.EqualError(t, err, expectedMsg)
}

func CheckErrSub(t *testing.T, err error, expectedSubstring string) {
	t.Helper()
	if expectedSubstring == "" {
		require.NoError(t, err)
		return
	}
	require.Contains(t, err.Error(), expectedSubstring)
}

// Run is a small wrapper around t.Run which enables parallel runs
// unconditionally
func Run(t *testing.T, name string, f func(*testing.T)) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		f(t)
	})
}

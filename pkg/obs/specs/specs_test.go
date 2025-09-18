/*
Copyright 2025 The Kubernetes Authors.

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

package specs_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/consts"
	"k8s.io/release/pkg/obs/specs"
)

var err = errors.New("error")

func TestValidate(t *testing.T) {
	testcases := []struct {
		prepare   func(options *specs.Options)
		shouldErr bool
	}{
		{ // default options
			prepare:   func(options *specs.Options) {},
			shouldErr: false,
		},
		{ // specs for package doesn't exist
			prepare: func(options *specs.Options) {
				options.SpecTemplatePath = "does_not_exist"
			},
			shouldErr: true,
		},
		{ // one of version or channel is required
			prepare: func(options *specs.Options) {
				options.Version = ""
				options.Channel = ""
			},
			shouldErr: true,
		},
		{ // selected channel is not supported
			prepare: func(options *specs.Options) {
				options.Channel = "not_supported"
			},
			shouldErr: true,
		},
		{ // revision is required
			prepare: func(options *specs.Options) {
				options.Revision = ""
			},
			shouldErr: true,
		},
		{ // architectures selection is not supported
			prepare: func(options *specs.Options) {
				options.Architectures = []string{"not_supported"}
			},
			shouldErr: true,
		},
		{ // output path dir doesn't exist
			prepare: func(options *specs.Options) {
				options.SpecOutputPath = "does_not_exist"
			},
			shouldErr: true,
		},
	}

	for _, tc := range testcases {
		options := specs.DefaultOptions()

		tc.prepare(options)

		newSpecPath(t)

		err = options.Validate()

		if tc.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func newSpecPath(t *testing.T) string {
	tempDir := t.TempDir()

	require.NoError(t, os.MkdirAll(
		consts.DefaultSpecTemplatePath,
		os.FileMode(0o755),
	))

	return tempDir
}

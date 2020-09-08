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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

func TestRunSetReleaseVersion(t *testing.T) {
	for _, tc := range []struct {
		anagoOpts *release.Options
		opts      *setReleaseVersionOptions
		shouldErr bool
		expected  string
	}{
		{
			anagoOpts: &release.Options{
				ReleaseType: release.ReleaseTypeOfficial,
			},
			opts: &setReleaseVersionOptions{
				buildVersion: "v1.19.1-rc.0.34+5f5b46a6e8ad56",
				branch:       "release-1.19",
			},
			shouldErr: false,
			expected: `declare -Ag RELEASE_VERSION
declare -ag ORDERED_RELEASE_KEYS
RELEASE_VERSION[official]="v1.19.1"
ORDERED_RELEASE_KEYS+=("official")
RELEASE_VERSION[rc]="v1.19.2-rc.0"
ORDERED_RELEASE_KEYS+=("rc")
export RELEASE_VERSION_PRIME=v1.19.1
`,
		},
		{
			anagoOpts: &release.Options{
				ReleaseType: release.ReleaseTypeRC,
			},
			opts: &setReleaseVersionOptions{
				buildVersion: "v1.19.1-rc.0.34+5f5b46a6e8ad56",
				branch:       "release-1.19",
			},
			shouldErr: false,
			expected: `declare -Ag RELEASE_VERSION
declare -ag ORDERED_RELEASE_KEYS
RELEASE_VERSION[rc]="v1.19.1-rc.1"
ORDERED_RELEASE_KEYS+=("rc")
export RELEASE_VERSION_PRIME=v1.19.1-rc.1
`,
		},
		{
			anagoOpts: &release.Options{
				ReleaseType: release.ReleaseTypeAlpha,
			},
			opts: &setReleaseVersionOptions{
				buildVersion: "v1.19.1-rc.0.34+5f5b46a6e8ad56",
				branch:       git.DefaultBranch,
			},
			shouldErr: true,
		},
		{
			anagoOpts: &release.Options{
				ReleaseType: release.ReleaseTypeAlpha,
			},
			opts: &setReleaseVersionOptions{
				buildVersion: "v1.20.0-alpha.0.1273+4e9bdd481e2400",
				branch:       git.DefaultBranch,
			},
			shouldErr: false,
			expected: `declare -Ag RELEASE_VERSION
declare -ag ORDERED_RELEASE_KEYS
RELEASE_VERSION[alpha]="v1.20.0-alpha.1"
ORDERED_RELEASE_KEYS+=("alpha")
export RELEASE_VERSION_PRIME=v1.20.0-alpha.1
`,
		},
		{
			anagoOpts: &release.Options{
				ReleaseType: release.ReleaseTypeBeta,
			},
			opts: &setReleaseVersionOptions{
				buildVersion: "v1.20.0-alpha.0.1273+4e9bdd481e2400",
				branch:       git.DefaultBranch,
			},
			shouldErr: false,
			expected: `declare -Ag RELEASE_VERSION
declare -ag ORDERED_RELEASE_KEYS
RELEASE_VERSION[beta]="v1.20.0-beta.0"
ORDERED_RELEASE_KEYS+=("beta")
export RELEASE_VERSION_PRIME=v1.20.0-beta.0
`,
		},
		{
			anagoOpts: &release.Options{
				ReleaseType: release.ReleaseTypeRC,
			},
			opts: &setReleaseVersionOptions{
				buildVersion: "v1.20.0-alpha.0.1273+4e9bdd481e2400",
				branch:       "release-1.20",
				parentBranch: git.DefaultBranch,
			},
			shouldErr: false,
			expected: `declare -Ag RELEASE_VERSION
declare -ag ORDERED_RELEASE_KEYS
RELEASE_VERSION[rc]="v1.20.0-rc.0"
ORDERED_RELEASE_KEYS+=("rc")
RELEASE_VERSION[alpha]="v1.21.0-alpha.0"
ORDERED_RELEASE_KEYS+=("alpha")
export RELEASE_VERSION_PRIME=v1.20.0-rc.0
`,
		},
	} {
		res, err := runSetReleaseVersion(tc.opts, tc.anagoOpts)
		if tc.shouldErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
			require.Equal(t, tc.expected, res)
		}
	}
}

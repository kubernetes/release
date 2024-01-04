/*
Copyright 2023 The Kubernetes Authors.

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

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/notes/options"
)

func TestPROptsValidateAndFinish(t *testing.T) {
	testOrg := "testOrg"
	testRepo := "testRepo"
	for _, tc := range []struct {
		name    string
		sut     checkPROptions
		mustErr bool
	}{
		{
			name: "good options",
			sut: checkPROptions{
				Options: options.Options{
					GithubOrg:  testOrg,
					GithubRepo: testRepo,
				},
				PullRequests: []int{1, 2},
			},
			mustErr: false,
		},
		{
			name: "missing repo",
			sut: checkPROptions{
				Options: options.Options{
					GithubOrg: testOrg,
				},
				PullRequests: []int{1, 2},
			},
			mustErr: true,
		},
		{
			name: "missing org",
			sut: checkPROptions{
				Options: options.Options{
					GithubRepo: testRepo,
				},
				PullRequests: []int{1, 2},
			},
			mustErr: true,
		},
		{
			name: "missing PRs",
			sut: checkPROptions{
				Options: options.Options{
					GithubOrg:  testOrg,
					GithubRepo: testRepo,
				},
				PullRequests: []int{},
			},
			mustErr: true,
		},
		{
			name: "invalid PR",
			sut: checkPROptions{
				Options: options.Options{
					GithubOrg:  testOrg,
					GithubRepo: testRepo,
				},
				PullRequests: []int{0},
			},
			mustErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mustErr {
				require.Error(t, tc.sut.ValidateAndFinish())
			} else {
				require.NoError(t, tc.sut.ValidateAndFinish())
			}
		})
	}
}

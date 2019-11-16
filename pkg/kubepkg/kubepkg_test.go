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

package kubepkg_test

import (
	"reflect"
	"strings"
	"testing"

	kpkg "k8s.io/release/pkg/kubepkg"
)

func TestGetKubeadmDependencies(t *testing.T) {
	testcases := []struct {
		name    string
		version string
		deps    []string
	}{
		{
			name:    "minimum supported kubernetes",
			version: "1.13.0",
			deps: []string{
				"kubelet (>= 1.13.0)",
				"kubectl (>= 1.13.0)",
				"kubernetes-cni (>= 0.7.5)",
				"cri-tools (>= 1.13.0)",
				"${misc:Depends}",
			},
		},
		{
			name:    "latest stable minor kubernetes",
			version: "1.16.0",
			deps: []string{
				"kubelet (>= 1.13.0)",
				"kubectl (>= 1.13.0)",
				"kubernetes-cni (>= 0.7.5)",
				"cri-tools (>= 1.13.0)",
				"${misc:Depends}",
			},
		},
		{
			name:    "latest alpha kubernetes",
			version: "1.17.0-alpha.0",
			deps: []string{
				"kubelet (>= 1.13.0)",
				"kubectl (>= 1.13.0)",
				"kubernetes-cni (>= 0.7.5)",
				"cri-tools (>= 1.13.0)",
				"${misc:Depends}",
			},
		},
		{
			name:    "next stable minor kubernetes",
			version: "1.17.0",
			deps: []string{
				"kubelet (>= 1.13.0)",
				"kubectl (>= 1.13.0)",
				"kubernetes-cni (>= 0.7.5)",
				"cri-tools (>= 1.13.0)",
				"${misc:Depends}",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			packageDef := kpkg.PackageDefinition{Version: tc.version}
			deps, err := kpkg.GetKubeadmDependencies(packageDef)
			if err != nil {
				t.Fatalf("did not expect an error: %v", err)
			}
			actual := strings.Split(deps, ", ")
			if len(actual) != len(tc.deps) {
				t.Fatalf("Expected %d deps but found %d", len(tc.deps), len(actual))
			}
			if !reflect.DeepEqual(actual, tc.deps) {
				t.Fatalf("expected %q but got %q", tc.deps, actual)
			}
		})
	}

}

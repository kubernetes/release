package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestGetKubeadmConfig(t *testing.T) {
	testcases := []struct {
		version      string
		expectConfig string
		expectErr    bool
	}{
		{
			"not-a-real-version",
			"",
			true,
		},
		{
			"1.12.0",
			"post-1.10/10-kubeadm.conf",
			false,
		},
		{
			"1.13.0",
			"post-1.10/10-kubeadm.conf",
			false,
		},
		{
			"1.15.0",
			"post-1.10/10-kubeadm.conf",
			false,
		},
	}

	for _, tc := range testcases {
		v := version{
			Version: tc.version,
		}
		kubeadmConfig, err := getKubeadmKubeletConfigFile(v)

		if err != nil {
			if !tc.expectErr {
				t.Errorf("getKubeadmConfig(%s) returned unwanted error: %v", tc.version, err)
			}
		} else {
			if kubeadmConfig != tc.expectConfig {
				t.Errorf("getKubeadmConfig(%s) got %q, wanted %q", tc.version, kubeadmConfig, tc.expectConfig)
			}
		}
	}
}

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
				"${misc:Depends}",
				"cri-tools (>= 1.13.0)",
			},
		},
		{
			name:    "latest stable minor kubernetes",
			version: "1.15.0",
			deps: []string{
				"kubelet (>= 1.13.0)",
				"kubectl (>= 1.13.0)",
				"kubernetes-cni (>= 0.7.5)",
				"${misc:Depends}",
				"cri-tools (>= 1.13.0)",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			v := version{Version: tc.version}
			deps, err := getKubeadmDependencies(v)
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

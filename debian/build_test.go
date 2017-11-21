package main

import (
	"testing"
)

func TestGetKubeadmConfig(t *testing.T) {
	testcases := []struct {
		version      string
		expectConfig string
		expectErr    bool
	}{
		{
			"1.6.10",
			"pre-1.8/10-kubeadm.conf",
			false,
		},
		{
			"1.7.10",
			"pre-1.8/10-kubeadm.conf",
			false,
		},
		{
			"1.7.0-beta.5",
			"pre-1.8/10-kubeadm.conf",
			false,
		},
		{
			"1.8.0",
			"post-1.8/10-kubeadm.conf",
			false,
		},
		{
			"1.8.0-alpha.0",
			"post-1.8/10-kubeadm.conf",
			false,
		},
		{
			"1.8.0-beta.2",
			"post-1.8/10-kubeadm.conf",
			false,
		},
		{
			"1.9.0",
			"post-1.8/10-kubeadm.conf",
			false,
		},
		{
			"not-a-real-version",
			"",
			true,
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

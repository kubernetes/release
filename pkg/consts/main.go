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

package consts

import (
	"slices"

	"github.com/sirupsen/logrus"
)

const (
	PackageCRITools      string = "cri-tools"
	PackageCRIO          string = "cri-o"
	PackageKubeadm       string = "kubeadm"
	PackageKubectl       string = "kubectl"
	PackageKubelet       string = "kubelet"
	PackageKubernetesCNI string = "kubernetes-cni"
)

const (
	ChannelTypeRelease    string = "release"
	ChannelTypePrerelease string = "prerelease"
	ChannelTypeNightly    string = "nightly"
)

const (
	ArchitectureI386  string = "386"
	ArchitectureAMD64 string = "amd64"
	ArchitectureARM   string = "arm"
	ArchitectureARM64 string = "arm64"
	ArchitecturePPC   string = "ppc"
	ArchitecturePPC64 string = "ppc64le"
	ArchitectureS390X string = "s390x"
	ArchitectureRISCV string = "riscv"
)

var (
	SupportedChannels = []string{
		ChannelTypeRelease,
		ChannelTypePrerelease,
		ChannelTypeNightly,
	}

	SupportedArchitectures = []string{
		ArchitectureAMD64,
		ArchitectureARM64,
		ArchitecturePPC64,
		ArchitectureS390X,
	}
	FastArchitectures = []string{
		ArchitectureAMD64,
	}
)

const (
	DefaultReleaseDownloadLinkBase = "https://dl.k8s.io/release"
	DefaultRevision                = "0"
	DefaultSpecTemplatePath        = "cmd/krel/templates/latest"
)

func IsSupported(field string, input, expected []string) bool {
	notSupported := []string{}

	for _, i := range input {
		supported := slices.Contains(expected, i)

		if !supported {
			notSupported = append(notSupported, i)
		}
	}

	if len(notSupported) > 0 {
		logrus.Infof(
			"Flag %s has an unsupported option: %v", field, notSupported,
		)

		return false
	}

	return true
}

func IsCoreKubernetesPackage(packageName string) bool {
	switch packageName {
	case PackageKubeadm, PackageKubectl, PackageKubelet:
		return true
	default:
		return false
	}
}

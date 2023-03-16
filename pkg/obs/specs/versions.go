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

package specs

import (
	"errors"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-utils/util"
)

var (
	minimumCNIVersion      = "1.1.1"
	minimumCRIToolsVersion = "1.26.0"

	kubeCNIVersionMatrix = map[string]string{
		">=1.0.0": "1.1.1",
	}

	kubeCRIToolsVersionMatrix = map[string]string{
		">=1.0.0": "1.26.0",
	}
)

// getKuebrnetesVersion is used to determine the Kubernetes version based on
// provided channel. If Kubernetes version is provided by the user, this
// function does nothing.
func (c *Client) GetKubernetesVersion(kubernetesVersion string, channel ChannelType) (string, error) {
	if kubernetesVersion != "" {
		return kubernetesVersion, nil
	}

	switch channel {
	case ChannelTesting:
		return c.impl.GetKubeVersion(release.VersionTypeStablePreRelease)
	case ChannelNightly:
		return c.impl.GetKubeVersion(release.VersionTypeCILatestCross)
	}

	return c.impl.GetKubeVersion(release.VersionTypeStable)
}

// GetPackageVersion determines the package version based on what package
// are we building specs for.
func (c *Client) GetPackageVersion(packageDef *PackageDefinition, kubeVersion string) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	switch packageDef.Name {
	case kubernetesCNIPackage:
		return getComponentVersion(packageDef, kubeVersion, c.options.CNIVersion, minimumCNIVersion)
	case criToolsPackage:
		return getComponentVersion(packageDef, kubeVersion, c.options.CRIToolsVersion, minimumCRIToolsVersion)
	}

	return util.TrimTagPrefix(kubeVersion), nil
}

// getComponentVersion is used to get version for kubernetes-cni and cri-tools.
func getComponentVersion(packageDef *PackageDefinition, kubeVersion, componentVersion, minimumVersion string) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	if componentVersion != "" {
		componentSemVer, err := util.TagStringToSemver(componentVersion)
		if err != nil {
			return "", fmt.Errorf("parsing %s version: %w", packageDef.Name, err)
		}
		minSemVer, err := util.TagStringToSemver(minimumVersion)
		if err != nil {
			return "", fmt.Errorf("parsing %s version as semver: %w", packageDef.Name, err)
		}

		if componentSemVer.LT(minSemVer) {
			return "", fmt.Errorf("specified %s version (%s) cannot be lower than %s", packageDef.Name, componentVersion, minimumCNIVersion)
		}

		logrus.Infof("Setting %s version to %s", packageDef.Name, componentVersion)
		return componentVersion, nil
	}

	matrix := map[string]string{}
	switch packageDef.Name {
	case kubernetesCNIPackage:
		matrix = kubeCNIVersionMatrix
	case criToolsPackage:
		matrix = kubeCRIToolsVersionMatrix
	}

	for cond, ver := range matrix {
		r, err := semver.ParseRange(cond)
		if err != nil {
			return "", fmt.Errorf("parsing semver range for %s: %w", packageDef.Name, err)
		}
		kubeSemVer, err := util.TagStringToSemver(kubeVersion)
		if err != nil {
			return "", fmt.Errorf("parsing kubernetes version: %w", err)
		}

		if r(kubeSemVer) {
			return getComponentVersion(packageDef, kubeVersion, ver, minimumVersion)
		}
	}

	return "", fmt.Errorf("unable to get %s version", packageDef.Name)
}

// GetDependencies is ued to get Kubernetes dependencies for a package.
// It might not include system dependencies.
func GetDependencies(packageDef *PackageDefinition) (map[string]string, error) {
	if packageDef == nil {
		return nil, errors.New("package definition cannot be nil")
	}

	deps := make(map[string]string)

	switch packageDef.Name {
	case "kubelet":
		deps[kubernetesCNIPackage] = minimumCNIVersion
	case "kubeadm":
		deps["kubelet"] = minimumKubernetesVersion
		deps["kubectl"] = minimumKubernetesVersion
		deps[kubernetesCNIPackage] = minimumCNIVersion
		deps[criToolsPackage] = minimumCRIToolsVersion
	}

	return deps, nil
}

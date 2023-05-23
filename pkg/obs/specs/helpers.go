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
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/obs/options"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-utils/util"
)

func isCoreKubernetesPackage(packageName string) bool {
	switch packageName {
	case options.PackageKubeadm:
		fallthrough
	case options.PackageKubectl:
		fallthrough
	case options.PackageKubelet:
		return true
	default:
		return false
	}
}

func getKubernetesChannelForVersion(kubernetesVersion string) (string, error) {
	kubeSemver, err := util.TagStringToSemver(kubernetesVersion)
	if err != nil {
		return "", fmt.Errorf("user-supplied kubernetes version is not valid semver: %w", err)
	}

	kubeVersionString := kubeSemver.String()
	kubeVersionParts := strings.Split(kubeVersionString, ".")

	switch {
	case len(kubeVersionParts) > 4:
		logrus.Info("User-supplied Kubernetes version is a CI version, using nightly channel")
		return options.ChannelTypeNightly, nil
	case len(kubeVersionParts) == 4:
		logrus.Info("User-supplied Kubernetes version is a pre-release version, using testing channel")
		return options.ChannelTypePrerelease, nil
	default:
		logrus.Info("User-supplied Kubernetes version is a release version, using release channel")
		return options.ChannelTypeRelease, nil
	}
}

// getKuebrnetesVersion is used to determine the Kubernetes version based on
// provided channel. If Kubernetes version is provided by the user, this
// function does nothing.
func (c *Client) getKubernetesVersionForChannel(channel string) (string, error) {
	switch channel {
	case options.ChannelTypePrerelease:
		return c.impl.GetKubeVersion(release.VersionTypeStablePreRelease)
	case options.ChannelTypeNightly:
		return c.impl.GetKubeVersion(release.VersionTypeCILatestCross)
	default:
		return c.impl.GetKubeVersion(release.VersionTypeStable)
	}
}

func (c *Client) getKubernetesDownloadLink(channel, baseURL, name, version, arch string) func() (string, error) {
	switch channel {
	case options.ChannelTypeNightly:
		return func() (string, error) {
			return c.getKubernetesCIDownloadLink(baseURL, name, version, arch)
		}
	default:
		return func() (string, error) {
			return getKubernetesReleaseDownloadLink(baseURL, name, version, arch), nil
		}
	}
}

func getKubernetesReleaseDownloadLink(baseURL, name, version, arch string) string {
	if baseURL == "" {
		baseURL = options.DefaultReleaseDownloadLinkBase
	}

	return fmt.Sprintf(
		"%s/%s/bin/linux/%s/%s",
		baseURL,
		util.AddTagPrefix(version),
		arch,
		name,
	)
}

func (c *Client) getKubernetesCIDownloadLink(baseURL, name, version, arch string) (string, error) {
	if baseURL == "" {
		baseURL = options.DefaultReleaseDownloadLinkBase
	}

	if version == "" {
		var err error
		version, err = c.impl.GetKubeVersion(release.VersionTypeCILatestCross)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf(
		"%s/ci/%s/bin/linux/%s/%s",
		baseURL, util.AddTagPrefix(version), arch, name), nil
}

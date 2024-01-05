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

	"k8s.io/release/pkg/consts"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-utils/util"
)

// GetKubernetesChannelForVersion returns channel for the given Kubernetes version.
func (s *Specs) GetKubernetesChannelForVersion(kubernetesVersion string) (string, error) {
	kubeSemver, err := s.impl.TagStringToSemver(kubernetesVersion)
	if err != nil {
		return "", fmt.Errorf("user-supplied kubernetes version is not valid semver: %w", err)
	}

	kubeVersionString := kubeSemver.String()
	kubeVersionParts := strings.Split(kubeVersionString, ".")

	switch {
	case len(kubeVersionParts) > 4:
		logrus.Info("User-supplied Kubernetes version is a CI version, using nightly channel")
		return consts.ChannelTypeNightly, nil
	case len(kubeVersionParts) == 4:
		logrus.Info("User-supplied Kubernetes version is a pre-release version, using testing channel")
		return consts.ChannelTypePrerelease, nil
	default:
		logrus.Info("User-supplied Kubernetes version is a release version, using release channel")
		return consts.ChannelTypeRelease, nil
	}
}

// GetKubernetesVersionForChannel is used to determine the Kubernetes version
// based on the provided channel.
func (s *Specs) GetKubernetesVersionForChannel(channel string) (string, error) {
	switch channel {
	case consts.ChannelTypePrerelease:
		return s.impl.GetKubeVersion(release.VersionTypeStablePreRelease)
	case consts.ChannelTypeNightly:
		return s.impl.GetKubeVersion(release.VersionTypeCILatestCross)
	default:
		return s.impl.GetKubeVersion(release.VersionTypeStable)
	}
}

// GetKubernetesDownloadLink gets the download link for Kubernetes packages
// based on given options.
func (s *Specs) GetKubernetesDownloadLink(channel, baseURL, name, version, arch string) func() (string, error) {
	switch channel {
	case consts.ChannelTypeNightly:
		return func() (string, error) {
			return s.GetKubernetesCIDownloadLink(baseURL, name, version, arch)
		}
	default:
		return func() (string, error) {
			return s.GetKubernetesReleaseDownloadLink(baseURL, name, version, arch), nil
		}
	}
}

// GetKubernetesReleaseDownloadLink gets the download link for release version
// of Kubernetes.
func (s *Specs) GetKubernetesReleaseDownloadLink(baseURL, name, version, arch string) string {
	if baseURL == "" {
		baseURL = consts.DefaultReleaseDownloadLinkBase
	}

	return fmt.Sprintf(
		"%s/%s/bin/linux/%s/%s",
		baseURL,
		util.AddTagPrefix(version),
		arch,
		name,
	)
}

// GetKubernetesCIDownloadLink gets the download link for CI version of
// Kubernetes.
func (s *Specs) GetKubernetesCIDownloadLink(baseURL, name, version, arch string) (string, error) {
	if baseURL == "" {
		baseURL = consts.DefaultReleaseDownloadLinkBase
	}

	if version == "" {
		var err error
		version, err = s.impl.GetKubeVersion(release.VersionTypeCILatestCross)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf(
		"%s/ci/%s/bin/linux/%s/%s",
		baseURL, util.AddTagPrefix(version), arch, name), nil
}

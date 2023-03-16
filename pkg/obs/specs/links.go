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

	"k8s.io/release/pkg/obs/options"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-utils/util"
)

func (c *Client) GetDownloadLinkBase(kubernetesVersion string, channel ChannelType) (string, error) {
	if kubernetesVersion == "" {
		return "", fmt.Errorf("getting download link base: kubernetes version is required")
	}

	if channel == ChannelNightly {
		return c.GetCIBuildsDownloadLinkBase(kubernetesVersion)
	}

	return GetDefaultReleaseDownloadLinkBase(kubernetesVersion)
}

func (c *Client) GetCIBuildsDownloadLinkBase(kubernetesVersion string) (string, error) {
	if kubernetesVersion == "" {
		return "", fmt.Errorf("getting ci download link base: kubernetes version is required")
	}

	ciVersion := kubernetesVersion
	if ciVersion == "" {
		var err error
		ciVersion, err = c.impl.GetKubeVersion(release.VersionTypeCILatestCross)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("https://dl.k8s.io/ci/%s/bin/linux/", util.AddTagPrefix(ciVersion)), nil
}

func GetDefaultReleaseDownloadLinkBase(kubernetesVersion string) (string, error) {
	if kubernetesVersion == "" {
		return "", fmt.Errorf("getting default download link base: kubernetes version is required")
	}

	return fmt.Sprintf(
		"%s/%s/bin/linux/",
		options.DefaultReleaseDownloadLinkBase,
		util.AddTagPrefix(kubernetesVersion),
	), nil
}

func GetCNIDownloadLink(version, arch string) (string, error) {
	if _, err := util.TagStringToSemver(version); err != nil {
		return "", fmt.Errorf("parsing cni version: %w", err)
	}
	version = util.TrimTagPrefix(version)

	return fmt.Sprintf("https://storage.googleapis.com/k8s-artifacts-cni/release/v%s/cni-plugins-linux-%s-v%s.tgz", version, arch, version), nil
}

func GetCRIToolsDownloadLink(version, arch string) (string, error) {
	if _, err := util.TagStringToSemver(version); err != nil {
		return "", fmt.Errorf("parsing cri-tools version: %w", err)
	}
	version = util.TrimTagPrefix(version)

	return fmt.Sprintf("https://storage.googleapis.com/k8s-artifacts-cri-tools/release/v%s/crictl-v%s-linux-%s.tar.gz", version, version, arch), nil
}

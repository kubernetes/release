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

package kubepkg

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/util"
)

type (
	BuildType   string
	ChannelType string
)

const (
	BuildDeb BuildType = "deb"
	BuildRpm BuildType = "rpm"
	BuildAll BuildType = "all"

	ChannelRelease ChannelType = "release"
	ChannelTesting ChannelType = "testing"
	ChannelNightly ChannelType = "nightly"

	minimumKubernetesVersion = "1.13.0"
	minimumCNIVersion        = "0.7.5"
	pre117CNIVersion         = "0.7.5"

	DefaultRevision = "0"

	packagesRootDir = "packages"

	kubeadmConf = "10-kubeadm.conf"
)

var (
	minimumCRIToolsVersion = minimumKubernetesVersion
	latestPackagesDir      = fmt.Sprintf("%s/%s", packagesRootDir, "latest")

	DefaultPackages      = []string{"kubelet", "kubectl", "kubeadm", "kubernetes-cni", "cri-tools"}
	DefaultChannels      = []string{"release", "testing", "nightly"}
	DefaultArchitectures = []string{"amd64", "arm", "arm64", "ppc64le", "s390x"}

	DefaultReleaseDownloadLinkBase = "https://dl.k8s.io"

	builtins = map[string]interface{}{
		"date": func() string {
			return time.Now().Format(time.RFC1123Z)
		},
	}

	keepTmp = flag.Bool("keep-tmp", false, "keep tmp dir after build")
)

type Build struct {
	Type        BuildType
	Package     string
	Definitions []*PackageDefinition
}

type PackageDefinition struct {
	Name     string
	Version  string
	Revision string

	Channel           ChannelType
	KubernetesVersion string
	KubeletCNIVersion string

	DownloadLinkBase         string
	KubeadmKubeletConfigFile string

	CNIDownloadLink string
}

type cfg struct {
	*PackageDefinition
	Type         BuildType
	Arch         string
	DebArch      string
	Package      string
	Dependencies string
}

func ConstructBuilds(buildType BuildType, packages, channels []string, kubeVersion, revision, cniVersion, criToolsVersion string) ([]Build, error) {
	logrus.Infof("Constructing builds...")

	builds := []Build{}

	for _, pkg := range packages {
		b := &Build{
			Type:    buildType,
			Package: pkg,
		}

		for _, channel := range channels {
			packageDef := &PackageDefinition{
				Revision: revision,
				Channel:  ChannelType(channel),
			}

			packageDef.KubernetesVersion = kubeVersion

			switch b.Package {
			case "kubernetes-cni":
				packageDef.Version = cniVersion
			case "cri-tools":
				packageDef.Version = criToolsVersion
			}

			b.Definitions = append(b.Definitions, packageDef)
		}

		builds = append(builds, *b)
	}

	logrus.Infof("Successfully constructed builds")
	return builds, nil
}

func WalkBuilds(builds []Build, architectures []string) error {
	logrus.Infof("Walking builds...")

	for _, arch := range architectures {
		for _, build := range builds {
			for _, packageDef := range build.Definitions {
				if err := buildPackage(build.Type, build.Package, arch, packageDef); err != nil {
					return err
				}
			}
		}
	}

	logrus.Infof("Successfully walked builds")
	return nil
}

func buildPackage(buildType BuildType, pkg, arch string, packageDef *PackageDefinition) error {
	if packageDef == nil {
		return errors.New("package definition cannot be nil")
	}

	c := cfg{
		PackageDefinition: packageDef,
		Type:              buildType,
		Package:           pkg,
		Arch:              arch,
	}

	c.Name = pkg

	var err error

	if c.KubernetesVersion != "" {
		logrus.Infof("Checking if user-supplied Kubernetes version (%s) is valid semver...", c.KubernetesVersion)
		kubeSemver, err := semver.Parse(c.KubernetesVersion)
		if err != nil {
			return errors.Wrap(err, "user-supplied Kubernetes version is not valid semver")
		}

		kubeVersionString := kubeSemver.String()
		kubeVersionParts := strings.Split(kubeVersionString, ".")

		switch {
		case len(kubeVersionParts) > 4:
			logrus.Info("User-supplied Kubernetes version is a CI version")
			logrus.Info("Setting channel to nightly")
			c.Channel = ChannelNightly
		case len(kubeVersionParts) == 4:
			logrus.Info("User-supplied Kubernetes version is a pre-release version")
			logrus.Info("Setting channel to testing")
			c.Channel = ChannelTesting
		default:
			logrus.Info("User-supplied Kubernetes version is a release version")
			logrus.Info("Setting channel to release")
			c.Channel = ChannelRelease
		}
	}

	c.KubernetesVersion, err = getKubernetesVersion(packageDef)
	if err != nil {
		return errors.Wrap(err, "getting Kubernetes version")
	}

	c.DownloadLinkBase, err = getDownloadLinkBase(packageDef)
	if err != nil {
		return errors.Wrap(err, "getting Kubernetes download link base")
	}

	logrus.Infof("Kubernetes download link base: %s", c.DownloadLinkBase)

	// For cases where a CI build version of Kubernetes is retrieved, replace instances
	// of "+" with "-", so that we build with a valid Debian package version.
	c.KubernetesVersion = strings.Replace(c.KubernetesVersion, "+", "-", 1)

	c.Version, err = getPackageVersion(packageDef)
	if err != nil {
		return errors.Wrap(err, "getting package version")
	}

	logrus.Infof("%s package version: %s", c.Name, c.Version)

	c.KubeletCNIVersion = minimumCNIVersion

	c.Dependencies, err = GetKubeadmDependencies(packageDef)
	if err != nil {
		return errors.Wrap(err, "getting kubeadm dependencies")
	}

	c.KubeadmKubeletConfigFile = kubeadmConf

	if c.Arch == "arm" {
		c.DebArch = "armhf"
	} else if c.Arch == "ppc64le" {
		c.DebArch = "ppc64el"
	} else {
		c.DebArch = c.Arch
	}

	c.CNIDownloadLink, err = getCNIDownloadLink(packageDef, c.Arch)
	if err != nil {
		return errors.Wrap(err, "getting CNI download link")
	}

	logrus.Infof("Building %s package for %s/%s architecture...", c.Package, c.Arch, c.DebArch)
	return c.run()
}

func (c cfg) run() error {
	// nolint
	// TODO: Get package directory for any version once package definitions are broken out
	srcdir := filepath.Join(latestPackagesDir, c.Package)
	dstdir, err := ioutil.TempDir(os.TempDir(), "debs")
	if err != nil {
		return err
	}

	if !*keepTmp {
		defer os.RemoveAll(dstdir)
	}

	_, err = buildTemplate(c, srcdir, dstdir)

	//nolint:godox
	// TODO: Move OS-specific logic into their own files
	switch c.Type {
	case BuildDeb:
		logrus.Infof("Running dpkg-buildpackage for %s (%s/%s)", c.Package, c.Arch, c.DebArch)

		dpkgStatus, dpkgErr := command.NewWithWorkDir(
			dstdir,
			"dpkg-buildpackage",
			"--unsigned-source",
			"--unsigned-changes",
			"--build=binary",
			"--host-arch",
			c.DebArch,
		).Run()

		if dpkgErr != nil {
			return dpkgErr
		}
		if !dpkgStatus.Success() {
			return errors.Errorf("dpkg-buildpackage command failed: %s", dpkgStatus.Error())
		}

		fileName := fmt.Sprintf("%s_%s-%s_%s.deb", c.Package, c.Version, c.Revision, c.DebArch)
		dstParts := []string{"bin", string(c.Channel), fileName}

		dstPath := filepath.Join(dstParts...)
		if mkdirErr := os.MkdirAll(dstPath, 0o777); mkdirErr != nil {
			return err
		}

		mvStatus, mvErr := command.New("mv", filepath.Join("/tmp", fileName), dstPath).Run()
		if mvErr != nil {
			return mvErr
		}
		if !mvStatus.Success() {
			return errors.Errorf("mv command failed: %s", mvStatus.Error())
		}

		logrus.Infof("Successfully built %s", dstPath)
	case BuildRpm:
		logrus.Fatal("Building rpms via kubepkg is not currently supported")
	}

	return nil
}

func getPackageVersion(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	logrus.Infof("Setting version for %s package...", packageDef.Name)
	switch packageDef.Name {
	case "kubernetes-cni":
		return getCNIVersion(packageDef)
	case "cri-tools":
		return getCRIToolsVersion(packageDef)
	}

	logrus.Infof("Using Kubernetes version for %s package", packageDef.Name)
	return packageDef.KubernetesVersion, nil
}

func getKubernetesVersion(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	if packageDef.KubernetesVersion != "" {
		logrus.Infof("Using Kubernetes version (%s) for %s package", packageDef.KubernetesVersion, packageDef.Name)
		return packageDef.KubernetesVersion, nil
	}
	switch packageDef.Channel {
	case ChannelTesting:
		return getTestingKubeVersion()
	case ChannelNightly:
		return getNightlyKubeVersion()
	}

	return getReleaseKubeVersion()
}

func getReleaseKubeVersion() (string, error) {
	logrus.Info("Retrieving Kubernetes release version...")
	return fetchVersion("https://dl.k8s.io/release/stable.txt")
}

func getTestingKubeVersion() (string, error) {
	logrus.Info("Retrieving Kubernetes testing version...")
	return fetchVersion("https://dl.k8s.io/release/latest.txt")
}

func getNightlyKubeVersion() (string, error) {
	logrus.Info("Retrieving Kubernetes nightly version...")
	return fetchVersion("https://dl.k8s.io/ci/k8s-master.txt")
}

func fetchVersion(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	versionBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}

	// Remove a newline and the v prefix from the string
	version := strings.Replace(strings.Replace(string(versionBytes), "v", "", 1), "\n", "", 1)

	logrus.Infof("Retrieved Kubernetes version: %s", version)
	return version, nil
}

func getCNIVersion(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	logrus.Infof("Getting CNI version...")

	kubeSemver, err := semver.Make(packageDef.KubernetesVersion)
	if err != nil {
		return "", err
	}

	v117, err := semver.Make("1.17.0-alpha.0")
	if err != nil {
		return "", err
	}

	if packageDef.Version != "" {
		if kubeSemver.LT(v117) {
			logrus.Infof("Kubernetes version earlier than 1.17 must use CNI version <= %s", pre117CNIVersion)
			logrus.Infof("Setting CNI version to %s", pre117CNIVersion)
			return pre117CNIVersion, nil
		}

		logrus.Infof("Setting CNI version to %s", packageDef.Version)
		return packageDef.Version, nil
	}

	logrus.Infof("Setting CNI version to %s", minimumCNIVersion)
	return minimumCNIVersion, nil
}

func getCRIToolsVersion(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	if packageDef.Version != "" {
		return packageDef.Version, nil
	}

	kubeSemver, err := semver.Parse(packageDef.KubernetesVersion)
	if err != nil {
		return "", err
	}

	logrus.Infof("Getting CRI version...")

	kubeVersionString := kubeSemver.String()
	kubeVersionParts := strings.Split(kubeVersionString, ".")

	criToolsMajor := kubeVersionParts[0]
	criToolsMinor := kubeVersionParts[1]

	// CRI tools releases are not published until after the corresponding Kubernetes release.
	// In instances where the Kubernetes version selected is a pre-release or CI build version, // we need to build from the previous minor version of CRI tools instead.
	//
	// Example:
	// Kubernetes version: 1.18.0-alpha.1
	// Initial CRI tools version: 1.18.0-alpha.1
	// Modified CRI tools version: 1.17.0
	if len(kubeVersionParts) >= 4 {
		criToolsMinorInt, err := strconv.Atoi(criToolsMinor)
		if err != nil {
			return "", err
		}

		criToolsMinorInt--
		criToolsMinor = strconv.Itoa(criToolsMinorInt)
	}

	criToolsVersion := fmt.Sprintf("%s.%s.0", criToolsMajor, criToolsMinor)

	releases, err := fetchReleases("kubernetes-sigs", "cri-tools", false)
	if err != nil {
		return "", err
	}

	var tags []string
	for _, release := range releases {
		criToolsReleaseTag := util.TrimTagPrefix(*release.TagName)
		criToolsReleaseVersionParts := strings.Split(criToolsReleaseTag, ".")
		criToolsReleaseMinor := criToolsReleaseVersionParts[1]

		if criToolsReleaseMinor == criToolsMinor {
			tags = append(tags, criToolsReleaseTag)
		}
	}

	for _, tag := range tags {
		tagSemver, err := semver.Parse(tag)
		if err != nil {
			return "", errors.Wrap(err, "could not parse tag semver")
		}

		criToolsSemver, err := semver.Parse(criToolsVersion)
		if err != nil {
			return "", errors.Wrap(err, "could not parse CRI tools semver")
		}

		if tagSemver.GTE(criToolsSemver) {
			criToolsVersion = tag
		}
	}

	logrus.Infof("Setting CRI tools version to %s", criToolsVersion)
	return criToolsVersion, nil
}

func fetchReleases(owner, repo string, includePrereleases bool) ([]*github.RepositoryRelease, error) {
	ghClient := github.NewClient(nil)

	allReleases, _, err := ghClient.Repositories.ListReleases(context.Background(), owner, repo, nil)
	if err != nil {
		return nil, err
	}

	var releases []*github.RepositoryRelease
	for _, release := range allReleases {
		if *release.Prerelease {
			if includePrereleases {
				releases = append(releases, release)
			}
		} else {
			releases = append(releases, release)
		}
	}

	return releases, nil
}

func getDownloadLinkBase(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	if packageDef.Channel == ChannelNightly {
		return getCIBuildsDownloadLinkBase(packageDef)
	}

	return getDefaultReleaseDownloadLinkBase(packageDef)
}

func getCIBuildsDownloadLinkBase(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	ciVersion := packageDef.KubernetesVersion
	if ciVersion == "" {
		var err error
		ciVersion, err = getNightlyKubeVersion()
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("https://dl.k8s.io/ci/v%s", ciVersion), nil
}

func getDefaultReleaseDownloadLinkBase(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	return fmt.Sprintf("%s/v%s", DefaultReleaseDownloadLinkBase, packageDef.KubernetesVersion), nil
}

func GetKubeadmDependencies(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	deps := []string{
		fmt.Sprintf("kubelet (>= %s)", minimumKubernetesVersion),
		fmt.Sprintf("kubectl (>= %s)", minimumKubernetesVersion),
		fmt.Sprintf("kubernetes-cni (>= %s)", minimumCNIVersion),
		fmt.Sprintf("cri-tools (>= %s)", minimumCRIToolsVersion),
		"${misc:Depends}",
	}

	return strings.Join(deps, ", "), nil
}

func getCNIDownloadLink(packageDef *PackageDefinition, arch string) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	sv, err := semver.Parse(packageDef.Version)
	if err != nil {
		return "", err
	}

	v075, err := semver.Make(pre117CNIVersion)
	if err != nil {
		return "", err
	}

	if sv.LTE(v075) {
		return fmt.Sprintf("https://github.com/containernetworking/plugins/releases/download/v%s/cni-plugins-%s-v%s.tgz", packageDef.Version, arch, packageDef.Version), nil
	}

	return fmt.Sprintf("https://github.com/containernetworking/plugins/releases/download/v%s/cni-plugins-linux-%s-v%s.tgz", packageDef.Version, arch, packageDef.Version), nil
}

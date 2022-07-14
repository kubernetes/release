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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	gogithub "github.com/google/go-github/v45/github"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/kubepkg/options"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-sdk/github"
	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/util"
)

type ChannelType string

const (
	ChannelRelease ChannelType = "release"
	ChannelTesting ChannelType = "testing"
	ChannelNightly ChannelType = "nightly"

	minimumKubernetesVersion = "1.13.0"
	CurrentCNIVersion        = "0.8.7"
	MinimumCNIVersion        = "0.8.6"

	kubeadmConf = "10-kubeadm.conf"
)

var (
	minimumCRIToolsVersion = minimumKubernetesVersion

	buildArchMap = map[string]map[options.BuildType]string{
		"amd64": {
			"deb": "amd64",
			"rpm": "x86_64",
		},
		"arm": {
			"deb": "armhf",
			"rpm": "armhfp",
		},
		"arm64": {
			"deb": "arm64",
			"rpm": "aarch64",
		},
		"ppc64le": {
			"deb": "ppc64el",
			"rpm": "ppc64le",
		},
		"s390x": {
			"deb": "s390x",
			"rpm": "s390x",
		},
	}

	builtins = map[string]interface{}{
		"date": func() string {
			return time.Now().Format(time.RFC1123Z)
		},
	}
)

type Client struct {
	options *options.Options
	impl    Impl
}

func New(o *options.Options) *Client {
	return &Client{
		options: o,
		impl:    &impl{},
	}
}

func (c *Client) SetImpl(impl Impl) {
	c.impl = impl
}

type impl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Impl
type Impl interface {
	RunSuccessWithWorkDir(workDir, cmd string, args ...string) error
	Releases(owner, repo string, includePrereleases bool) ([]*gogithub.RepositoryRelease, error)
	GetKubeVersion(versionType release.VersionType) (string, error)
	ReadFile(string) ([]byte, error)
	WriteFile(string, []byte, os.FileMode) error
}

func (i *impl) RunSuccessWithWorkDir(workDir, cmd string, args ...string) error {
	return command.NewWithWorkDir(workDir, cmd, args...).RunSuccess()
}

func (i *impl) Releases(owner, repo string, includePrereleases bool) ([]*gogithub.RepositoryRelease, error) {
	return github.New().Releases(owner, repo, includePrereleases)
}

func (i *impl) GetKubeVersion(versionType release.VersionType) (string, error) {
	return release.NewVersion().GetKubeVersion(versionType)
}

func (i *impl) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (i *impl) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

type Build struct {
	Type        options.BuildType
	Package     string
	Definitions []*PackageDefinition
	TemplateDir string
}

type PackageDefinition struct {
	Name     string
	Version  string
	Revision string

	Channel ChannelType

	KubernetesVersion string
	Dependencies      map[string]string

	DownloadLinkBase         string
	KubeadmKubeletConfigFile string

	CNIVersion      string
	CNIDownloadLink string
}

type buildConfig struct {
	*PackageDefinition
	Type      options.BuildType
	GoArch    string
	BuildArch string
	Package   string

	TemplateDir string
	workspace   string
	specOnly    bool
}

func (c *Client) ConstructBuilds() ([]Build, error) {
	logrus.Infof("Constructing builds...")

	builds := []Build{}

	for _, pkg := range c.options.Packages() {
		// TODO: Get package directory for any version once package definitions are broken out
		packageTemplateDir := filepath.Join(c.options.TemplateDir(), string(c.options.BuildType()), pkg)
		if _, err := os.Stat(packageTemplateDir); err != nil {
			return nil, fmt.Errorf("finding package template dir: %w", err)
		}

		b := &Build{
			Type:        c.options.BuildType(),
			Package:     pkg,
			TemplateDir: packageTemplateDir,
		}

		for _, channel := range c.options.Channels() {
			packageDef := &PackageDefinition{
				Revision: c.options.Revision(),
				Channel:  ChannelType(channel),
			}

			packageDef.KubernetesVersion = c.options.KubeVersion()

			switch b.Package {
			case "kubelet":
				packageDef.CNIVersion = c.options.CNIVersion()
			case "kubernetes-cni":
				packageDef.Version = c.options.CNIVersion()
			case "cri-tools":
				packageDef.Version = c.options.CRIToolsVersion()
			}

			b.Definitions = append(b.Definitions, packageDef)
		}

		builds = append(builds, *b)
	}

	logrus.Infof("Successfully constructed builds")
	return builds, nil
}

func (c *Client) WalkBuilds(builds []Build) (err error) {
	logrus.Infof("Walking builds...")

	workingDir := os.Getenv("KUBEPKG_WORKING_DIR")
	if workingDir == "" {
		workingDir, err = os.MkdirTemp("", "kubepkg")
		if err != nil {
			return err
		}
	}

	for _, arch := range c.options.Architectures() {
		for _, build := range builds {
			for _, packageDef := range build.Definitions {
				if err := c.buildPackage(build, packageDef, arch, workingDir); err != nil {
					return err
				}
			}
		}
	}
	if c.options.SpecOnly() {
		logrus.Infof("Package specs have been saved in %s", workingDir)
	}
	logrus.Infof("Successfully walked builds")
	return nil
}

func (c *Client) buildPackage(build Build, packageDef *PackageDefinition, arch, tmpDir string) error {
	if packageDef == nil {
		return errors.New("package definition cannot be nil")
	}

	pd := &PackageDefinition{}
	*pd = *packageDef

	bc := &buildConfig{
		PackageDefinition: pd,
		Type:              build.Type,
		Package:           build.Package,
		GoArch:            arch,
		TemplateDir:       build.TemplateDir,
		workspace:         tmpDir,
		specOnly:          c.options.SpecOnly(),
	}

	bc.Name = build.Package

	var err error

	if bc.KubernetesVersion != "" {
		logrus.Infof("Checking if user-supplied Kubernetes version (%s) is valid semver...", bc.KubernetesVersion)
		kubeSemver, err := util.TagStringToSemver(bc.KubernetesVersion)
		if err != nil {
			return fmt.Errorf("user-supplied Kubernetes version is not valid semver: %w", err)
		}

		kubeVersionString := kubeSemver.String()
		kubeVersionParts := strings.Split(kubeVersionString, ".")

		switch {
		case len(kubeVersionParts) > 4:
			logrus.Info("User-supplied Kubernetes version is a CI version")
			logrus.Info("Setting channel to nightly")
			bc.Channel = ChannelNightly
		case len(kubeVersionParts) == 4:
			logrus.Info("User-supplied Kubernetes version is a pre-release version")
			logrus.Info("Setting channel to testing")
			bc.Channel = ChannelTesting
		default:
			logrus.Info("User-supplied Kubernetes version is a release version")
			logrus.Info("Setting channel to release")
			bc.Channel = ChannelRelease
		}
	}

	bc.KubernetesVersion, err = c.GetKubernetesVersion(pd)
	if err != nil {
		return fmt.Errorf("getting Kubernetes version: %w", err)
	}

	bc.DownloadLinkBase, err = c.GetDownloadLinkBase(pd)
	if err != nil {
		return fmt.Errorf("getting Kubernetes download link base: %w", err)
	}

	logrus.Infof("Kubernetes download link base: %s", bc.DownloadLinkBase)

	// For cases where a CI build version of Kubernetes is retrieved, replace instances
	// of "+" with "-", so that we build with a valid Debian package version.
	bc.KubernetesVersion = strings.Replace(bc.KubernetesVersion, "+", "-", 1)

	bc.Version, err = c.GetPackageVersion(pd)
	if err != nil {
		return fmt.Errorf("getting package version: %w", err)
	}

	logrus.Infof("%s package version: %s", bc.Name, bc.Version)

	bc.Dependencies, err = GetDependencies(pd)
	if err != nil {
		return fmt.Errorf("getting dependencies: %w", err)
	}

	bc.KubeadmKubeletConfigFile = kubeadmConf

	bc.BuildArch = getBuildArch(bc.GoArch, bc.Type)

	bc.CNIVersion, err = GetCNIVersion(pd)
	if err != nil {
		return fmt.Errorf("getting CNI version: %w", err)
	}

	bc.CNIDownloadLink, err = GetCNIDownloadLink(pd.Version, bc.GoArch)
	if err != nil {
		return fmt.Errorf("getting CNI download link: %w", err)
	}

	logrus.Infof("Building %s package for %s/%s architecture...", bc.Package, bc.GoArch, bc.BuildArch)
	return c.run(bc)
}

func (c *Client) run(bc *buildConfig) error {
	workspaceInfo, err := os.Stat(bc.workspace)
	if err != nil {
		return err
	}

	specDir := filepath.Join(bc.workspace, string(bc.Channel), bc.Package)
	specDirWithArch := filepath.Join(specDir, bc.GoArch)

	if err := os.MkdirAll(specDirWithArch, workspaceInfo.Mode()); err != nil {
		return err
	}

	// TODO: keepTmp/cleanup needs to defined in kubepkg root
	if !bc.specOnly {
		defer os.RemoveAll(specDirWithArch)
	}

	if _, err := buildSpecs(bc, specDirWithArch); err != nil {
		return err
	}

	if bc.specOnly {
		logrus.Info("Spec-only mode was selected; kubepkg will now exit without building packages")
		return nil
	}

	// TODO: Move OS-specific logic into their own files
	switch bc.Type {
	case options.BuildDeb:
		logrus.Infof("Running dpkg-buildpackage for %s (%s/%s)", bc.Package, bc.GoArch, bc.BuildArch)

		if err := c.impl.RunSuccessWithWorkDir(
			specDirWithArch,
			"dpkg-buildpackage",
			"--unsigned-source",
			"--unsigned-changes",
			"--build=binary",
			"--host-arch",
			bc.BuildArch,
		); err != nil {
			return fmt.Errorf("running debian package build: %w", err)
		}

		fileName := fmt.Sprintf(
			"%s_%s-%s_%s.deb",
			bc.Package,
			bc.Version,
			bc.Revision,
			bc.BuildArch,
		)

		dstPath := filepath.Join("bin", string(bc.Channel), fileName)
		logrus.Infof("Using package destination path %s", dstPath)

		if err := os.MkdirAll(filepath.Dir(dstPath), os.FileMode(0o777)); err != nil {
			return fmt.Errorf("creating %s: %w", filepath.Dir(dstPath), err)
		}

		srcPath := filepath.Join(specDir, fileName)
		input, err := c.impl.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", srcPath, err)
		}

		err = c.impl.WriteFile(dstPath, input, os.FileMode(0o644))
		if err != nil {
			return fmt.Errorf("writing file to %s: %w", dstPath, err)
		}

		logrus.Infof("Successfully built %s", dstPath)
	case options.BuildRpm:
		logrus.Info("Building rpms via kubepkg is not currently supported")
	}

	return nil
}

func (c *Client) GetPackageVersion(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	logrus.Infof("Setting version for %s package...", packageDef.Name)
	switch packageDef.Name {
	case "kubernetes-cni":
		return GetCNIVersion(packageDef)
	case "cri-tools":
		return c.GetCRIToolsVersion(packageDef)
	}

	logrus.Infof(
		"Using Kubernetes version %s for %s package",
		packageDef.KubernetesVersion, packageDef.Name,
	)
	return util.TrimTagPrefix(packageDef.KubernetesVersion), nil
}

func (c *Client) GetKubernetesVersion(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	if packageDef.KubernetesVersion != "" {
		logrus.Infof("Using Kubernetes version (%s) for %s package", packageDef.KubernetesVersion, packageDef.Name)
		return packageDef.KubernetesVersion, nil
	}
	switch packageDef.Channel {
	case ChannelTesting:
		return c.impl.GetKubeVersion(release.VersionTypeStablePreRelease)
	case ChannelNightly:
		return c.impl.GetKubeVersion(release.VersionTypeCILatestCross)
	}

	return c.impl.GetKubeVersion(release.VersionTypeStable)
}

func GetCNIVersion(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	// TODO: Ensure version is not less than MinimumCNIVersion
	logrus.Infof("Getting CNI version...")
	if packageDef.CNIVersion != "" {
		cniSemVer, err := util.TagStringToSemver(packageDef.CNIVersion)
		if err != nil {
			return "", fmt.Errorf("parsing CNI version: %w", err)
		}
		minCNISemVer, err := util.TagStringToSemver(MinimumCNIVersion)
		if err != nil {
			return "", fmt.Errorf("parsing CNI version: %w", err)
		}

		if cniSemVer.LT(minCNISemVer) {
			return "", fmt.Errorf("specified CNI version (%s) cannot be lower than %s", packageDef.CNIVersion, MinimumCNIVersion)
		}

		logrus.Infof("Setting CNI version to %s", packageDef.CNIVersion)
		return packageDef.CNIVersion, nil
	}

	logrus.Infof("Setting CNI version to %s", MinimumCNIVersion)
	return MinimumCNIVersion, nil
}

func (c *Client) GetCRIToolsVersion(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	if packageDef.Version != "" {
		return packageDef.Version, nil
	}

	kubeSemver, err := util.TagStringToSemver(packageDef.KubernetesVersion)
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

	releases, err := c.impl.Releases("kubernetes-sigs", "cri-tools", false)
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
			return "", fmt.Errorf("could not parse tag semver: %w", err)
		}

		criToolsSemver, err := semver.Parse(criToolsVersion)
		if err != nil {
			return "", fmt.Errorf("could not parse CRI tools semver: %w", err)
		}

		if tagSemver.GTE(criToolsSemver) {
			criToolsVersion = tag
		}
	}

	logrus.Infof("Setting CRI tools version to %s", criToolsVersion)
	return criToolsVersion, nil
}

func (c *Client) GetDownloadLinkBase(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	if packageDef.Channel == ChannelNightly {
		return c.GetCIBuildsDownloadLinkBase(packageDef)
	}

	return GetDefaultReleaseDownloadLinkBase(packageDef)
}

func (c *Client) GetCIBuildsDownloadLinkBase(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	ciVersion := packageDef.KubernetesVersion
	if ciVersion == "" {
		var err error
		ciVersion, err = c.impl.GetKubeVersion(release.VersionTypeCILatestCross)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("https://dl.k8s.io/ci/%s", util.AddTagPrefix(ciVersion)), nil
}

func GetDefaultReleaseDownloadLinkBase(packageDef *PackageDefinition) (string, error) {
	if packageDef == nil {
		return "", errors.New("package definition cannot be nil")
	}

	return fmt.Sprintf(
		"%s/%s",
		options.DefaultReleaseDownloadLinkBase,
		util.AddTagPrefix(packageDef.KubernetesVersion),
	), nil
}

func GetDependencies(packageDef *PackageDefinition) (map[string]string, error) {
	if packageDef == nil {
		return nil, errors.New("package definition cannot be nil")
	}

	deps := make(map[string]string)

	switch packageDef.Name {
	case "kubelet":
		deps["kubernetes-cni"] = MinimumCNIVersion
	case "kubeadm":
		deps["kubelet"] = minimumKubernetesVersion
		deps["kubectl"] = minimumKubernetesVersion
		deps["kubernetes-cni"] = MinimumCNIVersion // deb based kubeadm still requires kubernetes-cni
		deps["cri-tools"] = minimumCRIToolsVersion
	}

	return deps, nil
}

func getBuildArch(goArch string, buildType options.BuildType) string {
	return buildArchMap[goArch][buildType]
}

func GetCNIDownloadLink(version, arch string) (string, error) {
	if _, err := util.TagStringToSemver(version); err != nil {
		return "", fmt.Errorf("parsing CNI version: %w", err)
	}

	return fmt.Sprintf("https://storage.googleapis.com/k8s-artifacts-cni/release/v%s/cni-plugins-linux-%s-v%s.tgz", version, arch, version), nil
}

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
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
)

type ChannelType string

const (
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

type work struct {
	src  string
	dst  string
	t    *template.Template
	info os.FileInfo
}

type Build struct {
	Package     string
	Definitions []PackageDefinition
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
	Arch         string
	DebArch      string
	Package      string
	Dependencies string
}

func ConstructBuilds(packages, channels []string, kubeVersion, revision, cniVersion, criToolsVersion string) ([]Build, error) {
	builds := []Build{}

	for _, pkg := range packages {
		b := &Build{
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

			b.Definitions = append(b.Definitions, *packageDef)
		}

		builds = append(builds, *b)
	}

	return builds, nil
}

func WalkBuilds(builds []Build, architectures []string) error {
	for _, arch := range architectures {
		for _, build := range builds {
			for _, packageDef := range build.Definitions {
				if err := buildPackage(build.Package, arch, packageDef); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func buildPackage(pkg, arch string, packageDef PackageDefinition) error {
	c := cfg{
		PackageDefinition: &packageDef,
		Package:           pkg,
		Arch:              arch,
	}

	c.Name = pkg

	var err error

	if c.KubernetesVersion != "" {
		log.Printf("checking k8s semver")
		kubeSemver, err := semver.Parse(c.KubernetesVersion)
		if err != nil {
			log.Fatalf("could not parse k8s semver: %v", err)
		}

		kubeVersionString := kubeSemver.String()
		kubeVersionParts := strings.Split(kubeVersionString, ".")

		log.Printf("%v, len: %d", kubeVersionParts, len(kubeVersionParts))
		switch {
		case len(kubeVersionParts) > 4:
			c.Channel = ChannelNightly
		case len(kubeVersionParts) == 4:
			c.Channel = ChannelTesting
		default:
			c.Channel = ChannelRelease
		}
	}

	c.KubernetesVersion, err = getKubernetesVersion(packageDef)
	if err != nil {
		log.Fatalf("error getting Kubernetes version: %v", err)
	}

	log.Printf("download link base is %s", c.DownloadLinkBase)
	c.DownloadLinkBase, err = getDownloadLinkBase(packageDef)
	if err != nil {
		log.Fatalf("error getting Kubernetes download link base: %v", err)
	}

	log.Printf("download link base is %s", c.DownloadLinkBase)

	// For cases where a CI build version of Kubernetes is retrieved, replace instances
	// of "+" with "-", so that we build with a valid Debian package version.
	c.KubernetesVersion = strings.Replace(c.KubernetesVersion, "+", "-", 1)

	c.Version, err = getPackageVersion(packageDef)
	if err != nil {
		log.Fatalf("error getting package version: %v", err)
	}

	log.Printf("package version is %s", c.Version)

	c.KubeletCNIVersion = minimumCNIVersion

	c.Dependencies, err = GetKubeadmDependencies(packageDef)
	if err != nil {
		log.Fatalf("error getting kubeadm dependencies: %v", err)
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
		log.Fatalf("error getting CNI download link: %v", err)
	}

	return c.run()
}

func (c cfg) run() error {
	log.Printf("!!!!!!!!! doing: %#v", c)
	var w []work

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

	if err := filepath.Walk(srcdir, func(srcfile string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dstfile := filepath.Join(dstdir, srcfile[len(srcdir):])
		if dstfile == dstdir {
			return nil
		}
		if f.IsDir() {
			log.Print(dstfile)
			return os.Mkdir(dstfile, f.Mode())
		}
		t, err := template.
			New("").
			Funcs(builtins).
			Option("missingkey=error").
			ParseFiles(srcfile)
		if err != nil {
			return err
		}
		w = append(w, work{
			src:  srcfile,
			dst:  dstfile,
			t:    t.Templates()[0],
			info: f,
		})

		return nil
	}); err != nil {
		return err
	}

	for _, w := range w {
		log.Printf("w: %#v", w)
		if err := func() error {
			f, err := os.OpenFile(w.dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0)
			if err != nil {
				return err
			}
			defer f.Close()

			if err := w.t.Execute(f, c); err != nil {
				return err
			}
			if err := os.Chmod(w.dst, w.info.Mode()); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}

	err = runCommand(dstdir, "dpkg-buildpackage", "-us", "-uc", "-b", "-a"+c.DebArch)
	if err != nil {
		return err
	}

	dstParts := []string{"bin", string(c.Channel)}

	dstPath := filepath.Join(dstParts...)
	if mkdirErr := os.MkdirAll(dstPath, 0o777); mkdirErr != nil {
		return err
	}

	fileName := fmt.Sprintf("%s_%s-%s_%s.deb", c.Package, c.Version, c.Revision, c.DebArch)
	err = runCommand("", "mv", filepath.Join("/tmp", fileName), dstPath)
	if err != nil {
		return err
	}

	return nil
}

func runCommand(pwd, command string, cmdArgs ...string) error {
	cmd := exec.Command(command, cmdArgs...)
	if pwd == "" {
		cmd.Dir = pwd
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getPackageVersion(packageDef PackageDefinition) (string, error) {
	log.Printf("Setting version for %s package...", packageDef.Name)
	switch packageDef.Name {
	case "kubernetes-cni":
		return getCNIVersion(packageDef)
	case "cri-tools":
		return getCRIToolsVersion(packageDef)
	}

	log.Printf("using Kubernetes version")
	return packageDef.KubernetesVersion, nil
}

func getKubernetesVersion(packageDef PackageDefinition) (string, error) {
	if packageDef.KubernetesVersion != "" {
		log.Printf("Using Kubernetes version (%s) for %s package...", packageDef.KubernetesVersion, packageDef.Name)
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
	log.Print("Retrieving Kubernetes release version...")
	return fetchVersion("https://dl.k8s.io/release/stable.txt")
}

func getTestingKubeVersion() (string, error) {
	log.Print("Retrieving Kubernetes testing version...")
	return fetchVersion("https://dl.k8s.io/release/latest.txt")
}

func getNightlyKubeVersion() (string, error) {
	log.Print("Retrieving Kubernetes nightly version...")
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
	return strings.Replace(strings.Replace(string(versionBytes), "v", "", 1), "\n", "", 1), nil
}

func getCNIVersion(packageDef PackageDefinition) (string, error) {
	log.Printf("using CNI version")

	kubeSemver, err := semver.Make(packageDef.KubernetesVersion)
	if err != nil {
		return "", err
	}

	v117, err := semver.Make("1.17.0-alpha.0")
	if err != nil {
		return "", err
	}

	log.Printf("checking kube version (%s) against %s", kubeSemver.String(), v117.String())
	if packageDef.Version != "" {
		if kubeSemver.LT(v117) {
			return pre117CNIVersion, nil
		}
		return packageDef.Version, nil
	}

	return minimumCNIVersion, nil
}

func getCRIToolsVersion(packageDef PackageDefinition) (string, error) {
	if packageDef.Version != "" {
		return packageDef.Version, nil
	}

	kubeSemver, err := semver.Parse(packageDef.KubernetesVersion)
	if err != nil {
		return "", err
	}

	log.Printf("using CRI version")
	kubeVersionString := kubeSemver.String()
	kubeVersionParts := strings.Split(kubeVersionString, ".")

	criToolsMajor := kubeVersionParts[0]
	criToolsMinor := kubeVersionParts[1]

	log.Printf("%v, len: %v", kubeVersionParts, len(kubeVersionParts))
	// v1.17.0-alpha.0.1809+ff8716f4cf6180
	if len(kubeVersionParts) >= 4 {
		criToolsMinorInt, err := strconv.Atoi(criToolsMinor)
		if err != nil {
			return "", err
		}

		log.Printf("CRI minor is %s", criToolsMinor)
		criToolsMinorInt--
		criToolsMinor = strconv.Itoa(criToolsMinorInt)
		log.Printf("CRI minor is %s", criToolsMinor)
	}

	criToolsVersion := fmt.Sprintf("%s.%s.0", criToolsMajor, criToolsMinor)

	releases, err := fetchReleases("kubernetes-sigs", "cri-tools", false)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	var tags []string
	for _, release := range releases {
		criToolsReleaseTag := strings.Trim(*release.TagName, "v")
		criToolsReleaseVersionParts := strings.Split(criToolsReleaseTag, ".")
		criToolsReleaseMinor := criToolsReleaseVersionParts[1]

		if criToolsReleaseMinor == criToolsMinor {
			tags = append(tags, criToolsReleaseTag)
		}
	}

	for _, tag := range tags {
		tagSemver, err := semver.Parse(tag)
		if err != nil {
			log.Fatalf("could not parse tag semver: %v", err)
		}

		criToolsSemver, err := semver.Parse(criToolsVersion)
		if err != nil {
			log.Fatalf("could not parse CRI tools semver: %v", err)
		}

		if tagSemver.GTE(criToolsSemver) {
			criToolsVersion = tag
		}
	}

	log.Printf("CRI tools version is %s", criToolsVersion)
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

func getDownloadLinkBase(packageDef PackageDefinition) (string, error) {
	if packageDef.Channel == ChannelNightly {
		return getCIBuildsDownloadLinkBase(packageDef)
	}

	return getDefaultReleaseDownloadLinkBase(packageDef)
}

func getCIBuildsDownloadLinkBase(packageDef PackageDefinition) (string, error) {
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

func getDefaultReleaseDownloadLinkBase(packageDef PackageDefinition) (string, error) {
	return fmt.Sprintf("%s/v%s", DefaultReleaseDownloadLinkBase, packageDef.KubernetesVersion), nil
}

func GetKubeadmDependencies(packageDef PackageDefinition) (string, error) {
	deps := []string{
		fmt.Sprintf("kubelet (>= %s)", minimumKubernetesVersion),
		fmt.Sprintf("kubectl (>= %s)", minimumKubernetesVersion),
		fmt.Sprintf("kubernetes-cni (>= %s)", minimumCNIVersion),
		fmt.Sprintf("cri-tools (>= %s)", minimumCRIToolsVersion),
		"${misc:Depends}",
	}

	return strings.Join(deps, ", "), nil
}

func getCNIDownloadLink(packageDef PackageDefinition, arch string) (string, error) {
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

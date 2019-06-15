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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver"
	"github.com/pkg/errors"
)

const (
	minimumKubernetesVersion       = "1.12.0-alpha.0"
	minimumStableKubernetesVersion = "1.12.0"

	minimumCNIVersion = "0.7.5"
	latestCNIVersion  = "0.8.1"

	BuildTypeAll        BuildType = "all"
	BuildTypeKubernetes BuildType = "kubernetes"
	BuildTypeStable     BuildType = "stable"
	BuildTypeUnstable   BuildType = "unstable"
	BuildTypeNightly    BuildType = "nightly"

	RepositoryDefault  RepositoryName = "kubernetes-xenial"
	RepositoryStable   RepositoryName = "kubernetes-stable"
	RepositoryUnstable RepositoryName = "kubernetes-unstable"
	RepositoryNightly  RepositoryName = "kubernetes-nightly"

	defaultDownloadLinkBase = "https://dl.k8s.io"
)

type BuildType string

type RepositoryName string

var DownloadLink string

var DownloadLinkBase string

type Build struct {
	Packages    []string
	PackageSets []PackageSet
}

type PackageSet struct {
	BuildType
	Name          string
	Version       string
	Revision      string
	Repositories  []RepositoryDefinition
	Architectures []string
	Dependencies  []string
}

type RepositoryDefinition struct {
	Name     RepositoryName
	LinkBase string
}

func GetDownloadLink(buildType BuildType, version string) (string, error) {
	switch buildType {
	case BuildTypeAll, BuildTypeKubernetes, BuildTypeStable, BuildTypeUnstable:
		downloadLink, err := getReleaseDownloadLink(version)
		if err != nil {
			return "", err
		}

		return downloadLink, nil
	case BuildTypeNightly:
		downloadLink, err := getCIBuildsDownloadLink()
		if err != nil {
			return "", err
		}

		return downloadLink, nil
	}

	return "", nil
}

type work struct {
	src  string
	dst  string
	t    *template.Template
	info os.FileInfo
}

/*
type cfg struct {
	//version
	DistroName   string
	Arch         string
	DebArch      string
	Package      string
	Dependencies string
}
*/

type stringList []string

func (ss *stringList) String() string {
	return strings.Join(*ss, ",")
}
func (ss *stringList) Set(v string) error {
	*ss = strings.Split(v, ",")
	return nil
}

var (
	SupportedPackages = []string{
		"kubeadm",
		"kubelet",
		"kubectl",
		"kubernetes-cni",
		"cri-tools",
	}

	KubernetesPackages = []string{
		"kubeadm",
		"kubelet",
		"kubectl",
	}

	SupportedArchitectures = []string{
		"amd64",
		"arm",
		"armhf",
		"arm64",
		"ppc64el",
		"ppc64le",
		"s390x",
	}
	providedArchitectures = stringList{}
	defaultArchitectures  = stringList{"amd64", "arm", "arm64", "ppc64le", "s390x"}
	// distros describes the Debian and Ubuntu versions that binaries will be built for.
	// Each distro build definition is currently symlinked to the most recent ubuntu build definition in the repo.
	// Build definitions should be kept up to date across release cycles, removing Debian/Ubuntu versions
	// that are no longer supported from the perspective of the OS distribution maintainers.
	//distros                 = stringList{"bionic", "xenial", "trusty", "stretch", "jessie", "sid"}
	kubeVersion              = ""
	revision                 = "00"
	providedDownloadLinkBase = ""
	builtins                 = map[string]interface{}{
		"date": func() string {
			return time.Now().Format(time.RFC1123Z)
		},
	}

	keepTmp = flag.Bool("keep-tmp", false, "keep tmp dir after build")

	KubeadmDependencies = []string{
		fmt.Sprintf("kubelet (>= %s)", minimumStableKubernetesVersion),
		fmt.Sprintf("kubectl (>= %s)", minimumStableKubernetesVersion),
		fmt.Sprintf("kubernetes-cni (>= %s)", minimumCNIVersion),
		fmt.Sprintf("cri-tools (>= %s)", minimumStableKubernetesVersion),
		"${misc:Depends}",
	}

	KubeletDependencies = []string{
		fmt.Sprintf("kubernetes-cni (>= %s)", minimumCNIVersion),
	}
)

func init() {
	flag.Var(&providedArchitectures, "arch", "Architectures to build for.")
	//flag.Var(&distros, "distros", "Distros to build for.")
	flag.StringVar(&kubeVersion, "kube-version", "", "Distros to build for.")
	flag.StringVar(&revision, "revision", "00", "Deb package revision.")
	flag.StringVar(&providedDownloadLinkBase, "release-download-link-base", defaultDownloadLinkBase, "Release download link base.")
}

func runCommand(pwd string, command string, cmdArgs ...string) error {
	cmd := exec.Command(command, cmdArgs...)
	if len(pwd) != 0 {
		cmd.Dir = pwd
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

/*
func (c cfg) run() error {
	log.Printf("!!!!!!!!! doing: %#v", c)
	var w []work

	srcdir := filepath.Join(c.DistroName, c.Package)
	dstdir, err := ioutil.TempDir(os.TempDir(), "debs")
	if err != nil {
		return err
	}
	if !*keepTmp {
		defer os.RemoveAll(dstdir)
	}

	// allow base package dir to by a symlink so we can reuse packages
	// that don't change between distros
	realSrcdir, err := filepath.EvalSymlinks(srcdir)
	if err != nil {
		return err
	}

	if err := filepath.Walk(realSrcdir, func(srcfile string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dstfile := filepath.Join(dstdir, srcfile[len(realSrcdir):])
		if dstfile == dstdir {
			return nil
		}
		if f.IsDir() {
			log.Printf(dstfile)
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

	dstParts := []string{"bin", string(c.Channel), c.DistroName}

	dstPath := filepath.Join(dstParts...)
	os.MkdirAll(dstPath, 0777)

	fileName := fmt.Sprintf("%s_%s-%s_%s.deb", c.Package, c.Version, c.Revision, c.DebArch)
	err = runCommand("", "mv", filepath.Join("/tmp", fileName), dstPath)
	if err != nil {
		return err
	}

	return nil
}
*/

/*
func walkBuilds(builds []build, f func(pkg, distro, arch string, v version) error) error {
	for _, a := range architectures {
		for _, b := range builds {
			for _, d := range b.Distros {
				for _, v := range b.Versions {
					// Populate the version if it doesn't exist
					if len(v.Version) == 0 && v.GetVersion != nil {
						var err error
						v.Version, err = v.GetVersion()
						if err != nil {
							return err
						}
					}

					// Populate the version if it doesn't exist
					if len(v.DownloadLinkBase) == 0 && v.GetDownloadLinkBase != nil {
						var err error
						v.DownloadLinkBase, err = v.GetDownloadLinkBase(v)
						if err != nil {
							return err
						}
					}

					if err := f(b.Package, d, a, v); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
*/

func fetchVersionFromLink(link string) (string, error) {
	res, err := http.Get(link)
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

func getStableKubeVersion() (string, error) {
	return fetchVersionFromLink("https://dl.k8s.io/release/stable.txt")
}

func getLatestKubeVersion() (string, error) {
	return fetchVersionFromLink("https://dl.k8s.io/release/latest.txt")
}

func getKubeCIVersion() (string, error) {
	latestVersion, err := getLatestKubeCIBuild()
	if err != nil {
		return "", err
	}

	// Replace the "+" with a "-" to make it semver-compliant
	return strings.Replace(latestVersion, "+", "-", 1), nil
}

func getCRIToolsLatestVersion() (string, error) {
	kv, err := getStableKubeVersion()
	if err != nil {
		return "", err
	}

	kubeSemver, err := semver.Parse(kv)
	if err != nil {
		return "", err
	}

	criToolsVersion := fmt.Sprintf("%s.%s.0", strconv.FormatUint(kubeSemver.Major, 10), strconv.FormatUint(kubeSemver.Minor, 10))
	if err != nil {
		return "", err
	}

	return criToolsVersion, nil
}

func getLatestKubeCIBuild() (string, error) {
	return fetchVersionFromLink("https://dl.k8s.io/ci-cross/latest.txt")
}

func getCIBuildsDownloadLink() (string, error) {
	latestCIVersion, err := getLatestKubeCIBuild()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://dl.k8s.io/ci-cross/v%s", latestCIVersion), nil
}

func getReleaseDownloadLink(version string) (string, error) {
	linkBase := GetDownloadLinkBase()
	link := fmt.Sprintf("%s/v%s", linkBase, version)

	_, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	return link, nil
}

func GetDownloadLinkBase() string {
	if providedDownloadLinkBase != "" {
		return providedDownloadLinkBase
	}

	return defaultDownloadLinkBase
}

func main() {
	flag.Parse()

	/*
		builds := []build{
			{
				Package: "kubectl",
				Distros: distros,
				Versions: []version{
					{
						GetVersion:          getStableKubeVersion,
						Revision:            revision,
						Channel:             RepositoryStable,
						GetDownloadLinkBase: getReleaseDownloadLink,
					},
					{
						GetVersion:          getLatestKubeVersion,
						Revision:            revision,
						Channel:             RepositoryUnstable,
						GetDownloadLinkBase: getReleaseDownloadLink,
					},
					{
						GetVersion:          getKubeCIVersion,
						Revision:            revision,
						Channel:             RepositoryNightly,
						GetDownloadLinkBase: getCIBuildsDownloadLink,
					},
				},
			},
			{
				Package: "kubelet",
				Distros: distros,
				Versions: []version{
					{
						GetVersion:          getStableKubeVersion,
						Revision:            revision,
						Channel:             RepositoryStable,
						GetDownloadLinkBase: getReleaseDownloadLink,
					},
					{
						GetVersion:          getLatestKubeVersion,
						Revision:            revision,
						Channel:             RepositoryUnstable,
						GetDownloadLinkBase: getReleaseDownloadLink,
					},
					{
						GetVersion:          getKubeCIVersion,
						Revision:            revision,
						Channel:             RepositoryNightly,
						GetDownloadLinkBase: getCIBuildsDownloadLink,
					},
				},
			},
			{
				Package: "kubernetes-cni",
				Distros: distros,
				Versions: []version{
					{
						Version:  minimumCNIVersion,
						Revision: revision,
						Channel:  RepositoryStable,
					},
					{
						Version:  minimumCNIVersion,
						Revision: revision,
						Channel:  RepositoryUnstable,
					},
					{
						Version:  minimumCNIVersion,
						Revision: revision,
						Channel:  RepositoryNightly,
					},
				},
			},
			{
				Package: "kubeadm",
				Distros: distros,
				Versions: []version{
					{
						GetVersion:          getStableKubeVersion,
						Revision:            revision,
						Channel:             RepositoryStable,
						GetDownloadLinkBase: getReleaseDownloadLink,
					},
					{
						GetVersion:          getLatestKubeVersion,
						Revision:            revision,
						Channel:             RepositoryUnstable,
						GetDownloadLinkBase: getReleaseDownloadLink,
					},
					{
						GetVersion:          getKubeCIVersion,
						Revision:            revision,
						Channel:             RepositoryNightly,
						GetDownloadLinkBase: getCIBuildsDownloadLink,
					},
				},
			},
			{
				Package: "cri-tools",
				Distros: distros,
				Versions: []version{
					{
						GetVersion: getCRIToolsLatestVersion,
						Revision:   revision,
						Channel:    RepositoryStable,
					},
					{
						GetVersion: getCRIToolsLatestVersion,
						Revision:   revision,
						Channel:    RepositoryUnstable,
					},
					{
						GetVersion: getCRIToolsLatestVersion,
						Revision:   revision,
						Channel:    RepositoryNightly,
					},
				},
			},
		}

		if kubeVersion != "" {
			getSpecifiedVersion := func() (string, error) {
				return kubeVersion, nil
			}
			builds = []build{
				{
					Package: "kubectl",
					Distros: distros,
					Versions: []version{
						{
							GetVersion:          getSpecifiedVersion,
							Revision:            revision,
							Channel:             RepositoryStable,
							GetDownloadLinkBase: getReleaseDownloadLink,
						},
					},
				},
				{
					Package: "kubelet",
					Distros: distros,
					Versions: []version{
						{
							GetVersion:          getSpecifiedVersion,
							Revision:            revision,
							Channel:             RepositoryStable,
							GetDownloadLinkBase: getReleaseDownloadLink,
						},
					},
				},
				{
					Package: "kubernetes-cni",
					Distros: distros,
					Versions: []version{
						{
							Version:  minimumCNIVersion,
							Revision: revision,
							Channel:  RepositoryStable,
						},
					},
				},
				{
					Package: "kubeadm",
					Distros: distros,
					Versions: []version{
						{
							GetVersion:          getSpecifiedVersion,
							Revision:            revision,
							Channel:             RepositoryStable,
							GetDownloadLinkBase: getReleaseDownloadLink,
						},
					},
				},
				{
					Package: "cri-tools",
					Distros: distros,
					Versions: []version{
						{
							GetVersion: getCRIToolsLatestVersion,
							Revision:   revision,
							Channel:    RepositoryStable,
						},
					},
				},
			}
		}

		if err := walkBuilds(builds, func(pkg, distro, arch string, v version) error {
			c := cfg{
				Package:    pkg,
				version:    v,
				DistroName: distro,
				Arch:       arch,
			}
			if c.Arch == "arm" {
				c.DebArch = "armhf"
			} else if c.Arch == "ppc64le" {
				c.DebArch = "ppc64el"
			} else {
				c.DebArch = c.Arch
			}

			var err error
			c.Dependencies = KubeadmDependencies
			if err != nil {
				log.Fatalf("error getting kubelet CNI Version: %v", err)
			}

			return c.run()
		}); err != nil {
			log.Fatalf("err: %v", err)
		}
	*/
}

// Helpers

func constructPackageSets(packages, arches []string, buildType BuildType, version, revision string) ([]PackageSet, error) {
	checkSupportedPackages(packages)
	checkSupportedArchitectures(arches)

	pkgSets := make([]PackageSet, 0, 5)

	for _, pkg := range packages {
		var pkgSet PackageSet
		pkgSet.Name = pkg
		pkgSet.Version = version
		pkgSet.Architectures = transformArchitectures(arches)

		for _, kpkg := range KubernetesPackages {
			if pkg == kpkg {
				switch buildType {
				case BuildTypeAll:
					pkgSet.BuildType = BuildTypeAll
				default:
					pkgSet.BuildType = buildType

					var rd RepositoryDefinition
					var err error
					rd.LinkBase, err = GetDownloadLink(pkgSet.BuildType, pkgSet.Version)
					if err != nil {
						return pkgSets, err
					}

					pkgSet.Repositories = append(pkgSet.Repositories, rd)
				}
			} else {
				pkgSet.BuildType = BuildTypeStable
			}
		}

		switch pkgSet.Name {
		case "kubeadm":
			pkgSet.Dependencies = KubeadmDependencies
		case "kubelet":
			pkgSet.Dependencies = KubeletDependencies
		}

		pkgSets = append(pkgSets, pkgSet)
	}

	/*
		for _, pkgSet := range b.PackageSets {
			for _, rd := range pkgSet.Repositories {
				var err error
				rd.LinkBase, err = rd.GetDownloadLink(pkgSet.Version)
			}

			switch pkgSet.Name {
			case "kubeadm":
				pkgSet.Dependencies = KubeadmDependencies
			case "kubelet":
				pkgSet.Dependencies = KubeletDependencies
			}
		}
	*/

	return pkgSets, nil
}

func checkSupportedPackages(packages []string) error {
	var unsupportedPackages []string
	for _, pkg := range packages {
		for i, sp := range SupportedPackages {
			if pkg == sp {
				break
			} else if i == len(SupportedPackages)-1 {
				unsupportedPackages = append(unsupportedPackages, pkg)
				fmt.Printf("%s is NOT supported\n", pkg)
			}
		}
	}

	if len(unsupportedPackages) > 0 {
		return errors.Errorf("The following packages are not supported by this tool: %s", strings.Join(unsupportedPackages, ","))
	}

	return nil
}

func checkSupportedArchitectures(arches []string) error {
	var unsupportedArchitectures []string
	for _, arch := range arches {
		for i, sa := range SupportedArchitectures {
			if arch == sa {
				break
			} else if i == len(SupportedArchitectures)-1 {
				unsupportedArchitectures = append(unsupportedArchitectures, arch)
				fmt.Printf("%s is NOT supported\n", arch)
			}
		}
	}

	if len(unsupportedArchitectures) > 0 {
		return errors.Errorf("The following architectures are not supported by this tool: %s", strings.Join(unsupportedArchitectures, ","))
	}

	return nil
}

func transformArchitectures(arches []string) []string {
	var validArches []string
	for _, arch := range arches {
		switch arch {
		case "arm":
			validArches = append(validArches, "armhf")
		case "ppc64le":
			validArches = append(validArches, "ppc64el")
		default:
			validArches = append(validArches, arch)
		}
	}

	return validArches
}

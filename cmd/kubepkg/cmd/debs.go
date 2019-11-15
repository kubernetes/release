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

package cmd

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

	"github.com/blang/semver"
	"github.com/google/go-github/v28/github"
	"github.com/spf13/cobra"
)

type debsOptions struct {
	branch    string
	masterRef string
	org       string
}

var debsOpts = &debsOptions{}

// debsCmd represents the base command when called without any subcommands
var debsCmd = &cobra.Command{
	Use:   "debs --branch <release-branch> [--ref <master-ref>] [--nomock] [--cleanup]",
	Short: "debs fast forwards a Kubernetes release branch",
	Long: `debs fast forwards a branch to a specified master object reference
(defaults to HEAD), and then prepares the branch as a Kubernetes release branch:

- Run hack/update-all.sh to ensure compliance of generated files`,
	Example: "kubepkg debs --branch release-1.16 39d0135e --ref HEAD --cleanup",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runDebs(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(debsCmd)
}

func runDebs() error {
	flag.Parse()

	// Replace the "+" with a "-" to make it semver-compliant
	kubeVersion = strings.TrimPrefix(kubeVersion, "v")

	builds, err := constructBuilds(packages, channels, kubeVersion, revision, cniVersion)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	if err := walkBuilds(builds); err != nil {
		log.Fatalf("err: %v", err)
	}

	// TODO: Fix return values
	return nil
}

func constructBuilds(packages, channels []string, kubeVersion, revision, cniVersion string) ([]build, error) {
	var builds []build

	for _, pkg := range packages {
		b := &build{
			Package: pkg,
		}

		for _, channel := range channels {
			packageDef := &packageDefinition{
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

func walkBuilds(builds []build) error {
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

func buildPackage(pkg, arch string, packageDef packageDefinition) error {
	c := cfg{
		packageDefinition: &packageDef,
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

	// TODO: Add note about this
	c.KubernetesVersion = strings.Replace(c.KubernetesVersion, "+", "-", 1)

	c.Version, err = getPackageVersion(packageDef)
	if err != nil {
		log.Fatalf("error getting package version: %v", err)
	}

	log.Printf("package version is %s", c.Version)

	c.KubeletCNIVersion = minimumCNIVersion

	c.Dependencies, err = getKubeadmDependencies(packageDef)
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

	dstParts := []string{"bin", string(c.Channel)}

	dstPath := filepath.Join(dstParts...)
	os.MkdirAll(dstPath, 0o777)

	fileName := fmt.Sprintf("%s_%s-%s_%s.deb", c.Package, c.Version, c.Revision, c.DebArch)
	err = runCommand("", "mv", filepath.Join("/tmp", fileName), dstPath)
	if err != nil {
		return err
	}

	return nil
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

func getPackageVersion(packageDef packageDefinition) (string, error) {
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

func getKubernetesVersion(packageDef packageDefinition) (string, error) {
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

func getCNIVersion(packageDef packageDefinition) (string, error) {
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

func getCRIToolsVersion(packageDef packageDefinition) (string, error) {
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

func getDownloadLinkBase(packageDef packageDefinition) (string, error) {
	switch packageDef.Channel {
	case ChannelNightly:
		return getCIBuildsDownloadLinkBase(packageDef)
	}

	return getReleaseDownloadLinkBase(packageDef)
}

func getCIBuildsDownloadLinkBase(packageDef packageDefinition) (string, error) {
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

func getReleaseDownloadLinkBase(packageDef packageDefinition) (string, error) {
	return fmt.Sprintf("%s/v%s", releaseDownloadLinkBase, packageDef.KubernetesVersion), nil
}

func getKubeadmDependencies(packageDef packageDefinition) (string, error) {
	deps := []string{
		fmt.Sprintf("kubelet (>= %s)", minimumKubernetesVersion),
		fmt.Sprintf("kubectl (>= %s)", minimumKubernetesVersion),
		fmt.Sprintf("kubernetes-cni (>= %s)", minimumCNIVersion),
		fmt.Sprintf("cri-tools (>= %s)", minimumCRIToolsVersion),
		"${misc:Depends}",
	}

	return strings.Join(deps, ", "), nil
}

func getCNIDownloadLink(packageDef packageDefinition, arch string) (string, error) {
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

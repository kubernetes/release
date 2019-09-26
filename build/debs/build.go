package main

import (
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
)

type DistributionType string

const (
	DistributionStable   DistributionType = "stable"
	DistributionTesting  DistributionType = "testing"
	DistributionUnstable DistributionType = "unstable"

	minimumKubernetesVersion = "1.13.0"
	minimumCNIVersion        = "0.7.5"
	pre117CNIVersion         = "0.7.5"

	defaultRevision = "0"

	packagesRootDir = "packages"

	kubeadmConf = "10-kubeadm.conf"
)

var (
	minimumCRIToolsVersion = minimumKubernetesVersion
	latestPackagesDir      = fmt.Sprintf("%s/%s", packagesRootDir, "latest")
)

type work struct {
	src  string
	dst  string
	t    *template.Template
	info os.FileInfo
}

type build struct {
	Package     string
	Definitions []packageDefinition
}

type packageDefinition struct {
	Name     string
	Version  string
	Revision string

	KubernetesVersion string
	KubeletCNIVersion string

	DownloadLinkBase         string
	Distribution             DistributionType
	KubeadmKubeletConfigFile string

	CNIDownloadLink string
}

type cfg struct {
	*packageDefinition
	Arch         string
	DebArch      string
	Package      string
	Dependencies string
}

type stringList []string

func (ss *stringList) String() string {
	return strings.Join(*ss, ",")
}
func (ss *stringList) Set(v string) error {
	*ss = strings.Split(v, ",")
	return nil
}

var (
	revision    string
	kubeVersion string
	cniVersion  string

	architectures           = stringList{"amd64", "arm", "arm64", "ppc64le", "s390x"}
	releaseDownloadLinkBase = "https://dl.k8s.io"

	builtins = map[string]interface{}{
		"date": func() string {
			return time.Now().Format(time.RFC1123Z)
		},
	}

	keepTmp = flag.Bool("keep-tmp", false, "keep tmp dir after build")
)

func init() {
	// TODO: Add flag support to build stable, testing, or unstable versions
	flag.Var(&architectures, "arch", "architectures to build for")
	flag.StringVar(&kubeVersion, "kube-version", "", "Kubernetes version to build")
	flag.StringVar(&revision, "revision", defaultRevision, "deb package revision.")
	flag.StringVar(&cniVersion, "cni-version", "", "CNI version to build")
	flag.StringVar(&releaseDownloadLinkBase, "release-download-link-base", "https://dl.k8s.io", "release download link base.")
}

func main() {
	flag.Parse()

	// Replace the "+" with a "-" to make it semver-compliant
	kubeVersion = strings.TrimPrefix(kubeVersion, "v")

	builds := []build{
		{
			Package: "kubectl",
			Definitions: []packageDefinition{
				{
					Revision:     revision,
					Distribution: DistributionStable,
				},
				{
					Revision:     revision,
					Distribution: DistributionTesting,
				},
				{
					Revision:     revision,
					Distribution: DistributionUnstable,
				},
			},
		},
		{
			Package: "kubelet",
			Definitions: []packageDefinition{
				{
					Revision:     revision,
					Distribution: DistributionStable,
				},
				{
					Revision:     revision,
					Distribution: DistributionTesting,
				},
				{
					Revision:     revision,
					Distribution: DistributionUnstable,
				},
			},
		},
		{
			Package: "kubernetes-cni",
			Definitions: []packageDefinition{
				{
					Version:      cniVersion,
					Revision:     revision,
					Distribution: DistributionStable,
				},
				{
					Version:      cniVersion,
					Revision:     revision,
					Distribution: DistributionTesting,
				},
				{
					Version:      cniVersion,
					Revision:     revision,
					Distribution: DistributionUnstable,
				},
			},
		},
		{
			Package: "kubeadm",
			Definitions: []packageDefinition{
				{
					Revision:     revision,
					Distribution: DistributionStable,
				},
				{
					Revision:     revision,
					Distribution: DistributionTesting,
				},
				{
					Revision:     revision,
					Distribution: DistributionUnstable,
				},
			},
		},
		{
			Package: "cri-tools",
			Definitions: []packageDefinition{
				{
					Revision:     revision,
					Distribution: DistributionStable,
				},
				{
					Revision:     revision,
					Distribution: DistributionTesting,
				},
				{
					Revision:     revision,
					Distribution: DistributionUnstable,
				},
			},
		},
	}

	if kubeVersion != "" {
		builds = []build{
			{
				Package: "kubectl",
				Definitions: []packageDefinition{
					{
						KubernetesVersion: kubeVersion,
						Revision:          revision,
						Distribution:      DistributionStable,
					},
				},
			},
			{
				Package: "kubelet",
				Definitions: []packageDefinition{
					{
						KubernetesVersion: kubeVersion,
						Revision:          revision,
						Distribution:      DistributionStable,
					},
				},
			},
			{
				Package: "kubernetes-cni",
				Definitions: []packageDefinition{
					{
						Version:           cniVersion,
						Revision:          revision,
						KubernetesVersion: kubeVersion,
						Distribution:      DistributionStable,
					},
				},
			},
			{
				Package: "kubeadm",
				Definitions: []packageDefinition{
					{
						KubernetesVersion: kubeVersion,
						Revision:          revision,
						Distribution:      DistributionStable,
					},
				},
			},
			{
				Package: "cri-tools",
				Definitions: []packageDefinition{
					{
						KubernetesVersion: kubeVersion,
						Revision:          revision,
						Distribution:      DistributionStable,
					},
				},
			},
		}
	}

	if err := walkBuilds(builds, func(pkg, arch string, packageDef packageDefinition) error {
		c := cfg{
			packageDefinition: &packageDef,
			Package:           pkg,
			Arch:              arch,
		}

		c.Name = pkg

		var err error

		// TODO: Allow building packages for a specific distro type

		if c.KubernetesVersion != "" {
			log.Printf("checking k8s semver")
			kubeSemver, err := semver.Parse(c.KubernetesVersion)
			if err != nil {
				log.Fatalf("could not parse k8s semver: %v", err)
			}

			kubeVersionString := kubeSemver.String()
			kubeVersionParts := strings.Split(kubeVersionString, ".")

			log.Printf("%v, len: %s", kubeVersionParts, len(kubeVersionParts))
			switch {
			case len(kubeVersionParts) > 4:
				c.Distribution = DistributionUnstable
			case len(kubeVersionParts) == 4:
				c.Distribution = DistributionTesting
			default:
				c.Distribution = DistributionStable
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
	}); err != nil {
		log.Fatalf("err: %v", err)
	}
}

func walkBuilds(builds []build, f func(pkg, arch string, packageDef packageDefinition) error) error {
	for _, a := range architectures {
		for _, b := range builds {
			for _, packageDef := range b.Definitions {
				if err := f(b.Package, a, packageDef); err != nil {
					return err
				}
			}
		}
	}
	return nil
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

	dstParts := []string{"bin", string(c.Distribution)}

	dstPath := filepath.Join(dstParts...)
	os.MkdirAll(dstPath, 0777)

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
	log.Printf("package name is %s", packageDef.Name)
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
		return packageDef.KubernetesVersion, nil
	}
	switch packageDef.Distribution {
	case DistributionTesting:
		return getTestingKubeVersion()
	case DistributionUnstable:
		return getUnstableKubeVersion()
	}

	return getStableKubeVersion()
}

func getStableKubeVersion() (string, error) {
	return fetchVersion("https://dl.k8s.io/release/stable.txt")
}

func getTestingKubeVersion() (string, error) {
	return fetchVersion("https://dl.k8s.io/release/latest.txt")
}

func getUnstableKubeVersion() (string, error) {
	latestCIVersion, err := getLatestKubeCIBuild()
	if err != nil {
		return "", err
	}

	return latestCIVersion, nil
}

func getLatestKubeCIBuild() (string, error) {
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
	log.Printf("CRI tools function")
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

	return fmt.Sprintf("%s.%s.0", criToolsMajor, criToolsMinor), nil
}

func getDownloadLinkBase(packageDef packageDefinition) (string, error) {
	switch packageDef.Distribution {
	case DistributionUnstable:
		return getCIBuildsDownloadLinkBase(packageDef)
	}

	return getReleaseDownloadLinkBase(packageDef)
}

func getCIBuildsDownloadLinkBase(packageDef packageDefinition) (string, error) {
	ciVersion := packageDef.KubernetesVersion
	if ciVersion == "" {
		var err error
		ciVersion, err = getLatestKubeCIBuild()
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

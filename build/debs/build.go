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
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver"
)

type DistributionType string

const (
	DistributionStable   DistributionType = "stable"
	DistributionUnstable DistributionType = "unstable"
	DistributionTesting  DistributionType = "testing"

	cniVersion      = "0.7.5"
	criToolsVersion = "1.13.0"

	packagesRootDir = "packages"

	kubeadmConf = "10-kubeadm.conf"
)

var latestPackagesDir = fmt.Sprintf("%s/%s", packagesRootDir, "latest")

type work struct {
	src, dst string
	t        *template.Template
	info     os.FileInfo
}

type build struct {
	Package  string
	Versions []version
}

type version struct {
	Version, Revision, DownloadLinkBase string
	Distribution                        DistributionType
	GetVersion                          func() (string, error)
	GetDownloadLinkBase                 func(v version) (string, error)
	KubeadmKubeletConfigFile            string
	KubeletCNIVersion                   string
}

type cfg struct {
	version
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
	architectures           = stringList{"amd64", "arm", "arm64", "ppc64le", "s390x"}
	kubeVersion             = ""
	revision                = "00"
	releaseDownloadLinkBase = "https://dl.k8s.io"

	builtins = map[string]interface{}{
		"date": func() string {
			return time.Now().Format(time.RFC1123Z)
		},
	}

	keepTmp = flag.Bool("keep-tmp", false, "keep tmp dir after build")
)

func init() {
	flag.Var(&architectures, "arch", "architectures to build for")
	flag.StringVar(&kubeVersion, "kube-version", "", "Kubernetes versions to build")
	flag.StringVar(&revision, "revision", "00", "deb package revision.")
	flag.StringVar(&releaseDownloadLinkBase, "release-download-link-base", "https://dl.k8s.io", "release download link base.")
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

func walkBuilds(builds []build, f func(pkg, arch string, v version) error) error {
	for _, a := range architectures {
		for _, b := range builds {
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

				if err := f(b.Package, a, v); err != nil {
					return err
				}
			}
		}
	}
	return nil
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

func getStableKubeVersion() (string, error) {
	return fetchVersion("https://dl.k8s.io/release/stable.txt")
}

func getLatestKubeVersion() (string, error) {
	return fetchVersion("https://dl.k8s.io/release/latest.txt")
}

func getLatestCIVersion() (string, error) {
	latestVersion, err := getLatestKubeCIBuild()
	if err != nil {
		return "", err
	}

	// Replace the "+" with a "-" to make it semver-compliant
	return strings.Replace(latestVersion, "+", "-", 1), nil
}

func getCRIToolsLatestVersion() (string, error) {
	return criToolsVersion, nil
}

func getLatestKubeCIBuild() (string, error) {
	return fetchVersion("https://dl.k8s.io/ci/k8s-master.txt")
}

func getCIBuildsDownloadLinkBase(_ version) (string, error) {
	latestCiVersion, err := getLatestKubeCIBuild()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://dl.k8s.io/ci/v%s", latestCiVersion), nil
}

func getReleaseDownloadLinkBase(v version) (string, error) {
	return fmt.Sprintf("%s/v%s", releaseDownloadLinkBase, v.Version), nil
}

func getKubeadmDependencies(v version) (string, error) {
	cniVersion, err := getKubeletCNIVersion(v)
	if err != nil {
		return "", err
	}

	deps := []string{
		"kubelet (>= 1.13.0)",
		"kubectl (>= 1.13.0)",
		fmt.Sprintf("kubernetes-cni (%s)", cniVersion),
		"${misc:Depends}",
	}
	sv, err := semver.Make(v.Version)
	if err != nil {
		return "", err
	}

	v1110, err := semver.Make("1.11.0-alpha.0")
	if err != nil {
		return "", err
	}

	if sv.GTE(v1110) {
		criToolsVersion, err := getCRIToolsVersion(v)
		if err != nil {
			return "", err
		}

		deps = append(deps, fmt.Sprintf("cri-tools (>= %s)", criToolsVersion))
		return strings.Join(deps, ", "), nil
	}
	return strings.Join(deps, ", "), nil
}

// CNI get bumped in 1.9, which is incompatible for kubelet<1.9.
// So we need to restrict the CNI version when install kubelet.
func getKubeletCNIVersion(v version) (string, error) {
	sv, err := semver.Make(v.Version)
	if err != nil {
		return "", err
	}

	v190, err := semver.Make("1.9.0-alpha.0")
	if err != nil {
		return "", err
	}

	if sv.GTE(v190) {
		return fmt.Sprintf(">= %s", cniVersion), nil
	}
	return fmt.Sprint("= 0.5.1"), nil
}

// getCRIToolsVersion assumes v coming in is >= 1.11.0-alpha.0
func getCRIToolsVersion(v version) (string, error) {
	sv, err := semver.Make(v.Version)
	if err != nil {
		return "", err
	}

	v1110, err := semver.Make("1.11.0-alpha.0")
	if err != nil {
		return "", err
	}
	v1121, err := semver.Make("1.12.1-alpha.0")
	if err != nil {
		return "", err
	}

	if sv.GTE(v1110) && sv.LT(v1121) {
		return "1.11.1", nil
	}
	return criToolsVersion, nil
}

func main() {
	flag.Parse()

	builds := []build{
		{
			Package: "kubectl",
			Versions: []version{
				{
					GetVersion:          getStableKubeVersion,
					Revision:            revision,
					Distribution:        DistributionStable,
					GetDownloadLinkBase: getReleaseDownloadLinkBase,
				},
				{
					GetVersion:          getLatestKubeVersion,
					Revision:            revision,
					Distribution:        DistributionUnstable,
					GetDownloadLinkBase: getReleaseDownloadLinkBase,
				},
				{
					GetVersion:          getLatestCIVersion,
					Revision:            revision,
					Distribution:        DistributionTesting,
					GetDownloadLinkBase: getCIBuildsDownloadLinkBase,
				},
			},
		},
		{
			Package: "kubelet",
			Versions: []version{
				{
					GetVersion:          getStableKubeVersion,
					Revision:            revision,
					Distribution:        DistributionStable,
					GetDownloadLinkBase: getReleaseDownloadLinkBase,
				},
				{
					GetVersion:          getLatestKubeVersion,
					Revision:            revision,
					Distribution:        DistributionUnstable,
					GetDownloadLinkBase: getReleaseDownloadLinkBase,
				},
				{
					GetVersion:          getLatestCIVersion,
					Revision:            revision,
					Distribution:        DistributionTesting,
					GetDownloadLinkBase: getCIBuildsDownloadLinkBase,
				},
			},
		},
		{
			Package: "kubernetes-cni",
			Versions: []version{
				{
					Version:      cniVersion,
					Revision:     revision,
					Distribution: DistributionStable,
				},
				{
					Version:      cniVersion,
					Revision:     revision,
					Distribution: DistributionUnstable,
				},
				{
					Version:      cniVersion,
					Revision:     revision,
					Distribution: DistributionTesting,
				},
			},
		},
		{
			Package: "kubeadm",
			Versions: []version{
				{
					GetVersion:          getStableKubeVersion,
					Revision:            revision,
					Distribution:        DistributionStable,
					GetDownloadLinkBase: getReleaseDownloadLinkBase,
				},
				{
					GetVersion:          getLatestKubeVersion,
					Revision:            revision,
					Distribution:        DistributionUnstable,
					GetDownloadLinkBase: getReleaseDownloadLinkBase,
				},
				{
					GetVersion:          getLatestCIVersion,
					Revision:            revision,
					Distribution:        DistributionTesting,
					GetDownloadLinkBase: getCIBuildsDownloadLinkBase,
				},
			},
		},
		{
			Package: "cri-tools",
			Versions: []version{
				{
					GetVersion:   getCRIToolsLatestVersion,
					Revision:     revision,
					Distribution: DistributionStable,
				},
				{
					GetVersion:   getCRIToolsLatestVersion,
					Revision:     revision,
					Distribution: DistributionUnstable,
				},
				{
					GetVersion:   getCRIToolsLatestVersion,
					Revision:     revision,
					Distribution: DistributionTesting,
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
				Versions: []version{
					{
						GetVersion:          getSpecifiedVersion,
						Revision:            revision,
						Distribution:        DistributionStable,
						GetDownloadLinkBase: getReleaseDownloadLinkBase,
					},
				},
			},
			{
				Package: "kubelet",
				Versions: []version{
					{
						GetVersion:          getSpecifiedVersion,
						Revision:            revision,
						Distribution:        DistributionStable,
						GetDownloadLinkBase: getReleaseDownloadLinkBase,
					},
				},
			},
			{
				Package: "kubernetes-cni",
				Versions: []version{
					{
						Version:      cniVersion,
						Revision:     revision,
						Distribution: DistributionStable,
					},
				},
			},
			{
				Package: "kubeadm",
				Versions: []version{
					{
						GetVersion:          getSpecifiedVersion,
						Revision:            revision,
						Distribution:        DistributionStable,
						GetDownloadLinkBase: getReleaseDownloadLinkBase,
					},
				},
			},
			{
				Package: "cri-tools",
				Versions: []version{
					{
						GetVersion:   getCRIToolsLatestVersion,
						Revision:     revision,
						Distribution: DistributionStable,
					},
				},
			},
		}
	}

	if err := walkBuilds(builds, func(pkg, arch string, v version) error {
		c := cfg{
			Package: pkg,
			version: v,
			Arch:    arch,
		}
		if c.Arch == "arm" {
			c.DebArch = "armhf"
		} else if c.Arch == "ppc64le" {
			c.DebArch = "ppc64el"
		} else {
			c.DebArch = c.Arch
		}

		c.KubeadmKubeletConfigFile = kubeadmConf

		var err error
		c.Dependencies, err = getKubeadmDependencies(v)
		if err != nil {
			log.Fatalf("error getting kubeadm dependencies: %v", err)
		}

		c.KubeletCNIVersion, err = getKubeletCNIVersion(v)
		if err != nil {
			log.Fatalf("error getting kubelet CNI Version: %v", err)
		}

		return c.run()
	}); err != nil {
		log.Fatalf("err: %v", err)
	}
}

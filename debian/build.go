package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"
)

type work struct {
	src, dst string
	t        *template.Template
	info     os.FileInfo
}

type build struct {
	Package  string
	Distros  []string
	Versions []version
}

type version struct {
	Version, Revision string
	Stable            bool
}

type cfg struct {
	version
	DistroName, Arch, DebArch, Package string
}

var (
	architectures = []string{"amd64", "arm", "arm64"}
	serverDistros = []string{"xenial"}
	allDistros    = []string{"xenial", "jessie", "precise", "sid", "stretch", "trusty",
		"utopic", "vivid", "wheezy", "wily", "yakkety"}

	builtins = map[string]interface{}{
		"date": func() string {
			return time.Now().Format(time.RFC1123Z)
		},
	}

	keepTmp = flag.Bool("keep_tmp", false, "keep tmp dir after build")
)

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

	dstParts := []string{"bin"}
	if c.Stable {
		dstParts = append(dstParts, "stable")
	} else {
		dstParts = append(dstParts, "unstable")
	}
	dstParts = append(dstParts, c.DistroName)

	dstPath := filepath.Join(dstParts...)
	os.MkdirAll(dstPath, 0777)

	fileName := fmt.Sprintf("%s_%s-%s_%s.deb", c.Package, c.Version, c.Revision, c.DebArch)
	err = runCommand("", "mv", filepath.Join("/tmp", fileName), dstPath)
	if err != nil {
		return err
	}

	return nil
}

func walkBuilds(builds []build, f func(pkg, distro, arch string, v version) error) error {
	for _, a := range architectures {
		for _, b := range builds {
			for _, d := range b.Distros {
				for _, v := range b.Versions {
					if err := f(b.Package, d, a, v); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func main() {
	flag.Parse()

	builds := []build{
		{
			Package: "kubectl",
			Distros: allDistros,
			Versions: []version{
				{
					Version:  "1.4.0",
					Revision: "00",
					Stable:   true,
				},
				{
					Version:  "1.4.0",
					Revision: "00",
					Stable:   false,
				},
			},
		},
		{
			Package: "kubelet",
			Distros: serverDistros,
			Versions: []version{
				{
					Version:  "1.4.0",
					Revision: "01",
					Stable:   true,
				},
				{
					Version:  "1.4.0",
					Revision: "01",
					Stable:   false,
				},
			},
		},
		{
			Package: "kubernetes-cni",
			Distros: serverDistros,
			Versions: []version{
				{
					Version:  "0.3.0.1-07a8a2",
					Revision: "00",
					Stable:   true,
				},
				{
					Version:  "0.3.0.1-07a8a2",
					Revision: "00",
					Stable:   true,
				},
			},
		},
		{
			Package: "kubeadm",
			Distros: serverDistros,
			Versions: []version{
				{
					Version:  "1.5.0-alpha.0-1534-gcf7301f",
					Revision: "00",
					Stable:   true,
				},
				{
					Version:  "1.5.0-alpha.0-1534-gcf7301f",
					Revision: "00",
					Stable:   false,
				},
			},
		},
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
		} else {
			c.DebArch = c.Arch
		}
		return c.run()
	}); err != nil {
		log.Fatalf("err: %v", err)
	}
}

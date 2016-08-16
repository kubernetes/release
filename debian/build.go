package main

import (
	"flag"
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

type cfg struct {
	Version, DistroName, Arch, Package, Revision string
}

var (
	builtins = map[string]interface{}{
		"date": func() string {
			return time.Now().Format(time.RFC1123Z)
		},
	}
)

func (c cfg) run() error {
	log.Printf("!!!!!!!!! doing: %#v", c)
	var w []work

	srcdir := filepath.Join(c.DistroName, c.Package)
	dstdir, err := ioutil.TempDir(os.TempDir(), "debs")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dstdir)

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

	cmd := exec.Command("dpkg-buildpackage", "-us", "-uc", "-b")
	cmd.Dir = dstdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	var (
		c = cfg{
			Revision: "00",
		}
	)
	flag.StringVar(&c.Version, "version", c.Version, "version")
	flag.StringVar(&c.DistroName, "distro_name", c.DistroName, "distro name")
	flag.StringVar(&c.Arch, "arch", c.Arch, "arch")
	flag.StringVar(&c.Package, "package", c.Package, "package")
	flag.StringVar(&c.Revision, "revision", c.Revision, "revision")
	flag.Parse()

	if err := c.run(); err != nil {
		log.Fatalf("err: %v", err)
	}
}

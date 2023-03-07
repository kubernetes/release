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
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// DownloadAndArchiveBinaries downloads and archives binaries for requested
// packages. The archive structure is according to the format required by
// spec files and Open Build Service.
func (c *Client) DownloadAndArchiveBinaries(pkgBuilder *PackageBuilder) error {
	if pkgBuilder == nil {
		return errors.New("package builder cannot be nil")
	}

	logrus.Info("Downloading binaries...")

	for _, pkg := range pkgBuilder.Definitions {
		for _, arch := range pkgBuilder.Architectures {
			logrus.Infof("Downloading %s/%s...", arch, pkg.Name)

			dlRootPath := filepath.Join(pkgBuilder.OutputDir, pkg.Name, arch)
			err := os.MkdirAll(dlRootPath, os.FileMode(0o755))
			if err != nil {
				if !os.IsExist(err) {
					return fmt.Errorf("creating directory to download %s: %w", pkg.Name, err)
				}
			}

			switch pkg.Name {
			case kubernetesCNIPackage:
				cniURL, err := GetCNIDownloadLink(pkg.Version, arch)
				if err != nil {
					return fmt.Errorf("parsing cni url: %w", err)
				}

				dlPath := filepath.Join(dlRootPath, "kubernetes-cni.tar.gz")
				if err := c.downloadBinaryFromURL(cniURL, dlPath, true); err != nil {
					return fmt.Errorf("downloading cni binaries: %w", err)
				}
			case criToolsPackage:
				criURL, err := GetCRIToolsDownloadLink(pkg.Version, arch)
				if err != nil {
					return fmt.Errorf("parsing cri-tools url: %w", err)
				}

				dlPath := filepath.Join(dlRootPath, "cri-tools.tar.gz")
				if err := c.downloadBinaryFromURL(criURL, dlPath, true); err != nil {
					return fmt.Errorf("downloading cri-tools binaries: %w", err)
				}
			default:
				kubeURL, err := url.JoinPath(pkgBuilder.DownloadLinkBase, arch, pkg.Name)
				if err != nil {
					return fmt.Errorf("parsing %s url: %w", pkg.Name, err)
				}

				dlPath := filepath.Join(dlRootPath, pkg.Name)
				if err := c.downloadBinaryFromURL(kubeURL, dlPath, false); err != nil {
					return fmt.Errorf("downloading %s binaries: %w", pkg.Name, err)
				}
			}

			logrus.Infof("Successfully downloaded %s/%s.", arch, pkg.Name)
		}

		logrus.Infof("Archiving binaries for %s...", pkg.Name)

		archiveSrc := filepath.Join(pkgBuilder.OutputDir, pkg.Name)
		archiveDst := filepath.Join(pkgBuilder.OutputDir, pkg.Name, fmt.Sprintf("%s_%s.orig.tar.gz", pkg.Name, pkg.Version))
		if err := createTarGzArchive(archiveSrc, archiveDst); err != nil {
			return fmt.Errorf("creating archive: %w", err)
		}

		logrus.Infof("Successfully archived binaries for %s!", pkg.Name)
	}

	logrus.Info("Binaries for all packages downloaded and archived!")

	return nil
}

func (c *Client) downloadBinaryFromURL(downloadURL, path string, extractTgz bool) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating download destination file: %w", err)
	}
	defer out.Close()

	resp, err := c.impl.GetRequest(downloadURL)
	if err != nil {
		return fmt.Errorf("downloading binary: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading binary: status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	if !extractTgz {
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("writing downloaded binary: %w", err)
		}
	} else {
		if err := extractTarGz(resp.Body, filepath.Dir(path)); err != nil {
			return fmt.Errorf("extracting .tar.gz archive: %w", err)
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing extracted archive: %w", err)
		}
	}

	return nil
}

func extractTarGz(gzipStream io.Reader, path string) error {
	gzipReader, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("creating gz reader failed: %w", err)
	}

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("unpacking .tar.gz archive: %w", err)
		}

		sp, err := sanitizeArchivePath(path, header.Name)
		if err != nil {
			return fmt.Errorf("sanitizing archive path: %w", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(sp, 0o755); err != nil {
				if !os.IsExist(err) {
					return fmt.Errorf("creating directory to extract: %w", err)
				}
			}
		case tar.TypeReg:
			outFile, err := os.Create(sp)
			if err != nil {
				return fmt.Errorf("creating file to extract: %w", err)
			}

			// This is to mitigate "G110: Potential DoS vulnerability via decompression bomb".
			for {
				_, err := io.CopyN(outFile, tarReader, 1024)
				if err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
			}

			outFile.Close()
		default:
			return fmt.Errorf("unknown file to extract %q (%q)", header.Typeflag, header.Name)
		}
	}

	return nil
}

func createTarGzArchive(src, dest string) error {
	// Create the archive file
	archiveFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating .tar.gz archive %q: %w", dest, err)
	}
	defer archiveFile.Close()

	// Create a gzip writer for the archive file
	gzipWriter, err := gzip.NewWriterLevel(archiveFile, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("creating gzip writer: %w", err)
	}
	defer gzipWriter.Close()

	// Create a tar writer for the gzip writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk through the directory and add files to the archive
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dest {
			return nil
		}
		// We don't compress spec files.
		if strings.HasSuffix(path, ".spec") {
			return nil
		}

		// Create a new tar header for the file
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return fmt.Errorf("getting tar file info header: %w", err)
		}

		// Set the header name to the relative path of the file
		header.Name, err = filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("getting relative destination path: %w", err)
		}

		// Write the header to the tar writer
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("writing tar header: %w", err)
		}

		// If the file is not a directory, write its contents to the tar writer
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}

			if err := cleanupFile(path); err != nil {
				return fmt.Errorf("cleaning up archive source file: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("archiving binaries: %w", err)
	}

	return nil
}

func cleanupFile(path string) error {
	// We don't clean up spec files.
	if strings.HasSuffix(path, ".spec") {
		return nil
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("cleaning up file: %w", err)
	}

	dirEnt, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	if len(dirEnt) == 0 {
		if err := os.RemoveAll(filepath.Dir(path)); err != nil {
			return fmt.Errorf("cleaning up root directory: %w", err)
		}
	}

	return nil
}

// sanitizeArchivePath sanitizes archive file path to prevent
// "G305: Zip Slip vulnerability" lint error.
// Ref: https://github.com/securego/gosec/issues/324#issuecomment-935927967
func sanitizeArchivePath(d, t string) (v string, err error) {
	v = filepath.Join(d, t)
	if strings.HasPrefix(v, filepath.Clean(d)) {
		return v, nil
	}

	return "", fmt.Errorf("%s: %s", "content filepath is tainted", t)
}

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
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/blang/semver/v4"
	"k8s.io/release/pkg/obs/metadata"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-sdk/object"
	khttp "sigs.k8s.io/release-utils/http"
	"sigs.k8s.io/release-utils/tar"
	"sigs.k8s.io/release-utils/util"
)

type defaultImpl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . impl
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt specsfakes/fake_impl.go > specsfakes/_fake_impl.go && mv specsfakes/_fake_impl.go specsfakes/fake_impl.go"
type impl interface {
	GetKubeVersion(versionType release.VersionType) (string, error)
	GetRequest(url string) (*http.Response, error)
	HeadRequest(url string) (*http.Response, error)
	CreateFile(name string) (*os.File, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	Mkdir(path string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	RemoveFile(name string) error
	RemoveAll(path string) error
	IsExist(err error) bool
	Stat(name string) (os.FileInfo, error)
	Walk(root string, fn filepath.WalkFunc) error
	Compress(tarFilePath, tarContentsPath string, excludes ...*regexp.Regexp) error
	Extract(tarFilePath, destinationPath string) error
	GCSCopyToLocal(gcsPath, dst string) error
	TagStringToSemver(tag string) (semver.Version, error)
	TrimTagPrefix(tag string) string
	LoadPackageMetadata(path string) (metadata.PackageMetadataList, error)
}

func (d *defaultImpl) GetKubeVersion(versionType release.VersionType) (string, error) {
	return release.NewVersion().GetKubeVersion(versionType)
}

func (d *defaultImpl) GetRequest(url string) (*http.Response, error) {
	return khttp.NewAgent().WithTimeout(3 * time.Minute).GetRequest(url)
}

func (d *defaultImpl) HeadRequest(url string) (*http.Response, error) {
	return khttp.NewAgent().WithTimeout(3 * time.Minute).HeadRequest(url)
}

func (d *defaultImpl) CreateFile(name string) (*os.File, error) {
	return os.Create(name)
}

func (d *defaultImpl) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (d *defaultImpl) Mkdir(path string, perm os.FileMode) error {
	return os.Mkdir(path, perm)
}

func (d *defaultImpl) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (d *defaultImpl) IsExist(err error) bool {
	return os.IsExist(err)
}

func (d *defaultImpl) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (d *defaultImpl) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

func (d *defaultImpl) RemoveFile(name string) error {
	return os.Remove(name)
}

func (d *defaultImpl) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (d *defaultImpl) Compress(tarFilePath, tarContentsPath string, excludes ...*regexp.Regexp) error {
	return tar.CompressWithoutPreservingPath(tarFilePath, tarContentsPath, excludes...)
}

func (d *defaultImpl) Extract(tarFilePath, destinationPath string) error {
	return tar.Extract(tarFilePath, destinationPath)
}

func (d *defaultImpl) GCSCopyToLocal(gcsPath, dst string) error {
	return object.NewGCS().CopyToLocal(gcsPath, dst)
}

func (d *defaultImpl) TagStringToSemver(tag string) (semver.Version, error) {
	return util.TagStringToSemver(tag)
}

func (d *defaultImpl) TrimTagPrefix(tag string) string {
	return util.TrimTagPrefix(tag)
}

func (d *defaultImpl) LoadPackageMetadata(path string) (metadata.PackageMetadataList, error) {
	return metadata.LoadPackageMetadata(path)
}

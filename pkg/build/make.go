/*
Copyright 2020 The Kubernetes Authors.

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

package build

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/release"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Make is the main structure for building Kubernetes releases.
type Make struct {
	impl
}

// New creates a new `Build` instance.
func NewMake() *Make {
	return &Make{&defaultMakeImpl{}}
}

// SetImpl can be used to set the internal implementation.
func (m *Make) SetImpl(impl impl) {
	m.impl = impl
}

type defaultMakeImpl struct{}

//counterfeiter:generate . impl
type impl interface {
	OpenRepo(repoPath string) (*git.Repo, error)
	Checkout(repo *git.Repo, rev string) error
	Command(cmd string, args ...string) error
	Rename(from, to string) error
}

func (d *defaultMakeImpl) OpenRepo(repoPath string) (*git.Repo, error) {
	return git.OpenRepo(repoPath)
}

func (d *defaultMakeImpl) Checkout(repo *git.Repo, rev string) error {
	return repo.Checkout(rev)
}

func (d *defaultMakeImpl) Command(cmd string, args ...string) error {
	return command.New(cmd, args...).RunSuccess()
}

func (d *defaultMakeImpl) Rename(from, to string) error {
	return os.Rename(from, to)
}

// MakeCross cross compiles Kubernetes binaries for the provided `versions` and
// `repoPath`.
func (m *Make) MakeCross(version string) error {
	repo, err := m.impl.OpenRepo(".")
	if err != nil {
		return errors.Wrap(err, "open Kubernetes repository")
	}

	logrus.Infof("Checking out version %s", version)
	if err := m.impl.Checkout(repo, version); err != nil {
		return errors.Wrapf(err, "checking out version %s", version)
	}

	logrus.Info("Building binaries")
	if err := m.impl.Command(
		"make",
		"cross-in-a-container",
		fmt.Sprintf("KUBE_DOCKER_IMAGE_TAG=%s", version),
	); err != nil {
		return errors.Wrapf(err, "build version %s", version)
	}

	newBuildDir := fmt.Sprintf("%s-%s", release.BuildDir, version)
	logrus.Infof("Moving build output to %s", newBuildDir)
	if err := m.impl.Rename(release.BuildDir, newBuildDir); err != nil {
		return errors.Wrap(err, "move build output")
	}

	logrus.Info("Building package tarballs")
	if err := m.impl.Command(
		"make",
		"package-tarballs",
		fmt.Sprintf("KUBE_DOCKER_IMAGE_TAG=%s", version),
		fmt.Sprintf("OUT_DIR=%s", newBuildDir),
	); err != nil {
		return errors.Wrap(err, "build package tarballs")
	}

	return nil
}

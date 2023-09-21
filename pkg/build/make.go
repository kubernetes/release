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

	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/release"
	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/command"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt buildfakes/fake_impl.go > buildfakes/_fake_impl.go && mv buildfakes/_fake_impl.go buildfakes/fake_impl.go"

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
		return fmt.Errorf("open Kubernetes repository: %w", err)
	}

	logrus.Infof("Checking out version %s", version)
	if err := m.impl.Checkout(repo, version); err != nil {
		return fmt.Errorf("checking out version %s: %w", version, err)
	}

	// Unset the build memory requirement for parallel builds
	// TODO: Remove this once the 1.20 release reaches EOL.
	const buildMemoryKey = "KUBE_PARALLEL_BUILD_MEMORY"
	logrus.Infof("Setting %s to force serial build", buildMemoryKey)
	os.Setenv(buildMemoryKey, "32")

	logrus.Info("Building binaries")
	if err := m.impl.Command(
		"make",
		"cross-in-a-container",
		fmt.Sprintf("KUBE_DOCKER_IMAGE_TAG=%s", version),
	); err != nil {
		return fmt.Errorf("build version %s: %w", version, err)
	}

	newBuildDir := fmt.Sprintf("%s-%s", release.BuildDir, version)
	logrus.Infof("Moving build output to %s", newBuildDir)
	if err := m.impl.Rename(release.BuildDir, newBuildDir); err != nil {
		return fmt.Errorf("move build output: %w", err)
	}

	logrus.Info("Building package tarballs")
	if err := m.impl.Command(
		"make",
		"package-tarballs",
		fmt.Sprintf("KUBE_DOCKER_IMAGE_TAG=%s", version),
		fmt.Sprintf("OUT_DIR=%s", newBuildDir),
	); err != nil {
		return fmt.Errorf("build package tarballs: %w", err)
	}

	return nil
}

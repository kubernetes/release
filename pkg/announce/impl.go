/*
Copyright 2024 The Kubernetes Authors.

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

package announce

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/util"

	"k8s.io/release/pkg/kubecross"
)

type defaultImpl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . impl
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt announcefakes/fake_impl.go > announcefakes/_fake_impl.go && mv announcefakes/_fake_impl.go announcefakes/fake_impl.go"

type impl interface {
	Create(workDir, subject, message string) error
	GetGoVersion(tag string) (string, error)
	ReadChangelogFile(file string) ([]byte, error)
}

func (i *defaultImpl) Create(workDir, subject, message string) error {
	subjectFile := filepath.Join(workDir, subjectFile)
	//nolint:gosec // TODO(gosec): G306: Expect WriteFile permissions to be
	// 0600 or less
	if err := os.WriteFile(
		subjectFile, []byte(subject), 0o755,
	); err != nil {
		return fmt.Errorf(
			"writing subject to file %s: %w",
			subjectFile,
			err,
		)
	}
	logrus.Debugf("Wrote file %s", subjectFile)

	announcementFile := filepath.Join(workDir, announcementFile)
	//nolint:gosec // TODO(gosec): G306: Expect WriteFile permissions to be
	// 0600 or less
	if err := os.WriteFile(
		announcementFile, []byte(message), 0o755,
	); err != nil {
		return fmt.Errorf(
			"writing announcement to file %s: %w",
			announcementFile,
			err,
		)
	}
	logrus.Debugf("Wrote file %s", announcementFile)

	return nil
}

// GetGoVersion runs kube-cross container and go version inside it.
// We're running kube-cross container because it's not guaranteed that
// k8s-cloud-builder container will be running the same Go version as
// the kube-cross container used to build the release.
func (i *defaultImpl) GetGoVersion(tag string) (string, error) {
	semver, err := util.TagStringToSemver(tag)
	if err != nil {
		return "", fmt.Errorf("parse version tag: %w", err)
	}

	branch := fmt.Sprintf("release-%d.%d", semver.Major, semver.Minor)
	kc := kubecross.New()
	kubecrossVer, err := kc.ForBranch(branch)
	if err != nil {
		kubecrossVer, err = kc.Latest()
		if err != nil {
			return "", fmt.Errorf("get kubecross version: %w", err)
		}
	}

	kubecrossImg := "registry.k8s.io/build-image/kube-cross:" + kubecrossVer

	res, err := command.New(
		"docker", "run", "--rm", kubecrossImg, "go", "version",
	).RunSilentSuccessOutput()
	if err != nil {
		return "", fmt.Errorf("get go version: %w", err)
	}

	versionRegex := regexp.MustCompile(`^?(\d+)(\.\d+)?(\.\d+)`)
	return versionRegex.FindString(strings.TrimSpace(res.OutputTrimNL())), nil
}

func (i *defaultImpl) ReadChangelogFile(file string) ([]byte, error) {
	return os.ReadFile(file)
}

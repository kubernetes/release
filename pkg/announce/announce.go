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

package announce

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/util"

	"k8s.io/release/pkg/kubecross"
)

const (
	announcementFile = "announcement.html"
	subjectFile      = "announcement-subject.txt"
)

const branchAnnouncement = `Kubernetes Community,
<p>
Kubernetes' %s branch has been created.
<p>
The release owner will be sending updates on how to interact with this branch
shortly. The <a href=https://git.k8s.io/community/contributors/devel/sig-release/cherry-picks.md>Cherrypick
Guide</A> has some general guidance on how things will proceed.
<p>
Announced by your
<a href=https://git.k8s.io/sig-release/release-managers.md>Kubernetes Release
Managers</A>.
`

const releaseAnnouncement = `Kubernetes Community,
<p>
Kubernetes <b>%s</b> has been built and pushed using Golang version <b>%s</b>.
<p>
The release notes have been updated in
<a href=https://git.k8s.io/kubernetes/%s>%s</a>, with a pointer to them on
<a href=https://github.com/kubernetes/kubernetes/releases/tag/%s>GitHub</a>:
<p>
<hr>
%s
<hr>
<p><br>
Contributors, the
<a href=https://git.k8s.io/kubernetes/%s>%s</a> has been bootstrapped with
%s release notes and you may edit now as needed.
<p><br><br>
Published by your
<a href=https://git.k8s.io/sig-release/release-managers.md>Kubernetes Release
Managers</a>.
`

func CreateForBranch(opts *Options) error {
	logrus.Infof(
		"Creating %s branch announcement in %s",
		opts.branch, opts.workDir,
	)

	if err := create(
		opts.workDir,
		fmt.Sprintf("Kubernetes %s branch has been created", opts.branch),
		fmt.Sprintf(branchAnnouncement, opts.branch),
	); err != nil {
		return errors.Wrap(err, "creating branch announcement")
	}

	// TODO: When we create a new branch, we notify the publishing-bot folks by
	// creating an issue for them (see anago)

	logrus.Infof("Branch announcement created")
	return nil
}

func CreateForRelease(opts *Options) error {
	logrus.Infof("Creating %s announcement in %s", opts.tag, opts.workDir)

	changelog := ""

	// Read the changelog from the specified file if we got one
	if opts.changelogFile != "" {
		changelogData, err := os.ReadFile(opts.changelogFile)
		if err != nil {
			return errors.Wrap(err, "reading changelog html file")
		}
		changelog = string(changelogData)
	}

	// ... unless it is overridden by passing the HTML directly
	if opts.changelogHTML != "" {
		changelog = opts.changelogHTML
	}

	logrus.Infof("Trying to get the Go version used to build %s...", opts.tag)
	goVersion, err := getGoVersion(opts.tag)
	if err != nil {
		return err
	}
	logrus.Infof("Found the following Go version: %s", goVersion)

	if err := create(
		opts.workDir,
		fmt.Sprintf("Kubernetes %s is live!", opts.tag),
		fmt.Sprintf(releaseAnnouncement,
			opts.tag, goVersion, opts.changelogPath,
			filepath.Base(opts.changelogPath), opts.tag, changelog,
			opts.changelogPath, filepath.Base(opts.changelogPath), opts.tag,
		),
	); err != nil {
		return errors.Wrap(err, "creating release announcement")
	}

	logrus.Infof("Release announcement created")
	return nil
}

func create(workDir, subject, message string) error {
	subjectFile := filepath.Join(workDir, subjectFile)
	if err := os.WriteFile(
		subjectFile, []byte(subject), 0o755,
	); err != nil {
		return errors.Wrapf(
			err, "writing subject to file %s", subjectFile,
		)
	}
	logrus.Debugf("Wrote file %s", subjectFile)

	announcementFile := filepath.Join(workDir, announcementFile)
	if err := os.WriteFile(
		announcementFile, []byte(message), 0o755,
	); err != nil {
		return errors.Wrapf(
			err, "writing announcement to file %s", announcementFile,
		)
	}
	logrus.Debugf("Wrote file %s", announcementFile)

	return nil
}

// getGoVersion runs kube-cross container and go version inside it.
// We're running kube-cross container because it's not guaranteed that
// k8s-cloud-builder container will be running the same Go version as
// the kube-cross container used to build the release.
func getGoVersion(tag string) (string, error) {
	semver, err := util.TagStringToSemver(tag)
	if err != nil {
		return "", errors.Wrap(err, "parse version tag")
	}

	branch := fmt.Sprintf("release-%d.%d", semver.Major, semver.Minor)
	kc := kubecross.New()
	kubecrossVer, err := kc.ForBranch(branch)
	if err != nil {
		kubecrossVer, err = kc.Latest()
		if err != nil {
			return "", errors.Wrap(err, "get kubecross version")
		}
	}

	kubecrossImg := fmt.Sprintf("k8s.gcr.io/build-image/kube-cross:%s", kubecrossVer)

	res, err := command.New(
		"docker", "run", "--rm", kubecrossImg, "go", "version",
	).RunSilentSuccessOutput()
	if err != nil {
		return "", errors.Wrap(err, "get go version")
	}

	versionRegex := regexp.MustCompile(`^?([0-9]+)(\.[0-9]+)?(\.[0-9]+)`)
	return versionRegex.FindString(strings.TrimSpace(res.OutputTrimNL())), nil
}

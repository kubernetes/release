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
	"errors"
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
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
<a href=https://git.k8s.io/website/content/en/releases/release-managers.md>Kubernetes Release
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
<a href=https://git.k8s.io/website/content/en/releases/release-managers.md>Kubernetes Release
Managers</a>.
`

type Announce struct {
	options *Options
	impl
}

// NewAnnounce returns a new Announce instance.
func NewAnnounce(opts *Options) *Announce {
	return &Announce{
		impl:    &defaultImpl{},
		options: opts,
	}
}

// SetImplementation sets the implementation to handle file operations.
func (a *Announce) SetImplementation(i impl) {
	a.impl = i
}

func (a *Announce) CreateForBranch() error {
	logrus.Infof(
		"Creating %s branch announcement in %s",
		a.options.branch, a.options.workDir,
	)

	if err := a.impl.Create(
		a.options.workDir,
		fmt.Sprintf("Kubernetes %s branch has been created", a.options.branch),
		fmt.Sprintf(branchAnnouncement, a.options.branch),
	); err != nil {
		return fmt.Errorf("creating branch announcement: %w", err)
	}

	// TODO: When we create a new branch, we notify the publishing-bot folks by
	// creating an issue for them (see anago)

	logrus.Infof("Branch announcement created")
	return nil
}

func (a *Announce) CreateForRelease() error {
	logrus.Infof("Creating %s announcement in %s", a.options.tag, a.options.workDir)

	changelog := ""

	// Read the changelog from the specified file if we got one
	if a.options.changelogFile != "" {
		changelogData, err := a.impl.ReadChangelogFile(a.options.changelogFile)
		if err != nil {
			return fmt.Errorf("reading changelog html file: %w", err)
		}
		if len(changelogData) == 0 {
			return fmt.Errorf("verifying that changelog html file '%s' is not empty", a.options.changelogFile)
		}
		changelog = string(changelogData)
	}

	// ... unless it is overridden by passing the HTML directly
	if a.options.changelogHTML != "" {
		changelog = a.options.changelogHTML
	}

	logrus.Infof("Trying to get the Go version used to build %s...", a.options.tag)
	goVersion, err := a.impl.GetGoVersion(a.options.tag)
	if err != nil {
		return err
	}
	if goVersion == "" {
		return errors.New("verifying Go version is not empty")
	}
	logrus.Infof("Found the following Go version: %s", goVersion)

	if err := a.impl.Create(
		a.options.workDir,
		fmt.Sprintf("Kubernetes %s is live!", a.options.tag),
		fmt.Sprintf(releaseAnnouncement,
			a.options.tag, goVersion, a.options.changelogPath,
			filepath.Base(a.options.changelogPath), a.options.tag, changelog,
			a.options.changelogPath, filepath.Base(a.options.changelogPath), a.options.tag,
		),
	); err != nil {
		return fmt.Errorf("creating release announcement: %w", err)
	}

	logrus.Infof("Release announcement created")
	return nil
}

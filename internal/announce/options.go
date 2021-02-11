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

type Options struct {
	// workDir is the directory where announcement.html and
	// announcement-subject.txt will be written
	workDir string
	// Release tag we will build the announcement for
	tag    string
	branch string
	// Changelog path is the path used to build the link to the changelog
	// in the announcement HTML. For example CHANGELOG/CHANGELOG-1.20.md
	changelogPath string
	// Changelog HTML is the changelog in HTML format. This will be
	// embedded in the announcement HTML, overrides changelogFile if
	// both are present.
	changelogHTML string
	// changelogFile is the path to an HTML file containing the changelog
	// which will be embedded in the announcement template
	changelogFile string
}

// NewOptions can be used to create a new Options instance
func NewOptions() *Options {
	return &Options{}
}

func (o *Options) WithWorkDir(workDir string) *Options {
	o.workDir = workDir
	return o
}

func (o *Options) WithTag(tag string) *Options {
	o.tag = tag
	return o
}

func (o *Options) WithBranch(branch string) *Options {
	o.branch = branch
	return o
}

func (o *Options) WithChangelogPath(changelogPath string) *Options {
	o.changelogPath = changelogPath
	return o
}

func (o *Options) WithChangelogHTML(changelogHTML string) *Options {
	o.changelogHTML = changelogHTML
	return o
}

func (o *Options) WithChangelogFile(changelogFile string) *Options {
	o.changelogFile = changelogFile
	return o
}

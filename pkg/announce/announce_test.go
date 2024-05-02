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

package announce_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/announce"
	"k8s.io/release/pkg/announce/announcefakes"
)

const (
	workdir       = "/workspace"
	changelogFile = "/workspace/src/release-notes.html"
	branch        = "release-1.30"
)

var err = errors.New("error")

var readChangelogFileStub = func(s string) ([]byte, error) {
	if s == changelogFile {
		return []byte("<b>changelog contents</b>"), nil
	} else {
		return nil, err
	}
}

func TestCreateForBranch(t *testing.T) {
	for _, tc := range []struct {
		name        string
		workdir     string
		branch      string
		prepare     func(*announcefakes.FakeImpl)
		shouldError bool
	}{
		{
			name:    "create announcement successfully",
			workdir: workdir,
			branch:  branch,
			prepare: func(mock *announcefakes.FakeImpl) {
			},
			shouldError: false,
		},
		{
			name:    "fails to create announcement file",
			workdir: workdir,
			branch:  branch,
			prepare: func(mock *announcefakes.FakeImpl) {
				mock.CreateReturns(err)
			},
			shouldError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := announce.NewOptions().
				WithWorkDir(tc.workdir).
				WithBranch(tc.branch)

			an := announce.NewAnnounce(opts)
			mock := &announcefakes.FakeImpl{}
			tc.prepare(mock)
			an.SetImplementation(mock)

			err := an.CreateForBranch()
			if tc.shouldError {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestCreateForRelease(t *testing.T) {
	for _, tc := range []struct {
		name          string
		workdir       string
		changelogFile string
		tag           string
		prepare       func(*announcefakes.FakeImpl)
		shouldError   bool
	}{
		{
			name:          "create announcement successfully",
			workdir:       workdir,
			changelogFile: changelogFile,
			tag:           "1.30.0",
			prepare: func(mock *announcefakes.FakeImpl) {
				mock.GetGoVersionReturns("1.22.0", nil)
				mock.ReadChangelogFileCalls(readChangelogFileStub)
			},
			shouldError: false,
		},
		{
			name:          "fails to get go version",
			workdir:       workdir,
			changelogFile: changelogFile,
			tag:           "1.30.0",
			prepare: func(mock *announcefakes.FakeImpl) {
				mock.GetGoVersionReturns("", err)
				mock.ReadChangelogFileCalls(readChangelogFileStub)
			},
			shouldError: true,
		},
		{
			name:          "gets empty go version",
			workdir:       workdir,
			changelogFile: changelogFile,
			tag:           "1.30.0",
			prepare: func(mock *announcefakes.FakeImpl) {
				mock.GetGoVersionReturns("", nil)
				mock.ReadChangelogFileCalls(readChangelogFileStub)
			},
			shouldError: true,
		},
		{
			name:          "fails to create announcement file",
			workdir:       workdir,
			changelogFile: changelogFile,
			tag:           "1.30.0",
			prepare: func(mock *announcefakes.FakeImpl) {
				mock.GetGoVersionReturns("1.22.0", nil)
				mock.CreateReturns(err)
				mock.ReadChangelogFileCalls(readChangelogFileStub)
			},
			shouldError: true,
		},
		{
			name:          "fails to read changelog file",
			workdir:       workdir,
			changelogFile: changelogFile,
			tag:           "1.30.0",
			prepare: func(mock *announcefakes.FakeImpl) {
				mock.GetGoVersionReturns("1.22.0", nil)
				mock.ReadChangelogFileReturns(nil, err)
			},
			shouldError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := announce.NewOptions().
				WithWorkDir(tc.workdir).
				WithChangelogFile(tc.changelogFile).
				WithTag(tc.tag)

			an := announce.NewAnnounce(opts)
			mock := &announcefakes.FakeImpl{}
			tc.prepare(mock)
			an.SetImplementation(mock)

			err := an.CreateForRelease()
			if tc.shouldError {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

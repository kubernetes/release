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

package changelog_test

import (
	"errors"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/changelog"
	"k8s.io/release/pkg/changelog/changelogfakes"
	"k8s.io/release/pkg/github"
	"k8s.io/release/pkg/notes"
)

func TestRun(t *testing.T) {
	err := errors.New("")
	for _, tc := range []struct {
		prepare   func(*changelogfakes.FakeImpl, *changelog.Options)
		shouldErr bool
	}{
		{ // success new official
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
			},
			shouldErr: false,
		},
		{ // success new first alpha
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.TagStringToSemverReturns(semver.Version{
					Pre: []semver.PRVersion{
						{VersionStr: "alpha"},
						{VersionNum: 1},
					},
				}, nil)
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.GatherReleaseNotesReturns(&notes.ReleaseNotes{}, nil)
			},
			shouldErr: false,
		},
		{ // success new patch
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.TagStringToSemverReturns(semver.Version{
					Major: 1,
					Minor: 19,
					Patch: 3,
				}, nil)
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.GatherReleaseNotesReturns(&notes.ReleaseNotes{}, nil)
			},
			shouldErr: false,
		},
		{ // success new pre release
			prepare: func(mock *changelogfakes.FakeImpl, opts *changelog.Options) {
				mock.TagStringToSemverReturns(semver.Version{
					Major: 1,
					Minor: 19,
					Pre: []semver.PRVersion{
						{VersionStr: "beta"},
						{VersionNum: 1},
					},
				}, nil)
				mock.LatestGitHubTagsPerBranchReturns(github.TagsPerBranch{
					"release-1.19": "v1.19.0-beta.0",
				}, nil)
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.GatherReleaseNotesReturns(&notes.ReleaseNotes{}, nil)
			},
			shouldErr: false,
		},
		{ // TagStringToSemver failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.TagStringToSemverReturns(semver.Version{}, err)
			},
			shouldErr: true,
		},
		{ // OpenRepo failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.OpenRepoReturns(nil, err)
			},
			shouldErr: true,
		},
		{ // RevParse failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.RevParseReturns("", err)
			},
			shouldErr: true,
		},
		{ // CreateDownloadsTable failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.CreateDownloadsTableReturns(err)
			},
			shouldErr: true,
		},
		{ // GetURLResponse 0 failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.GetURLResponseReturnsOnCall(0, "", err)
			},
			shouldErr: true,
		},
		{ // GenerateTOC failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.GenerateTOCReturns("", err)
			},
			shouldErr: true,
		},
		{ // DependencyChanges failed
			prepare: func(mock *changelogfakes.FakeImpl, opts *changelog.Options) {
				opts.Dependencies = true
				mock.DependencyChangesReturns("", err)
			},
			shouldErr: true,
		},
		{ // CurrentBranch failed
			prepare: func(mock *changelogfakes.FakeImpl, opts *changelog.Options) {
				mock.CurrentBranchReturns("", err)
			},
			shouldErr: true,
		},
		{ // Checkout failed
			prepare: func(mock *changelogfakes.FakeImpl, opts *changelog.Options) {
				mock.CheckoutReturns(err)
			},
			shouldErr: true,
		},
		{ // WriteFile failed
			prepare: func(mock *changelogfakes.FakeImpl, opts *changelog.Options) {
				mock.WriteFileReturns(err)
			},
			shouldErr: true,
		},
		{ // Stat failed
			prepare: func(mock *changelogfakes.FakeImpl, opts *changelog.Options) {
				mock.StatReturns(nil, err)
			},
			shouldErr: true,
		},
		{ // ReadFile failed
			prepare: func(mock *changelogfakes.FakeImpl, opts *changelog.Options) {
				mock.ReadFileReturns(nil, err)
			},
			shouldErr: true,
		},
		{ // notoc marker failed
			prepare:   func(*changelogfakes.FakeImpl, *changelog.Options) {},
			shouldErr: true,
		},
		{ // GenerateTOC failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.GenerateTOCReturns("", err)
			},
			shouldErr: true,
		},
		{ // MarkdownToHTML failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.MarkdownToHTMLReturns(err)
			},
			shouldErr: true,
		},
		{ // ParseHTMLTemplate failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.ParseHTMLTemplateReturns(nil, err)
			},
			shouldErr: true,
		},
		{ // TemplateExecute failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.TemplateExecuteReturns(err)
			},
			shouldErr: true,
		},
		{ // Abs failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.AbsReturns("", err)
			},
			shouldErr: true,
		},
		{ // WriteFile failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.WriteFileReturns(err)
			},
			shouldErr: true,
		},
		{ // Add failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.AddReturns(err)
			},
			shouldErr: true,
		},
		{ // Commit failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.CommitReturns(err)
			},
			shouldErr: true,
		},
		{ // Checkout 1 failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.CheckoutReturnsOnCall(1, err)
			},
			shouldErr: true,
		},
		{ // Checkout 2 failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.CheckoutReturnsOnCall(2, err)
			},
			shouldErr: true,
		},
		{ // Rm failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.RmReturns(err)
			},
			shouldErr: true,
		},
		{ // GatherReleaseNotes failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.TagStringToSemverReturns(semver.Version{
					Pre: []semver.PRVersion{
						{VersionStr: "alpha"},
						{VersionNum: 1},
					},
				}, nil)
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.GatherReleaseNotesReturns(nil, err)
			},
			shouldErr: true,
		},
		{ // NewDocumentReturns failed
			prepare: func(mock *changelogfakes.FakeImpl, _ *changelog.Options) {
				mock.TagStringToSemverReturns(semver.Version{
					Pre: []semver.PRVersion{
						{VersionStr: "alpha"},
						{VersionNum: 1},
					},
				}, nil)
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.GatherReleaseNotesReturns(&notes.ReleaseNotes{}, nil)
				mock.NewDocumentReturns(nil, err)
			},
			shouldErr: true,
		},
		{ // CloneCVEData returns error
			prepare: func(mock *changelogfakes.FakeImpl, o *changelog.Options) {
				o.CloneCVEMaps = true
				mock.TagStringToSemverReturns(semver.Version{
					Major: 1,
					Minor: 19,
					Patch: 3,
				}, nil)
				mock.CloneCVEDataReturns("", err)
			},
			shouldErr: true,
		},
		{ // CloneCVEData returns empty string
			prepare: func(mock *changelogfakes.FakeImpl, o *changelog.Options) {
				o.CloneCVEMaps = true
				mock.TagStringToSemverReturns(semver.Version{
					Major: 1,
					Minor: 19,
					Patch: 3,
				}, nil)
				mock.ReadFileReturns([]byte(changelog.TocEnd), nil)
				mock.GatherReleaseNotesReturns(&notes.ReleaseNotes{}, nil)
				mock.CloneCVEDataReturns("", nil)
			},
			shouldErr: false,
		},
	} {
		options := &changelog.Options{}
		sut := changelog.New(options)
		mock := &changelogfakes.FakeImpl{}
		tc.prepare(mock, options)
		sut.SetImpl(mock)

		err := sut.Run()
		if tc.shouldErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

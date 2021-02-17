/*
Copyright 2021 The Kubernetes Authors.

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

package changelog

import (
	"k8s.io/release/pkg/changelog/changelogfakes"
)

func NewFakeImpl() *changelogfakes.FakeImpl {
	fallback := &defaultImpl{}
	return &changelogfakes.FakeImpl{
		AbsStub:                       fallback.Abs,
		AddStub:                       fallback.Add,
		CheckoutStub:                  fallback.Checkout,
		CommitStub:                    fallback.Commit,
		CreateDownloadsTableStub:      fallback.CreateDownloadsTable,
		CurrentBranchStub:             fallback.CurrentBranch,
		DependencyChangesStub:         fallback.DependencyChanges,
		GatherReleaseNotesStub:        fallback.GatherReleaseNotes,
		GenerateTOCStub:               fallback.GenerateTOC,
		GetURLResponseStub:            fallback.GetURLResponse,
		LatestGitHubTagsPerBranchStub: fallback.LatestGitHubTagsPerBranch,
		MarkdownToHTMLStub:            fallback.MarkdownToHTML,
		NewDocumentStub:               fallback.NewDocument,
		OpenRepoStub:                  fallback.OpenRepo,
		ParseHTMLTemplateStub:         fallback.ParseHTMLTemplate,
		ReadFileStub:                  fallback.ReadFile,
		RenderMarkdownTemplateStub:    fallback.RenderMarkdownTemplate,
		RepoDirStub:                   fallback.RepoDir,
		RevParseStub:                  fallback.RevParse,
		RmStub:                        fallback.Rm,
		StatStub:                      fallback.Stat,
		TagStringToSemverStub:         fallback.TagStringToSemver,
		TemplateExecuteStub:           fallback.TemplateExecute,
		ValidateAndFinishStub:         fallback.ValidateAndFinish,
		WriteFileStub:                 fallback.WriteFile,
	}
}

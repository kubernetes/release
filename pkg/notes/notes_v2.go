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

package notes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"k8s.io/release/pkg/notes/options"

	"github.com/cheggaaa/pb/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/nozzle/throttler"
	"github.com/sirupsen/logrus"
)

type commitPrPair struct {
	Commit *gitobject.Commit
	PrNum  int
}

type releaseNotesAggregator struct {
	releaseNotes *ReleaseNotes
	sync.RWMutex
}

func (g *Gatherer) ListReleaseNotesV2() (*ReleaseNotes, error) {
	// left parent of Git commits is always the main branch parent
	pairs, err := g.listLeftParentCommits(g.options)
	if err != nil {
		return nil, errors.Wrap(err, "listing offline commits")
	}

	// load map providers specified in options
	mapProviders := []MapProvider{}
	for _, initString := range g.options.MapProviderStrings {
		provider, err := NewProviderFromInitString(initString)
		if err != nil {
			return nil, errors.Wrap(err, "while getting release notes map providers")
		}
		mapProviders = append(mapProviders, provider)
	}

	t := throttler.New(maxParallelRequests, len(pairs))

	aggregator := releaseNotesAggregator{
		releaseNotes: NewReleaseNotes(),
	}

	pairsCount := len(pairs)
	logrus.Infof("processing release notes for %d commits", pairsCount)
	bar := pb.Full.Start(pairsCount)

	for _, pair := range pairs {
		noteMaps := []*ReleaseNotesMap{}

		for _, provider := range mapProviders {
			noteMaps, err = provider.GetMapsForPR(pair.PrNum)
			if err != nil {
				logrus.Errorf("[ignored] pr: %d err: %v", pair.PrNum, err)
				noteMaps = []*ReleaseNotesMap{}
			}
		}

		go func() {
			releaseNote, err := g.buildReleaseNote(pair)
			if err == nil && releaseNote != nil {
				for _, noteMap := range noteMaps {
					if err := releaseNote.ApplyMap(noteMap); err != nil {
						logrus.Errorf("[ignored] pr: %d err: %v", pair.PrNum, err)
					}
				}
				aggregator.Lock()
				aggregator.releaseNotes.Set(pair.PrNum, releaseNote)
				aggregator.Unlock()
			} else if err != nil {
				logrus.Errorf("sha: %s pr: %d err: %v", pair.Commit.Hash.String(), pair.PrNum, err)
			}
			bar.Increment()
			t.Done(nil)
		}()

		if t.Throttle() > 0 {
			break
		}
	}

	if err := t.Err(); err != nil {
		return nil, err
	}

	bar.Finish()

	return aggregator.releaseNotes, nil
}

func (g *Gatherer) buildReleaseNote(pair *commitPrPair) (*ReleaseNote, error) {
	pr, _, err := g.client.GetPullRequest(g.context, g.options.GithubOrg, g.options.GithubRepo, pair.PrNum)
	if err != nil {
		return nil, err
	}

	prBody := pr.GetBody()

	text, err := noteTextFromString(prBody)
	if err != nil {
		logrus.Debugf("sha: %s pr: %d err: %v", pair.Commit.Hash.String(), pair.PrNum, err)
		return nil, nil
	}

	documentation := DocumentationFromString(prBody)

	author := pr.GetUser().GetLogin()
	authorURL := pr.GetUser().GetHTMLURL()
	prURL := pr.GetHTMLURL()
	isFeature := hasString(labelsWithPrefix(pr, "kind"), "feature")
	noteSuffix := prettifySIGList(labelsWithPrefix(pr, "sig"))

	isDuplicateSIG := false
	if len(labelsWithPrefix(pr, "sig")) > 1 {
		isDuplicateSIG = true
	}

	isDuplicateKind := false
	if len(labelsWithPrefix(pr, "kind")) > 1 {
		isDuplicateKind = true
	}

	// TODO(wilsonehusin): extract / follow original in ReleasenoteFromCommit
	indented := strings.ReplaceAll(text, "\n", "\n  ")
	markdown := fmt.Sprintf("%s ([#%d](%s), [@%s](%s))",
		indented, pr.GetNumber(), prURL, author, authorURL)

	if noteSuffix != "" {
		markdown = fmt.Sprintf("%s [%s]", markdown, noteSuffix)
	}

	// Uppercase the first character of the markdown to make it look uniform
	markdown = strings.ToUpper(string(markdown[0])) + markdown[1:]

	return &ReleaseNote{
		Commit:         pair.Commit.Hash.String(),
		Text:           text,
		Markdown:       markdown,
		Documentation:  documentation,
		Author:         author,
		AuthorURL:      authorURL,
		PrURL:          prURL,
		PrNumber:       pr.GetNumber(),
		SIGs:           labelsWithPrefix(pr, "sig"),
		Kinds:          labelsWithPrefix(pr, "kind"),
		Areas:          labelsWithPrefix(pr, "area"),
		Feature:        isFeature,
		Duplicate:      isDuplicateSIG,
		DuplicateKind:  isDuplicateKind,
		ActionRequired: labelExactMatch(pr, "release-note-action-required"),
		DoNotPublish:   labelExactMatch(pr, "release-note-none"),
	}, nil

}

func (g *Gatherer) listLeftParentCommits(opts *options.Options) ([]*commitPrPair, error) {
	localRepository, err := git.PlainOpen(opts.RepoPath)
	if err != nil {
		return nil, err
	}

	// opts.StartSHA points to a tag (e.g. 1.20.0) which is on a release branch (e.g. release-1.20)
	// this means traveling through commit history from opts.EndSHA will never reach opts.StartSHA

	// the stopping point to be set should be the last shared commit between release branch and primary (master) branch
	// usually, following the left / first parents, it would be
	// * tag: v1.20.0, some merge commit
	// |
	// * Anago GCB release commit (begin branch out of release-1.20)
	// |
	// * last shared commit

	// this means the stopping point is 2 commits behind the tag pointed by opts.StartSHA

	stopHash := plumbing.NewHash(opts.StartSHA)
	for i := 0; i < 2; i++ {
		commitObject, err := localRepository.CommitObject(stopHash)
		if err != nil {
			return nil, errors.Wrap(err, "finding last shared commit")
		}
		stopHash = commitObject.ParentHashes[0]
	}

	logrus.Infof("will stop at %s", stopHash)

	currentTagHash := plumbing.NewHash(opts.EndSHA)

	pairs := []*commitPrPair{}
	hashPointer := currentTagHash
	for hashPointer != stopHash {
		hashString := hashPointer.String()

		// Find and collect commit objects
		commitPointer, err := localRepository.CommitObject(hashPointer)
		if err != nil {
			return nil, errors.Wrap(err, "finding CommitObject")
		}

		// Find and collect PR number from commit message
		prNums, err := prsNumForCommitFromMessage(commitPointer.Message)
		if err == errNoPRIDFoundInCommitMessage {
			logrus.Debugf("sha: %s prs: []", hashString)

			// Advance pointer based on left parent
			hashPointer = commitPointer.ParentHashes[0]
			continue
		}
		if err != nil {
			logrus.Warnf("sha: %s err: %s (silenced)", hashString, err.Error())

			// Advance pointer based on left parent
			hashPointer = commitPointer.ParentHashes[0]
			continue
		}
		logrus.Debugf("sha: %s prs: %v", hashString, prNums)

		// Only taking the first one, assuming they are merged by Prow
		pairs = append(pairs, &commitPrPair{Commit: commitPointer, PrNum: prNums[0]})

		// Advance pointer based on left parent
		hashPointer = commitPointer.ParentHashes[0]
	}

	return pairs, nil
}

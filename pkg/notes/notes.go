/*
Copyright 2017 The Kubernetes Authors.

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
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // used for file integrity checks, NOT security
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/cheggaaa/pb/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
	gogithub "github.com/google/go-github/v84/github"
	"github.com/mattn/go-isatty"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"sigs.k8s.io/release-sdk/github"
	"sigs.k8s.io/yaml"

	"k8s.io/release/pkg/notes/options"
)

var (
	errNoPRIDFoundInCommitMessage       = errors.New("no PR IDs found in the commit message")
	errNoOriginPRIDFoundInPR            = errors.New("no origin PR IDs found in the PR")
	apiSleepTime                  int64 = 60
)

var regexK8sCherryPickBotBranch = regexp.MustCompile(`cherry-pick-(?P<number>\d+)-to`)

const (
	DefaultOrg  = "kubernetes"
	DefaultRepo = "kubernetes"

	// maxParallelRequests is the maximum parallel requests we shall make to the
	// GitHub API.
	maxParallelRequests = 10
)

const k8sCherryPickBotUsername = "k8s-infra-cherrypick-robot"

type (
	Notes []string
	Kind  string
)

const (
	KindAPIChange     Kind = "api-change"
	KindBug           Kind = "bug"
	KindCleanup       Kind = "cleanup"
	KindDeprecation   Kind = "deprecation"
	KindDesign        Kind = "design"
	KindDocumentation Kind = "documentation"
	KindFailingTest   Kind = "failing-test"
	KindFeature       Kind = "feature"
	KindFlake         Kind = "flake"
	KindRegression    Kind = "regression"
	KindOther         Kind = "Other (Cleanup or Flake)"
	KindUncategorized Kind = "Uncategorized"
)

// ReleaseNote is the type that represents the total sum of all the information
// we've gathered about a single release note.
type ReleaseNote struct {
	// Commit is the SHA of the commit which is the source of this note. This is
	// also effectively a unique ID for release notes.
	Commit string `json:"commit"`

	// Text is the actual content of the release note
	Text string `json:"text"`

	// Markdown is the markdown formatted note
	Markdown string `json:"markdown"`

	// Docs is additional documentation for the release note
	Documentation []*Documentation `json:"documentation,omitempty"`

	// Author is the GitHub username of the commit author
	Author string `json:"author"`

	// AuthorURL is the GitHub URL of the commit author
	AuthorURL string `json:"author_url"`

	// PrURL is a URL to the PR
	PrURL string `json:"pr_url"`

	// PrNumber is the number of the PR
	PrNumber int `json:"pr_number"`

	// Areas is a list of the labels beginning with area/
	Areas []string `json:"areas,omitempty"`

	// Kinds is a list of the labels beginning with kind/
	Kinds []string `json:"kinds,omitempty"`

	// SIGs is a list of the labels beginning with sig/
	SIGs []string `json:"sigs,omitempty"`

	// Indicates whether or not a note will appear as a new feature
	Feature bool `json:"feature,omitempty"`

	// Indicates whether or not a note is duplicated across SIGs
	Duplicate bool `json:"duplicate,omitempty"`

	// Indicates whether or not a note is duplicated across Kinds
	DuplicateKind bool `json:"duplicate_kind,omitempty"`

	// ActionRequired indicates whether or not the release-note-action-required
	// label was set on the PR
	ActionRequired bool `json:"action_required,omitempty"`

	// DoNotPublish by default represents release-note-none label on GitHub
	DoNotPublish bool `json:"do_not_publish,omitempty"`

	// DataFields a key indexed map of data fields
	DataFields map[string]ReleaseNotesDataField `json:"-"`

	// IsMapped is set if the note got modified from a map
	IsMapped bool `json:"is_mapped,omitempty"`

	// PRBody is the full PR body of the release note
	PRBody string `json:"pr_body,omitempty"`
}

type Documentation struct {
	// A description about the documentation
	Description string `json:"description,omitempty"`

	// The url to be linked
	URL string `json:"url"`

	// Classifies the link as something special, like a KEP
	Type DocType `json:"type"`
}

type DocType string

const (
	DocTypeExternal DocType = "external"
	DocTypeKEP      DocType = "KEP"
	DocTypeOfficial DocType = "official"
)

// ReleaseNotes is the main struct for collecting release notes.
type ReleaseNotes struct {
	byPR    ReleaseNotesByPR
	history ReleaseNotesHistory
}

// NewReleaseNotes can be used to create a new empty ReleaseNotes struct.
func NewReleaseNotes() *ReleaseNotes {
	return &ReleaseNotes{
		byPR: make(ReleaseNotesByPR),
	}
}

// ReleaseNotes is a map of PR numbers referencing notes.
// To avoid needless loops, we need to be able to reference things by PR
// When we have to merge old and new entries, we want to be able to override
// the old entries with the new ones efficiently.
type ReleaseNotesByPR map[int]*ReleaseNote

// ReleaseNotesHistory is the sorted list of PRs in the commit history.
type ReleaseNotesHistory []int

// History returns the ReleaseNotesHistory for the ReleaseNotes.
func (r *ReleaseNotes) History() ReleaseNotesHistory {
	return r.history
}

// ByPR returns the ReleaseNotesByPR for the ReleaseNotes.
func (r *ReleaseNotes) ByPR() ReleaseNotesByPR {
	return r.byPR
}

// Get returns the ReleaseNote for the provided prNumber.
func (r *ReleaseNotes) Get(prNumber int) *ReleaseNote {
	return r.byPR[prNumber]
}

// Set can be used to set a release note for the provided prNumber.
func (r *ReleaseNotes) Set(prNumber int, note *ReleaseNote) {
	r.byPR[prNumber] = note
	r.history = append(r.history, prNumber)
}

type commitPrPair struct {
	Commit *gitobject.Commit
	PrNum  int
}

type releaseNotesAggregator struct {
	sync.RWMutex

	releaseNotes *ReleaseNotes
}

type Gatherer struct {
	client       github.Client
	context      context.Context //nolint:containedctx // contained context is intentional
	options      *options.Options
	MapProviders []*MapProvider
}

// NewGatherer creates a new notes gatherer.
func NewGatherer(ctx context.Context, opts *options.Options) (*Gatherer, error) {
	client, err := opts.Client()
	if err != nil {
		return nil, fmt.Errorf("unable to create notes client: %w", err)
	}

	return &Gatherer{
		client:  client,
		context: ctx,
		options: opts,
	}, nil
}

// NewGathererWithClient creates a new notes gatherer with a specific client.
func NewGathererWithClient(ctx context.Context, c github.Client) *Gatherer {
	return &Gatherer{
		client:  c,
		context: ctx,
		options: options.New(),
	}
}

// GatherReleaseNotes creates a new gatherer and collects the release notes
// afterwards.
func GatherReleaseNotes(ctx context.Context, opts *options.Options) (*ReleaseNotes, error) {
	logrus.Info("Gathering release notes")

	gatherer, err := NewGatherer(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("retrieving notes gatherer: %w", err)
	}

	startTime := time.Now()

	releaseNotes, err := gatherer.ListReleaseNotes()
	if err != nil {
		return nil, fmt.Errorf("listing release notes: %w", err)
	}

	logrus.Infof("Finished gathering release notes in %v", time.Since(startTime))

	return releaseNotes, nil
}

// ListReleaseNotes produces a list of fully contextualized release notes
// starting from a given commit SHA and ending at starting a given commit SHA.
func (g *Gatherer) ListReleaseNotes() (*ReleaseNotes, error) {
	// left parent of Git commits is always the main branch parent
	pairs, err := g.listLeftParentCommits(g.options)
	if err != nil {
		return nil, fmt.Errorf("listing offline commits: %w", err)
	}

	// load map providers specified in options
	mapProviders := []MapProvider{}

	for _, initString := range g.options.MapProviderStrings {
		provider, err := NewProviderFromInitString(initString)
		if err != nil {
			return nil, fmt.Errorf("while getting release notes map providers: %w", err)
		}

		mapProviders = append(mapProviders, provider)
	}

	eg := new(errgroup.Group)
	eg.SetLimit(maxParallelRequests)

	aggregator := releaseNotesAggregator{
		releaseNotes: NewReleaseNotes(),
	}

	dedupeCache := struct {
		sync.Mutex

		seen map[string]struct{}
	}{seen: map[string]struct{}{}}

	pairsCount := len(pairs)
	logrus.Infof("processing release notes for %d commits", pairsCount)

	// display progress bar in stdout, since stderr is used by logger
	bar := pb.New(pairsCount).SetWriter(os.Stdout)

	// only display progress bar in user TTY
	if isatty.IsTerminal(os.Stdout.Fd()) {
		bar.Start()
	}

	for _, pair := range pairs {
		eg.Go(func() error {
			var noteMaps []*ReleaseNotesMap

			for _, provider := range mapProviders {
				providerMaps, err := provider.GetMapsForPR(pair.PrNum)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"pr": pair.PrNum,
					}).Errorf("ignore err: %v", err)

					continue
				}

				noteMaps = append(noteMaps, providerMaps...)
			}

			releaseNote, err := g.buildReleaseNote(pair)
			if err == nil {
				if releaseNote == nil {
					logrus.WithFields(logrus.Fields{
						"pr": pair.PrNum,
					}).Debugf("skip: empty release note")
				} else {
					for _, noteMap := range noteMaps {
						if err := releaseNote.ApplyMap(noteMap, g.options.AddMarkdownLinks); err != nil {
							logrus.WithFields(logrus.Fields{
								"pr": pair.PrNum,
							}).Errorf("ignore err: %v", err)
						}
					}

					dedupeCache.Lock()

					_, duplicate := dedupeCache.seen[releaseNote.Markdown]
					if !duplicate {
						dedupeCache.seen[releaseNote.Markdown] = struct{}{}
					}
					dedupeCache.Unlock()

					if duplicate {
						logrus.WithFields(logrus.Fields{
							"pr": pair.PrNum,
						}).Debugf("skip: duplicate release note")
					} else {
						logrus.WithFields(logrus.Fields{
							"pr":   pair.PrNum,
							"note": releaseNote.Text,
						}).Debugf("finalized release note")
						aggregator.Lock()
						aggregator.releaseNotes.Set(pair.PrNum, releaseNote)
						aggregator.Unlock()
					}
				}
			} else {
				logrus.WithFields(logrus.Fields{
					"sha": pair.Commit.Hash.String(),
					"pr":  pair.PrNum,
				}).Errorf("err: %v", err)
			}

			bar.Increment()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	bar.Finish()

	return aggregator.releaseNotes, nil
}

// noteTextFromString returns the text of the release note given a string which
// may contain the commit message, the PR description, etc.
// This is generally the content inside the ```release-note ``` stanza.
func noteTextFromString(s string) (string, error) {
	// check release note is not empty
	// Matches "release-notes" block with no meaningful content (ex. only whitespace, empty, just newlines)
	emptyExps := []*regexp.Regexp{
		regexp.MustCompile("(?i)```release-notes?\\s*```\\s*"),
	}

	if matchesFilter(s, emptyExps) {
		return "", errors.New("empty release note")
	}

	exps := []*regexp.Regexp{
		// (?s) is needed for '.' to be matching on newlines, by default that's disabled
		// we need to match ungreedy 'U', because after the notes a `docs` block can occur
		regexp.MustCompile("(?sU)```release-notes?\\r\\n(?P<note>.+)\\r\\n```"),
		regexp.MustCompile("(?sU)```dev-release-notes?\\r\\n(?P<note>.+)"),
		regexp.MustCompile("(?sU)```\\r\\n(?P<note>.+)\\r\\n```"),
		regexp.MustCompile("(?sU)```release-notes?\n(?P<note>.+)\n```"),
	}

	for _, exp := range exps {
		match := exp.FindStringSubmatch(s)
		if len(match) == 0 {
			continue
		}

		result := map[string]string{}

		for i, name := range exp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		note := strings.ReplaceAll(result["note"], "\r", "")
		note = stripActionRequired(note)
		note = dashify(note)
		note = unlist(note)
		note = strings.TrimSpace(note)

		return note, nil
	}

	return "", errors.New("no matches found when parsing note text from commit string")
}

func DocumentationFromString(s string) []*Documentation {
	regex := regexp.MustCompile("(?s)```docs\\r?\\n(?P<text>.+)\\r?\\n```")
	match := regex.FindStringSubmatch(s)

	if len(match) < 1 {
		// Nothing found, but we don't require it
		return nil
	}

	result := []*Documentation{}
	text := match[1]
	text = stripStar(text)
	text = stripDash(text)

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		const httpPrefix = "http"

		s := strings.SplitN(scanner.Text(), httpPrefix, 2)
		if len(s) != 2 {
			continue
		}

		description := strings.TrimRight(strings.TrimSpace(s[0]), " :-")
		urlString := httpPrefix + strings.TrimSpace(s[1])

		// Validate the URL
		parsedURL, err := url.Parse(urlString)
		if err != nil {
			continue
		}

		result = append(result, &Documentation{
			Description: description,
			URL:         urlString,
			Type:        classifyURL(parsedURL),
		})
	}

	return result
}

// classifyURL returns the correct DocType for the given url.
func classifyURL(u *url.URL) DocType {
	// Kubernetes Enhancement Proposals (KEPs)
	if strings.Contains(u.Host, "github.com") &&
		strings.Contains(u.Path, "/kubernetes/enhancements/") {
		return DocTypeKEP
	}

	// Official documentation
	if strings.Contains(u.Host, "kubernetes.io") &&
		strings.Contains(u.Path, "/docs/") {
		return DocTypeOfficial
	}

	return DocTypeExternal
}

// ReleaseNoteForPullRequest returns a release note from a pull request number.
// If the release note is blank or.
func (g *Gatherer) ReleaseNoteForPullRequest(prNr int) (*ReleaseNote, error) {
	pr, _, err := g.client.GetPullRequest(g.context, g.options.GithubOrg, g.options.GithubRepo, prNr)
	if err != nil {
		return nil, fmt.Errorf("reading PR from GitHub: %w", err)
	}

	prBody := pr.GetBody()

	// This will be true when the release note is NONE or the flag is set
	var doNotPublish bool

	// If we match exclusion filter (release-note-none), we don't look further,
	// instead we return that PR. We return that PR because it might have a map,
	// which is checked after this function returns
	if MatchesExcludeFilter(prBody) || len(labelsWithPrefix(pr, "release-note-none")) > 0 {
		doNotPublish = true
	}

	// If we didn't match the exclusion filter, try to extract the release note from the PR.
	// If we can't extract the release note, consider that the PR is invalid and take the next one
	s, err := noteTextFromString(prBody)
	if err != nil && !doNotPublish {
		return nil, fmt.Errorf("PR #%d does not seem to contain a valid release note: %w", pr.GetNumber(), err)
	}

	// If we found a valid release note, return the PR, otherwise, take the next one
	if s == "" && !doNotPublish {
		return nil, fmt.Errorf("PR #%d does not seem to contain a valid release note", pr.GetNumber())
	}

	if doNotPublish {
		s = ""
	}

	// Create the release notes object
	note := &ReleaseNote{
		Text:           s,
		Markdown:       s,
		Documentation:  []*Documentation{},
		Author:         *pr.GetUser().Login,
		AuthorURL:      fmt.Sprintf("%s%s", g.options.GithubBaseURL, *pr.GetUser().Login),
		PrURL:          fmt.Sprintf("%s%s/%s/pull/%d", g.options.GithubBaseURL, g.options.GithubOrg, g.options.GithubRepo, prNr),
		PrNumber:       prNr,
		SIGs:           labelsWithPrefix(pr, "sig"),
		Kinds:          labelsWithPrefix(pr, "kind"),
		Areas:          labelsWithPrefix(pr, "area"),
		Feature:        false,
		Duplicate:      false,
		DuplicateKind:  false,
		ActionRequired: false,
		DoNotPublish:   doNotPublish,
		DataFields:     map[string]ReleaseNotesDataField{},
		PRBody:         prBody,
	}

	if s != "" {
		logrus.Infof("PR #%d seems to contain a release note", pr.GetNumber())
	}

	return note, nil
}

func (g *Gatherer) buildReleaseNote(pair *commitPrPair) (*ReleaseNote, error) {
	pr, _, err := g.client.GetPullRequest(g.context, g.options.GithubOrg, g.options.GithubRepo, pair.PrNum)
	if err != nil {
		return nil, err
	}

	prBody := pr.GetBody()

	if MatchesExcludeFilter(prBody) {
		return nil, nil //nolint:nilnil // intentional nil,nil return
	}

	if len(g.options.IncludeLabels) > 0 && !matchesLabelFilter(pr.Labels, g.options.IncludeLabels) {
		return nil, nil //nolint:nilnil // intentional nil,nil return
	}

	text, err := noteTextFromString(prBody)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"sha": pair.Commit.Hash.String(),
			"pr":  pair.PrNum,
		}).Debugf("ignore err: %v", err)

		return nil, nil //nolint:nilnil // intentional nil,nil return
	}

	if isAutomatedCherryPickPR(pr) {
		logrus.Infof("PR #%d seems to be an automated cherry-pick, retrieving origin info", pr.GetNumber())

		originPRNum, err := originPrNumFromPr(pr)
		if err != nil {
			return nil, err
		}

		originPR, err := g.getPr(originPRNum)
		if err != nil {
			return nil, err
		}

		pr.User = originPR.GetUser()
	}

	documentation := DocumentationFromString(prBody)

	author := pr.GetUser().GetLogin()
	authorURL := pr.GetUser().GetHTMLURL()
	prURL := pr.GetHTMLURL()
	isFeature := slices.Contains(labelsWithPrefix(pr, "kind"), "feature")
	sigLabels := labelsWithPrefix(pr, "sig")
	noteSuffix := prettifySIGList(sigLabels)

	isDuplicateSIG := len(labelsWithPrefix(pr, "sig")) > 1

	isDuplicateKind := len(labelsWithPrefix(pr, "kind")) > 1

	indented := strings.ReplaceAll(text, "\n", "\n  ")
	markdown := fmt.Sprintf("%s (#%d, @%s)",
		indented, pr.GetNumber(), author)

	if g.options.AddMarkdownLinks {
		markdown = fmt.Sprintf("%s ([#%d](%s), [@%s](%s))",
			indented, pr.GetNumber(), prURL, author, authorURL)
	}

	if noteSuffix != "" {
		markdown = fmt.Sprintf("%s [%s]", markdown, noteSuffix)
	}

	// Uppercase the first character of the markdown to make it look uniform
	markdown = capitalizeString(markdown)

	return &ReleaseNote{
		Commit:         pair.Commit.Hash.String(),
		Text:           text,
		Markdown:       markdown,
		Documentation:  documentation,
		Author:         author,
		AuthorURL:      authorURL,
		PrURL:          prURL,
		PrNumber:       pr.GetNumber(),
		SIGs:           sigLabels,
		Kinds:          labelsWithPrefix(pr, "kind"),
		Areas:          labelsWithPrefix(pr, "area"),
		Feature:        isFeature,
		Duplicate:      isDuplicateSIG,
		DuplicateKind:  isDuplicateKind,
		ActionRequired: labelExactMatch(pr, "release-note-action-required"),
		DoNotPublish:   labelExactMatch(pr, "release-note-none"),
		PRBody:         prBody,
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

	// ^ master
	// |
	// * tag: 1.21.0-alpha.x / 1.21.0-beta.y
	// |
	// : :
	// | |
	// | * tag: v1.20.0, some merge commit pointed by opts.StartSHA
	// | |
	// | * Anago GCB release commit (begin branch out of release-1.20)
	// |/
	// x last shared commit

	// merge base would resolve to last shared commit, marked by (x)

	endCommit, err := localRepository.CommitObject(plumbing.NewHash(opts.EndSHA))
	if err != nil {
		return nil, fmt.Errorf("finding commit of EndSHA: %w", err)
	}

	// When SkipFirstCommit is used, the StartSHA is advanced to the next
	// commit which may not be on the first-parent chain. Use the original
	// tag SHA for the merge base calculation so the traversal can find a
	// valid stop point via first-parent walking.
	mergeBaseSHA := opts.StartSHA
	if opts.OriginalStartSHA != "" {
		mergeBaseSHA = opts.OriginalStartSHA
	}

	startCommit, err := localRepository.CommitObject(plumbing.NewHash(mergeBaseSHA))
	if err != nil {
		return nil, fmt.Errorf("finding commit of merge base SHA: %w", err)
	}

	logrus.Debugf("finding merge base (last shared commit) between the two SHAs")

	startTime := time.Now()

	lastSharedCommits, err := endCommit.MergeBase(startCommit)
	if err != nil {
		return nil, fmt.Errorf("finding shared commits: %w", err)
	}

	if len(lastSharedCommits) == 0 {
		return nil, errors.New("no shared commits between the provided SHAs")
	}

	logrus.Debugf("found merge base in %v", time.Since(startTime))

	stopHash := lastSharedCommits[0].Hash
	logrus.Infof("will stop at %s", stopHash.String())

	currentTagHash := plumbing.NewHash(opts.EndSHA)

	pairs := []*commitPrPair{}

	hashPointer := currentTagHash
	for hashPointer != stopHash {
		hashString := hashPointer.String()

		// Find and collect commit objects
		commitPointer, err := localRepository.CommitObject(hashPointer)
		if err != nil {
			return nil, fmt.Errorf("finding CommitObject: %w", err)
		}

		if len(commitPointer.ParentHashes) == 0 {
			return nil, fmt.Errorf("commit %s has no parents, cannot traverse further", hashString)
		}

		// Find and collect PR number from commit message
		prNums, err := prsNumForCommitFromMessage(commitPointer.Message)
		if errors.Is(err, errNoPRIDFoundInCommitMessage) {
			logrus.WithFields(logrus.Fields{
				"sha": hashString,
			}).Debug("no associated PR found")

			hashPointer = commitPointer.ParentHashes[0]

			continue
		}

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"sha": hashString,
			}).Warnf("ignore err: %v", err)

			hashPointer = commitPointer.ParentHashes[0]

			continue
		}

		logrus.WithFields(logrus.Fields{
			"sha": hashString,
			"prs": prNums,
		}).Debug("found PR from commit")

		// Only taking the first one, assuming they are merged by Prow
		pairs = append(pairs, &commitPrPair{Commit: commitPointer, PrNum: prNums[0]})

		hashPointer = commitPointer.ParentHashes[0]
	}

	return pairs, nil
}

// noteExclusionFilters is a list of regular expressions that match commits
// that do NOT contain release notes. Notably, this is all of the variations of
// "release note none" that appear in the commit log.
var noteExclusionFilters = []*regexp.Regexp{
	// 'none','n/a','na' case-insensitive with optional trailing
	// whitespace, wrapped in ``` with/without release-note identifier
	// the 'none','n/a','na' can also optionally be wrapped in quotes ' or "
	regexp.MustCompile("(?i)```release-notes?\\s*('\")?(none|n/a|na)('\")?\\s*```"),

	// simple '/release-note-none' tag
	regexp.MustCompile("/release-note-none"),
}

// MatchesExcludeFilter returns true if the string matches an excluded release note.
func MatchesExcludeFilter(msg string) bool {
	return matchesFilter(msg, noteExclusionFilters)
}

// matchesLabelFilter returns true if any of PR labels match the includeLabels.
func matchesLabelFilter(prLabels []*gogithub.Label, includeLabels []string) bool {
	for _, include := range includeLabels {
		for _, label := range prLabels {
			if label.GetName() == include {
				return true
			}
		}
	}

	return false
}

func matchesFilter(msg string, filters []*regexp.Regexp) bool {
	for _, filter := range filters {
		if filter.MatchString(msg) {
			return true
		}
	}

	return false
}

func isAutomatedCherryPickPR(pr *gogithub.PullRequest) bool {
	if pr == nil || pr.GetUser() == nil {
		return false
	}

	return pr.GetUser().GetLogin() == k8sCherryPickBotUsername
}

func originPrNumFromPr(pr *gogithub.PullRequest) (int, error) {
	if pr == nil || pr.GetHead() == nil {
		return 0, errNoOriginPRIDFoundInPR
	}

	originPR := prForRegex(regexK8sCherryPickBotBranch, pr.GetHead().GetLabel())
	if originPR == 0 {
		return 0, errNoOriginPRIDFoundInPR
	}

	return originPR, nil
}

// labelsWithPrefix is a helper for fetching all labels on a PR that start with
// a given string. This pattern is used often in the k/k repo and we can take
// advantage of this to contextualize release note generation with the kind, sig,
// area, etc labels.
func labelsWithPrefix(pr *gogithub.PullRequest, prefix string) []string {
	var labels []string

	for _, label := range pr.Labels {
		if strings.HasPrefix(*label.Name, prefix) {
			labels = append(labels, strings.TrimPrefix(*label.Name, prefix+"/"))
		}
	}

	return labels
}

// labelExactMatch indicates whether or not a matching label was found on PR.
func labelExactMatch(pr *gogithub.PullRequest, labelToFind string) bool {
	for _, label := range pr.Labels {
		if *label.Name == labelToFind {
			return true
		}
	}

	return false
}

func stripActionRequired(note string) string {
	expressions := []string{
		`(?i)\[action required\]\s`,
		`(?i)action required:\s`,
	}

	for _, exp := range expressions {
		re := regexp.MustCompile(exp)
		note = re.ReplaceAllString(note, "")
	}

	return note
}

func stripStar(note string) string {
	re := regexp.MustCompile(`(?i)\*\s`)

	return re.ReplaceAllString(note, "")
}

func stripDash(note string) string {
	re := regexp.MustCompile(`(?i)\-\s`)

	return re.ReplaceAllString(note, "")
}

const listPrefix = "- "

func dashify(note string) string {
	re := regexp.MustCompile(`(?m)(^\s*)\*\s`)

	return re.ReplaceAllString(note, "$1- ")
}

// unlist transforms a single markdown list entry to a flat note entry.
func unlist(note string) string {
	if !strings.HasPrefix(note, listPrefix) {
		return note
	}

	res := strings.Builder{}
	scanner := bufio.NewScanner(strings.NewReader(note))
	firstLine := true
	trim := true

	for scanner.Scan() {
		line := scanner.Text()

		// Per default strip the two dashes from the list
		prefix := "  "

		if strings.HasPrefix(line, listPrefix) {
			if firstLine {
				// First list item, strip the prefix
				prefix = listPrefix
				firstLine = false
			} else {
				// Another list item? Treat it as sublist and do not trim any
				// more.
				trim = false
			}
		}

		if trim {
			line = strings.TrimPrefix(line, prefix)
		}

		res.WriteString(line + "\n")
	}

	return res.String()
}

// canWaitAndRetry retruen true if the gatherer hit the GitHub API secondary rate limit.
func canWaitAndRetry(r *gogithub.Response, err error) bool {
	// If we hit the secondary rate limit...
	if r == nil {
		return false
	}

	if r.StatusCode == http.StatusForbidden &&
		strings.Contains(err.Error(), "secondary rate limit. Please wait") {
		// ... sleep for a minute plus a random bit so that workers don't
		// respawn all at the same time
		rtime, err := rand.Int(rand.Reader, big.NewInt(30))
		if err != nil {
			logrus.Error(err)

			return false
		}

		waitTime := rtime.Int64() + apiSleepTime
		logrus.Warnf("Hit the GitHub secondary rate limit, sleeping for %d secs.", waitTime)
		time.Sleep(time.Duration(waitTime) * time.Second)

		return true
	}

	return false
}

func (g *Gatherer) getPr(prNum int) (*gogithub.PullRequest, error) {
	for {
		res, resp, err := g.client.GetPullRequest(g.context, g.options.GithubOrg, g.options.GithubRepo, prNum)
		if err != nil {
			if !canWaitAndRetry(resp, err) {
				return nil, err
			}
		} else {
			return res, nil
		}
	}
}

func prsNumForCommitFromMessage(commitMessage string) (prs []int, err error) {
	// Thankfully k8s-merge-robot commits the PR number consistently. If this ever
	// stops being true, this definitely won't work anymore.
	regex := regexp.MustCompile(`Merge pull request #(?P<number>\d+)`)

	pr := prForRegex(regex, commitMessage)
	if pr != 0 {
		prs = append(prs, pr)
	}

	regex = regexp.MustCompile(`automated-cherry-pick-of-#(?P<number>\d+)`)

	pr = prForRegex(regex, commitMessage)
	if pr != 0 {
		prs = append(prs, pr)
	}

	regex = regexp.MustCompile(`\(#(?P<number>\d+)\)\s*\n\nThis reverts commit`)

	pr = prForRegex(regex, commitMessage)
	if pr != 0 {
		prs = append(prs, pr)
	}

	// If the PR was squash merged, the regexp is different
	regex = regexp.MustCompile(`\(#(?P<number>\d+)\)`)

	pr = prForRegex(regex, commitMessage)
	if pr != 0 {
		prs = append(prs, pr)
	}

	if prs == nil {
		return nil, errNoPRIDFoundInCommitMessage
	}

	return prs, nil
}

func prForRegex(regex *regexp.Regexp, commitMessage string) int {
	result := map[string]string{}
	match := regex.FindStringSubmatch(commitMessage)

	if match == nil {
		return 0
	}

	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	pr, err := strconv.Atoi(result["number"])
	if err != nil {
		return 0
	}

	return pr
}

// prettySIG takes a sig name as parsed by the `sig-foo` label and returns a
// "pretty" version of it that can be printed in documents.
func prettySIG(sig string) string {
	parts := strings.Split(sig, "-")
	for i, part := range parts {
		switch part {
		case "vsphere":
			parts[i] = "vSphere"
		case "vmware":
			parts[i] = "VMWare"
		case "openstack":
			parts[i] = "OpenStack"
		case "api", "aws", "cli", "gcp":
			parts[i] = strings.ToUpper(part)
		default:
			parts[i] = cases.Title(language.English).String(part)
		}
	}

	return strings.Join(parts, " ")
}

func prettifySIGList(sigs []string) string {
	sigList := ""

	// sort the list so that any group of SIGs with the same content gives us the
	// same result
	sort.Strings(sigs)

	for i, sig := range sigs {
		switch i {
		case 0:
			sigList = "SIG " + prettySIG(sig)

		case len(sigs) - 1:
			sigList = fmt.Sprintf("%s and %s", sigList, prettySIG(sig))

		default:
			sigList = fmt.Sprintf("%s, %s", sigList, prettySIG(sig))
		}
	}

	return sigList
}

// ApplyMap Modifies the content of the release using information from
//
//	a ReleaseNotesMap
func (rn *ReleaseNote) ApplyMap(noteMap *ReleaseNotesMap, markdownLinks bool) error {
	logrus.WithFields(logrus.Fields{
		"pr": rn.PrNumber,
	}).Debugf("Applying map to note")

	rn.IsMapped = true

	if noteMap.PRBody != nil {
		if rn.PRBody != "" && rn.PRBody != *noteMap.PRBody {
			logrus.Warnf("Original PR body of release note mapping changed for PR: #%d", rn.PrNumber)

			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(rn.PRBody, *noteMap.PRBody, false)
			logrus.Warnf("The diff between actual release note body and mapped one is:\n%s", dmp.DiffPrettyText(diffs))
		}

		rn.PRBody = *noteMap.PRBody
	}

	reRenderMarkdown := false

	if noteMap.ReleaseNote.Author != nil {
		rn.Author = *noteMap.ReleaseNote.Author
		rn.AuthorURL = "https://github.com/" + *noteMap.ReleaseNote.Author
		reRenderMarkdown = true
	}

	if noteMap.ReleaseNote.Text != nil {
		rn.Text = *noteMap.ReleaseNote.Text
		reRenderMarkdown = true
	}

	if noteMap.ReleaseNote.Documentation != nil {
		rn.Documentation = *noteMap.ReleaseNote.Documentation
	}

	if noteMap.ReleaseNote.Areas != nil {
		rn.Areas = *noteMap.ReleaseNote.Areas
	}

	if noteMap.ReleaseNote.Kinds != nil {
		rn.Kinds = *noteMap.ReleaseNote.Kinds
	}

	if noteMap.ReleaseNote.SIGs != nil {
		rn.SIGs = *noteMap.ReleaseNote.SIGs
		reRenderMarkdown = true
	}

	if noteMap.ReleaseNote.Feature != nil {
		rn.Feature = *noteMap.ReleaseNote.Feature
	}

	if noteMap.ReleaseNote.ActionRequired != nil {
		rn.ActionRequired = *noteMap.ReleaseNote.ActionRequired
	}

	if noteMap.ReleaseNote.DoNotPublish != nil {
		rn.DoNotPublish = *noteMap.ReleaseNote.DoNotPublish
	}

	// If there are datafields, add them
	if len(noteMap.DataFields) > 0 {
		rn.DataFields = make(map[string]ReleaseNotesDataField)
	}

	maps.Copy(rn.DataFields, noteMap.DataFields)

	// If parts of the markup where modified, change them
	// TODO: Spin this to sep function
	if reRenderMarkdown {
		indented := strings.ReplaceAll(rn.Text, "\n", "\n  ")
		markdown := fmt.Sprintf("%s (#%d, @%s)",
			indented, rn.PrNumber, rn.Author)

		if markdownLinks {
			markdown = fmt.Sprintf("%s ([#%d](%s), [@%s](%s))",
				indented, rn.PrNumber, rn.PrURL, rn.Author, rn.AuthorURL)
		}
		// Add sig labels to markdown
		if len(rn.SIGs) >= 1 {
			markdown = fmt.Sprintf("%s [%s]", markdown, prettifySIGList(rn.SIGs))
		}
		// Uppercase the first character of the markdown to make it look uniform
		rn.Markdown = capitalizeString(markdown)
	}

	return nil
}

// ToNoteMap returns the note's content as YAML code for use in a notemap.
func (rn *ReleaseNote) ToNoteMap() (string, error) {
	noteMap := &ReleaseNotesMap{
		PR:     rn.PrNumber,
		Commit: rn.Commit,
	}

	noteMap.ReleaseNote.Text = &rn.Text
	noteMap.ReleaseNote.Documentation = &rn.Documentation
	noteMap.ReleaseNote.Author = &rn.Author
	noteMap.ReleaseNote.Areas = &rn.Areas
	noteMap.ReleaseNote.Kinds = &rn.Kinds
	noteMap.ReleaseNote.SIGs = &rn.SIGs
	noteMap.ReleaseNote.Feature = &rn.Feature
	noteMap.ReleaseNote.ActionRequired = &rn.ActionRequired
	noteMap.ReleaseNote.DoNotPublish = &rn.DoNotPublish
	noteMap.PRBody = &rn.PRBody

	yamlCode, err := yaml.Marshal(&noteMap)
	if err != nil {
		return "", fmt.Errorf("marshalling release note to map: %w", err)
	}

	return string(yamlCode), nil
}

// ContentHash returns a sha1 hash derived from the note's content.
func (rn *ReleaseNote) ContentHash() (string, error) {
	// Convert the note to a map
	noteMap, err := rn.ToNoteMap()
	if err != nil {
		return "", fmt.Errorf("serializing note's content: %w", err)
	}

	//nolint:gosec // used for file integrity checks, NOT security
	// TODO(relnotes): Could we use SHA256 here instead?
	h := sha1.New()

	_, err = h.Write([]byte(noteMap))
	if err != nil {
		return "", fmt.Errorf("calculating content hash from map: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// capitalizeString returns a capitalized string of the input string.
func capitalizeString(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

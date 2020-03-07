/*
Copyright 2019 The Kubernetes Authors.

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

package document

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/release/pkg/notes"
	"k8s.io/release/pkg/release"
)

// Document represents the underlying structure of a release notes document.
type Document struct {
	NotesWithActionRequired Notes         `json:"action_required"`
	NotesByKind             NotesByKind   `json:"kinds"`
	Downloads               *FileMetadata `json:"downloads"`
	CurrentRevision         string        `json:"release_tag"`
	PreviousRevision        string
}

// FileMetadata contains metadata about files associated with the release.
type FileMetadata struct {
	// Files containing source code.
	Source []File

	// Client binaries
	Client []File

	// Server binaries
	Server []File

	// Node binaries
	Node []File
}

// fetchMetadata generates file metadata for k8s binaries in `dir`. Returns nil
// if `dir` is not given or when there are no matching well known k8s binaries
// in `dir`.
func fetchMetadata(dir, urlPrefix, tag string) (*FileMetadata, error) {
	if dir == "" {
		return nil, nil
	}
	if tag == "" {
		return nil, errors.New("release tags not specified")
	}
	if urlPrefix == "" {
		return nil, errors.New("url prefix not specified")
	}

	fm := new(FileMetadata)
	m := map[*[]File][]string{
		&fm.Source: {"kubernetes.tar.gz", "kubernetes-src.tar.gz"},
		&fm.Client: {"kubernetes-client*.tar.gz"},
		&fm.Server: {"kubernetes-server*.tar.gz"},
		&fm.Node:   {"kubernetes-node*.tar.gz"},
	}

	var fileCount int
	for fileType, patterns := range m {
		fileMetadata, err := fileInfo(dir, patterns, urlPrefix, tag)
		if err != nil {
			return nil, errors.Wrap(err, "fetching file metadata")
		}
		*fileType = append(*fileType, fileMetadata...)
		fileCount += len(fileMetadata)
	}

	if fileCount == 0 {
		return nil, nil
	}
	return fm, nil
}

func fileInfo(dir string, patterns []string, urlPrefix, tag string) ([]File, error) {
	var files []File
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, err
		}

		for _, filePath := range matches {
			f, err := os.Open(filePath)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			h := sha512.New()
			if _, err := io.Copy(h, f); err != nil {
				return nil, err
			}

			fileName := filepath.Base(filePath)
			files = append(files, File{
				Checksum: fmt.Sprintf("%x", h.Sum(nil)),
				Name:     fileName,
				URL:      fmt.Sprintf("%s/%s/%s", urlPrefix, tag, fileName),
			})
		}
	}
	return files, nil
}

// A File is a downloadable file.
type File struct {
	Checksum, Name, URL string
}

type Kind string
type NotesByKind map[Kind]Notes
type Notes []string

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
	// TODO: These should be same case as the others. Probably fix up prettyKind()??
	KindBugCleanupFlake Kind = "Other (Bug, Cleanup or Flake)"
	KindUncategorized   Kind = "Uncategorized"
)

var kindPriority = []Kind{
	KindDeprecation,
	KindAPIChange,
	KindFeature,
	KindDesign,
	KindDocumentation,
	KindFailingTest,
	KindBug,
	KindCleanup,
	KindFlake,
	KindBugCleanupFlake,
	KindUncategorized,
}

var kindMap = map[Kind]Kind{
	KindBug:     KindBugCleanupFlake,
	KindCleanup: KindBugCleanupFlake,
	KindFlake:   KindBugCleanupFlake,
}

// CreateDocument assembles an organized document from an unorganized set of
// release notes
func CreateDocument(releaseNotes notes.ReleaseNotes, history notes.ReleaseNotesHistory) (*Document, error) {
	doc := &Document{
		NotesWithActionRequired: Notes{},
		NotesByKind:             NotesByKind{},
	}

	for _, pr := range history {
		note := releaseNotes[pr]

		if note.DuplicateKind {
			kind := mapKind(highestPriorityKind(note.Kinds))
			existingNotes, ok := doc.NotesByKind[kind]
			if ok {
				doc.NotesByKind[kind] = append(existingNotes, note.Markdown)
			} else {
				doc.NotesByKind[kind] = []string{note.Markdown}
			}
		} else if note.ActionRequired {
			doc.NotesWithActionRequired = append(doc.NotesWithActionRequired, note.Markdown)
		} else {
			for _, kind := range note.Kinds {
				mappedKind := mapKind(Kind(kind))
				notesForKind, ok := doc.NotesByKind[mappedKind]
				if ok {
					doc.NotesByKind[mappedKind] = append(notesForKind, note.Markdown)
				} else {
					doc.NotesByKind[mappedKind] = []string{note.Markdown}
				}
			}

			if len(note.Kinds) == 0 {
				// the note has not been categorized so far
				kind := KindUncategorized
				if existingNotes, ok := doc.NotesByKind[kind]; ok {
					if ok {
						doc.NotesByKind[kind] = append(existingNotes, note.Markdown)
					} else {
						doc.NotesByKind[kind] = []string{note.Markdown}
					}
				}
			}
		}
	}

	sort.Strings(doc.NotesWithActionRequired)
	return doc, nil
}

// RenderMarkdownTemplate renders a document using the Go template in `goTemplate`.
func (d *Document) RenderMarkdownTemplate(bucket, fileDir, goTemplate string) (string, error) {
	urlPrefix := release.URLPrefixForBucket(bucket)
	fileMetadata, err := fetchMetadata(fileDir, urlPrefix, d.CurrentRevision)
	if err != nil {
		return "", errors.Wrap(err, "fetching downloads metadata")
	}
	d.Downloads = fileMetadata

	tmpl, err := template.New("markdown").
		Funcs(template.FuncMap{"prettyKind": prettyKind}).
		Parse(goTemplate)
	if err != nil {
		return "", errors.Wrap(err, "parsing template")
	}

	var s strings.Builder
	if err := tmpl.Execute(&s, d); err != nil {
		return "", errors.Wrapf(err, "rendering with template")
	}
	return s.String(), nil
}

// RenderMarkdown accepts a Document and writes a version of that document to
// supplied io.Writer in markdown format.
func (d *Document) RenderMarkdown(bucket, tars, prevTag, newTag string) (string, error) {
	o := &strings.Builder{}
	if err := CreateDownloadsTable(o, bucket, tars, prevTag, newTag); err != nil {
		return "", err
	}

	nl := func() {
		o.WriteRune('\n')
	}
	nlnl := func() {
		nl()
		nl()
	}

	// writeNote encapsulates the pre-processing that might happen on a note text
	// before it gets bulleted and written to the io.Writer
	writeNote := func(s string) {
		const prefix = "- "
		if !strings.HasPrefix(s, prefix) {
			o.WriteString(prefix)
		}
		o.WriteString(s)
		nl()
	}

	// notes with action required get their own section
	if len(d.NotesWithActionRequired) > 0 {
		o.WriteString("## Urgent Upgrade Notes")
		nlnl()
		o.WriteString("### (No, really, you MUST read this before you upgrade)")
		nlnl()
		for _, note := range d.NotesWithActionRequired {
			writeNote(note)
			nl()
		}
	}

	// each Kind gets a section
	sortedKinds := sortKinds(d.NotesByKind)
	if len(sortedKinds) > 0 {
		o.WriteString("## Changes by Kind")
		nlnl()
		for _, kind := range sortedKinds {
			o.WriteString("### ")
			o.WriteString(prettyKind(kind))
			nlnl()

			sort.Strings(d.NotesByKind[kind])
			for _, note := range d.NotesByKind[kind] {
				writeNote(note)
			}
			nl()
		}
		nlnl()
	}

	return strings.TrimSpace(o.String()), nil
}

// sortKinds sorts kinds by their priority and returns the result in a string
// slice
func sortKinds(notesByKind NotesByKind) []Kind {
	res := []Kind{}
	for kind := range notesByKind {
		res = append(res, kind)
	}

	indexOf := func(kind Kind) int {
		for i, prioKind := range kindPriority {
			if kind == prioKind {
				return i
			}
		}
		return -1
	}

	sort.Slice(res, func(i, j int) bool {
		return indexOf(res[i]) < indexOf(res[j])
	})

	return res
}

// CreateDownloadsTable creates the markdown table with the links to the tarballs.
// The function does nothing if the `tars` variable is empty.
func CreateDownloadsTable(w io.Writer, bucket, tars, prevTag, newTag string) error {
	if prevTag == "" || newTag == "" {
		return errors.New("release tags not specified")
	}

	urlPrefix := release.URLPrefixForBucket(bucket)
	fileMetadata, err := fetchMetadata(tars, urlPrefix, newTag)
	if fileMetadata == nil {
		// If directory is empty, doesn't contain matching files, or is not
		// given we will have a nil value. This is not an error in every
		// context. Return early so we do not modify markdown. This will be
		// removed once issue #1019 lands.
		fmt.Fprintf(w, "# %s\n\n", newTag)
		fmt.Fprintf(w, "## Changelog since %s\n\n", prevTag)
		return nil
	}

	if err != nil {
		return errors.Wrap(err, "fetching downloads metadata")
	}

	fmt.Fprintf(w, "# %s\n\n", newTag)
	fmt.Fprintf(w, "[Documentation](https://docs.k8s.io)\n\n")

	fmt.Fprintf(w, "## Downloads for %s\n\n", newTag)

	// Sort the files by their headers
	headers := [4]string{
		"", "Client Binaries", "Server Binaries", "Node Binaries",
	}
	files := map[string][]File{
		headers[0]: fileMetadata.Source,
		headers[1]: fileMetadata.Client,
		headers[2]: fileMetadata.Server,
		headers[3]: fileMetadata.Node,
	}

	for _, header := range headers {
		if header != "" {
			fmt.Fprintf(w, "### %s\n\n", header)
		}
		fmt.Fprintln(w, "filename | sha512 hash")
		fmt.Fprintln(w, "-------- | -----------")

		for _, f := range files[header] {
			fmt.Fprintf(w, "[%s](%s) | `%s`\n", f.Name, f.URL, f.Checksum)
		}
		fmt.Fprintln(w, "")
	}

	fmt.Fprintf(w, "## Changelog since %s\n\n", prevTag)
	return nil
}

func highestPriorityKind(kinds []string) Kind {
	for _, prioKind := range kindPriority {
		for _, k := range kinds {
			kind := Kind(k)
			if kind == prioKind {
				return kind
			}
		}
	}

	// Kind not in priority slice, returning the first one
	return Kind(kinds[0])
}

func mapKind(kind Kind) Kind {
	if newKind, ok := kindMap[kind]; ok {
		return newKind
	}
	return kind
}

func prettyKind(kind Kind) string {
	if kind == KindAPIChange {
		return "API Change"
	} else if kind == KindFailingTest {
		return "Failing Test"
	} else if kind == KindBugCleanupFlake {
		return string(KindBugCleanupFlake)
	}
	return strings.Title(string(kind))
}

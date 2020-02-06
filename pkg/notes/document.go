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

package notes

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// Document represents the underlying structure of a release notes document.
type Document struct {
	Kinds         map[string][]string `json:"kinds"`
	Uncategorized []string            `json:"uncategorized"`
}

const (
	KindAPIChange       = "api-change"
	KindBug             = "bug"
	KindCleanup         = "cleanup"
	KindDeprecation     = "deprecation"
	KindDesign          = "design"
	KindDocumentation   = "documentation"
	KindFailingTest     = "failing-test"
	KindFeature         = "feature"
	KindFlake           = "flake"
	KindBugCleanupFlake = "Other (Bug, Cleanup or Flake)"
)

var kindPriority = []string{
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
}

var kindMap = map[string]string{
	KindBug:     KindBugCleanupFlake,
	KindCleanup: KindBugCleanupFlake,
	KindFlake:   KindBugCleanupFlake,
}

// CreateDocument assembles an organized document from an unorganized set of
// release notes
func CreateDocument(notes ReleaseNotes, history ReleaseNotesHistory) (*Document, error) {
	doc := &Document{
		Kinds:         map[string][]string{},
		Uncategorized: []string{},
	}

	for _, pr := range history {
		note := notes[pr]

		if note.DuplicateKind {
			kind := mapKind(highestPriorityKind(note.Kinds))
			existingNotes, ok := doc.Kinds[kind]
			if ok {
				doc.Kinds[kind] = append(existingNotes, note.Markdown)
			} else {
				doc.Kinds[kind] = []string{note.Markdown}
			}
		} else {
			for _, kind := range note.Kinds {
				mappedKind := mapKind(kind)
				notesForKind, ok := doc.Kinds[mappedKind]
				if ok {
					doc.Kinds[mappedKind] = append(notesForKind, note.Markdown)
				} else {
					doc.Kinds[mappedKind] = []string{note.Markdown}
				}
			}

			if len(note.Kinds) == 0 {
				// the note has not been categorized so far
				doc.Uncategorized = append(doc.Uncategorized, note.Markdown)
			}
		}
	}
	return doc, nil
}

// RenderMarkdown accepts a Document and writes a version of that document to
// supplied io.Writer in markdown format.
func RenderMarkdown(doc *Document, bucket, tars, prevTag, newTag string) (string, error) {
	o := &strings.Builder{}
	if err := createDownloadsTable(o, bucket, tars, prevTag, newTag); err != nil {
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

	// each Kind gets a section
	sortedKinds := sortKinds(doc.Kinds)
	if len(sortedKinds) > 0 {
		o.WriteString("## Changes by Kind")
		nlnl()
		for _, kind := range sortedKinds {
			o.WriteString("### ")
			o.WriteString(prettyKind(kind))
			nlnl()

			sort.Strings(doc.Kinds[kind])
			for _, note := range doc.Kinds[kind] {
				writeNote(note)
			}
			nl()
		}
		nlnl()
	}

	// We call the uncategorized notes "Other Changes". These are changes
	// without any kind
	if len(doc.Uncategorized) > 0 {
		o.WriteString("## Other Changes")
		nlnl()
		for _, note := range doc.Uncategorized {
			writeNote(note)
		}
		nlnl()
	}

	return o.String(), nil
}

// sortKinds sorts kinds by their priority and returns the result in a string
// slice
func sortKinds(kinds map[string][]string) []string {
	res := []string{}
	for kind := range kinds {
		res = append(res, kind)
	}

	indexOf := func(kind string) int {
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

// prettySIG takes a sig name as parsed by the `sig-foo` label and returns a
// "pretty" version of it that can be printed in documents
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
			parts[i] = strings.Title(part)
		}
	}
	return strings.Join(parts, " ")
}

func prettifySigList(sigs []string) string {
	sigList := ""

	// sort the list so that any group of SIGs with the same content gives us the
	// same result
	sort.Strings(sigs)

	for i, sig := range sigs {
		if i == 0 {
			sigList = fmt.Sprintf("SIG %s", prettySIG(sig))
		} else if i == len(sigs)-1 {
			sigList = fmt.Sprintf("%s, and SIG %s", sigList, prettySIG(sig))
		} else {
			sigList = fmt.Sprintf("%s, SIG %s", sigList, prettySIG(sig))
		}
	}

	return sigList
}

// createDownloadsTable creates the markdown table with the links to the tarballs.
// The function does nothing if the `tars` variable is empty.
func createDownloadsTable(w io.Writer, bucket, tars, prevTag, newTag string) error {
	// Do not add the table if not explicitly requested
	if tars == "" {
		return nil
	}
	if prevTag == "" || newTag == "" {
		return errors.New("release tags not specified")
	}

	fmt.Fprintf(w, "# %s\n\n", newTag)
	fmt.Fprintf(w, "[Documentation](https://docs.k8s.io)\n\n")

	fmt.Fprintf(w, "## Downloads for %s\n\n", newTag)

	urlPrefix := fmt.Sprintf("https://storage.googleapis.com/%s/release", bucket)
	if bucket == "kubernetes-release" {
		urlPrefix = "https://dl.k8s.io"
	}

	for _, item := range []struct {
		heading  string
		patterns []string
	}{
		{"", []string{"kubernetes.tar.gz", "kubernetes-src.tar.gz"}},
		{"Client Binaries", []string{"kubernetes-client*.tar.gz"}},
		{"Server Binaries", []string{"kubernetes-server*.tar.gz"}},
		{"Node Binaries", []string{"kubernetes-node*.tar.gz"}},
	} {
		if item.heading != "" {
			fmt.Fprintf(w, "### %s\n\n", item.heading)
		}
		fmt.Fprintln(w, "filename | sha512 hash")
		fmt.Fprintln(w, "-------- | -----------")

		for _, pattern := range item.patterns {
			pattern := filepath.Join(tars, pattern)

			matches, err := filepath.Glob(pattern)
			if err != nil {
				return err
			}

			for _, file := range matches {
				f, err := os.Open(file)
				if err != nil {
					return err
				}
				defer f.Close()

				h := sha512.New()
				if _, err := io.Copy(h, f); err != nil {
					return err
				}

				fileName := filepath.Base(file)
				fmt.Fprintf(w,
					"[%s](%s/%s/%s) | `%x`\n",
					fileName, urlPrefix, newTag, fileName, h.Sum(nil),
				)
			}
		}

		fmt.Fprintln(w, "")
	}

	fmt.Fprintf(w, "## Changelog since %s\n\n", prevTag)
	return nil
}

func highestPriorityKind(kinds []string) string {
	for _, prioKind := range kindPriority {
		for _, kind := range kinds {
			if kind == prioKind {
				return kind
			}
		}
	}

	// Kind not in priotiy slice, returning the first one
	return kinds[0]
}

func mapKind(kind string) string {
	if newKind, ok := kindMap[kind]; ok {
		return newKind
	}
	return kind
}

func prettyKind(kind string) string {
	if kind == KindAPIChange {
		return "API Change"
	} else if kind == KindFailingTest {
		return "Failing Test"
	} else if kind == KindBugCleanupFlake {
		return KindBugCleanupFlake
	}
	return strings.Title(kind)
}

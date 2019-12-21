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
	NewFeatures    []string            `json:"new_features"`
	ActionRequired []string            `json:"action_required"`
	APIChanges     []string            `json:"api_changes"`
	Duplicates     map[string][]string `json:"duplicate_notes"`
	SIGs           map[string][]string `json:"sigs"`
	BugFixes       []string            `json:"bug_fixes"`
	Uncategorized  []string            `json:"uncategorized"`
}

// CreateDocument assembles an organized document from an unorganized set of
// release notes
func CreateDocument(notes ReleaseNotes, history ReleaseNotesHistory) (*Document, error) {
	doc := &Document{
		NewFeatures:    []string{},
		ActionRequired: []string{},
		APIChanges:     []string{},
		Duplicates:     map[string][]string{},
		SIGs:           map[string][]string{},
		BugFixes:       []string{},
		Uncategorized:  []string{},
	}

	for _, pr := range history {
		note := notes[pr]

		if note.ActionRequired {
			doc.ActionRequired = append(doc.ActionRequired, note.Markdown)
		} else if note.Feature {
			doc.NewFeatures = append(doc.NewFeatures, note.Markdown)
		} else if note.Duplicate {
			header := prettifySigList(note.SIGs)
			existingNotes, ok := doc.Duplicates[header]
			if ok {
				doc.Duplicates[header] = append(existingNotes, note.Markdown)
			} else {
				doc.Duplicates[header] = []string{note.Markdown}
			}
		} else {
			categorized := false

			for _, sig := range note.SIGs {
				categorized = true
				notesForSIG, ok := doc.SIGs[sig]
				if ok {
					doc.SIGs[sig] = append(notesForSIG, note.Markdown)
				} else {
					doc.SIGs[sig] = []string{note.Markdown}
				}
			}
			isBug := false
			for _, kind := range note.Kinds {
				switch kind {
				case "bug":
					// if the PR has kind/bug, we want to make a note of it, but we don't
					// include it in the Bug Fixes section until we haven't processed all
					// kinds and determined that it has no other categorization label.
					isBug = true
				case "feature":
					continue
				case "api-change", "new-api":
					categorized = true
					doc.APIChanges = append(doc.APIChanges, note.Markdown)
				}
			}

			// if the note has not been categorized so far, we can toss in one of two
			// buckets
			if !categorized {
				if isBug {
					doc.BugFixes = append(doc.BugFixes, note.Markdown)
				} else {
					doc.Uncategorized = append(doc.Uncategorized, note.Markdown)
				}
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

	// we always want to render the document with SIGs in alphabetical order
	sortedSIGs := []string{}
	for sig := range doc.SIGs {
		sortedSIGs = append(sortedSIGs, sig)
	}
	sort.Strings(sortedSIGs)

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

	// the "Action Required" section
	if len(doc.ActionRequired) > 0 {
		o.WriteString("## Action Required")
		nlnl()
		for _, note := range doc.ActionRequired {
			writeNote(note)
		}
		nlnl()
	}

	// the "New Feautres" section
	if len(doc.NewFeatures) > 0 {
		o.WriteString("## New Features")
		nlnl()
		for _, note := range doc.NewFeatures {
			writeNote(note)
		}
		nlnl()
	}

	// the "API Changes" section
	if len(doc.APIChanges) > 0 {
		o.WriteString("### API Changes")
		nlnl()
		for _, note := range doc.APIChanges {
			writeNote(note)
		}
		nlnl()
	}

	// the "Duplicate Notes" section
	if len(doc.Duplicates) > 0 {
		o.WriteString("### Notes from Multiple SIGs")
		nlnl()
		for header, notes := range doc.Duplicates {
			o.WriteString("#### ")
			o.WriteString(header)
			nlnl()
			for _, note := range notes {
				writeNote(note)
			}
			nl()
		}
		nl()
	}

	// each SIG gets a section (in alphabetical order)
	if len(sortedSIGs) > 0 {
		o.WriteString("### Notes from Individual SIGs")
		nlnl()
		for _, sig := range sortedSIGs {
			o.WriteString("#### SIG ")
			o.WriteString(prettySIG(sig))
			nlnl()
			for _, note := range doc.SIGs[sig] {
				writeNote(note)
			}
			nl()
		}
		nlnl()
	}

	// the "Bug Fixes" section
	if len(doc.BugFixes) > 0 {
		o.WriteString("### Bug Fixes")
		nlnl()
		for _, note := range doc.BugFixes {
			writeNote(note)
		}
		nlnl()
	}

	// we call the uncategorized notes "Other Notable Changes". ideally these
	// notes would at least have a SIG label.
	if len(doc.Uncategorized) > 0 {
		o.WriteString("### Other Notable Changes")
		nlnl()
		for _, note := range doc.Uncategorized {
			writeNote(note)
		}
		nlnl()
	}

	return o.String(), nil
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
		} else if i == (len(sigs) - 1) {
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

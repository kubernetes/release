package notes

import (
	"io"
	"sort"
	"strings"
)

// Document represents the underlying structure of a release notes document.
type Document struct {
	NewFeatures    []string            `json:"new_features"`
	ActionRequired []string            `json:"action_required"`
	APIChanges     []string            `json:"api_changes"`
	SIGs           map[string][]string `json:"sigs"`
	BugFixes       []string            `json:"bug_fixes"`
	Uncategorized  []string            `json:"uncategorized"`
}

// CreateDocument assembles an organized document from an unorganized set of
// release notes
func CreateDocument(notes []*ReleaseNote) (*Document, error) {
	doc := &Document{
		NewFeatures:    []string{},
		ActionRequired: []string{},
		APIChanges:     []string{},
		SIGs:           map[string][]string{},
		BugFixes:       []string{},
		Uncategorized:  []string{},
	}

	for _, note := range notes {
		categorized := false

		if note.ActionRequired {
			categorized = true
			doc.ActionRequired = append(doc.ActionRequired, note.Markdown)
		}

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
				categorized = true
				doc.NewFeatures = append(doc.NewFeatures, note.Markdown)
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

	return doc, nil
}

// RenderMarkdown accepts a Document and writes a version of that document to
// supplied io.Writer in markdown format.
func RenderMarkdown(doc *Document, w io.Writer) error {
	// we always want to render the document with SIGs in alphabetical order
	sortedSIGs := []string{}
	for sig, _ := range doc.SIGs {
		sortedSIGs = append(sortedSIGs, sig)
	}
	sort.Strings(sortedSIGs)

	// this is a helper so that we don't have to check err != nil on every write

	// first, we create a long-lived err that we can re-use
	var err error

	// write is a helper that writes a string to the in-scope io.Writer w
	write := func(s string) {
		// if write has already failed, just return and don't do anything
		if err != nil {
			return
		}
		// perform the write
		_, err = w.Write([]byte(s))
	}

	// writeNote encapsulates the pre-processing that might happen on a note text
	// before it gets bulleted and written to the io.Writer
	writeNote := func(s string) {
		if !strings.HasPrefix(s, "- ") {
			s = "- " + s
		}
		write(s + "\n")
	}

	// the "Action Required" section
	if len(doc.ActionRequired) > 0 {
		write("## Action Required\n\n")
		for _, note := range doc.ActionRequired {
			writeNote(note)
		}
		write("\n\n")
	}

	// the "New Feautres" section
	if len(doc.NewFeatures) > 0 {
		write("## New Features\n\n")
		for _, note := range doc.NewFeatures {
			writeNote(note)
		}
		write("\n\n")
	}

	// the "API Changes" section
	if len(doc.APIChanges) > 0 {
		write("## API Changes\n\n")
		for _, note := range doc.APIChanges {
			writeNote(note)
		}
		write("\n\n")
	}

	// each SIG gets a section (in alphabetical order)
	for _, sig := range sortedSIGs {
		write("## SIG " + prettySIG(sig) + "\n\n")
		for _, note := range doc.SIGs[sig] {
			writeNote(note)
		}
		write("\n\n")
	}

	// the "Bug Fixes" section
	if len(doc.BugFixes) > 0 {
		write("## Bug Fixes\n\n")
		for _, note := range doc.BugFixes {
			writeNote(note)
		}
		write("\n\n")
	}

	// we call the uncategorized notes "Other Notable Changes". ideally these
	// notes would at least have a SIG label.
	if len(doc.Uncategorized) > 0 {
		write("## Other Notable Changes\n\n")
		for _, note := range doc.Uncategorized {
			writeNote(note)
		}
		write("\n\n")
	}

	return err
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

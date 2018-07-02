package notes

import (
	"io"
	"sort"
	"strings"
)

type Document struct {
	SIGs           map[string][]string `json:"sigs"`
	BugFixes       []string            `json:"bug_fixes"`
	ActionRequired []string            `json:"action_required"`
	Uncategorized  []string            `json:"uncategorized"`
}

func CreateDocument(notes []*ReleaseNote) (*Document, error) {
	doc := &Document{
		SIGs:           map[string][]string{},
		BugFixes:       []string{},
		ActionRequired: []string{},
		Uncategorized:  []string{},
	}

	for _, note := range notes {
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

		for _, kind := range note.Kinds {
			if kind == "bug" {
				categorized = true
				doc.BugFixes = append(doc.BugFixes, note.Markdown)
			}
		}

		if note.ActionRequired {
			categorized = true
			doc.ActionRequired = append(doc.ActionRequired, note.Markdown)
		}

		if !categorized {
			doc.Uncategorized = append(doc.Uncategorized, note.Markdown)
		}
	}

	return doc, nil
}

func RenderMarkdown(doc *Document, w io.Writer) error {
	sortedSIGs := []string{}
	for sig, _ := range doc.SIGs {
		sortedSIGs = append(sortedSIGs, sig)
	}
	sort.Strings(sortedSIGs)

	var err error
	write := func(s string) {
		if err != nil {
			return
		}
		_, err = w.Write([]byte(s))
	}

	writeNote := func(s string) {
		if !strings.HasPrefix(s, "- ") {
			s = "- " + s
		}
		write(s + "\n")
	}

	if len(doc.ActionRequired) > 0 {
		write("## Action Required\n\n")
		for _, note := range doc.ActionRequired {
			writeNote(note)
		}
		write("\n\n")
	}

	for _, sig := range sortedSIGs {
		write("## SIG " + prettySIG(sig) + "\n\n")
		for _, note := range doc.SIGs[sig] {
			writeNote(note)
		}
		write("\n\n")
	}

	if len(doc.BugFixes) > 0 {
		write("## Bug Fixes\n\n")
		for _, note := range doc.BugFixes {
			writeNote(note)
		}
		write("\n\n")
	}

	if len(doc.Uncategorized) > 0 {
		write("## Other Notable Changes\n\n")
		for _, note := range doc.Uncategorized {
			writeNote(note)
		}
		write("\n\n")
	}

	return err
}

func prettySIG(sig string) string {
	parts := strings.Split(sig, "-")
	for i, part := range parts {
		switch part {
		case "vsphere":
			parts[i] = "vSphere"
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

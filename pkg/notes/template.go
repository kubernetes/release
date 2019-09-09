package notes

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type Template struct {
	UpgradeNotes [14]SIG
	Metrics      Metrics
	Features     Features
	APIChanges   []string
	Other        [14]SIG
}

type SIG struct {
	Header  string
	Content []string
}

var availableSIGs = map[string]definedSIG{
	"api-machinery":     {"API Machinery", 0},
	"apps":              {"Apps", 1},
	"auth":              {"Auth", 2},
	"autoscaling":       {"Autoscaling", 3},
	"cli":               {"CLI", 4},
	"cloud-provider":    {"Cloud Provider", 5},
	"cluster-lifecycle": {"Cluster Lifecycle", 6},
	"instrumentation":   {"Instrumentation", 7},
	"network":           {"Network", 8},
	"node":              {"Node", 9},
	"release":           {"Release", 10},
	"scheduling":        {"Scheduling", 11},
	"storage":           {"Storage", 12},
	"windows":           {"Windows", 13},
}

type definedSIG struct {
	name string
	id   int
}

type Metrics struct {
	Changes    []string
	Added      []string
	Removed    []string
	Deprecated []string
}

type Features struct {
	Stable              []string
	Beta                []string
	Alpha               []string
	StagingRepositories []string
	CliImprovements     []string
	Misc                []string
}

const notesTemplate = `## What’s New (Major Themes)

<!-- Add themes from Comms Blog here -->

## Known Issues

<!-- Add issues from known issues bucket (known-issues-bucket.Md) here -->

## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)
{{range .UpgradeNotes}}{{if .Header}}
#### {{.Header}}

{{range .Content}}{{.}}
{{end}}{{end}}{{end}}
## Deprecations and Removals

<!--
Add it in the format
- Component
  - deprecation
  - removal
-->

## Metrics Changes

{{range .Metrics.Changes}}{{.}}
{{end}}
### Added metrics

{{range .Metrics.Added}}{{.}}
{{end}}
### Removed metrics

{{range .Metrics.Removed}}{{.}}
{{end}}
### Deprecated/changed metrics

{{range .Metrics.Deprecated}}{{.}}
{{end}}
## Notable Features

### Stable

{{range .Features.Stable}}{{.}}
{{end}}
### Beta

{{range .Features.Beta}}{{.}}
{{end}}
### Alpha

{{range .Features.Alpha}}{{.}}
{{end}}
### Staging Repositories

{{range .Features.StagingRepositories}}{{.}}
{{end}}
### CLI Improvements

{{range .Features.CliImprovements}}{{.}}
{{end}}
### Misc

{{range .Features.Misc}}{{.}}
{{end}}
## API Changes

{{range .APIChanges}}{{.}}
{{end}}
## Other notable changes
{{range .Other}}{{if .Header}}
### {{ .Header }}

{{range .Content}}{{.}}
{{end}}{{end}}{{end}}
## Dependencies

<!-- Add by hand -->
`

func CreateFromTemplate(logger log.Logger, notes ReleaseNotes) (*Template, error) {
	res := &Template{}
	for _, note := range notes {
		for _, sig := range note.SIGs {

			definedSIG, ok := availableSIGs[sig]
			if !ok {
				level.Warn(logger).Log("msg", "unable to find ID for SIG "+sig)
				continue
			}
			i := definedSIG.id

			if note.ActionRequired {
				res.UpgradeNotes[i].setHeaderIfNeeded(definedSIG.name)
				res.UpgradeNotes[i].Content = markdownAppend(res.UpgradeNotes[i].Content, note)

			} else if wordIn(note, "metrics") {
				res.Metrics.classifyMetricsAppend(note)

			} else if wordIn(note, "api") {
				res.APIChanges = markdownAppend(res.APIChanges, note)

			} else if note.Feature {
				res.Features.classifyFeaturesAppend(note)

			} else {
				res.Other[i].setHeaderIfNeeded(definedSIG.name)
				res.Other[i].Content = markdownAppend(res.Other[i].Content, note)
			}
		}
	}

	return res, nil
}

func (m Metrics) classifyMetricsAppend(note *ReleaseNote) {
	if wordIn(note, "added") {
		m.Added = markdownAppend(m.Added, note)

	} else if wordIn(note, "removed") {
		m.Removed = markdownAppend(m.Removed, note)

	} else if wordIn(note, "deprecated") {
		m.Deprecated = markdownAppend(m.Deprecated, note)

	} else {
		m.Changes = markdownAppend(m.Changes, note)
	}
}

func (f Features) classifyFeaturesAppend(note *ReleaseNote) {
	if wordIn(note, "alpha") {
		f.Alpha = markdownAppend(f.Alpha, note)

	} else if wordIn(note, "beta") {
		f.Beta = markdownAppend(f.Beta, note)

	} else if wordIn(note, "stable") {
		f.Stable = markdownAppend(f.Stable, note)

	} else if wordIn(note, "staging") {
		f.StagingRepositories = markdownAppend(f.StagingRepositories, note)

	} else if wordIn(note, "cli") {
		f.CliImprovements = markdownAppend(f.CliImprovements, note)

	} else {
		f.Misc = markdownAppend(f.Misc, note)
	}
}

func wordIn(note *ReleaseNote, word string) bool {
	str, word := strings.ToUpper(note.Markdown), strings.ToUpper(word)
	return strings.Contains(str, fmt.Sprintf(" %s ", word))
}

func markdownAppend(input []string, note *ReleaseNote) []string {
	s := note.Markdown
	if !strings.HasPrefix(s, "- ") {
		s = "- " + s
	}
	// Do not add duplicates
	for _, n := range input {
		if n == s {
			return input
		}
	}
	return append(input, s)
}

func (s *SIG) setHeaderIfNeeded(to string) {
	if s.Header == "" {
		s.Header = to
	}
}

// RenderTemplate writes the template to the provided writer
func (t *Template) RenderTemplate(w io.Writer) error {
	const templateName = "notes"
	tpl, err := template.New(templateName).Parse(notesTemplate)
	if err != nil {
		return err
	}
	return tpl.ExecuteTemplate(w, templateName, t)
}

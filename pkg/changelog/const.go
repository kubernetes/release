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

package changelog

const (
	// The default CHANGELOG directory inside the k/k repository.
	RepoChangelogDir = "CHANGELOG"

	nl                   = "\n"
	tocStart             = "<!-- BEGIN MUNGE: GENERATED_TOC -->"
	TocEnd               = "<!-- END MUNGE: GENERATED_TOC -->"
	releaseNotesTemplate = `
{{- $CurrentRevision := .CurrentRevision -}}
{{- $PreviousRevision := .PreviousRevision -}}
# {{$CurrentRevision}}

{{if .Downloads}}
## Downloads for {{$CurrentRevision}}

{{- with .Downloads.Source }}

### Source Code

filename | sha512 hash
-------- | -----------
{{range .}}[{{.Name}}]({{.URL}}) | {{.Checksum}}{{println}}{{end}}
{{end}}

{{- with .Downloads.Client -}}
### Client Binaries

filename | sha512 hash
-------- | -----------
{{range .}}[{{.Name}}]({{.URL}}) | {{.Checksum}}{{println}}{{end}}
{{end}}

{{- with .Downloads.Server -}}
### Server Binaries

filename | sha512 hash
-------- | -----------
{{range .}}[{{.Name}}]({{.URL}}) | {{.Checksum}}{{println}}{{end}}
{{end}}

{{- with .Downloads.Node -}}
### Node Binaries

filename | sha512 hash
-------- | -----------
{{range .}}[{{.Name}}]({{.URL}}) | {{.Checksum}}{{println}}{{end}}
{{end -}}
{{- end -}}
## Changelog since {{$PreviousRevision}}

{{with .CVEList -}}
## Important Security Information

This release contains changes that address the following vulnerabilities:
{{range .}}
### {{.ID}}: {{.Title}}

{{.Description}}

**CVSS Rating:** {{.CVSSRating}} ({{.CVSSScore}}) [{{.CVSSVector}}]({{.CalcLink}})
{{- if .TrackingIssue -}}
<br>
**Tracking Issue:** {{.TrackingIssue}}
{{- end }}

{{ end }}
{{- end -}}

{{with .NotesWithActionRequired -}}
## Urgent Upgrade Notes

### (No, really, you MUST read this before you upgrade)

{{range .}} {{println "-" .}} {{end}}
{{end}}

{{- if .Notes -}}
## Changes by Kind
{{ range .Notes}}
### {{.Kind | prettyKind}}

{{range $note := .NoteEntries }}{{println "-" $note}}{{end}}
{{- end -}}
{{- end -}}
`

	htmlTemplate = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width" />
    <title>{{ .Title }}</title>
    <style type="text/css">
      table,
      th,
      tr,
      td {
        border: 1px solid gray;
        border-collapse: collapse;
        padding: 5px;
      }
    </style>
  </head>
  <body>
    {{ .Content }}
  </body>
</html>`
)

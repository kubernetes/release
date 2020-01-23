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

package patch

const hr = "\n\n----\n\n"

const MailHeadMarkdown = `
Below is a draft of the generated changelog for {{ .Version }}. If you submitted a cherrypick, please make sure it's listed and has an **accurate release note**.

If you have a pending cherrypick for {{ .Version }}, make sure it merges by end of day on **{{ dateFormatHuman .DateFreeze }}**.
Please tag {{ code .ReleaseManagerTag }} on the GitHub issue/PR, email {{ link .ReleaseManagerName (print "mailto:" .ReleaseManagerEmail) }}, or reach out in {{ link (print "#" .ReleaseManagerSlackChannel) (printf "https://kubernetes.slack.com/messages/%s/" .ReleaseManagerSlackChannel) }} on Slack if your cherrypick appears to be blocked on something out of your control.

If you've already spoken to the patch release team about PRs that are not yet merged or listed below, don't worry, we're tracking them.
`

const MailStyle = `
<style type="text/css">
body {
  font-family: "Verdana"
}
code, pre {
  background-color: #f2f2f2;
}
code {
  display: inline-block;
  padding: 2px;
}
pre {
  padding: 1em;
}
tr:nth-child(even) td {
  background-color: #f2f2f2;
}
td {
  padding: 1em;
}
hr {
  border: 4px solid #f2f2f2;
  margin-left: 1em;
  margin-right: 1em;
  border-radius: 2px;
}
</style>
`

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

package internal

import (
	"fmt"
	"strings"

	"k8s.io/release/pkg/log"
)

type Formatter struct {
	log.Mixin

	CommandCreator CommandCreator
	Style          string
}

const formatterScript = `
set -euo pipefail
pandoc \
	--standalone \
	--columns=10000000 \
	--from=gfm \
	--to=html5 \
	--metadata=pagetitle="${TITLE}" \
	--include-in-header=<(echo "${STYLE}") \
	--output=-
`

func (r *Formatter) MarkdownToHTML(markdown, title string) (string, error) {
	cmd := r.CommandCreator.create("bash", "-c", formatterScript)
	if cmd == nil {
		return "", fmt.Errorf("command is nil")
	}
	r.Logger().Debug("command created")

	cmd.SetStdin(strings.NewReader(markdown))
	cmd.SetEnv([]string{
		"TITLE=" + title,
		"STYLE=" + r.Style,
	})

	s, err := cmdOutput(cmd)
	if err != nil {
		r.Logger().WithError(err).Debug("execing & getting output failed")
		r.Logger().WithField("error", err.FullError()).Trace("full exec error")
		return "", err
	}
	return s, nil
}

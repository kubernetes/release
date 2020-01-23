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
	"path/filepath"

	"k8s.io/release/pkg/log"
)

const relnoteScript = `
set -euo pipefail
tmp="$( mktemp )"
trap 'rm -f -- "${tmp}"' EXIT
%q --htmlize-md --preview --quiet --markdown-file="${tmp}" >&2
cat "${tmp}"
`

type ReleaseNoter struct {
	log.Mixin

	ReleaseToolsDir string
	K8sDir          string
	GithubToken     string

	CommandCreator CommandCreator
}

func (r *ReleaseNoter) GetMarkdown() (string, error) {
	binPath, err := filepath.Abs(filepath.Join(r.ReleaseToolsDir, "relnotes"))
	if err != nil {
		return "", fmt.Errorf("could not determine current working directory")
	}
	r.Logger().WithField("binpath", binPath).Debug("binpath set")

	cmd := r.CommandCreator.create(
		"bash", "-c", fmt.Sprintf(relnoteScript, binPath),
	)
	if cmd == nil {
		return "", fmt.Errorf("command is nil")
	}
	r.Logger().Debug("command created")

	cmd.SetDir(r.K8sDir)
	cmd.SetEnv([]string{
		"GITHUB_TOKEN=" + r.GithubToken,
	})

	r.Logger().WithField("workdir", r.K8sDir).Info("starting release notes gatherer ... this may take a while ...")

	s, eerr := cmdOutput(cmd)
	if eerr != nil {
		r.Logger().WithError(eerr).Debug("execing & getting output failed")
		r.Logger().WithField("error", eerr.FullError()).Trace("full exec error")
		return "", eerr
	}
	return s, nil
}

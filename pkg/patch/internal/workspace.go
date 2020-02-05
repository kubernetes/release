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

type Workspace struct {
	log.Mixin

	K8sRepoPath    string
	CommandCreator CommandCreator
}

type status = map[string]string

func (w *Workspace) Status() (status, error) {
	execPath := "hack/print-workspace-status.sh"

	cmd := w.CommandCreator.create(execPath)
	if cmd == nil {
		return nil, fmt.Errorf("command is nil")
	}

	cmd.SetDir(w.K8sRepoPath)

	s, eerr := cmdOutput(cmd)
	if eerr != nil {
		w.Logger().WithError(eerr).Debug("execing & getting output failed")
		w.Logger().WithField("error", eerr.FullError()).Trace("full exec error")
		return nil, eerr
	}

	lines := strings.Split(s, "\n")
	statuses := make(status, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		t := strings.SplitN(line, " ", 2)
		if len(t) != 2 {
			return nil, fmt.Errorf("cannot parse workspace status line %q", line)
		}
		statuses[t[0]] = t[1]
	}

	return statuses, nil
}

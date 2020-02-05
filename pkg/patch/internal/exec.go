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

	exec "k8s.io/utils/exec"
)

//counterfeiter:generate . Cmd
type Cmd = exec.Cmd

type CommandCreator func(path string, args ...string) Cmd

func (c CommandCreator) create(path string, args ...string) Cmd {
	if c == nil {
		c = exec.New().Command
	}
	return c(path, args...)
}

func cmdOutput(cmd Cmd) (string, *execErr) {
	b, err := cmd.Output()
	s := string(b)
	if err != nil {
		return "", &execErr{
			Err:    err,
			stdout: s,
			stderr: getStderr(err),
		}
	}
	return s, nil
}

type execErr struct {
	Err    error
	stdout string
	stderr string
}

func (ee *execErr) Error() string {
	return fmt.Sprintf("%s %s", ee.Err, ee.stdio(35))
}
func (ee *execErr) FullError() string {
	return fmt.Sprintf("%s %s", ee.Err, ee.stdio(0))
}
func (ee *execErr) stdio(maxLen int) string {
	truncater := func(s string) string {
		if maxLen < 1 {
			return s
		}
		return trunc(maxLen, s)
	}
	var s string
	if o := ee.stdout; o != "" {
		s += "Stdout: " + truncater(o)
	} else {
		s += "no Stdout"
	}
	s += ", "
	if o := ee.stderr; o != "" {
		s += "Stderr: " + truncater(o)
	} else {
		s += "no Stderr"
	}
	return "(" + s + ")"
}

func trunc(maxLen int, s string) string {
	if len(s) <= maxLen {
		return s
	}
	partLen := (maxLen - 5) / 2
	return s[:partLen] + " ... " + s[len(s)-partLen:]
}

func getStderr(err error) string {
	if eer, ok := err.(*exec.ExitErrorWrapper); ok {
		return string(eer.Stderr)
	}
	return ""
}

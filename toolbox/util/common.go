// Copyright 2017 The Kubernetes Authors All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Shell runs a command and returns the result as a string.
func Shell(name string, arg ...string) (string, error) {
	c := exec.Command(name, arg...)
	bytes, err := c.CombinedOutput()
	return string(bytes), err
}

// GetSha256 calculates SHA256 for input file.
func GetSha256(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// RenderProgressBar renders a progress bar by rewriting the current (assuming
// stdout outputs to a terminal console). If initRender is true, the function
// writes a new line instead of rewrite the current line.
func RenderProgressBar(progress, total int, duration string, initRender bool) {
	barLen := 80
	var progressLen, arrowLen, remainLen int
	var rewrite string

	percentage := float64(progress) / float64(total)

	progressLen = int(percentage * float64(barLen))
	if progressLen < barLen {
		arrowLen = 1
	}
	remainLen = barLen - progressLen - arrowLen

	if !initRender {
		rewrite = "\r"
	}

	fmt.Printf("%s%12s [%s%s%s] %7.2f%%", rewrite, duration, strings.Repeat("=", progressLen), strings.Repeat(">", arrowLen), strings.Repeat("-", remainLen), percentage*100.0)
}

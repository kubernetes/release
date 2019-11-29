/*
Copyright 2019 The Kubernetes Authors.

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

package util

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

/*
#############################################################################
# Simple yes/no prompt
#
# @optparam default -n(default)/-y/-e (default to n, y or make (e)xplicit)
# @param message
common::askyorn () {
  local yorn
  local def=n
  local msg="y/N"

  case $1 in
  -y) # yes default
      def="y" msg="Y/n"
      shift
      ;;
  -e) # Explicit
      def="" msg="y/n"
      shift
      ;;
  -n) shift
      ;;
  esac

  while [[ $yorn != [yYnN] ]]; do
    logecho -n "$*? ($msg): "
    read yorn
    : ${yorn:=$def}
  done

  # Final test to set return code
  [[ $yorn == [yY] ]]
}
*/

func Ask(question, expectedResponse string, retries int) (answer string, success bool, err error) {
	attempts := 1

	if retries < 0 {
		log.Printf("Retries was set to a number less than zero (%d). Please specify a positive number of retries or zero, if you want to ask unconditionally.", retries)
	}

	for attempts <= retries {
		scanner := bufio.NewScanner(os.Stdin)
		log.Printf("%s (%d/%d) ", question, attempts, retries)

		scanner.Scan()
		answer = scanner.Text()

		if answer == expectedResponse {
			return answer, true, nil
		}

		log.Printf("Expected '%s', but got '%s'", expectedResponse, answer)

		attempts++
	}

	log.Printf("Expected response was not provided. Retries exceeded.")
	return answer, false, errors.New("expected response was not input. Retries exceeded")
}

// FakeGOPATH creates a temp directory, links the base directory into it and
// sets the GOPATH environment variable to it.
func FakeGOPATH(srcDir string) (string, error) {
	log.Printf("Linking repository into temp dir")
	baseDir, err := ioutil.TempDir("", "ff-")
	if err != nil {
		return "", err
	}

	log.Printf("New working directory is %q", baseDir)

	os.Setenv("GOPATH", baseDir)
	log.Printf("GOPATH: %s", os.Getenv("GOPATH"))

	gitRoot := fmt.Sprintf("%s/src/k8s.io", baseDir)
	if err := os.MkdirAll(gitRoot, 0o755); err != nil {
		return "", err
	}
	gitRoot = filepath.Join(gitRoot, "kubernetes")

	// link the repo into the working directory
	log.Printf("Creating symlink from %q to %q", srcDir, gitRoot)
	if err := os.Symlink(srcDir, gitRoot); err != nil {
		return "", err
	}

	log.Printf("Changing working directory to %s", gitRoot)
	if err := os.Chdir(gitRoot); err != nil {
		return "", err
	}

	return gitRoot, nil
}

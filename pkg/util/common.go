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
	"log"
	"os"
	"os/exec"
	"strings"
)

// Run wraps the exec.Cmd.Run() command and sets the standard output.
// TODO: Should this take an error code argument/return an error code?
func Run(c *exec.Cmd) {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		log.Fatalf("Command %q failed: %v", strings.Join(c.Args, " "), err)
	}
}

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

func Ask(question, expectedResponse string, retries int) (string, bool, error) {
	attempts := 1
	answer := ""

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
	return answer, false, errors.New("Expected response was not input. Retries exceeded.")
}

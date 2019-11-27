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

package command

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"
)

// A generic command abstraction
type Command struct {
	cmd *exec.Cmd
}

// A generic command exit status
type Status struct {
	exitCode syscall.WaitStatus
	output   string
}

// New creates a new command from the provided arguments.
func New(cmd ...string) *Command {
	return NewWithWorkDir("", cmd...)
}

// New creates a new command from the provided workDir and the command
// arguments.
func NewWithWorkDir(workDir string, cmd ...string) *Command {
	args := strings.Fields(strings.Join(cmd, " "))

	command := func() *Command {
		if len(args) == 0 {
			return &Command{exec.Command(cmd[0])}
		} else if len(args) > 1 {
			return &Command{exec.Command(args[0], args[1:]...)}
		}
		return &Command{exec.Command(args[0])}
	}()

	if workDir != "" {
		command.cmd.Dir = workDir
	}

	return command
}

// Run starts the command and waits for it to finish. It returns an error if
// the command execution was not possible at all, otherwise the Status.
// This method prints the commands output during execution
func (c *Command) Run() (res *Status, err error) {
	return c.run(true)
}

// Run starts the command and waits for it to finish. It returns an error if
// the command execution was not successful.
func (c *Command) RunSuccess() (err error) {
	res, err := c.run(true)
	if err != nil {
		return err
	}
	if !res.Success() {
		return errors.Errorf("command %v did not succeed", c.cmd)
	}
	return nil
}

// Run starts the command and waits for it to finish. It returns an error if
// the command execution was not possible at all, otherwise the Status.
// This method does not print the output of the command during its execution.
func (c *Command) RunSilent() (res *Status, err error) {
	return c.run(false)
}

// Run starts the command and waits for it to finish. It returns an error if
// the command execution was not successful.
// This method does not print the output of the command during its execution.
func (c *Command) RunSilentSuccess() (err error) {
	res, err := c.run(false)
	if err != nil {
		return err
	}
	if !res.Success() {
		return errors.Errorf("command %v did not succeed", c.cmd)
	}
	return nil
}

// run is the internal run method
func (c *Command) run(print bool) (res *Status, err error) {
	log.Printf("Running command: %v", c.cmd)
	outBuffer := bytes.Buffer{}

	var stdout, stderr io.ReadCloser
	if stdout, err = c.cmd.StdoutPipe(); err != nil {
		return nil, err
	}
	if stderr, err = c.cmd.StderrPipe(); err != nil {
		return nil, err
	}

	var writer io.Writer
	if print {
		writer = io.MultiWriter(os.Stdout, &outBuffer)
	} else {
		writer = io.MultiWriter(&outBuffer)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, err := io.Copy(writer, stdout); err != nil {
			log.Println("unable to copy command stdout")
		}
		if _, err := io.Copy(writer, stderr); err != nil {
			log.Println("unable to copy command stderr")
		}
	}()

	status, err := c.runStatus()
	wg.Wait()

	if err != nil {
		return nil, err
	}

	return &Status{exitCode: status, output: outBuffer.String()}, nil
}

// runStatus returns the syscall wait status for the command by running it
func (c *Command) runStatus() (syscall.WaitStatus, error) {
	err := c.cmd.Run()

	if err == nil {
		return 0, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status, nil
		}
	}

	return 0, err
}

// Success returns if a Status was successful
func (s *Status) Success() bool {
	return s.exitCode == 0
}

// ExitCode returns the exit status of the command status
func (s *Status) ExitCode() int {
	return s.exitCode.ExitStatus()
}

// Output returns the combined stdout and stderr of the command status
func (s *Status) Output() string {
	return s.output
}

// Execute is a convenience function which creates a new Command, executes it
// and evaluates its status.
func Execute(cmd ...string) error {
	status, err := New(cmd...).Run()
	if err != nil {
		return errors.Wrapf(err, "command %q is not executable", cmd)
	}
	if !status.Success() {
		return errors.Errorf(
			"command %q did not exit successful (%d)",
			cmd, status.ExitCode(),
		)
	}
	return nil
}

// Available verifies that the specified `commands` are available within the
// current `$PATH` environment and returns true if so. The function does not
// check for duplicates nor if the provided slice is empty.
func Available(commands ...string) (ok bool) {
	ok = true
	for _, command := range commands {
		if _, err := exec.LookPath(command); err != nil {
			log.Printf("Unable to %v", err)
			ok = false
		}
	}
	return ok
}

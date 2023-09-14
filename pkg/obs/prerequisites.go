/*
Copyright 2023 The Kubernetes Authors.

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

package obs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-sdk/osc"
	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/env"
)

// PrerequisitesChecker is the main type for checking the prerequisites for
// OBS operations.
type PrerequisitesChecker struct {
	impl prerequisitesCheckerImpl
	opts *PrerequisitesCheckerOptions
}

// Type prerequisites checker
type PrerequisitesCheckerOptions struct {
	CheckOBSPassword bool
}

var DefaultPrerequisitesCheckerOptions = &PrerequisitesCheckerOptions{
	CheckOBSPassword: true,
}

// NewPrerequisitesChecker creates a new PrerequisitesChecker instance.
func NewPrerequisitesChecker() *PrerequisitesChecker {
	return &PrerequisitesChecker{
		&defaultPrerequisitesChecker{},
		DefaultPrerequisitesCheckerOptions,
	}
}

// Options return the options from the prereq checker
func (p *PrerequisitesChecker) Options() *PrerequisitesCheckerOptions {
	return p.opts
}

// SetImpl can be used to set the internal PrerequisitesChecker implementation.
func (p *PrerequisitesChecker) SetImpl(impl prerequisitesCheckerImpl) {
	p.impl = impl
}

//counterfeiter:generate . prerequisitesCheckerImpl
type prerequisitesCheckerImpl interface {
	CommandAvailable(commands ...string) bool
	OSCOutput(args ...string) (string, error)
	IsEnvSet(key string) bool
	Usage(dir string) (*disk.UsageStat, error)
}

type defaultPrerequisitesChecker struct{}

func (*defaultPrerequisitesChecker) CommandAvailable(
	commands ...string,
) bool {
	return command.Available(commands...)
}

func (*defaultPrerequisitesChecker) OSCOutput(
	args ...string,
) (string, error) {
	return osc.Output("", args...)
}

func (*defaultPrerequisitesChecker) IsEnvSet(key string) bool {
	return env.IsSet(key)
}

func (*defaultPrerequisitesChecker) Usage(dir string) (*disk.UsageStat, error) {
	return disk.Usage(dir)
}

func (p *PrerequisitesChecker) Run(workdir string) error {
	// Command checks
	commands := []string{"osc"}
	logrus.Infof(
		"Verifying that the commands %s exist in $PATH.",
		strings.Join(commands, ", "),
	)

	if !p.impl.CommandAvailable(commands...) {
		return errors.New("not all commands available")
	}

	// osc checks
	logrus.Info("Verifying OpenBuildService access")
	ver, err := p.impl.OSCOutput("version")
	if err != nil {
		return fmt.Errorf("running osc version: %w", err)
	}
	logrus.Infof("Using osc version: %s", ver)

	user, err := p.impl.OSCOutput("whois")
	if err != nil {
		return fmt.Errorf("running osc whois: %w", err)
	}
	logrus.Infof("Using OpenBuildService user: %s", user)

	// Environment checks
	if p.opts.CheckOBSPassword {
		logrus.Infof(
			"Verifying that %s environment variable is set", OBSPasswordKey,
		)
		if !p.impl.IsEnvSet(OBSPasswordKey) {
			return fmt.Errorf("no %s env variable set", OBSPasswordKey)
		}
	}

	// Disk space check
	const minDiskSpaceGiB = 10
	logrus.Infof(
		"Checking available disk space (%dGB) for %s", minDiskSpaceGiB, workdir,
	)
	res, err := p.impl.Usage(workdir)
	if err != nil {
		return fmt.Errorf("check available disk space: %w", err)
	}
	diskSpaceGiB := res.Free / 1024 / 1024 / 1024
	if diskSpaceGiB < minDiskSpaceGiB {
		return fmt.Errorf(
			"not enough disk space available. Got %dGiB, need at least %dGiB",
			diskSpaceGiB, minDiskSpaceGiB,
		)
	}

	return nil
}

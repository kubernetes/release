/*
Copyright 2026 The Kubernetes Authors.

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

package vex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/vuln/scan"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt vexfakes/fake_scanner_implementation.go > vexfakes/_fake_scanner_implementation.go && mv vexfakes/_fake_scanner_implementation.go vexfakes/fake_scanner_implementation.go"

// Scanner generates OpenVEX documents by running govulncheck on source code.
type Scanner struct {
	impl scannerImplementation
}

// NewScanner returns a scanner that runs govulncheck in-process.
func NewScanner() *Scanner {
	return &Scanner{impl: &defaultScannerImpl{}}
}

// SetImpl replaces the scanner implementation, for testing.
func (s *Scanner) SetImpl(impl scannerImplementation) {
	s.impl = impl
}

//counterfeiter:generate . scannerImplementation
type scannerImplementation interface {
	// ScanSource runs the scanner in source mode on the pattern, resolved
	// inside sourceDir, and returns its OpenVEX output.
	ScanSource(ctx context.Context, sourceDir, pattern string) ([]byte, error)
}

// ScanSource runs the vulnerability scanner in source mode on the pattern,
// resolved inside sourceDir, and returns its OpenVEX output.
func (s *Scanner) RunScan(ctx context.Context, sourceDir, pattern string) ([]byte, error) {
	sourceDir = strings.TrimSpace(sourceDir)
	pattern = strings.TrimSpace(pattern)

	if sourceDir == "" {
		return nil, errors.New("source directory is required")
	}

	if pattern == "" {
		return nil, errors.New("source pattern is required")
	}

	return s.impl.ScanSource(ctx, sourceDir, pattern)
}

// defaultScannerImpl runs govulncheck through its library API.
type defaultScannerImpl struct{}

// ScanSource runs govulncheck in source mode on the pattern, resolved inside
// sourceDir, and returns its OpenVEX output.
func (*defaultScannerImpl) ScanSource(ctx context.Context, sourceDir, pattern string) ([]byte, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	if err := runGovulncheck(ctx, govulncheckArgs(sourceDir, pattern), stdout, stderr); err != nil {
		return nil, scanError(pattern, err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// runGovulncheck runs the govulncheck command, writing its output to stdout
// and errors to stderr.
func runGovulncheck(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cmd := scan.Command(ctx, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting govulncheck: %w", err)
	}

	return cmd.Wait()
}

// govulncheckArgs returns the command arguments of an OpenVEX source scan.
func govulncheckArgs(sourceDir, pattern string) []string {
	return []string{
		"-format=openvex",
		"-mode=source",
		"-C", sourceDir,
		pattern,
	}
}

// scanError describes a failed scan, including the scanner details written
// to stderr, when there are any.
func scanError(pattern string, err error, stderr string) error {
	if details := strings.TrimSpace(stderr); details != "" {
		return fmt.Errorf("running govulncheck for %q: %w: %s", pattern, err, details)
	}

	return fmt.Errorf("running govulncheck for %q: %w", pattern, err)
}

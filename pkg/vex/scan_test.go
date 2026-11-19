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
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/vex/vexfakes"
)

func TestScanSource(t *testing.T) {
	t.Parallel()

	var gotSourceDir, gotPattern string

	mock := &vexfakes.FakeScannerImplementation{}
	mock.ScanSourceCalls(func(_ context.Context, sourceDir, pattern string) ([]byte, error) {
		gotSourceDir, gotPattern = sourceDir, pattern

		return []byte(`{"@context":"https://openvex.dev/ns/v0.2.0"}`), nil
	})

	scanner := NewScanner()
	scanner.SetImpl(mock)

	result, err := scanner.RunScan(context.Background(), " /src/kubernetes ", " ./cmd/kube-apiserver ")
	require.NoError(t, err)
	require.JSONEq(t, `{"@context":"https://openvex.dev/ns/v0.2.0"}`, string(result))
	require.Equal(t, "/src/kubernetes", gotSourceDir)
	require.Equal(t, "./cmd/kube-apiserver", gotPattern)
	require.Equal(t, 1, mock.ScanSourceCallCount())
}

func TestScanSourceError(t *testing.T) {
	t.Parallel()

	scanErr := errors.New("scan failed")
	mock := &vexfakes.FakeScannerImplementation{}
	mock.ScanSourceReturns(nil, scanErr)

	scanner := NewScanner()
	scanner.SetImpl(mock)

	_, err := scanner.RunScan(context.Background(), "/src/kubernetes", "./cmd/kube-apiserver")
	require.ErrorIs(t, err, scanErr)
}

func TestScanSourceRequiresInputs(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sourceDir string
		pattern   string
		want      string
		calls     int
	}{
		"source directory": {pattern: "./cmd/kube-apiserver", want: "source directory is required"},
		"source pattern":   {sourceDir: "/src/kubernetes", want: "source pattern is required"},
		"no errors":        {sourceDir: "/src/kubernetes", pattern: "./cmd/kube-apiserver", calls: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := &vexfakes.FakeScannerImplementation{}
			scanner := NewScanner()
			scanner.SetImpl(mock)

			_, err := scanner.RunScan(context.Background(), tc.sourceDir, tc.pattern)
			if tc.want == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.want)
			}

			require.Equal(t, tc.calls, mock.ScanSourceCallCount())
		})
	}
}

func TestGovulncheckArgs(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{
		"-format=openvex",
		"-mode=source",
		"-C", "/src/kubernetes",
		"./cmd/kube-apiserver",
	}, govulncheckArgs("/src/kubernetes", "./cmd/kube-apiserver"))
}

func TestScanError(t *testing.T) {
	t.Parallel()

	err := scanError("./cmd/kube-apiserver", errors.New("scan failed"), "scanner details\n")
	require.ErrorContains(t, err, `running govulncheck for "./cmd/kube-apiserver"`)
	require.ErrorContains(t, err, "scan failed")
	require.ErrorContains(t, err, "scanner details")

	err = scanError("./cmd/kube-apiserver", errors.New("scan failed"), "  ")
	require.ErrorContains(t, err, `running govulncheck for "./cmd/kube-apiserver"`)
	require.NotContains(t, err.Error(), "scanner details")
}

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

package anago

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	intoto "github.com/in-toto/attestation/go/v1"
	openvex "github.com/openvex/go-vex/pkg/vex"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/vex"
	"k8s.io/release/pkg/vex/vexfakes"
)

// govulncheckScan is the OpenVEX output of a govulncheck source scan, with
// its placeholder product.
const govulncheckScan = `{
  "@context": "https://openvex.dev/ns/v0.2.0",
  "@id": "govulncheck/vex:old-id",
  "author": "Unknown Author",
  "timestamp": "2026-11-08T12:00:00Z",
  "version": 1,
  "tooling": "https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck",
  "statements": [
    {
      "vulnerability": {"name": "GO-2026-0001"},
      "products": [{
        "@id": "Unknown Product",
        "subcomponents": [{"@id": "pkg:golang/example.com/vulnerable@v1.2.3"}]
      }],
      "status": "affected"
    }
  ]
}`

const testImageDigest = "d1045e5c6d2f016797d22143eba7502e1bb712a4681836a7c35763a9c192dd70"

func TestBinaryVEXDocument(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "kube-apiserver")
	require.NoError(t, os.WriteFile(binaryPath, []byte("binary"), 0o600))

	sum := sha256.Sum256([]byte("binary"))
	digest := hex.EncodeToString(sum[:])

	doc, fileName, err := binaryVEXDocument(
		map[string][]byte{"kube-apiserver": []byte(govulncheckScan)},
		struct{ Path, Platform, Arch string }{binaryPath, "linux", "amd64"},
	)
	require.NoError(t, err)
	require.Equal(t, "vex-kube-apiserver-linux-amd64.json", fileName)

	parsed, err := openvex.Parse(doc)
	require.NoError(t, err)
	require.Equal(t, vex.Author, parsed.Author)
	require.Len(t, parsed.Statements, 1)
	require.Equal(t, map[openvex.Algorithm]openvex.Hash{
		openvex.SHA256: openvex.Hash(digest),
	}, parsed.Statements[0].Products[0].Hashes)
	require.Equal(t,
		"pkg:golang/example.com/vulnerable@v1.2.3",
		parsed.Statements[0].Products[0].Subcomponents[0].ID,
	)
}

func TestBinaryVEXDocumentRequiresScan(t *testing.T) {
	_, _, err := binaryVEXDocument(
		map[string][]byte{},
		struct{ Path, Platform, Arch string }{"/build/kube-apiserver", "linux", "amd64"},
	)
	require.ErrorContains(t, err, "no vulnerability scan of kube-apiserver")
}

func TestWindowsBinaryVEXDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kubectl.exe")
	require.NoError(t, os.WriteFile(path, []byte("windows binary"), 0o600))

	doc, fileName, err := binaryVEXDocument(
		map[string][]byte{"kubectl": []byte(govulncheckScan)},
		struct{ Path, Platform, Arch string }{path, "windows", "amd64"},
	)
	require.NoError(t, err)
	require.Equal(t, "vex-kubectl-windows-amd64.json", fileName)

	parsed, err := openvex.Parse(doc)
	require.NoError(t, err)

	sum := sha256.Sum256([]byte("windows binary"))
	require.Equal(t, openvex.Hash(hex.EncodeToString(sum[:])),
		parsed.Statements[0].Products[0].Hashes[openvex.SHA256])
}

func TestScanBinariesDoesNotPublishPartialScans(t *testing.T) {
	sourceDir := t.TempDir()
	for _, binary := range []string{"kube-apiserver", "kubectl"} {
		require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "cmd", binary), 0o755))
	}

	fake := &vexfakes.FakeScannerImplementation{}
	fake.ScanSourceCalls(func(_ context.Context, _, pattern string) ([]byte, error) {
		if pattern == "./cmd/kubectl" {
			return nil, errors.New("vulnerability database unavailable")
		}

		return []byte(govulncheckScan), nil
	})

	scanner := vex.NewScanner()
	scanner.SetImpl(fake)
	scans, err := scanBinaries(sourceDir, []string{"kube-apiserver", "kubectl"}, scanner)
	require.ErrorContains(t, err, "vulnerability database unavailable")
	require.Empty(t, scans)
	require.Equal(t, 2, fake.ScanSourceCallCount())
}

func TestScanBinariesOmitsCleanScans(t *testing.T) {
	sourceDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "cmd", "kubectl"), 0o755))

	fake := &vexfakes.FakeScannerImplementation{}
	fake.ScanSourceReturns([]byte(`{
		"@context": "https://openvex.dev/ns/v0.2.0",
		"@id": "govulncheck/vex:old-id",
		"author": "Unknown Author",
		"timestamp": "2026-11-08T12:00:00Z",
		"version": 1
	}`), nil)

	scanner := vex.NewScanner()
	scanner.SetImpl(fake)
	scans, err := scanBinaries(sourceDir, []string{"kubectl"}, scanner)
	require.NoError(t, err)
	require.Empty(t, scans)
	require.Equal(t, 1, fake.ScanSourceCallCount())
}

// TODO: do we need to run with the real scan delays? can we run with 100ms delay and normal test instead?
func BenchmarkScanBinariesRunsConcurrently(b *testing.B) {
	sourceDir := b.TempDir()

	binaries := []string{
		"kube-apiserver", "kube-controller-manager", "kube-scheduler",
		"kube-proxy", "kubelet", "kubectl", "kubectl-convert", "kubeadm",
	}
	for _, bin := range binaries {
		require.NoError(b, os.MkdirAll(filepath.Join(sourceDir, "cmd", bin), 0o755))
	}

	// Simulate a 20–30s scan per binary.
	const (
		scanLatency    = 20 * time.Second
		maxScanLatency = 30 * time.Second
	)

	fake := &vexfakes.FakeScannerImplementation{}
	fake.ScanSourceCalls(func(_ context.Context, _, _ string) ([]byte, error) {
		randomDurationUnder10S, err := rand.Int(rand.Reader, big.NewInt(int64(10)))
		require.NoError(b, err, "Randomizing the scan latency should not fail")

		d := scanLatency + time.Second*time.Duration(randomDurationUnder10S.Int64())
		time.Sleep(d)

		return []byte(govulncheckScan), nil
	})

	scanner := vex.NewScanner()
	scanner.SetImpl(fake)

	start := time.Now()
	scans, err := scanBinaries(sourceDir, binaries, scanner)
	elapsed := time.Since(start)

	require.NoError(b, err)
	require.Len(b, scans, len(binaries))

	// 8 binaries × 30s(max) = 4mn sequential.
	// With maxConcurrentScans=3 → ~3 waves ≈ 1mn30s.
	// Reject anything that looks sequential.
	sequential := time.Duration(len(binaries)) * maxScanLatency
	require.Less(b, elapsed, 2*time.Minute,
		"expected concurrent scans to finish faster than %v sequential, took %v", sequential, elapsed)
	b.Logf("Scanned %d binaries in %v (sequential estimate: %v)", len(binaries), elapsed, sequential)
}

func TestVersionGCSStageDir(t *testing.T) {
	version := "v1.36.0"
	require.Equal(t, filepath.Join(gitRoot, release.BuildDir+"-"+version, release.GCSStagePath, version), versionGCSStageDir(version))
}

func TestWriteVEXDocumentsContinuesAfterFailure(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "first.json"), 0o755))

	docs := map[string][]byte{"first.json": []byte("first"), "second.json": []byte("second")}
	err := writeVEXDocuments(docs, dir)
	require.ErrorContains(t, err, "staging first.json")
	second, err := os.ReadFile(filepath.Join(dir, "second.json"))
	require.NoError(t, err)
	require.Equal(t, []byte("second"), second)
}

func TestWriteVEXDocumentsRequiresStagingDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	require.ErrorContains(t, writeVEXDocuments(map[string][]byte{"vex.json": []byte("vex")}, dir), "checking VEX staging directory")
	_, err := os.Stat(dir)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestImageVEXDocument(t *testing.T) {
	for _, tc := range []struct {
		name     string
		subject  *intoto.ResourceDescriptor
		fileName string
	}{
		{
			name: "manifest list image",
			subject: &intoto.ResourceDescriptor{
				Name:   "gcr.io/k8s-staging-kubernetes/kube-apiserver",
				Digest: map[string]string{"sha256": testImageDigest},
			},
			fileName: "image-vex-kube-apiserver.json",
		},
		{
			name: "arch-specific image",
			subject: &intoto.ResourceDescriptor{
				Name:   "gcr.io/k8s-staging-kubernetes/kube-apiserver-amd64",
				Digest: map[string]string{"sha256": testImageDigest},
			},
			fileName: "image-vex-kube-apiserver-amd64.json",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scans := map[string][]byte{"kube-apiserver": []byte(govulncheckScan)}

			doc, fileName, err := imageVEXDocument(scans, tc.subject)
			require.NoError(t, err)
			require.Equal(t, tc.fileName, fileName)

			parsed, err := openvex.Parse(doc)
			require.NoError(t, err)
			require.Equal(t, vex.Author, parsed.Author)
			require.Len(t, parsed.Statements, 1)
			require.Equal(t,
				"pkg:oci/"+filepath.Base(tc.subject.GetName())+"@sha256:"+testImageDigest, parsed.Statements[0].Products[0].ID,
			)
		})
	}
}

func TestImageVEXDocumentRequiresScan(t *testing.T) {
	_, _, err := imageVEXDocument(
		map[string][]byte{},
		&intoto.ResourceDescriptor{
			Name:   "gcr.io/k8s-staging-kubernetes/conformance",
			Digest: map[string]string{"sha256": testImageDigest},
		},
	)
	require.ErrorContains(t, err, "no vulnerability scan of conformance")
}

func TestImageBinaryName(t *testing.T) {
	for _, tc := range []struct {
		imageName string
		want      string
	}{
		{imageName: "kube-apiserver", want: "kube-apiserver"},
		{imageName: "kube-apiserver-amd64", want: "kube-apiserver"},
		{imageName: "kube-apiserver-arm64", want: "kube-apiserver"},
		{imageName: "kube-apiserver-s390x", want: "kube-apiserver"},
		{imageName: "kubectl-convert", want: "kubectl-convert"},
		{imageName: "conformance", want: "conformance"},
	} {
		t.Run(tc.imageName, func(t *testing.T) {
			require.Equal(t, tc.want, imageBinaryName(tc.imageName))
		})
	}
}

func TestFileDigest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kubectl")
	require.NoError(t, os.WriteFile(path, []byte("kubectl"), 0o600))

	sum := sha256.Sum256([]byte("kubectl"))

	digest, err := fileDigest(path)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(sum[:]), digest)
}

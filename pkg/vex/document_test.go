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
	"strings"
	"testing"

	openvex "github.com/openvex/go-vex/pkg/vex"
	"github.com/stretchr/testify/require"
)

const govulncheckDocument = `{
  "@context": "https://openvex.dev/ns/v0.2.0",
  "@id": "govulncheck/vex:old-id",
  "author": "Unknown Author",
  "timestamp": "2026-11-08T12:00:00Z",
  "version": 1,
  "tooling": "https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck",
  "statements": [
    {
      "vulnerability": {
        "@id": "https://pkg.go.dev/vuln/GO-2026-0001",
        "name": "GO-2026-0001",
        "aliases": ["CVE-2026-0001"]
      },
      "products": [{
        "@id": "Unknown Product",
        "subcomponents": [{"@id": "pkg:golang/example.com/vulnerable@v1.2.3"}]
      }],
      "status": "affected"
    },
    {
      "vulnerability": {
        "@id": "https://pkg.go.dev/vuln/GO-2026-0002",
        "name": "GO-2026-0002"
      },
      "products": [{
        "@id": "Unknown Product",
        "subcomponents": [{"@id": "pkg:golang/example.com/unreachable@v2.0.0"}]
      }],
      "status": "not_affected",
      "justification": "vulnerable_code_not_in_execute_path",
      "impact_statement": "Govulncheck determined that the vulnerable code isn't called"
    }
  ]
}`

func TestPatchDocument(t *testing.T) {
	product := openvex.Component{
		Hashes: map[openvex.Algorithm]openvex.Hash{
			openvex.SHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		},
	}

	result, err := PatchDocument([]byte(govulncheckDocument), product)
	require.NoError(t, err)

	doc, err := openvex.Parse(result)
	require.NoError(t, err)
	require.Equal(t, Author, doc.Author)
	require.Equal(t, "https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck", doc.Tooling)
	require.NotEqual(t, "govulncheck/vex:old-id", doc.ID)
	require.True(t, strings.HasPrefix(doc.ID, IDNamespace+"/"),
		"document ID %q must be in the %s namespace", doc.ID, IDNamespace)
	require.Len(t, doc.Statements, 2)

	for _, statement := range doc.Statements {
		require.Len(t, statement.Products, 1)
		require.Equal(t, product, statement.Products[0].Component)
		require.Len(t, statement.Products[0].Subcomponents, 1)
	}

	require.Equal(t, "pkg:golang/example.com/vulnerable@v1.2.3", doc.Statements[0].Products[0].Subcomponents[0].ID)
	require.Equal(t, openvex.StatusAffected, doc.Statements[0].Status)
	require.Equal(t, "pkg:golang/example.com/unreachable@v2.0.0", doc.Statements[1].Products[0].Subcomponents[0].ID)
	require.Equal(t, openvex.StatusNotAffected, doc.Statements[1].Status)
	require.Equal(t, openvex.VulnerableCodeNotInExecutePath, doc.Statements[1].Justification)
}

func TestPatchDocumentNoStatements(t *testing.T) {
	// govulncheck omits the statements key entirely on a clean scan.
	cleanScan := `{
      "@context": "https://openvex.dev/ns/v0.2.0",
      "@id": "govulncheck/vex:old-id",
      "author": "Unknown Author",
      "timestamp": "2026-11-08T12:00:00Z",
      "version": 1,
      "tooling": "https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck"
    }`

	result, err := PatchDocument([]byte(cleanScan), openvex.Component{ID: "pkg:oci/example@sha256:abc"})
	require.NoError(t, err)

	require.Contains(t, string(result), `"statements": []`)

	doc, err := openvex.Parse(result)
	require.NoError(t, err)
	require.NotNil(t, doc.Statements)
	require.Empty(t, doc.Statements)
	require.Equal(t, Author, doc.Author)
	require.True(t, strings.HasPrefix(doc.ID, IDNamespace+"/"),
		"document ID %q must be in the %s namespace", doc.ID, IDNamespace)
}

func TestPatchDocumentRejectsUnexpectedProducts(t *testing.T) {
	tests := map[string]string{
		"no products": `{
          "@context":"https://openvex.dev/ns/v0.2.0",
          "timestamp":"2026-11-08T12:00:00Z",
          "version":1,
          "statements":[{"vulnerability":{"name":"GO-2026-0001"},"products":[],"status":"affected"}]
        }`,
		"multiple products": `{
          "@context":"https://openvex.dev/ns/v0.2.0",
          "timestamp":"2026-11-08T12:00:00Z",
          "version":1,
          "statements":[{"vulnerability":{"name":"GO-2026-0001"},"products":[{"@id":"one"},{"@id":"two"}],"status":"affected"}]
        }`,
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := PatchDocument([]byte(input), openvex.Component{ID: "pkg:oci/example@sha256%3Aabc"})
			require.ErrorContains(t, err, "expected one govulncheck placeholder product")
		})
	}
}

func TestPatchDocumentRequiresProduct(t *testing.T) {
	_, err := PatchDocument([]byte(govulncheckDocument), openvex.Component{})
	require.EqualError(t, err, "product must have an ID, hash, or identifier")
}

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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"sigs.k8s.io/bom/pkg/spdx"
)

// sourceDocument writes a source SBOM like the stage run does and returns its
// path and its package.
func sourceDocument(t *testing.T) (string, *spdx.Package) {
	t.Helper()

	doc := spdx.NewDocument()
	doc.Name = "kubernetes-v1.36.0"
	doc.Namespace = "https://sbom.k8s.io/v1.36.0/source"

	pkg := spdx.NewPackage()
	pkg.Name = "k8s.io/kubernetes"
	pkg.Version = "v1.36.0"
	pkg.BuildID(pkg.Name, pkg.Version)
	pkg.ExternalRefs = []spdx.ExternalRef{{
		Category: "PACKAGE-MANAGER",
		Type:     "purl",
		Locator:  "pkg:golang/k8s.io/kubernetes@v1.36.0",
	}}
	require.NoError(t, doc.AddPackage(pkg))

	path := filepath.Join(t.TempDir(), "source-bom-v1.36.0.spdx")
	require.NoError(t, doc.Write(path))

	return path, pkg
}

func TestSourcePackage(t *testing.T) {
	t.Parallel()

	path, pkg := sourceDocument(t)

	source, err := sourcePackage(path)
	require.NoError(t, err)
	require.Equal(t, pkg.SPDXID(), source.SPDXID())
	require.Equal(t, "k8s.io/kubernetes", source.Name)
	require.NotEqual(t, kubernetesPackageRef, source.SPDXID())

	// A missing source SBOM fails
	_, err = sourcePackage(filepath.Join(t.TempDir(), "missing.spdx"))
	require.Error(t, err)
}

func TestSetSourceReference(t *testing.T) {
	t.Parallel()

	doc := spdx.NewDocument()
	pkg := spdx.NewPackage()
	pkg.Name = "registry.k8s.io/kube-apiserver-amd64:v1.36.0"
	pkg.AddRelationship(&spdx.Relationship{
		PeerReference:    kubernetesPackageRef,
		PeerExtReference: "kubernetes-v1.36.0",
		Type:             spdx.GENERATED_FROM,
	})
	pkg.AddRelationship(&spdx.Relationship{
		PeerReference: "SPDXRef-Package-other",
		Type:          spdx.DEPENDS_ON,
	})
	require.NoError(t, doc.AddPackage(pkg))

	setSourceReference(doc, "kubernetes-v1.36.0", "SPDXRef-Package-k8s.io-kubernetes-v1.36.0")

	rels := *pkg.GetRelationships()
	require.Equal(t, "SPDXRef-Package-k8s.io-kubernetes-v1.36.0", rels[0].PeerReference)
	require.Equal(t, "SPDXRef-Package-other", rels[1].PeerReference)
}

func TestWriteSPDX3(t *testing.T) {
	t.Parallel()

	sourcePath, source := sourceDocument(t)
	extRef := spdx.ExternalDocumentRef{ID: "kubernetes-v1.36.0", URI: "https://sbom.k8s.io/v1.36.0/source"}
	require.NoError(t, extRef.ReadSourceFile(sourcePath))

	doc := spdx.NewDocument()
	doc.Name = "Kubernetes Release v1.36.0"
	doc.Namespace = "https://sbom.k8s.io/v1.36.0/release"

	pkg := spdx.NewPackage()
	pkg.Name = "registry.k8s.io/kube-apiserver-amd64:v1.36.0"
	pkg.AddRelationship(&spdx.Relationship{
		PeerReference:    source.SPDXID(),
		PeerExtReference: "kubernetes-v1.36.0",
		Type:             spdx.GENERATED_FROM,
	})
	require.NoError(t, doc.AddPackage(pkg))

	file := spdx.NewFile()
	file.Name = "bin/linux/amd64/kubectl"
	file.Checksum = map[string]string{"SHA256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}
	file.AddRelationship(&spdx.Relationship{
		PeerReference:    source.SPDXID(),
		PeerExtReference: "kubernetes-v1.36.0",
		Type:             spdx.GENERATED_FROM,
	})
	require.NoError(t, doc.AddFile(file))

	path := filepath.Join(t.TempDir(), "release-bom-v1.36.0.spdx3.json")
	require.NoError(t, writeSPDX3(doc, path, "v1.36.0", extRef, source))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var written struct {
		Context string           `json:"@context"`
		Graph   []map[string]any `json:"@graph"`
	}

	require.NoError(t, json.Unmarshal(data, &written))
	require.Equal(t, "https://spdx.org/rdf/3.0.1/spdx-context.jsonld", written.Context)

	names := map[string][]any{}
	ids := map[string]string{}
	purls := map[string]any{}
	generates := map[string][][]any{}

	for _, element := range written.Graph {
		typ, ok := element["type"].(string)
		if !ok {
			continue
		}

		names[typ] = append(names[typ], element["name"])

		if name, ok := element["name"].(string); ok {
			ids[name], _ = element["spdxId"].(string)
			purls[name] = element["software_packageUrl"]
		}

		if typ == "Relationship" && element["relationshipType"] == "generates" {
			from, _ := element["from"].(string)
			to, ok := element["to"].([]any)
			require.True(t, ok)

			generates[from] = append(generates[from], to)
		}
	}

	// The document has an IRI of its own
	require.Equal(t, []any{"Kubernetes Release v1.36.0"}, names["SpdxDocument"])
	require.Equal(t, "https://sbom.k8s.io/v1.36.0/release.spdx3.json", ids["Kubernetes Release v1.36.0"])
	require.Contains(t, names["software_Package"], "registry.k8s.io/kube-apiserver-amd64:v1.36.0")
	require.Contains(t, names["software_File"], "bin/linux/amd64/kubectl")

	// The source package generates the image and the binary, like in the
	// SPDX 2.3 document, in one relationship, and has the release version
	sourceIRI := ids["k8s.io/kubernetes"]
	require.NotEmpty(t, sourceIRI)
	require.Len(t, generates[sourceIRI], 1)
	require.ElementsMatch(t, []any{
		ids["registry.k8s.io/kube-apiserver-amd64:v1.36.0"],
		ids["bin/linux/amd64/kubectl"],
	}, generates[sourceIRI][0])
	require.Equal(t, "pkg:golang/k8s.io/kubernetes@v1.36.0", purls["k8s.io/kubernetes"])

	// A path that can't be created fails
	require.Error(t, writeSPDX3(doc, filepath.Join(t.TempDir(), "missing", "sbom.spdx3.json"), "v1.36.0", extRef, source))
}

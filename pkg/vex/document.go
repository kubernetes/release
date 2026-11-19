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
	"errors"
	"fmt"

	openvex "github.com/openvex/go-vex/pkg/vex"
)

// Author is the author recorded in Kubernetes release VEX documents.
const Author = "Kubernetes Release Engineering"

// IDNamespace is the namespace of the IDs of Kubernetes release VEX
// documents. The documents of the binaries are published under it, the
// documents of the container images are attached to the images they describe.
const IDNamespace = "https://dl.k8s.io/release/vex"

// PatchDocument replaces govulncheck's placeholder product with one release
// product while preserving the vulnerable Go modules as subcomponents.
func PatchDocument(data []byte, product openvex.Component) ([]byte, error) {
	doc, err := openvex.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing govulncheck OpenVEX: %w", err)
	}

	if !validProduct(product) {
		return nil, errors.New("product must have an ID, hash, or identifier")
	}

	for i := range doc.Statements {
		products := doc.Statements[i].Products
		if len(products) != 1 {
			return nil, fmt.Errorf(
				"statement %d has %d products, expected one govulncheck placeholder product",
				i, len(products),
			)
		}

		doc.Statements[i].Products = []openvex.Product{{
			Component:     product,
			Subcomponents: products[0].Subcomponents,
		}}
	}

	// govulncheck omits the statements key when it finds no vulnerabilities,
	// which parses back as nil and would serialize as a null field.
	if doc.Statements == nil {
		doc.Statements = []openvex.Statement{}
	}

	doc.Author = Author
	doc.ID = ""

	hash, err := doc.CanonicalHash()
	if err != nil {
		return nil, fmt.Errorf("hashing OpenVEX document: %w", err)
	}

	doc.ID = fmt.Sprintf("%s/%s", IDNamespace, hash)

	var result bytes.Buffer
	if err := doc.ToJSON(&result); err != nil {
		return nil, fmt.Errorf("serializing OpenVEX document: %w", err)
	}

	return result.Bytes(), nil
}

func validProduct(product openvex.Component) bool {
	return product.ID != "" || len(product.Hashes) > 0 || len(product.Identifiers) > 0
}

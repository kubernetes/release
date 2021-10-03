/*
Copyright 2021 The Kubernetes Authors.

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

package provenance_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/provenance"
	"k8s.io/release/pkg/provenance/provenancefakes"
)

// getPredicateSUT returns a predicate loaded with the test implementation
func getPredicateSUT() *provenance.Predicate {
	p := provenance.NewSLSAPredicate()
	p.SetImplementation(&provenancefakes.FakePredicateImplementation{})
	return &p
}

func TestWrite(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*provenancefakes.FakePredicateImplementation)
		shouldError bool
	}{
		{
			// Write errors
			prepare: func(mock *provenancefakes.FakePredicateImplementation) {
				mock.WriteReturns(errors.New("Fake error"))
			},
			shouldError: true,
		},
		{
			// Write succeeds
			prepare: func(mock *provenancefakes.FakePredicateImplementation) {
				mock.WriteReturns(nil)
			},
			shouldError: false,
		},
	} {
		p := getPredicateSUT()
		mock := &provenancefakes.FakePredicateImplementation{}
		tc.prepare(mock)
		p.SetImplementation(mock)
		res := p.Write("/tmp/mock")
		if tc.shouldError {
			require.NotNil(t, res)
		} else {
			require.Nil(t, res)
		}
	}
}

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

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/provenance"
	"k8s.io/release/pkg/provenance/provenancefakes"
)

func getStatementSUT() *provenance.Statement {
	sut := provenance.NewSLSAStatement()
	sut.SetImplementation(&provenancefakes.FakeStatementImplementation{})
	return sut
}

func TestReadSubjectsFromDir(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*provenancefakes.FakeStatementImplementation)
		shouldError bool
	}{
		{
			// Read errors
			prepare: func(mock *provenancefakes.FakeStatementImplementation) {
				mock.ReadSubjectsFromDirReturns(errors.New("mock error"))
			},
			shouldError: true,
		},
		{
			// Read succeeds
			prepare: func(mock *provenancefakes.FakeStatementImplementation) {
				mock.ReadSubjectsFromDirReturns(nil)
			},
			shouldError: false,
		},
	} {
		s := getStatementSUT()
		mock := &provenancefakes.FakeStatementImplementation{}
		tc.prepare(mock)
		s.SetImplementation(mock)
		res := s.ReadSubjectsFromDir("/tmp/mock/")
		if tc.shouldError {
			require.NotNil(t, res)
		} else {
			require.Nil(t, res)
		}
	}
}

func TestAddSubjectFromFile(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*provenancefakes.FakeStatementImplementation)
		shouldError bool
	}{
		{
			// Read errors
			prepare: func(mock *provenancefakes.FakeStatementImplementation) {
				mock.SubjectFromFileReturns(intoto.Subject{}, errors.New("mock error"))
			},
			shouldError: true,
		},
		{
			// Read succeeds
			prepare: func(mock *provenancefakes.FakeStatementImplementation) {
				mock.SubjectFromFileReturns(intoto.Subject{}, nil)
			},
			shouldError: false,
		},
	} {
		s := getStatementSUT()
		mock := &provenancefakes.FakeStatementImplementation{}
		tc.prepare(mock)
		s.SetImplementation(mock)
		res := s.AddSubjectFromFile("/tmp/mock/")
		if tc.shouldError {
			require.NotNil(t, res)
		} else {
			require.Nil(t, res)
		}
	}
}

func TestWriteStatement(t *testing.T) {
	for _, tc := range []struct {
		prepare     func(*provenancefakes.FakeStatementImplementation)
		shouldError bool
	}{
		{
			// Write errors
			prepare: func(mock *provenancefakes.FakeStatementImplementation) {
				mock.WriteReturns(errors.New("Fake error"))
			},
			shouldError: true,
		},
		{
			// Write succeeds
			prepare: func(mock *provenancefakes.FakeStatementImplementation) {
				mock.WriteReturns(nil)
			},
			shouldError: false,
		},
	} {
		p := getStatementSUT()
		mock := &provenancefakes.FakeStatementImplementation{}
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

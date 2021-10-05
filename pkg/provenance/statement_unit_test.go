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

package provenance

import (
	"os"
	"path/filepath"
	"testing"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/release-utils/util"
)

func TestReadSubjectsFromDir(t *testing.T) {
	s := NewSLSAStatement()
	testdata := []struct {
		filename string
		content  string
		hash     string
	}{
		{"en.txt", "Hello world", "64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f3c"},
		{"es.txt", "Hola mundo", "ca8f60b2cc7f05837d98b208b57fb6481553fc5f1219d59618fd025002a66f5c"},
		{"es/mx.txt", "Quiobos", "0ff2872124d43e90de9221ec849c94f3c797d8daf9254230055c8ebe41fc8b47"},
		{"de.txt", "Hallo Welt", "2d2da19605a34e037dbe82173f98a992a530a5fdd53dad882f570d4ba204ef30"},
		{"de/ch.txt", "Salü", "d64d0c924abb7b5bbc9352cab90676f69d36170deefc2f224b17fe3de71e6a53"},
	}

	// Create a directory with some files
	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	for _, testfile := range testdata {
		path := filepath.Join(dir, testfile.filename)
		if !util.Exists(filepath.Dir(path)) {
			require.Nil(t, os.Mkdir(filepath.Dir(path), os.FileMode(0o755)))
		}
		require.Nil(t, os.WriteFile(
			path, []byte(testfile.content), os.FileMode(0o644)),
			"writing test file",
		)
	}

	// Read the files as subjects of the predicate
	require.Nil(t, s.ReadSubjectsFromDir(dir), "Reading subjects")
	require.Equal(t, len(testdata), len(s.Subject))

	// Cycle all subjects and check the hashes match
	for _, subject := range s.Subject {
		seen := false
		for _, data := range testdata {
			if data.filename == subject.Name {
				seen = true
				require.Equal(t, data.hash, subject.Digest["sha256"], "invalid subject hash: "+subject.Name)
			}
		}
		require.True(t, seen, "file not found in subjects: "+subject.Name)
	}
}

func TestAddSubject(t *testing.T) {
	s := NewSLSAStatement()
	sha1 := "cd7f2fdcbd859060732c8a9677d9e838babfa6b9"
	s.AddSubject("https://www.example.com/", intoto.DigestSet{"sha1": sha1})
	require.Equal(t, 1, len(s.Subject))
	require.Equal(t, sha1, s.Subject[0].Digest["sha1"])
}

func TestLoadPredicate(t *testing.T) {
	prData := `{"builder":{"id":"Test@1.0"},"metadata":{"buildInvocationId":"CICD1234","completeness":{"arguments":false,"environment":false,"materials":false},"reproducible":false,"BuildStartedOn":null,"buildFinishedOn":null},"recipe":{"type":"","definedInMaterial":0,"entryPoint":"","arguments":null,"environment":null},"materials":[]}`

	file, err := os.CreateTemp("", "predicate")
	require.Nil(t, err)
	defer os.Remove(file.Name())
	require.Nil(t, os.WriteFile(file.Name(), []byte(prData), os.FileMode(0o644)))

	s := NewSLSAStatement()
	require.Nil(t, s.LoadPredicate(file.Name()), "loading predicate from file")

	require.Equal(t, "Test@1.0", s.Predicate.Builder.ID)
}

func TestSubjectFromFile(t *testing.T) {
	// Create a test file
	f, err := os.CreateTemp("", "")
	require.Nil(t, err)
	defer os.Remove(f.Name())
	require.Nil(t, os.WriteFile(f.Name(), []byte("Hello world"), os.FileMode(0o644)))

	// Create a subject from the temporary file
	si := defaultStatementImplementation{}
	subject, err := si.SubjectFromFile(f.Name())
	require.Nil(t, err, "creating subject from file")

	// Check the filename
	require.Equal(t, f.Name(), subject.Name)

	// Verify the hashes match the expected values
	require.Equal(
		t, "64ec88ca00b268e5ba1a35678a1b5316d212f4f366b2477232534a8aeca37f3c",
		subject.Digest["sha256"],
	)
	require.Equal(
		t, "b7f783baed8297f0db917462184ff4f08e69c2d5e5f79a942600f9725f58ce1f29c18139bf80b06c0fff2bdd34738452ecf40c488c22a7e3d80cdf6f9c1c0d47",
		subject.Digest["sha512"],
	)

	// Attempting a subject from a directory must fail
	_, err = si.SubjectFromFile(filepath.Dir(f.Name()))
	require.NotNil(t, err, "should err trying to create a subject from a dir")
}

func TestWriteStatement(t *testing.T) {
	s := NewSLSAStatement()
	s.Predicate.Builder.ID = "asd"
	tmp, err := os.CreateTemp("", "statement-test")
	require.Nil(t, err)
	defer os.Remove(tmp.Name())

	res := s.Write(tmp.Name())
	require.Nil(t, res)
	require.FileExists(t, tmp.Name())
	st, err := os.Stat(tmp.Name())
	require.Nil(t, err)
	require.Greater(t, st.Size(), int64(0))
}

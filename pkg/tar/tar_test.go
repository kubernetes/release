/*
Copyright 2020 The Kubernetes Authors.

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

package tar

import (
	"archive/tar"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompress(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "compress-")
	require.Nil(t, err)
	defer os.RemoveAll(baseTmpDir)

	for _, fileName := range []string{
		"1.txt", "2.bin", "3.md",
	} {
		require.Nil(t, ioutil.WriteFile(
			filepath.Join(baseTmpDir, fileName),
			[]byte{1, 2, 3},
			os.FileMode(0o644),
		))
	}

	subTmpDir := filepath.Join(baseTmpDir, "sub")
	require.Nil(t, os.MkdirAll(subTmpDir, os.FileMode(0o755)))

	for _, fileName := range []string{
		"4.txt", "5.bin", "6.md",
	} {
		require.Nil(t, ioutil.WriteFile(
			filepath.Join(subTmpDir, fileName),
			[]byte{4, 5, 6},
			os.FileMode(0o644),
		))
	}

	excludes := []*regexp.Regexp{
		regexp.MustCompile(".md"),
		regexp.MustCompile("5"),
	}

	tarFilePath := filepath.Join(baseTmpDir, "res.tar.gz")
	require.Nil(t, Compress(tarFilePath, baseTmpDir, excludes...))
	require.FileExists(t, tarFilePath)

	res := []string{"1.txt", "2.bin", "sub/4.txt"}
	require.Nil(t, iterateTarball(
		tarFilePath, func(_ *tar.Reader, header *tar.Header) bool {
			require.Equal(t, res[0], header.Name)
			res = res[1:]
			return false
		}),
	)
}

func TestReadFileFromGzippedTar(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "tar-read-file-")
	require.Nil(t, err)
	defer os.RemoveAll(baseTmpDir)

	const (
		testFilePath     = "test.txt"
		testFileContents = "test-file-contents"
	)
	testTarPath := filepath.Join(baseTmpDir, "test.tar.gz")

	require.Nil(t, ioutil.WriteFile(
		filepath.Join(baseTmpDir, testFilePath),
		[]byte(testFileContents),
		os.FileMode(0o644),
	))
	require.Nil(t, Compress(testTarPath, baseTmpDir, nil))

	type args struct {
		tarPath  string
		filePath string
	}
	type want struct {
		fileContents string
		shouldErr    bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FoundFileInTar": {
			args: args{
				tarPath:  testTarPath,
				filePath: testFilePath,
			},
			want: want{fileContents: testFileContents},
		},
		"FileNotInTar": {
			args: args{
				tarPath:  testTarPath,
				filePath: "badfile.txt",
			},
			want: want{shouldErr: true},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, err := ReadFileFromGzippedTar(tc.args.tarPath, tc.args.filePath)
			if tc.want.shouldErr {
				require.Nil(t, r)
				require.NotNil(t, err)
			} else {
				file, err := ioutil.ReadAll(r)
				require.Nil(t, err)
				require.Equal(t, tc.want.fileContents, string(file))
			}
		})
	}
}

/*
Copyright 2019 The Kubernetes Authors.

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

package util

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMoreRecent(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	// Create test files.
	testFileOne := filepath.Join(baseTmpDir, "testone.txt")
	require.Nil(t, ioutil.WriteFile(
		testFileOne,
		[]byte("file-one-contents"),
		0644,
	))

	time.Sleep(1 * time.Second)

	testFileTwo := filepath.Join(baseTmpDir, "testtwo.txt")
	require.Nil(t, ioutil.WriteFile(
		testFileTwo,
		[]byte("file-two-contents"),
		0644,
	))

	notFile := filepath.Join(baseTmpDir, "noexist.txt")

	defer cleanupTmp(t, baseTmpDir)

	type args struct {
		a string
		b string
	}
	type want struct {
		r   bool
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AIsRecent": {
			args: args{
				a: testFileTwo,
				b: testFileOne,
			},
			want: want{
				r:   true,
				err: nil,
			},
		},
		"AIsNotRecent": {
			args: args{
				a: testFileOne,
				b: testFileTwo,
			},
			want: want{
				r:   false,
				err: nil,
			},
		},
		"ADoesNotExist": {
			args: args{
				a: notFile,
				b: testFileTwo,
			},
			want: want{
				r:   false,
				err: nil,
			},
		},
		"BDoesNotExist": {
			args: args{
				a: testFileOne,
				b: notFile,
			},
			want: want{
				r:   true,
				err: nil,
			},
		},
		"NeitherExists": {
			args: args{
				a: notFile,
				b: notFile,
			},
			want: want{
				r:   false,
				err: errors.New("neither file exists"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			more, err := MoreRecent(tc.args.a, tc.args.b)
			require.Equal(t, tc.want.err, err)
			require.Equal(t, tc.want.r, more)
		})
	}
}

func TestReadFileFromGzippedTar(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	testFilePath := "test.txt"
	testFileContents := "test-file-contents"
	testTarPath := filepath.Join(baseTmpDir, "test.tar.gz")

	var b bytes.Buffer

	// Create a zip archive.
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	require.Nil(t, tw.WriteHeader(&tar.Header{
		Name: testFilePath,
		Size: int64(len(testFileContents)),
	}))
	_, err = tw.Write([]byte(testFileContents))
	require.Nil(t, err)
	require.Nil(t, gz.Close())
	require.Nil(t, tw.Close())
	require.Nil(t, ioutil.WriteFile(
		testTarPath,
		b.Bytes(),
		0644,
	))

	defer cleanupTmp(t, baseTmpDir)

	type args struct {
		tarPath  string
		filePath string
	}
	type want struct {
		fileContents string
		err          error
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
			want: want{
				fileContents: testFileContents,
				err:          nil,
			},
		},
		"FileNotInTar": {
			args: args{
				tarPath:  testTarPath,
				filePath: "badfile.txt",
			},
			want: want{
				err: errors.New("unable to find file in tarball"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, err := ReadFileFromGzippedTar(tc.args.tarPath, tc.args.filePath)
			require.Equal(t, tc.want.err, err)
			if tc.want.err == nil {
				file, err := ioutil.ReadAll(r)
				require.Nil(t, err)
				require.Equal(t, tc.want.fileContents, string(file))
			}
		})
	}
}

func cleanupTmp(t *testing.T, dir string) {
	require.Nil(t, os.RemoveAll(dir))
}

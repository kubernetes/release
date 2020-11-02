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

	"github.com/sirupsen/logrus"
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

	logrus.SetLevel(logrus.DebugLevel)

	require.Nil(t, os.Symlink(
		filepath.Join(baseTmpDir, "1.txt"),
		filepath.Join(subTmpDir, "link"),
	))

	excludes := []*regexp.Regexp{
		regexp.MustCompile(".md"),
		regexp.MustCompile("5"),
	}

	tarFilePath := filepath.Join(baseTmpDir, "res.tar.gz")
	require.Nil(t, Compress(tarFilePath, baseTmpDir, excludes...))
	require.FileExists(t, tarFilePath)

	res := []string{"1.txt", "2.bin", "sub/4.txt", "sub/link"}
	require.Nil(t, iterateTarball(
		tarFilePath, func(_ *tar.Reader, header *tar.Header) (bool, error) {
			require.Equal(t, res[0], header.Name)
			res = res[1:]
			return false, nil
		}),
	)
}

func TestExtract(t *testing.T) {
	tarball := []byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xec, 0xd7,
		0xdf, 0xea, 0x82, 0x30, 0x14, 0xc0, 0xf1, 0xfd, 0xfe, 0xd4, 0x73, 0xf4,
		0x02, 0xb9, 0xb3, 0x9d, 0xb9, 0xd9, 0xe3, 0xa8, 0x08, 0x49, 0x69, 0xe2,
		0x26, 0xf4, 0xf8, 0x31, 0x93, 0x2e, 0xba, 0x10, 0x8a, 0xe6, 0x08, 0xcf,
		0xe7, 0x66, 0x0c, 0xc4, 0x1d, 0x2f, 0xbe, 0x30, 0x45, 0xe2, 0xae, 0x8e,
		0x85, 0x05, 0x00, 0xa0, 0x95, 0xf2, 0xab, 0x30, 0x29, 0x8c, 0x7b, 0x71,
		0xdf, 0x4f, 0x90, 0x09, 0x34, 0x29, 0x00, 0x48, 0xd4, 0x9a, 0x81, 0x90,
		0x0a, 0x91, 0xed, 0x20, 0xf0, 0x5c, 0xa3, 0xc1, 0xba, 0xbc, 0x67, 0x00,
		0x36, 0xb7, 0xe5, 0x31, 0x9f, 0x7b, 0xae, 0xea, 0xed, 0xcc, 0x7b, 0xa6,
		0x2f, 0x79, 0xac, 0x5f, 0xe2, 0xe7, 0xf7, 0x2f, 0xf6, 0x08, 0x24, 0x22,
		0x99, 0x14, 0x75, 0x1b, 0xf8, 0x8c, 0x37, 0xfa, 0x47, 0x9d, 0x52, 0xff,
		0x4b, 0xa0, 0xfe, 0xd7, 0xcd, 0x0e, 0x05, 0x57, 0x81, 0xef, 0x00, 0xaf,
		0xf7, 0x8f, 0x52, 0x1a, 0xea, 0x7f, 0x09, 0xff, 0x9b, 0x6d, 0xec, 0x11,
		0x48, 0x44, 0xbe, 0xff, 0x73, 0xdd, 0x9e, 0x42, 0x9e, 0xe1, 0x7b, 0x30,
		0xc6, 0xcc, 0xf4, 0x0f, 0x4f, 0xfd, 0x1b, 0xed, 0xef, 0xff, 0x92, 0xbb,
		0xa6, 0xe3, 0xe5, 0xa5, 0xe9, 0xfa, 0xca, 0xda, 0xfd, 0x01, 0x33, 0x25,
		0x54, 0x86, 0x8a, 0x7f, 0xf2, 0xa7, 0x65, 0xe5, 0xfd, 0x13, 0x42, 0xd6,
		0xeb, 0x16, 0x00, 0x00, 0xff, 0xff, 0xe9, 0xde, 0xbe, 0xdf, 0x00, 0x12,
		0x00, 0x00,
	}
	file, err := ioutil.TempFile("", "tarball")
	require.Nil(t, err)
	defer os.Remove(file.Name())
	_, err = file.Write(tarball)
	require.Nil(t, err)

	baseTmpDir, err := ioutil.TempDir("", "extract-")
	require.Nil(t, err)
	require.Nil(t, os.RemoveAll(baseTmpDir))
	defer os.RemoveAll(baseTmpDir)

	require.Nil(t, Extract(file.Name(), baseTmpDir))
	res := []string{
		filepath.Base(baseTmpDir),
		"1.txt",
		"2.bin",
		"sub",
		"4.txt",
		"link",
	}
	require.Nil(t, filepath.Walk(
		baseTmpDir,
		func(filePath string, fileInfo os.FileInfo, err error) error {
			require.Equal(t, res[0], fileInfo.Name())
			if res[0] == "link" {
				require.True(t, fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink)
			}
			res = res[1:]
			return nil
		},
	))
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

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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestPackagesAvailableSuccess(t *testing.T) {
	testcases := [][]string{
		{"bash"},
		{"bash", "curl", "grep"},
		{},
	}

	for _, packages := range testcases {
		available, err := PackagesAvailable(packages...)
		require.Nil(t, err)
		require.True(t, available)
	}
}

func TestPackagesAvailableFailure(t *testing.T) {
	testcases := [][]string{
		{
			"fakepackagefoo",
		},
		{
			"fakepackagefoo",
			"fakepackagebar",
			"fakepackagebaz",
		},
		{
			"bash",
			"fakepackagefoo",
			"fakepackagebar",
		},
	}

	for _, packages := range testcases {
		actual, err := PackagesAvailable(packages...)
		require.Nil(t, err)
		require.False(t, actual)
	}
}

func TestMoreRecent(t *testing.T) {
	baseTmpDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	// Create test files.
	testFileOne := filepath.Join(baseTmpDir, "testone.txt")
	require.Nil(t, ioutil.WriteFile(
		testFileOne,
		[]byte("file-one-contents"),
		os.FileMode(0644),
	))

	time.Sleep(1 * time.Second)

	testFileTwo := filepath.Join(baseTmpDir, "testtwo.txt")
	require.Nil(t, ioutil.WriteFile(
		testFileTwo,
		[]byte("file-two-contents"),
		os.FileMode(0644),
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
			require.IsType(t, tc.want.err, err)
			require.Equal(t, tc.want.r, more)
		})
	}
}

func TestCopyFile(t *testing.T) {
	srcDir, err := ioutil.TempDir("", "src")
	require.Nil(t, err)
	dstDir, err := ioutil.TempDir("", "dst")
	require.Nil(t, err)

	// Create test file.
	srcFileOnePath := filepath.Join(srcDir, "testone.txt")
	require.Nil(t, ioutil.WriteFile(
		srcFileOnePath,
		[]byte("file-one-contents"),
		os.FileMode(0644),
	))

	dstFileOnePath := filepath.Join(dstDir, "testone.txt")

	defer cleanupTmp(t, srcDir)
	defer cleanupTmp(t, dstDir)

	type args struct {
		src      string
		dst      string
		required bool
	}
	cases := map[string]struct {
		args        args
		shouldError bool
	}{
		"CopyFileSuccess": {
			args: args{
				src:      srcFileOnePath,
				dst:      dstFileOnePath,
				required: true,
			},
			shouldError: false,
		},
		"CopyFileNotExistNotIgnore": {
			args: args{
				src:      "path/does/not/exit",
				dst:      dstFileOnePath,
				required: true,
			},
			shouldError: true,
		},
		"CopyFileNotExistIgnore": {
			args: args{
				src:      "path/does/not/exit",
				dst:      dstFileOnePath,
				required: false,
			},
			shouldError: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			copyErr := CopyFileLocal(tc.args.src, tc.args.dst, tc.args.required)
			if tc.shouldError {
				require.NotNil(t, copyErr)
			} else {
				require.Nil(t, copyErr)
			}
			if copyErr == nil {
				_, err := os.Stat(filepath.Join(tc.args.dst))
				if err != nil && tc.args.required {
					t.Fatal("file does not exist in destination")
				}
			}
		})
	}
}

func TestCopyDirContentLocal(t *testing.T) {
	srcDir, err := ioutil.TempDir("", "src")
	require.Nil(t, err)
	dstDir, err := ioutil.TempDir("", "dst")
	require.Nil(t, err)

	// Create test file.
	srcFileOnePath := filepath.Join(srcDir, "testone.txt")
	require.Nil(t, ioutil.WriteFile(
		srcFileOnePath,
		[]byte("file-one-contents"),
		os.FileMode(0644),
	))

	srcFileTwoPath := filepath.Join(srcDir, "testtwo.txt")
	require.Nil(t, ioutil.WriteFile(
		srcFileTwoPath,
		[]byte("file-two-contents"),
		os.FileMode(0644),
	))

	defer cleanupTmp(t, srcDir)
	defer cleanupTmp(t, dstDir)

	type args struct {
		src string
		dst string
	}
	type want struct {
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"CopyDirContentsSuccess": {
			args: args{
				src: srcDir,
				dst: dstDir,
			},
			want: want{
				err: nil,
			},
		},
		"CopyDirContentsSuccessDstNotExist": {
			args: args{
				src: srcDir,
				dst: filepath.Join(dstDir, "path/not/exist"),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			copyErr := CopyDirContentsLocal(tc.args.src, tc.args.dst)
			require.Equal(t, tc.want.err, copyErr)
		})
	}
}

func TestRemoveAndReplaceDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "rm")
	require.Nil(t, err)

	// Create test file.
	fileOnePath := filepath.Join(dir, "testone.txt")
	require.Nil(t, ioutil.WriteFile(
		fileOnePath,
		[]byte("file-one-contents"),
		os.FileMode(0644),
	))

	fileTwoPath := filepath.Join(dir, "testtwo.txt")
	require.Nil(t, ioutil.WriteFile(
		fileTwoPath,
		[]byte("file-two-contents"),
		os.FileMode(0644),
	))

	defer cleanupTmp(t, dir)

	type args struct {
		dir string
	}
	type want struct {
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"RemoveAndReplaceSuccess": {
			args: args{
				dir: dir,
			},
			want: want{
				err: nil,
			},
		},
		"RemoveAndReplaceNotExist": {
			args: args{
				dir: filepath.Join(dir, "not/exit"),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := RemoveAndReplaceDir(tc.args.dir)
			require.Equal(t, tc.want.err, err)
		})
	}
}

func TestExist(t *testing.T) {
	dir, err := ioutil.TempDir("", "rm")
	require.Nil(t, err)

	// Create test file.
	fileOnePath := filepath.Join(dir, "testone.txt")
	require.Nil(t, ioutil.WriteFile(
		fileOnePath,
		[]byte("file-one-contents"),
		os.FileMode(0644),
	))

	defer cleanupTmp(t, dir)

	type args struct {
		dir string
	}
	type want struct {
		exist bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"DirExists": {
			args: args{
				dir: dir,
			},
			want: want{
				exist: true,
			},
		},
		"FileExists": {
			args: args{
				dir: fileOnePath,
			},
			want: want{
				exist: true,
			},
		},
		"DirNotExists": {
			args: args{
				dir: filepath.Join(dir, "path/not/exit"),
			},
			want: want{
				exist: false,
			},
		},
		"FileNotExists": {
			args: args{
				dir: filepath.Join(dir, "notexist.txt"),
			},
			want: want{
				exist: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			exist := Exists(tc.args.dir)
			require.Equal(t, tc.want.exist, exist)
		})
	}
}

func cleanupTmp(t *testing.T, dir string) {
	require.Nil(t, os.RemoveAll(dir))
}

func TestTagStringToSemver(t *testing.T) {
	// Success
	version, err := TagStringToSemver("v1.2.3")
	require.Nil(t, err)
	require.Equal(t, version, semver.Version{Major: 1, Minor: 2, Patch: 3})

	// No Major.Minor.Patch elements found
	version, err = TagStringToSemver("invalid")
	require.NotNil(t, err)
	require.Equal(t, version, semver.Version{})

	// Version string empty
	version, err = TagStringToSemver("")
	require.NotNil(t, err)
	require.Equal(t, version, semver.Version{})
}

func TestSemverToTagString(t *testing.T) {
	version := semver.Version{Major: 1, Minor: 2, Patch: 3}
	require.Equal(t, SemverToTagString(version), "v1.2.3")
}

func TestAddTagPrefix(t *testing.T) {
	require.Equal(t, "v0.0.0", AddTagPrefix("0.0.0"))
	require.Equal(t, "v1.0.0", AddTagPrefix("v1.0.0"))
}

func TestTrimTagPrefix(t *testing.T) {
	require.Equal(t, "0.0.0", TrimTagPrefix("0.0.0"))
	require.Equal(t, "1.0.0", TrimTagPrefix("1.0.0"))
}

func TestWrapText(t *testing.T) {
	//nolint
	longText := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Ut molestie accumsan orci, id congue nibh sollicitudin in. Nulla condimentum arcu eu est hendrerit tempus. Nunc risus nibh, aliquam in ultrices fringilla, aliquet ac purus. Aenean non nibh magna. Nunc lacinia suscipit malesuada. Vivamus porta a leo vel ornare. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Morbi pellentesque orci magna, sed semper nulla fringilla at. Nam elementum ipsum maximus lectus tempor faucibus. Donec eu enim nulla. Integer egestas venenatis tristique. Curabitur id purus sem. Vivamus nec mollis lorem.`
	wrappedText := "Lorem ipsum dolor sit amet, consectetur\n"
	wrappedText += "adipiscing elit. Ut molestie accumsan\n"
	wrappedText += "orci, id congue nibh sollicitudin in.\n"
	wrappedText += "Nulla condimentum arcu eu est hendrerit\n"
	wrappedText += "tempus. Nunc risus nibh, aliquam in\n"
	wrappedText += "ultrices fringilla, aliquet ac purus.\n"
	wrappedText += "Aenean non nibh magna. Nunc lacinia\n"
	wrappedText += "suscipit malesuada. Vivamus porta a leo\n"
	wrappedText += "vel ornare. Orci varius natoque\n"
	wrappedText += "penatibus et magnis dis parturient\n"
	wrappedText += "montes, nascetur ridiculus mus. Morbi\n" //nolint
	wrappedText += "pellentesque orci magna, sed semper\n"
	wrappedText += "nulla fringilla at. Nam elementum ipsum\n"
	wrappedText += "maximus lectus tempor faucibus. Donec eu\n"
	wrappedText += "enim nulla. Integer egestas venenatis\n"
	wrappedText += "tristique. Curabitur id purus sem.\n"
	wrappedText += "Vivamus nec mollis lorem."
	require.Equal(t, WrapText(longText, 40), wrappedText)
}

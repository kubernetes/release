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

package binary_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/binary"
	"k8s.io/release/pkg/binary/binaryfakes"
)

type TestHeader struct {
	Bits int
	Arch string
	OS   string
	Data string
}

// GetTestHeaders returns an array of test binary fragments. The base64 encoded data
// corresponds to the first bytes (between 128 and 512) of the kubectl executables
// of the Kubernetes v1.20.2 release. They are note meant to be the full excecutables,
// only the first bytes to test the header analysis functions.
func GetTestHeaders() []TestHeader {
	return []TestHeader{
		{
			Bits: 64,
			Arch: "amd64",
			OS:   "linux",
			Data: "f0VMRgIBAQAAAAAAAAAAAAIAPgABAAAAwPZGAAAAAABAAAAAAAAAAJABAAAAAAAAAAAAAEAAOAAGAEAADQADAAYAAAAEAAAAQAAAAAAAAABAAEAAAAAAAEAAQAAAAAAAUAEAAAAAAABQAQAAAAAAAAAQAAAAAAAAAQAAAAUAAAA=",
		},
		{
			Bits: 64,
			Arch: "s390x",
			OS:   "linux",
			Data: "f0VMRgICAQAAAAAAAAAAAAACABYAAAABAAAAAAAIFwAAAAAAAAAAQAAAAAAAAAGQAAAAAQBAADgABgBAAA0AAwAAAAYAAAAEAAAAAAAAAEAAAAAAAAEAQAAAAAAAAQBAAAAAAA==",
		},
		{
			Bits: 64,
			Arch: "ppc64le",
			OS:   "linux",
			Data: "f0VMRgIBAQAAAAAAAAAAAAIAFQABAAAAAKwHAAAAAABAAAAAAAAAAJABAAAAAAAAAgAAAEAAOAAGAEAADQADAAYAAAAEAAAAQAAAAAAAAABAAAEAAAAAAEAAAQAAAAAAUAEAAA==",
		},
		{
			Bits: 64,
			Arch: "arm64",
			OS:   "linux",
			Data: "f0VMRgIBAQAAAAAAAAAAAAIAtwABAAAAkGQHAAAAAABAAAAAAAAAAJABAAAAAAAAAAAAAEAAOAAGAEAADQADAAYAAAAEAAAAQAAAAAAAAABAAAEAAAAAAEAAAQAAAAAAUAEAAA==",
		},
		{
			Bits: 32,
			Arch: "386",
			OS:   "linux",
			Data: "f0VMRgEBAQAAAAAAAAAAAAIAAwABAAAA8JsKCDQAAAD0AAAAAAAAADQAIAAGACgADQADAAYAAAA0AAAANIAECDSABAjAAAAAwAAAAAQAAAAAEAAAAQAAAAAAAAAAgAQIAIAECA==",
		},
		{
			Bits: 64,
			Arch: "amd64",
			OS:   "darwin",
			Data: "z/rt/gcAAAEDAAAAAgAAAAwAAACABwAAAAAAAAAAAAAZAAAASAAAAF9fUEFHRVpFUk8AAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
		},
		{
			Bits: 32,
			Arch: "386",
			OS:   "windows",
			Data: "TVqQAAMABAAAAAAA//8AAIsAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAA4fug4AtAnNIbgBTM0hVGhpcyBwcm9ncmFtIGNhbm5vdCBiZSBydW4gaW4gRE9TIG1vZGUuDQ0KJAAAAAAAAABQRQAATAEGAAAAAAAAqi8CAAAAAOAAAgMLAQMAAHgJAQAyHQAAAAAAMBYGAAAQAAAAUP8BAABAAAAQAAAAAgAABgABAAEAAAAGAAEAAAAAAAAgMgIABAAAAAAAAAMAQIEAABAAABAAAAAAEAAAEAAAAAAAABAAAAAAAAAAAAAAAA==",
		},
		{
			Bits: 64,
			Arch: "amd64",
			OS:   "windows",
			Data: "TVqQAAMABAAAAAAA//8AAIsAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAA4fug4AtAnNIbgBTM0hVGhpcyBwcm9ncmFtIGNhbm5vdCBiZSBydW4gaW4gRE9TIG1vZGUuDQ0KJAAAAAAAAABQRQAAZIYGAAAAAAAAXHgCAAAAAPAAIgILAgMAAAw9AQBOHgAAAAAAABMHAAAQAAAAAEAAAAAAAAAQAAAAAgAABgABAAEAAAAGAAEAAAAAAAAwfQIABgAAAAAAAAMAYIEAACAAAAAAAAAQAAAAAAAAAAAQAAAAAAAAEAAAAAAAAA==",
		},
	}
}

// writeTestBinary Writes a test binary and returns the path
func writeTestBinary(t *testing.T, base64Data *string) *os.File {
	f, err := os.CreateTemp("", "test-binary-")
	require.Nil(t, err)

	binData, err := base64.StdEncoding.DecodeString(*base64Data)
	require.Nil(t, err)

	_, err = f.Write(binData)
	require.Nil(t, err)

	_, err = f.Seek(0, 0)
	require.Nil(t, err)

	return f
}

func TestOS(t *testing.T) {
	mock := &binaryfakes.FakeBinaryImplementation{}
	mock.OSReturns("darwin")
	sut := &binary.Binary{}
	sut.SetImplementation(mock)

	require.Equal(t, "darwin", sut.OS())
}

func TestArch(t *testing.T) {
	mock := &binaryfakes.FakeBinaryImplementation{}
	mock.ArchReturns("amd64")
	sut := &binary.Binary{}
	sut.SetImplementation(mock)

	require.Equal(t, "amd64", sut.Arch())
}

func TestGetELFHeader(t *testing.T) {
	for _, testBin := range GetTestHeaders() {
		f := writeTestBinary(t, &testBin.Data)
		defer os.Remove(f.Name())
		header, err := binary.GetELFHeader(f.Name())
		require.Nil(t, err)
		if testBin.OS == "linux" {
			require.NotNil(t, header)
			require.Equal(t, testBin.Bits, header.WordLength())
		} else {
			require.Nil(t, header)
		}
	}
}

func TestGetMachOHeader(t *testing.T) {
	for _, testBin := range GetTestHeaders() {
		f := writeTestBinary(t, &testBin.Data)
		defer os.Remove(f.Name())
		header, err := binary.GetMachOHeader(f.Name())
		require.Nil(t, err)
		if testBin.OS == "darwin" {
			require.NotNil(t, header)
			require.Equal(t, testBin.Bits, header.WordLength())
		} else {
			require.Nil(t, header)
		}
	}
}

func TestGetPEHeader(t *testing.T) {
	for _, testBin := range GetTestHeaders() {
		f := writeTestBinary(t, &testBin.Data)
		defer os.Remove(f.Name())
		header, err := binary.GetPEHeader(f.Name())
		require.Nil(t, err)
		if testBin.OS == "windows" {
			require.NotNil(t, header, fmt.Sprintf("testing binary for %s/%s", testBin.OS, testBin.Arch))
			require.Equal(t, testBin.Bits, header.WordLength())
		} else {
			require.Nil(t, header)
		}
	}
}

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

package build

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/release"
)

const (
	testS3Version      = "v1.36.0-alpha.0.12+0123456789abcd"
	olderTestS3Version = "v1.36.0-alpha.0.10+0123456789abcd"
	newerTestS3Version = "v1.37.0-alpha.0.1+0123456789abcd"
	// testS3VersionKey is the key that S3 serves at the URL path testS3Version.
	testS3VersionKey = "v1.36.0-alpha.0.12 0123456789abcd"
)

// fakeS3 stores the objects in memory, by "<bucket>/<key>".
type fakeS3 struct {
	objects        map[string]string
	checkBucketErr error
}

func newFakeS3(objects map[string]string) *fakeS3 {
	if objects == nil {
		objects = map[string]string{}
	}

	return &fakeS3{objects: objects}
}

func (f *fakeS3) CheckBucket(string) error {
	return f.checkBucketErr
}

func (f *fakeS3) PathExists(bucket, keyPrefix string) (bool, error) {
	for name := range f.objects {
		if strings.HasPrefix(name, bucket+"/"+keyPrefix) {
			return true, nil
		}
	}

	return false, nil
}

func (f *fakeS3) ReadObject(bucket, key string) (string, error) {
	content, ok := f.objects[bucket+"/"+key]
	if !ok {
		return "", errors.New("not found")
	}

	return content, nil
}

func (f *fakeS3) UploadDir(dir, bucket, keyPrefix string) error {
	return filepath.WalkDir(dir, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		rel, err := filepath.Rel(dir, file)
		if err != nil {
			return err
		}

		return f.UploadVersionMarker(file, bucket, path.Join(keyPrefix, filepath.ToSlash(rel)))
	})
}

func (f *fakeS3) UploadVersionMarker(file, bucket, key string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	f.objects[bucket+"/"+key] = string(content)

	return nil
}

func newS3TestInstance(t *testing.T, s3 *fakeS3, opts *Options) *Instance {
	t.Helper()

	opts.Bucket = "s3://bucket/prefix"
	opts.BucketRoot = "kubernetes"
	opts.BuildDir = t.TempDir()
	opts.BuildType = "ci"

	stageDir := filepath.Join(opts.BuildDir, release.GCSStagePath, testS3Version)
	require.NoError(t, os.MkdirAll(filepath.Join(stageDir, "bin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(stageDir, release.KubernetesTar), []byte("tar"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(stageDir, "bin", "kubectl"), []byte("kubectl"), 0o600))

	return &Instance{opts: opts, s3: s3}
}

func TestPushToS3(t *testing.T) {
	t.Parallel()

	const markerPrefix = "bucket/prefix/kubernetes/"

	s3 := newFakeS3(map[string]string{
		markerPrefix + "latest.txt":   olderTestS3Version,
		markerPrefix + "latest-1.txt": newerTestS3Version,
	})
	sut := newS3TestInstance(t, s3, &Options{
		CI:                  true,
		ExtraVersionMarkers: []string{"extra"},
	})

	require.NoError(t, sut.pushToS3(testS3Version))

	require.Equal(t, map[string]string{
		markerPrefix + testS3VersionKey + "/" + release.KubernetesTar: "tar",
		markerPrefix + testS3VersionKey + "/bin/kubectl":              "kubectl",
		markerPrefix + "latest.txt":                                   testS3Version,
		markerPrefix + "latest-1.txt":                                 newerTestS3Version,
		markerPrefix + "latest-1.36.txt":                              testS3Version,
		markerPrefix + "extra.txt":                                    testS3Version,
	}, s3.objects)
}

func TestPushToS3WithoutVersionMarkers(t *testing.T) {
	t.Parallel()

	for name, opts := range map[string]*Options{
		"not CI":           {},
		"no latest update": {CI: true, NoUpdateLatest: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s3 := newFakeS3(nil)
			sut := newS3TestInstance(t, s3, opts)

			require.NoError(t, sut.pushToS3(testS3Version))
			require.Len(t, s3.objects, 2)
		})
	}
}

func TestPushToS3InvalidPublishedVersion(t *testing.T) {
	t.Parallel()

	s3 := newFakeS3(map[string]string{"bucket/prefix/kubernetes/latest.txt": "wrong"})
	sut := newS3TestInstance(t, s3, &Options{CI: true})

	require.ErrorContains(t, sut.pushToS3(testS3Version), "invalid version")
}

func TestS3BuildPathsExist(t *testing.T) {
	t.Parallel()

	s3 := newFakeS3(nil)
	sut := newS3TestInstance(t, s3, &Options{})

	found, existErrors := sut.s3BuildPathsExist(testS3Version)
	require.Equal(t, 0, found)
	require.Len(t, existErrors, 3)

	require.NoError(t, sut.pushToS3(testS3Version))

	found, existErrors = sut.s3BuildPathsExist(testS3Version)
	require.Equal(t, 3, found)
	require.Empty(t, existErrors)
}

func TestS3Key(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		bucket   string
		fast     bool
		expected string
	}{
		{bucket: "s3://bucket", expected: "ci/" + testS3VersionKey + "/kubernetes.tar.gz"},
		{bucket: "s3://bucket/a/b/", expected: "a/b/ci/" + testS3VersionKey + "/kubernetes.tar.gz"},
		{bucket: "s3://bucket", fast: true, expected: "ci/fast/" + testS3VersionKey + "/kubernetes.tar.gz"},
	} {
		sut := &Instance{opts: &Options{Bucket: tc.bucket, BucketRoot: "ci", Fast: tc.fast}}
		require.Equal(t, tc.expected, sut.s3Key(testS3Version, release.KubernetesTar))
	}
}

func TestCheckReleaseBucketS3(t *testing.T) {
	t.Parallel()

	s3 := newFakeS3(nil)
	sut := &Instance{opts: &Options{Bucket: "s3://bucket"}, s3: s3}
	require.NoError(t, sut.CheckReleaseBucket())

	s3.checkBucketErr = errors.New("forbidden")

	require.ErrorContains(t, sut.CheckReleaseBucket(), "forbidden")
}

func TestCheckS3OptionsRejectsPrivateBucket(t *testing.T) {
	t.Parallel()

	sut := &Instance{opts: &Options{Bucket: "s3://bucket", PrivateBucket: true}}
	require.ErrorContains(t, sut.checkS3Options(), "private buckets are not supported")
	require.ErrorContains(t, sut.Push(), "private buckets are not supported")
}

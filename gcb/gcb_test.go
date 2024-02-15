/*
Copyright 2024 The Kubernetes Authors.

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

package gcb

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/release/gcb/gcbfakes"
	"k8s.io/release/pkg/gcp/build"
)

func TestDirForJobType(t *testing.T) {
	t.Parallel()

	const testDir = "testDir"
	errTest := errors.New("test")

	for _, tc := range []struct {
		name, jobType string
		prepare       func(*gcbfakes.FakeImpl)
		assert        func(string, error)
	}{
		{
			name:    "success stage",
			jobType: JobTypeStage,
			prepare: func(mock *gcbfakes.FakeImpl) {
				mock.MkdirTempReturns(testDir, nil)
				mock.WriteFileCalls(func(name string, content []byte, mode fs.FileMode) error {
					assert.Contains(t, string(content), "- STAGE")
					assert.Equal(t, filepath.Join(testDir, build.DefaultCloudbuildFile), name)
					return nil
				})
			},
			assert: func(dir string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, testDir, dir)
			},
		},
		{
			name:    "success release",
			jobType: JobTypeRelease,
			prepare: func(mock *gcbfakes.FakeImpl) {
				mock.MkdirTempReturns(testDir, nil)
				mock.WriteFileCalls(func(name string, content []byte, mode fs.FileMode) error {
					assert.Contains(t, string(content), "- RELEASE")
					assert.Equal(t, filepath.Join(testDir, build.DefaultCloudbuildFile), name)
					return nil
				})
			},
			assert: func(dir string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, testDir, dir)
			},
		},
		{
			name:    "success fast forward",
			jobType: JobTypeFastForward,
			prepare: func(mock *gcbfakes.FakeImpl) {
				mock.MkdirTempReturns(testDir, nil)
				mock.WriteFileCalls(func(name string, content []byte, mode fs.FileMode) error {
					assert.Contains(t, string(content), "- FAST_FORWARD")
					assert.Equal(t, filepath.Join(testDir, build.DefaultCloudbuildFile), name)
					return nil
				})
			},
			assert: func(dir string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, testDir, dir)
			},
		},
		{
			name:    "success obs stage",
			jobType: JobTypeObsStage,
			prepare: func(mock *gcbfakes.FakeImpl) {
				mock.MkdirTempReturns(testDir, nil)
				mock.WriteFileCalls(func(name string, content []byte, mode fs.FileMode) error {
					assert.Contains(t, string(content), "- OBS_STAGE")
					assert.Equal(t, filepath.Join(testDir, build.DefaultCloudbuildFile), name)
					return nil
				})
			},
			assert: func(dir string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, testDir, dir)
			},
		},
		{
			name:    "success obs release",
			jobType: JobTypeObsRelease,
			prepare: func(mock *gcbfakes.FakeImpl) {
				mock.MkdirTempReturns(testDir, nil)
				mock.WriteFileCalls(func(name string, content []byte, mode fs.FileMode) error {
					assert.Contains(t, string(content), "- OBS_RELEASE")
					assert.Equal(t, filepath.Join(testDir, build.DefaultCloudbuildFile), name)
					return nil
				})
			},
			assert: func(dir string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, testDir, dir)
			},
		},
		{
			name:    "failure on temp dir creation",
			jobType: JobTypeStage,
			prepare: func(mock *gcbfakes.FakeImpl) {
				mock.MkdirTempReturns("", errTest)
			},
			assert: func(dir string, err error) {
				assert.Error(t, err)
				assert.Empty(t, dir)
			},
		},
		{
			name:    "failure on file write",
			jobType: JobTypeStage,
			prepare: func(mock *gcbfakes.FakeImpl) {
				mock.MkdirTempReturns(testDir, nil)
				mock.WriteFileReturns(errTest)
			},
			assert: func(dir string, err error) {
				assert.Error(t, err)
				assert.Empty(t, dir)
			},
		},
		{
			name:    "failure unknown job type",
			jobType: "wrong",
			prepare: func(mock *gcbfakes.FakeImpl) {
				mock.MkdirTempReturns(testDir, nil)
			},
			assert: func(dir string, err error) {
				assert.Error(t, err)
				assert.Empty(t, dir)
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mock := &gcbfakes.FakeImpl{}
			sut := New()
			sut.impl = mock
			tc.prepare(mock)

			dir, err := sut.DirForJobType(tc.jobType)
			tc.assert(dir, err)
		})
	}
}

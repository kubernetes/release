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

package release_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
)

func newJobCacheClientSut() (
	*release.JobCacheClient, *releasefakes.FakeGcpClient,
) {
	sut := release.NewJobCacheClient()
	client := &releasefakes.FakeGcpClient{}
	sut.SetClient(client)
	return sut, client
}

func newJobCacheClientSutMocked(t *testing.T, json string) (
	sut *release.JobCacheClient, cleanup func(),
) {
	sut, client := newJobCacheClientSut()
	f, err := ioutil.TempFile("", "job-cache-test-")
	require.Nil(t, err)
	_, err = f.WriteString(json)
	require.Nil(t, err)
	client.CopyJobCacheReturns(f.Name(), nil)
	return sut, func() { require.Nil(t, os.RemoveAll(f.Name())) }
}

func TestGetJobCacheSuccessSkip(t *testing.T) {
	sut, _ := newJobCacheClientSut()
	res, err := sut.GetJobCache("", true)
	require.Nil(t, err)
	require.Nil(t, res)
}

func TestGetJobCacheSuccess(t *testing.T) {
	sut, cleanup := newJobCacheClientSutMocked(t, testJobCache)
	defer cleanup()

	res, err := sut.GetJobCache("", false)
	require.Nil(t, err)
	require.Len(t, res, 294)
	require.Equal(t,
		"v1.18.3-beta.0.41+ec280c2f9e131c", res["1259303474592485376"],
	)
}

func TestGetJobCacheSuccessDedup(t *testing.T) {
	sut, cleanup := newJobCacheClientSutMocked(t, testJobCache)
	defer cleanup()

	res, err := sut.GetJobCache("", true)
	require.Nil(t, err)
	require.Len(t, res, 9)
	require.Equal(t,
		"v1.18.3-beta.0.58+d6e40f410ca91c", res["1263096730283413504"],
	)
}

func TestGetJobCacheFailure(t *testing.T) {
	sut, client := newJobCacheClientSut()
	client.CopyJobCacheReturns("", errors.New(""))
	res, err := sut.GetJobCache("", true)
	require.NotNil(t, err)
	require.Nil(t, res)
}

func TestGetJobCacheFailureWrongJson(t *testing.T) {
	sut, cleanup := newJobCacheClientSutMocked(t, "wrong")
	defer cleanup()

	res, err := sut.GetJobCache("", false)
	require.NotNil(t, err)
	require.Nil(t, res)
}

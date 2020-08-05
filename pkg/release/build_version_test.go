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
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/git"
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
	client.CopyJobCacheReturns(f.Name())
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
	require.Len(t, res.Versions, 294)
	require.Equal(t,
		"v1.18.4-rc.0.3+3ff09514d162b0", res.Versions[0],
	)
}

func TestGetJobCacheSuccessDedup(t *testing.T) {
	sut, cleanup := newJobCacheClientSutMocked(t, testJobCache)
	defer cleanup()

	res, err := sut.GetJobCache("", true)
	require.Nil(t, err)
	require.Len(t, res.Versions, 9)
	require.Equal(t,
		"v1.18.4-rc.0.3+3ff09514d162b0", res.Versions[0],
	)
}

func TestGetJobCacheFailureWrongJson(t *testing.T) {
	sut, cleanup := newJobCacheClientSutMocked(t, "wrong")
	defer cleanup()

	res, err := sut.GetJobCache("", false)
	require.NotNil(t, err)
	require.Nil(t, res)
}

func newBuildVersionClientSUT() (
	*release.BuildVersionClient,
	*releasefakes.FakeGithubClient,
	*releasefakes.FakeJobCacheClient,
	*releasefakes.FakeTestGridClient,
) {
	sut := release.NewBuildVersionClient()

	githubClient := &releasefakes.FakeGithubClient{}
	sut.SetGithubClient(githubClient)

	jobCacheClient := &releasefakes.FakeJobCacheClient{}
	sut.SetJobCacheClient(jobCacheClient)

	testGridClient := &releasefakes.FakeTestGridClient{}
	sut.SetTestGridClient(testGridClient)

	return sut, githubClient, jobCacheClient, testGridClient
}

func TestSetBuildVersionSuccess(t *testing.T) {
	const testVersion = "v1.18.4-rc.0.3+3ff09514d162b0"
	sut, _, jcMock, tgMock := newBuildVersionClientSUT()
	jcMock.GetJobCacheReturnsOnCall(0, &release.JobCache{
		Name:         "1",
		BuildNumbers: []string{"10"},
		Versions:     []string{testVersion},
	}, nil)
	jcMock.GetJobCacheReturnsOnCall(1, &release.JobCache{
		Name:         "2",
		BuildNumbers: []string{"10"},
		Versions:     []string{testVersion},
	}, nil)
	tgMock.BlockingTestsReturns([]string{"1", "2", "3"}, nil)

	version, err := sut.SetBuildVersion(git.DefaultBranch, []string{"3"})
	require.Nil(t, err)
	require.Equal(t, version, testVersion)
}

func TestSetBuildVersionFailureSecondBuildVersionNotAvailable(t *testing.T) {
	const testVersion = "v1.18.4-rc.0.3+3ff09514d162b0"
	sut, _, jcMock, tgMock := newBuildVersionClientSUT()
	jcMock.GetJobCacheReturnsOnCall(0, &release.JobCache{
		Name:         "1",
		BuildNumbers: []string{"10"},
		Versions:     []string{testVersion},
	}, nil)
	jcMock.GetJobCacheReturnsOnCall(1, &release.JobCache{
		Name:         "2",
		BuildNumbers: []string{"10"},
		Versions:     []string{},
	}, nil)
	tgMock.BlockingTestsReturns([]string{"1", "2", "3", "4"}, nil)

	version, err := sut.SetBuildVersion(git.DefaultBranch, []string{"4"})
	require.NotNil(t, err)
	require.Empty(t, version)
}

func TestSetBuildVersionFailureGetCommitDate(t *testing.T) {
	const testVersion = "v1.18.4-rc.0.3+3ff09514d162b0"
	sut, hcMock, jcMock, tgMock := newBuildVersionClientSUT()
	jcMock.GetJobCacheReturns(&release.JobCache{
		Name:         "kubernetes-ci-build",
		BuildNumbers: []string{"10", "11"},
		Versions:     []string{testVersion, testVersion},
	}, nil)
	tgMock.BlockingTestsReturns([]string{"1", "2", "3"}, nil)
	hcMock.GetCommitDateReturns(time.Time{}, errors.New(""))

	version, err := sut.SetBuildVersion(git.DefaultBranch, []string{"3"})
	require.NotNil(t, err)
	require.Empty(t, version)
}

func TestSetBuildVersionFailureMainVersionWrong(t *testing.T) {
	const testVersion = "wrong"
	sut, _, jcMock, tgMock := newBuildVersionClientSUT()
	jcMock.GetJobCacheReturns(&release.JobCache{
		Name:         "kubernetes-ci-build",
		BuildNumbers: []string{"10", "11"},
		Versions:     []string{testVersion},
	}, nil)
	tgMock.BlockingTestsReturns([]string{"1", "2", "3"}, nil)

	version, err := sut.SetBuildVersion(git.DefaultBranch, nil)
	require.NotNil(t, err)
	require.Empty(t, version)
}

func TestSetBuildVersionFailureNotFound(t *testing.T) {
	sut, _, jcMock, tgMock := newBuildVersionClientSUT()
	jcMock.GetJobCacheReturns(&release.JobCache{}, nil)
	tgMock.BlockingTestsReturns([]string{"1", "2", "3"}, nil)

	version, err := sut.SetBuildVersion(git.DefaultBranch, nil)
	require.NotNil(t, err)
	require.Empty(t, version)
}

func TestSetBuildVersionFailureExcludeRegexInvalid(t *testing.T) {
	sut, _, jcMock, tgMock := newBuildVersionClientSUT()
	jcMock.GetJobCacheReturns(&release.JobCache{}, nil)
	tgMock.BlockingTestsReturns([]string{"1", "2", "3"}, nil)

	version, err := sut.SetBuildVersion(git.DefaultBranch, []string{")"})
	require.NotNil(t, err)
	require.Empty(t, version)
}

func TestSetBuildVersionFailureBlockingTestsFailed(t *testing.T) {
	sut, _, _, tgMock := newBuildVersionClientSUT()
	tgMock.BlockingTestsReturns(nil, errors.New(""))

	version, err := sut.SetBuildVersion(git.DefaultBranch, nil)
	require.NotNil(t, err)
	require.Empty(t, version)
}

func TestSetBuildVersionFailureBlockingTestsEmpty(t *testing.T) {
	sut, _, _, tgMock := newBuildVersionClientSUT()
	tgMock.BlockingTestsReturns([]string{}, nil)

	version, err := sut.SetBuildVersion(git.DefaultBranch, nil)
	require.NotNil(t, err)
	require.Empty(t, version)
}

func TestSetBuildVersionFailureFirstGetJobCacheFails(t *testing.T) {
	sut, _, jcMock, tgMock := newBuildVersionClientSUT()
	tgMock.BlockingTestsReturns([]string{"1", "2", "3"}, nil)
	jcMock.GetJobCacheReturns(nil, errors.New(""))

	version, err := sut.SetBuildVersion(git.DefaultBranch, nil)
	require.NotNil(t, err)
	require.Empty(t, version)
}

func TestSetBuildVersionFailureSecondGetJobCacheFails(t *testing.T) {
	sut, _, jcMock, tgMock := newBuildVersionClientSUT()
	tgMock.BlockingTestsReturns([]string{"1", "2", "3"}, nil)
	jcMock.GetJobCacheReturnsOnCall(0, &release.JobCache{}, nil)
	jcMock.GetJobCacheReturnsOnCall(1, nil, errors.New(""))

	version, err := sut.SetBuildVersion(git.DefaultBranch, nil)
	require.NotNil(t, err)
	require.Empty(t, version)
}

func TestSetBuildVersionFailureMainJobCacheNil(t *testing.T) {
	sut, _, _, tgMock := newBuildVersionClientSUT()
	tgMock.BlockingTestsReturns([]string{"1", "2", "3"}, nil)
	version, err := sut.SetBuildVersion(git.DefaultBranch, nil)
	require.NotNil(t, err)
	require.Empty(t, version)
}

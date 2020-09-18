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

package testgrid_test

import (
	"errors"
	"testing"

	pb "github.com/GoogleCloudPlatform/testgrid/pb/config"
	"github.com/golang/protobuf/proto" // nolint: staticcheck
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/git"
	"k8s.io/release/pkg/testgrid"
	"k8s.io/release/pkg/testgrid/testgridfakes"
)

func newSut() (*testgrid.TestGrid, *testgridfakes.FakeClient) {
	client := &testgridfakes.FakeClient{}
	sut := testgrid.New()
	sut.SetClient(client)
	return sut, client
}

func TestBlockingTestsSuccess(t *testing.T) {
	// Given
	sut, client := newSut()
	httpRes, err := proto.Marshal(&pb.Configuration{
		Dashboards: []*pb.Dashboard{{
			Name: "sig-master-blocking",
			DashboardTab: []*pb.DashboardTab{
				{TestGroupName: "first"},
				{TestGroupName: "second"},
			},
		}},
	})
	require.Nil(t, err)
	client.GetURLResponseReturns(string(httpRes), nil)

	// When
	res, err := sut.BlockingTests(git.DefaultBranch)

	// Then
	require.Nil(t, err)
	require.Len(t, res, 2)
	require.Equal(t, "first", res[0])
	require.Equal(t, "second", res[1])
}

func TestBlockingTestsFailureDashboardNotFound(t *testing.T) {
	// Given
	sut, client := newSut()
	client.GetURLResponseReturns("", nil)

	// When
	res, err := sut.BlockingTests("")

	// Then
	require.NotNil(t, err)
	require.Nil(t, res)
}

func TestBlockingTestsFailureHTTP(t *testing.T) {
	// Given
	sut, client := newSut()
	client.GetURLResponseReturns("", errors.New(""))

	// When
	res, err := sut.BlockingTests("")

	// Then
	require.NotNil(t, err)
	require.Nil(t, res)
}

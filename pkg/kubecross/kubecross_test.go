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

package kubecross

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/kubecross/kubecrossfakes"
)

func TestLatest(t *testing.T) {
	for _, tc := range []struct {
		prepare func(*kubecrossfakes.FakeImpl)
		expect  func(res string, err error)
	}{
		{ // success
			prepare: func(mock *kubecrossfakes.FakeImpl) {
				mock.GetURLResponseReturns("test", nil)
			},
			expect: func(res string, err error) {
				require.Nil(t, err)
				require.Equal(t, "test", res)
			},
		},
		{ // failure GetURLResponse
			prepare: func(mock *kubecrossfakes.FakeImpl) {
				mock.GetURLResponseReturns("", errors.New(""))
			},
			expect: func(res string, err error) {
				require.NotNil(t, err)
				require.Empty(t, res)
			},
		},
	} {
		mock := &kubecrossfakes.FakeImpl{}
		tc.prepare(mock)

		kc := New()
		kc.impl = mock

		res, err := kc.Latest()
		tc.expect(res, err)
	}
}

func TestForBranch(t *testing.T) {
	for _, tc := range []struct {
		prepare func(*kubecrossfakes.FakeImpl)
		expect  func(res string, err error)
	}{
		{ // success
			prepare: func(mock *kubecrossfakes.FakeImpl) {
				mock.GetURLResponseReturns("test", nil)
			},
			expect: func(res string, err error) {
				require.Nil(t, err)
				require.Equal(t, "test", res)
			},
		},
		{ // failure GetURLResponse
			prepare: func(mock *kubecrossfakes.FakeImpl) {
				mock.GetURLResponseReturns("", errors.New(""))
			},
			expect: func(res string, err error) {
				require.NotNil(t, err)
				require.Empty(t, res)
			},
		},
	} {
		mock := &kubecrossfakes.FakeImpl{}
		tc.prepare(mock)

		kc := New()
		kc.impl = mock

		res, err := kc.ForBranch("")
		tc.expect(res, err)
	}
}

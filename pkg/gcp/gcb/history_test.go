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

package gcb_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/api/cloudbuild/v1"

	"k8s.io/release/pkg/gcp/gcb"
	"k8s.io/release/pkg/gcp/gcb/gcbfakes"
)

func TestHistoryRun(t *testing.T) {
	err := errors.New("error")
	for _, tc := range []struct {
		options   *gcb.HistoryOptions
		prepare   func(*gcbfakes.FakeHistoryImpl)
		shouldErr bool
	}{
		{ // success only date from
			options: &gcb.HistoryOptions{
				DateFrom: "2020-11-11",
			},
			prepare:   func(*gcbfakes.FakeHistoryImpl) {},
			shouldErr: false,
		},
		{ // success date from and to
			options: &gcb.HistoryOptions{
				DateFrom: "2020-11-11",
				DateTo:   "2020-11-12",
			},
			prepare: func(mock *gcbfakes.FakeHistoryImpl) {
				mock.GetJobsByTagReturns([]*cloudbuild.Build{
					{
						Tags: []string{"STAGE"},
						Timing: map[string]cloudbuild.TimeSpan{
							"BUILD": {
								StartTime: "2020-10-10",
								EndTime:   "2020-10-10",
							},
						},
					},
				}, nil)
			},
			shouldErr: false,
		},
		{ // failure no from date
			options:   &gcb.HistoryOptions{},
			prepare:   func(*gcbfakes.FakeHistoryImpl) {},
			shouldErr: true,
		},
		{ // failure parse from date
			options: &gcb.HistoryOptions{
				DateFrom: "wrong",
			},
			prepare: func(mock *gcbfakes.FakeHistoryImpl) {
				mock.ParseTimeReturnsOnCall(0, time.Time{}, err)
			},
			shouldErr: true,
		},
		{ // failure parse to date
			options: &gcb.HistoryOptions{
				DateFrom: "2020-11-11",
				DateTo:   "wrong",
			},
			prepare: func(mock *gcbfakes.FakeHistoryImpl) {
				mock.ParseTimeReturnsOnCall(1, time.Time{}, err)
			},
			shouldErr: true,
		},
		{ // failure GetJobsByTag
			options: &gcb.HistoryOptions{
				DateFrom: "2020-11-11",
			},
			prepare: func(mock *gcbfakes.FakeHistoryImpl) {
				mock.GetJobsByTagReturns(nil, err)
			},
			shouldErr: true,
		},
		{ // failure parse start date
			options: &gcb.HistoryOptions{
				DateFrom: "2020-11-11",
			},
			prepare: func(mock *gcbfakes.FakeHistoryImpl) {
				mock.GetJobsByTagReturns([]*cloudbuild.Build{
					{
						Tags: []string{"RELEASE"},
						Timing: map[string]cloudbuild.TimeSpan{
							"BUILD": {StartTime: "wrong"},
						},
					},
				}, nil)
				mock.ParseTimeReturnsOnCall(1, time.Time{}, err)
			},
			shouldErr: true,
		},
		{ // failure parse end date
			options: &gcb.HistoryOptions{
				DateFrom: "2020-11-11",
			},
			prepare: func(mock *gcbfakes.FakeHistoryImpl) {
				mock.GetJobsByTagReturns([]*cloudbuild.Build{
					{
						Tags: []string{"STAGE"},
						Timing: map[string]cloudbuild.TimeSpan{
							"BUILD": {EndTime: "wrong"},
						},
						Substitutions: map[string]string{"_NOMOCK": "true"},
					},
				}, nil)
				mock.ParseTimeReturnsOnCall(2, time.Time{}, err)
			},
			shouldErr: true,
		},
	} {
		sut := gcb.NewHistory(tc.options)
		mock := &gcbfakes.FakeHistoryImpl{}
		tc.prepare(mock)
		sut.SetImpl(mock)

		err := sut.Run()
		if tc.shouldErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

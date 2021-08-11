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

package internal_test

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/github/internal"
)

func TestMain(m *testing.M) {
	// logrus, shut up
	logrus.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func TestGithubRetryer(t *testing.T) {
	tests := map[string]struct {
		maxTries        int
		sleeper         func(time.Duration)
		errs            []error
		expectedResults []bool
	}{
		"never retry": {
			maxTries: 0,
		},
		"when error is nil, don't retry": {
			maxTries:        1,
			sleeper:         nilSleeper,
			errs:            []error{nil},
			expectedResults: []bool{false},
		},
		"when error is a random error, don't retry": {
			maxTries:        1,
			sleeper:         nilSleeper,
			errs:            []error{fmt.Errorf("some randm error")},
			expectedResults: []bool{false},
		},
		"when the error is a github abuse rate limit error, retry": {
			maxTries:        1,
			sleeper:         nilSleeper,
			errs:            []error{&github.AbuseRateLimitError{}},
			expectedResults: []bool{true},
		},
		"when the error is a github abuse rate limit error but max tries have been reached, don't retry": {
			maxTries: 2,
			sleeper:  nilSleeper,
			errs: []error{
				&github.AbuseRateLimitError{},
				&github.AbuseRateLimitError{},
				&github.AbuseRateLimitError{},
			},
			expectedResults: []bool{
				true, true, false,
			},
		},
		"when no RetryAfter is specified on the abuse rate limit error, sleep the default amount of time": {
			maxTries:        1,
			sleeper:         sleepChecker(t, 1*time.Minute),
			errs:            []error{&github.AbuseRateLimitError{}},
			expectedResults: []bool{true},
		},
		"when a RetryAfter is specified on the abuse rate limit error, sleep that amount of time": {
			maxTries:        1,
			sleeper:         sleepChecker(t, 42*time.Minute),
			errs:            []error{&github.AbuseRateLimitError{RetryAfter: durPtr(42 * time.Minute)}},
			expectedResults: []bool{true},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			shouldRetry := internal.GithubErrChecker(tc.maxTries, nilSleeper)

			for i, err := range tc.errs {
				if a, e := shouldRetry(err), tc.expectedResults[i]; e != a {
					t.Errorf("Expected to get %t, got: %t", e, a)
				}
			}
		})
	}
}

func sleepChecker(t *testing.T, expectedSleep time.Duration) func(time.Duration) {
	return func(d time.Duration) {
		if d != expectedSleep {
			t.Errorf("Expected the sleeper to be called with a duration %s, got called with %s", expectedSleep, d)
		}
	}
}

func nilSleeper(_ time.Duration) {
}

func durPtr(d time.Duration) *time.Duration {
	return &d
}

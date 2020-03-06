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

package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	khttp "k8s.io/release/pkg/http"
)

func TestGetURLResponseSuccess(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := io.WriteString(w, "")
			require.Nil(t, err)
		}))
	defer server.Close()

	// When
	actual, err := khttp.GetURLResponse(server.URL, false)

	// Then
	require.Nil(t, err)
	require.Empty(t, actual)
}

func TestGetURLResponseSuccessTrimmed(t *testing.T) {
	// Given
	const expected = "     some test     "
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := io.WriteString(w, expected)
			require.Nil(t, err)
		}))
	defer server.Close()

	// When
	actual, err := khttp.GetURLResponse(server.URL, true)

	// Then
	require.Nil(t, err)
	require.Equal(t, strings.TrimSpace(expected), actual)
}

func TestGetURLResponseFailedStatus(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
	defer server.Close()

	// When
	_, err := khttp.GetURLResponse(server.URL, true)

	// Then
	require.NotNil(t, err)
}

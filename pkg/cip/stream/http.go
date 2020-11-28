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

package stream

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// HTTP is a wrapper around the net/http's Request type.
type HTTP struct {
	Req *http.Request
	Res *http.Response
}

const (
	requestTimeoutSeconds = 3
)

// Produce runs the external process and returns two io.Readers (to stdout and
// stderr). In this case we equate the http.Respose "Body" with stdout.
func (h *HTTP) Produce() (stdOut, stdErr io.Reader, err error) {
	client := http.Client{
		Timeout: time.Second * requestTimeoutSeconds,
	}

	// We close the response body in Close().
	// nolint[bodyclose]
	h.Res, err = client.Do(h.Req)

	if err != nil {
		return nil, nil, err
	}

	if h.Res.StatusCode == http.StatusOK {
		return h.Res.Body, nil, nil
	}

	// Try to glean some additional information by reading from the response
	// body.
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(h.Res.Body)
	if err != nil {
		logrus.Errorf("could not read from HTTP response body")
		return nil, nil, fmt.Errorf(
			"problems encountered: unexpected response code %d",
			h.Res.StatusCode,
		)
	}

	return nil, nil, fmt.Errorf(
		"problems encountered: unexpected response code %d; body: %s",
		h.Res.StatusCode,
		buf.String(),
	)
}

// Close closes the http request. This is required because otherwise there will
// be a resource leak.
// See https://stackoverflow.com/a/33238755/437583
func (h *HTTP) Close() error {
	return h.Res.Body.Close()
}

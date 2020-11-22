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
	"io"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// Producer is an interface for anything that can generate an io.Reader from
// which we can read from (typically JSON output).
type Producer interface {
	// The first two io.Readers are expected to be the stdout and stderr streams
	// from the process, respectively.
	Produce() (io.Reader, io.Reader, error)
	Close() error
}

// Consumer is really only defined for symmetry with "Producer"; nothing
// actually uses it.
type Consumer interface {
	Consume(io.Reader) error
}

// An ExternalRequest is anything that can create and then consume any stream.
// The request comes bundled with something that can produce a stream
// (io.Reader), and something that can read from that stream to populate some
// arbitrary data structure.
type ExternalRequest struct {
	RequestParams  interface{}
	StreamProducer Producer
}

// BackoffDefault is the default Backoff behavior for network call retries.
//
// nolint[gomnd]
var BackoffDefault = wait.Backoff{
	Duration: time.Second,
	Factor:   2,
	Jitter:   0.1,
	Steps:    45,
	Cap:      time.Second * 60,
}

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

package cli

import "fmt"

var (
	// GitDescribe is stamped by bazel.
	GitDescribe string

	// GitCommit is stamped by bazel.
	GitCommit string

	// TimestampUtcRfc3339 is stamped by bazel.
	TimestampUtcRfc3339 string
)

type RootOptions struct {
	LogLevel string
	Version  bool
}

// nolint: deadcode,unused
func printVersion() {
	fmt.Printf("Built:   %s\n", TimestampUtcRfc3339)
	fmt.Printf("Version: %s\n", GitDescribe)
	fmt.Printf("Commit:  %s\n", GitCommit)
}

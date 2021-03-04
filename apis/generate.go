// +build generate

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

// See the below link for details on what is happening here.
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

//go:generate go run -tags generate k8s.io/code-generator/cmd/deepcopy-gen -i ./krel/v1alpha1 -o "." -O zz_generated.deepcopy --go-header-file ../hack/boilerplate/boilerplate.go.txt

package apis

import (
	_ "k8s.io/code-generator" //nolint:typecheck
)

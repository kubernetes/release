/*
Copyright 2022 The Kubernetes Authors.

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

package packages

import (
	"os"

	"sigs.k8s.io/release-sdk/object"
	"sigs.k8s.io/release-utils/command"
)

type defaultImpl struct{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . impl
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt packagesfakes/fake_impl.go > packagesfakes/_fake_impl.go && mv packagesfakes/_fake_impl.go packagesfakes/fake_impl.go"
type impl interface {
	Getwd() (string, error)
	Chdir(string) error
	RunCommand(string, ...string) error
	NormalizePath(*object.GCS, ...string) (string, error)
	CopyToRemote(*object.GCS, string, string) error
}

func (*defaultImpl) Getwd() (dir string, err error) {
	return os.Getwd()
}

func (*defaultImpl) Chdir(dir string) error {
	return os.Chdir(dir)
}

func (*defaultImpl) RunCommand(cmd string, args ...string) error {
	return command.New(cmd, args...).RunSuccess()
}

func (*defaultImpl) NormalizePath(store *object.GCS, parts ...string) (string, error) {
	return store.NormalizePath(parts...)
}

func (*defaultImpl) CopyToRemote(store *object.GCS, src, dst string) error {
	return store.CopyToRemote(src, dst)
}

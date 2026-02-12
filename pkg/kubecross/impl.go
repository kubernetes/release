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
	"bytes"

	"k8s.io/release/pkg/consts"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . impl
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt kubecrossfakes/fake_impl.go > kubecrossfakes/_fake_impl.go && mv kubecrossfakes/_fake_impl.go kubecrossfakes/fake_impl.go"
type impl interface {
	GetURLResponse(url string) (string, error)
}

type defaultImpl struct{}

func (*defaultImpl) GetURLResponse(url string) (string, error) {
	content, err := consts.NewHTTPAgent().Get(url)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(content)), nil
}

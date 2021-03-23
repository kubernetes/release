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
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/git"
)

// KubeCross is the main structure of this package.
type KubeCross struct {
	impl impl
}

// New creates a new KubeCross instance.
func New() *KubeCross {
	return &KubeCross{&defaultImpl{}}
}

// Latest returns the latest available kubecross version.
func (k *KubeCross) Latest() (string, error) {
	return k.ForBranch(git.DefaultBranch)
}

// ForBranch returns the kubecross version for the provided branch.
func (k *KubeCross) ForBranch(branch string) (string, error) {
	logrus.Infof("Trying to retrieve kube-cross version for branch %s", branch)

	const (
		baseURL     = "https://raw.githubusercontent.com/kubernetes/kubernetes"
		versionPath = "build/build-image/cross/VERSION"
	)

	url := fmt.Sprintf("%s/%s/%s", baseURL, branch, versionPath)
	version, err := k.impl.GetURLResponse(url, true)
	if err != nil {
		return "", errors.Wrap(err, "get URL response")
	}

	logrus.Infof("Retrieved kube-cross version: %s", version)
	return version, nil
}

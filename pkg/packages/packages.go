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
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/release-sdk/object"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/release/pkg/release"
)

type Packages struct {
	impl    impl
	version string
}

func New(version string) *Packages {
	return &Packages{
		impl:    &defaultImpl{},
		version: version,
	}
}

// SetImpl can be used to set the internal implementation, which is mainly used
// for testing.
func (p *Packages) SetImpl(impl impl) {
	p.impl = impl
}

const (
	// gitRoot is the root directory of the k/release repository.
	// We only require this for the script Invocation and should be removed
	// together with rapture.
	gitRootPath = "/workspace/go/src/k8s.io/release"

	// scriptPath is the relative path from the gitRoo to the rapture script.
	scriptPath = "./hack/rapture/build-packages.sh"

	// debPath is the path for the deb package results.
	debPath = gitRootPath + "/packages/deb/bin"

	// rpmPath is the path for the rpm package results.
	rpmPath = gitRootPath + "/packages/rpm/output"
)

// Build creates the packages.
func (p *Packages) Build() error {
	logrus.Infof("Building packages will be done in `Release`, " +
		"because rapture relies on available remote resources ")
	return nil
}

// Release pushes the packages to their final destination.
func (p *Packages) Release() error {
	logrus.Infof("Building and releasing packages for %s", p.version)
	wd, err := p.impl.Getwd()
	if err != nil {
		return fmt.Errorf("get current working dir: %w", err)
	}
	defer func() {
		if err := p.impl.Chdir(wd); err != nil {
			logrus.Errorf("Unable to restore current working directory: %v", err)
		}
	}()
	if err := p.impl.Chdir(gitRootPath); err != nil {
		return fmt.Errorf("switch to k/release git root: %w", err)
	}

	semverVersion, err := p.impl.TagStringToSemver(p.version)
	if err != nil {
		return fmt.Errorf(" parse semver version %s: %w", p.version, err)
	}

	// We dropped support for the arm architecture in Kubernetes v1.27:
	// https://github.com/kubernetes/kubernetes/pull/115742
	//
	// Means we have to add this workaround for currently supported release branches.
	// TODO(saschagrunert): Remove when v1.26 goes end of life (planned 2024-02-28)
	architectures := sets.New("amd64", "arm", "arm64", "ppc64le", "s390x")
	const droppedArmInMinor = 27
	if semverVersion.Minor >= droppedArmInMinor {
		logrus.Info("Removing arm architecture from default set")
		architectures.Delete("arm")
	}
	archList := architectures.UnsortedList()
	sort.Strings(archList)

	if err := p.impl.RunCommand(
		scriptPath, semverVersion.String(), strings.Join(archList, ","),
	); err != nil {
		return fmt.Errorf("run rapture: %w", err)
	}

	store := object.NewGCS()
	store.SetOptions(store.WithNoClobber(false))
	gcsRootPath, err := p.impl.NormalizePath(
		store, release.ProductionBucket, "release", p.version,
	)
	if err != nil {
		return fmt.Errorf("normalize GCS path: %w", err)
	}
	if err := p.impl.CopyToRemote(store, debPath, gcsRootPath+"/deb"); err != nil {
		return fmt.Errorf("copy deb packages to GCS: %w", err)
	}
	if err := p.impl.CopyToRemote(store, rpmPath, gcsRootPath+"/rpm"); err != nil {
		return fmt.Errorf("copy rpm packages to GCS: %w", err)
	}

	logrus.Infof("Done building and releasing packages to %q and %q", debPath, rpmPath)
	return nil
}

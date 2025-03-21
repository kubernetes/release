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

package release

import (
	"errors"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-sdk/git"
)

type BranchChecker struct {
	impl branchCheckerImpl
}

// NewBranchChecker creates a new release branch checker instance.
func NewBranchChecker() *BranchChecker {
	return &BranchChecker{&defaultBranchCheckerImpl{}}
}

// SetImpl can be used to set the internal BranchChecker implementation.
func (r *BranchChecker) SetImpl(impl branchCheckerImpl) {
	r.impl = impl
}

//counterfeiter:generate . branchCheckerImpl
type branchCheckerImpl interface {
	LSRemoteExec(repoURL string, args ...string) (string, error)
}

type defaultBranchCheckerImpl struct{}

func (*defaultBranchCheckerImpl) LSRemoteExec(
	repoURL string, args ...string,
) (string, error) {
	return git.LSRemoteExec(repoURL, args...)
}

// NeedsCreation returns if the provided release branch has to be created and
// checks if it's correct.
func (r *BranchChecker) NeedsCreation(
	branch, releaseType string, buildVersion semver.Version,
) (createReleaseBranch bool, err error) {
	logrus.Infof("Checking if branch %s exists on remote", branch)

	output, err := r.impl.LSRemoteExec(
		git.GetRepoURL(GetK8sOrg(), GetK8sRepo(), false),
		"refs/heads/"+branch,
	)
	if err != nil {
		return false, fmt.Errorf(
			"get remote commit for %s branch: %w",
			branch,
			err,
		)
	}

	if output != "" {
		logrus.Infof("Branch %s does already exist on remote location", branch)
	} else {
		logrus.Infof("Branch %s does not yet exist on remote location", branch)

		if releaseType == ReleaseTypeOfficial {
			return false, errors.New("can't do officials releases when creating a new branch")
		}

		createReleaseBranch = true
	}

	logrus.Infof("Release branch needs to be created: %v", createReleaseBranch)

	if branch == git.DefaultBranch {
		return createReleaseBranch, nil
	}

	// Verify the required release branch
	requiredReleaseBranch := fmt.Sprintf(
		"release-%d.%d", buildVersion.Major, buildVersion.Minor,
	)
	if branch != requiredReleaseBranch {
		return false, fmt.Errorf(
			"branch and build version does not match, got: %v, required: %v",
			branch,
			requiredReleaseBranch,
		)
	}

	return createReleaseBranch, nil
}

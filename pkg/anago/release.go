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

package anago

// releaseClient is a client for release a previously staged release.
//counterfeiter:generate . releaseClient
type releaseClient interface {
	// CheckPrerequisites verifies that a valid GITHUB_TOKEN environment
	// variable is set. It also checks for the existence and version of
	// required packages and if the correct Google Cloud project is set. A
	// basic hardware check will ensure that enough disk space is available,
	// too.
	CheckPrerequisites() error

	// SetBuildCandidate discovers the release branch, parent branch (if
	// available) and build version for this release.
	SetBuildCandidate() error

	// PrepareWorkspace verifies that the working directory is in the desired
	// state. This means that the staged sources will be downloaded from the
	// bucket which should contain a copy of the repository.
	PrepareWorkspace() error

	// PushArtifacts pushes the generated artifacts to the release bucket and
	// Google Container Registry.
	PushArtifacts() error

	// PushGitObjects pushes the new tags and branches to the repository remote
	// on GitHub.
	PushGitObjects() error

	// CreateAnnouncement creates the release announcement mail and update the
	// GitHub release page to contain the artifacts and their checksums.
	CreateAnnouncement() error

	// Archive copies the release process logs to a bucket and sets private
	// permissions on it.
	Archive() error
}

// DefaultRelease is the default staging implementation used in production.
type DefaultRelease struct {
	impl releaseImpl
}

// NewDefaultRelease creates a new defaultRelease instance.
func NewDefaultRelease() *DefaultRelease {
	return &DefaultRelease{&defaultReleaseImpl{}}
}

// SetClient can be used to set the internal release implementation.
func (d *DefaultRelease) SetClient(impl releaseImpl) {
	d.impl = impl
}

// defaultReleaseImpl is the default internal release client implementation.
type defaultReleaseImpl struct{}

// releaseImpl is the implementation of the release client.
//counterfeiter:generate . releaseImpl
type releaseImpl interface{}

func (d *DefaultRelease) CheckPrerequisites() error { return nil }

func (d *DefaultRelease) SetBuildCandidate() error { return nil }

func (d *DefaultRelease) PrepareWorkspace() error { return nil }

func (d *DefaultRelease) PushArtifacts() error { return nil }

func (d *DefaultRelease) PushGitObjects() error { return nil }

func (d *DefaultRelease) CreateAnnouncement() error { return nil }

func (d *DefaultRelease) Archive() error { return nil }

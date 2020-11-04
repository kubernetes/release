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

// stageClient is a client for staging releases.
//counterfeiter:generate . stageClient
type stageClient interface {
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
	// state. This means that the build directory is cleaned up and the checked
	// out repository is in a clean state.
	PrepareWorkspace() error

	// Build runs 'make cross-in-a-container' by using the latest kubecross
	// container image. This step also build all necessary release tarballs.
	Build() error

	// Generate release notes: Generate the CHANGELOG-x.y.md file and commit it
	// into the local working repository.
	GenerateReleaseNotes() error

	// StageArtifacts copies the build artifacts to a Google Cloud Bucket.
	StageArtifacts() error
}

// DefaultStage is the default staging implementation used in production.
type DefaultStage struct {
	impl stageImpl
}

// NewDefaultStage creates a new defaultStage instance.
func NewDefaultStage() *DefaultStage {
	return &DefaultStage{&defaultStageImpl{}}
}

// SetClient can be used to set the internal stage implementation.
func (d *DefaultStage) SetClient(impl stageImpl) {
	d.impl = impl
}

// defaultStageImpl is the default internal stage client implementation.
type defaultStageImpl struct{}

// stageImpl is the implementation of the stage client.
//counterfeiter:generate . stageImpl
type stageImpl interface{}

func (d *DefaultStage) CheckPrerequisites() error { return nil }

func (d *DefaultStage) SetBuildCandidate() error { return nil }

func (d *DefaultStage) PrepareWorkspace() error { return nil }

func (d *DefaultStage) Build() error { return nil }

func (d *DefaultStage) GenerateReleaseNotes() error { return nil }

func (d *DefaultStage) StageArtifacts() error { return nil }

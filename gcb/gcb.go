/*
Copyright 2024 The Kubernetes Authors.

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

package gcb

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"k8s.io/release/pkg/gcp/build"
)

// All available job types
const (
	JobTypeStage       = "stage"
	JobTypeRelease     = "release"
	JobTypeFastForward = "fast-forward"
	JobTypeObsStage    = "obs-stage"
	JobTypeObsRelease  = "obs-release"
)

var (
	//go:embed stage/cloudbuild.yaml
	stageCloudBuild []byte

	//go:embed release/cloudbuild.yaml
	releaseCloudBuild []byte

	//go:embed fast-forward/cloudbuild.yaml
	fastForwardCloudBuild []byte

	//go:embed obs-stage/cloudbuild.yaml
	obsStageCloudBuild []byte

	//go:embed obs-release/cloudbuild.yaml
	obsReleaseCloudBuild []byte
)

// CloudBuild is the main type of this package.
type CloudBuild struct {
	impl
}

// New creates a new CloudBuild instance.
func New() *CloudBuild {
	return &CloudBuild{impl: &defaultImpl{}}
}

// DirForJobType creates a temp directory containing the default cloudbuild
// file for the provided job type.
func (c *CloudBuild) DirForJobType(jobType string) (string, error) {
	tempDir, err := c.impl.MkdirTemp("", "krel-cloudbuild-*")
	if err != nil {
		return "", fmt.Errorf("create temp cloudbuild dir: %w", err)
	}

	var content []byte
	switch jobType {
	case JobTypeStage:
		content = stageCloudBuild
	case JobTypeRelease:
		content = releaseCloudBuild
	case JobTypeFastForward:
		content = fastForwardCloudBuild
	case JobTypeObsStage:
		content = obsStageCloudBuild
	case JobTypeObsRelease:
		content = obsReleaseCloudBuild
	default:
		return "", fmt.Errorf("unknown job type: %s", jobType)
	}

	if err := c.impl.WriteFile(
		filepath.Join(tempDir, build.DefaultCloudbuildFile),
		content,
		0o600,
	); err != nil {
		return "", fmt.Errorf("write cloudbuild file: %w", err)
	}

	return tempDir, nil
}

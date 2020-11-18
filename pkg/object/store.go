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

package object

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Store
type Store interface {
	// TODO: Implement interface
	// TODO: Consider a method to set options
	// TODO: Choose GCS-agnostic names
	CopyToGCS(src, gcsPath string) error
	CopyToLocal(gcsPath, dst string) error
	CopyBucketToBucket(src, dst string) error
	GetReleasePath(bucket, gcsRoot, version string, fast bool) (string, error)
	GetMarkerPath(bucket, gcsRoot string) (string, error)
	NormalizeGCSPath(gcsPathParts ...string) (string, error)
	IsPathNormalized(gcsPath string) bool
	RsyncRecursive(src, dst string) error
	PathExists(gcsPath string) (bool, error)
}

type DefaultStore struct {
	// TODO: Implement store
	opts *Options
}

func New(opts *Options) *DefaultStore {
	return &DefaultStore{opts}
}

// Options are the main options to pass to `Store`.
type Options struct {
	// TODO: Populate fields
}

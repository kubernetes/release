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

package gcs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/gcp"
	"k8s.io/utils/pointer"
)

var (
	// GcsPrefix url prefix for google cloud storage buckets
	GcsPrefix      = "gs://"
	concurrentFlag = "-m"
	recursiveFlag  = "-r"
	noClobberFlag  = "-n"
)

type Options struct {
	// gsutil options
	Concurrent *bool
	Recursive  *bool
	NoClobber  *bool

	// local options
	// AllowMissing allows a copy operation to be skipped if the source or
	// destination does not exist. This is useful for scenarios where copy
	// operations happen in a loop/channel, so a single "failure" does not block
	// the entire operation.
	AllowMissing *bool
}

// DefaultGCSCopyOptions have the default options for the GCS copy action
var DefaultGCSCopyOptions = &Options{
	Concurrent:   pointer.BoolPtr(true),
	Recursive:    pointer.BoolPtr(true),
	NoClobber:    pointer.BoolPtr(true),
	AllowMissing: pointer.BoolPtr(true),
}

// CopyToGCS copies a local directory to the specified GCS path
// TODO: Consider using IsPathNormalized here
func CopyToGCS(src, gcsPath string, opts *Options) error {
	logrus.Infof("Copying %s to GCS (%s)", src, gcsPath)
	gcsPath, gcsPathErr := NormalizeGCSPath(gcsPath)
	if gcsPathErr != nil {
		return errors.Wrap(gcsPathErr, "normalize GCS path")
	}

	_, err := os.Stat(src)
	if err != nil {
		logrus.Info("Unable to get local source directory info")

		if *opts.AllowMissing {
			logrus.Infof("Source directory (%s) does not exist. Skipping GCS upload.", src)
			return nil
		}

		return errors.New("source directory does not exist")
	}

	return bucketCopy(src, gcsPath, opts)
}

// CopyToLocal copies a GCS path to the specified local directory
// TODO: Consider using IsPathNormalized here
func CopyToLocal(gcsPath, dst string, opts *Options) error {
	logrus.Infof("Copying GCS (%s) to %s", gcsPath, dst)
	gcsPath, gcsPathErr := NormalizeGCSPath(gcsPath)
	if gcsPathErr != nil {
		return errors.Wrap(gcsPathErr, "normalize GCS path")
	}

	return bucketCopy(gcsPath, dst, opts)
}

// CopyBucketToBucket copies between two GCS paths.
// TODO: Consider using IsPathNormalized here
func CopyBucketToBucket(src, dst string, opts *Options) error {
	logrus.Infof("Copying %s to %s", src, dst)

	src, srcErr := NormalizeGCSPath(src)
	if srcErr != nil {
		return errors.Wrap(srcErr, "normalize GCS path")
	}

	dst, dstErr := NormalizeGCSPath(dst)
	if dstErr != nil {
		return errors.Wrap(dstErr, "normalize GCS path")
	}

	return bucketCopy(src, dst, opts)
}

func bucketCopy(src, dst string, opts *Options) error {
	args := []string{}

	if *opts.Concurrent {
		logrus.Debug("Setting GCS copy to run concurrently")
		args = append(args, concurrentFlag)
	}

	args = append(args, "cp")
	if *opts.Recursive {
		logrus.Debug("Setting GCS copy to run recursively")
		args = append(args, recursiveFlag)
	}
	if *opts.NoClobber {
		logrus.Debug("Setting GCS copy to not clobber existing files")
		args = append(args, noClobberFlag)
	}

	args = append(args, src, dst)

	if err := gcp.GSUtil(args...); err != nil {
		return errors.Wrap(err, "gcs copy")
	}

	return nil
}

// GetReleasePath returns a GCS path to retrieve builds from or push builds to
//
// Expected destination format:
//   gs://<bucket>/<buildType>[-<gcsSuffix>][/fast][/<version>]
func GetReleasePath(
	bucket, buildType, gcsSuffix, version string,
	fast bool) (string, error) {
	gcsPath, err := getPath(
		bucket,
		buildType,
		gcsSuffix,
		version,
		"release",
		fast,
	)
	if err != nil {
		return "", errors.Wrap(err, "normalize GCS path")
	}

	logrus.Infof("Release path is %s", gcsPath)
	return gcsPath, nil
}

// GetMarkerPath returns a GCS path where version markers should be stored
//
// Expected destination format:
//   gs://<bucket>/<buildType>[-<gcsSuffix>]
func GetMarkerPath(
	bucket, buildType, gcsSuffix string) (string, error) {
	gcsPath, err := getPath(
		bucket,
		buildType,
		gcsSuffix,
		"",
		"marker",
		false,
	)
	if err != nil {
		return "", errors.Wrap(err, "normalize GCS path")
	}

	logrus.Infof("Version marker path is %s", gcsPath)
	return gcsPath, nil
}

// GetReleasePath returns a GCS path to retrieve builds from or push builds to
//
// Expected destination format:
//   gs://<bucket>/<buildType>[-<gcsSuffix>][/fast][/<version>]
// TODO: Support "release" buildType
func getPath(
	bucket, buildType, gcsSuffix, version, pathType string,
	fast bool) (string, error) {
	gcsPathParts := []string{}

	gcsPathParts = append(gcsPathParts, bucket)

	releaseRoot := buildType
	if gcsSuffix != "" {
		releaseRoot += "-" + gcsSuffix
	}

	gcsPathParts = append(gcsPathParts, releaseRoot)

	if pathType == "release" {
		if fast {
			gcsPathParts = append(gcsPathParts, "fast")
		}

		if version != "" {
			gcsPathParts = append(gcsPathParts, version)
		}
	}

	// Ensure any constructed GCS path is prefixed with `gs://`
	return NormalizeGCSPath(gcsPathParts...)
}

// NormalizeGCSPath takes a GCS path and ensures that the `GcsPrefix` is
// prepended to it.
// TODO: Should there be an append function for paths to prevent multiple calls
//       like in build.checkBuildExists()?
func NormalizeGCSPath(gcsPathParts ...string) (string, error) {
	gcsPath := ""

	// Ensure there is at least one element in the gcsPathParts slice before
	// trying to construct a path
	if len(gcsPathParts) == 0 {
		return "", errors.New("must contain at least one path part")
	} else if len(gcsPathParts) == 1 {
		if gcsPathParts[0] == "" {
			return "", errors.New("path should not be an empty string")
		}

		gcsPath = gcsPathParts[0]
	} else {
		var emptyParts int

		for i, part := range gcsPathParts {
			if part == "" {
				emptyParts++
			}

			if i == 0 {
				continue
			}

			if strings.Contains(part, "gs:/") {
				return "", errors.New("one of the GCS path parts contained a `gs:/`, which may suggest a filepath.Join() error in the caller")
			}

			if i == len(gcsPathParts)-1 && emptyParts == len(gcsPathParts) {
				return "", errors.New("all paths provided were empty")
			}
		}

		gcsPath = filepath.Join(gcsPathParts...)
	}

	// Strip `gs://` if it was included in gcsPathParts
	gcsPath = strings.TrimPrefix(gcsPath, GcsPrefix)

	// Strip `gs:/` if:
	// - `gs://` was included in gcsPathParts
	// - gcsPathParts had more than element
	// - filepath.Join() was called somewhere in a caller's logic
	gcsPath = strings.TrimPrefix(gcsPath, "gs:/")

	// Strip `/`
	// This scenario may never happen, but let's catch it, just in case
	gcsPath = strings.TrimPrefix(gcsPath, "/")

	gcsPath = GcsPrefix + gcsPath

	isNormalized := IsPathNormalized(gcsPath)
	if !isNormalized {
		return gcsPath, errors.New("unknown error while trying to normalize GCS path")
	}

	return gcsPath, nil
}

// IsPathNormalized determines if a GCS path is prefixed with `gs://`.
// Use this function as pre-check for any gsutil/GCS functions that manipulate
// GCS bucket contents.
func IsPathNormalized(gcsPath string) bool {
	var errCount int

	if !strings.HasPrefix(gcsPath, GcsPrefix) {
		logrus.Errorf("GCS path (%s) should be prefixed with `gs://`", gcsPath)
		errCount++
	}

	strippedPath := strings.TrimPrefix(gcsPath, GcsPrefix)
	if strings.Contains(strippedPath, "gs:/") {
		logrus.Errorf("GCS path (%s) should be prefixed with `gs:/`", gcsPath)
		errCount++
	}

	// TODO: Add logic to handle invalid path characters

	if errCount > 0 {
		return false
	}

	return true
}

// RsyncRecursive runs `gsutil rsync` in recursive mode. The caller of this
// function has to ensure that the provided paths are prefixed with gs:// if
// necessary (see `NormalizeGCSPath()`).
// TODO: Implementation of `gsutil rsync` should support local directory copies
//       with `-d`
func RsyncRecursive(src, dst string) error {
	if !IsPathNormalized(src) || !IsPathNormalized(dst) {
		return errors.New("cannot run `gsutil rsync` as one or more paths does not begin with `gs://`")
	}

	return errors.Wrap(
		gcp.GSUtil(concurrentFlag, "rsync", recursiveFlag, src, dst),
		"running gsutil rsync",
	)
}

// PathExists returns true if the specified GCS path exists.
func PathExists(gcsPath string) (bool, error) {
	if !IsPathNormalized(gcsPath) {
		return false, errors.New("cannot run `gsutil ls` GCS path does not begin with `gs://`")
	}

	err := gcp.GSUtil(
		"ls",
		gcsPath,
	)
	if err != nil {
		return false, err
	}

	logrus.Infof("Found %s", gcsPath)
	return true, nil
}

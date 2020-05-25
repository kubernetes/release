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
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/gcp"
)

var (
	gcsPrefix      = "gs://"
	concurrentFlag = "-m"
	recursiveFlag  = "-r"
)

// CopyToGCS copies a local directory to the specified GCS path
func CopyToGCS(src, gcsPath string, recursive, concurrent bool) error {
	gcsPath = normalizeGCSPath(gcsPath)

	logrus.Infof("Copying %s to GCS (%s)", src, gcsPath)
	return bucketCopy(src, gcsPath, recursive, concurrent)
}

// CopyToLocal copies a GCS path to the specified local directory
func CopyToLocal(gcsPath, dst string, recursive, concurrent bool) error {
	gcsPath = normalizeGCSPath(gcsPath)

	logrus.Infof("Copying GCS (%s) to %s", gcsPath, dst)
	return bucketCopy(gcsPath, dst, recursive, concurrent)
}

func bucketCopy(src, dst string, recursive, concurrent bool) error {
	args := []string{}

	if concurrent {
		logrus.Info("Setting GCS copy to run concurrently")
		args = append(args, concurrentFlag)
	}

	args = append(args, "cp")
	if recursive {
		logrus.Info("Setting GCS copy to run recursively")
		args = append(args, recursiveFlag)
	}
	args = append(args, src, dst)

	cpErr := command.Execute(gcp.GSUtilExecutable, args...)
	if cpErr != nil {
		return errors.Wrap(cpErr, "gcs copy")
	}

	return nil
}

func normalizeGCSPath(gcsPath string) string {
	gcsPath = strings.TrimPrefix(gcsPath, gcsPrefix)
	gcsPath = gcsPrefix + gcsPath

	return gcsPath
}

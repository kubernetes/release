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

// CopyToGCS copies a local directory to the specified GCS path
func CopyToGCS(src, gcsPath string, opts *Options) error {
	logrus.Infof("Copying %s to GCS (%s)", src, gcsPath)
	gcsPath = normalizeGCSPath(gcsPath)

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
func CopyToLocal(gcsPath, dst string, opts *Options) error {
	logrus.Infof("Copying GCS (%s) to %s", gcsPath, dst)
	gcsPath = normalizeGCSPath(gcsPath)

	return bucketCopy(gcsPath, dst, opts)
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

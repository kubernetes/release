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
	"context"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/gcp/gcs"
	"k8s.io/release/pkg/util"
	"k8s.io/utils/pointer"
)

// PushBuild is the main structure for pushing builds.
type PushBuild struct {
	opts *PushBuildOptions
}

// PushBuildOptions are the main options to pass to `PushBuild`.
type PushBuildOptions struct {
	// Specify an alternate bucket for pushes (normally 'devel' or 'ci').
	Bucket string

	// Specify an alternate build directory (defaults to '_output').
	BuildDir string

	// If set, push docker images to specified registry/project.
	DockerRegistry string

	// Comma separated list which can be used to upload additional version
	// files to GCS. The path is relative and is append to a GCS path. (--ci
	// only).
	ExtraVersionMarkers string

	// Specify a suffix to append to the upload destination on GCS.
	GCSSuffix string

	// Specify an alternate bucket for pushes (normally 'devel' or 'ci').
	ReleaseType string

	// Append suffix to version name if set.
	VersionSuffix string

	// Do not exit error if the build already exists on the gcs path.
	AllowDup bool

	// Used when called from Jenkins (for ci runs).
	CI bool

	// Do not update the latest file.
	NoUpdateLatest bool

	// Do not mark published bits on GCS as publicly readable.
	PrivateBucket bool

	// Specifies a fast build (linux amd64 only).
	Fast bool

	// Specifies if we should push to the bucket or the user suffixed one.
	NoMock bool
}

type stageFile struct {
	srcPath  string
	dstPath  string
	required bool
}

var gcpStageFiles = []stageFile{
	{
		srcPath:  filepath.Join(GCEPath, "configure-vm.sh"),
		dstPath:  filepath.Join(GCSStagePath, "extra/gce"),
		required: false,
	},
	{
		srcPath:  filepath.Join(GCIPath, "node.yaml"),
		dstPath:  filepath.Join(GCSStagePath, "extra/gce"),
		required: true,
	},
	{
		srcPath:  filepath.Join(GCIPath, "master.yaml"),
		dstPath:  filepath.Join(GCSStagePath, "extra/gce"),
		required: true,
	},
	{
		srcPath:  filepath.Join(GCIPath, "configure.sh"),
		dstPath:  filepath.Join(GCSStagePath, "extra/gce"),
		required: true,
	},
	{
		srcPath:  filepath.Join(GCIPath, "shutdown.sh"),
		dstPath:  filepath.Join(GCSStagePath, "extra/gce"),
		required: false,
	},
}

var windowsStageFiles = []stageFile{
	{
		srcPath:  filepath.Join(WindowsLocalPath, "configure.ps1"),
		dstPath:  WindowsGCSPath,
		required: true,
	},
	{
		srcPath:  filepath.Join(WindowsLocalPath, "common.psm1"),
		dstPath:  WindowsGCSPath,
		required: true,
	},
	{
		srcPath:  filepath.Join(WindowsLocalPath, "k8s-node-setup.psm1"),
		dstPath:  WindowsGCSPath,
		required: true,
	},
	{
		srcPath:  filepath.Join(WindowsLocalPath, "testonly/install-ssh.psm1"),
		dstPath:  WindowsGCSPath,
		required: true,
	},
	{
		srcPath:  filepath.Join(WindowsLocalPath, "testonly/user-profile.psm1"),
		dstPath:  WindowsGCSPath,
		required: true,
	},
}

// NewPushBuild can be used to create a new PushBuild instnace.
func NewPushBuild(opts *PushBuildOptions) *PushBuild {
	return &PushBuild{opts}
}

// Push pushes the build by taking the internal options into account.
func (p *PushBuild) Push() error {
	var latest string

	// Check if latest build uses bazel
	dir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "get working directory")
	}

	isBazel, err := BuiltWithBazel(dir)
	if err != nil {
		return errors.Wrap(err, "identify if release built with Bazel")
	}

	if isBazel {
		logrus.Info("Using Bazel build version")
		version, err := ReadBazelVersion(dir)
		if err != nil {
			return errors.Wrap(err, "read Bazel build version")
		}
		latest = version
	} else {
		logrus.Info("Using Dockerized build version")
		version, err := ReadDockerizedVersion(dir)
		if err != nil {
			return errors.Wrap(err, "read Dockerized build version")
		}
		latest = version
	}

	logrus.Infof("Found build version: %s", latest)

	valid, err := IsValidReleaseBuild(latest)
	if err != nil {
		return errors.Wrap(err, "determine if release build version is valid")
	}
	if !valid {
		return errors.Errorf("build version %s is not valid for release", latest)
	}

	if p.opts.CI && IsDirtyBuild(latest) {
		return errors.New("refusing to push dirty build with --ci flag given")
	}

	if p.opts.VersionSuffix != "" {
		latest += "-" + p.opts.VersionSuffix
	}

	logrus.Infof("Latest version is %s", latest)

	releaseBucket := p.opts.Bucket
	if p.opts.NoMock {
		logrus.Infof("Running a *REAL* push with bucket %s", releaseBucket)
	} else {
		u, err := user.Current()
		if err != nil {
			return errors.Wrap(err, "identify current user")
		}

		releaseBucket += "-" + u.Username
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return errors.Wrap(err, "fetching gcloud credentials, try running \"gcloud auth application-default login\"")
	}

	bucket := client.Bucket(releaseBucket)
	if bucket == nil {
		return errors.Errorf(
			"identify specified bucket for artifacts: %s", releaseBucket,
		)
	}

	// Check if bucket exists and user has permissions
	requiredGCSPerms := []string{"storage.objects.create"}
	perms, err := bucket.IAM().TestPermissions(
		context.Background(), requiredGCSPerms,
	)
	if err != nil {
		return errors.Wrap(err, "find release artifact bucket")
	}
	if len(perms) != 1 {
		return errors.Errorf(
			"GCP user must have at least %s permissions on bucket %s",
			requiredGCSPerms, releaseBucket,
		)
	}

	buildDir := p.opts.BuildDir
	if err = util.RemoveAndReplaceDir(
		filepath.Join(buildDir, GCSStagePath),
	); err != nil {
		return errors.Wrap(err, "remove and replace GCS staging directory")
	}

	// Copy release tarballs to local GCS staging directory for push
	if err = util.CopyDirContentsLocal(
		filepath.Join(buildDir, ReleaseTarsPath),
		filepath.Join(buildDir, GCSStagePath),
	); err != nil {
		return errors.Wrap(err, "copy source directory into destination")
	}

	// Copy helpful GCP scripts to local GCS staging directory for push
	for _, file := range gcpStageFiles {
		if err := util.CopyFileLocal(
			filepath.Join(buildDir, file.srcPath),
			filepath.Join(buildDir, file.dstPath),
			file.required,
		); err != nil {
			return errors.Wrap(err, "copy GCP stage files")
		}
	}

	// Copy helpful Windows scripts to local GCS staging directory for push
	for _, file := range windowsStageFiles {
		if err := util.CopyFileLocal(
			filepath.Join(buildDir, file.srcPath),
			filepath.Join(buildDir, file.dstPath),
			file.required,
		); err != nil {
			return errors.Wrap(err, "copy Windows stage files")
		}
	}

	// Copy the "naked" binaries to GCS. This is useful for install scripts
	// that download the binaries directly and don't need tars.
	if err := CopyBinaries(
		filepath.Join(buildDir, ReleaseStagePath),
	); err != nil {
		return errors.Wrap(err, "stage binaries")
	}

	// Write the release checksums
	gcsStagePath := filepath.Join(buildDir, GCSStagePath, latest)
	if err := WriteChecksums(gcsStagePath); err != nil {
		return errors.Wrap(err, "write checksums")
	}

	// Publish container images
	gcsDest := p.opts.ReleaseType
	if p.opts.CI {
		gcsDest = "ci"
	}
	gcsDest += p.opts.GCSSuffix

	if p.opts.Fast {
		gcsDest = filepath.Join(gcsDest, "fast")
	}
	logrus.Infof("GCS destination is %s", gcsDest)

	copyOpts := gcs.DefaultGCSCopyOptions
	copyOpts.NoClobber = pointer.BoolPtr(p.opts.AllowDup)

	if err := gcs.CopyToGCS(
		gcsStagePath,
		filepath.Join(releaseBucket, gcsDest, latest),
		copyOpts,
	); err != nil {
		return errors.Wrap(err, "copy artifacts to GCS")
	}

	if p.opts.DockerRegistry != "" {
		if err := NewImages().Publish(
			p.opts.DockerRegistry,
			strings.ReplaceAll(latest, "+", "_"),
			buildDir,
		); err != nil {
			return errors.Wrap(err, "publish container images")
		}
	}

	if !p.opts.CI {
		logrus.Info("No CI flag set, we're done")
		return nil
	}

	// Publish release to GCS
	versionMarkers := strings.Split(p.opts.ExtraVersionMarkers, ",")
	if err := NewPublisher().PublishVersion(
		gcsDest, latest, buildDir, releaseBucket, versionMarkers,
		p.opts.PrivateBucket, p.opts.NoMock,
	); err != nil {
		return errors.Wrap(err, "publish release")
	}

	return nil
}

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

package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/tar"
	"k8s.io/release/pkg/util"
)

type stageFile struct {
	srcPath  string
	dstPath  string
	required bool
}

const extraDir = "extra"

// ExtraGcpStageFiles defines extra GCP files to be staged if `StageExtraFiles`
// in `Options` is set to `true`.
var ExtraGcpStageFiles = []stageFile{
	{
		srcPath:  filepath.Join(release.GCEPath, "configure-vm.sh"),
		dstPath:  extraDir + "/gce/configure-vm.sh",
		required: false,
	},
	{
		srcPath:  filepath.Join(release.GCIPath, "node.yaml"),
		dstPath:  extraDir + "/gce/node.yaml",
		required: true,
	},
	{
		srcPath:  filepath.Join(release.GCIPath, "master.yaml"),
		dstPath:  extraDir + "/gce/master.yaml",
		required: true,
	},
	{
		srcPath:  filepath.Join(release.GCIPath, "configure.sh"),
		dstPath:  extraDir + "/gce/configure.sh",
		required: true,
	},
	{
		srcPath:  filepath.Join(release.GCIPath, "shutdown.sh"),
		dstPath:  extraDir + "/gce/shutdown.sh",
		required: false,
	},
}

// ExtraWindowsStageFiles defines extra Windows files to be staged if
// `StageExtraFiles` in `Options` is set to `true`.
var ExtraWindowsStageFiles = []stageFile{
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "configure.ps1"),
		dstPath:  extraDir + "/gce/windows/configure.ps1",
		required: true,
	},
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "common.psm1"),
		dstPath:  extraDir + "/gce/windows/common.psm1",
		required: true,
	},
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "k8s-node-setup.psm1"),
		dstPath:  extraDir + "/gce/windows/k8s-node-setup.psm1",
		required: true,
	},
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "testonly/install-ssh.psm1"),
		dstPath:  extraDir + "/gce/windows/install-ssh.psm1",
		required: true,
	},
	{
		srcPath:  filepath.Join(release.WindowsLocalPath, "testonly/user-profile.psm1"),
		dstPath:  extraDir + "/gce/windows/user-profile.psm1",
		required: true,
	},
}

// Push pushes the build by taking the internal options into account.
func (bi *Instance) Push() error {
	version, err := bi.findLatestVersion()
	if err != nil {
		return errors.Wrap(err, "find latest version")
	}

	if version == "" {
		return errors.New("cannot push an empty version")
	}

	logrus.Infof("Latest version is %s", version)

	if err := bi.CheckReleaseBucket(); err != nil {
		return errors.Wrap(err, "check release bucket access")
	}

	if err := bi.StageLocalArtifacts(); err != nil {
		return errors.Wrap(err, "staging local artifacts")
	}

	if err := bi.PushContainerImages(); err != nil {
		return errors.Wrap(err, "push container images")
	}

	gcsDest, gcsDestErr := bi.getGCSBuildPath(version)
	if gcsDestErr != nil {
		return errors.Wrap(gcsDestErr, "get GCS destination")
	}

	if err := bi.PushReleaseArtifacts(
		filepath.Join(bi.opts.BuildDir, release.GCSStagePath, version),
		gcsDest,
	); err != nil {
		return errors.Wrap(err, "push release artifacts")
	}

	if !bi.opts.CI {
		logrus.Info("No CI flag set, we're done")
		return nil
	}

	if bi.opts.NoUpdateLatest {
		logrus.Info("Not updating version markers")
		return nil
	}

	// Publish release to GCS
	extraVersionMarkers := bi.opts.ExtraVersionMarkers
	if err := release.NewPublisher().PublishVersion(
		bi.opts.BuildType,
		version,
		bi.opts.BuildDir,
		bi.opts.Bucket,
		bi.opts.GCSRoot,
		extraVersionMarkers,
		bi.opts.PrivateBucket,
		bi.opts.Fast,
	); err != nil {
		return errors.Wrap(err, "publish release")
	}

	return nil
}

func (bi *Instance) findLatestVersion() (latestVersion string, err error) {
	// Check if latest build uses bazel
	if bi.opts.RepoRoot == "" {
		bi.opts.RepoRoot, err = os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "get working directory")
		}
	}

	isBazel, err := release.BuiltWithBazel(bi.opts.RepoRoot)
	if err != nil {
		return "", errors.Wrap(err, "identify if release built with Bazel")
	}

	latestVersion = bi.opts.Version
	if bi.opts.Version == "" {
		if isBazel {
			logrus.Info("Using Bazel build version")
			version, err := release.ReadBazelVersion(bi.opts.RepoRoot)
			if err != nil {
				return "", errors.Wrap(err, "read Bazel build version")
			}
			latestVersion = version
		} else {
			logrus.Info("Using Dockerized build version")
			version, err := release.ReadDockerizedVersion(bi.opts.RepoRoot)
			if err != nil {
				return "", errors.Wrap(err, "read Dockerized build version")
			}
			latestVersion = version
		}
	}

	logrus.Infof("Using build version: %s", latestVersion)

	valid, err := release.IsValidReleaseBuild(latestVersion)
	if err != nil {
		return "", errors.Wrap(
			err, "determine if release build version is valid",
		)
	}
	if !valid {
		return "", errors.Errorf(
			"build version %s is not valid for release", latestVersion,
		)
	}

	if bi.opts.CI && release.IsDirtyBuild(latestVersion) {
		return "", errors.Errorf(
			"refusing to push dirty build %s with --ci flag given",
			latestVersion,
		)
	}

	if bi.opts.VersionSuffix != "" {
		latestVersion += "-" + bi.opts.VersionSuffix
	}

	setupBuildDir(bi, isBazel)

	return strings.TrimSpace(latestVersion), nil
}

func setupBuildDir(bi *Instance, isBazel bool) {
	if bi.opts.BuildDir == "" {
		logrus.Info("BuildDir is not set, setting it automatically")
		if isBazel {
			logrus.Infof(
				"Release is build by bazel, using BuildDir as %s",
				release.BazelBuildDir,
			)
			bi.opts.BuildDir = release.BazelBuildDir
		} else {
			logrus.Infof(
				"Release is build in a container, using BuildDir as %s",
				release.BuildDir,
			)
			bi.opts.BuildDir = release.BuildDir
		}
	}
	// convert buildDir to an absolute path
	bi.opts.BuildDir = filepath.Join(bi.opts.RepoRoot, bi.opts.BuildDir)
	logrus.Infof(
		"Setting BuildDir to %s",
		bi.opts.BuildDir,
	)
}

// CheckReleaseBucket verifies that a release bucket exists and the current
// authenticated GCP user has write permissions to it.
func (bi *Instance) CheckReleaseBucket() error {
	logrus.Infof("Checking bucket %s for write permissions", bi.opts.Bucket)

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return errors.Wrap(err,
			"fetching gcloud credentials, try running "+
				`"gcloud auth application-default login"`,
		)
	}

	bucket := client.Bucket(bi.opts.Bucket)
	if bucket == nil {
		return errors.Errorf(
			"identify specified bucket for artifacts: %s", bi.opts.Bucket,
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
			requiredGCSPerms, bi.opts.Bucket,
		)
	}

	return nil
}

// StageLocalArtifacts locally stages the release artifacts
func (bi *Instance) StageLocalArtifacts() error {
	logrus.Info("Staging local artifacts")
	stageDir := filepath.Join(bi.opts.BuildDir, release.GCSStagePath, bi.opts.Version)

	logrus.Infof("Cleaning staging dir %s", stageDir)
	if err := util.RemoveAndReplaceDir(stageDir); err != nil {
		return errors.Wrap(err, "remove and replace GCS staging directory")
	}

	// Copy release tarballs to local GCS staging directory for push
	logrus.Info("Copying release tarballs")
	if err := util.CopyDirContentsLocal(
		filepath.Join(bi.opts.BuildDir, release.ReleaseTarsPath), stageDir,
	); err != nil {
		return errors.Wrap(err, "copy source directory into destination")
	}

	if bi.opts.StageExtraFiles {
		// Copy helpful GCP scripts to local GCS staging directory for push
		logrus.Info("Copying extra GCP stage files")
		if err := bi.copyStageFiles(stageDir, ExtraGcpStageFiles); err != nil {
			return errors.Wrapf(err, "copy GCP stage files")
		}

		// Copy helpful Windows scripts to local GCS staging directory for push
		logrus.Info("Copying extra Windows stage files")
		if err := bi.copyStageFiles(stageDir, ExtraWindowsStageFiles); err != nil {
			return errors.Wrapf(err, "copy Windows stage files")
		}
	}

	// Copy the plain binaries to GCS. This is useful for install scripts that
	// download the binaries directly and don't need tars.
	plainBinariesPath := filepath.Join(bi.opts.BuildDir, release.ReleaseStagePath)
	if util.Exists(plainBinariesPath) {
		logrus.Info("Copying plain binaries")
		if err := release.CopyBinaries(
			filepath.Join(bi.opts.BuildDir, release.ReleaseStagePath),
			stageDir,
		); err != nil {
			return errors.Wrap(err, "stage binaries")
		}
	} else {
		logrus.Infof(
			"Skipping not existing plain binaries dir %s", plainBinariesPath,
		)
	}

	// Write the release checksums
	logrus.Info("Writing checksums")
	if err := release.WriteChecksums(stageDir); err != nil {
		return errors.Wrap(err, "write checksums")
	}
	return nil
}

// copyStageFiles takes the staging dir and copies each file of `files` into
// it. It also ensures that the base dir exists before copying the file (if the
// file is `required`).
func (bi *Instance) copyStageFiles(stageDir string, files []stageFile) error {
	for _, file := range files {
		dstPath := filepath.Join(stageDir, file.dstPath)

		if file.required {
			if err := os.MkdirAll(
				filepath.Dir(dstPath), os.FileMode(0o755),
			); err != nil {
				return errors.Wrapf(
					err, "create destination path %s", file.dstPath,
				)
			}
		}

		if err := util.CopyFileLocal(
			filepath.Join(bi.opts.BuildDir, file.srcPath),
			dstPath, file.required,
		); err != nil {
			return errors.Wrapf(err, "copy stage file")
		}
	}

	return nil
}

// PushReleaseArtifacts can be used to push local artifacts from the `srcPath`
// to the remote `gcsPath`. The Bucket has to be set via the `Bucket` option.
func (bi *Instance) PushReleaseArtifacts(srcPath, gcsPath string) error {
	dstPath, dstPathErr := bi.objStore.NormalizePath(gcsPath)
	if dstPathErr != nil {
		return errors.Wrap(dstPathErr, "normalize GCS destination")
	}

	logrus.Infof("Pushing release artifacts from %s to %s", srcPath, dstPath)

	return errors.Wrap(
		bi.objStore.RsyncRecursive(srcPath, dstPath), "rsync artifacts to GCS",
	)
}

// PushContainerImages will publish container images into the set
// `Registry`. It also validates if the remove manifests are correct,
// which can be turned of by setting `ValidateRemoteImageDigests` to `false`.
func (bi *Instance) PushContainerImages() error {
	if bi.opts.Registry == "" {
		logrus.Info("Registry is not set, will not publish container images")
		return nil
	}

	images := release.NewImages()
	logrus.Infof("Publishing container images for %s", bi.opts.Version)

	if err := images.Publish(
		bi.opts.Registry, bi.opts.Version, bi.opts.BuildDir,
	); err != nil {
		return errors.Wrap(err, "publish container images")
	}

	if !bi.opts.ValidateRemoteImageDigests {
		logrus.Info("Will not validate remote image digests")
		return nil
	}

	if err := images.Validate(
		bi.opts.Registry, bi.opts.Version, bi.opts.BuildDir,
	); err != nil {
		return errors.Wrap(err, "validate container images")
	}

	return nil
}

// CopyStagedFromGCS copies artifacts from GCS and between buckets as needed.
// TODO: Investigate if it's worthwhile to use any of the bi.objStore.Get*Path()
//       functions here or create a new one to populate staging paths
func (bi *Instance) CopyStagedFromGCS(stagedBucket, buildVersion string) error {
	logrus.Info("Copy staged release artifacts from GCS")

	bi.objStore.SetOptions(
		bi.objStore.WithNoClobber(bi.opts.AllowDup),
		bi.objStore.WithAllowMissing(false),
	)

	gcsStageRoot := filepath.Join(bi.opts.Bucket, release.StagePath, buildVersion, bi.opts.Version)
	src := filepath.Join(gcsStageRoot, release.GCSStagePath, bi.opts.Version)

	gcsSrc, gcsSrcErr := bi.objStore.NormalizePath(src)
	if gcsSrcErr != nil {
		return errors.Wrap(gcsSrcErr, "normalize GCS source")
	}

	dst, dstErr := bi.objStore.NormalizePath(bi.opts.Bucket, "release", bi.opts.Version)
	if dstErr != nil {
		return errors.Wrap(dstErr, "normalize GCS destination")
	}

	logrus.Infof("Bucket to bucket rsync from %s to %s", gcsSrc, dst)
	if err := bi.objStore.RsyncRecursive(gcsSrc, dst); err != nil {
		return errors.Wrap(err, "copy stage to release bucket")
	}

	src = filepath.Join(src, release.KubernetesTar)
	dst = filepath.Join(bi.opts.BuildDir, release.GCSStagePath, bi.opts.Version, release.KubernetesTar)
	logrus.Infof("Copy kubernetes tarball %s to %s", src, dst)
	if err := bi.objStore.CopyToLocal(src, dst); err != nil {
		return errors.Wrapf(err, "copy to local")
	}

	src = filepath.Join(gcsStageRoot, release.ImagesPath)
	if err := os.MkdirAll(bi.opts.BuildDir, os.FileMode(0o755)); err != nil {
		return errors.Wrap(err, "create dst dir")
	}
	logrus.Infof("Copy container images %s to %s", src, bi.opts.BuildDir)
	if err := bi.objStore.CopyToLocal(src, bi.opts.BuildDir); err != nil {
		return errors.Wrapf(err, "copy to local")
	}

	return nil
}

// StageLocalSourceTree creates a src.tar.gz from the Kubernetes sources and
// uploads it to GCS.
func (bi *Instance) StageLocalSourceTree(workDir, buildVersion string) error {
	tarballPath := filepath.Join(workDir, release.SourcesTar)
	logrus.Infof("Creating source tree tarball %s", tarballPath)

	exclude, err := regexp.Compile(fmt.Sprintf(`.*/%s-.*`, release.BuildDir))
	if err != nil {
		return errors.Wrap(err, "compile tarball exclude regex")
	}

	if err := tar.Compress(
		tarballPath, filepath.Join(workDir, "src"), exclude,
	); err != nil {
		return errors.Wrap(err, "create tarball")
	}

	logrus.Infof("Uploading source tree tarball to GCS")
	bi.objStore.SetOptions(
		bi.objStore.WithAllowMissing(false),
		bi.objStore.WithNoClobber(false),
	)
	if err := bi.objStore.CopyToRemote(
		tarballPath,
		filepath.Join(bi.opts.Bucket, release.StagePath, buildVersion, release.SourcesTar),
	); err != nil {
		return errors.Wrap(err, "copy tarball to GCS")
	}

	logrus.Infof("Removing local source tree tarball")
	return errors.Wrap(os.RemoveAll(tarballPath), "remove local source tarball")
}

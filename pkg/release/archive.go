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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/command"
	"k8s.io/release/pkg/gcp"
	"k8s.io/release/pkg/object"

	"k8s.io/release/pkg/util"
)

const (
	archiveDirPrefix = "anago-"
)

// Archiver stores the release build directory in a bucket
// along with it's logs
type Archiver struct {
	impl archiverImpl
	opts *ArchiverOptions
}

// NewArchiver create a new archiver with the default implementation
func NewArchiver(opts *ArchiverOptions) *Archiver {
	return &Archiver{&defaultArchiverImpl{}, opts}
}

// SetImpl changes the archiver implementation
func (archiver *Archiver) SetImpl(impl archiverImpl) {
	archiver.impl = impl
}

// ArchiverOptions set the options used when archiving a release
type ArchiverOptions struct {
	ReleaseBuildDir string // Build directory that will be archived
	LogsDirectory   string // Subdirectory to get the logs from

	StageGCSPath   string // Stage path in the bucket // ie gs://kubernetes-release/stage
	ArchiveGCSPath string // Archive path in the bucket // ie gs://kubernetes-release/archive

	BuildVersion string // Version tag of the release we are archiving
}

// ArchiveBucketPath returns the bucket path we the release will be stored
func (o *ArchiverOptions) ArchiveBucketPath() string {
	// local archive_bucket="gs://$RELEASE_BUCKET/archive"
	if o.ArchiveGCSPath == "" || o.BuildVersion == "" {
		return ""
	}
	gcs := object.NewGCS()
	archiveBucketPath, err := gcs.NormalizePath(
		filepath.Join(o.ArchiveGCSPath, archiveDirPrefix+o.BuildVersion),
	)
	if err != nil {
		logrus.Error(err)
		return ""
	}
	return archiveBucketPath
}

// Validate checks if the set values are correct and complete to
// start running the archival process
func (o *ArchiverOptions) Validate() error {
	if o.LogsDirectory == "" {
		return errors.New("missing logs subdirectory in archive options")
	}
	if o.ArchiveGCSPath == "" {
		return errors.New("archival bucket location is missing from options")
	}
	if o.StageGCSPath == "" {
		return errors.New("stage bucket location is missing from options")
	}
	if !util.Exists(o.ReleaseBuildDir) {
		return errors.New("GCB worskapce directory does not exist")
	}
	if !util.Exists(filepath.Join(o.LogsDirectory)) {
		return errors.New("logs directory does not exist")
	}
	if o.BuildVersion == "" {
		return errors.New("release tag in archiver options is empty")
	}

	// Check if the tag is well formed
	_, err := util.TagStringToSemver(o.BuildVersion)
	if err != nil {
		return errors.Wrap(err, "verifying release tag")
	}

	return nil
}

//counterfeiter:generate . archiverImpl
type archiverImpl interface {
	CopyReleaseToBucket(string, string) error
	DeleteStalePasswordFiles(string) error
	MakeFilesPrivate(string, []string) error
	GetLogFiles(string) ([]string, error)
	ValidateOptions(*ArchiverOptions) error
	CopyReleaseLogs([]string, string) error
	CleanStagedBuilds(string, string) error
}

type defaultArchiverImpl struct{}

// ArchiveRelease stores the release directory and logs in a GCP
// bucket for archival purposes. Log files are sanitized and made private
func (archiver *Archiver) ArchiveRelease() error {
	// Verify options are complete
	if err := archiver.impl.ValidateOptions(archiver.opts); err != nil {
		return errors.Wrap(err, "validating archive options")
	}

	// local logfiles=$(ls $LOGFILE{,.[0-9]} 2>/dev/null || true)
	// Before moving anything, find the log files (full path)
	logFiles, err := archiver.impl.GetLogFiles(archiver.opts.LogsDirectory)
	if err != nil {
		return errors.Wrap(err, "getting files from logs directory")
	}

	// TODO: Is this still relevant?
	// local text="files"

	// copy_logs_to_workdir
	if err := archiver.impl.CopyReleaseLogs(
		logFiles, archiver.opts.ReleaseBuildDir,
	); err != nil {
		return errors.Wrap(err, "copying release logs to archive")
	}

	// # TODO: Copy $PROGSTATE as well to GCS and restore it if found
	// # also delete if complete or just delete once copied back to $TMPDIR
	// # This is so failures on GCB can be restarted / reentrant too.

	// if [[ $arg != "--files-only" ]]; then
	//  dash_args="-rc"
	//   text="contents"
	// fi

	// Remove temporary password file so not to alarm passers-by.
	if err := archiver.impl.DeleteStalePasswordFiles(
		archiver.opts.ReleaseBuildDir,
	); err != nil {
		return errors.Wrap(err, "looking for stale password files")
	}

	// Copy the logs to the bucket
	if err = archiver.impl.CopyReleaseToBucket(
		archiver.opts.ReleaseBuildDir,
		archiver.opts.ArchiveBucketPath(),
	); err != nil {
		return errors.Wrap(err, "while copying the release directory")
	}

	// Make the logs private (remove AllUsers from GCS ACL)
	if err := archiver.impl.MakeFilesPrivate(
		archiver.opts.ArchiveBucketPath(), logFiles,
	); err != nil {
		return errors.Wrapf(err, "setting private ACL on logs")
	}

	// Clean previous staged builds
	if err := archiver.impl.CleanStagedBuilds(
		archiver.opts.StageGCSPath,
		archiver.opts.BuildVersion,
	); err != nil {
		return errors.Wrap(err, "deleting previous staged builds")
	}

	logrus.Info("Release archive complete")
	return nil
}

// validateOptions runs the options validation
func (a *defaultArchiverImpl) ValidateOptions(o *ArchiverOptions) error {
	return errors.Wrap(o.Validate(), "validating options")
}

// makeFilesPrivate updates the ACL on the logs to ensure they do not remain worl-readable
func (a *defaultArchiverImpl) MakeFilesPrivate(
	archiveBucketPath string, logFiles []string,
) error {
	for _, logFile := range logFiles {
		logrus.Infof("Ensure PRIVATE ACL on %s/%s", archiveBucketPath, logFile)
		// logrun -s $GSUTIL acl ch -d AllUsers "$archive_bucket/$build_dir/${LOGFILE##*/}*" || true
		if err := gcp.GSUtil(
			"acl", "ch", "-d", "AllUsers", filepath.Join(archiveBucketPath, logFile),
		); err != nil {
			return errors.Wrapf(err, "removing public access from %s", logFile)
		}
	}
	return nil
}

// deleteStalePasswordFiles emoves temporary password file so not to alarm passers-by.
func (a *defaultArchiverImpl) DeleteStalePasswordFiles(releaseBuildDir string) error {
	if err := command.NewWithWorkDir(
		releaseBuildDir, "find", "-type", "f", "-name", "rsyncd.password", "-delete",
	).RunSuccess(); err != nil {
		return errors.Wrap(err, "deleting temporary password files")
	}
	return nil
}

// copyReleaseLogs gets a slice of log file names. Those files are
// sanitized to remove sensitive data and control characters and then are
// copied to the GCB working directory.
func (a *defaultArchiverImpl) CopyReleaseLogs(logFiles []string, targetDir string) error {
	for _, fileName := range logFiles {
		// Strip the logfiles from control chars and sensitive data
		if err := util.CleanLogFile(fileName); err != nil {
			return errors.Wrap(err, "sanitizing logfile")
		}

		logrus.Infof("Copying %s to %s", fileName, targetDir)
		if err := util.CopyFileLocal(
			fileName, filepath.Join(targetDir, filepath.Base(fileName)), true,
		); err != nil {
			return errors.Wrapf(err, "Copying logfile %s to %s", fileName, targetDir)
		}
	}
	return nil
}

// copyReleaseToBucket Copies the release directory to the specified bucket location
func (a *defaultArchiverImpl) CopyReleaseToBucket(releaseBuildDir, archiveBucketPath string) error {
	// TODO: Check if we have write access to the bucket?

	// Create a GCS cliente to copy the release
	gcs := object.NewGCS()

	logrus.Infof("Copy %s $text to %s...", releaseBuildDir, archiveBucketPath)

	// logrun $GSUTIL -mq cp $dash_args $WORKDIR/* $archive_bucket/$build_dir || true
	if err := gcs.CopyToRemote(releaseBuildDir, archiveBucketPath); err != nil {
		return errors.Wrap(err, "copying release directory to bucket")
	}

	return nil
}

// GetLogFiles reads a directory and returns the files that are anago logs
func (a *defaultArchiverImpl) GetLogFiles(logsDir string) ([]string, error) {
	logFiles := []string{}
	tmpContents, err := ioutil.ReadDir(logsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "searching for logfiles in %s", logsDir)
	}
	for _, finfo := range tmpContents {
		if strings.HasPrefix(finfo.Name(), "anago") &&
			strings.Contains(finfo.Name(), ".log") {
			logFiles = append(logFiles, filepath.Join(logsDir, finfo.Name()))
		}
	}
	return logFiles, nil
}

// CleanStagedBuilds removes all past staged builds from the same
// Major.Minor version we are running now
func (a *defaultArchiverImpl) CleanStagedBuilds(bucketPath, buildVersion string) error {
	// Build the prefix we will be looking for
	semver, err := util.TagStringToSemver(buildVersion)
	if err != nil {
		return errors.Wrap(err, "parsing semver from tag")
	}
	dirPrefix := fmt.Sprintf("%s%d.%d", util.TagPrefix, semver.Major, semver.Minor)

	// Normalize the bucket parh
	// Build a GCS object to delete old builds
	gcs := object.NewGCS()
	gcs.SetOptions(
		gcs.WithConcurrent(true),
		gcs.WithRecursive(true),
	)

	// Normalize the bucket path
	path, err := gcs.NormalizePath(bucketPath, dirPrefix+"*")
	if err != nil {
		return errors.Wrap(err, "normalizing stage path")
	}

	// Get all staged build that match the pattern
	output, err := gcp.GSUtilOutput("ls", "-d", path)
	if err != nil {
		return errors.Wrap(err, "listing bucket contents")
	}

	for _, line := range strings.Fields(output) {
		if strings.Contains(line, dirPrefix) && !strings.Contains(line, buildVersion) {
			logrus.Infof("Deleting previous staged build: %s", line)
			if err := gcs.DeletePath(line); err != nil {
				return errors.Wrap(err, "calling gsutil to delete build")
			}
		}
	}
	return nil
}

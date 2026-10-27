/*
Copyright 2026 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-utils/command"
	"sigs.k8s.io/release-utils/helpers"

	"k8s.io/release/pkg/release"
)

const s3Scheme = "s3://"

// s3Client models the S3 operations used to push builds.
type s3Client interface {
	CheckBucket(bucket string) error
	PathExists(bucket, keyPrefix string) (bool, error)
	ReadObject(bucket, key string) (string, error)
	UploadDir(dir, bucket, keyPrefix string) error
	UploadVersionMarker(file, bucket, key string) error
}

// awsCLI implements s3Client with the AWS CLI, which, like the gcloud CLI used
// for GCS, uploads large files in parallel parts. It needs no region, as it
// follows the redirects of S3 to the region of the bucket.
type awsCLI struct{}

func (*awsCLI) CheckBucket(bucket string) error {
	return command.New("aws", "s3api", "head-bucket", "--bucket", bucket).RunSuccess()
}

func (*awsCLI) PathExists(bucket, keyPrefix string) (bool, error) {
	out, err := command.New(
		"aws", "s3api", "list-objects-v2", "--bucket", bucket, "--prefix", keyPrefix,
		"--max-keys", "1", "--no-paginate", "--query", "KeyCount", "--output", "text",
	).RunSilentSuccessOutput()
	if err != nil {
		return false, err
	}

	return out.OutputTrimNL() != "0", nil
}

func (*awsCLI) ReadObject(bucket, key string) (string, error) {
	out, err := command.New("aws", "s3", "cp", s3URI(bucket, key), "-").RunSilentSuccessOutput()
	if err != nil {
		return "", err
	}

	return out.Output(), nil
}

func (*awsCLI) UploadDir(dir, bucket, keyPrefix string) error {
	return command.New(
		"aws", "s3", "sync", "--no-progress", dir, s3URI(bucket, keyPrefix)+"/",
	).RunSuccess()
}

func (*awsCLI) UploadVersionMarker(file, bucket, key string) error {
	// Same metadata as the GCS version markers, so that they are not cached.
	return command.New(
		"aws", "s3", "cp", "--no-progress",
		"--content-type", "text/plain",
		"--cache-control", "private, max-age=0, no-transform",
		file, s3URI(bucket, key),
	).RunSuccess()
}

func s3URI(bucket, key string) string {
	return s3Scheme + bucket + "/" + key
}

// S3ObjectKey returns the key of the object that S3 serves at the https:// URL
// path urlPath.
//
// For compatibility with its original REST API, S3 decodes a literal "+" in a
// URL path as a space, and only "%2B" as a plus sign. CI versions contain a
// "+", and consumers build the URLs of the artifacts by joining paths, which
// writes the "+" literally. Storing the artifacts under keys with a space in
// place of the "+" serves them at the same URL paths as on GCS.
//
// Every "+" in the key is replaced, including in the prefix of the bucket URL.
// Tools that address the objects by key, like the AWS CLI and SDKs, or URLs
// that encode the "+" as "%2B", have to replace it with a space as well.
func S3ObjectKey(urlPath string) string {
	return strings.ReplaceAll(strings.Trim(urlPath, "/"), "+", " ")
}

func (bi *Instance) isS3() bool {
	return strings.HasPrefix(bi.opts.Bucket, s3Scheme)
}

// checkS3Options rejects the options that cannot push to an s3:// bucket
// before building or pushing anything.
func (bi *Instance) checkS3Options() error {
	if !bi.isS3() {
		return nil
	}

	if bi.opts.PrivateBucket {
		return errors.New("private buckets are not supported on S3, use a bucket policy to restrict access instead")
	}

	if !command.Available("aws") {
		return errors.New("pushing to S3 requires the aws CLI, which is not in $PATH")
	}

	return nil
}

// s3Location splits the s3:// Bucket into the bucket name and a key prefix.
func (bi *Instance) s3Location() (bucket, prefix string) {
	bucket, prefix, _ = strings.Cut(strings.TrimPrefix(bi.opts.Bucket, s3Scheme), "/")

	return bucket, prefix
}

// s3Key returns the key of the object at
// <prefix>/<bucketRoot>[/fast][/<parts>], the S3 counterpart of the GCS paths.
func (bi *Instance) s3Key(parts ...string) string {
	_, prefix := bi.s3Location()

	elems := []string{prefix, bi.opts.BucketRoot}
	if bi.opts.Fast {
		elems = append(elems, "fast")
	}

	return S3ObjectKey(path.Join(append(elems, parts...)...))
}

// s3BuildPathsExist is the S3 counterpart of the GCS checks in
// checkBuildExists.
func (bi *Instance) s3BuildPathsExist(version string) (found int, existErrors []error) {
	bucket, _ := bi.s3Location()

	for _, key := range []string{
		bi.s3Key(version) + "/",
		bi.s3Key(version, release.KubernetesTar),
		bi.s3Key(version, "bin") + "/",
	} {
		logrus.Infof("Checking if S3 build path (%s) exists", s3URI(bucket, key))

		exists, err := bi.s3.PathExists(bucket, key)
		switch {
		case err != nil:
			existErrors = append(existErrors, err)
		case !exists:
			existErrors = append(existErrors, fmt.Errorf("%s does not exist", s3URI(bucket, key)))
		default:
			found++
		}
	}

	return found, existErrors
}

// pushToS3 is the S3 counterpart of pushing the staged artifacts and the
// version markers to GCS. Unlike on GCS, the objects get no ACLs, so they are
// only public if the bucket policy makes them.
func (bi *Instance) pushToS3(version string) error {
	bucket, _ := bi.s3Location()

	stageDir := filepath.Join(bi.opts.BuildDir, release.GCSStagePath, version)
	releaseKey := bi.s3Key(version)
	logrus.Infof("Pushing release artifacts from %s to %s", stageDir, s3URI(bucket, releaseKey))

	if err := bi.s3.UploadDir(stageDir, bucket, releaseKey); err != nil {
		return fmt.Errorf("push release artifacts to S3: %w", err)
	}

	if !bi.opts.CI {
		logrus.Info("No CI flag set, we're done")

		return nil
	}

	if bi.opts.NoUpdateLatest {
		logrus.Info("Not updating version markers")

		return nil
	}

	markers, err := release.VersionMarkers(
		bi.opts.BuildType, version, bi.opts.ExtraVersionMarkers, bi.opts.Fast,
	)
	if err != nil {
		return fmt.Errorf("get version markers: %w", err)
	}

	logrus.Infof("Publish version markers: %v", markers)

	for _, marker := range markers {
		if err := bi.publishS3VersionMarker(bucket, marker+".txt", version); err != nil {
			return fmt.Errorf("publish version marker %s: %w", marker, err)
		}
	}

	return nil
}

// publishS3VersionMarker updates the version marker file, unless it already
// points to a newer version, like release.Publisher.PublishVersion.
func (bi *Instance) publishS3VersionMarker(bucket, file, version string) error {
	key := bi.s3Key(file)

	exists, err := bi.s3.PathExists(bucket, key)
	if err != nil {
		return fmt.Errorf("check if %s exists: %w", s3URI(bucket, key), err)
	}

	if exists {
		published, err := bi.s3.ReadObject(bucket, key)
		if err != nil {
			return fmt.Errorf("read %s: %w", s3URI(bucket, key), err)
		}

		publishedVersion, err := helpers.TagStringToSemver(strings.TrimSpace(published))
		if err != nil {
			return fmt.Errorf("invalid version %q in %s", published, s3URI(bucket, key))
		}

		newVersion, err := helpers.TagStringToSemver(version)
		if err != nil {
			return fmt.Errorf("invalid version %s", version)
		}

		if release.IsUpToDate(publishedVersion, newVersion) {
			logrus.Infof("Not updating %s, because %s <= %s", file, version, published)

			return nil
		}
	}

	localFile, err := os.CreateTemp("", "version-marker-")
	if err != nil {
		return fmt.Errorf("create version marker file: %w", err)
	}
	defer os.Remove(localFile.Name())

	if _, err := localFile.WriteString(version); err != nil {
		localFile.Close()

		return fmt.Errorf("write %s: %w", localFile.Name(), err)
	}

	if err := localFile.Close(); err != nil {
		return fmt.Errorf("close %s: %w", localFile.Name(), err)
	}

	logrus.Infof("Publishing %s to %s", version, s3URI(bucket, key))

	return bi.s3.UploadVersionMarker(localFile.Name(), bucket, key)
}

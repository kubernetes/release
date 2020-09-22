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
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/release/pkg/gcp"
	"k8s.io/release/pkg/gcp/gcs"
	"k8s.io/release/pkg/http"
	"k8s.io/release/pkg/util"
)

// Publisher is the structure for publishing anything release related
type Publisher struct {
	client publisherClient
}

// NewPublisher creates a new Publisher instance
func NewPublisher() *Publisher {
	return &Publisher{&defaultPublisher{}}
}

// SetClient can be used to set the internal publisher client
func (p *Publisher) SetClient(client publisherClient) {
	p.client = client
}

// publisherClient is a client for working with GCS
//counterfeiter:generate . publisherClient
type publisherClient interface {
	GSUtil(args ...string) error
	GSUtilOutput(args ...string) (string, error)
	GetURLResponse(url string) (string, error)
}

type defaultPublisher struct{}

func (*defaultPublisher) GSUtil(args ...string) error {
	return gcp.GSUtil(args...)
}

func (*defaultPublisher) GSUtilOutput(args ...string) (string, error) {
	return gcp.GSUtilOutput(args...)
}

func (*defaultPublisher) GetURLResponse(url string) (string, error) {
	return http.GetURLResponse(url, true)
}

// Publish a new version, (latest or stable) but only if the files actually
// exist on GCS and the artifacts we're dealing with are newer than the
// contents in GCS.
// buildType - One of 'release' or 'ci'
// version - The version
// buildDir - build output directory
// bucket - GS bucket
func (p *Publisher) PublishVersion(
	buildType, version, buildDir, bucket string,
	versionMarkers []string,
	noMock, privateBucket bool,
) error {
	releaseType := "latest"

	if buildType == "release" {
		// For release/ targets, type should be 'stable'
		if !(strings.HasSuffix(version, ReleaseTypeAlpha) ||
			strings.HasSuffix(version, ReleaseTypeBeta) ||
			strings.HasSuffix(version, ReleaseTypeRC)) {
			releaseType = "stable"
		}
	}

	releasePath := gcs.GcsPrefix + filepath.Join(bucket, buildType, version)
	if err := p.client.GSUtil("ls", releasePath); err != nil {
		return errors.Wrapf(err, "release files dont exist at %s", releasePath)
	}

	sv, err := util.TagStringToSemver(version)
	if err != nil {
		return errors.Errorf("invalid version %s", version)
	}

	publishFiles := append([]string{
		releaseType,
		fmt.Sprintf("%s-%d", releaseType, sv.Major),
		fmt.Sprintf("%s-%d.%d", releaseType, sv.Major, sv.Minor),
	}, versionMarkers...)

	logrus.Infof("Publish version markers: %v", publishFiles)
	logrus.Infof("Publish official pointer text files to bucket %s", bucket)

	for _, file := range publishFiles {
		publishFile := filepath.Join(buildType, file+".txt")
		needsUpdate, err := p.VerifyLatestUpdate(
			publishFile, bucket, version,
		)
		if err != nil {
			return errors.Wrapf(err, "verify latest update for %s", publishFile)
		}
		// If there's a version that's above the one we're trying to release,
		// don't do anything, and just try the next one.
		if !needsUpdate {
			logrus.Infof(
				"Skipping %s for %s because it does not need to be updated",
				publishFile, version,
			)
			continue
		}

		if err := p.PublishToGcs(
			publishFile, buildDir, bucket, version, noMock, privateBucket,
		); err != nil {
			return errors.Wrap(err, "publish release to GCS")
		}
	}

	return nil
}

// VerifyLatestUpdate checks if the new version is greater than the version
// currently published on GCS. It returns `true` for `needsUpdate` if the remote
// version does not exist or needs to be updated.
// publishFile - the GCS location to look in
// bucket - GS bucket
// version - release version
// was releaselib.sh: release::gcs::verify_latest_update
func (p *Publisher) VerifyLatestUpdate(
	publishFile, bucket, version string,
) (needsUpdate bool, err error) {
	logrus.Infof("Testing %s > %s (published)", version, publishFile)

	publishFileDst := gcs.GcsPrefix + filepath.Join(bucket, publishFile)
	gcsVersion, err := p.client.GSUtilOutput("cat", publishFileDst)
	if err != nil {
		logrus.Infof("%s does not exist but will be created", publishFileDst)
		return true, nil
	}

	sv, err := util.TagStringToSemver(version)
	if err != nil {
		return false, errors.Errorf("invalid version format %s", version)
	}

	gcsSemverVersion, err := util.TagStringToSemver(gcsVersion)
	if err != nil {
		return false, errors.Errorf("invalid GCS version format %s", gcsVersion)
	}

	if sv.LTE(gcsSemverVersion) {
		logrus.Infof(
			"Not updating version, because %s <= %s", version, gcsVersion,
		)
		return false, nil
	}

	logrus.Infof("Updating version, because %s > %s", version, gcsVersion)
	return true, nil
}

// PublishToGcs publishes a release to GCS
// publishFile - the GCS location to look in
// buildDir - build output directory
// bucket - GS bucket
// version - release version
// was releaselib.sh: release::gcs::publish
func (p *Publisher) PublishToGcs(
	publishFile, buildDir, bucket, version string,
	noMock, privateBucket bool,
) error {
	releaseStage := filepath.Join(buildDir, ReleaseStagePath)
	publishFileDst := gcs.GcsPrefix + filepath.Join(bucket, publishFile)
	publicLink := filepath.Join(URLPrefixForBucket(bucket), publishFile)
	if bucket == ProductionBucket {
		publicLink = filepath.Join(ProductionBucketURL, publishFile)
	}

	uploadDir := filepath.Join(releaseStage, "upload")
	if err := os.MkdirAll(uploadDir, os.FileMode(0o755)); err != nil {
		return errors.Wrapf(err, "create upload dir %s", uploadDir)
	}

	latestFile := filepath.Join(uploadDir, "latest")
	if err := ioutil.WriteFile(
		latestFile, []byte(version), os.FileMode(0o644),
	); err != nil {
		return errors.Wrap(err, "write latest version file")
	}

	if err := p.client.GSUtil(
		"-m",
		"-h", "Content-Type:text/plain",
		"-h", "Cache-Control:private, max-age=0, no-transform",
		"cp",
		latestFile,
		publishFileDst,
	); err != nil {
		return errors.Wrapf(err, "copy %s to %s", latestFile, publishFileDst)
	}

	var content string
	if noMock && !privateBucket {
		// New Kubernetes infra buckets, like k8s-staging-kubernetes, have a
		// bucket-only ACL policy set, which means attempting to set the ACL on
		// an object will fail. We should skip this ACL change in those
		// instances, as new buckets already default to being publicly
		// readable.
		//
		// Ref:
		// - https://cloud.google.com/storage/docs/bucket-policy-only
		// - https://github.com/kubernetes/release/issues/904
		if strings.HasPrefix(bucket, "k8s-") {
			aclOutput, err := p.client.GSUtilOutput(
				"acl", "ch", "-R", "-g", "all:R", publishFileDst,
			)
			if err != nil {
				return errors.Wrapf(err, "change %s permissions", publishFileDst)
			}
			logrus.Infof("Making uploaded version file public: %s", aclOutput)
		}

		// If public, validate public link
		response, err := p.client.GetURLResponse(publicLink)
		if err != nil {
			return errors.Wrapf(err, "get content of %s", publicLink)
		}
		content = response
	} else {
		response, err := p.client.GSUtilOutput("cat", publicLink)
		if err != nil {
			return errors.Wrapf(err, "get content of %s", publicLink)
		}
		content = response
	}

	logrus.Infof("Validating uploaded version file at %s", publicLink)
	if version != content {
		return errors.Errorf(
			"version %s it not equal response %s",
			version, content,
		)
	}

	logrus.Info("Version equals response")
	return nil
}

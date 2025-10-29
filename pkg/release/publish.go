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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-sdk/gcli"
	"sigs.k8s.io/release-sdk/object"
	"sigs.k8s.io/release-utils/helpers"
	"sigs.k8s.io/release-utils/http"
)

// Publisher is the structure for publishing anything release related.
type Publisher struct {
	client publisherClient
}

// NewPublisher creates a new Publisher instance.
func NewPublisher() *Publisher {
	objStore := *object.NewGCS()
	objStore.SetOptions(
		objStore.WithNoClobber(false),
	)

	return &Publisher{
		client: &defaultPublisher{&objStore},
	}
}

// SetClient can be used to set the internal publisher client.
func (p *Publisher) SetClient(client publisherClient) {
	p.client = client
}

// publisherClient is a client for working with GCS
//
//counterfeiter:generate . publisherClient
type publisherClient interface {
	GSUtil(args ...string) error
	GSUtilOutput(args ...string) (string, error)
	GSUtilStatus(args ...string) (bool, error)
	GetURLResponse(url string) (string, error)
	GetReleasePath(bucket, gcsRoot, version string, fast bool) (string, error)
	GetMarkerPath(bucket, gcsRoot string, fast bool) (string, error)
	NormalizePath(pathParts ...string) (string, error)
	TempDir(dir, pattern string) (name string, err error)
	CopyToLocal(remote, local string) error
	ReadFile(filename string) ([]byte, error)
	Unmarshal(data []byte, v any) error
	Marshal(v any) ([]byte, error)
	TempFile(dir, pattern string) (f *os.File, err error)
	CopyToRemote(local, remote string) error
}

type defaultPublisher struct {
	objStore object.Store
}

func (*defaultPublisher) GSUtil(args ...string) error {
	return gcli.GSUtil(args...)
}

func (*defaultPublisher) GSUtilOutput(args ...string) (string, error) {
	return gcli.GSUtilOutput(args...)
}

func (*defaultPublisher) GSUtilStatus(args ...string) (bool, error) {
	status, err := gcli.GSUtilStatus(args...)
	if err != nil {
		return false, err
	}

	return status.Success(), nil
}

func (*defaultPublisher) GetURLResponse(url string) (string, error) {
	c, err := http.NewAgent().Get(url)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(c)), nil
}

func (d *defaultPublisher) GetReleasePath(
	bucket, gcsRoot, version string, fast bool,
) (string, error) {
	return d.objStore.GetReleasePath(bucket, gcsRoot, version, fast)
}

func (d *defaultPublisher) GetMarkerPath(
	bucket, gcsRoot string, fast bool,
) (string, error) {
	return d.objStore.GetMarkerPath(bucket, gcsRoot, fast)
}

func (d *defaultPublisher) NormalizePath(pathParts ...string) (string, error) {
	return d.objStore.NormalizePath(pathParts...)
}

func (*defaultPublisher) TempDir(dir, pattern string) (name string, err error) {
	return os.MkdirTemp(dir, pattern)
}

func (d *defaultPublisher) CopyToLocal(remote, local string) error {
	return d.objStore.CopyToLocal(remote, local)
}

func (*defaultPublisher) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (*defaultPublisher) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (*defaultPublisher) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (*defaultPublisher) TempFile(dir, pattern string) (f *os.File, err error) {
	return os.CreateTemp(dir, pattern)
}

func (d *defaultPublisher) CopyToRemote(local, remote string) error {
	return d.objStore.CopyToRemote(local, remote)
}

// Publish a new version, (latest or stable) but only if the files actually
// exist on GCS and the artifacts we're dealing with are newer than the
// contents in GCS.
// buildType - One of 'release' or 'ci'
// version - The version
// buildDir - build output directory
// bucket - GCS bucket
// gcsRoot - The top-level GCS directory builds will be released to
//
// Expected destination format:
//
//	gs://<bucket>/<gcsRoot>[/fast]/<version>
func (p *Publisher) PublishVersion(
	buildType, version, buildDir, bucket, gcsRoot string,
	extraVersionMarkers []string,
	privateBucket, fast bool,
) error {
	logrus.Info("Publishing version")

	releaseType := "latest"

	if buildType == "release" {
		// For release/ targets, type should be 'stable'
		if !strings.Contains(version, ReleaseTypeAlpha) && !strings.Contains(version, ReleaseTypeBeta) && !strings.Contains(version, ReleaseTypeRC) {
			releaseType = "stable"
		}
	}

	sv, err := helpers.TagStringToSemver(version)
	if err != nil {
		return fmt.Errorf("invalid version %s", version)
	}

	markerPath, markerPathErr := p.client.GetMarkerPath(
		bucket,
		gcsRoot,
		fast,
	)
	if markerPathErr != nil {
		return fmt.Errorf("get version marker path: %w", markerPathErr)
	}

	releasePath, releasePathErr := p.client.GetReleasePath(
		bucket,
		gcsRoot,
		version,
		fast,
	)
	if releasePathErr != nil {
		return fmt.Errorf("get release path: %w", releasePathErr)
	}

	// TODO: This should probably be a more thorough check of explicit files
	// TODO: This should explicitly do a `gsutil ls` via gcs.PathExists
	if err := p.client.GSUtil("ls", releasePath); err != nil {
		return fmt.Errorf("release files don't exist at %s: %w", releasePath, err)
	}

	var versionMarkers []string
	if fast {
		versionMarkers = append(
			versionMarkers,
			releaseType+"-fast",
		)
	} else {
		versionMarkers = append(
			versionMarkers,
			releaseType,
			fmt.Sprintf("%s-%d", releaseType, sv.Major),
			fmt.Sprintf("%s-%d.%d", releaseType, sv.Major, sv.Minor),
		)
	}

	if len(extraVersionMarkers) > 0 {
		versionMarkers = append(versionMarkers, extraVersionMarkers...)
	}

	logrus.Infof("Publish version markers: %v", versionMarkers)
	logrus.Infof("Publish official pointer text files to %s", markerPath)

	for _, file := range versionMarkers {
		versionMarker := file + ".txt"

		needsUpdate, err := p.VerifyLatestUpdate(
			versionMarker, markerPath, version,
		)
		if err != nil {
			return fmt.Errorf("verify latest update for %s: %w", versionMarker, err)
		}

		// If there's a version that's above the one we're trying to release,
		// don't do anything, and just try the next one.
		if !needsUpdate {
			logrus.Infof(
				"Skipping %s for %s because it does not need to be updated",
				versionMarker, version,
			)

			continue
		}

		if err := p.PublishToGcs(
			versionMarker, buildDir, markerPath, version, privateBucket,
		); err != nil {
			return fmt.Errorf("publish release to GCS: %w", err)
		}
	}

	return nil
}

// VerifyLatestUpdate checks if the new version is greater than the version
// currently published on GCS. It returns `true` for `needsUpdate` if the remote
// version does not exist or needs to be updated.
// publishFile - the version marker to look for
// markerPath - the GCS path to search for the version marker in
// version - release version.
func (p *Publisher) VerifyLatestUpdate(
	publishFile, markerPath, version string,
) (needsUpdate bool, err error) {
	logrus.Infof("Testing %s > %s (published)", version, publishFile)

	publishFileDst, publishFileDstErr := p.client.NormalizePath(markerPath, publishFile)
	if publishFileDstErr != nil {
		return false, fmt.Errorf("get marker file destination: %w", publishFileDstErr)
	}

	// TODO: Should we add a object.`GCS` method for `gsutil cat`?
	gcsVersion, err := p.client.GSUtilOutput("cat", publishFileDst)
	if err != nil {
		logrus.Infof("%s does not exist but will be created", publishFileDst)
		//nolint:nilerr // returning nil is intentional
		return true, nil
	}

	sv, err := helpers.TagStringToSemver(version)
	if err != nil {
		return false, fmt.Errorf("invalid version format %s", version)
	}

	gcsSemverVersion, err := helpers.TagStringToSemver(gcsVersion)
	if err != nil {
		return false, fmt.Errorf("invalid GCS version format %s", gcsVersion)
	}

	if IsUpToDate(gcsSemverVersion, sv) {
		logrus.Infof(
			"Not updating version, because %s <= %s", version, gcsVersion,
		)

		return false, nil
	}

	logrus.Infof("Updating version, because %s > %s", version, gcsVersion)

	return true, nil
}

func IsUpToDate(oldVersion, newVersion semver.Version) bool {
	oldPre := oldVersion.Pre
	newPre := newVersion.Pre

	oldStrippedPre := semver.Version{Major: oldVersion.Major, Minor: oldVersion.Minor, Patch: oldVersion.Patch}
	newStrippedPre := semver.Version{Major: newVersion.Major, Minor: newVersion.Minor, Patch: newVersion.Patch}

	// Verfy specific use case in our tagging logic:
	// 1.30.0-rc.2.10+00000000000000
	// needs to be considered lower than
	// 1.30.0-11+00000000000000
	// which is not given by newVersion.LTE(oldVersion) below.
	if len(oldPre) == 3 && // [rc 2 10]
		oldPre[0].String() == "rc" &&
		len(newPre) == 1 && // [11]
		newPre[0].IsNum &&
		// For example when:
		// oldVersion: 1.29.0-rc.0.20+00000000000000
		// newVersion: 1.28.1-2+00000000000000
		newStrippedPre.GE(oldStrippedPre) {
		return false
	}

	return newVersion.LTE(oldVersion)
}

// PublishToGcs publishes a release to GCS
// publishFile - the GCS location to look in
// buildDir - build output directory
// markerPath - the GCS path to publish a version marker to
// version - release version.
func (p *Publisher) PublishToGcs(
	publishFile, buildDir, markerPath, version string,
	privateBucket bool,
) error {
	releaseStage := filepath.Join(buildDir, ReleaseStagePath)

	publishFileDst, publishFileDstErr := p.client.NormalizePath(markerPath, publishFile)
	if publishFileDstErr != nil {
		return fmt.Errorf("get marker file destination: %w", publishFileDstErr)
	}

	logrus.Infof("Using marker path: %s", markerPath)

	publicLink := fmt.Sprintf("%s/%s", URLPrefixForBucket(markerPath), publishFile)
	if strings.HasSuffix(markerPath, ProductionBucket+"/release") {
		publicLink = fmt.Sprintf("%s/release/%s", ProductionBucketURL, publishFile)
	}

	logrus.Infof("Using public link: %s", publicLink)

	uploadDir := filepath.Join(releaseStage, "upload")
	if err := os.MkdirAll(uploadDir, os.FileMode(0o755)); err != nil {
		return fmt.Errorf("create upload dir %s: %w", uploadDir, err)
	}

	latestFile := filepath.Join(uploadDir, "latest")
	if err := os.WriteFile(
		latestFile, []byte(version), os.FileMode(0o644),
	); err != nil {
		return fmt.Errorf("write latest version file: %w", err)
	}

	logrus.Infof("Running `gsutil cp` from %s to: %s", latestFile, publishFileDst)

	if err := p.client.GSUtil(
		"-m",
		"-h", "Content-Type:text/plain",
		"-h", "Cache-Control:private, max-age=0, no-transform",
		"cp",
		latestFile,
		publishFileDst,
	); err != nil {
		return fmt.Errorf("copy %s to %s: %w", latestFile, publishFileDst, err)
	}

	var content string

	if !privateBucket {
		// If public, validate public link
		logrus.Infof("Validating uploaded version file using HTTP at %s", publicLink)

		response, err := p.client.GetURLResponse(publicLink)
		if err != nil {
			return fmt.Errorf("get content of %s: %w", publicLink, err)
		}

		content = response
	} else {
		// Use the private location
		logrus.Infof("Validating uploaded version file using `gsutil cat` at %s", publishFileDst)

		response, err := p.client.GSUtilOutput("cat", publishFileDst)
		if err != nil {
			return fmt.Errorf("get content of %s: %w", publishFileDst, err)
		}

		content = response
	}

	if version != content {
		return fmt.Errorf(
			"version %s it not equal response %s",
			version, content,
		)
	}

	logrus.Info("Version equals response")

	return nil
}

func FixPublicReleaseNotesURL(gcsPath string) string {
	const prefix = "gs://" + ProductionBucket
	if after, ok := strings.CutPrefix(gcsPath, prefix); ok {
		gcsPath = ProductionBucketURL + after
	}

	return gcsPath
}

// PublishReleaseNotesIndex updates or creates the release notes index JSON at
// the target `gcsIndexRootPath`.
func (p *Publisher) PublishReleaseNotesIndex(
	gcsIndexRootPath, gcsReleaseNotesPath, version string,
) error {
	logrus.Infof("Using index root path: %s", gcsIndexRootPath)
	logrus.Infof("Using GCS release notes path: %s", gcsReleaseNotesPath)

	const releaseNotesIndex = "/release-notes-index.json"

	indexFilePath, err := p.client.NormalizePath(
		gcsIndexRootPath, releaseNotesIndex,
	)
	if err != nil {
		return fmt.Errorf("normalize index file: %w", err)
	}

	logrus.Infof("Publishing release notes index %s", indexFilePath)

	success, err := p.client.GSUtilStatus("-q", "stat", indexFilePath)
	if err != nil {
		return fmt.Errorf("run gcsutil stat: %w", err)
	}

	logrus.Info("Building release notes index")

	versions := make(map[string]string)

	if success {
		logrus.Info("Modifying existing release notes index file")

		tempDir, err := p.client.TempDir("", "release-notes-index-")
		if err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}

		defer os.RemoveAll(tempDir)

		tempIndexFile := filepath.Join(tempDir, releaseNotesIndex)

		if err := p.client.CopyToLocal(
			indexFilePath, tempIndexFile,
		); err != nil {
			return fmt.Errorf("copy index file to local: %w", err)
		}

		indexBytes, err := p.client.ReadFile(tempIndexFile)
		if err != nil {
			return fmt.Errorf("read local index file: %w", err)
		}

		if err := p.client.Unmarshal(indexBytes, &versions); err != nil {
			return fmt.Errorf("unmarshal versions: %w", err)
		}
	} else {
		logrus.Info("Creating non existing release notes index file")
	}

	versions[version] = gcsReleaseNotesPath

	// Fixup the index to only use public URLS
	for v, releaseNotesPath := range versions {
		versions[v] = FixPublicReleaseNotesURL(releaseNotesPath)
	}

	versionJSON, err := p.client.Marshal(versions)
	if err != nil {
		return fmt.Errorf("marshal version JSON: %w", err)
	}

	logrus.Infof("Writing new release notes index: %s", string(versionJSON))

	tempFile, err := p.client.TempFile("", "release-notes-index-")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(versionJSON); err != nil {
		return fmt.Errorf("write temp index: %w", err)
	}

	logrus.Info("Uploading release notes index")

	if err := p.client.CopyToRemote(
		tempFile.Name(), indexFilePath,
	); err != nil {
		return fmt.Errorf("upload index file: %w", err)
	}

	return nil
}

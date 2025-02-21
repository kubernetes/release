/*
Copyright 2021 The Kubernetes Authors.

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

package cve

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"sigs.k8s.io/release-sdk/object"

	"k8s.io/release/pkg/notes"
)

//counterfeiter:generate . ClientImplementation
type ClientImplementation interface {
	CheckBucketPath(string, *ClientOptions) error
	CheckBucketWriteAccess(*ClientOptions) error
	DeleteFile(string, *ClientOptions) error
	CopyFile(string, string, *ClientOptions) error
	CheckID(string) error
	CopyToTemp(string, *ClientOptions) (*os.File, error)
	ValidateCVEMap(string, string, *ClientOptions) error
	CreateEmptyFile(string, *ClientOptions) (*os.File, error)
	EntryExists(string, *ClientOptions) (bool, error)
}

// defaultClientImplementation.
type defaultClientImplementation struct{}

// CheckBucketWriteAccess verifies if the current user has writeaccess to the bucket
// adapted from the build pkg.
func (impl *defaultClientImplementation) CheckBucketWriteAccess(opts *ClientOptions) error {
	logrus.Infof("Checking bucket %s for write permissions", opts.Bucket)

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return fmt.Errorf(
			"fetching gcloud credentials, try running "+
				`"gcloud auth application-default login: %w"`,
			err,
		)
	}

	bucket := client.Bucket(opts.Bucket)
	if bucket == nil {
		return fmt.Errorf(
			"unable to open CVE bucket: %s", opts.Bucket,
		)
	}

	// Check if bucket exists and user has permissions
	requiredGCSPerms := []string{"storage.objects.create"}

	perms, err := bucket.IAM().TestPermissions(
		context.Background(), requiredGCSPerms,
	)
	if err != nil {
		return fmt.Errorf("getting bucket permissions: %w", err)
	}

	if len(perms) != 1 {
		return fmt.Errorf(
			"GCP user must have at least %s permissions on bucket %s",
			requiredGCSPerms, opts.Bucket,
		)
	}

	return nil
}

// Delete file erases a file from the CVE bucket location.
func (impl *defaultClientImplementation) DeleteFile(
	path string, opts *ClientOptions,
) error {
	if err := impl.CheckBucketWriteAccess(opts); err != nil {
		return fmt.Errorf("checking bucket permissions to delete data: %w", err)
	}

	gcs := object.NewGCS()

	path, err := gcs.NormalizePath(path)
	if err != nil {
		return fmt.Errorf("normalizing bucket path: %w", err)
	}

	if err := impl.CheckBucketPath(path, opts); err != nil {
		return fmt.Errorf("checking path to delete file: %w", err)
	}

	if !strings.HasSuffix(path, ".yaml") {
		return errors.New("only yaml files can be deleted")
	}

	exists, err := gcs.PathExists(path)
	if err != nil {
		return fmt.Errorf("checking if cve entry exists: %w", err)
	}

	if !exists {
		return errors.New("specified CVE entry not found")
	}

	return gcs.DeletePath(path)
}

// CopyToTemp copies a CVE map file into a temporary file for editing.
func (impl *defaultClientImplementation) CopyToTemp(
	cve string, opts *ClientOptions,
) (*os.File, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "cve-maps-")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}

	gcs := object.NewGCS()
	if err := gcs.CopyToLocal(
		object.GcsPrefix+filepath.Join(
			opts.Bucket, opts.Directory, cve+mapExt,
		), dir,
	); err != nil {
		return nil, fmt.Errorf("copying CVE %s to tempfile: %w", cve, err)
	}

	return os.Open(filepath.Join(dir, cve+mapExt))
}

// CopyFile copies a file into the CVE location in the bucket.
func (impl *defaultClientImplementation) CopyFile(
	src, dest string, opts *ClientOptions,
) error {
	if err := impl.CheckBucketWriteAccess(opts); err != nil {
		return fmt.Errorf("checking bucket permissions to copy data: %w", err)
	}

	gcs := object.NewGCS()
	gcs.SetOptions(
		gcs.WithNoClobber(false),
	)

	path, err := gcs.NormalizePath(dest)
	if err != nil {
		return fmt.Errorf("normalizing bucket path: %w", err)
	}

	if err := impl.CheckBucketPath(path, opts); err != nil {
		return fmt.Errorf("checking path to copy file: %w", err)
	}

	if err := gcs.CopyToRemote(src, path); err != nil {
		return fmt.Errorf("copying %s to bucket: %w", path, err)
	}

	// Copy the file to the bucket
	return nil
}

// CheckBucketPath checks if a path is inside the cve location.
func (impl *defaultClientImplementation) CheckBucketPath(
	path string, opts *ClientOptions,
) error {
	g := object.NewGCS()

	path, err := g.NormalizePath(path)
	if err != nil {
		return fmt.Errorf("normalizing CVE bucket path: %w", err)
	}

	path = strings.TrimPrefix(path, object.GcsPrefix)

	if !strings.HasPrefix(path, filepath.Join(opts.Bucket, opts.Directory)) {
		return errors.New("invalid path, all paths must be in the cve location")
	}

	return nil
}

// CheckID checks if a string is a weel formed CVE identifier.
func (impl *defaultClientImplementation) CheckID(cveID string) error {
	if regexp.MustCompile(CVEIDRegExp).MatchString(cveID) {
		return nil
	}

	return errors.New("invalid CVE identifier")
}

// ValidateCVEData checks a cve map.
func (impl *defaultClientImplementation) ValidateCVEMap(
	cveID, path string, _ *ClientOptions,
) (err error) {
	// Parse the data map
	maps, err := notes.ParseReleaseNotesMap(path)
	if err != nil {
		return fmt.Errorf("parsing CVE data map: %w", err)
	}

	// Cycle all data maps in file
	for i, dataMap := range *maps {
		// Check if map has other the CVE field
		if _, ok := dataMap.DataFields["cve"]; !ok {
			return fmt.Errorf("data map #%d in file %s has no CVE data", i, path)
		}
		// Cast the datafield as CVE data
		cvedata := CVE{}
		if err := cvedata.ReadRawInterface(dataMap.DataFields["cve"]); err != nil {
			return fmt.Errorf("reading CVE data from YAML file: %w", err)
		}

		if err := cvedata.Validate(); err != nil {
			return fmt.Errorf("validating map #%d in file %s: %w", i, path, err)
		}

		if cvedata.ID != cveID {
			return fmt.Errorf(
				"CVE ID in map #%d in file %s does not match %s",
				i,
				path,
				cveID,
			)
		}
	}

	return nil
}

// CreateEmptyFile creates an empty CVE map.
func (impl *defaultClientImplementation) CreateEmptyFile(cve string, _ *ClientOptions) (
	file *os.File, err error,
) {
	if err := impl.CheckID(cve); err != nil {
		return nil, fmt.Errorf("checking new CVE ID: %w", err)
	}

	// Add a relnote-compatible struct with only the CVE data
	noteMap := struct {
		PR         int
		DataFields map[string]*CVE
	}{
		PR: 0,
		DataFields: map[string]*CVE{
			"cve": {
				ID: cve,
			},
		},
	}

	// Marshall the data struct into yaml
	yamlCode, err := yaml.Marshal(noteMap)
	if err != nil {
		return nil, fmt.Errorf("marshalling CVE data map: %w", err)
	}

	file, err = os.CreateTemp(os.TempDir(), "cve-data-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("creating new map file: %w", err)
	}

	if _, err := file.WriteString(newMapHeader); err != nil {
		return nil, fmt.Errorf("writing empty CVE header: %w", err)
	}

	if _, err := file.Write(yamlCode); err != nil {
		return nil, fmt.Errorf("writing yaml code to file: %w", err)
	}

	return file, nil
}

// EntryExists returns true if a CVE already exists.
func (impl *defaultClientImplementation) EntryExists(
	cveID string, opts *ClientOptions,
) (exists bool, err error) {
	// Check the ID string to be valid
	if err := ValidateID(cveID); err != nil {
		return exists, fmt.Errorf("checking CVE ID string: %w", err)
	}

	// Verify the expected file exists in the bucket
	gcs := object.NewGCS()
	// Normalizar the path to the CVE
	path, err := gcs.NormalizePath(
		object.GcsPrefix + filepath.Join(
			opts.Bucket, opts.Directory, cveID+mapExt,
		),
	)
	if err != nil {
		return exists, fmt.Errorf("checking if CVE entry already exists: %w", err)
	}

	return gcs.PathExists(path)
}

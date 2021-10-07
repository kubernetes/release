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

package release

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/release/pkg/provenance"
	"sigs.k8s.io/release-sdk/object"
	"sigs.k8s.io/release-utils/util"
)

func NewProvenanceChecker(opts *ProvenanceCheckerOptions) *ProvenanceChecker {
	p := &ProvenanceChecker{
		objStore: object.NewGCS(),
		options:  opts,
	}
	p.objStore.WithConcurrent(true)
	p.objStore.WithRecursive(true)
	p.impl = &defaultProvenanceCheckerImpl{}
	return p
}

// ProvenanceChecker
type ProvenanceChecker struct {
	objStore *object.GCS
	options  *ProvenanceCheckerOptions
	impl     provenanceCheckerImplementation
}

// CheckStageProvenance
func (pc *ProvenanceChecker) CheckStageProvenance(buildVersion string) error {
	// Init the local dir
	h := sha1.New()
	if _, err := h.Write([]byte(buildVersion)); err != nil {
		return errors.Wrap(err, "creating dir")
	}
	pc.options.StageDirectory = filepath.Join(pc.options.ScratchDirectory, fmt.Sprintf("%x", h.Sum(nil)))

	gcsPath, err := pc.objStore.NormalizePath(
		object.GcsPrefix + filepath.Join(
			pc.options.StageBucket, StagePath, buildVersion,
		) + string(filepath.Separator),
	)
	if err != nil {
		return errors.Wrap(err, "normalizing GCS stage path")
	}
	// Download all the artifacts from the bucket
	if err := pc.impl.downloadStagedArtifacts(pc.options, pc.objStore, gcsPath); err != nil {
		return errors.Wrap(err, "downloading staged artifacts")
	}

	// Preprocess the attestation file. We have to rewrite the paths
	// to strip the GCS prefix
	statement, err := pc.impl.processAttestation(pc.options, buildVersion)
	if err != nil {
		return errors.Wrap(err, "processing provenance attestation")
	}

	// Run the check of the artifacts
	return pc.impl.checkProvenance(pc.options, statement)
}

type ProvenanceCheckerOptions struct {
	StageBucket      string // Bucket where the artifacts are stored
	StageDirectory   string // Directory where artifacts will be downloaded
	ScratchDirectory string // Directory where StageDirectory will be created
}

type provenanceCheckerImplementation interface {
	downloadStagedArtifacts(*ProvenanceCheckerOptions, *object.GCS, string) error
	processAttestation(*ProvenanceCheckerOptions, string) (*provenance.Statement, error)
	checkProvenance(*ProvenanceCheckerOptions, *provenance.Statement) error
}

type defaultProvenanceCheckerImpl struct{}

// downloadReleaseArtifacts sybc
func (di *defaultProvenanceCheckerImpl) downloadStagedArtifacts(
	opts *ProvenanceCheckerOptions, objStore *object.GCS, path string,
) error {
	logrus.Infof("Synching stage from %s to %s", path, opts.StageDirectory)
	if !util.Exists(opts.StageDirectory) {
		if err := os.Mkdir(opts.StageDirectory, os.FileMode(0o755)); err != nil {
			return errors.Wrap(err, "creating local working directory")
		}
	}
	return errors.Wrap(
		objStore.CopyToLocal(path, opts.StageDirectory),
		"synching staged sources",
	)
}

// processAttestation
func (di *defaultProvenanceCheckerImpl) processAttestation(
	opts *ProvenanceCheckerOptions, buildVersion string) (s *provenance.Statement, err error) {
	// Load the downloaded statement
	s, err = provenance.LoadStatement(filepath.Join(opts.StageDirectory, buildVersion, ProvenanceFilename))
	if err != nil {
		return nil, errors.Wrap(err, "loading staging provenance file")
	}

	// We've downloaded all artifacts, so to check we need to strip
	// the gcs bucket prefix from the subjects to read from the local copy
	gcsPath := object.GcsPrefix + filepath.Join(opts.StageBucket, StagePath)

	newSubjects := []intoto.Subject{}

	for i, sub := range s.Subject {
		newSubjects = append(newSubjects, intoto.Subject{
			Name:   strings.TrimPrefix(sub.Name, gcsPath),
			Digest: sub.Digest,
		})
		s.Subject[i].Name = strings.TrimPrefix(sub.Name, gcsPath)
	}
	s.Subject = newSubjects
	return s, nil
}

func (di *defaultProvenanceCheckerImpl) checkProvenance(
	opts *ProvenanceCheckerOptions, s *provenance.Statement) error {
	return errors.Wrap(s.VerifySubjects(opts.StageDirectory), "checking subjects in attestation")
}

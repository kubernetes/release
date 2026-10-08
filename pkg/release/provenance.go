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
	"context"
	"crypto/sha1" //nolint:gosec // used for file integrity checks, NOT security
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	intoto "github.com/in-toto/attestation/go/v1"
	plattestation "github.com/policylabs/attestation"
	"github.com/policylabs/collector/envelope/bundle"
	sapi "github.com/policylabs/signer/api/v1"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	"github.com/sirupsen/logrus"
	"github.com/slsa-framework/verifier/pkg/slsa"
	"google.golang.org/protobuf/encoding/protojson"

	"sigs.k8s.io/bom/pkg/bom"
	"sigs.k8s.io/bom/pkg/provenance"
	"sigs.k8s.io/bom/pkg/spdx"
	"sigs.k8s.io/release-sdk/object"
	"sigs.k8s.io/release-utils/hash"
	"sigs.k8s.io/release-utils/helpers"

	"k8s.io/release/pkg/attestation"
)

const (
	// StageProvenanceSigner is the identity that signs the provenance of a
	// stage run, see the attestation step of gcb/stage/cloudbuild.yaml.
	StageProvenanceSigner = "sigstore::https://accounts.google.com::krel-staging@k8s-releng-prod.iam.gserviceaccount.com"

	// StageProvenanceBuilder is the builder ID of the provenance of a stage
	// run.
	StageProvenanceBuilder = "https://git.k8s.io/release/docs/krel"
)

// ProvenanceChecker is the main structure to check the provenance.
type ProvenanceChecker struct {
	objStore *object.GCS
	options  *ProvenanceCheckerOptions
	impl     provenanceCheckerImplementation
}

// NewProvenanceChecker returns a new ProvenanceChecker instance.
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

// CheckStageProvenance validates the provenance for the provided build version.
func (pc *ProvenanceChecker) CheckStageProvenance(buildVersion string) error {
	//nolint:gosec // used for file integrity checks, NOT security
	// Init the local dir
	h := sha1.New()
	if _, err := h.Write([]byte(buildVersion)); err != nil {
		return fmt.Errorf("creating dir: %w", err)
	}

	pc.options.StageDirectory = filepath.Join(pc.options.ScratchDirectory, hex.EncodeToString(h.Sum(nil)))

	gcsPath, err := pc.objStore.NormalizePath(
		object.GcsPrefix + filepath.Join(
			pc.options.StageBucket, StagePath, buildVersion,
		) + string(filepath.Separator),
	)
	if err != nil {
		return fmt.Errorf("normalizing GCS stage path: %w", err)
	}
	// Download all the artifacts from the bucket
	if err := pc.impl.downloadStagedArtifacts(pc.options, pc.objStore, gcsPath); err != nil {
		return fmt.Errorf("downloading staged artifacts: %w", err)
	}

	// Preprocess the attestation file. We have to rewrite the paths
	// to strip the GCS prefix
	statement, err := pc.impl.processAttestation(pc.options, buildVersion)
	if err != nil {
		return fmt.Errorf("processing provenance attestation: %w", err)
	}

	// Run the check of the artifacts
	if err := pc.impl.checkProvenance(pc.options, statement); err != nil {
		return fmt.Errorf("verifying provenance of staged artifacts: %w", err)
	}

	logrus.Infof(
		"Successfully verified provenance information of %d staged artifacts",
		len(statement.GetSubject()),
	)

	return nil
}

// GenerateFinalAttestation combines the stage provenance attestation
// with a release sbom to create the end-user provenance atteatation.
func (pc *ProvenanceChecker) GenerateFinalAttestation(buildVersion string, versions *Versions) error {
	statementPath := filepath.Join(pc.options.StageDirectory, buildVersion, ProvenanceFilename)
	for _, version := range versions.Ordered() {
		if err := pc.impl.generateFinalAttestation(
			pc.options,
			filepath.Join(
				pc.options.StageDirectory, buildVersion, version, GCSStagePath, version, "kubernetes-release.spdx",
			),
			statementPath, version,
		); err != nil {
			return fmt.Errorf("generating provenance data for %s: %w", version, err)
		}
	}

	return nil
}

type ProvenanceCheckerOptions struct {
	StageBucket      string // Bucket where the artifacts are stored
	StageDirectory   string // Directory where artifacts will be downloaded
	ScratchDirectory string // Directory where StageDirectory will be created
}

type provenanceCheckerImplementation interface {
	downloadStagedArtifacts(*ProvenanceCheckerOptions, *object.GCS, string) error
	processAttestation(*ProvenanceCheckerOptions, string) (*intoto.Statement, error)
	checkProvenance(*ProvenanceCheckerOptions, *intoto.Statement) error
	generateFinalAttestation(opts *ProvenanceCheckerOptions, sbom, stageProvenance, version string) error
}

type defaultProvenanceCheckerImpl struct {
	// verifySignatures verifies the signatures of the stage provenance and
	// records the outcome in it, against the sigstore public good instance
	// when nil.
	verifySignatures func(plattestation.Envelope) error
}

// downloadReleaseArtifacts sybc.
func (di *defaultProvenanceCheckerImpl) downloadStagedArtifacts(
	opts *ProvenanceCheckerOptions, objStore *object.GCS, path string,
) error {
	logrus.Infof("Synching stage from %s to %s", path, opts.StageDirectory)

	if !helpers.Exists(opts.StageDirectory) {
		if err := os.MkdirAll(opts.StageDirectory, os.FileMode(0o755)); err != nil {
			return fmt.Errorf("creating local working directory: %w", err)
		}
	}

	if err := objStore.CopyToLocal(path, opts.StageDirectory); err != nil {
		return fmt.Errorf("synching staged sources: %w", err)
	}

	return nil
}

// processAttestation reads the SLSA attestation generated during the stage
// run, verifies it, see verifyStageProvenance, and returns its statement,
// whose subjects are the staged artifacts to check.
func (di *defaultProvenanceCheckerImpl) processAttestation(
	opts *ProvenanceCheckerOptions, buildVersion string,
) (*intoto.Statement, error) {
	// Load the downloaded statement
	path := filepath.Join(opts.StageDirectory, buildVersion, ProvenanceFilename)

	bundleData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading staging provenance file: %w", err)
	}

	if err := di.verifyStageProvenance(bundleData, stageProvenanceSource()); err != nil {
		return nil, fmt.Errorf("verifying staging provenance file: %w", err)
	}

	data, err := stageStatement(bundleData)
	if err != nil {
		return nil, fmt.Errorf("unwrapping staging provenance file: %w", err)
	}

	s := &intoto.Statement{}
	if err := protojson.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parsing staging provenance file: %w", err)
	}

	// We've downloaded all artifacts, so to check we need to strip
	// the gcs bucket prefix from the subjects to read from the local copy
	gcsPath := object.GcsPrefix + filepath.Join(opts.StageBucket, StagePath)

	for _, sub := range s.GetSubject() {
		sub.Name = strings.TrimPrefix(sub.GetName(), gcsPath)
	}

	return s, nil
}

// stageProvenanceSource is the repository the stage run builds from, which
// is a fork in mock runs with K8S_ORG or K8S_REPO set.
func stageProvenanceSource() string {
	return "github.com/" + GetK8sOrg() + "/" + GetK8sRepo()
}

// verifyStageProvenance verifies that the provenance of the stage run is a
// sigstore bundle signed by StageProvenanceSigner, about a build of the
// source repository by StageProvenanceBuilder, with the SLSA verifier.
func (di *defaultProvenanceCheckerImpl) verifyStageProvenance(data []byte, source string) error {
	if !attestation.IsSigstoreBundle(data) {
		return errors.New("not a sigstore bundle, so it is not signed")
	}

	envelopes, err := (&bundle.Parser{}).Parse(data)
	if err != nil {
		return fmt.Errorf("parsing sigstore bundle: %w", err)
	}

	if len(envelopes) != 1 {
		return fmt.Errorf("expected one statement in the sigstore bundle, found %d", len(envelopes))
	}

	verify := di.verifySignatures
	if verify == nil {
		verify = func(envelope plattestation.Envelope) error {
			return envelope.Verify() //nolint:wrapcheck // wrapped by the caller
		}
	}

	if err := verify(envelopes[0]); err != nil {
		return fmt.Errorf("verifying signatures: %w", err)
	}

	return checkStageStatement(envelopes[0].GetStatement(), source)
}

// checkStageStatement checks a stage provenance statement whose signatures
// are verified against the signer, builder and source of a stage run. The
// verifier checks the builder only from SLSA build level 2 on, where a
// trusted builder has to be bound to its signer, so that is the level the
// check requires. The stage run still meets level 1 only, since it signs its
// provenance itself.
func checkStageStatement(statement plattestation.Statement, source string) error {
	if statement == nil {
		return errors.New("the sigstore bundle holds no in-toto statement")
	}

	signer, err := sapi.NewIdentityFromSpec(StageProvenanceSigner)
	if err != nil {
		return fmt.Errorf("parsing signer identity: %w", err)
	}

	verifier, err := slsa.New()
	if err != nil {
		return fmt.Errorf("creating SLSA verifier: %w", err)
	}

	res, err := verifier.Verify(context.Background(), statement,
		slsa.WithRequireSignatures(true),
		slsa.WithExpectedSigners([]*sapi.Identity{signer}),
		slsa.WithParam("trusted_builders", []string{StageProvenanceBuilder}),
		slsa.WithParam("expected_source", source),
		slsa.WithMinLevel(2),
		slsa.WithSkipBuildTypeChecks(true),
	)
	if err == nil && !res.Pass() {
		err = errors.New(failureReason(res))
	}

	if err != nil {
		return fmt.Errorf(
			"expected signer %s, builder %s and source %s: %w",
			StageProvenanceSigner, StageProvenanceBuilder, source, err,
		)
	}

	return nil
}

// failureReason lists the failed controls of a SLSA verifier result.
func failureReason(res *slsa.Result) string {
	var reasons []string

	for _, layer := range [][]*slsa.ControlResult{res.CoreResults, res.BuildTypeResults, res.UserResults} {
		for _, cr := range layer {
			if cr.Status != slsa.StatusFail && cr.Status != slsa.StatusError {
				continue
			}

			reason := cr.ID
			if cr.Message != "" {
				reason += " (" + cr.Message + ")"
			}

			reasons = append(reasons, reason)
		}
	}

	if res.Message != "" {
		reasons = append(reasons, res.Message)
	}

	if len(reasons) == 0 {
		return string(res.Status)
	}

	return strings.Join(reasons, ", ")
}

// hexRegex matches lowercase hex encoded digest values.
var hexRegex = regexp.MustCompile(`^[0-9a-f]+$`)

// checkProvenance verifies the hashes of the local copies of the staged
// artifacts against the subjects of the attestation.
func (di *defaultProvenanceCheckerImpl) checkProvenance(
	opts *ProvenanceCheckerOptions, s *intoto.Statement,
) error {
	hashers := map[string]struct {
		hashFile     func(string) (string, error)
		digestLength int
	}{
		intoto.AlgorithmSHA256.String(): {hash.SHA256ForFile, sha256.Size * 2},
		intoto.AlgorithmSHA512.String(): {hash.SHA512ForFile, sha512.Size * 2},
	}

	errs := 0

	for _, sub := range s.GetSubject() {
		if sub.GetName() == "" {
			logrus.Error("Found empty subject in provenance attestation")

			errs++

			continue
		}

		// checked records if at least one digest of the subject was
		// verified, failed if any of its digests failed to verify:
		checked, failed := false, false

		for algo, expected := range sub.GetDigest() {
			hasher, ok := hashers[algo]
			if !ok {
				continue
			}

			if len(expected) != hasher.digestLength || !hexRegex.MatchString(expected) {
				logrus.Errorf("Malformed %s digest in %s", algo, sub.GetName())

				errs++
				failed = true

				break
			}

			actual, err := hasher.hashFile(filepath.Join(opts.StageDirectory, sub.GetName()))
			if err != nil {
				logrus.Errorf("Hashing %s: %v", sub.GetName(), err)

				errs++
				failed = true

				break
			}

			if actual != expected {
				logrus.Errorf("Invalid %s hash in %s", algo, sub.GetName())

				errs++
				failed = true

				break
			}

			checked = true
		}

		// Every subject must have at least one verified digest
		if !checked && !failed {
			logrus.Errorf("Subject %s has no supported digest", sub.GetName())

			errs++
		}
	}

	if errs > 0 {
		return fmt.Errorf("%d errors verifying subjects of the provenance attestation", errs)
	}

	return nil
}

func (di *defaultProvenanceCheckerImpl) generateFinalAttestation(
	opts *ProvenanceCheckerOptions, sbom, stageProvenance, version string,
) error {
	pdoc, err := bom.Open(sbom)
	if err != nil {
		return fmt.Errorf("parsing sbom for version %s from %s: %w", version, sbom, err)
	}

	doc, err := spdx.FromProtobom(pdoc)
	if err != nil {
		return fmt.Errorf("converting sbom for version %s: %w", version, err)
	}

	// The SBOM only provides the subjects: bom's statements declare the
	// SLSA v0.2 predicate type and bom can't read the v1 predicate of the
	// stage run.
	subjects := doc.ToProvenanceStatement(spdx.DefaultProvenanceOptions).Subject

	// Rewrite the provenance sublects to list their full paths in the bucket
	for i, sub := range subjects {
		subjects[i].Name = object.GcsPrefix + filepath.Join(
			opts.StageBucket, "release", version, sub.GetName(),
		)
	}

	data, err := readStageStatement(stageProvenance)
	if err != nil {
		return fmt.Errorf("reading staging provenance: %w", err)
	}

	stageStatement := &intoto.Statement{}
	if err := protojson.Unmarshal(data, stageStatement); err != nil {
		return fmt.Errorf("parsing staging provenance %s: %w", stageProvenance, err)
	}

	statement := &intoto.Statement{
		Type:          intoto.StatementTypeUri,
		Subject:       subjects,
		PredicateType: stageStatement.GetPredicateType(),
		Predicate:     stageStatement.GetPredicate(),
	}

	if err := statement.Validate(); err != nil {
		return fmt.Errorf("checking final provenance attestation for %s: %w", version, err)
	}

	finalData, err := protojson.Marshal(statement)
	if err != nil {
		return fmt.Errorf("serializing final provenance attestation for %s: %w", version, err)
	}

	// The release pushes it from there, see PushArtifacts in pkg/anago.
	if err := os.WriteFile( //nolint:gosec // G303: a fixed name the release copies
		filepath.Join(os.TempDir(), fmt.Sprintf("provenance-%s.json", version)), finalData, 0o600,
	); err != nil {
		return fmt.Errorf("writing final provenance attestation for %s: %w", version, err)
	}

	return nil
}

// inTotoPayloadType is the DSSE payload type of in-toto statements.
const inTotoPayloadType = "application/vnd.in-toto+json"

// readStageStatement returns the in-toto statement of the stage provenance
// in path, see stageStatement.
func readStageStatement(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	return stageStatement(data)
}

// stageStatement returns the in-toto statement of the stage provenance. The
// stage run signs the provenance in place, so it is a sigstore bundle with
// the statement in its DSSE envelope.
func stageStatement(data []byte) ([]byte, error) {
	if !attestation.IsSigstoreBundle(data) {
		return nil, errors.New("not a sigstore bundle, so it is not signed")
	}

	sigstoreBundle := &protobundle.Bundle{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, sigstoreBundle); err != nil {
		return nil, fmt.Errorf("parsing sigstore bundle: %w", err)
	}

	envelope := sigstoreBundle.GetDsseEnvelope()
	if envelope == nil {
		return nil, errors.New("the sigstore bundle holds no DSSE envelope")
	}

	if envelope.GetPayloadType() != inTotoPayloadType {
		return nil, fmt.Errorf(
			"the sigstore bundle holds a %q payload, not an in-toto statement", envelope.GetPayloadType(),
		)
	}

	return envelope.GetPayload(), nil
}

type ProvenanceReader struct {
	options *ProvenanceReaderOptions
	impl    provenanceReaderImplementation
}

// NewProvenanceReader returns a new ProvenanceReader instance.
func NewProvenanceReader(opts *ProvenanceReaderOptions) *ProvenanceReader {
	return &ProvenanceReader{
		options: opts,
		impl:    &defaultProvenanceReaderImpl{},
	}
}

type provenanceReaderImplementation interface {
	GetStagingSubjects(*ProvenanceReaderOptions, string) ([]*intoto.ResourceDescriptor, error)
	GetBuildSubjects(*ProvenanceReaderOptions, string, string) ([]*intoto.ResourceDescriptor, error)
}

type ProvenanceReaderOptions struct {
	Bucket       string
	BuildVersion string
	WorkspaceDir string
}

// GetBuildSubjects returns all artifacts in the output directory
// as intoto subjects, ready to add to the attestation.
func (pr *ProvenanceReader) GetBuildSubjects(path, version string) ([]*intoto.ResourceDescriptor, error) {
	return pr.impl.GetBuildSubjects(pr.options, path, version)
}

// GetStagingSubjects reads artifacts from the GCB workspace and returns them
// as in-toto subjects, with their paths normalized to their final locations
// in the staging bucket.
func (pr *ProvenanceReader) GetStagingSubjects(path string) ([]*intoto.ResourceDescriptor, error) {
	return pr.impl.GetStagingSubjects(pr.options, path)
}

type defaultProvenanceReaderImpl struct{}

func (di *defaultProvenanceReaderImpl) GetStagingSubjects(
	opts *ProvenanceReaderOptions, path string,
) ([]*intoto.ResourceDescriptor, error) {
	// Create the dummy statement to read artifacts
	dummy := provenance.NewSLSAStatement()

	// The path in the bucket were built artifacts will be staged
	gcsPath := filepath.Join(opts.Bucket, StagePath, opts.BuildVersion)

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("checking artifact path to generate provenance subjects: %w", err)
	}

	if info.IsDir() {
		if err := dummy.ReadSubjectsFromDir(path); err != nil {
			return nil, fmt.Errorf("generating provenance subject from file %s: %w", path, err)
		}
	} else {
		if err := dummy.AddSubjectFromFile(path); err != nil {
			return nil, fmt.Errorf("generating provenance subject from file %s: %w", path, err)
		}
	}

	// Check if we are dealing with the sources tar and translate to the top
	if dummy.Subject[0].GetName() == filepath.Join(opts.WorkspaceDir, SourcesTar) {
		dummy.Subject[0].Name = SourcesTar
	}

	for i, s := range dummy.Subject {
		dummy.Subject[i].Name = object.GcsPrefix + filepath.Join(gcsPath, s.GetName())
	}

	return dummy.Subject, nil
}

func (di *defaultProvenanceReaderImpl) GetBuildSubjects(
	opts *ProvenanceReaderOptions, path, version string,
) ([]*intoto.ResourceDescriptor, error) {
	// The path in the bucket were built artifacts will be staged
	gcsPath := filepath.Join(opts.Bucket, StagePath, opts.BuildVersion)

	// When adding the output directory for a specific version, we need
	// to modiy the paths in the attestation to match the bucket names.
	// In order to do that, we create a dummy statement. Use that to read
	// the files and translate those to the final attestation with the paths
	// translated.
	dummy := provenance.NewSLSAStatement()
	if err := dummy.ReadSubjectsFromDir(path); err != nil {
		return nil, fmt.Errorf("reading output directory provenance subjects: %w", err)
	}

	// Cycle the subjects, translate the paths and copy them to the
	// real attestation:
	newSubjects := []*intoto.ResourceDescriptor{}

	for _, subject := range dummy.Subject {
		// If the artifact is not in the images or gcs-stage dir, skip
		if !strings.HasPrefix(subject.GetName(), ImagesPath) &&
			!strings.HasPrefix(subject.GetName(), GCSStagePath) {
			continue
		}

		// Now the tricky part. We need to re-append the version tag. Eg
		// gcs-stage/v1.23.0-alpha.4/file.txt should be
		// v1.23.0-alpha.4/gcs-stage/v1.23.0-alpha.4/file.txt should be
		subject.Name = object.GcsPrefix + filepath.Join(gcsPath, version, subject.GetName())

		newSubjects = append(newSubjects, subject)
	}

	return newSubjects, nil
}

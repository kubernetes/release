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

package release

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	slsa "github.com/in-toto/attestation/go/predicates/provenance/v1"
	intoto "github.com/in-toto/attestation/go/v1"
	sbundle "github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"sigs.k8s.io/release-sdk/object"
	"sigs.k8s.io/release-utils/command"

	"k8s.io/release/pkg/attestation"
)

const (
	// SigningBuildBuilder is the builder ID of the provenance the signing
	// build writes for a stage build.
	SigningBuildBuilder = "https://git.k8s.io/release/gcb/provenance"

	// SigningBuildType is the build type of that provenance, see
	// docs/krel/buildtypes/v2.
	SigningBuildType = "https://git.k8s.io/release/docs/krel/buildtypes/v2"

	// SigningBuildSigner is the service account that signs the provenance
	// of production stages.
	SigningBuildSigner = "krel-provenance@k8s-releng-prod.iam.gserviceaccount.com"

	// SigningBuildMockSigner is the service account that signs the
	// provenance of mock stages, which nothing trusts.
	SigningBuildMockSigner = "krel-provenance-mock@k8s-releng-prod.iam.gserviceaccount.com"

	// SigningBuildProvenanceFilename is the name of the provenance of the
	// staged files, next to ProvenanceFilename in the stage directory.
	SigningBuildProvenanceFilename = "provenance-v2.json"

	// SigningBuildImageProvenanceFilename is the name of the provenance of
	// the staged images, next to ImageProvenanceFilename.
	SigningBuildImageProvenanceFilename = "image-provenance-v2.json"

	// SigningBuildFinalProvenanceFilename is the name of the provenance of
	// the files of the release of a version, in the directory of the version
	// in the stage directory.
	SigningBuildFinalProvenanceFilename = "final-provenance-v2.json"

	// StageServiceAccount is the service account stage builds run as, from
	// their triggers.
	StageServiceAccount = "krel-stage@k8s-release.iam.gserviceaccount.com"

	// stageBuildTag tags the Cloud Build jobs krel submits for a stage.
	stageBuildTag = "STAGE"

	// stageEntryPoint is the Cloud Build configuration of a stage.
	stageEntryPoint = "https://git.k8s.io/release/gcb/stage/cloudbuild.yaml"

	// releaseSBOMFilename is the release SBOM of a version, which lists the
	// files of the release.
	releaseSBOMFilename = "kubernetes-release.spdx"
)

// SigningBuildOptions configures the signing build.
type SigningBuildOptions struct {
	// BuildID is the Cloud Build ID of the stage build.
	BuildID string

	// Project and Region are where the stage build ran.
	Project string
	Region  string

	// Signer is the service account the provenance is signed as, by
	// impersonating it.
	Signer string

	// StageTriggerID and MockStageTriggerID are the IDs of the Cloud Build
	// triggers of production and of mock stages: the production signer only
	// signs builds of the first, the mock signer only of the second.
	StageTriggerID     string
	MockStageTriggerID string

	// ScratchDirectory is where the staged artifacts are downloaded to.
	ScratchDirectory string

	// DryRun writes the unsigned statements to OutputDirectory and changes
	// nothing remote.
	DryRun          bool
	OutputDirectory string
}

// DefaultSigningBuildOptions returns the options for a production stage.
func DefaultSigningBuildOptions() *SigningBuildOptions {
	return &SigningBuildOptions{
		Project:          DefaultKubernetesStagingProject,
		Region:           "us-central1",
		Signer:           SigningBuildSigner,
		ScratchDirectory: filepath.Join(os.TempDir(), "krel-provenance"),
		OutputDirectory:  ".",
	}
}

// SigningBuild writes the provenance of a stage build from outside of it,
// for SLSA build level 3: it reads the record of the stage build from Cloud
// Build, computes the digests of the staged files and images itself and
// signs SLSA v1 provenance about them with its own builder and signer.
type SigningBuild struct {
	options *SigningBuildOptions
	impl    signingBuildImpl
}

// NewSigningBuild returns a signing build for the options.
func NewSigningBuild(opts *SigningBuildOptions) *SigningBuild {
	return &SigningBuild{options: opts, impl: &defaultSigningBuildImpl{}}
}

// SetImpl replaces the implementation, for testing.
func (s *SigningBuild) SetImpl(impl signingBuildImpl) {
	s.impl = impl
}

//counterfeiter:generate . signingBuildImpl
type signingBuildImpl interface {
	GetBuild(project, region, id string) (*cloudbuild.Build, error)
	DownloadStage(gcsPath, dir string) error
	FileSubjects(bucket, buildVersion, stageDir string, versions []string) ([]*intoto.ResourceDescriptor, error)
	ImageSubjects(registry, version, buildDir string) ([]*intoto.ResourceDescriptor, error)
	FinalSubjects(sbom, bucket, version string) ([]*intoto.ResourceDescriptor, error)
	StagedObjects(gcsPath string) (map[string]StagedObject, error)
	ImageUploads(repository string) (map[string]time.Time, error)
	ReadObject(gcsPath string) ([]byte, error)
	VerifyProvenance(data []byte, signer, builder string) ([]byte, error)
	Publish(signer, registry string, statements []*SigningBuildStatement) error
}

// SigningBuildStatement is a statement of the signing build and where its
// bundle goes.
type SigningBuildStatement struct {
	// Local is the unsigned statement, Remote the gs:// path of its bundle.
	Local  string
	Remote string

	// Attach says whether the bundle gets attached to the images the
	// statement is about.
	Attach bool

	// Existing is a bundle at Remote that verified and holds this
	// statement: it is kept and attached again, and nothing is signed.
	Existing []byte
}

// StagedObject is the metadata of an object in the stage directory.
type StagedObject struct {
	Created time.Time
	Updated time.Time
}

// stageBuild is what the signing build takes from the record of a stage.
type stageBuild struct {
	started       time.Time
	finished      time.Time
	buildVersion  string
	bucket        string
	registry      string
	releaseType   string
	releaseBranch string
	nomock        bool
}

// Run writes the provenance of the stage build.
func (s *SigningBuild) Run() error {
	build, err := s.impl.GetBuild(s.options.Project, s.options.Region, s.options.BuildID)
	if err != nil {
		return fmt.Errorf("reading stage build %s: %w", s.options.BuildID, err)
	}

	triggerID := s.options.StageTriggerID
	if s.options.Signer == SigningBuildMockSigner {
		triggerID = s.options.MockStageTriggerID
	}

	stage, err := parseStageBuild(build, triggerID)
	if err != nil {
		return fmt.Errorf("stage build %s: %w", s.options.BuildID, err)
	}

	// krel stage picks its bucket and registry from --nomock, not from the
	// substitutions the locations are checked against below, so each signer
	// also only signs its kind of stage. The provenance of a mock stage must
	// not be trusted, and the mock signer must not sign a nomock stage.
	if s.options.Signer == SigningBuildSigner && !stage.nomock {
		return fmt.Errorf("stage build %s is a mock stage, which %s does not sign", s.options.BuildID, SigningBuildSigner)
	}

	if s.options.Signer == SigningBuildMockSigner && stage.nomock {
		return fmt.Errorf("stage build %s is a nomock stage, which %s does not sign", s.options.BuildID, SigningBuildMockSigner)
	}

	if err := checkSignerLocations(s.options.Signer, stage); err != nil {
		return fmt.Errorf("stage build %s: %w", s.options.BuildID, err)
	}

	predicate, err := signingBuildPredicate(build, stage)
	if err != nil {
		return err
	}

	stagePath := object.GcsPrefix + filepath.Join(stage.bucket, StagePath, stage.buildVersion)
	if err := s.impl.DownloadStage(stagePath+"/", s.options.ScratchDirectory); err != nil {
		return fmt.Errorf("downloading the staged artifacts: %w", err)
	}

	stageDir := filepath.Join(s.options.ScratchDirectory, stage.buildVersion)

	versions, err := stageVersions(stageDir)
	if err != nil {
		return err
	}

	files, err := s.impl.FileSubjects(stage.bucket, stage.buildVersion, stageDir, versions)
	if err != nil {
		return fmt.Errorf("reading the staged files: %w", err)
	}

	files = slices.DeleteFunc(slices.Clone(files), isReleaseSignature)

	// Listed after the download: any write to a staged object after the
	// stage build, also one before the download, gives the current
	// generation a later time.
	objects, err := s.impl.StagedObjects(stagePath + "/")
	if err != nil {
		return fmt.Errorf("listing the staged files: %w", err)
	}

	if err := checkStagedFiles(files, objects, stage); err != nil {
		return err
	}

	images := []*intoto.ResourceDescriptor{}
	finals := map[string][]*intoto.ResourceDescriptor{}

	for _, version := range versions {
		subjects, err := s.impl.ImageSubjects(stage.registry, version, filepath.Join(stageDir, version))
		if err != nil {
			return fmt.Errorf("reading the staged images of %s: %w", version, err)
		}

		if err := s.checkStagedImages(subjects, stage); err != nil {
			return err
		}

		images = append(images, subjects...)

		final, err := s.impl.FinalSubjects(
			filepath.Join(stageDir, version, GCSStagePath, version, releaseSBOMFilename), stage.bucket, version,
		)
		if err != nil {
			return fmt.Errorf("reading the release files of %s: %w", version, err)
		}

		if err := checkFinalSubjects(final, files, stage, version); err != nil {
			return err
		}

		finals[version] = final
	}

	statements := []*namedStatement{
		{name: SigningBuildProvenanceFilename, remote: stagePath + "/" + SigningBuildProvenanceFilename, subjects: files},
		{
			name: SigningBuildImageProvenanceFilename, remote: stagePath + "/" + SigningBuildImageProvenanceFilename,
			subjects: images, attach: true,
		},
	}

	for _, version := range versions {
		statements = append(statements, &namedStatement{
			name:     version + "-" + SigningBuildFinalProvenanceFilename,
			remote:   stagePath + "/" + version + "/" + SigningBuildFinalProvenanceFilename,
			subjects: finals[version],
		})
	}

	dir := s.options.OutputDirectory
	if !s.options.DryRun {
		dir = filepath.Join(s.options.ScratchDirectory, "statements-"+s.options.BuildID)
	}

	if err := writeStatements(dir, statements, predicate); err != nil {
		return err
	}

	if s.options.DryRun {
		logrus.Infof(
			"Dry run: wrote %d statements to %s about %d files and %d images, nothing was signed",
			len(statements), dir, len(files), len(images),
		)

		return nil
	}

	pending, err := s.pending(statements)
	if err != nil {
		return err
	}

	if err := s.impl.Publish(s.options.Signer, stage.registry, pending); err != nil {
		return fmt.Errorf("publishing the provenance: %w", err)
	}

	return nil
}

// namedStatement is a statement the signing build writes.
type namedStatement struct {
	name      string
	remote    string
	subjects  []*intoto.ResourceDescriptor
	attach    bool
	statement *intoto.Statement
	local     string
}

// pending returns the statements to publish. A bundle that exists already,
// verifies as the signer's and holds the same statement is kept, so that a
// rerun for the same stage build signs nothing again; anything else is
// signed and replaced.
func (s *SigningBuild) pending(statements []*namedStatement) ([]*SigningBuildStatement, error) {
	pending := make([]*SigningBuildStatement, 0, len(statements))

	for _, st := range statements {
		entry := &SigningBuildStatement{Local: st.local, Remote: st.remote, Attach: st.attach}

		existing, err := s.impl.ReadObject(st.remote)

		switch {
		case errors.Is(err, os.ErrNotExist):
		case err != nil:
			return nil, fmt.Errorf("reading %s: %w", st.remote, err)
		default:
			if s.keeps(existing, st) {
				logrus.Infof("Keeping %s, it holds the same signed statement", st.remote)

				entry.Existing = existing
			} else {
				logrus.Infof("Signing %s again, the existing bundle doesn't hold this signed statement", st.remote)
			}
		}

		pending = append(pending, entry)
	}

	return pending, nil
}

// keeps returns whether the existing bundle verifies as the signer's and
// holds the statement.
func (s *SigningBuild) keeps(existing []byte, st *namedStatement) bool {
	payload, err := s.impl.VerifyProvenance(existing, s.options.Signer, SigningBuildBuilder)
	if err != nil {
		logrus.Infof("Existing bundle at %s doesn't verify: %v", st.remote, err)

		return false
	}

	signed := &intoto.Statement{}
	if err := protojson.Unmarshal(payload, signed); err != nil {
		return false
	}

	return proto.Equal(signed, st.statement)
}

// writeStatements writes the statements into dir.
func writeStatements(dir string, statements []*namedStatement, predicate *structpb.Struct) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	for _, st := range statements {
		st.statement = &intoto.Statement{
			Type:          intoto.StatementTypeUri,
			Subject:       st.subjects,
			PredicateType: "https://slsa.dev/provenance/v1",
			Predicate:     predicate,
		}

		if err := st.statement.Validate(); err != nil {
			return fmt.Errorf("checking %s: %w", st.name, err)
		}

		data, err := protojson.Marshal(st.statement)
		if err != nil {
			return fmt.Errorf("serializing %s: %w", st.name, err)
		}

		st.local = filepath.Join(dir, st.name)
		if err := os.WriteFile(st.local, data, 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", st.local, err)
		}
	}

	return nil
}

// parseStageBuild checks that the build is a successful stage, run by the
// stage trigger with the ID triggerID as StageServiceAccount, and reads what
// the signing build needs from its record. Cloud Build announces every build
// of the project, also the ones release managers submit by hand, which must
// not get signed provenance.
func parseStageBuild(build *cloudbuild.Build, triggerID string) (stage *stageBuild, err error) {
	if build.Status != "SUCCESS" {
		return nil, fmt.Errorf("its status is %s, not SUCCESS", build.Status)
	}

	if triggerID == "" {
		return nil, errors.New("no stage trigger ID is set to check it against")
	}

	if build.BuildTriggerId != triggerID {
		return nil, fmt.Errorf("it was started by trigger %q, not by the stage trigger %s", build.BuildTriggerId, triggerID)
	}

	if account := serviceAccountEmail(build.ServiceAccount); account != StageServiceAccount {
		return nil, fmt.Errorf("it ran as %q, not as %s", account, StageServiceAccount)
	}

	if !slices.Contains(build.Tags, stageBuildTag) {
		return nil, fmt.Errorf("it has no %s tag, so it is not a stage build", stageBuildTag)
	}

	if !slices.ContainsFunc(build.Steps, isKrelStageStep) {
		return nil, errors.New("none of its steps runs krel stage")
	}

	subs := build.Substitutions
	stage = &stageBuild{
		buildVersion:  subs["_BUILDVERSION"],
		registry:      subs["_STAGE_REGISTRY"],
		releaseType:   subs["_TYPE"],
		releaseBranch: subs["_RELEASE_BRANCH"],
		nomock:        stageStepNomock(build.Steps),
	}

	bucket, ok := strings.CutPrefix(subs["_STAGE_BUCKET"], object.GcsPrefix)
	if !ok || bucket == "" || strings.Contains(bucket, "/") {
		return nil, fmt.Errorf("its stage bucket %q is not a gs:// bucket", subs["_STAGE_BUCKET"])
	}

	stage.bucket = bucket

	if stage.started, err = time.Parse(time.RFC3339Nano, build.StartTime); err != nil {
		return nil, fmt.Errorf("parsing its start time: %w", err)
	}

	if stage.finished, err = time.Parse(time.RFC3339Nano, build.FinishTime); err != nil {
		return nil, fmt.Errorf("parsing its finish time: %w", err)
	}

	for key, value := range map[string]string{
		"_BUILDVERSION": stage.buildVersion, "_STAGE_REGISTRY": stage.registry,
		"_TYPE": stage.releaseType, "_RELEASE_BRANCH": stage.releaseBranch,
	} {
		if value == "" {
			return nil, fmt.Errorf("it has no %s substitution", key)
		}
	}

	if strings.Contains(stage.buildVersion, "/") || strings.Contains(stage.buildVersion, "..") {
		return nil, fmt.Errorf("its build version %q is not one", stage.buildVersion)
	}

	return stage, nil
}

// serviceAccountEmail returns the email of a service account given as email
// or as projects/<project>/serviceAccounts/<email>.
func serviceAccountEmail(account string) string {
	if _, email, ok := strings.Cut(account, "/serviceAccounts/"); ok {
		return email
	}

	return account
}

// stageStepNomock returns whether the krel stage step ran with --nomock, which
// decides the bucket and registry it staged to. The record holds the
// arguments with the substitutions applied.
func stageStepNomock(steps []*cloudbuild.BuildStep) bool {
	for _, step := range steps {
		if isKrelStageStep(step) {
			return slices.Contains(step.Args, "--nomock") || slices.Contains(step.Args, "--nomock=true")
		}
	}

	return false
}

// isKrelStageStep returns whether a build step runs krel stage.
func isKrelStageStep(step *cloudbuild.BuildStep) bool {
	return len(step.Args) >= 2 && filepath.Base(step.Args[0]) == "krel" && step.Args[1] == "stage"
}

// stageVersions returns the versions a stage directory holds artifacts of.
func stageVersions(stageDir string) ([]string, error) {
	entries, err := os.ReadDir(stageDir)
	if err != nil {
		return nil, fmt.Errorf("reading the stage directory: %w", err)
	}

	versions := []string{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if _, err := os.Stat(filepath.Join(stageDir, entry.Name(), GCSStagePath)); err == nil {
			versions = append(versions, entry.Name())
		}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no staged versions in %s", stageDir)
	}

	return versions, nil
}

// checkFinalSubjects checks that every file of the release of the version
// is a staged file with the same SHA-256 digest, so that the final
// provenance is about digests the signing build computed itself, not only
// the ones the release SBOM of the stage claims.
func checkFinalSubjects(final, files []*intoto.ResourceDescriptor, stage *stageBuild, version string) error {
	staged := make(map[string]string, len(files))
	for _, file := range files {
		staged[file.GetName()] = file.GetDigest()[intoto.AlgorithmSHA256.String()]
	}

	releasePrefix := object.GcsPrefix + filepath.Join(stage.bucket, "release", version) + "/"
	stagePrefix := object.GcsPrefix + filepath.Join(stage.bucket, StagePath, stage.buildVersion, version, GCSStagePath, version) + "/"

	for _, subject := range final {
		file, ok := strings.CutPrefix(subject.GetName(), releasePrefix)
		if !ok {
			return fmt.Errorf("release file %s of %s is not in %s", subject.GetName(), version, releasePrefix)
		}

		digest, ok := staged[stagePrefix+file]
		if !ok {
			return fmt.Errorf("release file %s of %s is not staged", file, version)
		}

		if want := subject.GetDigest()[intoto.AlgorithmSHA256.String()]; want == "" || want != digest {
			return fmt.Errorf("release file %s of %s has SHA-256 %q in the release SBOM, but the staged file %s", file, version, want, digest)
		}

		subject.Digest = map[string]string{intoto.AlgorithmSHA256.String(): digest}
	}

	return nil
}

// signerLocations are the bucket and registry each signer signs the stage of.
// They are fixed, not taken from the stage build, whose substitutions release
// managers set: otherwise a mock stage could get the provenance of the mock
// signer attached to production images, or the production signer could sign
// files in the public mock bucket.
var signerLocations = map[string]struct{ bucket, registry string }{
	SigningBuildSigner:     {bucket: ProductionBucket, registry: GCRIOPathStaging},
	SigningBuildMockSigner: {bucket: TestBucket, registry: GCRIOPathMock},
}

// checkSignerLocations checks that the stage build staged to the bucket and
// registry of the signer.
func checkSignerLocations(signer string, stage *stageBuild) error {
	want, ok := signerLocations[signer]
	if !ok {
		return fmt.Errorf("%s is neither the production nor the mock signer", signer)
	}

	if stage.bucket != want.bucket {
		return fmt.Errorf("it staged to bucket %s, but %s only signs stages in %s", stage.bucket, signer, want.bucket)
	}

	if stage.registry != want.registry {
		return fmt.Errorf("it staged to registry %s, but %s only signs stages in %s", stage.registry, signer, want.registry)
	}

	return nil
}

// isReleaseSignature reports whether a staged file is a signature or
// certificate krel sign blobs adds to the stage directory during the release,
// after the stage build, which the stage provenance doesn't name either.
func isReleaseSignature(file *intoto.ResourceDescriptor) bool {
	return strings.HasSuffix(file.GetName(), ".sig") || strings.HasSuffix(file.GetName(), ".cert")
}

// inStageWindow reports whether t is within the run of the stage build. GCS
// reports times in whole seconds, hence the second of slack.
func inStageWindow(t time.Time, stage *stageBuild) bool {
	return !t.Before(stage.started.Truncate(time.Second)) && !t.After(stage.finished.Add(time.Second))
}

// checkStagedFiles checks that every staged file was written while the stage
// build ran and not changed since, from the metadata of the objects: someone
// who can write the stage directory could replace a file after the stage
// build, and the provenance would vouch for it.
func checkStagedFiles(files []*intoto.ResourceDescriptor, objects map[string]StagedObject, stage *stageBuild) error {
	for _, file := range files {
		obj, ok := objects[file.GetName()]
		if !ok {
			return fmt.Errorf("staged file %s is not in the bucket", file.GetName())
		}

		if !inStageWindow(obj.Created, stage) || !inStageWindow(obj.Updated, stage) {
			return fmt.Errorf(
				"staged file %s was written at %s and updated at %s, outside of the stage build from %s to %s",
				file.GetName(), obj.Created.Format(time.RFC3339), obj.Updated.Format(time.RFC3339),
				stage.started.Format(time.RFC3339), stage.finished.Format(time.RFC3339),
			)
		}
	}

	return nil
}

// checkStagedImages checks that every image the stage tags resolved to was
// uploaded to its repository while the stage build ran: a tag moved to
// another image afterwards, a new one or an older one, fails.
func (s *SigningBuild) checkStagedImages(images []*intoto.ResourceDescriptor, stage *stageBuild) error {
	uploads := map[string]map[string]time.Time{}

	for _, image := range images {
		repo := image.GetName()

		if _, ok := uploads[repo]; !ok {
			repoUploads, err := s.impl.ImageUploads(repo)
			if err != nil {
				return fmt.Errorf("listing the images of %s: %w", repo, err)
			}

			uploads[repo] = repoUploads
		}

		digest := "sha256:" + image.GetDigest()[intoto.AlgorithmSHA256.String()]

		uploaded, ok := uploads[repo][digest]
		if !ok {
			return fmt.Errorf("staged image %s@%s is not in its repository", repo, digest)
		}

		if !inStageWindow(uploaded, stage) {
			return fmt.Errorf(
				"staged image %s@%s was uploaded at %s, outside of the stage build from %s to %s",
				repo, digest, uploaded.Format(time.RFC3339),
				stage.started.Format(time.RFC3339), stage.finished.Format(time.RFC3339),
			)
		}
	}

	return nil
}

// gitCommitDigest is the digest algorithm of git commits in SLSA provenance.
const gitCommitDigest = "gitCommit"

// fullCommit matches a full git commit hash.
var fullCommit = regexp.MustCompile(`^[0-9a-f]{40}$`)

// signingBuildPredicate returns the SLSA v1 provenance predicate of the stage
// build, from its Cloud Build record.
func signingBuildPredicate(build *cloudbuild.Build, stage *stageBuild) (*structpb.Struct, error) {
	steps := make([]any, 0, len(build.Steps))
	for _, step := range build.Steps {
		args := make([]any, 0, len(step.Args))
		for _, arg := range step.Args {
			args = append(args, arg)
		}

		steps = append(steps, map[string]any{"name": step.Name, "entrypoint": step.Entrypoint, "args": args})
	}

	// The provenance is public, so it leaves out who submitted the stage.
	substitutions := map[string]any{}

	for _, key := range slices.Sorted(maps.Keys(build.Substitutions)) {
		if key != "_GCP_USER_TAG" {
			substitutions[key] = build.Substitutions[key]
		}
	}

	external, err := structpb.NewStruct(map[string]any{
		"releaseType":   stage.releaseType,
		"releaseBranch": stage.releaseBranch,
		"buildVersion":  stage.buildVersion,
		"nomock":        stage.nomock,
		"entryPoint":    stageEntryPoint,
		"substitutions": substitutions,
		"steps":         steps,
	})
	if err != nil {
		return nil, fmt.Errorf("recording the parameters of the stage build: %w", err)
	}

	internal, err := structpb.NewStruct(map[string]any{"project": build.ProjectId})
	if err != nil {
		return nil, fmt.Errorf("recording the project of the stage build: %w", err)
	}

	predicate := &slsa.Provenance{
		BuildDefinition: &slsa.BuildDefinition{
			BuildType:            SigningBuildType,
			ExternalParameters:   external,
			InternalParameters:   internal,
			ResolvedDependencies: resolvedDependencies(build),
		},
		RunDetails: &slsa.RunDetails{
			Builder: &slsa.Builder{Id: SigningBuildBuilder},
			Metadata: &slsa.BuildMetadata{
				InvocationId: build.Id,
				StartedOn:    timestamppb.New(stage.started),
				FinishedOn:   timestamppb.New(stage.finished),
			},
		},
	}

	data, err := protojson.Marshal(predicate)
	if err != nil {
		return nil, fmt.Errorf("serializing the provenance predicate: %w", err)
	}

	res := &structpb.Struct{}
	if err := protojson.Unmarshal(data, res); err != nil {
		return nil, fmt.Errorf("converting the provenance predicate: %w", err)
	}

	return res, nil
}

// resolvedDependencies returns the dependencies of the stage build its record
// names: the source Cloud Build checked out, the kubernetes/release ref krel
// came from and the images of the steps, by digest.
func resolvedDependencies(build *cloudbuild.Build) []*intoto.ResourceDescriptor {
	deps := []*intoto.ResourceDescriptor{}

	subs := build.Substitutions

	// Kubernetes, from the Cloud Build source if that is the Kubernetes
	// repository, which records the commit, or else from the repository and
	// ref the stage checked out. A trigger's source is the repository of its
	// config, which is not Kubernetes.
	if org, repo := subs["_K8S_ORG"], subs["_K8S_REPO"]; org != "" && repo != "" {
		url := fmt.Sprintf("https://github.com/%s/%s", org, repo)

		if src := build.SourceProvenance; src != nil && src.ResolvedGitSource != nil &&
			strings.TrimSuffix(src.ResolvedGitSource.Url, ".git") == url && fullCommit.MatchString(src.ResolvedGitSource.Revision) {
			deps = append(deps, &intoto.ResourceDescriptor{
				Uri:    "git+" + url,
				Digest: map[string]string{gitCommitDigest: src.ResolvedGitSource.Revision},
			})
		} else {
			dep := &intoto.ResourceDescriptor{Uri: "git+" + url}
			if ref := subs["_K8S_REF"]; ref != "" {
				dep.Uri += "@" + ref
			}

			if fullCommit.MatchString(subs["_K8S_REF"]) {
				dep.Digest = map[string]string{gitCommitDigest: subs["_K8S_REF"]}
			} else {
				logrus.Warn("The stage build records no Kubernetes commit, only the ref it checked out")
			}

			deps = append(deps, dep)
		}
	}

	if org, repo, ref := subs["_TOOL_ORG"], subs["_TOOL_REPO"], subs["_TOOL_REF"]; org != "" && repo != "" && ref != "" {
		dep := &intoto.ResourceDescriptor{Uri: fmt.Sprintf("git+https://github.com/%s/%s@%s", org, repo, ref)}
		if fullCommit.MatchString(ref) {
			dep.Digest = map[string]string{gitCommitDigest: ref}
		}

		deps = append(deps, dep)
	}

	var images []string
	if build.Results != nil {
		images = build.Results.BuildStepImages
	}

	seen := map[string]bool{}

	for i, step := range build.Steps {
		if i >= len(images) {
			break
		}

		digest, ok := strings.CutPrefix(images[i], "sha256:")
		if !ok || seen[step.Name+digest] {
			continue
		}

		seen[step.Name+digest] = true
		deps = append(deps, &intoto.ResourceDescriptor{
			Name:   step.Name,
			Digest: map[string]string{intoto.AlgorithmSHA256.String(): digest},
		})
	}

	return deps
}

type defaultSigningBuildImpl struct{}

func (*defaultSigningBuildImpl) GetBuild(project, region, id string) (*cloudbuild.Build, error) {
	ctx := context.Background()

	service, err := cloudbuild.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating the Cloud Build client: %w", err)
	}

	build, err := service.Projects.Locations.Builds.Get(
		fmt.Sprintf("projects/%s/locations/%s/builds/%s", project, region, id),
	).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("getting the build: %w", err)
	}

	return build, nil
}

func (*defaultSigningBuildImpl) DownloadStage(gcsPath, dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	objStore := object.NewGCS()
	objStore.WithConcurrent(true)
	objStore.WithRecursive(true)

	logrus.Infof("Downloading %s to %s", gcsPath, dir)

	return objStore.CopyToLocal(gcsPath, dir) //nolint:wrapcheck // wrapped by the caller
}

// FileSubjects returns the staged files the provenance of a stage names: the
// source tarball and the staged files and image tarballs of each version.
func (*defaultSigningBuildImpl) FileSubjects(
	bucket, buildVersion, stageDir string, versions []string,
) ([]*intoto.ResourceDescriptor, error) {
	reader := NewProvenanceReader(&ProvenanceReaderOptions{
		Bucket: bucket, BuildVersion: buildVersion, WorkspaceDir: stageDir,
	})

	subjects, err := reader.GetStagingSubjects(filepath.Join(stageDir, SourcesTar))
	if err != nil {
		return nil, err
	}

	for _, version := range versions {
		versionSubjects, err := reader.GetBuildSubjects(filepath.Join(stageDir, version), version)
		if err != nil {
			return nil, err
		}

		subjects = append(subjects, versionSubjects...)
	}

	return subjects, nil
}

func (*defaultSigningBuildImpl) ImageSubjects(registry, version, buildDir string) ([]*intoto.ResourceDescriptor, error) {
	return NewImages().Subjects(registry, version, buildDir)
}

func (*defaultSigningBuildImpl) FinalSubjects(sbom, bucket, version string) ([]*intoto.ResourceDescriptor, error) {
	return finalSubjects(sbom, bucket, version)
}

// stagedObject is an object gcloud storage objects list reports.
type stagedObject struct {
	Bucket       string `json:"bucket"`
	Name         string `json:"name"`
	CreationTime string `json:"creation_time"`
	UpdateTime   string `json:"update_time"`
}

// gcloudTime is the time format of gcloud storage.
const gcloudTime = "2006-01-02T15:04:05-0700"

// StagedObjects lists the objects below gcsPath, by their gs:// path.
func (*defaultSigningBuildImpl) StagedObjects(gcsPath string) (map[string]StagedObject, error) {
	out, err := command.New("gcloud", "storage", "objects", "list", gcsPath+"**", "--format=json").
		RunSilentSuccessOutput()
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", gcsPath, err)
	}

	listed := []stagedObject{}
	if err := json.Unmarshal([]byte(out.Output()), &listed); err != nil {
		return nil, fmt.Errorf("parsing the objects of %s: %w", gcsPath, err)
	}

	objects := make(map[string]StagedObject, len(listed))

	for _, obj := range listed {
		created, err := time.Parse(gcloudTime, obj.CreationTime)
		if err != nil {
			return nil, fmt.Errorf("parsing the creation time of %s: %w", obj.Name, err)
		}

		updated, err := time.Parse(gcloudTime, obj.UpdateTime)
		if err != nil {
			return nil, fmt.Errorf("parsing the update time of %s: %w", obj.Name, err)
		}

		objects[object.GcsPrefix+obj.Bucket+"/"+obj.Name] = StagedObject{Created: created, Updated: updated}
	}

	return objects, nil
}

// ImageUploads returns when each manifest of the repository was uploaded, by
// digest, from the Google registry API.
func (*defaultSigningBuildImpl) ImageUploads(repository string) (map[string]time.Time, error) {
	repo, err := name.NewRepository(repository)
	if err != nil {
		return nil, fmt.Errorf("parsing repository %s: %w", repository, err)
	}

	tags, err := google.List(repo, google.WithAuthFromKeychain(google.Keychain))
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", repository, err)
	}

	uploads := make(map[string]time.Time, len(tags.Manifests))
	for digest, info := range tags.Manifests {
		uploads[digest] = info.Uploaded
	}

	return uploads, nil
}

func (*defaultSigningBuildImpl) ReadObject(gcsPath string) ([]byte, error) {
	objStore := object.NewGCS()

	exists, err := objStore.PathExists(gcsPath)
	if err != nil {
		return nil, fmt.Errorf("checking %s: %w", gcsPath, err)
	}

	if !exists {
		return nil, os.ErrNotExist
	}

	dir, err := os.MkdirTemp("", "krel-provenance-")
	if err != nil {
		return nil, fmt.Errorf("creating a temporary directory: %w", err)
	}
	defer os.RemoveAll(dir)

	local := filepath.Join(dir, filepath.Base(gcsPath))
	if err := objStore.CopyToLocal(gcsPath, local); err != nil {
		return nil, fmt.Errorf("downloading %s: %w", gcsPath, err)
	}

	return os.ReadFile(local) //nolint:gosec,wrapcheck // a file this function created
}

func (*defaultSigningBuildImpl) VerifyProvenance(data []byte, signer, builder string) ([]byte, error) {
	return verifyProvenanceBundle(data, nil, "sigstore::https://accounts.google.com::"+signer, builder, "")
}

// Publish signs the statements without an existing bundle as the signer,
// attaches the bundles that should be to their images below the registry
// and only then writes the new bundles.
func (*defaultSigningBuildImpl) Publish(signer, registry string, statements []*SigningBuildStatement) error {
	signerOpts := attestation.DefaultSignerOptions()
	signerOpts.ImpersonateServiceAccount = signer
	sgnr := attestation.NewSigner(signerOpts)
	// Only the bundle of the signer counts as attached: anyone who can write
	// to the registry could attach another one with the same statement.
	attacher := attestation.NewAttacher(&attestation.AttacherOptions{Repository: registry, ExactBundle: true})

	toSign := []string{}

	for _, st := range statements {
		if st.Existing == nil {
			toSign = append(toSign, st.Local)
		}
	}

	signed := map[string]*attestation.SignedStatement{}

	if len(toSign) > 0 {
		results, err := sgnr.SignFiles(toSign)
		if err != nil {
			return fmt.Errorf("signing: %w", err)
		}

		for _, result := range results {
			signed[result.Path] = result
		}
	}

	for _, st := range statements {
		if !st.Attach {
			continue
		}

		statement, bundle, err := bundleOf(sgnr, st, signed)
		if err != nil {
			return err
		}

		attached, err := attacher.AttachToImages(statement, bundle)
		if err != nil {
			return fmt.Errorf("attaching %s to its images: %w", st.Remote, err)
		}

		logrus.Infof("Attached %s to %d image(s)", st.Remote, attached)
	}

	for _, st := range statements {
		if st.Existing != nil {
			continue
		}

		if err := sgnr.WriteBundleFile(st.Remote, signed[st.Local].Bundle); err != nil {
			return fmt.Errorf("writing %s: %w", st.Remote, err)
		}

		logrus.Infof("Wrote %s", st.Remote)
	}

	return nil
}

// bundleOf returns the signed statement and serialized bundle of a statement,
// the existing one or the one just signed.
func bundleOf(
	sgnr *attestation.Signer, st *SigningBuildStatement, signed map[string]*attestation.SignedStatement,
) (*attestation.SignedStatement, []byte, error) {
	if st.Existing != nil {
		bndl := &sbundle.Bundle{}
		if err := bndl.UnmarshalJSON(st.Existing); err != nil {
			return nil, nil, fmt.Errorf("parsing the bundle at %s: %w", st.Remote, err)
		}

		return &attestation.SignedStatement{Path: st.Remote, Bundle: bndl}, st.Existing, nil
	}

	statement := signed[st.Local]

	var data bytes.Buffer
	if err := sgnr.WriteBundle(statement.Bundle, &data); err != nil {
		return nil, nil, fmt.Errorf("serializing the bundle of %s: %w", st.Remote, err)
	}

	return statement, data.Bytes(), nil
}

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

package release_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	slsa "github.com/in-toto/attestation/go/predicates/provenance/v1"
	intoto "github.com/in-toto/attestation/go/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/release/releasefakes"
)

const (
	sbBuildID      = "54ef86d9-5a89-4e0e-a008-10904a0f5a34"
	sbBuildVersion = "v1.38.0-beta.0.899+9f140ee0061e18"
	sbVersion      = "v1.38.0-beta.1"
	sbBucket       = release.ProductionBucket
	sbRegistry     = release.GCRIOPathStaging
	sbKrelCommit   = "9c79ae380875188f191f7d30ee01389772a10f3d"
	sbK8sCommit    = "9f140ee0061e181fa296a4257dda0044a63f1f6c"
	sbDigestA      = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	sbDigestB      = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	sbDigestC      = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	sbTrigger      = "11111111-1111-1111-1111-111111111111"
	sbMockTrigger  = "22222222-2222-2222-2222-222222222222"
)

// sbStageBuild returns the Cloud Build record of a nomock stage build.
func sbStageBuild() *cloudbuild.Build {
	return &cloudbuild.Build{
		Id:             sbBuildID,
		ProjectId:      "k8s-release",
		BuildTriggerId: sbTrigger,
		ServiceAccount: "projects/k8s-release/serviceAccounts/" + release.StageServiceAccount,
		Status:         "SUCCESS",
		Tags:           []string{"nomock", "STAGE", "1.38.0-beta.1"},
		StartTime:      "2026-11-25T15:56:49.360329688Z",
		FinishTime:     "2026-11-25T16:26:50.339390Z",
		Substitutions: map[string]string{
			"_BUILDVERSION":   sbBuildVersion,
			"_STAGE_BUCKET":   "gs://" + sbBucket,
			"_STAGE_REGISTRY": sbRegistry,
			"_TYPE":           "beta",
			"_RELEASE_BRANCH": "master",
			"_NOMOCK":         "--nomock",
			"_TOOL_ORG":       "kubernetes",
			"_TOOL_REPO":      "release",
			"_TOOL_REF":       sbKrelCommit,
			"_GCP_USER_TAG":   "someone-at-example-com",
			"_K8S_ORG":        "kubernetes",
			"_K8S_REPO":       "kubernetes",
			"_K8S_REF":        "HEAD",
		},
		Steps: []*cloudbuild.BuildStep{
			{Name: "gcr.io/cloud-builders/git", Args: []string{"clone", "https://github.com/kubernetes/release"}},
			{Name: "gcr.io/k8s-staging-releng/k8s-cloud-builder:v1.38.0", Args: []string{"bin/krel", "stage", "--nomock"}},
			{Name: "gcr.io/k8s-staging-releng/k8s-cloud-builder:v1.38.0", Args: []string{"bin/krel", "sign"}},
		},
		Results: &cloudbuild.Results{
			BuildStepImages: []string{"sha256:" + sbDigestA, "sha256:" + sbDigestB, "sha256:" + sbDigestB},
		},
		SourceProvenance: &cloudbuild.SourceProvenance{
			ResolvedGitSource: &cloudbuild.GitSource{
				Url: "https://github.com/kubernetes/kubernetes", Revision: sbK8sCommit,
			},
		},
	}
}

func sbStagePath(rest string) string {
	return sbStagePathIn(sbBucket, rest)
}

func sbStagePathIn(bucket, rest string) string {
	return "gs://" + filepath.Join(bucket, release.StagePath, sbBuildVersion, rest)
}

func sbSubject(name, digest string) *intoto.ResourceDescriptor {
	return &intoto.ResourceDescriptor{Name: name, Digest: map[string]string{"sha256": digest}}
}

// signingBuildTest is a signing build of a stage with one staged release
// file and one image, whose bucket holds objects and whose bundles verify
// to the payloads in verified.
type signingBuildTest struct {
	sut      *release.SigningBuild
	opts     *release.SigningBuildOptions
	mock     *releasefakes.FakeSigningBuildImpl
	objects  map[string][]byte
	verified map[string][]byte

	// writtenAt and uploadedAt are when the staged files and images were
	// written, by gs:// path and digest.
	writtenAt  map[string]time.Time
	uploadedAt map[string]time.Time
}

func newSigningBuildTest(t *testing.T, build *cloudbuild.Build) *signingBuildTest {
	t.Helper()

	scratch := t.TempDir()
	bucket := strings.TrimPrefix(build.Substitutions["_STAGE_BUCKET"], "gs://")
	stagePath := func(rest string) string { return sbStagePathIn(bucket, rest) }
	st := &signingBuildTest{
		mock:     &releasefakes.FakeSigningBuildImpl{},
		objects:  map[string][]byte{},
		verified: map[string][]byte{},
	}

	st.opts = release.DefaultSigningBuildOptions()
	st.opts.BuildID = sbBuildID
	st.opts.StageTriggerID = sbTrigger
	st.opts.MockStageTriggerID = sbMockTrigger
	st.opts.ScratchDirectory = scratch
	st.opts.OutputDirectory = filepath.Join(scratch, "out")

	st.mock.GetBuildReturns(build, nil)
	st.mock.DownloadStageCalls(func(_, dir string) error {
		return os.MkdirAll(filepath.Join(dir, sbBuildVersion, sbVersion, release.GCSStagePath), 0o750)
	})
	st.mock.FileSubjectsReturns([]*intoto.ResourceDescriptor{
		sbSubject(stagePath(release.SourcesTar), sbDigestC),
		// Written by the release after the stage, so neither attested nor
		// checked.
		sbSubject(stagePath(sbVersion+"/"+release.GCSStagePath+"/"+sbVersion+"/bin/linux/amd64/kubectl.sig"), sbDigestB),
		sbSubject(stagePath(sbVersion+"/"+release.GCSStagePath+"/"+sbVersion+"/bin/linux/amd64/kubectl.cert"), sbDigestB),
		sbSubject(stagePath(sbVersion+"/"+release.GCSStagePath+"/"+sbVersion+"/bin/linux/amd64/kubectl"), sbDigestA),
	}, nil)
	st.mock.ImageSubjectsCalls(func(registry, _, _ string) ([]*intoto.ResourceDescriptor, error) {
		return []*intoto.ResourceDescriptor{sbSubject(registry+"/kube-apiserver-amd64", sbDigestB)}, nil
	})
	st.finalSubjects(sbSubject("gs://"+bucket+"/release/"+sbVersion+"/bin/linux/amd64/kubectl", sbDigestA))
	st.mock.ReadObjectCalls(func(gcsPath string) ([]byte, error) {
		data, ok := st.objects[gcsPath]
		if !ok {
			return nil, os.ErrNotExist
		}

		return data, nil
	})
	st.mock.VerifyProvenanceCalls(func(data []byte, _, _ string) ([]byte, error) {
		payload, ok := st.verified[string(data)]
		if !ok {
			return nil, errors.New("does not verify")
		}

		return payload, nil
	})

	// Everything was written while the stage build ran.
	during := time.Date(2026, 11, 25, 16, 20, 26, 0, time.UTC)
	st.writtenAt = map[string]time.Time{
		stagePath(release.SourcesTar): during,
		stagePath(sbVersion + "/" + release.GCSStagePath + "/" + sbVersion + "/bin/linux/amd64/kubectl"): during,
	}
	st.uploadedAt = map[string]time.Time{"sha256:" + sbDigestB: during}

	st.mock.StagedObjectsCalls(func(gcsPath string) (map[string]release.StagedObject, error) {
		require.Equal(t, stagePath("")+"/", gcsPath)

		objects := map[string]release.StagedObject{}
		for name, at := range st.writtenAt {
			objects[name] = release.StagedObject{Created: at, Updated: at}
		}

		return objects, nil
	})
	st.mock.ImageUploadsCalls(func(string) (map[string]time.Time, error) {
		return st.uploadedAt, nil
	})

	st.sut = release.NewSigningBuild(st.opts)
	st.sut.SetImpl(st.mock)

	return st
}

func (st *signingBuildTest) finalSubjects(subjects ...*intoto.ResourceDescriptor) {
	st.mock.FinalSubjectsCalls(func(string, string, string) ([]*intoto.ResourceDescriptor, error) {
		// Fresh copies, since the signing build sets their digests.
		res := make([]*intoto.ResourceDescriptor, 0, len(subjects))
		for _, s := range subjects {
			res = append(res, sbSubject(s.GetName(), s.GetDigest()["sha256"]))
		}

		return res, nil
	})
}

func TestSigningBuildDryRun(t *testing.T) {
	t.Parallel()

	st := newSigningBuildTest(t, sbStageBuild())
	st.opts.DryRun = true

	require.NoError(t, st.sut.Run())
	require.Zero(t, st.mock.PublishCallCount())
	require.Zero(t, st.mock.ReadObjectCallCount())

	gcsPath, _ := st.mock.DownloadStageArgsForCall(0)
	require.Equal(t, sbStagePath("")+"/", gcsPath)

	out := st.opts.OutputDirectory
	files := sbReadStatement(t, filepath.Join(out, release.SigningBuildProvenanceFilename))
	images := sbReadStatement(t, filepath.Join(out, release.SigningBuildImageProvenanceFilename))
	final := sbReadStatement(t, filepath.Join(out, sbVersion+"-"+release.SigningBuildFinalProvenanceFilename))

	require.Len(t, files.GetSubject(), 2)
	require.Equal(t, []string{sbRegistry + "/kube-apiserver-amd64"}, sbSubjectNames(images))
	require.Equal(t, []string{"gs://" + sbBucket + "/release/v1.38.0-beta.1/bin/linux/amd64/kubectl"}, sbSubjectNames(final))

	predicate := &slsa.Provenance{}
	data, err := protojson.Marshal(files.GetPredicate())
	require.NoError(t, err)
	require.NoError(t, protojson.Unmarshal(data, predicate))

	require.Equal(t, release.SigningBuildType, predicate.GetBuildDefinition().GetBuildType())
	require.Equal(t, release.SigningBuildBuilder, predicate.GetRunDetails().GetBuilder().GetId())
	require.Equal(t, sbBuildID, predicate.GetRunDetails().GetMetadata().GetInvocationId())

	external := predicate.GetBuildDefinition().GetExternalParameters().GetFields()
	require.Equal(t, sbBuildVersion, external["buildVersion"].GetStringValue())
	require.True(t, external["nomock"].GetBoolValue())
	require.Len(t, external["steps"].GetListValue().GetValues(), 3)
	require.Equal(t, sbBuildVersion, external["substitutions"].GetStructValue().GetFields()["_BUILDVERSION"].GetStringValue())
	require.NotContains(t, external["substitutions"].GetStructValue().GetFields(), "_GCP_USER_TAG")

	// The Kubernetes source, krel and the step images, once per name and
	// digest.
	deps := predicate.GetBuildDefinition().GetResolvedDependencies()
	require.Len(t, deps, 4)
	require.Equal(t, "git+https://github.com/kubernetes/kubernetes", deps[0].GetUri())
	require.Equal(t, sbK8sCommit, deps[0].GetDigest()["gitCommit"])
	require.Equal(t, "git+https://github.com/kubernetes/release@"+sbKrelCommit, deps[1].GetUri())
	require.Equal(t, sbKrelCommit, deps[1].GetDigest()["gitCommit"])
}

func TestSigningBuildSignsAndKeeps(t *testing.T) {
	t.Parallel()

	st := newSigningBuildTest(t, sbStageBuild())

	// Nothing signed yet: every statement is signed, the image one attached.
	require.NoError(t, st.sut.Run())
	require.Equal(t, 1, st.mock.PublishCallCount())

	signer, registry, first := st.mock.PublishArgsForCall(0)
	require.Equal(t, release.SigningBuildSigner, signer)
	require.Equal(t, sbRegistry, registry)
	require.ElementsMatch(t, []string{
		sbStagePath(release.SigningBuildProvenanceFilename),
		sbStagePath(release.SigningBuildImageProvenanceFilename),
		sbStagePath(filepath.Join(sbVersion, release.SigningBuildFinalProvenanceFilename)),
	}, sbRemotes(first))

	for _, s := range first {
		require.Nil(t, s.Existing)
		require.Equal(t, strings.HasSuffix(s.Remote, "/"+release.SigningBuildImageProvenanceFilename), s.Attach, s.Remote)

		// The bundles written now verify and hold their statements.
		payload, err := os.ReadFile(s.Local)
		require.NoError(t, err)

		st.objects[s.Remote] = []byte("bundle of " + s.Remote)
		st.verified["bundle of "+s.Remote] = payload
	}

	// A rerun keeps them, and attaches the image provenance again.
	require.NoError(t, st.sut.Run())

	_, _, second := st.mock.PublishArgsForCall(1)
	for _, s := range second {
		require.Equal(t, st.objects[s.Remote], s.Existing, s.Remote)
	}

	// A bundle that doesn't verify, or holds another statement, is signed
	// again.
	delete(st.verified, "bundle of "+sbStagePath(release.SigningBuildProvenanceFilename))
	st.verified["bundle of "+sbStagePath(release.SigningBuildImageProvenanceFilename)] = []byte(
		`{"_type":"https://in-toto.io/Statement/v1"}`,
	)

	require.NoError(t, st.sut.Run())

	_, _, third := st.mock.PublishArgsForCall(2)
	for _, s := range third {
		final := s.Remote == sbStagePath(filepath.Join(sbVersion, release.SigningBuildFinalProvenanceFilename))
		require.Equal(t, final, s.Existing != nil, s.Remote)
	}
}

// sbMockStageBuild returns the record of a mock stage build.
func sbMockStageBuild() *cloudbuild.Build {
	build := sbStageBuild()
	delete(build.Substitutions, "_NOMOCK")
	build.Steps[1].Args = []string{"bin/krel", "stage", ""}
	build.Substitutions["_STAGE_BUCKET"] = "gs://" + release.TestBucket
	build.Substitutions["_STAGE_REGISTRY"] = release.GCRIOPathMock
	build.BuildTriggerId = sbMockTrigger

	return build
}

func TestSigningBuildMockStage(t *testing.T) {
	t.Parallel()

	// The production signer doesn't sign mock stages, even from its trigger.
	build := sbMockStageBuild()
	build.BuildTriggerId = sbTrigger

	st := newSigningBuildTest(t, build)
	require.ErrorContains(t, st.sut.Run(), "is a mock stage")
	require.Zero(t, st.mock.DownloadStageCallCount())
	require.Zero(t, st.mock.PublishCallCount())

	// The mock signer does, from the mock stage trigger.
	build.BuildTriggerId = sbMockTrigger
	st.opts.Signer = release.SigningBuildMockSigner
	require.NoError(t, st.sut.Run())

	signer, registry, _ := st.mock.PublishArgsForCall(0)
	require.Equal(t, release.SigningBuildMockSigner, signer)
	require.Equal(t, release.GCRIOPathMock, registry)
}

func TestSigningBuildChecksWhoRanTheStage(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		build   func() *cloudbuild.Build
		signer  string
		opts    func(*release.SigningBuildOptions)
		wantErr string
	}{
		"submitted by hand": {
			build: func() *cloudbuild.Build {
				b := sbStageBuild()
				b.BuildTriggerId = ""

				return b
			},
			signer:  release.SigningBuildSigner,
			wantErr: "not by the stage trigger " + sbTrigger,
		},
		"another trigger": {
			build: func() *cloudbuild.Build {
				b := sbStageBuild()
				b.BuildTriggerId = "33333333-3333-3333-3333-333333333333"

				return b
			},
			signer:  release.SigningBuildSigner,
			wantErr: "not by the stage trigger",
		},
		"mock trigger for the production signer": {
			build: func() *cloudbuild.Build {
				b := sbStageBuild()
				b.BuildTriggerId = sbMockTrigger

				return b
			},
			signer:  release.SigningBuildSigner,
			wantErr: "not by the stage trigger " + sbTrigger,
		},
		"production trigger for the mock signer": {
			build: func() *cloudbuild.Build {
				b := sbMockStageBuild()
				b.BuildTriggerId = sbTrigger

				return b
			},
			signer:  release.SigningBuildMockSigner,
			wantErr: "not by the stage trigger " + sbMockTrigger,
		},
		"another service account": {
			build: func() *cloudbuild.Build {
				b := sbStageBuild()
				b.ServiceAccount = "projects/k8s-release/serviceAccounts/304687256732@cloudbuild.gserviceaccount.com"

				return b
			},
			signer:  release.SigningBuildSigner,
			wantErr: `ran as "304687256732@cloudbuild.gserviceaccount.com"`,
		},
		"default service account": {
			build: func() *cloudbuild.Build {
				b := sbStageBuild()
				b.ServiceAccount = ""

				return b
			},
			signer:  release.SigningBuildSigner,
			wantErr: `ran as ""`,
		},
		"no trigger ID set": {
			build:   sbStageBuild,
			signer:  release.SigningBuildSigner,
			opts:    func(o *release.SigningBuildOptions) { o.StageTriggerID = "" },
			wantErr: "no stage trigger ID",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			st := newSigningBuildTest(t, tc.build())
			st.opts.Signer = tc.signer

			if tc.opts != nil {
				tc.opts(st.opts)
			}

			require.ErrorContains(t, st.sut.Run(), tc.wantErr)
			require.Zero(t, st.mock.DownloadStageCallCount())
			require.Zero(t, st.mock.PublishCallCount())
		})
	}
}

func TestSigningBuildAcceptsBareServiceAccount(t *testing.T) {
	t.Parallel()

	build := sbStageBuild()
	build.ServiceAccount = release.StageServiceAccount

	require.NoError(t, newSigningBuildTest(t, build).sut.Run())
}

func TestSigningBuildPinsSignerLocations(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		build   func() *cloudbuild.Build
		signer  string
		wantErr string
	}{
		"mock stage pushing to the production registry": {
			build: func() *cloudbuild.Build {
				b := sbMockStageBuild()
				b.Substitutions["_STAGE_REGISTRY"] = release.GCRIOPathStaging

				return b
			},
			signer:  release.SigningBuildMockSigner,
			wantErr: "staged to registry " + release.GCRIOPathStaging,
		},
		"nomock stage in the mock bucket": {
			build: func() *cloudbuild.Build {
				b := sbStageBuild()
				b.Substitutions["_STAGE_BUCKET"] = "gs://" + release.TestBucket

				return b
			},
			signer:  release.SigningBuildSigner,
			wantErr: "staged to bucket " + release.TestBucket,
		},
		"nomock stage for the mock signer": {
			build: func() *cloudbuild.Build {
				b := sbStageBuild()
				b.BuildTriggerId = sbMockTrigger

				return b
			},
			signer:  release.SigningBuildMockSigner,
			wantErr: "is a nomock stage",
		},
		// krel stage writes to production with --nomock, whatever the
		// substitutions say.
		"nomock stage with mock locations for the mock signer": {
			build: func() *cloudbuild.Build {
				b := sbMockStageBuild()
				b.Steps[1].Args = []string{"bin/krel", "stage", "--nomock"}

				return b
			},
			signer:  release.SigningBuildMockSigner,
			wantErr: "is a nomock stage",
		},
		"another signer": {
			build:   sbStageBuild,
			signer:  "someone@example.com",
			wantErr: "neither the production nor the mock signer",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			st := newSigningBuildTest(t, tc.build())
			st.opts.Signer = tc.signer

			require.ErrorContains(t, st.sut.Run(), tc.wantErr)
			require.Zero(t, st.mock.DownloadStageCallCount())
			require.Zero(t, st.mock.PublishCallCount())
		})
	}
}

func TestSigningBuildChecksReleaseFiles(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		subject *intoto.ResourceDescriptor
		wantErr string
	}{
		"digest differs": {
			subject: sbSubject("gs://"+sbBucket+"/release/v1.38.0-beta.1/bin/linux/amd64/kubectl", sbDigestB),
			wantErr: "but the staged file",
		},
		"not staged": {
			subject: sbSubject("gs://"+sbBucket+"/release/v1.38.0-beta.1/bin/linux/amd64/kubeadm", sbDigestA),
			wantErr: "is not staged",
		},
		"elsewhere": {
			subject: sbSubject("gs://other-bucket/release/v1.38.0-beta.1/bin/linux/amd64/kubectl", sbDigestA),
			wantErr: "is not in",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			st := newSigningBuildTest(t, sbStageBuild())
			st.finalSubjects(tc.subject)

			require.ErrorContains(t, st.sut.Run(), tc.wantErr)
			require.Zero(t, st.mock.PublishCallCount())
		})
	}
}

func TestSigningBuildRejectsTampering(t *testing.T) {
	t.Parallel()

	kubectl := sbStagePath(sbVersion + "/" + release.GCSStagePath + "/" + sbVersion + "/bin/linux/amd64/kubectl")
	image := "sha256:" + sbDigestB

	// The stage build ran from 15:56:49.36 to 16:26:50.33.
	for name, tc := range map[string]struct {
		prepare func(*signingBuildTest)
		wantErr string
	}{
		"file replaced after the stage": {
			prepare: func(st *signingBuildTest) { st.writtenAt[kubectl] = time.Date(2026, 11, 25, 17, 0, 0, 0, time.UTC) },
			wantErr: "outside of the stage build",
		},
		"file older than the stage": {
			prepare: func(st *signingBuildTest) { st.writtenAt[kubectl] = time.Date(2026, 11, 25, 15, 0, 0, 0, time.UTC) },
			wantErr: "outside of the stage build",
		},
		"file not in the bucket": {
			prepare: func(st *signingBuildTest) { delete(st.writtenAt, kubectl) },
			wantErr: "is not in the bucket",
		},
		"tag moved to an image uploaded later": {
			prepare: func(st *signingBuildTest) { st.uploadedAt[image] = time.Date(2026, 11, 25, 17, 0, 0, 0, time.UTC) },
			wantErr: "was uploaded at",
		},
		"tag moved to an older image": {
			prepare: func(st *signingBuildTest) { st.uploadedAt[image] = time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC) },
			wantErr: "was uploaded at",
		},
		"image not in its repository": {
			prepare: func(st *signingBuildTest) { delete(st.uploadedAt, image) },
			wantErr: "is not in its repository",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			st := newSigningBuildTest(t, sbStageBuild())
			tc.prepare(st)

			require.ErrorContains(t, st.sut.Run(), tc.wantErr)
			require.Zero(t, st.mock.PublishCallCount())
		})
	}
}

func TestSigningBuildStageWindowSlack(t *testing.T) {
	t.Parallel()

	// GCS reports whole seconds: a file written in the first second of the
	// stage build, or the last one, is within it.
	st := newSigningBuildTest(t, sbStageBuild())
	st.writtenAt[sbStagePath(release.SourcesTar)] = time.Date(2026, 11, 25, 15, 56, 49, 0, time.UTC)
	st.uploadedAt["sha256:"+sbDigestB] = time.Date(2026, 11, 25, 16, 26, 51, 0, time.UTC)

	require.NoError(t, st.sut.Run())
}

func TestSigningBuildFailures(t *testing.T) {
	t.Parallel()

	for name, prepare := range map[string]func(*releasefakes.FakeSigningBuildImpl){
		"get build": func(m *releasefakes.FakeSigningBuildImpl) { m.GetBuildReturns(nil, errors.New("boom")) },
		"download": func(m *releasefakes.FakeSigningBuildImpl) {
			m.DownloadStageCalls(nil)
			m.DownloadStageReturns(errors.New("boom"))
		},
		"file subjects": func(m *releasefakes.FakeSigningBuildImpl) { m.FileSubjectsReturns(nil, errors.New("boom")) },
		"image subjects": func(m *releasefakes.FakeSigningBuildImpl) {
			m.ImageSubjectsCalls(nil)
			m.ImageSubjectsReturns(nil, errors.New("boom"))
		},
		"final subjects": func(m *releasefakes.FakeSigningBuildImpl) {
			m.FinalSubjectsCalls(nil)
			m.FinalSubjectsReturns(nil, errors.New("boom"))
		},
		"read object": func(m *releasefakes.FakeSigningBuildImpl) {
			m.ReadObjectCalls(nil)
			m.ReadObjectReturns(nil, errors.New("boom"))
		},
		"publish": func(m *releasefakes.FakeSigningBuildImpl) { m.PublishReturns(errors.New("boom")) },
		"staged objects": func(m *releasefakes.FakeSigningBuildImpl) {
			m.StagedObjectsCalls(nil)
			m.StagedObjectsReturns(nil, errors.New("boom"))
		},
		"image uploads": func(m *releasefakes.FakeSigningBuildImpl) {
			m.ImageUploadsCalls(nil)
			m.ImageUploadsReturns(nil, errors.New("boom"))
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			st := newSigningBuildTest(t, sbStageBuild())
			prepare(st.mock)

			require.ErrorContains(t, st.sut.Run(), "boom")
		})
	}
}

func sbReadStatement(t *testing.T, path string) *intoto.Statement {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	s := &intoto.Statement{}
	require.NoError(t, protojson.Unmarshal(data, s))
	require.NoError(t, s.Validate())

	return s
}

func sbRemotes(statements []*release.SigningBuildStatement) []string {
	res := make([]string, 0, len(statements))
	for _, s := range statements {
		res = append(res, s.Remote)
	}

	return res
}

func sbSubjectNames(s *intoto.Statement) []string {
	res := make([]string, 0, len(s.GetSubject()))
	for _, sub := range s.GetSubject() {
		res = append(res, sub.GetName())
	}

	return res
}

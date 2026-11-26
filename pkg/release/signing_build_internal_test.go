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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/api/cloudbuild/v1"
)

const (
	testBuildID      = "54ef86d9-5a89-4e0e-a008-10904a0f5a34"
	testBuildVersion = "v1.38.0-beta.0.899+9f140ee0061e18"
	testVersion      = "v1.38.0-beta.1"
	testBucket       = "test-bucket"
	testKrelCommit   = "9c79ae380875188f191f7d30ee01389772a10f3d"
	testK8sCommit    = "9f140ee0061e181fa296a4257dda0044a63f1f6c"
	testDigestA      = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testDigestB      = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	testDigestC      = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

// testStageBuild returns the Cloud Build record of a nomock stage build.
func testStageBuild() *cloudbuild.Build {
	return &cloudbuild.Build{
		Id:             testBuildID,
		ProjectId:      "k8s-release",
		BuildTriggerId: "11111111-1111-1111-1111-111111111111",
		ServiceAccount: "projects/k8s-release/serviceAccounts/krel-stage@k8s-release.iam.gserviceaccount.com",
		Status:         "SUCCESS",
		Tags:           []string{"nomock", "STAGE", "1.38.0-beta.1"},
		StartTime:      "2026-11-25T15:56:49.360329688Z",
		FinishTime:     "2026-11-25T16:26:50.339390Z",
		Substitutions: map[string]string{
			"_BUILDVERSION":   testBuildVersion,
			"_STAGE_BUCKET":   "gs://" + testBucket,
			"_STAGE_REGISTRY": "gcr.io/k8s-staging-kubernetes",
			"_TYPE":           "beta",
			"_RELEASE_BRANCH": "master",
			"_NOMOCK":         "--nomock",
			"_TOOL_ORG":       "kubernetes",
			"_TOOL_REPO":      "release",
			"_TOOL_REF":       testKrelCommit,
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
			BuildStepImages: []string{"sha256:" + testDigestA, "sha256:" + testDigestB, "sha256:" + testDigestB},
		},
		SourceProvenance: &cloudbuild.SourceProvenance{
			ResolvedGitSource: &cloudbuild.GitSource{
				Url: "https://github.com/kubernetes/kubernetes", Revision: testK8sCommit,
			},
		},
	}
}

func TestParseStageBuild(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		change  func(*cloudbuild.Build)
		wantErr string
	}{
		"stage":         {change: func(*cloudbuild.Build) {}},
		"failed":        {change: func(b *cloudbuild.Build) { b.Status = "FAILURE" }, wantErr: "not SUCCESS"},
		"not tagged":    {change: func(b *cloudbuild.Build) { b.Tags = []string{"nomock"} }, wantErr: "no STAGE tag"},
		"no krel stage": {change: func(b *cloudbuild.Build) { b.Steps = b.Steps[:1] }, wantErr: "krel stage"},
		"no bucket":     {change: func(b *cloudbuild.Build) { b.Substitutions["_STAGE_BUCKET"] = "bucket" }, wantErr: "not a gs:// bucket"},
		"bucket path":   {change: func(b *cloudbuild.Build) { b.Substitutions["_STAGE_BUCKET"] = "gs://a/b" }, wantErr: "not a gs:// bucket"},
		"no version":    {change: func(b *cloudbuild.Build) { delete(b.Substitutions, "_BUILDVERSION") }, wantErr: "_BUILDVERSION"},
		"version path":  {change: func(b *cloudbuild.Build) { b.Substitutions["_BUILDVERSION"] = "../x" }, wantErr: "is not one"},
		"no start time": {change: func(b *cloudbuild.Build) { b.StartTime = "" }, wantErr: "start time"},
		"no trigger":    {change: func(b *cloudbuild.Build) { b.BuildTriggerId = "" }, wantErr: "not by the stage trigger"},
		"other account": {change: func(b *cloudbuild.Build) { b.ServiceAccount = "someone@example.com" }, wantErr: `ran as "someone@example.com"`},
		"no registry":   {change: func(b *cloudbuild.Build) { delete(b.Substitutions, "_STAGE_REGISTRY") }, wantErr: "_STAGE_REGISTRY"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			build := testStageBuild()
			tc.change(build)

			stage, err := parseStageBuild(build, "11111111-1111-1111-1111-111111111111")
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, &stageBuild{
				buildVersion: testBuildVersion, bucket: testBucket, registry: "gcr.io/k8s-staging-kubernetes",
				releaseType: "beta", releaseBranch: "master", nomock: true,
				started:  time.Date(2026, 11, 25, 15, 56, 49, 360329688, time.UTC),
				finished: time.Date(2026, 11, 25, 16, 26, 50, 339390000, time.UTC),
			}, stage)
		})
	}
}

func TestResolvedDependenciesWithoutSource(t *testing.T) {
	t.Parallel()

	// A stage without a git source, like the ones before Kubernetes became
	// their Cloud Build source, and krel from a branch.
	build := testStageBuild()
	build.SourceProvenance = nil
	build.Substitutions["_TOOL_REF"] = "master"

	deps := resolvedDependencies(build)
	require.Equal(t, "git+https://github.com/kubernetes/kubernetes@HEAD", deps[0].GetUri())
	require.Empty(t, deps[0].GetDigest())
	require.Equal(t, "git+https://github.com/kubernetes/release@master", deps[1].GetUri())
	require.Empty(t, deps[1].GetDigest())
	require.Len(t, deps, 4)
}

func TestResolvedDependenciesTriggerSource(t *testing.T) {
	t.Parallel()

	// A trigger's source is the repository of its config: it is not the
	// Kubernetes source.
	build := testStageBuild()
	build.SourceProvenance.ResolvedGitSource.Url = "https://github.com/kubernetes/release"

	deps := resolvedDependencies(build)
	require.Equal(t, "git+https://github.com/kubernetes/kubernetes@HEAD", deps[0].GetUri())
	require.Empty(t, deps[0].GetDigest())

	// With Kubernetes as its source, the commit is recorded.
	build.SourceProvenance.ResolvedGitSource.Url = "https://github.com/kubernetes/kubernetes.git"

	deps = resolvedDependencies(build)
	require.Equal(t, "git+https://github.com/kubernetes/kubernetes", deps[0].GetUri())
	require.Equal(t, testK8sCommit, deps[0].GetDigest()["gitCommit"])
}

func TestStageStepNomock(t *testing.T) {
	t.Parallel()

	stage := func(args ...string) []*cloudbuild.BuildStep {
		return []*cloudbuild.BuildStep{{Args: append([]string{"bin/krel", "stage"}, args...)}}
	}

	require.True(t, stageStepNomock(stage("--submit=false", "--nomock")))
	require.True(t, stageStepNomock(stage("--nomock=true")))
	require.False(t, stageStepNomock(stage("--submit=false", "")))
	require.False(t, stageStepNomock([]*cloudbuild.BuildStep{{Args: []string{"bin/krel", "sign", "--nomock"}}}))
}

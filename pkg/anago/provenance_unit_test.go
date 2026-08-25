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

package anago

import (
	"testing"
	"time"

	slsa "github.com/in-toto/attestation/go/predicates/provenance/v1"
	intoto "github.com/in-toto/attestation/go/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"sigs.k8s.io/release-sdk/git"
	"sigs.k8s.io/release-utils/command"

	buildtypev1 "k8s.io/release/pkg/anago/buildtypes/v1"
	"k8s.io/release/pkg/release"
)

func TestProvenanceStructRoundTrip(t *testing.T) {
	t.Parallel()

	original := &slsa.Provenance{
		BuildDefinition: &slsa.BuildDefinition{
			BuildType: "https://example.com/buildtype@v1",
			ResolvedDependencies: []*intoto.ResourceDescriptor{{
				Uri:    "git+https://github.com/kubernetes/kubernetes",
				Digest: map[string]string{"gitCommit": "abcdef"},
			}},
		},
		RunDetails: &slsa.RunDetails{
			Builder: &slsa.Builder{Id: "https://example.com/builder"},
			Metadata: &slsa.BuildMetadata{
				InvocationId: "1234",
				StartedOn:    timestamppb.New(time.Date(2026, 8, 22, 1, 2, 3, 0, time.UTC)),
			},
		},
	}

	asStruct, err := protoToStruct(original, false)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/buildtype@v1",
		asStruct.GetFields()["buildDefinition"].GetStructValue().GetFields()["buildType"].GetStringValue(),
	)

	// Unset fields of the predicate must not be recorded as empty values
	depFields := asStruct.GetFields()["buildDefinition"].GetStructValue().
		GetFields()["resolvedDependencies"].GetListValue().GetValues()[0].GetStructValue().GetFields()
	require.Contains(t, depFields, "uri")
	require.Contains(t, depFields, "digest")
	require.NotContains(t, depFields, "name")
	require.NotContains(t, depFields, "mediaType")
	require.NotContains(t, depFields, "downloadLocation")

	parsed, err := provenanceFromStruct(asStruct)
	require.NoError(t, err)
	require.True(t, proto.Equal(original, parsed), "round-tripped predicate differs from original")

	// A missing predicate must not fail, returns empty provenance
	parsed, err = provenanceFromStruct(nil)
	require.NoError(t, err)
	require.Nil(t, parsed.GetRunDetails())
}

func TestExternalParametersToStruct(t *testing.T) {
	t.Parallel()

	params, err := protoToStruct(&buildtypev1.ExternalParameters{
		ReleaseType:   "alpha",
		ReleaseBranch: "master",
		BuildVersion:  "v1.36.0-alpha.1.10+abcdef",
		Nomock:        false,
	}, true)
	require.NoError(t, err)

	fields := params.GetFields()
	require.Equal(t, "alpha", fields["releaseType"].GetStringValue())
	require.Equal(t, "master", fields["releaseBranch"].GetStringValue())
	require.Equal(t, "v1.36.0-alpha.1.10+abcdef", fields["buildVersion"].GetStringValue())

	// Default values must be recorded explicitly in the parameters
	require.Contains(t, fields, "nomock")
	require.False(t, fields["nomock"].GetBoolValue())
	require.Contains(t, fields, "entryPoint")
}

// testGit runs a git command in dir with a fixed identity and no signing,
// returning its trimmed output.
func testGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	out, err := command.NewWithWorkDir(dir, "git", append([]string{
		"-c", "user.name=test", "-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false",
	}, args...)...).RunSilentSuccessOutput()
	require.NoError(t, err)

	return out.OutputTrimNL()
}

func TestSourceDependencies(t *testing.T) {
	t.Parallel()

	// A repository with a build point commit tagged as alpha and an
	// empty release commit on top of it tagged as the official release,
	// as done on release branches.
	dir := t.TempDir()
	testGit(t, dir, "init", "-q", "-b", "release-1.29")
	testGit(t, dir, "commit", "-q", "--allow-empty", "-m", "build point")
	buildPointSHA := testGit(t, dir, "rev-parse", "HEAD")
	testGit(t, dir, "tag", "-a", "v1.30.0-alpha.1", "-m", "alpha")
	testGit(t, dir, "commit", "-q", "--allow-empty", "-m", "release commit")
	releaseSHA := testGit(t, dir, "rev-parse", "HEAD")
	testGit(t, dir, "tag", "-a", "v1.29.1", "-m", "official")

	repo, err := git.OpenRepo(dir)
	require.NoError(t, err)

	repoURI := "git+" + git.GetRepoURL(release.GetK8sOrg(), release.GetK8sRepo(), false)
	digest := func(sha string) map[string]string {
		return map[string]string{"sha1": sha, "gitCommit": sha}
	}

	t.Run("one descriptor per release tag", func(t *testing.T) {
		t.Parallel()

		state := DefaultStageState()
		state.versions = release.NewReleaseVersions("", "v1.29.1", "", "", "v1.30.0-alpha.1")

		sources, err := sourceDependencies(repo, state)
		require.NoError(t, err)
		require.Len(t, sources, 2)

		// The release tags in order, resolved to the tagged commits
		require.Equal(t, repoURI+"@refs/tags/v1.29.1", sources[0].GetUri())
		require.Equal(t, digest(releaseSHA), sources[0].GetDigest())
		require.Equal(t, repoURI+"@refs/tags/v1.30.0-alpha.1", sources[1].GetUri())
		require.Equal(t, digest(buildPointSHA), sources[1].GetDigest())
	})

	t.Run("no versions", func(t *testing.T) {
		t.Parallel()

		_, err := sourceDependencies(repo, DefaultStageState())
		require.Error(t, err)
	})

	t.Run("missing tag", func(t *testing.T) {
		t.Parallel()

		state := DefaultStageState()
		state.versions = release.NewReleaseVersions("", "v9.9.9", "", "", "")

		_, err := sourceDependencies(repo, state)
		require.Error(t, err)
	})
}

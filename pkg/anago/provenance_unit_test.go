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

	buildtypev1 "k8s.io/release/pkg/anago/buildtypes/v1"
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

	asStruct, err := protoToStruct(original)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/buildtype@v1",
		asStruct.GetFields()["buildDefinition"].GetStructValue().GetFields()["buildType"].GetStringValue(),
	)

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
	})
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

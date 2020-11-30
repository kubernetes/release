/*
Copyright 2019 The Kubernetes Authors.

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

package audit_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/cip/audit"
	reg "k8s.io/release/pkg/cip/dockerregistry"
	"k8s.io/release/pkg/cip/logclient"
	"k8s.io/release/pkg/cip/remotemanifest"
	"k8s.io/release/pkg/cip/report"
	"k8s.io/release/pkg/cip/stream"
)

func TestParsePubSubMessageBody(t *testing.T) {
	toPsm := func(s string) audit.PubSubMessage {
		return audit.PubSubMessage{
			Message: audit.PubSubMessageInner{
				Data: []byte(s),
				ID:   "1",
			},
			Subscription: "2",
		}
	}

	shouldBeValid := []struct {
		name string
		body string
	}{
		{
			"nothing (missing keys, but is OK for just parsing from JSON)",
			`{}`,
		},
		{
			"regular insert, FQIN only",
			`{"action": "INSERT", "digest": "gcr.io/foo/bar@sha256:000"}`,
		},
		{
			"regular insert with both FQIN and PQIN",
			`{"action": "INSERT", "digest": "gcr.io/foo/bar@sha256:000", "tag":"gcr.io/foo/bar:1.0"}`,
		},
		{
			"deletion",
			`{"action": "DELETE", "tag": "gcr.io/foo/bar:1.0"}`,
		},
	}

	for _, test := range shouldBeValid {
		psm := toPsm(test.body)
		psmBytes, err := json.Marshal(psm)
		require.Nil(t, err)

		_, gotErr := audit.ParsePubSubMessageBody(psmBytes)
		require.Nil(t, gotErr)
	}

	shouldBeInvalid := []struct {
		name        string
		body        string
		expectedErr error
	}{
		{
			"malformed JSON",
			`{`,
			fmt.Errorf("json.Unmarshal (message data): unexpected end of JSON input"),
		},
		{
			"incompatible type (int for string)",
			`{"action": 1}`,
			fmt.Errorf("json.Unmarshal (message data): json: cannot unmarshal number into Go struct field GCRPubSubPayload.action of type string"),
		},
	}

	for _, test := range shouldBeInvalid {
		psm := toPsm(test.body)
		psmBytes, err := json.Marshal(psm)
		require.Nil(t, err)

		_, gotErr := audit.ParsePubSubMessageBody(psmBytes)
		require.Equal(t, gotErr, test.expectedErr)
	}
}

func TestValidatePayload(t *testing.T) {
	shouldBeValid := []reg.GCRPubSubPayload{
		{
			Action: "INSERT",
			FQIN:   "gcr.io/foo/bar@sha256:0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			Action: "INSERT",
			FQIN:   "gcr.io/foo/bar@sha256:0000000000000000000000000000000000000000000000000000000000000000",
			PQIN:   "gcr.io/foo/bar:1.0",
		},
	}

	for _, input := range shouldBeValid {
		input := input
		gotErr := audit.ValidatePayload(&input)
		require.Nil(t, gotErr)
	}

	shouldBeInValid := []struct {
		input    reg.GCRPubSubPayload
		expected error
	}{
		{
			reg.GCRPubSubPayload{
				Action: "INSERT",
			},
			fmt.Errorf(
				`{Action: "INSERT", FQIN: "", PQIN: "", Path: "", Digest: "", Tag: ""}: neither 'digest' nor 'tag' was specified`,
			),
		},
		{
			reg.GCRPubSubPayload{
				FQIN: "gcr.io/foo/bar@sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
			fmt.Errorf(
				`{Action: "", FQIN: "gcr.io/foo/bar@sha256:0000000000000000000000000000000000000000000000000000000000000000", PQIN: "", Path: "gcr.io/foo/bar", Digest: "sha256:0000000000000000000000000000000000000000000000000000000000000000", Tag: ""}: Action not specified`,
			),
		},
		{
			reg.GCRPubSubPayload{
				Action: "DELETE",
				FQIN:   "gcr.io/foo/bar@sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
			fmt.Errorf(
				`{Action: "DELETE", FQIN: "gcr.io/foo/bar@sha256:0000000000000000000000000000000000000000000000000000000000000000", PQIN: "", Path: "gcr.io/foo/bar", Digest: "sha256:0000000000000000000000000000000000000000000000000000000000000000", Tag: ""}: deletions are prohibited`,
			),
		},
		{
			reg.GCRPubSubPayload{
				Action: "WOOF",
				FQIN:   "gcr.io/foo/bar@sha256:0000000000000000000000000000000000000000000000000000000000000000",
			},
			fmt.Errorf(
				`{Action: "WOOF", FQIN: "gcr.io/foo/bar@sha256:0000000000000000000000000000000000000000000000000000000000000000", PQIN: "", Path: "gcr.io/foo/bar", Digest: "sha256:0000000000000000000000000000000000000000000000000000000000000000", Tag: ""}: unknown action "WOOF"`,
			),
		},
	}

	for _, test := range shouldBeInValid {
		gotErr := audit.ValidatePayload(&test.input)
		require.Equal(t, gotErr, test.expected)
	}
}

// nolint[gocyclo]
func TestAudit(t *testing.T) {
	// Regression test case for
	// https://github.com/kubernetes-sigs/k8s-container-image-promoter/issues/191.
	manifests1 := []reg.Manifest{
		{
			Registries: []reg.RegistryContext{
				{
					Name: "gcr.io/k8s-staging-kas-network-proxy",
					Src:  true,
				},
				{
					Name:           "us.gcr.io/k8s-artifacts-prod/kas-network-proxy",
					ServiceAccount: "foobar@google-containers.iam.gserviceaccount.com",
				},
			},

			Images: []reg.Image{
				{
					ImageName: "proxy-agent",
					Dmap: reg.DigestTags{
						"sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32": {"v0.0.8"},
					},
				},
			},
		},
	}

	readRepo1 := map[string]string{
		"gcr.io/k8s-staging-kas-network-proxy": `{
  "child": [
    "proxy-agent"
  ],
  "manifest": {},
  "name": "k8s-staging-kas-network-proxy",
  "tags": []
}`,
		"gcr.io/k8s-staging-kas-network-proxy/proxy-agent": `{
  "child": [],
  "manifest": {
    "sha256:43273b274ee48f7fd7fc09bc82e7e75ddc596ca219fd9b522b1701bebec6ceff": {
      "imageSizeBytes": "6843680",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1583451840426",
      "timeUploadedMs": "1583475320110"
    },
    "sha256:7bcbdf4cb26400ac576b33718000f0b630290dcf6380be3f60e33e5ba0461d31": {
      "imageSizeBytes": "7367874",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1583451717939",
      "timeUploadedMs": "1583475314214"
    },
    "sha256:8735603bbd7153b8bfc8d2460481282bb44e2e830e5b237738e5c3e2a58c8f45": {
      "imageSizeBytes": "7396163",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1583451882087",
      "timeUploadedMs": "1583475321761"
    },
    "sha256:99bade313218f3e6e63fdeb87bcddbf3a134aaa9e45e633be5ee5e60ddaac667": {
      "imageSizeBytes": "6888230",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1583451799250",
      "timeUploadedMs": "1583475318193"
    },
    "sha256:c1ccf44d6b6fe49fc8506f7571f4a988ad69eb00c7747cd2b307b5e5b125a1f1": {
      "imageSizeBytes": "6888983",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1583451758583",
      "timeUploadedMs": "1583475316361"
    },
    "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32": {
      "imageSizeBytes": "0",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
      "tag": [
        "v0.0.8"
      ],
      "timeCreatedMs": "0",
      "timeUploadedMs": "1583475321879"
    }
  },
  "name": "k8s-staging-kas-network-proxy/proxy-agent",
  "tags": [
    "v0.0.8"
  ]
}`,
	}

	// This is the response for reading the manifest for the parent
	// image by digest.
	readManifestList1 := map[string]string{
		"gcr.io/k8s-staging-kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32": `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
   "manifests": [
      {
         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
         "size": 528,
         "digest": "sha256:7bcbdf4cb26400ac576b33718000f0b630290dcf6380be3f60e33e5ba0461d31",
         "platform": {
            "architecture": "amd64",
            "os": "linux"
         }
      },
      {
         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
         "size": 528,
         "digest": "sha256:c1ccf44d6b6fe49fc8506f7571f4a988ad69eb00c7747cd2b307b5e5b125a1f1",
         "platform": {
            "architecture": "arm",
            "os": "linux"
         }
      },
      {
         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
         "size": 528,
         "digest": "sha256:99bade313218f3e6e63fdeb87bcddbf3a134aaa9e45e633be5ee5e60ddaac667",
         "platform": {
            "architecture": "arm64",
            "os": "linux"
         }
      },
      {
         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
         "size": 528,
         "digest": "sha256:43273b274ee48f7fd7fc09bc82e7e75ddc596ca219fd9b522b1701bebec6ceff",
         "platform": {
            "architecture": "ppc64le",
            "os": "linux"
         }
      },
      {
         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
         "size": 528,
         "digest": "sha256:8735603bbd7153b8bfc8d2460481282bb44e2e830e5b237738e5c3e2a58c8f45",
         "platform": {
            "architecture": "s390x",
            "os": "linux"
         }
      }
   ]
}`,
	}

	// Regression test for the case where we are promoting
	// bcdd5657b1edc1a2eb27356f33dd66b9400d4a084209c33461c7a7da0a32ebb3 (etcd
	// tag "3.4.7-2"). It is crucial that we read fom *ALL* source registries,
	// not just the first one that has a match for the destination registry
	// "us.gcr.io/k8s-artifacts-prod" (because both manifests below promote to
	// that same destination, but only one of them have the child images of
	// 3.4.7-2).
	manifests2 := []reg.Manifest{
		{
			Registries: []reg.RegistryContext{
				{
					Name: "gcr.io/google-containers",
					Src:  true,
				},
				{
					// Root promotion
					Name:           "us.gcr.io/k8s-artifacts-prod",
					ServiceAccount: "foobar@google-containers.iam.gserviceaccount.com",
				},
			},

			Images: []reg.Image{
				{
					ImageName: "etcd",
					Dmap: reg.DigestTags{
						"sha256:12f377200949c25fde1e54bba639d34d119edd7cfcfb1d117526dba677c03c85": {"3.4.7", "3.4.7-0"},
					},
				},
			},
		},
		{
			Registries: []reg.RegistryContext{
				{
					Name: "gcr.io/k8s-staging-etcd",
					Src:  true,
				},
				{
					// Root promotion
					Name:           "us.gcr.io/k8s-artifacts-prod",
					ServiceAccount: "foobar@google-containers.iam.gserviceaccount.com",
				},
				{
					// Non-root promotion
					Name:           "us.gcr.io/k8s-artifacts-prod/kubernetes",
					ServiceAccount: "foobar@google-containers.iam.gserviceaccount.com",
				},
			},

			Images: []reg.Image{
				{
					ImageName: "etcd",
					Dmap: reg.DigestTags{
						"sha256:bcdd5657b1edc1a2eb27356f33dd66b9400d4a084209c33461c7a7da0a32ebb3": {"3.4.7-2"},
					},
				},
			},
		},
	}

	readRepo2 := map[string]string{
		"gcr.io/k8s-staging-etcd": `{
  "child": [
    "etcd"
  ],
  "manifest": {},
  "name": "k8s-staging-etcd",
  "tags": []
}`,
		"gcr.io/k8s-staging-etcd/etcd": `{
  "child": [],
  "manifest": {
    "sha256:0873d877318546c6569e1abfafd75e0625c202d24435299c4d2e57eeebea52ee": {
      "imageSizeBytes": "86603886",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1589951279359",
      "timeUploadedMs": "1589952169591"
    },
    "sha256:18f3242ebdefb8c2cbb9da24bb1845001f031c222a06255a06a57b541f7b45ad": {
      "imageSizeBytes": "134330709",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1589951693859",
      "timeUploadedMs": "1589952171472"
    },
    "sha256:2fb9a8348e5318142ea54c031bfc294c1042009d8f141e0e1d41c332386cc299": {
      "imageSizeBytes": "133303788",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1589951904418",
      "timeUploadedMs": "1589952172561"
    },
    "sha256:54654da17593ef5e930e57e6ff4e03c0139aeeb0e2f3a6f4f7de248a937369e7": {
      "imageSizeBytes": "134756383",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1589951485669",
      "timeUploadedMs": "1589952170488"
    },
    "sha256:bcdd5657b1edc1a2eb27356f33dd66b9400d4a084209c33461c7a7da0a32ebb3": {
      "imageSizeBytes": "0",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
      "tag": [
        "3.4.7-2"
      ],
      "timeCreatedMs": "0",
      "timeUploadedMs": "1589952174599"
    },
    "sha256:edc07fe4241d4d745fdd4aaf4bbef4a8568a427693ff7af9e6572335b45c272f": {
      "imageSizeBytes": "143520849",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1589952118346",
      "timeUploadedMs": "1589952173579"
    }
  },
  "name": "k8s-staging-etcd/etcd",
  "tags": [
    "3.4.7-2"
  ]
}`,
		"gcr.io/google-containers": `{
  "child": [
    "etcd"
  ],
  "manifest": {},
  "name": "google-containers",
  "tags": []
}`,
		"gcr.io/google-containers/etcd": `{
  "child": [],
  "manifest": {
    "sha256:12f377200949c25fde1e54bba639d34d119edd7cfcfb1d117526dba677c03c85": {
      "imageSizeBytes": "0",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
      "tag": [
        "3.4.7",
        "3.4.7-0"
      ],
      "timeCreatedMs": "0",
      "timeUploadedMs": "1586803705534"
    },
    "sha256:a5250021a52e8d2300b6c1c5111a12a3b2f70c463eac9e628e9589578c25cd7a": {
      "imageSizeBytes": "104216218",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1586542481169",
      "timeUploadedMs": "1586803661224"
    },
    "sha256:50ca5007f4def90e14c5558481a0ed4049ec9f172c92f7d20206c2ccedab6fcf": {
      "imageSizeBytes": "154231381",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1586542603683",
      "timeUploadedMs": "1586803673694"
    },
    "sha256:969f6212eb70c23c8c2305f7972940129302efc8bafd81dbf610e47c400acaa3": {
      "imageSizeBytes": "155511148",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1586542725090",
      "timeUploadedMs": "1586803683899"
    },
    "sha256:9da2721aa7dcfd425a2c2d1dc22c59cd80f1f556f677fc78f36d88fbbbbb155f": {
      "imageSizeBytes": "158254458",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1586451856473",
      "timeUploadedMs": "1586803695175"
    },
    "sha256:7fe05534ed4147675e18179a628743e6be966bba794a6cf8a1b88763b8ae0169": {
      "imageSizeBytes": "169795464",
      "layerId": "",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "tag": [],
      "timeCreatedMs": "1586542977948",
      "timeUploadedMs": "1586803705133"
    }
  },
  "name": "google-containers/etcd",
  "tags": [
    "3.4.7",
    "3.4.7-0"
  ]
}`,
	}

	readManifestList2 := map[string]string{
		"gcr.io/k8s-staging-etcd/etcd@sha256:bcdd5657b1edc1a2eb27356f33dd66b9400d4a084209c33461c7a7da0a32ebb3": `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "manifests": [
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 1372,
      "digest": "sha256:0873d877318546c6569e1abfafd75e0625c202d24435299c4d2e57eeebea52ee",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 1372,
      "digest": "sha256:54654da17593ef5e930e57e6ff4e03c0139aeeb0e2f3a6f4f7de248a937369e7",
      "platform": {
        "architecture": "arm",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 1372,
      "digest": "sha256:18f3242ebdefb8c2cbb9da24bb1845001f031c222a06255a06a57b541f7b45ad",
      "platform": {
        "architecture": "arm64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 1373,
      "digest": "sha256:2fb9a8348e5318142ea54c031bfc294c1042009d8f141e0e1d41c332386cc299",
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 1372,
      "digest": "sha256:edc07fe4241d4d745fdd4aaf4bbef4a8568a427693ff7af9e6572335b45c272f",
      "platform": {
        "architecture": "s390x",
        "os": "linux"
      }
    }
  ]
}`,
		"gcr.io/google-containers/etcd@sha256:12f377200949c25fde1e54bba639d34d119edd7cfcfb1d117526dba677c03c85": `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "manifests": [
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 952,
      "digest": "sha256:a5250021a52e8d2300b6c1c5111a12a3b2f70c463eac9e628e9589578c25cd7a",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 953,
      "digest": "sha256:50ca5007f4def90e14c5558481a0ed4049ec9f172c92f7d20206c2ccedab6fcf",
      "platform": {
        "architecture": "arm",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 953,
      "digest": "sha256:969f6212eb70c23c8c2305f7972940129302efc8bafd81dbf610e47c400acaa3",
      "platform": {
        "architecture": "arm64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 953,
      "digest": "sha256:9da2721aa7dcfd425a2c2d1dc22c59cd80f1f556f677fc78f36d88fbbbbb155f",
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 953,
      "digest": "sha256:7fe05534ed4147675e18179a628743e6be966bba794a6cf8a1b88763b8ae0169",
      "platform": {
        "architecture": "s390x",
        "os": "linux"
      }
    }
  ]
}`,
	}

	type expectedPatterns struct {
		report []string
		info   []string
		error  []string
		alert  []string
	}

	var shouldBeValid = []struct {
		name             string
		manifests        []reg.Manifest
		payload          reg.GCRPubSubPayload
		readRepo         map[string]string
		readManifestList map[string]string
		expectedPatterns expectedPatterns
	}{
		{
			"direct manifest (tagless image)",
			manifests1,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32",
				PQIN:   "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:v0.0.8",
			},
			readRepo1,
			readManifestList1,
			expectedPatterns{
				report: nil,
				info:   []string{`TRANSACTION VERIFIED: {Action: "INSERT", FQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", PQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:v0.0.8", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent", Digest: "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", Tag: "v0.0.8"}: agrees with manifest`},
				error:  nil,
				alert:  nil,
			},
		},
		{
			"child manifest (tagless child image, digest not in promoter manifest, but parent image is in promoter manifest)",
			manifests1,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:8735603bbd7153b8bfc8d2460481282bb44e2e830e5b237738e5c3e2a58c8f45",
				PQIN:   "",
			},
			readRepo1,
			readManifestList1,
			expectedPatterns{
				report: nil,
				info:   []string{`TRANSACTION VERIFIED: {Action: "INSERT", FQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:8735603bbd7153b8bfc8d2460481282bb44e2e830e5b237738e5c3e2a58c8f45", PQIN: "", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent", Digest: "sha256:8735603bbd7153b8bfc8d2460481282bb44e2e830e5b237738e5c3e2a58c8f45", Tag: ""}: agrees with manifest \(parent digest sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32\)`},
				error:  nil,
				alert:  nil,
			},
		},
		{
			"image not found (no path match, even though the digest is found in the promoter manifest)",
			manifests1,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent-white-powder@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32",
				PQIN:   "",
			},
			readRepo1,
			readManifestList1,
			expectedPatterns{
				report: []string{`TRANSACTION REJECTED: could not find matching source registry for us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent-white-powder@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32`},
				info:   []string{`could not find direct manifest entry for {Action: "INSERT", FQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent-white-powder@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", PQIN: "", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent-white-powder", Digest: "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", Tag: ""}; assuming child manifest`},
				error:  nil,
				alert:  []string{`TRANSACTION REJECTED: could not find matching source registry for us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent-white-powder@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32`},
			},
		},
		{
			"image not found (path and digest match, but not the tag)",
			manifests1,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32",
				PQIN:   "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:evil",
			},
			readRepo1,
			readManifestList1,
			expectedPatterns{
				report: []string{`TRANSACTION REJECTED: {Action: "INSERT", FQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", PQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:evil", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent", Digest: "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", Tag: "evil"}: could not validate`},
				info:   []string{`could not find direct manifest entry for {Action: "INSERT", FQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", PQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:evil", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent", Digest: "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", Tag: "evil"}; assuming child manifest`},
				error:  nil,
				alert:  []string{`TRANSACTION REJECTED: {Action: "INSERT", FQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", PQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:evil", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent", Digest: "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", Tag: "evil"}: could not validate`},
			},
		},
		{
			"image has been completely deleted (digest removed)",
			manifests1,
			reg.GCRPubSubPayload{
				Action: "DELETE",
				FQIN:   "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32",
				PQIN:   "",
			},
			readRepo1,
			readManifestList1,
			expectedPatterns{
				report: []string{`TRANSACTION REJECTED: validation failure: {Action: "DELETE", FQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", PQIN: "", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent", Digest: "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", Tag: ""}: deletions are prohibited`},
				info:   nil,
				error:  nil,
				alert:  []string{`TRANSACTION REJECTED: validation failure: {Action: "DELETE", FQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", PQIN: "", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent", Digest: "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", Tag: ""}: deletions are prohibited`},
			},
		},
		{
			"image has been untagged (tag removed)",
			manifests1,
			reg.GCRPubSubPayload{
				Action: "DELETE",
				FQIN:   "",
				PQIN:   "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:v0.0.8",
			},
			readRepo1,
			readManifestList1,
			expectedPatterns{
				report: []string{`TRANSACTION REJECTED: validation failure: {Action: "DELETE", FQIN: "", PQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:v0.0.8", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent", Digest: "", Tag: "v0.0.8"}: deletions are prohibited`},
				info:   nil,
				error:  nil,
				alert:  []string{`TRANSACTION REJECTED: validation failure: {Action: "DELETE", FQIN: "", PQIN: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:v0.0.8", Path: "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent", Digest: "", Tag: "v0.0.8"}: deletions are prohibited`},
			},
		},
		{
			"image has been completely deleted (digest removed for UNTRACKED image)",
			manifests1,
			reg.GCRPubSubPayload{
				Action: "DELETE",
				FQIN:   "us.gcr.io/k8s-artifacts-prod/secret@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32",
				PQIN:   "",
			},
			readRepo1,
			readManifestList1,
			expectedPatterns{
				report: []string{`TRANSACTION REJECTED: validation failure: {Action: "DELETE", FQIN: "us.gcr.io/k8s-artifacts-prod/secret@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", PQIN: "", Path: "us.gcr.io/k8s-artifacts-prod/secret", Digest: "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", Tag: ""}: deletions are prohibited`},
				info:   nil,
				error:  nil,
				alert:  []string{`TRANSACTION REJECTED: validation failure: {Action: "DELETE", FQIN: "us.gcr.io/k8s-artifacts-prod/secret@sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", PQIN: "", Path: "us.gcr.io/k8s-artifacts-prod/secret", Digest: "sha256:c419394f3fa40c32352be5a6ec5865270376d4351a3756bb1893be3f28fcba32", Tag: ""}: deletions are prohibited`},
			},
		},
		{
			"image has been untagged (tag removed for UNTRACKED image)",
			manifests1,
			reg.GCRPubSubPayload{
				Action: "DELETE",
				FQIN:   "",
				PQIN:   "us.gcr.io/k8s-artifacts-prod/secret:v0.0.8",
			},
			readRepo1,
			readManifestList1,
			expectedPatterns{
				report: []string{`TRANSACTION REJECTED: validation failure: {Action: "DELETE", FQIN: "", PQIN: "us.gcr.io/k8s-artifacts-prod/secret:v0.0.8", Path: "us.gcr.io/k8s-artifacts-prod/secret", Digest: "", Tag: "v0.0.8"}: deletions are prohibited`},
				info:   nil,
				error:  nil,
				alert:  []string{`TRANSACTION REJECTED: validation failure: {Action: "DELETE", FQIN: "", PQIN: "us.gcr.io/k8s-artifacts-prod/secret:v0.0.8", Path: "us.gcr.io/k8s-artifacts-prod/secret", Digest: "", Tag: "v0.0.8"}: deletions are prohibited`},
			},
		},
		{
			"child image promoted (two manifests both promote to the same destination, but only one source repo actually has the child's parent image referenced in the promoter manifest; payload refers to non-root insertion)",
			manifests2,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/k8s-artifacts-prod/kubernetes/etcd@sha256:0873d877318546c6569e1abfafd75e0625c202d24435299c4d2e57eeebea52ee",
				PQIN:   "",
			},
			readRepo2,
			readManifestList2,
			expectedPatterns{
				report: nil,
				info:   []string{`TRANSACTION VERIFIED: {Action: "INSERT", FQIN: "us.gcr.io/k8s-artifacts-prod/kubernetes/etcd@sha256:0873d877318546c6569e1abfafd75e0625c202d24435299c4d2e57eeebea52ee", PQIN: "", Path: "us.gcr.io/k8s-artifacts-prod/kubernetes/etcd", Digest: "sha256:0873d877318546c6569e1abfafd75e0625c202d24435299c4d2e57eeebea52ee", Tag: ""}: agrees with manifest \(parent digest sha256:bcdd5657b1edc1a2eb27356f33dd66b9400d4a084209c33461c7a7da0a32ebb3\)`},
				error:  nil,
				alert:  nil,
			},
		},
		{
			"child image promoted (two manifests both promote to the same destination, but only one source repo actually has the child's parent image referenced in the promoter manifest; payload refers to root insertion)",
			manifests2,
			reg.GCRPubSubPayload{
				Action: "INSERT",
				FQIN:   "us.gcr.io/k8s-artifacts-prod/etcd@sha256:0873d877318546c6569e1abfafd75e0625c202d24435299c4d2e57eeebea52ee",
				PQIN:   "",
			},
			readRepo2,
			readManifestList2,
			expectedPatterns{
				report: nil,
				info:   []string{`TRANSACTION VERIFIED: {Action: "INSERT", FQIN: "us.gcr.io/k8s-artifacts-prod/etcd@sha256:0873d877318546c6569e1abfafd75e0625c202d24435299c4d2e57eeebea52ee", PQIN: "", Path: "us.gcr.io/k8s-artifacts-prod/etcd", Digest: "sha256:0873d877318546c6569e1abfafd75e0625c202d24435299c4d2e57eeebea52ee", Tag: ""}: agrees with manifest \(parent digest sha256:bcdd5657b1edc1a2eb27356f33dd66b9400d4a084209c33461c7a7da0a32ebb3\)`},
				error:  nil,
				alert:  nil,
			},
		},
	}

	for _, test := range shouldBeValid {
		// Create a new ResponseRecorder to record the response from Audit().
		w := httptest.NewRecorder()

		// Create a new Request to pass to the handler, which incorporates the
		// GCRPubSubPayload.
		payload, err := json.Marshal(test.payload)
		require.Nil(t, err)

		psm := audit.PubSubMessage{
			Message: audit.PubSubMessageInner{
				Data: payload,
				ID:   "1"},
			Subscription: "2"}
		b, err := json.Marshal(psm)
		require.Nil(t, err)

		r, err := http.NewRequest("POST", "/", bytes.NewBuffer(b))
		require.Nil(t, err)

		// test is used to pin the "test" variable from the outer "range" scope
		// (see scopelint) into the fakeReadRepo (in a sense it ensures that
		// fakeReadRepo closes over "test" in the outer scope, as a closure
		// should).
		test := test
		fakeReadRepo := func(sc *reg.SyncContext, rc reg.RegistryContext) stream.Producer {
			var sr stream.Fake

			_, domain, repoPath := reg.GetTokenKeyDomainRepoPath(rc.Name)
			key := fmt.Sprintf("%s/%s", domain, repoPath)

			fakeHTTPBody, ok := test.readRepo[key]
			require.NotEmpty(t, ok)

			sr.Bytes = []byte(fakeHTTPBody)
			return &sr
		}

		fakeReadManifestList := func(sc *reg.SyncContext, gmlc *reg.GCRManifestListContext) stream.Producer {
			var sr stream.Fake

			_, domain, repoPath := reg.GetTokenKeyDomainRepoPath(gmlc.RegistryContext.Name)
			key := fmt.Sprintf("%s/%s/%s@%s",
				domain,
				repoPath,
				gmlc.ImageName,
				gmlc.Digest)
			fakeHTTPBody, ok := test.readManifestList[key]
			require.NotEmpty(t, ok)

			sr.Bytes = []byte(fakeHTTPBody)
			return &sr
		}

		reportingFacility := report.NewFakeReportingClient()
		loggingFacility := logclient.NewFakeLogClient()

		s := initFakeServerContext(
			test.manifests,
			reportingFacility,
			loggingFacility,
			fakeReadRepo,
			fakeReadManifestList,
		)

		// Handle the request.
		s.Audit(w, r)

		// Check what happened!
		require.Equal(t, w.Code, http.StatusOK)

		// Check all buffers for how the output should look like.
		reportBuffer := reportingFacility.GetReportBuffer()
		infoLogBuffer := loggingFacility.GetInfoBuffer()
		errorLogBuffer := loggingFacility.GetErrorBuffer()
		alertLogBuffer := loggingFacility.GetAlertBuffer()

		if len(test.expectedPatterns.report) > 0 {
			for _, pattern := range test.expectedPatterns.report {
				re := regexp.MustCompile(pattern)
				require.Regexp(t, reportBuffer.Bytes(), re)
			}
		} else {
			require.Empty(t, reportBuffer.String())
		}

		if len(test.expectedPatterns.info) > 0 {
			for _, pattern := range test.expectedPatterns.info {
				re := regexp.MustCompile(pattern)
				require.Regexp(t, infoLogBuffer.Bytes(), re)
			}
		} else {
			require.Empty(t, infoLogBuffer.String())
		}

		if len(test.expectedPatterns.error) > 0 {
			for _, pattern := range test.expectedPatterns.error {
				re := regexp.MustCompile(pattern)
				require.Regexp(t, errorLogBuffer.Bytes(), re)
			}
		} else {
			require.Empty(t, errorLogBuffer.String())
		}

		if len(test.expectedPatterns.alert) > 0 {
			for _, pattern := range test.expectedPatterns.alert {
				re := regexp.MustCompile(pattern)
				require.Regexp(t, alertLogBuffer.Bytes(), re)
			}
		} else {
			require.Empty(t, alertLogBuffer.String())
		}
	}
}

func initFakeServerContext(
	manifests []reg.Manifest,
	reportingFacility report.ReportingFacility,
	loggingFacility logclient.LoggingFacility,
	fakeReadRepo func(*reg.SyncContext, reg.RegistryContext) stream.Producer,
	fakeReadManifestList func(*reg.SyncContext, *reg.GCRManifestListContext) stream.Producer,
) audit.ServerContext {
	remoteManifestFacility := remotemanifest.NewFake(manifests)

	serverContext := audit.ServerContext{
		ID:                     "cafec0ffee",
		RemoteManifestFacility: remoteManifestFacility,
		ErrorReportingFacility: reportingFacility,
		LoggingFacility:        loggingFacility,
		GcrReadingFacility: audit.GcrReadingFacility{
			ReadRepo:         fakeReadRepo,
			ReadManifestList: fakeReadManifestList,
		},
	}

	return serverContext
}

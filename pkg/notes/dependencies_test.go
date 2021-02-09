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

package notes_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/v1/pkg/notes"
	"k8s.io/release/v1/pkg/notes/notesfakes"
)

const expected = `## Dependencies

### Added
- github.com/ajstarks/svgo: [644b8db](https://github.com/ajstarks/svgo/tree/644b8db)
- github.com/census-instrumentation/opencensus-proto: [v0.2.1](https://github.com/census-instrumentation/opencensus-proto/tree/v0.2.1)
- github.com/cilium/ebpf: [95b36a5](https://github.com/cilium/ebpf/tree/95b36a5)
- github.com/envoyproxy/go-control-plane: [5f8ba28](https://github.com/envoyproxy/go-control-plane/tree/5f8ba28)
- github.com/envoyproxy/protoc-gen-validate: [v0.1.0](https://github.com/envoyproxy/protoc-gen-validate/tree/v0.1.0)
- github.com/fogleman/gg: [0403632](https://github.com/fogleman/gg/tree/0403632)
- github.com/golang/freetype: [e2365df](https://github.com/golang/freetype/tree/e2365df)
- github.com/jung-kurt/gofpdf: [24315ac](https://github.com/jung-kurt/gofpdf/tree/24315ac)
- gonum.org/v1/plot: e2840ee
- rsc.io/pdf: v0.1.1
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.0.7
- sigs.k8s.io/structured-merge-diff/v3: v3.0.0

### Changed
- github.com/Microsoft/go-winio: [v0.4.11 → v0.4.14](https://github.com/Microsoft/go-winio/compare/v0.4.11...v0.4.14)
- github.com/aws/aws-sdk-go: [v1.16.26 → v1.28.2](https://github.com/aws/aws-sdk-go/compare/v1.16.26...v1.28.2)
- github.com/checkpoint-restore/go-criu: [bdb7599 → 17b0214](https://github.com/checkpoint-restore/go-criu/compare/bdb7599...17b0214)
- github.com/coredns/corefile-migration: [v1.0.4 → v1.0.6](https://github.com/coredns/corefile-migration/compare/v1.0.4...v1.0.6)
- github.com/docker/libnetwork: [f0e46a7 → c8a5fca](https://github.com/docker/libnetwork/compare/f0e46a7...c8a5fca)
- github.com/elazarl/goproxy: [c4fc265 → 947c36d](https://github.com/elazarl/goproxy/compare/c4fc265...947c36d)
- github.com/gogo/protobuf: [65acae2 → v1.3.1](https://github.com/gogo/protobuf/compare/65acae2...v1.3.1)
- github.com/golang/mock: [v1.2.0 → v1.3.1](https://github.com/golang/mock/compare/v1.2.0...v1.3.1)
- github.com/google/gofuzz: [v1.0.0 → v1.1.0](https://github.com/google/gofuzz/compare/v1.0.0...v1.1.0)
- github.com/googleapis/gnostic: [0c51083 → v0.1.0](https://github.com/googleapis/gnostic/compare/0c51083...v0.1.0)
- github.com/morikuni/aec: [3977121 → v1.0.0](https://github.com/morikuni/aec/compare/3977121...v1.0.0)
- github.com/onsi/ginkgo: [v1.10.1 → v1.11.0](https://github.com/onsi/ginkgo/compare/v1.10.1...v1.11.0)
- github.com/opencontainers/runc: [v1.0.0-rc9 → v1.0.0-rc10](https://github.com/opencontainers/runc/compare/v1.0.0-rc9...v1.0.0-rc10)
- github.com/prometheus/client_model: [fd36f42 → v0.2.0](https://github.com/prometheus/client_model/compare/fd36f42...v0.2.0)
- github.com/smartystreets/goconvey: [68dc04a → v1.6.4](https://github.com/smartystreets/goconvey/compare/68dc04a...v1.6.4)
- golang.org/x/crypto: 60c769a → bac4c82
- gonum.org/v1/gonum: 3d26580 → v0.6.2
- google.golang.org/genproto: 54afdca → 24fa4b2
- google.golang.org/grpc: v1.23.1 → v1.26.0
- gopkg.in/yaml.v2: v2.2.4 → v2.2.8
- k8s.io/gengo: 26a6646 → 36b2048
- k8s.io/kube-openapi: 30be4d1 → bf4fb3b
- k8s.io/utils: e782cd3 → a9aa75a
- sigs.k8s.io/yaml: v1.1.0 → v1.2.0

### Removed
- sigs.k8s.io/structured-merge-diff: b1b620d
`

func TestDependencyChangesSuccess(t *testing.T) {
	moDiff := &notesfakes.FakeMoDiff{}
	moDiff.RunReturns(expected, nil)
	sut := notes.NewDependencies()
	sut.SetMoDiff(moDiff)

	res, err := sut.Changes("v1.17.0", "v1.18.0")
	require.Nil(t, err)
	require.Equal(t, expected, res)
}

func TestDependencyChangesFailure(t *testing.T) {
	moDiff := &notesfakes.FakeMoDiff{}
	moDiff.RunReturns("", errors.New(""))
	sut := notes.NewDependencies()
	sut.SetMoDiff(moDiff)

	res, err := sut.Changes("v1.17.0", "v1.18.0")
	require.NotNil(t, err)
	require.Empty(t, res)
}

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

package notes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const testInput = `
# v1.16.4

[Documentation](https://docs.k8s.io)

## Downloads for v1.16.4

| filename | sha512 hash |
| -------- | ----------- |


### Client Binaries

| filename | sha512 hash |
| -------- | ----------- |


### Server Binaries

| filename | sha512 hash |
| -------- | ----------- |


### Node Binaries

| filename | sha512 hash |
| -------- | ----------- |


## Changelog since v1.16.3

### API Changes

- For x-kubernetes-list-type=set a scalar or atomic item type is now required, as documented. Persisted, invalid data is tolerated. ([#85385](https://github.com/kubernetes/kubernetes/pull/85385), [@sttts](https://github.com/sttts))
  ` + "```" + `
  # A code block
  ` + "```" + `

### Notes from Multiple SIGs

#### SIG API Machinery, SIG Cloud Provider, and SIG Scalability

- Fixes a performance issue when using server-side apply with objects with very large atomic maps. ([#85462](https://github.com/kubernetes/kubernetes/pull/85462), [@jennybuckley](https://github.com/jennybuckley))

#### SIG Apps, and ` + "`SIG`" + ` Network

- kube-controller-manager: Fixes bug setting headless service labels on endpoints ([#85361](https://github.com/kubernetes/kubernetes/pull/85361), [@liggitt](https://github.com/liggitt))

### Notes from Individual SIGs

#### SIG API Machinery

- Filter published OpenAPI schema by making nullable, required fields non-required in order to avoid kubectl to wrongly reject null values. ([#85733](https://github.com/kubernetes/kubernetes/pull/85733), [@sttts](https://github.com/sttts))
- For x-kubernetes-list-type=set a scalar or atomic item type is now required, as documented. Persisted, invalid data is tolerated. ([#85385](https://github.com/kubernetes/kubernetes/pull/85385), [@sttts](https://github.com/sttts))

#### SIG Cloud Provider

- azure: update disk lock logic per vm during attach/detach to allow concurrent updates for different nodes. ([#85115](https://github.com/kubernetes/kubernetes/pull/85115), [@aramase](https://github.com/aramase))
- fix vmss dirty cache issue in disk attach/detach on vmss node ([#85158](https://github.com/kubernetes/kubernetes/pull/85158), [@andyzhangx](https://github.com/andyzhangx))
- fix race condition when attach/delete azure disk in same time ([#84917](https://github.com/kubernetes/kubernetes/pull/84917), [@andyzhangx](https://github.com/andyzhangx))
- Ensure health probes are created for local traffic policy UDP services on Azure ([#85189](https://github.com/kubernetes/kubernetes/pull/85189), [@nilo19](https://github.com/nilo19))
- Change GCP ILB firewall names to contain the "k8s-fw-" prefix like the rest of the firewall rules. This is needed for consistency and also for other components to identify the firewall rule as k8s/service-controller managed. ([#85102](https://github.com/kubernetes/kubernetes/pull/85102), [@prameshj](https://github.com/prameshj))

#### SIG Cluster Lifecycle

- Fixed issue with addon-resizer using deprecated extensions APIs ([#85865](https://github.com/kubernetes/kubernetes/pull/85865), [@liggitt](https://github.com/liggitt))
- kubeadm: prevent infinite hang on "kubeadm join" using token discovery ([#85292](https://github.com/kubernetes/kubernetes/pull/85292), [@neolit123](https://github.com/neolit123))
- In cases where the CoreDNS migration isn't supported and the user chooses to ignore the warnings from the preflight check, the migration will be skipped and the ConfigMap and Deployment of CoreDNS will be retained. ([#85096](https://github.com/kubernetes/kubernetes/pull/85096), [@rajansandeep](https://github.com/rajansandeep))
- kubeadm: fix skipped etcd upgrade on secondary control-plane nodes when the command "kubeadm upgrade node" is used. ([#85024](https://github.com/kubernetes/kubernetes/pull/85024), [@neolit123](https://github.com/neolit123))

#### SIG Network

- Change kube-proxy's default node IP back to 127.0.0.1, if this is incorrect, please use --bind-address to set the correct address ([#84391](https://github.com/kubernetes/kubernetes/pull/84391), [@zouyee](https://github.com/zouyee))

# v1.16.3
## Downloads for v1.16.3
### Client Binaries
### Server Binaries
### Node Binaries
## Changelog since v1.16.2

# v1.16.2
## Downloads for v1.16.2
### Client Binaries
### Server Binaries
### Node Binaries
## Changelog since v1.16.1`

const expectedOutput = `- [v1.16.4](#v1164)
  - [Downloads for v1.16.4](#downloads-for-v1164)
    - [Client Binaries](#client-binaries)
    - [Server Binaries](#server-binaries)
    - [Node Binaries](#node-binaries)
  - [Changelog since v1.16.3](#changelog-since-v1163)
    - [API Changes](#api-changes)
    - [Notes from Multiple SIGs](#notes-from-multiple-sigs)
      - [SIG API Machinery, SIG Cloud Provider, and SIG Scalability](#sig-api-machinery-sig-cloud-provider-and-sig-scalability)
      - [SIG Apps, and ` + "`SIG`" + ` Network](#sig-apps-and-sig-network)
    - [Notes from Individual SIGs](#notes-from-individual-sigs)
      - [SIG API Machinery](#sig-api-machinery)
      - [SIG Cloud Provider](#sig-cloud-provider)
      - [SIG Cluster Lifecycle](#sig-cluster-lifecycle)
      - [SIG Network](#sig-network)
- [v1.16.3](#v1163)
  - [Downloads for v1.16.3](#downloads-for-v1163)
    - [Client Binaries](#client-binaries-1)
    - [Server Binaries](#server-binaries-1)
    - [Node Binaries](#node-binaries-1)
  - [Changelog since v1.16.2](#changelog-since-v1162)
- [v1.16.2](#v1162)
  - [Downloads for v1.16.2](#downloads-for-v1162)
    - [Client Binaries](#client-binaries-2)
    - [Server Binaries](#server-binaries-2)
    - [Node Binaries](#node-binaries-2)
  - [Changelog since v1.16.1](#changelog-since-v1161)
`

func TestGenerateTOC(t *testing.T) {
	toc, err := GenerateTOC(testInput)
	require.Nil(t, err)
	require.Equal(t, toc, expectedOutput)
}

func TestGenerateTOCBackTickInHeading(t *testing.T) {
	toc, err := GenerateTOC("# `markdown` solves all our problems, they said")
	require.Nil(t, err)
	require.Equal(t, toc, "- [`markdown` solves all our problems, they said](#markdown-solves-all-our-problems-they-said)\n")
}

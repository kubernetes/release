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

package cmd

const patchReleaseExpectedTOC = `<!-- BEGIN MUNGE: GENERATED_TOC -->

- [v1.16.3](#v1163)
  - [Changelog since v1.16.2](#changelog-since-v1162)
  - [Changes by Kind](#changes-by-kind)`

const patchReleaseExpectedContent = `## Changes by Kind

### Feature

- Azure: Add allow unsafe read from cache ([#83685](https://github.com/kubernetes/kubernetes/pull/83685), [@aramase](https://github.com/aramase)) [SIG Cloud Provider]

### Failing Test

- Removed dependency on test/e2e/common from test/e2e/storage/testsuites ([#83776](https://github.com/kubernetes/kubernetes/pull/83776), [@avalluri](https://github.com/avalluri)) [SIG Testing]

### Other (Bug, Cleanup or Flake)

- Add data cache flushing during unmount device for GCE-PD driver in Windows Server. ([#83591](https://github.com/kubernetes/kubernetes/pull/83591), [@jingxu97](https://github.com/jingxu97)) [SIG Storage and Windows]
- Adds a metric apiserver_request_error_total to kube-apiserver. This metric tallies the number of request_errors encountered by verb, group, version, resource, subresource, scope, component, and code. ([#83427](https://github.com/kubernetes/kubernetes/pull/83427), [@logicalhan](https://github.com/logicalhan)) [SIG API Machinery and Instrumentation]
- Bumps metrics-server version to v0.3.6 for following bugfix:
  - Don't break metric storage when duplicate pod metrics encountered causing hpa to fail ([#84223](https://github.com/kubernetes/kubernetes/pull/84223), [@olagacek](https://github.com/olagacek)) [SIG Cluster Lifecycle]
- CSI detach timeout increased from 10 seconds to 2 minutes ([#84321](https://github.com/kubernetes/kubernetes/pull/84321), [@cduchesne](https://github.com/cduchesne)) [SIG Storage]
- Change kube-proxy's default node IP back to 127.0.0.1, if this is incorrect, please use --bind-address to set the correct address ([#84391](https://github.com/kubernetes/kubernetes/pull/84391), [@zouyee](https://github.com/zouyee)) [SIG Network]
- Fix handling tombstones in pod-disruption-budged controller. ([#83951](https://github.com/kubernetes/kubernetes/pull/83951), [@zouyee](https://github.com/zouyee)) [SIG Apps]
- Fix kubelet metrics gathering on non-English Windows hosts ([#84156](https://github.com/kubernetes/kubernetes/pull/84156), [@wawa0210](https://github.com/wawa0210)) [SIG Node and Windows]
- Fixed an issue with informers missing an ` + "`Added`" + ` event if a recently deleted object was immediately recreated at the same time the informer dropped a watch and relisted. ([#83911](https://github.com/kubernetes/kubernetes/pull/83911), [@matte21](https://github.com/matte21)) [SIG API Machinery]
- Fixed binding of block PersistentVolumes / PersistentVolumeClaims when BlockVolume feature is off. ([#84175](https://github.com/kubernetes/kubernetes/pull/84175), [@jsafrane](https://github.com/jsafrane)) [SIG Apps]
- Fixed panic when accessing CustomResources of a CRD with x-kubernetes-int-or-string. ([#83789](https://github.com/kubernetes/kubernetes/pull/83789), [@sttts](https://github.com/sttts)) [SIG API Machinery]
- Kube-apiserver: Fixed a regression accepting patch requests > 1MB ([#84963](https://github.com/kubernetes/kubernetes/pull/84963), [@liggitt](https://github.com/liggitt)) [SIG API Machinery and Testing]
- Kube-apiserver: fixed a bug that could cause a goroutine leak if the apiserver encountered an encoding error serving a watch to a websocket watcher ([#84960](https://github.com/kubernetes/kubernetes/pull/84960), [@liggitt](https://github.com/liggitt)) [SIG API Machinery]
- Kube-scheduler now fallbacks to emitting events using core/v1 Events when events.k8s.io/v1beta1 is disabled. ([#83692](https://github.com/kubernetes/kubernetes/pull/83692), [@yastij](https://github.com/yastij)) [SIG API Machinery, Apps, Scheduling and Testing]
- Kubeadm: fix skipped etcd upgrade on secondary control-plane nodes when the command "kubeadm upgrade node" is used. ([#85024](https://github.com/kubernetes/kubernetes/pull/85024), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Restores compatibility of kube-scheduler with clusters that do not enable the events.k8s.io/v1beta1 API ([#84465](https://github.com/kubernetes/kubernetes/pull/84465), [@yastij](https://github.com/yastij)) [SIG API Machinery and Scheduling]
- Switched intstr.Type to sized integer to follow API guidelines and improve compatibility with proto libraries ([#83956](https://github.com/kubernetes/kubernetes/pull/83956), [@liggitt](https://github.com/liggitt)) [SIG API Machinery]
- Update Cluster Autoscaler version to 1.16.2 (CA release docs: https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.16.2) ([#84038](https://github.com/kubernetes/kubernetes/pull/84038), [@losipiuk](https://github.com/losipiuk)) [SIG Cluster Lifecycle]
- Update to use go1.12.12 ([#84064](https://github.com/kubernetes/kubernetes/pull/84064), [@cblecker](https://github.com/cblecker)) [SIG Release and Testing]
- Upgrade to etcd client 3.3.17 to fix bug where etcd client does not parse IPv6 addresses correctly when members are joining, and to fix bug where failover on multi-member etcd cluster fails certificate check on DNS mismatch ([#83968](https://github.com/kubernetes/kubernetes/pull/83968), [@jpbetz](https://github.com/jpbetz)) [SIG API Machinery and Cloud Provider]`

const alphaReleaseExpectedTOC = `<!-- BEGIN MUNGE: GENERATED_TOC -->

- \[v1.18.0-alpha.3\]\(\#v1180-alpha3\)
  - \[Changelog since v.*\]\(\#changelog-since-.*\)
  - \[Changes by Kind\]\(\#changes-by-kind\)
    - \[Deprecation\]\(\#deprecation\)
    - \[API Change\]\(\#api-change\)
    - \[Feature\]\(\#feature\)
    - \[Other \(Bug, Cleanup or Flake\)\]\(\#other-bug-cleanup-or-flake\)`

const alphaReleaseExpectedContent = `## Changes by Kind

### Deprecation

- Kubeadm: kube-dns is deprecated and will not be supported in a future version ([#86574](https://github.com/kubernetes/kubernetes/pull/86574), [@SataQiu](https://github.com/SataQiu)) [SIG Cluster Lifecycle]
- Remove all the generators from kubectl run. It will now only create pods. Additionally, deprecates all the flags that are not relevant anymore. ([#87077](https://github.com/kubernetes/kubernetes/pull/87077), [@soltysh](https://github.com/soltysh)) [SIG Architecture, CLI and Testing]

### API Change

- --enable-cadvisor-endpoints is now disabled by default. If you need access to the cAdvisor v1 Json API please enable it explicitly in the kubelet command line. Please note that this flag was deprecated in 1.15 and will be removed in 1.19. ([#87440](https://github.com/kubernetes/kubernetes/pull/87440), [@dims](https://github.com/dims)) [SIG Instrumentation, Node and Testing]
- Add kubescheduler.config.k8s.io/v1alpha2 ([#87628](https://github.com/kubernetes/kubernetes/pull/87628), [@alculquicondor](https://github.com/alculquicondor)) [SIG Scheduling]
- The following feature gates are removed, because the associated features were unconditionally enabled in previous releases: CustomResourceValidation, CustomResourceSubresources, CustomResourceWebhookConversion, CustomResourcePublishOpenAPI, CustomResourceDefaulting ([#87475](https://github.com/kubernetes/kubernetes/pull/87475), [@liggitt](https://github.com/liggitt)) [SIG API Machinery]

### Feature

- API request throttling (due to a high rate of requests) is now reported in client-go logs at log level 2.  The messages are of the form
  ` + `
  Throttling request took 1.50705208s, request: GET:<URL>
  ` + `
  The presence of these messages, may indicate to the administrator the need to tune the cluster accordingly. ([#87740](https://github.com/kubernetes/kubernetes/pull/87740), [@jennybuckley](https://github.com/jennybuckley)) [SIG API Machinery]
- Add support for pre-allocated hugepages for more than one page size ([#82820](https://github.com/kubernetes/kubernetes/pull/82820), [@odinuge](https://github.com/odinuge)) [SIG Apps]
- Added more details to taint toleration errors ([#87250](https://github.com/kubernetes/kubernetes/pull/87250), [@starizard](https://github.com/starizard)) [SIG Apps and Scheduling]
- Aggregation api will have alpha support for network proxy ([#87515](https://github.com/kubernetes/kubernetes/pull/87515), [@Sh4d1](https://github.com/Sh4d1)) [SIG API Machinery]
- DisableAvailabilitySetNodes is added to avoid VM list for VMSS clusters. It should only be used when vmType is "vmss" and all the nodes (including masters) are VMSS virtual machines. ([#87685](https://github.com/kubernetes/kubernetes/pull/87685), [@feiskyer](https://github.com/feiskyer)) [SIG Cloud Provider]
- Kube-apiserver metrics will now include request counts, latencies, and response sizes for /healthz, /livez, and /readyz requests. ([#83598](https://github.com/kubernetes/kubernetes/pull/83598), [@jktomer](https://github.com/jktomer)) [SIG API Machinery]
- Kubeadm: reject a node joining the cluster if a node with the same name already exists ([#81056](https://github.com/kubernetes/kubernetes/pull/81056), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Scheduler: Add DefaultBinder plugin ([#87430](https://github.com/kubernetes/kubernetes/pull/87430), [@alculquicondor](https://github.com/alculquicondor)) [SIG Scheduling and Testing]
- Skip default spreading scoring plugin for pods that define TopologySpreadConstraints ([#87566](https://github.com/kubernetes/kubernetes/pull/87566), [@skilxn-go](https://github.com/skilxn-go)) [SIG Scheduling]
- The kubectl --dry-run flag now accepts the values 'client', 'server', and 'none', to support client-side and server-side dry-run strategies. The boolean and unset values for the --dry-run flag are deprecated and a value will be required in a future version. ([#87580](https://github.com/kubernetes/kubernetes/pull/87580), [@julianvmodesto](https://github.com/julianvmodesto)) [SIG CLI]
- Update CNI version to v0.8.5 ([#78819](https://github.com/kubernetes/kubernetes/pull/78819), [@justaugustus](https://github.com/justaugustus)) [SIG API Machinery, Cluster Lifecycle, Network, Release and Testing]

### Other (Bug, Cleanup or Flake)

- "kubectl describe statefulsets.apps" prints garbage for rolling update partition ([#85846](https://github.com/kubernetes/kubernetes/pull/85846), [@phil9909](https://github.com/phil9909)) [SIG CLI]
- Update cri-tools to v1.17.0 ([#86305](https://github.com/kubernetes/kubernetes/pull/86305), [@saschagrunert](https://github.com/saschagrunert)) [SIG Cluster Lifecycle and Release]
- Fix regression in statefulset conversion which prevented applying a statefulset multiple times. ([#87706](https://github.com/kubernetes/kubernetes/pull/87706), [@liggitt](https://github.com/liggitt)) [SIG Apps and Testing]
- Fix the masters rolling upgrade causing thundering herd of LISTs on etcd leading to control plane unavailability. ([#86430](https://github.com/kubernetes/kubernetes/pull/86430), [@wojtek-t](https://github.com/wojtek-t)) [SIG API Machinery, Node and Testing]
- Fixed two scheduler metrics (pending_pods and schedule_attempts_total) not being recorded ([#87692](https://github.com/kubernetes/kubernetes/pull/87692), [@everpeace](https://github.com/everpeace)) [SIG Scheduling]
- For volumes that allow attaches across multiple nodes, attach and detach operations across different nodes are now executed in parallel. ([#87258](https://github.com/kubernetes/kubernetes/pull/87258), [@verult](https://github.com/verult)) [SIG Apps, Node and Storage]
- Kubeadm: apply further improvements to the tentative support for concurrent etcd member join. Fixes a bug where multiple members can receive the same hostname. Increase the etcd client dial timeout and retry timeout for add/remove/... operations. ([#87505](https://github.com/kubernetes/kubernetes/pull/87505), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Kubeadm: remove the deprecated CoreDNS feature-gate. It was set to "true" since v1.11 when the feature went GA. In v1.13 it was marked as deprecated and hidden from the CLI. ([#87400](https://github.com/kubernetes/kubernetes/pull/87400), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Removed the 'client' label from apiserver_request_total. ([#87669](https://github.com/kubernetes/kubernetes/pull/87669), [@logicalhan](https://github.com/logicalhan)) [SIG API Machinery and Instrumentation]
- Resolved a performance issue in the node authorizer index maintenance. ([#87693](https://github.com/kubernetes/kubernetes/pull/87693), [@liggitt](https://github.com/liggitt)) [SIG Auth]
- Reverted a kubectl azure auth module change where oidc claim spn: prefix was omitted resulting a breaking behavior with existing Azure AD OIDC enabled api-server ([#87507](https://github.com/kubernetes/kubernetes/pull/87507), [@weinong](https://github.com/weinong)) [SIG API Machinery, Auth and Cloud Provider]
- Shared informers are now more reliable in the face of network disruption. ([#86015](https://github.com/kubernetes/kubernetes/pull/86015), [@squeed](https://github.com/squeed)) [SIG API Machinery]
- The CSR signing cert/key pairs will be reloaded from disk like the kube-apiserver cert/key pairs ([#86816](https://github.com/kubernetes/kubernetes/pull/86816), [@deads2k](https://github.com/deads2k)) [SIG API Machinery, Apps and Auth]
- To reduce chances of throttling, VM cache is set to nil when Azure node provisioning state is deleting ([#87635](https://github.com/kubernetes/kubernetes/pull/87635), [@feiskyer](https://github.com/feiskyer)) [SIG Cloud Provider]`

const minorReleaseExpectedTOC = `<!-- BEGIN MUNGE: GENERATED_TOC -->

- [v1.17.0](#v1170)
  - [Changelog since v1.16.0](#changelog-since-v1160)
- [Kubernetes v1.17.0 Release Notes](#kubernetes-v1170-release-notes)
  - [What’s New (Major Themes)](#what’s-new-major-themes)
    - [Cloud Provider Labels reach General Availability](#cloud-provider-labels-reach-general-availability)
    - [Volume Snapshot Moves to Beta](#volume-snapshot-moves-to-beta)
    - [CSI Migration Beta](#csi-migration-beta)
  - [Known Issues](#known-issues)
  - [Urgent Upgrade Notes](#urgent-upgrade-notes)
    - [(No, really, you MUST read this before you upgrade)](#no-really-you-must-read-this-before-you-upgrade)
      - [Cluster Lifecycle](#cluster-lifecycle)
      - [Network](#network)
      - [Scheduling](#scheduling)
      - [Storage](#storage)
      - [Windows](#windows)
    - [Deprecations and Removals](#deprecations-and-removals)
    - [Metrics Changes](#metrics-changes)
      - [Added metrics](#added-metrics)
      - [Deprecated/changed metrics](#deprecated/changed-metrics)
    - [Notable Features](#notable-features)
      - [Stable](#stable)
      - [Beta](#beta)
      - [CLI Improvements](#cli-improvements)
    - [API Changes](#api-changes)
    - [Other notable changes](#other-notable-changes)
      - [API Machinery](#api-machinery)
      - [Apps](#apps)
      - [Auth](#auth)
      - [CLI](#cli)
      - [Cloud Provider](#cloud-provider)
      - [Cluster Lifecycle](#cluster-lifecycle-1)
      - [Instrumentation](#instrumentation)
      - [Network](#network-1)
      - [Node](#node)
      - [Release](#release)
      - [Scheduling](#scheduling-1)
      - [Storage](#storage-1)
      - [Windows](#windows-1)
      - [Dependencies](#dependencies)
      - [Detailed go Dependency Changes](#detailed-go-dependency-changes)
        - [Added](#added)
        - [Changed](#changed)
        - [Removed](#removed)`

const minorReleaseExpectedContent = `# Kubernetes v1.17.0 Release Notes
A complete changelog for the release notes is now hosted in a customizable format at [relnotes.k8s.io](https://relnotes.k8s.io). Check it out and please give us your feedback!

## What’s New (Major Themes)

### Cloud Provider Labels reach General Availability

Added as a beta feature way back in v1.2, v1.17 sees the general availability of cloud provider labels.

### Volume Snapshot Moves to Beta

The Kubernetes Volume Snapshot feature is now beta in Kubernetes v1.17. It was introduced as alpha in Kubernetes v1.12, with a second alpha with breaking changes in Kubernetes v1.13.

### CSI Migration Beta

The Kubernetes in-tree storage plugin to Container Storage Interface (CSI) migration infrastructure is now beta in Kubernetes v1.17. CSI migration was introduced as alpha in Kubernetes v1.14.`

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

### Bug or Regression

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

const patchReleaseDeps = `## Dependencies

### Added
- github.com/OpenPeeDeeP/depguard: [v1.0.1](https://github.com/OpenPeeDeeP/depguard/tree/v1.0.1)
- github.com/StackExchange/wmi: [5d04971](https://github.com/StackExchange/wmi/tree/5d04971)
- github.com/anmitsu/go-shlex: [648efa6](https://github.com/anmitsu/go-shlex/tree/648efa6)
- github.com/bazelbuild/rules_go: [6dae44d](https://github.com/bazelbuild/rules_go/tree/6dae44d)
- github.com/bradfitz/go-smtpd: [deb6d62](https://github.com/bradfitz/go-smtpd/tree/deb6d62)
- github.com/gliderlabs/ssh: [v0.1.1](https://github.com/gliderlabs/ssh/tree/v0.1.1)
- github.com/go-critic/go-critic: [1df3008](https://github.com/go-critic/go-critic/tree/1df3008)
- github.com/go-lintpack/lintpack: [v0.5.2](https://github.com/go-lintpack/lintpack/tree/v0.5.2)
- github.com/go-ole/go-ole: [v1.2.1](https://github.com/go-ole/go-ole/tree/v1.2.1)
- github.com/go-toolsmith/astcast: [v1.0.0](https://github.com/go-toolsmith/astcast/tree/v1.0.0)
- github.com/go-toolsmith/astcopy: [v1.0.0](https://github.com/go-toolsmith/astcopy/tree/v1.0.0)
- github.com/go-toolsmith/astequal: [v1.0.0](https://github.com/go-toolsmith/astequal/tree/v1.0.0)
- github.com/go-toolsmith/astfmt: [v1.0.0](https://github.com/go-toolsmith/astfmt/tree/v1.0.0)
- github.com/go-toolsmith/astinfo: [9809ff7](https://github.com/go-toolsmith/astinfo/tree/9809ff7)
- github.com/go-toolsmith/astp: [v1.0.0](https://github.com/go-toolsmith/astp/tree/v1.0.0)
- github.com/go-toolsmith/pkgload: [v1.0.0](https://github.com/go-toolsmith/pkgload/tree/v1.0.0)
- github.com/go-toolsmith/strparse: [v1.0.0](https://github.com/go-toolsmith/strparse/tree/v1.0.0)
- github.com/go-toolsmith/typep: [v1.0.0](https://github.com/go-toolsmith/typep/tree/v1.0.0)
- github.com/gobwas/glob: [v0.2.3](https://github.com/gobwas/glob/tree/v0.2.3)
- github.com/golangci/check: [cfe4005](https://github.com/golangci/check/tree/cfe4005)
- github.com/golangci/dupl: [3e9179a](https://github.com/golangci/dupl/tree/3e9179a)
- github.com/golangci/errcheck: [ef45e06](https://github.com/golangci/errcheck/tree/ef45e06)
- github.com/golangci/go-misc: [927a3d8](https://github.com/golangci/go-misc/tree/927a3d8)
- github.com/golangci/go-tools: [e32c541](https://github.com/golangci/go-tools/tree/e32c541)
- github.com/golangci/goconst: [041c5f2](https://github.com/golangci/goconst/tree/041c5f2)
- github.com/golangci/gocyclo: [2becd97](https://github.com/golangci/gocyclo/tree/2becd97)
- github.com/golangci/gofmt: [0b8337e](https://github.com/golangci/gofmt/tree/0b8337e)
- github.com/golangci/golangci-lint: [v1.18.0](https://github.com/golangci/golangci-lint/tree/v1.18.0)
- github.com/golangci/gosec: [66fb7fc](https://github.com/golangci/gosec/tree/66fb7fc)
- github.com/golangci/ineffassign: [42439a7](https://github.com/golangci/ineffassign/tree/42439a7)
- github.com/golangci/lint-1: [ee948d0](https://github.com/golangci/lint-1/tree/ee948d0)
- github.com/golangci/maligned: [b1d8939](https://github.com/golangci/maligned/tree/b1d8939)
- github.com/golangci/misspell: [950f5d1](https://github.com/golangci/misspell/tree/950f5d1)
- github.com/golangci/prealloc: [215b22d](https://github.com/golangci/prealloc/tree/215b22d)
- github.com/golangci/revgrep: [d9c87f5](https://github.com/golangci/revgrep/tree/d9c87f5)
- github.com/golangci/unconvert: [28b1c44](https://github.com/golangci/unconvert/tree/28b1c44)
- github.com/google/go-github: [v17.0.0+incompatible](https://github.com/google/go-github/tree/v17.0.0)
- github.com/google/go-querystring: [v1.0.0](https://github.com/google/go-querystring/tree/v1.0.0)
- github.com/gostaticanalysis/analysisutil: [v0.0.3](https://github.com/gostaticanalysis/analysisutil/tree/v0.0.3)
- github.com/jellevandenhooff/dkim: [f50fe3d](https://github.com/jellevandenhooff/dkim/tree/f50fe3d)
- github.com/klauspost/compress: [v1.4.1](https://github.com/klauspost/compress/tree/v1.4.1)
- github.com/logrusorgru/aurora: [a7b3b31](https://github.com/logrusorgru/aurora/tree/a7b3b31)
- github.com/mattn/goveralls: [v0.0.2](https://github.com/mattn/goveralls/tree/v0.0.2)
- github.com/mitchellh/go-ps: [4fdf99a](https://github.com/mitchellh/go-ps/tree/4fdf99a)
- github.com/mozilla/tls-observatory: [8791a20](https://github.com/mozilla/tls-observatory/tree/8791a20)
- github.com/nbutton23/zxcvbn-go: [eafdab6](https://github.com/nbutton23/zxcvbn-go/tree/eafdab6)
- github.com/quasilyte/go-consistent: [c6f3937](https://github.com/quasilyte/go-consistent/tree/c6f3937)
- github.com/ryanuber/go-glob: [256dc44](https://github.com/ryanuber/go-glob/tree/256dc44)
- github.com/shirou/gopsutil: [c95755e](https://github.com/shirou/gopsutil/tree/c95755e)
- github.com/shirou/w32: [bb4de01](https://github.com/shirou/w32/tree/bb4de01)
- github.com/shurcooL/go-goon: [37c2f52](https://github.com/shurcooL/go-goon/tree/37c2f52)
- github.com/shurcooL/go: [9e1955d](https://github.com/shurcooL/go/tree/9e1955d)
- github.com/sourcegraph/go-diff: [v0.5.1](https://github.com/sourcegraph/go-diff/tree/v0.5.1)
- github.com/tarm/serial: [98f6abe](https://github.com/tarm/serial/tree/98f6abe)
- github.com/timakin/bodyclose: [87058b9](https://github.com/timakin/bodyclose/tree/87058b9)
- github.com/ultraware/funlen: [v0.0.2](https://github.com/ultraware/funlen/tree/v0.0.2)
- github.com/valyala/bytebufferpool: [v1.0.0](https://github.com/valyala/bytebufferpool/tree/v1.0.0)
- github.com/valyala/fasthttp: [v1.2.0](https://github.com/valyala/fasthttp/tree/v1.2.0)
- github.com/valyala/quicktemplate: [v1.1.1](https://github.com/valyala/quicktemplate/tree/v1.1.1)
- github.com/valyala/tcplisten: [ceec8f9](https://github.com/valyala/tcplisten/tree/ceec8f9)
- go.etcd.io/bbolt: v1.3.3
- go4.org: 417644f
- golang.org/x/build: 2835ba2
- golang.org/x/perf: 6e6d33e
- golang.org/x/xerrors: a985d34
- grpc.go4.org: 11d0a25
- mvdan.cc/interfacer: c200402
- mvdan.cc/lint: adc824a
- mvdan.cc/unparam: fbb5962
- sourcegraph.com/sqs/pbtypes: d3ebe8f

### Changed
- github.com/bazelbuild/bazel-gazelle: [c728ce9 → 70208cb](https://github.com/bazelbuild/bazel-gazelle/compare/c728ce9...70208cb)
- github.com/bazelbuild/buildtools: [80c7f0d → 69366ca](https://github.com/bazelbuild/buildtools/compare/80c7f0d...69366ca)
- github.com/coreos/bbolt: [v1.3.1-coreos.6 → v1.3.3](https://github.com/coreos/bbolt/compare/v1.3.1-coreos.6...v1.3.3)
- github.com/coreos/etcd: [v3.3.15+incompatible → v3.3.17+incompatible](https://github.com/coreos/etcd/compare/v3.3.15...v3.3.17)
- github.com/coreos/go-systemd: [39ca1b0 → c6f51f8](https://github.com/coreos/go-systemd/compare/39ca1b0...c6f51f8)
- github.com/go-openapi/jsonpointer: [v0.19.2 → v0.19.3](https://github.com/go-openapi/jsonpointer/compare/v0.19.2...v0.19.3)
- github.com/go-openapi/swag: [v0.19.2 → v0.19.5](https://github.com/go-openapi/swag/compare/v0.19.2...v0.19.5)
- github.com/gregjones/httpcache: [787624d → 9cad4c3](https://github.com/gregjones/httpcache/compare/787624d...9cad4c3)
- github.com/heketi/heketi: [v9.0.0+incompatible → c2e2a4a](https://github.com/heketi/heketi/compare/v9.0.0...c2e2a4a)
- github.com/mailru/easyjson: [94de47d → b2ccc51](https://github.com/mailru/easyjson/compare/94de47d...b2ccc51)
- github.com/mattn/go-isatty: [v0.0.3 → v0.0.9](https://github.com/mattn/go-isatty/compare/v0.0.3...v0.0.9)
- github.com/pkg/errors: [v0.8.0 → v0.8.1](https://github.com/pkg/errors/compare/v0.8.0...v0.8.1)
- github.com/spf13/pflag: [v1.0.3 → v1.0.5](https://github.com/spf13/pflag/compare/v1.0.3...v1.0.5)
- golang.org/x/crypto: e84da03 → bac4c82
- golang.org/x/lint: 8f45f77 → 959b441
- golang.org/x/net: cdfb69a → 13f9640
- golang.org/x/oauth2: 9f33145 → 0f29369
- golang.org/x/sync: 42b3178 → cd5d95a
- golang.org/x/sys: 3b52091 → fde4db3
- golang.org/x/text: e6919f6 → v0.3.2
- golang.org/x/time: f51c127 → 9d24e82
- golang.org/x/tools: 6e04913 → 65e3620
- gopkg.in/inf.v0: v0.9.0 → v0.9.1
- gopkg.in/yaml.v2: v2.2.4 → v2.2.8
- k8s.io/klog: v0.4.0 → v1.0.0
- k8s.io/kube-openapi: 743ec37 → 594e756
- k8s.io/repo-infra: 00fe14e → v0.0.1-alpha.1
- sigs.k8s.io/structured-merge-diff: 6149e45 → v1.0.2

### Removed
- github.com/heketi/rest: [aa6a652](https://github.com/heketi/rest/tree/aa6a652)
- github.com/heketi/utils: [435bc5b](https://github.com/heketi/utils/tree/435bc5b)`

const patchReleaseExpectedHTML = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width" />
    <title>v1.16.3</title>
    <style type="text/css">
      table,
      th,
      tr,
      td {
        border: 1px solid gray;
        border-collapse: collapse;
        padding: 5px;
      }
    </style>
  </head>
  <body>
    <h1>v1.16.3</h1>
<h2>Changelog since v1.16.2</h2>
<h2>Changes by Kind</h2>
<h3>Feature</h3>
<ul>
<li>Azure: Add allow unsafe read from cache (<a href="https://github.com/kubernetes/kubernetes/pull/83685">#83685</a>, <a href="https://github.com/aramase">@aramase</a>) [SIG Cloud Provider]</li>
</ul>
<h3>Failing Test</h3>
<ul>
<li>Removed dependency on test/e2e/common from test/e2e/storage/testsuites (<a href="https://github.com/kubernetes/kubernetes/pull/83776">#83776</a>, <a href="https://github.com/avalluri">@avalluri</a>) [SIG Testing]</li>
</ul>
<h3>Bug or Regression</h3>
<ul>
<li>Add data cache flushing during unmount device for GCE-PD driver in Windows Server. (<a href="https://github.com/kubernetes/kubernetes/pull/83591">#83591</a>, <a href="https://github.com/jingxu97">@jingxu97</a>) [SIG Storage and Windows]</li>
<li>Adds a metric apiserver_request_error_total to kube-apiserver. This metric tallies the number of request_errors encountered by verb, group, version, resource, subresource, scope, component, and code. (<a href="https://github.com/kubernetes/kubernetes/pull/83427">#83427</a>, <a href="https://github.com/logicalhan">@logicalhan</a>) [SIG API Machinery and Instrumentation]</li>
<li>Bumps metrics-server version to v0.3.6 for following bugfix:
<ul>
<li>Don't break metric storage when duplicate pod metrics encountered causing hpa to fail (<a href="https://github.com/kubernetes/kubernetes/pull/84223">#84223</a>, <a href="https://github.com/olagacek">@olagacek</a>) [SIG Cluster Lifecycle]</li>
</ul>
</li>
<li>CSI detach timeout increased from 10 seconds to 2 minutes (<a href="https://github.com/kubernetes/kubernetes/pull/84321">#84321</a>, <a href="https://github.com/cduchesne">@cduchesne</a>) [SIG Storage]</li>
<li>Change kube-proxy's default node IP back to 127.0.0.1, if this is incorrect, please use --bind-address to set the correct address (<a href="https://github.com/kubernetes/kubernetes/pull/84391">#84391</a>, <a href="https://github.com/zouyee">@zouyee</a>) [SIG Network]</li>
<li>Fix handling tombstones in pod-disruption-budged controller. (<a href="https://github.com/kubernetes/kubernetes/pull/83951">#83951</a>, <a href="https://github.com/zouyee">@zouyee</a>) [SIG Apps]</li>
<li>Fix kubelet metrics gathering on non-English Windows hosts (<a href="https://github.com/kubernetes/kubernetes/pull/84156">#84156</a>, <a href="https://github.com/wawa0210">@wawa0210</a>) [SIG Node and Windows]</li>
<li>Fixed an issue with informers missing an <code>Added</code> event if a recently deleted object was immediately recreated at the same time the informer dropped a watch and relisted. (<a href="https://github.com/kubernetes/kubernetes/pull/83911">#83911</a>, <a href="https://github.com/matte21">@matte21</a>) [SIG API Machinery]</li>
<li>Fixed binding of block PersistentVolumes / PersistentVolumeClaims when BlockVolume feature is off. (<a href="https://github.com/kubernetes/kubernetes/pull/84175">#84175</a>, <a href="https://github.com/jsafrane">@jsafrane</a>) [SIG Apps]</li>
<li>Fixed panic when accessing CustomResources of a CRD with x-kubernetes-int-or-string. (<a href="https://github.com/kubernetes/kubernetes/pull/83789">#83789</a>, <a href="https://github.com/sttts">@sttts</a>) [SIG API Machinery]</li>
<li>Kube-apiserver: Fixed a regression accepting patch requests &gt; 1MB (<a href="https://github.com/kubernetes/kubernetes/pull/84963">#84963</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG API Machinery and Testing]</li>
<li>Kube-apiserver: fixed a bug that could cause a goroutine leak if the apiserver encountered an encoding error serving a watch to a websocket watcher (<a href="https://github.com/kubernetes/kubernetes/pull/84960">#84960</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG API Machinery]</li>
<li>Kube-scheduler now fallbacks to emitting events using core/v1 Events when events.k8s.io/v1beta1 is disabled. (<a href="https://github.com/kubernetes/kubernetes/pull/83692">#83692</a>, <a href="https://github.com/yastij">@yastij</a>) [SIG API Machinery, Apps, Scheduling and Testing]</li>
<li>Kubeadm: fix skipped etcd upgrade on secondary control-plane nodes when the command &quot;kubeadm upgrade node&quot; is used. (<a href="https://github.com/kubernetes/kubernetes/pull/85024">#85024</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Restores compatibility of kube-scheduler with clusters that do not enable the events.k8s.io/v1beta1 API (<a href="https://github.com/kubernetes/kubernetes/pull/84465">#84465</a>, <a href="https://github.com/yastij">@yastij</a>) [SIG API Machinery and Scheduling]</li>
<li>Switched intstr.Type to sized integer to follow API guidelines and improve compatibility with proto libraries (<a href="https://github.com/kubernetes/kubernetes/pull/83956">#83956</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG API Machinery]</li>
<li>Update Cluster Autoscaler version to 1.16.2 (CA release docs: <a href="https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.16.2">https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.16.2</a>) (<a href="https://github.com/kubernetes/kubernetes/pull/84038">#84038</a>, <a href="https://github.com/losipiuk">@losipiuk</a>) [SIG Cluster Lifecycle]</li>
<li>Update to use go1.12.12 (<a href="https://github.com/kubernetes/kubernetes/pull/84064">#84064</a>, <a href="https://github.com/cblecker">@cblecker</a>) [SIG Release and Testing]</li>
<li>Upgrade to etcd client 3.3.17 to fix bug where etcd client does not parse IPv6 addresses correctly when members are joining, and to fix bug where failover on multi-member etcd cluster fails certificate check on DNS mismatch (<a href="https://github.com/kubernetes/kubernetes/pull/83968">#83968</a>, <a href="https://github.com/jpbetz">@jpbetz</a>) [SIG API Machinery and Cloud Provider]</li>
</ul>
<h2>Dependencies</h2>
<h3>Added</h3>
<ul>
<li>github.com/OpenPeeDeeP/depguard: <a href="https://github.com/OpenPeeDeeP/depguard/tree/v1.0.1">v1.0.1</a></li>
<li>github.com/StackExchange/wmi: <a href="https://github.com/StackExchange/wmi/tree/5d04971">5d04971</a></li>
<li>github.com/anmitsu/go-shlex: <a href="https://github.com/anmitsu/go-shlex/tree/648efa6">648efa6</a></li>
<li>github.com/bazelbuild/rules_go: <a href="https://github.com/bazelbuild/rules_go/tree/6dae44d">6dae44d</a></li>
<li>github.com/bradfitz/go-smtpd: <a href="https://github.com/bradfitz/go-smtpd/tree/deb6d62">deb6d62</a></li>
<li>github.com/gliderlabs/ssh: <a href="https://github.com/gliderlabs/ssh/tree/v0.1.1">v0.1.1</a></li>
<li>github.com/go-critic/go-critic: <a href="https://github.com/go-critic/go-critic/tree/1df3008">1df3008</a></li>
<li>github.com/go-lintpack/lintpack: <a href="https://github.com/go-lintpack/lintpack/tree/v0.5.2">v0.5.2</a></li>
<li>github.com/go-ole/go-ole: <a href="https://github.com/go-ole/go-ole/tree/v1.2.1">v1.2.1</a></li>
<li>github.com/go-toolsmith/astcast: <a href="https://github.com/go-toolsmith/astcast/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/go-toolsmith/astcopy: <a href="https://github.com/go-toolsmith/astcopy/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/go-toolsmith/astequal: <a href="https://github.com/go-toolsmith/astequal/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/go-toolsmith/astfmt: <a href="https://github.com/go-toolsmith/astfmt/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/go-toolsmith/astinfo: <a href="https://github.com/go-toolsmith/astinfo/tree/9809ff7">9809ff7</a></li>
<li>github.com/go-toolsmith/astp: <a href="https://github.com/go-toolsmith/astp/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/go-toolsmith/pkgload: <a href="https://github.com/go-toolsmith/pkgload/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/go-toolsmith/strparse: <a href="https://github.com/go-toolsmith/strparse/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/go-toolsmith/typep: <a href="https://github.com/go-toolsmith/typep/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/gobwas/glob: <a href="https://github.com/gobwas/glob/tree/v0.2.3">v0.2.3</a></li>
<li>github.com/golangci/check: <a href="https://github.com/golangci/check/tree/cfe4005">cfe4005</a></li>
<li>github.com/golangci/dupl: <a href="https://github.com/golangci/dupl/tree/3e9179a">3e9179a</a></li>
<li>github.com/golangci/errcheck: <a href="https://github.com/golangci/errcheck/tree/ef45e06">ef45e06</a></li>
<li>github.com/golangci/go-misc: <a href="https://github.com/golangci/go-misc/tree/927a3d8">927a3d8</a></li>
<li>github.com/golangci/go-tools: <a href="https://github.com/golangci/go-tools/tree/e32c541">e32c541</a></li>
<li>github.com/golangci/goconst: <a href="https://github.com/golangci/goconst/tree/041c5f2">041c5f2</a></li>
<li>github.com/golangci/gocyclo: <a href="https://github.com/golangci/gocyclo/tree/2becd97">2becd97</a></li>
<li>github.com/golangci/gofmt: <a href="https://github.com/golangci/gofmt/tree/0b8337e">0b8337e</a></li>
<li>github.com/golangci/golangci-lint: <a href="https://github.com/golangci/golangci-lint/tree/v1.18.0">v1.18.0</a></li>
<li>github.com/golangci/gosec: <a href="https://github.com/golangci/gosec/tree/66fb7fc">66fb7fc</a></li>
<li>github.com/golangci/ineffassign: <a href="https://github.com/golangci/ineffassign/tree/42439a7">42439a7</a></li>
<li>github.com/golangci/lint-1: <a href="https://github.com/golangci/lint-1/tree/ee948d0">ee948d0</a></li>
<li>github.com/golangci/maligned: <a href="https://github.com/golangci/maligned/tree/b1d8939">b1d8939</a></li>
<li>github.com/golangci/misspell: <a href="https://github.com/golangci/misspell/tree/950f5d1">950f5d1</a></li>
<li>github.com/golangci/prealloc: <a href="https://github.com/golangci/prealloc/tree/215b22d">215b22d</a></li>
<li>github.com/golangci/revgrep: <a href="https://github.com/golangci/revgrep/tree/d9c87f5">d9c87f5</a></li>
<li>github.com/golangci/unconvert: <a href="https://github.com/golangci/unconvert/tree/28b1c44">28b1c44</a></li>
<li>github.com/google/go-github: <a href="https://github.com/google/go-github/tree/v17.0.0">v17.0.0+incompatible</a></li>
<li>github.com/google/go-querystring: <a href="https://github.com/google/go-querystring/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/gostaticanalysis/analysisutil: <a href="https://github.com/gostaticanalysis/analysisutil/tree/v0.0.3">v0.0.3</a></li>
<li>github.com/jellevandenhooff/dkim: <a href="https://github.com/jellevandenhooff/dkim/tree/f50fe3d">f50fe3d</a></li>
<li>github.com/klauspost/compress: <a href="https://github.com/klauspost/compress/tree/v1.4.1">v1.4.1</a></li>
<li>github.com/logrusorgru/aurora: <a href="https://github.com/logrusorgru/aurora/tree/a7b3b31">a7b3b31</a></li>
<li>github.com/mattn/goveralls: <a href="https://github.com/mattn/goveralls/tree/v0.0.2">v0.0.2</a></li>
<li>github.com/mitchellh/go-ps: <a href="https://github.com/mitchellh/go-ps/tree/4fdf99a">4fdf99a</a></li>
<li>github.com/mozilla/tls-observatory: <a href="https://github.com/mozilla/tls-observatory/tree/8791a20">8791a20</a></li>
<li>github.com/nbutton23/zxcvbn-go: <a href="https://github.com/nbutton23/zxcvbn-go/tree/eafdab6">eafdab6</a></li>
<li>github.com/quasilyte/go-consistent: <a href="https://github.com/quasilyte/go-consistent/tree/c6f3937">c6f3937</a></li>
<li>github.com/ryanuber/go-glob: <a href="https://github.com/ryanuber/go-glob/tree/256dc44">256dc44</a></li>
<li>github.com/shirou/gopsutil: <a href="https://github.com/shirou/gopsutil/tree/c95755e">c95755e</a></li>
<li>github.com/shirou/w32: <a href="https://github.com/shirou/w32/tree/bb4de01">bb4de01</a></li>
<li>github.com/shurcooL/go-goon: <a href="https://github.com/shurcooL/go-goon/tree/37c2f52">37c2f52</a></li>
<li>github.com/shurcooL/go: <a href="https://github.com/shurcooL/go/tree/9e1955d">9e1955d</a></li>
<li>github.com/sourcegraph/go-diff: <a href="https://github.com/sourcegraph/go-diff/tree/v0.5.1">v0.5.1</a></li>
<li>github.com/tarm/serial: <a href="https://github.com/tarm/serial/tree/98f6abe">98f6abe</a></li>
<li>github.com/timakin/bodyclose: <a href="https://github.com/timakin/bodyclose/tree/87058b9">87058b9</a></li>
<li>github.com/ultraware/funlen: <a href="https://github.com/ultraware/funlen/tree/v0.0.2">v0.0.2</a></li>
<li>github.com/valyala/bytebufferpool: <a href="https://github.com/valyala/bytebufferpool/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/valyala/fasthttp: <a href="https://github.com/valyala/fasthttp/tree/v1.2.0">v1.2.0</a></li>
<li>github.com/valyala/quicktemplate: <a href="https://github.com/valyala/quicktemplate/tree/v1.1.1">v1.1.1</a></li>
<li>github.com/valyala/tcplisten: <a href="https://github.com/valyala/tcplisten/tree/ceec8f9">ceec8f9</a></li>
<li>go.etcd.io/bbolt: v1.3.3</li>
<li>go4.org: 417644f</li>
<li>golang.org/x/build: 2835ba2</li>
<li>golang.org/x/perf: 6e6d33e</li>
<li>golang.org/x/xerrors: a985d34</li>
<li>grpc.go4.org: 11d0a25</li>
<li>mvdan.cc/interfacer: c200402</li>
<li>mvdan.cc/lint: adc824a</li>
<li>mvdan.cc/unparam: fbb5962</li>
<li>sourcegraph.com/sqs/pbtypes: d3ebe8f</li>
</ul>
<h3>Changed</h3>
<ul>
<li>github.com/bazelbuild/bazel-gazelle: <a href="https://github.com/bazelbuild/bazel-gazelle/compare/c728ce9...70208cb">c728ce9 → 70208cb</a></li>
<li>github.com/bazelbuild/buildtools: <a href="https://github.com/bazelbuild/buildtools/compare/80c7f0d...69366ca">80c7f0d → 69366ca</a></li>
<li>github.com/coreos/bbolt: <a href="https://github.com/coreos/bbolt/compare/v1.3.1-coreos.6...v1.3.3">v1.3.1-coreos.6 → v1.3.3</a></li>
<li>github.com/coreos/etcd: <a href="https://github.com/coreos/etcd/compare/v3.3.15...v3.3.17">v3.3.15+incompatible → v3.3.17+incompatible</a></li>
<li>github.com/coreos/go-systemd: <a href="https://github.com/coreos/go-systemd/compare/39ca1b0...c6f51f8">39ca1b0 → c6f51f8</a></li>
<li>github.com/go-openapi/jsonpointer: <a href="https://github.com/go-openapi/jsonpointer/compare/v0.19.2...v0.19.3">v0.19.2 → v0.19.3</a></li>
<li>github.com/go-openapi/swag: <a href="https://github.com/go-openapi/swag/compare/v0.19.2...v0.19.5">v0.19.2 → v0.19.5</a></li>
<li>github.com/gregjones/httpcache: <a href="https://github.com/gregjones/httpcache/compare/787624d...9cad4c3">787624d → 9cad4c3</a></li>
<li>github.com/heketi/heketi: <a href="https://github.com/heketi/heketi/compare/v9.0.0...c2e2a4a">v9.0.0+incompatible → c2e2a4a</a></li>
<li>github.com/mailru/easyjson: <a href="https://github.com/mailru/easyjson/compare/94de47d...b2ccc51">94de47d → b2ccc51</a></li>
<li>github.com/mattn/go-isatty: <a href="https://github.com/mattn/go-isatty/compare/v0.0.3...v0.0.9">v0.0.3 → v0.0.9</a></li>
<li>github.com/pkg/errors: <a href="https://github.com/pkg/errors/compare/v0.8.0...v0.8.1">v0.8.0 → v0.8.1</a></li>
<li>github.com/spf13/pflag: <a href="https://github.com/spf13/pflag/compare/v1.0.3...v1.0.5">v1.0.3 → v1.0.5</a></li>
<li>golang.org/x/crypto: e84da03 → bac4c82</li>
<li>golang.org/x/lint: 8f45f77 → 959b441</li>
<li>golang.org/x/net: cdfb69a → 13f9640</li>
<li>golang.org/x/oauth2: 9f33145 → 0f29369</li>
<li>golang.org/x/sync: 42b3178 → cd5d95a</li>
<li>golang.org/x/sys: 3b52091 → fde4db3</li>
<li>golang.org/x/text: e6919f6 → v0.3.2</li>
<li>golang.org/x/time: f51c127 → 9d24e82</li>
<li>golang.org/x/tools: 6e04913 → 65e3620</li>
<li>gopkg.in/inf.v0: v0.9.0 → v0.9.1</li>
<li>gopkg.in/yaml.v2: v2.2.4 → v2.2.8</li>
<li>k8s.io/klog: v0.4.0 → v1.0.0</li>
<li>k8s.io/kube-openapi: 743ec37 → 594e756</li>
<li>k8s.io/repo-infra: 00fe14e → v0.0.1-alpha.1</li>
<li>sigs.k8s.io/structured-merge-diff: 6149e45 → v1.0.2</li>
</ul>
<h3>Removed</h3>
<ul>
<li>github.com/heketi/rest: <a href="https://github.com/heketi/rest/tree/aa6a652">aa6a652</a></li>
<li>github.com/heketi/utils: <a href="https://github.com/heketi/utils/tree/435bc5b">435bc5b</a></li>
</ul>

  </body>
</html>`

const alpha1ReleaseExpectedTOC = `<!-- BEGIN MUNGE: GENERATED_TOC -->

- \[v1.19.0-alpha.1\]\(\#v1190-alpha1\)
  - \[Changelog since v1.18.0\]\(\#changelog-since-v1180\)
  - \[Changes by Kind\]\(\#changes-by-kind\)
    - \[Deprecation\]\(\#deprecation\)
    - \[API Change\]\(\#api-change\)
    - \[Feature\]\(\#feature\)
    - \[Documentation\]\(\#documentation\)
    - \[Bug or Regression\]\(\#bug-or-regression\)
    - \[Other \(Cleanup or Flake\)\]\(\#other-cleanup-or-flake\)`

const alphaReleaseExpectedTOC = `<!-- BEGIN MUNGE: GENERATED_TOC -->

- \[v1.18.0-alpha.3\]\(\#v1180-alpha3\)
  - \[Changelog since v.*\]\(\#changelog-since-.*\)
  - \[Changes by Kind\]\(\#changes-by-kind\)
    - \[Deprecation\]\(\#deprecation\)
    - \[API Change\]\(\#api-change\)
    - \[Feature\]\(\#feature\)
    - \[Bug or Regression\]\(\#bug-or-regression\)
    - \[Other \(Cleanup or Flake\)\]\(\#other-cleanup-or-flake\)`

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

### Bug or Regression

- "kubectl describe statefulsets.apps" prints garbage for rolling update partition ([#85846](https://github.com/kubernetes/kubernetes/pull/85846), [@phil9909](https://github.com/phil9909)) [SIG CLI]
- Fix regression in statefulset conversion which prevented applying a statefulset multiple times. ([#87706](https://github.com/kubernetes/kubernetes/pull/87706), [@liggitt](https://github.com/liggitt)) [SIG Apps and Testing]
- Fix the masters rolling upgrade causing thundering herd of LISTs on etcd leading to control plane unavailability. ([#86430](https://github.com/kubernetes/kubernetes/pull/86430), [@wojtek-t](https://github.com/wojtek-t)) [SIG API Machinery, Node and Testing]
- Fixed two scheduler metrics (pending_pods and schedule_attempts_total) not being recorded ([#87692](https://github.com/kubernetes/kubernetes/pull/87692), [@everpeace](https://github.com/everpeace)) [SIG Scheduling]
- For volumes that allow attaches across multiple nodes, attach and detach operations across different nodes are now executed in parallel. ([#87258](https://github.com/kubernetes/kubernetes/pull/87258), [@verult](https://github.com/verult)) [SIG Apps, Node and Storage]
- Kubeadm: apply further improvements to the tentative support for concurrent etcd member join. Fixes a bug where multiple members can receive the same hostname. Increase the etcd client dial timeout and retry timeout for add/remove/... operations. ([#87505](https://github.com/kubernetes/kubernetes/pull/87505), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Resolved a performance issue in the node authorizer index maintenance. ([#87693](https://github.com/kubernetes/kubernetes/pull/87693), [@liggitt](https://github.com/liggitt)) [SIG Auth]
- Reverted a kubectl azure auth module change where oidc claim spn: prefix was omitted resulting a breaking behavior with existing Azure AD OIDC enabled api-server ([#87507](https://github.com/kubernetes/kubernetes/pull/87507), [@weinong](https://github.com/weinong)) [SIG API Machinery, Auth and Cloud Provider]
- Shared informers are now more reliable in the face of network disruption. ([#86015](https://github.com/kubernetes/kubernetes/pull/86015), [@squeed](https://github.com/squeed)) [SIG API Machinery]
- To reduce chances of throttling, VM cache is set to nil when Azure node provisioning state is deleting ([#87635](https://github.com/kubernetes/kubernetes/pull/87635), [@feiskyer](https://github.com/feiskyer)) [SIG Cloud Provider]
- ` + "`" + `kubectl diff` + "`" + ` now returns 1 only on diff finding changes, and >1 on kubectl errors. The "exit status code 1" message as also been muted. ([#87437](https://github.com/kubernetes/kubernetes/pull/87437), [@apelisse](https://github.com/apelisse)) [SIG CLI and Testing]

### Other (Cleanup or Flake)

- Kubeadm: remove the deprecated CoreDNS feature-gate. It was set to "true" since v1.11 when the feature went GA. In v1.13 it was marked as deprecated and hidden from the CLI. ([#87400](https://github.com/kubernetes/kubernetes/pull/87400), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Removed the 'client' label from apiserver_request_total. ([#87669](https://github.com/kubernetes/kubernetes/pull/87669), [@logicalhan](https://github.com/logicalhan)) [SIG API Machinery and Instrumentation]
- The CSR signing cert/key pairs will be reloaded from disk like the kube-apiserver cert/key pairs ([#86816](https://github.com/kubernetes/kubernetes/pull/86816), [@deads2k](https://github.com/deads2k)) [SIG API Machinery, Apps and Auth]
- Update cri-tools to v1.17.0 ([#86305](https://github.com/kubernetes/kubernetes/pull/86305), [@saschagrunert](https://github.com/saschagrunert)) [SIG Cluster Lifecycle and Release]
`

const alpha1ExpectedHTML = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width" />
    <title>v1.19.0-alpha.1</title>
    <style type="text/css">
      table,
      th,
      tr,
      td {
        border: 1px solid gray;
        border-collapse: collapse;
        padding: 5px;
      }
    </style>
  </head>
  <body>
    <h1>v1.19.0-alpha.1</h1>
<h2>Changelog since v1.18.0</h2>
<h2>Changes by Kind</h2>
<h3>Deprecation</h3>
<ul>
<li>Kubeadm: remove the deprecated &quot;kubeadm alpha kubelet config enable-dynamic&quot; command. To continue using the feature please defer to the guide for &quot;Dynamic Kubelet Configuration&quot; at k8s.io. This change also removes the parent command &quot;kubeadm alpha kubelet&quot; as there are no more sub-commands under it for the time being. (<a href="https://github.com/kubernetes/kubernetes/pull/94668">#94668</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: remove the deprecated --kubelet-config flag for the command &quot;kubeadm upgrade node&quot; (<a href="https://github.com/kubernetes/kubernetes/pull/94869">#94869</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Kubelet's deprecated endpoint <code>metrics/resource/v1alpha1</code> has been removed, please adopt to <code>metrics/resource</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/94272">#94272</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>) [SIG Instrumentation and Node]</li>
<li>The v1alpha1 PodPreset API and admission plugin has been removed with no built-in replacement. Admission webhooks can be used to modify pods on creation. (<a href="https://github.com/kubernetes/kubernetes/pull/94090">#94090</a>, <a href="https://github.com/deads2k">@deads2k</a>) [SIG API Machinery, Apps, CLI, Cloud Provider, Scalability and Testing]</li>
</ul>
<h3>API Change</h3>
<ul>
<li>A new <code>nofuzz</code> go build tag now disables gofuzz support. Release binaries enable this. (<a href="https://github.com/kubernetes/kubernetes/pull/92491">#92491</a>, <a href="https://github.com/BenTheElder">@BenTheElder</a>) [SIG API Machinery]</li>
<li>External facing API podresources is now available under k8s.io/kubelet/pkg/apis/ (<a href="https://github.com/kubernetes/kubernetes/pull/92632">#92632</a>, <a href="https://github.com/RenaudWasTaken">@RenaudWasTaken</a>) [SIG Node and Testing]</li>
<li>Fix conversions for custom metrics. (<a href="https://github.com/kubernetes/kubernetes/pull/94481">#94481</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>) [SIG API Machinery and Instrumentation]</li>
<li>Kube-controller-manager: volume plugins can be restricted from contacting local and loopback addresses by setting <code>--volume-host-allow-local-loopback=false</code>, or from contacting specific CIDR ranges by setting <code>--volume-host-cidr-denylist</code> (for example, <code>--volume-host-cidr-denylist=127.0.0.1/28,feed::/16</code>) (<a href="https://github.com/kubernetes/kubernetes/pull/91785">#91785</a>, <a href="https://github.com/mattcary">@mattcary</a>) [SIG API Machinery, Apps, Auth, CLI, Network, Node, Storage and Testing]</li>
<li>Migrate scheduler, controller-manager and cloud-controller-manager to use LeaseLock (<a href="https://github.com/kubernetes/kubernetes/pull/94603">#94603</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>) [SIG API Machinery, Apps, Cloud Provider and Scheduling]</li>
<li>Modify DNS-1123 error messages to indicate that RFC 1123 is not followed exactly (<a href="https://github.com/kubernetes/kubernetes/pull/94182">#94182</a>, <a href="https://github.com/mattfenwick">@mattfenwick</a>) [SIG API Machinery, Apps, Auth, Network and Node]</li>
<li>The ServiceAccountIssuerDiscovery feature gate is now Beta and enabled by default. (<a href="https://github.com/kubernetes/kubernetes/pull/91921">#91921</a>, <a href="https://github.com/mtaufen">@mtaufen</a>) [SIG Auth]</li>
<li>WinOverlay feature graduated to beta (<a href="https://github.com/kubernetes/kubernetes/pull/94807">#94807</a>, <a href="https://github.com/ksubrmnn">@ksubrmnn</a>) [SIG Windows]</li>
</ul>
<h3>Feature</h3>
<ul>
<li>Add metrics for azure service operations (route and loadbalancer). (<a href="https://github.com/kubernetes/kubernetes/pull/94124">#94124</a>, <a href="https://github.com/nilo19">@nilo19</a>) [SIG Cloud Provider and Instrumentation]</li>
<li>Add network rule support in Azure account creation (<a href="https://github.com/kubernetes/kubernetes/pull/94239">#94239</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>) [SIG Cloud Provider]</li>
<li>Kubeadm: Add a preflight check that the control-plane node has at least 1700MB of RAM (<a href="https://github.com/kubernetes/kubernetes/pull/93275">#93275</a>, <a href="https://github.com/xlgao-zju">@xlgao-zju</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: add the &quot;--cluster-name&quot; flag to the &quot;kubeadm alpha kubeconfig user&quot; to allow configuring the cluster name in the generated kubeconfig file (<a href="https://github.com/kubernetes/kubernetes/pull/93992">#93992</a>, <a href="https://github.com/prabhu43">@prabhu43</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: add the &quot;--kubeconfig&quot; flag to the &quot;kubeadm init phase upload-certs&quot; command to allow users to pass a custom location for a kubeconfig file. (<a href="https://github.com/kubernetes/kubernetes/pull/94765">#94765</a>, <a href="https://github.com/zhanw15">@zhanw15</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: make etcd pod request 100m CPU, 100Mi memory and 100Mi ephemeral_storage by default (<a href="https://github.com/kubernetes/kubernetes/pull/94479">#94479</a>, <a href="https://github.com/knight42">@knight42</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: make the command &quot;kubeadm alpha kubeconfig user&quot; accept a &quot;--config&quot; flag and remove the following flags:
<ul>
<li>apiserver-advertise-address / apiserver-bind-port: use either localAPIEndpoint from InitConfiguration or controlPlaneEndpoint from ClusterConfiguration.</li>
<li>cluster-name: use clusterName from ClusterConfiguration</li>
<li>cert-dir: use certificatesDir from ClusterConfiguration (<a href="https://github.com/kubernetes/kubernetes/pull/94879">#94879</a>, <a href="https://github.com/knight42">@knight42</a>) [SIG Cluster Lifecycle]</li>
</ul>
</li>
<li>Kubemark now supports both real and hollow nodes in a single cluster. (<a href="https://github.com/kubernetes/kubernetes/pull/93201">#93201</a>, <a href="https://github.com/ellistarn">@ellistarn</a>) [SIG Scalability]</li>
<li>Kubernetes is now built using go1.15.2
<ul>
<li>
<p>build: Update to <a href="mailto:k/repo-infra@v0.1.1">k/repo-infra@v0.1.1</a> (supports go1.15.2)</p>
</li>
<li>
<p>build: Use go-runner:buster-v2.0.1 (built using go1.15.1)</p>
</li>
<li>
<p>bazel: Replace --features with Starlark build settings flag</p>
</li>
<li>
<p>hack/lib/util.sh: some bash cleanups</p>
<ul>
<li>switched one spot to use kube::logging</li>
<li>make kube::util::find-binary return an error when it doesn't find
anything so that hack scripts fail fast instead of with '' binary not
found errors.</li>
<li>this required deleting some genfeddoc stuff. the binary no longer
exists in k/k repo since we removed federation/, and I don't see it
in <a href="https://github.com/kubernetes-sigs/kubefed/">https://github.com/kubernetes-sigs/kubefed/</a> either. I'm assuming
that it's gone for good now.</li>
</ul>
</li>
<li>
<p>bazel: output go_binary rule directly from go_binary_conditional_pure</p>
<p>From: @mikedanese:
Instead of aliasing. Aliases are annoying in a number of ways. This is
specifically bugging me now because they make the action graph harder to
analyze programmatically. By using aliases here, we would need to handle
potentially aliased go_binary targets and dereference to the effective
target.</p>
<p>The comment references an issue with <code>pure = select(...)</code> which appears
to be resolved considering this now builds.</p>
</li>
<li>
<p>make kube::util::find-binary not dependent on bazel-out/ structure</p>
<p>Implement an aspect that outputs go_build_mode metadata for go binaries,
and use that during binary selection. (<a href="https://github.com/kubernetes/kubernetes/pull/94449">#94449</a>, <a href="https://github.com/justaugustus">@justaugustus</a>) [SIG Architecture, CLI, Cluster Lifecycle, Node, Release and Testing]</p>
</li>
</ul>
</li>
<li>Only update Azure data disks when attach/detach (<a href="https://github.com/kubernetes/kubernetes/pull/94265">#94265</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>) [SIG Cloud Provider]</li>
<li>Promote SupportNodePidsLimit to GA to provide node to pod pid isolation
Promote SupportPodPidsLimit to GA to provide ability to limit pids per pod (<a href="https://github.com/kubernetes/kubernetes/pull/94140">#94140</a>, <a href="https://github.com/derekwaynecarr">@derekwaynecarr</a>) [SIG Node and Testing]</li>
<li>Support [service.beta.kubernetes.io/azure-pip-ip-tags] annotations to allow customers to specify ip-tags to influence public-ip creation in Azure [Tag1=Value1, Tag2=Value2, etc.] (<a href="https://github.com/kubernetes/kubernetes/pull/94114">#94114</a>, <a href="https://github.com/MarcPow">@MarcPow</a>) [SIG Cloud Provider]</li>
</ul>
<h3>Documentation</h3>
<ul>
<li>Kubelet: remove alpha warnings for CNI flags. (<a href="https://github.com/kubernetes/kubernetes/pull/94508">#94508</a>, <a href="https://github.com/andrewsykim">@andrewsykim</a>) [SIG Network and Node]</li>
</ul>
<h3>Bug or Regression</h3>
<ul>
<li>
<p>Add kubectl wait  --ignore-not-found flag (<a href="https://github.com/kubernetes/kubernetes/pull/90969">#90969</a>, <a href="https://github.com/zhouya0">@zhouya0</a>) [SIG CLI]</p>
</li>
<li>
<p>Adding fix to the statefulset controller to wait for pvc deletion before creating pods. (<a href="https://github.com/kubernetes/kubernetes/pull/93457">#93457</a>, <a href="https://github.com/ymmt2005">@ymmt2005</a>) [SIG Apps]</p>
</li>
<li>
<p>Azure ARM client: don't segfault on empty response and http error (<a href="https://github.com/kubernetes/kubernetes/pull/94078">#94078</a>, <a href="https://github.com/bpineau">@bpineau</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Azure: fix a bug that kube-controller-manager would panic if wrong Azure VMSS name is configured (<a href="https://github.com/kubernetes/kubernetes/pull/94306">#94306</a>, <a href="https://github.com/knight42">@knight42</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Both apiserver_request_duration_seconds metrics and RequestReceivedTimestamp field of an audit event take
into account the time a request spends in the apiserver request filters. (<a href="https://github.com/kubernetes/kubernetes/pull/94903">#94903</a>, <a href="https://github.com/tkashem">@tkashem</a>) [SIG API Machinery, Auth and Instrumentation]</p>
</li>
<li>
<p>Build/lib/release: Explicitly use '--platform' in building server images</p>
<p>When we switched to go-runner for building the apiserver,
controller-manager, and scheduler server components, we no longer
reference the individual architectures in the image names, specifically
in the 'FROM' directive of the server image Dockerfiles.</p>
<p>As a result, server images for non-amd64 images copy in the go-runner
amd64 binary instead of the go-runner that matches that architecture.</p>
<p>This commit explicitly sets the '--platform=linux/${arch}' to ensure
we're pulling the correct go-runner arch from the manifest list.</p>
<p>Before:
<code>FROM ${base_image}</code></p>
<p>After:
<code>FROM --platform=linux/${arch} ${base_image}</code> (<a href="https://github.com/kubernetes/kubernetes/pull/94552">#94552</a>, <a href="https://github.com/justaugustus">@justaugustus</a>) [SIG Release]</p>
</li>
<li>
<p>CSIDriver object can be deployed during volume attachment. (<a href="https://github.com/kubernetes/kubernetes/pull/93710">#93710</a>, <a href="https://github.com/Jiawei0227">@Jiawei0227</a>) [SIG Apps, Node, Storage and Testing]</p>
</li>
<li>
<p>Do not fail sorting empty elements. (<a href="https://github.com/kubernetes/kubernetes/pull/94666">#94666</a>, <a href="https://github.com/soltysh">@soltysh</a>) [SIG CLI]</p>
</li>
<li>
<p>Dual-stack: make nodeipam compatible with existing single-stack clusters when dual-stack feature gate become enabled by default (<a href="https://github.com/kubernetes/kubernetes/pull/90439">#90439</a>, <a href="https://github.com/SataQiu">@SataQiu</a>) [SIG API Machinery]</p>
</li>
<li>
<p>Ensure backoff step is set to 1 for Azure armclient. (<a href="https://github.com/kubernetes/kubernetes/pull/94180">#94180</a>, <a href="https://github.com/feiskyer">@feiskyer</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Ensure getPrimaryInterfaceID not panic when network interfaces for Azure VMSS are null (<a href="https://github.com/kubernetes/kubernetes/pull/94355">#94355</a>, <a href="https://github.com/feiskyer">@feiskyer</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Fix HandleCrash order (<a href="https://github.com/kubernetes/kubernetes/pull/93108">#93108</a>, <a href="https://github.com/lixiaobing1">@lixiaobing1</a>) [SIG API Machinery]</p>
</li>
<li>
<p>Fix a concurrent map writes error in kubelet (<a href="https://github.com/kubernetes/kubernetes/pull/93773">#93773</a>, <a href="https://github.com/knight42">@knight42</a>) [SIG Node]</p>
</li>
<li>
<p>Fix a regression where kubeadm bails out with a fatal error when an optional version command line argument is supplied to the &quot;kubeadm upgrade plan&quot; command (<a href="https://github.com/kubernetes/kubernetes/pull/94421">#94421</a>, <a href="https://github.com/rosti">@rosti</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Fix azure file migration panic (<a href="https://github.com/kubernetes/kubernetes/pull/94853">#94853</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Fix bug where loadbalancer deletion gets stuck because of missing resource group #75198 (<a href="https://github.com/kubernetes/kubernetes/pull/93962">#93962</a>, <a href="https://github.com/phiphi282">@phiphi282</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Fix etcd_object_counts metric reported by kube-apiserver (<a href="https://github.com/kubernetes/kubernetes/pull/94773">#94773</a>, <a href="https://github.com/tkashem">@tkashem</a>) [SIG API Machinery]</p>
</li>
<li>
<p>Fix incorrectly reported verbs for kube-apiserver metrics for CRD objects (<a href="https://github.com/kubernetes/kubernetes/pull/93523">#93523</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>) [SIG API Machinery and Instrumentation]</p>
</li>
<li>
<p>Fix kubectl SchemaError on CRDs with schema using x-kubernetes-preserve-unknown-fields on array types. (<a href="https://github.com/kubernetes/kubernetes/pull/94888">#94888</a>, <a href="https://github.com/sttts">@sttts</a>) [SIG API Machinery]</p>
</li>
<li>
<p>Fix missing csi annotations on node during parallel csinode update. (<a href="https://github.com/kubernetes/kubernetes/pull/94389">#94389</a>, <a href="https://github.com/pacoxu">@pacoxu</a>) [SIG Storage]</p>
</li>
<li>
<p>Fix the <code>cloudprovider_azure_api_request_duration_seconds</code> metric buckets to correctly capture the latency metrics. Previously, the majority of the calls would fall in the &quot;+Inf&quot; bucket. (<a href="https://github.com/kubernetes/kubernetes/pull/94873">#94873</a>, <a href="https://github.com/marwanad">@marwanad</a>) [SIG Cloud Provider and Instrumentation]</p>
</li>
<li>
<p>Fix: azure disk resize error if source does not exist (<a href="https://github.com/kubernetes/kubernetes/pull/93011">#93011</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Fix: detach azure disk broken on Azure Stack (<a href="https://github.com/kubernetes/kubernetes/pull/94885">#94885</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Fix: use sensitiveOptions on Windows mount (<a href="https://github.com/kubernetes/kubernetes/pull/94126">#94126</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>) [SIG Cloud Provider and Storage]</p>
</li>
<li>
<p>Fixed Ceph RBD volume expansion when no ceph.conf exists (<a href="https://github.com/kubernetes/kubernetes/pull/92027">#92027</a>, <a href="https://github.com/juliantaylor">@juliantaylor</a>) [SIG Storage]</p>
</li>
<li>
<p>Fixed a bug where improper storage and comparison of endpoints led to excessive API traffic from the endpoints controller (<a href="https://github.com/kubernetes/kubernetes/pull/94112">#94112</a>, <a href="https://github.com/damemi">@damemi</a>) [SIG Apps, Network and Testing]</p>
</li>
<li>
<p>Fixed a panic in kubectl debug when pod has multiple init containers or ephemeral containers (<a href="https://github.com/kubernetes/kubernetes/pull/94580">#94580</a>, <a href="https://github.com/kiyoshim55">@kiyoshim55</a>) [SIG CLI]</p>
</li>
<li>
<p>Fixed a regression that sometimes prevented <code>kubectl portforward</code> to work when TCP and UDP services were configured on the same port (<a href="https://github.com/kubernetes/kubernetes/pull/94728">#94728</a>, <a href="https://github.com/amorenoz">@amorenoz</a>) [SIG CLI]</p>
</li>
<li>
<p>Fixed bug in reflector that couldn't recover from &quot;Too large resource version&quot; errors with API servers 1.17.0-1.18.5 (<a href="https://github.com/kubernetes/kubernetes/pull/94316">#94316</a>, <a href="https://github.com/janeczku">@janeczku</a>) [SIG API Machinery]</p>
</li>
<li>
<p>Fixed bug where kubectl top pod output is not sorted when --sort-by and --containers flags are used together (<a href="https://github.com/kubernetes/kubernetes/pull/93692">#93692</a>, <a href="https://github.com/brianpursley">@brianpursley</a>) [SIG CLI]</p>
</li>
<li>
<p>Fixed kubelet creating extra sandbox for pods with RestartPolicyOnFailure after all containers succeeded (<a href="https://github.com/kubernetes/kubernetes/pull/92614">#92614</a>, <a href="https://github.com/tnqn">@tnqn</a>) [SIG Node and Testing]</p>
</li>
<li>
<p>Fixes a bug where EndpointSlices would not be recreated after rapid Service recreation. (<a href="https://github.com/kubernetes/kubernetes/pull/94730">#94730</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG Apps, Network and Testing]</p>
</li>
<li>
<p>Fixes a race condition in kubelet pod handling (<a href="https://github.com/kubernetes/kubernetes/pull/94751">#94751</a>, <a href="https://github.com/auxten">@auxten</a>) [SIG Node]</p>
</li>
<li>
<p>Fixes an issue proxying to ipv6 pods without specifying a port (<a href="https://github.com/kubernetes/kubernetes/pull/94834">#94834</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG API Machinery and Network]</p>
</li>
<li>
<p>Ignore root user check when windows pod starts (<a href="https://github.com/kubernetes/kubernetes/pull/92355">#92355</a>, <a href="https://github.com/wawa0210">@wawa0210</a>) [SIG Node and Windows]</p>
</li>
<li>
<p>Increased maximum IOPS of AWS EBS io1 volumes to 64,000 (current AWS maximum). (<a href="https://github.com/kubernetes/kubernetes/pull/90014">#90014</a>, <a href="https://github.com/jacobmarble">@jacobmarble</a>) [SIG Cloud Provider and Storage]</p>
</li>
<li>
<p>K8s.io/apimachinery: runtime.DefaultUnstructuredConverter.FromUnstructured now handles converting integer fields to typed float values (<a href="https://github.com/kubernetes/kubernetes/pull/93250">#93250</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG API Machinery]</p>
</li>
<li>
<p>Kube-proxy now trims extra spaces found in loadBalancerSourceRanges to match Service validation. (<a href="https://github.com/kubernetes/kubernetes/pull/94107">#94107</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG Network]</p>
</li>
<li>
<p>Kubeadm now makes sure the etcd manifest is regenerated upon upgrade even when no etcd version change takes place (<a href="https://github.com/kubernetes/kubernetes/pull/94395">#94395</a>, <a href="https://github.com/rosti">@rosti</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubeadm: avoid a panic when determining if the running version of CoreDNS is supported during upgrades (<a href="https://github.com/kubernetes/kubernetes/pull/94299">#94299</a>, <a href="https://github.com/zouyee">@zouyee</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubeadm: ensure &quot;kubeadm reset&quot; does not unmount the root &quot;/var/lib/kubelet&quot; directory if it is mounted by the user (<a href="https://github.com/kubernetes/kubernetes/pull/93702">#93702</a>, <a href="https://github.com/thtanaka">@thtanaka</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubeadm: ensure the etcd data directory is created with 0700 permissions during control-plane init and join (<a href="https://github.com/kubernetes/kubernetes/pull/94102">#94102</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubeadm: fix the bug that kubeadm tries to call 'docker info' even if the CRI socket was for another CR (<a href="https://github.com/kubernetes/kubernetes/pull/94555">#94555</a>, <a href="https://github.com/SataQiu">@SataQiu</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubeadm: make the kubeconfig files for the kube-controller-manager and kube-scheduler use the LocalAPIEndpoint instead of the ControlPlaneEndpoint. This makes kubeadm clusters more reseliant to version skew problems during immutable upgrades: <a href="https://kubernetes.io/docs/setup/release/version-skew-policy/&amp;#35;kube-controller-manager-kube-scheduler-and-cloud-controller-manager">https://kubernetes.io/docs/setup/release/version-skew-policy/&amp;#35;kube-controller-manager-kube-scheduler-and-cloud-controller-manager</a> (<a href="https://github.com/kubernetes/kubernetes/pull/94398">#94398</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubeadm: relax the validation of kubeconfig server URLs. Allow the user to define custom kubeconfig server URLs without erroring out during validation of existing kubeconfig files (e.g. when using external CA mode). (<a href="https://github.com/kubernetes/kubernetes/pull/94816">#94816</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubelet: assume that swap is disabled when <code>/proc/swaps</code> does not exist (<a href="https://github.com/kubernetes/kubernetes/pull/93931">#93931</a>, <a href="https://github.com/SataQiu">@SataQiu</a>) [SIG Node]</p>
</li>
<li>
<p>NONE (<a href="https://github.com/kubernetes/kubernetes/pull/71269">#71269</a>, <a href="https://github.com/DeliangFan">@DeliangFan</a>) [SIG Node]</p>
</li>
<li>
<p>New Azure instance types do now have correct max data disk count information. (<a href="https://github.com/kubernetes/kubernetes/pull/94340">#94340</a>, <a href="https://github.com/ialidzhikov">@ialidzhikov</a>) [SIG Cloud Provider and Storage]</p>
</li>
<li>
<p>Require feature flag CustomCPUCFSQuotaPeriod if setting a non-default cpuCFSQuotaPeriod in kubelet config. (<a href="https://github.com/kubernetes/kubernetes/pull/94687">#94687</a>, <a href="https://github.com/karan">@karan</a>) [SIG Node]</p>
</li>
<li>
<p>The <code>/debug/api_priority_and_fairness/dump_requests</code> path at an apiserver will no longer return a phantom line for each exempt priority level. (<a href="https://github.com/kubernetes/kubernetes/pull/93406">#93406</a>, <a href="https://github.com/MikeSpreitzer">@MikeSpreitzer</a>) [SIG API Machinery]</p>
</li>
<li>
<p>The kubelet recognizes the --containerd-namespace flag to configure the namespace used by cadvisor. (<a href="https://github.com/kubernetes/kubernetes/pull/87054">#87054</a>, <a href="https://github.com/changyaowei">@changyaowei</a>) [SIG Node]</p>
</li>
<li>
<p>Update Calico to v3.15.2 (<a href="https://github.com/kubernetes/kubernetes/pull/94241">#94241</a>, <a href="https://github.com/lmm">@lmm</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Update default etcd server version to 3.4.13 (<a href="https://github.com/kubernetes/kubernetes/pull/94287">#94287</a>, <a href="https://github.com/jingyih">@jingyih</a>) [SIG API Machinery, Cloud Provider, Cluster Lifecycle and Testing]</p>
</li>
<li>
<p>Use NLB Subnet CIDRs instead of VPC CIDRs in Health Check SG Rules (<a href="https://github.com/kubernetes/kubernetes/pull/93515">#93515</a>, <a href="https://github.com/t0rr3sp3dr0">@t0rr3sp3dr0</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Users will see increase in time for deletion of pods and also guarantee that removal of pod from api server  would mean deletion of all the resources from container runtime. (<a href="https://github.com/kubernetes/kubernetes/pull/92817">#92817</a>, <a href="https://github.com/kmala">@kmala</a>) [SIG Node]</p>
</li>
<li>
<p>Very large patches may now be specified to <code>kubectl patch</code> with the <code>--patch-file</code> flag instead of including them directly on the command line. The <code>--patch</code> and <code>--patch-file</code> flags are mutually exclusive. (<a href="https://github.com/kubernetes/kubernetes/pull/93548">#93548</a>, <a href="https://github.com/smarterclayton">@smarterclayton</a>) [SIG CLI]</p>
</li>
</ul>
<h3>Other (Cleanup or Flake)</h3>
<ul>
<li>Adds a bootstrapping ClusterRole, ClusterRoleBinding and group for /metrics, /livez/<em>, /readyz/</em>, &amp; /healthz/- endpoints. (<a href="https://github.com/kubernetes/kubernetes/pull/93311">#93311</a>, <a href="https://github.com/logicalhan">@logicalhan</a>) [SIG API Machinery, Auth, Cloud Provider and Instrumentation]</li>
<li>Base-images: Update to debian-iptables:buster-v1.3.0
<ul>
<li>Uses iptables 1.8.5</li>
<li>base-images: Update to debian-base:buster-v1.2.0</li>
<li>cluster/images/etcd: Build etcd:3.4.13-1 image
<ul>
<li>Uses debian-base:buster-v1.2.0 (<a href="https://github.com/kubernetes/kubernetes/pull/94733">#94733</a>, <a href="https://github.com/justaugustus">@justaugustus</a>) [SIG API Machinery, Release and Testing]</li>
</ul>
</li>
</ul>
</li>
<li>Fix kubelet to properly log when a container is started. Before, sometimes the log said that a container is dead and was restarted when it was started for the first time. This only happened when using pods with initContainers and regular containers. (<a href="https://github.com/kubernetes/kubernetes/pull/91469">#91469</a>, <a href="https://github.com/rata">@rata</a>) [SIG Node]</li>
<li>Fixes the flooding warning messages about setting volume ownership for configmap/secret volumes (<a href="https://github.com/kubernetes/kubernetes/pull/92878">#92878</a>, <a href="https://github.com/jvanz">@jvanz</a>) [SIG Instrumentation, Node and Storage]</li>
<li>Fixes the message about no auth for metrics in scheduler. (<a href="https://github.com/kubernetes/kubernetes/pull/94035">#94035</a>, <a href="https://github.com/zhouya0">@zhouya0</a>) [SIG Scheduling]</li>
<li>Kubeadm: Separate argument key/value in log msg (<a href="https://github.com/kubernetes/kubernetes/pull/94016">#94016</a>, <a href="https://github.com/mrueg">@mrueg</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: remove support for the &quot;ci/k8s-master&quot; version label. This label has been removed in the Kubernetes CI release process and would no longer work in kubeadm. You can use the &quot;ci/latest&quot; version label instead. See kubernetes/test-infra#18517 (<a href="https://github.com/kubernetes/kubernetes/pull/93626">#93626</a>, <a href="https://github.com/vikkyomkar">@vikkyomkar</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: remove the CoreDNS check for known image digests when applying the addon (<a href="https://github.com/kubernetes/kubernetes/pull/94506">#94506</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Lock ExternalPolicyForExternalIP to default, this feature gate will be removed in 1.22. (<a href="https://github.com/kubernetes/kubernetes/pull/94581">#94581</a>, <a href="https://github.com/knabben">@knabben</a>) [SIG Network]</li>
<li>Service.beta.kubernetes.io/azure-load-balancer-disable-tcp-reset is removed.  All Standard load balancers will always enable tcp resets. (<a href="https://github.com/kubernetes/kubernetes/pull/94297">#94297</a>, <a href="https://github.com/MarcPow">@MarcPow</a>) [SIG Cloud Provider]</li>
<li>Stop propagating SelfLink (deprecated in 1.16) in kube-apiserver (<a href="https://github.com/kubernetes/kubernetes/pull/94397">#94397</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>) [SIG API Machinery and Testing]</li>
<li>Strip unnecessary security contexts on Windows (<a href="https://github.com/kubernetes/kubernetes/pull/93475">#93475</a>, <a href="https://github.com/ravisantoshgudimetla">@ravisantoshgudimetla</a>) [SIG Node, Testing and Windows]</li>
<li>To ensure the code be strong,  add unit test for GetAddressAndDialer (<a href="https://github.com/kubernetes/kubernetes/pull/93180">#93180</a>, <a href="https://github.com/FreeZhang61">@FreeZhang61</a>) [SIG Node]</li>
<li>Update CNI plugins to v0.8.7 (<a href="https://github.com/kubernetes/kubernetes/pull/94367">#94367</a>, <a href="https://github.com/justaugustus">@justaugustus</a>) [SIG Cloud Provider, Network, Node, Release and Testing]</li>
<li>Update cri-tools to <a href="https://github.com/kubernetes-sigs/cri-tools/releases/tag/v1.19.0">v1.19.0</a> (<a href="https://github.com/kubernetes/kubernetes/pull/94307">#94307</a>, <a href="https://github.com/xmudrii">@xmudrii</a>) [SIG Cloud Provider]</li>
<li>Update etcd client side to v3.4.13 (<a href="https://github.com/kubernetes/kubernetes/pull/94259">#94259</a>, <a href="https://github.com/jingyih">@jingyih</a>) [SIG API Machinery and Cloud Provider]</li>
<li><code>kubectl get ingress</code> now prefers the <code>networking.k8s.io/v1</code> over <code>extensions/v1beta1</code> (deprecated since v1.14). To explicitly request the deprecated version, use <code>kubectl get ingress.v1beta1.extensions</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/94309">#94309</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG API Machinery and CLI]</li>
</ul>

  </body>
</html>`

const alphaReleaseExpectedHTMLHead = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width" />
    <title>v1.18.0-alpha.3</title>
    <style type="text/css">
      table,
      th,
      tr,
      td {
        border: 1px solid gray;
        border-collapse: collapse;
        padding: 5px;
      }
    </style>
  </head>
  <body>
    <h1>v1.18.0-alpha.3</h1>
<h2>Changelog since `

const alphaReleaseExpectedHTMLBottom = `
<h2>Changes by Kind</h2>
<h3>Deprecation</h3>
<ul>
<li>Kubeadm: kube-dns is deprecated and will not be supported in a future version (<a href="https://github.com/kubernetes/kubernetes/pull/86574">#86574</a>, <a href="https://github.com/SataQiu">@SataQiu</a>) [SIG Cluster Lifecycle]</li>
<li>Remove all the generators from kubectl run. It will now only create pods. Additionally, deprecates all the flags that are not relevant anymore. (<a href="https://github.com/kubernetes/kubernetes/pull/87077">#87077</a>, <a href="https://github.com/soltysh">@soltysh</a>) [SIG Architecture, CLI and Testing]</li>
</ul>
<h3>API Change</h3>
<ul>
<li>--enable-cadvisor-endpoints is now disabled by default. If you need access to the cAdvisor v1 Json API please enable it explicitly in the kubelet command line. Please note that this flag was deprecated in 1.15 and will be removed in 1.19. (<a href="https://github.com/kubernetes/kubernetes/pull/87440">#87440</a>, <a href="https://github.com/dims">@dims</a>) [SIG Instrumentation, Node and Testing]</li>
<li>Add kubescheduler.config.k8s.io/v1alpha2 (<a href="https://github.com/kubernetes/kubernetes/pull/87628">#87628</a>, <a href="https://github.com/alculquicondor">@alculquicondor</a>) [SIG Scheduling]</li>
<li>The following feature gates are removed, because the associated features were unconditionally enabled in previous releases: CustomResourceValidation, CustomResourceSubresources, CustomResourceWebhookConversion, CustomResourcePublishOpenAPI, CustomResourceDefaulting (<a href="https://github.com/kubernetes/kubernetes/pull/87475">#87475</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG API Machinery]</li>
</ul>
<h3>Feature</h3>
<ul>
<li>
<p>API request throttling (due to a high rate of requests) is now reported in client-go logs at log level 2.  The messages are of the form</p>
<p>Throttling request took 1.50705208s, request: GET:<!-- raw HTML omitted --></p>
<p>The presence of these messages, may indicate to the administrator the need to tune the cluster accordingly. (<a href="https://github.com/kubernetes/kubernetes/pull/87740">#87740</a>, <a href="https://github.com/jennybuckley">@jennybuckley</a>) [SIG API Machinery]</p>
</li>
<li>
<p>Add support for pre-allocated hugepages for more than one page size (<a href="https://github.com/kubernetes/kubernetes/pull/82820">#82820</a>, <a href="https://github.com/odinuge">@odinuge</a>) [SIG Apps]</p>
</li>
<li>
<p>Added more details to taint toleration errors (<a href="https://github.com/kubernetes/kubernetes/pull/87250">#87250</a>, <a href="https://github.com/starizard">@starizard</a>) [SIG Apps and Scheduling]</p>
</li>
<li>
<p>Aggregation api will have alpha support for network proxy (<a href="https://github.com/kubernetes/kubernetes/pull/87515">#87515</a>, <a href="https://github.com/Sh4d1">@Sh4d1</a>) [SIG API Machinery]</p>
</li>
<li>
<p>DisableAvailabilitySetNodes is added to avoid VM list for VMSS clusters. It should only be used when vmType is &quot;vmss&quot; and all the nodes (including masters) are VMSS virtual machines. (<a href="https://github.com/kubernetes/kubernetes/pull/87685">#87685</a>, <a href="https://github.com/feiskyer">@feiskyer</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Kube-apiserver metrics will now include request counts, latencies, and response sizes for /healthz, /livez, and /readyz requests. (<a href="https://github.com/kubernetes/kubernetes/pull/83598">#83598</a>, <a href="https://github.com/jktomer">@jktomer</a>) [SIG API Machinery]</p>
</li>
<li>
<p>Kubeadm: reject a node joining the cluster if a node with the same name already exists (<a href="https://github.com/kubernetes/kubernetes/pull/81056">#81056</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Scheduler: Add DefaultBinder plugin (<a href="https://github.com/kubernetes/kubernetes/pull/87430">#87430</a>, <a href="https://github.com/alculquicondor">@alculquicondor</a>) [SIG Scheduling and Testing]</p>
</li>
<li>
<p>Skip default spreading scoring plugin for pods that define TopologySpreadConstraints (<a href="https://github.com/kubernetes/kubernetes/pull/87566">#87566</a>, <a href="https://github.com/skilxn-go">@skilxn-go</a>) [SIG Scheduling]</p>
</li>
<li>
<p>The kubectl --dry-run flag now accepts the values 'client', 'server', and 'none', to support client-side and server-side dry-run strategies. The boolean and unset values for the --dry-run flag are deprecated and a value will be required in a future version. (<a href="https://github.com/kubernetes/kubernetes/pull/87580">#87580</a>, <a href="https://github.com/julianvmodesto">@julianvmodesto</a>) [SIG CLI]</p>
</li>
<li>
<p>Update CNI version to v0.8.5 (<a href="https://github.com/kubernetes/kubernetes/pull/78819">#78819</a>, <a href="https://github.com/justaugustus">@justaugustus</a>) [SIG API Machinery, Cluster Lifecycle, Network, Release and Testing]</p>
</li>
</ul>
<h3>Bug or Regression</h3>
<ul>
<li>&quot;kubectl describe statefulsets.apps&quot; prints garbage for rolling update partition (<a href="https://github.com/kubernetes/kubernetes/pull/85846">#85846</a>, <a href="https://github.com/phil9909">@phil9909</a>) [SIG CLI]</li>
<li>Fix regression in statefulset conversion which prevented applying a statefulset multiple times. (<a href="https://github.com/kubernetes/kubernetes/pull/87706">#87706</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG Apps and Testing]</li>
<li>Fix the masters rolling upgrade causing thundering herd of LISTs on etcd leading to control plane unavailability. (<a href="https://github.com/kubernetes/kubernetes/pull/86430">#86430</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>) [SIG API Machinery, Node and Testing]</li>
<li>Fixed two scheduler metrics (pending_pods and schedule_attempts_total) not being recorded (<a href="https://github.com/kubernetes/kubernetes/pull/87692">#87692</a>, <a href="https://github.com/everpeace">@everpeace</a>) [SIG Scheduling]</li>
<li>For volumes that allow attaches across multiple nodes, attach and detach operations across different nodes are now executed in parallel. (<a href="https://github.com/kubernetes/kubernetes/pull/87258">#87258</a>, <a href="https://github.com/verult">@verult</a>) [SIG Apps, Node and Storage]</li>
<li>Kubeadm: apply further improvements to the tentative support for concurrent etcd member join. Fixes a bug where multiple members can receive the same hostname. Increase the etcd client dial timeout and retry timeout for add/remove/... operations. (<a href="https://github.com/kubernetes/kubernetes/pull/87505">#87505</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Resolved a performance issue in the node authorizer index maintenance. (<a href="https://github.com/kubernetes/kubernetes/pull/87693">#87693</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG Auth]</li>
<li>Reverted a kubectl azure auth module change where oidc claim spn: prefix was omitted resulting a breaking behavior with existing Azure AD OIDC enabled api-server (<a href="https://github.com/kubernetes/kubernetes/pull/87507">#87507</a>, <a href="https://github.com/weinong">@weinong</a>) [SIG API Machinery, Auth and Cloud Provider]</li>
<li>Shared informers are now more reliable in the face of network disruption. (<a href="https://github.com/kubernetes/kubernetes/pull/86015">#86015</a>, <a href="https://github.com/squeed">@squeed</a>) [SIG API Machinery]</li>
<li>To reduce chances of throttling, VM cache is set to nil when Azure node provisioning state is deleting (<a href="https://github.com/kubernetes/kubernetes/pull/87635">#87635</a>, <a href="https://github.com/feiskyer">@feiskyer</a>) [SIG Cloud Provider]</li>
<li><code>kubectl diff</code> now returns 1 only on diff finding changes, and &gt;1 on kubectl errors. The &quot;exit status code 1&quot; message as also been muted. (<a href="https://github.com/kubernetes/kubernetes/pull/87437">#87437</a>, <a href="https://github.com/apelisse">@apelisse</a>) [SIG CLI and Testing]</li>
</ul>
<h3>Other (Cleanup or Flake)</h3>
<ul>
<li>Kubeadm: remove the deprecated CoreDNS feature-gate. It was set to &quot;true&quot; since v1.11 when the feature went GA. In v1.13 it was marked as deprecated and hidden from the CLI. (<a href="https://github.com/kubernetes/kubernetes/pull/87400">#87400</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Removed the 'client' label from apiserver_request_total. (<a href="https://github.com/kubernetes/kubernetes/pull/87669">#87669</a>, <a href="https://github.com/logicalhan">@logicalhan</a>) [SIG API Machinery and Instrumentation]</li>
<li>The CSR signing cert/key pairs will be reloaded from disk like the kube-apiserver cert/key pairs (<a href="https://github.com/kubernetes/kubernetes/pull/86816">#86816</a>, <a href="https://github.com/deads2k">@deads2k</a>) [SIG API Machinery, Apps and Auth]</li>
<li>Update cri-tools to v1.17.0 (<a href="https://github.com/kubernetes/kubernetes/pull/86305">#86305</a>, <a href="https://github.com/saschagrunert">@saschagrunert</a>) [SIG Cluster Lifecycle and Release]</li>
<li><code>(*&quot;k8s.io/client-go/rest&quot;.Request).{Do,DoRaw,Stream,Watch}</code> now require callers to pass a <code>context.Context</code> as an argument. The context is used for timeout and cancellation signaling and to pass supplementary information to round trippers in the wrapped transport chain. If you don't need any of this functionality, it is sufficient to pass a context created with <code>context.Background()</code> to these functions. The <code>(*&quot;k8s.io/client-go/rest&quot;.Request).Context</code> method is removed now that all methods that execute a request accept a context directly. (<a href="https://github.com/kubernetes/kubernetes/pull/87597">#87597</a>, <a href="https://github.com/mikedanese">@mikedanese</a>) [SIG API Machinery, Apps, Auth, Autoscaling, CLI, Cloud Provider, Cluster Lifecycle, Instrumentation, Network, Node, Scheduling, Storage and Testing]</li>
</ul>

  </body>
</html>`

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

// nolint: misspell
const minorReleaseExpectedHTML = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width" />
    <title>v1.17.0</title>
    <style type="text/css">
      table,
      th,
      tr,
      td {
        border: 1px solid gray;
        border-collapse: collapse;
        padding: 5px;
      }
    </style>
  </head>
  <body>
    <h1>v1.17.0</h1>
<h2>Changelog since v1.16.0</h2>
<h1>Kubernetes v1.17.0 Release Notes</h1>
<p>A complete changelog for the release notes is now hosted in a customizable format at <a href="https://relnotes.k8s.io">relnotes.k8s.io</a>. Check it out and please give us your feedback!</p>
<h2>What’s New (Major Themes)</h2>
<h3>Cloud Provider Labels reach General Availability</h3>
<p>Added as a beta feature way back in v1.2, v1.17 sees the general availability of cloud provider labels.</p>
<h3>Volume Snapshot Moves to Beta</h3>
<p>The Kubernetes Volume Snapshot feature is now beta in Kubernetes v1.17. It was introduced as alpha in Kubernetes v1.12, with a second alpha with breaking changes in Kubernetes v1.13.</p>
<h3>CSI Migration Beta</h3>
<p>The Kubernetes in-tree storage plugin to Container Storage Interface (CSI) migration infrastructure is now beta in Kubernetes v1.17. CSI migration was introduced as alpha in Kubernetes v1.14.</p>
<h2>Known Issues</h2>
<ul>
<li>volumeDevices mapping ignored when container is privileged</li>
<li>The <code>Should recreate evicted statefulset</code> conformance <a href="https://github.com/kubernetes/kubernetes/blob/master/test/e2e/apps/statefulset.go">test</a> fails because <code>Pod ss-0 expected to be re-created at least once</code>. This was caused by the <code>Predicate PodFitsHostPorts failed</code> scheduling error. The root cause was a host port conflict for port <code>21017</code>. This port was in-use as an ephemeral port by another application running on the node. This will be looked at for the 1.18 release.</li>
</ul>
<h2>Urgent Upgrade Notes</h2>
<h3>(No, really, you MUST read this before you upgrade)</h3>
<h4>Cluster Lifecycle</h4>
<ul>
<li>Kubeadm: add a new <code>kubelet-finalize</code> phase as part of the <code>init</code> workflow and an experimental sub-phase to enable automatic kubelet client certificate rotation on primary control-plane nodes.
Prior to 1.17 and for existing nodes created by <code>kubeadm init</code> where kubelet client certificate rotation is desired, you must modify <code>/etc/kubernetes/kubelet.conf</code> to point to the PEM symlink for rotation:
<code>client-certificate: /var/lib/kubelet/pki/kubelet-client-current.pem</code> and <code>client-key: /var/lib/kubelet/pki/kubelet-client-current.pem</code>, replacing the embedded client certificate and key. (<a href="https://github.com/kubernetes/kubernetes/pull/84118">#84118</a>, <a href="https://github.com/neolit123">@neolit123</a>)</li>
</ul>
<h4>Network</h4>
<ul>
<li>EndpointSlices: If upgrading a cluster with EndpointSlices already enabled, any EndpointSlices that should be managed by the EndpointSlice controller should have a <code>http://endpointslice.kubernetes.io/managed-by</code> label set to <code>endpointslice-controller.k8s.io</code>.</li>
</ul>
<h4>Scheduling</h4>
<ul>
<li>Kubeadm: when adding extra apiserver authorization-modes, the defaults <code>Node,RBAC</code> are no longer prepended in the resulting static Pod manifests and a full override is allowed. (<a href="https://github.com/kubernetes/kubernetes/pull/82616">#82616</a>, <a href="https://github.com/ghouscht">@ghouscht</a>)</li>
</ul>
<h4>Storage</h4>
<ul>
<li>All nodes need to be drained before upgrading Kubernetes cluster, because paths used for block volumes are changed in this release, so on-line upgrade of nodes aren't allowed. (<a href="https://github.com/kubernetes/kubernetes/pull/74026">#74026</a>, <a href="https://github.com/mkimuram">@mkimuram</a>)</li>
</ul>
<h4>Windows</h4>
<ul>
<li>The Windows containers RunAsUsername feature is now beta.</li>
<li>Windows worker nodes in a Kubernetes cluster now support Windows Server version 1903 in addition to the existing support for Windows Server 2019</li>
<li>The RuntimeClass scheduler can now simplify steering Linux or Windows pods to appropriate nodes</li>
<li>All Windows nodes now get the new label <code>node.kubernetes.io/windows-build</code> that reflects the Windows major, minor, and build number that are needed to match compatibility between Windows containers and Windows worker nodes.</li>
</ul>
<h2>Deprecations and Removals</h2>
<ul>
<li><code>kubeadm.k8s.io/v1beta1</code> has been deprecated, you should update your config to use newer non-deprecated API versions. (<a href="https://github.com/kubernetes/kubernetes/pull/83276">#83276</a>, <a href="https://github.com/Klaven">@Klaven</a>)</li>
<li>The deprecated feature gates GCERegionalPersistentDisk, EnableAggregatedDiscoveryTimeout and PersistentLocalVolumes are now unconditionally enabled and can no longer be specified in component invocations. (<a href="https://github.com/kubernetes/kubernetes/pull/82472">#82472</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>Deprecate the default service IP CIDR. The previous default was <code>10.0.0.0/24</code> which will be removed in 6 months/2 releases. Cluster admins must specify their own desired value, by using <code>--service-cluster-ip-range</code> on kube-apiserver. (<a href="https://github.com/kubernetes/kubernetes/pull/81668">#81668</a>, <a href="https://github.com/darshanime">@darshanime</a>)</li>
<li>Remove deprecated &quot;include-uninitialized&quot; flag. (<a href="https://github.com/kubernetes/kubernetes/pull/80337">#80337</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>All resources within the <code>rbac.authorization.k8s.io/v1alpha1</code> and <code>rbac.authorization.k8s.io/v1beta1</code> API groups are deprecated in favor of <code>rbac.authorization.k8s.io/v1</code>, and will no longer be served in v1.20. (<a href="https://github.com/kubernetes/kubernetes/pull/84758">#84758</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>The certificate signer no longer accepts ca.key passwords via the <code>CFSSL_CA_PK_PASSWORD</code> environment variable. This capability was not prompted by user request, never advertised, and recommended against in the security audit. (<a href="https://github.com/kubernetes/kubernetes/pull/84677">#84677</a>, <a href="https://github.com/mikedanese">@mikedanese</a>)</li>
<li>Deprecate the instance type beta label (<code>beta.kubernetes.io/instance-type</code>) in favor of its GA equivalent: <code>node.kubernetes.io/instance-type</code> (<a href="https://github.com/kubernetes/kubernetes/pull/82049">#82049</a>, <a href="https://github.com/andrewsykim">@andrewsykim</a>)</li>
<li>The built-in system:csi-external-provisioner and system:csi-external-attacher cluster roles are removed as of 1.17 release (<a href="https://github.com/kubernetes/kubernetes/pull/84282">#84282</a>, <a href="https://github.com/tedyu">@tedyu</a>)</li>
<li>The in-tree GCE PD plugin <code>kubernetes.io/gce-pd</code> is now deprecated and will be removed in 1.21. Users that self-deploy Kubernetes on GCP should enable CSIMigration + CSIMigrationGCE features and install the GCE PD CSI Driver (<a href="https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver">https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver</a>) to avoid disruption to existing Pod and PVC objects at that time. Users should start using the GCE PD CSI CSI Driver directly for any new volumes. (<a href="https://github.com/kubernetes/kubernetes/pull/85231">#85231</a>, <a href="https://github.com/davidz627">@davidz627</a>)</li>
<li>The in-tree AWS EBS plugin <code>kubernetes.io/aws-ebs</code> is now deprecated and will be removed in 1.21. Users that self-deploy Kubernetes on AWS should enable CSIMigration + CSIMigrationAWS features and install the AWS EBS CSI Driver (<a href="https://github.com/kubernetes-sigs/aws-ebs-csi-driver">https://github.com/kubernetes-sigs/aws-ebs-csi-driver</a>) to avoid disruption to existing Pod and PVC objects at that time. Users should start using the AWS EBS CSI CSI Driver directly for any new volumes. (<a href="https://github.com/kubernetes/kubernetes/pull/85237">#85237</a>, <a href="https://github.com/leakingtapan">@leakingtapan</a>)</li>
<li>The CSINodeInfo feature gate is deprecated and will be removed in a future release. The storage.k8s.io/v1beta1 CSINode object is deprecated and will be removed in a future release. (<a href="https://github.com/kubernetes/kubernetes/pull/83474">#83474</a>, <a href="https://github.com/msau42">@msau42</a>)</li>
<li>Removed Alpha feature <code>MountContainers</code> (<a href="https://github.com/kubernetes/kubernetes/pull/84365">#84365</a>, <a href="https://github.com/codenrhoden">@codenrhoden</a>)</li>
<li>Removed plugin watching of the deprecated directory <code>{kubelet_root_dir}/plugins</code> and CSI V0 support in accordance with deprecation announcement in <a href="https://v1-13.docs.kubernetes.io/docs/setup/release/notes">https://v1-13.docs.kubernetes.io/docs/setup/release/notes</a> (<a href="https://github.com/kubernetes/kubernetes/pull/84533">#84533</a>, <a href="https://github.com/davidz627">@davidz627</a>)</li>
<li>kubeadm deprecates the use of the hyperkube image (<a href="https://github.com/kubernetes/kubernetes/pull/85094">#85094</a>, <a href="https://github.com/rosti">@rosti</a>)</li>
</ul>
<h2>Metrics Changes</h2>
<h3>Added metrics</h3>
<ul>
<li>Add <code>scheduler_goroutines</code> metric to track number of kube-scheduler binding and prioritizing goroutines (<a href="https://github.com/kubernetes/kubernetes/pull/83535">#83535</a>, <a href="https://github.com/wgliang">@wgliang</a>)</li>
<li>Adding initial EndpointSlice metrics. (<a href="https://github.com/kubernetes/kubernetes/pull/83257">#83257</a>, <a href="https://github.com/robscott">@robscott</a>)</li>
<li>Adds a metric <code>apiserver_request_error_total</code> to kube-apiserver. This metric tallies the number of <code>request_errors</code> encountered by verb, group, version, resource, subresource, scope, component, and code. (<a href="https://github.com/kubernetes/kubernetes/pull/83427">#83427</a>, <a href="https://github.com/logicalhan">@logicalhan</a>)</li>
<li>A new <code>kubelet_preemptions</code> metric is reported from Kubelets to track the number of preemptions occuring over time, and which resource is triggering those preemptions. (<a href="https://github.com/kubernetes/kubernetes/pull/84120">#84120</a>, <a href="https://github.com/smarterclayton">@smarterclayton</a>)</li>
<li>Kube-apiserver: Added metrics <code>authentication_latency_seconds</code> that can be used to understand the latency of authentication. (<a href="https://github.com/kubernetes/kubernetes/pull/82409">#82409</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
<li>Add <code>plugin_execution_duration_seconds</code> metric for scheduler framework plugins. (<a href="https://github.com/kubernetes/kubernetes/pull/84522">#84522</a>, <a href="https://github.com/liu-cong">@liu-cong</a>)</li>
<li>Add <code>permit_wait_duration_seconds</code> metric to the scheduler. (<a href="https://github.com/kubernetes/kubernetes/pull/84011">#84011</a>, <a href="https://github.com/liu-cong">@liu-cong</a>)</li>
</ul>
<h3>Deprecated/changed metrics</h3>
<ul>
<li>etcd version monitor metrics are now marked as with the ALPHA stability level. (<a href="https://github.com/kubernetes/kubernetes/pull/83283">#83283</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
<li>Change <code>pod_preemption_victims</code> metric from Gauge to Histogram. (<a href="https://github.com/kubernetes/kubernetes/pull/83603">#83603</a>, <a href="https://github.com/Tabrizian">@Tabrizian</a>)</li>
<li>Following metrics from kubelet are now marked as with the ALPHA stability level:
<code>kubelet_container_log_filesystem_used_bytes</code>
<code>kubelet_volume_stats_capacity_bytes</code>
<code>kubelet_volume_stats_available_bytes</code>
<code>kubelet_volume_stats_used_bytes</code>
<code>kubelet_volume_stats_inodes</code>
<code>kubelet_volume_stats_inodes_free</code>
<code>kubelet_volume_stats_inodes_used</code>
<code>plugin_manager_total_plugins</code>
<code>volume_manager_total_volumes</code>
(<a href="https://github.com/kubernetes/kubernetes/pull/84907">#84907</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
<li>Deprecated metric <code>rest_client_request_latency_seconds</code> has been turned off. (<a href="https://github.com/kubernetes/kubernetes/pull/83836">#83836</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
<li>Following metrics from kubelet are now marked as with the ALPHA stability level:
<code>node_cpu_usage_seconds_total</code>
<code>node_memory_working_set_bytes</code>
<code>container_cpu_usage_seconds_total</code>
<code>container_memory_working_set_bytes</code>
<code>scrape_error</code>
(<a href="https://github.com/kubernetes/kubernetes/pull/84987">#84987</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
<li>Deprecated prometheus request meta-metrics have been removed
<code>http_request_duration_microseconds</code> <code>http_request_duration_microseconds_sum</code> <code>http_request_duration_microseconds_count</code>
<code>http_request_size_bytes</code>
<code>http_request_size_bytes_sum</code>
<code>http_request_size_bytes_count</code>
<code>http_requests_total, http_response_size_bytes</code>
<code>http_response_size_bytes_sum</code>
<code>http_response_size_bytes_count</code>
due to removal from the prometheus client library. Prometheus http request meta-metrics are now generated from <a href="https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#InstrumentMetricHandler"><code>promhttp.InstrumentMetricHandler</code></a> instead.</li>
<li>Following metrics from kube-controller-manager are now marked as with the ALPHA stability level:
<code>storage_count_attachable_volumes_in_use</code>
<code>attachdetach_controller_total_volumes</code>
<code>pv_collector_bound_pv_count</code>
<code>pv_collector_unbound_pv_count</code>
<code>pv_collector_bound_pvc_count</code>
<code>pv_collector_unbound_pvc_count</code>
(<a href="https://github.com/kubernetes/kubernetes/pull/84896">#84896</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
<li>Following metrics have been turned off:
<code>apiserver_request_count</code>
<code>apiserver_request_latencies</code>
<code>apiserver_request_latencies_summary</code>
<code>apiserver_dropped_requests</code>
<code>etcd_request_latencies_summary</code>
<code>apiserver_storage_transformation_latencies_microseconds</code>
<code>apiserver_storage_data_key_generation_latencies_microseconds</code>
<code>apiserver_storage_transformation_failures_total</code>
(<a href="https://github.com/kubernetes/kubernetes/pull/83837">#83837</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
<li>Following metrics have been turned off:
<code>scheduler_scheduling_latency_seconds</code>
<code>scheduler_e2e_scheduling_latency_microseconds</code>
<code>scheduler_scheduling_algorithm_latency_microseconds</code>
<code>scheduler_scheduling_algorithm_predicate_evaluation</code>
<code>scheduler_scheduling_algorithm_priority_evaluation</code>
<code>scheduler_scheduling_algorithm_preemption_evaluation</code>
<code>scheduler_scheduling_binding_latency_microseconds ([#83838](https://github.com/kubernetes/kubernetes/pull/83838</code>), <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
<li>Deprecated metric <code>kubeproxy_sync_proxy_rules_latency_microseconds</code> has been turned off. (<a href="https://github.com/kubernetes/kubernetes/pull/83839">#83839</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
</ul>
<h2>Notable Features</h2>
<h3>Stable</h3>
<ul>
<li>Graduate ScheduleDaemonSetPods to GA. (feature gate will be removed in 1.18) (<a href="https://github.com/kubernetes/kubernetes/pull/82795">#82795</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>Graduate TaintNodesByCondition to GA in 1.17. (feature gate will be removed in 1.18) (<a href="https://github.com/kubernetes/kubernetes/pull/82703">#82703</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>The WatchBookmark feature is promoted to GA. With WatchBookmark feature, clients are able to request watch events with BOOKMARK type. See <a href="https://kubernetes.io/docs/reference/using-api/api-concepts/#watch-bookmarks">https://kubernetes.io/docs/reference/using-api/api-concepts/#watch-bookmarks</a> for more details. (<a href="https://github.com/kubernetes/kubernetes/pull/83195">#83195</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>)</li>
<li>Promote NodeLease feature to GA.
The feature make Lease object changes an additional healthiness signal from Node. Together with that, we reduce frequency of NodeStatus updates to 5m by default in case of no changes to status itself (<a href="https://github.com/kubernetes/kubernetes/pull/84351">#84351</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>)</li>
<li>CSI Topology feature is GA. (<a href="https://github.com/kubernetes/kubernetes/pull/83474">#83474</a>, <a href="https://github.com/msau42">@msau42</a>)</li>
<li>The VolumeSubpathEnvExpansion feature is graduating to GA. The <code>VolumeSubpathEnvExpansion</code> feature gate is unconditionally enabled, and will be removed in v1.19. (<a href="https://github.com/kubernetes/kubernetes/pull/82578">#82578</a>, <a href="https://github.com/kevtaylor">@kevtaylor</a>)</li>
<li>Node-specific volume limits has graduated to GA. (<a href="https://github.com/kubernetes/kubernetes/pull/83568">#83568</a>, <a href="https://github.com/bertinatto">@bertinatto</a>)</li>
<li>The ResourceQuotaScopeSelectors feature has graduated to GA. The <code>ResourceQuotaScopeSelectors</code> feature gate is now unconditionally enabled and will be removed in 1.18. (<a href="https://github.com/kubernetes/kubernetes/pull/82690">#82690</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
</ul>
<h3>Beta</h3>
<ul>
<li>The Kubernetes Volume Snapshot feature has been moved to beta. The VolumeSnapshotDataSource feature gate is on by default in this release. This feature enables you to take a snapshot of a volume (if supported by the CSI driver), and use the snapshot to provision a new volume, pre-populated with data from the snapshot.</li>
<li>Feature gates CSIMigration to Beta (on by default) and CSIMigrationGCE to Beta (off by default since it requires installation of the GCE PD CSI Driver) (<a href="https://github.com/kubernetes/kubernetes/pull/85231">#85231</a>, <a href="https://github.com/davidz627">@davidz627</a>)</li>
<li>EndpointSlices are now beta but not yet enabled by default. Use the EndpointSlice feature gate to enable this feature. (<a href="https://github.com/kubernetes/kubernetes/pull/85365">#85365</a>, <a href="https://github.com/robscott">@robscott</a>)</li>
<li>Promote CSIMigrationAWS to Beta (off by default since it requires installation of the AWS EBS CSI Driver) (<a href="https://github.com/kubernetes/kubernetes/pull/85237">#85237</a>, <a href="https://github.com/leakingtapan">@leakingtapan</a>)</li>
<li>Moving Windows RunAsUserName feature to beta (<a href="https://github.com/kubernetes/kubernetes/pull/84882">#84882</a>, <a href="https://github.com/marosset">@marosset</a>)</li>
</ul>
<h3>CLI Improvements</h3>
<ul>
<li>The kubectl's api-resource command now has a <code>--sort-by</code> flag to sort resources by name or kind. (<a href="https://github.com/kubernetes/kubernetes/pull/81971">#81971</a>, <a href="https://github.com/laddng">@laddng</a>)</li>
<li>A new <code>--prefix</code> flag added into kubectl logs which prepends each log line with information about it's source (pod name and container name) (<a href="https://github.com/kubernetes/kubernetes/pull/76471">#76471</a>, <a href="https://github.com/m1kola">@m1kola</a>)</li>
</ul>
<h2>API Changes</h2>
<ul>
<li>CustomResourceDefinitions now validate documented API semantics of <code>x-kubernetes-list-type</code> and <code>x-kubernetes-map-type</code> atomic to reject non-atomic sub-types. (<a href="https://github.com/kubernetes/kubernetes/pull/84722">#84722</a>, <a href="https://github.com/sttts">@sttts</a>)</li>
<li>Kube-apiserver: The <code>AdmissionConfiguration</code> type accepted by <code>--admission-control-config-file</code> has been promoted to <code>apiserver.config.k8s.io/v1</code> with no schema changes. (<a href="https://github.com/kubernetes/kubernetes/pull/85098">#85098</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>Fixed EndpointSlice port name validation to match Endpoint port name validation (allowing port names longer than 15 characters) (<a href="https://github.com/kubernetes/kubernetes/pull/84481">#84481</a>, <a href="https://github.com/robscott">@robscott</a>)</li>
<li>CustomResourceDefinitions introduce <code>x-kubernetes-map-type</code> annotation as a CRD API extension. Enables this particular validation for server-side apply. (<a href="https://github.com/kubernetes/kubernetes/pull/84113">#84113</a>, <a href="https://github.com/enxebre">@enxebre</a>)</li>
</ul>
<h2>Other notable changes</h2>
<h3>API Machinery</h3>
<ul>
<li>kube-apiserver: the <code>--runtime-config</code> flag now supports an <code>api/beta=false</code> value which disables all built-in REST API versions matching <code>v[0-9]+beta[0-9]+</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/84304">#84304</a>, <a href="https://github.com/liggitt">@liggitt</a>)
The <code>--feature-gates</code> flag now supports an <code>AllBeta=false</code> value which disables all beta feature gates. (<a href="https://github.com/kubernetes/kubernetes/pull/84304">#84304</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>New flag <code>--show-hidden-metrics-for-version</code> in kube-apiserver can be used to show all hidden metrics that deprecated in the previous minor release. (<a href="https://github.com/kubernetes/kubernetes/pull/84292">#84292</a>, <a href="https://github.com/RainbowMango">@RainbowMango</a>)</li>
<li>kube-apiserver: Authentication configuration for mutating and validating admission webhooks referenced from an <code>--admission-control-config-file</code> can now be specified with <code>apiVersion: apiserver.config.k8s.io/v1, kind: WebhookAdmissionConfiguration</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/85138">#85138</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>kube-apiserver: The <code>ResourceQuota</code> admission plugin configuration referenced from <code>--admission-control-config-file</code> admission config has been promoted to <code>apiVersion: apiserver.config.k8s.io/v1</code>, <code>kind: ResourceQuotaConfiguration</code> with no schema changes. (<a href="https://github.com/kubernetes/kubernetes/pull/85099">#85099</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>kube-apiserver: fixed a bug that could cause a goroutine leak if the apiserver encountered an encoding error serving a watch to a websocket watcher (<a href="https://github.com/kubernetes/kubernetes/pull/84693">#84693</a>, <a href="https://github.com/tedyu">@tedyu</a>)</li>
<li>Fix the bug that EndpointSlice for masters wasn't created after enabling EndpointSlice feature on a pre-existing cluster. (<a href="https://github.com/kubernetes/kubernetes/pull/84421">#84421</a>, <a href="https://github.com/tnqn">@tnqn</a>)</li>
<li>Switched intstr.Type to sized integer to follow API guidelines and improve compatibility with proto libraries (<a href="https://github.com/kubernetes/kubernetes/pull/83956">#83956</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>Client-go: improved allocation behavior of the delaying workqueue when handling objects with far-future ready times. (<a href="https://github.com/kubernetes/kubernetes/pull/83945">#83945</a>, <a href="https://github.com/barkbay">@barkbay</a>)</li>
<li>Fixed an issue with informers missing an <code>Added</code> event if a recently deleted object was immediately recreated at the same time the informer dropped a watch and relisted. (<a href="https://github.com/kubernetes/kubernetes/pull/83911">#83911</a>, <a href="https://github.com/matte21">@matte21</a>)</li>
<li>Fixed panic when accessing CustomResources of a CRD with <code>x-kubernetes-int-or-string</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/83787">#83787</a>, <a href="https://github.com/sttts">@sttts</a>)</li>
<li>The resource version option, when passed to a list call, is now consistently interpreted as the minimum allowed resource version.  Previously when listing resources that had the watch cache disabled clients could retrieve a snapshot at that exact resource version.  If the client requests a resource version newer than the current state, a TimeoutError is returned suggesting the client retry in a few seconds.  This behavior is now consistent for both single item retrieval and list calls, and for when the watch cache is enabled or disabled. (<a href="https://github.com/kubernetes/kubernetes/pull/72170">#72170</a>, <a href="https://github.com/jpbetz">@jpbetz</a>)</li>
<li>Fixes a goroutine leak in kube-apiserver when a request times out. (<a href="https://github.com/kubernetes/kubernetes/pull/83333">#83333</a>, <a href="https://github.com/lavalamp">@lavalamp</a>)</li>
<li>Fixes the bug in informer-gen that it produces incorrect code if a type has nonNamespaced tag set. (<a href="https://github.com/kubernetes/kubernetes/pull/80458">#80458</a>, <a href="https://github.com/tatsuhiro-t">@tatsuhiro-t</a>)</li>
<li>Resolves bottleneck in internal API server communication that can cause increased goroutines and degrade API Server performance (<a href="https://github.com/kubernetes/kubernetes/pull/80465">#80465</a>, <a href="https://github.com/answer1991">@answer1991</a>)</li>
<li>Resolves regression generating informers for packages whose names contain <code>.</code> characters (<a href="https://github.com/kubernetes/kubernetes/pull/82410">#82410</a>, <a href="https://github.com/nikhita">@nikhita</a>)</li>
<li>Resolves issue with <code>/readyz</code> and <code>/livez</code> not including etcd and kms health checks (<a href="https://github.com/kubernetes/kubernetes/pull/82713">#82713</a>, <a href="https://github.com/logicalhan">@logicalhan</a>)</li>
<li>Fixes regression in logging spurious stack traces when proxied connections are closed by the backend (<a href="https://github.com/kubernetes/kubernetes/pull/82588">#82588</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>Kube-apiserver now reloads serving certificates from disk every minute to allow rotation without restarting the server process (<a href="https://github.com/kubernetes/kubernetes/pull/84200">#84200</a>, <a href="https://github.com/jackkleeman">@jackkleeman</a>)</li>
<li>Client-ca bundles for the all generic-apiserver based servers will dynamically reload from disk on content changes (<a href="https://github.com/kubernetes/kubernetes/pull/83579">#83579</a>, <a href="https://github.com/deads2k">@deads2k</a>)</li>
<li>Client-go: Clients can request protobuf and json and correctly negotiate with the server for JSON for CRD objects, allowing all client libraries to request protobuf if it is available.  If an error occurs negotiating a watch with the server, the error is immediately return by the client <code>Watch()</code> method instead of being sent as an <code>Error</code> event on the watch stream. (<a href="https://github.com/kubernetes/kubernetes/pull/84692">#84692</a>, <a href="https://github.com/smarterclayton">@smarterclayton</a>)
Renamed FeatureGate RequestManagement to APIPriorityAndFairness.  This feature gate is an alpha and has not yet been associated with any actual functionality. (<a href="https://github.com/kubernetes/kubernetes/pull/85260">#85260</a>, <a href="https://github.com/MikeSpreitzer">@MikeSpreitzer</a>)</li>
<li>Filter published OpenAPI schema by making nullable, required fields non-required in order to avoid kubectl to wrongly reject null values. (<a href="https://github.com/kubernetes/kubernetes/pull/85722">#85722</a>, <a href="https://github.com/sttts">@sttts</a>)</li>
<li>kube-apiserver: fixed a conflict error encountered attempting to delete a pod with <code>gracePeriodSeconds=0</code> and a resourceVersion precondition (<a href="https://github.com/kubernetes/kubernetes/pull/85516">#85516</a>, <a href="https://github.com/michaelgugino">@michaelgugino</a>)</li>
<li>Use context to check client closed instead of http.CloseNotifier in processing watch request which will reduce 1 goroutine for each request if proto is HTTP/2.x . (<a href="https://github.com/kubernetes/kubernetes/pull/85408">#85408</a>, <a href="https://github.com/answer1991">@answer1991</a>)</li>
<li>Reload apiserver SNI certificates from disk every minute (<a href="https://github.com/kubernetes/kubernetes/pull/84303">#84303</a>, <a href="https://github.com/jackkleeman">@jackkleeman</a>)</li>
<li>The mutating and validating admission webhook plugins now read configuration from the admissionregistration.k8s.io/v1 API. (<a href="https://github.com/kubernetes/kubernetes/pull/80883">#80883</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>kube-proxy: a configuration file specified via <code>--config</code> is now loaded with strict deserialization, which fails if the config file contains duplicate or unknown fields. This protects against accidentally running with config files that are malformed, mis-indented, or have typos in field names, and getting unexpected behavior. (<a href="https://github.com/kubernetes/kubernetes/pull/82927">#82927</a>, <a href="https://github.com/obitech">@obitech</a>)</li>
<li>When registering with a 1.17+ API server, MutatingWebhookConfiguration and ValidatingWebhookConfiguration objects can now request that only <code>v1</code> AdmissionReview requests be sent to them. Previously, webhooks were required to support receiving <code>v1beta1</code> AdmissionReview requests as well for compatibility with API servers &lt;= 1.15.
<ul>
<li>When registering with a 1.17+ API server, a CustomResourceDefinition conversion webhook can now request that only <code>v1</code> ConversionReview requests be sent to them. Previously, conversion webhooks were required to support receiving <code>v1beta1</code> ConversionReview requests as well for compatibility with API servers &lt;= 1.15. (<a href="https://github.com/kubernetes/kubernetes/pull/82707">#82707</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
</ul>
</li>
<li>OpenAPI v3 format in CustomResourceDefinition schemas are now documented. (<a href="https://github.com/kubernetes/kubernetes/pull/85381">#85381</a>, <a href="https://github.com/sttts">@sttts</a>)</li>
<li>kube-apiserver: Fixed a regression accepting patch requests &gt; 1MB (<a href="https://github.com/kubernetes/kubernetes/pull/84963">#84963</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>The example API server has renamed its <code>wardle.k8s.io</code> API group to <code>wardle.example.com</code> (<a href="https://github.com/kubernetes/kubernetes/pull/81670">#81670</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>CRDs defaulting is promoted to GA. Note: the feature gate CustomResourceDefaulting will be removed in 1.18. (<a href="https://github.com/kubernetes/kubernetes/pull/84713">#84713</a>, <a href="https://github.com/sttts">@sttts</a>)</li>
<li>Restores compatibility with &lt;=1.15.x custom resources by not publishing OpenAPI for non-structural custom resource definitions (<a href="https://github.com/kubernetes/kubernetes/pull/82653">#82653</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>If given an IPv6 bind-address, kube-apiserver will now advertise an IPv6 endpoint for the kubernetes.default service. (<a href="https://github.com/kubernetes/kubernetes/pull/84727">#84727</a>, <a href="https://github.com/danwinship">@danwinship</a>)</li>
<li>Add table convertor to component status. (<a href="https://github.com/kubernetes/kubernetes/pull/85174">#85174</a>, <a href="https://github.com/zhouya0">@zhouya0</a>)</li>
<li>Scale custom resource unconditionally if resourceVersion is not provided (<a href="https://github.com/kubernetes/kubernetes/pull/80572">#80572</a>, <a href="https://github.com/knight42">@knight42</a>)</li>
<li>When the go-client reflector relists, the ResourceVersion list option is set to the reflector's latest synced resource version to ensure the reflector does not &quot;go back in time&quot; and reprocess events older than it has already processed. If the server responds with an HTTP 410 (Gone) status code response, the relist falls back to using <code>resourceVersion=&quot;&quot;</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/83520">#83520</a>, <a href="https://github.com/jpbetz">@jpbetz</a>)</li>
<li>Fix unsafe JSON construction in a number of locations in the codebase (<a href="https://github.com/kubernetes/kubernetes/pull/81158">#81158</a>, <a href="https://github.com/zouyee">@zouyee</a>)</li>
<li>Fixes a flaw (CVE-2019-11253) in json/yaml decoding where large or malformed documents could consume excessive server resources. Request bodies for normal API requests (create/delete/update/patch operations of regular resources) are now limited to 3MB. (<a href="https://github.com/kubernetes/kubernetes/pull/83261">#83261</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>CRDs can have fields named <code>type</code> with value <code>array</code> and nested array with <code>items</code> fields without validation to fall over this. (<a href="https://github.com/kubernetes/kubernetes/pull/85223">#85223</a>, <a href="https://github.com/sttts">@sttts</a>)</li>
</ul>
<h3>Apps</h3>
<ul>
<li>Support Service Topology (<a href="https://github.com/kubernetes/kubernetes/pull/72046">#72046</a>, <a href="https://github.com/m1093782566">@m1093782566</a>)</li>
<li>Finalizer Protection for Service LoadBalancers is now in GA (enabled by default). This feature ensures the Service resource is not fully deleted until the correlating load balancer resources are deleted. (<a href="https://github.com/kubernetes/kubernetes/pull/85023">#85023</a>, <a href="https://github.com/MrHohn">@MrHohn</a>)</li>
<li>Pod process namespace sharing is now Generally Available. The <code>PodShareProcessNamespace</code> feature gate is now deprecated and will be removed in Kubernetes 1.19. (<a href="https://github.com/kubernetes/kubernetes/pull/84356">#84356</a>, <a href="https://github.com/verb">@verb</a>)</li>
<li>Fix handling tombstones in pod-disruption-budged controller. (<a href="https://github.com/kubernetes/kubernetes/pull/83951">#83951</a>, <a href="https://github.com/zouyee">@zouyee</a>)</li>
<li>Fixed the bug that deleted services were processed by EndpointSliceController repeatedly even their cleanup were successful. (<a href="https://github.com/kubernetes/kubernetes/pull/82996">#82996</a>, <a href="https://github.com/tnqn">@tnqn</a>)</li>
<li>Add <code>RequiresExactMatch</code> for <code>label.Selector</code> (<a href="https://github.com/kubernetes/kubernetes/pull/85048">#85048</a>, <a href="https://github.com/shaloulcy">@shaloulcy</a>)</li>
<li>Adds a new label to indicate what is managing an EndpointSlice. (<a href="https://github.com/kubernetes/kubernetes/pull/83965">#83965</a>, <a href="https://github.com/robscott">@robscott</a>)</li>
<li>Fix handling tombstones in pod-disruption-budged controller. (<a href="https://github.com/kubernetes/kubernetes/pull/83951">#83951</a>, <a href="https://github.com/zouyee">@zouyee</a>)</li>
<li>Fixed the bug that deleted services were processed by EndpointSliceController repeatedly even their cleanup were successful. (<a href="https://github.com/kubernetes/kubernetes/pull/82996">#82996</a>, <a href="https://github.com/tnqn">@tnqn</a>)</li>
<li>An end-user may choose to request logs without confirming the identity of the backing kubelet.  This feature can be disabled by setting the <code>AllowInsecureBackendProxy</code> feature-gate to false. (<a href="https://github.com/kubernetes/kubernetes/pull/83419">#83419</a>, <a href="https://github.com/deads2k">@deads2k</a>)</li>
<li>When scaling down a ReplicaSet, delete doubled up replicas first, where a &quot;doubled up replica&quot; is defined as one that is on the same node as an active replica belonging to a related ReplicaSet.  ReplicaSets are considered &quot;related&quot; if they have a common controller (typically a Deployment). (<a href="https://github.com/kubernetes/kubernetes/pull/80004">#80004</a>, <a href="https://github.com/Miciah">@Miciah</a>)</li>
<li>Kube-controller-manager: Fixes bug setting headless service labels on endpoints (<a href="https://github.com/kubernetes/kubernetes/pull/85361">#85361</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>People can see the right log and note. (<a href="https://github.com/kubernetes/kubernetes/pull/84637">#84637</a>, <a href="https://github.com/zhipengzuo">@zhipengzuo</a>)</li>
<li>Clean duplicate GetPodServiceMemberships function (<a href="https://github.com/kubernetes/kubernetes/pull/83902">#83902</a>, <a href="https://github.com/gongguan">@gongguan</a>)</li>
</ul>
<h3>Auth</h3>
<ul>
<li>K8s docker config json secrets are now compatible with docker config desktop authentication credentials files (<a href="https://github.com/kubernetes/kubernetes/pull/82148">#82148</a>, <a href="https://github.com/bbourbie">@bbourbie</a>)</li>
<li>Kubelet and aggregated API servers now use v1 TokenReview and SubjectAccessReview endpoints to check authentication/authorization. (<a href="https://github.com/kubernetes/kubernetes/pull/84768">#84768</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>Kube-apiserver can now specify <code>--authentication-token-webhook-version=v1</code> or <code>--authorization-webhook-version=v1</code> to use <code>v1</code> TokenReview and SubjectAccessReview API objects when communicating with authentication and authorization webhooks. (<a href="https://github.com/kubernetes/kubernetes/pull/84768">#84768</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>Authentication token cache size is increased (from 4k to 32k) to support clusters with many nodes or many namespaces with active service accounts. (<a href="https://github.com/kubernetes/kubernetes/pull/83643">#83643</a>, <a href="https://github.com/lavalamp">@lavalamp</a>)</li>
<li>Apiservers based on k8s.io/apiserver with delegated authn based on cluster authentication will automatically update to new authentication information when the authoritative configmap is updated. (<a href="https://github.com/kubernetes/kubernetes/pull/85004">#85004</a>, <a href="https://github.com/deads2k">@deads2k</a>)</li>
<li>Configmaps/extension-apiserver-authentication in kube-system is continuously updated by kube-apiservers, instead of just at apiserver start (<a href="https://github.com/kubernetes/kubernetes/pull/82705">#82705</a>, <a href="https://github.com/deads2k">@deads2k</a>)</li>
</ul>
<h3>CLI</h3>
<ul>
<li>Fixed kubectl endpointslice output for get requests (<a href="https://github.com/kubernetes/kubernetes/pull/82603">#82603</a>, <a href="https://github.com/robscott">@robscott</a>)</li>
<li>Gives the right error message when using <code>kubectl delete</code> a wrong resource. (<a href="https://github.com/kubernetes/kubernetes/pull/83825">#83825</a>, <a href="https://github.com/zhouya0">@zhouya0</a>)</li>
<li>If a bad flag is supplied to a kubectl command, only a tip to run <code>--help</code> is printed, instead of the usage menu.  Usage menu is printed upon running <code>kubectl command --help</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/82423">#82423</a>, <a href="https://github.com/sallyom">@sallyom</a>)</li>
<li>Commands like <code>kubectl apply</code> now return errors if schema-invalid annotations are specified, rather than silently dropping the entire annotations section. (<a href="https://github.com/kubernetes/kubernetes/pull/83552">#83552</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>Fixes spurious 0 revisions listed when running <code>kubectl rollout history</code> for a StatefulSet (<a href="https://github.com/kubernetes/kubernetes/pull/82643">#82643</a>, <a href="https://github.com/ZP-AlwaysWin">@ZP-AlwaysWin</a>)</li>
<li>Correct a reference to a not/no longer used kustomize subcommand in the documentation (<a href="https://github.com/kubernetes/kubernetes/pull/82535">#82535</a>, <a href="https://github.com/demobox">@demobox</a>)</li>
<li>Kubectl set resources will no longer return an error if passed an empty change for a resource. kubectl set subject will no longer return an error if passed an empty change for a resource. (<a href="https://github.com/kubernetes/kubernetes/pull/85490">#85490</a>, <a href="https://github.com/sallyom">@sallyom</a>)</li>
<li>Kubectl: --resource-version now works properly in label/annotate/set selector commands when racing with other clients to update the target object (<a href="https://github.com/kubernetes/kubernetes/pull/85285">#85285</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>The <code>--certificate-authority</code> flag now correctly overrides existing skip-TLS or CA data settings in the kubeconfig file (<a href="https://github.com/kubernetes/kubernetes/pull/83547">#83547</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
</ul>
<h3>Cloud Provider</h3>
<ul>
<li>Azure: update disk lock logic per vm during attach/detach to allow concurrent updates for different nodes. (<a href="https://github.com/kubernetes/kubernetes/pull/85115">#85115</a>, <a href="https://github.com/aramase">@aramase</a>)</li>
<li>Fix vmss dirty cache issue in disk attach/detach on vmss node (<a href="https://github.com/kubernetes/kubernetes/pull/85158">#85158</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>)</li>
<li>Fix race condition when attach/delete azure disk in same time (<a href="https://github.com/kubernetes/kubernetes/pull/84917">#84917</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>)</li>
<li>Change GCP ILB firewall names to contain the <code>k8s-fw-</code> prefix like the rest of the firewall rules. This is needed for consistency and also for other components to identify the firewall rule as k8s/service-controller managed. (<a href="https://github.com/kubernetes/kubernetes/pull/84622">#84622</a>, <a href="https://github.com/prameshj">@prameshj</a>)</li>
<li>Ensure health probes are created for local traffic policy UDP services on Azure (<a href="https://github.com/kubernetes/kubernetes/pull/84802">#84802</a>, <a href="https://github.com/feiskyer">@feiskyer</a>)</li>
<li>Openstack: Do not delete managed LB in case of security group reconciliation errors (<a href="https://github.com/kubernetes/kubernetes/pull/82264">#82264</a>, <a href="https://github.com/multi-io">@multi-io</a>)</li>
<li>Fix aggressive VM calls for Azure VMSS (<a href="https://github.com/kubernetes/kubernetes/pull/83102">#83102</a>, <a href="https://github.com/feiskyer">@feiskyer</a>)</li>
<li>Fix: azure disk detach failure if node not exists (<a href="https://github.com/kubernetes/kubernetes/pull/82640">#82640</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>)</li>
<li>Add azure disk encryption(SSE+CMK) support (<a href="https://github.com/kubernetes/kubernetes/pull/84605">#84605</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>)</li>
<li>Update Azure SDK versions to v35.0.0 (<a href="https://github.com/kubernetes/kubernetes/pull/84543">#84543</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>)</li>
<li>Azure: Add allow unsafe read from cache (<a href="https://github.com/kubernetes/kubernetes/pull/83685">#83685</a>, <a href="https://github.com/aramase">@aramase</a>)</li>
<li>Reduces the number of calls made to the Azure API when requesting the instance view of a virtual machine scale set node. (<a href="https://github.com/kubernetes/kubernetes/pull/82496">#82496</a>, <a href="https://github.com/hasheddan">@hasheddan</a>)</li>
<li>Added cloud operation count metrics to azure cloud controller manager. (<a href="https://github.com/kubernetes/kubernetes/pull/82574">#82574</a>, <a href="https://github.com/kkmsft">@kkmsft</a>)</li>
<li>On AWS nodes with multiple network interfaces, kubelet should now more reliably report the same primary node IP. (<a href="https://github.com/kubernetes/kubernetes/pull/80747">#80747</a>, <a href="https://github.com/danwinship">@danwinship</a>)</li>
<li>Update Azure load balancer to prevent orphaned public IP addresses (<a href="https://github.com/kubernetes/kubernetes/pull/82890">#82890</a>, <a href="https://github.com/chewong">@chewong</a>)</li>
</ul>
<h3>Cluster Lifecycle</h3>
<ul>
<li>
<p>Kubeadm alpha certs command now skip missing files (<a href="https://github.com/kubernetes/kubernetes/pull/85092">#85092</a>, <a href="https://github.com/fabriziopandini">@fabriziopandini</a>)</p>
</li>
<li>
<p>Kubeadm: the command &quot;kubeadm token create&quot; now has a &quot;--certificate-key&quot; flag that can be used for the formation of join commands for control-planes with automatic copy of certificates (<a href="https://github.com/kubernetes/kubernetes/pull/84591">#84591</a>, <a href="https://github.com/TheLastProject">@TheLastProject</a>)</p>
</li>
<li>
<p>Kubeadm: Fix a bug where kubeadm cannot parse kubelet's version if the latter dumps logs on the standard error. (<a href="https://github.com/kubernetes/kubernetes/pull/85351">#85351</a>, <a href="https://github.com/rosti">@rosti</a>)</p>
</li>
<li>
<p>Kubeadm: added retry to all the calls to the etcd API so kubeadm will be more resilient to network glitches (<a href="https://github.com/kubernetes/kubernetes/pull/85201">#85201</a>, <a href="https://github.com/fabriziopandini">@fabriziopandini</a>)</p>
</li>
<li>
<p>Fixes a bug in kubeadm that caused init and join to hang indefinitely in specific conditions. (<a href="https://github.com/kubernetes/kubernetes/pull/85156">#85156</a>, <a href="https://github.com/chuckha">@chuckha</a>)</p>
</li>
<li>
<p>Kubeadm now includes CoreDNS version 1.6.5</p>
<ul>
<li><code>kubernetes</code> plugin adds metrics to measure kubernetes control plane latency.</li>
<li>the <code>health</code> plugin now includes the <code>lameduck</code> option by default, which waits for a duration before shutting down. (<a href="https://github.com/kubernetes/kubernetes/pull/85109">#85109</a>, <a href="https://github.com/rajansandeep">@rajansandeep</a>)</li>
</ul>
</li>
<li>
<p>Fixed bug when using kubeadm alpha certs commands with clusters using external etcd (<a href="https://github.com/kubernetes/kubernetes/pull/85091">#85091</a>, <a href="https://github.com/fabriziopandini">@fabriziopandini</a>)</p>
</li>
<li>
<p>Kubeadm no longer defaults or validates the component configs of the kubelet or kube-proxy (<a href="https://github.com/kubernetes/kubernetes/pull/79223">#79223</a>, <a href="https://github.com/rosti">@rosti</a>)</p>
</li>
<li>
<p>Kubeadm: remove the deprecated <code>--cri-socket</code> flag for <code>kubeadm upgrade apply</code>. The flag has been deprecated since v1.14. (<a href="https://github.com/kubernetes/kubernetes/pull/85044">#85044</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>Kubeadm: prevent potential hanging of commands such as &quot;kubeadm reset&quot; if the apiserver endpoint is not reachable. (<a href="https://github.com/kubernetes/kubernetes/pull/84648">#84648</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>Kubeadm: fix skipped etcd upgrade on secondary control-plane nodes when the command <code>kubeadm upgrade node</code> is used. (<a href="https://github.com/kubernetes/kubernetes/pull/85024">#85024</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>Kubeadm: fix an issue with the kube-proxy container env. variables (<a href="https://github.com/kubernetes/kubernetes/pull/84888">#84888</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>Utilize diagnostics tool to dump GKE windows test logs (<a href="https://github.com/kubernetes/kubernetes/pull/83517">#83517</a>, <a href="https://github.com/YangLu1031">@YangLu1031</a>)</p>
</li>
<li>
<p>Kubeadm: always mount the kube-controller-manager hostPath volume that is given by the <code>--flex-volume-plugin-dir</code> flag. (<a href="https://github.com/kubernetes/kubernetes/pull/84468">#84468</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>Update Cluster Autoscaler version to 1.16.2 (CA release docs: <a href="https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.16.2">https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.16.2</a>) (<a href="https://github.com/kubernetes/kubernetes/pull/84038">#84038</a>, <a href="https://github.com/losipiuk">@losipiuk</a>)</p>
</li>
<li>
<p>Kubeadm no longer removes /etc/cni/net.d as it does not install it. Users should remove files from it manually or rely on the component that created them (<a href="https://github.com/kubernetes/kubernetes/pull/83950">#83950</a>, <a href="https://github.com/yastij">@yastij</a>)</p>
</li>
<li>
<p>Kubeadm: fix wrong default value for the <code>upgrade node --certificate-renewal</code> flag. (<a href="https://github.com/kubernetes/kubernetes/pull/83528">#83528</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>Bump metrics-server to v0.3.5 (<a href="https://github.com/kubernetes/kubernetes/pull/83015">#83015</a>, <a href="https://github.com/olagacek">@olagacek</a>)</p>
</li>
<li>
<p>Dashboard: disable the dashboard Deployment on non-Linux nodes. This step is required to support Windows worker nodes. (<a href="https://github.com/kubernetes/kubernetes/pull/82975">#82975</a>, <a href="https://github.com/wawa0210">@wawa0210</a>)</p>
</li>
<li>
<p>Fixes a panic in kube-controller-manager cleaning up bootstrap tokens (<a href="https://github.com/kubernetes/kubernetes/pull/82887">#82887</a>, <a href="https://github.com/tedyu">@tedyu</a>)</p>
</li>
<li>
<p>Kubeadm: add a new <code>kubelet-finalize</code> phase as part of the <code>init</code> workflow and an experimental sub-phase to enable automatic kubelet client certificate rotation on primary control-plane nodes.</p>
<p>Prior to 1.17 and for existing nodes created by <code>kubeadm init</code> where kubelet client certificate rotation is desired, you must modify &quot;/etc/kubernetes/kubelet.conf&quot; to point to the PEM symlink for rotation:
<code>client-certificate: /var/lib/kubelet/pki/kubelet-client-current.pem</code> and <code>client-key: /var/lib/kubelet/pki/kubelet-client-current.pem</code>, replacing the embedded client certificate and key. (<a href="https://github.com/kubernetes/kubernetes/pull/84118">#84118</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>Kubeadm: add a upgrade health check that deploys a Job (<a href="https://github.com/kubernetes/kubernetes/pull/81319">#81319</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>Kubeadm now supports automatic calculations of dual-stack node cidr masks to kube-controller-manager. (<a href="https://github.com/kubernetes/kubernetes/pull/85609">#85609</a>, <a href="https://github.com/Arvinderpal">@Arvinderpal</a>)</p>
</li>
<li>
<p>Kubeadm: reset raises warnings if it cannot delete folders (<a href="https://github.com/kubernetes/kubernetes/pull/85265">#85265</a>, <a href="https://github.com/SataQiu">@SataQiu</a>)</p>
</li>
<li>
<p>Kubeadm: enable the usage of the secure kube-scheduler and kube-controller-manager ports for health checks. For kube-scheduler was 10251, becomes 10259. For kube-controller-manager was 10252, becomes 10257. (<a href="https://github.com/kubernetes/kubernetes/pull/85043">#85043</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>A new kubelet command line option, <code>--reserved-cpus</code>, is introduced to explicitly define the CPU list that will be reserved for system. For example, if <code>--reserved-cpus=0,1,2,3</code> is specified, then cpu 0,1,2,3 will be reserved for the system.  On a system with 24 CPUs, the user may specify <code>isolcpus=4-23</code> for the kernel option and use CPU 4-23 for the user containers. (<a href="https://github.com/kubernetes/kubernetes/pull/83592">#83592</a>, <a href="https://github.com/jianzzha">@jianzzha</a>)</p>
</li>
<li>
<p>Kubelet: a configuration file specified via <code>--config</code> is now loaded with strict deserialization, which fails if the config file contains duplicate or unknown fields. This protects against accidentally running with config files that are malformed, mis-indented, or have typos in field names, and getting unexpected behavior. (<a href="https://github.com/kubernetes/kubernetes/pull/83204">#83204</a>, <a href="https://github.com/obitech">@obitech</a>)</p>
</li>
<li>
<p>Kubeadm now propagates proxy environment variables to kube-proxy (<a href="https://github.com/kubernetes/kubernetes/pull/84559">#84559</a>, <a href="https://github.com/yastij">@yastij</a>)</p>
</li>
<li>
<p>Update the latest validated version of Docker to 19.03 (<a href="https://github.com/kubernetes/kubernetes/pull/84476">#84476</a>, <a href="https://github.com/neolit123">@neolit123</a>)</p>
</li>
<li>
<p>Update to Ingress-GCE v1.6.1 (<a href="https://github.com/kubernetes/kubernetes/pull/84018">#84018</a>, <a href="https://github.com/rramkumar1">@rramkumar1</a>)</p>
</li>
<li>
<p>Kubeadm: enhance certs check-expiration to show the expiration info of related CAs (<a href="https://github.com/kubernetes/kubernetes/pull/83932">#83932</a>, <a href="https://github.com/SataQiu">@SataQiu</a>)</p>
</li>
<li>
<p>Kubeadm: implemented structured output of 'kubeadm token list' in JSON, YAML, Go template and JsonPath formats (<a href="https://github.com/kubernetes/kubernetes/pull/78764">#78764</a>, <a href="https://github.com/bart0sh">@bart0sh</a>)</p>
</li>
<li>
<p>Kubeadm: add support for <code>127.0.0.1</code> as advertise address. kubeadm will automatically replace this value with matching global unicast IP address on the loopback interface. (<a href="https://github.com/kubernetes/kubernetes/pull/83475">#83475</a>, <a href="https://github.com/fabriziopandini">@fabriziopandini</a>)</p>
</li>
<li>
<p>Kube-scheduler: a configuration file specified via <code>--config</code> is now loaded with strict deserialization, which fails if the config file contains duplicate or unknown fields. This protects against accidentally running with config files that are malformed, mis-indented, or have typos in field names, and getting unexpected behavior. (<a href="https://github.com/kubernetes/kubernetes/pull/83030">#83030</a>, <a href="https://github.com/obitech">@obitech</a>)</p>
</li>
<li>
<p>Kubeadm: use the <code>--service-cluster-ip-range</code> flag to init or use the ServiceSubnet field in the kubeadm config to pass a comma separated list of Service CIDRs. (<a href="https://github.com/kubernetes/kubernetes/pull/82473">#82473</a>, <a href="https://github.com/Arvinderpal">@Arvinderpal</a>)</p>
</li>
<li>
<p>Update crictl to v1.16.1. (<a href="https://github.com/kubernetes/kubernetes/pull/82856">#82856</a>, <a href="https://github.com/Random-Liu">@Random-Liu</a>)</p>
</li>
<li>
<p>Bump addon-resizer to 1.8.7 to fix issues with using deprecated extensions APIs (<a href="https://github.com/kubernetes/kubernetes/pull/85864">#85864</a>, <a href="https://github.com/liggitt">@liggitt</a>)</p>
</li>
<li>
<p>Simple script based hyperkube image that bundles all the necessary binaries. This is an equivalent replacement for the image based on the go based hyperkube command + image. (<a href="https://github.com/kubernetes/kubernetes/pull/84662">#84662</a>, <a href="https://github.com/dims">@dims</a>)</p>
</li>
<li>
<p>Hyperkube will now be available in a new Github repository and will not be included in the kubernetes release from 1.17 onwards (<a href="https://github.com/kubernetes/kubernetes/pull/83454">#83454</a>, <a href="https://github.com/dims">@dims</a>)</p>
</li>
<li>
<p>Remove prometheus cluster monitoring addon from kube-up (<a href="https://github.com/kubernetes/kubernetes/pull/83442">#83442</a>, <a href="https://github.com/serathius">@serathius</a>)</p>
</li>
<li>
<p>SourcesReady provides the readiness of kubelet configuration sources such as apiserver update readiness. (<a href="https://github.com/kubernetes/kubernetes/pull/81344">#81344</a>, <a href="https://github.com/zouyee">@zouyee</a>)</p>
</li>
<li>
<p>This PR sets the --cluster-dns flag value to kube-dns service IP whether or not NodeLocal DNSCache is enabled. NodeLocal DNSCache will listen on both the link-local as well as the service IP. (<a href="https://github.com/kubernetes/kubernetes/pull/84383">#84383</a>, <a href="https://github.com/prameshj">@prameshj</a>)</p>
</li>
<li>
<p>kube-dns add-on:</p>
<ul>
<li>All containers are now being executed under more restrictive privileges.</li>
<li>Most of the containers now run as non-root user and has the root filesystem set as read-only.</li>
<li>The remaining container running as root only has the minimum Linux capabilities it requires to run.</li>
<li>Privilege escalation has been disabled for all containers. (<a href="https://github.com/kubernetes/kubernetes/pull/82347">#82347</a>, <a href="https://github.com/pjbgf">@pjbgf</a>)</li>
</ul>
</li>
<li>
<p>Kubernetes no longer monitors firewalld. On systems using firewalld for firewall
maintenance, kube-proxy will take slightly longer to recover from disruptive
firewalld operations that delete kube-proxy's iptables rules.</p>
<p>As a side effect of these changes, kube-proxy's
<code>sync_proxy_rules_last_timestamp_seconds</code> metric no longer behaves the
way it used to; now it will only change when services or endpoints actually
change, rather than reliably updating every 60 seconds (or whatever). If you
are trying to monitor for whether iptables updates are failing, the
<code>sync_proxy_rules_iptables_restore_failures_total</code> metric may be more useful. (<a href="https://github.com/kubernetes/kubernetes/pull/81517">#81517</a>, <a href="https://github.com/danwinship">@danwinship</a>)</p>
</li>
</ul>
<h3>Instrumentation</h3>
<ul>
<li>Bump version of event-exporter to 0.3.1, to switch it to protobuf. (<a href="https://github.com/kubernetes/kubernetes/pull/83396">#83396</a>, <a href="https://github.com/loburm">@loburm</a>)</li>
<li>Bumps metrics-server version to v0.3.6 with following bugfix:
<ul>
<li>Don't break metric storage when duplicate pod metrics encountered causing hpa to fail (<a href="https://github.com/kubernetes/kubernetes/pull/83907">#83907</a>, <a href="https://github.com/olagacek">@olagacek</a>)</li>
</ul>
</li>
<li>addons: elasticsearch discovery supports IPv6 (<a href="https://github.com/kubernetes/kubernetes/pull/85543">#85543</a>, <a href="https://github.com/SataQiu">@SataQiu</a>)</li>
<li>Update Cluster Autoscaler to 1.17.0; changelog: <a href="https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.17.0">https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.17.0</a> (<a href="https://github.com/kubernetes/kubernetes/pull/85610">#85610</a>, <a href="https://github.com/losipiuk">@losipiuk</a>)</li>
</ul>
<h3>Network</h3>
<ul>
<li>
<p>The official kube-proxy image (used by kubeadm, among other things) is now compatible with systems running iptables 1.8 in &quot;nft&quot; mode, and will autodetect which mode it should use. (<a href="https://github.com/kubernetes/kubernetes/pull/82966">#82966</a>, <a href="https://github.com/danwinship">@danwinship</a>)</p>
</li>
<li>
<p>Kubenet: added HostPort IPv6 support.  HostPortManager: operates only with one IP family, failing if receives port mapping entries with different IP families.  HostPortSyncer: operates only with one IP family, skipping portmap entries with different IP families (<a href="https://github.com/kubernetes/kubernetes/pull/80854">#80854</a>, <a href="https://github.com/aojea">@aojea</a>)</p>
</li>
<li>
<p>Kube-proxy now supports DualStack feature with EndpointSlices and IPVS. (<a href="https://github.com/kubernetes/kubernetes/pull/85246">#85246</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>Remove redundant API validation when using Service Topology with externalTrafficPolicy=Local (<a href="https://github.com/kubernetes/kubernetes/pull/85346">#85346</a>, <a href="https://github.com/andrewsykim">@andrewsykim</a>)</p>
</li>
<li>
<p>Update github.com/vishvananda/netlink to v1.0.0 (<a href="https://github.com/kubernetes/kubernetes/pull/83576">#83576</a>, <a href="https://github.com/andrewsykim">@andrewsykim</a>)</p>
</li>
<li>
<p><code>-- kube-controller-manager</code>
<code>--node-cidr-mask-size-ipv4 int32</code>     Default: 24. Mask size for IPv4 node-cidr in dual-stack cluster.
<code>--node-cidr-mask-size-ipv6 int32</code>     Default: 64. Mask size for IPv6 node-cidr in dual-stack cluster.</p>
<p>These 2 flags can be used only for dual-stack clusters. For non dual-stack clusters, continue to use  <code>--node-cidr-mask-size</code> flag to configure the mask size.</p>
<p>The default node cidr mask size for IPv6 was 24 which is now changed to 64. (<a href="https://github.com/kubernetes/kubernetes/pull/79993">#79993</a>, <a href="https://github.com/aramase">@aramase</a>)</p>
</li>
<li>
<p>deprecate cleanup-ipvs flag (<a href="https://github.com/kubernetes/kubernetes/pull/83832">#83832</a>, <a href="https://github.com/gongguan">@gongguan</a>)</p>
</li>
<li>
<p>Kube-proxy: emits a warning when a malformed component config file is used with v1alpha1. (<a href="https://github.com/kubernetes/kubernetes/pull/84143">#84143</a>, <a href="https://github.com/phenixblue">@phenixblue</a>)</p>
</li>
<li>
<p>Set config.BindAddress to IPv4 address <code>127.0.0.1</code> if not specified (<a href="https://github.com/kubernetes/kubernetes/pull/83822">#83822</a>, <a href="https://github.com/zouyee">@zouyee</a>)</p>
</li>
<li>
<p>Updated kube-proxy ipvs README with correct grep argument to list loaded ipvs modules (<a href="https://github.com/kubernetes/kubernetes/pull/83677">#83677</a>, <a href="https://github.com/pete911">@pete911</a>)</p>
</li>
<li>
<p>The userspace mode of kube-proxy no longer confusingly logs messages about deleting endpoints that it is actually adding. (<a href="https://github.com/kubernetes/kubernetes/pull/83644">#83644</a>, <a href="https://github.com/danwinship">@danwinship</a>)</p>
</li>
<li>
<p>Kube-proxy iptables probabilities are now more granular and will result in better distribution beyond 319 endpoints. (<a href="https://github.com/kubernetes/kubernetes/pull/83599">#83599</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>Significant kube-proxy performance improvements for non UDP ports. (<a href="https://github.com/kubernetes/kubernetes/pull/83208">#83208</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>Improved performance of kube-proxy with EndpointSlice enabled with more efficient sorting. (<a href="https://github.com/kubernetes/kubernetes/pull/83035">#83035</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>EndpointSlices are now beta for better Network Endpoint performance at scale. (<a href="https://github.com/kubernetes/kubernetes/pull/84390">#84390</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>Updated EndpointSlices to use PublishNotReadyAddresses from Services. (<a href="https://github.com/kubernetes/kubernetes/pull/84573">#84573</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>When upgrading to 1.17 with a cluster with EndpointSlices enabled, the <code>endpointslice.kubernetes.io/managed-by</code> label needs to be set on each EndpointSlice. (<a href="https://github.com/kubernetes/kubernetes/pull/85359">#85359</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>Adds FQDN addressType support for EndpointSlice. (<a href="https://github.com/kubernetes/kubernetes/pull/84091">#84091</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>Fix incorrect network policy description suggesting that pods are isolated when a network policy has no rules of a given type (<a href="https://github.com/kubernetes/kubernetes/pull/84194">#84194</a>, <a href="https://github.com/jackkleeman">@jackkleeman</a>)</p>
</li>
<li>
<p>Fix bug where EndpointSlice controller would attempt to modify shared objects. (<a href="https://github.com/kubernetes/kubernetes/pull/85368">#85368</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>Splitting IP address type into IPv4 and IPv6 for EndpointSlices (<a href="https://github.com/kubernetes/kubernetes/pull/84971">#84971</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>Added appProtocol field to EndpointSlice Port (<a href="https://github.com/kubernetes/kubernetes/pull/83815">#83815</a>, <a href="https://github.com/howardjohn">@howardjohn</a>)</p>
</li>
<li>
<p>The docker container runtime now enforces a 220 second timeout on container network operations. (<a href="https://github.com/kubernetes/kubernetes/pull/71653">#71653</a>, <a href="https://github.com/liucimin">@liucimin</a>)</p>
</li>
<li>
<p>Fix panic in kubelet when running IPv4/IPv6 dual-stack mode with a CNI plugin (<a href="https://github.com/kubernetes/kubernetes/pull/82508">#82508</a>, <a href="https://github.com/aanm">@aanm</a>)</p>
</li>
<li>
<p>EndpointSlice hostname is now set in the same conditions Endpoints hostname is. (<a href="https://github.com/kubernetes/kubernetes/pull/84207">#84207</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
<li>
<p>Improving the performance of Endpoint and EndpointSlice controllers by caching Service Selectors (<a href="https://github.com/kubernetes/kubernetes/pull/84280">#84280</a>, <a href="https://github.com/gongguan">@gongguan</a>)</p>
</li>
<li>
<p>Significant kube-proxy performance improvements when using Endpoint Slices at scale. (<a href="https://github.com/kubernetes/kubernetes/pull/83206">#83206</a>, <a href="https://github.com/robscott">@robscott</a>)</p>
</li>
</ul>
<h3>Node</h3>
<ul>
<li>
<p>Mirror pods now include an ownerReference for the node that created them. (<a href="https://github.com/kubernetes/kubernetes/pull/84485">#84485</a>, <a href="https://github.com/tallclair">@tallclair</a>)</p>
</li>
<li>
<p>Fixed a bug in the single-numa-policy of the TopologyManager. Previously, best-effort pods would result in a terminated state with a TopologyAffinity error. Now they will run as expected. (<a href="https://github.com/kubernetes/kubernetes/pull/83777">#83777</a>, <a href="https://github.com/lmdaly">@lmdaly</a>)</p>
</li>
<li>
<p>Fixed a bug in the single-numa-node policy of the TopologyManager.  Previously, pods that only requested CPU resources and did not request any third-party devices would fail to launch with a TopologyAffinity error. Now they will launch successfully. (<a href="https://github.com/kubernetes/kubernetes/pull/83697">#83697</a>, <a href="https://github.com/klueska">@klueska</a>)</p>
</li>
<li>
<p>Fix error where metrics related to dynamic kubelet config isn't registered (<a href="https://github.com/kubernetes/kubernetes/pull/83184">#83184</a>, <a href="https://github.com/odinuge">@odinuge</a>)</p>
</li>
<li>
<p>If container fails because ContainerCannotRun, do not utilize the FallbackToLogsOnError TerminationMessagePolicy, as it masks more useful logs. (<a href="https://github.com/kubernetes/kubernetes/pull/81280">#81280</a>, <a href="https://github.com/yqwang-ms">@yqwang-ms</a>)</p>
</li>
<li>
<p>Use online nodes instead of possible nodes when discovering available NUMA nodes (<a href="https://github.com/kubernetes/kubernetes/pull/83196">#83196</a>, <a href="https://github.com/zouyee">@zouyee</a>)</p>
</li>
<li>
<p>Use IPv4 in wincat port forward. (<a href="https://github.com/kubernetes/kubernetes/pull/83036">#83036</a>, <a href="https://github.com/liyanhui1228">@liyanhui1228</a>)</p>
</li>
<li>
<p>Single static pod files and pod files from http endpoints cannot be larger than 10 MB. HTTP probe payloads are now truncated to 10KB. (<a href="https://github.com/kubernetes/kubernetes/pull/82669">#82669</a>, <a href="https://github.com/rphillips">@rphillips</a>)</p>
</li>
<li>
<p>Limit the body length of exec readiness/liveness probes. remote CRIs and Docker shim read a max of 16MB output of which the exec probe itself inspects 10kb. (<a href="https://github.com/kubernetes/kubernetes/pull/82514">#82514</a>, <a href="https://github.com/dims">@dims</a>)</p>
</li>
<li>
<p>Kubelet: Added kubelet serving certificate metric <code>server_rotation_seconds</code> which is a histogram reporting the age of a just rotated serving certificate in seconds. (<a href="https://github.com/kubernetes/kubernetes/pull/84534">#84534</a>, <a href="https://github.com/sambdavidson">@sambdavidson</a>)</p>
</li>
<li>
<p>Reduce default NodeStatusReportFrequency to 5 minutes. With this change, periodic node status updates will be send every 5m if node status doesn't change (otherwise they are still send with 10s).</p>
<p>Bump NodeProblemDetector version to v0.8.0 to reduce forced NodeStatus updates frequency to 5 minutes. (<a href="https://github.com/kubernetes/kubernetes/pull/84007">#84007</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>)</p>
</li>
<li>
<p>The topology manager aligns resources for pods of all QoS classes with respect to NUMA locality, not just Guaranteed QoS pods. (<a href="https://github.com/kubernetes/kubernetes/pull/83492">#83492</a>, <a href="https://github.com/ConnorDoyle">@ConnorDoyle</a>)</p>
</li>
<li>
<p>Fix a bug that a node Lease object may have been created without OwnerReference. (<a href="https://github.com/kubernetes/kubernetes/pull/84998">#84998</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>)</p>
</li>
<li>
<p>External facing APIs in plugin registration and device plugin packages are now available under k8s.io/kubelet/pkg/apis/ (<a href="https://github.com/kubernetes/kubernetes/pull/83551">#83551</a>, <a href="https://github.com/dims">@dims</a>)</p>
</li>
</ul>
<h3>Release</h3>
<ul>
<li>Added the <code>crictl</code> Windows binaries as well as the Linux 32bit binary to the release archives (<a href="https://github.com/kubernetes/kubernetes/pull/83944">#83944</a>, <a href="https://github.com/saschagrunert">@saschagrunert</a>)</li>
<li>Bumps the minimum version of Go required for building Kubernetes to 1.12.4. (<a href="https://github.com/kubernetes/kubernetes/pull/83596">#83596</a>, <a href="https://github.com/jktomer">@jktomer</a>)</li>
<li>The deprecated mondo <code>kubernetes-test</code> tarball is no longer built. Users running Kubernetes e2e tests should use the <code>kubernetes-test-portable</code> and <code>kubernetes-test-{OS}-{ARCH}</code> tarballs instead. (<a href="https://github.com/kubernetes/kubernetes/pull/83093">#83093</a>, <a href="https://github.com/ixdy">@ixdy</a>)</li>
</ul>
<h3>Scheduling</h3>
<ul>
<li>Only validate duplication of the RequestedToCapacityRatio custom priority and allow other custom predicates/priorities (<a href="https://github.com/kubernetes/kubernetes/pull/84646">#84646</a>, <a href="https://github.com/liu-cong">@liu-cong</a>)</li>
<li>Scheduler policy configs can no longer be declared multiple times (<a href="https://github.com/kubernetes/kubernetes/pull/83963">#83963</a>, <a href="https://github.com/damemi">@damemi</a>)</li>
<li>TaintNodesByCondition was graduated to GA, CheckNodeMemoryPressure, CheckNodePIDPressure, CheckNodeDiskPressure, CheckNodeCondition were accidentally removed since 1.12, the replacement is to use CheckNodeUnschedulablePred (<a href="https://github.com/kubernetes/kubernetes/pull/84152">#84152</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>[migration phase 1] PodFitsHostPorts as filter plugin (<a href="https://github.com/kubernetes/kubernetes/pull/83659">#83659</a>, <a href="https://github.com/wgliang">@wgliang</a>)</li>
<li>[migration phase 1] PodFitsResources as framework plugin (<a href="https://github.com/kubernetes/kubernetes/pull/83650">#83650</a>, <a href="https://github.com/wgliang">@wgliang</a>)</li>
<li>[migration phase 1] PodMatchNodeSelector/NodAffinity as filter plugin (<a href="https://github.com/kubernetes/kubernetes/pull/83660">#83660</a>, <a href="https://github.com/wgliang">@wgliang</a>)</li>
<li>Add more tracing steps in generic_scheduler (<a href="https://github.com/kubernetes/kubernetes/pull/83539">#83539</a>, <a href="https://github.com/wgliang">@wgliang</a>)</li>
<li>[migration phase 1] PodFitsHost as filter plugin (<a href="https://github.com/kubernetes/kubernetes/pull/83662">#83662</a>, <a href="https://github.com/wgliang">@wgliang</a>)</li>
<li>Fixed a scheduler panic when using PodAffinity. (<a href="https://github.com/kubernetes/kubernetes/pull/82841">#82841</a>, <a href="https://github.com/Huang-Wei">@Huang-Wei</a>)</li>
<li>Take the context as the first argument of Schedule. (<a href="https://github.com/kubernetes/kubernetes/pull/82119">#82119</a>, <a href="https://github.com/wgliang">@wgliang</a>)</li>
<li>Fixed an issue that the correct PluginConfig.Args is not passed to the corresponding PluginFactory in kube-scheduler when multiple PluginConfig items are defined. (<a href="https://github.com/kubernetes/kubernetes/pull/82483">#82483</a>, <a href="https://github.com/everpeace">@everpeace</a>)</li>
<li>Profiling is enabled by default in the scheduler (<a href="https://github.com/kubernetes/kubernetes/pull/84835">#84835</a>, <a href="https://github.com/denkensk">@denkensk</a>)</li>
<li>Scheduler now reports metrics on cache size including nodes, pods, and assumed pods (<a href="https://github.com/kubernetes/kubernetes/pull/83508">#83508</a>, <a href="https://github.com/damemi">@damemi</a>)</li>
<li>User can now use component config to configure NodeLabel plugin for the scheduler framework. (<a href="https://github.com/kubernetes/kubernetes/pull/84297">#84297</a>, <a href="https://github.com/liu-cong">@liu-cong</a>)</li>
<li>Optimize inter-pod affinity preferredDuringSchedulingIgnoredDuringExecution type, up to 4x in some cases. (<a href="https://github.com/kubernetes/kubernetes/pull/84264">#84264</a>, <a href="https://github.com/ahg-g">@ahg-g</a>)</li>
<li>Filter plugin for cloud provider storage predicate (<a href="https://github.com/kubernetes/kubernetes/pull/84148">#84148</a>, <a href="https://github.com/gongguan">@gongguan</a>)</li>
<li>Refactor scheduler's framework permit API. (<a href="https://github.com/kubernetes/kubernetes/pull/83756">#83756</a>, <a href="https://github.com/hex108">@hex108</a>)</li>
<li>Add incoming pods metrics to scheduler queue. (<a href="https://github.com/kubernetes/kubernetes/pull/83577">#83577</a>, <a href="https://github.com/liu-cong">@liu-cong</a>)</li>
<li>Allow dynamically set glog logging level of kube-scheduler (<a href="https://github.com/kubernetes/kubernetes/pull/83910">#83910</a>, <a href="https://github.com/mrkm4ntr">@mrkm4ntr</a>)</li>
<li>Add latency and request count metrics for scheduler framework. (<a href="https://github.com/kubernetes/kubernetes/pull/83569">#83569</a>, <a href="https://github.com/liu-cong">@liu-cong</a>)</li>
<li>Expose SharedInformerFactory in the framework handle (<a href="https://github.com/kubernetes/kubernetes/pull/83663">#83663</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>Add per-pod scheduling metrics across 1 or more schedule attempts. (<a href="https://github.com/kubernetes/kubernetes/pull/83674">#83674</a>, <a href="https://github.com/liu-cong">@liu-cong</a>)</li>
<li>Add <code>podInitialBackoffDurationSeconds</code> and <code>podMaxBackoffDurationSeconds</code> to the scheduler config API (<a href="https://github.com/kubernetes/kubernetes/pull/81263">#81263</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>Expose kubernetes client in the scheduling framework handle. (<a href="https://github.com/kubernetes/kubernetes/pull/82432">#82432</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>Remove MaxPriority in the scheduler API, please use MaxNodeScore or MaxExtenderPriority instead. (<a href="https://github.com/kubernetes/kubernetes/pull/83386">#83386</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>Consolidate ScoreWithNormalizePlugin into the ScorePlugin interface (<a href="https://github.com/kubernetes/kubernetes/pull/83042">#83042</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>New APIs to allow adding/removing pods from pre-calculated prefilter state in the scheduling framework (<a href="https://github.com/kubernetes/kubernetes/pull/82912">#82912</a>, <a href="https://github.com/ahg-g">@ahg-g</a>)</li>
<li>Added Clone method to the scheduling framework's PluginContext and ContextData. (<a href="https://github.com/kubernetes/kubernetes/pull/82951">#82951</a>, <a href="https://github.com/ahg-g">@ahg-g</a>)</li>
<li>Modified the scheduling framework's Filter API. (<a href="https://github.com/kubernetes/kubernetes/pull/82842">#82842</a>, <a href="https://github.com/ahg-g">@ahg-g</a>)</li>
<li>Critical pods can now be created in namespaces other than kube-system. To limit critical pods to the kube-system namespace, cluster admins should create an admission configuration file limiting critical pods by default, and a matching quota object in the <code>kube-system</code> namespace permitting critical pods in that namespace. See <a href="https://kubernetes.io/docs/concepts/policy/resource-quotas/&amp;#35;limit-priority-class-consumption-by-default">https://kubernetes.io/docs/concepts/policy/resource-quotas/&amp;#35;limit-priority-class-consumption-by-default</a> for details. (<a href="https://github.com/kubernetes/kubernetes/pull/76310">#76310</a>, <a href="https://github.com/ravisantoshgudimetla">@ravisantoshgudimetla</a>)</li>
<li>Scheduler ComponentConfig fields are now pointers (<a href="https://github.com/kubernetes/kubernetes/pull/83619">#83619</a>, <a href="https://github.com/damemi">@damemi</a>)</li>
<li>Scheduler Policy API has a new recommended apiVersion <code>apiVersion: kubescheduler.config.k8s.io/v1</code> which is consistent with the scheduler API group <code>kubescheduler.config.k8s.io</code>. It holds the same API as the old apiVersion <code>apiVersion: v1</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/83578">#83578</a>, <a href="https://github.com/Huang-Wei">@Huang-Wei</a>)</li>
<li>Rename PluginContext to CycleState in the scheduling framework (<a href="https://github.com/kubernetes/kubernetes/pull/83430">#83430</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
<li>Some scheduler extender API fields are moved from <code>pkg/scheduler/api</code> to <code>pkg/scheduler/apis/extender/v1</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/83262">#83262</a>, <a href="https://github.com/Huang-Wei">@Huang-Wei</a>)</li>
<li>Kube-scheduler: emits a warning when a malformed component config file is used with v1alpha1. (<a href="https://github.com/kubernetes/kubernetes/pull/84129">#84129</a>, <a href="https://github.com/obitech">@obitech</a>)</li>
<li>Kube-scheduler now falls back to emitting events using core/v1 Events when events.k8s.io/v1beta1 is disabled. (<a href="https://github.com/kubernetes/kubernetes/pull/83692">#83692</a>, <a href="https://github.com/yastij">@yastij</a>)</li>
<li>Expand scheduler priority functions and scheduling framework plugins' node score range to [0, 100]. Note: this change is internal and does not affect extender and RequestedToCapacityRatio custom priority, which are still expected to provide a [0, 10] range. (<a href="https://github.com/kubernetes/kubernetes/pull/83522">#83522</a>, <a href="https://github.com/draveness">@draveness</a>)</li>
</ul>
<h3>Storage</h3>
<ul>
<li>
<p>Bump CSI version to 1.2.0 (<a href="https://github.com/kubernetes/kubernetes/pull/84832">#84832</a>, <a href="https://github.com/gnufied">@gnufied</a>)</p>
</li>
<li>
<p>CSI Migration: Fixes issue where all volumes with the same inline volume inner spec name were staged in the same path. Migrated inline volumes are now staged at a unique path per unique volume. (<a href="https://github.com/kubernetes/kubernetes/pull/84754">#84754</a>, <a href="https://github.com/davidz627">@davidz627</a>)</p>
</li>
<li>
<p>CSI Migration: GCE PD access mode now reflects read only status of inline volumes - this allows multi-attach for read only many PDs (<a href="https://github.com/kubernetes/kubernetes/pull/84809">#84809</a>, <a href="https://github.com/davidz627">@davidz627</a>)</p>
</li>
<li>
<p>CSI detach timeout increased from 10 seconds to 2 minutes (<a href="https://github.com/kubernetes/kubernetes/pull/84321">#84321</a>, <a href="https://github.com/cduchesne">@cduchesne</a>)</p>
</li>
<li>
<p>Ceph RBD volume plugin now does not use any keyring (<code>/etc/ceph/ceph.client.lvs01cinder.keyring</code>, <code>/etc/ceph/ceph.keyring</code>, <code>/etc/ceph/keyring</code>, <code>/etc/ceph/keyring.bin</code>) for authentication. Ceph user credentials must be provided in PersistentVolume objects and referred Secrets. (<a href="https://github.com/kubernetes/kubernetes/pull/75588">#75588</a>, <a href="https://github.com/smileusd">@smileusd</a>)</p>
</li>
<li>
<p>Validate Gluster IP (<a href="https://github.com/kubernetes/kubernetes/pull/83104">#83104</a>, <a href="https://github.com/zouyee">@zouyee</a>)</p>
</li>
<li>
<p>PersistentVolumeLabel admission plugin, responsible for labeling <code>PersistentVolumes</code> with topology labels, now does not overwrite existing labels on PVs that were dynamically provisioned. It trusts the  dynamic provisioning that it provided the correct labels to the <code>PersistentVolume</code>, saving one potentially expensive cloud API call. <code>PersistentVolumes</code> created manually by users are labelled by the admission plugin in the same way as before. (<a href="https://github.com/kubernetes/kubernetes/pull/82830">#82830</a>, <a href="https://github.com/jsafrane">@jsafrane</a>)</p>
</li>
<li>
<p>Existing PVs are converted to use volume topology if migration is enabled. (<a href="https://github.com/kubernetes/kubernetes/pull/83394">#83394</a>, <a href="https://github.com/bertinatto">@bertinatto</a>)</p>
</li>
<li>
<p>local: support local filesystem volume with block resource reconstruction (<a href="https://github.com/kubernetes/kubernetes/pull/84218">#84218</a>, <a href="https://github.com/cofyc">@cofyc</a>)</p>
</li>
<li>
<p>Fixed binding of block PersistentVolumes / PersistentVolumeClaims when BlockVolume feature is off. (<a href="https://github.com/kubernetes/kubernetes/pull/84049">#84049</a>, <a href="https://github.com/jsafrane">@jsafrane</a>)</p>
</li>
<li>
<p>Report non-confusing error for negative storage size in PVC spec. (<a href="https://github.com/kubernetes/kubernetes/pull/82759">#82759</a>, <a href="https://github.com/sttts">@sttts</a>)</p>
</li>
<li>
<p>Fixed &quot;requested device X but found Y&quot; attach error on AWS. (<a href="https://github.com/kubernetes/kubernetes/pull/85675">#85675</a>, <a href="https://github.com/jsafrane">@jsafrane</a>)</p>
</li>
<li>
<p>Reduced frequency of DescribeVolumes calls of AWS API when attaching/detaching a volume. (<a href="https://github.com/kubernetes/kubernetes/pull/84181">#84181</a>, <a href="https://github.com/jsafrane">@jsafrane</a>)</p>
</li>
<li>
<p>Fixed attachment of AWS volumes that have just been detached. (<a href="https://github.com/kubernetes/kubernetes/pull/83567">#83567</a>, <a href="https://github.com/jsafrane">@jsafrane</a>)</p>
</li>
<li>
<p>Fix possible fd leak and closing of dirs when using openstack (<a href="https://github.com/kubernetes/kubernetes/pull/82873">#82873</a>, <a href="https://github.com/odinuge">@odinuge</a>)</p>
</li>
<li>
<p>local: support local volume block mode reconstruction (<a href="https://github.com/kubernetes/kubernetes/pull/84173">#84173</a>, <a href="https://github.com/cofyc">@cofyc</a>)</p>
</li>
<li>
<p>Fixed cleanup of raw block devices after kubelet restart. (<a href="https://github.com/kubernetes/kubernetes/pull/83451">#83451</a>, <a href="https://github.com/jsafrane">@jsafrane</a>)</p>
</li>
<li>
<p>Add data cache flushing during unmount device for GCE-PD driver in Windows Server. (<a href="https://github.com/kubernetes/kubernetes/pull/83591">#83591</a>, <a href="https://github.com/jingxu97">@jingxu97</a>)</p>
</li>
</ul>
<h3>Windows</h3>
<ul>
<li>Adds Windows Server build information as a label on the node. (<a href="https://github.com/kubernetes/kubernetes/pull/84472">#84472</a>, <a href="https://github.com/gab-satchi">@gab-satchi</a>)</li>
<li>Fixes kube-proxy bug accessing self nodeip:port on windows (<a href="https://github.com/kubernetes/kubernetes/pull/83027">#83027</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>When using Containerd on Windows, the <code>TerminationMessagePath</code> file will now be mounted in the Windows Pod. (<a href="https://github.com/kubernetes/kubernetes/pull/83057">#83057</a>, <a href="https://github.com/bclau">@bclau</a>)</li>
<li>Fix kubelet metrics gathering on non-English Windows hosts (<a href="https://github.com/kubernetes/kubernetes/pull/84156">#84156</a>, <a href="https://github.com/wawa0210">@wawa0210</a>)</li>
</ul>
<h3>Dependencies</h3>
<ul>
<li>Update etcd client side to v3.4.3 (<a href="https://github.com/kubernetes/kubernetes/pull/83987">#83987</a>, <a href="https://github.com/wenjiaswe">@wenjiaswe</a>)</li>
<li>Kubernetes now requires go1.13.4+ to build (<a href="https://github.com/kubernetes/kubernetes/pull/82809">#82809</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>Update to use go1.12.12 (<a href="https://github.com/kubernetes/kubernetes/pull/84064">#84064</a>, <a href="https://github.com/cblecker">@cblecker</a>)</li>
<li>Update to go 1.12.10 (<a href="https://github.com/kubernetes/kubernetes/pull/83139">#83139</a>, <a href="https://github.com/cblecker">@cblecker</a>)</li>
<li>Update default etcd server version to 3.4.3 (<a href="https://github.com/kubernetes/kubernetes/pull/84329">#84329</a>, <a href="https://github.com/jingyih">@jingyih</a>)</li>
<li>Upgrade default etcd server version to 3.3.17 (<a href="https://github.com/kubernetes/kubernetes/pull/83804">#83804</a>, <a href="https://github.com/jpbetz">@jpbetz</a>)</li>
<li>Upgrade to etcd client 3.3.17 to fix bug where etcd client does not parse IPv6 addresses correctly when members are joining, and to fix bug where failover on multi-member etcd cluster fails certificate check on DNS mismatch (<a href="https://github.com/kubernetes/kubernetes/pull/83801">#83801</a>, <a href="https://github.com/jpbetz">@jpbetz</a>)</li>
</ul>
<h3>Detailed go Dependency Changes</h3>
<h4>Added</h4>
<ul>
<li>github.com/OpenPeeDeeP/depguard: v1.0.1</li>
<li>github.com/StackExchange/wmi: 5d04971</li>
<li>github.com/agnivade/levenshtein: v1.0.1</li>
<li>github.com/alecthomas/template: a0175ee</li>
<li>github.com/alecthomas/units: 2efee85</li>
<li>github.com/andreyvit/diff: c7f18ee</li>
<li>github.com/anmitsu/go-shlex: 648efa6</li>
<li>github.com/bazelbuild/rules_go: 6dae44d</li>
<li>github.com/bgentry/speakeasy: v0.1.0</li>
<li>github.com/bradfitz/go-smtpd: deb6d62</li>
<li>github.com/cockroachdb/datadriven: 80d97fb</li>
<li>github.com/creack/pty: v1.1.7</li>
<li>github.com/gliderlabs/ssh: v0.1.1</li>
<li>github.com/go-critic/go-critic: 1df3008</li>
<li>github.com/go-kit/kit: v0.8.0</li>
<li>github.com/go-lintpack/lintpack: v0.5.2</li>
<li>github.com/go-logfmt/logfmt: v0.3.0</li>
<li>github.com/go-ole/go-ole: v1.2.1</li>
<li>github.com/go-stack/stack: v1.8.0</li>
<li>github.com/go-toolsmith/astcast: v1.0.0</li>
<li>github.com/go-toolsmith/astcopy: v1.0.0</li>
<li>github.com/go-toolsmith/astequal: v1.0.0</li>
<li>github.com/go-toolsmith/astfmt: v1.0.0</li>
<li>github.com/go-toolsmith/astinfo: 9809ff7</li>
<li>github.com/go-toolsmith/astp: v1.0.0</li>
<li>github.com/go-toolsmith/pkgload: v1.0.0</li>
<li>github.com/go-toolsmith/strparse: v1.0.0</li>
<li>github.com/go-toolsmith/typep: v1.0.0</li>
<li>github.com/gobwas/glob: v0.2.3</li>
<li>github.com/golangci/check: cfe4005</li>
<li>github.com/golangci/dupl: 3e9179a</li>
<li>github.com/golangci/errcheck: ef45e06</li>
<li>github.com/golangci/go-misc: 927a3d8</li>
<li>github.com/golangci/go-tools: e32c541</li>
<li>github.com/golangci/goconst: 041c5f2</li>
<li>github.com/golangci/gocyclo: 2becd97</li>
<li>github.com/golangci/gofmt: 0b8337e</li>
<li>github.com/golangci/golangci-lint: v1.18.0</li>
<li>github.com/golangci/gosec: 66fb7fc</li>
<li>github.com/golangci/ineffassign: 42439a7</li>
<li>github.com/golangci/lint-1: ee948d0</li>
<li>github.com/golangci/maligned: b1d8939</li>
<li>github.com/golangci/misspell: 950f5d1</li>
<li>github.com/golangci/prealloc: 215b22d</li>
<li>github.com/golangci/revgrep: d9c87f5</li>
<li>github.com/golangci/unconvert: 28b1c44</li>
<li>github.com/google/go-github: v17.0.0+incompatible</li>
<li>github.com/google/go-querystring: v1.0.0</li>
<li>github.com/gostaticanalysis/analysisutil: v0.0.3</li>
<li>github.com/jellevandenhooff/dkim: f50fe3d</li>
<li>github.com/julienschmidt/httprouter: v1.2.0</li>
<li>github.com/klauspost/compress: v1.4.1</li>
<li>github.com/kr/logfmt: b84e30a</li>
<li>github.com/logrusorgru/aurora: a7b3b31</li>
<li>github.com/mattn/go-runewidth: v0.0.2</li>
<li>github.com/mattn/goveralls: v0.0.2</li>
<li>github.com/mitchellh/go-ps: 4fdf99a</li>
<li>github.com/mozilla/tls-observatory: 8791a20</li>
<li>github.com/mwitkow/go-conntrack: cc309e4</li>
<li>github.com/nbutton23/zxcvbn-go: eafdab6</li>
<li>github.com/olekukonko/tablewriter: a0225b3</li>
<li>github.com/quasilyte/go-consistent: c6f3937</li>
<li>github.com/rogpeppe/fastuuid: 6724a57</li>
<li>github.com/ryanuber/go-glob: 256dc44</li>
<li>github.com/sergi/go-diff: v1.0.0</li>
<li>github.com/shirou/gopsutil: c95755e</li>
<li>github.com/shirou/w32: bb4de01</li>
<li>github.com/shurcooL/go-goon: 37c2f52</li>
<li>github.com/shurcooL/go: 9e1955d</li>
<li>github.com/sourcegraph/go-diff: v0.5.1</li>
<li>github.com/tarm/serial: 98f6abe</li>
<li>github.com/tidwall/pretty: v1.0.0</li>
<li>github.com/timakin/bodyclose: 87058b9</li>
<li>github.com/ultraware/funlen: v0.0.2</li>
<li>github.com/urfave/cli: v1.20.0</li>
<li>github.com/valyala/bytebufferpool: v1.0.0</li>
<li>github.com/valyala/fasthttp: v1.2.0</li>
<li>github.com/valyala/quicktemplate: v1.1.1</li>
<li>github.com/valyala/tcplisten: ceec8f9</li>
<li>github.com/vektah/gqlparser: v1.1.2</li>
<li>go.etcd.io/etcd: 3cf2f69</li>
<li>go.mongodb.org/mongo-driver: v1.1.2</li>
<li>go4.org: 417644f</li>
<li>golang.org/x/build: 2835ba2</li>
<li>golang.org/x/perf: 6e6d33e</li>
<li>golang.org/x/xerrors: a985d34</li>
<li>gopkg.in/alecthomas/kingpin.v2: v2.2.6</li>
<li>gopkg.in/cheggaaa/pb.v1: v1.0.25</li>
<li>gopkg.in/resty.v1: v1.12.0</li>
<li>grpc.go4.org: 11d0a25</li>
<li>k8s.io/system-validators: v1.0.4</li>
<li>mvdan.cc/interfacer: c200402</li>
<li>mvdan.cc/lint: adc824a</li>
<li>mvdan.cc/unparam: fbb5962</li>
<li>sourcegraph.com/sqs/pbtypes: d3ebe8f</li>
</ul>
<h4>Changed</h4>
<ul>
<li>github.com/Azure/azure-sdk-for-go: v32.5.0+incompatible → v35.0.0+incompatible</li>
<li>github.com/Microsoft/go-winio: v0.4.11 → v0.4.14</li>
<li>github.com/bazelbuild/bazel-gazelle: c728ce9 → 70208cb</li>
<li>github.com/bazelbuild/buildtools: 80c7f0d → 69366ca</li>
<li>github.com/beorn7/perks: 3a771d9 → v1.0.0</li>
<li>github.com/container-storage-interface/spec: v1.1.0 → v1.2.0</li>
<li>github.com/coredns/corefile-migration: v1.0.2 → v1.0.4</li>
<li>github.com/coreos/etcd: v3.3.17+incompatible → v3.3.10+incompatible</li>
<li>github.com/coreos/go-systemd: 39ca1b0 → 95778df</li>
<li>github.com/docker/go-units: v0.3.3 → v0.4.0</li>
<li>github.com/docker/libnetwork: a9cd636 → f0e46a7</li>
<li>github.com/fatih/color: v1.6.0 → v1.7.0</li>
<li>github.com/ghodss/yaml: c7ce166 → v1.0.0</li>
<li>github.com/go-openapi/analysis: v0.19.2 → v0.19.5</li>
<li>github.com/go-openapi/jsonpointer: v0.19.2 → v0.19.3</li>
<li>github.com/go-openapi/jsonreference: v0.19.2 → v0.19.3</li>
<li>github.com/go-openapi/loads: v0.19.2 → v0.19.4</li>
<li>github.com/go-openapi/runtime: v0.19.0 → v0.19.4</li>
<li>github.com/go-openapi/spec: v0.19.2 → v0.19.3</li>
<li>github.com/go-openapi/strfmt: v0.19.0 → v0.19.3</li>
<li>github.com/go-openapi/swag: v0.19.2 → v0.19.5</li>
<li>github.com/go-openapi/validate: v0.19.2 → v0.19.5</li>
<li>github.com/godbus/dbus: v4.1.0+incompatible → 2ff6f7f</li>
<li>github.com/golang/protobuf: v1.3.1 → v1.3.2</li>
<li>github.com/google/btree: 4030bb1 → v1.0.0</li>
<li>github.com/google/cadvisor: v0.34.0 → v0.35.0</li>
<li>github.com/gregjones/httpcache: 787624d → 9cad4c3</li>
<li>github.com/grpc-ecosystem/go-grpc-middleware: cfaf568 → f849b54</li>
<li>github.com/grpc-ecosystem/grpc-gateway: v1.3.0 → v1.9.5</li>
<li>github.com/heketi/heketi: v9.0.0+incompatible → c2e2a4a</li>
<li>github.com/json-iterator/go: v1.1.7 → v1.1.8</li>
<li>github.com/mailru/easyjson: 94de47d → v0.7.0</li>
<li>github.com/mattn/go-isatty: v0.0.3 → v0.0.9</li>
<li>github.com/mindprince/gonvml: fee913c → 9ebdce4</li>
<li>github.com/mrunalp/fileutils: 4ee1cc9 → 7d4729f</li>
<li>github.com/munnerz/goautoneg: a547fc6 → a7dc8b6</li>
<li>github.com/onsi/ginkgo: v1.8.0 → v1.10.1</li>
<li>github.com/onsi/gomega: v1.5.0 → v1.7.0</li>
<li>github.com/opencontainers/runc: 6cc5158 → v1.0.0-rc9</li>
<li>github.com/opencontainers/selinux: v1.2.2 → 5215b18</li>
<li>github.com/pkg/errors: v0.8.0 → v0.8.1</li>
<li>github.com/prometheus/client_golang: v0.9.2 → v1.0.0</li>
<li>github.com/prometheus/client_model: 5c3871d → fd36f42</li>
<li>github.com/prometheus/common: 4724e92 → v0.4.1</li>
<li>github.com/prometheus/procfs: 1dc9a6c → v0.0.2</li>
<li>github.com/soheilhy/cmux: v0.1.3 → v0.1.4</li>
<li>github.com/spf13/pflag: v1.0.3 → v1.0.5</li>
<li>github.com/stretchr/testify: v1.3.0 → v1.4.0</li>
<li>github.com/syndtr/gocapability: e7cb7fa → d983527</li>
<li>github.com/vishvananda/netlink: b2de5d1 → v1.0.0</li>
<li>github.com/vmware/govmomi: v0.20.1 → v0.20.3</li>
<li>github.com/xiang90/probing: 07dd2e8 → 43a291a</li>
<li>go.uber.org/atomic: 8dc6146 → v1.3.2</li>
<li>go.uber.org/multierr: ddea229 → v1.1.0</li>
<li>go.uber.org/zap: 67bc79d → v1.10.0</li>
<li>golang.org/x/crypto: e84da03 → 60c769a</li>
<li>golang.org/x/lint: 8f45f77 → 959b441</li>
<li>golang.org/x/net: cdfb69a → 13f9640</li>
<li>golang.org/x/oauth2: 9f33145 → 0f29369</li>
<li>golang.org/x/sync: 42b3178 → cd5d95a</li>
<li>golang.org/x/sys: 3b52091 → fde4db3</li>
<li>golang.org/x/text: e6919f6 → v0.3.2</li>
<li>golang.org/x/time: f51c127 → 9d24e82</li>
<li>golang.org/x/tools: 6e04913 → 65e3620</li>
<li>google.golang.org/grpc: v1.23.0 → v1.23.1</li>
<li>gopkg.in/inf.v0: v0.9.0 → v0.9.1</li>
<li>k8s.io/klog: v0.4.0 → v1.0.0</li>
<li>k8s.io/kube-openapi: 743ec37 → 30be4d1</li>
<li>k8s.io/repo-infra: 00fe14e → v0.0.1-alpha.1</li>
<li>k8s.io/utils: 581e001 → e782cd3</li>
<li>sigs.k8s.io/structured-merge-diff: 6149e45 → b1b620d</li>
</ul>
<h4>Removed</h4>
<ul>
<li>github.com/cloudflare/cfssl: 56268a6</li>
<li>github.com/coreos/bbolt: v1.3.3</li>
<li>github.com/coreos/rkt: v1.30.0</li>
<li>github.com/globalsign/mgo: eeefdec</li>
<li>github.com/google/certificate-transparency-go: v1.0.21</li>
<li>github.com/heketi/rest: aa6a652</li>
<li>github.com/heketi/utils: 435bc5b</li>
<li>github.com/pborman/uuid: v1.2.0</li>
</ul>

  </body>
</html>`

const rcReleaseExpectedTOC = `<!-- BEGIN MUNGE: GENERATED_TOC -->

- [v1.16.0-rc.1](#v1160-rc1)`

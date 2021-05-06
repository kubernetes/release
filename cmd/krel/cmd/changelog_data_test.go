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
- github.com/evanphx/json-patch: [v4.2.0+incompatible → v4.9.0+incompatible](https://github.com/evanphx/json-patch/compare/v4.2.0...v4.9.0)
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
<li>github.com/evanphx/json-patch: <a href="https://github.com/evanphx/json-patch/compare/v4.2.0...v4.9.0">v4.2.0+incompatible → v4.9.0+incompatible</a></li>
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
<p>Kubeadm: make the kubeconfig files for the kube-controller-manager and kube-scheduler use the LocalAPIEndpoint instead of the ControlPlaneEndpoint. This makes kubeadm clusters more reseliant to version skew problems during immutable upgrades: <a href="https://kubernetes.io/docs/setup/release/version-skew-policy/#kube-controller-manager-kube-scheduler-and-cloud-controller-manager">https://kubernetes.io/docs/setup/release/version-skew-policy/#kube-controller-manager-kube-scheduler-and-cloud-controller-manager</a> (<a href="https://github.com/kubernetes/kubernetes/pull/94398">#94398</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubeadm: relax the validation of kubeconfig server URLs. Allow the user to define custom kubeconfig server URLs without erroring out during validation of existing kubeconfig files (e.g. when using external CA mode). (<a href="https://github.com/kubernetes/kubernetes/pull/94816">#94816</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubelet: assume that swap is disabled when <code>/proc/swaps</code> does not exist (<a href="https://github.com/kubernetes/kubernetes/pull/93931">#93931</a>, <a href="https://github.com/SataQiu">@SataQiu</a>) [SIG Node]</p>
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
<li>Adds a bootstrapping ClusterRole, ClusterRoleBinding and group for /metrics, /livez/<em>, /readyz/</em>, &amp; /healthz/* endpoints. (<a href="https://github.com/kubernetes/kubernetes/pull/93311">#93311</a>, <a href="https://github.com/logicalhan">@logicalhan</a>) [SIG API Machinery, Auth, Cloud Provider and Instrumentation]</li>
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

- [v1.21.0](#v1210)
  - [Changelog since v1.20.0](#changelog-since-v1200)
- [Release notes for v1.21.0-rc.0](#release-notes-for-v1210-rc0)
- [Changelog since v1.20.0](#changelog-since-v1200-1)
  - [What's New (Major Themes)](#whats-new-major-themes)
    - [Deprecation of PodSecurityPolicy](#deprecation-of-podsecuritypolicy)
    - [Kubernetes API Reference Documentation](#kubernetes-api-reference-documentation)
    - [Kustomize Updates in Kubectl](#kustomize-updates-in-kubectl)
    - [Default Container Labels](#default-container-labels)
    - [Immutable Secrets and ConfigMaps](#immutable-secrets-and-configmaps)
    - [Structured Logging in Kubelet](#structured-logging-in-kubelet)
    - [Storage Capacity Tracking](#storage-capacity-tracking)
    - [Generic Ephemeral Volumes](#generic-ephemeral-volumes)
    - [CSI Service Account Token](#csi-service-account-token)
    - [CSI Health Monitoring](#csi-health-monitoring)
  - [Known Issues](#known-issues)
    - [<code>TopologyAwareHints</code> feature falls back to default behavior](#-feature-falls-back-to-default-behavior)
  - [Urgent Upgrade Notes](#urgent-upgrade-notes)
    - [(No, really, you MUST read this before you upgrade)](#no-really-you-must-read-this-before-you-upgrade)
  - [Changes by Kind](#changes-by-kind)
    - [Deprecation](#deprecation)
    - [API Change](#api-change)
    - [Feature](#feature)
    - [Documentation](#documentation)
    - [Failing Test](#failing-test)
    - [Bug or Regression](#bug-or-regression)
    - [Other (Cleanup or Flake)](#other-cleanup-or-flake)
    - [Uncategorized](#uncategorized)
  - [Dependencies](#dependencies)
    - [Added](#added)
    - [Changed](#changed)
    - [Removed](#removed)
`

const minorReleaseExpectedContent = `# Changelog since v1.20.0

## What's New (Major Themes)

### Deprecation of PodSecurityPolicy

PSP as an admission controller resource is being deprecated. Deployed PodSecurityPolicy's will keep working until version 1.25, their target removal from the codebase. A new feature, with a working title of "PSP replacement policy", is being developed in [KEP-2579](https://features.k8s.io/2579). To learn more, read [PodSecurityPolicy Deprecation: Past, Present, and Future](https://blog.k8s.io/2021/04/06/podsecuritypolicy-deprecation-past-present-and-future/).`

// nolint: misspell
const minorReleaseExpectedHTML = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width" />
    <title>v1.21.0</title>
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
    <h1>v1.21.0</h1>
<h2>Changelog since v1.20.0</h2>
<h1>Release notes for v1.21.0-rc.0</h1>
<p><a href="https://docs.k8s.io/docs/home">Documentation</a></p>
<h1>Changelog since v1.20.0</h1>
<h2>What's New (Major Themes)</h2>
<h3>Deprecation of PodSecurityPolicy</h3>
<p>PSP as an admission controller resource is being deprecated. Deployed PodSecurityPolicy's will keep working until version 1.25, their target removal from the codebase. A new feature, with a working title of &quot;PSP replacement policy&quot;, is being developed in <a href="https://features.k8s.io/2579">KEP-2579</a>. To learn more, read <a href="https://blog.k8s.io/2021/04/06/podsecuritypolicy-deprecation-past-present-and-future/">PodSecurityPolicy Deprecation: Past, Present, and Future</a>.</p>
<h3>Kubernetes API Reference Documentation</h3>
<p>The API reference is now generated with <a href="https://github.com/kubernetes-sigs/reference-docs/tree/c96658d89fb21037b7d00d27e6dbbe6b32375837/gen-resourcesdocs"><code>gen-resourcesdocs</code></a> and it is moving to <a href="https://docs.k8s.io/reference/kubernetes-api/">Kubernetes API</a></p>
<h3>Kustomize Updates in Kubectl</h3>
<p><a href="https://github.com/kubernetes-sigs/kustomize">Kustomize</a> version in kubectl had a jump from v2.0.3 to <a href="https://github.com/kubernetes/kubernetes/pull/98946">v4.0.5</a>. Kustomize is now treated as a library and future updates will be less sporadic.</p>
<h3>Default Container Labels</h3>
<p>Pod with multiple containers can use <code>kubectl.kubernetes.io/default-container</code> label to have a container preselected for kubectl commands. More can be read in <a href="https://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/2227-kubectl-default-container/README.md">KEP-2227</a>.</p>
<h3>Immutable Secrets and ConfigMaps</h3>
<p>Immutable Secrets and ConfigMaps graduates to GA. This feature allows users to specify that the contents of a particular Secret or ConfigMap is immutable for its object lifetime. For such instances, Kubelet will not watch/poll for changes and therefore reducing apiserver load.</p>
<h3>Structured Logging in Kubelet</h3>
<p>Kubelet has adopted structured logging, thanks to community effort in accomplishing this within the release timeline. Structured logging in the project remains an ongoing effort -- for folks interested in participating, <a href="https://groups.google.com/g/kubernetes-dev/c/y4WIw-ntUR8">keep an eye / chime in to the mailing list discussion</a>.</p>
<h3>Storage Capacity Tracking</h3>
<p>Traditionally, the Kubernetes scheduler was based on the assumptions that additional persistent storage is available everywhere in the cluster and has infinite capacity. Topology constraints addressed the first point, but up to now pod scheduling was still done without considering that the remaining storage capacity may not be enough to start a new pod. <a href="https://docs.k8s.io/concepts/storage/storage-capacity/">Storage capacity tracking</a> addresses that by adding an API for a CSI driver to report storage capacity and uses that information in the Kubernetes scheduler when choosing a node for a pod. This feature serves as a stepping stone for supporting dynamic provisioning for local volumes and other volume types that are more capacity constrained.</p>
<h3>Generic Ephemeral Volumes</h3>
<p><a href="https://docs.k8s.io/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes">Generic ephermeral volumes</a> feature allows any existing storage driver that supports dynamic provisioning to be used as an ephemeral volume with the volume’s lifecycle bound to the Pod. It can be used to provide scratch storage that is different from the root disk, for example persistent memory, or a separate local disk on that node. All StorageClass parameters for volume provisioning are supported. All features supported with PersistentVolumeClaims are supported, such as storage capacity tracking, snapshots and restore, and volume resizing.</p>
<h3>CSI Service Account Token</h3>
<p>CSI Service Account Token feature moves to Beta in 1.21. This feature improves the security posture and allows CSI drivers to receive pods' <a href="https://github.com/kubernetes/enhancements/blob/master/keps/sig-auth/1205-bound-service-account-tokens/README.md">bound service account tokens</a>. This feature also provides a knob to re-publish volumes so that short-lived volumes can be refreshed.</p>
<h3>CSI Health Monitoring</h3>
<p>The CSI health monitoring feature is being released as a second Alpha in Kubernetes 1.21. This feature enables CSI Drivers to share abnormal volume conditions from the underlying storage systems with Kubernetes so that they can be reported as events on PVCs or Pods. This feature serves as a stepping stone towards programmatic detection and resolution of individual volume health issues by Kubernetes.</p>
<h2>Known Issues</h2>
<h3><code>TopologyAwareHints</code> feature falls back to default behavior</h3>
<p>The feature gate currently falls back to the default behavior in most cases. Enabling the feature gate will add hints to <code>EndpointSlices</code>, but functional differences are only observed in non-dual stack kube-proxy implementation. <a href="https://github.com/kubernetes/kubernetes/pull/100804">The fix will be available in coming releases</a>.</p>
<h2>Urgent Upgrade Notes</h2>
<h3>(No, really, you MUST read this before you upgrade)</h3>
<ul>
<li>Kube-proxy's IPVS proxy mode no longer sets the net.ipv4.conf.all.route_localnet sysctl parameter. Nodes upgrading will have net.ipv4.conf.all.route_localnet set to 1 but new nodes will inherit the system default (usually 0). If you relied on any behavior requiring net.ipv4.conf.all.route_localnet, you must set ensure it is enabled as kube-proxy will no longer set it automatically. This change helps to further mitigate CVE-2020-8558. (<a href="https://github.com/kubernetes/kubernetes/pull/92938">#92938</a>, <a href="https://github.com/lbernail">@lbernail</a>) [SIG Network and Release]</li>
<li>Kubeadm: during &quot;init&quot; an empty cgroupDriver value in the KubeletConfiguration is now always set to &quot;systemd&quot; unless the user is explicit about it. This requires existing machine setups to configure the container runtime to use the &quot;systemd&quot; driver. Documentation on this topic can be found here: <a href="https://kubernetes.io/docs/setup/production-environment/container-runtimes/">https://kubernetes.io/docs/setup/production-environment/container-runtimes/</a>. When upgrading existing clusters / nodes using &quot;kubeadm upgrade&quot; the old cgroupDriver value is preserved, but in 1.22 this change will also apply to &quot;upgrade&quot;. For more information on migrating to the &quot;systemd&quot; driver or remaining on the &quot;cgroupfs&quot; driver see: <a href="https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/configure-cgroup-driver/">https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/configure-cgroup-driver/</a>. (<a href="https://github.com/kubernetes/kubernetes/pull/99471">#99471</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Newly provisioned PVs by EBS plugin will no longer use the deprecated &quot;failure-domain.beta.kubernetes.io/zone&quot; and &quot;failure-domain.beta.kubernetes.io/region&quot; labels. It will use &quot;topology.kubernetes.io/zone&quot; and &quot;topology.kubernetes.io/region&quot; labels instead. (<a href="https://github.com/kubernetes/kubernetes/pull/99130">#99130</a>, <a href="https://github.com/ayberk">@ayberk</a>) [SIG Cloud Provider, Storage and Testing]</li>
<li>Newly provisioned PVs by OpenStack Cinder plugin will no longer use the deprecated &quot;failure-domain.beta.kubernetes.io/zone&quot; and &quot;failure-domain.beta.kubernetes.io/region&quot; labels. It will use &quot;topology.kubernetes.io/zone&quot; and &quot;topology.kubernetes.io/region&quot; labels instead. (<a href="https://github.com/kubernetes/kubernetes/pull/99719">#99719</a>, <a href="https://github.com/jsafrane">@jsafrane</a>) [SIG Cloud Provider and Storage]</li>
<li>Newly provisioned PVs by gce-pd will no longer have the beta FailureDomain label. gce-pd volume plugin will start to have GA topology label instead. (<a href="https://github.com/kubernetes/kubernetes/pull/98700">#98700</a>, <a href="https://github.com/Jiawei0227">@Jiawei0227</a>) [SIG Cloud Provider, Storage and Testing]</li>
<li>OpenStack Cinder CSI migration is on by default, Clinder CSI driver must be installed on clusters on OpenStack for Cinder volumes to work. (<a href="https://github.com/kubernetes/kubernetes/pull/98538">#98538</a>, <a href="https://github.com/dims">@dims</a>) [SIG Storage]</li>
<li>Remove alpha <code>CSIMigrationXXComplete</code> flag and add alpha <code>InTreePluginXXUnregister</code> flag. Deprecate <code>CSIMigrationvSphereComplete</code> flag and it will be removed in v1.22. (<a href="https://github.com/kubernetes/kubernetes/pull/98243">#98243</a>, <a href="https://github.com/Jiawei0227">@Jiawei0227</a>)</li>
<li>Remove storage metrics <code>storage_operation_errors_total</code>, since we already have <code>storage_operation_status_count</code>.And add new field <code>status</code> for <code>storage_operation_duration_seconds</code>, so that we can know about all status storage operation latency. (<a href="https://github.com/kubernetes/kubernetes/pull/98332">#98332</a>, <a href="https://github.com/JornShen">@JornShen</a>) [SIG Instrumentation and Storage]</li>
<li>The metric <code>storage_operation_errors_total</code> is not removed, but is marked deprecated, and the metric <code>storage_operation_status_count</code> is marked deprecated. In both cases the <code>storage_operation_duration_seconds</code> metric can be used to recover equivalent counts (using <code>status=fail-unknown</code> in the case of <code>storage_operations_errors_total</code>). (<a href="https://github.com/kubernetes/kubernetes/pull/99045">#99045</a>, <a href="https://github.com/mattcary">@mattcary</a>)</li>
<li><code>ServiceNodeExclusion</code>, <code>NodeDisruptionExclusion</code> and <code>LegacyNodeRoleBehavior</code> features have been promoted to GA. <code>ServiceNodeExclusion</code> and <code>NodeDisruptionExclusion</code> are now unconditionally enabled, while <code>LegacyNodeRoleBehavior</code> is unconditionally disabled. To prevent control plane nodes from being added to load balancers automatically, upgrade users need to add &quot;node.kubernetes.io/exclude-from-external-load-balancers&quot; label to control plane nodes. (<a href="https://github.com/kubernetes/kubernetes/pull/97543">#97543</a>, <a href="https://github.com/pacoxu">@pacoxu</a>)</li>
</ul>
<h2>Changes by Kind</h2>
<h3>Deprecation</h3>
<ul>
<li>
<p>Aborting the drain command in a list of nodes will be deprecated. The new behavior will make the drain command go through all nodes even if one or more nodes failed during the drain. For now, users can try such experience by enabling --ignore-errors flag. (<a href="https://github.com/kubernetes/kubernetes/pull/98203">#98203</a>, <a href="https://github.com/yuzhiquan">@yuzhiquan</a>)</p>
</li>
<li>
<p>Delete deprecated <code>service.beta.kubernetes.io/azure-load-balancer-mixed-protocols</code> mixed procotol annotation in favor of the MixedProtocolLBService feature (<a href="https://github.com/kubernetes/kubernetes/pull/97096">#97096</a>, <a href="https://github.com/nilo19">@nilo19</a>) [SIG Cloud Provider]</p>
</li>
<li>
<p>Deprecate the <code>topologyKeys</code> field in Service. This capability will be replaced with upcoming work around Topology Aware Subsetting and Service Internal Traffic Policy. (<a href="https://github.com/kubernetes/kubernetes/pull/96736">#96736</a>, <a href="https://github.com/andrewsykim">@andrewsykim</a>) [SIG Apps]</p>
</li>
<li>
<p>Kube-proxy: remove deprecated --cleanup-ipvs flag of kube-proxy, and make --cleanup flag always to flush IPVS (<a href="https://github.com/kubernetes/kubernetes/pull/97336">#97336</a>, <a href="https://github.com/maaoBit">@maaoBit</a>) [SIG Network]</p>
</li>
<li>
<p>Kubeadm: deprecated command &quot;alpha selfhosting pivot&quot; is now removed. (<a href="https://github.com/kubernetes/kubernetes/pull/97627">#97627</a>, <a href="https://github.com/knight42">@knight42</a>)</p>
</li>
<li>
<p>Kubeadm: graduate the command <code>kubeadm alpha kubeconfig user</code> to <code>kubeadm kubeconfig user</code>. The <code>kubeadm alpha kubeconfig user</code> command is deprecated now. (<a href="https://github.com/kubernetes/kubernetes/pull/97583">#97583</a>, <a href="https://github.com/knight42">@knight42</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubeadm: the &quot;kubeadm alpha certs&quot; command is removed now, please use &quot;kubeadm certs&quot; instead. (<a href="https://github.com/kubernetes/kubernetes/pull/97706">#97706</a>, <a href="https://github.com/knight42">@knight42</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubeadm: the deprecated kube-dns is no longer supported as an option. If &quot;ClusterConfiguration.dns.type&quot; is set to &quot;kube-dns&quot; kubeadm will now throw an error. (<a href="https://github.com/kubernetes/kubernetes/pull/99646">#99646</a>, <a href="https://github.com/rajansandeep">@rajansandeep</a>) [SIG Cluster Lifecycle]</p>
</li>
<li>
<p>Kubectl: The deprecated <code>kubectl alpha debug</code> command is removed. Use <code>kubectl debug</code> instead. (<a href="https://github.com/kubernetes/kubernetes/pull/98111">#98111</a>, <a href="https://github.com/pandaamanda">@pandaamanda</a>) [SIG CLI]</p>
</li>
<li>
<p>Official support to build kubernetes with docker-machine / remote docker is removed. This change does not affect building kubernetes with docker locally. (<a href="https://github.com/kubernetes/kubernetes/pull/97935">#97935</a>, <a href="https://github.com/adeniyistephen">@adeniyistephen</a>) [SIG Release and Testing]</p>
</li>
<li>
<p>Remove deprecated <code>--generator, --replicas, --service-generator, --service-overrides, --schedule</code> from <code>kubectl run</code>
Deprecate <code>--serviceaccount, --hostport, --requests, --limits</code> in <code>kubectl run</code> (<a href="https://github.com/kubernetes/kubernetes/pull/99732">#99732</a>, <a href="https://github.com/soltysh">@soltysh</a>)</p>
</li>
<li>
<p>Remove the deprecated metrics &quot;scheduling_algorithm_preemption_evaluation_seconds&quot; and &quot;binding_duration_seconds&quot;, suggest to use &quot;scheduler_framework_extension_point_duration_seconds&quot; instead. (<a href="https://github.com/kubernetes/kubernetes/pull/96447">#96447</a>, <a href="https://github.com/chendave">@chendave</a>) [SIG Cluster Lifecycle, Instrumentation, Scheduling and Testing]</p>
</li>
<li>
<p>Removing experimental windows container hyper-v support with Docker (<a href="https://github.com/kubernetes/kubernetes/pull/97141">#97141</a>, <a href="https://github.com/wawa0210">@wawa0210</a>) [SIG Node and Windows]</p>
</li>
<li>
<p>Rename metrics <code>etcd_object_counts</code> to <code>apiserver_storage_object_counts</code> and mark it as stable. The original <code>etcd_object_counts</code> metrics name is marked as &quot;Deprecated&quot; and will be removed in the future. (<a href="https://github.com/kubernetes/kubernetes/pull/99785">#99785</a>, <a href="https://github.com/erain">@erain</a>) [SIG API Machinery, Instrumentation and Testing]</p>
</li>
<li>
<p>The GA TokenRequest and TokenRequestProjection feature gates have been removed and are unconditionally enabled. Remove explicit use of those feature gates in CLI invocations. (<a href="https://github.com/kubernetes/kubernetes/pull/97148">#97148</a>, <a href="https://github.com/wawa0210">@wawa0210</a>) [SIG Node]</p>
</li>
<li>
<p>The PodSecurityPolicy API is deprecated in 1.21, and will no longer be served starting in 1.25. (<a href="https://github.com/kubernetes/kubernetes/pull/97171">#97171</a>, <a href="https://github.com/deads2k">@deads2k</a>) [SIG Auth and CLI]</p>
</li>
<li>
<p>The <code>batch/v2alpha1</code> CronJob type definitions and clients are deprecated and removed. (<a href="https://github.com/kubernetes/kubernetes/pull/96987">#96987</a>, <a href="https://github.com/soltysh">@soltysh</a>) [SIG API Machinery, Apps, CLI and Testing]</p>
</li>
<li>
<p>The <code>export</code> query parameter (inconsistently supported by API resources and deprecated in v1.14) is fully removed.  Requests setting this query parameter will now receive a 400 status response. (<a href="https://github.com/kubernetes/kubernetes/pull/98312">#98312</a>, <a href="https://github.com/deads2k">@deads2k</a>) [SIG API Machinery, Auth and Testing]</p>
</li>
<li>
<p><code>audit.k8s.io/v1beta1</code> and <code>audit.k8s.io/v1alpha1</code> audit policy configuration and audit events are deprecated in favor of <code>audit.k8s.io/v1</code>, available since v1.13. kube-apiserver invocations that specify alpha or beta policy configurations with <code>--audit-policy-file</code>, or explicitly request alpha or beta audit events with <code>--audit-log-version</code> / <code>--audit-webhook-version</code> must update to use <code>audit.k8s.io/v1</code> and accept <code>audit.k8s.io/v1</code> events prior to v1.24. (<a href="https://github.com/kubernetes/kubernetes/pull/98858">#98858</a>, <a href="https://github.com/carlory">@carlory</a>) [SIG Auth]</p>
</li>
<li>
<p><code>discovery.k8s.io/v1beta1</code> EndpointSlices are deprecated in favor of <code>discovery.k8s.io/v1</code>, and will no longer be served in Kubernetes v1.25. (<a href="https://github.com/kubernetes/kubernetes/pull/100472">#100472</a>, <a href="https://github.com/liggitt">@liggitt</a>)</p>
</li>
<li>
<p><code>diskformat</code> storage class parameter for in-tree vSphere volume plugin is deprecated as of v1.21 release. Please consider updating storageclass and remove <code>diskformat</code> parameter. vSphere CSI Driver does not support diskformat storageclass parameter.</p>
<p>vSphere releases less than 67u3 are deprecated as of v1.21. Please consider upgrading vSphere to 67u3 or above. vSphere CSI Driver requires minimum vSphere 67u3.</p>
<p>VM Hardware version less than 15 is deprecated as of v1.21. Please consider upgrading the Node VM Hardware version to 15 or above. vSphere CSI Driver recommends Node VM's Hardware version set to at least vmx-15.</p>
<p>Multi vCenter support is deprecated as of v1.21. If you have a Kubernetes cluster spanning across multiple vCenter servers, please consider moving all k8s nodes to a single vCenter Server. vSphere CSI Driver does not support Kubernetes deployment spanning across multiple vCenter servers.</p>
<p>Support for these deprecations will be available till Kubernetes v1.24. (<a href="https://github.com/kubernetes/kubernetes/pull/98546">#98546</a>, <a href="https://github.com/divyenpatel">@divyenpatel</a>)</p>
</li>
</ul>
<h3>API Change</h3>
<ul>
<li>
<ol>
<li>PodAffinityTerm includes a namespaceSelector field to allow selecting eligible namespaces based on their labels.</li>
<li>A new CrossNamespacePodAffinity quota scope API that allows restricting which namespaces allowed to use PodAffinityTerm with corss-namespace reference via namespaceSelector or namespaces fields. (<a href="https://github.com/kubernetes/kubernetes/pull/98582">#98582</a>, <a href="https://github.com/ahg-g">@ahg-g</a>) [SIG API Machinery, Apps, Auth and Testing]</li>
</ol>
</li>
<li>Add Probe-level terminationGracePeriodSeconds field (<a href="https://github.com/kubernetes/kubernetes/pull/99375">#99375</a>, <a href="https://github.com/ehashman">@ehashman</a>) [SIG API Machinery, Apps, Node and Testing]</li>
<li>Added <code>.spec.completionMode</code> field to Job, with accepted values <code>NonIndexed</code> (default) and <code>Indexed</code>. This is an alpha field and is only honored by servers with the <code>IndexedJob</code> feature gate enabled. (<a href="https://github.com/kubernetes/kubernetes/pull/98441">#98441</a>, <a href="https://github.com/alculquicondor">@alculquicondor</a>) [SIG Apps and CLI]</li>
<li>Adds support for endPort field in NetworkPolicy (<a href="https://github.com/kubernetes/kubernetes/pull/97058">#97058</a>, <a href="https://github.com/rikatz">@rikatz</a>) [SIG Apps and Network]</li>
<li>CSIServiceAccountToken graduates to Beta and enabled by default. (<a href="https://github.com/kubernetes/kubernetes/pull/99298">#99298</a>, <a href="https://github.com/zshihang">@zshihang</a>)</li>
<li>Cluster admins can now turn off <code>/debug/pprof</code> and <code>/debug/flags/v</code> endpoint in kubelet by setting <code>enableProfilingHandler</code> and <code>enableDebugFlagsHandler</code> to <code>false</code> in the Kubelet configuration file. Options <code>enableProfilingHandler</code> and <code>enableDebugFlagsHandler</code> can be set to <code>true</code> only when <code>enableDebuggingHandlers</code> is also set to <code>true</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/98458">#98458</a>, <a href="https://github.com/SaranBalaji90">@SaranBalaji90</a>)</li>
<li>DaemonSets accept a MaxSurge integer or percent on their rolling update strategy that will launch the updated pod on nodes and wait for those pods to go ready before marking the old out-of-date pods as deleted. This allows workloads to avoid downtime during upgrades when deployed using DaemonSets. This feature is alpha and is behind the DaemonSetUpdateSurge feature gate. (<a href="https://github.com/kubernetes/kubernetes/pull/96441">#96441</a>, <a href="https://github.com/smarterclayton">@smarterclayton</a>) [SIG Apps and Testing]</li>
<li>Enable SPDY pings to keep connections alive, so that <code>kubectl exec</code> and <code>kubectl portforward</code> won't be interrupted. (<a href="https://github.com/kubernetes/kubernetes/pull/97083">#97083</a>, <a href="https://github.com/knight42">@knight42</a>) [SIG API Machinery and CLI]</li>
<li>FieldManager no longer owns fields that get reset before the object is persisted (e.g. &quot;status wiping&quot;). (<a href="https://github.com/kubernetes/kubernetes/pull/99661">#99661</a>, <a href="https://github.com/kevindelgado">@kevindelgado</a>) [SIG API Machinery, Auth and Testing]</li>
<li>Fixes server-side apply for APIService resources. (<a href="https://github.com/kubernetes/kubernetes/pull/98576">#98576</a>, <a href="https://github.com/kevindelgado">@kevindelgado</a>)</li>
<li>Generic ephemeral volumes are beta. (<a href="https://github.com/kubernetes/kubernetes/pull/99643">#99643</a>, <a href="https://github.com/pohly">@pohly</a>) [SIG API Machinery, Apps, Auth, CLI, Node, Storage and Testing]</li>
<li>Hugepages request values are limited to integer multiples of the page size. (<a href="https://github.com/kubernetes/kubernetes/pull/98515">#98515</a>, <a href="https://github.com/lala123912">@lala123912</a>) [SIG Apps]</li>
<li>Implement the GetAvailableResources in the podresources API. (<a href="https://github.com/kubernetes/kubernetes/pull/95734">#95734</a>, <a href="https://github.com/fromanirh">@fromanirh</a>) [SIG Instrumentation, Node and Testing]</li>
<li>IngressClass resource can now reference a resource in a specific namespace
for implementation-specific configuration (previously only Cluster-level resources were allowed).
This feature can be enabled using the IngressClassNamespacedParams feature gate. (<a href="https://github.com/kubernetes/kubernetes/pull/99275">#99275</a>, <a href="https://github.com/hbagdi">@hbagdi</a>)</li>
<li>Jobs API has a new <code>.spec.suspend</code> field that can be used to suspend and resume Jobs. This is an alpha field which is only honored by servers with the <code>SuspendJob</code> feature gate enabled. (<a href="https://github.com/kubernetes/kubernetes/pull/98727">#98727</a>, <a href="https://github.com/adtac">@adtac</a>)</li>
<li>Kubelet Graceful Node Shutdown feature graduates to Beta and enabled by default. (<a href="https://github.com/kubernetes/kubernetes/pull/99735">#99735</a>, <a href="https://github.com/bobbypage">@bobbypage</a>)</li>
<li>Kubernetes is now built using go1.15.7 (<a href="https://github.com/kubernetes/kubernetes/pull/98363">#98363</a>, <a href="https://github.com/cpanato">@cpanato</a>) [SIG Cloud Provider, Instrumentation, Node, Release and Testing]</li>
<li>Namespace API objects now have a <code>kubernetes.io/metadata.name</code> label matching their metadata.name field to allow selecting any namespace by its name using a label selector. (<a href="https://github.com/kubernetes/kubernetes/pull/96968">#96968</a>, <a href="https://github.com/jayunit100">@jayunit100</a>) [SIG API Machinery, Apps, Cloud Provider, Storage and Testing]</li>
<li>One new field &quot;InternalTrafficPolicy&quot; in Service is added.
It specifies if the cluster internal traffic should be routed to all endpoints or node-local endpoints only.
&quot;Cluster&quot; routes internal traffic to a Service to all endpoints.
&quot;Local&quot; routes traffic to node-local endpoints only, and traffic is dropped if no node-local endpoints are ready.
The default value is &quot;Cluster&quot;. (<a href="https://github.com/kubernetes/kubernetes/pull/96600">#96600</a>, <a href="https://github.com/maplain">@maplain</a>) [SIG API Machinery, Apps and Network]</li>
<li>PodDisruptionBudget API objects can now contain conditions in status. (<a href="https://github.com/kubernetes/kubernetes/pull/98127">#98127</a>, <a href="https://github.com/mortent">@mortent</a>) [SIG API Machinery, Apps, Auth, CLI, Cloud Provider, Cluster Lifecycle and Instrumentation]</li>
<li>PodSecurityPolicy only stores &quot;generic&quot; as allowed volume type if the GenericEphemeralVolume feature gate is enabled (<a href="https://github.com/kubernetes/kubernetes/pull/98918">#98918</a>, <a href="https://github.com/pohly">@pohly</a>) [SIG Auth and Security]</li>
<li>Promote CronJobs to batch/v1 (<a href="https://github.com/kubernetes/kubernetes/pull/99423">#99423</a>, <a href="https://github.com/soltysh">@soltysh</a>) [SIG API Machinery, Apps, CLI and Testing]</li>
<li>Promote Immutable Secrets/ConfigMaps feature to Stable. This allows to set <code>immutable</code> field in Secret or ConfigMap object to mark their contents as immutable. (<a href="https://github.com/kubernetes/kubernetes/pull/97615">#97615</a>, <a href="https://github.com/wojtek-t">@wojtek-t</a>) [SIG Apps, Architecture, Node and Testing]</li>
<li>Remove support for building Kubernetes with bazel. (<a href="https://github.com/kubernetes/kubernetes/pull/99561">#99561</a>, <a href="https://github.com/BenTheElder">@BenTheElder</a>) [SIG API Machinery, Apps, Architecture, Auth, Autoscaling, CLI, Cloud Provider, Cluster Lifecycle, Instrumentation, Network, Node, Release, Scalability, Scheduling, Storage, Testing and Windows]</li>
<li>Scheduler extender filter interface now can report unresolvable failed nodes in the new field <code>FailedAndUnresolvableNodes</code> of  <code>ExtenderFilterResult</code> struct. Nodes in this map will be skipped in the preemption phase. (<a href="https://github.com/kubernetes/kubernetes/pull/92866">#92866</a>, <a href="https://github.com/cofyc">@cofyc</a>) [SIG Scheduling]</li>
<li>Services can specify loadBalancerClass to use a custom load balancer (<a href="https://github.com/kubernetes/kubernetes/pull/98277">#98277</a>, <a href="https://github.com/XudongLiuHarold">@XudongLiuHarold</a>)</li>
<li>Storage capacity tracking (= the CSIStorageCapacity feature) graduates to Beta and enabled by default, storage.k8s.io/v1alpha1/VolumeAttachment and storage.k8s.io/v1alpha1/CSIStorageCapacity objects are deprecated (<a href="https://github.com/kubernetes/kubernetes/pull/99641">#99641</a>, <a href="https://github.com/pohly">@pohly</a>)</li>
<li>Support for Indexed Job: a Job that is considered completed when Pods associated to indexes from 0 to (.spec.completions-1) have succeeded. (<a href="https://github.com/kubernetes/kubernetes/pull/98812">#98812</a>, <a href="https://github.com/alculquicondor">@alculquicondor</a>) [SIG Apps and CLI]</li>
<li>The BoundServiceAccountTokenVolume feature has been promoted to beta, and enabled by default.
<ul>
<li>This changes the tokens provided to containers at <code>/var/run/secrets/kubernetes.io/serviceaccount/token</code> to be time-limited, auto-refreshed, and invalidated when the containing pod is deleted.</li>
<li>Clients should reload the token from disk periodically (once per minute is recommended) to ensure they continue to use a valid token. <code>k8s.io/client-go</code> version v11.0.0+ and v0.15.0+ reload tokens automatically.</li>
<li>By default, injected tokens are given an extended lifetime so they remain valid even after a new refreshed token is provided. The metric <code>serviceaccount_stale_tokens_total</code> can be used to monitor for workloads that are depending on the extended lifetime and are continuing to use tokens even after a refreshed token is provided to the container. If that metric indicates no existing workloads are depending on extended lifetimes, injected token lifetime can be shortened to 1 hour by starting <code>kube-apiserver</code> with <code>--service-account-extend-token-expiration=false</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/95667">#95667</a>, <a href="https://github.com/zshihang">@zshihang</a>) [SIG API Machinery, Auth, Cluster Lifecycle and Testing]</li>
</ul>
</li>
<li>The EndpointSlice Controllers are now GA. The <code>EndpointSliceController</code> will not populate the <code>deprecatedTopology</code> field and will only provide topology information through the <code>zone</code> and <code>nodeName</code> fields. (<a href="https://github.com/kubernetes/kubernetes/pull/99870">#99870</a>, <a href="https://github.com/swetharepakula">@swetharepakula</a>)</li>
<li>The Endpoints controller will now set the <code>endpoints.kubernetes.io/over-capacity</code> annotation to &quot;warning&quot; when an Endpoints resource contains more than 1000 addresses. In a future release, the controller will truncate Endpoints that exceed this limit. The EndpointSlice API can be used to support significantly larger number of addresses. (<a href="https://github.com/kubernetes/kubernetes/pull/99975">#99975</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG Apps and Network]</li>
<li>The PodDisruptionBudget API has been promoted to policy/v1 with no schema changes. The only functional change is that an empty selector (<code>{}</code>) written to a policy/v1 PodDisruptionBudget now selects all pods in the namespace. The behavior of the policy/v1beta1 API remains unchanged. The policy/v1beta1 PodDisruptionBudget API is deprecated and will no longer be served in 1.25+. (<a href="https://github.com/kubernetes/kubernetes/pull/99290">#99290</a>, <a href="https://github.com/mortent">@mortent</a>) [SIG API Machinery, Apps, Auth, Autoscaling, CLI, Cloud Provider, Cluster Lifecycle, Instrumentation, Scheduling and Testing]</li>
<li>The <code>EndpointSlice</code> API is now GA. The <code>EndpointSlice</code> topology field has been removed from the GA API and will be replaced by a new per Endpoint Zone field. If the topology field was previously used, it will be converted into an annotation in the v1 Resource. The <code>discovery.k8s.io/v1alpha1</code> API is removed. (<a href="https://github.com/kubernetes/kubernetes/pull/99662">#99662</a>, <a href="https://github.com/swetharepakula">@swetharepakula</a>)</li>
<li>The <code>controller.kubernetes.io/pod-deletion-cost</code> annotation can be set to offer a hint on the cost of deleting a <code>Pod</code> compared to other pods belonging to the same ReplicaSet. Pods with lower deletion cost are deleted first. This is an alpha feature. (<a href="https://github.com/kubernetes/kubernetes/pull/99163">#99163</a>, <a href="https://github.com/ahg-g">@ahg-g</a>)</li>
<li>The kube-apiserver now resets <code>managedFields</code> that got corrupted by a mutating admission controller. (<a href="https://github.com/kubernetes/kubernetes/pull/98074">#98074</a>, <a href="https://github.com/kwiesmueller">@kwiesmueller</a>)</li>
<li>Topology Aware Hints are now available in alpha and can be enabled with the <code>TopologyAwareHints</code> feature gate. (<a href="https://github.com/kubernetes/kubernetes/pull/99522">#99522</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG API Machinery, Apps, Auth, Instrumentation, Network and Testing]</li>
<li>Users might specify the <code>kubectl.kubernetes.io/default-exec-container</code> annotation in a Pod to preselect container for kubectl commands. (<a href="https://github.com/kubernetes/kubernetes/pull/97099">#97099</a>, <a href="https://github.com/pacoxu">@pacoxu</a>) [SIG CLI]</li>
</ul>
<h3>Feature</h3>
<ul>
<li>A client-go metric, rest_client_exec_plugin_call_total, has been added to track total calls to client-go credential plugins. (<a href="https://github.com/kubernetes/kubernetes/pull/98892">#98892</a>, <a href="https://github.com/ankeesler">@ankeesler</a>) [SIG API Machinery, Auth, Cluster Lifecycle and Instrumentation]</li>
<li>A new histogram metric to track the time it took to delete a job by the <code>TTLAfterFinished</code> controller (<a href="https://github.com/kubernetes/kubernetes/pull/98676">#98676</a>, <a href="https://github.com/ahg-g">@ahg-g</a>)</li>
<li>AWS cloud provider supports auto-discovering subnets without any <code>kubernetes.io/cluster/&lt;clusterName&gt;</code> tags. It also supports additional service annotation <code>service.beta.kubernetes.io/aws-load-balancer-subnets</code> to manually configure the subnets. (<a href="https://github.com/kubernetes/kubernetes/pull/97431">#97431</a>, <a href="https://github.com/kishorj">@kishorj</a>)</li>
<li>Aborting the drain command in a list of nodes will be deprecated. The new behavior will make the drain command go through all nodes even if one or more nodes failed during the drain. For now, users can try such experience by enabling --ignore-errors flag. (<a href="https://github.com/kubernetes/kubernetes/pull/98203">#98203</a>, <a href="https://github.com/yuzhiquan">@yuzhiquan</a>)</li>
<li>Add --permit-address-sharing flag to <code>kube-apiserver</code> to listen with <code>SO_REUSEADDR</code>. While allowing to listen on wildcard IPs like 0.0.0.0 and specific IPs in parallel, it avoids waiting for the kernel to release socket in <code>TIME_WAIT</code> state, and hence, considerably reducing <code>kube-apiserver</code> restart times under certain conditions. (<a href="https://github.com/kubernetes/kubernetes/pull/93861">#93861</a>, <a href="https://github.com/sttts">@sttts</a>)</li>
<li>Add <code>csi_operations_seconds</code> metric on kubelet that exposes CSI operations duration and status for node CSI operations. (<a href="https://github.com/kubernetes/kubernetes/pull/98979">#98979</a>, <a href="https://github.com/Jiawei0227">@Jiawei0227</a>) [SIG Instrumentation and Storage]</li>
<li>Add <code>migrated</code> field into <code>storage_operation_duration_seconds</code> metric (<a href="https://github.com/kubernetes/kubernetes/pull/99050">#99050</a>, <a href="https://github.com/Jiawei0227">@Jiawei0227</a>) [SIG Apps, Instrumentation and Storage]</li>
<li>Add flag --lease-reuse-duration-seconds for kube-apiserver to config etcd lease reuse duration. (<a href="https://github.com/kubernetes/kubernetes/pull/97009">#97009</a>, <a href="https://github.com/lingsamuel">@lingsamuel</a>) [SIG API Machinery and Scalability]</li>
<li>Add metric etcd_lease_object_counts for kube-apiserver to observe max objects attached to a single etcd lease. (<a href="https://github.com/kubernetes/kubernetes/pull/97480">#97480</a>, <a href="https://github.com/lingsamuel">@lingsamuel</a>) [SIG API Machinery, Instrumentation and Scalability]</li>
<li>Add support to generate client-side binaries for new darwin/arm64 platform (<a href="https://github.com/kubernetes/kubernetes/pull/97743">#97743</a>, <a href="https://github.com/dims">@dims</a>) [SIG Release and Testing]</li>
<li>Added <code>ephemeral_volume_controller_create[_failures]_total</code> counters to kube-controller-manager metrics (<a href="https://github.com/kubernetes/kubernetes/pull/99115">#99115</a>, <a href="https://github.com/pohly">@pohly</a>) [SIG API Machinery, Apps, Cluster Lifecycle, Instrumentation and Storage]</li>
<li>Added support for installing <code>arm64</code> node artifacts. (<a href="https://github.com/kubernetes/kubernetes/pull/99242">#99242</a>, <a href="https://github.com/liu-cong">@liu-cong</a>)</li>
<li>Adds alpha feature <code>VolumeCapacityPriority</code> which makes the scheduler prioritize nodes based on the best matching size of statically provisioned PVs across multiple topologies. (<a href="https://github.com/kubernetes/kubernetes/pull/96347">#96347</a>, <a href="https://github.com/cofyc">@cofyc</a>) [SIG Apps, Network, Scheduling, Storage and Testing]</li>
<li>Adds the ability to pass --strict-transport-security-directives to the kube-apiserver to set the HSTS header appropriately.  Be sure you understand the consequences to browsers before setting this field. (<a href="https://github.com/kubernetes/kubernetes/pull/96502">#96502</a>, <a href="https://github.com/249043822">@249043822</a>) [SIG Auth]</li>
<li>Adds two new metrics to cronjobs, a histogram to track the time difference when a job is created and the expected time when it should be created, as well as a gauge for the missed schedules of a cronjob (<a href="https://github.com/kubernetes/kubernetes/pull/99341">#99341</a>, <a href="https://github.com/alaypatel07">@alaypatel07</a>)</li>
<li>Alpha implementation of Kubectl Command Headers: SIG CLI KEP 859 enabled when KUBECTL_COMMAND_HEADERS environment variable set on the client command line. (<a href="https://github.com/kubernetes/kubernetes/pull/98952">#98952</a>, <a href="https://github.com/seans3">@seans3</a>)</li>
<li>Base-images: Update to debian-iptables:buster-v1.4.0
<ul>
<li>Uses iptables 1.8.5</li>
<li>base-images: Update to debian-base:buster-v1.3.0</li>
<li>cluster/images/etcd: Build etcd:3.4.13-2 image
<ul>
<li>Uses debian-base:buster-v1.3.0 (<a href="https://github.com/kubernetes/kubernetes/pull/98401">#98401</a>, <a href="https://github.com/pacoxu">@pacoxu</a>) [SIG Testing]</li>
</ul>
</li>
</ul>
</li>
<li>CRIContainerLogRotation graduates to GA and unconditionally enabled. (<a href="https://github.com/kubernetes/kubernetes/pull/99651">#99651</a>, <a href="https://github.com/umohnani8">@umohnani8</a>)</li>
<li>Component owner can configure the allowlist of metric label with flag '--allow-metric-labels'. (<a href="https://github.com/kubernetes/kubernetes/pull/99385">#99385</a>, <a href="https://github.com/YoyinZyc">@YoyinZyc</a>) [SIG API Machinery, CLI, Cloud Provider, Cluster Lifecycle, Instrumentation and Release]</li>
<li>Component owner can configure the allowlist of metric label with flag '--allow-metric-labels'. (<a href="https://github.com/kubernetes/kubernetes/pull/99738">#99738</a>, <a href="https://github.com/YoyinZyc">@YoyinZyc</a>) [SIG API Machinery, Cluster Lifecycle and Instrumentation]</li>
<li>EmptyDir memory backed volumes are sized as the the minimum of pod allocatable memory on a host and an optional explicit user provided value. (<a href="https://github.com/kubernetes/kubernetes/pull/100319">#100319</a>, <a href="https://github.com/derekwaynecarr">@derekwaynecarr</a>) [SIG Node]</li>
<li>Enables Kubelet to check volume condition and log events to corresponding pods. (<a href="https://github.com/kubernetes/kubernetes/pull/99284">#99284</a>, <a href="https://github.com/fengzixu">@fengzixu</a>) [SIG Apps, Instrumentation, Node and Storage]</li>
<li>EndpointSliceNodeName graduates to GA and thus will be unconditionally enabled -- NodeName will always be available in the v1beta1 API. (<a href="https://github.com/kubernetes/kubernetes/pull/99746">#99746</a>, <a href="https://github.com/swetharepakula">@swetharepakula</a>)</li>
<li>Export <code>NewDebuggingRoundTripper</code> function and <code>DebugLevel</code> options in the k8s.io/client-go/transport package. (<a href="https://github.com/kubernetes/kubernetes/pull/98324">#98324</a>, <a href="https://github.com/atosatto">@atosatto</a>)</li>
<li>Kube-proxy iptables: new metric sync_proxy_rules_iptables_total that exposes the number of rules programmed per table in each iteration (<a href="https://github.com/kubernetes/kubernetes/pull/99653">#99653</a>, <a href="https://github.com/aojea">@aojea</a>) [SIG Instrumentation and Network]</li>
<li>Kube-scheduler now logs plugin scoring summaries at --v=4 (<a href="https://github.com/kubernetes/kubernetes/pull/99411">#99411</a>, <a href="https://github.com/damemi">@damemi</a>) [SIG Scheduling]</li>
<li>Kubeadm now includes CoreDNS v1.8.0. (<a href="https://github.com/kubernetes/kubernetes/pull/96429">#96429</a>, <a href="https://github.com/rajansandeep">@rajansandeep</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: IPv6DualStack feature gate graduates to Beta and enabled by default (<a href="https://github.com/kubernetes/kubernetes/pull/99294">#99294</a>, <a href="https://github.com/pacoxu">@pacoxu</a>)</li>
<li>Kubeadm: a warning to user as ipv6 site-local is deprecated (<a href="https://github.com/kubernetes/kubernetes/pull/99574">#99574</a>, <a href="https://github.com/pacoxu">@pacoxu</a>) [SIG Cluster Lifecycle and Network]</li>
<li>Kubeadm: add support for certificate chain validation. When using kubeadm in external CA mode, this allows an intermediate CA to be used to sign the certificates. The intermediate CA certificate must be appended to each signed certificate for this to work correctly. (<a href="https://github.com/kubernetes/kubernetes/pull/97266">#97266</a>, <a href="https://github.com/robbiemcmichael">@robbiemcmichael</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: amend the node kernel validation to treat CGROUP_PIDS, FAIR_GROUP_SCHED as required and CFS_BANDWIDTH, CGROUP_HUGETLB as optional (<a href="https://github.com/kubernetes/kubernetes/pull/96378">#96378</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle and Node]</li>
<li>Kubeadm: apply the &quot;node.kubernetes.io/exclude-from-external-load-balancers&quot; label on control plane nodes during &quot;init&quot;, &quot;join&quot; and &quot;upgrade&quot; to preserve backwards compatibility with the lagacy LB mode where nodes labeled as &quot;master&quot; where excluded. To opt-out you can remove the label from a node. See #97543 and the linked KEP for more details. (<a href="https://github.com/kubernetes/kubernetes/pull/98269">#98269</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: if the user has customized their image repository via the kubeadm configuration, pass the custom pause image repository and tag to the kubelet via --pod-infra-container-image not only for Docker but for all container runtimes. This flag tells the kubelet that it should not garbage collect the image. (<a href="https://github.com/kubernetes/kubernetes/pull/99476">#99476</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: perform pre-flight validation on host/node name upon <code>kubeadm init</code> and <code>kubeadm join</code>, showing warnings on non-compliant names (<a href="https://github.com/kubernetes/kubernetes/pull/99194">#99194</a>, <a href="https://github.com/pacoxu">@pacoxu</a>)</li>
<li>Kubectl version changed to write a warning message to stderr if the client and server version difference exceeds the supported version skew of +/-1 minor version. (<a href="https://github.com/kubernetes/kubernetes/pull/98250">#98250</a>, <a href="https://github.com/brianpursley">@brianpursley</a>) [SIG CLI]</li>
<li>Kubectl: Add <code>--use-protocol-buffers</code> flag to kubectl top pods and nodes. (<a href="https://github.com/kubernetes/kubernetes/pull/96655">#96655</a>, <a href="https://github.com/serathius">@serathius</a>)</li>
<li>Kubectl: <code>kubectl get</code> will omit managed fields by default now. Users could set <code>--show-managed-fields</code> to true to show managedFields when the output format is either <code>json</code> or <code>yaml</code>. (<a href="https://github.com/kubernetes/kubernetes/pull/96878">#96878</a>, <a href="https://github.com/knight42">@knight42</a>) [SIG CLI and Testing]</li>
<li>Kubectl: a Pod can be preselected as default container using <code>kubectl.kubernetes.io/default-container</code> annotation (<a href="https://github.com/kubernetes/kubernetes/pull/99833">#99833</a>, <a href="https://github.com/mengjiao-liu">@mengjiao-liu</a>)</li>
<li>Kubectl: add bash-completion for comma separated list on <code>kubectl get</code> (<a href="https://github.com/kubernetes/kubernetes/pull/98301">#98301</a>, <a href="https://github.com/phil9909">@phil9909</a>)</li>
<li>Kubernetes is now built using go1.15.8 (<a href="https://github.com/kubernetes/kubernetes/pull/98834">#98834</a>, <a href="https://github.com/cpanato">@cpanato</a>) [SIG Cloud Provider, Instrumentation, Release and Testing]</li>
<li>Kubernetes is now built with Golang 1.16 (<a href="https://github.com/kubernetes/kubernetes/pull/98572">#98572</a>, <a href="https://github.com/justaugustus">@justaugustus</a>) [SIG API Machinery, Auth, CLI, Cloud Provider, Cluster Lifecycle, Instrumentation, Node, Release and Testing]</li>
<li>Kubernetes is now built with Golang 1.16.1 (<a href="https://github.com/kubernetes/kubernetes/pull/100106">#100106</a>, <a href="https://github.com/justaugustus">@justaugustus</a>) [SIG Cloud Provider, Instrumentation, Release and Testing]</li>
<li>Metrics can now be disabled explicitly via a command line flag (i.e. '--disabled-metrics=metric1,metric2') (<a href="https://github.com/kubernetes/kubernetes/pull/99217">#99217</a>, <a href="https://github.com/logicalhan">@logicalhan</a>)</li>
<li>New admission controller <code>DenyServiceExternalIPs</code> is available.  Clusters which do not <em>need</em> the Service <code>externalIPs</code> feature should enable this controller and be more secure. (<a href="https://github.com/kubernetes/kubernetes/pull/97395">#97395</a>, <a href="https://github.com/thockin">@thockin</a>)</li>
<li>Overall, enable the feature of <code>PreferNominatedNode</code> will  improve the performance of scheduling where preemption might frequently happen, but in theory, enable the feature of <code>PreferNominatedNode</code>, the pod might not be scheduled to the best candidate node in the cluster. (<a href="https://github.com/kubernetes/kubernetes/pull/93179">#93179</a>, <a href="https://github.com/chendave">@chendave</a>) [SIG Scheduling and Testing]</li>
<li>Persistent Volumes formatted with the btrfs filesystem will now automatically resize when expanded. (<a href="https://github.com/kubernetes/kubernetes/pull/99361">#99361</a>, <a href="https://github.com/Novex">@Novex</a>) [SIG Storage]</li>
<li>Port the devicemanager to Windows node to allow device plugins like directx (<a href="https://github.com/kubernetes/kubernetes/pull/93285">#93285</a>, <a href="https://github.com/aarnaud">@aarnaud</a>) [SIG Node, Testing and Windows]</li>
<li>Removes cAdvisor JSON metrics (/stats/container, /stats/<!-- raw HTML omitted -->/<!-- raw HTML omitted -->, /stats/<!-- raw HTML omitted -->/<!-- raw HTML omitted -->/<!-- raw HTML omitted -->/<!-- raw HTML omitted -->) from the kubelet. (<a href="https://github.com/kubernetes/kubernetes/pull/99236">#99236</a>, <a href="https://github.com/pacoxu">@pacoxu</a>)</li>
<li>Rename metrics <code>etcd_object_counts</code> to <code>apiserver_storage_object_counts</code> and mark it as stable. The original <code>etcd_object_counts</code> metrics name is marked as &quot;Deprecated&quot; and will be removed in the future. (<a href="https://github.com/kubernetes/kubernetes/pull/99785">#99785</a>, <a href="https://github.com/erain">@erain</a>) [SIG API Machinery, Instrumentation and Testing]</li>
<li>Sysctls graduates to General Availability and thus unconditionally enabled. (<a href="https://github.com/kubernetes/kubernetes/pull/99158">#99158</a>, <a href="https://github.com/wgahnagl">@wgahnagl</a>)</li>
<li>The Kubernetes pause image manifest list now contains an image for Windows Server 20H2. (<a href="https://github.com/kubernetes/kubernetes/pull/97322">#97322</a>, <a href="https://github.com/claudiubelu">@claudiubelu</a>) [SIG Windows]</li>
<li>The NodeAffinity plugin implements the PreFilter extension, offering enhanced performance for Filter. (<a href="https://github.com/kubernetes/kubernetes/pull/99213">#99213</a>, <a href="https://github.com/AliceZhang2016">@AliceZhang2016</a>) [SIG Scheduling]</li>
<li>The <code>CronJobControllerV2</code> feature flag graduates to Beta and set to be enabled by default. (<a href="https://github.com/kubernetes/kubernetes/pull/98878">#98878</a>, <a href="https://github.com/soltysh">@soltysh</a>)</li>
<li>The <code>EndpointSlice</code> mirroring controller mirrors endpoints annotations and labels to the generated endpoint slices, it also ensures that updates on any of these fields are mirrored.
The well-known annotation <code>endpoints.kubernetes.io/last-change-trigger-time</code> is skipped and not mirrored. (<a href="https://github.com/kubernetes/kubernetes/pull/98116">#98116</a>, <a href="https://github.com/aojea">@aojea</a>)</li>
<li>The <code>RunAsGroup</code> feature has been promoted to GA in this release. (<a href="https://github.com/kubernetes/kubernetes/pull/94641">#94641</a>, <a href="https://github.com/krmayankk">@krmayankk</a>) [SIG Auth and Node]</li>
<li>The <code>ServiceAccountIssuerDiscovery</code> feature has graduated to GA, and is unconditionally enabled. The <code>ServiceAccountIssuerDiscovery</code> feature-gate will be removed in 1.22. (<a href="https://github.com/kubernetes/kubernetes/pull/98553">#98553</a>, <a href="https://github.com/mtaufen">@mtaufen</a>) [SIG API Machinery, Auth and Testing]</li>
<li>The <code>TTLAfterFinished</code> feature flag is now beta and enabled by default (<a href="https://github.com/kubernetes/kubernetes/pull/98678">#98678</a>, <a href="https://github.com/ahg-g">@ahg-g</a>)</li>
<li>The apimachinery util/net function used to detect the bind address <code>ResolveBindAddress()</code> takes into consideration global IP addresses on loopback interfaces when 1) the host has default routes, or 2) there are no global IPs on those interfaces in order to support more complex network scenarios like BGP Unnumbered RFC 5549 (<a href="https://github.com/kubernetes/kubernetes/pull/95790">#95790</a>, <a href="https://github.com/aojea">@aojea</a>) [SIG Network]</li>
<li>The feature gate <code>RootCAConfigMap</code> graduated to GA in v1.21 and therefore will be unconditionally enabled. This flag will be removed in v1.22 release. (<a href="https://github.com/kubernetes/kubernetes/pull/98033">#98033</a>, <a href="https://github.com/zshihang">@zshihang</a>)</li>
<li>The pause image upgraded to <code>v3.4.1</code> in kubelet and kubeadm for both Linux and Windows. (<a href="https://github.com/kubernetes/kubernetes/pull/98205">#98205</a>, <a href="https://github.com/pacoxu">@pacoxu</a>)</li>
<li>Update pause container to run as pseudo user and group <code>65535:65535</code>. This implies the release of version 3.5 of the container images. (<a href="https://github.com/kubernetes/kubernetes/pull/97963">#97963</a>, <a href="https://github.com/saschagrunert">@saschagrunert</a>) [SIG CLI, Cloud Provider, Cluster Lifecycle, Node, Release, Security and Testing]</li>
<li>Update the latest validated version of Docker to 20.10 (<a href="https://github.com/kubernetes/kubernetes/pull/98977">#98977</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG CLI, Cluster Lifecycle and Node]</li>
<li>Upgrade node local dns to 1.17.0 for better IPv6 support (<a href="https://github.com/kubernetes/kubernetes/pull/99749">#99749</a>, <a href="https://github.com/pacoxu">@pacoxu</a>) [SIG Cloud Provider and Network]</li>
<li>Upgrades <code>IPv6Dualstack</code> to <code>Beta</code> and turns it on by default. New clusters or existing clusters are not be affected until an actor starts adding secondary Pods and service CIDRS CLI flags as described here: <a href="https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/563-dual-stack">IPv4/IPv6 Dual-stack</a> (<a href="https://github.com/kubernetes/kubernetes/pull/98969">#98969</a>, <a href="https://github.com/khenidak">@khenidak</a>)</li>
<li>Users might specify the <code>kubectl.kubernetes.io/default-container</code> annotation in a Pod to preselect container for kubectl commands. (<a href="https://github.com/kubernetes/kubernetes/pull/99581">#99581</a>, <a href="https://github.com/mengjiao-liu">@mengjiao-liu</a>) [SIG CLI]</li>
<li>When downscaling ReplicaSets, ready and creation timestamps are compared in a logarithmic scale. (<a href="https://github.com/kubernetes/kubernetes/pull/99212">#99212</a>, <a href="https://github.com/damemi">@damemi</a>) [SIG Apps and Testing]</li>
<li>When the kubelet is watching a ConfigMap or Secret purely in the context of setting environment variables
for containers, only hold that watch for a defined duration before cancelling it. This change reduces the CPU
and memory usage of the kube-apiserver in large clusters. (<a href="https://github.com/kubernetes/kubernetes/pull/99393">#99393</a>, <a href="https://github.com/chenyw1990">@chenyw1990</a>) [SIG API Machinery, Node and Testing]</li>
<li>WindowsEndpointSliceProxying feature gate has graduated to beta and is enabled by default. This means kube-proxy will  read from EndpointSlices instead of Endpoints on Windows by default. (<a href="https://github.com/kubernetes/kubernetes/pull/99794">#99794</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG Network]</li>
<li><code>kubectl wait</code> ensures that observedGeneration &gt;= generation to prevent stale state reporting. An example scenario can be found on CRD updates. (<a href="https://github.com/kubernetes/kubernetes/pull/97408">#97408</a>, <a href="https://github.com/KnicKnic">@KnicKnic</a>)</li>
</ul>
<h3>Documentation</h3>
<ul>
<li>Azure file migration graduates to beta, with CSIMigrationAzureFile flag off by default
as it requires installation of AzureFile CSI Driver. Users should enable CSIMigration and
CSIMigrationAzureFile features and install the <a href="https://github.com/kubernetes-sigs/azurefile-csi-driver">AzureFile CSI Driver</a>
to avoid disruption to existing Pod and PVC objects at that time. Azure File CSI driver does not support using same persistent
volume with different fsgroups. When CSI migration is enabled for azurefile driver, such case is not supported.
(there is a case we support where volume is mounted with 0777 and then it readable/writable by everyone) (<a href="https://github.com/kubernetes/kubernetes/pull/96293">#96293</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>)</li>
<li>Official support to build kubernetes with docker-machine / remote docker is removed. This change does not affect building kubernetes with docker locally. (<a href="https://github.com/kubernetes/kubernetes/pull/97935">#97935</a>, <a href="https://github.com/adeniyistephen">@adeniyistephen</a>) [SIG Release and Testing]</li>
<li>Set kubelet option <code>--volume-stats-agg-period</code> to negative value to disable volume calculations. (<a href="https://github.com/kubernetes/kubernetes/pull/96675">#96675</a>, <a href="https://github.com/pacoxu">@pacoxu</a>) [SIG Node]</li>
</ul>
<h3>Failing Test</h3>
<ul>
<li>Escape the special characters like <code>[</code>, <code>]</code> and <code> </code> that exist in vsphere windows path (<a href="https://github.com/kubernetes/kubernetes/pull/98830">#98830</a>, <a href="https://github.com/liyanhui1228">@liyanhui1228</a>) [SIG Storage and Windows]</li>
<li>Kube-proxy: fix a bug on UDP <code>NodePort</code> Services where stale connection tracking entries may blackhole the traffic directed to the <code>NodePort</code> (<a href="https://github.com/kubernetes/kubernetes/pull/98305">#98305</a>, <a href="https://github.com/aojea">@aojea</a>)</li>
<li>Kubelet: fixes a bug in the HostPort dockershim implementation that caused the conformance test &quot;HostPort validates that there is no conflict between pods with same hostPort but different hostIP and protocol&quot; to fail. (<a href="https://github.com/kubernetes/kubernetes/pull/98755">#98755</a>, <a href="https://github.com/aojea">@aojea</a>) [SIG Cloud Provider, Network and Node]</li>
</ul>
<h3>Bug or Regression</h3>
<ul>
<li>AcceleratorStats will be available in the Summary API of kubelet when cri_stats_provider is used. (<a href="https://github.com/kubernetes/kubernetes/pull/96873">#96873</a>, <a href="https://github.com/ruiwen-zhao">@ruiwen-zhao</a>) [SIG Node]</li>
<li>All data is no longer automatically deleted when a failure is detected during creation of the volume data file on a CSI volume. Now only the data file and volume path is removed. (<a href="https://github.com/kubernetes/kubernetes/pull/96021">#96021</a>, <a href="https://github.com/huffmanca">@huffmanca</a>)</li>
<li>Clean ReplicaSet by revision instead of creation timestamp in deployment controller (<a href="https://github.com/kubernetes/kubernetes/pull/97407">#97407</a>, <a href="https://github.com/waynepeking348">@waynepeking348</a>) [SIG Apps]</li>
<li>Cleanup subnet in frontend IP configs to prevent huge subnet request bodies in some scenarios. (<a href="https://github.com/kubernetes/kubernetes/pull/98133">#98133</a>, <a href="https://github.com/nilo19">@nilo19</a>) [SIG Cloud Provider]</li>
<li>Client-go exec credential plugins will pass stdin only when interactive terminal is detected on stdin. This fixes a bug where previously it was checking if <strong>stdout</strong> is an interactive terminal. (<a href="https://github.com/kubernetes/kubernetes/pull/99654">#99654</a>, <a href="https://github.com/ankeesler">@ankeesler</a>)</li>
<li>Cloud-controller-manager: routes controller should not depend on --allocate-node-cidrs (<a href="https://github.com/kubernetes/kubernetes/pull/97029">#97029</a>, <a href="https://github.com/andrewsykim">@andrewsykim</a>) [SIG Cloud Provider and Testing]</li>
<li>Cluster Autoscaler version bump to v1.20.0 (<a href="https://github.com/kubernetes/kubernetes/pull/97011">#97011</a>, <a href="https://github.com/towca">@towca</a>)</li>
<li>Creating a PVC with DataSource should fail for non-CSI plugins. (<a href="https://github.com/kubernetes/kubernetes/pull/97086">#97086</a>, <a href="https://github.com/xing-yang">@xing-yang</a>) [SIG Apps and Storage]</li>
<li>EndpointSlice controller is now less likely to emit FailedToUpdateEndpointSlices events. (<a href="https://github.com/kubernetes/kubernetes/pull/99345">#99345</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG Apps and Network]</li>
<li>EndpointSlice controllers are less likely to create duplicate EndpointSlices. (<a href="https://github.com/kubernetes/kubernetes/pull/100103">#100103</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG Apps and Network]</li>
<li>EndpointSliceMirroring controller is now less likely to emit FailedToUpdateEndpointSlices events. (<a href="https://github.com/kubernetes/kubernetes/pull/99756">#99756</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG Apps and Network]</li>
<li>Ensure all vSphere nodes are are tracked by volume attach-detach controller (<a href="https://github.com/kubernetes/kubernetes/pull/96689">#96689</a>, <a href="https://github.com/gnufied">@gnufied</a>)</li>
<li>Ensure empty string annotations are copied over in rollbacks. (<a href="https://github.com/kubernetes/kubernetes/pull/94858">#94858</a>, <a href="https://github.com/waynepeking348">@waynepeking348</a>)</li>
<li>Ensure only one LoadBalancer rule is created when HA mode is enabled (<a href="https://github.com/kubernetes/kubernetes/pull/99825">#99825</a>, <a href="https://github.com/feiskyer">@feiskyer</a>) [SIG Cloud Provider]</li>
<li>Ensure that client-go's EventBroadcaster is safe (non-racy) during shutdown. (<a href="https://github.com/kubernetes/kubernetes/pull/95664">#95664</a>, <a href="https://github.com/DirectXMan12">@DirectXMan12</a>) [SIG API Machinery]</li>
<li>Explicitly pass <code>KUBE_BUILD_CONFORMANCE=y</code> in <code>package-tarballs</code> to reenable building the conformance tarballs. (<a href="https://github.com/kubernetes/kubernetes/pull/100571">#100571</a>, <a href="https://github.com/puerco">@puerco</a>)</li>
<li>Fix Azure file migration e2e test failure when CSIMigration is turned on. (<a href="https://github.com/kubernetes/kubernetes/pull/97877">#97877</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>)</li>
<li>Fix CSI-migrated inline EBS volumes failing to mount if their volumeID is prefixed by aws:// (<a href="https://github.com/kubernetes/kubernetes/pull/96821">#96821</a>, <a href="https://github.com/wongma7">@wongma7</a>) [SIG Storage]</li>
<li>Fix CVE-2020-8555 for Gluster client connections. (<a href="https://github.com/kubernetes/kubernetes/pull/97922">#97922</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG Storage]</li>
<li>Fix NPE in ephemeral storage eviction (<a href="https://github.com/kubernetes/kubernetes/pull/98261">#98261</a>, <a href="https://github.com/wzshiming">@wzshiming</a>) [SIG Node]</li>
<li>Fix PermissionDenied issue on SMB mount for Windows (<a href="https://github.com/kubernetes/kubernetes/pull/99550">#99550</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>)</li>
<li>Fix bug that would let the Horizontal Pod Autoscaler scale down despite at least one metric being unavailable/invalid (<a href="https://github.com/kubernetes/kubernetes/pull/99514">#99514</a>, <a href="https://github.com/mikkeloscar">@mikkeloscar</a>) [SIG Apps and Autoscaling]</li>
<li>Fix cgroup handling for systemd with cgroup v2 (<a href="https://github.com/kubernetes/kubernetes/pull/98365">#98365</a>, <a href="https://github.com/odinuge">@odinuge</a>) [SIG Node]</li>
<li>Fix counting error in service/nodeport/loadbalancer quota check (<a href="https://github.com/kubernetes/kubernetes/pull/97451">#97451</a>, <a href="https://github.com/pacoxu">@pacoxu</a>) [SIG API Machinery, Network and Testing]</li>
<li>Fix errors when accessing Windows container stats for Dockershim (<a href="https://github.com/kubernetes/kubernetes/pull/98510">#98510</a>, <a href="https://github.com/jsturtevant">@jsturtevant</a>) [SIG Node and Windows]</li>
<li>Fix kube-proxy container image architecture for non amd64 images. (<a href="https://github.com/kubernetes/kubernetes/pull/98526">#98526</a>, <a href="https://github.com/saschagrunert">@saschagrunert</a>)</li>
<li>Fix missing cadvisor machine metrics. (<a href="https://github.com/kubernetes/kubernetes/pull/97006">#97006</a>, <a href="https://github.com/lingsamuel">@lingsamuel</a>) [SIG Node]</li>
<li>Fix nil VMSS name when setting service to auto mode (<a href="https://github.com/kubernetes/kubernetes/pull/97366">#97366</a>, <a href="https://github.com/nilo19">@nilo19</a>) [SIG Cloud Provider]</li>
<li>Fix privileged config of Pod Sandbox which was previously ignored. (<a href="https://github.com/kubernetes/kubernetes/pull/96877">#96877</a>, <a href="https://github.com/xeniumlee">@xeniumlee</a>)</li>
<li>Fix the panic when kubelet registers if a node object already exists with no Status.Capacity or Status.Allocatable (<a href="https://github.com/kubernetes/kubernetes/pull/95269">#95269</a>, <a href="https://github.com/SataQiu">@SataQiu</a>) [SIG Node]</li>
<li>Fix the regression with the slow pods termination. Before this fix pods may take an additional time to terminate - up to one minute. Reversing the change that ensured that CNI resources cleaned up when the pod is removed on API server. (<a href="https://github.com/kubernetes/kubernetes/pull/97980">#97980</a>, <a href="https://github.com/SergeyKanzhelev">@SergeyKanzhelev</a>) [SIG Node]</li>
<li>Fix to recover CSI volumes from certain dangling attachments (<a href="https://github.com/kubernetes/kubernetes/pull/96617">#96617</a>, <a href="https://github.com/yuga711">@yuga711</a>) [SIG Apps and Storage]</li>
<li>Fix: azure file latency issue for metadata-heavy workloads (<a href="https://github.com/kubernetes/kubernetes/pull/97082">#97082</a>, <a href="https://github.com/andyzhangx">@andyzhangx</a>) [SIG Cloud Provider and Storage]</li>
<li>Fixed Cinder volume IDs on OpenStack Train (<a href="https://github.com/kubernetes/kubernetes/pull/96673">#96673</a>, <a href="https://github.com/jsafrane">@jsafrane</a>) [SIG Cloud Provider]</li>
<li>Fixed FibreChannel volume plugin corrupting filesystems on detach of multipath volumes. (<a href="https://github.com/kubernetes/kubernetes/pull/97013">#97013</a>, <a href="https://github.com/jsafrane">@jsafrane</a>) [SIG Storage]</li>
<li>Fixed a bug in kubelet that will saturate CPU utilization after containerd got restarted. (<a href="https://github.com/kubernetes/kubernetes/pull/97174">#97174</a>, <a href="https://github.com/hanlins">@hanlins</a>) [SIG Node]</li>
<li>Fixed a bug that causes smaller number of conntrack-max being used under CPU static policy. (#99225, @xh4n3) (<a href="https://github.com/kubernetes/kubernetes/pull/99613">#99613</a>, <a href="https://github.com/xh4n3">@xh4n3</a>) [SIG Network]</li>
<li>Fixed a bug that on k8s nodes, when the policy of INPUT chain in filter table is not ACCEPT, healthcheck nodeport would not work.
Added iptables rules to allow healthcheck nodeport traffic. (<a href="https://github.com/kubernetes/kubernetes/pull/97824">#97824</a>, <a href="https://github.com/hanlins">@hanlins</a>) [SIG Network]</li>
<li>Fixed a bug that the kubelet cannot start on BtrfS. (<a href="https://github.com/kubernetes/kubernetes/pull/98042">#98042</a>, <a href="https://github.com/gjkim42">@gjkim42</a>) [SIG Node]</li>
<li>Fixed a race condition on API server startup ensuring previously created webhook configurations are effective before the first write request is admitted. (<a href="https://github.com/kubernetes/kubernetes/pull/95783">#95783</a>, <a href="https://github.com/roycaihw">@roycaihw</a>) [SIG API Machinery]</li>
<li>Fixed an issue with garbage collection failing to clean up namespaced children of an object also referenced incorrectly by cluster-scoped children (<a href="https://github.com/kubernetes/kubernetes/pull/98068">#98068</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG API Machinery and Apps]</li>
<li>Fixed authentication_duration_seconds metric scope. Previously, it included whole apiserver request duration which yields inaccurate results. (<a href="https://github.com/kubernetes/kubernetes/pull/99944">#99944</a>, <a href="https://github.com/marseel">@marseel</a>)</li>
<li>Fixed bug in CPUManager with race on container map access (<a href="https://github.com/kubernetes/kubernetes/pull/97427">#97427</a>, <a href="https://github.com/klueska">@klueska</a>) [SIG Node]</li>
<li>Fixed bug that caused cAdvisor to incorrectly detect single-socket multi-NUMA topology. (<a href="https://github.com/kubernetes/kubernetes/pull/99315">#99315</a>, <a href="https://github.com/iwankgb">@iwankgb</a>) [SIG Node]</li>
<li>Fixed cleanup of block devices when /var/lib/kubelet is a symlink. (<a href="https://github.com/kubernetes/kubernetes/pull/96889">#96889</a>, <a href="https://github.com/jsafrane">@jsafrane</a>) [SIG Storage]</li>
<li>Fixed no effect namespace when exposing deployment with --dry-run=client. (<a href="https://github.com/kubernetes/kubernetes/pull/97492">#97492</a>, <a href="https://github.com/masap">@masap</a>) [SIG CLI]</li>
<li>Fixed provisioning of Cinder volumes migrated to CSI when StorageClass with AllowedTopologies was used. (<a href="https://github.com/kubernetes/kubernetes/pull/98311">#98311</a>, <a href="https://github.com/jsafrane">@jsafrane</a>) [SIG Storage]</li>
<li>Fixes a bug of identifying the correct containerd process. (<a href="https://github.com/kubernetes/kubernetes/pull/97888">#97888</a>, <a href="https://github.com/pacoxu">@pacoxu</a>)</li>
<li>Fixes add-on manager leader election to use leases instead of endpoints, similar to what kube-controller-manager does in 1.20 (<a href="https://github.com/kubernetes/kubernetes/pull/98968">#98968</a>, <a href="https://github.com/liggitt">@liggitt</a>)</li>
<li>Fixes connection errors when using <code>--volume-host-cidr-denylist</code> or <code>--volume-host-allow-local-loopback</code> (<a href="https://github.com/kubernetes/kubernetes/pull/98436">#98436</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG Network and Storage]</li>
<li>Fixes problem where invalid selector on <code>PodDisruptionBudget</code> leads to a nil pointer dereference that causes the Controller manager to crash loop. (<a href="https://github.com/kubernetes/kubernetes/pull/98750">#98750</a>, <a href="https://github.com/mortent">@mortent</a>)</li>
<li>Fixes spurious errors about IPv6 in <code>kube-proxy</code> logs on nodes with IPv6 disabled. (<a href="https://github.com/kubernetes/kubernetes/pull/99127">#99127</a>, <a href="https://github.com/danwinship">@danwinship</a>)</li>
<li>Fixing a bug where a failed node may not have the NoExecute taint set correctly (<a href="https://github.com/kubernetes/kubernetes/pull/96876">#96876</a>, <a href="https://github.com/howieyuen">@howieyuen</a>) [SIG Apps and Node]</li>
<li>GCE Internal LoadBalancer sync loop will now release the ILB IP address upon sync failure. An error in ILB forwarding rule creation will no longer leak IP addresses. (<a href="https://github.com/kubernetes/kubernetes/pull/97740">#97740</a>, <a href="https://github.com/prameshj">@prameshj</a>) [SIG Cloud Provider and Network]</li>
<li>Ignore update pod with no new images in alwaysPullImages admission controller (<a href="https://github.com/kubernetes/kubernetes/pull/96668">#96668</a>, <a href="https://github.com/pacoxu">@pacoxu</a>) [SIG Apps, Auth and Node]</li>
<li>Improve speed of vSphere PV provisioning and reduce number of API calls (<a href="https://github.com/kubernetes/kubernetes/pull/100054">#100054</a>, <a href="https://github.com/gnufied">@gnufied</a>) [SIG Cloud Provider and Storage]</li>
<li>KUBECTL_EXTERNAL_DIFF now accepts equal sign for additional parameters. (<a href="https://github.com/kubernetes/kubernetes/pull/98158">#98158</a>, <a href="https://github.com/dougsland">@dougsland</a>) [SIG CLI]</li>
<li>Kube-apiserver: an update of a pod with a generic ephemeral volume dropped that volume if the feature had been disabled since creating the pod with such a volume (<a href="https://github.com/kubernetes/kubernetes/pull/99446">#99446</a>, <a href="https://github.com/pohly">@pohly</a>) [SIG Apps, Node and Storage]</li>
<li>Kube-proxy: remove deprecated --cleanup-ipvs flag of kube-proxy, and make --cleanup flag always to flush IPVS (<a href="https://github.com/kubernetes/kubernetes/pull/97336">#97336</a>, <a href="https://github.com/maaoBit">@maaoBit</a>) [SIG Network]</li>
<li>Kubeadm installs etcd v3.4.13 when creating cluster v1.19 (<a href="https://github.com/kubernetes/kubernetes/pull/97244">#97244</a>, <a href="https://github.com/pacoxu">@pacoxu</a>)</li>
<li>Kubeadm: Fixes a kubeadm upgrade bug that could cause a custom CoreDNS configuration to be replaced with the default. (<a href="https://github.com/kubernetes/kubernetes/pull/97016">#97016</a>, <a href="https://github.com/rajansandeep">@rajansandeep</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: Some text in the <code>kubeadm upgrade plan</code> output has changed. If you have scripts or other automation that parses this output, please review these changes and update your scripts to account for the new output. (<a href="https://github.com/kubernetes/kubernetes/pull/98728">#98728</a>, <a href="https://github.com/stmcginnis">@stmcginnis</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: fix a bug in the host memory detection code on 32bit Linux platforms (<a href="https://github.com/kubernetes/kubernetes/pull/97403">#97403</a>, <a href="https://github.com/abelbarrera15">@abelbarrera15</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: fix a bug where &quot;kubeadm join&quot; would not properly handle missing names for existing etcd members. (<a href="https://github.com/kubernetes/kubernetes/pull/97372">#97372</a>, <a href="https://github.com/ihgann">@ihgann</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: fix a bug where &quot;kubeadm upgrade&quot; commands can fail if CoreDNS v1.8.0 is installed. (<a href="https://github.com/kubernetes/kubernetes/pull/97919">#97919</a>, <a href="https://github.com/neolit123">@neolit123</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: fix a bug where external credentials in an existing admin.conf prevented the CA certificate to be written in the cluster-info ConfigMap. (<a href="https://github.com/kubernetes/kubernetes/pull/98882">#98882</a>, <a href="https://github.com/kvaps">@kvaps</a>) [SIG Cluster Lifecycle]</li>
<li>Kubeadm: get k8s CI version markers from k8s infra bucket (<a href="https://github.com/kubernetes/kubernetes/pull/98836">#98836</a>, <a href="https://github.com/hasheddan">@hasheddan</a>) [SIG Cluster Lifecycle and Release]</li>
<li>Kubeadm: skip validating pod subnet against node-cidr-mask when allocate-node-cidrs is set to be false (<a href="https://github.com/kubernetes/kubernetes/pull/98984">#98984</a>, <a href="https://github.com/SataQiu">@SataQiu</a>) [SIG Cluster Lifecycle]</li>
<li>Kubectl logs: <code>--ignore-errors</code> is now honored by all containers, maintaining consistency with parallelConsumeRequest behavior. (<a href="https://github.com/kubernetes/kubernetes/pull/97686">#97686</a>, <a href="https://github.com/wzshiming">@wzshiming</a>)</li>
<li>Kubectl-convert: Fix <code>no kind &quot;Ingress&quot; is registered for version</code> error (<a href="https://github.com/kubernetes/kubernetes/pull/97754">#97754</a>, <a href="https://github.com/wzshiming">@wzshiming</a>)</li>
<li>Kubectl: Fixed panic when describing an ingress backend without an API Group (<a href="https://github.com/kubernetes/kubernetes/pull/100505">#100505</a>, <a href="https://github.com/lauchokyip">@lauchokyip</a>) [SIG CLI]</li>
<li>Kubelet now cleans up orphaned volume directories automatically (<a href="https://github.com/kubernetes/kubernetes/pull/95301">#95301</a>, <a href="https://github.com/lorenz">@lorenz</a>) [SIG Node and Storage]</li>
<li>Kubelet.exe on Windows now checks that the process running as administrator and the executing user account is listed in the built-in administrators group.  This is the equivalent to checking the process is running as uid 0. (<a href="https://github.com/kubernetes/kubernetes/pull/96616">#96616</a>, <a href="https://github.com/perithompson">@perithompson</a>) [SIG Node and Windows]</li>
<li>Kubelet: Fix kubelet from panic after getting the wrong signal (<a href="https://github.com/kubernetes/kubernetes/pull/98200">#98200</a>, <a href="https://github.com/wzshiming">@wzshiming</a>) [SIG Node]</li>
<li>Kubelet: Fix repeatedly acquiring the inhibit lock (<a href="https://github.com/kubernetes/kubernetes/pull/98088">#98088</a>, <a href="https://github.com/wzshiming">@wzshiming</a>) [SIG Node]</li>
<li>Kubelet: Fixed the bug of getting the number of cpu when the number of cpu logical processors is more than 64 in windows (<a href="https://github.com/kubernetes/kubernetes/pull/97378">#97378</a>, <a href="https://github.com/hwdef">@hwdef</a>) [SIG Node and Windows]</li>
<li>Limits lease to have 1000 maximum attached objects. (<a href="https://github.com/kubernetes/kubernetes/pull/98257">#98257</a>, <a href="https://github.com/lingsamuel">@lingsamuel</a>)</li>
<li>Mitigate CVE-2020-8555 for kube-up using GCE by preventing local loopback folume hosts. (<a href="https://github.com/kubernetes/kubernetes/pull/97934">#97934</a>, <a href="https://github.com/mattcary">@mattcary</a>) [SIG Cloud Provider and Storage]</li>
<li>On single-stack configured (IPv4 or IPv6, but not both) clusters, Services which are both headless (no clusterIP) and selectorless (empty or undefined selector) will report <code>ipFamilyPolicy RequireDualStack</code> and will have entries in <code>ipFamilies[]</code> for both IPv4 and IPv6.  This is a change from alpha, but does not have any impact on the manually-specified Endpoints and EndpointSlices for the Service. (<a href="https://github.com/kubernetes/kubernetes/pull/99555">#99555</a>, <a href="https://github.com/thockin">@thockin</a>) [SIG Apps and Network]</li>
<li>Performance regression #97685 has been fixed. (<a href="https://github.com/kubernetes/kubernetes/pull/97860">#97860</a>, <a href="https://github.com/MikeSpreitzer">@MikeSpreitzer</a>) [SIG API Machinery]</li>
<li>Pod Log stats for windows now reports metrics (<a href="https://github.com/kubernetes/kubernetes/pull/99221">#99221</a>, <a href="https://github.com/jsturtevant">@jsturtevant</a>) [SIG Node, Storage, Testing and Windows]</li>
<li>Pod status updates faster when reacting on probe results. The first readiness probe will be called faster when startup probes succeeded, which will make Pod status as ready faster. (<a href="https://github.com/kubernetes/kubernetes/pull/98376">#98376</a>, <a href="https://github.com/matthyx">@matthyx</a>)</li>
<li>Readjust <code>kubelet_containers_per_pod_count</code> buckets to only show metrics greater than 1. (<a href="https://github.com/kubernetes/kubernetes/pull/98169">#98169</a>, <a href="https://github.com/wawa0210">@wawa0210</a>)</li>
<li>Remove CSI topology from migrated in-tree gcepd volume. (<a href="https://github.com/kubernetes/kubernetes/pull/97823">#97823</a>, <a href="https://github.com/Jiawei0227">@Jiawei0227</a>) [SIG Cloud Provider and Storage]</li>
<li>Requests with invalid timeout parameters in the request URL now appear in the audit log correctly. (<a href="https://github.com/kubernetes/kubernetes/pull/96901">#96901</a>, <a href="https://github.com/tkashem">@tkashem</a>) [SIG API Machinery and Testing]</li>
<li>Resolve a &quot;concurrent map read and map write&quot; crashing error in the kubelet (<a href="https://github.com/kubernetes/kubernetes/pull/95111">#95111</a>, <a href="https://github.com/choury">@choury</a>) [SIG Node]</li>
<li>Resolves spurious <code>Failed to list *v1.Secret</code> or <code>Failed to list *v1.ConfigMap</code> messages in kubelet logs. (<a href="https://github.com/kubernetes/kubernetes/pull/99538">#99538</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG Auth and Node]</li>
<li>ResourceQuota of an entity now inclusively calculate Pod overhead (<a href="https://github.com/kubernetes/kubernetes/pull/99600">#99600</a>, <a href="https://github.com/gjkim42">@gjkim42</a>)</li>
<li>Return zero time (midnight on Jan. 1, 1970) instead of negative number when reporting startedAt and finishedAt of the not started or a running Pod when using <code>dockershim</code> as a runtime. (<a href="https://github.com/kubernetes/kubernetes/pull/99585">#99585</a>, <a href="https://github.com/Iceber">@Iceber</a>)</li>
<li>Reverts breaking change to inline AzureFile volumes; referenced secrets are now searched for in the same namespace as the pod as in previous releases. (<a href="https://github.com/kubernetes/kubernetes/pull/100563">#100563</a>, <a href="https://github.com/msau42">@msau42</a>)</li>
<li>Scores from InterPodAffinity have stronger differentiation. (<a href="https://github.com/kubernetes/kubernetes/pull/98096">#98096</a>, <a href="https://github.com/leileiwan">@leileiwan</a>) [SIG Scheduling]</li>
<li>Specifying the KUBE_TEST_REPO environment variable when e2e tests are executed will instruct the test infrastructure to load that image from a location within the specified repo, using a predefined pattern. (<a href="https://github.com/kubernetes/kubernetes/pull/93510">#93510</a>, <a href="https://github.com/smarterclayton">@smarterclayton</a>) [SIG Testing]</li>
<li>Static pods will be deleted gracefully. (<a href="https://github.com/kubernetes/kubernetes/pull/98103">#98103</a>, <a href="https://github.com/gjkim42">@gjkim42</a>) [SIG Node]</li>
<li>Sync node status during kubelet node shutdown.
Adds an pod admission handler that rejects new pods when the node is in progress of shutting down. (<a href="https://github.com/kubernetes/kubernetes/pull/98005">#98005</a>, <a href="https://github.com/wzshiming">@wzshiming</a>) [SIG Node]</li>
<li>The calculation of pod UIDs for static pods has changed to ensure each static pod gets a unique value - this will cause all static pod containers to be recreated/restarted if an in-place kubelet upgrade from 1.20 to 1.21 is performed. Note that draining pods before upgrading the kubelet across minor versions is the supported upgrade path. (<a href="https://github.com/kubernetes/kubernetes/pull/87461">#87461</a>, <a href="https://github.com/bboreham">@bboreham</a>) [SIG Node]</li>
<li>The maximum number of ports allowed in EndpointSlices has been increased from 100 to 20,000 (<a href="https://github.com/kubernetes/kubernetes/pull/99795">#99795</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG Network]</li>
<li>Truncates a message if it hits the <code>NoteLengthLimit</code> when the scheduler records an event for the pod that indicates the pod has failed to schedule. (<a href="https://github.com/kubernetes/kubernetes/pull/98715">#98715</a>, <a href="https://github.com/carlory">@carlory</a>)</li>
<li>Updated k8s.gcr.io/ingress-gce-404-server-with-metrics-amd64 to a version that serves /metrics endpoint on a non-default port. (<a href="https://github.com/kubernetes/kubernetes/pull/97621">#97621</a>, <a href="https://github.com/vbannai">@vbannai</a>) [SIG Cloud Provider]</li>
<li>Updates the commands ` + "`" + `
<ul>
<li>kubectl kustomize {arg}</li>
<li>kubectl apply -k {arg}
` + "`" + `to use same code as kustomize CLI <a href="https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv4.0.5">v4.0.5</a> (<a href="https://github.com/kubernetes/kubernetes/pull/98946">#98946</a>, <a href="https://github.com/monopole">@monopole</a>)</li>
</ul>
</li>
<li>Use force unmount for NFS volumes if regular mount fails after 1 minute timeout (<a href="https://github.com/kubernetes/kubernetes/pull/96844">#96844</a>, <a href="https://github.com/gnufied">@gnufied</a>) [SIG Storage]</li>
<li>Use network.Interface.VirtualMachine.ID to get the binded VM
Skip standalone VM when reconciling LoadBalancer (<a href="https://github.com/kubernetes/kubernetes/pull/97635">#97635</a>, <a href="https://github.com/nilo19">@nilo19</a>) [SIG Cloud Provider]</li>
<li>Using exec auth plugins with kubectl no longer results in warnings about constructing many client instances from the same exec auth config. (<a href="https://github.com/kubernetes/kubernetes/pull/97857">#97857</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG API Machinery and Auth]</li>
<li>When a CNI plugin returns dual-stack pod IPs, kubelet will now try to respect the
&quot;primary IP family&quot; of the cluster by picking a primary pod IP of the same family
as the (primary) node IP, rather than assuming that the CNI plugin returned the IPs
in the order the administrator wanted (since some CNI plugins don't allow
configuring this). (<a href="https://github.com/kubernetes/kubernetes/pull/97979">#97979</a>, <a href="https://github.com/danwinship">@danwinship</a>) [SIG Network and Node]</li>
<li>When dynamically provisioning Azure File volumes for a premium account, the requested size will be set to 100GB if the request is initially lower than this value to accommodate Azure File requirements. (<a href="https://github.com/kubernetes/kubernetes/pull/99122">#99122</a>, <a href="https://github.com/huffmanca">@huffmanca</a>) [SIG Cloud Provider and Storage]</li>
<li>When using <code>Containerd</code> on Windows, the <code>C:\Windows\System32\drivers\etc\hosts</code> file will now be managed by kubelet. (<a href="https://github.com/kubernetes/kubernetes/pull/83730">#83730</a>, <a href="https://github.com/claudiubelu">@claudiubelu</a>)</li>
<li><code>VolumeBindingArgs</code> now allow <code>BindTimeoutSeconds</code> to be set as zero, while the value zero indicates no waiting for the checking of volume binding operation. (<a href="https://github.com/kubernetes/kubernetes/pull/99835">#99835</a>, <a href="https://github.com/chendave">@chendave</a>) [SIG Scheduling and Storage]</li>
<li><code>kubectl exec</code> and <code>kubectl attach</code> now honor the <code>--quiet</code> flag which suppresses output from the local binary that could be confused by a script with the remote command output (all non-failure output is hidden). In addition, print inline with exec and attach the list of alternate containers when we default to the first spec.container. (<a href="https://github.com/kubernetes/kubernetes/pull/99004">#99004</a>, <a href="https://github.com/smarterclayton">@smarterclayton</a>) [SIG CLI]</li>
</ul>
<h3>Other (Cleanup or Flake)</h3>
<ul>
<li>APIs for kubelet annotations and labels from <code>k8s.io/kubernetes/pkg/kubelet/apis</code> are now moved under <code>k8s.io/kubelet/pkg/apis/</code> (<a href="https://github.com/kubernetes/kubernetes/pull/98931">#98931</a>, <a href="https://github.com/michaelbeaumont">@michaelbeaumont</a>)</li>
<li>Apiserver_request_duration_seconds is promoted to stable status. (<a href="https://github.com/kubernetes/kubernetes/pull/99925">#99925</a>, <a href="https://github.com/logicalhan">@logicalhan</a>) [SIG API Machinery, Instrumentation and Testing]</li>
<li>Bump github.com/Azure/go-autorest/autorest to v0.11.12 (<a href="https://github.com/kubernetes/kubernetes/pull/97033">#97033</a>, <a href="https://github.com/patrickshan">@patrickshan</a>) [SIG API Machinery, CLI, Cloud Provider and Cluster Lifecycle]</li>
<li>Clients required to use go1.15.8+ or go1.16+ if kube-apiserver has the goaway feature enabled to avoid unexpected data race condition. (<a href="https://github.com/kubernetes/kubernetes/pull/98809">#98809</a>, <a href="https://github.com/answer1991">@answer1991</a>)</li>
<li>Delete deprecated <code>service.beta.kubernetes.io/azure-load-balancer-mixed-protocols</code> mixed procotol annotation in favor of the MixedProtocolLBService feature (<a href="https://github.com/kubernetes/kubernetes/pull/97096">#97096</a>, <a href="https://github.com/nilo19">@nilo19</a>) [SIG Cloud Provider]</li>
<li>EndpointSlice generation is now incremented when labels change. (<a href="https://github.com/kubernetes/kubernetes/pull/99750">#99750</a>, <a href="https://github.com/robscott">@robscott</a>) [SIG Network]</li>
<li>Featuregate AllowInsecureBackendProxy graduates to GA and unconditionally enabled. (<a href="https://github.com/kubernetes/kubernetes/pull/99658">#99658</a>, <a href="https://github.com/deads2k">@deads2k</a>)</li>
<li>Increase timeout for pod lifecycle test to reach pod status=ready (<a href="https://github.com/kubernetes/kubernetes/pull/96691">#96691</a>, <a href="https://github.com/hh">@hh</a>)</li>
<li>Increased <code>CSINodeIDMaxLength</code> from 128 bytes to 192 bytes. (<a href="https://github.com/kubernetes/kubernetes/pull/98753">#98753</a>, <a href="https://github.com/Jiawei0227">@Jiawei0227</a>)</li>
<li>Kube-apiserver: The OIDC authenticator no longer waits 10 seconds before attempting to fetch the metadata required to verify tokens. (<a href="https://github.com/kubernetes/kubernetes/pull/97693">#97693</a>, <a href="https://github.com/enj">@enj</a>) [SIG API Machinery and Auth]</li>
<li>Kube-proxy: Traffic from the cluster directed to ExternalIPs is always sent directly to the Service. (<a href="https://github.com/kubernetes/kubernetes/pull/96296">#96296</a>, <a href="https://github.com/aojea">@aojea</a>) [SIG Network and Testing]</li>
<li>Kubeadm: change the default image repository for CI images from 'gcr.io/kubernetes-ci-images' to 'gcr.io/k8s-staging-ci-images' (<a href="https://github.com/kubernetes/kubernetes/pull/97087">#97087</a>, <a href="https://github.com/SataQiu">@SataQiu</a>) [SIG Cluster Lifecycle]</li>
<li>Kubectl: The deprecated <code>kubectl alpha debug</code> command is removed. Use <code>kubectl debug</code> instead. (<a href="https://github.com/kubernetes/kubernetes/pull/98111">#98111</a>, <a href="https://github.com/pandaamanda">@pandaamanda</a>) [SIG CLI]</li>
<li>Kubelet command line flags related to dockershim are now showing deprecation message as they will be removed along with dockershim in future release. (<a href="https://github.com/kubernetes/kubernetes/pull/98730">#98730</a>, <a href="https://github.com/dims">@dims</a>)</li>
<li>Official support to build kubernetes with docker-machine / remote docker is removed. This change does not affect building kubernetes with docker locally. (<a href="https://github.com/kubernetes/kubernetes/pull/97618">#97618</a>, <a href="https://github.com/jherrera123">@jherrera123</a>) [SIG Release and Testing]</li>
<li>Process start time on Windows now uses current process information (<a href="https://github.com/kubernetes/kubernetes/pull/97491">#97491</a>, <a href="https://github.com/jsturtevant">@jsturtevant</a>) [SIG API Machinery, CLI, Cloud Provider, Cluster Lifecycle, Instrumentation and Windows]</li>
<li>Resolves flakes in the Ingress conformance tests due to conflicts with controllers updating the Ingress object (<a href="https://github.com/kubernetes/kubernetes/pull/98430">#98430</a>, <a href="https://github.com/liggitt">@liggitt</a>) [SIG Network and Testing]</li>
<li>The <code>AttachVolumeLimit</code> feature gate (GA since v1.17) has been removed and now unconditionally enabled. (<a href="https://github.com/kubernetes/kubernetes/pull/96539">#96539</a>, <a href="https://github.com/ialidzhikov">@ialidzhikov</a>)</li>
<li>The <code>CSINodeInfo</code> feature gate that is GA since v1.17 is unconditionally enabled, and can no longer be specified via the <code>--feature-gates</code> argument. (<a href="https://github.com/kubernetes/kubernetes/pull/96561">#96561</a>, <a href="https://github.com/ialidzhikov">@ialidzhikov</a>) [SIG Apps, Auth, Scheduling, Storage and Testing]</li>
<li>The <code>apiserver_request_total</code> metric is promoted to stable status and no longer has a content-type dimensions, so any alerts/charts which presume the existence of this will fail. This is however, unlikely to be the case since it was effectively an unbounded dimension in the first place. (<a href="https://github.com/kubernetes/kubernetes/pull/99788">#99788</a>, <a href="https://github.com/logicalhan">@logicalhan</a>)</li>
<li>The default delegating authorization options now allow unauthenticated access to healthz, readyz, and livez.  A system:masters user connecting to an authz delegator will not perform an authz check. (<a href="https://github.com/kubernetes/kubernetes/pull/98325">#98325</a>, <a href="https://github.com/deads2k">@deads2k</a>) [SIG API Machinery, Auth, Cloud Provider and Scheduling]</li>
<li>The deprecated feature gates <code>CSIDriverRegistry</code>, <code>BlockVolume</code> and <code>CSIBlockVolume</code> are now unconditionally enabled and can no longer be specified in component invocations. (<a href="https://github.com/kubernetes/kubernetes/pull/98021">#98021</a>, <a href="https://github.com/gavinfish">@gavinfish</a>) [SIG Storage]</li>
<li>The deprecated feature gates <code>RotateKubeletClientCertificate</code>, <code>AttachVolumeLimit</code>, <code>VolumePVCDataSource</code> and <code>EvenPodsSpread</code> are now unconditionally enabled and can no longer be specified in component invocations. (<a href="https://github.com/kubernetes/kubernetes/pull/97306">#97306</a>, <a href="https://github.com/gavinfish">@gavinfish</a>) [SIG Node, Scheduling and Storage]</li>
<li>The e2e suite can be instructed not to wait for pods in kube-system to be ready or for all nodes to be ready by passing <code>--allowed-not-ready-nodes=-1</code> when invoking the e2e.test program. This allows callers to run subsets of the e2e suite in scenarios other than perfectly healthy clusters. (<a href="https://github.com/kubernetes/kubernetes/pull/98781">#98781</a>, <a href="https://github.com/smarterclayton">@smarterclayton</a>) [SIG Testing]</li>
<li>The feature gates <code>WindowsGMSA</code> and <code>WindowsRunAsUserName</code> that are GA since v1.18 are now removed. (<a href="https://github.com/kubernetes/kubernetes/pull/96531">#96531</a>, <a href="https://github.com/ialidzhikov">@ialidzhikov</a>) [SIG Node and Windows]</li>
<li>The new <code>-gce-zones</code> flag on the <code>e2e.test</code> binary instructs tests that check for information about how the cluster interacts with the cloud to limit their queries to the provided zone list. If not specified, the current behavior of asking the cloud provider for all available zones in multi zone clusters is preserved. (<a href="https://github.com/kubernetes/kubernetes/pull/98787">#98787</a>, <a href="https://github.com/smarterclayton">@smarterclayton</a>) [SIG API Machinery, Cluster Lifecycle and Testing]</li>
<li>Update cri-tools to <a href="https://github.com/kubernetes-sigs/cri-tools/releases/tag/v1.20.0">v1.20.0</a> (<a href="https://github.com/kubernetes/kubernetes/pull/97967">#97967</a>, <a href="https://github.com/rajibmitra">@rajibmitra</a>) [SIG Cloud Provider]</li>
<li>Windows nodes on GCE will take longer to start due to dependencies installed at node creation time. (<a href="https://github.com/kubernetes/kubernetes/pull/98284">#98284</a>, <a href="https://github.com/pjh">@pjh</a>) [SIG Cloud Provider]</li>
<li><code>apiserver_storage_objects</code> (a newer version of <code>etcd_object_counts</code>) is promoted and marked as stable. (<a href="https://github.com/kubernetes/kubernetes/pull/100082">#100082</a>, <a href="https://github.com/logicalhan">@logicalhan</a>)</li>
</ul>
<h3>Uncategorized</h3>
<ul>
<li>GCE L4 Loadbalancers now handle &gt; 5 ports in service spec correctly. (<a href="https://github.com/kubernetes/kubernetes/pull/99595">#99595</a>, <a href="https://github.com/prameshj">@prameshj</a>) [SIG Cloud Provider]</li>
<li>The DownwardAPIHugePages feature is beta.  Users may use the feature if all workers in their cluster are min 1.20 version.  The feature will be enabled by default in all installations in 1.22. (<a href="https://github.com/kubernetes/kubernetes/pull/99610">#99610</a>, <a href="https://github.com/derekwaynecarr">@derekwaynecarr</a>) [SIG Node]</li>
</ul>
<h2>Dependencies</h2>
<h3>Added</h3>
<ul>
<li>github.com/go-errors/errors: <a href="https://github.com/go-errors/errors/tree/v1.0.1">v1.0.1</a></li>
<li>github.com/gobuffalo/here: <a href="https://github.com/gobuffalo/here/tree/v0.6.0">v0.6.0</a></li>
<li>github.com/google/shlex: <a href="https://github.com/google/shlex/tree/e7afc7f">e7afc7f</a></li>
<li>github.com/markbates/pkger: <a href="https://github.com/markbates/pkger/tree/v0.17.1">v0.17.1</a></li>
<li>github.com/moby/spdystream: <a href="https://github.com/moby/spdystream/tree/v0.2.0">v0.2.0</a></li>
<li>github.com/monochromegane/go-gitignore: <a href="https://github.com/monochromegane/go-gitignore/tree/205db1a">205db1a</a></li>
<li>github.com/niemeyer/pretty: <a href="https://github.com/niemeyer/pretty/tree/a10e7ca">a10e7ca</a></li>
<li>github.com/xlab/treeprint: <a href="https://github.com/xlab/treeprint/tree/a009c39">a009c39</a></li>
<li>go.starlark.net: 8dd3e2e</li>
<li>golang.org/x/term: 6a3ed07</li>
<li>sigs.k8s.io/kustomize/api: v0.8.5</li>
<li>sigs.k8s.io/kustomize/cmd/config: v0.9.7</li>
<li>sigs.k8s.io/kustomize/kustomize/v4: v4.0.5</li>
<li>sigs.k8s.io/kustomize/kyaml: v0.10.15</li>
</ul>
<h3>Changed</h3>
<ul>
<li>dmitri.shuralyov.com/gpu/mtl: 666a987 → 28db891</li>
<li>github.com/Azure/go-autorest/autorest: <a href="https://github.com/Azure/go-autorest/autorest/compare/v0.11.1...v0.11.12">v0.11.1 → v0.11.12</a></li>
<li>github.com/NYTimes/gziphandler: <a href="https://github.com/NYTimes/gziphandler/compare/56545f4...v1.1.1">56545f4 → v1.1.1</a></li>
<li>github.com/cilium/ebpf: <a href="https://github.com/cilium/ebpf/compare/1c8d4c9...v0.2.0">1c8d4c9 → v0.2.0</a></li>
<li>github.com/container-storage-interface/spec: <a href="https://github.com/container-storage-interface/spec/compare/v1.2.0...v1.3.0">v1.2.0 → v1.3.0</a></li>
<li>github.com/containerd/console: <a href="https://github.com/containerd/console/compare/v1.0.0...v1.0.1">v1.0.0 → v1.0.1</a></li>
<li>github.com/containerd/containerd: <a href="https://github.com/containerd/containerd/compare/v1.4.1...v1.4.4">v1.4.1 → v1.4.4</a></li>
<li>github.com/coredns/corefile-migration: <a href="https://github.com/coredns/corefile-migration/compare/v1.0.10...v1.0.11">v1.0.10 → v1.0.11</a></li>
<li>github.com/creack/pty: <a href="https://github.com/creack/pty/compare/v1.1.7...v1.1.11">v1.1.7 → v1.1.11</a></li>
<li>github.com/docker/docker: <a href="https://github.com/docker/docker/compare/bd33bbf...v20.10.2">bd33bbf → v20.10.2+incompatible</a></li>
<li>github.com/go-logr/logr: <a href="https://github.com/go-logr/logr/compare/v0.2.0...v0.4.0">v0.2.0 → v0.4.0</a></li>
<li>github.com/go-openapi/spec: <a href="https://github.com/go-openapi/spec/compare/v0.19.3...v0.19.5">v0.19.3 → v0.19.5</a></li>
<li>github.com/go-openapi/strfmt: <a href="https://github.com/go-openapi/strfmt/compare/v0.19.3...v0.19.5">v0.19.3 → v0.19.5</a></li>
<li>github.com/go-openapi/validate: <a href="https://github.com/go-openapi/validate/compare/v0.19.5...v0.19.8">v0.19.5 → v0.19.8</a></li>
<li>github.com/gogo/protobuf: <a href="https://github.com/gogo/protobuf/compare/v1.3.1...v1.3.2">v1.3.1 → v1.3.2</a></li>
<li>github.com/golang/mock: <a href="https://github.com/golang/mock/compare/v1.4.1...v1.4.4">v1.4.1 → v1.4.4</a></li>
<li>github.com/google/cadvisor: <a href="https://github.com/google/cadvisor/compare/v0.38.5...v0.39.0">v0.38.5 → v0.39.0</a></li>
<li>github.com/heketi/heketi: <a href="https://github.com/heketi/heketi/compare/c2e2a4a...v10.2.0">c2e2a4a → v10.2.0+incompatible</a></li>
<li>github.com/kisielk/errcheck: <a href="https://github.com/kisielk/errcheck/compare/v1.2.0...v1.5.0">v1.2.0 → v1.5.0</a></li>
<li>github.com/konsorten/go-windows-terminal-sequences: <a href="https://github.com/konsorten/go-windows-terminal-sequences/compare/v1.0.3...v1.0.2">v1.0.3 → v1.0.2</a></li>
<li>github.com/kr/text: <a href="https://github.com/kr/text/compare/v0.1.0...v0.2.0">v0.1.0 → v0.2.0</a></li>
<li>github.com/mattn/go-runewidth: <a href="https://github.com/mattn/go-runewidth/compare/v0.0.2...v0.0.7">v0.0.2 → v0.0.7</a></li>
<li>github.com/miekg/dns: <a href="https://github.com/miekg/dns/compare/v1.1.4...v1.1.35">v1.1.4 → v1.1.35</a></li>
<li>github.com/moby/sys/mountinfo: <a href="https://github.com/moby/sys/mountinfo/compare/v0.1.3...v0.4.0">v0.1.3 → v0.4.0</a></li>
<li>github.com/moby/term: <a href="https://github.com/moby/term/compare/672ec06...df9cb8a">672ec06 → df9cb8a</a></li>
<li>github.com/mrunalp/fileutils: <a href="https://github.com/mrunalp/fileutils/compare/abd8a0e...v0.5.0">abd8a0e → v0.5.0</a></li>
<li>github.com/olekukonko/tablewriter: <a href="https://github.com/olekukonko/tablewriter/compare/a0225b3...v0.0.4">a0225b3 → v0.0.4</a></li>
<li>github.com/opencontainers/runc: <a href="https://github.com/opencontainers/runc/compare/v1.0.0-rc92...v1.0.0-rc93">v1.0.0-rc92 → v1.0.0-rc93</a></li>
<li>github.com/opencontainers/runtime-spec: <a href="https://github.com/opencontainers/runtime-spec/compare/4d89ac9...e6143ca">4d89ac9 → e6143ca</a></li>
<li>github.com/opencontainers/selinux: <a href="https://github.com/opencontainers/selinux/compare/v1.6.0...v1.8.0">v1.6.0 → v1.8.0</a></li>
<li>github.com/sergi/go-diff: <a href="https://github.com/sergi/go-diff/compare/v1.0.0...v1.1.0">v1.0.0 → v1.1.0</a></li>
<li>github.com/sirupsen/logrus: <a href="https://github.com/sirupsen/logrus/compare/v1.6.0...v1.7.0">v1.6.0 → v1.7.0</a></li>
<li>github.com/syndtr/gocapability: <a href="https://github.com/syndtr/gocapability/compare/d983527...42c35b4">d983527 → 42c35b4</a></li>
<li>github.com/willf/bitset: <a href="https://github.com/willf/bitset/compare/d5bec33...v1.1.11">d5bec33 → v1.1.11</a></li>
<li>github.com/yuin/goldmark: <a href="https://github.com/yuin/goldmark/compare/v1.1.27...v1.2.1">v1.1.27 → v1.2.1</a></li>
<li>golang.org/x/crypto: 7f63de1 → 5ea612d</li>
<li>golang.org/x/exp: 6cc2880 → 85be41e</li>
<li>golang.org/x/mobile: d2bd2a2 → e6ae53a</li>
<li>golang.org/x/mod: v0.3.0 → ce943fd</li>
<li>golang.org/x/net: 69a7880 → 3d97a24</li>
<li>golang.org/x/sync: cd5d95a → 67f06af</li>
<li>golang.org/x/sys: 5cba982 → a50acf3</li>
<li>golang.org/x/time: 3af7569 → f8bda1e</li>
<li>golang.org/x/tools: c1934b7 → v0.1.0</li>
<li>gopkg.in/check.v1: 41f04d3 → 8fa4692</li>
<li>gopkg.in/yaml.v2: v2.2.8 → v2.4.0</li>
<li>gotest.tools/v3: v3.0.2 → v3.0.3</li>
<li>k8s.io/gengo: 83324d8 → b6c5ce2</li>
<li>k8s.io/klog/v2: v2.4.0 → v2.8.0</li>
<li>k8s.io/kube-openapi: d219536 → 591a79e</li>
<li>k8s.io/system-validators: v1.2.0 → v1.4.0</li>
<li>sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.0.14 → v0.0.15</li>
<li>sigs.k8s.io/structured-merge-diff/v4: v4.0.2 → v4.1.0</li>
</ul>
<h3>Removed</h3>
<ul>
<li>github.com/codegangsta/negroni: <a href="https://github.com/codegangsta/negroni/tree/v1.0.0">v1.0.0</a></li>
<li>github.com/docker/spdystream: <a href="https://github.com/docker/spdystream/tree/449fdfc">449fdfc</a></li>
<li>github.com/golangplus/bytes: <a href="https://github.com/golangplus/bytes/tree/45c989f">45c989f</a></li>
<li>github.com/golangplus/fmt: <a href="https://github.com/golangplus/fmt/tree/2a5d6d7">2a5d6d7</a></li>
<li>github.com/gorilla/context: <a href="https://github.com/gorilla/context/tree/v1.1.1">v1.1.1</a></li>
<li>github.com/kr/pty: <a href="https://github.com/kr/pty/tree/v1.1.5">v1.1.5</a></li>
<li>rsc.io/quote/v3: v3.1.0</li>
<li>rsc.io/sampler: v1.3.0</li>
<li>sigs.k8s.io/kustomize: v2.0.3+incompatible</li>
</ul>

  </body>
</html>`

const rcReleaseExpectedTOC = `<!-- BEGIN MUNGE: GENERATED_TOC -->

- [v1.16.0-rc.1](#v1160-rc1)`

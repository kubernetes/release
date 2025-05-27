---
name: Dependency update - Golang
about: Create a tracking issue for updating Golang dependencies
title: Dependency update - Golang 1.xy.z/1.xy.z
labels: kind/feature, sig/release, area/release-eng, area/dependency

---
<!--
Please only use this template if you are a Release Manager updating
Golang dependencies.
-->

### Tracking info

<!-- Search query: https://github.com/kubernetes/release/issues?q=is%3Aissue+Dependency+update+-+Golang -->
<!-- Example: https://github.com/kubernetes/release/issues/3383 -->
Link to any previous tracking issue: 

<!-- golang-announce mailing list: https://groups.google.com/g/golang-announce -->
Golang mailing list announcement: 

SIG Release Slack thread: 

### Work items

<!-- Example: https://github.com/kubernetes/release/pull/3388 -->
- [ ] `kube-cross`, `go-runner`, `releng-ci` image updates: 

  <!-- Example: https://github.com/kubernetes/k8s.io/pull/6153 -->
  - [ ] `kube-cross` image promotion: 

  <!-- Example: https://github.com/kubernetes/k8s.io/pull/6154 -->
  - [ ] `go-runner` image promotion: 

  <!-- Example: https://github.com/kubernetes/k8s.io/pull/6155 -->
  - [ ] `releng-ci` image promotion: 

#### After go-runner image promotion

<!-- Example: https://github.com/kubernetes/release/pull/3389 -->
- [ ] `distroless-iptables` image update: 

  <!-- Example: https://github.com/kubernetes/k8s.io/pull/6164 -->
  - [ ] image promotion: 

#### After kube-cross and distroless-iptables image promotions

<!-- Example: https://github.com/kubernetes/kubernetes/pull/122201 -->
- [ ] kubernetes/kubernetes update (`master`): 

  Ensure the following have been updated within the PR:

  - [ ] `.go-version` file
  - [ ] kube-cross image
  - [ ] go-runner image
  - [ ] distroless-iptables image
  - [ ] publishing bot rules

> **Note**
> This update may require an update to go.sum files, for example: https://github.com/kubernetes/kubernetes/pull/118507
> This will require an API Review approval.

#### After kubernetes/kubernetes (master) has been updated

<!--
Notice: Always use the oldest supported distribution release (Debian bullseye as
time of writing) to achieve maximum glibc compatibility of the kubelet. Other
images can use the latest available release.

Example: https://github.com/kubernetes/release/pull/3390
-->
- [ ] `k8s-cloud-builder` and `k8s-ci-builder` image updates: 

<!-- Example: https://github.com/kubernetes/test-infra/pull/31387 -->
- [ ] `kubekins`/`krte` image variants update: 

### Cherry picks

<!--
Depending on the Golang release type, this section may not be required.

General rule of thumb:
Only cherry pick Golang patch releases to branches that have the same Golang
minor release version.

Concrete example:
At the time of this template's creation, go1.15.5 was just merged on our
primary development branch and the following Golang versions were active on
in-support kubernetes/kubernetes release branches:
- `master`: go1.15.5
- `release-1.19`: go1.15.2
- `release-1.18`: go1.13.15
- `release-1.17`: go1.13.15

In this case, we would only cherry pick the go1.15.5 to the `release-1.19`
branch, since it is the only other branch with a go1.15 minor version on it.
-->

- [ ] Kubernetes 1.y-1: 
- [ ] Kubernetes 1.y-2: 
- [ ] Kubernetes 1.y-3: 

<!--
  If the Golang version of the active development branch (`master`) is newer than
any of the Golang versions on _active_ release branches, then the current
Golang versions for all release branches need to be updated within publishing
bot rules.
  Concrete example:
  - `master` was just updated from go1.16.6 to go1.16.7
  - cherry picks were issued to the 1.22 and 1.21 branches
  - `release-1.20` was also updated from go1.15.14 to go1.15.15
  - these changes were cherry picked to the 1.19 branch

  In this case, because we updated the default go version on `master` to
go1.16.7, there's no action required for staging repositories using go1.16.
  However, for staging repository branches using go1.15, the `master` branch's
publishing bot rules need to be updated to learn about the Golang update that
happened for the 1.20 and 1.19 Kubernetes release branches.
  PR: https://github.com/kubernetes/kubernetes/pull/122299
-->
- [ ] publishing bot rule updates for active Golang versions: 


#### After kubernetes/kubernetes (release branches) has been updated

<!--
Notice: Always use the oldest supported distribution release (Debian bullseye as
time of writing) to achieve maximum glibc compatibility of the kubelet. Other
images can use the latest available release.

Example: https://github.com/kubernetes/release/pull/3394
-->
- [ ] `k8s-cloud-builder` and `k8s-ci-builder` image updates: 

<!-- Example: https://github.com/kubernetes/test-infra/pull/31398 -->
- [ ] `kubekins`/`krte` image updates: 

### Follow-up items

<!--
Use this section to list out process improvements or items that need to be
addressed before the next Golang update.
-->

- [ ] Ensure the Golang issue template is updated with any new requirements
- [ ] <Any other follow up items>

/assign
cc: @kubernetes/release-engineering

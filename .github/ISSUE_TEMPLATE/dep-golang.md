---
name: Dependency update - Golang
about: Create a tracking issue for updating Golang dependencies
labels: kind/feature, sig/release, area/release-eng, area/dependency

---
<!--
Please only use this template if you are a Release Manager updating
Golang dependencies.
-->

### Tracking info

Link to any previous tracking issue: 

Golang mailing list announcement: 

SIG Release Slack thread: 

### Work items for `gox.y.z`

<!-- Example: https://github.com/kubernetes/release/pull/1732 -->
- [ ] `kube-cross` image update: 

  <!-- Example: https://github.com/kubernetes/k8s.io/pull/1423 -->
  - [ ] image promotion: 

<!-- Example: https://github.com/kubernetes/release/pull/1780 -->
- [ ] `go-runner` image update: 

  <!-- Example: https://github.com/kubernetes/k8s.io/pull/1449 -->
  - [ ] image promotion: 

<!-- Example: https://github.com/kubernetes/release/pull/1732 -->
- [ ] `vulndash` image update: 

  <!-- Example: https://github.com/kubernetes/k8s.io/pull/1423 -->
  - [ ] image promotion: 

<!-- Example: https://github.com/kubernetes/release/pull/1732 -->
- [ ] `releng-ci` image update: 

  <!-- Example: https://github.com/kubernetes/k8s.io/pull/1423 -->
  - [ ] image promotion: 

#### After kube-cross image promotion

<!-- Example: https://github.com/kubernetes/kubernetes/pull/95776 -->
- [ ] kubernetes/kubernetes update (`master`): 

  Ensure the following have been updated within the PR:

  - [ ] kube-cross image
  - [ ] go-runner image
  - [ ] publishing bot rules

#### After kubernetes/kubernetes has been updated

<!-- Example: https://github.com/kubernetes/release/pull/1784 -->
- [ ] `k8s-cloud-builder` image update: 

<!-- Example: https://github.com/kubernetes/release/pull/1784 -->
- [ ] `k8s-ci-builder` image variants update: 

<!-- Example: https://github.com/kubernetes/test-infra/pull/20038 -->
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

- [ ] Kubernetes x.y-1: 
- [ ] Kubernetes x.y-2: 
- [ ] Kubernetes x.y-3: 

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
  PR: https://github.com/kubernetes/kubernetes/pull/104226
-->
- [ ] publishing bot rule updates for active Golang versions: 


### Follow-up items

<!--
Use this section to list out process improvements or items that need to be
addressed before the next Golang update.
-->

- [ ] Ensure the Golang issue template is updated with any new requirements
- [ ] 

/assign
cc: @kubernetes/release-engineering

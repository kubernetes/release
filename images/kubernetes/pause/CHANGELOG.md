# CHANGELOG

## v3.4.0

- Image building has been migrated to kubernetes/release
- Built using `kube-cross:v1.15.2-1`
- Updated to use image tags that are in line with other Release Engineering
  built images: `vx.y.z`

## 3.3

- The pause container is now built on Kubernetes Community infra with kube-cross:v1.13.9-2 ([#87954](https://github.com/kubernetes/kubernetes/pull/90665), [@justaugustus](https://github.com/justaugustus))

## 3.2

- The pause container is built with the correct "Architecture" metadata. ([#87954](https://prs.k8s.io/87954), [@BenTheElder](https://github.com/BenTheElder))

## 3.1

- The pause container gains a signal handler to clean up orphaned zombie processes. ([#36853](https://prs.k8s.io/36853), [@verb](https://github.com/verb))
- `pause -v` will return build information for the pause binary. ([#56762](https://prs.k8s.io/56762), [@verb](https://github.com/verb))

## 3.0

- The pause container was rewritten entirely in C. ([#23009](https://prs.k8s.io/23009), [@uluyol](https://github.com/uluyol))

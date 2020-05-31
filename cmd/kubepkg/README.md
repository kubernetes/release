# kubepkg <!-- omit in toc -->

`kubepkg` is a tool for building deb and rpm packages for Kubernetes components.

**NOTE: `kubepkg` is currently in development and its' design is expected to rapidly change. If you encounter errors, please file an issue in this repo.**

- [Installation](#installation)
- [Usage](#usage)
  - [Example: Building nightly kubeadm debs for amd64 architecture](#example-building-nightly-kubeadm-debs-for-amd64-architecture)
  - [Example: Building deb specs for all packages, all channels, and all architectures](#example-building-deb-specs-for-all-packages-all-channels-and-all-architectures)
- [Known Issues](#known-issues)
  - [Building rpms is not _currently_ supported](#building-rpms-is-not-currently-supported)

## Installation

From this directory:

```shell
go install ./...
```

## Usage

`kubepkg [command]`

```shell
Available Commands:
  debs        debs creates Debian-based packages for Kubernetes components
  help        Help about any command
  rpms        rpms creates RPMs for Kubernetes components

Flags:
      --arch strings                        architectures to build for (default [amd64,arm,arm64,ppc64le,s390x])
      --channels strings                    channels to build for (default [release,testing,nightly])
      --cni-version string                  CNI version to build
      --cri-tools-version string            CRI tools version to build
  -h, --help                                help for kubepkg
      --kube-version string                 Kubernetes version to build
      --log-level string                    the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace' (default "info")
      --packages strings                    packages to build (default [kubelet,kubectl,kubeadm,kubernetes-cni,cri-tools])
      --release-download-link-base string   release download link base (default "https://dl.k8s.io")
      --revision string                     deb package revision. (default "0")
      --spec-only                           only create specs instead of building packages
      --template-dir string                 template directory (default "templates/latest")
```

### Example: Building nightly kubeadm debs for amd64 architecture

```shell
kubepkg debs --packages kubeadm --channels nightly --arch amd64
```

### Example: Building deb specs for all packages, all channels, and all architectures

```shell
kubepkg debs --spec-only
```

## Known Issues

### Building rpms is not _currently_ supported

We haven't written the logic for building rpms yet.

Right now, you can build rpm specs using the `--spec-only` flag and then use a tool of your choice to build the rpms using the specs produced.

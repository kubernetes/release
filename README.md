# Kubernetes Release Tooling

[![PkgGoDev](https://pkg.go.dev/badge/k8s.io/release)](https://pkg.go.dev/k8s.io/release)
[![Go Report Card](https://goreportcard.com/badge/k8s.io/release)](https://goreportcard.com/report/k8s.io/release)
[![Slack](https://img.shields.io/badge/Slack-%23release--management-blueviolet)](https://kubernetes.slack.com/archives/C2C40FMNF)

This repo contains the tooling and infrastructure configurations for creating
Kubernetes releases from the [kubernetes/kubernetes] main repository.

There are several scripts and helpers in this repository a Release Manager will
find useful when managing all kinds of releases (alpha, beta, official, rc)
across branches.

- [Release Management](#release-management)
  - [`krel`](#krel)
  - [`schedule-builder`](#schedule-builder)
- [Artifact Management](#artifact-management)
  - [`kpromo`](#kpromo)
  - [`kubepkg`](#kubepkg)
  - [`cip-mm`](#cip-mm)
  - [`gh2gcs`](#gh2gcs)
  - [`vulndash`](#vulndash)
- [End User](#end-user)
  - [`release-notes`](#release-notes)
  - [`gcbuilder`](#gcbuilder)
  - [`publish-release`](#publish-release)
- [Legacy](#legacy)
  - [`push-build.sh`](#push-buildsh)
- [Contributing](#contributing)

**Each of the headings below links to a tool's location in the repository.**

## Release Management

### [`krel`](/cmd/krel)

**K**ubernetes **rel**ease Toolbox: tooling for releasing Kubernetes

Status: Feature Complete

Audience: [Release Managers][release-managers]

Details: [Documentation](/docs/krel/README.md)

### [`schedule-builder`](/cmd/schedule-builder)

Generate a Markdown schedule for Kubernetes releases.

Status: In Progress

Audience: [Release Managers][release-managers]

Details: [Documentation](/cmd/schedule-builder/README.md)

## Artifact Management

### [`kpromo`](/cmd/kpromo)

**K**ubernetes artifact **promo**tion tooling: tooling for promoting artifacts

Status: In Progress

Audience: [Release Managers][release-managers] and subproject maintainers
responsible for promoting file or container artifacts

Details: [Documentation](/cmd/kpromo/README.md)

### [`kubepkg`](/cmd/kubepkg)

Create Kubernetes deb/rpm packages.

Status: In Progress

Audience: [Release Managers][release-managers]

Details: [Documentation](/cmd/kubepkg/README.md)

### [`cip-mm`](/cmd/cip-mm)

Modify container image manifests for promotion.

Status: In Progress

Details: [Documentation](/cmd/cip-mm/README.md)

### [`gh2gcs`](/cmd/gh2gcs)

Upload GitHub release assets to Google Cloud Storage.

Status: In Progress

Audience: [Release Managers][release-managers] and subproject maintainers
responsible for promoting container artifacts

Details: [Documentation](/cmd/gh2gcs/README.md)

### [`vulndash`](/cmd/vulndash)

Generate a dashboard of container image vulnerabilities.

Status: In Progress

Audience: [Release Managers][release-managers]

Details: [Documentation](/docs/vuln-dashboard.md)

## End User

### [`release-notes`](/cmd/release-notes)

Scrape GitHub pull requests for release notes.

Status: Feature Complete

Details: [Documentation](/cmd/release-notes/README.md)

### [`gcbuilder`](/cmd/gcbuilder)

General purpose tool for triggering Google Cloud Build (GCB) runs with
substitutions.

Status: Unused

Details: [Documentation](/cmd/gcbuilder/README.md)

### [`publish-release`](/cmd/publish-release)

A tool to announce software releases. Currently supports updating the
release page on GitHub based on templates and updating release artifacts.

Details: [Documentation](cmd/publish-release/README.md)

## Legacy

### [`push-build.sh`](push-build.sh)

Push a CI build of Kubernetes to Google Cloud Storage (GCS).

Status: Deprecated (but still in use)

Audience: [Release Managers][release-managers], Prowjobs

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute.

[kubernetes/kubernetes]: https://git.k8s.io/kubernetes
[Branch Manager Handbook]: https://git.k8s.io/sig-release/release-engineering/role-handbooks/branch-manager.md
[release-managers]: https://git.k8s.io/sig-release/release-managers.md

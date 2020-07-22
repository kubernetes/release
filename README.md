<!-- BEGIN MUNGE: GENERATED_TOC -->
- [Kubernetes Release](#kubernetes-release)
  - [Intro](#intro)
  - [Tools](#tools)
  - [Release Notes Gathering](#release-notes-gathering)
  - [Building Linux Packages](#building-linux-packages)
  - [Contributing](#contributing)
<!-- END MUNGE: GENERATED_TOC -->

# Kubernetes Release

This repo contains the tooling and infrastructure to create Kubernetes releases from the [kubernetes/kubernetes] main repository.

## Intro

There are several scripts and helpers in this repository a release
manager will find useful when managing all kinds of releases (alpha,
beta, official, rc) across branches.

## Tools

Most tools in this repo run by default in *mock* mode to allow for ease in
development and testing.

Tools | Description
 :---: | --
[`krel`](/cmd/krel)                     | Kubernetes Release Toolbox<br/>This is the new golang based tool for managing releases
[`anago`](/anago)                       | Release Tool:<br/>The main driver for creating staged builds and releases. This is what runs inside GCB after a job is submitted using `krel gcbmgr`.
[`find_green_build`](/find_green_build) | Asks Jenkins for a good build to use.
[`release-notes`](/cmd/release-notes)   | Scrape GitHub for release notes.<br/>See [Release Notes Gathering](#release-notes-gathering) for more information.
[`prin`](/prin)                         | To show release tags of a particular PR or commit.
[`push-build.sh`](/push-build.sh)       | Pushes a developer build or CI Jenkins build up to GCS.
[`script-template`](/script-template)   | Generates a script template in the kubernetes/release ecosystem.
[`testgridshot`](/testgridshot)         | Screenshots failing testgrid dashboards and creates a markdown stub that can be copied and pasted into a GitHub issue comment.<br/>This makes it easier to create comments like [this][shot-issue] as part of the release process.

For information on how to use `krel` and `anago`, see the [Branch Manager Handbook]

[kubernetes/kubernetes]: https://git.k8s.io/kubernetes
[Branch Manager Handbook]: https://git.k8s.io/sig-release/release-engineering/role-handbooks/branch-manager.md
[shot-issue]: https://github.com/kubernetes/sig-release/issues/756#issuecomment-520721968

## Release Notes Gathering

For more extensive build and usage documentation for the `release-notes` tool, see the [documentation](./cmd/release-notes/README.md).

Once the tool is installed, use `-h` or `--help` to see the command usage:

```
$ release-notes -h
release-notes - The Kubernetes Release Notes Generator

Usage:
  release-notes [flags]

Flags:
      --branch master            Select which branch to scrape. Defaults to master (default "master")
      --debug                    Enable debug logging
      --dependencies             Add dependency report (default true)
      --discover string          The revision discovery mode for automatic revision retrieval (options: none, mergebase-to-latest, patch-to-patch, minor-to-minor) (default "none")
      --end-rev string           The git revision to end at.
      --format string            The format for notes output (options: json, markdown) (default "markdown")
      --github-org string        Name of github organization (default "kubernetes")
      --github-repo string       Name of github repository (default "kubernetes")
      --go-template string       The go template to be used if --format=markdown (options: go-template:default, go-template:inline:<template>, go-template:<file.template>) (default "go-template:default")
  -h, --help                     help for release-notes
      --output string            The path to the where the release notes will be printed
      --record string            Record the API into a directory
      --release-bucket string    Specify gs bucket to point to in generated notes (default "kubernetes-release")
      --release-tars string      Directory of tars to sha512 sum for display
      --release-version string   Which release version to tag the entries as.
      --replay string            Replay a previously recorded API from a directory
      --repo-path string         Path to a local Kubernetes repository, used only for tag discovery. (default "/tmp/k8s-repo")
      --required-author string   Only commits from this GitHub user are considered. Set to empty string to include all users (default "k8s-ci-robot")
      --start-rev string         The git revision to start at.
      --toc                      Enable the rendering of the table of contents
```

## Building Linux Packages

See the [`kubepkg`](/cmd/kubepkg/README.md) documentation for instructions on how to build debs and rpms.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute.

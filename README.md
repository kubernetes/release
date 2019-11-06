<!-- BEGIN MUNGE: GENERATED_TOC -->
- [Kubernetes Release](#kubernetes-release)
  - [Intro](#intro)
  - [Tools](#tools)
  - [Release Notes Gathering](#release-notes-gathering)
  - [Building Linux Packages](#building-linux-packages)
    - [For Debian](#for-debian)
    - [For Fedora, CentOS, Red Hat Enterprise Linux](#for-fedora-centos-red-hat-enterprise-linux)
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
[`gcbmgr`](/gcbmgr)                     | Google Cloud Builder Manager:<br/>This is the main entry point for release managers for producing releases. All release types can be staged and later released using this method.
[`anago`](/anago)                       | Release Tool:<br/>The main driver for creating staged builds and releases. This is what runs inside GCB after a job is submitted using `gcbmgr`.
[`branchff`](/branchff)                 | Fast-forward branching helper:<br/>A tool used to pull new patches onto the release branch.
[`find_green_build`](/find_green_build) | Asks Jenkins for a good build to use.
[`release-notes`](/cmd/release-notes)   | Scrape GitHub for release notes.<br/>See [Release Notes Gathering](#release-notes-gathering) for more information.
[`prin`](/prin)                         | To show release tags of a particular PR or commit.
[`changelog-update`](/changelog-update) | Updates [CHANGELOG.md] version entries by rescanning github for text and label changes.
[`push-build.sh`](/push-build.sh)       | Pushes a developer build or CI Jenkins build up to GCS.
[`script-template`](/script-template)   | Generates a script template in the kubernetes/release ecosystem.
[`testgridshot`](/testgridshot)         | Screenshots failing testgrid dashboards and creates a markdown stub that can be copied and pasted into a GitHub issue comment.<br/>This makes it easier to create comments like [this][shot-issue] as part of the release process.

For information on how to use `gcbmgr`, `anago` and `branchff`, see the [Branch Manager Handbook]

[kubernetes/kubernetes]: https://git.k8s.io/kubernetes
[Branch Manager Handbook]: https://git.k8s.io/sig-release/release-engineering/role-handbooks/branch-manager.md
[CHANGELOG.md]: https://git.k8s.io/kubernetes/blob/master/CHANGELOG.md
[shot-issue]: https://github.com/kubernetes/sig-release/issues/756#issuecomment-520721968

## Release Notes Gathering

For more extensive build and usage documentation for the `release-notes` tool, see the [documentation](./cmd/release-notes/README.md).

Once the tool is installed, use `-h` or `--help` to see the command usage:

```
$ release-notes -h
Usage of release-notes:
  -end-sha string
        The commit hash to end at
  -format string
        The format for notes output (options: markdown, json) (default "markdown")
  -github-token string
        A personal GitHub access token (required)
  -output string
        The path to the where the release notes will be printed
  -start-sha string
        The commit hash to start at
```

## Building Linux Packages

### For Debian

You can build the deb packages in a Docker container like this:
```
make build-debs
```

The build runs for a while, after it's done you will find the output in `_output/debs`.

### For Fedora, CentOS, Red Hat Enterprise Linux

You can build the rpm packages in a Docker container with:

```
make build-rpms
```

Resulting rpms, and a pre-generated yum repository will be generated in `_output/rpms`.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute.

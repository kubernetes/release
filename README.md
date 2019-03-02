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

This repo contains the tooling and infrastructure to create Kubernetes releases from the [kubernetes/kubernetes](https://github.com/kubernetes/kubernetes) main repository.

## Intro

There are several scripts and helpers in this repository a release
manager will find useful when managing all kinds of releases (alpha,
beta, official, rc) across branches.

## Tools

Most tools in this repo run by default in *mock* mode to allow for ease in
development and testing.

Tools | Description
 :---: | --
[`gcbmgr`](https://github.com/kubernetes/release/blob/master/gcbmgr) | Google Cloud Builder Manager: <br/><br/> This is the main entry point for release managers for producing releases. All release types can be staged and later released using this method.
[`anago`](https://github.com/kubernetes/release/blob/master/anago) | Release Tool: <br/><br/> The main driver for creating staged builds and releases. This is what runs inside GCB after a job is submitted using `gcbmgr`.
[`branchff`](https://github.com/kubernetes/release/blob/master/branchff) | Fast-forward branching helper : <br/><br/> A tool used to pull new patches onto the release branch.
<br/> [`find_green_build`](https://github.com/kubernetes/release/blob/master/find_green_build) <br/><br/> | Asks Jenkins for a good build to use.
<br/> [`release-notes`](https://github.com/kubernetes/release/blob/master/cmd/release-notes) <br/><br/> | Scrape GitHub for release notes. See [Release Notes Gathering](#release-notes-gathering) for more information.
<br/> [`prin`](https://github.com/kubernetes/release/blob/master/prin) <br/><br/> | To show release tags of a particular PR or commit.
<br/> [`changelog-update`](https://github.com/kubernetes/release/blob/master/changelog-update) <br/><br/> | Updates [CHANGELOG.md](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG.md) version entries by rescanning github for text and label changes.
<br/> [`push-build.sh`](https://github.com/kubernetes/release/blob/master/push-build.sh) <br/><br/> | Pushes a developer build or CI Jenkins build up to GCS.
<br/> [`script-template`](https://github.com/kubernetes/release/blob/master/script-template) <br/><br/> | Generates a script template in the kubernetes/release ecosystem.

For information on how to use `gcbmgr`, `anago` and `branchff`, see the [Branch Manager Handbook](https://github.com/kubernetes/sig-release/tree/master/release-team/role-handbooks/branch-manager#branch-manager-handbook)

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
docker build --tag=debian-packager debian
docker run --volume="$(pwd)/debian:/src" debian-packager
```

The build runs for a while, after it's done you will find the output in `debian/bin`.

### For Fedora, CentOS, Red Hat Enterprise Linux

You can build the rpm packages in a Docker container with:

```
cd rpm
./docker-build.sh
```

Resulting rpms, and a pre-generated yum repository will be generated in rpm/output/x86_64.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute.

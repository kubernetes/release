<!-- BEGIN MUNGE: GENERATED_TOC -->
- [Kubernetes Release](#kubernetes-release)
  - [Intro](#intro)
  - [Primary Tools](#primary-tools)
    - [gcbmgr](#gcbmgr)
    - [anago](#anago)
    - [branchff](#branchff)
  - [Common Workflows](#common-workflows)
    - [Alpha release](#alpha-release)
    - [Official release](#official-release)
    - [Release with --nomock](#release-with---nomock)
  - [All Tools](#all-tools)
  - [Release Notes Gathering](#release-notes-gathering)
  - [Building Linux Packages](#building-linux-packages)
    - [For Debian](#for-debian)
    - [For Fedora, CentOS, Red Hat Enterprise Linux](#for-fedora-centos-red-hat-enterprise-linux)
<!-- END MUNGE: GENERATED_TOC -->

# Kubernetes Release

This repo contains the tooling and infrastructure to create Kubernetes releases from the [kubernetes/kubernetes](https://github.com/kubernetes/kubernetes) main repository.

## Intro

There are several scripts and helpers in this repository a release
manager will find useful when managing all kinds of releases (alpha,
beta, official, rc) across branches.

## Primary Tools

Most tools in this repo run by default in *mock* mode to allow for ease in
development and testing.

The tools also include a familiar [\*nix-style man
page](https://github.com/kubernetes/release/blob/master/anago) as the header of the file, accessed via the command-line via `-man`.

Ex.
```
$ gcbmgr -man
```

### gcbmgr

Stage and release via GCB (Google Cloud Builder).  This is the main entry point
for release managers for producing releases in the cloud.  All release types
can be staged and later released using this method.

### anago

The main driver for created staged builds and releases.  This is what runs
inside GCB after a job is submitted using `gcbmgr`.

### branchff

The branch fast-forward tool used between after a new branch as been created.

See the [Playbook](ADD LINK) for more details.

## Common Workflows

### Alpha release

**Stage it**
```
$ gcbmgr stage master
```
**Release it**
(The complete invocation below is output at the end of the staging log)
```
$ gcbmgr release master --buildversion=<staged version>
```

**Announce it**
(The complete invocation below is output at the end of the release log)
```
$ release-notify <release version>
```

### Official release

**Stage it**
```
$ gcbmgr stage release-1.10 --official
```
**Release it**
(The complete invocation below is output at the end of the staging log)
```
$ gcbmgr release release-1.10 --buildversion=<staged version>
```

**Announce it**
(The complete invocation below is output at the end of the release log)
```
$ release-notify <release version>
```

### Release with --nomock

The above example workflows run *mock* versions of the release.  To produce
a fully deployed and announced release, add `--nomock` to the command line.


## All Tools

* [gcbmgr](https://github.com/kubernetes/release/blob/master/gcbmgr) : GCB manager - find status and initiate builds
* [anago](https://github.com/kubernetes/release/blob/master/anago) : Release Tool
* [branchff](https://github.com/kubernetes/release/blob/master/branchff) : Fast-forward branching helper
* [find_green_build](https://github.com/kubernetes/release/blob/master/find_green_build) : Ask Jenkins for a good build to use
* [relnotes](https://github.com/kubernetes/release/blob/master/relnotes) : Scrape github for release notes \(See below for more info\)
* [prin](https://github.com/kubernetes/release/blob/master/prin) : What tags/releases is my PR IN?
* [changelog-update](https://github.com/kubernetes/release/blob/master/changelog-update) : Update CHANGELOG.md version entries by rescanning github for text and label changes
* [push-build.sh](https://github.com/kubernetes/release/blob/master/push-build.sh) : Push a developer (or CI) build up to GCS
* [script-template](https://github.com/kubernetes/release/blob/master/script-template) : Generate a script template in the kubernetes/release ecosystem

## Release Notes Gathering

To build the `release-notes` tool, check out this repo to your `$GOOPATH`:

```
git clone git@github.com:kubernetes/release.git $GOPATH/src/k8s.io/release
```

Run the following from the root of the repository to install dependencies:

```
go get -u github.com/golang/dep/cmd/dep
dep ensure -vendor-only
```

Run the following from the root of the repository to build the `release-notes` binary:

```
go build ./cmd/release-notes
```

Use the `-h` flag for help:

```
./release-notes -h
```

To generate contextualized for a commit range, run:

```
$ export GITHUB_TOKEN=a_github_api_token
$ ./release-notes \
  -start-sha d0a17cb4bbdf608559f257a76acfaa9acb054903 \
	-end-sha 91e7b4fd31fcd3d5f436da26c980becec37ceefe
level=info msg="fetching all commits. this might take a while..."
level=info msg="got the commits, performing rendering"
level=info msg="release notes JSON written to file" path=/var/folders/wp/6fkmvjf11gv18tdprv4g2mk40000gn/T/release-notes-048706664
```

You can then use a variet of tools (such as `jq`) to slice and dice the output:

```
$ cat /var/folders/wp/6fkmvjf11gv18tdprv4g2mk40000gn/T/release-notes-048706664 | jq ".[:3]"
[
  {
    "text": "kubectl convert previous created a list inside of a list.  Now it is only wrapped once."
  },
  {
    "text": "fixes a regression in kube-scheduler to properly load client connection information from a `--config` file that references a kubeconfig file",
    "kinds": [
      "bug"
    ],
    "sigs": [
      "scheduling"
    ]
  },
  {
    "text": "Fixed cleanup of CSI metadata files."
  }
]
$
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

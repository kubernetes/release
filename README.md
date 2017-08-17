Table of Contents
=================
* [Intro](#intro)
* [Instructions (Quick Start)](#instructions-quick-start)
* [Other Tools](#other-tools)
  * [Tools](#tools)
    * [Release Notes Gathering](#release-notes-gathering)

# Kubernetes Release Process

This repo contains the release infrastructure for
[Kubernetes](https://github.com/kubernetes/kubernetes).

## Intro

Live Kubernetes releases are done by the Kubernetes team at Google due to
permissions and other restrictions.  This may expand eventually to allow
other Kubernetes contributors to generate releases.

The current tooling runs by default in *mock* mode and anyone should
be able to run it in this mode to see exactly how the process works.
In *mock* mode all the code paths are followed for a release, but nothing
is pushed to repositories.

Sticking with the ancient Greek theme, the release script is called `anago`.
Anago means, in the context of navigators and shipping:
"to launch out, set sail, put to sea."

Tools in this repository includes a familiar [\*nix-style man
page](https://github.com/kubernetes/release/blob/master/anago) with usage,
process and examples.  The link shows how the self-contained doc/man page
makes up the header of the script itself and the same info is available
on the command-line (or get usage simply by calling the script with no options):

```
$ anago -man
```

The idea is that no external doc updates should be necessary and the
tool itself contains all of the details and instructions and prerequisite
checks needed for anyone to run the tool in at least mock mode.

There is a simple $USER check to ensure that no one but a certain few people can
run the script with --nomock to perform a real release.

## Instructions (Quick Start)

The tool was designed to require minimal inputs.
The only information the tool needs is to know where you want to create a
release with one optional flag `[--official]` \(used on release-\* branches only\).
The [--official] flag indicates that a standard patch release will be created
on a branch.  Without the [--official] flag, a beta would be created.

Try an alpha release:
```
$ anago master
```

Try a beta release on a branch:
```
$ anago release-1.2
```

Try an official release on a branch:
```
$ anago release-1.2 --official
```

Try a beta release on a new branch:
```
$ anago release-9.9
```

Try creating a new branch and beta for an emergency zero-day fix.
See [docs/branching.md](docs/branching.md) for more details.

(The branch name should reflect the branch point/tag. So if branching at the
v9.9.9 tag on the release-9.9 branch, create a release-9.9.9 branch):
```
$ anago release-9.9.9
```

## Live Releases

Anago is currently locked down to only run for a specific set of individuals.
when ```--nomock``` is specified.

Adding that flag to the command-line indicates the release will push
tags and artifacts.  The user is still prompted before a push occurs, however.

## FAQ

### How can I manually specify a build to use when anago (find_green_build) can't automatically locate a green run?

The output from anago (or find_green_build) displays --buildversion in its
output while trying to locate a build.  The user is welcome to pass any
--buildversion value to anago to create a release at a particular hash.

## Other Tools

All standalone scripts have embedded man pages.  Just use `-man` to view or
your favorite editor.

### Tools

* [prin](https://github.com/kubernetes/release/blob/master/prin) : What tags/releases is my PR IN?
* [find_green_build](https://github.com/kubernetes/release/blob/master/find_green_build) : Ask Jenkins for a good build to use
* [script-template](https://github.com/kubernetes/release/blob/master/script-template) : Generate a script template in the kubernetes/release ecosystem
* [relnotes](https://github.com/kubernetes/release/blob/master/relnotes) : Scrape github for release notes \(See below for more info\)
* [branchff](https://github.com/kubernetes/release/blob/master/branchff) : Fast-forward branching helper
* [changelog-update](https://github.com/kubernetes/release/blob/master/changelog-update) : Update CHANGELOG.md version entries by rescanning github for text and label changes
* [push-build.sh](https://github.com/kubernetes/release/blob/master/push-build.sh) : Push a developer (or CI) build up to GCS

### Release Notes Gathering

```
# get details on how to use the tool
$ relnotes -man
$ cd /kubernetes

# Show release notes from the last release on a branch to HEAD
$ relnotes

# Show release notes from the last release on a specific branch to branch HEAD
$ relnotes --branch=release-1.2

# Show release notes between two specific releases
$ relnotes v1.2.0..v1.2.1 --branch=release-1.2
```

Please report *any* [issues](https://github.com/kubernetes/release/issues)
you encounter.

## Building Packages

### For Debian

You can build the deb packages in a Docker container like this:
```
docker build --tag=debian-packager debian
docker run --volume="$(pwd)/debian:/src:z" debian-packager
```

The build runs for a while, after it's done you will find the output in `debian/bin`.

### For Fedora, CentOS, Red Hat Enterprise Linux

You can build the rpm packages in a Docker container with:

```
cd rpm
./docker-build.sh
```

Resulting rpms, and a pre-generated yum repository will be generated in rpm/output/x86_64.


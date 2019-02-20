# Kubernetes Release Notes Generator

This directory contains a tool called `release-notes` and a set of library utilities at which aim to provide a simple and extensible set of tools for fetching, contextualizing, and rendering release notes for the [Kubernetes](https://github.com/kubernetes/kubernetes) repository.

## Install

The simplest way to install the `release-notes` CLI is via `go get`:

```
go get k8s.io/release/cmd/release-notes
```

This will install `release-notes` to `$GOPATH/bin/release-notes`. If you're new to Go, `$GOPATH` default to `~/go`, so look for the binary at `~/go/bin/release-notes`.

## Usage

To generate release notes for a commit range, run:

```
$ export GITHUB_TOKEN=a_github_api_token
$ release-notes \
  -start-sha d0a17cb4bbdf608559f257a76acfaa9acb054903 \
  -end-sha   91e7b4fd31fcd3d5f436da26c980becec37ceefe
level=info msg="fetching all commits. this might take a while..."
level=info msg="got the commits, performing rendering"
level=info msg="release notes markdown written to file" path=/var/folders/wp/6fkmvjf11gv18tdprv4g2mk40000gn/T/release-notes-048706664
```

You can also generate the raw notes data into JSON. You can then use a variety of tools (such as `jq`) to slice and dice the output:

```json
[
  {
    "text": "fixed incorrect OpenAPI schema for CustomResourceDefinition objects",
    "author": "liggitt",
    "author_url": "https://github.com/liggitt",
    "pr_url": "https://github.com/kubernetes/kubernetes/pull/65256",
    "pr_number": 65256,
    "kinds": [
      "bug"
    ],
    "sigs": [
      "api-machinery"
    ]
  }
]
```

## Building From Source

To build the `release-notes` tool, check out this repo to your `$GOPATH`:

```
git clone git@github.com:kubernetes/release.git $GOPATH/src/k8s.io/release
```

Run the following from the root of the repository to build the `release-notes` binary:

```
bazel build //cmd/release-notes
```

Use the `-h` flag for help:

```
./bazel-bin/cmd/release-notes/darwin_amd64_stripped/release-notes -h
```

Install the binary into your path:

```
cp ./bazel-bin/cmd/release-notes/darwin_amd64_stripped/release-notes /usr/local/bin/release-notes
```


## FAQ

### What do generated notes look like?

Check out the rendering of 1.11's release notes [here](https://gist.github.com/marpaia/acfdb889f362195bb683e9e09ce196bc).

### Why formats are supported?

Right now the tool can output release notes in Markdown and JSON.

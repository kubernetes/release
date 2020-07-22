# Kubernetes Release Notes Generator

This directory contains a tool called `release-notes` and a set of library utilities at which aim to provide a simple and extensible set of tools for fetching, contextualizing, and rendering release notes for the [Kubernetes](https://github.com/kubernetes/kubernetes) repository.

## Install

The simplest way to install the `release-notes` CLI is via `go get`:

```
GO111MODULE=on go get k8s.io/release/cmd/release-notes
```

This will install `release-notes` to `$(go env GOPATH)/bin/release-notes`.

## Usage

To generate release notes for a commit range, run:

```bash
$ export GITHUB_TOKEN=a_github_api_token
$ release-notes --start-rev v1.18.1 --end-rev v1.18.2 --branch release-1.18
…
INFO release notes written to file  	format=markdown path=/tmp/release-notes-659889201
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

## Options

| Flag                    | Env Variable    | Default Value       | Required | Description                                                                                                                       |
| ----------------------- | --------------- | ------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------- |
| **GITHUB REPO OPTIONS** |
|                         | GITHUB_TOKEN    |                     | Yes      | A personal GitHub access token                                                                                                    |
| github-org              | GITHUB_ORG      | kubernetes          | Yes      | Name of GitHub organization                                                                                                       |
| github-repo             | GITHUB_REPO     | kubernetes          | Yes      | Name of GitHub repository                                                                                                         |
| required-author         | REQUIRED_AUTHOR | k8s-ci-robot        | Yes      | Only commits from this GitHub user are considered. Set to empty string to include all users                                       |
| branch                  | BRANCH          | master              | Yes      | The GitHub repository branch to scrape                                                                                            |
| repo-path               | REPO_PATH       | /tmp/k8s-repo       | No       | Path to a local Kubernetes repository, used only for tag discovery                                                                |
| start-rev               | START_REV       |                     | No       | The git revision to start at.
| env-rev                 | END_REV         |                     | No       | The git revision to end at.
| discover                | DISCOVER        | none                | No       | The revision discovery mode for automatic revision retrieval (options: none, mergebase-to-latest, patch-to-patch, patch-to-latest, minor-to-minor) |
| release-bucket          | RELEASE_BUCKET  | kubernetes-release  | No       | Specify gs bucket to point to in generated notes (default "kubernetes-release")                                                   |
| release-tars            | RELEASE_TARS    |                     | No       | Directory of tars to sha512 sum for display                                                                                       |
| **OUTPUT OPTIONS**      |
| output                  | OUTPUT          |                     | No       | The path where the release notes will be written                                                                                  |
| format                  | FORMAT          | markdown            | No       | The format for notes output (options: json, markdown)                                                                             |
| go-template             | GO_TEMPLATE     | go-template:default | No       | The go template if `--format=markdown` (options: go-template:default, go-template:inline:<template-string> go-template:<file.template>) |
| release-version         | RELEASE_VERSION |                     | No       | The release version to tag the notes with                                                                                         |
| dependencies            |                 | true                | No       | Add dependency report                                                                                                             |
| **LOG OPTIONS**         |
| debug                   | DEBUG           | false               | No       | Enable debug logging (options: true, false)                                                                                       |

## Building From Source

To build the `release-notes` tool, check out this repo to your `$GOPATH`:

```
git clone git@github.com:kubernetes/release.git $(go env GOPATH)/src/k8s.io/release
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

### What formats are supported?

Right now the tool can output release notes in Markdown and JSON. The tool
also supports arbitrary formats using go-templates. The template has access
to fields in the `Document` struct. For an example, see the default markdown
template (`pkg/notes/internal/template.go`) used to render the stock format.


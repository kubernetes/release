# GitHub Releases to Google Cloud Storage

This directory contains a tool called `gh2gcs` that downloads a release published in GitHub to a Google Cloud Storage Bucket.

## Requirements

This tool directly depends on `gcloud` and `gsutil` to be installed on the system.

Google Cloud has [documentation on installing and configuring the Google Cloud SDK CLI tools](https://cloud.google.com/sdk/docs/quickstarts).

## Install

The simplest way to install the `gh2gcs` CLI is via `go get`:

```
$ go get k8s.io/release/cmd/gh2gcs
```

This will install `gh2gcs` to `$(go env GOPATH)/bin/gh2gcs`.

## Usage

To use this tool to copy a release to a Google Cloud Storage, run:

```bash
$ gh2gcs \
  --org kubernetes --repo kubernetes --bucket my-test-bucket \
  --release-dir release --tags v1.18.0

INFO Validating gh2gcs options...
INFO Downloading assets for the following kubernetes/kubernetes release tags: v1.18.0
INFO Download assets for kubernetes/kubernetes@v1.18.0
INFO Writing assets to /tmp/test/kubernetes/kubernetes/v1.18.0
...

```

## Use case

GitHub has a rate limit on downloading release artifacts. The artifacts that are often used (e.g. in CI) are uploaded to Google Cloud Storage, as GCS doesn't have read rate limits. This improves test stability as flakes related to hitting rate limits are avoided.

## SIG Release managed buckets

The following GCS buckets are managed by SIG Release:

- k8s-artifacts-cni - contains [CNI plugins](https://github.com/containernetworking/plugins) artifacts
- k8s-artifacts-cri-tools - contains [CRI tools](https://github.com/kubernetes-sigs/cri-tools) artifacts (`crictl` and `critest`)

The artifacts are pushed to GCS by [Release Managers](https://github.com/kubernetes/sig-release/blob/master/release-managers.md). The pushing is done manually by running the appropriate `gh2gcs` command. It's recommended for Release Managers to watch the appropriate repositories for new releases.

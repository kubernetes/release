# krel push

Push Kubernetes release artifacts to Google Cloud Storage (GCS) or Amazon S3

- [Summary](#summary)
- [Installation](#installation)
- [Usage](#usage)
  - [Command line flags](#command-line-flags)
  - [Examples](#examples)
- [Important Notes](#important-notes)

## Summary

Used for pushing developer builds and Jenkins' continuous builds.

## Installation

Simply [install krel](README.md#installation).

## Usage

```
  krel push [--noupdatelatest] [--ci] [--bucket=<GS bucket>] [--private-bucket] [flags]
```

Developer pushes simply run as they do pushing to devel/ on GCS.
In `--ci` mode, 'push' runs in mock mode by default. Use `--nomock` to do a real push.

### Command line flags

```
Flags:
      --allow-dup                       Do not exit error if the build already exists on the gcs path
      --bucket string                   Specify an alternate bucket for pushes (normally 'devel' or 'ci'), or an s3://<bucket>[/<prefix>] URL to push to Amazon S3, which stores every '+' in the object keys as a space (default "devel")
      --bucket-root string              Specify an alternate path in the bucket to push artifacts to
      --buildDir string                 Specify an alternate build directory (defaults to '_output') (default "_output")
      --ci                              Used when called from Jenkins (for ci runs)
      --extra-version-markers strings   Comma separated list which can be used to upload additional version files to GCS. The path is relative and is append to a GCS path. (--ci only)
      --fast                            Specifies a fast build (linux/amd64 only)
  -h, --help                            help for push
      --noupdatelatest                  Do not update the latest file
      --private-bucket                  Do not mark published bits on GCS as publicly readable (not supported for S3)
      --registry string                 If set, push docker images to specified registry/project
      --validate-images                 Validate that the remote image digests exists
      --version-suffix string           Append suffix to version name if set

Global Flags:
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warning', 'info', 'debug', 'trace' (default "info")
      --nomock             run the command to target the production environment
```

### Examples

```bash
krel push                                   # Do a developer push
krel push --ci                              # Do a CI push
krel push --nomock --ci                     # Do a non-mocked CI push
krel push --bucket=kubernetes-release-$USER # Do a developer push to kubernetes-release-$USER
krel push --ci --bucket=s3://bucket/prefix  # Do a CI push to s3://bucket/prefix/ci
```

## Important Notes

### Pushing to Amazon S3

An `s3://<bucket>[/<prefix>]` URL as `--bucket` pushes the artifacts to Amazon
S3 in the same layout as on GCS, using the `aws` CLI, which has to be in
`$PATH`. Unlike on GCS:

- The objects get no ACLs, so they are only public if the bucket policy makes
  them public. `--private-bucket` is rejected, use the bucket policy to restrict
  access instead.
- S3 decodes a literal `+` in a URL path as a space. To serve the artifacts of
  CI versions like `v1.36.0-alpha.0.12+0123abc` at the same URL paths as on
  GCS, every `+` in the object keys, including in the prefix, is stored as a
  space, e.g. `ci/v1.36.0-alpha.0.12 0123abc/kubernetes.tar.gz`. The version
  markers keep the `+`, so tools that address the objects by key, like the AWS
  CLI and SDKs, or URLs that encode the `+` as `%2B`, have to replace it with a
  space as well.

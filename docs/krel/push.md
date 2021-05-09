# krel push

Push Kubernetes release artifacts to Google Cloud Storage (GCS)

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
      --bucket string                   Specify an alternate bucket for pushes (normally 'devel' or 'ci') (default "devel")
      --buildDir string                 Specify an alternate build directory (defaults to '_output') (default "_output")
      --ci                              Used when called from Jenkins (for ci runs)
      --extra-version-markers strings   Comma separated list which can be used to upload additional version files to GCS. The path is relative and is append to a GCS path. (--ci only)
      --fast                            Specifies a fast build (linux/amd64 only)
      --gcs-root string                 Specify an alternate GCS path to push artifacts to
  -h, --help                            help for push
      --noupdatelatest                  Do not update the latest file
      --private-bucket                  Do not mark published bits on GCS as publicly readable
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
```

## Important Notes

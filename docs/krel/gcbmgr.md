# gcbmgr (GCB Manager) <!-- omit in toc -->

`gcbmgr` is a subcommand of `krel` for Release Managers to submit Kubernetes staging and release jobs to Google Cloud Build (GCB).

`krel gcbmgr` is a wrapper around the `gcloud builds submit` command which passes a set of GCB substitutions (variables) to a GCB configuration for usage in [`anago`](/anago).

**NOTE: `krel gcbmgr` is currently in development and its design may rapidly change. If you encounter errors, please file an issue in this repo.**

- [Installation](#installation)
- [Usage](#usage)
- [Important notes](#important-notes)
- [List Jobs](#list-jobs)
- [Alpha](#alpha)
  - [Alpha Stage](#alpha-stage)
  - [Alpha Release](#alpha-release)
- [Beta](#beta)
  - [Beta Stage](#beta-stage)
  - [Beta Release](#beta-release)
- [Release Candidate](#release-candidate)
  - [Release Candidate (RC) Stage](#release-candidate-rc-stage)
  - [Release Candidate (RC) Release](#release-candidate-rc-release)
- [Official](#official)
  - [Official Stage](#official-stage)
  - [Official Release](#official-release)

## Installation

Simply [install krel](README.md#installation).

<!-- TODO(vdf): Need to reference K8s Infra projects in usage examples -->

## Usage

`krel gcbmgr [flags]`

```
Flags:
      --branch string          branch to run the specified GCB run against (default "master")
      --build-version string   build version
      --gcb-config string      if specified, this will be used as the name of the Google Cloud Build config file (default "cloudbuild.yaml")
      --gcp-user string        if specified, this will be used as the GCP_USER_TAG
  -h, --help                   help for gcbmgr
      --list-jobs int          list the last N build jobs in the project (default 5)
      --project string         GCP project to run GCB in (default "kubernetes-release-test")
      --release                submit a release run to GCB
      --stage                  submit a stage run to GCB
      --stream                 if specified, GCB will run synchronously, tailing its logs to stdout
      --type string            release type, must be one of: 'alpha', 'beta', 'rc', 'official' (default "alpha")

Global Flags:
      --cleanup            cleanup flag
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace' (default "info")
      --nomock             nomock flag
      --repo string        the local path to the repository to be used (default "/tmp/k8s")
```

## Important notes

- Default executions of `krel gcbmgr` run in mock mode. To run an actual stage or release, you **MUST** provide the `--nomock` flag.
- Note that the default `--branch` is set to `master`, which means that it needs
  to be set to the release branch if necessary.
- Always execute the release process in the following order:
  - mock stage: `krel gcbmgr --stage`
  - mock release: `krel gcbmgr --release`
  - nomock stage: `krel gcbmgr --stage --nomock`
  - nomock release: `krel gcbmgr --release --nomock`
- For release jobs, you **MUST** specify the build version that is output as a result of a successful staging run.
- The following environment variables can be used to test the staging/release process against a contributor's fork: `TOOL_ORG`, `TOOL_REPO`, `TOOL_BRANCH`

  Example:

  ```shell
  TOOL_ORG=justaugustus \
  TOOL_REPO=release \
  TOOL_BRANCH=great-new-feature-branch \
  krel gcbmgr --stage \
    --type beta \
    --branch release-x.y \
    --project kubernetes-release-test
  ```

## List Jobs

```shell
krel gcbmgr \
  --list-jobs 10
```

## Alpha

### Alpha Stage

```shell
krel gcbmgr --stage \
  --type alpha \
  --project kubernetes-release-test
```

### Alpha Release

```shell
krel gcbmgr --release \
  --type alpha \
  --project kubernetes-release-test \
  --build-version <build-version>
```

## Beta

### Beta Stage

```shell
krel gcbmgr --stage \
  --type beta \
  --project kubernetes-release-test
```

### Beta Release

```shell
krel gcbmgr --release \
  --type beta \
  --project kubernetes-release-test \
  --build-version <build-version>
```

## Release Candidate

### Release Candidate (RC) Stage

```shell
krel gcbmgr --stage \
  --type rc \
  --branch release-x.y \
  --project kubernetes-release-test
```

### Release Candidate (RC) Release

```shell
krel gcbmgr --release \
  --type rc \
  --branch release-x.y \
  --project kubernetes-release-test \
  --build-version <build-version>
```

## Official

### Official Stage

```shell
krel gcbmgr --stage \
  --type official \
  --branch release-x.y \
  --project kubernetes-release-test
```

### Official Release

```shell
krel gcbmgr --release \
  --type official \
  --branch release-x.y \
  --project kubernetes-release-test \
  --build-version <build-version>
```

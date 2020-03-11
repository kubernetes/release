# gcbmgr (GCB Manager) <!-- omit in toc -->

`gcbmgr` is a subcommand of `krel` for Release Managers to submit Kubernetes staging and release jobs to Google Cloud Build (GCB).

`krel gcbmgr` is a wrapper around the `gcloud builds submit` command which passes a set of GCB substitutions (variables) to a GCB configuration for usage in [`anago`](/anago).

**NOTE: `krel gcbmgr` is currently in development and its design may rapidly change. If you encounter errors, please file an issue in this repo.**

- [Installation](#installation)
- [Usage](#usage)
- [Important notes](#important-notes)
- [Alpha](#alpha)
  - [Alpha Stage](#alpha-stage)
  - [Alpha Release](#alpha-release)
- [Beta](#beta)
  - [Beta Stage](#beta-stage)
    - [Branch cut (`x.y.0-beta.0`)](#branch-cut-xy0-beta0)
    - [Post-branch cut (`x.y.0-beta.1` and beyond)](#post-branch-cut-xy0-beta1-and-beyond)
  - [Beta Release](#beta-release)
- [Release Candidate](#release-candidate)
  - [Release Candidate (RC) Stage](#release-candidate-rc-stage)
  - [Release Candidate (RC) Release](#release-candidate-rc-release)
- [Official](#official)
  - [Official Stage](#official-stage)
  - [Official Release](#official-release)

## Installation

From this root of this repo:

```shell
./compile-release-tools krel
```

<!-- TODO(vdf): Need to reference K8s Infra projects in usage examples -->
## Usage

`krel gcbmgr [flags]`

```
Flags:
      --branch string          Branch to run the specified GCB run against
      --build-version string   Build version
      --gcb-config string      If provided, this will be used as the name of the Google Cloud Build config file. (default "cloudbuild.yaml")
      --gcp-user string        If provided, this will be used as the GCP_USER_TAG.
  -h, --help                   help for gcbmgr
      --project string         GCP project to run GCB in (default "kubernetes-release-test")
      --release                Submit a release run to GCB
      --stage                  Submit a stage run to GCB
      --stream                 If specified, GCB will run synchronously, tailing its' logs to stdout
      --type string            Release type (must be one of: 'prerelease', 'rc', 'official') (default "prerelease")

Global Flags:
      --cleanup            cleanup flag
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace' (default "info")
      --nomock             nomock flag
      --repo string        the local path to the repository to be used (default "/tmp/k8s")
```

## Important notes

- Default executions of `krel gcbmgr` run in mock mode. To run an actual stage or release, you **MUST** provide the `--nomock` flag.
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
    --branch release-x.y \
    --project kubernetes-release-test
  ```

## Alpha

### Alpha Stage

```shell
krel gcbmgr --stage \
  --branch master \
  --project kubernetes-release-test
```

### Alpha Release

```shell
krel gcbmgr --release \
  --branch master \
  --project kubernetes-release-test \
  --build-version <build-version>
```

## Beta

### Beta Stage

#### Branch cut (`x.y.0-beta.0`)

```shell
krel gcbmgr --stage \
  --branch release-x.y \
  --project kubernetes-release-test
```

#### Post-branch cut (`x.y.0-beta.1` and beyond)

```shell
krel gcbmgr --stage \
  --branch release-x.y \
  --project kubernetes-release-test
```

### Beta Release

```shell
krel gcbmgr --release \
  --branch release-x.y \
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

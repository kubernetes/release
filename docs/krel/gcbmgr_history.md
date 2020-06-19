# history <!-- omit in toc -->

`history` is a subcommand of `krel gcbmgr` for Release Managers to build a markdown table that contains all the jobs and necessary information to complete the Cut Release GitHub Issue.

`krel gcbmgr history` collect all the jobs that run in `google cloudbuild` for a specific release in a specific date. The intention of this subcommand is to automate the tasks of filling the data in the GitHub Issue by building the entire table when all the jobs ran, instead of doing one by one manually.

- [Installation](#installation)
- [Usage](#usage)

## Installation

Simply [install krel](README.md#installation).


## Usage

`krel gcbmgr history [flags]`

```
Flags:
      --date-from string   Get the jobs starting from a specific date. Format to use yyyy-mm-dd
      --date-to string     Get the jobs ending from a specific date. Format to use yyyy-mm-dd
  -h, --help               help for generate

Global Flags:
      --branch string          branch to run the specified GCB run against (default "master")
      --build-version string   build version
      --cleanup                cleanup flag
      --gcb-config string      if specified, this will be used as the name of the Google Cloud Build config file (default "cloudbuild.yaml")
      --gcp-user string        if specified, this will be used as the GCP_USER_TAG
      --list-jobs int          list the last N build jobs in the project (default 5)
      --log-level string       the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace' (default "info")
      --nomock                 nomock flag
      --project string         GCP project to run GCB in (default "kubernetes-release-test")
      --release                submit a release run to GCB
      --repo string            the local path to the repository to be used (default "/var/folders/qk/9n4vv2sx1vj1rl9gq2tc829r0000gn/T/k8s")
      --stage                  submit a stage run to GCB
      --stream                 if specified, GCB will run synchronously, tailing its logs to stdout
      --type string            release type, must be one of: 'alpha', 'beta', 'rc', 'official' (default "alpha")
```

```
$ krel gcbmgr history --date-from 2020-06-17 --branch release-1.16

|      STEP      |                                                        COMMAND                                                        |                                                     LINK                                                      |             START              | DURATION | SUCCEEDED? |
|----------------|-----------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|--------------------------------|----------|------------|
| `MOCK STAGE`   | `krel gcbmgr --stage --type official --branch release-1.16 --buildversion=v1.16.11-rc.0.46+1a6b7d0d26472f`            | https://console.cloud.google.com/cloud-build/builds/9b1ffe44-14fc-4051-9fe2-91583b8edf8e?project=648026197307 | 2020-06-17T09:23:58.519142310Z | 01:33:22 | SUCCESS    |
| `MOCK RELEASE` | `krel gcbmgr --release --type official --branch release-1.16 --buildversion=v1.16.11-rc.0.46+1a6b7d0d26472f`          | https://console.cloud.google.com/cloud-build/builds/3d039388-ec45-487c-8a41-e621b7ed70e7?project=648026197307 | 2020-06-17T11:04:28.710554932Z | 00:29:08 | SUCCESS    |
| `STAGE`        | `krel gcbmgr --stage --type official --branch release-1.16 --buildversion=v1.16.11-rc.0.46+1a6b7d0d26472f --nomock`   | https://console.cloud.google.com/cloud-build/builds/335ea874-2eef-454a-874c-1f1c9c8b50e5?project=648026197307 | 2020-06-17T11:38:07.062174282Z | 01:29:32 | SUCCESS    |
| `RELEASE`      | `krel gcbmgr --release --type official --branch release-1.16 --buildversion=v1.16.11-rc.0.46+1a6b7d0d26472f --nomock` | https://console.cloud.google.com/cloud-build/builds/2100825a-0f95-4dff-9022-1119c236593e?project=648026197307 | 2020-06-17T15:34:43.000190228Z | 00:32:21 | SUCCESS    |
```

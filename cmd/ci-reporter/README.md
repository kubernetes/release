# CI signal report

You can get the current overview for CI signal report by running

```
GITHUB_TOKEN=xxx go run cmd/ci-reporter/main.go
```

It needs a GitHub token to be able to query the project board for CI signal. Make sure to check `read:org` under `admin:org` for the permissions when creating the token. For some reason even though those boards are available for public view, the APIs require auth. See [this documentation](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line) to set up your access token.

## Prerequisites

-   GoLang >=1.16

## Run the report

```
git clone git@github.com:kubernetes/release <folder>
cd <folder>
GITHUB_TOKEN=xxx go run cmd/ci-reporter/main.go
```

### Flags and Commands

```bash
$ go run cmd/ci-reporter/main.go --help
```

```bash
CI-Signal reporter that generates github and testgrid reports.

Usage:
  reporter [flags]
  reporter [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  github      Github report generator
  help        Help about any command
  testgrid    Testgrid report generator

Flags:
  -f, --file string              Specify a filepath to write the report to a file
  -h, --help                     help for reporter
      --json                     Report output in json format
  -v, --release-version string   Specify a Kubernetes release versions like '1.22' which will populate the report additionally
  -s, --short                    A short report for mails and slack
```

### Command for generating the weekly Ci Signal Report 

Replace `-v 1.25` with the current release cycle. 

```bash
$ go run cmd/ci-reporter/main.go -s -v 1.25
```

## Rate limits

GitHub API has rate limits, to see how much you have used you can query like this (replace User with your GH user and Token with your Auth Token):

```
curl \
  -u GIT_HUB-USER:GIT_HUB_TOKEN -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/rate_limit & curl \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/rate_limit
```

## Example table output

```
GITHUB_TOKEN=xxx go run cmd/ci-reporter/main.go -s
GITHUB REPORT

|  TESTGRID BOARD  |               TITLE                |  STATUS   |                                 STATUS DETAILS                                  |
|------------------|------------------------------------|-----------|---------------------------------------------------------------------------------|
| master-blocking  | [Flaky Tests] Various              | OBSERVING | kind/flake,sig/testing,needs-triage                                             |
|                  | sig-release-master-blocking        |           |                                                                                 |
|                  | jobs with "Unknown: Build is       |           |                                                                                 |
|                  | too old to process"                |           |                                                                                 |
| master-informing | [Failing Test]                     | FLAKY     | sig/cluster-lifecycle,kind/flake,needs-triage                                   |
|                  | periodic-conformance-main-k8s-main |           |                                                                                 |
| master-informing | [Flaking Test]                     | FLAKY     | sig/storage,kind/flake,triage/accepted                                          |
|                  | gce-cos-master-serial              |           |                                                                                 |
| master-informing | [Flaky Test] [sig-windows]         | FLAKY     | priority/important-soon,kind/flake,sig/windows,lifecycle/rotten,triage/accepted |
|                  | Services should be able to         |           |                                                                                 |
|                  | create a functioning NodePort      |           |                                                                                 |
|                  | service for Windows                |           |                                                                                 |
| master-informing | [Flaky] HostProcess containers     | OBSERVING | kind/flake,sig/windows,needs-triage                                             |
|                  | container stats validation         |           |                                                                                 |

SUMMARY - Total:5 OBSERVING:2 FLAKY:3 

TESTGRID REPORT

|        TESTGRID BOARD        |                    TITLE                    | STATUS | STATUS DETAILS  |
|------------------------------|---------------------------------------------|--------|-----------------|
| sig-release-master-blocking  | gce-cos-master-scalability-100              | FLAKY  | 8 of 9 (88.9%)  |
| sig-release-master-blocking  | integration-master                          | FLAKY  | 7 of 9 (77.8%)  |
| sig-release-master-blocking  | ci-kubernetes-unit                          | FLAKY  | 8 of 10 (80.0%) |
| sig-release-master-blocking  | gce-cos-master-reboot                       | FLAKY  | 8 of 9 (88.9%)  |
| sig-release-master-blocking  | kind-master-parallel                        | FLAKY  | 8 of 9 (88.9%)  |
| sig-release-master-blocking  | build-master                                | FLAKY  | 8 of 9 (88.9%)  |
| sig-release-master-informing | gce-master-scale-correctness                | FLAKY  | 7 of 10 (70.0%) |
| sig-release-master-informing | gce-master-scale-performance                | FLAKY  | 9 of 10 (90.0%) |
| sig-release-master-informing | post-kubernetes-push-image-etcd             | FLAKY  | 1 of 2 (50.0%)  |
| sig-release-master-informing | capz-windows-containerd-master              | FLAKY  | 8 of 10 (80.0%) |
| sig-release-master-informing | gce-cos-master-serial                       | FLAKY  | 5 of 9 (55.6%)  |
| sig-release-master-informing | kubeadm-kinder-latest                       | FLAKY  | 9 of 10 (90.0%) |
| sig-release-master-informing | periodic-conformance-main-k8s-main          | FLAKY  | 4 of 10 (40.0%) |
| sig-release-master-informing | post-release-push-image-distroless-iptables | FLAKY  | 2 of 3 (66.7%)  |
| sig-release-master-informing | post-release-push-image-go-runner           | FLAKY  | 1 of 2 (50.0%)  |
| sig-release-master-informing | capg-conformance-main-ci-artifacts          | FLAKY  | 8 of 9 (88.9%)  |
| sig-release-master-informing | post-release-push-image-debian-iptables     | FLAKY  | 3 of 4 (75.0%)  |
| sig-release-master-informing | post-release-push-image-kube-cross          | FLAKY  | 1 of 3 (33.3%)  |

SUMMARY - Total:18 FLAKY:18 
```

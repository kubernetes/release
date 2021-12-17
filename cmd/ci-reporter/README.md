# CI signal report

You can get the current overview for CI signal report by running

```
GITHUB_TOKEN=xxx go run cmd/ci-reporter/main.go
```

It needs a GitHub token to be able to query the project board for CI signal. For some reason even though those boards are available for public view, the APIs require auth. See [this documentation](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line) to set up your access token.

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
  completion  generate the autocompletion script for the specified shell
  github      Github report generator
  help        Help about any command
  testgrid    Testgrid report generator

Flags:
  -h, --help                     help for reporter
      --json                     Report output in json format
  -v, --release-version string   Specify a Kubernetes release versions like '1.22' which will populate the report additionally
  -s, --short                    A short report for mails and slack

Use "reporter [command] --help" for more information about a command.
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
     ID    |                  TITLE                  |       CATEGORY        | STATUS
-----------+-----------------------------------------+-----------------------+---------
    100230 | [Flaky Test]                            | New/Not Yet Started   |
           | [sig-cloud-provider-gcp] Nodes          |                       |
           | [Disruptive] Resize [Slow]              |                       |
           | should be able to delete nodes          |                       |
    105677 | HPA Custom metrics tests are            |                       |
           | failing                                 |                       |
    106278 | New Windows kubelet stats               | In flight             |
           | collection test flaking                 |                       |
     98180 | [Flaky Test] [sig-apps]                 |                       |
           | Deployment should run the               |                       |
           | lifecycle of a Deployment               |                       |
     97783 | Device manager for Windows              |                       |
           | passes when run on cluster              |                       |
           | that does not have a GPU but            |                       |
           | cuases cascading errors                 |                       |
    103742 | [Flaking Test] [sig-scalability]        |                       |
           | restarting konnectivity-agent           |                       |
           | (ci-kubernetes-e2e-gci-gce-scalability) |                       |
-----------+-----------------------------------------+-----------------------+---------
  TOTAL: 6 |                                           NEW/NOT YET STARTED:2 |
           |                                                IN FLIGHT:4      |
-----------+-----------------------------------------+-----------------------+---------

TESTGRID REPORT
               ID              |                  TITLE                  | CATEGORY  |      STATUS
-------------------------------+-----------------------------------------+-----------+--------------------
  sig-release-master-blocking  | ci-kubernetes-unit                      | FLAKY     | 8 of 9 (88.9%)
                               | gce-device-plugin-gpu-master            | FAILING   | 0 of 10 (0.0%)
                               | gci-gce-ingress                         | FLAKY     | 8 of 10 (80.0%)
                               | kind-master-parallel                    |           | 8 of 9 (88.9%)
                               | integration-master                      |           |
                               | verify-master                           |           | 9 of 9 (100.0%)
  sig-release-master-informing | post-release-push-image-debian-base     |           | 0 of 1 (0.0%)
                               | post-release-push-image-setcap          |           |
                               | capg-conformance-main-ci-artifacts      |           | 10 of 10 (100.0%)
                               | capg-conformance-v1beta1-ci-artifacts   |           | 8 of 10 (80.0%)
                               | ci-crio-cgroupv1-node-e2e-conformance   |           | 9 of 10 (90.0%)
                               | gce-ubuntu-master-default               | FAILING   | 0 of 9 (0.0%)
                               | kubeadm-kinder-upgrade-1-23-latest      | FLAKY     | 8 of 9 (88.9%)
                               | post-release-push-image-debian-iptables |           | 0 of 1 (0.0%)
                               | aks-engine-windows-containerd-master    |           | 7 of 10 (70.0%)
                               | post-release-push-image-go-runner       |           | 4 of 9 (44.4%)
                               | gce-cos-master-slow                     |           | 8 of 10 (80.0%)
                               | gce-master-scale-performance            |           | 9 of 10 (90.0%)
                               | periodic-conformance-main-k8s-main      | FAILING   | 8 of 10 (80.0%)
-------------------------------+-----------------------------------------+-----------+--------------------
           TOTAL: 19           |                                           FLAKY:16  |
                               |                                           FAILING:3 |
                               |                                                     |
-------------------------------+-----------------------------------------+-----------+--------------------
```

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
     ID    |                  TITLE                  |       CATEGORY        | STATUS |              SIGS              |                          URL                           | TS
-----------+-----------------------------------------+-----------------------+--------+--------------------------------+--------------------------------------------------------+-----
    106278 | New Windows kubelet stats               | In flight             |        | [sig windows]                  | https://github.com/kubernetes/kubernetes/issues/106278 |
           | collection test flaking                 |                       |        |                                |                                                        |
     98180 | [Flaky Test] [sig-apps]                 |                       |        | [sig apps]                     | https://github.com/kubernetes/kubernetes/issues/98180  |
           | Deployment should run the               |                       |        |                                |                                                        |
           | lifecycle of a Deployment               |                       |        |                                |                                                        |
     97783 | Device manager for Windows              |                       |        | [sig windows]                  | https://github.com/kubernetes/kubernetes/issues/97783  |
           | passes when run on cluster              |                       |        |                                |                                                        |
           | that does not have a GPU but            |                       |        |                                |                                                        |
           | cuases cascading errors                 |                       |        |                                |                                                        |
     97071 | [Flaky test] [sig-storage]              |                       |        | [sig scalability sig storage]  | https://github.com/kubernetes/kubernetes/issues/97071  |
           | In-tree Volumes [Driver:                |                       |        |                                |                                                        |
           | gcepd] [Testpattern:                    |                       |        |                                |                                                        |
           | Pre-provisioned PV                      |                       |        |                                |                                                        |
           | (xfs)][Slow] volumes should             |                       |        |                                |                                                        |
           | store data                              |                       |        |                                |                                                        |
    103742 | [Flaking Test] [sig-scalability]        |                       |        | [sig scalability sig           | https://github.com/kubernetes/kubernetes/issues/103742 |
           | restarting konnectivity-agent           |                       |        | cloud-provider]                |                                                        |
           | (ci-kubernetes-e2e-gci-gce-scalability) |                       |        |                                |                                                        |
     93740 | [Flaky Test][sig-network]               |                       |        | [sig network]                  | https://github.com/kubernetes/kubernetes/issues/93740  |
           | Loadbalancing: L7 GCE [Slow]            |                       |        |                                |                                                        |
           | [Feature:Ingress] should                |                       |        |                                |                                                        |
           | conform to Ingress spec                 |                       |        |                                |                                                        |
    100230 | [Flaky Test]                            | New/Not Yet Started   |        | [sig cloud-provider]           | https://github.com/kubernetes/kubernetes/issues/100230 |
           | [sig-cloud-provider-gcp] Nodes          |                       |        |                                |                                                        |
           | [Disruptive] Resize [Slow]              |                       |        |                                |                                                        |
           | should be able to delete nodes          |                       |        |                                |                                                        |
    105677 | HPA Custom metrics tests are            |                       |        | [sig autoscaling sig testing]  | https://github.com/kubernetes/kubernetes/issues/105677 |
           | failing                                 |                       |        |                                |                                                        |
-----------+-----------------------------------------+-----------------------+--------+--------------------------------+--------------------------------------------------------+-----
  TOTAL: 8 |                                            IN FLIGHT:6 NEW/NOT  |
           |                                               YET STARTED:2     |
-----------+-----------------------------------------+-----------------------+--------+--------------------------------+--------------------------------------------------------+-----
TESTGRID REPORT
                    ID                    |            TITLE             | CATEGORY  |      STATUS       | SIGS |                                             URL                                              |           TS
------------------------------------------+------------------------------+-----------+-------------------+------+----------------------------------------------------------------------------------------------+--------------------------
  gce-ubuntu-master-default               | sig-release-master-informing | FAILING   | 0 of 10 (0.0%)    | []   | https://testgrid.k8s.io/sig-release-master-informing#gce-ubuntu-master-default               | 2021-12-16 19:32:53 CET
  post-release-push-image-setcap          |                              | FLAKY     | 0 of 1 (0.0%)     |      | https://testgrid.k8s.io/sig-release-master-informing#post-release-push-image-setcap          | 2021-11-16 20:53:34 CET
  capg-conformance-main-ci-artifacts      |                              |           | 10 of 10 (100.0%) |      | https://testgrid.k8s.io/sig-release-master-informing#capg-conformance-main-ci-artifacts      | 2021-12-16 12:59:28 CET
  ci-crio-cgroupv1-node-e2e-conformance   |                              |           | 9 of 10 (90.0%)   |      | https://testgrid.k8s.io/sig-release-master-informing#ci-crio-cgroupv1-node-e2e-conformance   | 2021-12-16 19:06:16 CET
  gce-master-scale-performance            |                              |           | 8 of 9 (88.9%)    |      | https://testgrid.k8s.io/sig-release-master-informing#gce-master-scale-performance            | 2021-12-16 18:01:16 CET
  post-release-push-image-go-runner       |                              |           | 4 of 9 (44.4%)    |      | https://testgrid.k8s.io/sig-release-master-informing#post-release-push-image-go-runner       | 2021-12-12 14:52:01 CET
  periodic-conformance-main-k8s-main      |                              | FAILING   | 8 of 10 (80.0%)   |      | https://testgrid.k8s.io/sig-release-master-informing#periodic-conformance-main-k8s-main      | 2021-12-16 15:31:06 CET
  post-release-push-image-debian-base     |                              | FLAKY     | 0 of 1 (0.0%)     |      | https://testgrid.k8s.io/sig-release-master-informing#post-release-push-image-debian-base     | 2021-11-16 20:53:34 CET
  post-release-push-image-debian-iptables |                              |           |                   |      | https://testgrid.k8s.io/sig-release-master-informing#post-release-push-image-debian-iptables |
  capg-conformance-v1beta1-ci-artifacts   |                              |           | 9 of 10 (90.0%)   |      | https://testgrid.k8s.io/sig-release-master-informing#capg-conformance-v1beta1-ci-artifacts   | 2021-12-16 12:59:15 CET
  gce-cos-master-slow                     |                              |           | 7 of 9 (77.8%)    |      | https://testgrid.k8s.io/sig-release-master-informing#gce-cos-master-slow                     | 2021-12-16 19:48:46 CET
  kubeadm-kinder-upgrade-1-23-latest      |                              |           | 8 of 9 (88.9%)    |      | https://testgrid.k8s.io/sig-release-master-informing#kubeadm-kinder-upgrade-1-23-latest      | 2021-12-16 19:41:46 CET
  kubeadm-kinder-latest                   |                              |           |                   |      | https://testgrid.k8s.io/sig-release-master-informing#kubeadm-kinder-latest                   | 2021-12-16 19:42:46 CET
  aks-engine-windows-containerd-master    |                              |           | 4 of 9 (44.4%)    |      | https://testgrid.k8s.io/sig-release-master-informing#aks-engine-windows-containerd-master    | 2021-12-16 18:55:16 CET
  verify-master                           | sig-release-master-blocking  |           | 7 of 9 (77.8%)    |      | https://testgrid.k8s.io/sig-release-master-blocking#verify-master                            | 2021-12-16 18:45:16 CET
  gce-device-plugin-gpu-master            |                              | FAILING   | 0 of 10 (0.0%)    |      | https://testgrid.k8s.io/sig-release-master-blocking#gce-device-plugin-gpu-master             | 2021-12-16 18:30:16 CET
  kind-ipv6-master-parallel               |                              | FLAKY     | 9 of 10 (90.0%)   |      | https://testgrid.k8s.io/sig-release-master-blocking#kind-ipv6-master-parallel                | 2021-12-16 19:08:16 CET
  ci-kubernetes-unit                      |                              |           | 8 of 10 (80.0%)   |      | https://testgrid.k8s.io/sig-release-master-blocking#ci-kubernetes-unit                       | 2021-12-16 19:14:16 CET
------------------------------------------+------------------------------+-----------+-------------------+------+----------------------------------------------------------------------------------------------+--------------------------
                 TOTAL: 18                |                                FAILING:3 |
                                          |                                FLAKY:15  |
------------------------------------------+------------------------------+-----------+-------------------+------+----------------------------------------------------------------------------------------------+--------------------------

```

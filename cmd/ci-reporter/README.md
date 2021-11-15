# CI signal report

You can get the current overview for CI signal report by running

```
GITHUB_TOKEN=xxx go run cmd/ci-reporter/main.go
```

It needs a GitHub token to be able to query the project board for CI signal. For some reason even though those boards are available for public view, the APIs require auth. See [this documentation](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line) to set up your access token.

## Prerequisites

- GoLang >=1.16

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

Flags:
  -h, --help                     help for reporter
  -v, --release-version string   Specify a Kubernetes release versions like '1.22' which will populate the report additionally
  -s, --short                    A short report for mails and slack
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

## Example output

```
GITHUB_TOKEN=xxx go run cmd/ci-reporter/main.go -s

In flight
#106278 [sig windows]
 - New Windows kubelet stats collection test flaking
 - https://github.com/kubernetes/kubernetes/issues/106278
#98180 [sig apps]
 - [Flaky Test] [sig-apps] Deployment should run the lifecycle of a Deployment
 - https://github.com/kubernetes/kubernetes/issues/98180
#97783 [sig windows]
 - Device manager for Windows passes when run on cluster that does not have a GPU but cuases cascading errors
 - https://github.com/kubernetes/kubernetes/issues/97783
#97071 [sig scalability sig storage]
 - [Flaky test] [sig-storage] In-tree Volumes [Driver: gcepd] [Testpattern: Pre-provisioned PV (xfs)][Slow] volumes should store data
 - https://github.com/kubernetes/kubernetes/issues/97071

New/Not Yet Started
#100230 [sig cloud-provider]
 - [Flaky Test] [sig-cloud-provider-gcp] Nodes [Disruptive] Resize [Slow] should be able to delete nodes
 - https://github.com/kubernetes/kubernetes/issues/100230
#105677 [sig autoscaling sig testing]
 - HPA Custom metrics tests are failing
 - https://github.com/kubernetes/kubernetes/issues/105677


Failures in Master-Blocking
	18 jobs total
	13 are passing
	5 are flaking
	0 are failing
	0 are stale


Failures in Master-Informing
	23 jobs total
	11 are passing
	8 are flaking
	4 are failing
	0 are stale

```

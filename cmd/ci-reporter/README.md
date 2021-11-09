# CI signal report

You can get the current overview for CI signal report by running

```
GITHUB_AUTH_TOKEN=xxx go run report.go
```

It needs a GitHub token to be able to query the project board for CI signal. For some reason even though those boards are available for public view, the APIs require auth. See [this documentation](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line) to set up your access token.

## Prerequisites
- GoLang >=1.16

## Run the report
```
git clone git@github.com:alenkacz/ci-signal-report.git <folder>
cd <folder>
GITHUB_AUTH_TOKEN=xxx go run report.go
```

### Other version statistics
By adding `RELEASE_VERSION=xxx` where the XXX can be like `1.21`, the report statistics get extended for the choosen version.

```
GITHUB_AUTH_TOKEN=xxx RELEASE_VERSION=xxx go run report.go
```

### Short report
You can also output a short version of the report with the flag `-short`. This reduces the report to `New/Not Yet Started` and `In Flight` issues.

```
GITHUB_AUTH_TOKEN=xxx go run report.go -short
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
GITHUB_AUTH_TOKEN=yourFavouriteGitHubTokenLivesHere RELEASE_VERSION=1.21 go run report.go -short

New/Not Yet Started
SIG network
#80719 https://api.github.com/repos/kubernetes/kubernetes/issues/80719 [sig-network] Services should only allow access from service loadbalancer source ranges [Slow]

SIG storage
#80717 https://api.github.com/repos/kubernetes/kubernetes/issues/80717  [sig-storage] CSI Volumes [Driver: csi-hostpath] Snapshot Tests

In flight
SIG testing
#79662 https://api.github.com/repos/kubernetes/kubernetes/issues/79662   Nodes resize test failing in master-blocking

SIG cluster-lifecycle
#78907 https://api.github.com/repos/kubernetes/kubernetes/issues/78907 [Flaky Tests] task-06-upgrade is failing on master-informing

SIG scheduling
#74931 https://api.github.com/repos/kubernetes/kubernetes/issues/74931 Scheduler TestPreemptionRaces is flaky

Observing
SIG apps
#79740 https://api.github.com/repos/kubernetes/kubernetes/issues/79740  Test Deployment deployment should support rollback is failing on master informing

SIG cli
#79533 https://api.github.com/repos/kubernetes/kubernetes/issues/79533  Kubectl client Conformance test failing

Resolved
SIG cluster-lifecycle
#80434 https://api.github.com/repos/kubernetes/kubernetes/issues/80434  Errors bringing up kube-proxy in CI

Failures in Master-Blocking
    14 jobs total
    9 are passing
    2 are flaking
    3 are failing
    0 are stale

Failures in Master-Informing
    14 jobs total
    7 are passing
    4 are flaking
    3 are failing
    0 are stale

Failures in 1.21-Blocking
    14 jobs total
    9 are passing
    2 are flaking
    3 are failing
    0 are stale

Failures in 1.21-Informing
    14 jobs total
    7 are passing
    4 are flaking
    3 are failing
    0 are stale
 ```

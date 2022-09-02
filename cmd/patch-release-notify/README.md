# Patch Release Notify

This simple tool has the objective to send an notification email when we are closer to
the patch release cycle to let people know that the cherry pick deadline is approaching.

## Install

The simplest way to install the `patch-release-notify` CLI is via `go get`:

```
$ go get k8s.io/release/cmd/patch-release-notify
```

This will install `patch-release-notify` to `$(go env GOPATH)/bin/patch-release-notify`.

Also if you have the `kubernetes/release` cloned you can run the `make release-tools` to build all the tools.

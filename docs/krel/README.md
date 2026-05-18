# krel — The Kubernetes Release Toolbox

`krel` is the new golang based tool for managing releases.

- [Summary](#summary)
- [Installation](#installation)
- [Usage:](#usage)
  - [Available Commands:](#available-commands)
- [Important Notes](#important-notes)

## Summary

The purpose of krel is to provide a toolkit for managing the different steps needed to create
Kubernetes Releases. This includes manually executed tasks like generating the Release Notes during the release cycle and performing automated tasks like pushing the Kubernetes release artifacts to Google Cloud Storage.

## Installation

Choose one of the following options:

**Repo-local (binary in `./bin/krel`):**

```shell
./hack/get-krel
```

Use `./bin/krel` or add `./bin` to your `PATH`.

**Build from source (binary in `./bin/krel`):**

```shell
./compile-release-tools krel
```

**Go install (binary in your Go bin directory):**

```shell
go install k8s.io/release/cmd/krel@latest
```

This installs to `$(go env GOBIN)` (or `$(go env GOPATH)/bin` when `GOBIN` is unset).
Add that directory to your `PATH` if needed:

```shell
export PATH="$PATH:$(go env GOBIN):$(go env GOPATH)/bin"
```

Note: Examples below assume `krel` is on your `PATH`. For repo-local installs, use `./bin/krel`.
## Usage:

krel has several subcommands that perform various tasks during the release lifecycle:

`krel [subcommand]`

### Available Commands:

| Subcommand                          | Description                                                                                 |
| ----------------------------------- | --------------------------------------------------------------------------------------------|
| announce                            | Build and announce Kubernetes releases                                                      |
| ci-build                            | Build Kubernetes in CI and push release artifacts to Google Cloud Storage (GCS)             |
| cve                                 | Add and edit CVE information                                                                |
| [ff](ff.md)                         | Fast forward a Kubernetes release branch                                                    |
| history                             | Run history to build a list of commands that ran when cutting a specific Kubernetes release |
| [push](push.md)                     | Push Kubernetes release artifacts to Google Cloud Storage (GCS)                             |
| release                             | Release a staged Kubernetes version                                                         |
| [release-notes](release-notes.md)   | The subcommand of choice for the Release Notes subteam of SIG Release                       |
| stage                               | Stage a new Kubernetes version                                                              |
| testgridshot                        | Take a screenshot of the testgrid dashboards                                                |

## Sending Release Announcements

The `krel announce send` command sends release announcement emails via the
Gmail API using Google OAuth. No additional setup is required — a browser
window will open for authorization on each run.

In mock mode (default), emails are sent to
[kubernetes-announce-test](https://groups.google.com/g/kubernetes-announce-test).
In production mode (`--nomock`), emails are sent to
[kubernetes-announce](https://groups.google.com/g/kubernetes-announce) and
[dev@kubernetes.io](https://groups.google.com/a/kubernetes.io/g/dev).

```shell
# Mock run (sends to kubernetes-announce-test only):
krel announce send --tag v1.35.1

# Production run (sends to kubernetes-announce and dev):
krel announce send --tag v1.35.1 --nomock

# Print announcement content without sending:
krel announce send --tag v1.35.1 --print-only
```

### Headless environments (`--no-browser`)

If you are running `krel` in a headless environment (e.g. SSH session), use the
`--no-browser` flag:

```shell
krel announce send --tag v1.35.1 --no-browser
```

This prints the OAuth authorization URL to the terminal instead of opening a
browser. Open the URL in any browser (can be on another machine), authorize the
application, and the browser will redirect to `http://localhost?code=...`. Since
no local server is listening, the page will fail to load — this is expected.
Copy the full URL from the browser's address bar and paste it back into the
terminal to complete authentication.

The OAuth app is managed in the
[k8s-release](https://console.cloud.google.com/auth/clients?project=k8s-release)
Google Cloud project.

## Important Notes

Some of the krel subcommands are under development and their usage may already differ from these docs.

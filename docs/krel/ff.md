# krel ff

Fast forward a Kubernetes release branch

- [Summary](#summary)
- [Installation](#installation)
- [Usage](#usage)
- [Important notes](#important-notes)

## Summary

`ff` fast forwards a branch to a specified git object (defaults to origin/master).

`krel ff` pre-checks that the local branch to be forwarded is an actual
'release-x.y' branch and that the branch exists remotely. If that is not the
case, `krel ff` will fail.

After that preflight-check, the release branch will be checked out and krel
verifies that the latest merge base tag is the same for the master and the
release branch. This means that only the latest release branch can be fast
forwarded.

krel merges the provided ref into the release branch and asks for a final
confirmation if the push should really happen. The push will only be executed
as real push if the `--nomock` flag is specified.

## Installation

Simply [install krel](README.md#installation).

## Usage

```
  krel ff --branch <release-branch> [--ref <master-ref>] [--nomock] [--cleanup] [flags]
```

### Command Line Flags

```
Flags:
      --branch string   branch
      --cleanup         cleanup the repository after the run
  -h, --help            help for ff
      --ref string      ref on the main branch (default "origin/master")
      --repo string     the local path to the repository to be used (default "/tmp/k8s")

Global Flags:
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warning', 'info', 'debug', 'trace' (default "info")
      --nomock             run the command to target the production environment
```

### Example

```bash
krel ff --branch release-1.17 --ref origin/master --cleanup
```

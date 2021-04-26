# krel changelog

Automate the lifecycle of CHANGELOG-x.y.{md,html} files in a k/k repository

- [Summary](#summary)
- [Installation](#installation)
- [Usage](#usage)
- [Important notes](#important-notes)

## Summary

The `changelog` subcommand of `krel` does the following things by utilizing
the golang based `release-notes` tool:

1. Generate the release notes for either a patch or a new minor release. Minor
   releases can be alpha, beta or rc’s, too.

   a) Create a new `CHANGELOG-x.y.md` file if not existing.

   b) Correctly prepend the generated notes to the existing `CHANGELOG-x.y.md`
   file if already existing. This also includes the modification of the
   table of contents.

2. Convert the markdown release notes into a HTML equivalent on purpose of
   sending it by mail to the announce list. The HTML file will be dropped into
   the current working directly as `CHANGELOG-x.y.html`. Sending the
   announcement is done by another subcommand of 'krel', not 'changelog'.

3. Commit the modified CHANGELOG-x.y.md` into the master branch as well as the
   corresponding release-branch of kubernetes/kubernetes. The release branch
   will be pruned from all other CHANGELOG-\*.md files which do not belong to
   this release branch.

## Installation

Simply [install krel](README.md#installation).

## Usage

```
krel changelog [flags]
```

### Command Line Flags

```
Flags:
      --branch string      The branch to be used. Will be automatically inherited by the tag if not set.
      --bucket string      Specify gs bucket to point to in generated notes (default "kubernetes-release")
      --dependencies       Add dependency report (default true)
  -h, --help               help for changelog
      --html-file string   The target html file to be written. If empty, then it will be CHANGELOG-x.y.html in the current path.
      --record string      Record the API into a directory
      --replay string      Replay a previously recorded API from a directory
      --repo string        the local path to the repository to be used (default "/tmp/k8s")
      --tag string         The version tag of the release, for example v1.17.0-rc.1
      --tars string        Directory of tars to SHA512 sum for display (default ".")

Global Flags:
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warning', 'info', 'debug', 'trace' (default "info")
      --nomock             run the command to target the production environment
```

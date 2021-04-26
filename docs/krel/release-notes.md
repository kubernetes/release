# krel release-notes

The subcommand of choice for the Release Notes subteam of SIG Release

- [Summary](#summary)
- [Installation](#installation)
- [Usage](#usage)
- [Important notes and issues](#important-notes-and-issues)

## Summary

The `release-notes` subcommand of krel is used to generate the release notes
when a new Kubernetes version is cut. The subcommand can be used to generate
release notes pull requests to update the Release Notes Draft and the 
Kubernetes Release Notes Website.

This subcommand can be used to generate the release notes draft for the
current development version of kubernetes and the JSON version that power
relnotes.k8s.io. `krel release-notes` will create a branch in a user's fork of
the corresponding repositories. It will commit and push the changes and, finally,
it will file pull requests on your behalf to the proper kubernetes org repositories.

## Installation

After [installing krel](README.md#installation), you will need to [get a GitHub token](https://github.com/settings/tokens) to run the release-notes subcommand.

If you want to generate the JSON patches for relnotes.k8s.io you will need to have [npm](https://www.npmjs.com/) and a recent version of [node.js](https://nodejs.org/) installed to run the JSON formatter.

## Usage

Before running `krel release-notes` export your GitHub token to \$GITHUB_TOKEN:

```
  export GITHUB_TOKEN=YOURTOKENHERE
  krel release-notes [flags]
```

### Command line flags

```
Flags:
      --create-draft-pr     update the Release Notes draft and create a PR in k/sig-release
      --create-website-pr   patch the relnotes.k8s.io sources and generate a PR with the changes
      --dependencies        add dependency report (default true)
      --fix                 fix release notes
      --fork string         the user's fork in the form org/repo. Used to submit Pull Requests for the website and draft
  -h, --help                help for release-notes
      --list-v2             enable experimental implementation to list commits (ListReleaseNotesV2)
  -m, --maps-from strings   specify a location to recursively look for release notes *.y[a]ml file mappings
      --repo string         the local path to the repository to be used (default "/tmp/k8s")
  -t, --tag string          version tag for the notes

Global Flags:
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warning', 'info', 'debug', 'trace' (default "info")
      --nomock             run the command to target the production environment
```

### Examples

`krel release-notes` has two main modes of operation:

#### Update the Release Notes markdown draft

This invocation will generate the Release Notes Draft [published in sig-release](https://github.com/kubernetes/sig-release/blob/master/releases/release-1.19/release-notes-draft.md).
It will generate the draft for the current branch, starting from the first RC of the previous
minor Kubernetes release. For example, if you are generating the draft for `v1.19.0-beta.1`, `krel`
will generate the release notes draft starting from `v1.18.0-rc.1`.

The draft will be written into your fork of
[kubernetes/sig-release](https://github.com/kubernetes/sig-release). `krel release-notes` will
clone k/sig-release, create a branch, write the draft markdown and then push the new branch
back to your fork in GitHub. Finally __it will create a pull request__ on your behalf.

To update the Release Notes Draft, run `krel release-notes` with `--create-draft-pr` and set 
`--fork` to your GitHub user (the organization that owns your fork of k/sig-release):

```bash
krel release-notes --create-draft-pr --fork=kubefriend --tag v1.19.0-beta.1 
```

The `--tag` flag is optional, it will default to the latest minor version if it is not defined.
If, for any reason, your fork of k/sig-release is not named _sig-release_ you can set the name
of your repository by specifying the full repo slug `--fork=myorg/myrepo`.

#### Update the relnotes.k8s.io website

The subcommand can also generate the notes and modify the necessary files to update the
[release notes website](https://relnotes.k8s.io/). This invocation will clone the
[release-notes repo](https://github.com/kubernetes-sigs/release-notes) and add your fork as
a remote (kubefriend/release-notes). It will then create a feature branch to commit the notes
up to the tag you have defined and update the website files. If successful, it will push the
changes to your GitHub repository (defined by the `--fork` flag) and create a pull request:

```bash
krel release-notes --create-website-pr --fork=kubefriend --tag v1.19.0-beta.1 
```

As with `--create-draft-pr`, `--tag` is optional and will default to the latest release.
You can override the name of your fork of kubernetes-sigs/release-notes by specifying
the full repository slug: `--fork=myorg/myreponame`.

### Usage notes

You can run `--create-draft-pr` and `--create-website-pr` in the same invocation of krel.
Note that `krel` will use the same GitHub organization (`--fork`) to get the forks of
k-sigs/relese-notes and  k/sig release when doing so. Note that you cannot override the
name of the repositories when generating both PRs in the same invacation. 

## Important notes and issues

- Make sure [git `author.email`](https://help.github.com/en/github/setting-up-and-managing-your-github-user-account/setting-your-commit-email-address)
matches the address you used to sign the CNCF's CLA. Otherwise the
PR tests will fail.

- ~~If krel is interrupted while cloning, it might leave the repository at an unusable state. We recommend deleting any incomplete repositories before running krel again~~ (Fixed in #[1126](https://github.com/kubernetes/release/pull/1126))

- ~~If krel was run without forking `kubernetes/sig-release` and `kubernetes-sigs/release-notes`, the commit will be created but pushing to your fork will fail~~ (Fixed in #[1287](https://github.com/kubernetes/release/pull/1287))

- ~~The `release-notes` subcommand will eventually create the PR for you~~ (Implemented in #[1304](https://github.com/kubernetes/release/pull/1304))

- ~~The /tmp defaults shown in the flags above are for Linux. Other platform will
default to Go's `os.TempDir()` (for example, on a the Mac the actual path will be
under `/var/folders/…` )~~ (Deprecated in #[1350](https://github.com/kubernetes/release/pull/1350))


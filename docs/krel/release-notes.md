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

After [installing krel](README.md#installation), you will need to [get a GitHub token](https://github.com/settings/tokens) to run the release-notes subcommand specifying only the `repo/public_repo` scope:

- [ ] repo
  - [ ] repo: status
  - [ ] repo_deployment
  - [X] public_repo
  - [ ] repo:invite
  - [ ] security_events

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
      --create-website-pr   [DEPRECATED] patch the relnotes.k8s.io sources and generate a PR with the changes
      --dependencies        add dependency report (default true)
      --fix                 fix release notes
      --fork string         the user's fork in the form org/repo. Used to submit Pull Requests for the website and draft
  -h, --help                help for release-notes
      --list-v2             use git graph traversal to list commits instead of GitHub API date-based filtering (default true)
  -m, --maps-from strings   specify a location to recursively look for release notes *.y[a]ml file mappings
      --repo string         the local path to the repository to be used (default "/tmp/k8s")
  -t, --tag string          version tag for the notes

Global Flags:
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warning', 'info', 'debug', 'trace' (default "info")
      --nomock             run the command to target the production environment
```

### Examples

`krel release-notes` has three main modes of operation:

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

#### Rerun against an existing draft branch

After the initial `--create-draft-pr` run creates a PR against k/sig-release, reviewers may suggest changes. To incorporate those changes (via [map files](../release-notes-maps.md)), you need to regenerate the draft against the existing branch.

The `rerun` subcommand handles this workflow — it fetches an existing draft branch from any fork, regenerates the notes with maps applied, and optionally pushes the result to a destination fork.

This solves the problem where re-running `--create-draft-pr` fails because the branch already exists on the fork, and also supports handoffs where a different team member needs to pick up the rerun work from a colleague's fork.

The process:

1. Clone upstream k/sig-release and fetch the draft branch from `--draft-pr-source-fork`
2. Gather release notes from k/k for the given `--tag` range
3. Apply map files (from the branch and any extra `--maps-from` paths)
4. Write the updated markdown and JSON drafts, then commit
5. If `--draft-pr-push-fork` is set, push the branch there (updating any open PR)

The local clone is preserved after the run so you can inspect or push manually.

```bash
# Rerun and push to your own fork (updates the open PR):
krel release-notes rerun \
  --tag v1.36.0-beta.0 \
  --draft-pr-source-fork colleague \
  --draft-pr-push-fork myfork

# Rerun with additional local maps, no push (inspect locally):
krel release-notes rerun \
  --tag v1.36.0-beta.0 \
  --draft-pr-source-fork myfork \
  --maps-from /path/to/extra/maps
```

The `--draft-pr-source-fork` flag accepts either an org name (`myorg`, which expands to `myorg/sig-release`) or a full slug (`myorg/myrepo`).  
The same applies to `--draft-pr-push-fork`.

If `--draft-pr-source-branch` is not specified, it defaults to `release-notes-draft-<tag>`.

Flags specific to `rerun`:

```
      --draft-pr-source-fork string     k/sig-release fork to fetch the existing draft branch from (required)
      --draft-pr-source-branch string   branch to fetch (default: release-notes-draft-<tag>)
      --draft-pr-push-fork string       k/sig-release fork to push the updated branch to (omit to skip push)
      --draft-pr-push-branch string     branch to push to on the destination fork (default: same as source branch)
```

Inherited flags from the parent command (`--tag`, `--maps-from`, `--repo`, `--use-ssh`, `--update-repo`, `--include-labels`) are also available.

### Usage notes

You can run `--create-draft-pr` and `--create-website-pr` in the same invocation of krel.
Note that `krel` will use the same GitHub organization (`--fork`) to get the forks of
k-sigs/relese-notes and  k/sig release when doing so. Note that you cannot override the
name of the repositories when generating both PRs in the same invacation. 

## Important notes and issues

- Make sure [git `user.email`](https://help.github.com/en/github/setting-up-and-managing-your-github-user-account/setting-your-commit-email-address)
matches the address you used to sign the CNCF's CLA. Otherwise the
PR tests will fail.

- ~~If krel is interrupted while cloning, it might leave the repository at an unusable state. We recommend deleting any incomplete repositories before running krel again~~ (Fixed in #[1126](https://github.com/kubernetes/release/pull/1126))

- ~~If krel was run without forking `kubernetes/sig-release` and `kubernetes-sigs/release-notes`, the commit will be created but pushing to your fork will fail~~ (Fixed in #[1287](https://github.com/kubernetes/release/pull/1287))

- ~~The `release-notes` subcommand will eventually create the PR for you~~ (Implemented in #[1304](https://github.com/kubernetes/release/pull/1304))

- ~~The /tmp defaults shown in the flags above are for Linux. Other platform will
default to Go's `os.TempDir()` (for example, on a the Mac the actual path will be
under `/var/folders/…` )~~ (Deprecated in #[1350](https://github.com/kubernetes/release/pull/1350))


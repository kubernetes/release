# krel release-notes
The subcommand of choice for the Release Notes subteam of SIG Release

- [Summary](#summary)
- [Installation](#installation)
- [Usage](#usage)
- [Important notes](#important-notes)

## Summary

The `release-notes` subcommand of krel is used to generate the release notes 
draft of a kubernetes version. The subcommand can be used to generate release notes
files for a specific tag or a range of tags. It can output the notes files to Markdown or JSON formats.

This subcommand can be used to generate the release notes draft for the current development version of kubernetes and the JSON version that power relnotes.k8s.io.
`krel release-notes` will create a branch in a user's fork of the corresponding repositories, commit and push the changes. Filing the final PRs to the kubernetes org repositories is in development and will soon be ready.

## Installation
After [installing krel](README.md#installation), you will need to [get a GitHub token](https://github.com/settings/tokens) to run the release-notes subcommand.

If you want to generate the JSON patches for relnotes.k8s.io you will need to have [npm](https://www.npmjs.com/) installed to run the JSON formatter.

## Usage

Before running `krel release-notes` export your GitHub token to $GITHUB_TOKEN:
```
  export GITHUB_TOKEN=YOURTOKENHERE
  krel release-notes [flags]
```

### Command line flags
```
      --create-draft-pr                    create the Release Notes Draft PR. --draft-org and --draft-repo must be set along with this option
      --create-website-pr                  generate the Releas Notes to a local fork of relnotes.k8s.io and create a PR.  --draft-org and --draft-repo must be set along with this option
      --draft-org string                   a Github organization owner of the fork of k/sig-release where the Release Notes Draft PR will be created
      --draft-repo string                  the name of the fork of k/sig-release, the Release Notes Draft PR will be created from this repository (default "sig-release")
      --format string                      The format for notes output (options: markdown, json) (default "markdown")
  -h, --help                               help for release-notes
      --kubernetes-sigs-fork-path string   fork kubernetes-sigs/release-notes and output a copy of the json release notes to this directory (default "/tmp/k8s-sigs")
  -o, --output-dir string                  output a copy of the release notes to this directory (default ".")
      --sigrelease-fork-path string        fork k/sig-release and output a copy of the release notes draft to this directory (default "/tmp/k8s-sigrelease")
  -t, --tag string                         version tag for the notes
      --website-org string                 a Github organization owner of the fork of kuberntets-sigs/release-notes where the Website PR will be created
      --website-repo string                the name of the fork of kuberntets-sigs/release-notes, the Release Notes Draft PR will be created from this repository (default "release-notes")
```


### Examples

`krel release-notes` has three main modes of operation:

#### Ouput the Releas Notes to a single file
The following command will generate the release notes for the 1.18 branch up to beta.2 in /var/www/html/

```bash
krel release-notes -o /var/www/html/ --format markdown --tag v1.18.0-beta.2
```
#### Generate the current markdown draft
This invocation will generate the release notes draft [published in sig-release](https://github.com/kubernetes/sig-release/blob/master/releases/release-1.18/release-notes-draft.md). It will generate the draft for the v1.18 branch up to rc.1. The draft will be written in my local fork of kubernetes/sig-release. `krel release-notes` will clone k/sig-release, create a branch, write the draft and then push the changes back to github, in `kubefriend/sig-release`.

```bash
krel release-notes --create-draft-pr --tag v1.18.0-rc.1 --draft-org=kubefriend
```

#### Update the relnotes.k8s.io website
The subcommand can also generate the notes and modify the necessary files to update the [release notes website](https://relnotes.k8s.io/). This invocation will clone the [release-notes repo](https://github.com/kubernetes-sigs/release-notes) and add my fork as a remote (kubefriend/release-notes). It will then create a feature branch to commit the notes up to v1.18 alpha.1 and update the website files. Finally it will push the changes to my GitHub repo:

```bash
krel release-notes --tag v1.18.0-alpha.1 --website-org=kubefriend --create-website-pr
```
## Important notes

The /tmp defaults shown in the flags above are for Linux. Other platform will default to Go's `os.TempDir()` (for example, on a the Mac the actual path will be under `/var/folders/…` ).

The release-notes subcommand will eventually create the PR for you. This is still under development though.
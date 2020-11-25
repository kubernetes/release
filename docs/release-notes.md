# Kubernetes Release Notes Workflow

Release notes are captured during the development cycle from [PR titles (or body
blocks)](https://github.com/kubernetes/kubernetes/blob/master/.github/PULL_REQUEST_TEMPLATE.md) using [release notes
labels](https://github.com/kubernetes/community/blob/master/contributors/guide/release-notes.md) with compliance across master and release- branches using [test-infra/prow](https://github.com/kubernetes/test-infra/tree/master/prow/plugins/releasenote)
on the main Kubernetes repository (and future repositories later).

Releases are built and published by `krel` in [this repo](https://github.com/kubernetes/release) with the release notes published in [kubernetes/CHANGELOG/CHANGELOG-x.y.md files](https://git.k8s.io/kubernetes/CHANGELOG/README.md).

The automated release notes gathered in this way are considered complete
for alpha, beta and official patch releases.

For new branch, major milestones (x.y.0), the automated release
note tooling provides a complete set of changes since the last major
milestone release by referencing all alphas and betas for the minor .X. version.

This bootstraps a milestone release in CHANGELOG-x.y.md with a template for
use by the Release Czar and the subteam area leads to further expand
on the release notes with more detail as needed.

An example major milestone release (x.y.0) in CHANGELOG-x.y.md:

```
# vx.y.0
[Documentation](https://docs.k8s.io/)

## Downloads

binary | sha1 hash | md5 hash
------ | --------- | --------
[kubernetes.tar.gz](https://dl.k8s.io/vx.y.0/kubernetes.tar.gz) | `7deca064d2e277a0beed802c7cfe32d152434aed` | `0b735e8b9fd64064caa578d49062820d`

## Major Themes
* TBD

## Other notable improvements
* TBD

## Known Issues
* TBD

## Provider-specific Notes
* TBD

## Changelog since vx.y.0-beta.2

### Action Required
* This very important item to watch out for

### Other notable changes
* this bug fix
* that bug fix
* this cool feature

### Previous Releases Included in vx.y.0
- [vx.y.0-beta.2](#vxy0-beta2)
- [vx.y.0-beta.1](#vxy0-beta1)
- [vx.y.0-alpha.5](#vxy0-alpha5)
- [vx.y.0-alpha.4](#vxy0-alpha4)
- [vx.y.0-alpha.3](#v1N0-alpha3)
- [vx.y.0-alpha.2](#v1N0-alpha2)
- [vx.y.0-alpha.1](#v1N0-alpha1)
```

## FAQ

### Why do we publish a bootstrapped/templated set of release notes for milestone releases?

For 1.2.0, the release notes in their entirety were held up in a team-wide
editing session that lasted weeks.  This delay between the actual release
and the publishing of the release notes was widely disliked.

Newer releases have the benefit of release note labeling and automated
publishing leading up to a milestone release.

By providing this mostly complete set of notes at release time, we
provide immediate value to the community and to users.

The templated portion of the release notes let users know that additional
details on the release are forthcoming.


### Where can I keep Known Issues leading up to a release?

"Known Issues" can often simply be categorized as "Action required"
and there's a release note label for that if your known issue is covered
by a PR.

If your known issue requires more detail than can fit in a PR title, see the
next FAQ item below.

If your known issue isn't tracked by a PR at all, please contact your area lead and cloud-kubernetes-release@google.com to discuss.

### Where can I expand on a release note for a PR when the content doesn't fit in a PR title?

Use PR [body blocks](https://github.com/kubernetes/kubernetes/blob/master/.github/PULL_REQUEST_TEMPLATE.md) for expanded release notes.

You can add or update a body block at any time in a PR body prior to a
release and it will end up in the release notes.


## Related

* [Original release notes proposal](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/release-notes.md)
* [Behind The Scenes: Kubernetes Release Notes Tips & Tricks - Mike Arpaia, Kolide (KubeCon 2018 Lightning Talk)](https://www.youtube.com/watch?v=n62oPohOyYs)

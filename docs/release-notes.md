# Kubernetes Release Notes Workflow

Release notes are captured during the development cycle from [PR titles (or body
blocks)](https://github.com/kubernetes/kubernetes/blob/master/.github/PULL_REQUEST_TEMPLATE.md) using [release notes
labels](https://github.com/kubernetes/kubernetes/blob/master/docs/devel/pull-requests.md#release-notes) with compliance across master and release- branches using the munger in [contrib/mungegithub](https://github.com/kubernetes/contrib/blob/master/mungegithub/README.md#submit-queue)
on the main Kubernetes repository (and future repositories later).

Releases are built and published by the anago tool in [this repo](https://github.com/kubernetes/release) with the release notes published in [kubernetes/CHANGELOG.md](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG.md).

The automated release notes gathered in this way are considered complete
for alpha, beta and official patch releases.

For new branch, major milestones (1.X.0), the automated release
note tooling provides a complete set of changes since the last major
milestone release by referencing all alphas and betas for the minor .X. version.

This bootstraps a milestone release in CHANGELOG.md with a template for
use by the Release Czar and the subteam area leads to then further expand
on the release notes with more detail as needed.

A major milestone release (1.X.0) in CHANGELOG.md might look like:

* Changes since v1.2.0
   * 1.3.0-beta.2
   * 1.3.0-beta.1
   * v1.3.0-alpha.5
   * ...
   * ...
* Major Themes
  * ...
* Experimental Features
  * ...more detail on experimental features...
* Action required
  * ...more detail on action-required features...
* Known Issues
   * Docker Known Issues
* Provider-specific Notes


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

Use PR [body blocks]((https://github.com/kubernetes/kubernetes/blob/master/.github/PULL_REQUEST_TEMPLATE.md) for expanded release notes.

You can add or update a body block at any time in a PR body prior to a
release and it will end up in the release notes.


## Related

* [Original release notes proposal](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/release-notes.md))

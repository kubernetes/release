# Kubernetes Releases

This repo contains the tooling and documentation for the release of
the core Kubernetes project.  

In the future it is expected that the functionality will expand and be
generalized to support release infrastructure for all of the kubernetes
sub-projects as well.

The tooling and doc here is expected to change often as requirements
change and the project(s) evolve.

The doc and tooling in this repo is NOT designed to address the planning,
coordination of releases.  For more info on feature and release planning see:
* [Kubernetes Community Wiki](https://github.com/kubernetes/community/wiki)
* [Kubernetes Feature Tracking](https://github.com/kubernetes/features)


## Types of Releases

* Alpha releases (`vX.Y.0-alpha.W`) are cut directly from `master`.
* Beta releases (`vX.Y.Z-beta.W`) are cut from their respective release branch,
  `release-X.Y`.
* Official releases (`vX.Y.Z`) are cut from their respective release branch,
  `release-X.Y`.
* Emergency releases (`vX.Y.Z`) are cut from a new release-X.Y.Z branch based on a tag

## Release Schedule

| Type      | Versioning     | Branch        | Frequency                  |
| ----      | ----------     | ------        | ---------                  |
| alpha     | vX.Y.0-alpha.W | master        | every ~2 weeks             |
| beta      | vX.Y.Z-beta.W  | release-N.N   | as needed (at branch time) |
| official  | vX.Y.Z         | release-N.N   | as needed (post beta)      |
| emergency | vX.Y.Z         | release-N.N.N | as needed                  |

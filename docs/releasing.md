# Kubernetes Releases

This repo contains the tooling and documentation for the release of
the core Kubernetes project.  

In the future it is expected that the functionality will expand and be
generalized to support release infrastrcture for all of the kubernetes
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
* New release series are also cut directly from `master`.

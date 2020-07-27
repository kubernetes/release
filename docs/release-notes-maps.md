# Release Notes Map Files

The release notes libraries and command libraries now support incorporating
data from sources outside of the data gathered from pull requests and the
GitHub API through the use of _map files_.

A release note map file is a YAML construct intended to add and/or modify
the information in a release note gathered from GitHub. The motivations for
developing the maps are to give the release notes team more flexibility 
to edit the notes during the release cycle and to be able to add more
data and domain-specific context to the release notes.

## Reading Map Files with `release-notes` or `krel` 

To read the map files, a new `--maps-from` flag has been added
to the `release-notes` command and the `krel release-notes` subcommand.
This new flag takes as value a location from which to read YAML files:

```console
release-notes --maps-from=/path/to/yaml/files/

# Future map locations would be prefixed with a URL-like schema.
# An example, to read from a GCP bucket (not yet implemented):

krel release-notes --maps-from=gs://bucket-name/path/
```

The logic to read from each location is handled by a MapProvider (see below).

## Release Notes Map Format

A release notes map has two parts, each one optional. The first one
allows us to override the data fields defined in the associated 
Pull Request. The second part provides a mechanism for adding more data
to a release note to provide more context or additional
information to a release note.

### Overriding Fields: the `release-note` section

The first part of the YAML in a map file overrides the data defined in the release note pull request. Each of the fields corresponds to a PR data field. A map file
can contain any of them. An example:

```yaml
---
pr: 123
commit: 1a89038915fe77d73bf7c9cfa8f2ce123a464c82
release-note:  
  text: Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
  author: kubernetes-ci-robot
  areas:
    - release-eng
  kinds:
    - feature
  sigs:
    - release
  feature: true
  action_required: false
  release_version: v1.19.0
  documentation:
    - description: Release Notes Improvements
      url: https://github.com/kubernetes/enhancements/tree/master/keps/sig-release/1733-release-notes
      type: kep
```
Any of the fields defined in a map file fully overrides its PR counterpart. 

### Adding data: The `datafields` section

The second section in a release notes map is the `datafields` struct. 
A data field is a free form YAML structure defined under a key in the YAML
file. The idea of a datafield is to be able to have a central and extensible
way to add content and improve the context of a note.

The interpretation and rendering of each data field are specific to their
domain. While our libraries will read the data fields under a new key, 
rendering and use need to be implemented manually.

The initial implementation of a data field was written to add CVE
vulnerability data to a release note (see issue [#1354](https://github.com/kubernetes/release/issues/1354)). An example of the CVE information in a map file:

```yaml
datafields:
  cve:
    id: CVE-2019-1010260
    title: "CVE-2020-8555: Half-Blind SSRF in kube-controller-manager"
    published: 2020-05-28
    score: 5.2
    rating: CVSS:3.0/AV:N/AC:H/PR:L/UI:N/S:C/C:H/I:N/A:N
    linkedPRs:
    - 89794
    - 89796
    - 89837
    - 89838
    - 89839
    description:
        There exists a Server Side Request Forgery (SSRF) vulnerability in kube-controller-manager that allows certain authorized users to leak up to 500 bytes of arbitrary information from unprotected endpoints within the master's host network (such as link-local or loopback services).
        An attacker with permissions to create a pod with certain built-in Volume types (GlusterFS, Quobyte, StorageOS, ScaleIO) or permissions to create a StorageClass can cause kube-controller-manager to make GET requests or POST requests without an attacker controlled request body from the master's host network.
```

## Finding Maps: The `MapProvider` Interface

Release notes maps are simple YAML files. In order to find and read them, the 
release notes libraries have implemented a MapProvider interface. The job of
a MapProvider is to initialize itself from a location string and implement 
a single hook function `GetMapsForPR(int)`. The release notes libraries will
request maps by PR number from each provider defined. It will request the
maps after gathering the data from GitHub and just before rendering.

In theory, any number of map providers can be attached to the release notes
process by invoking the `--maps-from` flag:

```console
release-notes --maps-from=/path/to/mapfiles/ 
```

The motivation of having a MapProvider interface is to be able to _read_
maps from different sources. Currently, we have a single provider `DirectoryMapProvider` which takes a directory name as a location and reads
the YAML files found in it.

Additional ideas in flight for new MapProviders include a `GitHubMapProvider`
and `GoogleCloudStorageProvider` which would add the ability to read from
a (possibly private) GitHub repo and a Google Cloud Storage bucket.

To add a new provider, create a new URL-like init string to be associated 
with the provider by its schema (for example "gs://"). Then hack the 
`notes.NewProviderFromInitString` function to recognize it. Finally,
add your file reading logic and implement the `GetMapsForPR()` hook to
return the data when called.


# krel — The Kubernetes Release Toolbox

`krel` is the new golang based tool for managing releases.

- [Summary](#summary)
- [Installation](#installation)
- [Usage](#usage)
- [Important notes](#important-notes)

## Summary

The purpose of krel is to provide a toolkit for managing the different steps needed to create
Kubernetes Releases. This includes manually executed tasks like generating the Release Notes during the release cycle and performing automated tasks like pushing the Kubernetes release artifacts to Google Cloud Storage.

## Installation

Compile krel by running the `compile-release-tools` script from the root of this repo:

```shell
./compile-release-tools krel
```

## Usage:

krel has several subcommands that perform various tasks during the release lifecycle:

`krel [subcommand]`

### Available Commands:

| Subcommand                          | Description                                                                                 |
| ----------------------------------- | --------------------------------------------------------------------------------------------|
| announce                            | Build and announce Kubernetes releases                                                      |
| [changelog](changelog.md)           | Automate the lifecycle of CHANGELOG-x.y.{md,html} files in a k/k repository                 |
| ci-build                            | Build Kubernetes in CI and push release artifacts to Google Cloud Storage (GCS)             |
| cve                                 | Add and edit CVE information                                                                |
| [ff](ff.md)                         | Fast forward a Kubernetes release branch                                                    |
| history                             | Run history to build a list of commands that ran when cutting a specific Kubernetes release |
| promote-images                      | Starts an image promotion for a tag of kubernetes images                                    |
| [push](push.md)                     | Push Kubernetes release artifacts to Google Cloud Storage (GCS)                             |
| release                             | Release a staged Kubernetes version                                                         |
| [release-notes](release-notes.md)   | The subcommand of choice for the Release Notes subteam of SIG Release                       |
| stage                               | Stage a new Kubernetes version                                                              |
| testgridshot                        | Take a screenshot of the testgrid dashboards                                                |

## Important Notes

Some of the krel subcommands are under development and their usage may already differ from these docs.

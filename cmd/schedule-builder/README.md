# Schedule Builder

This simple tool has the objective to parse the yaml file located in
[k/website](https://github.com/kubernetes/website/blob/main/data/releases/schedule.yaml),
which shows the scheduled and past patch releases of the current Kubernetes
Release cycle as well as Kubernetes Release Cycles schedules ([like 1.23](https://github.com/kubernetes/sig-release/blob/master/releases/release-1.23/README.md)) in machine readable format.

## Install

The simplest way to install the `schedule-builder` CLI is via `go get`:

```
$ go get k8s.io/release/cmd/schedule-builder
```

This will install `schedule-builder` to `$(go env GOPATH)/bin/schedule-builder`.

Also if you have the `kubernetes/release` cloned you can run the `make release-tools` to build all the tools.

## Usage

To run this tool you can just do, assuming you have cloned both `website` and `release` repositories, like

```
.
+-- k8s.io
|   +-- release
|   +-- website
```

### For Patch Release Schedule
```bash
$ schedule-builder --config-path ../website/data/releases/schedule.yaml --type patch
```

The output will be something similar to this

```
### Timeline

### 1.18

Next patch release is **1.18.4**

End of Life for **1.18** is **TBD**

| PATCH RELEASE | CHERRY PICK DEADLINE  | TARGET DATE |
|---------------|-----------------------|-------------|
| 1.18.4        | 2020-06-12            | 2020-06-17  |
| 1.18.3        | 2020-05-15            | 2020-05-20  |
| 1.18.2        | 2020-04-13            | 2020-04-16  |
| 1.18.1        | 2020-04-06            | 2020-04-08  |

### 1.17

Next patch release is **1.17.7**

End of Life for **1.17** is **TBD**

| PATCH RELEASE |                                 CHERRY PICK DEADLINE                                  | TARGET DATE |
|---------------|---------------------------------------------------------------------------------------|-------------|
| 1.17.7        | 2020-06-12                                                                            | 2020-06-17  |
| 1.17.6        | 2020-05-15                                                                            | 2020-05-20  |
| 1.17.5        | 2020-04-13                                                                            | 2020-04-16  |
| 1.17.4        | 2020-03-09                                                                            | 2020-03-12  |
| 1.17.3        | 2020-02-07                                                                            | 2020-02-11  |
| 1.17.2        | No-op release https://groups.google.com/d/topic/kubernetes-dev/Mhpx-loSBns/discussion | 2020-01-21  |
| 1.17.1        | 2020-01-10                                                                            | 2020-01-14  |

### 1.16

Next patch release is **1.16.11**

End of Life for **1.16** is **TBD**

| PATCH RELEASE |                                 CHERRY PICK DEADLINE                                  | TARGET DATE |
|---------------|---------------------------------------------------------------------------------------|-------------|
| 1.16.11       | 2020-06-12                                                                            | 2020-06-17  |
| 1.16.10       | 2020-05-15                                                                            | 2020-05-20  |
| 1.16.9        | 2020-04-13                                                                            | 2020-04-16  |
| 1.16.8        | 2020-03-09                                                                            | 2020-03-12  |
| 1.16.7        | 2020-02-07                                                                            | 2020-02-11  |
| 1.16.6        | No-op release https://groups.google.com/d/topic/kubernetes-dev/Mhpx-loSBns/discussion | 2020-01-21  |
| 1.16.5        | 2020-01-10                                                                            | 2020-01-14  |
| 1.16.4        | 2019-12-06                                                                            | 2019-12-11  |
| 1.16.3        | 2019-11-08                                                                            | 2019-11-13  |
| 1.16.2        | 2019-10-11                                                                            | 2019-10-15  |
| 1.16.1        | 2019-09-27                                                                            | 2019-10-02  |
```

### For Release Cycle Schedule
```bash
$ schedule-builder --config-path testdata/rel-schedule.yaml --type release
```
The output will be something similar to this

```
# Kubernetes 1.23

#### Links


* [This document](https://git.k8s.io/sig-release/releases/release-1.23/README.md)
* [Release Team](https://github.com/kubernetes/sig-release/blob/master/releases/release-1.23/release-team.md)
* [Meeting Minutes](http://bit.ly/k8s123-releasemtg) (join [kubernetes-sig-release@] to receive meeting invites)
* [v1.23 Release Calendar](https://bit.ly/k8s-release-cal)
* Contact: [#sig-release] on slack, [kubernetes-release-team@] on e-mail
* [Internal Contact Info][Internal Contact Info] (accessible only to members of [kubernetes-release-team@])

#### Tracking docs


* [Enhancements Tracking Sheet](https://bit.ly/k8s123-enhancements)
* [Feature blog Tracking Sheet](TBD)
* [Bug Triage Tracking Sheet](TBD)
* CI Signal Report: TODO
* [Retrospective Document][Retrospective Document]
* [kubernetes/sig-release v1.23 milestone](https://github.com/kubernetes/kubernetes/milestone/56)

#### Guides


* [Targeting Issues and PRs to This Milestone](https://git.k8s.io/community/contributors/devel/sig-release/release.md)
* [Triaging and Escalating Test Failures](https://git.k8s.io/community/contributors/devel/sig-testing/testing.md#troubleshooting-a-failure)

## TL;DR

The 1.23 release cycle is proposed as follows:

- **Mon August 23, 2021**: week 1- Release cycle begins
- **Thu September 2, 2021**: week 2- [Production Readiness Soft Freeze](https://groups.google.com/g/kubernetes-sig-architecture/c/a6_y81N49aQ)
- **Thu September 9, 2021**: week 3 - [Enhancements Freeze](../release_phases.md#enhancements-freeze)
- **Tue November 16, 2021**: week 13 - [Code Freeze](../release_phases.md#code-freeze)
- **Tue November 23, 2021**: week 14 - [Test Freeze](../release_phases.md#test-freeze)
- **Tue November 30, 2021**: week 15 - Docs must be completed and reviewed
- **Tue December 7, 2021**: week 16 - Kubernetes v1.23.0 released
- **Tue December 14, 2021**: week 17 - [Release Retrospective][Retrospective Document] part 1
- **Wed December 15, 2021**: week 17 - [Release Retrospective][Retrospective Document] part 2
- **Thu December 16, 2021**: week 17 - [Release Retrospective][Retrospective Document] part 3

## Timeline

|                                 **WHAT**                                 |            **WHO**            |       **WHEN**        |  **WEEK**   |                  **CI SIGNAL**                   |  |
|--------------------------------------------------------------------------|-------------------------------|-----------------------|-------------|--------------------------------------------------|--|
| Start of Release Cycle                                                   | Lead                          | Mon August 23, 2021   | week 1      | master-blocking                                  |  |
| Start Enhancements Tracking                                              | Enhancements Lead             | Mon August 23, 2021   | week 1      |                                                  |  |
| 1.23.0-alpha.1 released                                                  | Branch Manager                | Wed August 25, 2021   | week 1      |                                                  |  |
| Schedule finalized                                                       | Lead                          | Thu August 26, 2021   | week 1      |                                                  |  |
| Team finalized                                                           | Lead                          | Fri August 27, 2021   | week 1      |                                                  |  |
| Production Readiness Soft Freeze                                         | Enhancements Lead             | Thu September 2, 2021 | week 2      |                                                  |  |
| Start Release Notes Draft                                                | Release Notes Lead            | Tue September 7, 2021 | week 3      |                                                  |  |
| Begin Enhancements Freeze (23:59 PDT)                                    | Enhancements Lead             | Thu September 9, 2021 | week 3      | master-blocking, master-informing                |  |
| 1.23.0-alpha.2 released                                                  | Branch Manager                | TBD                   | TBD         |                                                  |  |
| 1.23.0-alpha.3 released                                                  | Branch Manager                | TBD                   | TBD         |                                                  |  |
| KubeCon NA + Co-located events                                           |                               | October 11-15, 2021   | week 8      |                                                  |  |
| 1.23.0-alpha.4 released                                                  | Branch Manager                | TBD                   | TBD         |                                                  |  |
| Begin Burndown (MWF meetings)                                            | Lead                          | Mon November 1, 2021  | week 11     | 1.23-blocking, master-blocking, master-informing |  |
| Call for Exceptions                                                      | Lead                          | Mon November 1, 2021  | 2021	week 11 |                                                  |  |
| Brace Yourself, Code Freeze is Coming                                    | Comms / Bug Triage            | Mon November 1, 2021  | week 11     |                                                  |  |
| Begin Feature blog freeze (23:59 PDT)                                    | Comms Lead                    | Tue November 2, 2021  | week 11     |                                                  |  |
| Burndown Meetings daily                                                  | Lead                          | Mon November 15, 2021 | week 13     |                                                  |  |
| Begin Code Freeze (18:00 PST)                                            | Branch Manager                | Tue November 16, 2021 | week 13     |                                                  |  |
| 1.23.0-beta.0 released                                                   | Branch Manager                | Tue November 16, 2021 | week 13     |                                                  |  |
| Docs deadline - Open placeholder PRs                                     | Docs Lead                     | Thu November 18, 2021 | week 13     |                                                  |  |
| Test Freeze (18:00 PST)                                                  | Branch Manager                | Tue November 23, 2021 | week 14     |                                                  |  |
| Docs deadline - PRs ready for review                                     | Docs Lead                     | Tue November 23, 2021 | week 14     |                                                  |  |
| 1.23.0-rc.0 released                                                     | Branch Manager                | Tue November 23, 2021 | week 14     |                                                  |  |
| release-1.23 branch created                                              | Branch Manager                | Tue November 23, 2021 | week 14     |                                                  |  |
| Start final draft of Release Notes                                       | Release Notes Lead            | Tue November 23, 2021 | week 14     |                                                  |  |
| Release blog ready to review (23:59 PST)                                 | Comms / Docs                  | Tue November 23, 2021 | week 14     |                                                  |  |
| Docs complete - All PRs reviewed and ready to merge                      | Docs Lead                     | Tue November 30, 2021 | week 15     |                                                  |  |
| Feature blogs ready to review (23:59 PST)                                | Enhancement Owner / SIG Leads | Tue November 30, 2021 | week 15     |                                                  |  |
| Release Notes complete - reviewed & merged to k/sig-release (23:59 PST)  | Release Notes Lead            | Thu December 2, 2021  | week 15     |                                                  |  |
| v1.23.0 released                                                         | Branch Manager                | Tue December 7, 2021  | week 16     |                                                  |  |
| Release blog published                                                   | Comms                         | Tue December 7, 2021  | week 16     |                                                  |  |
| Thaw                                                                     | Branch Manager                | Tue December 7, 2021  | week 16     |                                                  |  |
| Release retrospective part 1 (7:30am PST during the SIG Release meeting) | Community                     | Tue December 14, 2021 | week 17     |                                                  |  |
| Release retrospective part 2 (10:00am PST)                               | Community                     | Wed December 15, 2021 | week 17     |                                                  |  |
| Release retrospective part 3 (10:00am PST)                               | Community                     | Thu December 16, 2021 | week 17     |                                                  |  |

## Phases


Please refer to the [release phases document](../release_phases.md).

[k8s1.23-calendar]: https://bit.ly/k8s-release-cal
[Internal Contact Info]: https://bit.ly/k8s123-contacts
[Retrospective Document]: https://bit.ly/k8s123-retro

[Enhancements Freeze]: ../release_phases.md#enhancements-freeze
[Burndown]: ../release_phases.md#burndown
[Code Freeze]: ../release_phases.md#code-freeze
[Exception]: ../release_phases.md#exceptions
[Thaw]: ../release_phases.md#thaw
[Test Freeze]: ../release_phases.md#test-freeze
[release-team@]: https://groups.google.com/a/kubernetes.io/g/release-team
[kubernetes-sig-release@]: https://groups.google.com/forum/#!forum/kubernetes-sig-release
[#sig-release]: https://kubernetes.slack.com/messages/sig-release/
[kubernetes-release-calendar]: https://bit.ly/k8s-release-cal
[kubernetes/kubernetes]: https://github.com/kubernetes/kubernetes
[master-blocking]: https://testgrid.k8s.io/sig-release-master-blocking#Summary
[master-informing]: https://testgrid.k8s.io/sig-release-master-informing#Summary
[123-blocking]: https://testgrid.k8s.io/sig-release-1.23-blocking#Summary
[exception requests]: ../EXCEPTIONS.md
[release phases document]: ../release_phases.md
```

Also can save the schedule in a file, to do that, you can set the `--output-file` flag together with the filename.

```
$ schedule-builder --config-path ../website/data/releases/schedule.yaml --type patch--output-file my-schedule.md
```

/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const expectedPatchSchedule = `### Timeline

### X.Y

Next patch release is **X.Y.ZZZ**

End of Life for **X.Y** is **NOW**

| PATCH RELEASE | CHERRY PICK DEADLINE | TARGET DATE | NOTE |
|---------------|----------------------|-------------|------|
| X.Y.ZZZ       | 2020-06-12           | 2020-06-17  |      |
| X.Y.XXX       | 2020-05-15           | 2020-05-20  | honk |
| X.Y.YYY       | 2020-04-13           | 2020-04-16  |      |
`

const expectedReleaseSchedule = `# Kubernetes X.Y

#### Links


* [This document](https://git.k8s.io/sig-release/releases/release-X.Y/README.md)
* [Release Team](https://github.com/kubernetes/sig-release/blob/master/releases/release-X.Y/release-team.md)
* [Meeting Minutes](http://bit.ly/k8sXY-releasemtg) (join [kubernetes-sig-release@] to receive meeting invites)
* [vX.Y Release Calendar](https://bit.ly/k8s-release-cal)
* Contact: [#sig-release] on slack, [kubernetes-release-team@] on e-mail
* [Internal Contact Info][Internal Contact Info] (accessible only to members of [kubernetes-release-team@])

#### Tracking docs


* [Enhancements Tracking Sheet](https://bit.ly/k8sXY-enhancements)
* [Feature blog Tracking Sheet](TBD)
* [Bug Triage Tracking Sheet](TBD)
* CI Signal Report: TODO
* [Retrospective Document][Retrospective Document]
* [kubernetes/sig-release vX.Y milestone](https://github.com/kubernetes/kubernetes/milestone/56)

#### Guides


* [Targeting Issues and PRs to This Milestone](https://git.k8s.io/community/contributors/devel/sig-release/release.md)
* [Triaging and Escalating Test Failures](https://git.k8s.io/community/contributors/devel/sig-testing/testing.md#troubleshooting-a-failure)

## TL;DR

The X.Y release cycle is proposed as follows:

- **2020-06-17**: week 1- Release cycle begins
- **2020-06-20**: week 1- [Production Readiness Soft Freeze](https://groups.google.com/g/kubernetes-sig-architecture/c/a6_y81N49aQ)
- **2020-06-21**: week 1 - [Enhancements Freeze](../release_phases.md#enhancements-freeze)
- **2020-06-22**: week 1 - [Code Freeze](../release_phases.md#code-freeze)
- **2020-06-25**: week 2 - [Test Freeze](../release_phases.md#test-freeze)
- **2020-06-26**: week 2 - Docs must be completed and reviewed
- **2020-06-27**: week 2 - Kubernetes vX.Y.0 released
- **2020-06-27**: week 2 - [Release Retrospective][Retrospective Document] part 1
- **2020-06-27**: week 2 - [Release Retrospective][Retrospective Document] part 2
- **2020-06-28**: week 2 - [Release Retrospective][Retrospective Document] part 3

## Timeline

| **WHAT**  | **WHO** |  **WHEN**  | **WEEK** | **CI SIGNAL** |  |
|-----------|---------|------------|----------|---------------|--|
| Testing-A | tester  | 2020-06-17 | week 1   | green         |  |
| Testing-B | tester  | 2020-06-19 | week 1   | green         |  |
| Testing-C | tester  | 2020-06-20 | week 1   | green         |  |
| Testing-D | tester  | 2020-06-21 | week 1   | green         |  |
| Testing-E | tester  | 2020-06-22 | week 1   | green         |  |
| Testing-F | tester  | 2020-06-25 | week 2   | green         |  |
| Testing-G | tester  | 2020-06-26 | week 2   | green         |  |
| Testing-H | tester  | 2020-06-27 | week 2   | green         |  |
| Testing-I | tester  | 2020-06-27 | week 2   | green         |  |
| Testing-J | tester  | 2020-06-27 | week 2   | green         |  |
| Testing-K | tester  | 2020-06-28 | week 2   | green         |  |
| Testing-L | tester  | 2020-06-28 | week 2   | green         |  |

## Phases


Please refer to the [release phases document](../release_phases.md).

[k8sX.Y-calendar]: https://bit.ly/k8s-release-cal
[Internal Contact Info]: https://bit.ly/k8sXY-contacts
[Retrospective Document]: https://bit.ly/k8sXY-retro

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
[XY-blocking]: https://testgrid.k8s.io/sig-release-X.Y-blocking#Summary
[exception requests]: ../EXCEPTIONS.md
[release phases document]: ../release_phases.md`

func TestParseSchedule(t *testing.T) {
	testcases := []struct {
		name     string
		schedule PatchSchedule
	}{
		{
			name: "next patch is not in previous patch list",
			schedule: PatchSchedule{
				Schedules: []Schedule{
					{
						Release:            "X.Y",
						Next:               "X.Y.ZZZ",
						CherryPickDeadline: "2020-06-12",
						TargetDate:         "2020-06-17",
						EndOfLifeDate:      "NOW",
						PreviousPatches: []PreviousPatches{
							{
								Release:            "X.Y.XXX",
								CherryPickDeadline: "2020-05-15",
								TargetDate:         "2020-05-20",
								Note:               "honk",
							},
							{
								Release:            "X.Y.YYY",
								CherryPickDeadline: "2020-04-13",
								TargetDate:         "2020-04-16",
							},
						},
					},
				},
			},
		},
		{
			name: "next patch is in previous patch list",
			schedule: PatchSchedule{
				Schedules: []Schedule{
					{
						Release:            "X.Y",
						Next:               "X.Y.ZZZ",
						CherryPickDeadline: "2020-06-12",
						TargetDate:         "2020-06-17",
						EndOfLifeDate:      "NOW",
						PreviousPatches: []PreviousPatches{
							{
								Release:            "X.Y.ZZZ",
								CherryPickDeadline: "2020-06-12",
								TargetDate:         "2020-06-17",
							},
							{
								Release:            "X.Y.XXX",
								CherryPickDeadline: "2020-05-15",
								TargetDate:         "2020-05-20",
								Note:               "honk",
							},
							{
								Release:            "X.Y.YYY",
								CherryPickDeadline: "2020-04-13",
								TargetDate:         "2020-04-16",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		fmt.Printf("Test case: %s\n", tc.name)
		out := parseSchedule(tc.schedule)
		require.Equal(t, out, expectedPatchSchedule)
	}
}

func TestParseReleaseSchedule(t *testing.T) {
	testcases := []struct {
		name     string
		schedule ReleaseSchedule
	}{
		{
			name: "test of release cycle of X.Y version",
			schedule: ReleaseSchedule{
				Releases: []Release{
					{
						Version: "X.Y",
						Timeline: []Timeline{
							{
								What:     "Testing-A",
								Who:      "tester",
								When:     "2020-06-17",
								Week:     "week 1",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-B",
								Who:      "tester",
								When:     "2020-06-19",
								Week:     "week 1",
								CISignal: "green",
							},
							{
								What:     "Testing-C",
								Who:      "tester",
								When:     "2020-06-20",
								Week:     "week 1",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-D",
								Who:      "tester",
								When:     "2020-06-21",
								Week:     "week 1",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-E",
								Who:      "tester",
								When:     "2020-06-22",
								Week:     "week 1",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-F",
								Who:      "tester",
								When:     "2020-06-25",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-G",
								Who:      "tester",
								When:     "2020-06-26",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-H",
								Who:      "tester",
								When:     "2020-06-27",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-I",
								Who:      "tester",
								When:     "2020-06-27",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-J",
								Who:      "tester",
								When:     "2020-06-27",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-K",
								Who:      "tester",
								When:     "2020-06-28",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     "true",
							},
							{
								What:     "Testing-L",
								Who:      "tester",
								When:     "2020-06-28",
								Week:     "week 2",
								CISignal: "green",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		fmt.Printf("Test case: %s\n", tc.name)
		out := parseReleaseSchedule(tc.schedule)
		require.Equal(t, out, expectedReleaseSchedule)
	}
}

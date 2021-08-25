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
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
)

// runs with `--type=patch` to retrun the patch schedule
func parseSchedule(patchSchedule PatchSchedule) string {
	output := []string{}
	output = append(output, "### Timeline\n")
	for _, releaseSchedule := range patchSchedule.Schedules {
		output = append(output, fmt.Sprintf("### %s\n", releaseSchedule.Release),
			fmt.Sprintf("Next patch release is **%s**\n", releaseSchedule.Next),
			fmt.Sprintf("End of Life for **%s** is **%s**\n", releaseSchedule.Release, releaseSchedule.EndOfLifeDate))

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"Patch Release", "Cherry Pick Deadline", "Target Date", "Note"})

		// Check if the next patch release is in the Previous Patch list, if yes dont read in the output
		if !patchReleaseInPreviousList(releaseSchedule.Next, releaseSchedule.PreviousPatches) {
			table.Append([]string{strings.TrimSpace(releaseSchedule.Next), strings.TrimSpace(releaseSchedule.CherryPickDeadline), strings.TrimSpace(releaseSchedule.TargetDate), ""})
		}

		for _, previous := range releaseSchedule.PreviousPatches {
			table.Append([]string{strings.TrimSpace(previous.Release), strings.TrimSpace(previous.CherryPickDeadline), strings.TrimSpace(previous.TargetDate), strings.TrimSpace(previous.Note)})
		}
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
		table.Render()

		output = append(output, tableString.String())
	}

	scheduleOut := strings.Join(output, "\n")

	logrus.Info("Schedule parsed")
	println(scheduleOut)

	return scheduleOut
}

// runs with `--type=release` to retrun the release cycle schedule
func parseReleaseSchedule(releaseSchedule ReleaseSchedule) string {
	k8VersionWithDot := releaseSchedule.Releases[0].Version
	k8VersionWithoutDot := removeDotfromVersion(releaseSchedule.Releases[0].Version)
	output := []string{}
	output = append(output, fmt.Sprintf("# Kubernetes %s\n", k8VersionWithDot), "#### Links\n",
		`
* [This document](https://git.k8s.io/sig-release/releases/release-`+k8VersionWithDot+`/README.md)
* [Release Team](https://github.com/kubernetes/sig-release/blob/master/releases/release-`+k8VersionWithDot+`/release-team.md)
* [Meeting Minutes](http://bit.ly/k8s`+k8VersionWithoutDot+`-releasemtg) (join [kubernetes-sig-release@] to receive meeting invites)
* [v`+k8VersionWithDot+` Release Calendar](https://bit.ly/k8s-release-cal)
* Contact: [#sig-release] on slack, [kubernetes-release-team@] on e-mail
* [Internal Contact Info][Internal Contact Info] (accessible only to members of [kubernetes-release-team@])
`, "#### Tracking docs\n",
		`
* [Enhancements Tracking Sheet](https://bit.ly/k8s`+k8VersionWithoutDot+`-enhancements)
* [Feature blog Tracking Sheet](TBD)
* [Bug Triage Tracking Sheet](TBD)
* CI Signal Report: TODO
* [Retrospective Document][Retrospective Document]
* [kubernetes/sig-release v`+k8VersionWithDot+` milestone](https://github.com/kubernetes/kubernetes/milestone/56)
`, "#### Guides\n",
		`
* [Targeting Issues and PRs to This Milestone](https://git.k8s.io/community/contributors/devel/sig-release/release.md)
* [Triaging and Escalating Test Failures](https://git.k8s.io/community/contributors/devel/sig-testing/testing.md#troubleshooting-a-failure)
`)

	arr := []Timeline{}
	for _, releaseSchedule := range releaseSchedule.Releases {
		for _, timeline := range releaseSchedule.Timeline {
			if timeline.Tldr == "true" {
				arr = append(arr, timeline)
			}
		}
	}
	output = append(output, "## TL;DR\n",
		`The `+k8VersionWithDot+` release cycle is proposed as follows:

- **`+arr[0].When+`**: `+arr[0].Week+`- Release cycle begins
- **`+arr[1].When+`**: `+arr[1].Week+`- [Production Readiness Soft Freeze](https://groups.google.com/g/kubernetes-sig-architecture/c/a6_y81N49aQ)
- **`+arr[2].When+`**: `+arr[2].Week+` - [Enhancements Freeze](../release_phases.md#enhancements-freeze)
- **`+arr[3].When+`**: `+arr[3].Week+` - [Code Freeze](../release_phases.md#code-freeze)
- **`+arr[4].When+`**: `+arr[4].Week+` - [Test Freeze](../release_phases.md#test-freeze)
- **`+arr[5].When+`**: `+arr[5].Week+` - Docs must be completed and reviewed
- **`+arr[6].When+`**: `+arr[6].Week+` - Kubernetes v`+k8VersionWithDot+`.0 released
- **`+arr[7].When+`**: `+arr[7].Week+` - [Release Retrospective][Retrospective Document] part 1
- **`+arr[8].When+`**: `+arr[8].Week+` - [Release Retrospective][Retrospective Document] part 2
- **`+arr[9].When+`**: `+arr[9].Week+` - [Release Retrospective][Retrospective Document] part 3
`, "## Timeline\n")
	for _, releaseSchedule := range releaseSchedule.Releases {
		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetAutoWrapText(false)
		table.SetHeader([]string{"**What**", "**Who**", "**When**", "**WEEK**", "**CI Signal**"})

		for _, timeline := range releaseSchedule.Timeline {
			table.Append([]string{strings.TrimSpace(timeline.What), strings.TrimSpace(timeline.Who), strings.TrimSpace(timeline.When), strings.TrimSpace(timeline.Week), strings.TrimSpace(timeline.CISignal), ""})
		}

		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
		table.Render()

		output = append(output, tableString.String())
	}

	output = append(output, "## Phases\n",
		`
Please refer to the [release phases document](../release_phases.md).

[k8s`+k8VersionWithDot+`-calendar]: https://bit.ly/k8s-release-cal
[Internal Contact Info]: https://bit.ly/k8s`+k8VersionWithoutDot+`-contacts
[Retrospective Document]: https://bit.ly/k8s`+k8VersionWithoutDot+`-retro

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
[`+k8VersionWithoutDot+`-blocking]: https://testgrid.k8s.io/sig-release-`+k8VersionWithDot+`-blocking#Summary
[exception requests]: ../EXCEPTIONS.md
[release phases document]: ../release_phases.md`)

	scheduleOut := strings.Join(output, "\n")

	logrus.Info("Release Schedule parsed")
	println(scheduleOut)

	return scheduleOut
}

func patchReleaseInPreviousList(a string, previousPatches []PreviousPatches) bool {
	for _, b := range previousPatches {
		if b.Release == a {
			return true
		}
	}
	return false
}

func removeDotfromVersion(a string) string {
	return strings.ReplaceAll(a, ".", "")
}

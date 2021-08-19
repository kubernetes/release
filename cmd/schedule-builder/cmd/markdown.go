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

func parseReleaseSchedule(releaseSchedule ReleaseSchedule) string {
	output := []string{}
	output = append(output, fmt.Sprintf("# Kubernetes %s\n", releaseSchedule.Releases[0].Version))

	output = append(output, fmt.Sprintf("#### Links\n"))
	for _, link := range releaseSchedule.Releases[0].Links {
		output = append(output, fmt.Sprintf("* [%s](%s)\n", link.Text, link.Href))
	}
	output = append(output, fmt.Sprintf("#### Tracking docs\n"))
	for _, trackingDocs := range releaseSchedule.Releases[0].TrackingDocs {
		output = append(output, fmt.Sprintf("* [%s](%s)\n", trackingDocs.Text, trackingDocs.Href))
	}

	output = append(output, fmt.Sprintf("#### Guides\n"))
	for _, guide := range releaseSchedule.Releases[0].Guides {
		output = append(output, fmt.Sprintf("* [%s](%s)\n", guide.Text, guide.Href))
	}

	output = append(output, fmt.Sprintf("## Timeline\n"))
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

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
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
)

//go:embed templates/*.tmpl
var tpls embed.FS

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
	type RelSched struct {
		K8VersionWithDot    string
		K8VersionWithoutDot string
		Arr                 []Timeline
		TimelineOutput      string
	}

	var relSched RelSched

	relSched.K8VersionWithDot = releaseSchedule.Releases[0].Version
	relSched.K8VersionWithoutDot = removeDotfromVersion(releaseSchedule.Releases[0].Version)
	relSched.Arr = []Timeline{}
	for _, releaseSchedule := range releaseSchedule.Releases {
		for _, timeline := range releaseSchedule.Timeline {
			if timeline.Tldr {
				relSched.Arr = append(relSched.Arr, timeline)
			}
		}
	}
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

		relSched.TimelineOutput = tableString.String()
	}

	scheduleOut := ProcessFile("templates/rel-schedule.tmpl", relSched)

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

// process applies the data structure 'vars' onto an already
// parsed template 't', and returns the resulting string.
func process(t *template.Template, vars interface{}) string {
	var tmplBytes bytes.Buffer

	err := t.Execute(&tmplBytes, vars)
	if err != nil {
		panic(err)
	}
	return tmplBytes.String()
}

func ProcessFile(fileName string, vars interface{}) string {
	tmpl, err := template.ParseFS(tpls, fileName)
	if err != nil {
		panic(err)
	}
	return process(tmpl, vars)
}

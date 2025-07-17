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
	"os"
	"strings"
	"text/template" // NOLINT // Mark text/template as not to be checked for producing yaml.
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/release-utils/util"
	"sigs.k8s.io/yaml"
)

//go:embed templates/*.tmpl
var tpls embed.FS

// runs with `--type=patch` to return the patch schedule.
func parsePatchSchedule(patchSchedule PatchSchedule) string {
	output := []string{}

	if len(patchSchedule.UpcomingReleases) > 0 {
		output = append(output, "### Upcoming Monthly Releases\n")
		tableString := &strings.Builder{}
		table := tablewriter.NewTable(tableString,
			tablewriter.WithConfig(tablewriter.Config{
				Header: tw.CellConfig{
					Alignment: tw.CellAlignment{Global: tw.AlignLeft},
				},
			}),
			tablewriter.WithHeader([]string{"Monthly Patch Release", "Cherry Pick Deadline", "Target Date"}),
			tablewriter.WithRenderer(renderer.NewMarkdown()),
			tablewriter.WithRendition(tw.Rendition{
				Symbols: tw.NewSymbols(tw.StyleMarkdown),
				Borders: tw.Border{
					Left:   tw.On,
					Top:    tw.Off,
					Right:  tw.On,
					Bottom: tw.Off,
				},
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenRows: tw.On,
					},
				},
			}),
			tablewriter.WithRowAutoWrap(tw.WrapNone),
		)

		for _, upcoming := range patchSchedule.UpcomingReleases {
			targetDate, err := time.Parse(refDate, upcoming.TargetDate)
			if err != nil {
				logrus.Errorf("Unable to parse upcoming target date %q: %v", upcoming.TargetDate, err)

				continue
			}

			_ = table.Append([]string{
				targetDate.Format(refDateMonthly),
				strings.TrimSpace(upcoming.CherryPickDeadline),
				strings.TrimSpace(upcoming.TargetDate),
			})
		}

		_ = table.Render()

		output = append(output, tableString.String())
	}

	output = append(output, "### Timeline\n")

	for _, releaseSchedule := range patchSchedule.Schedules {
		output = append(output, fmt.Sprintf("### %s\n", releaseSchedule.Release),
			fmt.Sprintf("Next patch release is **%s**\n", releaseSchedule.Next.Release),
			fmt.Sprintf("**%s** enters maintenance mode on **%s** and End of Life is on **%s**.\n",
				releaseSchedule.Release, releaseSchedule.MaintenanceModeStartDate, releaseSchedule.EndOfLifeDate,
			),
		)

		tableString := &strings.Builder{}
		table := tablewriter.NewTable(tableString,
			tablewriter.WithConfig(tablewriter.Config{
				Header: tw.CellConfig{
					Alignment: tw.CellAlignment{Global: tw.AlignLeft},
				},
				Row: tw.CellConfig{
					Alignment: tw.CellAlignment{Global: tw.AlignLeft},
				},
			}),
			tablewriter.WithHeader([]string{"Patch Release", "Cherry Pick Deadline", "Target Date", "Note"}),
			tablewriter.WithRenderer(renderer.NewMarkdown()),
			tablewriter.WithRendition(tw.Rendition{
				Symbols: tw.NewSymbols(tw.StyleMarkdown),
				Borders: tw.Border{
					Left:   tw.On,
					Top:    tw.Off,
					Right:  tw.On,
					Bottom: tw.Off,
				},
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenRows: tw.On,
					},
				},
			}),
			tablewriter.WithRowAutoWrap(tw.WrapNone),
		)

		// Check if the next patch release is in the Previous Patch list, if yes dont read in the output
		if !patchReleaseInPreviousList(releaseSchedule.Next.Release, releaseSchedule.PreviousPatches) {
			_ = table.Append([]string{strings.TrimSpace(releaseSchedule.Next.Release), strings.TrimSpace(releaseSchedule.Next.CherryPickDeadline), strings.TrimSpace(releaseSchedule.Next.TargetDate), ""})
		}

		for _, previous := range releaseSchedule.PreviousPatches {
			_ = table.Append([]string{strings.TrimSpace(previous.Release), strings.TrimSpace(previous.CherryPickDeadline), strings.TrimSpace(previous.TargetDate), strings.TrimSpace(previous.Note)})
		}

		_ = table.Render()

		output = append(output, tableString.String())
	}

	scheduleOut := strings.Join(output, "\n")

	logrus.Info("Schedule parsed")

	return scheduleOut
}

// runs with `--type=release` to return the release cycle schedule.
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
		table := tablewriter.NewTable(tableString,
			tablewriter.WithConfig(tablewriter.Config{
				Header: tw.CellConfig{
					Alignment: tw.CellAlignment{Global: tw.AlignCenter},
				},
				Row: tw.CellConfig{
					Alignment: tw.CellAlignment{Global: tw.AlignLeft},
				},
			}),
			tablewriter.WithHeader([]string{"**What**", "**Who**", "**When**", "**WEEK**", "**CI Signal**"}),
			tablewriter.WithRenderer(renderer.NewMarkdown()),
			tablewriter.WithRendition(tw.Rendition{
				Symbols: tw.NewSymbols(tw.StyleMarkdown),
				Borders: tw.Border{
					Left:   tw.On,
					Top:    tw.Off,
					Right:  tw.On,
					Bottom: tw.Off,
				},
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenRows:    tw.On,
						BetweenColumns: tw.Off,
					},
				},
			}),
			tablewriter.WithRowAutoWrap(tw.WrapNone),
		)

		for _, timeline := range releaseSchedule.Timeline {
			_ = table.Append([]string{strings.TrimSpace(timeline.What), strings.TrimSpace(timeline.Who), strings.TrimSpace(timeline.When), strings.TrimSpace(timeline.Week), strings.TrimSpace(timeline.CISignal), ""})
		}

		_ = table.Render()

		relSched.TimelineOutput = tableString.String()
	}

	scheduleOut := processFile("templates/rel-schedule.tmpl", relSched)

	logrus.Info("Release Schedule parsed")

	return scheduleOut
}

func patchReleaseInPreviousList(a string, previousPatches []*PatchRelease) bool {
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

func processFile(fileName string, vars interface{}) string {
	tmpl, err := template.ParseFS(tpls, fileName)
	if err != nil {
		panic(err)
	}

	return process(tmpl, vars)
}

const (
	refDate        = "2006-01-02"
	refDateMonthly = "January 2006"
	markdownHelp   = `# Use "schedule-builder" to maintain this file:
# https://github.com/kubernetes/release/tree/master/cmd/schedule-builder
# For example by running:
# schedule-builder -uc data/releases/schedule.yaml -e data/releases/eol.yaml
---
`
)

func updatePatchSchedule(refTime time.Time, schedule PatchSchedule, eolBranches EolBranches, filePath, eolFilePath string) error {
	removeSchedules := []int{}

	for i, sched := range schedule.Schedules {
		for {
			if sched.Next == nil {
				logrus.Warnf("Next release not set for %s, skipping", sched.Release)

				break
			}

			eolDate, err := time.Parse(refDate, sched.EndOfLifeDate)
			if err != nil {
				return fmt.Errorf("parse end of life date: %w", err)
			}

			if refTime.After(eolDate) {
				if eolFilePath == "" {
					logrus.Infof("Skipping end of life release: %s", sched.Release)

					break
				}

				logrus.Infof("Moving %s to end of life", sched.Release)
				eolBranches.Branches = append([]*EolBranch{{
					Release:           sched.Release,
					FinalPatchRelease: sched.Next.Release,
					EndOfLifeDate:     sched.Next.TargetDate,
				}}, eolBranches.Branches...)

				removeSchedules = append(removeSchedules, i)

				break
			}

			targetDate, err := time.Parse(refDate, sched.Next.TargetDate)
			if err != nil {
				return fmt.Errorf("parse target date: %w", err)
			}

			if targetDate.After(refTime) {
				break
			}

			// Copy the release to the previousPatches section
			sched.PreviousPatches = append([]*PatchRelease{sched.Next}, sched.PreviousPatches...)

			// Create a new next release
			nextReleaseVersion, err := util.TagStringToSemver(sched.Next.Release)
			if err != nil {
				return fmt.Errorf("parse semver version: %w", err)
			}

			if err := nextReleaseVersion.IncrementPatch(); err != nil {
				return fmt.Errorf("increment patch version: %w", err)
			}

			cherryPickDeadline, err := time.Parse(refDate, sched.Next.CherryPickDeadline)
			if err != nil {
				return fmt.Errorf("parse cherry pick deadline: %w", err)
			}

			cherryPickDeadlinePlusOneMonth := cherryPickDeadline.AddDate(0, 1, 0)

			var (
				newCherryPickDeadline time.Time
				newTargetDate         time.Time
				found                 bool
			)

			for _, u := range schedule.UpcomingReleases {
				if u == nil {
					continue
				}

				upcomingCherryPickDeadline, err := time.Parse(refDate, u.CherryPickDeadline)
				if err != nil {
					return fmt.Errorf("parse upcoming cherry pick deadline: %w", err)
				}

				if cherryPickDeadlinePlusOneMonth.Month() == upcomingCherryPickDeadline.Month() &&
					cherryPickDeadlinePlusOneMonth.Year() == upcomingCherryPickDeadline.Year() {
					newCherryPickDeadline = upcomingCherryPickDeadline

					newTargetDate, err = time.Parse(refDate, u.TargetDate)
					if err != nil {
						return fmt.Errorf("parse upcoming release date: %w", err)
					}

					found = true

					break
				}
			}

			if !found {
				cherryPickDay := firstFriday(cherryPickDeadlinePlusOneMonth)
				newCherryPickDeadline = time.Date(cherryPickDeadlinePlusOneMonth.Year(), cherryPickDeadlinePlusOneMonth.Month(), cherryPickDay, 0, 0, 0, 0, time.UTC)

				targetDatePlusOneMonth := targetDate.AddDate(0, 1, 0)
				targetDateDay := secondTuesday(targetDatePlusOneMonth)
				newTargetDate = time.Date(targetDatePlusOneMonth.Year(), targetDatePlusOneMonth.Month(), targetDateDay, 0, 0, 0, 0, time.UTC)
			}

			sched.Next = &PatchRelease{
				Release:            nextReleaseVersion.String(),
				CherryPickDeadline: newCherryPickDeadline.Format(refDate),
				TargetDate:         newTargetDate.Format(refDate),
			}

			logrus.Infof("Adding release schedule: %+v", sched.Next)
		}
	}

	newSchedules := []*Schedule{}

	for i, sched := range schedule.Schedules {
		appendItem := true

		for _, k := range removeSchedules {
			if i == k {
				appendItem = false

				break
			}
		}

		if appendItem {
			newSchedules = append(newSchedules, sched)
		}
	}

	schedule.Schedules = newSchedules

	newUpcomingReleases := []*PatchRelease{}
	latestDate := refTime

	for _, upcomingRelease := range schedule.UpcomingReleases {
		upcomingTargetDate, err := time.Parse(refDate, upcomingRelease.TargetDate)
		if err != nil {
			return fmt.Errorf("parse upcoming release target date: %w", err)
		}

		if refTime.After(upcomingTargetDate) {
			logrus.Infof("Skipping outdated upcoming release for %s", upcomingRelease.TargetDate)

			continue
		}

		logrus.Infof("Using existing upcoming release for %s", upcomingRelease.TargetDate)
		newUpcomingReleases = append(newUpcomingReleases, upcomingRelease)
		latestDate = upcomingTargetDate
	}

	for {
		if len(newUpcomingReleases) >= 3 {
			logrus.Infof("Got 3 new upcoming releases, not adding any more")

			break
		}

		latestDate = latestDate.AddDate(0, 1, 0)
		cherryPickDay := firstFriday(latestDate)
		targetDateDay := secondTuesday(latestDate)
		nextCherryPickDeadline := time.Date(latestDate.Year(), latestDate.Month(), cherryPickDay, 0, 0, 0, 0, time.UTC)
		nextTargetDate := time.Date(latestDate.Year(), latestDate.Month(), targetDateDay, 0, 0, 0, 0, time.UTC)

		logrus.Infof("Adding new upcoming release for %s", nextTargetDate.Format(refDateMonthly))

		newUpcomingReleases = append(newUpcomingReleases, &PatchRelease{
			CherryPickDeadline: nextCherryPickDeadline.Format(refDate),
			TargetDate:         nextTargetDate.Format(refDate),
		})
	}

	schedule.UpcomingReleases = newUpcomingReleases

	yamlBytes, err := yaml.Marshal(schedule)
	if err != nil {
		return fmt.Errorf("marshal schedule YAML: %w", err)
	}

	yamlBytes = append([]byte(markdownHelp), yamlBytes...)

	//nolint:gocritic,gosec
	if err := os.WriteFile(filePath, yamlBytes, 0o644); err != nil {
		return fmt.Errorf("write schedule YAML: %w", err)
	}

	if eolFilePath != "" {
		logrus.Infof("Writing end of life branches: %s", eolFilePath)

		yamlBytes, err := yaml.Marshal(eolBranches)
		if err != nil {
			return fmt.Errorf("marshal end of life YAML: %w", err)
		}

		yamlBytes = append([]byte(markdownHelp), yamlBytes...)

		//nolint:gocritic,gosec
		if err := os.WriteFile(eolFilePath, yamlBytes, 0o644); err != nil {
			return fmt.Errorf("write end of life YAML: %w", err)
		}
	}

	logrus.Infof("Wrote schedule YAML to: %v", filePath)

	return nil
}

func secondTuesday(t time.Time) int {
	return firstMonday(t) + 8
}

func firstFriday(t time.Time) int {
	return firstMonday(t) + 4
}

func firstMonday(from time.Time) int {
	t := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)

	return (8-int(t.Weekday()))%7 + 1
}

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
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/yaml"
)

const expectedPatchSchedule = `### Upcoming Monthly Releases

| MONTHLY PATCH RELEASE | CHERRY PICK DEADLINE | TARGET DATE |
|:----------------------|:---------------------|:------------|
| June 2020             | 2020-06-12           | 2020-06-17  |

### Timeline

### X.Y

Next patch release is **X.Y.ZZZ**

**X.Y** enters maintenance mode on **THEN** and End of Life is on **NOW**.

| PATCH RELEASE | CHERRY PICK DEADLINE | TARGET DATE | NOTE |
|:--------------|:---------------------|:------------|:-----|
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

- **2020-06-17**: week 1 - Release cycle begins
- **2020-06-20**: week 1 - [Production Readiness Soft Freeze](https://groups.google.com/g/kubernetes-sig-architecture/c/a6_y81N49aQ)
- **2020-06-21**: week 1 - [Enhancements Freeze](../release_phases.md#enhancements-freeze)
- **2020-06-22**: week 1 - [Code Freeze](../release_phases.md#code-freeze)
- **2020-06-25**: week 2 - [Test Freeze](../release_phases.md#test-freeze)
- **2020-06-26**: week 2 - Docs must be completed and reviewed
- **2020-06-27**: week 2 - Kubernetes vX.Y.0 released
- **2020-06-27**: week 2 - [Release Retrospective][Retrospective Document] part 1
- **2020-06-27**: week 2 - [Release Retrospective][Retrospective Document] part 2
- **2020-06-28**: week 2 - [Release Retrospective][Retrospective Document] part 3

## Timeline

| ** WHAT ** | ** WHO ** | ** WHEN ** | ** WEEK ** | ** CI SIGNAL ** |   |
|:-----------|:----------|:-----------|:-----------|:----------------|:--|
| Testing-A  | tester    | 2020-06-17 | week 1     | green           |   |
| Testing-B  | tester    | 2020-06-19 | week 1     | green           |   |
| Testing-C  | tester    | 2020-06-20 | week 1     | green           |   |
| Testing-D  | tester    | 2020-06-21 | week 1     | green           |   |
| Testing-E  | tester    | 2020-06-22 | week 1     | green           |   |
| Testing-F  | tester    | 2020-06-25 | week 2     | green           |   |
| Testing-G  | tester    | 2020-06-26 | week 2     | green           |   |
| Testing-H  | tester    | 2020-06-27 | week 2     | green           |   |
| Testing-I  | tester    | 2020-06-27 | week 2     | green           |   |
| Testing-J  | tester    | 2020-06-27 | week 2     | green           |   |
| Testing-K  | tester    | 2020-06-28 | week 2     | green           |   |
| Testing-L  | tester    | 2020-06-28 | week 2     | green           |   |

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
[release phases document]: ../release_phases.md
`

func TestParsePatchSchedule(t *testing.T) {
	testcases := []struct {
		name     string
		schedule PatchSchedule
	}{
		{
			name: "next patch is not in previous patch list",
			schedule: PatchSchedule{
				Schedules: []*Schedule{
					{
						Release: "X.Y",
						Next: &PatchRelease{
							Release:            "X.Y.ZZZ",
							CherryPickDeadline: "2020-06-12",
							TargetDate:         "2020-06-17",
						},
						EndOfLifeDate:            "NOW",
						MaintenanceModeStartDate: "THEN",
						PreviousPatches: []*PatchRelease{
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
				UpcomingReleases: []*PatchRelease{
					{
						CherryPickDeadline: "2020-06-12",
						TargetDate:         "2020-06-17",
					},
				},
			},
		},
		{
			name: "next patch is in previous patch list",
			schedule: PatchSchedule{
				Schedules: []*Schedule{
					{
						Release: "X.Y",
						Next: &PatchRelease{
							Release:            "X.Y.ZZZ",
							CherryPickDeadline: "2020-06-12",
							TargetDate:         "2020-06-17",
						},
						EndOfLifeDate:            "NOW",
						MaintenanceModeStartDate: "THEN",
						PreviousPatches: []*PatchRelease{
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
				UpcomingReleases: []*PatchRelease{
					{
						CherryPickDeadline: "2020-06-12",
						TargetDate:         "2020-06-17",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		fmt.Printf("Test case: %s\n", tc.name)
		out := parsePatchSchedule(tc.schedule)
		require.Equal(t, expectedPatchSchedule, out)
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
								Tldr:     true,
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
								Tldr:     true,
							},
							{
								What:     "Testing-D",
								Who:      "tester",
								When:     "2020-06-21",
								Week:     "week 1",
								CISignal: "green",
								Tldr:     true,
							},
							{
								What:     "Testing-E",
								Who:      "tester",
								When:     "2020-06-22",
								Week:     "week 1",
								CISignal: "green",
								Tldr:     true,
							},
							{
								What:     "Testing-F",
								Who:      "tester",
								When:     "2020-06-25",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     true,
							},
							{
								What:     "Testing-G",
								Who:      "tester",
								When:     "2020-06-26",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     true,
							},
							{
								What:     "Testing-H",
								Who:      "tester",
								When:     "2020-06-27",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     true,
							},
							{
								What:     "Testing-I",
								Who:      "tester",
								When:     "2020-06-27",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     true,
							},
							{
								What:     "Testing-J",
								Who:      "tester",
								When:     "2020-06-27",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     true,
							},
							{
								What:     "Testing-K",
								Who:      "tester",
								When:     "2020-06-28",
								Week:     "week 2",
								CISignal: "green",
								Tldr:     true,
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
		out, err := parseReleaseSchedule(tc.schedule)
		require.NoError(t, err)
		require.Equal(t, expectedReleaseSchedule, out)
	}
}

func TestUpdatePatchSchedule(t *testing.T) {
	for _, tc := range []struct {
		name                            string
		refTime                         time.Time
		givenSchedule, expectedSchedule PatchSchedule
		expectedEolBranches             EolBranches
	}{
		{
			name:    "succeed to update the schedule",
			refTime: time.Date(2024, 4, 3, 0, 0, 0, 0, time.UTC),
			givenSchedule: PatchSchedule{
				Schedules: []*Schedule{
					{ // Needs multiple updates
						Release: "1.30",
						Next: &PatchRelease{
							Release:            "1.30.1",
							CherryPickDeadline: "2024-01-05",
							TargetDate:         "2024-01-09",
						},
						EndOfLifeDate:            "2025-01-01",
						MaintenanceModeStartDate: "2024-12-01",
					},
					{ // next not set
						Release:       "1.29",
						EndOfLifeDate: "2025-02-28",
					},
					{ // next and previous not set & EOL
						Release:       "1.17",
						EndOfLifeDate: "2021-01-13",
					},
					{ // next not set & EOL
						Release:       "1.23",
						EndOfLifeDate: "2023-02-28",
						PreviousPatches: []*PatchRelease{
							{
								Release:            "1.23.17",
								CherryPickDeadline: "2023-02-10",
								TargetDate:         "2023-02-15",
							},
							{
								Release:            "1.23.16",
								CherryPickDeadline: "2023-01-13",
								TargetDate:         "2023-01-18",
							},
							{
								Release:            "1.23.15",
								CherryPickDeadline: "2022-12-02",
								TargetDate:         "2022-12-08",
							},
						},
					},
					{ // EOL
						Release:       "1.20",
						EndOfLifeDate: "2023-01-01",
						Next: &PatchRelease{
							Release:            "1.20.10",
							CherryPickDeadline: "2023-12-08",
							TargetDate:         "2023-12-12",
						},
					},
				},
				UpcomingReleases: []*PatchRelease{
					{
						CherryPickDeadline: "2024-03-08",
						TargetDate:         "2024-03-12",
					},
					{
						CherryPickDeadline: "2024-04-12",
						TargetDate:         "2024-04-17",
					},
					{
						CherryPickDeadline: "2024-05-10",
						TargetDate:         "2024-05-14",
					},
				},
			},
			expectedSchedule: PatchSchedule{
				Schedules: []*Schedule{
					{
						Release: "1.30",
						Next: &PatchRelease{
							Release:            "1.30.4",
							CherryPickDeadline: "2024-04-12",
							TargetDate:         "2024-04-17",
						},
						EndOfLifeDate:            "2025-01-01",
						MaintenanceModeStartDate: "2024-12-01",
						PreviousPatches: []*PatchRelease{
							{
								Release:            "1.30.3",
								CherryPickDeadline: "2024-03-08",
								TargetDate:         "2024-03-12",
							},
							{
								Release:            "1.30.2",
								CherryPickDeadline: "2024-02-09",
								TargetDate:         "2024-02-13",
							},
							{
								Release:            "1.30.1",
								CherryPickDeadline: "2024-01-05",
								TargetDate:         "2024-01-09",
							},
						},
					},
					{
						Release:       "1.29",
						EndOfLifeDate: "2025-02-28",
					},
					{
						Release:       "1.17",
						EndOfLifeDate: "2021-01-13",
					},
				},
				UpcomingReleases: []*PatchRelease{
					{
						CherryPickDeadline: "2024-04-12",
						TargetDate:         "2024-04-17",
					},
					{
						CherryPickDeadline: "2024-05-10",
						TargetDate:         "2024-05-14",
					},
					{
						CherryPickDeadline: "2024-06-07",
						TargetDate:         "2024-06-11",
					},
				},
			},
			expectedEolBranches: EolBranches{
				Branches: []*EolBranch{
					{
						Release:           "1.20",
						FinalPatchRelease: "1.20.10",
						EndOfLifeDate:     "2023-01-01",
					},
					{
						Release:           "1.23",
						FinalPatchRelease: "1.23.17",
						EndOfLifeDate:     "2023-02-28",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scheduleFile, err := os.CreateTemp(t.TempDir(), "schedule-")
			require.NoError(t, err)
			require.NoError(t, scheduleFile.Close())

			eolFile, err := os.CreateTemp(t.TempDir(), "eol-")
			require.NoError(t, err)
			require.NoError(t, eolFile.Close())

			require.NoError(t, updatePatchSchedule(tc.refTime, tc.givenSchedule, EolBranches{}, scheduleFile.Name(), eolFile.Name()))

			scheduleYamlBytes, err := os.ReadFile(scheduleFile.Name()) //nolint:gosec // G703 - temp file path is safe
			require.NoError(t, err)

			patchRes := PatchSchedule{}
			require.NoError(t, yaml.UnmarshalStrict(scheduleYamlBytes, &patchRes))

			assert.Equal(t, tc.expectedSchedule, patchRes)

			eolYamlBytes, err := os.ReadFile(eolFile.Name()) //nolint:gosec // G703 - temp file path is safe
			require.NoError(t, err)

			eolRes := EolBranches{}
			require.NoError(t, yaml.UnmarshalStrict(eolYamlBytes, &eolRes))

			assert.Equal(t, tc.expectedEolBranches, eolRes)
		})
	}
}

func TestUpdatePatchScheduleMaintenanceWindow(t *testing.T) {
	refTime := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)

	// Four supported minors with the oldest (1.30) sitting in its
	// maintenance window relative to refTime. Each Next.TargetDate is
	// after refTime so the per-schedule patch-bump loop short-circuits.
	baseSchedule := func() PatchSchedule {
		return PatchSchedule{
			Schedules: []*Schedule{
				{
					Release: "1.33",
					Next: &PatchRelease{
						Release:            "1.33.1",
						CherryPickDeadline: "2026-06-05",
						TargetDate:         "2026-06-09",
					},
					EndOfLifeDate:            "2027-06-01",
					MaintenanceModeStartDate: "2027-03-01",
				},
				{
					Release: "1.32",
					Next: &PatchRelease{
						Release:            "1.32.5",
						CherryPickDeadline: "2026-06-05",
						TargetDate:         "2026-06-09",
					},
					EndOfLifeDate:            "2027-02-01",
					MaintenanceModeStartDate: "2026-11-01",
				},
				{
					Release: "1.31",
					Next: &PatchRelease{
						Release:            "1.31.10",
						CherryPickDeadline: "2026-06-05",
						TargetDate:         "2026-06-09",
					},
					EndOfLifeDate:            "2026-10-01",
					MaintenanceModeStartDate: "2026-08-01",
				},
				{
					Release: "1.30",
					Next: &PatchRelease{
						Release:            "1.30.15",
						CherryPickDeadline: "2026-06-05",
						TargetDate:         "2026-06-09",
					},
					EndOfLifeDate:            "2026-07-01",
					MaintenanceModeStartDate: "2026-04-01",
				},
			},
		}
	}

	for _, tc := range []struct {
		name              string
		confirm           func(*Schedule) (bool, error)
		mutate            func(PatchSchedule) PatchSchedule
		wantUpcomingCount int
		wantErr           string
	}{
		{
			name:              "overlap accepted - extends upcoming to 4",
			confirm:           func(*Schedule) (bool, error) { return true, nil },
			wantUpcomingCount: 4,
		},
		{
			name:              "overlap declined - upcoming stays at 3",
			confirm:           func(*Schedule) (bool, error) { return false, nil },
			wantUpcomingCount: 3,
		},
		{
			name: "TBD maintenance date - prompter never called",
			confirm: func(*Schedule) (bool, error) {
				t.Fatal("confirmExtraUpcoming must not be called when MaintenanceModeStartDate is TBD")

				return false, nil
			},
			mutate: func(s PatchSchedule) PatchSchedule {
				s.Schedules[3].MaintenanceModeStartDate = "TBD"

				return s
			},
			wantUpcomingCount: 3,
		},
		{
			name:    "prompt error bubbles up",
			confirm: func(*Schedule) (bool, error) { return false, errors.New("boom") },
			wantErr: "boom",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			orig := confirmExtraUpcoming

			t.Cleanup(func() {
				confirmExtraUpcoming = orig
			})

			confirmExtraUpcoming = tc.confirm

			sched := baseSchedule()
			if tc.mutate != nil {
				sched = tc.mutate(sched)
			}

			scheduleFile, err := os.CreateTemp(t.TempDir(), "schedule-")
			require.NoError(t, err)
			require.NoError(t, scheduleFile.Close())

			err = updatePatchSchedule(refTime, sched, EolBranches{}, scheduleFile.Name(), "")
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			scheduleYamlBytes, err := os.ReadFile(scheduleFile.Name()) //nolint:gosec // G304 - temp file path is safe
			require.NoError(t, err)

			patchRes := PatchSchedule{}
			require.NoError(t, yaml.UnmarshalStrict(scheduleYamlBytes, &patchRes))

			assert.Len(t, patchRes.UpcomingReleases, tc.wantUpcomingCount)
			assert.Len(t, patchRes.Schedules, 4, "schedules[] count must be unchanged")
		})
	}
}

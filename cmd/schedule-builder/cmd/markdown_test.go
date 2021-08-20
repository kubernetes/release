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

* [XYZ](https://example.com/XYZ)

* [ABC](https://example.com/ABC)

#### Tracking docs

* [DEF](https://example.com/DEF)

* [GHI](https://example.com/GHI)

#### Guides

* [JKL](https://example.com/JKL)

* [MNO](https://example.com/MNO)

## Timeline

| **WHAT**  | **WHO** |  **WHEN**  | **WEEK** | **CI SIGNAL** |  |
|-----------|---------|------------|----------|---------------|--|
| Testing-A | tester  | 2020-06-17 | week 1   | green         |  |
| Testing-B | tester  | 2020-06-19 | week 1   | green         |  |
`

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
						Links: []Link{
							{
								Href: "https://example.com/XYZ",
								Text: "XYZ",
							},
							{
								Href: "https://example.com/ABC",
								Text: "ABC",
							},
						},
						TrackingDocs: []TrackingDoc{
							{
								Href: "https://example.com/DEF",
								Text: "DEF",
							},
							{
								Href: "https://example.com/GHI",
								Text: "GHI",
							},
						},
						Guides: []Guide{
							{
								Href: "https://example.com/JKL",
								Text: "JKL",
							},
							{
								Href: "https://example.com/MNO",
								Text: "MNO",
							},
						},
						Timeline: []Timeline{
							{
								What:     "Testing-A",
								Who:      "tester",
								When:     "2020-06-17",
								Week:     "week 1",
								CISignal: "green",
							},
							{
								What:     "Testing-B",
								Who:      "tester",
								When:     "2020-06-19",
								Week:     "week 1",
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

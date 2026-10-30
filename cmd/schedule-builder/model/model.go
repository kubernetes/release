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

package model

// PatchSchedule main struct to hold the schedules.
type PatchSchedule struct {
	UpcomingReleases []*PatchRelease `json:"upcoming_releases,omitempty" yaml:"upcoming_releases,omitempty"`
	Schedules        []*Schedule     `json:"schedules,omitempty"         yaml:"schedules,omitempty"`
}

// PatchRelease struct to define the patch schedules.
type PatchRelease struct {
	Release            string `json:"release,omitempty"            yaml:"release,omitempty"`
	CherryPickDeadline string `json:"cherryPickDeadline,omitempty" yaml:"cherryPickDeadline,omitempty"`
	TargetDate         string `json:"targetDate,omitempty"         yaml:"targetDate,omitempty"`
	Note               string `json:"note,omitempty"               yaml:"note,omitempty"`
}

// Schedule struct to define the release schedule for a specific version.
type Schedule struct {
	Release                  string          `json:"release,omitempty"                  yaml:"release,omitempty"`
	ReleaseDate              string          `json:"releaseDate,omitempty"              yaml:"releaseDate,omitempty"`
	Next                     *PatchRelease   `json:"next,omitempty"                     yaml:"next,omitempty"`
	EndOfLifeDate            string          `json:"endOfLifeDate,omitempty"            yaml:"endOfLifeDate,omitempty"`
	MaintenanceModeStartDate string          `json:"maintenanceModeStartDate,omitempty" yaml:"maintenanceModeStartDate,omitempty"`
	PreviousPatches          []*PatchRelease `json:"previousPatches,omitempty"          yaml:"previousPatches,omitempty"`
}

// EolBranches is main struct to hold the end of life branches.
type EolBranches struct {
	Branches []*EolBranch `json:"branches,omitempty" yaml:"branches,omitempty"`
}

// EolBranch struct to define the end of life release branches.
type EolBranch struct {
	Release           string `json:"release,omitempty"           yaml:"release,omitempty"`
	FinalPatchRelease string `json:"finalPatchRelease,omitempty" yaml:"finalPatchRelease,omitempty"`
	EndOfLifeDate     string `json:"endOfLifeDate,omitempty"     yaml:"endOfLifeDate,omitempty"`
	Note              string `json:"note,omitempty"              yaml:"note,omitempty"`
}

type ReleaseSchedule struct {
	Releases []Release `yaml:"releases"`
}

type Release struct {
	Version  string     `yaml:"version"`
	Timeline []Timeline `yaml:"timeline"`
}

type Timeline struct {
	What     string `yaml:"what"`
	Who      string `yaml:"who"`
	When     string `yaml:"when"`
	Week     string `yaml:"week"`
	CISignal string `yaml:"ciSignal"`
	Tldr     bool   `yaml:"tldr"`
}

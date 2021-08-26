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

// PatchSchedule main struct to hold the schedules
type PatchSchedule struct {
	Schedules []Schedule `yaml:"schedules"`
}

// PreviousPatches struct to define the old patch schedules
type PreviousPatches struct {
	Release            string `yaml:"release"`
	CherryPickDeadline string `yaml:"cherryPickDeadline"`
	TargetDate         string `yaml:"targetDate"`
	Note               string `yaml:"note"`
}

// Schedule struct to define the release schedule for a specific version
type Schedule struct {
	Release            string            `yaml:"release"`
	Next               string            `yaml:"next"`
	CherryPickDeadline string            `yaml:"cherryPickDeadline"`
	TargetDate         string            `yaml:"targetDate"`
	EndOfLifeDate      string            `yaml:"endOfLifeDate"`
	PreviousPatches    []PreviousPatches `yaml:"previousPatches"`
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

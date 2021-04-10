/*
Copyright 2021 The Kubernetes Authors.

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

package docs

import (
	"reflect"
	"testing"
)

func TestStructureData(t *testing.T) {
	indexes := []int{0, 6, 8, 9, 10}

	mockSpreadSheetData := [][]interface{}{
		// Complete data
		{"19", "Cronjobs", "Tracked", "Somtochi", "Has Docs", "Complete(Merge)", "https://github.com/kubernetes/website/pull/24885", "DONE", "apps", "person", "https://github.com/kubernetes/kubernetes/pull/93370"},
		// Missing k/k pr
		{"18", "Kubelet Credential Provider", "Tracked", "Somtochi", "Has Docs", "Complete(Merge)", "https://github.com/kubernetes/website/pull/24885", "DONE", "apps", "person1"},
		{"17", "Kubelet CRI support", "Tracked", "Somtochi", "None", "Complete(Merge)", "", "DONE", "node", "person2"},
	}

	expectedData := []*PrData{
		{
			KEP:          19,
			KubernetesPR: 93370,
			WebsitePR:    24885,
			Sig:          "apps",
			Assignee:     "person",
		},
		{
			KEP:       18,
			WebsitePR: 24885,
			Sig:       "apps",
			Assignee:  "person1",
		},
		{
			KEP:      17,
			Sig:      "node",
			Assignee: "person2",
		},
	}

	prData, err := StructureData(mockSpreadSheetData, indexes)
	if err != nil {
		t.Fatalf("error structuring spreadsheet data %v:", err)
	}

	for i, data := range prData {
		if !reflect.DeepEqual(data, expectedData[i]) {
			t.Errorf("expected %v, got %v", data, expectedData[i])
		}
	}
}

func TestConvertColumnLettersToInt(t *testing.T) {
	tests := []struct {
		letters []string
		ids     []int
	}{
		{
			letters: []string{"A", "B", "C", "D", "E"},
			ids:     []int{0, 1, 2, 3, 4},
		},
		{
			letters: []string{"A", "G", "I", "J", "K"},
			ids:     []int{0, 6, 8, 9, 10},
		},
		{
			letters: []string{"q", "r", "b", "j", "x"},
			ids:     []int{16, 17, 1, 9, 23},
		},
	}

	for _, tt := range tests {
		actual, err := ConvertColumnLettersToInt(tt.letters)
		if err != nil {
			t.Fatalf("unexpected err: %s", err)
		}

		if !reflect.DeepEqual(actual, tt.ids) {
			t.Errorf("expected %v, got %v", tt.ids, actual)
		}
	}
}

func TestToInt(t *testing.T) {
	// Test case sensitivity too
	columnHeaders := "ABCDEFGHIJKLMNOPqrstuvwxyz"
	for i, letter := range columnHeaders {
		id, err := toInt(string(letter))
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if id != i {
			t.Errorf("expected id of column %v to be %v but got %v", letter, i, id)
		}
	}
}

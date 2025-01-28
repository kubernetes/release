/*
Copyright 2024 The Kubernetes Authors.

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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"

	"k8s.io/release/pkg/notes"
)

type releaseNotesValidateOptions struct {
	pathToReleaseNotes string
}

var releaseNotesValidateOpts = &releaseNotesValidateOptions{}

func init() {
	validateCmd.PersistentFlags().StringVar(
		&releaseNotesValidateOpts.pathToReleaseNotes,
		"path-to-release-notes",
		"",
		"The path to the release notes to validate. Can be a top level directory or a specific file.",
	)

	// Add the validation subcommand to the release-notes command
	releaseNotesCmd.AddCommand(validateCmd)
}

// validate represents the subcommand for `krel release-notes validate`.
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "The subcommand for validating release notes for the Release Notes subteam of SIG Release",
	Long: `krel release-notes validate <path-to-release-notes>

The 'validate' subcommand of krel has been developed to:

1. Check release notes maps for valid yaml.

2. Check release notes maps for valid punctuation.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure exactly one argument is provided
		if releaseNotesValidateOpts.pathToReleaseNotes == "" {
			return errors.New("path to release notes must be provided via --path-to-release-notes")
		}

		// Run the PR creation function
		return runValidateReleaseNotes(releaseNotesValidateOpts.pathToReleaseNotes)
	},
}

func runValidateReleaseNotes(releaseNotesPath string) (err error) {
	// Ensure the path is not empty
	if releaseNotesPath == "" {
		return errors.New("release notes path cannot be empty")
	}

	// Check if the directory exists
	if _, err := os.Stat(releaseNotesPath); os.IsNotExist(err) {
		return fmt.Errorf("release notes path %s does not exist", releaseNotesPath)
	}

	// Validate the YAML files in the directory
	err = filepath.Walk(releaseNotesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Only process YAML files
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			fmt.Printf("Validating YAML file: %s\n", path)

			// Validate YAML
			if err := ValidateYamlMap(path); err != nil {
				return fmt.Errorf("validating YAML file %s: %w", path, err)
			}

			fmt.Printf("YAML file %s is valid.\n", path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("validating release notes: %w", err)
	}

	fmt.Println("All release notes are valid.")

	return nil
}

// ValidateYamlMap reads a YAML map file, unmarshals it into a map, and then re-marshals it
// to validate the correctness of the content.
func ValidateYamlMap(filePath string) error {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", filePath, err)
	}

	// Unmarshal the YAML data into a map for manipulation and validation
	var testMap notes.ReleaseNotesMap
	if err := yaml.Unmarshal(data, &testMap); err != nil {
		return fmt.Errorf("YAML unmarshaling %s: %w", filePath, err)
	}

	// Check the map for valid punctuation in the "text" field
	if err := validateTextFieldPunctuation(&testMap); err != nil {
		return fmt.Errorf("punctuation check for file %s: %w", filePath, err)
	}

	// TODO: Add custom validation checks (tense, grammar, spelling, etc.) https://github.com/kubernetes/release/issues/3767

	// Re-marshall the YAML to check if it can be successfully serialized again
	_, err = yaml.Marshal(testMap)
	if err != nil {
		return fmt.Errorf("while re-marshaling map for file %s:%w", filePath, err)
	}

	fmt.Printf("File %s is valid YAML.\n", filePath)

	return nil
}

// validateTextFieldPunctuation checks if the "text" field in a YAML map
// ends with valid punctuation (., !, ?).
func validateTextFieldPunctuation(data *notes.ReleaseNotesMap) error {
	validPunctuation := regexp.MustCompile(`[.!?]$`)

	if data == nil {
		return errors.New("the release notes map is nil")
	}

	if data.ReleaseNote.Text == nil {
		return errors.New("the 'text' release notes map field is nil")
	}

	text := *data.ReleaseNote.Text
	if !validPunctuation.MatchString(strings.TrimSpace(text)) {
		return fmt.Errorf("the 'text' field does not end with valid punctuation: '%s'", text)
	}

	return nil
}

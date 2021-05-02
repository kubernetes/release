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

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/cve"
	"sigs.k8s.io/release-utils/editor"
)

// releaseNotesCmd represents the subcommand for `krel release-notes`
var cveCmd = &cobra.Command{
	Use:   "cve",
	Short: "Add and edit CVE information",
	Long: `krel cve
Subcommand to work with CVE data maps used to publish vulnerability information.
This subcommand enables a Release Manager to write and import new data maps with
CVE vulnerability information.

The command enables creatin, editing and deleting existing CVE entries in the 
release bucket. See each subcommand for more information.
`,
	SilenceUsage:  false,
	SilenceErrors: false,
}

var cveDeleteCmd = &cobra.Command{
	Use:           "delete",
	Short:         "Delete an existing cve map",
	Long:          `Deletes an existing CVE map from the release bucket`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteCVE(cveOpts)
	},
	Args: argFunc,
}

var cveEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a CVE map file",
	Long: `The edit command opens an editor pulls a CVE map which has already been published and
opens it for editing in the user's editor of choice (defined by the $EDITOR 
or $KUBEEDITOR env vars). When saving and exiting the editor, krel will check
the new CVE entry and upload it to the release bucket.

To abort the editing process, do no change anything in the file or simply
delete all content from the file.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return editCVE(cveOpts)
	},
	Args: argFunc,
}

type cveOptions struct {
	CVE      string   // CVE identifier to work on
	mapFiles []string // List of mapfiles
}

var argFunc = func(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("command takes only one argument: a CVE identifier")
	}
	cveOpts.CVE = strings.ToUpper(args[0])
	if err := cve.NewClient().CheckID(cveOpts.CVE); err != nil {
		return errors.New(fmt.Sprintf("Invalid CVE ID. Format must match %s", cve.CVEIDRegExp))
	}
	return nil
}

var cveOpts = &cveOptions{}

func init() {
	cveCmd.PersistentFlags().StringSliceVarP(
		&cveOpts.mapFiles,
		"file",
		"f",
		[]string{},
		"version tag for the notes",
	)

	cveCmd.AddCommand(cveEditCmd, cveDeleteCmd)
	rootCmd.AddCommand(cveCmd)
}

// writeNewCVE opens an editor to edit a new CVE entry interactively
func writeNewCVE(opts *cveOptions) (err error) {
	client := cve.NewClient()

	file, err := client.CreateEmptyMap(opts.CVE)
	if err != nil {
		return errors.Wrap(err, "creating new cve data map")
	}

	oldFile, err := os.ReadFile(file.Name())
	if err != nil {
		return errors.Wrap(err, "reading local copy of CVE entry")
	}

	kubeEditor := editor.NewDefaultEditor([]string{"KUBE_EDITOR", "EDITOR"})
	changes, tempFilePath, err := kubeEditor.LaunchTempFile(
		"cve-datamap-", ".yaml", bytes.NewReader(oldFile),
	)
	if err != nil {
		return errors.Wrap(err, "launching editor")
	}

	if string(changes) == string(oldFile) || string(changes) == "" {
		logrus.Info("CVE information not modified")
		return nil
	}

	logrus.Infof("Creating %s entry", opts.CVE)

	// If the file was changed, re-write it:
	return client.Write(opts.CVE, tempFilePath)
}

// writeCVEFiles handles non interactive file writes
func writeCVEFiles(opts *cveOptions) error {
	client := cve.NewClient()
	for _, mapFile := range opts.mapFiles {
		if err := client.Write(opts.CVE, mapFile); err != nil {
			return errors.Wrapf(err, "writing map file %s", mapFile)
		}
	}
	return nil
}

// deleteCVE removes an existing map file
func deleteCVE(opts *cveOptions) (err error) {
	client := cve.NewClient()
	return client.Delete(opts.CVE)
}

// editCVE main edit funcion
func editCVE(opts *cveOptions) (err error) {
	client := cve.NewClient()

	// If yaml files were specified, skip the interactive mode
	if len(opts.mapFiles) != 0 {
		return writeCVEFiles(opts)
	}

	// If we're editing interactively, check if it is a new CVE
	// or we should first pull the data from the bucket
	exists, err := client.EntryExists(opts.CVE)
	if err != nil {
		return errors.Wrap(err, "checking if cve entry exists")
	}

	if exists {
		return editExistingCVE(opts)
	}

	return writeNewCVE(opts)
}

// editExistingCVE loads an existing map from the bucket and opens is
// in the user's default editor
func editExistingCVE(opts *cveOptions) (err error) {
	client := cve.NewClient()
	file, err := client.CopyToTemp(opts.CVE)
	if err != nil {
		return errors.Wrap(err, "copying CVE entry for edting")
	}
	oldFile, err := os.ReadFile(file.Name())
	if err != nil {
		return errors.Wrap(err, "reading local copy of CVE entry")
	}

	kubeEditor := editor.NewDefaultEditor([]string{"KUBE_EDITOR", "EDITOR"})
	changes, tempFilePath, err := kubeEditor.LaunchTempFile(
		"cve-datamap-", ".yaml", bytes.NewReader(oldFile),
	)
	if err != nil {
		return errors.Wrap(err, "launching editor")
	}

	if string(changes) == string(oldFile) || string(changes) == "" {
		logrus.Info("CVE information not modified")
		return nil
	}

	logrus.Infof("Updating %s entry", opts.CVE)

	// If the file was changed, re-write it:
	return client.Write(opts.CVE, tempFilePath)
}

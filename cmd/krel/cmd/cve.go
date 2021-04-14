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

var cveWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Write a new CVE vulnerability entry",
	Long: `Adds or updates a CVE map file. The write subcommand will open a new empty CVE data
map in the user's editor of choice. When saved, krel will verify the entry and uploaded
to the release bucket.

A non interactive mode for use in scripts, and processes can be invoked by
passing the write subcommand a datamap using the --file flag. In this case, krel will 
verify the map and if the CVE data is correct, it will upload the data to the bucket
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return writeCVE(cveOpts)
	},
	Args: argFunc,
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
	Short: "Edit an already published CVE map file",
	Long: `The edit command pulls a CVE map which has already been published and
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

	cveCmd.AddCommand(cveWriteCmd, cveEditCmd, cveDeleteCmd)
	rootCmd.AddCommand(cveCmd)
}

func writeCVE(opts *cveOptions) (err error) {
	client := cve.NewClient()

	if len(opts.mapFiles) == 0 {
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

	for _, mapFile := range opts.mapFiles {
		if err := client.Write(opts.CVE, mapFile); err != nil {
			return errors.Wrapf(err, "writing map file %s", mapFile)
		}
	}
	return nil
}

func deleteCVE(opts *cveOptions) (err error) {
	client := cve.NewClient()
	return client.Delete(opts.CVE)
}

// editCVE
func editCVE(opts *cveOptions) (err error) {
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

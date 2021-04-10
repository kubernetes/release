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
	"errors"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	"k8s.io/release/pkg/docs"
)

const (
	k8sOrg      string = "kubernetes"
	websiteRepo string = "website"
	ghTokVar           = "GITHUB_TOKEN"
)

type docsOptions struct {
	credFile  string
	sheetID   string
	sheetName string
	rangeCol  string
	colID     []string
	release   string
}

var docsOpts = docsOptions{
	credFile:  getCredentialFile(),
	sheetName: "Docs",
	rangeCol:  "A1:K",
	colID:     []string{"A", "G", "I", "J", "K"},
}

// announceCmd represents the subcommand for `krel announce`
var automateCmd = &cobra.Command{
	Use:   "docs",
	Short: "Automate some work for release docs",
	Long: `The docs command carries out some actions on the kubernetes website pull request at various  stages of the release.
The docs command needs access to both the Github API and Sheets API.
Export your github access token as an env var and save your Google credentials(the json file)
to ~/.krel. You can also pass in a different path with the --cred-file command`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	automateCmd.PersistentFlags().StringVarP(
		&docsOpts.credFile,
		"cred-file",
		"",
		docsOpts.credFile,
		"credentials for the Google Sheets API")

	automateCmd.PersistentFlags().StringVarP(
		&docsOpts.sheetID,
		"sheet-id",
		"",
		"",
		"Sheet ID of the spreadsheet, normally found in the url of the spreadsheet")

	automateCmd.PersistentFlags().StringVarP(
		&docsOpts.sheetName,
		"name",
		"",
		docsOpts.sheetName,
		"The name of the particular sheet that we want to read from. Found at the bottom of the sheet")

	automateCmd.PersistentFlags().StringVarP(
		&docsOpts.rangeCol,
		"range",
		"",
		docsOpts.rangeCol,
		"The range of columns from which we want to read from in the format <start-column><start-rol>:<end-column> e.g A1:J")

	automateCmd.PersistentFlags().StringArrayVar(
		&docsOpts.colID,
		"col-ids",
		docsOpts.colID,
		`the letters of the columns that contain the kep, k/w, sigs, assignee and k/k pr information respectively
e.g A,C,D,E,J where A = kep column, C = website pr column, D = sig column etc`)

	automateCmd.PersistentFlags().StringVarP(
		&docsOpts.release,
		"release",
		"",
		"",
		"The kubernetes release e.g 1.20")

	if err := automateCmd.MarkFlagRequired("sheet-id"); err != nil {
		logrus.Warnf("error setting sheet-id flag to required: %s", err)
	}
	if err := automateCmd.MarkFlagRequired("release"); err != nil {
		logrus.Warnf("error setting sheet-id flag to required: %s", err)
	}

	rootCmd.AddCommand(automateCmd)
}

func getPrStruct(sheetID, readRange string, headers []string) ([]*docs.PrData, error) {
	if len(headers) == 4 {
		return nil, errors.New("column headers should be 5")
	}

	srv, err := docs.GetSheetService(docsOpts.credFile)
	if err != nil {
		return nil, err
	}

	resp, err := srv.Spreadsheets.Values.Get(sheetID, readRange).Do()
	if err != nil {
		return nil, err
	}

	if len(resp.Values) == 0 {
		logrus.Info("No data found")
		return nil, nil
	}

	headersIdx, err := docs.ConvertColumnLettersToInt(headers)
	if err != nil {
		return nil, err
	}

	allPrData, err := docs.StructureData(resp.Values, headersIdx)
	if err != nil {
		return nil, err
	}

	return allPrData, nil
}

func getCredentialFile() string {
	if homeDir := homedir.HomeDir(); homeDir != "" {
		return filepath.Join(homeDir, ".krel", "credentials.json")
	}

	return ""
}

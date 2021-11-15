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
	"context"
	"os"

	"github.com/google/go-github/v34/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"sigs.k8s.io/release-utils/env"
)

var rootCmd = &cobra.Command{
	Use:    "reporter",
	Short:  "Github and Testgrid report generator",
	Long:   "CI-Signal reporter that generates github and testgrid reports.",
	PreRun: setGithubConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunReport(*cfg)
	},
}

// Execute executes the ci-reporter root command.
func Execute() error {
	return rootCmd.Execute()
}

// ReporterConfig configuration that is getting injected into ci-signal report functions
type ReporterConfig struct {
	GithubClient   *github.Client
	GithubToken    string
	ReleaseVersion string
	ShortReport    bool
}

var cfg = &ReporterConfig{
	GithubClient:   &github.Client{},
	GithubToken:    "",
	ReleaseVersion: "",
	ShortReport:    false,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfg.ReleaseVersion, "release-version", "v", "", "Specify a Kubernetes release versions like '1.22' which will populate the report additionally")
	rootCmd.PersistentFlags().BoolVarP(&cfg.ShortReport, "short", "s", false, "A short report for mails and slack")
}

func setGithubConfig(cmd *cobra.Command, args []string) {
	// look for token in environment variables
	cfg.GithubToken = env.Default("GITHUB_TOKEN", "")
	if cfg.GithubToken == "" {
		logrus.Fatal("Please specify your Github access token via the environment variable 'GITHUB_TOKEN' to generate a ci-report")
		os.Exit(1)
	}

	// create a github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.GithubToken})
	tc := oauth2.NewClient(ctx, ts)
	cfg.GithubClient = github.NewClient(tc)
}

// RunReport used to execute
func RunReport(cfg ReporterConfig) error {
	githubReportData, err := GetGithubReportData(cfg)
	if err != nil {
		return err
	}
	PrintGithubReportData(githubReportData)

	testgridReportData, err := GetTestgridReportData(cfg)
	if err != nil {
		return err
	}
	PrintTestgridReportData(testgridReportData)

	return nil
}

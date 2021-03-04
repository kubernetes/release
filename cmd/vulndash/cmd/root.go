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
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	adapter "k8s.io/release/pkg/vulndash/adapter"
	"sigs.k8s.io/release-utils/log"
)

var validRegistryHostnames = []string{"gcr.io", "us.gcr.io", "asia.gcr.io", "eu.gcr.io"}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "vulndash --project <project-name> --bucket <bucket> --dashboard-file-path <path>",
	Short:             "vulndash generates a dashboard of container image vulnerabilities within a GCP project",
	Example:           "vulndash --project <project-name> --bucket <bucket> --dashboard-file-path <path>",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return checkRequiredFlags(cmd.Flags())
	},
	RunE: func(*cobra.Command, []string) error {
		return run(opts)
	},
}

type options struct {
	project           string
	bucket            string
	dashboardFilePath string
	registryHostname  string
	pageSize          int32
	logLevel          string
}

var opts = &options{}

var (
	projectFlag           = "project"
	bucketFlag            = "bucket"
	dashboardFilePathFlag = "dashboard-file-path"
	registryHostnameFlag  = "registry-hostname"
	pageSizeFlag          = "page-size"

	// requiredFlags only if the config flag is not set
	requiredFlags = []string{
		projectFlag,
		bucketFlag,
		dashboardFilePathFlag,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&opts.project,
		projectFlag,
		"",
		"the project which the vulnerability dashboard will display information for",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.bucket,
		bucketFlag,
		"",
		"GCS bucket to upload dashboard files to",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.dashboardFilePath,
		dashboardFilePathFlag,
		"",
		"the path to the local dashboard files",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.logLevel,
		"log-level",
		"info",
		fmt.Sprintf("the logging verbosity, either %s", log.LevelNames()),
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.registryHostname,
		registryHostnameFlag,
		"",
		"the registry hostname for where the images are located to use to filter "+
			"when getting the reports, ie. us.gcr.io, asia.gcr.io or eu.gcr.io",
	)

	rootCmd.PersistentFlags().Int32Var(
		&opts.pageSize,
		pageSizeFlag,
		200,
		"the page size when getting the list of vulnerabilities",
	)
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(opts.logLevel)
}

func checkRequiredFlags(flags *pflag.FlagSet) error {
	checkRequiredFlags := []string{}
	flags.VisitAll(func(flag *pflag.Flag) {
		for _, requiredflag := range requiredFlags {
			if requiredflag == flag.Name && !flag.Changed {
				checkRequiredFlags = append(checkRequiredFlags, requiredflag)
			}
		}
	})

	if len(checkRequiredFlags) != 0 {
		return errors.New("Required flag(s) `" + strings.Join(checkRequiredFlags, ", ") + "` not set")
	}

	return nil
}

func validateFlags(opts *options) error {
	if opts.registryHostname != "" {
		for _, registry := range validRegistryHostnames {
			if registry == opts.registryHostname {
				return nil
			}
		}

		return errors.New("--registry-hostname needs to be one of " + strings.Join(validRegistryHostnames, ", ") +
			" and was set " + opts.registryHostname)
	}

	return nil
}

func run(opts *options) error {
	if err := validateFlags(opts); err != nil {
		return errors.Wrap(err, "validating the flags")
	}

	logrus.Info("Updating the vulnerability dashboard...")

	updateErr := adapter.UpdateVulnerabilityDashboard(
		opts.dashboardFilePath,
		opts.project,
		opts.registryHostname,
		opts.bucket,
		opts.pageSize,
	)
	if updateErr != nil {
		return errors.Wrap(updateErr, "updating vulnerability dashboard")
	}

	logrus.Info("Finished vulnerability dashboard updates")

	return nil
}

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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/release/pkg/log"
	"k8s.io/release/pkg/patch"
	"k8s.io/release/pkg/util"
)

// slap the subcommand onto the parent/root
func init() {
	cmd := patchAnnounceCommand()
	rootCmd.AddCommand(cmd)
}

func patchAnnounceCommand() *cobra.Command {
	opts := patch.AnnounceOptions{}

	cmd := &cobra.Command{
		Use:           "patch-announce",
		Short:         "Send out patch release announcement mails",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(0), // no additional/positional args allowed
	}

	cobra.OnInitialize(initConfig) // ?

	// setup local flags
	cmd.PersistentFlags().StringVarP(&opts.SenderName, "sender-name", "n", "", "email sender's name")
	cmd.PersistentFlags().StringVarP(&opts.SenderEmail, "sender-email", "e", "", "email sender's address")
	cmd.PersistentFlags().StringVarP(&opts.FreezeDate, "freeze-date", "f", "", "date when no CPs are allowed anymore")
	cmd.PersistentFlags().StringVarP(&opts.CutDate, "cut-date", "c", "", "date when the patch release is planned to be cut")
	cmd.PersistentFlags().StringVarP(&opts.ReleaseRepoPath, "release-repo", "r", "./release", "local path of the k/release checkout")

	// TODO: figure out, how we can read env vars and also be able to set the flags to required in a cobra-native way
	cmd.PersistentFlags().StringVarP(&opts.SendgridAPIKey, "sendgrid-api-key", "s", util.EnvDefault("SENDGRID_API_KEY", ""), "API key for sendgrid")
	cmd.PersistentFlags().StringVarP(&opts.GithubToken, "github-token", "g", util.EnvDefault("GITHUB_TOKEN", ""), "a GitHub token, used r/o for generating the release notes")

	cmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		// TODO: make github-token & sendgrid-api-key required too
		if err := setFlagsRequired(cmd, "sender-name", "sender-email", "freeze-date", "cut-date"); err != nil {
			return err
		}

		var err error
		if opts.Nomock, err = cmd.Flags().GetBool("nomock"); err != nil {
			return err
		}
		if opts.K8sRepoPath, err = cmd.Flags().GetString("repo"); err != nil {
			return err
		}
		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Get the global logger, add the command's name as an initial tracing
		// field and use that from here on
		localLogger := logrus.NewEntry(logrus.StandardLogger())
		logger := log.AddTracePath(localLogger, cmd.Name()).WithField("mock", !opts.Nomock)

		announcer := &patch.Announcer{
			Opts: opts,
		}
		announcer.SetLogger(logger, "announcer")

		logger.Debug("run announcer")
		return announcer.Run()
	}

	return cmd
}

func setFlagsRequired(cmd *cobra.Command, flags ...string) error {
	for _, f := range flags {
		if err := cmd.MarkPersistentFlagRequired(f); err != nil {
			return err
		}
	}
	return nil
}

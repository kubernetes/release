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
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/release-utils/env"
	"sigs.k8s.io/release-utils/helpers"
	"sigs.k8s.io/release-utils/http"

	"k8s.io/release/pkg/mail"
	"k8s.io/release/pkg/release"
)

const (
	sendgridAPIKeyEnvKey = "SENDGRID_API_KEY" //nolint:gosec // it's just the key
	nameFlag             = "name"
	emailFlag            = "email"
)

// announceCmd represents the subcommand for `krel announce`.
var sendAnnounceCmd = &cobra.Command{
	Use:   "send",
	Short: "Announce Kubernetes releases",
	Long: fmt.Sprintf(`krel announce send

krel announce send can be used to mail an announcement of an already
built Kubernetes release to the %q and %q Google Groups.

By default the mail will be sent only to a test Google Group %q,
ie: the announcement run will only be a mock run.  To do an
official announcement, use the --nomock flag.

It is necessary to export the $%s environment variable. An API key can be created by
registering a sendgrid.com account and adding the key here:

https://app.sendgrid.com/settings/api_keys

Beside this, if the flags for a valid sender name (--%s,-n) and sender email
address (--%s,-e) are not set, then it tries to retrieve those values directly
from the Sendgrid API.

Setting a valid Kubernetes tag (--%s,-t) is always necessary.

If --%s,-p is given, then krel announce will only print the email content
without doing anything else.`,
		mail.KubernetesAnnounceGoogleGroup,
		mail.KubernetesDevGoogleGroup,
		mail.KubernetesAnnounceTestGoogleGroup,
		sendgridAPIKeyEnvKey,
		nameFlag,
		emailFlag,
		tagFlag,
		printOnlyFlag,
	),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnnounce(sendAnnounceOpts, announceOpts, rootOpts)
	},
}

type sendAnnounceOptions struct {
	sendgridAPIKey string
	name           string
	email          string
}

var sendAnnounceOpts = &sendAnnounceOptions{}

func init() {
	sendAnnounceOpts.sendgridAPIKey = env.Default(sendgridAPIKeyEnvKey, "")

	sendAnnounceCmd.PersistentFlags().StringVarP(
		&sendAnnounceOpts.name,
		nameFlag,
		"n",
		"",
		"mail sender name",
	)

	sendAnnounceCmd.PersistentFlags().StringVarP(
		&sendAnnounceOpts.email,
		emailFlag,
		"e",
		"",
		"email address",
	)

	announceCmd.AddCommand(sendAnnounceCmd)
}

func runAnnounce(opts *sendAnnounceOptions, announceRootOpts *announceOptions, rootOpts *rootOptions) error {
	if err := announceRootOpts.Validate(); err != nil {
		return fmt.Errorf("validating announcement send options: %w", err)
	}

	logrus.Info("Retrieving release announcement from Google Cloud Bucket")

	tag := helpers.AddTagPrefix(announceRootOpts.tag)
	u := fmt.Sprintf(
		"%s/release/%s/announcement.html",
		release.URLPrefixForBucket(release.ProductionBucket), tag,
	)
	logrus.Infof("Using announcement remote URL: %s", u)

	content, err := http.NewAgent().Get(u)
	if err != nil {
		return fmt.Errorf(
			"unable to retrieve release announcement form url: %s: %w", u, err,
		)
	}

	if announceRootOpts.printOnly {
		logrus.Infof("The email content is:")
		fmt.Print(string(content))

		return nil
	}

	if opts.sendgridAPIKey == "" {
		return fmt.Errorf(
			"$%s is not set", sendgridAPIKeyEnvKey,
		)
	}

	logrus.Info("Preparing mail sender")

	m := mail.NewSender(opts.sendgridAPIKey)

	if opts.name != "" && opts.email != "" {
		if err := m.SetSender(opts.name, opts.email); err != nil {
			return fmt.Errorf("unable to set mail sender: %w", err)
		}
	} else {
		logrus.Info("Retrieving default sender from sendgrid API")

		if err := m.SetDefaultSender(); err != nil {
			return fmt.Errorf("setting default sender: %w", err)
		}
	}

	groups := []mail.GoogleGroup{mail.KubernetesAnnounceTestGoogleGroup}
	if rootOpts.nomock {
		groups = []mail.GoogleGroup{
			mail.KubernetesAnnounceGoogleGroup,
			mail.KubernetesDevGoogleGroup,
		}
	}

	logrus.Infof("Using Google Groups as announcement target: %v", groups)

	if err := m.SetGoogleGroupRecipients(groups...); err != nil {
		return fmt.Errorf("unable to set mail recipients: %w", err)
	}

	logrus.Info("Sending mail")

	subject := fmt.Sprintf("Kubernetes %s is live!", tag)

	yes := true

	if rootOpts.nomock {
		_, yes, err = helpers.Ask("Send email? (y/N)", "y:Y:yes|n:N:no|N", 10)
		if err != nil {
			return err
		}
	}

	if yes {
		if err := m.Send(string(content), subject); err != nil {
			return fmt.Errorf("unable to send mail: %w", err)
		}
	}

	return nil
}

func (o *announceOptions) Validate() error {
	if o.tag == "" {
		return errors.New("need to specify a tag value")
	}

	return nil
}

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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/pkg/http"
	"k8s.io/release/pkg/mail"
	"k8s.io/release/pkg/release"
	"k8s.io/release/pkg/util"
)

const (
	sendgridAPIKeyFlag   = "sendgrid-api-key"
	sendgridAPIKeyEnvKey = "SENDGRID_API_KEY"
	nameFlag             = "name"
	emailFlag            = "email"
	tagFlag              = "tag"
)

// announceCmd represents the subcommand for `krel announce`
var announceCmd = &cobra.Command{
	Use:   "announce",
	Short: "Announce Kubernetes releases",
	Long: fmt.Sprintf(`krel announce

krel announce can be used to announce already built Kubernetes releases to the
%q and %q Google Group.

If --nomock=true (the default), then the mail will be sent only to a test
Google Group %q.

It is necessary to either set a valid --%s,-s or export the
$%s environment variable. An API key can be created by
registering a sendgrid.com account and adding the key here:

https://app.sendgrid.com/settings/api_keys

Beside this, the flags for a valid sender name (--%s,-n), sender email
address (--%s,-e) and a valid Kubernetes tag (--%s,-t) have to be set
as well.`,
		mail.KubernetesAnnounceGoogleGroup,
		mail.KubernetesDevGoogleGroup,
		mail.KubernetesAnnounceTestGoogleGroup,
		sendgridAPIKeyFlag,
		sendgridAPIKeyEnvKey,
		nameFlag,
		emailFlag,
		tagFlag,
	),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnnounce(announceOpts, rootOpts)
	},
}

type announceOptions struct {
	sendgridAPIKey string
	name           string
	email          string
	tag            string
}

var announceOpts = &announceOptions{}

func init() {
	announceCmd.PersistentFlags().StringVarP(
		&announceOpts.sendgridAPIKey,
		sendgridAPIKeyFlag,
		"s",
		util.EnvDefault(sendgridAPIKeyEnvKey, ""),
		fmt.Sprintf(
			"API key for sendgrid, can be set via %s too",
			sendgridAPIKeyEnvKey,
		),
	)

	announceCmd.PersistentFlags().StringVarP(
		&announceOpts.name,
		nameFlag,
		"n",
		"",
		"mail sender name",
	)
	if err := announceCmd.MarkPersistentFlagRequired(nameFlag); err != nil {
		logrus.Fatal(err)
	}

	announceCmd.PersistentFlags().StringVarP(
		&announceOpts.email,
		emailFlag,
		"e",
		"",
		"email address",
	)
	if err := announceCmd.MarkPersistentFlagRequired(emailFlag); err != nil {
		logrus.Fatal(err)
	}

	announceCmd.PersistentFlags().StringVarP(
		&announceOpts.tag,
		tagFlag,
		"t",
		"",
		"built tag to be announced, will be used for fetching the "+
			"announcement from the google cloud bucket",
	)
	if err := announceCmd.MarkPersistentFlagRequired(tagFlag); err != nil {
		logrus.Fatal(err)
	}

	rootCmd.AddCommand(announceCmd)
}

func runAnnounce(opts *announceOptions, rootOpts *rootOptions) error {
	if opts.sendgridAPIKey == "" {
		return errors.Errorf(
			"Neither --sendgrid-api-key,-s nor $%s is set", sendgridAPIKeyEnvKey,
		)
	}

	logrus.Info("Retrieving release announcement from Google Cloud Bucket")

	tag := util.AddTagPrefix(opts.tag)
	u := fmt.Sprintf(
		"%s/archive/anago-%s/announcement.html",
		release.URLPrefixForBucket(release.ProductionBucket), tag,
	)
	logrus.Infof("Using announcement remote URL: %s", u)

	content, err := http.GetURLResponse(u, false)
	if err != nil {
		return errors.Wrapf(err,
			"unable to retrieve release announcement form url: %s", u,
		)
	}

	logrus.Info("Preparing mail sender")
	m := mail.Sender{APIKey: opts.sendgridAPIKey}

	if err := m.SetSender(opts.name, opts.email); err != nil {
		return errors.Wrap(err, "unable to set mail sender")
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
		return errors.Wrap(err, "unable to set mail recipients")
	}

	logrus.Info("Sending mail")
	subject := fmt.Sprintf("Kubernetes %s is live!", tag)
	if err := m.Send(content, subject); err != nil {
		return errors.Wrap(err, "unable to send mail")
	}

	return nil
}

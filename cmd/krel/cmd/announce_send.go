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
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/release-utils/helpers"
	"sigs.k8s.io/release-utils/http"

	"k8s.io/release/pkg/mail"
	"k8s.io/release/pkg/release"
)

const noBrowserFlag = "no-browser"

// sendAnnounceCmd represents the subcommand for `krel announce send`.
var sendAnnounceCmd = &cobra.Command{
	Use:   "send",
	Short: "Announce Kubernetes releases",
	Long: fmt.Sprintf(`krel announce send

krel announce send can be used to mail an announcement of an already
built Kubernetes release to the %q and %q Google Groups.

By default the mail will be sent only to a test Google Group %q,
ie: the announcement run will only be a mock run. To do an
official announcement, use the --nomock flag.

Email is sent via the Gmail API using Google OAuth. A browser window
will open for authorization on each run. No additional setup is
needed.

Use --%s to print the authorization URL for manual copy/paste
(useful in headless environments). After authorizing in the browser,
the redirect will fail to load. Copy the full URL from the browser's
address bar and paste it back into the terminal.

Setting a valid Kubernetes tag (--%s,-t) is always necessary.

If --%s,-p is given, then krel announce will only print the email
content without doing anything else.`,
		mail.KubernetesAnnounceGoogleGroup,
		mail.KubernetesDevGoogleGroup,
		mail.KubernetesAnnounceTestGoogleGroup,
		noBrowserFlag,
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
	noBrowser bool
}

var sendAnnounceOpts = &sendAnnounceOptions{}

func init() {
	sendAnnounceCmd.PersistentFlags().BoolVar(
		&sendAnnounceOpts.noBrowser,
		noBrowserFlag,
		false,
		"disable automatic browser opening for OAuth (manual URL copy/paste)",
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
			"unable to retrieve release announcement from url: %s: %w", u, err,
		)
	}

	if announceRootOpts.printOnly {
		logrus.Infof("The email content is:")
		fmt.Print(string(content))

		return nil
	}

	logrus.Info("Starting Gmail OAuth flow")

	sender, err := mail.NewGmailSender(context.Background(), opts.noBrowser)
	if err != nil {
		return fmt.Errorf("creating Gmail sender: %w", err)
	}

	groups := []mail.GoogleGroup{mail.KubernetesAnnounceTestGoogleGroup}
	if rootOpts.nomock {
		groups = []mail.GoogleGroup{
			mail.KubernetesAnnounceGoogleGroup,
			mail.KubernetesDevGoogleGroup,
		}
	}

	logrus.Infof("Using Google Groups as announcement target: %v", groups)

	sender.SetGoogleGroupRecipients(groups...)

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
		if err := sender.Send(string(content), subject); err != nil {
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

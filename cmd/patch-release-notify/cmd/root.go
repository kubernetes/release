/*
Copyright 2022 The Kubernetes Authors.

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
	"embed"
	"fmt"
	"html/template"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/release/cmd/schedule-builder/model"
	"k8s.io/release/pkg/mail"
	"sigs.k8s.io/release-utils/env"
	"sigs.k8s.io/release-utils/log"
	"sigs.k8s.io/yaml"
)

//go:embed templates/*.tmpl
var tpls embed.FS

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "patch-release-notify --schedule-path /path/to/schedule.yaml",
	Short:             "patch-release-notify check the cherry pick deadline and send an email to notify",
	Example:           "patch-release-notify --schedule-path /path/to/schedule.yaml",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initLogging,
	RunE: func(*cobra.Command, []string) error {
		return run(opts)
	},
}

type options struct {
	sendgridAPIKey string
	schedulePath   string
	dayToalert     int
	name           string
	email          string
	nomock         bool
	logLevel       string
}

var opts = &options{}

const (
	sendgridAPIKeyEnvKey = "SENDGRID_API_KEY" // nolint: gosec
	layout               = "2006-01-02"

	schedulePathFlag = "schedule-path"
	nameFlag         = "name"
	emailFlag        = "email"
	dayToalertFlag   = "days-to-alert"
)

var requiredFlags = []string{
	schedulePathFlag,
}

type Template struct {
	Releases []TemplateRelease
}

type TemplateRelease struct {
	Release            string
	CherryPickDeadline string
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	opts.sendgridAPIKey = env.Default(sendgridAPIKeyEnvKey, "")

	rootCmd.PersistentFlags().StringVar(
		&opts.schedulePath,
		schedulePathFlag,
		"",
		"path where can find the schedule.yaml file",
	)

	rootCmd.PersistentFlags().BoolVar(
		&opts.nomock,
		"nomock",
		false,
		"run the command to target the production environment",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.logLevel,
		"log-level",
		"info",
		fmt.Sprintf("the logging verbosity, either %s", log.LevelNames()),
	)

	rootCmd.PersistentFlags().StringVarP(
		&opts.name,
		nameFlag,
		"n",
		"",
		"mail sender name",
	)

	rootCmd.PersistentFlags().IntVar(
		&opts.dayToalert,
		dayToalertFlag,
		3,
		"day to before the deadline to send the notification. Defaults to 3 days.",
	)

	rootCmd.PersistentFlags().StringVarP(
		&opts.email,
		emailFlag,
		"e",
		"",
		"email address",
	)

	for _, flag := range requiredFlags {
		if err := rootCmd.MarkPersistentFlagRequired(flag); err != nil {
			logrus.Fatal(err)
		}
	}
}

func initLogging(*cobra.Command, []string) error {
	return log.SetupGlobalLogger(opts.logLevel)
}

func run(opts *options) error {
	if err := opts.SetAndValidate(); err != nil {
		return errors.Wrap(err, "validating schedule-path options")
	}

	if opts.sendgridAPIKey == "" {
		return errors.Errorf(
			"$%s is not set", sendgridAPIKeyEnvKey,
		)
	}

	data, err := loadFileOrURL(opts.schedulePath)
	if err != nil {
		return errors.Wrap(err, "failed to read the file")
	}

	patchSchedule := &model.PatchSchedule{}

	logrus.Info("Parsing the schedule...")

	if err := yaml.UnmarshalStrict(data, &patchSchedule); err != nil {
		return errors.Wrap(err, "failed to decode the file")
	}

	output := &Template{}

	shouldSendEmail := false

	for _, patch := range patchSchedule.Schedules {
		t, err := time.Parse(layout, patch.CherryPickDeadline)
		if err != nil {
			return errors.Wrap(err, "parsing schedule time")
		}

		currentTime := time.Now().UTC()
		days := t.Sub(currentTime).Hours() / 24
		intDay, _ := math.Modf(days)
		if int(intDay) == opts.dayToalert {
			output.Releases = append(output.Releases, TemplateRelease{
				Release:            patch.Release,
				CherryPickDeadline: patch.CherryPickDeadline,
			})
			shouldSendEmail = true
		}
	}

	tmpl, err := template.ParseFS(tpls, "templates/email.tmpl")
	if err != nil {
		return errors.Wrap(err, "parsing template")
	}

	var tmplBytes bytes.Buffer
	err = tmpl.Execute(&tmplBytes, output)
	if err != nil {
		return errors.Wrap(err, "parsing values to the template")
	}

	if shouldSendEmail {
		if !opts.nomock {
			logrus.Info("This is a mock only, will print out the email before sending to a test mailing list")
			fmt.Println(tmplBytes.String())
		}

		logrus.Info("Preparing mail sender")
		m := mail.NewSender(opts.sendgridAPIKey)

		if opts.name != "" && opts.email != "" {
			if err := m.SetSender(opts.name, opts.email); err != nil {
				return errors.Wrap(err, "unable to set mail sender")
			}
		} else {
			logrus.Info("Retrieving default sender from sendgrid API")
			if err := m.SetDefaultSender(); err != nil {
				return errors.Wrap(err, "setting default sender")
			}
		}

		groups := []mail.GoogleGroup{mail.KubernetesAnnounceTestGoogleGroup}
		if opts.nomock {
			groups = []mail.GoogleGroup{
				mail.KubernetesDevGoogleGroup,
			}
		}
		logrus.Infof("Using Google Groups as announcement target: %v", groups)

		if err := m.SetGoogleGroupRecipients(groups...); err != nil {
			return errors.Wrap(err, "unable to set mail recipients")
		}

		logrus.Info("Sending mail")
		subject := "[Please Read] Patch Releases cherry-pick deadline"

		if err := m.Send(tmplBytes.String(), subject); err != nil {
			return errors.Wrap(err, "unable to send mail")
		}
	} else {
		logrus.Info("No email is needed to send")
	}

	return nil
}

// SetAndValidate sets some default options and verifies if options are valid
func (o *options) SetAndValidate() error {
	logrus.Info("Validating schedule-path options...")

	if o.schedulePath == "" {
		return errors.Errorf("need to set the schedule-path")
	}

	return nil
}

func loadFileOrURL(fileRef string) ([]byte, error) {
	var raw []byte
	var err error
	if strings.HasPrefix(fileRef, "http://") || strings.HasPrefix(fileRef, "https://") {
		// #nosec G107
		resp, err := http.Get(fileRef)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		raw, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		raw, err = os.ReadFile(filepath.Clean(fileRef))
		if err != nil {
			return nil, err
		}
	}
	return raw, nil
}

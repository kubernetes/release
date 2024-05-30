/*
Copyright 2024 The Kubernetes Authors.

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

package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gomodules.xyz/envconfig"
	"gopkg.in/gomail.v2"
	"gopkg.in/yaml.v3"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/sirupsen/logrus"

	"k8s.io/release/cmd/schedule-builder/model"
)

//go:embed templates/email.tmpl
var tpls embed.FS

type Config struct {
	FromEmail    string `envconfig:"FROM_EMAIL"    required:"true"`
	ToEmail      string `envconfig:"TO_EMAIL"      required:"true"`
	SchedulePath string `envconfig:"SCHEDULE_PATH" required:"true"`
	DaysToAlert  int    `envconfig:"DAYS_TO_ALERT" required:"true"`

	NoMock bool `default:"false" envconfig:"NO_MOCK" required:"true"`

	AWSRegion string `envconfig:"AWS_REGION" required:"true"`
}

type Options struct {
	AWSSess *session.Session
	Config  *Config
	Context context.Context
}

const (
	layout = "2006-01-02"
)

type Template struct {
	Releases []TemplateRelease
}

type TemplateRelease struct {
	Release            string
	CherryPickDeadline string
}

func main() {
	lambda.Start(handler)
}

func getConfig() (*Config, error) {
	var c Config
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func NewOptions(ctx context.Context) (*Options, error) {
	config, err := getConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// create new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWSRegion),
	})
	if err != nil {
		return nil, err
	}

	return &Options{
		AWSSess: sess,
		Config:  config,
		Context: ctx,
	}, nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) { //nolint: gocritic
	o, err := NewOptions(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       `{"status": "nok"}`,
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	logrus.Infof("Will pull the patch release schedule from: %s", o.Config.SchedulePath)
	data, err := loadFileOrURL(o.Config.SchedulePath)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       `{"status": "nok"}`,
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("reading the file: %w", err)
	}

	patchSchedule := &model.PatchSchedule{}

	logrus.Info("Parsing the schedule...")

	if err := yaml.Unmarshal(data, &patchSchedule); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       `{"status": "nok"}`,
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("decoding the file: %w", err)
	}

	output := &Template{}

	shouldSendEmail := false

	for _, patch := range patchSchedule.Schedules {
		t, err := time.Parse(layout, patch.Next.CherryPickDeadline)
		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       `{"status": "nok"}`,
				StatusCode: http.StatusInternalServerError,
			}, fmt.Errorf("parsing schedule time: %w", err)
		}

		currentTime := time.Now().UTC()
		days := t.Sub(currentTime).Hours() / 24
		intDay, _ := math.Modf(days)
		logrus.Infof("Cherry pick deadline: %d, days to alert: %d", int(intDay), o.Config.DaysToAlert)
		if int(intDay) == o.Config.DaysToAlert {
			output.Releases = append(output.Releases, TemplateRelease{
				Release:            patch.Release,
				CherryPickDeadline: patch.Next.CherryPickDeadline,
			})
			shouldSendEmail = true
		}
	}

	if !shouldSendEmail {
		logrus.Info("No email is needed to send")
		return events.APIGatewayProxyResponse{
			Body:       `{"status": "ok"}`,
			StatusCode: http.StatusOK,
		}, nil
	}

	tmpl, err := template.ParseFS(tpls, "templates/email.tmpl")
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       `{"status": "nok"}`,
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("parsing template: %w", err)
	}

	var tmplBytes bytes.Buffer
	err = tmpl.Execute(&tmplBytes, output)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       `{"status": "nok"}`,
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("executing the template: %w", err)
	}

	logrus.Info("Sending mail")
	subject := "[Please Read] Upcoming Patch Releases Cherry-Pick Deadline for Kubernetes"
	fromEmail := o.Config.FromEmail

	recipient := Recipient{
		toEmails: []string{o.Config.ToEmail},
	}

	if !o.Config.NoMock {
		logrus.Info("This is a mock only, will print out the email before sending to a test mailing list")
		fmt.Println(tmplBytes.String())
		// if is a mock we send the email to ourselves to test
		recipient.toEmails = []string{o.Config.FromEmail}
	}

	err = o.SendEmailRawSES(tmplBytes.String(), subject, fromEmail, recipient)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       `{"status": "nok"}`,
			StatusCode: http.StatusInternalServerError,
		}, fmt.Errorf("sending the email: %w", err)
	}

	return events.APIGatewayProxyResponse{
		Body:       `{"status": "ok"}`,
		StatusCode: 200,
	}, nil
}

// Recipient struct to hold email IDs
type Recipient struct {
	toEmails  []string
	ccEmails  []string
	bccEmails []string
}

// SendEmailSES sends email to specified email IDs
func (o *Options) SendEmailRawSES(messageBody, subject, fromEmail string, recipient Recipient) error {
	// create raw message
	msg := gomail.NewMessage()

	// set to section
	recipients := make([]*string, 0, len(recipient.toEmails))
	for _, r := range recipient.toEmails {
		recipient := r
		recipients = append(recipients, &recipient)
	}

	// cc mails mentioned
	if len(recipient.ccEmails) != 0 {
		// Need to add cc mail IDs also in recipient list
		for _, r := range recipient.ccEmails {
			recipient := r
			recipients = append(recipients, &recipient)
		}
		msg.SetHeader("cc", recipient.ccEmails...)
	}

	// bcc mails mentioned
	if len(recipient.bccEmails) != 0 {
		// Need to add bcc mail IDs also in recipient list
		for _, r := range recipient.bccEmails {
			recipient := r
			recipients = append(recipients, &recipient)
		}
		msg.SetHeader("bcc", recipient.bccEmails...)
	}

	// create an SES session.
	svc := ses.New(o.AWSSess)

	msg.SetAddressHeader("From", fromEmail, "Release Managers")
	msg.SetHeader("To", recipient.toEmails...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", messageBody)

	// create a new buffer to add raw data
	var emailRaw bytes.Buffer
	_, err := msg.WriteTo(&emailRaw)
	if err != nil {
		logrus.Errorf("Failed to write mail: %v", err)
		return err
	}

	// create new raw message
	message := ses.RawMessage{Data: emailRaw.Bytes()}

	input := &ses.SendRawEmailInput{Source: &fromEmail, Destinations: recipients, RawMessage: &message}

	// send raw email
	_, err = svc.SendRawEmail(input)
	if err != nil {
		logrus.Errorf("Error sending mail - %v", err)
		return err
	}

	logrus.Infof("Email sent successfully to: %q", recipient.toEmails)
	return nil
}

func loadFileOrURL(fileRef string) ([]byte, error) {
	var raw []byte
	var err error
	if strings.HasPrefix(fileRef, "http://") || strings.HasPrefix(fileRef, "https://") {
		resp, err := http.Get(fileRef) //nolint:gosec // we are not using user input we set via env var
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

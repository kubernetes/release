/*
Copyright 2026 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/sirupsen/logrus"

	"sigs.k8s.io/yaml"

	"k8s.io/release/cmd/schedule-builder/model"
)

const (
	dateLayout   = "2006-01-02"
	emailSubject = "[Please Read] Upcoming Patch Releases Cherry-Pick Deadline for Kubernetes"
	fromName     = "Release Managers"

	// maxScheduleSize caps how much of the schedule we read, the real file is
	// a few KiB.
	maxScheduleSize = 1 << 20
	httpTimeout     = 30 * time.Second
)

//go:embed templates/email.tmpl
var tpls embed.FS

var emailTemplate = template.Must(template.ParseFS(tpls, "templates/email.tmpl"))

var httpClient = &http.Client{Timeout: httpTimeout}

// config is read from the Lambda environment.
type config struct {
	// FROM_EMAIL is the SES verified identity used as sender.
	fromEmail string
	// TO_EMAIL is the recipient used when NO_MOCK is true.
	toEmail string
	// SCHEDULE_PATH is an https URL or a local path to schedule.yaml.
	schedulePath string
	// DAYS_TO_ALERT is how many calendar days before the cherry-pick
	// deadline the notification is sent.
	daysToAlert int
	// NO_MOCK sends the email to TO_EMAIL. When false (the default) the
	// email is only sent to FROM_EMAIL so it can be checked first.
	noMock bool
}

type templateData struct {
	Releases []templateRelease
}

type templateRelease struct {
	Release            string
	CherryPickDeadline string
}

// emailSender is the subset of the SES v2 client we use.
type emailSender interface {
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

func main() {
	lambda.Start(handler)
}

// handler is invoked by an EventBridge schedule, the event payload is unused.
func handler(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// The region is provided by the Lambda runtime via AWS_REGION.
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("loading AWS config: %w", err)
	}

	return run(ctx, cfg, time.Now(), sesv2.NewFromConfig(awsCfg))
}

func run(ctx context.Context, cfg *config, now time.Time, sender emailSender) error {
	logrus.Infof("Pulling the patch release schedule from: %s", cfg.schedulePath)

	data, err := loadSchedule(ctx, cfg.schedulePath)
	if err != nil {
		return fmt.Errorf("reading the schedule: %w", err)
	}

	schedule := &model.PatchSchedule{}
	if err := yaml.Unmarshal(data, schedule); err != nil {
		return fmt.Errorf("decoding the schedule: %w", err)
	}

	releases, err := releasesToNotify(schedule, now, cfg.daysToAlert)
	if err != nil {
		return err
	}

	if len(releases) == 0 {
		logrus.Info("No cherry-pick deadline matches, no email to send")

		return nil
	}

	var body bytes.Buffer
	if err := emailTemplate.Execute(&body, templateData{Releases: releases}); err != nil {
		return fmt.Errorf("executing the template: %w", err)
	}

	to := cfg.toEmail
	if !cfg.noMock {
		logrus.Infof("Mock mode, sending the email to %s instead of %s", cfg.fromEmail, cfg.toEmail)
		logrus.Info(body.String())

		to = cfg.fromEmail
	}

	if err := sendEmail(ctx, sender, cfg.fromEmail, to, body.String()); err != nil {
		return fmt.Errorf("sending the email: %w", err)
	}

	logrus.Infof("Email sent successfully to: %s", to)

	return nil
}

func loadConfig() (*config, error) {
	var errs []error

	required := func(key string) string {
		v, ok := os.LookupEnv(key)
		if !ok || strings.TrimSpace(v) == "" {
			errs = append(errs, fmt.Errorf("%s is required", key))
		}

		return strings.TrimSpace(v)
	}

	cfg := &config{
		fromEmail:    required("FROM_EMAIL"),
		toEmail:      required("TO_EMAIL"),
		schedulePath: required("SCHEDULE_PATH"),
	}

	for key, v := range map[string]string{"FROM_EMAIL": cfg.fromEmail, "TO_EMAIL": cfg.toEmail} {
		if v == "" {
			continue
		}

		if _, err := mail.ParseAddress(v); err != nil {
			errs = append(errs, fmt.Errorf("parsing %s: %w", key, err))
		}
	}

	if v := required("DAYS_TO_ALERT"); v != "" {
		days, err := strconv.Atoi(v)
		if err != nil || days < 0 {
			errs = append(errs, fmt.Errorf("DAYS_TO_ALERT must be a non-negative integer, got %q", v))
		}

		cfg.daysToAlert = days
	}

	if v, ok := os.LookupEnv("NO_MOCK"); ok && v != "" {
		noMock, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("parsing NO_MOCK: %w", err))
		}

		cfg.noMock = noMock
	}

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	return cfg, nil
}

// releasesToNotify returns the branches whose next cherry-pick deadline is
// exactly daysToAlert calendar days (UTC) after now.
func releasesToNotify(schedule *model.PatchSchedule, now time.Time, daysToAlert int) ([]templateRelease, error) {
	today := now.UTC().Truncate(24 * time.Hour)

	var releases []templateRelease

	for _, sched := range schedule.Schedules {
		if sched == nil || sched.Next == nil || sched.Next.CherryPickDeadline == "" {
			continue
		}

		deadline, err := time.Parse(dateLayout, sched.Next.CherryPickDeadline)
		if err != nil {
			return nil, fmt.Errorf("parsing cherry-pick deadline for %s: %w", sched.Release, err)
		}

		days := int(deadline.Sub(today).Hours() / 24)
		logrus.Infof("Release %s: cherry-pick deadline %s is in %d days, days to alert: %d", sched.Release, sched.Next.CherryPickDeadline, days, daysToAlert)

		if days == daysToAlert {
			releases = append(releases, templateRelease{
				Release:            sched.Release,
				CherryPickDeadline: sched.Next.CherryPickDeadline,
			})
		}
	}

	return releases, nil
}

func loadSchedule(ctx context.Context, ref string) ([]byte, error) {
	if strings.HasPrefix(ref, "http://") {
		return nil, errors.New("refusing to fetch the schedule over plain http, use https")
	}

	if !strings.HasPrefix(ref, "https://") {
		return os.ReadFile(filepath.Clean(ref))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating the request: %w", err)
	}

	resp, err := httpClient.Do(req) //nolint:gosec // the URL comes from the Lambda configuration
	if err != nil {
		return nil, fmt.Errorf("fetching the schedule: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching the schedule: unexpected status %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxScheduleSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading the response body: %w", err)
	}

	if len(data) > maxScheduleSize {
		return nil, fmt.Errorf("schedule is larger than %d bytes", maxScheduleSize)
	}

	return data, nil
}

func sendEmail(ctx context.Context, sender emailSender, from, to, htmlBody string) error {
	fromAddr := (&mail.Address{Name: fromName, Address: from}).String()

	_, err := sender.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(fromAddr),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(emailSubject), Charset: aws.String("UTF-8")},
				Body: &types.Body{
					Html: &types.Content{Data: aws.String(htmlBody), Charset: aws.String("UTF-8")},
				},
			},
		},
	})

	return err
}

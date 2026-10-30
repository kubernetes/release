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
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/release/cmd/schedule-builder/model"
)

const testSchedule = `
schedules:
- release: "1.31"
  next:
    release: 1.31.2
    cherryPickDeadline: "2024-10-11"
    targetDate: "2024-10-16"
- release: "1.30"
  next:
    release: 1.30.6
    cherryPickDeadline: "2024-10-11"
    targetDate: "2024-10-16"
- release: "1.29"
  next:
    release: 1.29.10
    cherryPickDeadline: "2024-10-18"
    targetDate: "2024-10-23"
- release: "1.28"
`

type fakeSender struct {
	inputs []*sesv2.SendEmailInput
	err    error
}

func (f *fakeSender) SendEmail(_ context.Context, in *sesv2.SendEmailInput, _ ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	f.inputs = append(f.inputs, in)

	return &sesv2.SendEmailOutput{}, f.err
}

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()

	for _, key := range []string{"FROM_EMAIL", "TO_EMAIL", "SCHEDULE_PATH", "DAYS_TO_ALERT", "NO_MOCK"} {
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}

	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoadConfig(t *testing.T) {
	valid := map[string]string{
		"FROM_EMAIL":    "release-managers@example.com",
		"TO_EMAIL":      "dev@example.com",
		"SCHEDULE_PATH": "https://example.com/schedule.yaml",
		"DAYS_TO_ALERT": "3",
	}

	with := func(k, v string) map[string]string {
		m := maps.Clone(valid)
		m[k] = v

		return m
	}

	for _, tc := range []struct {
		name    string
		env     map[string]string
		want    *config
		wantErr string
	}{
		{
			name: "defaults to mock",
			env:  valid,
			want: &config{
				fromEmail:    "release-managers@example.com",
				toEmail:      "dev@example.com",
				schedulePath: "https://example.com/schedule.yaml",
				daysToAlert:  3,
			},
		},
		{
			name: "no mock",
			env:  with("NO_MOCK", "true"),
			want: &config{
				fromEmail:    "release-managers@example.com",
				toEmail:      "dev@example.com",
				schedulePath: "https://example.com/schedule.yaml",
				daysToAlert:  3,
				noMock:       true,
			},
		},
		{name: "missing everything", env: map[string]string{}, wantErr: "FROM_EMAIL is required"},
		{name: "invalid days", env: with("DAYS_TO_ALERT", "three"), wantErr: "DAYS_TO_ALERT"},
		{name: "negative days", env: with("DAYS_TO_ALERT", "-1"), wantErr: "DAYS_TO_ALERT"},
		{name: "invalid no mock", env: with("NO_MOCK", "yes please"), wantErr: "NO_MOCK"},
		{name: "invalid email", env: with("TO_EMAIL", "not an email"), wantErr: "TO_EMAIL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, tc.env)

			got, err := loadConfig()
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReleasesToNotify(t *testing.T) {
	schedule := &model.PatchSchedule{
		Schedules: []*model.Schedule{
			{Release: "1.31", Next: &model.PatchRelease{CherryPickDeadline: "2024-10-11"}},
			{Release: "1.30", Next: &model.PatchRelease{CherryPickDeadline: "2024-10-11"}},
			{Release: "1.29", Next: &model.PatchRelease{CherryPickDeadline: "2024-10-18"}},
			{Release: "1.28"},
			nil,
		},
	}

	for _, tc := range []struct {
		name string
		now  time.Time
		days int
		want []string
	}{
		{
			// The Lambda runs at 16:00 UTC, the time of day must not shift
			// the result to another day.
			name: "matches calendar days regardless of time of day",
			now:  time.Date(2024, 10, 8, 16, 0, 0, 0, time.UTC),
			days: 3,
			want: []string{"1.31", "1.30"},
		},
		{
			name: "matches just before midnight",
			now:  time.Date(2024, 10, 8, 23, 59, 59, 0, time.UTC),
			days: 3,
			want: []string{"1.31", "1.30"},
		},
		{
			name: "non UTC time zone",
			now:  time.Date(2024, 10, 8, 20, 0, 0, 0, time.FixedZone("PDT", -7*3600)),
			days: 2,
			want: []string{"1.31", "1.30"},
		},
		{
			name: "deadline day",
			now:  time.Date(2024, 10, 18, 1, 0, 0, 0, time.UTC),
			days: 0,
			want: []string{"1.29"},
		},
		{
			name: "no match",
			now:  time.Date(2024, 10, 9, 16, 0, 0, 0, time.UTC),
			days: 3,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := releasesToNotify(schedule, tc.now, tc.days)
			require.NoError(t, err)

			names := make([]string, 0, len(got))
			for _, r := range got {
				names = append(names, r.Release)
			}

			assert.ElementsMatch(t, tc.want, names)
		})
	}

	t.Run("invalid deadline", func(t *testing.T) {
		_, err := releasesToNotify(&model.PatchSchedule{
			Schedules: []*model.Schedule{{Release: "1.31", Next: &model.PatchRelease{CherryPickDeadline: "Oct 11"}}},
		}, time.Now(), 3)
		require.ErrorContains(t, err, "1.31")
	})
}

func TestEmailTemplate(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, emailTemplate.Execute(&buf, templateData{Releases: []templateRelease{
		{Release: "1.31", CherryPickDeadline: "2024-10-11"},
		{Release: `<script>alert(1)</script>`, CherryPickDeadline: "2024-10-11"},
	}}))

	out := buf.String()
	assert.Contains(t, out, "The cherry-pick deadline for the <b>1.31</b> branch is <b>2024-10-11 EOD PT.</b>")
	assert.Contains(t, out, "base%3Arelease-1.31+label%3Ado-not-merge%2Fcherry-pick-not-approved")
	assert.Contains(t, out, "-label%3Ado-not-merge%2Fhold")
	assert.NotContains(t, out, "<script>")
}

func TestRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schedule.yaml")
	require.NoError(t, os.WriteFile(path, []byte(testSchedule), 0o600))

	cfg := &config{
		fromEmail:    "release-managers@example.com",
		toEmail:      "dev@example.com",
		schedulePath: path,
		daysToAlert:  3,
	}
	matching := time.Date(2024, 10, 8, 16, 0, 0, 0, time.UTC)

	t.Run("mock sends to the sender address", func(t *testing.T) {
		sender := &fakeSender{}
		require.NoError(t, run(t.Context(), cfg, matching, sender))
		require.Len(t, sender.inputs, 1)

		in := sender.inputs[0]
		assert.Equal(t, []string{"release-managers@example.com"}, in.Destination.ToAddresses)
		assert.Equal(t, `"Release Managers" <release-managers@example.com>`, *in.FromEmailAddress)
		assert.Equal(t, emailSubject, *in.Content.Simple.Subject.Data)
		assert.Contains(t, *in.Content.Simple.Body.Html.Data, "<b>1.31</b>")
		assert.Contains(t, *in.Content.Simple.Body.Html.Data, "<b>1.30</b>")
		assert.NotContains(t, *in.Content.Simple.Body.Html.Data, "<b>1.29</b>")
	})

	t.Run("no mock sends to the recipient", func(t *testing.T) {
		sender := &fakeSender{}
		noMock := *cfg
		noMock.noMock = true

		require.NoError(t, run(t.Context(), &noMock, matching, sender))
		require.Len(t, sender.inputs, 1)
		assert.Equal(t, []string{"dev@example.com"}, sender.inputs[0].Destination.ToAddresses)
	})

	t.Run("no matching deadline sends nothing", func(t *testing.T) {
		sender := &fakeSender{}
		require.NoError(t, run(t.Context(), cfg, matching.AddDate(0, 0, 1), sender))
		assert.Empty(t, sender.inputs)
	})

	t.Run("send error is returned", func(t *testing.T) {
		sender := &fakeSender{err: errors.New("throttled")}
		require.ErrorContains(t, run(t.Context(), cfg, matching, sender), "throttled")
	})

	t.Run("invalid schedule", func(t *testing.T) {
		bad := filepath.Join(t.TempDir(), "bad.yaml")
		require.NoError(t, os.WriteFile(bad, []byte("schedules: {"), 0o600))

		badCfg := *cfg
		badCfg.schedulePath = bad
		require.ErrorContains(t, run(t.Context(), &badCfg, matching, &fakeSender{}), "decoding the schedule")
	})
}

func TestLoadSchedule(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/schedule.yaml":
			_, _ = w.Write([]byte(testSchedule))
		case "/big.yaml":
			_, _ = w.Write([]byte(strings.Repeat("a", maxScheduleSize+1)))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	orig := httpClient
	httpClient = srv.Client()

	t.Cleanup(func() { httpClient = orig })

	data, err := loadSchedule(t.Context(), srv.URL+"/schedule.yaml")
	require.NoError(t, err)
	assert.Equal(t, testSchedule, string(data))

	_, err = loadSchedule(t.Context(), srv.URL+"/missing.yaml")
	require.ErrorContains(t, err, "404")

	_, err = loadSchedule(t.Context(), srv.URL+"/big.yaml")
	require.ErrorContains(t, err, "larger than")

	_, err = loadSchedule(t.Context(), "http://example.com/schedule.yaml")
	require.ErrorContains(t, err, "plain http")
}

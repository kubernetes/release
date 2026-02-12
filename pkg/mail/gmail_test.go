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

package mail_test

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/api/gmail/v1"

	"k8s.io/release/pkg/mail"
	"k8s.io/release/pkg/mail/mailfakes"
)

func TestBuildMessage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sender     mail.Recipient
		recipients []mail.Recipient
		subject    string
		body       string
		expect     func(t *testing.T, msg string)
	}{
		"full message with sender": {
			sender:     mail.Recipient{Name: "Release Manager", Address: "rm@example.com"},
			recipients: []mail.Recipient{{Name: "announce", Address: "announce@googlegroups.com"}},
			subject:    "Kubernetes v1.35.0 is live!",
			body:       "<h1>Release</h1>",
			expect: func(t *testing.T, msg string) {
				require.Contains(t, msg, "From: Release Manager <rm@example.com>\r\n")
				require.Contains(t, msg, "To: announce <announce@googlegroups.com>\r\n")
				require.Contains(t, msg, "Subject: Kubernetes v1.35.0 is live!\r\n")
				require.Contains(t, msg, "MIME-Version: 1.0\r\n")
				require.Contains(t, msg, "Content-Type: text/html; charset=\"UTF-8\"\r\n")
				require.True(t, strings.HasSuffix(msg, "<h1>Release</h1>"))
			},
		},
		"no sender omits From header": {
			recipients: []mail.Recipient{{Name: "test", Address: "test@example.com"}},
			subject:    "Test",
			body:       "<p>body</p>",
			expect: func(t *testing.T, msg string) {
				require.NotContains(t, msg, "From:")
				require.True(t, strings.HasPrefix(msg, "To:"))
			},
		},
		"recipient without name uses address only": {
			recipients: []mail.Recipient{{Address: "bare@example.com"}},
			subject:    "Test",
			body:       "body",
			expect: func(t *testing.T, msg string) {
				require.Contains(t, msg, "To: bare@example.com\r\n")
				require.NotContains(t, msg, "To:  <bare@example.com>")
			},
		},
		"multiple recipients comma-separated": {
			sender: mail.Recipient{Name: "RM", Address: "rm@example.com"},
			recipients: []mail.Recipient{
				{Name: "announce", Address: "announce@googlegroups.com"},
				{Address: "dev@kubernetes.io"},
				{Name: "test", Address: "test@googlegroups.com"},
			},
			subject: "Release",
			body:    "body",
			expect: func(t *testing.T, msg string) {
				require.Contains(t, msg, "announce <announce@googlegroups.com>, dev@kubernetes.io, test <test@googlegroups.com>")
			},
		},
		"empty body": {
			recipients: []mail.Recipient{{Name: "r", Address: "r@example.com"}},
			subject:    "Subject",
			body:       "",
			expect: func(t *testing.T, msg string) {
				require.Contains(t, msg, "Subject: Subject\r\n")
				require.True(t, strings.HasSuffix(msg, "\r\n\r\n"))
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			msg := mail.BuildMessage(tc.sender, tc.recipients, tc.subject, tc.body)
			tc.expect(t, msg)
		})
	}
}

func TestGmailSenderSend(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		description string
		recipients  []mail.Recipient
		body        string
		subject     string
		sendErr     error
		expectErr   bool
		errContains string
	}{
		{
			description: "successful send with single recipient",
			recipients:  []mail.Recipient{{Name: "k8s-announce", Address: "announce@googlegroups.com"}},
			body:        "<h1>Release</h1>",
			subject:     "Kubernetes v1.30.0 is live!",
		},
		{
			description: "successful send with multiple recipients",
			recipients: []mail.Recipient{
				{Name: "announce", Address: "announce@googlegroups.com"},
				{Name: "dev", Address: "dev@kubernetes.io"},
			},
			body:    "<h1>Release</h1>",
			subject: "Kubernetes v1.30.0 is live!",
		},
		{
			description: "recipient without name",
			recipients:  []mail.Recipient{{Address: "bare@example.com"}},
			body:        "<p>test</p>",
			subject:     "Test",
		},
		{
			description: "no recipients returns error",
			body:        "<h1>Release</h1>",
			subject:     "Test",
			expectErr:   true,
			errContains: "no recipients set",
		},
		{
			description: "Gmail API error is propagated",
			recipients:  []mail.Recipient{{Name: "test", Address: "test@example.com"}},
			body:        "<h1>Release</h1>",
			subject:     "Test",
			sendErr:     errors.New("gmail API error"),
			expectErr:   true,
			errContains: "sending email via Gmail API",
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			fakeService := &mailfakes.FakeGmailService{}
			fakeService.SendMessageReturns(&gmail.Message{}, tc.sendErr)

			g := mail.NewGmailSenderWithService(fakeService)

			if len(tc.recipients) > 0 {
				g.SetRecipients(tc.recipients)
			}

			err := g.Send(tc.body, tc.subject)

			if tc.expectErr {
				require.Error(t, err)

				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}

				return
			}

			require.NoError(t, err)
			require.Equal(t, 1, fakeService.SendMessageCallCount())

			userID, msg := fakeService.SendMessageArgsForCall(0)
			require.Equal(t, "me", userID)
			require.NotEmpty(t, msg.Raw)

			decoded, err := base64.URLEncoding.DecodeString(msg.Raw)
			require.NoError(t, err)

			rawMsg := string(decoded)
			require.Contains(t, rawMsg, "Subject: "+tc.subject)
			require.Contains(t, rawMsg, tc.body)
			require.Contains(t, rawMsg, "Content-Type: text/html")
			require.NotContains(t, rawMsg, "From:")

			for _, r := range tc.recipients {
				require.Contains(t, rawMsg, r.Address)
			}
		})
	}
}

func TestGmailSenderSetGoogleGroupRecipients(t *testing.T) {
	t.Parallel()

	fakeService := &mailfakes.FakeGmailService{}
	fakeService.SendMessageReturns(&gmail.Message{}, nil)

	g := mail.NewGmailSenderWithService(fakeService)
	g.SetGoogleGroupRecipients(
		mail.KubernetesAnnounceGoogleGroup,
		mail.KubernetesDevGoogleGroup,
	)

	err := g.Send("<h1>Test</h1>", "Test Subject")
	require.NoError(t, err)

	_, msg := fakeService.SendMessageArgsForCall(0)
	decoded, err := base64.URLEncoding.DecodeString(msg.Raw)
	require.NoError(t, err)

	rawMsg := string(decoded)
	require.Contains(t, rawMsg, "kubernetes-announce@googlegroups.com")
	require.Contains(t, rawMsg, "dev@kubernetes.io")
}

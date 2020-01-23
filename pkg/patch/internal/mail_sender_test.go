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

package internal_test

import (
	"fmt"
	"testing"

	"github.com/sendgrid/rest"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/patch/internal"
	"k8s.io/release/pkg/patch/internal/internalfakes"
	it "k8s.io/release/pkg/patch/internal/testing"
)

func TestMailSender(t *testing.T) {
	t.Parallel()

	it.Run(t, "SetRecipients", testRecipient)
	it.Run(t, "SetSender", testSender)
	it.Run(t, "Send", testSend)

	it.Run(t, "main", func(t *testing.T) {
		m := &internal.MailSender{
			SendgridClientCreator: func(_ string) internal.SendgridClient {
				c := &internalfakes.FakeSendgridClient{}
				c.SendReturns(&rest.Response{
					Body:       "some API response",
					StatusCode: 202,
				}, nil)
				return c
			},
		}
		require.NoError(t, m.SetSender("Jane Doe", "djane@example.org"))
		require.NoError(t, m.SetRecipients("Max Mustermann", "mmustermann@example.org"))
		require.NoError(t, m.Send("some content", "some subject"))
	})
}

func testSend(t *testing.T) {
	tests := map[string]struct {
		sendgridSendResponse *rest.Response
		sendgridSendErr      error
		message              string
		subject              string
		apiKey               string

		expectedSendgridAPIKey string
		expectedErr            string
	}{
		"the token is used": {
			apiKey:                 "some key",
			sendgridSendResponse:   simpleRespons("", 202),
			expectedSendgridAPIKey: "some key",
		},
		"when #Send returns an error, bubble it up": {
			sendgridSendErr: fmt.Errorf("some sendgrid err"),
			expectedErr:     "some sendgrid err",
		},
		"when #Send returns an empty response, an error is returned": {
			expectedErr: "empty API response",
		},
		"when #Send returns an invalid status code, an error holding the API response is returned": {
			sendgridSendResponse: simpleRespons("some API response", 500),
			expectedErr:          "some API response",
		},
	}

	for name, tc := range tests {
		tc := tc

		it.Run(t, name, func(t *testing.T) {
			m := &internal.MailSender{
				APIKey: tc.apiKey,
			}

			sgClient := &internalfakes.FakeSendgridClient{}
			sgClient.SendReturns(tc.sendgridSendResponse, tc.sendgridSendErr)

			m.SendgridClientCreator = func(apiKey string) internal.SendgridClient {
				require.Equal(t, tc.expectedSendgridAPIKey, apiKey, "SendgridClient#creator arg")
				return sgClient
			}

			err := m.Send(tc.message, tc.subject)
			it.CheckErrSub(t, err, tc.expectedErr)

			require.Equal(t, 1, sgClient.SendCallCount(), "SendgridClient#Send call count")

			mail := sgClient.SendArgsForCall(0)
			require.Equalf(t, tc.subject, mail.Subject, "the mail's subject")
			require.Equalf(t, tc.message, mail.Content[0].Value, "the mail's body")
		})
	}
}

func simpleRespons(body string, code int) *rest.Response {
	return &rest.Response{Body: body, StatusCode: code}
}

func testSender(t *testing.T) {
	tests := map[string]struct {
		senderName  string
		senderEmail string
		expectedErr string
	}{
		"happy path": {
			senderName:  "name",
			senderEmail: "email",
		},
		"when email is empty, error": {
			senderName:  "name",
			senderEmail: "",
			expectedErr: "email must not be empty",
		},
	}

	for name, tc := range tests {
		tc := tc
		it.Run(t, name, func(t *testing.T) {
			m := &internal.MailSender{}
			err := m.SetSender(tc.senderName, tc.senderEmail)
			it.CheckErr(t, err, tc.expectedErr)
		})
	}
}

func testRecipient(t *testing.T) {
	tests := map[string]struct {
		recipientArgs [][]string
		expectedErr   string
	}{
		"when # of recipient args is even, succeed": {
			recipientArgs: [][]string{
				{},
				{"name", "email"},
				{"name", "email", "otherName", "otherEmail"},
			},
		},
		"when # of recipients args is not even, error": {
			recipientArgs: [][]string{
				{"one"},
				{"one", "two", "three"},
			},
			expectedErr: "must be called with alternating recipient's names and email addresses",
		},
		"when email is empty, error": {
			recipientArgs: [][]string{
				{"name", ""},
				{"name", "email", "otherName", ""},
			},
			expectedErr: "email must not be empty",
		},
	}

	for name, tc := range tests {
		tc := tc

		it.Run(t, name, func(t *testing.T) {
			for _, args := range tc.recipientArgs {
				it.Run(t, "", func(t *testing.T) {
					m := &internal.MailSender{}
					err := m.SetRecipients(args...)
					it.CheckErr(t, err, tc.expectedErr)
				})
			}
		})
	}
}

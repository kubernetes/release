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

package mail_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sendgrid/rest"
	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/mail"
	"k8s.io/release/pkg/mail/mailfakes"
	it "k8s.io/release/pkg/testing"
)

func TestSetDefaultSender(t *testing.T) {
	for _, tc := range []struct {
		description string
		prep        func(*mailfakes.FakeAPIClient)
		expect      func(error, string)
	}{
		{
			description: "Should succeed",
			prep: func(c *mailfakes.FakeAPIClient) {
				c.APIReturnsOnCall(0, &rest.Response{
					StatusCode: 200,
					Body:       `{"email": "me@mail.com"}`,
				}, nil)
				c.APIReturnsOnCall(1, &rest.Response{
					StatusCode: 200,
					Body:       `{"first_name": "Firstname", "last_name": "Lastname"}`,
				}, nil)
			},
			expect: func(err error, desc string) { require.Nil(t, err, desc) },
		},
		{
			description: "Should fail with wrong JSON in second response",
			prep: func(c *mailfakes.FakeAPIClient) {
				c.APIReturnsOnCall(0, &rest.Response{
					StatusCode: 200,
					Body:       `{"email": "me@mail.com"}`,
				}, nil)
				c.APIReturnsOnCall(1, &rest.Response{
					StatusCode: 200,
					Body:       "wrong-json",
				}, nil)
			},
			expect: func(err error, desc string) { require.NotNil(t, err, desc) },
		},
		{
			description: "Should fail with wrong status code in second response",
			prep: func(c *mailfakes.FakeAPIClient) {
				c.APIReturnsOnCall(0, &rest.Response{
					StatusCode: 200,
					Body:       `{"email": "me@mail.com"}`,
				}, nil)
				c.APIReturnsOnCall(1, &rest.Response{StatusCode: 400}, nil)
			},
			expect: func(err error, desc string) { require.NotNil(t, err, desc) },
		},
		{
			description: "Should fail with failing second response",
			prep: func(c *mailfakes.FakeAPIClient) {
				c.APIReturnsOnCall(0, &rest.Response{
					StatusCode: 200,
					Body:       `{"email": "me@mail.com"}`,
				}, nil)
				c.APIReturnsOnCall(1, nil, errors.New(""))
			},
			expect: func(err error, desc string) { require.NotNil(t, err, desc) },
		},
		{
			description: "Should fail with wrong JSON in first response",
			prep: func(c *mailfakes.FakeAPIClient) {
				c.APIReturnsOnCall(0, &rest.Response{
					StatusCode: 200,
					Body:       "wrong-json",
				}, nil)
			},
			expect: func(err error, desc string) { require.NotNil(t, err, desc) },
		},
		{
			description: "Should fail with wrong status code in first response",
			prep: func(c *mailfakes.FakeAPIClient) {
				c.APIReturnsOnCall(0, &rest.Response{StatusCode: 400}, nil)
			},
			expect: func(err error, desc string) { require.NotNil(t, err, desc) },
		},
		{
			description: "Should fail with failing first response",
			prep: func(c *mailfakes.FakeAPIClient) {
				c.APIReturnsOnCall(0, nil, errors.New(""))
			},
			expect: func(err error, desc string) { require.NotNil(t, err, desc) },
		},
	} {
		m := mail.NewSender("")
		c := &mailfakes.FakeAPIClient{}
		tc.prep(c)
		m.SetAPIClient(c)

		err := m.SetDefaultSender()

		tc.expect(err, tc.description)
	}
}

func TestMailSender(t *testing.T) {
	t.Parallel()

	it.Run(t, "SetRecipients", testRecipient)
	it.Run(t, "SetSender", testSender)
	it.Run(t, "Send", testSend)

	it.Run(t, "main", func(t *testing.T) {
		m := mail.NewSender("")
		c := &mailfakes.FakeSendClient{}
		c.SendReturns(&rest.Response{
			Body:       "some API response",
			StatusCode: 202,
		}, nil)
		m.SetSendClient(c)
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
			m := mail.NewSender(tc.apiKey)

			sgClient := &mailfakes.FakeSendClient{}
			sgClient.SendReturns(tc.sendgridSendResponse, tc.sendgridSendErr)

			m.SetSendClient(sgClient)
			require.Equal(t, tc.expectedSendgridAPIKey, tc.apiKey, "SendgridClient#creator arg")

			err := m.Send(tc.message, tc.subject)
			it.CheckErrSub(t, err, tc.expectedErr)

			require.Equal(t, 1, sgClient.SendCallCount(), "SendgridClient#Send call count")

			email := sgClient.SendArgsForCall(0)
			require.Equalf(t, tc.subject, email.Subject, "the mail's subject")
			require.Equalf(t, tc.message, email.Content[0].Value, "the mail's body")
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
			m := &mail.Sender{}
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
					m := &mail.Sender{}
					err := m.SetRecipients(args...)
					it.CheckErr(t, err, tc.expectedErr)
				})
			}
		})
	}
}

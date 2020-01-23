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

package internal

import (
	"fmt"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"k8s.io/release/pkg/log"
)

type MailSender struct {
	log.Mixin

	SendgridClientCreator SendgridClientCreator
	APIKey                string

	sender     *mail.Email
	recipients []*mail.Email
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . SendgridClient
type SendgridClient interface {
	Send(*mail.SGMailV3) (*rest.Response, error)
}

type SendgridClientCreator func(apiKey string) SendgridClient

func (c SendgridClientCreator) create(apiKey string) SendgridClient {
	if c == nil {
		c = defaultSendgridClientCreator
	}
	return c(apiKey)
}

var defaultSendgridClientCreator = func(apiKey string) SendgridClient {
	return sendgrid.NewSendClient(apiKey)
}

func (m *MailSender) Send(body, subject string) error {
	html := mail.NewContent("text/html", body)

	p := mail.NewPersonalization()
	p.AddTos(m.recipients...)

	msg := mail.NewV3Mail().
		SetFrom(m.sender).
		AddContent(html).
		AddPersonalizations(p)
	msg.Subject = subject

	m.Logger().WithField("message", msg).Trace("message prepared")

	client := m.SendgridClientCreator.create(m.APIKey)
	res, err := client.Send(msg)
	if err != nil {
		return err
	}
	if res == nil {
		return &SendError{code: -1, resBody: "<empty API response>"}
	}
	if c := res.StatusCode; c < 200 || c >= 300 {
		return &SendError{code: res.StatusCode, resBody: res.Body, resHeaders: fmt.Sprintf("%#v", res.Headers)}
	}

	m.Logger().Debug("mail successfully sent")
	return nil
}

type SendError struct {
	code       int
	resBody    string
	resHeaders string
}

func (e *SendError) Error() string {
	return fmt.Sprintf("got code %d while sending: Body: %q, Header: %q", e.code, e.resBody, e.resHeaders)
}

func (m *MailSender) SetSender(name, email string) error {
	if email == "" {
		return fmt.Errorf("email must not be empty")
	}
	m.sender = mail.NewEmail(name, email)
	m.Logger().WithField("sender", m.sender).Debugf("sender set")
	return nil
}

func (m *MailSender) SetRecipients(recipientArgs ...string) error {
	l := len(recipientArgs)

	if l%2 != 0 {
		return fmt.Errorf("must be called with alternating recipient's names and email addresses")
	}

	recipients := make([]*mail.Email, l/2)

	for i := range recipients {
		name := recipientArgs[i*2]
		email := recipientArgs[i*2+1]
		if email == "" {
			return fmt.Errorf("email must not be empty")
		}
		recipients[i] = mail.NewEmail(name, email)
	}

	m.recipients = recipients
	m.Logger().WithField("recipients", m.sender).Debugf("recipients set")

	return nil
}

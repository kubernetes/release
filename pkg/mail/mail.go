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

package mail

import "fmt"

// EmailSender is a generic interface for sending emails.
// Implementations include GmailSender (Gmail OAuth) with SMTP planned
// for the future.
type EmailSender interface {
	Send(body, subject string) error
}

// Recipient represents an email recipient with a name and address.
type Recipient struct {
	Name    string
	Address string
}

// GoogleGroup is a simple google group representation.
type GoogleGroup string

const (
	KubernetesAnnounceGoogleGroup     GoogleGroup = "kubernetes-announce"
	KubernetesDevGoogleGroup          GoogleGroup = "dev"
	KubernetesAnnounceTestGoogleGroup GoogleGroup = "kubernetes-announce-test"
)

// GoogleGroupRecipients resolves GoogleGroup values into Recipients.
func GoogleGroupRecipients(groups ...GoogleGroup) []Recipient {
	recipients := make([]Recipient, 0, len(groups))
	for _, group := range groups {
		var addr string
		if group == KubernetesDevGoogleGroup {
			addr = fmt.Sprintf("%s@kubernetes.io", group)
		} else {
			addr = fmt.Sprintf("%s@googlegroups.com", group)
		}

		recipients = append(recipients, Recipient{
			Name:    string(group),
			Address: addr,
		})
	}

	return recipients
}

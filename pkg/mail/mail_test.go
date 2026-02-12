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
	"testing"

	"github.com/stretchr/testify/require"

	"k8s.io/release/pkg/mail"
)

func TestGoogleGroupRecipients(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		groups   []mail.GoogleGroup
		expected []mail.Recipient
	}{
		"announce and dev groups": {
			groups: []mail.GoogleGroup{
				mail.KubernetesAnnounceGoogleGroup,
				mail.KubernetesDevGoogleGroup,
			},
			expected: []mail.Recipient{
				{
					Name:    "kubernetes-announce",
					Address: "kubernetes-announce@googlegroups.com",
				},
				{
					Name:    "dev",
					Address: "dev@kubernetes.io",
				},
			},
		},
		"test group": {
			groups: []mail.GoogleGroup{
				mail.KubernetesAnnounceTestGoogleGroup,
			},
			expected: []mail.Recipient{
				{
					Name:    "kubernetes-announce-test",
					Address: "kubernetes-announce-test@googlegroups.com",
				},
			},
		},
		"empty groups": {
			groups:   []mail.GoogleGroup{},
			expected: []mail.Recipient{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			recipients := mail.GoogleGroupRecipients(tc.groups...)
			require.Equal(t, tc.expected, recipients)
		})
	}
}

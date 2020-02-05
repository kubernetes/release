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

package patch_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/release/pkg/patch"
	"k8s.io/release/pkg/patch/internal/internalfakes"
	it "k8s.io/release/pkg/patch/internal/testing"
)

type opts = patch.AnnounceOptions

type testCase struct {
	releaseNoterErr        error
	releaseNoterOutput     string
	mailerSenderErr        error
	mailerSetRecipientsErr error
	mailerSetSenderErr     error
	formatterOutput        string
	formattterErr          error
	workspaceStatus        map[string]string
	workspaceErr           error
	opts                   opts

	expectedReleaseNoterNOTToBeCalled bool
	expectedFormatterNOTToBeCalled    bool
	expectedMailerNOTToBeCalled       bool

	expectedFormatterMarkdown []*regexp.Regexp
	expectedMailerBody        []*regexp.Regexp
	expectedFormatterSubject  []*regexp.Regexp
	expectedMailerSubject     []*regexp.Regexp
	expectedErrMsg            string
	expectedRecipients        *[]string
	expectedSender            [2]string
}

func getOpts(funcs ...func(*opts)) opts {
	o := patch.AnnounceOptions{
		FreezeDate: "2010-11-05",
		CutDate:    "2010-11-12",
	}
	for _, f := range funcs {
		f(&o)
	}
	return o
}

func res(ss ...string) []*regexp.Regexp {
	res := make([]*regexp.Regexp, len(ss))
	for i, s := range ss {
		res[i] = regexp.MustCompile(s)
	}
	return res
}

func TestAnnounce(t *testing.T) {
	t.Parallel()

	tests := map[string]testCase{
		"happy path": {
			opts:                     getOpts(),
			workspaceStatus:          map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			releaseNoterOutput:       "some release notes content",
			expectedFormatterSubject: res("^Kubernetes v1.13.10 cut planned for Friday, 2010-11-12$"),
			expectedFormatterMarkdown: res(
				"v1.13.10",
				"Friday, 2010-11-05",
				"some release notes content",
			),
			formatterOutput:       "some formatted html",
			expectedMailerBody:    res("^some formatted html$"),
			expectedMailerSubject: res("^Kubernetes v1.13.10 cut planned for Friday, 2010-11-12$"),
		},
		"when getting the workspace status returns an error, the error bubbles up and the mail is never sent": {
			workspaceErr:                      fmt.Errorf("git describe err"),
			expectedErrMsg:                    "git describe err",
			expectedReleaseNoterNOTToBeCalled: true,
			expectedMailerNOTToBeCalled:       true,
			expectedFormatterNOTToBeCalled:    true,
		},
		"when the workspace status does not hold a git version, the error bubbles up and the mail is never sent": {
			workspaceStatus:                   map[string]string{},
			expectedErrMsg:                    "has no field",
			expectedReleaseNoterNOTToBeCalled: true,
			expectedMailerNOTToBeCalled:       true,
			expectedFormatterNOTToBeCalled:    true,
		},
		"when release notes returns an error, the error bubbles up and the mail is never sent": {
			workspaceStatus:                map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			opts:                           getOpts(),
			expectedFormatterNOTToBeCalled: true,
			expectedMailerNOTToBeCalled:    true,
			releaseNoterErr:                fmt.Errorf("some release notes error"),
			expectedErrMsg:                 "some release notes error",
		},
		"when the formatter fails, the error bubbles up and the mail is never sent": {
			workspaceStatus:             map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			opts:                        getOpts(),
			formattterErr:               fmt.Errorf("some formatter error"),
			expectedMailerNOTToBeCalled: true,
			expectedErrMsg:              "some formatter error",
		},
		"when the mail sender fails, the error bubbles up": {
			workspaceStatus: map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			opts:            getOpts(),
			mailerSenderErr: fmt.Errorf("some mail sender error"),
			expectedErrMsg:  "some mail sender error",
		},
		"when in mock mode, sets the sender as the recipient": {
			workspaceStatus: map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			opts: getOpts(func(o *opts) {
				o.SenderEmail = "sender email"
				o.SenderName = "sender name"
			}),
			expectedRecipients: &[]string{"sender name", "sender email"},
			expectedSender:     [...]string{"sender name", "sender email"},
		},
		"when in nomock mode, sets the mailinglists as the recipient": {
			workspaceStatus: map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			opts:            getOpts(func(o *opts) { o.Nomock = true }),
			expectedRecipients: &[]string{
				patch.KDevName, patch.KDevEmail,
				patch.KDevAnnounceName, patch.KDevAnnounceEmail,
			},
		},
		"when setting the recipients fails, the error bubbles up": {
			workspaceStatus:             map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			opts:                        getOpts(),
			expectedMailerNOTToBeCalled: true,
			mailerSetRecipientsErr:      fmt.Errorf("some recipients error"),
			expectedErrMsg:              "some recipients error",
		},
		"when setting the sender fails, the error bubbles up": {
			workspaceStatus:             map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			opts:                        getOpts(),
			expectedMailerNOTToBeCalled: true,
			mailerSetSenderErr:          fmt.Errorf("some sender error"),
			expectedErrMsg:              "some sender error",
		},
		"when cut date parsing fails, the error bubbles up": {
			workspaceStatus:                   map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			opts:                              getOpts(func(o *opts) { o.CutDate = "invalid cut date" }),
			expectedReleaseNoterNOTToBeCalled: true,
			expectedMailerNOTToBeCalled:       true,
			expectedFormatterNOTToBeCalled:    true,
			expectedErrMsg:                    `cannot parse "invalid cut date"`,
		},
		"when freeze date parsing fails, the error bubbles up": {
			workspaceStatus:                   map[string]string{"gitVersion": "v1.13.10-beta.0-16-g48844ef5e7"},
			opts:                              getOpts(func(o *opts) { o.FreezeDate = "invalid freeze date" }),
			expectedReleaseNoterNOTToBeCalled: true,
			expectedMailerNOTToBeCalled:       true,
			expectedFormatterNOTToBeCalled:    true,
			expectedErrMsg:                    `cannot parse "invalid freeze date"`,
		},
	}

	for name, tc := range tests {
		tc := tc

		it.Run(t, name, func(t *testing.T) {
			ws := &internalfakes.FakeWorkspace{}
			ws.StatusReturns(tc.workspaceStatus, tc.workspaceErr)

			rn := &internalfakes.FakeReleaseNoter{}
			rn.GetMarkdownReturns(tc.releaseNoterOutput, tc.releaseNoterErr)

			f := &internalfakes.FakeFormatter{}
			f.MarkdownToHTMLReturns(tc.formatterOutput, tc.formattterErr)

			ms := &internalfakes.FakeMailSender{}
			ms.SendReturns(tc.mailerSenderErr)
			ms.SetRecipientsReturns(tc.mailerSetRecipientsErr)
			ms.SetSenderReturns(tc.mailerSetSenderErr)

			announcer := &patch.Announcer{
				Opts:         tc.opts,
				Workspace:    ws,
				ReleaseNoter: rn,
				MailSender:   ms,
				Formatter:    f,
			}

			err := announcer.Run()
			it.CheckErrSub(t, err, tc.expectedErrMsg)

			require.Equal(t, 1, ws.StatusCallCount(), "Workspace#Status call count")

			checkReleaseNoter(t, rn, &tc)
			checkFormatter(t, f, &tc)
			checkMailSender(t, ms, &tc)
		})
	}
}

func checkReleaseNoter(t *testing.T, rn *internalfakes.FakeReleaseNoter, tc *testCase) {
	cc := rn.GetMarkdownCallCount()
	if tc.expectedReleaseNoterNOTToBeCalled {
		require.Equal(t, 0, cc, "ReleaseNoter#GetMarkdown call count")
		return
	}
	require.Equal(t, 1, cc, "ReleaseNoter#GetMarkdown call count")
}

func checkFormatter(t *testing.T, f *internalfakes.FakeFormatter, tc *testCase) {
	cc := f.MarkdownToHTMLCallCount()
	if tc.expectedFormatterNOTToBeCalled {
		require.Equal(t, 0, cc, "Formatter#MarkdownToHTML call count")
		return
	}
	require.Equal(t, 1, cc, "Formatter#MarkdownToHTML call count")

	content, subject := f.MarkdownToHTMLArgsForCall(0)

	for _, re := range tc.expectedFormatterMarkdown {
		require.Regexp(t, re, content)
	}
	for _, re := range tc.expectedFormatterSubject {
		require.Regexp(t, re, subject)
	}
}

func checkMailSender(t *testing.T, ms *internalfakes.FakeMailSender, tc *testCase) {
	cc := ms.SendCallCount()
	if tc.expectedMailerNOTToBeCalled {
		require.Equal(t, 0, cc, "Mailer#Send call count")
		return
	}
	require.Equal(t, 1, cc, "Mailer#Send call count")

	body, subject := ms.SendArgsForCall(0)

	for _, re := range tc.expectedMailerBody {
		require.Regexp(t, re, body)
	}
	for _, re := range tc.expectedMailerSubject {
		require.Regexp(t, re, subject)
	}

	if r := tc.expectedRecipients; r != nil {
		require.Equal(t, *r, ms.SetRecipientsArgsForCall(0))
	}

	sName, sEmail := ms.SetSenderArgsForCall(0)
	require.Equalf(t, tc.expectedSender[0], sName, "Sender name")
	require.Equalf(t, tc.expectedSender[1], sEmail, "Sender email")
}

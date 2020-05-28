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

package patch

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"k8s.io/release/pkg/log"
	"k8s.io/release/pkg/mail"
	"k8s.io/release/pkg/patch/internal"
)

type AnnounceOptions struct {
	SenderEmail     string
	SenderName      string
	FreezeDate      string
	CutDate         string
	Nomock          bool
	K8sRepoPath     string
	ReleaseRepoPath string
	SendgridAPIKey  string
	GithubToken     string
}

type Announcer struct {
	log.Mixin

	Opts         AnnounceOptions
	ReleaseNoter ReleaseNoter
	MailSender   MailSender
	Formatter    Formatter
	Workspace    Workspace
}

const (
	KDevName          = "Kubernetes developer/contributor discussion"
	KDevEmail         = "kubernetes-dev@googlegroups.com"
	KDevAnnounceName  = "kubernetes-dev-announce"
	KDevAnnounceEmail = "kubernetes-dev-announce@googlegroups.com"

	ReleaseManagerName         = "Kubernetes Release Managers"
	ReleaseManagerTag          = "@kubernetes/patch-release-team"
	ReleaseManagerEmail        = "release-managers@kubernetes.io"
	ReleaseManagerSlackChannel = "release-management"
)

func (a *Announcer) Run() error {
	ver, err := a.getUpcomingVer()
	if err != nil {
		a.Logger().WithError(err).Debug("getting upcoming version failed")
		return err
	}

	a.Logger().Infof("Running for %q", ver)

	freezeDate, err := time.Parse(dateLayoutISO8601, a.Opts.FreezeDate)
	if err != nil {
		a.Logger().WithError(err).Debug("parsing freeze date failed")
		return err
	}
	cutDate, err := time.Parse(dateLayoutISO8601, a.Opts.CutDate)
	if err != nil {
		a.Logger().WithError(err).Debug("parsing cut date failed")
		return err
	}

	subject := "Kubernetes " + ver + " cut planned for " + dateFormatHuman(cutDate)

	head, err := a.getMailHead(ver, freezeDate, cutDate)
	if err != nil {
		a.Logger().WithError(err).Debug("getting mail head failed")
		return err
	}

	relNotes, err := a.getReleaseNotes()
	if err != nil {
		a.Logger().WithError(err).Debug("getting release notes failed")
		return err
	}

	body, err := a.formatAsHTML(subject, head, relNotes)
	if err != nil {
		a.Logger().WithError(err).Debug("formatting mail as html failed")
		return err
	}

	a.Logger().
		WithField("emailBody", body).
		WithField("emailSubject", subject).
		Trace("email content generated")

	if err := a.sendMail(body, subject); err != nil {
		a.Logger().WithError(err).Debug("sending mail failed")
		return err
	}

	return nil
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o internal/internalfakes/fake_mail_sender.go . MailSender
type MailSender interface {
	SetSender(name, email string) error
	SetRecipients(recipients ...string) error
	Send(content, subject string) error
}

//counterfeiter:generate -o internal/internalfakes/fake_release_noter.go . ReleaseNoter
type ReleaseNoter interface {
	GetMarkdown() (relnotes string, err error)
}

//counterfeiter:generate -o internal/internalfakes/fake_formatter.go . Formatter
type Formatter interface {
	MarkdownToHTML(markdown, title string) (html string, err error)
}

//counterfeiter:generate -o internal/internalfakes/fake_workspace.go . Workspace
type Workspace interface {
	Status() (status map[string]string, err error)
}

// The date/time layouts to parse and format `time.Time`s
// Reference date is 'Mon Jan 2 15:04:05 MST 2006'
const (
	dateLayoutISO8601    = "2006-01-02"
	dateLayoutDayISO8601 = "Monday, 2006-01-02"
)

func dateFormatHuman(t time.Time) string {
	return t.Format(dateLayoutDayISO8601)
}

func loadTemplate(tmplString string) (*template.Template, error) {
	funcs := template.FuncMap{
		"dateFormatHuman": dateFormatHuman,
		"code":            func(s string) string { return "`" + s + "`" },
		"codeBlock":       func(s string) string { return "```\n" + s + "\n```" },
		"link":            func(n, t string) string { return "[" + n + "](" + t + ")" },
	}
	return template.New("main").Funcs(funcs).Parse(tmplString)
}

func (a *Announcer) getMailHead(version string, freezeDate, cutDate time.Time) (string, error) {
	tmpl, err := loadTemplate(MailHeadMarkdown)
	if err != nil {
		return "", err
	}

	templated := &bytes.Buffer{}
	templateData := struct {
		DateFreeze                 time.Time
		DateCut                    time.Time
		Version                    string
		ReleaseManagerName         string
		ReleaseManagerEmail        string
		ReleaseManagerTag          string
		ReleaseManagerSlackChannel string
	}{
		DateFreeze:                 freezeDate,
		DateCut:                    cutDate,
		Version:                    version,
		ReleaseManagerName:         ReleaseManagerName,
		ReleaseManagerTag:          ReleaseManagerTag,
		ReleaseManagerEmail:        ReleaseManagerEmail,
		ReleaseManagerSlackChannel: ReleaseManagerSlackChannel,
	}

	err = tmpl.Execute(templated, templateData)
	if err != nil {
		return "", err
	}

	return templated.String(), nil
}

func (a *Announcer) getUpcomingVer() (string, error) {
	if a.Workspace == nil {
		w := &internal.Workspace{
			K8sRepoPath: a.Opts.K8sRepoPath,
		}
		a.Workspace = w
		a.Logger().Debug("new workspace created")
	}

	status, err := a.Workspace.Status()
	if err != nil {
		return "", err
	}

	// v1.18.0-alpha.2.121+e92a7cfd2a82cd-dirty
	gitVersion, ok := status["gitVersion"]
	if !ok {
		return "", fmt.Errorf("workspace status has no field 'gitVersion': %q", status)
	}

	i := strings.IndexRune(gitVersion, '-')
	if i < 0 {
		return "", fmt.Errorf("cannot extract upcoming version from gitVersion %q", gitVersion)
	}

	return gitVersion[:i], nil
}

func (a *Announcer) formatAsHTML(title string, parts ...string) (string, error) {
	if a.Formatter == nil {
		f := &internal.Formatter{
			Style: MailStyle,
		}
		a.Formatter = f
		a.Logger().Debug("new formatter instance created")
	}

	html, err := a.Formatter.MarkdownToHTML(strings.Join(parts, hr), title)
	if err != nil {
		return "", err
	}

	return html, nil
}

func (a *Announcer) sendMail(content, subject string) error {
	if a.MailSender == nil {
		ms := mail.NewSender(a.Opts.SendgridAPIKey)
		a.MailSender = ms
		a.Logger().Debug("new instance of mail sender created")
	}

	if err := a.MailSender.SetSender(a.Opts.SenderName, a.Opts.SenderEmail); err != nil {
		return err
	}

	receipients := []string{a.Opts.SenderName, a.Opts.SenderEmail}
	if a.Opts.Nomock {
		receipients = []string{KDevName, KDevEmail, KDevAnnounceName, KDevAnnounceEmail}
		a.Logger().WithField("receipients", receipients).Info("setting mail recipients in nomock mode")
	}
	if err := a.MailSender.SetRecipients(receipients...); err != nil {
		return err
	}

	a.Logger().Debug("calling the mail sender")
	return a.MailSender.Send(content, subject)
}

func (a *Announcer) getReleaseNotes() (string, error) {
	if a.ReleaseNoter == nil {
		rn := &internal.ReleaseNoter{
			K8sDir:          a.Opts.K8sRepoPath,
			ReleaseToolsDir: a.Opts.ReleaseRepoPath,
			GithubToken:     a.Opts.GithubToken,
		}
		a.ReleaseNoter = rn
		a.Logger().Debug("new instance of release-noter created")
	}

	a.Logger().Debug("calling relase note generator")
	return a.ReleaseNoter.GetMarkdown()
}

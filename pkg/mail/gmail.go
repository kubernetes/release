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

package mail

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Default OAuth2 client credentials for krel. These are intentionally
// embedded in source code and are NOT secrets. Google documents this
// for "Desktop app" (installed) type OAuth clients:
// https://developers.google.com/identity/protocols/oauth2#installed
//
// These values only identify the application to Google. User security
// is provided by the interactive OAuth consent flow and the resulting
// access token, which is kept in memory only and valid for ~1 hour.
//
// The OAuth app is managed in the "k8s-release" Google Cloud project:
// https://console.cloud.google.com/auth/clients?project=k8s-release
const (
	defaultClientID     = "304687256732-2h7gv4smd613vscc3sic9n8kemq7m5fk.apps.googleusercontent.com" //nolint:gosec // not a secret
	defaultClientSecret = "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"                                      //nolint:gosec // not a secret
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . GmailService
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt mailfakes/fake_gmail_service.go > mailfakes/_fake_gmail_service.go && mv mailfakes/_fake_gmail_service.go mailfakes/fake_gmail_service.go"

// GmailService wraps the Gmail API send operation for testability.
type GmailService interface {
	// SendMessage sends a Gmail message on behalf of the given user.
	// The userID is typically "me" to indicate the authenticated user.
	SendMessage(userID string, message *gmail.Message) (*gmail.Message, error)
}

// GmailSender sends email through the Gmail API using OAuth2.
// The From header is automatically set by Gmail to the authenticated
// user's address.
type GmailSender struct {
	recipients []Recipient
	service    GmailService
}

// defaultGmailService wraps the real Gmail API service.
type defaultGmailService struct {
	svc *gmail.Service
}

func (d *defaultGmailService) SendMessage(userID string, message *gmail.Message) (*gmail.Message, error) {
	return d.svc.Users.Messages.Send(userID, message).Do()
}

// NewGmailSender creates a GmailSender by performing an OAuth2 flow
// using the embedded default client credentials and creating the Gmail
// API service. If noBrowser is true, the user will be prompted to
// manually open the authorization URL and paste the redirect URL.
func NewGmailSender(ctx context.Context, noBrowser bool) (*GmailSender, error) {
	config := &oauth2.Config{
		ClientID:     defaultClientID,
		ClientSecret: defaultClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.GmailSendScope},
	}

	var (
		authCode string
		err      error
	)

	if noBrowser {
		authCode, err = getAuthCodeManual(config)
	} else {
		authCode, err = getAuthCodeBrowser(ctx, config)
	}

	if err != nil {
		return nil, fmt.Errorf("getting authorization code: %w", err)
	}

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("exchanging auth code for token: %w", err)
	}

	client := config.Client(ctx, token)

	svc, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating Gmail service: %w", err)
	}

	return &GmailSender{
		service: &defaultGmailService{svc: svc},
	}, nil
}

// NewGmailSenderWithService creates a GmailSender with a pre-configured
// GmailService. This is primarily useful for testing.
func NewGmailSenderWithService(service GmailService) *GmailSender {
	return &GmailSender{service: service}
}

// SetRecipients sets the email recipients.
func (g *GmailSender) SetRecipients(recipients []Recipient) {
	g.recipients = recipients
}

// SetGoogleGroupRecipients sets Google Groups as recipients.
func (g *GmailSender) SetGoogleGroupRecipients(groups ...GoogleGroup) {
	g.recipients = GoogleGroupRecipients(groups...)
}

// Send sends an HTML email via the Gmail API. It satisfies the
// EmailSender interface.
func (g *GmailSender) Send(body, subject string) error {
	if len(g.recipients) == 0 {
		return errors.New("no recipients set")
	}

	raw := BuildMessage(Recipient{}, g.recipients, subject, body)

	message := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(raw)),
	}

	logrus.WithField("recipients", g.recipients).Debug("Sending message via Gmail API")

	if _, err := g.service.SendMessage("me", message); err != nil {
		return fmt.Errorf("sending email via Gmail API: %w", err)
	}

	logrus.Debug("Mail successfully sent via Gmail")

	return nil
}

// BuildMessage constructs an RFC 2822 email message with HTML content.
func BuildMessage(sender Recipient, recipients []Recipient, subject, body string) string {
	toAddrs := make([]string, 0, len(recipients))
	for _, r := range recipients {
		if r.Name != "" {
			toAddrs = append(toAddrs, fmt.Sprintf("%s <%s>", r.Name, r.Address))
		} else {
			toAddrs = append(toAddrs, r.Address)
		}
	}

	fromHeader := ""
	if sender.Address != "" {
		fromHeader = fmt.Sprintf("From: %s <%s>\r\n", sender.Name, sender.Address)
	}

	return fmt.Sprintf(
		"%sTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		fromHeader,
		strings.Join(toAddrs, ", "),
		subject,
		body,
	)
}

// getAuthCodeBrowser starts a local HTTP server, opens the browser for
// OAuth authorization, and waits for the callback with the auth code.
func getAuthCodeBrowser(ctx context.Context, config *oauth2.Config) (string, error) {
	lc := net.ListenConfig{}

	listener, err := lc.Listen(ctx, "tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("listening on local port: %w", err)
	}

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return "", fmt.Errorf("unexpected listener address type: %T", listener.Addr())
	}

	config.RedirectURL = fmt.Sprintf("http://localhost:%d/", tcpAddr.Port)

	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("generating state token: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(stateBytes)
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOnline)

	logrus.Info("Opening browser for OAuth authorization")

	if err := openBrowserURL(ctx, authURL); err != nil {
		logrus.Warnf("Could not open browser: %v", err)
		fmt.Printf("Please open this URL manually:\n\n  %s\n\n", authURL)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Authentication successful! You may close this window.")

			codeCh <- r.URL.Query().Get("code")
		}),
	}

	go func() {
		if serveErr := srv.Serve(listener); serveErr != http.ErrServerClosed {
			errCh <- serveErr
		}
	}()

	select {
	case code := <-codeCh:
		srv.Close()

		return code, nil
	case err := <-errCh:
		return "", fmt.Errorf("local OAuth server: %w", err)
	}
}

// getAuthCodeManual prints the auth URL and prompts the user to paste
// the redirect URL containing the authorization code.
func getAuthCodeManual(config *oauth2.Config) (string, error) {
	config.RedirectURL = "http://localhost"
	authURL := config.AuthCodeURL("", oauth2.AccessTypeOnline)

	fmt.Println("Open the following URL in your browser and authorize the application:")
	fmt.Printf("\n  %s\n\n", authURL)
	fmt.Println("After authorization, paste the full redirect URL here:")

	var redirectURLStr string
	if _, err := fmt.Scan(&redirectURLStr); err != nil {
		return "", fmt.Errorf("reading redirect URL: %w", err)
	}

	parsedURL, err := url.Parse(redirectURLStr)
	if err != nil {
		return "", fmt.Errorf("parsing redirect URL: %w", err)
	}

	code := parsedURL.Query().Get("code")
	if code == "" {
		return "", errors.New("authorization code not found in URL")
	}

	return code, nil
}

// openBrowserURL opens the given URL in the default browser.
func openBrowserURL(ctx context.Context, rawURL string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.CommandContext(ctx, "xdg-open", rawURL).Start()
	case "darwin":
		return exec.CommandContext(ctx, "open", rawURL).Start()
	case "windows":
		return exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", rawURL).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

package mailer

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type MailerConfig struct {
	From         string
	ClientID     string
	ClientSecret string
	RefreshToken string
}

type Mailer struct {
	mConfig MailerConfig
	svc     *gmail.Service
}

func New(ctx context.Context, mConfig MailerConfig) (*Mailer, error) {
	conf := &oauth2.Config{
		ClientID:     mConfig.ClientID,
		ClientSecret: mConfig.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.GmailSendScope},
	}

	tokenSource := conf.TokenSource(ctx, &oauth2.Token{
		RefreshToken: mConfig.RefreshToken,
	})

	svc, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("mailer: creating gmail service: %w", err)
	}

	return &Mailer{mConfig: mConfig, svc: svc}, nil
}

func (m *Mailer) send(to, subject, htmlBody string) error {
	headers := strings.Join([]string{
		"From: " + m.mConfig.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
	}, "\r\n")

	raw := headers + "\r\n\r\n" + htmlBody

	msg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(raw)),
	}

	_, err := m.svc.Users.Messages.Send("me", msg).Do()
	if err != nil {
		return fmt.Errorf("mailer: sending message: %w", err)
	}

	return nil
}

// SendOTP sends
func (m *Mailer) SendOTP(to, code, purpose string) error {
	subject := "Your verification code"

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; color: #222222; line-height: 1.5; max-width: 480px; margin: 0 auto; padding: 24px;">
  <p>Hello,</p>
  <p>Your verification code for <strong>%s</strong> is:</p>
  <p style="font-size: 28px; font-weight: bold; letter-spacing: 6px; margin: 20px 0; background: #f4f4f5; padding: 12px 16px; border-radius: 6px; display: inline-block;">%s</p>
  <p>This code expires in 10 minutes. Do not share it with anyone.</p>
  <p>If you did not request this code, please ignore this email.</p>
  <p style="margin-top: 24px;">— The Nex Play Team</p>
</body>
</html>`, purpose, code)

	return m.send(to, subject, body)
}

package mailer

import (
	"context"
	"fmt"
	"time"

	"github.com/mailersend/mailersend-go"
)

type MailerConfig struct {
	APIKey   string
	From     string
	FromName string
}

type Mailer struct {
	client *mailersend.Mailersend
	config MailerConfig
}

func New(config MailerConfig) *Mailer {
	return &Mailer{
		client: mailersend.NewMailersend(config.APIKey),
		config: config,
	}
}

func (m *Mailer) send(to, subject, text, html string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := m.client.Email.NewMessage()

	message.SetFrom(mailersend.From{
		Name:  m.config.FromName,
		Email: m.config.From,
	})

	message.SetRecipients([]mailersend.Recipient{{
		Email: to,
	}})

	message.SetSubject(subject)
	message.SetText(text)
	message.SetHTML(html)

	_, err := m.client.Email.Send(ctx, message)
	return err
}

func (m *Mailer) SendOTP(to, code, purpose string) error {
	subject := "Your verification code"

	text := fmt.Sprintf(`Hello,

Your verification code for %s is:

    %s

This code expires in 10 minutes.
Do not share this code with anyone.

If you did not request this code, please ignore this email.

— The Nex Play Team`, purpose, code)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family:Arial,sans-serif;">
  <h2>Your verification code</h2>
  <p>Your verification code for <strong>%s</strong> is:</p>
  <div style="font-size:32px;font-weight:bold;letter-spacing:4px;padding:16px;background:#f5f5f5;display:inline-block;">%s</div>
  <p>This code expires in <strong>10 minutes</strong>.</p>
  <p>If you didn't request this code, you can safely ignore this email.</p>
  <p>— The Nex Play Team</p>
</body>
</html>`, purpose, code)

	return m.send(to, subject, text, html)
}

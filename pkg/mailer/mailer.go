package mailer

import (
	"fmt"
	"net/smtp"
	"strings"
)

type MailerConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

type Mailer struct {
	mConfig MailerConfig
}

// New mailer
func New(mConfig MailerConfig) *Mailer {

	return &Mailer{mConfig: mConfig}
}

// Send
func (m *Mailer) send(to, subject, body string) error {

	addr := fmt.Sprintf("%s:%d", m.mConfig.Host, m.mConfig.Port)

	auth := smtp.PlainAuth("", m.mConfig.User, m.mConfig.Pass, m.mConfig.Host)

	headers := strings.Join([]string{

		"From: " + m.mConfig.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}, "\r\n")

	msg := headers + "\r\n\r\n" + body

	return smtp.SendMail(addr, auth, m.mConfig.From, []string{to}, []byte(msg))
}

// Send OTP
func (m *Mailer) SendOTP(to, code, purpose string) error {

	subject := "Your verification code"

	body := fmt.Sprintf(`Hello,

Your verification code for %s is:

    %s

This code expires in 10 minutes. Do not share it with anyone.

If you did not request this code, please ignore this email.

— The Nex Play Team
`, purpose, code)

	return m.send(to, subject, body)
}

package auth

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/config"
)

type Mailer interface {
	Send(to, subject, body string) error
}

// newMailer returns an SMTP sender when SMTP_HOST is set, otherwise a
// dev-mode mailer that logs messages instead of sending them.
func newMailer() Mailer {
	host := config.Env("SMTP_HOST", "")
	if host == "" {
		return LogMailer{}
	}
	return SMTPMailer{
		host: host,
		port: config.Env("SMTP_PORT", "587"),
		user: config.Env("SMTP_USER", ""),
		pass: config.Env("SMTP_PASSWORD", ""),
		from: config.Env("SMTP_FROM", "no-reply@ai-buddy"),
	}
}

type SMTPMailer struct {
	host, port, user, pass, from string
}

func (m SMTPMailer) Send(to, subject, body string) error {
	addr := m.host + ":" + m.port
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	return smtp.SendMail(addr, auth, m.from, []string{to}, buildMessage(m.from, to, subject, body))
}

type LogMailer struct{}

func (LogMailer) Send(to, subject, body string) error {
	slog.Info("email not sent: SMTP is not configured (dev mode)", "to", to, "subject", subject, "body", body)
	return nil
}

func buildMessage(from, to, subject, body string) []byte {
	encoded := base64.StdEncoding.EncodeToString([]byte(subject))
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: =?UTF-8?B?" + encoded + "?=",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
	return []byte(msg)
}

// confirmationLinkText builds the plain-text letter body with the link.
func confirmationLinkText(origin, token string) string {
	return fmt.Sprintf("Подтвердите почту, перейдя по ссылке: %s/api/v1/auth/confirm?token=%s\n"+
		"Ссылка действует 24 часа.", strings.TrimRight(origin, "/"), token)
}

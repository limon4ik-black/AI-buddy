package auth

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"

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

// Send dials IPv4 explicitly: gmail answers AAAA records too, and in networks
// without a working IPv6 route the v6 attempt fails before v4 is tried.
func (m SMTPMailer) Send(to, subject, body string) error {
	conn, err := m.dial()
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return err
	}
	defer c.Close()
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
			return err
		}
	}
	if m.user != "" {
		if err = c.Auth(smtp.PlainAuth("", m.user, m.pass, m.host)); err != nil {
			return err
		}
	}
	if err = c.Mail(m.from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(buildMessage(m.from, to, subject, body)); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func (m SMTPMailer) dial() (net.Conn, error) {
	if ips, err := net.LookupIP(m.host); err == nil {
		var last error
		for _, ip := range ips {
			if v4 := ip.To4(); v4 == nil {
				continue
			}
			d := net.Dialer{Timeout: 15 * time.Second}
			conn, err := d.Dial("tcp", net.JoinHostPort(ip.String(), m.port))
			if err == nil {
				return conn, nil
			}
			last = err
		}
		if last != nil {
			return nil, last
		}
	}
	return net.DialTimeout("tcp", net.JoinHostPort(m.host, m.port), 15*time.Second)
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

// loginLinkText builds the plain-text letter body with the link.
func loginLinkText(origin, token string) string {
	return fmt.Sprintf("Войти в AI-buddy: %s/api/v1/auth/confirm?token=%s\n"+
		"Ссылка действует 15 минут и одноразовая. Если вход не ты — просто игнорируй письмо.",
		strings.TrimRight(origin, "/"), token)
}

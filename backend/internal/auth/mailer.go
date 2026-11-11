package auth

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/config"
	"golang.org/x/net/proxy"
)

type Mailer interface {
	Send(to, subject, text, html string) error
}

func newMailer() Mailer {
	host := config.Env("SMTP_HOST", "")
	if host == "" {
		return LogMailer{}
	}
	m := SMTPMailer{
		host: host,
		port: config.Env("SMTP_PORT", "587"),
		user: config.Env("SMTP_USER", ""),
		pass: config.Env("SMTP_PASSWORD", ""),
		from: config.Env("SMTP_FROM", "no-reply@ai-buddy"),
	}
	if raw := config.Env("HTTPS_PROXY", ""); strings.HasPrefix(raw, "socks5://") {
		if u, err := url.Parse(raw); err == nil && u.Host != "" {
			if d, err := proxy.SOCKS5("tcp", u.Host, nil, proxy.Direct); err == nil {
				m.dialer = d
			}
		}
	}
	return m
}

type SMTPMailer struct {
	host, port, user, pass, from string
	dialer                       proxy.Dialer
}

func (m SMTPMailer) Send(to, subject, text, html string) error {
	conn, err := m.dial()
	if err != nil {
		return err
	}
	conn.SetDeadline(time.Now().Add(30 * time.Second))
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
	if _, err = w.Write(buildMessage(m.from, to, subject, text, html)); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func (m SMTPMailer) dial() (net.Conn, error) {
	if m.dialer != nil {
		return m.dialer.Dial("tcp", net.JoinHostPort(m.host, m.port))
	}
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

func (LogMailer) Send(to, subject, text, _ string) error {
	slog.Info("email not sent: SMTP is not configured (dev mode)", "to", to, "subject", subject, "body", text)
	return nil
}

func buildMessage(from, to, subject, text, html string) []byte {
	encoded := base64.StdEncoding.EncodeToString([]byte(subject))
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: =?UTF-8?B?" + encoded + "?=",
		"MIME-Version: 1.0",
		`Content-Type: multipart/alternative; boundary="buddy-mail"`,
		"",
		"--buddy-mail",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: base64",
		"",
		base64.StdEncoding.EncodeToString([]byte(text)),
		"--buddy-mail",
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: base64",
		"",
		base64.StdEncoding.EncodeToString([]byte(html)),
		"--buddy-mail--",
		"",
	}, "\r\n")
	return []byte(msg)
}

func loginCodeText(code string) string {
	return fmt.Sprintf("Код для входа в AI-buddy: %s\n"+
		"Код действует 7 минут и одноразовый. Если вход не ты — просто игнорируй письмо.", code)
}

func loginCodeHTML(code string) string {
	dot := func(size, color, style string) string {
		return `<span style="display:inline-block;width:` + size + `px;height:` + size + `px;border-radius:50%;background:` + color + `;` + style + `"></span>`
	}
	logo := dot("14", "#00aaff", "position:relative;top:2px;") +
		dot("10", "#ff4053", "margin:0 2px 0 3px;position:relative;top:-4px;") +
		dot("10", "#97cf26", "margin-right:3px;") +
		dot("15", "#965eeb", "position:relative;top:-1px;")
	return `<!doctype html>
<html><body style="margin:0;padding:0;background:#f2f3f5;font-family:Arial,Helvetica,sans-serif;">
<div style="max-width:520px;margin:0 auto;padding:36px 16px;">
  <table role="presentation" width="100%" style="background:#ffffff;border:1px solid #e9e9ee;border-radius:20px;box-shadow:0 8px 40px rgba(0,0,0,0.03);border-collapse:collapse;">
    <tr><td style="padding:26px 30px 0;">
      <span style="font-size:24px;font-weight:750;letter-spacing:-1px;color:#171717;">` + logo + `&nbsp;ai&nbsp;buddy<span style="color:#00aaff;">.</span></span>
    </td></tr>
    <tr><td style="padding:6px 30px 0;">
      <span style="display:inline-block;background:#f3f6ed;border:1px solid #e9eedf;color:#7c8d60;border-radius:20px;padding:6px 10px;font-size:9px;font-weight:600;letter-spacing:1px;">ВХОД В ПРОСТРАНСТВО</span>
    </td></tr>
    <tr><td style="padding:14px 30px 0;">
      <h1 style="margin:0 0 10px;font-size:27px;letter-spacing:-0.8px;color:#171717;">Твой код для входа</h1>
      <p style="margin:0 0 22px;color:#444a53;font-size:12px;line-height:1.8;">
        Введи эти шесть цифр на странице входа — и ты внутри.
        Код одноразовый и живёт 7&nbsp;минут.
      </p>
    </td></tr>
    <tr><td style="padding:0 30px 8px;">
      <div style="background:#f7f7f8;border:1px solid #ebebed;border-radius:14px;padding:24px 14px;text-align:center;">
        <span style="font-size:42px;font-weight:750;letter-spacing:12px;color:#171717;">` + code + `</span>
      </div>
    </td></tr>
    <tr><td style="padding:16px 30px 30px;">
      <p style="margin:0;color:#777783;font-size:11px;line-height:1.8;">
        Код нужен только для этого входа и привязан к твоему браузеру.
        Если вход в AI Buddy не ты — просто игнорируй письмо.
      </p>
    </td></tr>
    <tr><td style="padding:0 30px 24px;">
      <p style="margin:0;border-top:1px solid #ebebed;padding-top:14px;color:#9aa0aa;font-size:9px;letter-spacing:2px;">
        AI BUDDY · ПРОСТРАНСТВО АДАПТАЦИИ · ХАКАТОН 2026
      </p>
    </td></tr>
  </table>
  <p style="margin:16px 0 0;text-align:center;color:#9aa0aa;font-size:10px;">Если письмо попало к тебе случайно — просто удали его.</p>
</div>
</body></html>`
}

package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string

	FromEmail string
	FromName  string

	UseTLS bool // STARTTLS
}

type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) SendOTP(ctx context.Context, toEmail string, code string) error {
	subject := "Код входа V2B"
	text := fmt.Sprintf(
		"Ваш код подтверждения: %s\n\nКод действует ограниченное время.\nЕсли вы не запрашивали код — просто проигнорируйте это письмо.",
		code,
	)

	fromHeader := m.cfg.FromEmail
	if m.cfg.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.FromEmail)
	}

	msg := strings.Join([]string{
		"From: " + fromHeader,
		"To: " + toEmail,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		text,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	d := net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return err
	}
	defer func() { _ = c.Quit() }()

	// STARTTLS
	if m.cfg.UseTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			tlsCfg := &tls.Config{ServerName: m.cfg.Host}
			if err := c.StartTLS(tlsCfg); err != nil {
				return err
			}
		}
	}

	// Auth
	if m.cfg.Username != "" {
		auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
		if err := c.Auth(auth); err != nil {
			return err
		}
	}

	if err := c.Mail(m.cfg.FromEmail); err != nil {
		return err
	}
	if err := c.Rcpt(toEmail); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

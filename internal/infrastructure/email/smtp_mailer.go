// Package email implements the domain's ports.Mailer using stdlib
// transports only, matching this repo's dependency philosophy.
package email

import (
	"fmt"
	"net/mail"
	"net/smtp"

	"github.com/zaidmasri/business-planning-tool/internal/domain/ports"
)

// SMTPMailer sends email via net/smtp. It authenticates with PLAIN auth
// when a username is configured, and sends unauthenticated otherwise (e.g.
// a local relay/sandbox SMTP server).
type SMTPMailer struct {
	host, port, username, password, from string
}

func NewSMTPMailer(host, port, username, password, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, username: username, password: password, from: from}
}

var _ ports.Mailer = (*SMTPMailer)(nil)

func (m *SMTPMailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)

	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	envelopeFrom := m.from
	if parsed, err := mail.ParseAddress(m.from); err == nil {
		envelopeFrom = parsed.Address
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		m.from, to, subject, body,
	)

	if err := smtp.SendMail(addr, auth, envelopeFrom, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}
	return nil
}

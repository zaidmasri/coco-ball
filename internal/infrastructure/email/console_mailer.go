package email

import (
	"log"

	"github.com/zaidmasri/business-planning-tool/internal/domain/ports"
)

// ConsoleMailer logs emails instead of sending them. It's the default when
// SMTP isn't configured, so local development and the worker's smoke tests
// don't need a real mail server.
type ConsoleMailer struct{}

func NewConsoleMailer() *ConsoleMailer { return &ConsoleMailer{} }

var _ ports.Mailer = (*ConsoleMailer)(nil)

func (m *ConsoleMailer) Send(to, subject, body string) error {
	log.Printf("email (console): to=%s subject=%q\n%s", to, subject, body)
	return nil
}

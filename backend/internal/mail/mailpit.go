package mail

import (
	"fmt"
	"net/smtp"

	"github.com/joaopsramos/fincon/internal/config"
)

type MailPit struct {
	addr     string
	auth     smtp.Auth
	defaults config.MailDefaults
}

func NewMailPit(defaults config.MailDefaults) Mailer {
	host := "localhost"
	port := "1025"
	auth := smtp.PlainAuth("", "", "", host)

	return &MailPit{
		addr:     host + ":" + port,
		auth:     auth,
		defaults: defaults,
	}
}

func (m *MailPit) Send(email Email) error {
	var sender string

	if email.From.Email == "" {
		sender = m.defaults.From.Email
	}

	body, err := email.BuildBody()
	if err != nil {
		return err
	}

	mime := "Content-Type: text/html; charset=\"UTF-8\";"
	headers := "Subject: " + email.Subject

	message := fmt.Sprintf("%s\r\n%s\r\n%s\r\n\r\n", mime, headers, body)

	return smtp.SendMail(m.addr, m.auth, sender, []string{email.To.Email}, []byte(message))
}

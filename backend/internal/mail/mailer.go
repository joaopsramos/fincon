package mail

import (
	"bytes"
	"html/template"
	"path/filepath"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/types"
)

type (
	EmailTemplate string
	EmailSubject  string
)

const (
	ForgotPasswordTemplate EmailTemplate = "forgot_password"

	ForgotPasswordSubject EmailSubject = "[Fincon] Recuperação de senha"
)

type Email struct {
	To       types.MailContact
	From     types.MailContact
	Subject  EmailSubject
	Template EmailTemplate
	Data     any
}

type Mailer interface {
	Send(email Email) error
}

func NewMailer() Mailer {
	mailConfig := config.NewMailConfig()

	switch mailConfig.Driver {
	case config.MailPit:
		return NewMailPit(mailConfig.Defaults)
	case config.SES:
		// TODO: implement
		return &SES{}
	default:
		panic("invalid mail driver")
	}
}

func ParseTemplate(tmpl EmailTemplate) (*template.Template, error) {
	templatePath := filepath.Join("internal", "mail", "templates", string(tmpl)+".html")
	return template.ParseFiles(templatePath)
}

func (m *Email) BuildBody() (string, error) {
	tmpl, err := ParseTemplate(m.Template)
	if err != nil {
		return "", err
	}

	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, m.Data); err != nil {
		return "", err
	}

	return htmlBuffer.String(), nil
}

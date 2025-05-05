package config

import (
	"os"

	"github.com/joaopsramos/fincon/internal/types"
)

type Driver string

const (
	MailPit Driver = "mailpit"
	SES     Driver = "ses"
)

type MailConfig struct {
	Driver   Driver
	Defaults MailDefaults
}

type MailDefaults struct {
	From types.MailContact
}

func NewMailConfig() *MailConfig {
	driver := os.Getenv("MAIL_DRIVER")
	name := os.Getenv("MAIL_FROM_NAME")
	email := os.Getenv("MAIL_FROM_EMAIL")

	return &MailConfig{
		Driver: Driver(driver),
		Defaults: MailDefaults{
			From: types.MailContact{Name: name, Email: email},
		},
	}
}

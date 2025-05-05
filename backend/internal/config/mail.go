package config

import (
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
	cfg := Get()

	return &MailConfig{
		Driver: Driver(cfg.MailDriver),
		Defaults: MailDefaults{
			From: types.MailContact{Name: cfg.MailFromName, Email: cfg.MailFromEmail},
		},
	}
}

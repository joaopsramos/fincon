package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/honeybadger-io/honeybadger-go"
	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/mail"
)

func init() {
	config.Load(".")
}

func main() {
	cfg := config.Get()

	honeybadger.Configure(honeybadger.Configuration{APIKey: cfg.HoneybadgerAPIKey})
	defer honeybadger.Monitor()

	db := config.NewPostgresConn(cfg.PostgresDSN())
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mailer := mail.NewMailer()

	api := api.NewApp(db, logger, mailer)

	api.SetupAll()
	log.Fatal(api.Listen())
}

package main

import (
	"log"

	"github.com/honeybadger-io/honeybadger-go"
	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/config"
)

func init() {
	config.Load(".")
}

func main() {
	cfg := config.Get()

	honeybadger.Configure(honeybadger.Configuration{APIKey: cfg.HoneybadgerAPIKey})
	defer honeybadger.Monitor()

	db := config.NewPostgresConn(cfg.PostgresDSN())
	api := api.NewApp(db)
	api.SetupAll()
	log.Fatal(api.Listen())
}

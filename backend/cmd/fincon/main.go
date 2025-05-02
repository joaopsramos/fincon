package main

import (
	"log"
	"os"

	"github.com/honeybadger-io/honeybadger-go"
	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/config"
)

func init() {
	config.LoadEnv(".")
}

func main() {
	honeybadger.Configure(honeybadger.Configuration{APIKey: os.Getenv("HONEYBADGER_API_KEY")})
	defer honeybadger.Monitor()

	db := config.NewPostgresConn(config.PostgresDSNFromEnv())
	api := api.NewApp(db)
	api.SetupAll()
	log.Fatal(api.Listen())
}

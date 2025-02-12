package main

import (
	"log"

	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/config"
)

func init() {
	config.LoadEnv(".")
}

func main() {
	db := config.NewPostgresConn(config.PostgresDSNFromEnv())

	api := api.NewApi(db)
	api.SetupAll()
	log.Fatal(api.Listen())
}

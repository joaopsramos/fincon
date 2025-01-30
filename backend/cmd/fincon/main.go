package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	dns := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASS"),
		os.Getenv("POSTGRES_DB"),
	)
	db := config.NewPostgresConn(dns)

	api := api.NewApi(db)
	api.Setup()
	log.Fatal(api.Listen())
}

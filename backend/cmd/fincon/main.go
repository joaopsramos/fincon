package main

import (
	"os"

	"github.com/joaopsramos/fincon/internal/api"
	"github.com/joaopsramos/fincon/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	db := config.ConnectAndSetup(os.Getenv("SQLITE_PATH"))

	api := api.NewApi(db)
	api.SetupAndListen()
}

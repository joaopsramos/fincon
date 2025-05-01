package main

import (
	"fmt"
	"os"

	"github.com/joaopsramos/fincon/internal/config"
	"golang.org/x/exp/slog"
)

func init() {
	config.LoadEnv(".")
}

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASS"),
	)

	db := config.NewPostgresConn(dsn)

	slog.Info("Creating datbase...")
	tx := db.Exec("CREATE DATABASE " + os.Getenv("POSTGRES_DB"))
	if tx.Error != nil {
		panic(tx.Error)
	}
}

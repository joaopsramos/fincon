package main

import (
	"github.com/joaopsramos/fincon/internal/config"
	"golang.org/x/exp/slog"
)

func init() {
	config.Load(".")
}

func main() {
	cfg := config.Get()
	db := config.NewPostgresConn(cfg.PostgresDSN())

	slog.Info("Creating database...")
	tx := db.Exec("CREATE DATABASE " + cfg.Database.Name)
	if err := tx.Error; err != nil {
		panic(err)
	}
}

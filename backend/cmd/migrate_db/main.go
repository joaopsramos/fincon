package main

import (
	"log/slog"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/domain"
)

func init() {
	config.LoadEnv(".")
}

func main() {
	db := config.NewPostgresConn(config.PostgresDSNFromEnv())

	slog.Info("Creating extension 'unnaccent'")
	db.Exec("CREATE EXTENSION IF NOT EXISTS unaccent")

	slog.Info("Auto migrating...")
	db.AutoMigrate(&domain.Goal{}, &domain.Salary{}, &domain.Expense{})
}

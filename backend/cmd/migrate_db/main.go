package main

import (
	"log"
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

	slog.Info("Creating extension 'citext'")
	db.Exec("CREATE EXTENSION IF NOT EXISTS citext")

	slog.Info("Auto migrating...")
	err := db.AutoMigrate(&domain.User{}, &domain.Goal{}, &domain.Salary{}, &domain.Expense{})
	if err != nil {
		log.Fatal(err)
	}
}

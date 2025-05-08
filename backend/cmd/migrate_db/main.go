package main

import (
	"log"
	"log/slog"

	"github.com/joaopsramos/fincon/internal/config"
	"github.com/joaopsramos/fincon/internal/domain"
)

func init() {
	config.Load(".")
}

func main() {
	db := config.NewPostgresConn(config.Get().PostgresDSN())

	slog.Info("Creating extension 'unnaccent'")
	db.Exec("CREATE EXTENSION IF NOT EXISTS unaccent")

	slog.Info("Creating extension 'citext'")
	db.Exec("CREATE EXTENSION IF NOT EXISTS citext")

	slog.Info("Auto migrating...")
	err := db.AutoMigrate(
		&domain.User{},
		&domain.Goal{},
		&domain.Salary{},
		&domain.Expense{},
		&domain.UserToken{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

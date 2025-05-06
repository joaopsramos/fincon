package main

import (
	"fmt"

	"github.com/joaopsramos/fincon/internal/config"
	"golang.org/x/exp/slog"
)

func init() {
	config.Load(".")
}

func main() {
	cfg := config.Get()
	dbCfg := cfg.Database
	// Can't use config.PostgresDSN() because it contains the database name, which doesn't exist yet
	db := config.NewPostgresConn(fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=disable",
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.User,
		dbCfg.Pass,
	))

	slog.Info("Creating database...")
	tx := db.Exec("CREATE DATABASE " + cfg.Database.Name)
	if err := tx.Error; err != nil {
		panic(err)
	}
}

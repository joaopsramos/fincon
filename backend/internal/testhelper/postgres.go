package testhelper

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/joaopsramos/fincon/internal/config"
	"gorm.io/gorm"
)

func NewTestPostgresTx(t *testing.T) *gorm.DB {
	os.Setenv("APP_ENV", "test")
	config.LoadEnv(path.Join("..", ".."))

	dns := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UCT",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASS"),
		os.Getenv("POSTGRES_DB"),
	)
	db := config.NewPostgresConn(dns)

	tx := db.Begin()

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}
